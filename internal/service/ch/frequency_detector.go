package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultWindowSizeSeconds    = 60 // 1 minute windows
	defaultSignalCountThreshold = 12 // Minimum signals per window for activity
)

// FrequencyDetector detects segments using frequency analysis of signal updates
// Analyzes signal update patterns to identify vehicle activity periods
type FrequencyDetector struct {
	conn clickhouse.Conn
}

// ActiveWindow represents a time window with sufficient signal activity
type ActiveWindow struct {
	WindowStart         time.Time
	WindowEnd           time.Time
	SignalCount         uint32
	DistinctSignalCount uint16
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

	// Query active windows from pre-aggregated table
	windows, err := d.getActiveWindows(ctx, tokenID, from, to, signalThreshold)
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

// getActiveWindows queries pre-aggregated window data
func (d *FrequencyDetector) getActiveWindows(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	signalThreshold int,
) ([]ActiveWindow, error) {
	// Always use 1-minute windows (60 seconds) for finest granularity
	// Coarser windows (5m, 1h) available for future optimization
	windowSize := defaultWindowSizeSeconds

	query := `
SELECT
    window_start,
    window_start + INTERVAL window_size_seconds second AS window_end,
    signal_count,
    distinct_signal_count
FROM signal_window_aggregates
WHERE token_id = ?
  AND window_size_seconds = ?
  AND window_start >= ?
  AND window_start < ?
  AND signal_count >= ?
ORDER BY window_start`

	rows, err := d.conn.Query(ctx, query, tokenID, windowSize, from, to, signalThreshold)
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
