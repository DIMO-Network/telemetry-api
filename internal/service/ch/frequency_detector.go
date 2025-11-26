package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultWindowSizeSeconds            = 60 // 1 minute windows
	defaultSignalCountThreshold         = 12 // Minimum signals per window for activity
	defaultDistinctSignalCountThreshold = 2  // Minimum distinct signal types per window
)

// FrequencyDetector detects segments using frequency analysis of signal updates
// Analyzes signal update patterns to identify vehicle activity periods
type FrequencyDetector struct {
	conn clickhouse.Conn
}

// NewFrequencyDetector creates a new FrequencyDetector with the given connection
func NewFrequencyDetector(conn clickhouse.Conn) *FrequencyDetector {
	return &FrequencyDetector{conn: conn}
}

// ActiveWindow represents a time window with sufficient signal activity
type ActiveWindow struct {
	WindowStart         time.Time
	WindowEnd           time.Time
	SignalCount         uint64
	DistinctSignalCount uint64
}

// DetectSegments implements frequency-based segment detection
func (d *FrequencyDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*Segment, error) {
	// Apply configuration defaults
	signalThreshold := defaultSignalCountThreshold
	maxGap := defaultMinIdleSeconds
	minDuration := defaultMinSegmentDurationSeconds

	if config != nil {
		if config.MinIdleSeconds != nil {
			maxGap = *config.MinIdleSeconds
		}
		if config.MinSegmentDurationSeconds != nil {
			minDuration = *config.MinSegmentDurationSeconds
		}
		if config.SignalCountThreshold != nil {
			signalThreshold = *config.SignalCountThreshold
		}
	}

	// Query active windows with dynamic window size
	windowSize := defaultWindowSizeSeconds
	distinctSignalThreshold := defaultDistinctSignalCountThreshold
	windows, err := d.getActiveWindows(ctx, tokenID, from, to, windowSize, signalThreshold, distinctSignalThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get active windows: %w", err)
	}

	if len(windows) == 0 {
		return nil, nil
	}

	// Merge consecutive active windows into segments (in Go for flexibility)
	segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)

	return segments, nil
}

// getActiveWindows queries signal data with deduplication
// Uses FINAL to ensure ReplacingMergeTree deduplication before aggregation
func (d *FrequencyDetector) getActiveWindows(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	windowSizeSeconds int,
	signalThreshold int,
	distinctSignalThreshold int,
) ([]ActiveWindow, error) {
	// Query signal table directly with FINAL for accurate counts
	// Uses toStartOfInterval for flexible window sizes
	query := `
SELECT
    toStartOfInterval(timestamp, INTERVAL ? second) AS window_start,
    toStartOfInterval(timestamp, INTERVAL ? second) + INTERVAL ? second AS window_end,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal FINAL
WHERE token_id = ?
  AND timestamp >= ?
  AND timestamp < ?
GROUP BY window_start
HAVING signal_count >= ? AND distinct_signal_count >= ?
ORDER BY window_start`

	rows, err := d.conn.Query(ctx, query, windowSizeSeconds, windowSizeSeconds, windowSizeSeconds, tokenID, from, to, signalThreshold, distinctSignalThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed querying active windows: %w", err)
	}
	defer rows.Close()

	var windows []ActiveWindow
	for rows.Next() {
		var w ActiveWindow
		err := rows.Scan(&w.WindowStart, &w.WindowEnd, &w.SignalCount, &w.DistinctSignalCount)
		if err != nil {
			return nil, fmt.Errorf("failed scanning window row: %w", err)
		}
		windows = append(windows, w)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("window row error: %w", rows.Err())
	}

	return windows, nil
}

// mergeWindowsIntoSegments merges consecutive active windows with gap tolerance
func mergeWindowsIntoSegments(
	tokenID uint32,
	windows []ActiveWindow,
	from, to time.Time,
	maxGapSeconds int,
	minDurationSeconds int,
) []*Segment {
	if len(windows) == 0 {
		return nil
	}

	var segments []*Segment
	currentStart := windows[0].WindowStart
	currentEnd := windows[0].WindowEnd

	for i := 1; i < len(windows); i++ {
		gapSeconds := int(windows[i].WindowStart.Sub(currentEnd).Seconds())

		if gapSeconds <= maxGapSeconds {
			// Gap is small enough - extend current segment
			currentEnd = windows[i].WindowEnd
		} else {
			// Gap exceeds threshold - finalize current segment and start new one
			duration := int32(currentEnd.Sub(currentStart).Seconds())
			if int(duration) >= minDurationSeconds {
				endTime := currentEnd
				segments = append(segments, &Segment{
					TokenID:            tokenID,
					StartTime:          currentStart,
					EndTime:            &endTime,
					DurationSeconds:    duration,
					IsOngoing:          false,
					StartedBeforeRange: currentStart.Before(from),
				})
			}

			currentStart = windows[i].WindowStart
			currentEnd = windows[i].WindowEnd
		}
	}

	// Finalize last segment
	duration := int32(currentEnd.Sub(currentStart).Seconds())
	if int(duration) >= minDurationSeconds {
		// Check if segment extends beyond query range (ongoing)
		isOngoing := currentEnd.Equal(to) || currentEnd.After(to)

		// Also consider ongoing if we are querying near real-time (to is close to Now)
		// and the segment ended recently (within maxGapSeconds).
		// This handles the case where the vehicle is still active but we haven't seen
		// the *next* window yet, or the gap hasn't exceeded maxGapSeconds.
		if !isOngoing {
			now := time.Now()
			idleDuration := time.Duration(maxGapSeconds) * time.Second
			// If 'to' is within recent history (to >= Now - idle_time)
			// AND the segment ended within recent history (currentEnd >= Now - idle_time)
			if to.After(now.Add(-idleDuration)) && currentEnd.After(now.Add(-idleDuration)) {
				isOngoing = true
			}
		}

		var endTime *time.Time
		if !isOngoing {
			endTime = &currentEnd
		} else {
			// Ongoing segment: use query 'to' time as end
			endTime = nil
			duration = int32(to.Sub(currentStart).Seconds())
		}

		segments = append(segments, &Segment{
			TokenID:            tokenID,
			StartTime:          currentStart,
			EndTime:            endTime,
			DurationSeconds:    duration,
			IsOngoing:          isOngoing,
			StartedBeforeRange: currentStart.Before(from),
		})
	}

	return segments
}

// GetMechanismName returns the name of this detection mechanism
func (d *FrequencyDetector) GetMechanismName() string {
	return "frequencyAnalysis"
}
