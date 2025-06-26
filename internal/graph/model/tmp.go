package model

func SetAggregationField(aggregations *SignalAggregations, signal *AggSignal) {
	if _, ok := aggregations.SignalArgs.FloatArgs[signal.Handle]; ok {
		aggregations.ValueNumbers[signal.Handle] = signal.ValueNumber
		return
	}
	if _, ok := aggregations.SignalArgs.FloatArgs[signal.Handle]; ok {
		aggregations.ValueStrings[signal.Handle] = signal.ValueString
		return
	}
}
