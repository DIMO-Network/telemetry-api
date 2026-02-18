package ch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSmoothSamples(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("window=1 returns input unchanged", func(t *testing.T) {
		samples := []levelSample{{ts: min(0), value: 10}, {ts: min(1), value: 20}}
		result := smoothSamples(samples, 1)
		require.Equal(t, samples, result)
	})

	t.Run("window larger than samples returns input unchanged", func(t *testing.T) {
		samples := []levelSample{{ts: min(0), value: 10}, {ts: min(1), value: 20}}
		result := smoothSamples(samples, 5)
		require.Equal(t, samples, result)
	})

	t.Run("window=3 computes rolling average", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 10},
			{ts: min(1), value: 20},
			{ts: min(2), value: 30},
			{ts: min(3), value: 40},
			{ts: min(4), value: 50},
		}
		result := smoothSamples(samples, 3)
		require.Len(t, result, 3) // 5 - 3 + 1
		// avg of [10,20,30] = 20, ts from center (min(1))
		require.InDelta(t, 20.0, result[0].value, 0.01)
		require.Equal(t, min(1), result[0].ts)
		// avg of [20,30,40] = 30
		require.InDelta(t, 30.0, result[1].value, 0.01)
		// avg of [30,40,50] = 40
		require.InDelta(t, 40.0, result[2].value, 0.01)
	})
}

func TestFindTroughToPeakRanges(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("empty returns nil", func(t *testing.T) {
		require.Nil(t, findTroughToPeakRanges(nil, 1.0, 61))
		require.Nil(t, findTroughToPeakRanges([]levelSample{{ts: min(0), value: 10}}, 1.0, 61))
	})

	t.Run("single rise detected", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 20},
			{ts: min(2), value: 22},
			{ts: min(4), value: 25},
		}
		ranges := findTroughToPeakRanges(samples, 1.0, 60)
		require.Len(t, ranges, 1)
		require.Equal(t, min(0), ranges[0].start)
		require.Equal(t, min(4), ranges[0].end)
	})

	t.Run("rise below minRisePct filtered", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 50},
			{ts: min(2), value: 50.5}, // rise of 0.5 < 1.0
		}
		ranges := findTroughToPeakRanges(samples, 1.0, 0)
		require.Empty(t, ranges)
	})

	t.Run("rise below minDuration filtered", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 20},
			{ts: min(0) .Add(30 * time.Second), value: 30}, // 30s < 61s
		}
		ranges := findTroughToPeakRanges(samples, 1.0, 61)
		require.Empty(t, ranges)
	})

	t.Run("two rises with dip between", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 20},
			{ts: min(5), value: 30}, // peak 1
			{ts: min(10), value: 25}, // dip
			{ts: min(15), value: 40}, // peak 2
		}
		ranges := findTroughToPeakRanges(samples, 1.0, 60)
		require.Len(t, ranges, 2)
	})
}

func TestFilterRangesBySocAndOdo(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("keeps range with SoC increase and no odometer change", func(t *testing.T) {
		ranges := []timeRange{{start: min(0), end: min(10)}}
		soc := []levelSample{{ts: min(0), value: 20}, {ts: min(10), value: 80}}
		odo := []levelSample{{ts: min(0), value: 1000}, {ts: min(10), value: 1000}}
		result := filterRangesBySocAndOdo(ranges, soc, odo)
		require.Len(t, result, 1)
	})

	t.Run("filters range where SoC decreases", func(t *testing.T) {
		ranges := []timeRange{{start: min(0), end: min(10)}}
		soc := []levelSample{{ts: min(0), value: 80}, {ts: min(10), value: 60}}
		odo := []levelSample{{ts: min(0), value: 1000}, {ts: min(10), value: 1000}}
		result := filterRangesBySocAndOdo(ranges, soc, odo)
		require.Empty(t, result)
	})

	t.Run("filters range where odometer increases beyond epsilon", func(t *testing.T) {
		ranges := []timeRange{{start: min(0), end: min(10)}}
		soc := []levelSample{{ts: min(0), value: 20}, {ts: min(10), value: 80}}
		odo := []levelSample{{ts: min(0), value: 1000}, {ts: min(10), value: 1002}} // 2km > 0.5 epsilon
		result := filterRangesBySocAndOdo(ranges, soc, odo)
		require.Empty(t, result)
	})

	t.Run("odometer within epsilon is kept", func(t *testing.T) {
		ranges := []timeRange{{start: min(0), end: min(10)}}
		soc := []levelSample{{ts: min(0), value: 20}, {ts: min(10), value: 80}}
		odo := []levelSample{{ts: min(0), value: 1000}, {ts: min(10), value: 1000.3}} // 0.3 < 0.5
		result := filterRangesBySocAndOdo(ranges, soc, odo)
		require.Len(t, result, 1)
	})

	t.Run("no odometer data keeps range", func(t *testing.T) {
		ranges := []timeRange{{start: min(0), end: min(10)}}
		soc := []levelSample{{ts: min(0), value: 20}, {ts: min(10), value: 80}}
		result := filterRangesBySocAndOdo(ranges, soc, nil)
		require.Len(t, result, 1)
	})
}

func TestLevelFirstLastInRange(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }

	t.Run("returns first and last in range", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 10},
			{ts: min(5), value: 50},
			{ts: min(10), value: 90},
		}
		first, last, ok := levelFirstLastInRange(samples, min(0), min(10))
		require.True(t, ok)
		require.Equal(t, 10.0, first)
		require.Equal(t, 90.0, last)
	})

	t.Run("no samples in range", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 10},
		}
		_, _, ok := levelFirstLastInRange(samples, min(5), min(10))
		require.False(t, ok)
	})

	t.Run("empty samples", func(t *testing.T) {
		_, _, ok := levelFirstLastInRange(nil, min(0), min(10))
		require.False(t, ok)
	})
}
