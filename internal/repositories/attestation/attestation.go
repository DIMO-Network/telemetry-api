package attestation

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type indexRepoService interface {
	GetAllCloudEvents(ctx context.Context, filter *grpc.SearchOptions) ([]cloudevent.CloudEvent[json.RawMessage], error)
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
func (r *Repository) GetAttestations(ctx context.Context, vehicleTokenID uint32, source *common.Address, filter *model.AttestationFilter) ([]*model.Attestation, error) {
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
	if source != nil {
		opts.Source = &wrapperspb.StringValue{Value: source.Hex()}
	}

	if filter != nil {
		if filter.Producer != nil {
			opts.Producer = &wrapperspb.StringValue{Value: filter.Producer.Hex()}
		}

		if filter.EffectiveAt != nil {
			opts.After = timestamppb.New(*filter.EffectiveAt)
		}

		if filter.ExpiresAt != nil {
			opts.Before = timestamppb.New(*filter.ExpiresAt)
		}

		if filter.DataVersion != nil {
			opts.DataVersion = &wrapperspb.StringValue{Value: *filter.DataVersion}
		}
	}

	cloudEvents, err := r.indexService.GetAllCloudEvents(ctx, opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to get cloud events")
		return nil, errors.New("internal error")
	}

	tknID := int(vehicleTokenID)
	var attestations []*model.Attestation
	for _, ce := range cloudEvents {
		attestations = append(attestations, &model.Attestation{
			VehicleTokenID: tknID,
			RecordedAt:     ce.Time,
			Attestation:    string(ce.Data),
		})

	}

	return attestations, nil
}
