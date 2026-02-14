package ch

import (
	"testing"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/require"
)

func TestFindIdleRpmRanges(t *testing.T) {
	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	min := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }
	from := base
	to := base.Add(time.Hour)

	maxIdleRpm := 1000
	maxGap := 300   // 5 minutes
	minDuration := 240 // 4 minutes

	t.Run("empty samples returns nil", func(t *testing.T) {
		result := findIdleRpmRanges(nil, maxIdleRpm, maxGap, minDuration, from, to)
		require.Nil(t, result)
	})

	t.Run("single contiguous idle run", func(t *testing.T) {
		// RPM at 700 for 10 minutes
		var samples []levelSample
		for i := 0; i <= 10; i++ {
			samples = append(samples, levelSample{ts: min(i), value: 700})
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Len(t, result, 1)
		require.Equal(t, min(0), result[0].start)
		require.Equal(t, min(10), result[0].end)
	})

	t.Run("RPM exactly at maxIdleRpm boundary is idle", func(t *testing.T) {
		var samples []levelSample
		for i := 0; i <= 10; i++ {
			samples = append(samples, levelSample{ts: min(i), value: 1000}) // exactly at boundary
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Len(t, result, 1)
	})

	t.Run("RPM just above maxIdleRpm is not idle", func(t *testing.T) {
		var samples []levelSample
		for i := 0; i <= 10; i++ {
			samples = append(samples, levelSample{ts: min(i), value: 1001})
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Empty(t, result)
	})

	t.Run("RPM at 0 is not idle", func(t *testing.T) {
		// RPM=0 means engine off, not idling
		var samples []levelSample
		for i := 0; i <= 10; i++ {
			samples = append(samples, levelSample{ts: min(i), value: 0})
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Empty(t, result)
	})

	t.Run("gap larger than maxGap splits segments", func(t *testing.T) {
		samples := []levelSample{
			// Run 1: min(0) to min(5)
			{ts: min(0), value: 700},
			{ts: min(1), value: 700},
			{ts: min(2), value: 700},
			{ts: min(3), value: 700},
			{ts: min(4), value: 700},
			{ts: min(5), value: 700},
			// Gap of 6 minutes (> 5 min maxGap)
			// Run 2: min(11) to min(16)
			{ts: min(11), value: 700},
			{ts: min(12), value: 700},
			{ts: min(13), value: 700},
			{ts: min(14), value: 700},
			{ts: min(15), value: 700},
			{ts: min(16), value: 700},
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Len(t, result, 2)
	})

	t.Run("gap within maxGap keeps single segment", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 700},
			{ts: min(1), value: 700},
			{ts: min(2), value: 700},
			// Gap of 4 minutes (< 5 min maxGap)
			{ts: min(6), value: 700},
			{ts: min(7), value: 700},
			{ts: min(8), value: 700},
			{ts: min(9), value: 700},
			{ts: min(10), value: 700},
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Len(t, result, 1)
		require.Equal(t, min(0), result[0].start)
		require.Equal(t, min(10), result[0].end)
	})

	t.Run("short segment filtered by minDuration", func(t *testing.T) {
		// Only 3 minutes (180s < 240s minDuration)
		samples := []levelSample{
			{ts: min(0), value: 700},
			{ts: min(1), value: 700},
			{ts: min(2), value: 700},
			{ts: min(3), value: 700},
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Empty(t, result)
	})

	t.Run("time range clipping clips start before from", func(t *testing.T) {
		clipFrom := min(3)
		var samples []levelSample
		for i := 0; i <= 10; i++ {
			samples = append(samples, levelSample{ts: min(i), value: 700})
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, clipFrom, to)
		require.Len(t, result, 1)
		require.Equal(t, clipFrom, result[0].start) // clipped to from
	})

	t.Run("time range clipping clips end after to", func(t *testing.T) {
		clipTo := min(7)
		var samples []levelSample
		for i := 0; i <= 10; i++ {
			samples = append(samples, levelSample{ts: min(i), value: 700})
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, clipTo)
		require.Len(t, result, 1)
		require.Equal(t, clipTo, result[0].end) // clipped to to
	})

	t.Run("non-idle samples break segment", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 700},
			{ts: min(1), value: 700},
			{ts: min(2), value: 700},
			{ts: min(3), value: 700},
			{ts: min(4), value: 700},
			{ts: min(5), value: 3000}, // not idle
			{ts: min(6), value: 700},
			{ts: min(7), value: 700},
			{ts: min(8), value: 700},
			{ts: min(9), value: 700},
			{ts: min(10), value: 700},
			{ts: min(11), value: 700},
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		// First run: min(0)-min(4) = 4 min, second run: min(6)-min(11) = 5 min
		require.Len(t, result, 2)
	})

	t.Run("mixed idle and high RPM", func(t *testing.T) {
		samples := []levelSample{
			{ts: min(0), value: 700},
			{ts: min(5), value: 700},
			{ts: min(6), value: 5000},  // driving
			{ts: min(10), value: 5000}, // driving
			{ts: min(15), value: 700},
			{ts: min(20), value: 700},
		}
		result := findIdleRpmRanges(samples, maxIdleRpm, maxGap, minDuration, from, to)
		require.Len(t, result, 2)
	})
}

func TestResolveBaseConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		rc := resolveBaseConfig(nil)
		require.Equal(t, defaultMaxGapSeconds, rc.maxGapSeconds)
		require.Equal(t, defaultMinSegmentDurationSeconds, rc.minDuration)
	})

	t.Run("config overrides applied", func(t *testing.T) {
		gap := 120
		dur := 60
		rc := resolveBaseConfig(&model.SegmentConfig{
			MaxGapSeconds:             &gap,
			MinSegmentDurationSeconds: &dur,
		})
		require.Equal(t, 120, rc.maxGapSeconds)
		require.Equal(t, 60, rc.minDuration)
	})
}
