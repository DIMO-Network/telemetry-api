package model

import (
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
)

// SignalAggregations is the Go struct corresponding to the GraphQL type
// SignalAggregations. Most of its subfields are signal aggregation results
// that are returned by resolvers, so they do not appear on the model.
type SignalAggregations struct {
	Timestamp time.Time `json:"timestamp"`

	// Signals holds the {name, agg, value} entries computed from the request's
	// signalRequests argument. One entry per supplied request that produced a
	// value in this bucket and that the caller has permission to see.
	Signals []*SignalAggregationValue `json:"signals"`

	// Alias to value
	ValueNumbers map[string]float64 `json:"-"`
	// Alias to value
	ValueStrings map[string]string `json:"-"`
	// ValueLocations maps location field alias to location value.
	//
	// For approximate location, the value stored here is not yet obfuscated. It is
	// the responsibility of the resolver to obfuscate the location.
	ValueLocations map[string]vss.Location `json:"-"`
}
