// Code generated  with `make gql-model` DO NOT EDIT.
package graph

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

func (r *signalAggregationsResolver) ChassisAxleRow1WheelLeftTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelLeftTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) ChassisAxleRow1WheelRightTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow1WheelRightTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) ChassisAxleRow2WheelLeftTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow2WheelLeftTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) ChassisAxleRow2WheelRightTirePressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "chassisAxleRow2WheelRightTirePressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) CurrentLocationAltitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationAltitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) CurrentLocationLatitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationLatitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) CurrentLocationLongitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationLongitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) CurrentLocationTimestamp(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "currentLocationTimestamp", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) DIMOAftermarketHDOP(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "dimoAftermarketHDOP", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) DIMOAftermarketNSAT(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "dimoAftermarketNSAT", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) DIMOAftermarketSSID(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "dimoAftermarketSSID", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) DIMOAftermarketWPAState(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "dimoAftermarketWPAState", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) DIMOIsLocationRedacted(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "dimoIsLocationRedacted", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) ExteriorAirTemperature(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "exteriorAirTemperature", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) LowVoltageBatteryCurrentVoltage(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "lowVoltageBatteryCurrentVoltage", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) OBDBarometricPressure(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdBarometricPressure", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) OBDEngineLoad(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdEngineLoad", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) OBDIntakeTemp(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdIntakeTemp", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) OBDRunTime(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "obdRunTime", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainCombustionEngineECT(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineECT", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainCombustionEngineEngineOilLevel(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "powertrainCombustionEngineEngineOilLevel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) PowertrainCombustionEngineMAF(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineMAF", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainCombustionEngineSpeed(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineSpeed", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainCombustionEngineTPS(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainCombustionEngineTPS", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainFuelSystemAbsoluteLevel(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainFuelSystemAbsoluteLevel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainFuelSystemSupportedFuelTypes(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "powertrainFuelSystemSupportedFuelTypes", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) PowertrainRange(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainRange", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainTractionBatteryChargingChargeLimit(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryChargingChargeLimit", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainTractionBatteryChargingIsCharging(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryChargingIsCharging", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainTractionBatteryCurrentPower(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryCurrentPower", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainTractionBatteryGrossCapacity(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryGrossCapacity", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainTractionBatteryStateOfChargeCurrent(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTractionBatteryStateOfChargeCurrent", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainTransmissionTravelledDistance(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "powertrainTransmissionTravelledDistance", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) PowertrainType(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "powertrainType", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) Speed(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "speed", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}

func (r *signalAggregationsResolver) VehicleIdentificationBrand(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "vehicleIdentificationBrand", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) VehicleIdentificationModel(ctx context.Context, obj *model.SignalAggregations, agg model.StringAggregation) (*string, error) {
	vs, ok := obj.ValueStrings[model.AliasKey{Name: "vehicleIdentificationModel", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vs, nil
}

func (r *signalAggregationsResolver) VehicleIdentificationYear(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	vn, ok := obj.ValueNumbers[model.AliasKey{Name: "vehicleIdentificationYear", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	return &vn, nil
}
