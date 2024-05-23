package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
)

var (
	twoWeeks       = 14 * 24 * time.Hour
	errInvalidArgs = errors.New("invalid arguments")
)

func validateAggSigArgs(args *model.AggregatedSignalArgs) error {
	if args.FromTS.IsZero() {
		return fmt.Errorf("%w from timestamp is zero", errInvalidArgs)
	}
	if args.ToTS.IsZero() {
		return fmt.Errorf("%w to timestamp is zero", errInvalidArgs)
	}

	// check if time range is greater than 2 weeks
	if args.ToTS.Sub(args.FromTS) > twoWeeks {
		return fmt.Errorf("%w time range is greater than 2 weeks", errInvalidArgs)
	}

	if time.Duration(args.Interval) < 1 {
		return fmt.Errorf("%w interval less than 1 millisecond are not supported", errInvalidArgs)
	}
	return validateSignalArgs(&args.SignalArgs)
}

func validateSignalArgs(args *model.SignalArgs) error {
	if args.TokenID < 1 {
		return fmt.Errorf("%w tokenId is not a non-zero uint32", errInvalidArgs)
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
			return fmt.Errorf("%w source '%s', is not a valid value", errInvalidArgs, *filter.Source)
		}
	}
	return nil
}
