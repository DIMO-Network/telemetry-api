package ch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestApplyCUSUM(t *testing.T) {
	detector := &ChangePointDetector{}
	now := time.Now()

	t.Run("empty windows returns nil", func(t *testing.T) {
		result := detector.applyCUSUM(nil)
		require.Nil(t, result)

		result = detector.applyCUSUM([]CUSUMWindow{})
		require.Nil(t, result)
	})

	t.Run("low signal count windows not marked active", func(t *testing.T) {
		// Signal counts below baseline + drift + threshold won't trigger
		windows := []CUSUMWindow{
			{WindowStart: now, WindowEnd: now.Add(time.Minute), SignalCount: 1},
			{WindowStart: now.Add(time.Minute), WindowEnd: now.Add(2 * time.Minute), SignalCount: 2},
		}

		result := detector.applyCUSUM(windows)
		require.Empty(t, result)
	})

	t.Run("high signal count windows marked active", func(t *testing.T) {
		// High signal counts should accumulate CUSUM and trigger active state
		windows := []CUSUMWindow{
			{WindowStart: now, WindowEnd: now.Add(time.Minute), SignalCount: 20},
			{WindowStart: now.Add(time.Minute), WindowEnd: now.Add(2 * time.Minute), SignalCount: 25},
			{WindowStart: now.Add(2 * time.Minute), WindowEnd: now.Add(3 * time.Minute), SignalCount: 30},
		}

		result := detector.applyCUSUM(windows)
		require.NotEmpty(t, result)
		// All windows should be active due to high signal counts
		require.Len(t, result, 3)
	})

	t.Run("CUSUM resets after idle period", func(t *testing.T) {
		// High activity followed by low activity should reset CUSUM
		windows := []CUSUMWindow{
			{WindowStart: now, WindowEnd: now.Add(time.Minute), SignalCount: 50},
			{WindowStart: now.Add(time.Minute), WindowEnd: now.Add(2 * time.Minute), SignalCount: 50},
			// Gap in data (simulated by low counts)
			{WindowStart: now.Add(10 * time.Minute), WindowEnd: now.Add(11 * time.Minute), SignalCount: 0},
			{WindowStart: now.Add(11 * time.Minute), WindowEnd: now.Add(12 * time.Minute), SignalCount: 0},
			{WindowStart: now.Add(12 * time.Minute), WindowEnd: now.Add(13 * time.Minute), SignalCount: 0},
			// New activity
			{WindowStart: now.Add(20 * time.Minute), WindowEnd: now.Add(21 * time.Minute), SignalCount: 50},
		}

		result := detector.applyCUSUM(windows)
		// Should have active windows from first burst and second burst
		require.NotEmpty(t, result)
	})

	t.Run("gradual increase triggers detection", func(t *testing.T) {
		// Gradual increase should eventually trigger CUSUM threshold
		windows := []CUSUMWindow{
			{WindowStart: now, WindowEnd: now.Add(time.Minute), SignalCount: 5},
			{WindowStart: now.Add(time.Minute), WindowEnd: now.Add(2 * time.Minute), SignalCount: 8},
			{WindowStart: now.Add(2 * time.Minute), WindowEnd: now.Add(3 * time.Minute), SignalCount: 12},
			{WindowStart: now.Add(3 * time.Minute), WindowEnd: now.Add(4 * time.Minute), SignalCount: 15},
			{WindowStart: now.Add(4 * time.Minute), WindowEnd: now.Add(5 * time.Minute), SignalCount: 20},
		}

		result := detector.applyCUSUM(windows)
		// Later windows should be marked active as CUSUM accumulates
		require.NotEmpty(t, result)
	})

	t.Run("preserves window timing information", func(t *testing.T) {
		start := now
		end := now.Add(time.Minute)
		windows := []CUSUMWindow{
			{WindowStart: start, WindowEnd: end, SignalCount: 100, DistinctSignalCount: 5},
		}

		result := detector.applyCUSUM(windows)
		require.Len(t, result, 1)
		require.Equal(t, start, result[0].WindowStart)
		require.Equal(t, end, result[0].WindowEnd)
		require.Equal(t, uint64(100), result[0].SignalCount)
		require.Equal(t, uint64(5), result[0].DistinctSignalCount)
	})
}

func TestCUSUMWindowIsActiveFlag(t *testing.T) {
	detector := &ChangePointDetector{}
	now := time.Now()

	t.Run("IsActive flag set correctly on windows", func(t *testing.T) {
		windows := []CUSUMWindow{
			{WindowStart: now, WindowEnd: now.Add(time.Minute), SignalCount: 1},                        // Should be inactive
			{WindowStart: now.Add(time.Minute), WindowEnd: now.Add(2 * time.Minute), SignalCount: 100}, // Should be active
		}

		_ = detector.applyCUSUM(windows)

		// First window should be inactive (low count)
		require.False(t, windows[0].IsActive)
		// Second window should be active (high count triggers CUSUM)
		require.True(t, windows[1].IsActive)
	})
}
