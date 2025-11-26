package ch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFilterNoise(t *testing.T) {
	detector := &IgnitionDetector{}
	now := time.Now()
	minIdle := 300 // 5 minutes

	t.Run("empty input returns nil", func(t *testing.T) {
		result := detector.filterNoise(nil, minIdle)
		require.Nil(t, result)

		result = detector.filterNoise([]StateChange{}, minIdle)
		require.Nil(t, result)
	})

	t.Run("keeps all ON signals", func(t *testing.T) {
		changes := []StateChange{
			{Timestamp: now, State: 1, PrevState: 0},
			{Timestamp: now.Add(time.Minute), State: 1, PrevState: 0},
		}

		result := detector.filterNoise(changes, minIdle)
		require.Len(t, result, 2)
	})

	t.Run("filters short OFF followed by ON", func(t *testing.T) {
		// OFF at T=0, ON at T=1min (< 5min idle) - OFF should be filtered
		changes := []StateChange{
			{Timestamp: now, State: 1, PrevState: 0},
			{Timestamp: now.Add(time.Minute), State: 0, PrevState: 1},
			{Timestamp: now.Add(2 * time.Minute), State: 1, PrevState: 0}, // Only 1 min gap
		}

		result := detector.filterNoise(changes, minIdle)
		// Should keep ON signals, filter the short OFF
		require.Len(t, result, 2)
		require.Equal(t, float64(1), result[0].State)
		require.Equal(t, float64(1), result[1].State)
	})

	t.Run("keeps long OFF signals", func(t *testing.T) {
		// OFF at T=0, ON at T=10min (> 5min idle) - OFF should be kept
		changes := []StateChange{
			{Timestamp: now, State: 1, PrevState: 0},
			{Timestamp: now.Add(time.Minute), State: 0, PrevState: 1},
			{Timestamp: now.Add(10 * time.Minute), State: 1, PrevState: 0}, // 9 min gap > 5 min
		}

		result := detector.filterNoise(changes, minIdle)
		require.Len(t, result, 3)
	})

	t.Run("keeps final OFF with no following ON", func(t *testing.T) {
		changes := []StateChange{
			{Timestamp: now, State: 1, PrevState: 0},
			{Timestamp: now.Add(10 * time.Minute), State: 0, PrevState: 1},
		}

		result := detector.filterNoise(changes, minIdle)
		require.Len(t, result, 2)
		require.Equal(t, float64(0), result[1].State)
	})
}

func TestBuildSegmentsWithDebouncing(t *testing.T) {
	detector := &IgnitionDetector{}
	now := time.Now()
	tokenID := uint32(1)
	minIdle := 300    // 5 minutes
	minDuration := 60 // 1 minute

	t.Run("empty input returns nil", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		result := detector.buildSegmentsWithDebouncing(tokenID, nil, from, to, minIdle, minDuration)
		require.Nil(t, result)

		result = detector.buildSegmentsWithDebouncing(tokenID, []StateChange{}, from, to, minIdle, minDuration)
		require.Nil(t, result)
	})

	t.Run("simple ON/OFF creates segment", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		changes := []StateChange{
			{Timestamp: from.Add(10 * time.Minute), State: 1, PrevState: 0},
			{Timestamp: from.Add(20 * time.Minute), State: 0, PrevState: 1},
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Len(t, result, 1)
		require.Equal(t, from.Add(10*time.Minute), result[0].StartTime)
		require.NotNil(t, result[0].EndTime)
		require.Equal(t, from.Add(20*time.Minute), *result[0].EndTime)
		require.False(t, result[0].IsOngoing)
	})

	t.Run("multiple segments", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		changes := []StateChange{
			// Segment 1: 10-20 min
			{Timestamp: from.Add(10 * time.Minute), State: 1, PrevState: 0},
			{Timestamp: from.Add(20 * time.Minute), State: 0, PrevState: 1},
			// Segment 2: 40-50 min
			{Timestamp: from.Add(40 * time.Minute), State: 1, PrevState: 0},
			{Timestamp: from.Add(50 * time.Minute), State: 0, PrevState: 1},
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Len(t, result, 2)
	})

	t.Run("ongoing segment without OFF", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		changes := []StateChange{
			{Timestamp: from.Add(10 * time.Minute), State: 1, PrevState: 0},
			// No OFF signal
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Len(t, result, 1)
		require.True(t, result[0].IsOngoing)
		require.Nil(t, result[0].EndTime)
	})

	t.Run("filters short segments", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		changes := []StateChange{
			// Very short segment (30 seconds < minDuration)
			{Timestamp: from.Add(10 * time.Minute), State: 1, PrevState: 0},
			{Timestamp: from.Add(10*time.Minute + 30*time.Second), State: 0, PrevState: 1},
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Empty(t, result)
	})

	t.Run("ignores initial ON with prev_state -1", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		changes := []StateChange{
			// ON with prev_state -1 means unknown previous state
			{Timestamp: from.Add(10 * time.Minute), State: 1, PrevState: -1},
			// This should not create an ongoing segment
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Empty(t, result)
	})

	t.Run("startedBeforeRange flag set correctly", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		// Segment starts before the query range
		changes := []StateChange{
			{Timestamp: from.Add(-10 * time.Minute), State: 1, PrevState: 0}, // Before 'from'
			{Timestamp: from.Add(10 * time.Minute), State: 0, PrevState: 1},
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Len(t, result, 1)
		require.True(t, result[0].StartedBeforeRange)
	})

	t.Run("duration calculated correctly", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		start := from.Add(10 * time.Minute)
		end := from.Add(20 * time.Minute)
		changes := []StateChange{
			{Timestamp: start, State: 1, PrevState: 0},
			{Timestamp: end, State: 0, PrevState: 1},
		}

		result := detector.buildSegmentsWithDebouncing(tokenID, changes, from, to, minIdle, minDuration)
		require.Len(t, result, 1)
		require.Equal(t, int32(600), result[0].DurationSeconds) // 10 minutes = 600 seconds
	})
}

