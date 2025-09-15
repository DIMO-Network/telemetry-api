package attestation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type indexRepoService interface {
	GetAllCloudEvents(ctx context.Context, filter *grpc.AdvancedSearchOptions, limit int32) ([]cloudevent.CloudEvent[json.RawMessage], error)
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
	if !auth.ValidRequest(ctx, filter) {
		return nil, errorhandler.NewUnauthorizedError(ctx, errors.New("invalid claims"))
	}
	vehicleDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(vehicleTokenID)),
	}.String()
	opts := &grpc.AdvancedSearchOptions{
		Type: &grpc.StringFilterOption{
			In: []string{cloudevent.TypeAttestation},
		},
		Subject: &grpc.StringFilterOption{
			In: []string{vehicleDID},
		},
	}

	limit := 10
	if filter != nil {
		if filter.Source != nil {
			opts.Source = &grpc.StringFilterOption{
				In: []string{filter.Source.Hex()},
			}
		}

		if filter.Producer != nil {
			opts.Producer = &grpc.StringFilterOption{
				In: []string{*filter.Producer},
			}
		}

		if filter.After != nil {
			opts.After = timestamppb.New(*filter.After)
		}

		if filter.Before != nil {
			opts.Before = timestamppb.New(*filter.Before)
		}

		if filter.DataVersion != nil {
			opts.DataVersion = &grpc.StringFilterOption{
				In: []string{*filter.DataVersion},
			}
		}

		if filter.Limit != nil {
			limit = *filter.Limit
		}

		if filter.ID != nil {
			opts.Id = &grpc.StringFilterOption{
				In: []string{*filter.ID},
			}
		}

		if filter.Tags != nil {
			opts.Tags = toFetchAPIArrayFilterOption(filter.Tags)
		}
	}

	cloudEvents, err := r.indexService.GetAllCloudEvents(ctx, opts, int32(limit))
	if err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to get cloud events: %w", err), "internal error")
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

		attestation.Signature = ce.Signature
		attestations = append(attestations, attestation)
	}

	return attestations, nil
}

func toFetchAPIArrayFilterOption(filter *model.StringArrayFilter) *grpc.ArrayFilterOption {
	if filter == nil {
		return nil
	}
	orOptions := make([]*grpc.ArrayFilterOption, len(filter.Or))
	for i, or := range filter.Or {
		orOptions[i] = toFetchAPIArrayFilterOption(or)
	}
	return &grpc.ArrayFilterOption{
		ContainsAny:    filter.ContainsAny,
		ContainsAll:    filter.ContainsAll,
		NotContainsAny: filter.NotContainsAny,
		NotContainsAll: filter.NotContainsAll,
		Or:             orOptions,
	}
}
