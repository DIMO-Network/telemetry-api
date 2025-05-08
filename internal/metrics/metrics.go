package metrics

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	existStatusFailure = "failure"
	exitStatusSuccess  = "success"
)

type metricsKey struct{}

var (
	requestStartedCounter    prometheus.Counter
	requestCompletedCounter  prometheus.Counter
	resolverStartedCounter   *prometheus.CounterVec
	resolverCompletedCounter *prometheus.CounterVec
	timeToResolveField       *prometheus.HistogramVec
	timeToHandleRequest      *prometheus.HistogramVec
	fieldsPerRequest         *prometheus.HistogramVec
)

// FieldCountRange categorizes requests by field count.
type FieldCountRange string

const (
	// FieldCountSmall represents requests with 1-5 fields.
	FieldCountSmall FieldCountRange = "small" // 1-5 fields
	// FieldCountMedium represents requests with 6-10 fields.
	FieldCountMedium FieldCountRange = "medium" // 6-10 fields
	// FieldCountLarge represents requests with 11-50 fields.
	FieldCountLarge FieldCountRange = "large" // 11-50 fields
	// FieldCountHuge represents requests with 51+ fields.
	FieldCountHuge FieldCountRange = "huge" // 51+ fields
)

// GetFieldCountRange returns a string representation of the field count range.
func GetFieldCountRange(count int) string {
	switch {
	case count <= 5:
		return string(FieldCountSmall)
	case count <= 10:
		return string(FieldCountMedium)
	case count <= 50:
		return string(FieldCountLarge)
	default:
		return string(FieldCountHuge)
	}
}

// Tracer provides a GraphQL middleware for collecting Prometheus metrics.
type (
	Tracer struct{}
)

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

	registerer.MustRegister(
		requestStartedCounter,
		requestCompletedCounter,
		resolverStartedCounter,
		resolverCompletedCounter,
		timeToResolveField,
		timeToHandleRequest,
		fieldsPerRequest,
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
	requestStartedCounter.Inc()

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
	errList := graphql.GetErrors(ctx)

	var exitStatus string
	if len(errList) > 0 {
		exitStatus = existStatusFailure
	} else {
		exitStatus = exitStatusSuccess
	}

	oc := graphql.GetOperationContext(ctx)
	observerStart := oc.Stats.OperationStart

	// Get request metrics from context.
	metrics, ok := ctx.Value(metricsKey{}).(*requestMetrics)
	fieldCountRange := "unknown"
	if ok {
		// Track the number of fields in this request.
		fieldsPerRequest.WithLabelValues(exitStatus).Observe(float64(metrics.fieldCount.Load()))
		fieldCountRange = GetFieldCountRange(int(metrics.fieldCount.Load()))
	}

	timeToHandleRequest.With(prometheus.Labels{
		"exitStatus":      exitStatus,
		"fieldCountRange": fieldCountRange,
	}).Observe(float64(time.Since(observerStart).Nanoseconds() / int64(time.Millisecond)))

	requestCompletedCounter.Inc()

	return next(ctx)
}

// InterceptField intercepts GraphQL field resolution to track metrics.
func (a Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
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
