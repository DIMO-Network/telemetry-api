package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/DIMO-Network/model-garage/pkg/schema"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/rs/zerolog"
)

var (
	errInternal = errors.New("internal error")
	errTimeout  = errors.New("request exceeded or is estimated to exceed the maximum execution time")
	unixEpoch   = time.Unix(0, 0).UTC()
)

// TODO(elffjs): Get rid of this when we have device addresses in CH.
var ManufacturerSourceTranslations = map[string]string{
	"AutoPi":  "autopi",
	"Hashdog": "macaron",
	"Ruptela": "ruptela",
}

// CHService is the interface for the ClickHouse service.
//
//go:generate mockgen -source=./repositories.go -destination=repositories_mocks_test.go -package=repositories_test
type CHService interface {
	GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.AggSignal, error)
	GetLatestSignals(ctx context.Context, latestArgs *model.LatestSignalsArgs) ([]*vss.Signal, error)
	GetAvailableSignals(ctx context.Context, tokenID uint32, filter *model.SignalFilter) ([]string, error)
}

// Repository is the base repository for all repositories.
type Repository struct {
	queryableSignals map[string]struct{}
	chService        CHService
	log              *zerolog.Logger
	lastSeenBin      time.Duration
}

// NewRepository creates a new base repository.
// clientCAs is optional and can be nil.
func NewRepository(logger *zerolog.Logger, chService CHService, lastSeenBin int64) (*Repository, error) {
	definitions, err := schema.LoadDefinitionFile(strings.NewReader(schema.DefinitionsYAML()))
	if err != nil {
		return nil, fmt.Errorf("error reading definition file: %w", err)
	}
	queryableSignals := make(map[string]struct{}, len(definitions.FromName))
	for vssName := range definitions.FromName {
		queryableSignals[schema.VSSToJSONName(vssName)] = struct{}{}
	}

	return &Repository{
		chService:        chService,
		log:              logger,
		queryableSignals: queryableSignals,
		lastSeenBin:      time.Duration(lastSeenBin) * time.Hour,
	}, nil

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
		// ClickHouse returns the Unix epoch for max(timestamp) if there are no rows.
		if signal.Name == model.LastSeenField && !signal.Timestamp.Equal(unixEpoch) {
			coll.LastSeen = &signal.Timestamp
			continue
		}
		model.SetCollectionField(coll, signal)
	}

	return coll, nil
}

// GetDeviceActivity returns device status activity level.
func (r *Repository) GetDeviceActivity(ctx context.Context, vehicleTokenID int, adMfrName string) (*model.DeviceActivity, error) {
	source, ok := ManufacturerSourceTranslations[adMfrName]
	if !ok {
		return nil, fmt.Errorf("unrecognized manufacturer name %s", adMfrName)
	}

	args := &model.LatestSignalsArgs{
		IncludeLastSeen: true,
		SignalArgs: model.SignalArgs{
			TokenID: uint32(vehicleTokenID),
			Filter: &model.SignalFilter{
				Source: &source,
			},
		},
	}

	latest, err := r.GetSignalLatest(ctx, args)
	if err != nil {
		return nil, err
	}

	var out model.DeviceActivity

	if latest.LastSeen != nil {
		binned := latest.LastSeen.Truncate(r.lastSeenBin)
		out.LastActive = &binned
	}

	return &out, nil
}

// GetAvailableSignals returns the available signals for the given tokenID and filter.
// If no signals are found, a nil slice is returned.
func (r *Repository) GetAvailableSignals(ctx context.Context, tokenID uint32, filter *model.SignalFilter) ([]string, error) {
	allSignals, err := r.chService.GetAvailableSignals(ctx, tokenID, filter)
	if err != nil {
		return nil, handleDBError(err, r.log)
	}
	var retSignals []string
	for _, signal := range allSignals {
		if _, ok := r.queryableSignals[signal]; ok {
			retSignals = append(retSignals, signal)
		}
	}
	return retSignals, nil
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
