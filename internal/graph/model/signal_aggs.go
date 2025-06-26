package model

import "time"

// AliasKey identifies the combination of signal name and aggregation
// for a certain time bucket.
type AliasKey struct {
	Name string
	Agg  string
}

// SignalAggregations is the Go struct corresponding to the GraphQL type
// SignalAggregations. Most of its subfields are signal aggregation results
// that are returned by resolvers, so they do not appear on the model.
type SignalAggregations struct {
	Timestamp time.Time `json:"timestamp"`

	// Alias to value
	ValueNumbers map[string]float64 `json:"-"`
	ValueStrings map[string]string  `json:"-"`

	SignalArgs *AggregatedSignalArgs `json:"-"`
}

type FieldInfo struct {
	Alias string
	Name  string
}

// AggSignal holds the value of an aggregation for a signal in a certain
// time bucket. Only one of ValueNumber and ValueString contains a meaningful
// value, determined by Name.
type AggSignal struct {
	Handle      string
	Timestamp   time.Time
	ValueNumber float64
	ValueString string
}
