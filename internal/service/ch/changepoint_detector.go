package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultCUSUMWindowSeconds                = 60  // 1-minute windows for frequency measurement
	defaultCUSUMThreshold                    = 5.0 // CUSUM threshold for detecting change points
	defaultCUSUMDrift                        = 0.5 // Drift parameter (half of expected change magnitude)
	defaultCUSUMBaselineSignalCount          = 1.0 // Baseline signal count per window when idle
	defaultCUSUMSignalCountThreshold         = 10  // Minimum signals per window for activity
	defaultCUSUMDistinctSignalCountThreshold = 2   // Minimum distinct signal types per window
)

// ChangePointDetector detects segments using CUSUM (Cumulative Sum) change point detection.
// CUSUM monitors cumulative deviation from expected baseline to detect regime changes.
// When vehicle becomes active, signal frequency increases significantly.
type ChangePointDetector struct {
	conn clickhouse.Conn
}

// NewChangePointDetector creates a new ChangePointDetector with the given connection.
func NewChangePointDetector(conn clickhouse.Conn) *ChangePointDetector {
	return &ChangePointDetector{conn: conn}
}

// DetectSegments implements CUSUM-based change point detection
func (d *ChangePointDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*model.Segment, error) {
	rc := resolveBaseConfig(config)
	maxGap := rc.maxGapSeconds
	minDuration := rc.minDuration

	// Look back maxGap seconds before 'from' to detect segments that started before the query range.
	lookbackFrom := from.Add(-time.Duration(maxGap) * time.Second)
	windows, err := getWindowedSignalCounts(ctx, d.conn, tokenID, lookbackFrom, to, defaultCUSUMWindowSeconds, defaultCUSUMSignalCountThreshold, defaultCUSUMDistinctSignalCountThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query window signal counts: %w", err)
	}

	if len(windows) == 0 {
		return []*model.Segment{}, nil
	}

	// Apply CUSUM algorithm in Go (requires sequential processing)
	activeWindows := d.applyCUSUM(windows)

	if len(activeWindows) == 0 {
		return []*model.Segment{}, nil
	}

	// Merge active windows into segments
	segments := mergeWindowsIntoSegments(activeWindows, from, to, maxGap, minDuration)

	return segments, nil
}

// applyCUSUM applies the CUSUM algorithm to detect change points in signal frequency.
// Returns only windows where the cumulative deviation from baseline exceeds the threshold.
//
// Algorithm:
//
//	S[t] = max(0, S[t-1] + (x[t] - μ - k))
//	where:
//	  - x[t] is the signal count at time t
//	  - μ is the baseline (expected idle count)
//	  - k is the drift parameter (allowable deviation)
//	  - S[t] > threshold indicates active period
func (d *ChangePointDetector) applyCUSUM(windows []ActiveWindow) []ActiveWindow {
	n := len(windows)
	if n == 0 {
		return nil
	}

	baseline := defaultCUSUMBaselineSignalCount
	drift := defaultCUSUMDrift
	threshold := defaultCUSUMThreshold

	active := make([]ActiveWindow, 0, n/2)
	cusumStat := 0.0

	for i := range windows {
		deviation := float64(windows[i].SignalCount) - baseline - drift
		cusumStat = max(0, cusumStat+deviation)
		if cusumStat > threshold {
			active = append(active, windows[i])
		}
	}

	return active
}

// GetMechanismName returns the name of this detection mechanism
func (d *ChangePointDetector) GetMechanismName() string {
	return "changePointDetection"
}
