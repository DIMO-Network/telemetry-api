package repositories

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/schema"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/ethereum/go-ethereum/common"
	"github.com/uber/h3-go/v4"
)

const approximateLocationResolution = 6

var unixEpoch = time.Unix(0, 0).UTC()

// CHService is the interface for the ClickHouse service.
type CHService interface {
	GetAggregatedSignals(ctx context.Context, subject string, aggArgs *model.AggregatedSignalArgs) ([]*ch.AggSignal, error)
	GetAggregatedSignalsForRanges(ctx context.Context, subject string, ranges []ch.TimeRange, globalFrom, globalTo time.Time, floatArgs []model.FloatSignalArgs, locationArgs []model.LocationSignalArgs) ([]*ch.AggSignalForRange, error)
	GetLatestSignals(ctx context.Context, subject string, latestArgs *model.LatestSignalsArgs) ([]*vss.Signal, error)
	GetAllLatestSignals(ctx context.Context, subject string, filter *model.SignalFilter) ([]*vss.Signal, error)
	GetAvailableSignals(ctx context.Context, subject string, filter *model.SignalFilter) ([]string, error)
	GetSignalSummaries(ctx context.Context, subject string, filter *model.SignalFilter) ([]*model.SignalDataSummary, error)
	GetEvents(ctx context.Context, subject string, from, to time.Time, filter *model.EventFilter) ([]*vss.Event, error)
	GetEventCounts(ctx context.Context, subject string, from, to time.Time, eventNames []string) ([]*ch.EventCount, error)
	GetEventCountsForRanges(ctx context.Context, subject string, ranges []ch.TimeRange, eventNames []string) ([]*ch.EventCountForRange, error)
	GetEventSummaries(ctx context.Context, subject string) ([]*ch.EventSummary, error)
	GetSegments(ctx context.Context, subject string, from, to time.Time, mechanism model.DetectionMechanism, config *model.SegmentConfig) ([]*model.Segment, error)
}

// Repository is the base repository for all repositories.
type Repository struct {
	queryableSignals map[string]struct{}
	chService        CHService
	chainID          uint64
	vehicleAddress   common.Address
}

// NewRepository creates a new base repository.
// clientCAs is optional and can be nil.
func NewRepository(chService CHService, settings config.Settings) (*Repository, error) {
	definitions, err := schema.LoadDefinitionFile(strings.NewReader(schema.DefaultDefinitionsYAML()))
	if err != nil {
		return nil, fmt.Errorf("error reading definition file: %w", err)
	}
	queryableSignals := make(map[string]struct{}, len(definitions.FromName))
	for vssName := range definitions.FromName {
		queryableSignals[schema.VSSToJSONName(vssName)] = struct{}{}
	}

	return &Repository{
		chService:        chService,
		queryableSignals: queryableSignals,
		chainID:          settings.ChainID,
		vehicleAddress:   settings.VehicleNFTAddress,
	}, nil

}

// toSubject converts a vehicle token id to a DID subject string.
func (r *Repository) toSubject(tokenID uint32) string {
	return cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         big.NewInt(int64(tokenID)),
	}.String()
}

// GetSignal returns the aggregated signals for the given tokenID, interval, from, to and filter.
func (r *Repository) GetSignal(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.SignalAggregations, error) {
	if err := validateAggSigArgs(aggArgs); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}

	subject := r.toSubject(aggArgs.TokenID)
	signals, err := r.chService.GetAggregatedSignals(ctx, subject, aggArgs)
	if err != nil {
		return nil, handleDBError(ctx, err)
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
				Timestamp:      signal.Timestamp,
				ValueNumbers:   make(map[string]float64),
				ValueStrings:   make(map[string]string),
				ValueLocations: make(map[string]vss.Location),
			}
			allAggs = append(allAggs, currAggs)
		}

		switch signal.SignalType {
		case ch.FloatType:
			if len(aggArgs.FloatArgs) <= int(signal.SignalIndex) {
				return nil, fmt.Errorf("only %d float signal requests, but the query returned index %d", len(aggArgs.FloatArgs), signal.SignalIndex)
			}
			currAggs.ValueNumbers[aggArgs.FloatArgs[signal.SignalIndex].Alias] = signal.ValueNumber
		case ch.StringType:
			if len(aggArgs.StringArgs) <= int(signal.SignalIndex) {
				return nil, fmt.Errorf("only %d string signal requests, but the query returned index %d", len(aggArgs.StringArgs), signal.SignalIndex)
			}
			currAggs.ValueStrings[aggArgs.StringArgs[signal.SignalIndex].Alias] = signal.ValueString
		case ch.LocType:
			if len(aggArgs.LocationArgs) <= int(signal.SignalIndex) {
				return nil, fmt.Errorf("only %d location signal requests, but the query returned index %d", len(aggArgs.LocationArgs), signal.SignalIndex)
			}
			currAggs.ValueLocations[aggArgs.LocationArgs[signal.SignalIndex].Alias] = signal.ValueLocation
		default:
			return nil, fmt.Errorf("scanned a row with unrecognized type number %d", signal.SignalType)
		}
	}

	return allAggs, nil
}

// GetSignalLatest returns the latest signals for the given tokenID and filter.
func (r *Repository) GetSignalLatest(ctx context.Context, latestArgs *model.LatestSignalsArgs) (*model.SignalCollection, error) {
	if err := validateLatestSigArgs(latestArgs); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	subject := r.toSubject(latestArgs.TokenID)
	signals, err := r.chService.GetLatestSignals(ctx, subject, latestArgs)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	coll := &model.SignalCollection{}
	for _, signal := range signals {
		// ClickHouse returns the Unix epoch for max(timestamp) if there are no rows.
		if signal.Data.Name == model.LastSeenField && !signal.Data.Timestamp.Equal(unixEpoch) {
			coll.LastSeen = &signal.Data.Timestamp
			continue
		}
		model.SetCollectionField(coll, signal)
	}
	setApproximateLocationInCollection(coll)
	return coll, nil
}

// GetAvailableSignals returns the available signals for the given tokenID and filter.
// If no signals are found, a nil slice is returned.
func (r *Repository) GetAvailableSignals(ctx context.Context, tokenID uint32, filter *model.SignalFilter) ([]string, error) {
	subject := r.toSubject(tokenID)
	allSignals, err := r.chService.GetAvailableSignals(ctx, subject, filter)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	var retSignals []string
	for _, signal := range allSignals {
		if _, ok := r.queryableSignals[signal]; ok {
			retSignals = append(retSignals, signal)
		}
	}
	return retSignals, nil
}

// GetSignalSnapshot returns the latest value for every available signal for the given tokenID.
func (r *Repository) GetSignalSnapshot(ctx context.Context, tokenID uint32, filter *model.SignalFilter) (*model.SignalsSnapshotResponse, error) {
	if tokenID < 1 {
		return nil, errorhandler.NewBadRequestError(ctx, ValidationError("tokenID is not a positive integer"))
	}
	if err := validateFilter(filter); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	subject := r.toSubject(tokenID)
	signals, err := r.chService.GetAllLatestSignals(ctx, subject, filter)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}

	resp := &model.SignalsSnapshotResponse{}
	var rawLocationSignal *vss.Signal
	for _, signal := range signals {
		if signal.Data.Name == model.LastSeenField && !signal.Data.Timestamp.Equal(unixEpoch) {
			resp.LastSeen = &signal.Data.Timestamp
			continue
		}
		// Only include queryable signals.
		if _, ok := r.queryableSignals[signal.Data.Name]; !ok {
			continue
		}
		ls := model.SignalToLatestSignal(signal)
		if ls == nil {
			continue
		}
		resp.Signals = append(resp.Signals, ls)
		if signal.Data.Name == vss.FieldCurrentLocationCoordinates {
			rawLocationSignal = signal
		}
	}

	// Emit approximate location entry derived from raw coordinates.
	if rawLocationSignal != nil {
		loc := rawLocationSignal.Data.ValueLocation
		approx := GetApproximateLoc(loc.Latitude, loc.Longitude)
		if approx != nil {
			resp.Signals = append(resp.Signals, &model.LatestSignal{
				Name:      model.ApproximateCoordinatesField,
				Timestamp: rawLocationSignal.Data.Timestamp,
				ValueLocation: &model.Location{
					Latitude:  approx.Lat,
					Longitude: approx.Lng,
					Hdop:      loc.HDOP,
				},
			})
		}
	}

	return resp, nil
}

// GetDataSummary returns the signal and event metadata for the given tokenID and filter.
func (r *Repository) GetDataSummary(ctx context.Context, tokenID uint32, filter *model.SignalFilter) (*model.DataSummary, error) {
	subject := r.toSubject(tokenID)
	signalDataSummary, err := r.chService.GetSignalSummaries(ctx, subject, filter)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	eventSummaries, err := r.chService.GetEventSummaries(ctx, subject)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	eventDataSummary := make([]*model.EventDataSummary, len(eventSummaries))
	for i, es := range eventSummaries {
		eventDataSummary[i] = &model.EventDataSummary{
			Name:           es.Name,
			NumberOfEvents: es.Count,
			FirstSeen:      es.FirstSeen,
			LastSeen:       es.LastSeen,
		}
	}
	totalCount := uint64(0)
	minTimestamp := time.Now().UTC()
	maxTimestamp := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	availableSignals := make([]string, len(signalDataSummary))
	for i, metadata := range signalDataSummary {
		availableSignals[i] = metadata.Name
		totalCount += metadata.NumberOfSignals
		if metadata.FirstSeen.Before(minTimestamp) {
			minTimestamp = metadata.FirstSeen
		}
		if metadata.LastSeen.After(maxTimestamp) {
			maxTimestamp = metadata.LastSeen
		}
	}
	for _, es := range eventSummaries {
		if es.FirstSeen.Before(minTimestamp) {
			minTimestamp = es.FirstSeen
		}
		if es.LastSeen.After(maxTimestamp) {
			maxTimestamp = es.LastSeen
		}
	}
	return &model.DataSummary{
		NumberOfSignals:   totalCount,
		FirstSeen:         minTimestamp,
		LastSeen:          maxTimestamp,
		AvailableSignals:  availableSignals,
		SignalDataSummary: signalDataSummary,
		EventDataSummary:  eventDataSummary,
	}, nil
}

// GetEvents returns the events for the given tokenID, from, to and filter.
func (r *Repository) GetEvents(ctx context.Context, tokenID int, from, to time.Time, filter *model.EventFilter) ([]*model.Event, error) {
	if err := validateEventArgs(tokenID, from, to, filter); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	subject := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         big.NewInt(int64(tokenID)),
	}.String()
	allEvents, err := r.chService.GetEvents(ctx, subject, from, to, filter)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	retEvents := make([]*model.Event, len(allEvents))
	for i, event := range allEvents {
		retEvents[i] = &model.Event{
			Timestamp:  event.Data.Timestamp,
			Name:       event.Data.Name,
			Source:     event.Source,
			DurationNs: int(event.Data.DurationNs),
		}
		if event.Data.Metadata != "" {
			retEvents[i].Metadata = &event.Data.Metadata
		}
	}
	return retEvents, nil
}

// handleDBError logs the error and returns a generic error message.
func handleDBError(ctx context.Context, err error) error {
	exceptionErr := &proto.Exception{}
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &exceptionErr) && exceptionErr.Code == ch.TimeoutErrCode) {
		return errorhandler.NewBadRequestErrorWithMsg(ctx, err, "request exceeded or is estimated to exceed the maximum execution time")
	}
	return errorhandler.NewInternalErrorWithMsg(ctx, err, "failed to query db")
}

// GetApproximateLoc returns the approximate location for the given latitude and longitude.
// The result is nil if some H3 library error occurs. We should probably be more concerned about this.
func GetApproximateLoc(lat, long float64) *h3.LatLng {
	h3LatLng := h3.NewLatLng(lat, long)
	cell, err := h3.LatLngToCell(h3LatLng, approximateLocationResolution)
	if err != nil {
		return nil
	}
	latLong, err := h3.CellToLatLng(cell)
	if err != nil {
		return nil
	}
	return &latLong
}

func setApproximateLocationInCollection(coll *model.SignalCollection) {
	if coll == nil || coll.CurrentLocationCoordinates == nil {
		return
	}
	loc := coll.CurrentLocationCoordinates
	latLong := GetApproximateLoc(loc.Value.Latitude, loc.Value.Longitude)
	if latLong == nil {
		return
	}
	coll.CurrentLocationApproximateCoordinates = &model.SignalLocation{
		Timestamp: loc.Timestamp,
		Value: &model.Location{
			Latitude:  latLong.Lat,
			Longitude: latLong.Lng,
			Hdop:      loc.Value.Hdop,
		},
	}
}
