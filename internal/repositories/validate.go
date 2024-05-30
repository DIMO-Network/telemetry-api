package repositories

import (
	"fmt"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
)

var twoWeeks = 14 * 24 * time.Hour

// ValidationError is an error type for validation errors.
type ValidationError string

func (v ValidationError) Error() string { return "invalid argument: " + string(v) }

func validateAggSigArgs(args *model.AggregatedSignalArgs) error {
	if args.FromTS.IsZero() {
		return ValidationError("from timestamp is zero")
	}
	if args.ToTS.IsZero() {
		return ValidationError("to timestamp is zero")
	}
	if args.FromTS.After(args.ToTS) {
		return ValidationError("from timestamp is after to timestamp")
	}

	// check if time range is greater than 2 weeks
	if args.ToTS.Sub(args.FromTS) > twoWeeks {
		return ValidationError("time range is greater than two weeks")
	}

	if args.Interval < 1 {
		return ValidationError("interval is not a positive integer")
	}
	return validateSignalArgs(&args.SignalArgs)
}

func validateSignalArgs(args *model.SignalArgs) error {
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
