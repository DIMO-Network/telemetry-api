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

	ValueNumbers map[AliasKey]float64 `json:"-"`
	ValueStrings map[AliasKey]string  `json:"-"`
}

// AggSignal holds the value of an aggregation for a signal in a certain
// time bucket. Only one of ValueNumber and ValueString contains a meaningful
// value, determined by Name.
type AggSignal struct {
	Name        string
	Agg         string
	Timestamp   time.Time
	ValueNumber float64
	ValueString string
}
