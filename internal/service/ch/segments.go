package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	maxDateRangeDays = 30
)

// GetSegments returns segments detected using the specified mechanism
func (s *Service) GetSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	mechanism model.DetectionMechanism,
	filter *model.SignalFilter,
	config *model.SegmentConfig,
) ([]*Segment, error) {
	// Validate date range
	days := int(to.Sub(from).Hours() / 24)
	if days > maxDateRangeDays {
		return nil, fmt.Errorf("date range cannot exceed %d days (from: %s, to: %s)", maxDateRangeDays, from.Format(time.RFC3339), to.Format(time.RFC3339))
	}

	// Get appropriate detector based on mechanism
	var detector SegmentDetector
	switch mechanism {
	case model.DetectionMechanismIgnitionDetection:
		detector = &IgnitionDetector{conn: s.conn}
	case model.DetectionMechanismFrequencyAnalysis:
		detector = &FrequencyDetector{conn: s.conn}
	case model.DetectionMechanismChangePointDetection:
		detector = &ChangePointDetector{conn: s.conn}
	default:
		return nil, fmt.Errorf("unknown detection mechanism: %s", mechanism)
	}

	// Detect segments using mechanism-specific logic
	segments, err := detector.DetectSegments(ctx, tokenID, from, to, filter, config)
	if err != nil {
		return nil, err
	}

	return segments, nil
}
