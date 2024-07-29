package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/rs/zerolog"
)

var (
	errInternal = errors.New("internal error")
	errTimeout  = errors.New("request exceeded or is estimated to exceed the maximum execution time")
)

// CHService is the interface for the ClickHouse service.
//
//go:generate mockgen -source=./repositories.go -destination=repositories_mocks_test.go -package=repositories_test
type CHService interface {
	GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.AggSignal, error)
	GetLatestSignals(ctx context.Context, latestArgs *model.LatestSignalsArgs) ([]*vss.Signal, error)
	GetDeviceActivity(ctx context.Context, vehicleTokenID int, adManuf string) ([]*model.DeviceActivity, error)
}

// Repository is the base repository for all repositories.
type Repository struct {
	chService CHService
	log       *zerolog.Logger
}

// NewRepository creates a new base repository.
// clientCAs is optional and can be nil.
func NewRepository(logger *zerolog.Logger, chService CHService) *Repository {
	return &Repository{
		chService: chService,
		log:       logger,
	}
}

// GetSignal returns the aggregated signals for the given tokenID, interval, from, to and filter.
func (r *Repository) GetSignal(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.SignalAggregations, error) {
	if err := validateAggSigArgs(aggArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	signals, err := r.chService.GetAggregatedSignals(ctx, aggArgs)
	if err != nil {
		return nil, handleDBError(err, r.log)
	}

	// combine signals with the same timestamp by iterating over all signals
	// if the timestamp differs from the previous signal, create a new SignalAggregations object
	var allAggs []*model.SignalAggregations
	var currAggs *model.SignalAggregations
	lastTS := time.Time{}

	for _, signal := range signals {
		if !lastTS.Equal(signal.Timestamp) {
			lastTS = signal.Timestamp
			currAggs = &model.SignalAggregations{
				Timestamp:    signal.Timestamp,
				ValueNumbers: make(map[model.AliasKey]float64),
				ValueStrings: make(map[model.AliasKey]string),
			}
			allAggs = append(allAggs, currAggs)
		}

		model.SetAggregationField(currAggs, signal)
	}

	return allAggs, nil
}

// GetSignalLatest returns the latest signals for the given tokenID and filter.
func (r *Repository) GetSignalLatest(ctx context.Context, latestArgs *model.LatestSignalsArgs) (*model.SignalCollection, error) {
	if err := validateLatestSigArgs(latestArgs); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	signals, err := r.chService.GetLatestSignals(ctx, latestArgs)
	if err != nil {
		return nil, handleDBError(err, r.log)
	}
	coll := &model.SignalCollection{}
	for _, signal := range signals {
		if signal.Name == model.LastSeenField {
			coll.LastSeen = &signal.Timestamp
			continue
		}
		model.SetCollectionField(coll, signal)
	}

	return coll, nil
}

// GetDeviceActivity returns device status activity level.
func (r *Repository) GetDeviceActivity(ctx context.Context, vehicleTokenID int, adManuf string) (*model.DeviceActivity, error) {
	resp, err := r.chService.GetDeviceActivity(ctx, vehicleTokenID, adManuf)
	if err != nil {
		return nil, handleDBError(err, r.log)
	}

	if len(resp) == 0 {
		return nil, handleDBError(errors.New("no device activity found"), r.log)
	}

	return resp[len(resp)-1], nil
}

// handleDBError logs the error and returns a generic error message.
func handleDBError(err error, log *zerolog.Logger) error {
	exceptionErr := &proto.Exception{}
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &exceptionErr) && exceptionErr.Code == ch.TimeoutErrCode) {
		log.Error().Err(err).Msg("failed to query db")
		return errTimeout
	}
	log.Error().Err(err).Msg("failed to query db")
	return errInternal
}
