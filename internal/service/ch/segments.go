package ch

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/prometheus/client_golang/prometheus"
)

// newDetector returns a SegmentDetector for the given mechanism.
// Detectors are stateless beyond holding a conn, so constructing per-call is cheap.
func newDetector(conn clickhouse.Conn, mechanism model.DetectionMechanism) (SegmentDetector, error) {
	switch mechanism {
	case model.DetectionMechanismIgnitionDetection:
		return NewIgnitionDetector(conn), nil
	case model.DetectionMechanismFrequencyAnalysis:
		return NewFrequencyDetector(conn), nil
	case model.DetectionMechanismChangePointDetection:
		return NewChangePointDetector(conn), nil
	case model.DetectionMechanismIdling:
		return NewIdlingDetector(conn), nil
	case model.DetectionMechanismRefuel:
		return NewRefuelDetector(conn), nil
	case model.DetectionMechanismRecharge:
		return NewRechargeDetector(conn), nil
	default:
		return nil, fmt.Errorf("unknown detection mechanism: %s", mechanism)
	}
}

// GetSegments returns segments detected using the specified mechanism.
func (s *Service) GetSegments(
	ctx context.Context,
	tokenID uint32,
	from, to time.Time,
	mechanism model.DetectionMechanism,
	config *model.SegmentConfig,
) ([]*model.Segment, error) {
	detector, err := newDetector(s.conn, mechanism)
	if err != nil {
		return nil, err
	}

	timer := prometheus.NewTimer(GetSegmentsLatency.WithLabelValues(mechanism.String()))
	segments, err := detector.DetectSegments(ctx, tokenID, from, to, config)
	timer.ObserveDuration()
	if err != nil {
		return nil, err
	}
	return segments, nil
}
