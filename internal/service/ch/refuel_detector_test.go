package ch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFindRefuelTroughAndPeak(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("basic refuel rise", func(t *testing.T) {
		// Fuel drops to 20 then jumps to 80
		samples := []levelSample{
			{ts: min(0), value: 50},
			{ts: min(1), value: 40},
			{ts: min(2), value: 25},
			{ts: min(3), value: 20}, // trough
			{ts: min(4), value: 60},
			{ts: min(5), value: 75},
			{ts: min(6), value: 80}, // peak
			{ts: min(7), value: 80},
		}
		trough, peak, absRise := findRefuelTroughAndPeak(samples, min(3), min(5))
		require.Equal(t, min(3), trough)
		require.Equal(t, min(7), peak) // walks forward until stabilization or end
		require.InDelta(t, 60.0, absRise, 0.01)
	})

	t.Run("single sample returns zero", func(t *testing.T) {
		samples := []levelSample{{ts: min(0), value: 50}}
		trough, peak, absRise := findRefuelTroughAndPeak(samples, min(0), min(5))
		// With only one sample, trough is found but peak search starts beyond riseEnd and finds nothing
		require.True(t, trough.IsZero() || peak.IsZero() || absRise == 0)
	})

	t.Run("peak search capped by deadline", func(t *testing.T) {
		// Peak is far beyond the 30-min deadline
		samples := []levelSample{
			{ts: min(0), value: 10},
			{ts: min(5), value: 60},
			{ts: min(60), value: 95}, // beyond deadline
		}
		trough, peak, _ := findRefuelTroughAndPeak(samples, min(0), min(5))
		// Peak should be at min(5) since min(60) is past the 30-min deadline from min(5)
		require.False(t, trough.IsZero())
		require.Equal(t, min(5), peak)
	})

	t.Run("trough walk-back finds local minimum", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 60},
			{ts: min(1), value: 50},
			{ts: min(2), value: 30},
			{ts: min(3), value: 15}, // local min
			{ts: min(4), value: 20}, // riseStart
			{ts: min(5), value: 70}, // riseEnd
		}
		trough, _, _ := findRefuelTroughAndPeak(samples, min(4), min(5))
		require.Equal(t, min(3), trough) // should walk back to 15
	})

	t.Run("peak stabilization detection", func(t *testing.T) {
		// Fuel rises then drops slightly (> 1.0) indicating stabilization
		samples := []levelSample{
			{ts: min(0), value: 10},
			{ts: min(5), value: 70},
			{ts: min(6), value: 75},
			{ts: min(7), value: 73}, // drop of 2.0 from 75 → triggers early stabilization
			{ts: min(8), value: 72},
		}
		_, peak, _ := findRefuelTroughAndPeak(samples, min(0), min(5))
		require.Equal(t, min(6), peak) // stabilized at 75
	})
}

func TestRefuelMergeTimeRanges(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("overlapping ranges merged", func(t *testing.T) {
		ranges := []timeRange{
			{start: min(0), end: min(10)},
			{start: min(5), end: min(15)},
		}
		merged := mergeTimeRanges(ranges, 0, 0, base, min(30), nil)
		require.Len(t, merged, 1)
		require.Equal(t, min(0), merged[0].start)
		require.Equal(t, min(15), merged[0].end)
	})

	t.Run("non-overlapping ranges stay separate", func(t *testing.T) {
		ranges := []timeRange{
			{start: min(0), end: min(5)},
			{start: min(10), end: min(15)},
		}
		merged := mergeTimeRanges(ranges, 0, 0, base, min(30), nil)
		require.Len(t, merged, 2)
	})

	t.Run("short ranges filtered by minDuration", func(t *testing.T) {
		ranges := []timeRange{
			{start: min(0), end: min(1)}, // 60s < 240s default
		}
		merged := mergeTimeRanges(ranges, 0, 240, base, min(30), nil)
		require.Empty(t, merged)
	})
}
