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

		// check for approximate location fields and force pull the latitude and longitude
		if field.Name == model.ApproximateLatField || field.Name == model.ApproximateLongField {
			if err := addSignalAggregation(&aggArgs, child, vss.FieldCurrentLocationLatitude); err != nil {
				return nil, err
			}
			if err := addSignalAggregation(&aggArgs, child, vss.FieldCurrentLocationLongitude); err != nil {
				return nil, err
			}
			continue
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
	switch typedAgg := agg.(type) {
	case model.FloatAggregation:
		aggArgs.FloatArgs = append(aggArgs.FloatArgs, model.FloatSignalArgs{
			Name: name,
			Agg:  typedAgg,
		})
	case model.StringAggregation:
		aggArgs.StringArgs = append(aggArgs.StringArgs, model.StringSignalArgs{
			Name: name,
			Agg:  typedAgg,
		})
	default:
		return fmt.Errorf("unknown aggregation type: %T", agg)
	}
	return nil
}

// latestArgsFromContext creates a latest signals arguments from the context and the provided arguments.
func latestArgsFromContext(ctx context.Context, tokenID int, filter *model.SignalFilter) (*model.LatestSignalsArgs, error) {
	latestArgs := model.LatestSignalsArgs{
		SignalArgs: model.SignalArgs{
			TokenID: uint32(tokenID),
			Filter:  filter,
		},
	}
	fields := graphql.CollectFieldsCtx(ctx, nil)
	for _, field := range fields {
		if !isSignal(field) {
			switch field.Name {
			case model.LastSeenField:
				latestArgs.IncludeLastSeen = true
			case model.ApproximateLatField:
				fallthrough
			case model.ApproximateLongField:
				latestArgs.SignalNames = append(latestArgs.SignalNames, vss.FieldCurrentLocationLatitude, vss.FieldCurrentLocationLongitude)
			}
			continue
		}
		latestArgs.SignalNames = append(latestArgs.SignalNames, field.Name)
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
