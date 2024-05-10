package model

import (
	"time"
)

// SignalWithID is the collection of signals.
// This struct is used to force the generation of the SignalCollection, and SignalAggregations resolver.
type SignalsWithID struct {
	TokenID uint32 `json:"tokenID"`
	SigArgs SignalArgs
}

// SignalArgs is the base arguments for querying signals.
type SignalArgs struct {
	FromTS  time.Time
	ToTS    time.Time
	Filter  *SignalFilter
	Name    string
	TokenID uint32
}

// FloatSignalArgs is the arguments for querying a float signals.
type FloatSignalArgs struct {
	Agg FloatAggregation
	SignalArgs
}

// StringSignalArgs is the arguments for querying a string signals.
type StringSignalArgs struct {
	Agg StringAggregation
	SignalArgs
}
