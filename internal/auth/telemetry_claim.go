package auth

import (
	"context"
	"slices"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/shared/pkg/privileges"
	"github.com/DIMO-Network/shared/pkg/set"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/ethereum/go-ethereum/common"
)

// TelemetryClaimContextKey is a custom key for the context to store the custom claims.
type TelemetryClaimContextKey struct{}

// TelemetryClaim is a custom claim for the telemetry API.
type TelemetryClaim struct {
	privileges set.Set[model.Privilege]
	tokenclaims.CustomClaims
}

// Validate function is required to implement the validator.CustomClaims interface.
func (t *TelemetryClaim) Validate(context.Context) error {
	return nil
}

// SetPrivileges populates the set of GraphQL privileges on the claim object. To do this,
// it combines the address and privilege ids on the token together with the given map.
func (t *TelemetryClaim) SetPrivileges(contractPrivMaps map[common.Address]map[privileges.Privilege]model.Privilege) {
	t.privileges = set.New[model.Privilege]()

	contractClaims, ok := contractPrivMaps[t.ContractAddress]
	if !ok {
		return
	}

	for _, privID := range t.PrivilegeIDs {
		modelPriv, ok := contractClaims[privID]
		if !ok {
			continue
		}
		t.privileges.Add(modelPriv)
	}
}

func getTelemetryClaim(ctx context.Context) (*TelemetryClaim, error) {
	claim, ok := ctx.Value(TelemetryClaimContextKey{}).(*TelemetryClaim)
	if !ok || claim == nil {
		return nil, jwtmiddleware.ErrJWTMissing
	}
	return claim, nil
}

func ValidAttestationClaim(ctx context.Context, filter *model.AttestationFilter) bool {
	claim, err := getTelemetryClaim(ctx)
	if err != nil || claim.CloudEvents == nil {
		return false
	}

	source := tokenclaims.GlobalIdentifier
	id := tokenclaims.GlobalIdentifier

	if filter != nil {
		if filter.Source != nil {
			source = filter.Source.Hex()
		}

		if filter.ID != nil {
			id = *filter.ID
		}
	}

	var validClaim bool
	for _, ce := range claim.CloudEvents.Events {
		if ce.EventType == cloudevent.TypeAttestation || ce.EventType == tokenclaims.GlobalIdentifier {
			if ce.Source == source || ce.Source == tokenclaims.GlobalIdentifier {
				if slices.Contains(ce.IDs, id) || slices.Contains(ce.IDs, tokenclaims.GlobalIdentifier) {
					validClaim = true
				}
			}
		}
	}

	return validClaim
}
