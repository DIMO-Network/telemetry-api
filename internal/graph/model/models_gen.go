// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type Pomvc struct {
	// vehicleTokenId is the token ID of the vehicle.
	VehicleTokenID *int `json:"vehicleTokenId,omitempty"`
	// recordedBy is the entity that recorded the VIN.
	RecordedBy *string `json:"recordedBy,omitempty"`
	// vehicleContractAddress is the address of the vehicle contract.
	VehicleContractAddress *string `json:"vehicleContractAddress,omitempty"`
	// validFrom is the time the VC is valid from.
	ValidFrom *time.Time `json:"validFrom,omitempty"`
	// rawVC is the raw VC JSON.
	RawVc string `json:"rawVC"`
}

// The root query type for the GraphQL schema.
type Query struct {
}

type SignalCollection struct {
	// The last time any signal was seen matching the filter.
	LastSeen *time.Time `json:"lastSeen,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelLeftTirePressure *SignalFloat `json:"chassisAxleRow1WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelRightTirePressure *SignalFloat `json:"chassisAxleRow1WheelRightTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelLeftTirePressure *SignalFloat `json:"chassisAxleRow2WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelRightTirePressure *SignalFloat `json:"chassisAxleRow2WheelRightTirePressure,omitempty"`
	// Current altitude relative to WGS 84 reference ellipsoid, as measured at the position of GNSS receiver antenna.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationAltitude *SignalFloat `json:"currentLocationAltitude,omitempty"`
	// Indicates if the latitude and longitude signals at the current timestamp have been redacted using a privacy zone.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationIsRedacted *SignalFloat `json:"currentLocationIsRedacted,omitempty"`
	// Current latitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLatitude *SignalFloat `json:"currentLocationLatitude,omitempty"`
	// Current longitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLongitude *SignalFloat `json:"currentLocationLongitude,omitempty"`
	// Horizontal dilution of precision of GPS
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketHDOP *SignalFloat `json:"dimoAftermarketHDOP,omitempty"`
	// Number of sync satellites for GPS
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketNSAT *SignalFloat `json:"dimoAftermarketNSAT,omitempty"`
	// Service Set Identifier for the wifi.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketSSID *SignalString `json:"dimoAftermarketSSID,omitempty"`
	// Indicate the current WPA state for the device's wifi
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketWPAState *SignalString `json:"dimoAftermarketWPAState,omitempty"`
	// Air temperature outside the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ExteriorAirTemperature *SignalFloat `json:"exteriorAirTemperature,omitempty"`
	// Current Voltage of the low voltage battery.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	LowVoltageBatteryCurrentVoltage *SignalFloat `json:"lowVoltageBatteryCurrentVoltage,omitempty"`
	// PID 33 - Barometric pressure
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDBarometricPressure *SignalFloat `json:"obdBarometricPressure,omitempty"`
	// PID 04 - Engine load in percent - 0 = no load, 100 = full load
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDEngineLoad *SignalFloat `json:"obdEngineLoad,omitempty"`
	// PID 0F - Intake temperature
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDIntakeTemp *SignalFloat `json:"obdIntakeTemp,omitempty"`
	// PID 1F - Engine run time
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDRunTime *SignalFloat `json:"obdRunTime,omitempty"`
	// Engine coolant temperature.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineECT *SignalFloat `json:"powertrainCombustionEngineECT,omitempty"`
	// Engine oil level.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineEngineOilLevel *SignalString `json:"powertrainCombustionEngineEngineOilLevel,omitempty"`
	// Engine oil level as a percentage.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineEngineOilRelativeLevel *SignalFloat `json:"powertrainCombustionEngineEngineOilRelativeLevel,omitempty"`
	// Grams of air drawn into engine per second.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineMAF *SignalFloat `json:"powertrainCombustionEngineMAF,omitempty"`
	// Engine speed measured as rotations per minute.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineSpeed *SignalFloat `json:"powertrainCombustionEngineSpeed,omitempty"`
	// Current throttle position.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineTPS *SignalFloat `json:"powertrainCombustionEngineTPS,omitempty"`
	// Level in fuel tank as percent of capacity. 0 = empty. 100 = full.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemRelativeLevel *SignalFloat `json:"powertrainFuelSystemRelativeLevel,omitempty"`
	// High level information of fuel types supported
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemSupportedFuelTypes *SignalString `json:"powertrainFuelSystemSupportedFuelTypes,omitempty"`
	// Remaining range in meters using all energy sources available in the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainRange *SignalFloat `json:"powertrainRange,omitempty"`
	// Target charge limit (state of charge) for battery.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingChargeLimit *SignalFloat `json:"powertrainTractionBatteryChargingChargeLimit,omitempty"`
	// True if charging is ongoing. Charging is considered to be ongoing if energy is flowing from charger to vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingIsCharging *SignalFloat `json:"powertrainTractionBatteryChargingIsCharging,omitempty"`
	// Current electrical energy flowing in/out of battery. Positive = Energy flowing in to battery, e.g. during charging. Negative = Energy flowing out of battery, e.g. during driving.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryCurrentPower *SignalFloat `json:"powertrainTractionBatteryCurrentPower,omitempty"`
	// Gross capacity of the battery.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryGrossCapacity *SignalFloat `json:"powertrainTractionBatteryGrossCapacity,omitempty"`
	// Physical state of charge of the high voltage battery, relative to net capacity. This is not necessarily the state of charge being displayed to the customer.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryStateOfChargeCurrent *SignalFloat `json:"powertrainTractionBatteryStateOfChargeCurrent,omitempty"`
	// Odometer reading, total distance travelled during the lifetime of the transmission.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTransmissionTravelledDistance *SignalFloat `json:"powertrainTransmissionTravelledDistance,omitempty"`
	// Defines the powertrain type of the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainType *SignalString `json:"powertrainType,omitempty"`
	// Vehicle speed.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	Speed *SignalFloat `json:"speed,omitempty"`
}

// SignalFilter holds the filter parameters for the signal querys.
type SignalFilter struct {
	// Filter signals by source type.
	// avalible sources are: "autopi", "macaron", "smartcar", "tesla"
	Source *string `json:"source,omitempty"`
}

type SignalFloat struct {
	// timestamp of when this data was colllected
	Timestamp time.Time `json:"timestamp"`
	// value of the signal
	Value float64 `json:"value"`
}

type SignalString struct {
	// timestamp of when this data was colllected
	Timestamp time.Time `json:"timestamp"`
	// value of the signal
	Value string `json:"value"`
}

type Vinvc struct {
	// vehicleTokenId is the token ID of the vehicle.
	VehicleTokenID *int `json:"vehicleTokenId,omitempty"`
	// vin is the vehicle identification number.
	Vin *string `json:"vin,omitempty"`
	// recordedBy is the entity that recorded the VIN.
	RecordedBy *string `json:"recordedBy,omitempty"`
	// The time the VIN was recorded.
	RecordedAt *time.Time `json:"recordedAt,omitempty"`
	// countryCode is the country code that the VIN belongs to.
	CountryCode *string `json:"countryCode,omitempty"`
	// vehicleContractAddress is the address of the vehicle contract.
	VehicleContractAddress *string `json:"vehicleContractAddress,omitempty"`
	// validFrom is the time the VC is valid from.
	ValidFrom *time.Time `json:"validFrom,omitempty"`
	// validTo is the time the VC is valid to.
	ValidTo *time.Time `json:"validTo,omitempty"`
	// rawVC is the raw VC JSON.
	RawVc string `json:"rawVC"`
}

type FloatAggregation string

const (
	FloatAggregationAvg   FloatAggregation = "AVG"
	FloatAggregationMed   FloatAggregation = "MED"
	FloatAggregationMax   FloatAggregation = "MAX"
	FloatAggregationMin   FloatAggregation = "MIN"
	FloatAggregationRand  FloatAggregation = "RAND"
	FloatAggregationFirst FloatAggregation = "FIRST"
	FloatAggregationLast  FloatAggregation = "LAST"
)

var AllFloatAggregation = []FloatAggregation{
	FloatAggregationAvg,
	FloatAggregationMed,
	FloatAggregationMax,
	FloatAggregationMin,
	FloatAggregationRand,
	FloatAggregationFirst,
	FloatAggregationLast,
}

func (e FloatAggregation) IsValid() bool {
	switch e {
	case FloatAggregationAvg, FloatAggregationMed, FloatAggregationMax, FloatAggregationMin, FloatAggregationRand, FloatAggregationFirst, FloatAggregationLast:
		return true
	}
	return false
}

func (e FloatAggregation) String() string {
	return string(e)
}

func (e *FloatAggregation) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FloatAggregation(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FloatAggregation", str)
	}
	return nil
}

func (e FloatAggregation) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Privilege string

const (
	PrivilegeVehicleNonLocationData Privilege = "VEHICLE_NON_LOCATION_DATA"
	PrivilegeVehicleCommands        Privilege = "VEHICLE_COMMANDS"
	PrivilegeVehicleCurrentLocation Privilege = "VEHICLE_CURRENT_LOCATION"
	PrivilegeVehicleAllTimeLocation Privilege = "VEHICLE_ALL_TIME_LOCATION"
	PrivilegeVehicleVinCredential   Privilege = "VEHICLE_VIN_CREDENTIAL"
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

type StringAggregation string

const (
	// Randomly select a value from the group.
	StringAggregationRand StringAggregation = "RAND"
	// Select the most frequently occurring value in the group.
	StringAggregationTop StringAggregation = "TOP"
	// Return a list of unique values in the group.
	StringAggregationUnique StringAggregation = "UNIQUE"
	// Return value in group associated with the minimum time value.
	StringAggregationFirst StringAggregation = "FIRST"
	// Return value in group associated with the maximum time value.
	StringAggregationLast StringAggregation = "LAST"
)

var AllStringAggregation = []StringAggregation{
	StringAggregationRand,
	StringAggregationTop,
	StringAggregationUnique,
	StringAggregationFirst,
	StringAggregationLast,
}

func (e StringAggregation) IsValid() bool {
	switch e {
	case StringAggregationRand, StringAggregationTop, StringAggregationUnique, StringAggregationFirst, StringAggregationLast:
		return true
	}
	return false
}

func (e StringAggregation) String() string {
	return string(e)
}

func (e *StringAggregation) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = StringAggregation(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid StringAggregation", str)
	}
	return nil
}

func (e StringAggregation) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
