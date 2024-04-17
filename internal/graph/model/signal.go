package model

// SignalWithID is the collection of signals.
// This struct is used to force the generation of the SignalCollection, and SignalAggregations resolver.
type SignalsWithID struct {
	TokenID uint32 `json:"tokenID"`
}
