package auth

import (
	"context"

	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
)

// TelemetryClaimContextKey is a custom key for the context to store the custom claims.
type TelemetryClaimContextKey struct{}

// TelemetryClaim is a custom claim for the telemetry API.
type TelemetryClaim struct {
	privileges map[model.Privilege]struct{}
	privilegetoken.CustomClaims
}

// SetPrivileges sets the privileges from the embedded CustomClaims.
func (t *TelemetryClaim) SetPrivileges() {
	t.privileges = make(map[model.Privilege]struct{}, len(t.CustomClaims.PrivilegeIDs))
	for _, priv := range t.CustomClaims.PrivilegeIDs {
		t.privileges[privToAPI[priv]] = struct{}{}
	}
}

// Validate function is required to implement the validator.CustomClaims interface.
func (t *TelemetryClaim) Validate(context.Context) error {
	return nil
}

func getTelemetryClaim(ctx context.Context) (*TelemetryClaim, error) {
	claim, ok := ctx.Value(TelemetryClaimContextKey{}).(*TelemetryClaim)
	if !ok || claim == nil {
		return nil, jwtmiddleware.ErrJWTMissing
	}
	return claim, nil
}
