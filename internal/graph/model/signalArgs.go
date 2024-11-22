package model

import (
	"time"
)

const (
	// LastSeenField is the field name for the last seen timestamp.
	LastSeenField = "lastSeen"
	// ApproximateLongField is the field name for the approximate longitude.
	ApproximateLongField = "currentLocationApproximateLongitude"
	// ApproximateLatField is the field name for the approximate latitude.
	ApproximateLatField = "currentLocationApproximateLatitude"
)

// SignalArgs is the base arguments for querying signals.
type SignalArgs struct {
	// Filter  optional filter for the signals.
	Filter *SignalFilter
	// TokenID is the vehicles's NFT token ID.
	TokenID uint32
}

// LatestSignalsArgs is the arguments for querying the latest signals.
type LatestSignalsArgs struct {
	SignalArgs
	// SignalNames is the list of signal names to query.
	SignalNames map[string]struct{}
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
	// Interval in which the data is aggregated in milliseconds.
	Interval int64
	// FloatArgs represents arguments for each float signal.
	FloatArgs map[FloatSignalArgs]struct{}
	// StringArgs represents arguments for each string signal.
	StringArgs map[StringSignalArgs]struct{}
}

// FloatSignalArgs is the arguments for querying a float signals.
type FloatSignalArgs struct {
	// Name is the signal name.
	Name string
	// Agg is the aggregation type.
	Agg FloatAggregation
}

// StringSignalArgs is the arguments for querying a string signals.
type StringSignalArgs struct {
	// Name is the signal name.
	Name string
	// Agg is the aggregation type.
	Agg StringAggregation
}
