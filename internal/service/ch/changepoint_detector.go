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

// ChangePointDetector detects segments using CUSUM (Cumulative Sum) change point detection
// CUSUM monitors cumulative deviation from expected baseline to detect regime changes
// When vehicle becomes active, signal frequency increases significantly
type ChangePointDetector struct {
	conn clickhouse.Conn
}

// NewChangePointDetector creates a new ChangePointDetector with the given connection
func NewChangePointDetector(conn clickhouse.Conn) *ChangePointDetector {
	return &ChangePointDetector{conn: conn}
}

// CUSUMWindow represents a time window with CUSUM statistic
type CUSUMWindow struct {
	WindowStart         time.Time
	WindowEnd           time.Time
	SignalCount         uint64
	DistinctSignalCount uint64
	CUSUMStat           float64 // Cumulative sum statistic
	IsActive            bool    // Whether CUSUM exceeded threshold
}

// DetectSegments implements CUSUM-based change point detection
func (d *ChangePointDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*Segment, error) {
	// Apply configuration defaults
	maxGap := defaultMinIdleSeconds
	minDuration := defaultMinSegmentDurationSeconds

	if config != nil {
		if config.MinIdleSeconds != nil {
			maxGap = *config.MinIdleSeconds
		}
		if config.MinSegmentDurationSeconds != nil {
			minDuration = *config.MinSegmentDurationSeconds
		}
	}

	// Query signal counts per window with dynamic window size
	// Look back maxGap seconds before 'from' to detect segments that started before the query range
	// This allows us to properly set StartedBeforeRange for ongoing trips
	windowSize := defaultCUSUMWindowSeconds
	signalThreshold := defaultCUSUMSignalCountThreshold
	distinctSignalThreshold := defaultCUSUMDistinctSignalCountThreshold
	lookbackFrom := from.Add(-time.Duration(maxGap) * time.Second)
	windowCounts, err := d.getWindowSignalCounts(ctx, tokenID, lookbackFrom, to, windowSize, signalThreshold, distinctSignalThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get window signal counts: %w", err)
	}

	if len(windowCounts) == 0 {
		return nil, nil
	}

	// Apply CUSUM algorithm in Go (requires sequential processing)
	activeWindows := d.applyCUSUM(windowCounts)

	if len(activeWindows) == 0 {
		return nil, nil
	}

	// Merge active windows into segments
	segments := mergeWindowsIntoSegments(tokenID, activeWindows, from, to, maxGap, minDuration)

	return segments, nil
}

// getWindowSignalCounts gets signal counts per time window
// Uses FINAL to ensure ReplacingMergeTree deduplication before aggregation
//
// Performance notes:
// - PREWHERE filters on primary key (token_id) before FINAL merge
// - Pre-allocates result slice based on expected window count
func (d *ChangePointDetector) getWindowSignalCounts(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	windowSizeSeconds int,
	signalThreshold int,
	distinctSignalThreshold int,
) ([]CUSUMWindow, error) {
	// Query signal table directly with FINAL for accurate counts
	// PREWHERE on token_id filters before FINAL merge (primary key optimization)
	query := `
SELECT
    toStartOfInterval(timestamp, INTERVAL ? second) AS window_start,
    toStartOfInterval(timestamp, INTERVAL ? second) + INTERVAL ? second AS window_end,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal FINAL
PREWHERE token_id = ?
WHERE timestamp >= ?
  AND timestamp < ?
GROUP BY window_start
HAVING signal_count >= ? AND distinct_signal_count >= ?
ORDER BY window_start`

	rows, err := d.conn.Query(ctx, query, windowSizeSeconds, windowSizeSeconds, windowSizeSeconds, tokenID, from, to, signalThreshold, distinctSignalThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed querying window counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Pre-allocate based on expected number of windows
	expectedWindows := int(to.Sub(from).Seconds()) / windowSizeSeconds
	windows := make([]CUSUMWindow, 0, expectedWindows)

	for rows.Next() {
		var w CUSUMWindow
		err := rows.Scan(&w.WindowStart, &w.WindowEnd, &w.SignalCount, &w.DistinctSignalCount)
		if err != nil {
			return nil, fmt.Errorf("failed scanning window: %w", err)
		}
		windows = append(windows, w)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("window row error: %w", rows.Err())
	}

	return windows, nil
}

// applyCUSUM applies the CUSUM algorithm to detect change points in signal frequency
// CUSUM detects when cumulative deviation from baseline exceeds threshold
//
// Algorithm:
//
//	S[t] = max(0, S[t-1] + (x[t] - μ - k))
//	where:
//	  - x[t] is the signal count at time t
//	  - μ is the baseline (expected idle count)
//	  - k is the drift parameter (allowable deviation)
//	  - S[t] > threshold indicates active period
func (d *ChangePointDetector) applyCUSUM(windows []CUSUMWindow) []ActiveWindow {
	n := len(windows)
	if n == 0 {
		return nil
	}

	// CUSUM parameters
	baseline := defaultCUSUMBaselineSignalCount // Expected signal count when idle
	drift := defaultCUSUMDrift                  // Allowable deviation (k parameter)
	threshold := defaultCUSUMThreshold          // Detection threshold (h parameter)

	// Pre-allocate: assume ~50% of windows will be active (reasonable estimate)
	activeWindows := make([]ActiveWindow, 0, n/2)
	cusumStat := 0.0

	for i := range windows {
		// Calculate deviation from baseline
		deviation := float64(windows[i].SignalCount) - baseline - drift

		// Update CUSUM statistic (non-negative cumulative sum)
		cusumStat = max(0, cusumStat+deviation)

		// Store statistic for debugging/analysis
		windows[i].CUSUMStat = cusumStat

		// Check if CUSUM exceeds threshold (detected change point / active period)
		if cusumStat > threshold {
			windows[i].IsActive = true

			// Convert to ActiveWindow for segment merging
			activeWindows = append(activeWindows, ActiveWindow{
				WindowStart:         windows[i].WindowStart,
				WindowEnd:           windows[i].WindowEnd,
				SignalCount:         windows[i].SignalCount,
				DistinctSignalCount: windows[i].DistinctSignalCount,
			})
		} else {
			windows[i].IsActive = false
		}
	}

	return activeWindows
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// GetMechanismName returns the name of this detection mechanism
func (d *ChangePointDetector) GetMechanismName() string {
	return "changePointDetection"
}
