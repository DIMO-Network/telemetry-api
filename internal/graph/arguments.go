package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// getIntervalMS parses the interval string and returns the milliseconds.
func getIntervalMS(interval string) (int64, error) {
	dur, err := time.ParseDuration(interval)
	if err != nil {
		return 0, fmt.Errorf("failed parsing interval: %w", err)
	}
	return dur.Milliseconds(), nil
}

// AgregtaionArgsFromContext creates an aggregated signals arguments from the context and the provided arguments.
func AgregtaionArgsFromContext(ctx context.Context, tokenID int, interval string, from time.Time, to time.Time, filter *model.SignalFilter) (*model.AggregatedSignalArgs, error) {
	intervalInt, err := getIntervalMS(interval)
	if err != nil {
		return nil, err
	}
	aggArgs := model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
		FromTS:   from,
		ToTS:     to,
		Interval: intervalInt,
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
		if err := addSignalAggregation(&aggArgs, child); err != nil {
			return nil, err
		}
	}
	return &aggArgs, nil
}

// addSignalAggregation gets the aggregation arguments from the child field and adds them to the aggregated signal arguments as eiter a float or string aggregation.
func addSignalAggregation(aggArgs *model.AggregatedSignalArgs, child *graphql.FieldContext) error {
	agg := child.Args["agg"]
	switch typedAgg := agg.(type) {
	case model.FloatAggregation:
		aggArgs.FloatArgs = append(aggArgs.FloatArgs, model.FloatSignalArgs{
			Name: child.Field.Name,
			Agg:  typedAgg,
		})
	case model.StringAggregation:
		aggArgs.StringArgs = append(aggArgs.StringArgs, model.StringSignalArgs{
			Name: child.Field.Name,
			Agg:  typedAgg,
		})
	default:
		return fmt.Errorf("unknown aggregation type: %T", agg)
	}
	return nil
}

// LatestArgsFromContext creates a latest signals arguments from the context and the provided arguments.
func LatestArgsFromContext(ctx context.Context, tokenID int, filter *model.SignalFilter) (*model.LatestSignalsArgs, error) {
	latestArgs := model.LatestSignalsArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
	}
	fields := graphql.CollectFieldsCtx(ctx, nil)
	for _, field := range fields {
		if !isSignal(field) {
			if field.Name == model.LastSeenField {
				latestArgs.IncludeLastSeen = true
			}
			continue
		}

		latestArgs.SignalNames = append(latestArgs.SignalNames, field.Name)
	}
	return &latestArgs, nil
}

// isSignal checks if the field has the isSignal directive.
func isSignal(field graphql.CollectedField) bool {
	return field.Definition.Directives.ForName("isSignal") != nil
}

// hasAggregations checks if the field has the hasAggregation directive.
func hasAggregations(field graphql.CollectedField) bool {
	return field.Definition.Directives.ForName("hasAggregation") != nil
}
