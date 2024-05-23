package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// ChassisAxleRow1WheelLeftTirePressure is the resolver for the chassisAxleRow1WheelLeftTirePressure field.
func (r *signalAggregationsResolver) ChassisAxleRow1WheelLeftTirePressure(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// ChassisAxleRow1WheelRightTirePressure is the resolver for the chassisAxleRow1WheelRightTirePressure field.
func (r *signalAggregationsResolver) ChassisAxleRow1WheelRightTirePressure(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// ChassisAxleRow2WheelLeftTirePressure is the resolver for the chassisAxleRow2WheelLeftTirePressure field.
func (r *signalAggregationsResolver) ChassisAxleRow2WheelLeftTirePressure(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// ChassisAxleRow2WheelRightTirePressure is the resolver for the chassisAxleRow2WheelRightTirePressure field.
func (r *signalAggregationsResolver) ChassisAxleRow2WheelRightTirePressure(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// CurrentLocationAltitude is the resolver for the currentLocationAltitude field.
func (r *signalAggregationsResolver) CurrentLocationAltitude(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// CurrentLocationLatitude is the resolver for the currentLocationLatitude field.
func (r *signalAggregationsResolver) CurrentLocationLatitude(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// CurrentLocationLongitude is the resolver for the currentLocationLongitude field.
func (r *signalAggregationsResolver) CurrentLocationLongitude(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// CurrentLocationTimestamp is the resolver for the currentLocationTimestamp field.
func (r *signalAggregationsResolver) CurrentLocationTimestamp(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// DIMOAftermarketHdop is the resolver for the dIMOAftermarketHDOP field.
func (r *signalAggregationsResolver) DIMOAftermarketHdop(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// DIMOAftermarketNsat is the resolver for the dIMOAftermarketNSAT field.
func (r *signalAggregationsResolver) DIMOAftermarketNsat(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// DIMOAftermarketSsid is the resolver for the dIMOAftermarketSSID field.
func (r *signalAggregationsResolver) DIMOAftermarketSsid(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// DIMOAftermarketWPAState is the resolver for the dIMOAftermarketWPAState field.
func (r *signalAggregationsResolver) DIMOAftermarketWPAState(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// ExteriorAirTemperature is the resolver for the exteriorAirTemperature field.
func (r *signalAggregationsResolver) ExteriorAirTemperature(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// LowVoltageBatteryCurrentVoltage is the resolver for the lowVoltageBatteryCurrentVoltage field.
func (r *signalAggregationsResolver) LowVoltageBatteryCurrentVoltage(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// OBDBarometricPressure is the resolver for the oBDBarometricPressure field.
func (r *signalAggregationsResolver) OBDBarometricPressure(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// OBDEngineLoad is the resolver for the oBDEngineLoad field.
func (r *signalAggregationsResolver) OBDEngineLoad(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// OBDIntakeTemp is the resolver for the oBDIntakeTemp field.
func (r *signalAggregationsResolver) OBDIntakeTemp(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// OBDRunTime is the resolver for the oBDRunTime field.
func (r *signalAggregationsResolver) OBDRunTime(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainCombustionEngineEct is the resolver for the powertrainCombustionEngineECT field.
func (r *signalAggregationsResolver) PowertrainCombustionEngineEct(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainCombustionEngineEngineOilLevel is the resolver for the powertrainCombustionEngineEngineOilLevel field.
func (r *signalAggregationsResolver) PowertrainCombustionEngineEngineOilLevel(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// PowertrainCombustionEngineMaf is the resolver for the powertrainCombustionEngineMAF field.
func (r *signalAggregationsResolver) PowertrainCombustionEngineMaf(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainCombustionEngineSpeed is the resolver for the powertrainCombustionEngineSpeed field.
func (r *signalAggregationsResolver) PowertrainCombustionEngineSpeed(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainCombustionEngineTps is the resolver for the powertrainCombustionEngineTPS field.
func (r *signalAggregationsResolver) PowertrainCombustionEngineTps(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainFuelSystemAbsoluteLevel is the resolver for the powertrainFuelSystemAbsoluteLevel field.
func (r *signalAggregationsResolver) PowertrainFuelSystemAbsoluteLevel(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainFuelSystemSupportedFuelTypes is the resolver for the powertrainFuelSystemSupportedFuelTypes field.
func (r *signalAggregationsResolver) PowertrainFuelSystemSupportedFuelTypes(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// PowertrainRange is the resolver for the powertrainRange field.
func (r *signalAggregationsResolver) PowertrainRange(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainTractionBatteryChargingChargeLimit is the resolver for the powertrainTractionBatteryChargingChargeLimit field.
func (r *signalAggregationsResolver) PowertrainTractionBatteryChargingChargeLimit(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainTractionBatteryChargingIsCharging is the resolver for the powertrainTractionBatteryChargingIsCharging field.
func (r *signalAggregationsResolver) PowertrainTractionBatteryChargingIsCharging(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// PowertrainTractionBatteryGrossCapacity is the resolver for the powertrainTractionBatteryGrossCapacity field.
func (r *signalAggregationsResolver) PowertrainTractionBatteryGrossCapacity(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainTractionBatteryStateOfChargeCurrent is the resolver for the powertrainTractionBatteryStateOfChargeCurrent field.
func (r *signalAggregationsResolver) PowertrainTractionBatteryStateOfChargeCurrent(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainTransmissionTravelledDistance is the resolver for the powertrainTransmissionTravelledDistance field.
func (r *signalAggregationsResolver) PowertrainTransmissionTravelledDistance(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// PowertrainType is the resolver for the powertrainType field.
func (r *signalAggregationsResolver) PowertrainType(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// Speed is the resolver for the speed field.
func (r *signalAggregationsResolver) Speed(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// VehicleIdentificationBrand is the resolver for the vehicleIdentificationBrand field.
func (r *signalAggregationsResolver) VehicleIdentificationBrand(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// VehicleIdentificationModel is the resolver for the vehicleIdentificationModel field.
func (r *signalAggregationsResolver) VehicleIdentificationModel(ctx context.Context, obj *model.Signals, agg model.StringAggregation) ([]*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalString(ctx, &strArgs)
}

// VehicleIdentificationYear is the resolver for the vehicleIdentificationYear field.
func (r *signalAggregationsResolver) VehicleIdentificationYear(ctx context.Context, obj *model.Signals, agg model.FloatAggregation) ([]*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		Agg:        agg,
		SignalArgs: obj.SigArgs,
	}
	return r.GetSignalFloats(ctx, &floatArgs)
}

// ChassisAxleRow1WheelLeftTirePressure is the Collection resolver for the chassisAxleRow1WheelLeftTirePressure field.
func (r *signalCollectionResolver) ChassisAxleRow1WheelLeftTirePressure(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// ChassisAxleRow1WheelRightTirePressure is the Collection resolver for the chassisAxleRow1WheelRightTirePressure field.
func (r *signalCollectionResolver) ChassisAxleRow1WheelRightTirePressure(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// ChassisAxleRow2WheelLeftTirePressure is the Collection resolver for the chassisAxleRow2WheelLeftTirePressure field.
func (r *signalCollectionResolver) ChassisAxleRow2WheelLeftTirePressure(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// ChassisAxleRow2WheelRightTirePressure is the Collection resolver for the chassisAxleRow2WheelRightTirePressure field.
func (r *signalCollectionResolver) ChassisAxleRow2WheelRightTirePressure(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// CurrentLocationAltitude is the Collection resolver for the currentLocationAltitude field.
func (r *signalCollectionResolver) CurrentLocationAltitude(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// CurrentLocationLatitude is the Collection resolver for the currentLocationLatitude field.
func (r *signalCollectionResolver) CurrentLocationLatitude(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// CurrentLocationLongitude is the Collection resolver for the currentLocationLongitude field.
func (r *signalCollectionResolver) CurrentLocationLongitude(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// CurrentLocationTimestamp is the Collection resolver for the currentLocationTimestamp field.
func (r *signalCollectionResolver) CurrentLocationTimestamp(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// DIMOAftermarketHdop is the Collection resolver for the DIMOAftermarketHdop field.
func (r *signalCollectionResolver) DIMOAftermarketHdop(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// DIMOAftermarketNsat is the Collection resolver for the DIMOAftermarketNsat field.
func (r *signalCollectionResolver) DIMOAftermarketNsat(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// DIMOAftermarketSsid is the Collection resolver for the DIMOAftermarketSsid field.
func (r *signalCollectionResolver) DIMOAftermarketSsid(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// DIMOAftermarketWPAState is the Collection resolver for the dIMOAftermarketWPAState field.
func (r *signalCollectionResolver) DIMOAftermarketWPAState(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// ExteriorAirTemperature is the Collection resolver for the exteriorAirTemperature field.
func (r *signalCollectionResolver) ExteriorAirTemperature(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// LowVoltageBatteryCurrentVoltage is the Collection resolver for the lowVoltageBatteryCurrentVoltage field.
func (r *signalCollectionResolver) LowVoltageBatteryCurrentVoltage(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// OBDBarometricPressure is the Collection resolver for the oBDBarometricPressure field.
func (r *signalCollectionResolver) OBDBarometricPressure(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// OBDEngineLoad is the Collection resolver for the oBDEngineLoad field.
func (r *signalCollectionResolver) OBDEngineLoad(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// OBDIntakeTemp is the Collection resolver for the oBDIntakeTemp field.
func (r *signalCollectionResolver) OBDIntakeTemp(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// OBDRunTime is the Collection resolver for the oBDRunTime field.
func (r *signalCollectionResolver) OBDRunTime(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainCombustionEngineEct is the Collection resolver for the PowertrainCombustionEngineEct field.
func (r *signalCollectionResolver) PowertrainCombustionEngineEct(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainCombustionEngineEngineOilLevel is the Collection resolver for the powertrainCombustionEngineEngineOilLevel field.
func (r *signalCollectionResolver) PowertrainCombustionEngineEngineOilLevel(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// PowertrainCombustionEngineMaf is the Collection resolver for the PowertrainCombustionEngineMaf field.
func (r *signalCollectionResolver) PowertrainCombustionEngineMaf(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainCombustionEngineSpeed is the Collection resolver for the powertrainCombustionEngineSpeed field.
func (r *signalCollectionResolver) PowertrainCombustionEngineSpeed(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainCombustionEngineTps is the Collection resolver for the PowertrainCombustionEngineTps field.
func (r *signalCollectionResolver) PowertrainCombustionEngineTps(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainFuelSystemAbsoluteLevel is the Collection resolver for the powertrainFuelSystemAbsoluteLevel field.
func (r *signalCollectionResolver) PowertrainFuelSystemAbsoluteLevel(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainFuelSystemSupportedFuelTypes is the Collection resolver for the powertrainFuelSystemSupportedFuelTypes field.
func (r *signalCollectionResolver) PowertrainFuelSystemSupportedFuelTypes(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// PowertrainRange is the Collection resolver for the powertrainRange field.
func (r *signalCollectionResolver) PowertrainRange(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainTractionBatteryChargingChargeLimit is the Collection resolver for the powertrainTractionBatteryChargingChargeLimit field.
func (r *signalCollectionResolver) PowertrainTractionBatteryChargingChargeLimit(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainTractionBatteryChargingIsCharging is the Collection resolver for the powertrainTractionBatteryChargingIsCharging field.
func (r *signalCollectionResolver) PowertrainTractionBatteryChargingIsCharging(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// PowertrainTractionBatteryGrossCapacity is the Collection resolver for the powertrainTractionBatteryGrossCapacity field.
func (r *signalCollectionResolver) PowertrainTractionBatteryGrossCapacity(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainTractionBatteryStateOfChargeCurrent is the Collection resolver for the powertrainTractionBatteryStateOfChargeCurrent field.
func (r *signalCollectionResolver) PowertrainTractionBatteryStateOfChargeCurrent(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainTransmissionTravelledDistance is the Collection resolver for the powertrainTransmissionTravelledDistance field.
func (r *signalCollectionResolver) PowertrainTransmissionTravelledDistance(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// PowertrainType is the Collection resolver for the powertrainType field.
func (r *signalCollectionResolver) PowertrainType(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// Speed is the Collection resolver for the speed field.
func (r *signalCollectionResolver) Speed(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// VehicleIdentificationBrand is the Collection resolver for the vehicleIdentificationBrand field.
func (r *signalCollectionResolver) VehicleIdentificationBrand(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// VehicleIdentificationModel is the Collection resolver for the vehicleIdentificationModel field.
func (r *signalCollectionResolver) VehicleIdentificationModel(ctx context.Context, obj *model.Signals) (*model.SignalString, error) {
	strArgs := model.StringSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalString(ctx, &strArgs)
}

// VehicleIdentificationYear is the Collection resolver for the vehicleIdentificationYear field.
func (r *signalCollectionResolver) VehicleIdentificationYear(ctx context.Context, obj *model.Signals) (*model.SignalFloat, error) {
	floatArgs := model.FloatSignalArgs{
		Name:       graphql.GetFieldContext(ctx).Field.Name,
		SignalArgs: obj.SigArgs,
	}
	return r.GetLatestSignalFloat(ctx, &floatArgs)
}

// SignalAggregations returns SignalAggregationsResolver implementation.
func (r *Resolver) SignalAggregations() SignalAggregationsResolver {
	return &signalAggregationsResolver{r}
}

type signalAggregationsResolver struct{ *Resolver }