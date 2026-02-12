package auth

import (
	"context"
	"fmt"
	"slices"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
)

// TelemetryClaimContextKey is a custom key for the context to store the custom claims.
type TelemetryClaimContextKey struct{}

// TelemetryClaim is a custom claim for the telemetry API.
type TelemetryClaim struct {
	AssetDID cloudevent.ERC721DID
	tokenclaims.CustomClaims
}

// Validate function is required to implement the validator.CustomClaims interface.
func (t *TelemetryClaim) Validate(context.Context) error {
	var nftErr error
	t.AssetDID, nftErr = cloudevent.DecodeERC721DID(t.Asset)
	if nftErr == nil {
		return nil
	}
	ethrDID, ethrErr := cloudevent.DecodeEthrDID(t.Asset)
	if ethrErr == nil {
		t.AssetDID = cloudevent.ERC721DID{
			ChainID:         ethrDID.ChainID,
			ContractAddress: ethrDID.ContractAddress,
			TokenID:         nil,
		}
		return nil
	}
	return fmt.Errorf("unauthorized: failed to decode Asset as ERC721 DID: %w or Ethr DID: %v", nftErr, ethrErr)
}

func getTelemetryClaim(ctx context.Context) (*TelemetryClaim, error) {
	claim, ok := ctx.Value(TelemetryClaimContextKey{}).(*TelemetryClaim)
	if !ok || claim == nil {
		return nil, jwtmiddleware.ErrJWTMissing
	}
	return claim, nil
}

func ValidRequest(ctx context.Context, subject string, filter *model.AttestationFilter) bool {
	claim, err := getTelemetryClaim(ctx)
	if err != nil || claim.CloudEvents == nil {
		return false
	}
	if subject != claim.Asset {
		return false
	}

	return validCloudEventRequest(claim, cloudevent.TypeAttestation, filter)
}

func validCloudEventRequest(claim *TelemetryClaim, cloudEvtType string, filter *model.AttestationFilter) bool {
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

	for _, ce := range claim.CloudEvents.Events {
		if ce.EventType == cloudEvtType || ce.EventType == tokenclaims.GlobalIdentifier {
			if ce.Source == source || ce.Source == tokenclaims.GlobalIdentifier {
				if slices.Contains(ce.IDs, id) || slices.Contains(ce.IDs, tokenclaims.GlobalIdentifier) {
					return true
				}
			}
		}
	}

	return false
}
