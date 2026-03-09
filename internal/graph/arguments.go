package graph

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// resolveSubject resolves the subject from the provided tokenID or subject arguments.
// Exactly one of tokenID or subject must be provided.
func (r *queryResolver) resolveSubject(ctx context.Context, tokenID *int, subject *string) (string, error) {
	if subject != nil && tokenID != nil {
		return "", errorhandler.NewBadRequestError(ctx, errors.New("provide either tokenId or subject, not both"))
	}
	if subject != nil {
		return *subject, nil
	}
	if tokenID != nil {
		return r.BaseRepo.ToSubject(uint32(*tokenID)), nil
	}
	return "", errorhandler.NewBadRequestError(ctx, errors.New("tokenId or subject is required"))
}

// aggregationArgsFromContext creates an aggregated signals arguments from the context and the provided arguments.
func aggregationArgsFromContext(ctx context.Context, subject string, interval string, from time.Time, to time.Time, filter *model.SignalFilter) (*model.AggregatedSignalArgs, error) {
	intervalInt, err := getIntervalMicroseconds(interval)
	if err != nil {
		return nil, err
	}
	aggArgs := model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			Subject: subject,
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
		// TODO(elffjs): The casts here and in the location case are worrisome. Should we panic?
		filter, _ := child.Args["filter"].(*model.SignalFloatFilter)
		aggArgs.FloatArgs = append(aggArgs.FloatArgs, model.FloatSignalArgs{
			Name:   name,
			Agg:    typedAgg,
			Alias:  alias,
			Filter: filter,
		})
	case model.StringAggregation:
		aggArgs.StringArgs = append(aggArgs.StringArgs, model.StringSignalArgs{
			Name:  name,
			Agg:   typedAgg,
			Alias: alias,
		})
	case model.LocationAggregation:
		var filter *model.SignalLocationFilter
		dbSignalName := name
		if name == model.ApproximateCoordinatesField {
			dbSignalName = vss.FieldCurrentLocationCoordinates
		} else {
			filter, _ = child.Args["filter"].(*model.SignalLocationFilter)
		}
		aggArgs.LocationArgs = append(aggArgs.LocationArgs, model.LocationSignalArgs{
			Name:   dbSignalName,
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
func latestArgsFromContext(ctx context.Context, subject string, filter *model.SignalFilter) (*model.LatestSignalsArgs, error) {
	fields := graphql.CollectFieldsCtx(ctx, nil)
	latestArgs := model.LatestSignalsArgs{
		SignalArgs: model.SignalArgs{
			Subject: subject,
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

		if field.Name == model.ApproximateCoordinatesField {
			latestArgs.LocationSignalNames[vss.FieldCurrentLocationCoordinates] = struct{}{}
		} else if field.Definition.Type.Name() == "SignalLocation" {
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
