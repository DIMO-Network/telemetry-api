package model

import (
	"time"
)

const (
	// LastSeenField is the field name for the last seen timestamp.
	LastSeenField = "lastSeen"
	// ApproximateCoordinatesField is the field name for the approximate current location.
	// This is treated specially because there is no underlying ClickHouse table row carrying
	// this name.
	ApproximateCoordinatesField = "currentLocationApproximateCoordinates"
)

// SignalArgs is the base arguments for querying signals.
type SignalArgs struct {
	// Filter is an optional filter for the signals.
	Filter *SignalFilter
	// TokenID is the vehicle's NFT token ID.
	TokenID uint32
}

// LatestSignalsArgs is the arguments for querying the latest signals.
//
// We don't need to store as much information as we do for aggregations, since
// the lack of aggregation and filtering means that each signal will appear at
// most once in the resulting database query.
type LatestSignalsArgs struct {
	SignalArgs
	// SignalNames is the list of signal names to query.
	SignalNames map[string]struct{}
	// LocationSignalNames is the list of location signal names to query.
	LocationSignalNames map[string]struct{}
	// IncludeLastSeen is a flag to include a new signal for the last seen signal.
	IncludeLastSeen bool
}

// AggregatedSignalArgs is the arguments for querying aggregated signals.
type AggregatedSignalArgs struct {
	SignalArgs
	// FromTS is the start timestamp for the data range.
	FromTS time.Time
	// ToTS is the end timestamp for the data range.
	ToTS time.Time
	// Interval in which the data is aggregated in microseconds.
	Interval int64
	// FloatArgs represents arguments for each float signal.
	FloatArgs []FloatSignalArgs
	// StringArgs represents arguments for each string signal.
	StringArgs []StringSignalArgs
	// LocationArgs represents arguments for each location signal.
	LocationArgs []LocationSignalArgs
}

type LocationSignalArgs struct {
	// Name is the VSS name for the location field. This is the signal name in the database.
	//
	// This will match the GraphQL field name except in the case of approximate location.
	Name string
	Agg  LocationAggregation
	// Alias is the name used in the GraphQL query.
	//
	// Often this will be the same as Name, but these will necessarily differ for
	// queries in which the same signal is requested with different aggregations
	// and filters.
	Alias string
	// Filter optionally constrains the values fed into the aggregation. This will always
	// be null for approximate location queries.
	Filter *SignalLocationFilter
}

// FloatSignalArgs is the arguments for querying a float signals.
type FloatSignalArgs struct {
	// Name is the signal name. This is the field name in the API.
	Name string
	// Agg is the aggregation type.
	Agg FloatAggregation
	// Alias is the GraphQL field alias. If the client doesn't specify
	// an alias then this will be the same as Name.
	Alias string
	// Filter is an optional set of float value filters.
	Filter *SignalFloatFilter
}

// StringSignalArgs is the arguments for querying a string signals.
type StringSignalArgs struct {
	// Name is the signal name.
	Name string
	// Agg is the aggregation type.
	Agg StringAggregation
	// Alias is the GraphQL field alias. If the client doesn't specify
	// an alias then this will be the same as Name.
	Alias string
}
