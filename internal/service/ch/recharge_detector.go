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
	rechargeDefaultMinDurationSeconds = 60            // shorter default than other detectors — charge sessions can be brief
	rechargeSessionGapMax             = 2 * time.Hour // merge consecutive segments if gap ≤ this and odometer unchanged
	rechargeOdometerEpsilonKm         = 0.5           // allow odometer increase ≤ this (noise) to still count as stationary
	rechargeMinRisePct                = 1.0           // start-to-peak must rise at least this much to be a session (and always more than rechargeRunTolerancePct)
	rechargeRunTolerancePct           = 1.0           // a monotone run survives dips up to this far below its peak (OBD integer staircase, Tesla float jitter)
	rechargeMaxRatePctPerHour         = 600.0         // physical cap: a real session on a small pack at 350 kW stays under ~400 %/h; sensor glitches are far above
)

// RechargeDetector detects recharge segments by finding rises in the raw SoC curve while the vehicle is stationary.
type RechargeDetector struct {
	conn clickhouse.Conn
}

// NewRechargeDetector creates a new RechargeDetector with the given connection.
func NewRechargeDetector(conn clickhouse.Conn) *RechargeDetector {
	return &RechargeDetector{conn: conn}
}

// DetectSegments loads SoC and odometer samples and runs the pure recharge detection over them.
func (d *RechargeDetector) DetectSegments(
	ctx context.Context,
	subject string,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*model.Segment, error) {
	rc := resolveBaseConfig(config)
	// Use a shorter default minDuration for recharge; still honor explicit user override.
	if config == nil || config.MinSegmentDurationSeconds == nil {
		rc.minDuration = rechargeDefaultMinDurationSeconds
	}
	minRisePct := rechargeMinRisePct
	if config != nil && config.MinIncreasePercent != nil && *config.MinIncreasePercent > 0 {
		minRisePct = float64(*config.MinIncreasePercent)
	}

	socSamples, err := getLevelSamples(ctx, d.conn, subject, vss.FieldPowertrainTractionBatteryStateOfChargeCurrent, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query SoC samples: %w", err)
	}
	if len(socSamples) < 2 {
		return []*model.Segment{}, nil
	}
	odoSamples, err := getLevelSamples(ctx, d.conn, subject, vss.FieldPowertrainTransmissionTravelledDistance, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query odometer samples: %w", err)
	}

	sessions := detectRechargeSessions(socSamples, odoSamples, rc.minDuration, minRisePct)
	return rechargeSessionsToSegments(sessions, from), nil
}

// GetMechanismName returns the name of this detection mechanism.
func (d *RechargeDetector) GetMechanismName() string {
	return "recharge"
}

// rechargeSession is a detected charging session with its longest unobserved interval.
type rechargeSession struct {
	start, end          time.Time
	maxSampleGapSeconds int
}

// detectRechargeSessions is the pure detection core: it walks raw SoC samples for monotone runs, validates each
// run against the odometer (the rise must happen while the car is stationary) and a physical rate cap, then merges
// sessions that are close in time with an unchanged odometer.
//
// Devices that sleep while charging report nothing between the last parked reading and the wake-up, so a session
// may consist of just two samples; no smoothing is applied and sample count is never used as a proxy for time.
// soc and odo must be sorted by ts.
func detectRechargeSessions(soc, odo []levelSample, minDuration int, minRisePct float64) []rechargeSession {
	if len(soc) < 2 {
		return nil
	}
	var candidates []timeRange
	for _, run := range findMonotoneRuns(soc) {
		if tr, ok := validateRechargeRun(soc, odo, run, minDuration, minRisePct); ok {
			candidates = append(candidates, tr)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	shouldMerge := func(a, b timeRange) bool {
		_, odoCurEnd, ok1 := levelFirstLastInRange(odo, a.start, a.end)
		odoNextStart, _, ok2 := levelFirstLastInRange(odo, b.start, b.end)
		return ok1 && ok2 && odoCurEnd == odoNextStart
	}
	// Zero from/to: no clipping, the samples were already loaded for [from, to).
	merged := mergeTimeRanges(candidates, rechargeSessionGapMax, minDuration, time.Time{}, time.Time{}, shouldMerge)
	out := make([]rechargeSession, 0, len(merged))
	for _, tr := range merged {
		out = append(out, rechargeSession{start: tr.start, end: tr.end, maxSampleGapSeconds: maxSampleGapSeconds(soc, tr)})
	}
	return out
}

// monotoneRun is a candidate rise over raw samples: indices into the SoC slice.
type monotoneRun struct {
	troughIdx int // lowest sample before the rise
	peakIdx   int // first sample reaching the run's maximum
}

// findMonotoneRuns splits the SoC series into rises. A run starts at a trough and continues while SoC does not
// fall more than rechargeRunTolerancePct below the run's peak. The peak is the first sample reaching the maximum,
// so the run never bleeds into the departure drive that follows a sleeping charge.
func findMonotoneRuns(soc []levelSample) []monotoneRun {
	var runs []monotoneRun
	troughIdx, peakIdx := 0, 0
	for i := 1; i < len(soc); i++ {
		v := soc[i].value
		peakVal := soc[peakIdx].value
		switch {
		case v > peakVal:
			peakIdx = i
		case v < peakVal-rechargeRunTolerancePct:
			// The run is over; a new one starts at this lower sample.
			if peakIdx > troughIdx {
				runs = append(runs, monotoneRun{troughIdx: troughIdx, peakIdx: peakIdx})
			}
			troughIdx, peakIdx = i, i
		case v < soc[troughIdx].value:
			// Still drifting down within tolerance and below the trough: the rise has not started yet.
			troughIdx, peakIdx = i, i
		}
	}
	if peakIdx > troughIdx {
		runs = append(runs, monotoneRun{troughIdx: troughIdx, peakIdx: peakIdx})
	}
	return runs
}

// validateRechargeRun anchors a run on its stationary core and applies the rise, duration and rate checks.
// The session start reading is the last SoC sample at or before the car last moved (bounded below by the trough),
// so a reading taken while still driving to the charger counts and energy is not under-reported; equal readings
// after that point are skipped so the session starts when SoC last sat at its start value.
func validateRechargeRun(soc, odo []levelSample, run monotoneRun, minDuration int, minRisePct float64) (timeRange, bool) {
	peak := soc[run.peakIdx]
	startIdx := run.troughIdx
	if stationaryStart, ok := stationaryStartBefore(odo, peak.ts); ok {
		// Last SoC sample at or before the car stopped moving.
		idx := sort.Search(len(soc), func(i int) bool { return soc[i].ts.After(stationaryStart) }) - 1
		if idx > startIdx {
			startIdx = idx
		}
	}
	// Skip the flat lead-in: while the next sample has not risen above the start reading the charge has not
	// begun. This also absorbs sub-epsilon drives that an integer odometer cannot show.
	for startIdx+1 < run.peakIdx && soc[startIdx+1].value <= soc[startIdx].value {
		startIdx++
	}
	start := soc[startIdx]
	rise := peak.value - start.value
	// A rise within the run tolerance is indistinguishable from quantization flicker (46,46,47,46 on integer OBD).
	if rise < minRisePct || rise <= rechargeRunTolerancePct {
		return timeRange{}, false
	}
	dur := peak.ts.Sub(start.ts)
	if int(dur.Seconds()) < minDuration {
		return timeRange{}, false
	}
	if rise/dur.Hours() > rechargeMaxRatePctPerHour {
		return timeRange{}, false
	}
	return timeRange{start: start.ts, end: peak.ts}, true
}

// stationaryStartBefore walks odometer samples backward from t and returns the timestamp of the earliest sample
// after which the odometer stayed within rechargeOdometerEpsilonKm of its value at t, i.e. when the car last moved.
// ok is false when there is no odometer sample at or before t (no odometer data: caller keeps the trough).
func stationaryStartBefore(odo []levelSample, t time.Time) (time.Time, bool) {
	last := sort.Search(len(odo), func(i int) bool { return odo[i].ts.After(t) }) - 1
	if last < 0 {
		return time.Time{}, false
	}
	odoAtPeak := odo[last].value
	k := last
	for k > 0 && odoAtPeak-odo[k-1].value <= rechargeOdometerEpsilonKm {
		k--
	}
	return odo[k].ts, true
}

// maxSampleGapSeconds returns the longest interval within tr with no SoC sample, including the lead-in from
// tr.start to the first sample and the tail from the last sample to tr.end.
func maxSampleGapSeconds(soc []levelSample, tr timeRange) int {
	first := sort.Search(len(soc), func(i int) bool { return !soc[i].ts.Before(tr.start) })
	prev := tr.start
	maxGap := time.Duration(0)
	for i := first; i < len(soc) && !soc[i].ts.After(tr.end); i++ {
		if g := soc[i].ts.Sub(prev); g > maxGap {
			maxGap = g
		}
		prev = soc[i].ts
	}
	if g := tr.end.Sub(prev); g > maxGap {
		maxGap = g
	}
	return int(maxGap.Seconds())
}

// rechargeSessionsToSegments converts sessions to model segments, setting MaxSampleGapSeconds.
// Returns an empty (non-nil) slice when there are no sessions.
func rechargeSessionsToSegments(sessions []rechargeSession, from time.Time) []*model.Segment {
	out := make([]*model.Segment, 0, len(sessions))
	for _, s := range sessions {
		end := s.end
		seg := newSegment(s.start, &end, int32(end.Sub(s.start).Seconds()), false, !s.start.After(from))
		gap := s.maxSampleGapSeconds
		seg.MaxSampleGapSeconds = &gap
		out = append(out, seg)
	}
	return out
}
