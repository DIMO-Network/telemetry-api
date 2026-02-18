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
	defaultMaxIdleRpm = 1000 // max RPM to count as idle
	minRunningRpm     = 0    // RPM must be strictly above this to count as engine running (excludes key-on-engine-off)
)

// IdlingDetector detects segments where engine RPM remains in idle range.
// Processes RPM samples in-memory for exact segment boundaries (no window discretization).
// Note: Detection is RPM-only. Callers (e.g. repository) filter out segments with speed > 0.
type IdlingDetector struct {
	conn clickhouse.Conn
}

// NewIdlingDetector creates a new IdlingDetector with the given connection.
func NewIdlingDetector(conn clickhouse.Conn) *IdlingDetector {
	return &IdlingDetector{conn: conn}
}

// DetectSegments fetches RPM samples (1 CH query) and finds contiguous runs of idle RPM in-memory.
func (d *IdlingDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*model.Segment, error) {
	rc := resolveBaseConfig(config)
	maxGap := rc.maxGapSeconds
	minDuration := rc.minDuration
	maxIdleRpm := defaultMaxIdleRpm
	if config != nil && config.MaxIdleRpm != nil {
		maxIdleRpm = *config.MaxIdleRpm
	}

	lookbackFrom := from.Add(-time.Duration(maxGap) * time.Second)
	// Single CH query: RPM samples (returned sorted by CH)
	samples, err := getLevelSamples(ctx, d.conn, tokenID, vss.FieldPowertrainCombustionEngineSpeed, lookbackFrom, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query RPM samples: %w", err)
	}
	if len(samples) == 0 {
		return []*model.Segment{}, nil
	}

	ranges := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
	return timeRangesToSegments(ranges, from), nil
}

// GetMechanismName returns the name of this detection mechanism.
func (d *IdlingDetector) GetMechanismName() string {
	return "idling"
}

// findIdleRpmRanges walks sorted RPM samples and finds contiguous runs where 0 < RPM <= maxIdleRpm.
// A gap between consecutive idle samples larger than maxGap seconds ends the current run.
// Only runs with duration >= minDuration are emitted. Ranges are clipped to [from, to].
func findIdleRpmRanges(samples []levelSample, maxIdleRpm, maxGap, minDuration int, from, to time.Time) []timeRange {
	maxGapDur := time.Duration(maxGap) * time.Second
	var ranges []timeRange
	var runStart, runEnd time.Time
	inRun := false

	for _, s := range samples {
		isIdle := s.value > minRunningRpm && s.value <= float64(maxIdleRpm)
		if isIdle {
			if !inRun {
				runStart = s.ts
				runEnd = s.ts
				inRun = true
			} else if s.ts.Sub(runEnd) > maxGapDur {
				appendIdleRange(runStart, runEnd, minDuration, from, to, &ranges)
				runStart = s.ts
				runEnd = s.ts
			} else {
				runEnd = s.ts
			}
		} else {
			if inRun {
				appendIdleRange(runStart, runEnd, minDuration, from, to, &ranges)
				inRun = false
			}
		}
	}
	if inRun {
		appendIdleRange(runStart, runEnd, minDuration, from, to, &ranges)
	}
	return ranges
}

// appendIdleRange appends a timeRange if the run meets duration and time-range criteria.
func appendIdleRange(runStart, runEnd time.Time, minDuration int, from, to time.Time, out *[]timeRange) {
	if tr, ok := clipTimeRange(timeRange{start: runStart, end: runEnd}, from, to, minDuration); ok {
		*out = append(*out, tr)
	}
}
