// Code generated by "model-garage" DO NOT EDIT.
package model

import "github.com/DIMO-Network/model-garage/pkg/vss"

// SetCollectionField find the matching field based on the signal name and set the value based on the signal value.
func SetCollectionField(collection *SignalCollection, signal *vss.Signal) {
	if collection == nil || signal == nil {
		return
	}
	switch signal.Name {
	case "angularVelocityYaw":
		collection.AngularVelocityYaw = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow1DriverSideIsOpen":
		collection.CabinDoorRow1DriverSideIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow1DriverSideWindowIsOpen":
		collection.CabinDoorRow1DriverSideWindowIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow1PassengerSideIsOpen":
		collection.CabinDoorRow1PassengerSideIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow1PassengerSideWindowIsOpen":
		collection.CabinDoorRow1PassengerSideWindowIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow2DriverSideIsOpen":
		collection.CabinDoorRow2DriverSideIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow2DriverSideWindowIsOpen":
		collection.CabinDoorRow2DriverSideWindowIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow2PassengerSideIsOpen":
		collection.CabinDoorRow2PassengerSideIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "cabinDoorRow2PassengerSideWindowIsOpen":
		collection.CabinDoorRow2PassengerSideWindowIsOpen = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "chassisAxleRow1WheelLeftSpeed":
		collection.ChassisAxleRow1WheelLeftSpeed = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "chassisAxleRow1WheelLeftTirePressure":
		collection.ChassisAxleRow1WheelLeftTirePressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "chassisAxleRow1WheelRightSpeed":
		collection.ChassisAxleRow1WheelRightSpeed = &SignalFloat{
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
	case "currentLocationHeading":
		collection.CurrentLocationHeading = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "currentLocationIsRedacted":
		collection.CurrentLocationIsRedacted = &SignalFloat{
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
	case "exteriorAirTemperature":
		collection.ExteriorAirTemperature = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "isIgnitionOn":
		collection.IsIgnitionOn = &SignalFloat{
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
	case "obdCommandedEGR":
		collection.OBDCommandedEGR = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdCommandedEVAP":
		collection.OBDCommandedEVAP = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdDTCList":
		collection.OBDDTCList = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "obdDistanceSinceDTCClear":
		collection.OBDDistanceSinceDTCClear = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdDistanceWithMIL":
		collection.OBDDistanceWithMIL = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdEngineLoad":
		collection.OBDEngineLoad = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdFuelPressure":
		collection.OBDFuelPressure = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdIntakeTemp":
		collection.OBDIntakeTemp = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdLongTermFuelTrim1":
		collection.OBDLongTermFuelTrim1 = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdMAP":
		collection.OBDMAP = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdO2WRSensor1Voltage":
		collection.OBDO2WRSensor1Voltage = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdO2WRSensor2Voltage":
		collection.OBDO2WRSensor2Voltage = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdRunTime":
		collection.OBDRunTime = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdShortTermFuelTrim1":
		collection.OBDShortTermFuelTrim1 = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "obdWarmupsSinceDTCClear":
		collection.OBDWarmupsSinceDTCClear = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineDieselExhaustFluidCapacity":
		collection.PowertrainCombustionEngineDieselExhaustFluidCapacity = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineDieselExhaustFluidLevel":
		collection.PowertrainCombustionEngineDieselExhaustFluidLevel = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineECT":
		collection.PowertrainCombustionEngineECT = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineEOP":
		collection.PowertrainCombustionEngineEOP = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineEOT":
		collection.PowertrainCombustionEngineEOT = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainCombustionEngineEngineOilLevel":
		collection.PowertrainCombustionEngineEngineOilLevel = &SignalString{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueString,
		}
	case "powertrainCombustionEngineEngineOilRelativeLevel":
		collection.PowertrainCombustionEngineEngineOilRelativeLevel = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
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
	case "powertrainCombustionEngineTorque":
		collection.PowertrainCombustionEngineTorque = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainFuelSystemAbsoluteLevel":
		collection.PowertrainFuelSystemAbsoluteLevel = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainFuelSystemRelativeLevel":
		collection.PowertrainFuelSystemRelativeLevel = &SignalFloat{
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
	case "powertrainTractionBatteryChargingAddedEnergy":
		collection.PowertrainTractionBatteryChargingAddedEnergy = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryChargingChargeCurrentAC":
		collection.PowertrainTractionBatteryChargingChargeCurrentAC = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryChargingChargeLimit":
		collection.PowertrainTractionBatteryChargingChargeLimit = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryChargingChargeVoltageUnknownType":
		collection.PowertrainTractionBatteryChargingChargeVoltageUnknownType = &SignalFloat{
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
	case "powertrainTractionBatteryCurrentVoltage":
		collection.PowertrainTractionBatteryCurrentVoltage = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryGrossCapacity":
		collection.PowertrainTractionBatteryGrossCapacity = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryRange":
		collection.PowertrainTractionBatteryRange = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryStateOfChargeCurrent":
		collection.PowertrainTractionBatteryStateOfChargeCurrent = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryStateOfChargeCurrentEnergy":
		collection.PowertrainTractionBatteryStateOfChargeCurrentEnergy = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTractionBatteryTemperatureAverage":
		collection.PowertrainTractionBatteryTemperatureAverage = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTransmissionCurrentGear":
		collection.PowertrainTransmissionCurrentGear = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "powertrainTransmissionTemperature":
		collection.PowertrainTransmissionTemperature = &SignalFloat{
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
	case "serviceDistanceToService":
		collection.ServiceDistanceToService = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	case "speed":
		collection.Speed = &SignalFloat{
			Timestamp: signal.Timestamp,
			Value:     signal.ValueNumber,
		}
	}
}

// SetAggregationField find the matching field based on the signal name and set the value based on the signal value.
func SetAggregationField(aggregations *SignalAggregations, signal *AggSignal) {
	if aggregations == nil || signal == nil {
		return
	}
	switch signal.Name {
	case "angularVelocityYaw":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow1DriverSideIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow1DriverSideWindowIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow1PassengerSideIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow1PassengerSideWindowIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow2DriverSideIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow2DriverSideWindowIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow2PassengerSideIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "cabinDoorRow2PassengerSideWindowIsOpen":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "chassisAxleRow1WheelLeftSpeed":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "chassisAxleRow1WheelLeftTirePressure":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "chassisAxleRow1WheelRightSpeed":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "chassisAxleRow1WheelRightTirePressure":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "chassisAxleRow2WheelLeftTirePressure":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "chassisAxleRow2WheelRightTirePressure":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "currentLocationAltitude":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "currentLocationHeading":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "currentLocationIsRedacted":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "currentLocationLatitude":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "currentLocationLongitude":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "dimoAftermarketHDOP":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "dimoAftermarketNSAT":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "dimoAftermarketSSID":
		aggregations.ValueStrings[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueString
	case "dimoAftermarketWPAState":
		aggregations.ValueStrings[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueString
	case "exteriorAirTemperature":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "isIgnitionOn":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "lowVoltageBatteryCurrentVoltage":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdBarometricPressure":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdCommandedEGR":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdCommandedEVAP":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdDTCList":
		aggregations.ValueStrings[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueString
	case "obdDistanceSinceDTCClear":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdDistanceWithMIL":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdEngineLoad":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdFuelPressure":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdIntakeTemp":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdLongTermFuelTrim1":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdMAP":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdO2WRSensor1Voltage":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdO2WRSensor2Voltage":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdRunTime":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdShortTermFuelTrim1":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "obdWarmupsSinceDTCClear":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineDieselExhaustFluidCapacity":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineDieselExhaustFluidLevel":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineECT":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineEOP":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineEOT":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineEngineOilLevel":
		aggregations.ValueStrings[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueString
	case "powertrainCombustionEngineEngineOilRelativeLevel":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineMAF":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineSpeed":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineTPS":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainCombustionEngineTorque":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainFuelSystemAbsoluteLevel":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainFuelSystemRelativeLevel":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainFuelSystemSupportedFuelTypes":
		aggregations.ValueStrings[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueString
	case "powertrainRange":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryChargingAddedEnergy":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryChargingChargeCurrentAC":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryChargingChargeLimit":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryChargingChargeVoltageUnknownType":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryChargingIsCharging":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryCurrentPower":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryCurrentVoltage":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryGrossCapacity":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryRange":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryStateOfChargeCurrent":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryStateOfChargeCurrentEnergy":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTractionBatteryTemperatureAverage":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTransmissionCurrentGear":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTransmissionTemperature":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainTransmissionTravelledDistance":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "powertrainType":
		aggregations.ValueStrings[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueString
	case "serviceDistanceToService":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	case "speed":
		aggregations.ValueNumbers[AliasKey{Name: signal.Name, Agg: signal.Agg}] = signal.ValueNumber
	}
}
