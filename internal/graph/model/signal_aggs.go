package model

import "time"

type SignalAggregations struct {
	Timestamp time.Time `json:"timestamp"`

	NumberValues map[AliasKey]float64
	StringValues map[AliasKey]string
}

type AliasKey struct {
	Name string
	Agg  string
}
