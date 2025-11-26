package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultMinIdleSeconds            = 600 // 5 minutes
	defaultMinSegmentDurationSeconds = 150 // 5 minutes
)

// IgnitionDetector detects segments using ignition state transitions
type IgnitionDetector struct {
	conn clickhouse.Conn
}

// NewIgnitionDetector creates a new IgnitionDetector with the given connection
func NewIgnitionDetector(conn clickhouse.Conn) *IgnitionDetector {
	return &IgnitionDetector{conn: conn}
}

// StateChange represents a single state change from the signal_state_changes table
type StateChange struct {
	Timestamp time.Time
	State     float64
	PrevState float64
}

// DetectSegments implements ignition-based segment detection
func (d *IgnitionDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*Segment, error) {
	// Apply defaults
	minIdle := defaultMinIdleSeconds
	minDuration := defaultMinSegmentDurationSeconds

	if config != nil {
		if config.MinIdleSeconds != nil {
			minIdle = *config.MinIdleSeconds
		}
		if config.MinSegmentDurationSeconds != nil {
			minDuration = *config.MinSegmentDurationSeconds
		}
	}

	// Fetch all state changes with a single query that includes:
	// 1. The most recent state change before 'from' (to detect ongoing trips)
	// 2. All state changes within the query range [from, to)
	stmt, args := d.getStateChangesQueryWithLookback(tokenID, from, to)

	rows, err := d.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse for state changes: %w", err)
	}
	defer rows.Close()

	var stateChanges []StateChange
	for rows.Next() {
		var sc StateChange
		err := rows.Scan(&sc.Timestamp, &sc.State, &sc.PrevState)
		if err != nil {
			return nil, fmt.Errorf("failed scanning state change row: %w", err)
		}
		stateChanges = append(stateChanges, sc)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("state change row error: %w", rows.Err())
	}

	// Process state changes in Go to build segments with debouncing
	segments := d.buildSegmentsWithDebouncing(tokenID, stateChanges, from, to, minIdle, minDuration)

	return segments, nil
}

// getStateChangesQueryWithLookback builds a single query that fetches:
// 1. The most recent state change before 'from' (if ignition was ON, we have an ongoing trip)
// 2. All state changes within the range [from, to)
// Results are ordered by timestamp.
func (d *IgnitionDetector) getStateChangesQueryWithLookback(tokenID uint32, from, to time.Time) (string, []any) {
	// Use UNION ALL to combine:
	// - Last state change before 'from' (LIMIT 1, ordered DESC then re-ordered)
	// - All state changes in range [from, to)
	query := `
SELECT timestamp, new_state, prev_state FROM (
  -- Most recent state change before 'from' (to detect ongoing trips)
  SELECT timestamp, new_state, prev_state
  FROM signal_state_changes FINAL
  WHERE token_id = ?
    AND signal_name = 'isIgnitionOn'
    AND timestamp < ?
    AND prev_state != new_state
  ORDER BY timestamp DESC
  LIMIT 1

  UNION ALL

  -- All state changes within the query range
  SELECT timestamp, new_state, prev_state
  FROM signal_state_changes FINAL
  WHERE token_id = ?
    AND signal_name = 'isIgnitionOn'
    AND timestamp >= ?
    AND timestamp < ?
    AND prev_state != new_state
)
ORDER BY timestamp`

	args := []any{tokenID, from, tokenID, from, to}

	return query, args
}

// buildSegmentsWithDebouncing processes state changes and applies debouncing logic
// to merge consecutive short segments separated by less than minIdle seconds
func (d *IgnitionDetector) buildSegmentsWithDebouncing(tokenID uint32, stateChanges []StateChange, from, to time.Time, minIdle, minDuration int) []*Segment {
	if len(stateChanges) == 0 {
		return nil
	}

	// First pass: filter out noise (OFF signals followed by ON within minIdle seconds)
	filtered := d.filterNoise(stateChanges, minIdle)

	// Second pass: build segments from cleaned state changes
	var segments []*Segment
	var currentSegmentStart *time.Time
	var startedWithPrevMinus1 bool

	for _, sc := range filtered {
		if sc.State == 1 {
			// ON signal - start a new segment if we don't have one
			if currentSegmentStart == nil && sc.PrevState != 1 {
				currentSegmentStart = &sc.Timestamp
				startedWithPrevMinus1 = (sc.PrevState == -1)
			}
		} else if sc.State == 0 && currentSegmentStart != nil {
			// OFF signal - end the segment
			segmentEnd := sc.Timestamp
			duration := int32(segmentEnd.Sub(*currentSegmentStart).Seconds())

			if int(duration) >= minDuration {
				segments = append(segments, &Segment{
					TokenID:            tokenID,
					StartTime:          *currentSegmentStart,
					EndTime:            &segmentEnd,
					DurationSeconds:    duration,
					IsOngoing:          false,
					StartedBeforeRange: currentSegmentStart.Before(from),
				})
			}

			currentSegmentStart = nil
			startedWithPrevMinus1 = false
		}
	}

	// Handle ongoing segment (started but no end signal)
	if currentSegmentStart != nil && !startedWithPrevMinus1 {
		duration := int32(to.Sub(*currentSegmentStart).Seconds())
		if int(duration) >= minDuration {
			segments = append(segments, &Segment{
				TokenID:            tokenID,
				StartTime:          *currentSegmentStart,
				EndTime:            nil,
				DurationSeconds:    duration,
				IsOngoing:          true,
				StartedBeforeRange: currentSegmentStart.Before(from),
			})
		}
	}

	return segments
}

// filterNoise removes OFF signals that are followed by ON within minIdle seconds
func (d *IgnitionDetector) filterNoise(stateChanges []StateChange, minIdle int) []StateChange {
	if len(stateChanges) == 0 {
		return nil
	}

	filtered := make([]StateChange, 0, len(stateChanges))
	minIdleDuration := time.Duration(minIdle) * time.Second

	for i := 0; i < len(stateChanges); i++ {
		sc := stateChanges[i]

		// Keep all ON signals
		if sc.State == 1 {
			filtered = append(filtered, sc)
			continue
		}

		// For OFF signals, check if next ON is within minIdle
		if sc.State == 0 {
			// Find next ON signal
			keep := true
			for j := i + 1; j < len(stateChanges); j++ {
				if stateChanges[j].State == 1 {
					gap := stateChanges[j].Timestamp.Sub(sc.Timestamp)
					if gap < minIdleDuration {
						// Gap too short - this OFF is noise, skip it
						keep = false
					}
					break
				}
			}
			if keep {
				filtered = append(filtered, sc)
			}
		}
	}

	return filtered
}

// GetMechanismName returns the name of this detection mechanism
func (d *IgnitionDetector) GetMechanismName() string {
	return "ignitionDetection"
}
