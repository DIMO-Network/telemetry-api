package attestation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

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
	GetLatestCloudEvent(ctx context.Context, filter *grpc.SearchOptions) (cloudevent.CloudEvent[json.RawMessage], error)
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
	logger := r.getLogger(ctx, vehicleTokenID)
	vehicleDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(vehicleTokenID)),
	}.String()
	opts := &grpc.SearchOptions{
		Type:    &wrapperspb.StringValue{Value: cloudevent.TypeAttestation},
		Subject: &wrapperspb.StringValue{Value: vehicleDID},
	}

	logger.Info().Msgf("fetching attestations: %s", vehicleDID)
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
		logger.Error().Err(err).Msg("failed to get cloud events")
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
			logger.Info().Str("id", attestation.ID).Str("source", attestation.Source.Hex()).Msg("failed to pull signature")
			return nil, fmt.Errorf("invalid format: attestation signature missing")
		}

		attestation.Signature = signature
		attestations = append(attestations, attestation)
	}

	return attestations, nil
}

// GetAttestations fetches attestations for the given vehicle.
func (r *Repository) GetAttestation(ctx context.Context, vehicleTokenID int, source common.Address, id string) (*model.Attestation, error) {
	logger := r.getLogger(ctx, vehicleTokenID)
	vehicleDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(vehicleTokenID)),
	}.String()
	opts := &grpc.SearchOptions{
		Type:    &wrapperspb.StringValue{Value: cloudevent.TypeAttestation},
		Subject: &wrapperspb.StringValue{Value: vehicleDID},
	}

	opts.Source = &wrapperspb.StringValue{Value: source.Hex()}
	opts.Id = &wrapperspb.StringValue{Value: id}

	cloudEvent, err := r.indexService.GetLatestCloudEvent(ctx, opts)
	if err != nil {
		logger.Error().Err(err).Msgf("failed to get cloudevent %s from source: %s", id, source)
		return nil, errors.New("internal error")
	}

	tknID := int(vehicleTokenID)
	att := &model.Attestation{
		ID:             cloudEvent.ID,
		VehicleTokenID: tknID,
		Time:           cloudEvent.Time,
		Attestation:    string(cloudEvent.Data),
		Type:           cloudEvent.Type,
		Source:         common.HexToAddress(cloudEvent.Source),
		DataVersion:    cloudEvent.DataVersion,
	}

	signature, ok := cloudEvent.Extras["signature"].(string)
	if !ok {
		logger.Info().Str("id", cloudEvent.ID).Str("source", cloudEvent.Source).Msg("failed to pull signature")
		return nil, fmt.Errorf("invalid format: attestation signature missing")
	}
	att.Signature = signature

	return att, nil
}

func (r *Repository) getLogger(ctx context.Context, vehicleTokenID int) zerolog.Logger {
	return zerolog.Ctx(ctx).With().Str("component", "attestation").Int("vehicleTokenId", vehicleTokenID).Logger()
}
