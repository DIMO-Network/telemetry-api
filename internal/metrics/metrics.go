package metrics

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

const (
	existStatusFailure = "failure"
	exitStatusSuccess  = "success"
)

// ResponseSizeRange categorizes responses by size in bytes.
type ResponseSizeRange string

const (
	// ResponseSizeTiny represents responses with 0-10KB.
	ResponseSizeTiny ResponseSizeRange = "tiny"
	// ResponseSizeSmall represents responses with 10-100KB.
	ResponseSizeSmall ResponseSizeRange = "small"
	// ResponseSizeMedium represents responses with 100KB-1MB.
	ResponseSizeMedium ResponseSizeRange = "medium"
	// ResponseSizeLarge represents responses with 1-10MB.
	ResponseSizeLarge ResponseSizeRange = "large"
	// ResponseSizeHuge represents responses with 10MB-1GB.
	ResponseSizeHuge ResponseSizeRange = "huge"
)

// GetResponseSizeRange returns a string representation of the response size range.
func GetResponseSizeRange(size int) string {
	switch {
	case size <= 10*1024: // 10KB
		return string(ResponseSizeTiny)
	case size <= 100*1024: // 100KB
		return string(ResponseSizeSmall)
	case size <= 1024*1024: // 1MB
		return string(ResponseSizeMedium)
	case size <= 10*1024*1024: // 10MB
		return string(ResponseSizeLarge)
	case size <= 1024*1024*1024: // 1GB
		return string(ResponseSizeHuge)
	default:
		return string(ResponseSizeHuge) // Anything over 1GB is still considered huge
	}
}

type metricsKey struct{}

var (
	requestStartedCounter    prometheus.Counter
	requestCompletedCounter  prometheus.Counter
	resolverStartedCounter   *prometheus.CounterVec
	resolverCompletedCounter *prometheus.CounterVec
	timeToResolveField       *prometheus.HistogramVec
	timeToHandleRequest      *prometheus.HistogramVec
	fieldsPerRequest         *prometheus.HistogramVec
	// Field range specific counters
	requestStartedTinyCounter   prometheus.Counter
	requestStartedSmallCounter  prometheus.Counter
	requestStartedMediumCounter prometheus.Counter
	requestStartedLargeCounter  prometheus.Counter
	requestStartedHugeCounter   prometheus.Counter
	// Response size specific counters
	responseSizeTinyCounter   prometheus.Counter
	responseSizeSmallCounter  prometheus.Counter
	responseSizeMediumCounter prometheus.Counter
	responseSizeLargeCounter  prometheus.Counter
	responseSizeHugeCounter   prometheus.Counter
)

// FieldCountRange categorizes requests by field count.
type FieldCountRange string

const (
	// FieldCountTiny represents requests with 0-5 fields.
	FieldCountTiny FieldCountRange = "tiny"
	// FieldCountSmall represents requests with 6-10 fields.
	FieldCountSmall FieldCountRange = "small"
	// FieldCountMedium represents requests with 11-20 fields.
	FieldCountMedium FieldCountRange = "medium"
	// FieldCountLarge represents requests with 21-40 fields.
	FieldCountLarge FieldCountRange = "large"
	// FieldCountHuge represents requests with 41+ fields.
	FieldCountHuge FieldCountRange = "huge"
)

// GetFieldCountRange returns a string representation of the field count range.
func GetFieldCountRange(count int) string {
	switch {
	case count <= 5:
		return string(FieldCountTiny)
	case count <= 10:
		return string(FieldCountSmall)
	case count <= 20:
		return string(FieldCountMedium)
	case count <= 40:
		return string(FieldCountLarge)
	default:
		return string(FieldCountHuge)
	}
}

// Tracer provides a GraphQL middleware for collecting Prometheus metrics.
type Tracer struct {
	maxResponseSize atomic.Int64
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = Tracer{}

// Register registers metrics with the default Prometheus registerer.
func Register() {
	RegisterOn(prometheus.DefaultRegisterer)
}

// RegisterOn registers metrics with the provided Prometheus registerer.
func RegisterOn(registerer prometheus.Registerer) {
	requestStartedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_total",
			Help: "Total number of requests started on the graphql server.",
		},
	)

	requestStartedTinyCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_tiny_total",
			Help: "Total number of tiny requests (0-5 fields) started on the graphql server.",
		},
	)

	requestStartedSmallCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_small_total",
			Help: "Total number of small requests (6-10 fields) started on the graphql server.",
		},
	)

	requestStartedMediumCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_medium_total",
			Help: "Total number of medium requests (11-20 fields) started on the graphql server.",
		},
	)

	requestStartedLargeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_large_total",
			Help: "Total number of large requests (21-40 fields) started on the graphql server.",
		},
	)

	requestStartedHugeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_started_huge_total",
			Help: "Total number of huge requests (41+ fields) started on the graphql server.",
		},
	)

	requestCompletedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_request_completed_total",
			Help: "Total number of requests completed on the graphql server.",
		},
	)

	resolverStartedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "graphql_resolver_started_total",
			Help: "Total number of resolver started on the graphql server.",
		},
		[]string{"object", "field"},
	)

	resolverCompletedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "graphql_resolver_completed_total",
			Help: "Total number of resolver completed on the graphql server.",
		},
		[]string{"object", "field"},
	)

	timeToResolveField = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "graphql_resolver_duration_ms",
		Help:    "The time taken to resolve a field by graphql server.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 11),
	}, []string{"exitStatus", "object", "field"})

	timeToHandleRequest = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "graphql_request_duration_ms",
		Help:    "The time taken to handle a request by graphql server.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 11),
	}, []string{"exitStatus", "fieldCountRange"})

	fieldsPerRequest = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "graphql_fields_per_request",
		Help:    "The number of fields resolved in a request.",
		Buckets: prometheus.LinearBuckets(1, 5, 60), // From 1 to 300 fields in buckets of 5
	}, []string{"exitStatus"})

	responseSizeTinyCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_response_size_tiny_total",
			Help: "Total number of tiny responses (0-10KB) completed on the graphql server.",
		},
	)

	responseSizeSmallCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_response_size_small_total",
			Help: "Total number of small responses (10-100KB) completed on the graphql server.",
		},
	)

	responseSizeMediumCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_response_size_medium_total",
			Help: "Total number of medium responses (100KB-1MB) completed on the graphql server.",
		},
	)

	responseSizeLargeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_response_size_large_total",
			Help: "Total number of large responses (1-10MB) completed on the graphql server.",
		},
	)

	responseSizeHugeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "graphql_response_size_huge_total",
			Help: "Total number of huge responses (10MB-1GB) completed on the graphql server.",
		},
	)

	registerer.MustRegister(
		requestStartedCounter,
		requestCompletedCounter,
		resolverStartedCounter,
		resolverCompletedCounter,
		timeToResolveField,
		timeToHandleRequest,
		fieldsPerRequest,
		requestStartedTinyCounter,
		requestStartedSmallCounter,
		requestStartedMediumCounter,
		requestStartedLargeCounter,
		requestStartedHugeCounter,
		responseSizeTinyCounter,
		responseSizeSmallCounter,
		responseSizeMediumCounter,
		responseSizeLargeCounter,
		responseSizeHugeCounter,
	)
}

// UnRegister unregisters all metrics from the default Prometheus registerer.
func UnRegister() {
	UnRegisterFrom(prometheus.DefaultRegisterer)
}

// UnRegisterFrom unregisters all metrics from the provided Prometheus registerer.
func UnRegisterFrom(registerer prometheus.Registerer) {
	registerer.Unregister(requestStartedCounter)
	registerer.Unregister(requestCompletedCounter)
	registerer.Unregister(resolverStartedCounter)
	registerer.Unregister(resolverCompletedCounter)
	registerer.Unregister(timeToResolveField)
	registerer.Unregister(timeToHandleRequest)
	registerer.Unregister(fieldsPerRequest)
	registerer.Unregister(requestStartedTinyCounter)
	registerer.Unregister(requestStartedSmallCounter)
	registerer.Unregister(requestStartedMediumCounter)
	registerer.Unregister(requestStartedLargeCounter)
	registerer.Unregister(requestStartedHugeCounter)
	registerer.Unregister(responseSizeTinyCounter)
	registerer.Unregister(responseSizeSmallCounter)
	registerer.Unregister(responseSizeMediumCounter)
	registerer.Unregister(responseSizeLargeCounter)
	registerer.Unregister(responseSizeHugeCounter)
}

// ExtensionName returns the name of this extension.
func (a Tracer) ExtensionName() string {
	return "Prometheus"
}

// Validate validates the GraphQL schema.
func (a Tracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

type requestMetrics struct {
	fieldCount atomic.Int64
}

// InterceptOperation intercepts GraphQL operations to track metrics.
func (a Tracer) InterceptOperation(
	ctx context.Context,
	next graphql.OperationHandler,
) graphql.ResponseHandler {
	complexity := extension.GetComplexityStats(ctx)
	if complexity != nil {
		switch GetFieldCountRange(complexity.Complexity) {
		case string(FieldCountTiny):
			requestStartedTinyCounter.Inc()
		case string(FieldCountSmall):
			requestStartedSmallCounter.Inc()
		case string(FieldCountMedium):
			requestStartedMediumCounter.Inc()
		case string(FieldCountLarge):
			requestStartedLargeCounter.Inc()
		case string(FieldCountHuge):
			requestStartedHugeCounter.Inc()
		}
	}
	requestStartedCounter.Inc()
	return next(ctx)

	// Record initial resource usage for this request.
	metrics := &requestMetrics{
		fieldCount: atomic.Int64{},
	}

	// Store metrics in context.
	metricsCtx := context.WithValue(ctx, metricsKey{}, metrics)

	return next(metricsCtx)
}

// InterceptResponse intercepts GraphQL responses to record metrics.
func (a Tracer) InterceptResponse(
	ctx context.Context,
	next graphql.ResponseHandler,
) *graphql.Response {
	response := next(ctx)
	// errList := graphql.GetErrors(ctx)

	// var exitStatus string
	// if len(errList) > 0 {
	// 	exitStatus = existStatusFailure
	// } else {
	// 	exitStatus = exitStatusSuccess
	// }

	// oc := graphql.GetOperationContext(ctx)
	// observerStart := oc.Stats.OperationStart

	// // Get request metrics from context.
	// metrics, ok := ctx.Value(metricsKey{}).(*requestMetrics)
	// fieldCountRange := "unknown"
	// if ok {
	// 	// Track the number of fields in this request.
	// 	fieldsPerRequest.WithLabelValues(exitStatus).Observe(float64(metrics.fieldCount.Load()))
	// 	fieldCountRange = GetFieldCountRange(int(metrics.fieldCount.Load()))
	// }

	// Calculate response size and increment appropriate counter
	if response != nil && response.Data != nil {
		responseSize := len(response.Data)
		if int64(responseSize) > a.maxResponseSize.Load() {
			logger := zerolog.Ctx(ctx)
			logger.Info().
				Int("previous_max_bytes", int(a.maxResponseSize.Load())).
				Int("new_max_bytes", responseSize).
				Msg("New maximum response size recorded")
			a.maxResponseSize.Store(int64(responseSize))
		}
		switch GetResponseSizeRange(responseSize) {
		case string(ResponseSizeTiny):
			responseSizeTinyCounter.Inc()
		case string(ResponseSizeSmall):
			responseSizeSmallCounter.Inc()
		case string(ResponseSizeMedium):
			responseSizeMediumCounter.Inc()
		case string(ResponseSizeLarge):
			responseSizeLargeCounter.Inc()
		case string(ResponseSizeHuge):
			responseSizeHugeCounter.Inc()
		}
	}

	// timeToHandleRequest.With(prometheus.Labels{
	// 	"exitStatus":      exitStatus,
	// 	"fieldCountRange": fieldCountRange,
	// }).Observe(float64(time.Since(observerStart).Nanoseconds() / int64(time.Millisecond)))

	requestCompletedCounter.Inc()

	return response
}

// InterceptField intercepts GraphQL field resolution to track metrics.
func (a Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
	return next(ctx)
	fc := graphql.GetFieldContext(ctx)

	resolverStartedCounter.WithLabelValues(fc.Object, fc.Field.Name).Inc()

	// Increment field count in context.
	if metrics, ok := ctx.Value(metricsKey{}).(*requestMetrics); ok {
		metrics.fieldCount.Add(1)
	}

	observerStart := time.Now()

	res, err := next(ctx)

	var exitStatus string
	if err != nil {
		exitStatus = existStatusFailure
	} else {
		exitStatus = exitStatusSuccess
	}

	timeToResolveField.WithLabelValues(exitStatus, fc.Object, fc.Field.Name).
		Observe(float64(time.Since(observerStart).Nanoseconds() / int64(time.Millisecond)))

	resolverCompletedCounter.WithLabelValues(fc.Object, fc.Field.Name).Inc()

	return res, err
}
