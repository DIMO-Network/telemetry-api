package auth

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/ethereum/go-ethereum/common"
)

// TelemetryClaimContextKey is a custom key for the context to store the custom claims.
type TelemetryClaimContextKey struct{}

// TelemetryClaim is a custom claim for the telemetry API.
type TelemetryClaim struct {
	privileges map[model.Privilege]struct{}
	privilegetoken.CustomClaims
	contractToPrivs  map[common.Address]map[model.Privilege]struct{}
	vehicleNFTAddr   common.Address
	manufacturerAddr common.Address
}

// SetPrivileges sets the privileges from the embedded CustomClaims.
func (t *TelemetryClaim) SetPrivileges() {
	for _, priv := range t.CustomClaims.PrivilegeIDs {
		privName, okV := vehiclePrivileges[priv]
		if okV {
			t.privileges[privName] = struct{}{}
		}
		// privName, okM := manufacturerPrivileges[priv]
		// if okM {
		// 	t.privileges[privName] = struct{}{}
		// }
	}
}

// Validate function is required to implement the validator.CustomClaims interface.
func (t *TelemetryClaim) Validate(context.Context) error {
	if t.CustomClaims.ContractAddress != t.vehicleNFTAddr && t.CustomClaims.ContractAddress != t.manufacturerAddr {
		return fmt.Errorf("%w: unexpected contract address passed", errUnauthorized)
	}
	return nil
}

func getTelemetryClaim(ctx context.Context) (*TelemetryClaim, error) {
	claim, ok := ctx.Value(TelemetryClaimContextKey{}).(*TelemetryClaim)
	if !ok || claim == nil {
		return nil, jwtmiddleware.ErrJWTMissing
	}
	return claim, nil
}
