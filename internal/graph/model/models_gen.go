// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

// The root query type for the GraphQL schema.
type Query struct {
}

type SignalAggregations struct {
	// Timestamp of the aggregated data.
	Timestamp time.Time `json:"timestamp"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelLeftTirePressure *float64 `json:"chassisAxleRow1WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelRightTirePressure *float64 `json:"chassisAxleRow1WheelRightTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelLeftTirePressure *float64 `json:"chassisAxleRow2WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelRightTirePressure *float64 `json:"chassisAxleRow2WheelRightTirePressure,omitempty"`
	// Current altitude relative to WGS 84 reference ellipsoid, as measured at the position of GNSS receiver antenna.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationAltitude *float64 `json:"currentLocationAltitude,omitempty"`
	// Current latitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLatitude *float64 `json:"currentLocationLatitude,omitempty"`
	// Current longitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLongitude *float64 `json:"currentLocationLongitude,omitempty"`
	// Timestamp from GNSS system for current location, formatted according to ISO 8601 with UTC time zone.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationTimestamp *string `json:"currentLocationTimestamp,omitempty"`
	// Horizontal dilution of precision of GPS
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketHDOP *float64 `json:"dIMOAftermarketHDOP,omitempty"`
	// Number of sync satellites for GPS
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketNSAT *float64 `json:"dIMOAftermarketNSAT,omitempty"`
	// Service Set Ientifier for the wifi.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketSSID *string `json:"dIMOAftermarketSSID,omitempty"`
	// Indicate the current wpa state for the devices wifi
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketWPAState *string `json:"dIMOAftermarketWPAState,omitempty"`
	// Air temperature outside the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ExteriorAirTemperature *float64 `json:"exteriorAirTemperature,omitempty"`
	// Current Voltage of the low voltage battery.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	LowVoltageBatteryCurrentVoltage *float64 `json:"lowVoltageBatteryCurrentVoltage,omitempty"`
	// PID 33 - Barometric pressure
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDBarometricPressure *float64 `json:"oBDBarometricPressure,omitempty"`
	// PID 04 - Engine load in percent - 0 = no load, 100 = full load
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDEngineLoad *float64 `json:"oBDEngineLoad,omitempty"`
	// PID 0F - Intake temperature
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDIntakeTemp *float64 `json:"oBDIntakeTemp,omitempty"`
	// PID 1F - Engine run time
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDRunTime *float64 `json:"oBDRunTime,omitempty"`
	// Engine coolant temperature.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineECT *float64 `json:"powertrainCombustionEngineECT,omitempty"`
	// Engine oil level.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineEngineOilLevel *string `json:"powertrainCombustionEngineEngineOilLevel,omitempty"`
	// Grams of air drawn into engine per second.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineMAF *float64 `json:"powertrainCombustionEngineMAF,omitempty"`
	// Engine speed measured as rotations per minute.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineSpeed *float64 `json:"powertrainCombustionEngineSpeed,omitempty"`
	// Current throttle position.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineTPS *float64 `json:"powertrainCombustionEngineTPS,omitempty"`
	// Current available fuel in the fuel tank expressed in liters.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemAbsoluteLevel *float64 `json:"powertrainFuelSystemAbsoluteLevel,omitempty"`
	// High level information of fuel types supported
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemSupportedFuelTypes *string `json:"powertrainFuelSystemSupportedFuelTypes,omitempty"`
	// Remaining range in meters using all energy sources available in the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainRange *float64 `json:"powertrainRange,omitempty"`
	// Target charge limit (state of charge) for battery.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingChargeLimit *float64 `json:"powertrainTractionBatteryChargingChargeLimit,omitempty"`
	// True if charging is ongoing. Charging is considered to be ongoing if energy is flowing from charger to vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingIsCharging *string `json:"powertrainTractionBatteryChargingIsCharging,omitempty"`
	// Gross capacity of the battery.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryGrossCapacity *float64 `json:"powertrainTractionBatteryGrossCapacity,omitempty"`
	// Physical state of charge of the high voltage battery, relative to net capacity. This is not necessarily the state of charge being displayed to the customer.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryStateOfChargeCurrent *float64 `json:"powertrainTractionBatteryStateOfChargeCurrent,omitempty"`
	// Odometer reading, total distance travelled during the lifetime of the transmission.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTransmissionTravelledDistance *float64 `json:"powertrainTransmissionTravelledDistance,omitempty"`
	// Defines the powertrain type of the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainType *string `json:"powertrainType,omitempty"`
	// Vehicle speed.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	Speed *float64 `json:"speed,omitempty"`
	// Vehicle brand or manufacturer.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	VehicleIdentificationBrand *string `json:"vehicleIdentificationBrand,omitempty"`
	// Vehicle model.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	VehicleIdentificationModel *string `json:"vehicleIdentificationModel,omitempty"`
	// Model year of the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	VehicleIdentificationYear *float64 `json:"vehicleIdentificationYear,omitempty"`
}

type SignalCollection struct {
	// The last time any signal was seen matching the filter.
	LastSeen *time.Time `json:"lastSeen,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelLeftTirePressure *SignalFloat `json:"chassisAxleRow1WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelRightTirePressure *SignalFloat `json:"chassisAxleRow1WheelRightTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelLeftTirePressure *SignalFloat `json:"chassisAxleRow2WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelRightTirePressure *SignalFloat `json:"chassisAxleRow2WheelRightTirePressure,omitempty"`
	// Current altitude relative to WGS 84 reference ellipsoid, as measured at the position of GNSS receiver antenna.
	// Required Privlieges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationAltitude *SignalFloat `json:"currentLocationAltitude,omitempty"`
	// Current latitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Required Privlieges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLatitude *SignalFloat `json:"currentLocationLatitude,omitempty"`
	// Current longitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Required Privlieges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLongitude *SignalFloat `json:"currentLocationLongitude,omitempty"`
	// Timestamp from GNSS system for current location, formatted according to ISO 8601 with UTC time zone.
	// Required Privlieges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationTimestamp *SignalString `json:"currentLocationTimestamp,omitempty"`
	// Horizontal dilution of precision of GPS
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketHDOP *SignalFloat `json:"dIMOAftermarketHDOP,omitempty"`
	// Number of sync satellites for GPS
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketNSAT *SignalFloat `json:"dIMOAftermarketNSAT,omitempty"`
	// Service Set Ientifier for the wifi.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketSSID *SignalString `json:"dIMOAftermarketSSID,omitempty"`
	// Indicate the current wpa state for the devices wifi
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	DIMOAftermarketWPAState *SignalString `json:"dIMOAftermarketWPAState,omitempty"`
	// Air temperature outside the vehicle.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	ExteriorAirTemperature *SignalFloat `json:"exteriorAirTemperature,omitempty"`
	// Current Voltage of the low voltage battery.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	LowVoltageBatteryCurrentVoltage *SignalFloat `json:"lowVoltageBatteryCurrentVoltage,omitempty"`
	// PID 33 - Barometric pressure
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	OBDBarometricPressure *SignalFloat `json:"oBDBarometricPressure,omitempty"`
	// PID 04 - Engine load in percent - 0 = no load, 100 = full load
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	OBDEngineLoad *SignalFloat `json:"oBDEngineLoad,omitempty"`
	// PID 0F - Intake temperature
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	OBDIntakeTemp *SignalFloat `json:"oBDIntakeTemp,omitempty"`
	// PID 1F - Engine run time
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	OBDRunTime *SignalFloat `json:"oBDRunTime,omitempty"`
	// Engine coolant temperature.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineECT *SignalFloat `json:"powertrainCombustionEngineECT,omitempty"`
	// Engine oil level.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineEngineOilLevel *SignalString `json:"powertrainCombustionEngineEngineOilLevel,omitempty"`
	// Grams of air drawn into engine per second.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineMAF *SignalFloat `json:"powertrainCombustionEngineMAF,omitempty"`
	// Engine speed measured as rotations per minute.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineSpeed *SignalFloat `json:"powertrainCombustionEngineSpeed,omitempty"`
	// Current throttle position.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineTPS *SignalFloat `json:"powertrainCombustionEngineTPS,omitempty"`
	// Current available fuel in the fuel tank expressed in liters.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemAbsoluteLevel *SignalFloat `json:"powertrainFuelSystemAbsoluteLevel,omitempty"`
	// High level information of fuel types supported
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemSupportedFuelTypes *SignalString `json:"powertrainFuelSystemSupportedFuelTypes,omitempty"`
	// Remaining range in meters using all energy sources available in the vehicle.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainRange *SignalFloat `json:"powertrainRange,omitempty"`
	// Target charge limit (state of charge) for battery.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingChargeLimit *SignalFloat `json:"powertrainTractionBatteryChargingChargeLimit,omitempty"`
	// True if charging is ongoing. Charging is considered to be ongoing if energy is flowing from charger to vehicle.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingIsCharging *SignalString `json:"powertrainTractionBatteryChargingIsCharging,omitempty"`
	// Gross capacity of the battery.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryGrossCapacity *SignalFloat `json:"powertrainTractionBatteryGrossCapacity,omitempty"`
	// Physical state of charge of the high voltage battery, relative to net capacity. This is not necessarily the state of charge being displayed to the customer.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryStateOfChargeCurrent *SignalFloat `json:"powertrainTractionBatteryStateOfChargeCurrent,omitempty"`
	// Odometer reading, total distance travelled during the lifetime of the transmission.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTransmissionTravelledDistance *SignalFloat `json:"powertrainTransmissionTravelledDistance,omitempty"`
	// Defines the powertrain type of the vehicle.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainType *SignalString `json:"powertrainType,omitempty"`
	// Vehicle speed.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	Speed *SignalFloat `json:"speed,omitempty"`
	// Vehicle brand or manufacturer.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	VehicleIdentificationBrand *SignalString `json:"vehicleIdentificationBrand,omitempty"`
	// Vehicle model.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	VehicleIdentificationModel *SignalString `json:"vehicleIdentificationModel,omitempty"`
	// Model year of the vehicle.
	// Required Privlieges: [VEHICLE_NON_LOCATION_DATA]
	VehicleIdentificationYear *SignalFloat `json:"vehicleIdentificationYear,omitempty"`
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

type FloatAggregation string

const (
	FloatAggregationAvg  FloatAggregation = "AVG"
	FloatAggregationMed  FloatAggregation = "MED"
	FloatAggregationMax  FloatAggregation = "MAX"
	FloatAggregationMin  FloatAggregation = "MIN"
	FloatAggregationRand FloatAggregation = "RAND"
)

var AllFloatAggregation = []FloatAggregation{
	FloatAggregationAvg,
	FloatAggregationMed,
	FloatAggregationMax,
	FloatAggregationMin,
	FloatAggregationRand,
}

func (e FloatAggregation) IsValid() bool {
	switch e {
	case FloatAggregationAvg, FloatAggregationMed, FloatAggregationMax, FloatAggregationMin, FloatAggregationRand:
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
)

var AllStringAggregation = []StringAggregation{
	StringAggregationRand,
	StringAggregationTop,
	StringAggregationUnique,
}

func (e StringAggregation) IsValid() bool {
	switch e {
	case StringAggregationRand, StringAggregationTop, StringAggregationUnique:
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
