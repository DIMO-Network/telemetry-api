package ch

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// GetSegmentsLatency measures latency of segment detection by mechanism
	GetSegmentsLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telemetry_ch_get_segments_latency_seconds",
			Help:    "Latency of GetSegments (segment detection) in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"mechanism"},
	)

	// GetAggregatedSignalsForRangesLatency measures latency of batch signal aggregation for segment summaries
	GetAggregatedSignalsForRangesLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "telemetry_ch_get_aggregated_signals_for_ranges_latency_seconds",
			Help:    "Latency of GetAggregatedSignalsForRanges in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// GetEventCountsForRangesLatency measures latency of batch event counts for segment summaries
	GetEventCountsForRangesLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "telemetry_ch_get_event_counts_for_ranges_latency_seconds",
			Help:    "Latency of GetEventCountsForRanges in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)
