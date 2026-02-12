package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// signalLocationType is the name of the GraphQL SignalLocation type.
const signalLocationType = "SignalLocation"

// aggregationArgsFromContext creates an aggregated signals arguments from the context and the provided arguments.
func aggregationArgsFromContext(ctx context.Context, tokenID int, interval string, from time.Time, to time.Time, filter *model.SignalFilter) (*model.AggregatedSignalArgs, error) {
	// 1h 1s
	intervalInt, err := getIntervalMicroseconds(interval)
	if err != nil {
		return nil, err
	}
	aggArgs := model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
		FromTS:        from,
		ToTS:          to,
		Interval:      intervalInt,
		ApproxLocArgs: make(map[model.FloatAggregation]struct{}),
	}

	fields := graphql.CollectFieldsCtx(ctx, nil)
	parentCtx := graphql.GetFieldContext(ctx)
	for _, field := range fields {
		if !isSignal(field) || !hasAggregations(field) {
			continue
		}
		child, err := parentCtx.Child(ctx, field)
		if err != nil {
			return nil, fmt.Errorf("failed to get child field: %w", err)
		}

		if err := addSignalAggregation(&aggArgs, child, child.Field.Name); err != nil {
			return nil, err
		}
	}
	return &aggArgs, nil
}

// addSignalAggregation gets the aggregation arguments from the child field and adds them to the aggregated signal arguments as eiter a float or string aggregation.
func addSignalAggregation(aggArgs *model.AggregatedSignalArgs, child *graphql.FieldContext, name string) error {
	agg := child.Args["agg"]
	alias := child.Field.Alias
	switch typedAgg := agg.(type) {
	case model.FloatAggregation:
		if name == model.ApproximateLongField || name == model.ApproximateLatField {
			aggArgs.ApproxLocArgs[typedAgg] = struct{}{}
		} else {
			filter, _ := child.Args["filter"].(*model.SignalFloatFilter)
			aggArgs.FloatArgs = append(aggArgs.FloatArgs, model.FloatSignalArgs{
				Name:   name,
				Agg:    typedAgg,
				Alias:  alias,
				Filter: filter,
			})
		}
	case model.StringAggregation:
		aggArgs.StringArgs = append(aggArgs.StringArgs, model.StringSignalArgs{
			Name:  name,
			Agg:   typedAgg,
			Alias: alias,
		})
	case model.LocationAggregation:
		filter, _ := child.Args["filter"].(*model.SignalLocationFilter)
		aggArgs.LocationArgs = append(aggArgs.LocationArgs, model.LocationSignalArgs{
			Name:   name,
			Agg:    typedAgg,
			Alias:  alias,
			Filter: filter,
		})
	default:
		return fmt.Errorf("unknown aggregation type: %T", agg)
	}
	return nil
}

// latestArgsFromContext creates a latest signals arguments from the context and the provided arguments.
func latestArgsFromContext(ctx context.Context, tokenID int, filter *model.SignalFilter) (*model.LatestSignalsArgs, error) {
	fields := graphql.CollectFieldsCtx(ctx, nil)
	latestArgs := model.LatestSignalsArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
		SignalNames:         make(map[string]struct{}),
		LocationSignalNames: make(map[string]struct{}),
	}
	for _, field := range fields {
		if !isSignal(field) {
			if field.Name == model.LastSeenField {
				latestArgs.IncludeLastSeen = true
			}
			continue
		}

		if field.Name == model.ApproximateLatField || field.Name == model.ApproximateLongField || field.Name == vss.FieldCurrentLocationLatitude || field.Name == vss.FieldCurrentLocationLongitude {
			latestArgs.LocationSignalNames[vss.FieldCurrentLocationCoordinates] = struct{}{}
		} else if field.Definition.Type.Name() == signalLocationType {
			latestArgs.LocationSignalNames[field.Name] = struct{}{}
		} else {
			latestArgs.SignalNames[field.Name] = struct{}{}
		}
	}
	return &latestArgs, nil
}

// getIntervalMicroseconds parses the interval string and returns the number
// of microseconds the interval contains.
//
// We use microseconds because the ClickHouse column is DateTime64(6, 'UTC').
// Go stores durations in nanoseconds, so we do lose some precision.
func getIntervalMicroseconds(interval string) (int64, error) {
	dur, err := time.ParseDuration(interval)
	if err != nil {
		return 0, fmt.Errorf("failed parsing interval: %w", err)
	}
	return dur.Microseconds(), nil
}

// isSignal checks if the field has the isSignal directive.
func isSignal(field graphql.CollectedField) bool {
	return field.Definition.Directives.ForName("isSignal") != nil
}

// hasAggregations checks if the field has the hasAggregation directive.
func hasAggregations(field graphql.CollectedField) bool {
	return field.Definition.Directives.ForName("hasAggregation") != nil
}
