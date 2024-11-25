// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// The AftermarketDeviceBy input is used to specify a unique aftermarket device to query for last active status.
type AftermarketDeviceBy struct {
	TokenID *int            `json:"tokenId,omitempty"`
	Address *common.Address `json:"address,omitempty"`
	Serial  *string         `json:"serial,omitempty"`
}

type DeviceActivity struct {
	// lastActive indicates the start of a 3 hour block during which the device was last active.
	LastActive *time.Time `json:"lastActive,omitempty"`
}

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
	// Approximate Latitude of vehicle in WGS 84 geodetic coordinates.
	// This returned location is the center of the h3 cell with resolution 6 that the location is in.
	// More Info on H3: https://h3geo.org/
	// Unit: 'degrees' Min: '-90' Max: '90'
	// Required Privileges: [VEHICLE_APPROXIMATE_LOCATION OR VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationApproximateLatitude *SignalFloat `json:"currentLocationApproximateLatitude,omitempty"`
	// Approximate Longitude of vehicle in WGS 84 geodetic coordinates.
	// This returned location is the center of the h3 cell with resolution 6 that the location is in.
	// More Info on H3: https://h3geo.org/
	// Unit: 'degrees' Min: '-180' Max: '180'
	// Required Privileges: [VEHICLE_APPROXIMATE_LOCATION OR VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationApproximateLongitude *SignalFloat `json:"currentLocationApproximateLongitude,omitempty"`
	// Vehicle rotation rate along Z (vertical).
	// Unit: 'degrees/s'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	AngularVelocityYaw *SignalFloat `json:"angularVelocityYaw,omitempty"`
	// Rotational speed of a vehicle's wheel.
	// Unit: 'km/h'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelLeftSpeed *SignalFloat `json:"chassisAxleRow1WheelLeftSpeed,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelLeftTirePressure *SignalFloat `json:"chassisAxleRow1WheelLeftTirePressure,omitempty"`
	// Rotational speed of a vehicle's wheel.
	// Unit: 'km/h'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelRightSpeed *SignalFloat `json:"chassisAxleRow1WheelRightSpeed,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow1WheelRightTirePressure *SignalFloat `json:"chassisAxleRow1WheelRightTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelLeftTirePressure *SignalFloat `json:"chassisAxleRow2WheelLeftTirePressure,omitempty"`
	// Tire pressure in kilo-Pascal.
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ChassisAxleRow2WheelRightTirePressure *SignalFloat `json:"chassisAxleRow2WheelRightTirePressure,omitempty"`
	// Current altitude relative to WGS 84 reference ellipsoid, as measured at the position of GNSS receiver antenna.
	// Unit: 'm'
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationAltitude *SignalFloat `json:"currentLocationAltitude,omitempty"`
	// Indicates if the latitude and longitude signals at the current timestamp have been redacted using a privacy zone.
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationIsRedacted *SignalFloat `json:"currentLocationIsRedacted,omitempty"`
	// Current latitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Unit: 'degrees' Min: '-90' Max: '90'
	// Required Privileges: [VEHICLE_ALL_TIME_LOCATION]
	CurrentLocationLatitude *SignalFloat `json:"currentLocationLatitude,omitempty"`
	// Current longitude of vehicle in WGS 84 geodetic coordinates, as measured at the position of GNSS receiver antenna.
	// Unit: 'degrees' Min: '-180' Max: '180'
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
	// Unit: 'celsius'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ExteriorAirTemperature *SignalFloat `json:"exteriorAirTemperature,omitempty"`
	// Current Voltage of the low voltage battery.
	// Unit: 'V'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	LowVoltageBatteryCurrentVoltage *SignalFloat `json:"lowVoltageBatteryCurrentVoltage,omitempty"`
	// PID 33 - Barometric pressure
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDBarometricPressure *SignalFloat `json:"obdBarometricPressure,omitempty"`
	// PID 2C - Commanded exhaust gas recirculation (EGR)
	// Unit: 'percent'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDCommandedEGR *SignalFloat `json:"obdCommandedEGR,omitempty"`
	// PID 2E - Commanded evaporative purge (EVAP) valve
	// Unit: 'percent'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDCommandedEVAP *SignalFloat `json:"obdCommandedEVAP,omitempty"`
	// PID 31 - Distance traveled since codes cleared
	// Unit: 'km'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDDistanceSinceDTCClear *SignalFloat `json:"obdDistanceSinceDTCClear,omitempty"`
	// PID 21 - Distance traveled with MIL on
	// Unit: 'km'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDDistanceWithMIL *SignalFloat `json:"obdDistanceWithMIL,omitempty"`
	// PID 04 - Engine load in percent - 0 = no load, 100 = full load
	// Unit: 'percent'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDEngineLoad *SignalFloat `json:"obdEngineLoad,omitempty"`
	// PID 0A - Fuel pressure
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDFuelPressure *SignalFloat `json:"obdFuelPressure,omitempty"`
	// PID 0F - Intake temperature
	// Unit: 'celsius'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDIntakeTemp *SignalFloat `json:"obdIntakeTemp,omitempty"`
	// PID 07 - Long Term (learned) Fuel Trim - Bank 1 - negative percent leaner, positive percent richer
	// Unit: 'percent'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDLongTermFuelTrim1 *SignalFloat `json:"obdLongTermFuelTrim1,omitempty"`
	// PID 0B - Intake manifold pressure
	// Unit: 'kPa'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDMAP *SignalFloat `json:"obdMAP,omitempty"`
	// PID 2x (byte CD) - Voltage for wide range/band oxygen sensor
	// Unit: 'V'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDO2WRSensor1Voltage *SignalFloat `json:"obdO2WRSensor1Voltage,omitempty"`
	// PID 2x (byte CD) - Voltage for wide range/band oxygen sensor
	// Unit: 'V'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDO2WRSensor2Voltage *SignalFloat `json:"obdO2WRSensor2Voltage,omitempty"`
	// PID 1F - Engine run time
	// Unit: 's'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDRunTime *SignalFloat `json:"obdRunTime,omitempty"`
	// PID 06 - Short Term (immediate) Fuel Trim - Bank 1 - negative percent leaner, positive percent richer
	// Unit: 'percent'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDShortTermFuelTrim1 *SignalFloat `json:"obdShortTermFuelTrim1,omitempty"`
	// PID 30 - Number of warm-ups since codes cleared
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	OBDWarmupsSinceDTCClear *SignalFloat `json:"obdWarmupsSinceDTCClear,omitempty"`
	// Capacity in liters of the Diesel Exhaust Fluid Tank.
	// Unit: 'l'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineDieselExhaustFluidCapacity *SignalFloat `json:"powertrainCombustionEngineDieselExhaustFluidCapacity,omitempty"`
	// Level of the Diesel Exhaust Fluid tank as percent of capacity. 0 = empty. 100 = full.
	// Unit: 'percent' Min: '0' Max: '100'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineDieselExhaustFluidLevel *SignalFloat `json:"powertrainCombustionEngineDieselExhaustFluidLevel,omitempty"`
	// Engine coolant temperature.
	// Unit: 'celsius'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineECT *SignalFloat `json:"powertrainCombustionEngineECT,omitempty"`
	// Engine oil level.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineEngineOilLevel *SignalString `json:"powertrainCombustionEngineEngineOilLevel,omitempty"`
	// Engine oil level as a percentage.
	// Unit: 'percent' Min: '0' Max: '100'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineEngineOilRelativeLevel *SignalFloat `json:"powertrainCombustionEngineEngineOilRelativeLevel,omitempty"`
	// Grams of air drawn into engine per second.
	// Unit: 'g/s'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineMAF *SignalFloat `json:"powertrainCombustionEngineMAF,omitempty"`
	// Engine speed measured as rotations per minute.
	// Unit: 'rpm'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineSpeed *SignalFloat `json:"powertrainCombustionEngineSpeed,omitempty"`
	// Current throttle position.
	// Unit: 'percent' Max: '100'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineTPS *SignalFloat `json:"powertrainCombustionEngineTPS,omitempty"`
	// Current engine torque. Shall be reported as 0 during engine breaking.
	// Unit: 'Nm'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainCombustionEngineTorque *SignalFloat `json:"powertrainCombustionEngineTorque,omitempty"`
	// Current available fuel in the fuel tank expressed in liters.
	// Unit: 'l'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemAbsoluteLevel *SignalFloat `json:"powertrainFuelSystemAbsoluteLevel,omitempty"`
	// Level in fuel tank as percent of capacity. 0 = empty. 100 = full.
	// Unit: 'percent' Min: '0' Max: '100'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemRelativeLevel *SignalFloat `json:"powertrainFuelSystemRelativeLevel,omitempty"`
	// High level information of fuel types supported
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainFuelSystemSupportedFuelTypes *SignalString `json:"powertrainFuelSystemSupportedFuelTypes,omitempty"`
	// Remaining range in meters using all energy sources available in the vehicle.
	// Unit: 'm'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainRange *SignalFloat `json:"powertrainRange,omitempty"`
	// Amount of charge added to the high voltage battery during the current charging session, expressed in kilowatt-hours.
	// Unit: 'kWh'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingAddedEnergy *SignalFloat `json:"powertrainTractionBatteryChargingAddedEnergy,omitempty"`
	// Target charge limit (state of charge) for battery.
	// Unit: 'percent' Min: '0' Max: '100'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingChargeLimit *SignalFloat `json:"powertrainTractionBatteryChargingChargeLimit,omitempty"`
	// True if charging is ongoing. Charging is considered to be ongoing if energy is flowing from charger to vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryChargingIsCharging *SignalFloat `json:"powertrainTractionBatteryChargingIsCharging,omitempty"`
	// Current electrical energy flowing in/out of battery. Positive = Energy flowing in to battery, e.g. during charging. Negative = Energy flowing out of battery, e.g. during driving.
	// Unit: 'W'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryCurrentPower *SignalFloat `json:"powertrainTractionBatteryCurrentPower,omitempty"`
	// Current Voltage of the battery.
	// Unit: 'V'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryCurrentVoltage *SignalFloat `json:"powertrainTractionBatteryCurrentVoltage,omitempty"`
	// Gross capacity of the battery.
	// Unit: 'kWh'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryGrossCapacity *SignalFloat `json:"powertrainTractionBatteryGrossCapacity,omitempty"`
	// Remaining range in meters using only battery.
	// Unit: 'm'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryRange *SignalFloat `json:"powertrainTractionBatteryRange,omitempty"`
	// Physical state of charge of the high voltage battery, relative to net capacity. This is not necessarily the state of charge being displayed to the customer.
	// Unit: 'percent' Min: '0' Max: '100.0'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryStateOfChargeCurrent *SignalFloat `json:"powertrainTractionBatteryStateOfChargeCurrent,omitempty"`
	// Current average temperature of the battery cells.
	// Unit: 'celsius'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTractionBatteryTemperatureAverage *SignalFloat `json:"powertrainTractionBatteryTemperatureAverage,omitempty"`
	// The current gear. 0=Neutral, 1/2/..=Forward, -1/-2/..=Reverse.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTransmissionCurrentGear *SignalFloat `json:"powertrainTransmissionCurrentGear,omitempty"`
	// The current gearbox temperature.
	// Unit: 'celsius'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTransmissionTemperature *SignalFloat `json:"powertrainTransmissionTemperature,omitempty"`
	// Odometer reading, total distance travelled during the lifetime of the transmission.
	// Unit: 'km'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainTransmissionTravelledDistance *SignalFloat `json:"powertrainTransmissionTravelledDistance,omitempty"`
	// Defines the powertrain type of the vehicle.
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	PowertrainType *SignalString `json:"powertrainType,omitempty"`
	// Remaining distance to service (of any kind). Negative values indicate service overdue.
	// Unit: 'km'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	ServiceDistanceToService *SignalFloat `json:"serviceDistanceToService,omitempty"`
	// Vehicle speed.
	// Unit: 'km/h'
	// Required Privileges: [VEHICLE_NON_LOCATION_DATA]
	Speed *SignalFloat `json:"speed,omitempty"`
}

// SignalFilter holds the filter parameters for the signal querys.
type SignalFilter struct {
	// Filter signals by source type.
	// available sources are: "autopi", "macaron", "ruptela", "smartcar", "tesla"
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
	PrivilegeVehicleNonLocationData     Privilege = "VEHICLE_NON_LOCATION_DATA"
	PrivilegeVehicleCommands            Privilege = "VEHICLE_COMMANDS"
	PrivilegeVehicleCurrentLocation     Privilege = "VEHICLE_CURRENT_LOCATION"
	PrivilegeVehicleAllTimeLocation     Privilege = "VEHICLE_ALL_TIME_LOCATION"
	PrivilegeVehicleVinCredential       Privilege = "VEHICLE_VIN_CREDENTIAL"
	PrivilegeVehicleApproximateLocation Privilege = "VEHICLE_APPROXIMATE_LOCATION"
	PrivilegeManufacturerDeviceLastSeen Privilege = "MANUFACTURER_DEVICE_LAST_SEEN"
)

var AllPrivilege = []Privilege{
	PrivilegeVehicleNonLocationData,
	PrivilegeVehicleCommands,
	PrivilegeVehicleCurrentLocation,
	PrivilegeVehicleAllTimeLocation,
	PrivilegeVehicleVinCredential,
	PrivilegeVehicleApproximateLocation,
	PrivilegeManufacturerDeviceLastSeen,
}

func (e Privilege) IsValid() bool {
	switch e {
	case PrivilegeVehicleNonLocationData, PrivilegeVehicleCommands, PrivilegeVehicleCurrentLocation, PrivilegeVehicleAllTimeLocation, PrivilegeVehicleVinCredential, PrivilegeVehicleApproximateLocation, PrivilegeManufacturerDeviceLastSeen:
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
