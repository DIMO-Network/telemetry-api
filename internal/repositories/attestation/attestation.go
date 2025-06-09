package attestation

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/set"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/pkg/errorhandler"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type indexRepoService interface {
	GetAllCloudEvents(ctx context.Context, filter *grpc.SearchOptions, limit int32) ([]cloudevent.CloudEvent[json.RawMessage], error)
}
type Repository struct {
	indexService   indexRepoService
	chainID        uint64
	vehicleAddress common.Address
}

// New creates a new instance of Service.
func New(indexService indexRepoService, chainID uint64, vehicleAddress common.Address) *Repository {
	return &Repository{
		indexService:   indexService,
		chainID:        chainID,
		vehicleAddress: vehicleAddress,
	}
}

// GetAttestations fetches attestations for the given vehicle.
func (r *Repository) GetAttestations(ctx context.Context, vehicleTokenID int, filter *model.AttestationFilter) ([]*model.Attestation, error) {
	claimMap, err := auth.GetAttestationClaimMap(ctx)
	if err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("no claims found in jwt for vehicle token: %d", vehicleTokenID), "internal error")
	}
	vehicleDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(vehicleTokenID)),
	}.String()
	opts := &grpc.SearchOptions{
		Type:    &wrapperspb.StringValue{Value: cloudevent.TypeAttestation},
		Subject: &wrapperspb.StringValue{Value: vehicleDID},
	}

	limit := 10
	if filter != nil {
		if filter.Source != nil {
			opts.Source = &wrapperspb.StringValue{Value: filter.Source.Hex()}
		}

		if filter.Producer != nil {
			opts.Producer = &wrapperspb.StringValue{Value: *filter.Producer}
		}

		if filter.After != nil {
			opts.After = timestamppb.New(*filter.After)
		}

		if filter.Before != nil {
			opts.Before = timestamppb.New(*filter.Before)
		}

		if filter.DataVersion != nil {
			opts.DataVersion = &wrapperspb.StringValue{Value: *filter.DataVersion}
		}

		if filter.Limit != nil {
			limit = *filter.Limit
		}

		if filter.ID != nil {
			opts.Id = &wrapperspb.StringValue{Value: *filter.ID}
		}
	}

	cloudEvents, err := r.indexService.GetAllCloudEvents(ctx, opts, int32(limit))
	if err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to get cloud events: %w", err), "internal error")
	}

	tknID := int(vehicleTokenID)
	var attestations []*model.Attestation
	for _, ce := range cloudEvents {
		if !validClaim(claimMap, ce.Source, ce.ID) {
			return nil, fmt.Errorf("no claim found for requested attestation: %s %s", ce.Source, ce.ID)
		}

		attestation := &model.Attestation{
			ID:             ce.ID,
			VehicleTokenID: tknID,
			Time:           ce.Time,
			Attestation:    string(ce.Data),
			Type:           ce.Type,
			Source:         common.HexToAddress(ce.Source),
			DataVersion:    ce.DataVersion,
		}

		if ce.Producer != "" {
			attestation.Producer = &ce.Producer
		}

		signature, ok := ce.Extras["signature"].(string)
		if !ok {
			return nil, errorhandler.NewBadRequestErrorWithMsg(ctx, fmt.Errorf("invalid signature from %s on attestation %s", attestation.ID, attestation.Source), "invalid signature")
		}

		attestation.Signature = signature
		attestations = append(attestations, attestation)
	}

	return attestations, nil
}
func validClaim(claims map[string]*set.StringSet, source, id string) bool {
	accessBySource, ok := claims[source]
	if !ok {
		globalAccess, ok := claims[tokenclaims.GlobalIdentifier]
		if !ok {
			return false
		}

		return globalAccess.Contains(id) || globalAccess.Contains(tokenclaims.GlobalIdentifier)
	}

	return accessBySource.Contains(id) || accessBySource.Contains(tokenclaims.GlobalIdentifier)
}
