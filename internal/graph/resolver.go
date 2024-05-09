package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*repositories.Repository
}

func getFloatArgs(ctx context.Context, obj *model.SignalsWithID, agg model.FloatAggregation) (*repositories.FloatSignalArgs, error) {
	args, err := getSignalArgs(ctx, obj)
	if err != nil {
		return nil, err
	}
	return &repositories.FloatSignalArgs{
		Agg:        agg,
		SignalArgs: *args,
	}, nil
}

func getStringArgs(ctx context.Context, obj *model.SignalsWithID, agg model.StringAggregation) (*repositories.StringSignalArgs, error) {
	args, err := getSignalArgs(ctx, obj)
	if err != nil {
		return nil, err
	}
	return &repositories.StringSignalArgs{
		Agg:        agg,
		SignalArgs: *args,
	}, nil
}

// getFloatSignalArgs returns the arguments for the float signal query.
func getSignalArgs(ctx context.Context, obj *model.SignalsWithID) (*repositories.SignalArgs, error) {
	args := &repositories.SignalArgs{}
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return args, fmt.Errorf("no field context found")
	}
	name := fCtx.Field.Name
	var ok bool
	args.Name, ok = vss.JSONName2CHName[name]
	if !ok {
		return args, fmt.Errorf("field %s not found", name)
	}

	args.FromTS = getTimeArg(fCtx.Parent.Args, "from")
	args.ToTS = getTimeArg(fCtx.Parent.Args, "to")
	args.Filter = getFilterArg(fCtx.Parent.Args)
	args.TokenID = obj.TokenID

	return args, nil
}

func getTimeArg(args map[string]any, name string) time.Time {
	val, ok := args[name]
	if !ok {
		return time.Time{}
	}
	timeArg, ok := val.(time.Time)
	if !ok {
		return time.Time{}
	}
	return timeArg
}

func getFilterArg(args map[string]any) *model.SignalFilter {
	filterArg, ok := args["filter"]
	if !ok {
		return nil
	}
	filter, ok := filterArg.(*model.SignalFilter)
	if !ok {
		return nil
	}
	return filter
}
