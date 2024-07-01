// Code generated  with `make gql-model` DO NOT EDIT.
package graph

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// ChassisAxleRow1WheelLeftTirePressure is the resolver for the chassisAxleRow1WheelLeftTirePressure
func (r *signalAggregationsResolver) ChassisAxleRow1WheelLeftTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelLeftTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ChassisAxleRow1WheelRightTirePressure is the resolver for the chassisAxleRow1WheelRightTirePressure
func (r *signalAggregationsResolver) ChassisAxleRow1WheelRightTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelRightTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ChassisAxleRow2WheelLeftTirePressure is the resolver for the chassisAxleRow2WheelLeftTirePressure
func (r *signalAggregationsResolver) ChassisAxleRow2WheelLeftTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow2WheelLeftTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ChassisAxleRow2WheelRightTirePressure is the resolver for the chassisAxleRow2WheelRightTirePressure
func (r *signalAggregationsResolver) ChassisAxleRow2WheelRightTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow2WheelRightTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// CurrentLocationAltitude is the resolver for the currentLocationAltitude
func (r *signalAggregationsResolver) CurrentLocationAltitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationAltitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// CurrentLocationLatitude is the resolver for the currentLocationLatitude
func (r *signalAggregationsResolver) CurrentLocationLatitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationLatitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// CurrentLocationLongitude is the resolver for the currentLocationLongitude
func (r *signalAggregationsResolver) CurrentLocationLongitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationLongitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// CurrentLocationTimestamp is the resolver for the currentLocationTimestamp
func (r *signalAggregationsResolver) CurrentLocationTimestamp(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "currentLocationTimestamp", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// DimoAftermarketHdop is the resolver for the dimoAftermarketHDOP
func (r *signalAggregationsResolver) DimoAftermarketHdop(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "dimoAftermarketHDOP", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// DimoAftermarketNsat is the resolver for the dimoAftermarketNSAT
func (r *signalAggregationsResolver) DimoAftermarketNsat(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "dimoAftermarketNSAT", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// DimoAftermarketSsid is the resolver for the dimoAftermarketSSID
func (r *signalAggregationsResolver) DimoAftermarketSsid(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "dimoAftermarketSSID", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// DimoAftermarketWPAState is the resolver for the dimoAftermarketWPAState
func (r *signalAggregationsResolver) DimoAftermarketWPAState(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "dimoAftermarketWPAState", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// DimoIsLocationRedacted is the resolver for the dimoIsLocationRedacted
func (r *signalAggregationsResolver) DimoIsLocationRedacted(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "dimoIsLocationRedacted", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ExteriorAirTemperature is the resolver for the exteriorAirTemperature
func (r *signalAggregationsResolver) ExteriorAirTemperature(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "exteriorAirTemperature", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// LowVoltageBatteryCurrentVoltage is the resolver for the lowVoltageBatteryCurrentVoltage
func (r *signalAggregationsResolver) LowVoltageBatteryCurrentVoltage(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "lowVoltageBatteryCurrentVoltage", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdBarometricPressure is the resolver for the obdBarometricPressure
func (r *signalAggregationsResolver) ObdBarometricPressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdBarometricPressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdEngineLoad is the resolver for the obdEngineLoad
func (r *signalAggregationsResolver) ObdEngineLoad(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdEngineLoad", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdIntakeTemp is the resolver for the obdIntakeTemp
func (r *signalAggregationsResolver) ObdIntakeTemp(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdIntakeTemp", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdRunTime is the resolver for the obdRunTime
func (r *signalAggregationsResolver) ObdRunTime(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdRunTime", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainCombustionEngineEct is the resolver for the powertrainCombustionEngineECT
func (r *signalAggregationsResolver) PowertrainCombustionEngineEct(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineECT", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainCombustionEngineEngineOilLevel is the resolver for the powertrainCombustionEngineEngineOilLevel
func (r *signalAggregationsResolver) PowertrainCombustionEngineEngineOilLevel(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "powertrainCombustionEngineEngineOilLevel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// PowertrainCombustionEngineMaf is the resolver for the powertrainCombustionEngineMAF
func (r *signalAggregationsResolver) PowertrainCombustionEngineMaf(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineMAF", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainCombustionEngineSpeed is the resolver for the powertrainCombustionEngineSpeed
func (r *signalAggregationsResolver) PowertrainCombustionEngineSpeed(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineSpeed", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainCombustionEngineTps is the resolver for the powertrainCombustionEngineTPS
func (r *signalAggregationsResolver) PowertrainCombustionEngineTps(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineTPS", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainFuelSystemAbsoluteLevel is the resolver for the powertrainFuelSystemAbsoluteLevel
func (r *signalAggregationsResolver) PowertrainFuelSystemAbsoluteLevel(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainFuelSystemAbsoluteLevel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainFuelSystemSupportedFuelTypes is the resolver for the powertrainFuelSystemSupportedFuelTypes
func (r *signalAggregationsResolver) PowertrainFuelSystemSupportedFuelTypes(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "powertrainFuelSystemSupportedFuelTypes", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// PowertrainRange is the resolver for the powertrainRange
func (r *signalAggregationsResolver) PowertrainRange(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainRange", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainTractionBatteryChargingChargeLimit is the resolver for the powertrainTractionBatteryChargingChargeLimit
func (r *signalAggregationsResolver) PowertrainTractionBatteryChargingChargeLimit(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryChargingChargeLimit", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainTractionBatteryChargingIsCharging is the resolver for the powertrainTractionBatteryChargingIsCharging
func (r *signalAggregationsResolver) PowertrainTractionBatteryChargingIsCharging(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryChargingIsCharging", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainTractionBatteryCurrentPower is the resolver for the powertrainTractionBatteryCurrentPower
func (r *signalAggregationsResolver) PowertrainTractionBatteryCurrentPower(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryCurrentPower", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainTractionBatteryGrossCapacity is the resolver for the powertrainTractionBatteryGrossCapacity
func (r *signalAggregationsResolver) PowertrainTractionBatteryGrossCapacity(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryGrossCapacity", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainTractionBatteryStateOfChargeCurrent is the resolver for the powertrainTractionBatteryStateOfChargeCurrent
func (r *signalAggregationsResolver) PowertrainTractionBatteryStateOfChargeCurrent(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryStateOfChargeCurrent", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainTransmissionTravelledDistance is the resolver for the powertrainTransmissionTravelledDistance
func (r *signalAggregationsResolver) PowertrainTransmissionTravelledDistance(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTransmissionTravelledDistance", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// PowertrainType is the resolver for the powertrainType
func (r *signalAggregationsResolver) PowertrainType(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "powertrainType", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// Speed is the resolver for the speed
func (r *signalAggregationsResolver) Speed(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "speed", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// VehicleIdentificationBrand is the resolver for the vehicleIdentificationBrand
func (r *signalAggregationsResolver) VehicleIdentificationBrand(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "vehicleIdentificationBrand", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// VehicleIdentificationModel is the resolver for the vehicleIdentificationModel
func (r *signalAggregationsResolver) VehicleIdentificationModel(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "vehicleIdentificationModel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

// VehicleIdentificationYear is the resolver for the vehicleIdentificationYear
func (r *signalAggregationsResolver) VehicleIdentificationYear(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "vehicleIdentificationYear", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}
