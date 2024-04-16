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

func getFloatArgs(ctx context.Context, obj *model.SignalCollection, agg *model.FloatAggregation) (repositories.FloatSignalArgs, error) {
	if agg == nil {
		return repositories.FloatSignalArgs{}, fmt.Errorf("aggregation type is nil")
	}
	args, err := getSignalArgs(ctx, obj)
	if err != nil {
		return repositories.FloatSignalArgs{}, err
	}
	return repositories.FloatSignalArgs{
		Agg:        *agg,
		SignalArgs: args,
	}, nil
}

func getStringArgs(ctx context.Context, obj *model.SignalCollection, agg *model.StringAggregation) (repositories.StringSignalArgs, error) {
	if agg == nil {
		return repositories.StringSignalArgs{}, fmt.Errorf("aggregation type is nil")
	}
	args, err := getSignalArgs(ctx, obj)
	if err != nil {
		return repositories.StringSignalArgs{}, err
	}
	return repositories.StringSignalArgs{
		Agg:        *agg,
		SignalArgs: args,
	}, nil
}

// getFloatSignalArgs returns the arguments for the float signal query.
func getSignalArgs(ctx context.Context, obj *model.SignalCollection) (repositories.SignalArgs, error) {
	var args repositories.SignalArgs
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return args, fmt.Errorf("no field context found")
	}
	name := fCtx.Field.Name
	var ok bool
	args.Name, ok = vss.DimoJSONName2CHName[name]
	if !ok {
		return args, fmt.Errorf("field %s not found", name)
	}
	var err error
	args.FromTS, err = getTimeArg(fCtx.Parent.Args, "from")
	if err != nil {
		return args, err
	}
	args.ToTS, err = getTimeArg(fCtx.Parent.Args, "to")
	if err != nil {
		return args, err
	}
	args.TokenID = obj.TokenID
	return args, nil
}

func getTimeArg(args map[string]any, name string) (time.Time, error) {
	val, ok := args[name]
	if !ok {
		return time.Time{}, fmt.Errorf("'%s' argument not found", name)
	}
	timeArg, ok := val.(*time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("'%s' argument is not a time.Time", name)
	}
	if timeArg == nil {
		return time.Time{}, fmt.Errorf("'%s' argument is nil", name)
	}
	return *timeArg, nil
}
