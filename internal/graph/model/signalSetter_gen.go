// Code generated by "model-garage" DO NOT EDIT.
package model

import "github.com/DIMO-Network/model-garage/pkg/vss"

// SetCollectionField find the matching field based on the signal name and set the value based on the signal value.
func SetCollectionField(collection *SignalCollection, signal *vss.Signal) {
	if collection == nil || signal == nil {
		return
	}
	switch signal.Name {
	case "chassisAxleRow1WheelLeftTirePressure":
		collection.ChassisAxleRow1WheelLeftTirePressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "chassisAxleRow1WheelRightTirePressure":
		collection.ChassisAxleRow1WheelRightTirePressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "chassisAxleRow2WheelLeftTirePressure":
		collection.ChassisAxleRow2WheelLeftTirePressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "chassisAxleRow2WheelRightTirePressure":
		collection.ChassisAxleRow2WheelRightTirePressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "currentLocationAltitude":
		collection.CurrentLocationAltitude = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "currentLocationLatitude":
		collection.CurrentLocationLatitude = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "currentLocationLongitude":
		collection.CurrentLocationLongitude = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "currentLocationTimestamp":
		collection.CurrentLocationTimestamp = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "dimoAftermarketHDOP":
		collection.DIMOAftermarketHDOP = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "dimoAftermarketNSAT":
		collection.DIMOAftermarketNSAT = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "dimoAftermarketSSID":
		collection.DIMOAftermarketSSID = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "dimoAftermarketWPAState":
		collection.DIMOAftermarketWPAState = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "dimoIsLocationRedacted":
		collection.DIMOIsLocationRedacted = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "exteriorAirTemperature":
		collection.ExteriorAirTemperature = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "lowVoltageBatteryCurrentVoltage":
		collection.LowVoltageBatteryCurrentVoltage = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdBarometricPressure":
		collection.OBDBarometricPressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdEngineLoad":
		collection.OBDEngineLoad = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdIntakeTemp":
		collection.OBDIntakeTemp = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdRunTime":
		collection.OBDRunTime = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineECT":
		collection.PowertrainCombustionEngineECT = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineEngineOilLevel":
		collection.PowertrainCombustionEngineEngineOilLevel = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "powertrainCombustionEngineMAF":
		collection.PowertrainCombustionEngineMAF = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineSpeed":
		collection.PowertrainCombustionEngineSpeed = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineTPS":
		collection.PowertrainCombustionEngineTPS = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainFuelSystemAbsoluteLevel":
		collection.PowertrainFuelSystemAbsoluteLevel = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainFuelSystemSupportedFuelTypes":
		collection.PowertrainFuelSystemSupportedFuelTypes = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "powertrainRange":
		collection.PowertrainRange = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryChargingChargeLimit":
		collection.PowertrainTractionBatteryChargingChargeLimit = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryChargingIsCharging":
		collection.PowertrainTractionBatteryChargingIsCharging = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryCurrentPower":
		collection.PowertrainTractionBatteryCurrentPower = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryGrossCapacity":
		collection.PowertrainTractionBatteryGrossCapacity = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryStateOfChargeCurrent":
		collection.PowertrainTractionBatteryStateOfChargeCurrent = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTransmissionTravelledDistance":
		collection.PowertrainTransmissionTravelledDistance = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainType":
		collection.PowertrainType = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "speed":
		collection.Speed = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "vehicleIdentificationBrand":
		collection.VehicleIdentificationBrand = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "vehicleIdentificationModel":
		collection.VehicleIdentificationModel = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "vehicleIdentificationYear":
		collection.VehicleIdentificationYear = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	}
}

// SetAggregationField find the matching field based on the signal name and set the value based on the signal value.
func SetAggregationField(aggregations *SignalAggregations, signal *vss.Signal) {
	if aggregations == nil || signal == nil {
		return
	}
	switch signal.Name {
	case "chassisAxleRow1WheelLeftTirePressure":
		aggregations.ChassisAxleRow1WheelLeftTirePressure = &signal.ValueNumber
	case "chassisAxleRow1WheelRightTirePressure":
		aggregations.ChassisAxleRow1WheelRightTirePressure = &signal.ValueNumber
	case "chassisAxleRow2WheelLeftTirePressure":
		aggregations.ChassisAxleRow2WheelLeftTirePressure = &signal.ValueNumber
	case "chassisAxleRow2WheelRightTirePressure":
		aggregations.ChassisAxleRow2WheelRightTirePressure = &signal.ValueNumber
	case "currentLocationAltitude":
		aggregations.CurrentLocationAltitude = &signal.ValueNumber
	case "currentLocationLatitude":
		aggregations.CurrentLocationLatitude = &signal.ValueNumber
	case "currentLocationLongitude":
		aggregations.CurrentLocationLongitude = &signal.ValueNumber
	case "currentLocationTimestamp":
		aggregations.CurrentLocationTimestamp = &signal.ValueString
	case "dimoAftermarketHDOP":
		aggregations.DIMOAftermarketHDOP = &signal.ValueNumber
	case "dimoAftermarketNSAT":
		aggregations.DIMOAftermarketNSAT = &signal.ValueNumber
	case "dimoAftermarketSSID":
		aggregations.DIMOAftermarketSSID = &signal.ValueString
	case "dimoAftermarketWPAState":
		aggregations.DIMOAftermarketWPAState = &signal.ValueString
	case "dimoIsLocationRedacted":
		aggregations.DIMOIsLocationRedacted = &signal.ValueNumber
	case "exteriorAirTemperature":
		aggregations.ExteriorAirTemperature = &signal.ValueNumber
	case "lowVoltageBatteryCurrentVoltage":
		aggregations.LowVoltageBatteryCurrentVoltage = &signal.ValueNumber
	case "obdBarometricPressure":
		aggregations.OBDBarometricPressure = &signal.ValueNumber
	case "obdEngineLoad":
		aggregations.OBDEngineLoad = &signal.ValueNumber
	case "obdIntakeTemp":
		aggregations.OBDIntakeTemp = &signal.ValueNumber
	case "obdRunTime":
		aggregations.OBDRunTime = &signal.ValueNumber
	case "powertrainCombustionEngineECT":
		aggregations.PowertrainCombustionEngineECT = &signal.ValueNumber
	case "powertrainCombustionEngineEngineOilLevel":
		aggregations.PowertrainCombustionEngineEngineOilLevel = &signal.ValueString
	case "powertrainCombustionEngineMAF":
		aggregations.PowertrainCombustionEngineMAF = &signal.ValueNumber
	case "powertrainCombustionEngineSpeed":
		aggregations.PowertrainCombustionEngineSpeed = &signal.ValueNumber
	case "powertrainCombustionEngineTPS":
		aggregations.PowertrainCombustionEngineTPS = &signal.ValueNumber
	case "powertrainFuelSystemAbsoluteLevel":
		aggregations.PowertrainFuelSystemAbsoluteLevel = &signal.ValueNumber
	case "powertrainFuelSystemSupportedFuelTypes":
		aggregations.PowertrainFuelSystemSupportedFuelTypes = &signal.ValueString
	case "powertrainRange":
		aggregations.PowertrainRange = &signal.ValueNumber
	case "powertrainTractionBatteryChargingChargeLimit":
		aggregations.PowertrainTractionBatteryChargingChargeLimit = &signal.ValueNumber
	case "powertrainTractionBatteryChargingIsCharging":
		aggregations.PowertrainTractionBatteryChargingIsCharging = &signal.ValueNumber
	case "powertrainTractionBatteryCurrentPower":
		aggregations.PowertrainTractionBatteryCurrentPower = &signal.ValueNumber
	case "powertrainTractionBatteryGrossCapacity":
		aggregations.PowertrainTractionBatteryGrossCapacity = &signal.ValueNumber
	case "powertrainTractionBatteryStateOfChargeCurrent":
		aggregations.PowertrainTractionBatteryStateOfChargeCurrent = &signal.ValueNumber
	case "powertrainTransmissionTravelledDistance":
		aggregations.PowertrainTransmissionTravelledDistance = &signal.ValueNumber
	case "powertrainType":
		aggregations.PowertrainType = &signal.ValueString
	case "speed":
		aggregations.Speed = &signal.ValueNumber
	case "vehicleIdentificationBrand":
		aggregations.VehicleIdentificationBrand = &signal.ValueString
	case "vehicleIdentificationModel":
		aggregations.VehicleIdentificationModel = &signal.ValueString
	case "vehicleIdentificationYear":
		aggregations.VehicleIdentificationYear = &signal.ValueNumber
	}
}
