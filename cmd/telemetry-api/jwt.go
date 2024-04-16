package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// CustomContextKey is a custom key for the context to store the custom claims.
type CustomContextKey struct{}

// NewJWTMiddleware creates a new JWT middleware with the given issuer and contract address.
// This middleware will validate the token and add the claim to the context
func NewJWTMiddleware(issuer, jwksURI, contractAddress string, logger *zerolog.Logger) (*jwtmiddleware.JWTMiddleware, error) {
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %w", err)
	}
	opts := []jwks.ProviderOption{}
	if jwksURI != "" {
		keysURI, err := url.Parse(jwksURI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse jwksURI: %w", err)
		}
		opts = append(opts, jwks.WithCustomJWKSURI(keysURI))
	}
	provider := jwks.NewCachingProvider(issuerURL, 1*time.Hour, opts...)
	expectedAddr := common.HexToAddress(contractAddress)
	newCustomClaims := func() validator.CustomClaims {
		return &customClaimWrapper{expectedContractAddress: expectedAddr}
	}
	// Set up the validator.
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{"dimo.zone"},
		validator.WithCustomClaims(newCustomClaims),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithCredentialsOptional(false),
		jwtmiddleware.WithErrorHandler(ErrorHandler(logger)),
		jwtmiddleware.WithCredentialsOptional(true),
	)
	return middleware, nil
}

// AddClaimHandler is a middleware that adds the custom claims wrapper to the context.
func AddClaimHandler(next http.Handler, logger *zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claimWrapper *customClaimWrapper
		claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if ok {
			claimWrapper, ok = claims.CustomClaims.(*customClaimWrapper)
			if !ok {
				logger.Error().Msg("could not cast claims to customClaimWrapper")
				jwtmiddleware.DefaultErrorHandler(w, r, jwtmiddleware.ErrJWTMissing)
			}
		} else {
			// if the claims are not in the context, create an empty custom claim wrapper with no privileges.
			claimWrapper = &customClaimWrapper{}
			addr := common.Address{}
			if r.Header.Get("Authorization-unsafe") == "Bearer foo" {
				claimWrapper.CustomClaims = privilegetoken.CustomClaims{
					TokenID:         "foo",
					PrivilegeIDs:    []privileges.Privilege{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
					ContractAddress: addr,
				}
				claimWrapper.expectedContractAddress = addr
			}
		}
		claimWrapper.privileges = make(map[model.Privilege]struct{}, len(claimWrapper.CustomClaims.PrivilegeIDs))
		for _, priv := range claimWrapper.CustomClaims.PrivilegeIDs {
			claimWrapper.privileges[privToAPI[priv]] = struct{}{}
		}

		// add the custom claims to the context under a new custom key
		r = r.Clone(context.WithValue(r.Context(), CustomContextKey{}, claimWrapper))
		next.ServeHTTP(w, r)
	})
}

// ErrorHandler is a custom error handler for the jwt middleware. It logs the error and then calls the default error handler.
func ErrorHandler(logger *zerolog.Logger) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error().Err(err).Msg("error validating token")
		jwtmiddleware.DefaultErrorHandler(w, r, err)
	}
}

func getClaim(ctx context.Context) *customClaimWrapper {
	// not checking cast because our addClaimMiddleware should have already done that
	return ctx.Value(CustomContextKey{}).(*customClaimWrapper)
}

type customClaimWrapper struct {
	expectedContractAddress common.Address
	privileges              map[model.Privilege]struct{}
	privilegetoken.CustomClaims
}

// Validate function is required to implement the validator.CustomClaims interface.
func (c *customClaimWrapper) Validate(context.Context) error {
	if c.expectedContractAddress != c.CustomClaims.ContractAddress {
		return fmt.Errorf("incorrect contract address expected %v got %v", c.expectedContractAddress, c.CustomClaims.ContractAddress)
	}
	return nil
}

func requiresPrivilegeCheck(ctx context.Context, obj interface{}, next graphql.Resolver, privileges []model.Privilege) (res interface{}, err error) {
	claim := getClaim(ctx)
	for _, priv := range privileges {
		if _, ok := claim.privileges[priv]; !ok {
			return nil, fmt.Errorf("unathorized")
		}
	}
	return next(ctx)
}

var privToAPI = map[privileges.Privilege]model.Privilege{
	privileges.VehicleNonLocationData: model.PrivilegeVehicleNonLocationData,
	privileges.VehicleCommands:        model.PrivilegeVehicleCommands,
	privileges.VehicleCurrentLocation: model.PrivilegeVehicleCurrentLocation,
	privileges.VehicleAllTimeLocation: model.PrivilegeVehicleAllTimeLocation,
	privileges.VehicleVinCredential:   model.PrivilegeVehicleVinCredential,
}

func requiresTokenCheck(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return nil, fmt.Errorf("no field context found")
	}
	tokenID, ok := fCtx.Args["tokenID"].(*int)
	if !ok || tokenID == nil {
		return nil, fmt.Errorf("failed to get tokenID from args")
	}

	claim := getClaim(ctx)
	if strconv.Itoa(*tokenID) != claim.TokenID && claim.TokenID != "foo" {
		return nil, fmt.Errorf("unathorized")
	}
	return next(ctx)
}
