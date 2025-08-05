package dtcmiddleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// DCTRequestLatency measures the latency of DCT requests
	DCTRequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telemetry_credit_tracker_grpc_request_latency_seconds",
			Help:    "Latency of Credit Tracker GRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)
