package attestation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type indexRepoService interface {
	GetAllCloudEvents(ctx context.Context, filter *grpc.SearchOptions, limit int32) ([]cloudevent.CloudEvent[json.RawMessage], error)
}
type Repository struct {
	logger         *zerolog.Logger
	indexService   indexRepoService
	chainID        uint64
	vehicleAddress common.Address
}

// New creates a new instance of Service.
func New(indexService indexRepoService, chainID uint64, vehicleAddress common.Address, logger *zerolog.Logger) *Repository {
	return &Repository{
		indexService:   indexService,
		chainID:        chainID,
		vehicleAddress: vehicleAddress,
		logger:         logger,
	}
}

// GetAttestations fetches attestations for the given vehicle.
func (r *Repository) GetAttestations(ctx context.Context, vehicleTokenID uint32, filter *model.AttestationFilter) ([]*model.Attestation, error) {
	vehicleDID := cloudevent.NFTDID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         vehicleTokenID,
	}.String()
	opts := &grpc.SearchOptions{
		Type:    &wrapperspb.StringValue{Value: cloudevent.TypeAttestation},
		Subject: &wrapperspb.StringValue{Value: vehicleDID},
	}
	r.logger.Info().Msgf("fetching attestations: %s", vehicleDID)

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
	}

	cloudEvents, err := r.indexService.GetAllCloudEvents(ctx, opts, int32(limit))
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to get cloud events")
		return nil, errors.New("internal error")
	}

	tknID := int(vehicleTokenID)
	var attestations []*model.Attestation
	for _, ce := range cloudEvents {
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
			r.logger.Info().Str("id", attestation.ID).Str("source", attestation.Source.Hex()).Msg("failed to pull signature")
			return nil, fmt.Errorf("invalid signature from %s on attestation %s", attestation.ID, attestation.Source)
		}

		attestation.Signature = signature
		attestations = append(attestations, attestation)
	}

	return attestations, nil
}
