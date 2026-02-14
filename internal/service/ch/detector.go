package ch

import (
	"context"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultMaxGapSeconds             = 300 // 5 minutes — max gap between samples/windows before splitting
	defaultMinSegmentDurationSeconds = 240 // 4 minutes — minimum segment duration to keep
)

// segmentEndTime returns the end timestamp of a segment, or zero time if End is nil.
func segmentEndTime(seg *model.Segment) time.Time {
	if seg != nil && seg.End != nil {
		return seg.End.Timestamp
	}
	return time.Time{}
}

// newSegment builds a model.Segment with only time bounds set (Signals/EventCounts/Start.Value/End.Value filled by repo).
func newSegment(startTime time.Time, endTime *time.Time, durationSec int32, isOngoing, startedBeforeRange bool) *model.Segment {
	start := &model.SignalLocation{Timestamp: startTime, Value: nil}
	var end *model.SignalLocation
	if endTime != nil {
		end = &model.SignalLocation{Timestamp: *endTime, Value: nil}
	}
	return &model.Segment{
		Start:              start,
		End:                end,
		Duration:           int(durationSec),
		IsOngoing:          isOngoing,
		StartedBeforeRange: startedBeforeRange,
	}
}

// resolvedConfig holds the resolved (default + override) config values shared across detectors.
type resolvedConfig struct {
	maxGapSeconds int // max gap between samples/windows before splitting segments (seconds)
	minDuration   int // minimum segment duration to keep (seconds)
}

// resolveBaseConfig applies config overrides on top of defaults for maxGapSeconds and minDuration.
func resolveBaseConfig(config *model.SegmentConfig) resolvedConfig {
	rc := resolvedConfig{
		maxGapSeconds: defaultMaxGapSeconds,
		minDuration:   defaultMinSegmentDurationSeconds,
	}
	if config != nil {
		if config.MaxGapSeconds != nil {
			rc.maxGapSeconds = *config.MaxGapSeconds
		}
		if config.MinSegmentDurationSeconds != nil {
			rc.minDuration = *config.MinSegmentDurationSeconds
		}
	}
	return rc
}

// SegmentDetector defines the interface for different segment detection mechanisms.
type SegmentDetector interface {
	// DetectSegments identifies vehicle usage segments using mechanism-specific logic
	DetectSegments(
		ctx context.Context,
		tokenID uint32,
		from, to time.Time,
		config *model.SegmentConfig,
	) ([]*model.Segment, error)

	// GetMechanismName returns the name of this detection mechanism
	GetMechanismName() string
}
