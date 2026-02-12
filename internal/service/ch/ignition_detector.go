package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultMinIdleSeconds            = 300 // 5 minutes
	defaultMinSegmentDurationSeconds = 240 // 4 minutes
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
	defer func() { _ = rows.Close() }()

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

// maxLookbackDays limits how far back we search for prior state changes
const maxLookbackDays = 30

// getStateChangesQueryWithLookback builds a single query that fetches:
// 1. The most recent state change before 'from' (if ignition was ON, we have an ongoing trip)
// 2. All state changes within the range [from, to)
// Results are ordered by timestamp.
//
// Performance notes:
// - PREWHERE filters on primary key columns before FINAL merge (much faster)
// - Lookback is bounded to maxLookbackDays to prevent unbounded scans
func (d *IgnitionDetector) getStateChangesQueryWithLookback(tokenID uint32, from, to time.Time) (string, []any) {
	// Bound the lookback to prevent scanning unlimited history
	lookbackLimit := from.AddDate(0, 0, -maxLookbackDays)

	// Use UNION ALL to combine:
	// - Last state change before 'from' (LIMIT 1, ordered DESC then re-ordered)
	// - All state changes in range [from, to)
	//
	// PREWHERE on token_id filters before FINAL merge, significantly reducing work
	query := `
SELECT timestamp, new_state, prev_state FROM (
  -- Most recent state change before 'from' (to detect ongoing trips)
  SELECT timestamp, new_state, prev_state
  FROM signal_state_changes FINAL
  PREWHERE token_id = ?
  WHERE signal_name = 'isIgnitionOn'
    AND timestamp >= ?
    AND timestamp < ?
    AND prev_state != new_state
  ORDER BY timestamp DESC
  LIMIT 1

  UNION ALL

  -- All state changes within the query range
  SELECT timestamp, new_state, prev_state
  FROM signal_state_changes FINAL
  PREWHERE token_id = ?
  WHERE signal_name = 'isIgnitionOn'
    AND timestamp >= ?
    AND timestamp < ?
    AND prev_state != new_state
)
ORDER BY timestamp`

	args := []any{tokenID, lookbackLimit, from, tokenID, from, to}

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
// O(n) algorithm: pre-compute next ON index for each position, then single pass
func (d *IgnitionDetector) filterNoise(stateChanges []StateChange, minIdle int) []StateChange {
	n := len(stateChanges)
	if n == 0 {
		return nil
	}

	// Pre-compute next ON signal index for each position (O(n) reverse scan)
	nextON := make([]int, n)
	lastON := -1
	for i := n - 1; i >= 0; i-- {
		if stateChanges[i].State == 1 {
			lastON = i
		}
		nextON[i] = lastON
	}

	filtered := make([]StateChange, 0, n)
	minIdleDuration := time.Duration(minIdle) * time.Second

	for i := 0; i < n; i++ {
		sc := stateChanges[i]

		// Keep all ON signals
		if sc.State == 1 {
			filtered = append(filtered, sc)
			continue
		}

		// For OFF signals, check if next ON is within minIdle (O(1) lookup)
		if sc.State == 0 {
			keep := true
			if j := nextON[i]; j > i {
				gap := stateChanges[j].Timestamp.Sub(sc.Timestamp)
				if gap < minIdleDuration {
					keep = false
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
