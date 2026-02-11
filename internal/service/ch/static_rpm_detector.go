package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultIdleWindowSizeSeconds     = 60   // 1 minute windows (same as frequency detector)
	defaultMaxIdleRpm                = 1500
	engineSpeedSignalName            = "powertrainCombustionEngineSpeed" // fixed; not configurable
	defaultSignalCountThresholdIdle  = 2    // powertrainCombustionEngineSpeed ~2/min; min samples per 1min window
	defaultMinIdleRpmForEngineRunning = 0   // min_rpm > this to exclude engine-off
)

// StaticRpmDetector detects segments where engine RPM remains in idle range (static/low RPM).
// Uses repeated windows of idle RPM merged like trips.
type StaticRpmDetector struct {
	conn clickhouse.Conn
}

// NewStaticRpmDetector creates a new StaticRpmDetector with the given connection.
func NewStaticRpmDetector(conn clickhouse.Conn) *StaticRpmDetector {
	return &StaticRpmDetector{conn: conn}
}

// DetectSegments implements idle-RPM-based segment detection.
func (d *StaticRpmDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*Segment, error) {
	maxGap := defaultMinIdleSeconds
	minDuration := defaultMinSegmentDurationSeconds
	maxIdleRpm := defaultMaxIdleRpm
	signalThreshold := defaultSignalCountThresholdIdle

	if config != nil {
		if config.MinIdleSeconds != nil {
			maxGap = *config.MinIdleSeconds
		}
		if config.MinSegmentDurationSeconds != nil {
			minDuration = *config.MinSegmentDurationSeconds
		}
		if config.MaxIdleRpm != nil {
			maxIdleRpm = *config.MaxIdleRpm
		}
		if config.SignalCountThreshold != nil {
			signalThreshold = *config.SignalCountThreshold
		}
	}

	windowSize := defaultIdleWindowSizeSeconds
	lookbackFrom := from.Add(-time.Duration(maxGap) * time.Second)
	windows, err := d.getIdleWindows(ctx, tokenID, lookbackFrom, to, windowSize, maxIdleRpm, signalThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get idle windows: %w", err)
	}

	if len(windows) == 0 {
		return nil, nil
	}

	segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
	return segments, nil
}

// getIdleWindows returns time windows where engine speed is in idle band (0 < rpm <= maxIdleRpm).
// Uses signal FINAL; groups by window and keeps windows with sample_count >= signalThreshold and max(rpm) <= maxIdleRpm and min(rpm) > 0.
func (d *StaticRpmDetector) getIdleWindows(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	windowSizeSeconds int,
	maxIdleRpm int,
	signalThreshold int,
) ([]ActiveWindow, error) {
	query := `
SELECT
    toStartOfInterval(timestamp, INTERVAL ? second) AS window_start,
    toStartOfInterval(timestamp, INTERVAL ? second) + INTERVAL ? second AS window_end,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal FINAL
PREWHERE token_id = ?
WHERE name = ?
  AND timestamp >= ?
  AND timestamp < ?
GROUP BY window_start
HAVING signal_count >= ? AND max(value_number) <= ? AND min(value_number) > ?
ORDER BY window_start`

	rows, err := d.conn.Query(ctx, query,
		windowSizeSeconds, windowSizeSeconds, windowSizeSeconds,
		tokenID, engineSpeedSignalName, from, to,
		signalThreshold, maxIdleRpm, defaultMinIdleRpmForEngineRunning)
	if err != nil {
		return nil, fmt.Errorf("failed querying idle windows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	expectedWindows := int(to.Sub(from).Seconds()) / windowSizeSeconds
	if expectedWindows <= 0 {
		expectedWindows = 1
	}
	windows := make([]ActiveWindow, 0, expectedWindows)

	for rows.Next() {
		var w ActiveWindow
		err := rows.Scan(&w.WindowStart, &w.WindowEnd, &w.SignalCount, &w.DistinctSignalCount)
		if err != nil {
			return nil, fmt.Errorf("failed scanning idle window row: %w", err)
		}
		windows = append(windows, w)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("idle window row error: %w", rows.Err())
	}

	return windows, nil
}

// GetMechanismName returns the name of this detection mechanism.
func (d *StaticRpmDetector) GetMechanismName() string {
	return "staticRpm"
}
