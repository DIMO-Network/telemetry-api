// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type FloatAggregation struct {
	// Aggregation type.
	Type FloatAggregationType `json:"type"`
	// interval is a time span that used for aggregatting the data with.
	// A duration string is a sequence of decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m". Valid time units are "ms", "s", "m", "h", "d", "w", "y".
	Interval string `json:"interval"`
}

// The root query type for the GraphQL schema.
type Query struct {
}

type SignalFloat struct {
	// timestamp of when this data was colllected
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// value of the signal
	Value *float64 `json:"value,omitempty"`
}

type SignalString struct {
	// timestamp of when this data was colllected
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// value of the signal
	Value *string `json:"value,omitempty"`
}

type StringAggregation struct {
	// Aggregation type.
	Type StringAggregationType `json:"type"`
	// interval is a time span that used for aggregatting the data with.
	// A duration string is a sequence of decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m". Valid time units are "ms", "s", "m", "h", "d", "w", "y".
	Interval string `json:"interval"`
}

type FloatAggregationType string

const (
	FloatAggregationTypeAvg  FloatAggregationType = "avg"
	FloatAggregationTypeMed  FloatAggregationType = "med"
	FloatAggregationTypeMax  FloatAggregationType = "max"
	FloatAggregationTypeMin  FloatAggregationType = "min"
	FloatAggregationTypeRand FloatAggregationType = "rand"
)

var AllFloatAggregationType = []FloatAggregationType{
	FloatAggregationTypeAvg,
	FloatAggregationTypeMed,
	FloatAggregationTypeMax,
	FloatAggregationTypeMin,
	FloatAggregationTypeRand,
}

func (e FloatAggregationType) IsValid() bool {
	switch e {
	case FloatAggregationTypeAvg, FloatAggregationTypeMed, FloatAggregationTypeMax, FloatAggregationTypeMin, FloatAggregationTypeRand:
		return true
	}
	return false
}

func (e FloatAggregationType) String() string {
	return string(e)
}

func (e *FloatAggregationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FloatAggregationType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FloatAggregationType", str)
	}
	return nil
}

func (e FloatAggregationType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Privilege string

const (
	PrivilegeVehicleNonLocationData Privilege = "VehicleNonLocationData"
	PrivilegeVehicleCommands        Privilege = "VehicleCommands"
	PrivilegeVehicleCurrentLocation Privilege = "VehicleCurrentLocation"
	PrivilegeVehicleAllTimeLocation Privilege = "VehicleAllTimeLocation"
	PrivilegeVehicleVinCredential   Privilege = "VehicleVinCredential"
)

var AllPrivilege = []Privilege{
	PrivilegeVehicleNonLocationData,
	PrivilegeVehicleCommands,
	PrivilegeVehicleCurrentLocation,
	PrivilegeVehicleAllTimeLocation,
	PrivilegeVehicleVinCredential,
}

func (e Privilege) IsValid() bool {
	switch e {
	case PrivilegeVehicleNonLocationData, PrivilegeVehicleCommands, PrivilegeVehicleCurrentLocation, PrivilegeVehicleAllTimeLocation, PrivilegeVehicleVinCredential:
		return true
	}
	return false
}

func (e Privilege) String() string {
	return string(e)
}

func (e *Privilege) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Privilege(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Privilege", str)
	}
	return nil
}

func (e Privilege) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type StringAggregationType string

const (
	// Randomly select a value from the group.
	StringAggregationTypeRand StringAggregationType = "rand"
	// Select the most frequently occurring value in the group.
	StringAggregationTypeTop StringAggregationType = "top"
	// Return a list of unique values in the group.
	StringAggregationTypeUnique StringAggregationType = "unique"
)

var AllStringAggregationType = []StringAggregationType{
	StringAggregationTypeRand,
	StringAggregationTypeTop,
	StringAggregationTypeUnique,
}

func (e StringAggregationType) IsValid() bool {
	switch e {
	case StringAggregationTypeRand, StringAggregationTypeTop, StringAggregationTypeUnique:
		return true
	}
	return false
}

func (e StringAggregationType) String() string {
	return string(e)
}

func (e *StringAggregationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = StringAggregationType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid StringAggregationType", str)
	}
	return nil
}

func (e StringAggregationType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
