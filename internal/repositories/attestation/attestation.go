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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (r *Repository) GetAttestations(ctx context.Context, vehicleTokenID uint32, signer *string) ([]*model.Attestation, error) {
	vehicleDID := cloudevent.NFTDID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         vehicleTokenID,
	}.String()
	opts := &grpc.SearchOptions{
		Type:    &wrapperspb.StringValue{Value: cloudevent.TypeAttestation},
		Subject: &wrapperspb.StringValue{Value: vehicleDID},
	}

	if signer != nil {
		if !common.IsHexAddress(*signer) {
			r.logger.Info().Msgf("invalid attestation signer: %s", *signer)
			return nil, errors.New("invalid attestation signer")
		}
		opts.Source = &wrapperspb.StringValue{Value: *signer}
	}

	cloudEvents, err := r.indexService.GetAllCloudEvents(ctx, opts)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil //nolint // nil is a valid response
		}
		r.logger.Error().Err(err).Msg("failed to fetch vehicle attestations")
		return nil, errors.New("internal error")
	}

	tknID := int(vehicleTokenID)
	var attestations []*model.Attestation
	for _, ce := range cloudEvents {
		attestations = append(attestations, &model.Attestation{
			VehicleTokenID: &tknID,
			RecordedAt:     &ce.Time,
			Attestation:    string(ce.Data),
		})

	}

	return attestations, nil
}
