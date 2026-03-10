package repositories

import (
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/schema"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// we must load these tags before starting the application
var eventTags = func() map[string]struct{} {
	filter, err := schema.GetDefaultEventNames()
	if err != nil {
		panic(err)
	}
	tags := make(map[string]struct{}, len(filter))
	for _, tag := range filter {
		tags[tag.Name] = struct{}{}
	}
	return tags
}()

// eventNamePattern matches exactly 2 dotted segments, e.g. "behavior.harshBraking".
var eventNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*\.[a-zA-Z][a-zA-Z0-9]*$`)

// eventNamePrefixPattern matches a category with optional name segment, e.g. "behavior." or "behavior.harsh".
var eventNamePrefixPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*\.([a-zA-Z][a-zA-Z0-9]*)?$`)

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

	if len(args.FloatArgs) > math.MaxUint16 {
		return ValidationError("too many float aggregations")
	}
	if len(args.StringArgs) > math.MaxUint16 {
		return ValidationError("too many string aggregations")
	}
	if len(args.LocationArgs) > math.MaxUint16 {
		return ValidationError("too many location aggregations")
	}

	// TODO(elffjs): Awkward place to put this. Certainly this would get
	// worse if we allowed ORs.
	for _, locArg := range args.LocationArgs {
		if fil := locArg.Filter; fil != nil {
			// TODO(elffjs): Should we check polygon orientation?
			// Could apply isFilterLocationValid here, but failure there
			// doesn't actually break queries.
			if len(fil.InPolygon) != 0 && len(fil.InPolygon) < 3 {
				return ValidationError("not enough points in geofence filter")
			}

			if fil.InCircle != nil {
				if !isFilterLocationValid(fil.InCircle.Center) {
					return ValidationError("invalid circle filter location")
				}
				// Could think about checking for radius < 0, but nothing
				// actually breaks.
			}
		}
	}

	return validateSignalArgs(&args.SignalArgs)
}

func isFilterLocationValid(loc *model.FilterLocation) bool {
	return -90 <= loc.Latitude && loc.Latitude <= 90 && -180 <= loc.Longitude && loc.Longitude <= 180
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
	if filter.Source != nil {
		if _, err := cloudevent.DecodeEthrDID(*filter.Source); err != nil {
			return ValidationError(fmt.Sprintf("source '%s' is not a valid ethr DID", *filter.Source))
		}
	}
	return nil
}

func validateEventArgs(tokenID int, from, to time.Time, filter *model.EventFilter) error {
	if tokenID < 1 {
		return ValidationError("tokenID is not a positive integer")
	}
	if from.IsZero() {
		return ValidationError("from timestamp is zero")
	}
	if to.IsZero() {
		return ValidationError("to timestamp is zero")
	}
	if from.After(to) {
		return ValidationError("from timestamp is after to timestamp")
	}
	if filter != nil {
		if err := validateEventNameFilter(filter.Name); err != nil {
			return err
		}
		if err := validateTags(filter.Tags); err != nil {
			return err
		}
	}
	return nil
}

func validateEventNameFilter(filter *model.StringValueFilter) error {
	if filter == nil {
		return nil
	}
	if filter.Eq != nil {
		if !eventNamePattern.MatchString(*filter.Eq) {
			return ValidationError(fmt.Sprintf("event name '%s' does not match namespace pattern (e.g. 'behavior.harshBraking')", *filter.Eq))
		}
	}
	for _, name := range filter.In {
		if !eventNamePattern.MatchString(name) {
			return ValidationError(fmt.Sprintf("event name '%s' does not match namespace pattern (e.g. 'behavior.harshBraking')", name))
		}
	}
	if filter.StartsWith != nil {
		if !eventNamePrefixPattern.MatchString(*filter.StartsWith) {
			return ValidationError(fmt.Sprintf("event name prefix '%s' does not match namespace pattern (e.g. 'behavior.' or 'behavior.harsh')", *filter.StartsWith))
		}
	}
	for _, orFilter := range filter.Or {
		if err := validateEventNameFilter(orFilter); err != nil {
			return err
		}
	}
	return nil
}

func validateTags(stringArrayFilter *model.StringArrayFilter) error {
	if stringArrayFilter == nil {
		return nil
	}
	for _, tag := range stringArrayFilter.ContainsAll {
		if _, ok := eventTags[tag]; !ok {
			return ValidationError(fmt.Sprintf("tag '%s', is not a valid value", tag))
		}
	}
	for _, tag := range stringArrayFilter.ContainsAny {
		if _, ok := eventTags[tag]; !ok {
			return ValidationError(fmt.Sprintf("tag '%s', is not a valid value", tag))
		}
	}
	for _, tag := range stringArrayFilter.NotContainsAny {
		if _, ok := eventTags[tag]; !ok {
			return ValidationError(fmt.Sprintf("tag '%s', is not a valid value", tag))
		}
	}
	for _, tag := range stringArrayFilter.NotContainsAll {
		if _, ok := eventTags[tag]; !ok {
			return ValidationError(fmt.Sprintf("tag '%s', is not a valid value", tag))
		}
	}
	for _, tag := range stringArrayFilter.Or {
		if err := validateTags(tag); err != nil {
			return err
		}
	}
	return nil
}
