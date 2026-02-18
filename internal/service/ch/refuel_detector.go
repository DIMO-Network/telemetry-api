package ch

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	refuelWindowMinutes      = 5    // window length for fuel rise detection
	refuelMinRisePercent     = 30   // fuel must rise more than this % (relative) in the window
	refuelMinFuelEpsilon     = 1e-6
	refuelMinAbsoluteRisePct = 20.0 // trough-to-peak must rise at least this much in absolute % to be a real refuel
	refuelPeakSearchMaxMin      = 30  // max minutes to search forward from rise window for the peak
	refuelPeakStabilizationDrop = 1.0 // if fuel drops more than this from current sample to next, consider current sample the peak
)

// RefuelDetector detects refuel segments by finding large fuel rises and emitting segments from
// the last low reading (trough) to the first stable high reading (peak).
type RefuelDetector struct {
	conn clickhouse.Conn
}

// NewRefuelDetector creates a new RefuelDetector with the given connection.
func NewRefuelDetector(conn clickhouse.Conn) *RefuelDetector {
	return &RefuelDetector{conn: conn}
}

// DetectSegments finds 5-min windows with >30% fuel rise, then for each rise emits a segment
// from the trough (last low sample before the jump) to the peak (first stable high after).
// 1 CH query (fuel only).
func (d *RefuelDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*model.Segment, error) {
	rc := resolveBaseConfig(config)
	minDuration := rc.minDuration
	minRise := refuelMinRisePercent
	if config != nil && config.MinIncreasePercent != nil && *config.MinIncreasePercent > 0 {
		minRise = *config.MinIncreasePercent
	}
	windowDur := time.Duration(refuelWindowMinutes) * time.Minute
	fuelFrom := from.Add(-windowDur)
	fuelTo := to.Add(windowDur)

	// Single CH query: fuel samples (returned sorted by CH)
	samples, err := getLevelSamples(ctx, d.conn, tokenID, vss.FieldPowertrainFuelSystemRelativeLevel, fuelFrom, fuelTo)
	if err != nil {
		return nil, fmt.Errorf("failed to query fuel samples: %w", err)
	}
	if len(samples) < 2 {
		return []*model.Segment{}, nil
	}

	// Scan 5-min windows for large rises; track sample index incrementally
	var raw []timeRange
	t := from.Truncate(time.Minute)
	if t.Before(from) {
		t = t.Add(time.Minute)
	}
	for !t.Add(windowDur).After(to) {
		windowEnd := t.Add(windowDur)
		fuelStart := sampleAtOrBefore(samples, t)
		fuelEnd := sampleAtOrBefore(samples, windowEnd)
		if fuelStart >= refuelMinFuelEpsilon {
			risePct := (fuelEnd - fuelStart) / fuelStart * 100
			if risePct > float64(minRise) {
				troughTime, peakTime, absRise := findRefuelTroughAndPeak(samples, t, windowEnd)
				if !troughTime.IsZero() && !peakTime.IsZero() && peakTime.After(troughTime) && absRise >= refuelMinAbsoluteRisePct {
					if troughTime.Before(from) {
						troughTime = from
					}
					if peakTime.After(to) {
						peakTime = to
					}
					if int(peakTime.Sub(troughTime).Seconds()) >= minDuration {
						raw = append(raw, timeRange{start: troughTime, end: peakTime})
					}
				}
			}
		}
		t = t.Add(time.Minute)
	}

	merged := mergeTimeRanges(raw, 0, minDuration, from, to, nil)
	return timeRangesToSegments(merged, from), nil
}

// GetMechanismName returns the name of this detection mechanism.
func (d *RefuelDetector) GetMechanismName() string {
	return "refuel"
}

// findRefuelTroughAndPeak finds the trough (last low sample at or before riseStart) and
// peak (first sample where fuel stabilizes high after riseEnd) around a detected fuel rise.
// Uses binary search to jump to the relevant indices. samples must be sorted by ts.
func findRefuelTroughAndPeak(samples []levelSample, riseStart, riseEnd time.Time) (trough, peak time.Time, absRise float64) {
	peakDeadline := riseEnd.Add(time.Duration(refuelPeakSearchMaxMin) * time.Minute)

	// Find trough: binary search to first index at or before riseStart, then walk backward for local min.
	startIdx := sort.Search(len(samples), func(i int) bool { return samples[i].ts.After(riseStart) })
	if startIdx > 0 {
		startIdx--
	}
	troughIdx := -1
	troughVal := 0.0
	for i := startIdx; i >= 0; i-- {
		if troughIdx == -1 || samples[i].value <= troughVal {
			troughIdx = i
			troughVal = samples[i].value
		} else {
			break
		}
	}

	// Find peak: binary search to first index at or after riseEnd, then walk forward capped at deadline.
	peakStart := sort.Search(len(samples), func(i int) bool { return !samples[i].ts.Before(riseEnd) })
	peakIdx := -1
	peakVal := 0.0
	for i := peakStart; i < len(samples); i++ {
		if samples[i].ts.After(peakDeadline) {
			break
		}
		if peakIdx == -1 || samples[i].value >= peakVal {
			peakIdx = i
			peakVal = samples[i].value
		}
		if i+1 < len(samples) && !samples[i+1].ts.After(peakDeadline) && samples[i+1].value < samples[i].value-refuelPeakStabilizationDrop {
			peakIdx = i
			break
		}
	}

	if troughIdx < 0 || peakIdx < 0 {
		return time.Time{}, time.Time{}, 0
	}
	rise := samples[peakIdx].value - samples[troughIdx].value
	return samples[troughIdx].ts, samples[peakIdx].ts, rise
}
