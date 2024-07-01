package model

import "time"

type AliasKey struct {
	Name string
	Agg  string
}

type SignalAggregations struct {
	Timestamp time.Time `json:"timestamp"`

	ValueNumbers map[AliasKey]float64 `json:"-"`
	ValueStrings map[AliasKey]string  `json:"-"`
}

type AggSignal struct {
	Name        string
	Agg         string
	Timestamp   time.Time
	ValueNumber float64
	ValueString string
}
