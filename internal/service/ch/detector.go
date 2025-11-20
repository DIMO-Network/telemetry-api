package ch

import (
	"context"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// Segment represents a detected vehicle usage segment
type Segment struct {
	TokenID            uint32
	SegmentID          string
	StartTime          time.Time
	EndTime            *time.Time
	DurationSeconds    int32
	IsOngoing          bool
	StartedBeforeRange bool
}

// SegmentDetector defines the interface for different segment detection mechanisms
type SegmentDetector interface {
	// DetectSegments identifies vehicle usage segments using mechanism-specific logic
	DetectSegments(
		ctx context.Context,
		tokenID uint32,
		from, to time.Time,
		config *model.SegmentConfig,
	) ([]*Segment, error)

	// GetMechanismName returns the name of this detection mechanism
	GetMechanismName() string
}
