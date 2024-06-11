package repositories

import (
	"fmt"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
)

// ValidationError is an error type for validation errors.
type ValidationError string

func (v ValidationError) Error() string { return "invalid argument: " + string(v) }

func validateAggSigArgs(args *model.AggregatedSignalArgs) error {
	if args == nil {
		return ValidationError("aggregated signal args not provided")
	}

	if args.FromTS.IsZero() {
		return ValidationError("from timestamp is zero")
	}
	if args.ToTS.IsZero() {
		return ValidationError("to timestamp is zero")
	}
	if args.FromTS.After(args.ToTS) {
		return ValidationError("from timestamp is after to timestamp")
	}

	if args.Interval < 1 {
		return ValidationError("interval is not a positive integer")
	}
	return validateSignalArgs(&args.SignalArgs)
}

func validateLatestSigArgs(args *model.LatestSignalsArgs) error {
	if args == nil {
		return ValidationError("latest signal args not provided")
	}
	return validateSignalArgs(&args.SignalArgs)
}

func validateSignalArgs(args *model.SignalArgs) error {
	if args == nil {
		return ValidationError("signal args not provided")
	}

	if args.TokenID < 1 {
		return ValidationError("tokenID is not a positive integer")
	}

	return validateFilter(args.Filter)
}

func validateFilter(filter *model.SignalFilter) error {
	if filter == nil {
		return nil
	}
	// TODO: remove this check when we move to storing the device address as source
	if filter.Source != nil {
		if _, ok := ch.SourceTranslations[*filter.Source]; !ok {
			return ValidationError(fmt.Sprintf("source '%s', is not a valid value", *filter.Source))
		}
	}
	return nil
}
