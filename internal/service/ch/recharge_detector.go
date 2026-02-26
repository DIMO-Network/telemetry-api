package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	rechargeDefaultMinDurationSeconds = 60            // shorter default than other detectors — charge sessions can be brief
	rechargeSessionGapMax             = 2 * time.Hour // merge consecutive segments if gap ≤ this and odometer unchanged
	rechargeOdometerEpsilonKm         = 0.5           // allow odometer increase ≤ this (noise) to still count as stationary
	rechargeSmoothWindow              = 11            // rolling average window for SoC smoothing (~11 samples ≈ 11 min)
	rechargeMinRisePct                = 1.0           // trough-to-peak must rise at least this much to be a candidate
)

// RechargeDetector detects recharge segments by finding trough-to-peak rises in the SoC curve.
type RechargeDetector struct {
	conn clickhouse.Conn
}

// NewRechargeDetector creates a new RechargeDetector with the given connection.
func NewRechargeDetector(conn clickhouse.Conn) *RechargeDetector {
	return &RechargeDetector{conn: conn}
}

// DetectSegments finds periods where state of charge rises (trough to peak), filters by odometer, and merges nearby sessions.
func (d *RechargeDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
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
	return detectRechargeSegments(ctx, d.conn, tokenID, from, to, rc.minDuration, minRisePct)
}

// GetMechanismName returns the name of this detection mechanism.
func (d *RechargeDetector) GetMechanismName() string {
	return "recharge"
}

// detectRechargeSegments: 2 CH queries (SoC + odometer), then all processing in-memory.
func detectRechargeSegments(ctx context.Context, conn clickhouse.Conn, tokenID uint32, from, to time.Time, minDuration int, minRisePct float64) ([]*model.Segment, error) {
	// Query 1: SoC samples (returned sorted by CH)
	socSamples, err := getLevelSamples(ctx, conn, tokenID, vss.FieldPowertrainTractionBatteryStateOfChargeCurrent, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query SoC samples: %w", err)
	}
	if len(socSamples) < rechargeSmoothWindow+2 {
		return []*model.Segment{}, nil
	}

	// Query 2: Odometer samples (returned sorted by CH)
	odoSamples, err := getLevelSamples(ctx, conn, tokenID, vss.FieldPowertrainTransmissionTravelledDistance, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query odometer samples: %w", err)
	}

	// Step 1: Smooth SoC to eliminate per-sample noise
	smoothed := smoothSamples(socSamples, rechargeSmoothWindow)

	// Step 2: Find trough-to-peak ranges from smoothed curve
	candidates := findTroughToPeakRanges(smoothed, minRisePct, minDuration)

	// Step 3: Filter by SoC increase and odometer non-increase
	filtered := filterRangesBySocAndOdo(candidates, socSamples, odoSamples)

	// Step 4: Merge consecutive sessions (with odometer check)
	shouldMerge := func(a, b timeRange) bool {
		_, odoCurEnd, ok1 := levelFirstLastInRange(odoSamples, a.start, a.end)
		odoNextStart, _, ok2 := levelFirstLastInRange(odoSamples, b.start, b.end)
		return ok1 && ok2 && odoCurEnd == odoNextStart
	}
	// Merge with zero from/to to skip clipping (already filtered/clipped upstream)
	merged := mergeTimeRanges(filtered, rechargeSessionGapMax, minDuration, time.Time{}, time.Time{}, shouldMerge)

	return timeRangesToSegments(merged, from), nil
}

// smoothSamples applies a rolling average over the given window size.
// Timestamps are taken from the center sample of each window.
// Uses per-position summation for exact floating-point reproducibility.
func smoothSamples(samples []levelSample, window int) []levelSample {
	if window <= 1 || len(samples) <= window {
		return samples
	}
	half := window / 2
	wf := float64(window)
	out := make([]levelSample, 0, len(samples)-window+1)
	for i := half; i < len(samples)-half; i++ {
		sum := 0.0
		for j := i - half; j <= i+half; j++ {
			sum += samples[j].value
		}
		out = append(out, levelSample{ts: samples[i].ts, value: sum / wf})
	}
	return out
}

// findTroughToPeakRanges walks smoothed SoC samples and finds every rise from a local trough to a local peak.
func findTroughToPeakRanges(samples []levelSample, minRisePct float64, minDuration int) []timeRange {
	if len(samples) < 2 {
		return nil
	}

	const (
		dirRising  = 1
		dirFalling = -1
	)

	var ranges []timeRange
	dir := 0
	troughIdx := 0
	peakIdx := 0

	for i := 1; i < len(samples); i++ {
		diff := samples[i].value - samples[i-1].value
		if diff > 0 {
			if dir == dirFalling {
				troughIdx = i - 1
			}
			dir = dirRising
			peakIdx = i
		} else if diff < 0 {
			if dir == dirRising {
				appendTroughToPeak(samples, troughIdx, peakIdx, minRisePct, minDuration, &ranges)
			}
			dir = dirFalling
		}
	}
	if dir == dirRising {
		appendTroughToPeak(samples, troughIdx, peakIdx, minRisePct, minDuration, &ranges)
	}
	return ranges
}

// appendTroughToPeak appends a timeRange if the rise meets minimum criteria.
func appendTroughToPeak(samples []levelSample, troughIdx, peakIdx int, minRisePct float64, minDuration int, out *[]timeRange) {
	rise := samples[peakIdx].value - samples[troughIdx].value
	if rise < minRisePct {
		return
	}
	start := samples[troughIdx].ts
	end := samples[peakIdx].ts
	if int(end.Sub(start).Seconds()) < minDuration {
		return
	}
	*out = append(*out, timeRange{start: start, end: end})
}

// filterRangesBySocAndOdo keeps only ranges where SoC increased and odometer did not increase beyond epsilon.
func filterRangesBySocAndOdo(ranges []timeRange, socSamples, odoSamples []levelSample) []timeRange {
	out := make([]timeRange, 0, len(ranges))
	for _, tr := range ranges {
		socFirst, socLast, socOk := levelFirstLastInRange(socSamples, tr.start, tr.end)
		if !socOk || socLast <= socFirst {
			continue
		}
		odoFirst, odoLast, odoOk := levelFirstLastInRange(odoSamples, tr.start, tr.end)
		if odoOk && (odoLast-odoFirst) > rechargeOdometerEpsilonKm {
			continue
		}
		out = append(out, tr)
	}
	return out
}
