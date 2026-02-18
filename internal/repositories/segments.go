package repositories

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"golang.org/x/sync/errgroup"
)

const (
	maxDateRangeDays = 31
	maxSegmentLimit  = 200
)

// validateSegmentArgs validates the arguments for segment queries.
func validateSegmentArgs(tokenID int, from, to time.Time) error {
	if tokenID <= 0 {
		return fmt.Errorf("invalid tokenID: %d", tokenID)
	}

	if from.After(to) {
		return fmt.Errorf("from time must be before to time")
	}

	if from.Equal(to) {
		return fmt.Errorf("from and to times cannot be equal")
	}

	if to.After(time.Now()) {
		return fmt.Errorf("to time cannot be in the future")
	}

	maxDuration := maxDateRangeDays * 24 * time.Hour
	if to.Sub(from) > maxDuration {
		return fmt.Errorf("date range exceeds maximum of %d days", maxDateRangeDays)
	}

	return nil
}

// validateSegmentConfig validates the segment configuration parameters.
// When mechanism is idling, also validates idling-specific fields; refuel/recharge validate minIncreasePercent.
func validateSegmentConfig(config *model.SegmentConfig, mechanism model.DetectionMechanism) error {
	if config == nil {
		return nil
	}

	if config.MaxGapSeconds != nil {
		if *config.MaxGapSeconds < 60 || *config.MaxGapSeconds > 3600 {
			return fmt.Errorf("maxGapSeconds must be between 60 and 3600")
		}
	}

	if config.MinSegmentDurationSeconds != nil {
		if *config.MinSegmentDurationSeconds < 60 || *config.MinSegmentDurationSeconds > 3600 {
			return fmt.Errorf("minSegmentDurationSeconds must be between 60 and 3600")
		}
	}

	if config.SignalCountThreshold != nil {
		if *config.SignalCountThreshold < 1 || *config.SignalCountThreshold > 3600 {
			return fmt.Errorf("signalCountThreshold must be between 1 and 3600")
		}
	}

	if mechanism == model.DetectionMechanismIdling {
		if config.MaxIdleRpm != nil {
			if *config.MaxIdleRpm < 300 || *config.MaxIdleRpm > 3000 {
				return fmt.Errorf("maxIdleRpm must be between 300 and 3000")
			}
		}
	}

	if mechanism == model.DetectionMechanismRefuel || mechanism == model.DetectionMechanismRecharge {
		if config.MinIncreasePercent != nil {
			if *config.MinIncreasePercent < 1 || *config.MinIncreasePercent > 100 {
				return fmt.Errorf("minIncreasePercent must be between 1 and 100")
			}
		}
	}

	return nil
}

// validateSegmentLimit validates optional pagination limit. If non-nil, must be in [1, maxSegmentLimit].
func validateSegmentLimit(limit *int) error {
	if limit == nil {
		return nil
	}
	if *limit < 1 || *limit > maxSegmentLimit {
		return fmt.Errorf("limit must be between 1 and %d", maxSegmentLimit)
	}
	return nil
}

// Signal request building blocks used to compose mechanism-specific default sets.
var (
	sigSpeed     = &model.SegmentSignalRequest{Name: vss.FieldSpeed, Agg: model.FloatAggregationMax}
	sigFuelFirst = &model.SegmentSignalRequest{Name: vss.FieldPowertrainFuelSystemRelativeLevel, Agg: model.FloatAggregationFirst}
	sigFuelLast  = &model.SegmentSignalRequest{Name: vss.FieldPowertrainFuelSystemRelativeLevel, Agg: model.FloatAggregationLast}
	sigSoCFirst  = &model.SegmentSignalRequest{Name: vss.FieldPowertrainTractionBatteryStateOfChargeCurrent, Agg: model.FloatAggregationFirst}
	sigSoCLast   = &model.SegmentSignalRequest{Name: vss.FieldPowertrainTractionBatteryStateOfChargeCurrent, Agg: model.FloatAggregationLast}
	sigOdoFirst  = &model.SegmentSignalRequest{Name: vss.FieldPowertrainTransmissionTravelledDistance, Agg: model.FloatAggregationFirst}
	sigOdoLast   = &model.SegmentSignalRequest{Name: vss.FieldPowertrainTransmissionTravelledDistance, Agg: model.FloatAggregationLast}

	// mechanismSignalSets maps detection mechanism to its default signal set.
	// Mechanisms not listed fall through to baseSignalSet.
	mechanismSignalSets = map[model.DetectionMechanism][]*model.SegmentSignalRequest{
		model.DetectionMechanismIdling:   {sigSpeed, sigFuelFirst, sigFuelLast, sigOdoFirst, sigOdoLast},
		model.DetectionMechanismRefuel:   {sigSpeed, sigFuelFirst, sigFuelLast, sigOdoFirst, sigOdoLast},
		model.DetectionMechanismRecharge: {sigSpeed, sigSoCFirst, sigSoCLast, sigOdoFirst, sigOdoLast},
	}

	baseSignalSet = []*model.SegmentSignalRequest{sigSpeed, sigFuelFirst, sigFuelLast, sigSoCFirst, sigSoCLast, sigOdoFirst, sigOdoLast}
)

// defaultSegmentSignalSet returns the default signal set when summary is requested but signalRequests is omitted or empty.
func defaultSegmentSignalSet(mechanism model.DetectionMechanism) []*model.SegmentSignalRequest {
	if set, ok := mechanismSignalSets[mechanism]; ok {
		return set
	}
	return baseSignalSet
}

// buildAggArgs converts segment signal requests into the float and location arg slices needed by the CH service.
func buildAggArgs(signalReqs []*model.SegmentSignalRequest) ([]model.FloatSignalArgs, []model.LocationSignalArgs) {
	floatArgs := make([]model.FloatSignalArgs, 0, len(signalReqs))
	for _, req := range signalReqs {
		floatArgs = append(floatArgs, model.FloatSignalArgs{
			Name:  req.Name,
			Agg:   req.Agg,
			Alias: req.Name + "_" + string(req.Agg),
		})
	}
	locationArgs := []model.LocationSignalArgs{
		{Name: vss.FieldCurrentLocationCoordinates, Agg: model.LocationAggregationFirst, Alias: "startLoc"},
		{Name: vss.FieldCurrentLocationCoordinates, Agg: model.LocationAggregationLast, Alias: "endLoc"},
	}
	return floatArgs, locationArgs
}

// noDataLocation returns a zero-value location used as a fallback when no real location data is available.
// A new instance is returned each call to avoid shared-pointer aliasing.
func noDataLocation() *model.Location {
	return &model.Location{Latitude: 0, Longitude: 0, Hdop: 0}
}

// mergeSegmentSignalRequests returns defaultSet plus any client request not already present (deduped by name+agg). Order: defaultSet first, then new client requests.
func mergeSegmentSignalRequests(defaultSet []*model.SegmentSignalRequest, clientRequests []*model.SegmentSignalRequest) []*model.SegmentSignalRequest {
	key := func(r *model.SegmentSignalRequest) string {
		if r == nil {
			return ""
		}
		return r.Name + ":" + string(r.Agg)
	}
	seen := make(map[string]struct{})
	var out []*model.SegmentSignalRequest
	for _, r := range append(defaultSet, clientRequests...) {
		if r == nil {
			continue
		}
		k := key(r)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, r)
	}
	return out
}

// sortSegmentSignals orders signals by name, then by agg (e.g. FIRST, LAST, MAX) for stable, readable output.
func sortSegmentSignals(signals []*model.SignalAggregationValue) {
	sort.Slice(signals, func(i, j int) bool {
		if signals[i].Name != signals[j].Name {
			return signals[i].Name < signals[j].Name
		}
		return signals[i].Agg < signals[j].Agg
	})
}

// GetSegments returns segments detected using the specified mechanism in the time range.
// Pagination: pass after (exclusive cursor = startTime of last segment from previous page) and limit (default 100, max 200).
// Segments are ordered by startTime ascending. When after is set, only segments with startTime > after are requested from CH.
func (r *Repository) GetSegments(ctx context.Context, tokenID int, from, to time.Time, mechanism model.DetectionMechanism, config *model.SegmentConfig, signalRequests []*model.SegmentSignalRequest, eventRequests []*model.SegmentEventRequest, limit *int, after *time.Time) ([]*model.Segment, error) {
	if err := validateSegmentArgs(tokenID, from, to); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	if err := validateSegmentConfig(config, mechanism); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	if err := validateSegmentLimit(limit); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}
	// Cursor: only request segments with startTime > after so CH returns fewer rows
	if after != nil && after.Before(to) {
		cursorFrom := (*after).Add(time.Nanosecond) // exclusive: first segment start > after
		if cursorFrom.After(from) {
			from = cursorFrom
		}
	}

	chSegments, err := r.chService.GetSegments(ctx, uint32(tokenID), from, to, mechanism, config)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}
	// Apply limit before building ranges and batch queries so we don't run agg/event-count for segments we'll drop.
	if limit != nil && len(chSegments) > *limit {
		chSegments = chSegments[:*limit]
	}

	// Default signal set is always included; client signalRequests are added on top (deduped by name+agg).
	defaultReqs := defaultSegmentSignalSet(mechanism)
	signalReqs := mergeSegmentSignalRequests(defaultReqs, signalRequests)
	wantSummary := len(signalReqs) > 0 || len(eventRequests) > 0
	var eventNames []string
	if len(eventRequests) > 0 {
		eventNames = make([]string, len(eventRequests))
		for i, e := range eventRequests {
			eventNames[i] = e.Name
		}
	}

	subject := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         big.NewInt(int64(tokenID)),
	}.String()

	// For refuel/recharge, segment end is when the event is "done"; telemetry often has
	// the first fuel/charge/odometer reading shortly after. Use an extended summary
	// window for signal aggregation so those readings are included.
	const summaryEndBuffer = 2 * time.Minute
	extendSummaryEnd := mechanism == model.DetectionMechanismRefuel || mechanism == model.DetectionMechanismRecharge

	var eventCountsBySeg map[int]map[string]int
	var aggsBySeg map[int][]*ch.AggSignal
	if wantSummary && len(chSegments) > 0 {
		ranges := make([]ch.TimeRange, len(chSegments))
		aggRanges := make([]ch.TimeRange, len(chSegments))
		var globalFrom, globalTo time.Time
		for i, seg := range chSegments {
			segTo := to
			if seg.End != nil {
				segTo = seg.End.Timestamp
			}
			ranges[i] = ch.TimeRange{From: seg.Start.Timestamp, To: segTo}
			summaryTo := segTo
			if extendSummaryEnd {
				summaryTo = segTo.Add(summaryEndBuffer)
			}
			aggRanges[i] = ch.TimeRange{From: seg.Start.Timestamp, To: summaryTo}
			if i == 0 {
				globalFrom, globalTo = seg.Start.Timestamp, summaryTo
			} else {
				if seg.Start.Timestamp.Before(globalFrom) {
					globalFrom = seg.Start.Timestamp
				}
				if summaryTo.After(globalTo) {
					globalTo = summaryTo
				}
			}
		}
		floatArgs, locationArgs := buildAggArgs(signalReqs)
		var batchCounts []*ch.EventCountForRange
		var batchAggs []*ch.AggSignalForRange
		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			var err error
			batchCounts, err = r.chService.GetEventCountsForRanges(gctx, subject, ranges, eventNames)
			return err
		})
		g.Go(func() error {
			var err error
			batchAggs, err = r.chService.GetAggregatedSignalsForRanges(gctx, uint32(tokenID), aggRanges, globalFrom, globalTo, floatArgs, locationArgs)
			return err
		})
		if err := g.Wait(); err != nil {
			return nil, handleDBError(ctx, err)
		}
		eventCountsBySeg = make(map[int]map[string]int, len(chSegments))
		for _, ec := range batchCounts {
			if eventCountsBySeg[ec.SegIndex] == nil {
				eventCountsBySeg[ec.SegIndex] = make(map[string]int)
			}
			eventCountsBySeg[ec.SegIndex][ec.Name] = ec.Count
		}
		aggsBySeg = make(map[int][]*ch.AggSignal, len(chSegments))
		for _, a := range batchAggs {
			aggsBySeg[a.SegIndex] = append(aggsBySeg[a.SegIndex], &ch.AggSignal{
				SignalType:    a.SignalType,
				SignalIndex:   a.SignalIndex,
				ValueNumber:   a.ValueNumber,
				ValueString:   a.ValueString,
				ValueLocation: a.ValueLocation,
			})
		}
	}

	segments := chSegments
	for i, seg := range segments {
		if wantSummary {
			var eventCounts []*ch.EventCount
			if eventCountsBySeg != nil {
				m := eventCountsBySeg[i]
				eventCounts = make([]*ch.EventCount, 0, len(m))
				for name, count := range m {
					eventCounts = append(eventCounts, &ch.EventCount{Name: name, Count: count})
				}
			}
			var preFetchedAggs []*ch.AggSignal
			if aggsBySeg != nil {
				preFetchedAggs = aggsBySeg[i]
				if preFetchedAggs == nil {
					preFetchedAggs = []*ch.AggSignal{}
				}
			}
			summary, err := r.segmentSummary(ctx, uint32(tokenID), subject, seg, to, signalReqs, eventNames, eventCounts, preFetchedAggs)
			if err != nil {
				return nil, err
			}
			seg.Signals = summary.Signals
			seg.EventCounts = summary.EventCounts
			if summary.StartLocation != nil {
				seg.Start.Value = summary.StartLocation
			}
			if seg.End != nil && summary.EndLocation != nil {
				seg.End.Value = summary.EndLocation
			}
		}
		// Ensure non-nil location for GraphQL (schema: value: Location!)
		if seg.Start.Value == nil {
			seg.Start.Value = noDataLocation()
		}
		if seg.End != nil && seg.End.Value == nil {
			seg.End.Value = noDataLocation()
		}
	}

	// Idling: exclude segments where speed > 0 (RPM-based detection can include driving segments for EVs/hybrids or bad data).
	if mechanism == model.DetectionMechanismIdling {
		segments = filterIdlingSegmentsBySpeed(segments, 0)
	}
	return segments, nil
}

// segmentMaxSpeed returns the MAX speed value from segment signals, or -1 if not found.
func segmentMaxSpeed(signals []*model.SignalAggregationValue) float64 {
	for _, s := range signals {
		if s != nil && s.Name == vss.FieldSpeed && s.Agg == "MAX" {
			return s.Value
		}
	}
	return -1
}

// filterIdlingSegmentsBySpeed keeps only segments with speed <= maxSpeedKph or no speed signal.
// Returns a new slice; does not mutate the input.
func filterIdlingSegmentsBySpeed(segments []*model.Segment, maxSpeedKph float64) []*model.Segment {
	out := make([]*model.Segment, 0, len(segments))
	for _, seg := range segments {
		maxSpeed := segmentMaxSpeed(seg.Signals)
		if maxSpeed < 0 || maxSpeed <= maxSpeedKph {
			out = append(out, seg)
		}
	}
	return out
}

type segmentSummaryResult struct {
	Signals       []*model.SignalAggregationValue
	StartLocation *model.Location
	EndLocation   *model.Location
	EventCounts   []*model.EventCount
}

// buildSummaryFromAggs extracts signal values and start/end locations from aggregated signals.
func buildSummaryFromAggs(aggs []*ch.AggSignal, floatArgs []model.FloatSignalArgs) ([]*model.SignalAggregationValue, *model.Location, *model.Location) {
	signals := make([]*model.SignalAggregationValue, 0, len(floatArgs))
	var startLoc, endLoc *model.Location
	for _, a := range aggs {
		if a.SignalType == ch.FloatType && int(a.SignalIndex) < len(floatArgs) {
			signals = append(signals, &model.SignalAggregationValue{
				Name:  floatArgs[a.SignalIndex].Name,
				Agg:   string(floatArgs[a.SignalIndex].Agg),
				Value: a.ValueNumber,
			})
		}
		if a.SignalType == ch.LocType {
			loc := &model.Location{
				Latitude:  a.ValueLocation.Latitude,
				Longitude: a.ValueLocation.Longitude,
				Hdop:      a.ValueLocation.HDOP,
			}
			if a.SignalIndex == 0 {
				startLoc = loc
			} else {
				endLoc = loc
			}
		}
	}
	sortSegmentSignals(signals)
	return signals, startLoc, endLoc
}

// buildEventSummary builds a sorted event count slice from a count map and optional requested names.
// When eventNames is non-empty, only those names appear in the result (with count 0 for missing).
// When eventNames is empty, all entries from eventCountMap are returned.
func buildEventSummary(eventCountMap map[string]int, eventNames []string) []*model.EventCount {
	if len(eventNames) > 0 {
		out := make([]*model.EventCount, len(eventNames))
		for i, name := range eventNames {
			out[i] = &model.EventCount{Name: name, Count: eventCountMap[name]}
		}
		return out
	}
	out := make([]*model.EventCount, 0, len(eventCountMap))
	for name, count := range eventCountMap {
		out = append(out, &model.EventCount{Name: name, Count: count})
	}
	return out
}

// eventCountsToMap converts a slice of EventCount to a map keyed by name.
func eventCountsToMap(counts []*ch.EventCount) map[string]int {
	m := make(map[string]int, len(counts))
	for _, ec := range counts {
		m[ec.Name] = ec.Count
	}
	return m
}

func (r *Repository) segmentSummary(ctx context.Context, tokenID uint32, subject string, seg *model.Segment, queryTo time.Time, signalReqs []*model.SegmentSignalRequest, eventNames []string, preFetchedEventCounts []*ch.EventCount, preFetchedAggs []*ch.AggSignal) (*segmentSummaryResult, error) {
	segFrom := seg.Start.Timestamp
	segTo := queryTo
	if seg.End != nil {
		segTo = seg.End.Timestamp
	}
	intervalMicro := segTo.Sub(segFrom).Microseconds()
	if intervalMicro <= 0 {
		intervalMicro = 1
	}

	floatArgs, locationArgs := buildAggArgs(signalReqs)
	var aggs []*ch.AggSignal
	if preFetchedAggs != nil {
		aggs = preFetchedAggs
	} else {
		aggArgs := &model.AggregatedSignalArgs{
			SignalArgs:   model.SignalArgs{TokenID: tokenID},
			FromTS:       segFrom,
			ToTS:         segTo,
			Interval:     intervalMicro,
			FloatArgs:    floatArgs,
			LocationArgs: locationArgs,
		}
		var err error
		aggs, err = r.chService.GetAggregatedSignals(ctx, aggArgs)
		if err != nil {
			return nil, handleDBError(ctx, err)
		}
	}

	signalSummary, startLoc, endLoc := buildSummaryFromAggs(aggs, floatArgs)

	var eventCountMap map[string]int
	if preFetchedEventCounts != nil {
		eventCountMap = eventCountsToMap(preFetchedEventCounts)
	} else {
		eventCounts, err := r.chService.GetEventCounts(ctx, subject, segFrom, segTo, eventNames)
		if err != nil {
			return nil, handleDBError(ctx, err)
		}
		eventCountMap = eventCountsToMap(eventCounts)
	}

	return &segmentSummaryResult{
		Signals:       signalSummary,
		StartLocation: startLoc,
		EndLocation:   endLoc,
		EventCounts:   buildEventSummary(eventCountMap, eventNames),
	}, nil
}

// GetDailyActivity returns one record per calendar day in the requested date range, including days with zero segments.
// mechanism must be ignitionDetection, frequencyAnalysis, or changePointDetection; idling, refuel, recharge return 400.
func (r *Repository) GetDailyActivity(ctx context.Context, tokenID int, from, to time.Time, mechanism model.DetectionMechanism, config *model.SegmentConfig, signalRequests []*model.SegmentSignalRequest, eventRequests []*model.SegmentEventRequest, timezone *string) ([]*model.DailyActivity, error) {
	if mechanism == model.DetectionMechanismIdling || mechanism == model.DetectionMechanismRefuel || mechanism == model.DetectionMechanismRecharge {
		return nil, errorhandler.NewBadRequestError(ctx, fmt.Errorf("dailyActivity does not accept mechanism %s; use ignitionDetection, frequencyAnalysis, or changePointDetection", mechanism))
	}
	loc := time.UTC
	if timezone != nil && *timezone != "" {
		var err error
		loc, err = time.LoadLocation(*timezone)
		if err != nil {
			return nil, errorhandler.NewBadRequestError(ctx, fmt.Errorf("invalid timezone %q: %w", *timezone, err))
		}
	}
	fromInLoc := from.In(loc)
	toInLoc := to.In(loc)
	fromDate := time.Date(fromInLoc.Year(), fromInLoc.Month(), fromInLoc.Day(), 0, 0, 0, 0, loc)
	toDate := time.Date(toInLoc.Year(), toInLoc.Month(), toInLoc.Day(), 0, 0, 0, 0, loc)
	if !fromDate.Before(toDate) {
		return nil, errorhandler.NewBadRequestError(ctx, fmt.Errorf("from date must be before to date"))
	}
	if toDate.After(time.Now().In(loc)) {
		return nil, errorhandler.NewBadRequestError(ctx, fmt.Errorf("to date cannot be in the future"))
	}
	if toDate.Sub(fromDate) > maxDateRangeDays*24*time.Hour {
		return nil, errorhandler.NewBadRequestError(ctx, fmt.Errorf("date range exceeds maximum of %d days", maxDateRangeDays))
	}
	rangeStart := fromDate
	rangeEnd := toDate.Add(24 * time.Hour)

	defaultReqs := defaultSegmentSignalSet(mechanism)
	signalReqs := mergeSegmentSignalRequests(defaultReqs, signalRequests)
	var eventNames []string
	if len(eventRequests) > 0 {
		eventNames = make([]string, len(eventRequests))
		for i, e := range eventRequests {
			eventNames[i] = e.Name
		}
	}

	segments, err := r.GetSegments(ctx, tokenID, rangeStart, rangeEnd, mechanism, config, signalReqs, eventRequests, nil, nil)
	if err != nil {
		return nil, err
	}
	subject := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         big.NewInt(int64(tokenID)),
	}.String()

	var out []*model.DailyActivity
	for d := fromDate; d.Before(toDate); d = d.Add(24 * time.Hour) {
		dayStart := d
		dayEnd := d.Add(24 * time.Hour)
		dayStartUTC := dayStart.UTC()
		dayEndUTC := dayEnd.UTC()

		var segmentCount int
		var totalActiveSeconds int
		var firstSeg, lastSeg *model.Segment
		for _, seg := range segments {
			segEnd := dayEndUTC
			if seg.End != nil && seg.End.Timestamp.Before(dayEndUTC) {
				segEnd = seg.End.Timestamp
			}
			if seg.Start.Timestamp.After(dayEndUTC) || segEnd.Before(dayStartUTC) || !segEnd.After(seg.Start.Timestamp) {
				continue
			}
			segmentCount++
			overlapStart := seg.Start.Timestamp
			if overlapStart.Before(dayStartUTC) {
				overlapStart = dayStartUTC
			}
			overlapEnd := segEnd
			if overlapEnd.After(dayEndUTC) {
				overlapEnd = dayEndUTC
			}
			totalActiveSeconds += int(overlapEnd.Sub(overlapStart).Seconds())
			if firstSeg == nil {
				firstSeg = seg
			}
			lastSeg = seg
		}

		signalSummary, startLoc, endLoc, eventSummary, err := r.daySummary(ctx, uint32(tokenID), subject, dayStartUTC, dayEndUTC, signalReqs, eventNames)
		if err != nil {
			return nil, err
		}
		if firstSeg != nil && firstSeg.Start != nil && firstSeg.Start.Value != nil {
			startLoc = firstSeg.Start.Value
		}
		if lastSeg != nil && lastSeg.End != nil && lastSeg.End.Value != nil {
			endLoc = lastSeg.End.Value
		}
		if startLoc == nil {
			startLoc = noDataLocation()
		}
		if endLoc == nil {
			endLoc = noDataLocation()
		}

		startSignalLoc := &model.SignalLocation{Timestamp: dayStartUTC, Value: startLoc}
		endSignalLoc := &model.SignalLocation{Timestamp: dayEndUTC, Value: endLoc}
		out = append(out, &model.DailyActivity{
			SegmentCount: segmentCount,
			Duration:     totalActiveSeconds,
			Start:        startSignalLoc,
			End:          endSignalLoc,
			Signals:      signalSummary,
			EventCounts:  eventSummary,
		})
	}
	if out == nil {
		out = []*model.DailyActivity{}
	}
	return out, nil
}

func (r *Repository) daySummary(ctx context.Context, tokenID uint32, subject string, dayStart, dayEnd time.Time, signalReqs []*model.SegmentSignalRequest, eventNames []string) ([]*model.SignalAggregationValue, *model.Location, *model.Location, []*model.EventCount, error) {
	intervalMicro := dayEnd.Sub(dayStart).Microseconds()
	if intervalMicro <= 0 {
		intervalMicro = 1
	}
	floatArgs, locationArgs := buildAggArgs(signalReqs)
	aggArgs := &model.AggregatedSignalArgs{
		SignalArgs:   model.SignalArgs{TokenID: tokenID},
		FromTS:       dayStart,
		ToTS:         dayEnd,
		Interval:     intervalMicro,
		FloatArgs:    floatArgs,
		LocationArgs: locationArgs,
	}
	aggs, err := r.chService.GetAggregatedSignals(ctx, aggArgs)
	if err != nil {
		return nil, nil, nil, nil, handleDBError(ctx, err)
	}
	signalSummary, startLoc, endLoc := buildSummaryFromAggs(aggs, floatArgs)

	eventCounts, err := r.chService.GetEventCounts(ctx, subject, dayStart, dayEnd, eventNames)
	if err != nil {
		return nil, nil, nil, nil, handleDBError(ctx, err)
	}
	eventSummary := buildEventSummary(eventCountsToMap(eventCounts), eventNames)
	return signalSummary, startLoc, endLoc, eventSummary, nil
}
