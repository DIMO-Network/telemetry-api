package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// AngularVelocityYaw is the resolver for the angularVelocityYaw
func (r *signalAggregationsResolver) AngularVelocityYaw(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "angularVelocityYaw", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ChassisAxleRow1WheelLeftSpeed is the resolver for the chassisAxleRow1WheelLeftSpeed
func (r *signalAggregationsResolver) ChassisAxleRow1WheelLeftSpeed(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelLeftSpeed", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ChassisAxleRow1WheelLeftTirePressure is the resolver for the chassisAxleRow1WheelLeftTirePressure
func (r *signalAggregationsResolver) ChassisAxleRow1WheelLeftTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelLeftTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ChassisAxleRow1WheelRightSpeed is the resolver for the chassisAxleRow1WheelRightSpeed
func (r *signalAggregationsResolver) ChassisAxleRow1WheelRightSpeed(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelRightSpeed", Agg: agg.String()}]
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

// CurrentLocationIsRedacted is the resolver for the currentLocationIsRedacted
func (r *signalAggregationsResolver) CurrentLocationIsRedacted(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationIsRedacted", Agg: agg.String()}]
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

// ObdCommandedEgr is the resolver for the obdCommandedEGR
func (r *signalAggregationsResolver) ObdCommandedEgr(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdCommandedEGR", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdDistanceSinceDTCClear is the resolver for the obdDistanceSinceDTCClear
func (r *signalAggregationsResolver) ObdDistanceSinceDTCClear(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdDistanceSinceDTCClear", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdDistanceWithMil is the resolver for the obdDistanceWithMIL
func (r *signalAggregationsResolver) ObdDistanceWithMil(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdDistanceWithMIL", Agg: agg.String()}]
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

// ObdLongTermFuelTrim1 is the resolver for the obdLongTermFuelTrim1
func (r *signalAggregationsResolver) ObdLongTermFuelTrim1(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdLongTermFuelTrim1", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdMap is the resolver for the obdMAP
func (r *signalAggregationsResolver) ObdMap(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdMAP", Agg: agg.String()}]
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

// ObdShortTermFuelTrim1 is the resolver for the obdShortTermFuelTrim1
func (r *signalAggregationsResolver) ObdShortTermFuelTrim1(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdShortTermFuelTrim1", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

// ObdWarmupsSinceDTCClear is the resolver for the obdWarmupsSinceDTCClear
func (r *signalAggregationsResolver) ObdWarmupsSinceDTCClear(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdWarmupsSinceDTCClear", Agg: agg.String()}]
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

// PowertrainCombustionEngineEngineOilRelativeLevel is the resolver for the powertrainCombustionEngineEngineOilRelativeLevel
func (r *signalAggregationsResolver) PowertrainCombustionEngineEngineOilRelativeLevel(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineEngineOilRelativeLevel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
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

// PowertrainFuelSystemRelativeLevel is the resolver for the powertrainFuelSystemRelativeLevel
func (r *signalAggregationsResolver) PowertrainFuelSystemRelativeLevel(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainFuelSystemRelativeLevel", Agg: agg.String()}]
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

// PowertrainTractionBatteryTemperatureAverage is the resolver for the powertrainTractionBatteryTemperatureAverage
func (r *signalAggregationsResolver) PowertrainTractionBatteryTemperatureAverage(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryTemperatureAverage", Agg: agg.String()}]
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
