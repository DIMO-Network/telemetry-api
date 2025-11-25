package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	maxDateRangeDays = 30
)

// validateSegmentArgs validates the arguments for segment queries
func validateSegmentArgs(tokenID int, from, to time.Time) error {
	if tokenID <= 0 {
		return fmt.Errorf("invalid tokenID: %d", tokenID)
	}

	if from.After(to) {
		return fmt.Errorf("from time must be before to time")
	}

	if from.Equal(to) {
		return fmt.Errorf("from and to times cannot be equal")
	}

	if to.After(time.Now()) {
		return fmt.Errorf("to time cannot be in the future")
	}

	maxDuration := maxDateRangeDays * 24 * time.Hour
	if to.Sub(from) > maxDuration {
		return fmt.Errorf("date range exceeds maximum of %d days", maxDateRangeDays)
	}

	return nil
}

// validateSegmentConfig validates the segment configuration parameters
func validateSegmentConfig(config *model.SegmentConfig) error {
	if config == nil {
		return nil
	}

	if config.MinIdleSeconds != nil {
		if *config.MinIdleSeconds < 60 || *config.MinIdleSeconds > 3600 {
			return fmt.Errorf("minIdleSeconds must be between 60 and 3600")
		}
	}

	if config.MinSegmentDurationSeconds != nil {
		if *config.MinSegmentDurationSeconds < 1 || *config.MinSegmentDurationSeconds > 3600 {
			return fmt.Errorf("minSegmentDurationSeconds must be between 1 and 3600")
		}
	}

	if config.SignalCountThreshold != nil {
		if *config.SignalCountThreshold < 3 || *config.SignalCountThreshold > 100 {
			return fmt.Errorf("signalCountThreshold must be between 3 and 100")
		}
	}

	return nil
}

// GetSegments returns segments detected using the specified mechanism in the time range
func (r *Repository) GetSegments(ctx context.Context, tokenID int, from, to time.Time, mechanism model.DetectionMechanism, config *model.SegmentConfig) ([]*model.Segment, error) {
	// Validate inputs
	if err := validateSegmentArgs(tokenID, from, to); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}

	if err := validateSegmentConfig(config); err != nil {
		return nil, errorhandler.NewBadRequestError(ctx, err)
	}

	// Query from ClickHouse service
	chSegments, err := r.chService.GetSegments(ctx, uint32(tokenID), from, to, mechanism, config)
	if err != nil {
		return nil, handleDBError(ctx, err)
	}

	// Convert to GraphQL model
	segments := make([]*model.Segment, len(chSegments))
	for i, chSegment := range chSegments {
		segments[i] = &model.Segment{
			StartTime:          chSegment.StartTime,
			EndTime:            chSegment.EndTime,
			DurationSeconds:    int(chSegment.DurationSeconds),
			IsOngoing:          chSegment.IsOngoing,
			StartedBeforeRange: chSegment.StartedBeforeRange,
		}
	}

	return segments, nil
}
