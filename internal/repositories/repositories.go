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

// TODO(elffjs): Get rid of this when we have device addresses in CH.
var ManufacturerSourceTranslations = map[string]string{
	"AutoPi":  "autopi",
	"Hashdog": "macaron",
	"Ruptela": "ruptela",
}

// CHService is the interface for the ClickHouse service.
type CHService interface {
	GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*ch.AggSignal, error)
	GetLatestSignals(ctx context.Context, latestArgs *model.LatestSignalsArgs) ([]*vss.Signal, error)
	GetAvailableSignals(ctx context.Context, tokenID uint32, filter *model.SignalFilter) ([]string, error)
	GetSignalMetadata(ctx context.Context, tokenID uint32, filter *model.SignalFilter) ([]*model.SignalMetadata, error)
	GetEvents(ctx context.Context, subject string, from, to time.Time, filter *model.EventFilter) ([]*vss.Event, error)
}

// Repository is the base repository for all repositories.
type Repository struct {
	queryableSignals map[string]struct{}
	chService        CHService
	lastSeenBin      time.Duration
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
		lastSeenBin:      time.Duration(settings.DeviceLastSeenBinHrs) * time.Hour,
		chainID:          settings.ChainID,
		vehicleAddress:   settings.VehicleNFTAddress,
	}, nil

}

// GetSignal returns the aggregated signals for the given tokenID, interval, from, to and filter.
func (r *Repository) GetSignal(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.SignalAggregations, error) {
	if err := validateAggSigArgs(aggArgs); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}

	signals, err := r.chService.GetAggregatedSignals(ctx, aggArgs)
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
				AppLocNumbers:  make(map[model.AppLocKey]float64),
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
				return nil, fmt.Errorf("only %d string signal requests, but the query returned index %d", len(aggArgs.FloatArgs), signal.SignalIndex)
			}
			currAggs.ValueStrings[aggArgs.StringArgs[signal.SignalIndex].Alias] = signal.ValueString
		case ch.AppLocType:
			aggIndex, aggParity := signal.SignalIndex/2, signal.SignalIndex%2
			if int(aggIndex) >= len(model.AllFloatAggregation) {
				return nil, fmt.Errorf("scanned an approximate location row with aggregation index %d, but there are only %d types", signal.SignalIndex, len(model.AllFloatAggregation))
			}
			name := parityToLocationSignalName[aggParity]
			agg := model.AllFloatAggregation[aggIndex]
			currAggs.AppLocNumbers[model.AppLocKey{Aggregation: agg, Name: name}] = signal.ValueNumber
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

var parityToLocationSignalName = [2]string{vss.FieldCurrentLocationLatitude, vss.FieldCurrentLocationLongitude}

// GetSignalLatest returns the latest signals for the given tokenID and filter.
func (r *Repository) GetSignalLatest(ctx context.Context, latestArgs *model.LatestSignalsArgs) (*model.SignalCollection, error) {
	if err := validateLatestSigArgs(latestArgs); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	signals, err := r.chService.GetLatestSignals(ctx, latestArgs)
	if err != nil {
		return nil, handleDBError(ctx, err)
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
	setApproximateLocationInCollection(coll)
	return coll, nil
}

// GetDeviceActivity returns device status activity level.
func (r *Repository) GetDeviceActivity(ctx context.Context, vehicleTokenID int, adMfrName string) (*model.DeviceActivity, error) {
	source, ok := ManufacturerSourceTranslations[adMfrName]
	if !ok {
		return nil, errorhandler.NewBadRequestError(ctx, fmt.Errorf("unrecognized manufacturer name %s", adMfrName))
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

// GetSignalMetadata returns the signal metadata for the given tokenID and filter.
func (r *Repository) GetSignalMetadata(ctx context.Context, tokenID uint32, filter *model.SignalFilter) (*model.SignalsMetadata, error) {
	signalMetadata, err := r.chService.GetSignalMetadata(ctx, tokenID, filter)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	totalCount := 0
	minTimestamp := time.Now().UTC()
	maxTimestamp := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	availableSignals := make([]string, len(signalMetadata))
	for i, metadata := range signalMetadata {
		availableSignals[i] = metadata.Name
		totalCount += metadata.NumberOfSignals
		if metadata.FirstSeen.Before(minTimestamp) {
			minTimestamp = metadata.FirstSeen
		}
		if metadata.LastSeen.After(maxTimestamp) {
			maxTimestamp = metadata.LastSeen
		}
	}
	return &model.SignalsMetadata{
		NumberOfSignals:  totalCount,
		FirstSeen:        minTimestamp,
		LastSeen:         maxTimestamp,
		AvailableSignals: availableSignals,
		SignalMetadata:   signalMetadata,
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
			Timestamp:  event.Timestamp,
			Name:       event.Name,
			Source:     event.Source,
			DurationNs: int(event.DurationNs),
		}
		if event.Metadata != "" {
			retEvents[i].Metadata = &event.Metadata
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
	if coll == nil || coll.CurrentLocationLatitude == nil || coll.CurrentLocationLongitude == nil {
		return
	}
	latLong := GetApproximateLoc(coll.CurrentLocationLatitude.Value, coll.CurrentLocationLongitude.Value)
	coll.CurrentLocationApproximateLatitude = &model.SignalFloat{
		Timestamp: coll.CurrentLocationLatitude.Timestamp,
		Value:     latLong.Lat,
	}
	coll.CurrentLocationApproximateLongitude = &model.SignalFloat{
		Timestamp: coll.CurrentLocationLongitude.Timestamp,
		Value:     latLong.Lng,
	}
}
