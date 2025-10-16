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
func NewJWTMiddleware(issuer, jwksURI string) (*jwtmiddleware.JWTMiddleware, error) {
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %w", err)
	}
	opts := []any{}
	if jwksURI != "" {
		keysURI, err := url.Parse(jwksURI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse jwksURI: %w", err)
		}
		opts = append(opts, jwks.WithCustomJWKSURI(keysURI))
	}
	provider := jwks.NewCachingProvider(issuerURL, 1*time.Minute, opts...)
	newCustomClaims := func() validator.CustomClaims {
		return &TelemetryClaim{}
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
		jwtmiddleware.WithErrorHandler(ErrorHandler),
		jwtmiddleware.WithCredentialsOptional(true),
	)
	return middleware, nil
}

// AddClaimHandler is a middleware that fills in GraphQL-friendly privilege information on
// the *TelemetryClaim object in the context.
func AddClaimHandler(next http.Handler, vehicleAddr, mfrAddr common.Address) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetValidatedClaims(r.Context())
		if !ok || claims.CustomClaims == nil {
			// unauthorized calls will not have a claims.
			next.ServeHTTP(w, r)
			return
		}

		telClaim, ok := claims.CustomClaims.(*TelemetryClaim)
		if !ok {
			zerolog.Ctx(r.Context()).Error().Msg("Could not cast claims to TelemetryClaim")
			jwtmiddleware.DefaultErrorHandler(w, r, jwtmiddleware.ErrJWTMissing)
			return
		}
		// add the custom claims to the context under a new custom key
		r = r.Clone(context.WithValue(r.Context(), TelemetryClaimContextKey{}, telClaim))
		next.ServeHTTP(w, r)
	})
}

// GetValidatedClaims returns the validated claims from the request context.
func GetValidatedClaims(ctx context.Context) (*validator.ValidatedClaims, bool) {
	claim := ctx.Value(jwtmiddleware.ContextKey{})
	if claim == nil {
		return nil, false
	}
	validateClaim, ok := claim.(*validator.ValidatedClaims)
	if !ok {
		return nil, false
	}
	return validateClaim, true
}

// ErrorHandler is a custom error handler for the jwt middleware. It logs the error and then calls the default error handler.
func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	zerolog.Ctx(r.Context()).Error().Err(err).Msg("error validating token")
	jwtmiddleware.DefaultErrorHandler(w, r, err)
}
