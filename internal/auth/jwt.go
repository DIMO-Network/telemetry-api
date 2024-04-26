package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// NewJWTMiddleware creates a new JWT middleware with the given issuer and contract address.
// This middleware will validate the token and add the claim to the context.
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
		return &TelemetryClaim{expectedContractAddress: expectedAddr}
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

// AddClaimHandler is a middleware that adds the telemetry claim to the context.
func AddClaimHandler(next http.Handler, logger *zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if !ok || claims == nil || claims.CustomClaims == nil {
			// unathorized calls will not have a claims.
			next.ServeHTTP(w, r)
			return
		}

		telClaim, ok := claims.CustomClaims.(*TelemetryClaim)
		if !ok {
			logger.Error().Msg("could not cast claims to telemetyClaim")
			jwtmiddleware.DefaultErrorHandler(w, r, jwtmiddleware.ErrJWTMissing)
		}
		telClaim.SetPrivileges()

		// add the custom claims to the context under a new custom key
		r = r.Clone(context.WithValue(r.Context(), TelemetryClaimContextKey{}, telClaim))
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
