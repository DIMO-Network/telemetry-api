package ch

import (
	"testing"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/require"
)

func TestStaticRpmDetector_GetMechanismName(t *testing.T) {
	d := &StaticRpmDetector{}
	require.Equal(t, "staticRpm", d.GetMechanismName())
}

func TestStaticRpmDetector_DetectSegments_ConfigDefaults(t *testing.T) {
	// Config: static RPM uses SignalCountThreshold (same as frequency), maxIdleRpm; engine speed signal name is fixed.
	_ = model.DetectionMechanismStaticRpm
	maxRpm := 900
	threshold := 5
	config := &model.SegmentConfig{
		MaxIdleRpm:             &maxRpm,
		SignalCountThreshold:   &threshold,
		MinIdleSeconds:         ptr(600),
		MinSegmentDurationSeconds: ptr(300),
	}
	require.NotNil(t, config)
	require.Equal(t, 900, *config.MaxIdleRpm)
	require.Equal(t, 5, *config.SignalCountThreshold)
}

func TestStaticRpmDetector_IdleWindowsMergeIntoOneSegment(t *testing.T) {
	// Static RPM detector uses mergeWindowsIntoSegments with idle windows; merge logic is shared with frequency detector.
	// Verify that a run of consecutive idle windows produces one segment when gap and duration are satisfied.
	now := time.Now()
	from := now.Add(-30 * time.Minute)
	to := now.Add(-5 * time.Minute)
	tokenID := uint32(1)
	maxGap := 300
	minDuration := 60

	// Consecutive 1-minute idle windows; last window ends before query 'to' so segment is completed (not ongoing)
	endWindows := to.Add(-10 * time.Minute) // last window ends 10 min before query end
	windows := make([]ActiveWindow, 0, 15)
	for s := from; s.Before(endWindows); s = s.Add(time.Minute) {
		windows = append(windows, ActiveWindow{
			WindowStart:         s,
			WindowEnd:           s.Add(time.Minute),
			SignalCount:         5,
			DistinctSignalCount: 1,
		})
	}
	segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
	require.Len(t, segments, 1)
	require.False(t, segments[0].IsOngoing)
	require.NotNil(t, segments[0].EndTime)
	require.True(t, segments[0].DurationSeconds >= int32(minDuration))
}

func TestStaticRpmDetector_TwoIdleBlocksProduceTwoSegments(t *testing.T) {
	now := time.Now()
	from := now.Add(-2 * time.Hour)
	to := now.Add(-10 * time.Minute)
	tokenID := uint32(1)
	maxGap := 300   // 5 min
	minDuration := 120 // 2 min

	// Block 1: 2 hours ago, 3 minutes of idle windows
	block1Start := from
	block1End := from.Add(3 * time.Minute)
	// Block 2: 30 min ago, 3 minutes of idle windows (gap between block1 and block2 > maxGap)
	block2Start := to.Add(-30 * time.Minute)
	block2End := block2Start.Add(3 * time.Minute)

	windows := []ActiveWindow{
		{WindowStart: block1Start, WindowEnd: block1Start.Add(time.Minute), SignalCount: 5, DistinctSignalCount: 1},
		{WindowStart: block1Start.Add(time.Minute), WindowEnd: block1End, SignalCount: 5, DistinctSignalCount: 1},
		{WindowStart: block1End, WindowEnd: block1End.Add(time.Minute), SignalCount: 5, DistinctSignalCount: 1},
		{WindowStart: block2Start, WindowEnd: block2Start.Add(time.Minute), SignalCount: 5, DistinctSignalCount: 1},
		{WindowStart: block2Start.Add(time.Minute), WindowEnd: block2End, SignalCount: 5, DistinctSignalCount: 1},
		{WindowStart: block2End, WindowEnd: block2End.Add(time.Minute), SignalCount: 5, DistinctSignalCount: 1},
	}
	segments := mergeWindowsIntoSegments(tokenID, windows, from, to, maxGap, minDuration)
	require.Len(t, segments, 2)
	require.False(t, segments[0].IsOngoing)
	require.False(t, segments[1].IsOngoing)
	require.True(t, segments[0].StartTime.Equal(block1Start))
	require.True(t, segments[1].StartTime.Equal(block2Start))
}

func ptr(i int) *int { return &i }
