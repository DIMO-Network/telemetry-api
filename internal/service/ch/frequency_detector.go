package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const (
	defaultWindowSizeSeconds            = 60 // 1 minute windows
	defaultSignalCountThreshold         = 10 // Minimum signals per window for activity
	defaultDistinctSignalCountThreshold = 2  // Minimum distinct signal types per window
)

// FrequencyDetector detects segments using frequency analysis of signal updates.
// Analyzes signal update patterns to identify vehicle activity periods.
type FrequencyDetector struct {
	conn clickhouse.Conn
}

// NewFrequencyDetector creates a new FrequencyDetector with the given connection.
func NewFrequencyDetector(conn clickhouse.Conn) *FrequencyDetector {
	return &FrequencyDetector{conn: conn}
}

// DetectSegments implements frequency-based segment detection
func (d *FrequencyDetector) DetectSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	config *model.SegmentConfig,
) ([]*model.Segment, error) {
	rc := resolveBaseConfig(config)
	maxGap := rc.maxGapSeconds
	minDuration := rc.minDuration
	signalThreshold := defaultSignalCountThreshold
	if config != nil && config.SignalCountThreshold != nil {
		signalThreshold = *config.SignalCountThreshold
	}

	// Look back maxGap seconds before 'from' to detect segments that started before the query range.
	lookbackFrom := from.Add(-time.Duration(maxGap) * time.Second)
	windows, err := getWindowedSignalCounts(ctx, d.conn, tokenID, lookbackFrom, to, defaultWindowSizeSeconds, signalThreshold, defaultDistinctSignalCountThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query active windows: %w", err)
	}

	if len(windows) == 0 {
		return []*model.Segment{}, nil
	}

	// Merge consecutive active windows into segments (in Go for flexibility)
	segments := mergeWindowsIntoSegments(windows, from, to, maxGap, minDuration)

	return segments, nil
}

// GetMechanismName returns the name of this detection mechanism.
func (d *FrequencyDetector) GetMechanismName() string {
	return "frequencyAnalysis"
}
