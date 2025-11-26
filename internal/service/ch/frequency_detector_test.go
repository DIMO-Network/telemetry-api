package ch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergeWindowsIntoSegments(t *testing.T) {
	// We can't inject time.Now() anymore, so we rely on relative times from time.Now() for tests that depend on "now"
	now := time.Now()
	tokenID := uint32(1)
	maxGap := 300 // 5 minutes
	minDuration := 60

	t.Run("normal segment - historical", func(t *testing.T) {
		// Use time far in the past so it doesn't trigger "near real-time" logic
		historicalNow := now.Add(-24 * time.Hour)
		from := historicalNow.Add(-time.Hour)
		to := historicalNow

		// Windows from hour ago to 30 mins ago (relative to historicalNow)
		windows := []ActiveWindow{
			{WindowStart: from, WindowEnd: from.Add(10 * time.Minute)},
			{WindowStart: from.Add(10 * time.Minute), WindowEnd: from.Add(30 * time.Minute)},
		}

		segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
		require.Len(t, segments, 1)
		require.False(t, segments[0].IsOngoing)
		require.NotNil(t, segments[0].EndTime)
		require.Equal(t, from.Add(30*time.Minute), *segments[0].EndTime)
	})

	t.Run("ongoing segment - hits to time", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		// Windows from hour ago to now
		windows := []ActiveWindow{
			{WindowStart: from, WindowEnd: now},
		}

		segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
		require.Len(t, segments, 1)
		require.True(t, segments[0].IsOngoing)
		require.Nil(t, segments[0].EndTime)
	})

	t.Run("ongoing segment - near real-time logic", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		// Window ends 1 minute ago (within 5 min gap)
		lastWindowEnd := now.Add(-time.Minute)
		windows := []ActiveWindow{
			{WindowStart: from, WindowEnd: lastWindowEnd},
		}

		segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
		require.Len(t, segments, 1)
		require.True(t, segments[0].IsOngoing)
		require.Nil(t, segments[0].EndTime)
		// Duration should be from start to 'to' (now)
		// Allow small delta for time precision
		expectedDuration := int32(to.Sub(from).Seconds())
		require.InDelta(t, expectedDuration, segments[0].DurationSeconds, 1)
	})

	t.Run("completed segment - outside real-time gap", func(t *testing.T) {
		from := now.Add(-time.Hour)
		to := now

		// Window ends 10 minutes ago (outside 5 min gap)
		lastWindowEnd := now.Add(-10 * time.Minute)
		windows := []ActiveWindow{
			{WindowStart: from, WindowEnd: lastWindowEnd},
		}

		segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
		require.Len(t, segments, 1)
		require.False(t, segments[0].IsOngoing)
		require.NotNil(t, segments[0].EndTime)
		require.Equal(t, lastWindowEnd, *segments[0].EndTime)
	})

	t.Run("historical query - not ongoing even if gap small relative to query time", func(t *testing.T) {
		// Query for yesterday
		queryTo := now.Add(-24 * time.Hour)
		from := queryTo.Add(-time.Hour)

		// Window ends 1 minute before queryTo
		lastWindowEnd := queryTo.Add(-time.Minute)
		windows := []ActiveWindow{
			{WindowStart: from, WindowEnd: lastWindowEnd},
		}

		segments := mergeWindowsIntoSegments(tokenID, windows, from, queryTo, maxGap, minDuration)
		require.Len(t, segments, 1)
		require.False(t, segments[0].IsOngoing)
		require.NotNil(t, segments[0].EndTime)
	})
}
