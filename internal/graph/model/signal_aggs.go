package model

import "time"

type AliasKey struct {
	Name string
	Agg  string
}

type StringSignalMap map[AliasKey]string

type FloatSignalMap map[AliasKey]float64

type AggSignal struct {
	Name        string
	Agg         string
	Timestamp   time.Time
	ValueNumber float64
	ValueString string
}
