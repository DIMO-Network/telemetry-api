package ch

import (
	"context"
	"fmt"
	"math"
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
		candidates = validateRechargeRun(soc, odo, run, minDuration, minRisePct, candidates)
	}
	if len(candidates) == 0 {
		return nil
	}
	// Two sessions merge only if the car sat at the same odometer for both. Compare the odometer at each
	// peak (the stationary value) rather than at the boundary samples, which may be shared between sessions.
	shouldMerge := func(a, b timeRange) bool {
		odoA, okA := odometerAtOrBefore(odo, a.end)
		odoB, okB := odometerAtOrBefore(odo, b.end)
		return okA && okB && math.Abs(odoA-odoB) <= rechargeOdometerEpsilonKm
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

// validateRechargeRun anchors a run on its stationary core, applies the rise, duration and rate checks, and
// appends the resulting session (if any) to out.
//
// The session start reading is the lowest SoC sample between the last moving and the first stationary odometer
// sample (bounded below by the trough), so a reading taken while still rolling to the charger counts and a sparse
// odometer does not lag the start; equal readings after that point are skipped so the session starts when SoC
// last sat at its start value. If the car moved after the trough, the part of the run before that movement is
// evaluated on its own so a charge followed by a short hop to a second charger is not folded into the second.
func validateRechargeRun(soc, odo []levelSample, run monotoneRun, minDuration int, minRisePct float64, out []timeRange) []timeRange {
	peak := soc[run.peakIdx]
	// The anchored start can only be at or above the trough, so the raw rise bounds the real one.
	if rawRise := peak.value - soc[run.troughIdx].value; rawRise < minRisePct || rawRise <= rechargeRunTolerancePct {
		return out
	}
	startIdx := run.troughIdx
	core, ok := stationaryCoreBefore(odo, peak.ts)
	if ok {
		idx := lowestSampleInWindow(soc, core.lastMoving, core.start)
		if idx < 0 {
			// No SoC sample while the car came to rest: last sample at or before it stopped.
			idx = sort.Search(len(soc), func(i int) bool { return soc[i].ts.After(core.start) }) - 1
		}
		if idx > startIdx {
			startIdx = idx
		}
		// The car moved after the trough: the earlier part of the run may be its own session.
		if core.hasLastMoving && core.lastMoving.After(soc[run.troughIdx].ts) {
			endIdx := min(startIdx, run.peakIdx-1)
			out = validateRechargeRun(soc, odo, monotoneRun{troughIdx: run.troughIdx, peakIdx: firstMaxIdx(soc, run.troughIdx, endIdx)}, minDuration, minRisePct, out)
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
		return out
	}
	dur := peak.ts.Sub(start.ts)
	if int(dur.Seconds()) < minDuration {
		return out
	}
	if rise/dur.Hours() > rechargeMaxRatePctPerHour {
		return out
	}
	return append(out, timeRange{start: start.ts, end: peak.ts})
}

// stationaryCore describes the stretch before a peak during which the odometer did not move.
type stationaryCore struct {
	start         time.Time // first odometer sample within epsilon of the value at the peak
	lastMoving    time.Time // the odometer sample before start (the car was still moving at this time)
	hasLastMoving bool      // false when the odometer series begins inside the stationary stretch
}

// stationaryCoreBefore walks odometer samples backward from t while they stay within rechargeOdometerEpsilonKm of
// the value at t. ok is false when there is no odometer sample at or before t (no odometer data: caller keeps the trough).
func stationaryCoreBefore(odo []levelSample, t time.Time) (stationaryCore, bool) {
	last := sort.Search(len(odo), func(i int) bool { return odo[i].ts.After(t) }) - 1
	if last < 0 {
		return stationaryCore{}, false
	}
	odoAtPeak := odo[last].value
	k := last
	for k > 0 && math.Abs(odoAtPeak-odo[k-1].value) <= rechargeOdometerEpsilonKm {
		k--
	}
	core := stationaryCore{start: odo[k].ts}
	if k > 0 {
		core.lastMoving, core.hasLastMoving = odo[k-1].ts, true
	}
	return core, true
}

// lowestSampleInWindow returns the index of the first lowest sample with ts in (after, until], or -1 if none.
func lowestSampleInWindow(soc []levelSample, after, until time.Time) int {
	best := -1
	for i := sort.Search(len(soc), func(i int) bool { return soc[i].ts.After(after) }); i < len(soc) && !soc[i].ts.After(until); i++ {
		if best < 0 || soc[i].value < soc[best].value {
			best = i
		}
	}
	return best
}

// firstMaxIdx returns the index of the first sample holding the maximum value in soc[from..to].
func firstMaxIdx(soc []levelSample, from, to int) int {
	best := from
	for i := from + 1; i <= to; i++ {
		if soc[i].value > soc[best].value {
			best = i
		}
	}
	return best
}

// odometerAtOrBefore returns the odometer value at or before t. ok is false if there is none.
func odometerAtOrBefore(odo []levelSample, t time.Time) (float64, bool) {
	idx := sort.Search(len(odo), func(i int) bool { return odo[i].ts.After(t) }) - 1
	if idx < 0 {
		return 0, false
	}
	return odo[idx].value, true
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
