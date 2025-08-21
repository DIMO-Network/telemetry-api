package model

import (
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
)

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
	// Alias to value
	ValueStrings map[string]string `json:"-"`

	// Aggregation cross non-approximate field name to value
	AppLocNumbers map[AppLocKey]float64 `json:"-"`

	// Alias to value
	ValueLocations map[string]vss.Location `json:"-"`
}

type AppLocKey struct {
	Aggregation FloatAggregation
	Name        string
}
