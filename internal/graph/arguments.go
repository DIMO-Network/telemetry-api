package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// aggregationArgsFromContext creates an aggregated signals arguments from the context and the provided arguments.
func aggregationArgsFromContext(ctx context.Context, tokenID int, interval string, from time.Time, to time.Time, filter *model.SignalFilter) (*model.AggregatedSignalArgs, error) {
	// 1h 1s
	intervalInt, err := getIntervalMS(interval)
	if err != nil {
		return nil, err
	}
	aggArgs := model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
		FromTS:     from,
		ToTS:       to,
		Interval:   intervalInt,
		FloatArgs:  make(map[string]model.FloatSignalArgs),
		StringArgs: make(map[string]model.StringSignalArgs),
	}

	fields := graphql.CollectFieldsCtx(ctx, nil)
	parentCtx := graphql.GetFieldContext(ctx)

	floatIndex := 0
	stringIndex := 0

	for _, field := range fields {
		if !isSignal(field) || !hasAggregations(field) {
			continue
		}
		child, err := parentCtx.Child(ctx, field)
		if err != nil {
			return nil, fmt.Errorf("failed to get child field: %w", err)
		}

		agg := child.Args["agg"]
		alias := child.Field.Alias
		switch typedAgg := agg.(type) {
		case model.FloatAggregation:
			handle := fmt.Sprintf("float%d", floatIndex)
			filter := child.Args["filter"].(*model.SignalFloatFilter)
			aggArgs.FloatArgs[child.Field.Alias] = model.FloatSignalArgs{
				Name:        child.Field.Name,
				Agg:         typedAgg,
				Filter:      filter,
				QueryHandle: handle,
			}
			aggArgs.AliasToHandle[alias] = handle
			floatIndex++
		case model.StringAggregation:
			handle := fmt.Sprintf("string%d", floatIndex)
			aggArgs.StringArgs[child.Field.Alias] = model.StringSignalArgs{
				Name:        child.Field.Name,
				Agg:         typedAgg,
				QueryHandle: handle,
			}
			aggArgs.AliasToHandle[alias] = handle
			stringIndex++
		default:
			return nil, fmt.Errorf("unknown aggregation type: %T", agg)
		}
	}
	return &aggArgs, nil
}

// latestArgsFromContext creates a latest signals arguments from the context and the provided arguments.
func latestArgsFromContext(ctx context.Context, tokenID int, filter *model.SignalFilter) (*model.LatestSignalsArgs, error) {
	fields := graphql.CollectFieldsCtx(ctx, nil)
	latestArgs := model.LatestSignalsArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
		SignalNames: make(map[string]struct{}, len(fields)),
	}
	for _, field := range fields {
		if !isSignal(field) {
			if field.Name == model.LastSeenField {
				latestArgs.IncludeLastSeen = true
			}
			continue
		}
		if field.Name == model.ApproximateLatField || field.Name == model.ApproximateLongField {
			latestArgs.SignalNames[vss.FieldCurrentLocationLatitude] = struct{}{}
			latestArgs.SignalNames[vss.FieldCurrentLocationLongitude] = struct{}{}
			continue
		}
		latestArgs.SignalNames[field.Name] = struct{}{}
	}
	return &latestArgs, nil
}

// getIntervalMS parses the interval string and returns the milliseconds.
func getIntervalMS(interval string) (int64, error) {
	dur, err := time.ParseDuration(interval)
	if err != nil {
		return 0, fmt.Errorf("failed parsing interval: %w", err)
	}
	return dur.Milliseconds(), nil
}

// isSignal checks if the field has the isSignal directive.
func isSignal(field graphql.CollectedField) bool {
	return field.Definition.Directives.ForName("isSignal") != nil
}

// hasAggregations checks if the field has the hasAggregation directive.
func hasAggregations(field graphql.CollectedField) bool {
	return field.Definition.Directives.ForName("hasAggregation") != nil
}
