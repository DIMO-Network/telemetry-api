package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.56

import (
	"context"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// Signals is the resolver for the Signals field.
func (r *queryResolver) Signals(ctx context.Context, tokenID int, interval string, from time.Time, to time.Time, filter *model.SignalFilter) ([]*model.SignalAggregations, error) {
	aggArgs, err := aggregationArgsFromContext(ctx, tokenID, interval, from, to, filter)
	if err != nil {
		return nil, err
	}
	return r.Repository.GetSignal(ctx, aggArgs)
}

// SignalsLatest is the resolver for the SignalsLatest field.
func (r *queryResolver) SignalsLatest(ctx context.Context, tokenID int, filter *model.SignalFilter) (*model.SignalCollection, error) {
	latestArgs, err := latestArgsFromContext(ctx, tokenID, filter)
	if err != nil {
		return nil, err
	}
	return r.Repository.GetSignalLatest(ctx, latestArgs)
}

// AvailableSignals is the resolver for the AvailableSignals field.
func (r *queryResolver) AvailableSignals(ctx context.Context, tokenID int, filter *model.SignalFilter) ([]string, error) {
	return r.Repository.GetAvailableSignals(ctx, uint32(tokenID), filter)
}

// CurrentLocationApproximateLatitude is the resolver for the CurrentLocationApproximateLatitude
func (r *signalAggregationsResolver) CurrentLocationApproximateLatitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	latLong := approximateLocationSignalAggregations(obj, agg)
	if latLong == nil {
		return nil, nil
	}
	return &latLong.Lat, nil
}

// CurrentLocationApproximateLongitude is the resolver for the CurrentLocationApproximateLongitude
func (r *signalAggregationsResolver) CurrentLocationApproximateLongitude(ctx context.Context, obj *model.SignalAggregations, agg model.FloatAggregation) (*float64, error) {
	latLong := approximateLocationSignalAggregations(obj, agg)
	if latLong == nil {
		return nil, nil
	}
	return &latLong.Lng, nil
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// SignalAggregations returns SignalAggregationsResolver implementation.
func (r *Resolver) SignalAggregations() SignalAggregationsResolver {
	return &signalAggregationsResolver{r}
}

type queryResolver struct{ *Resolver }
type signalAggregationsResolver struct{ *Resolver }
