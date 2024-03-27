package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

func GetPreloads(ctx context.Context) []string {
	return GetNestedPreloads(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

func GetNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, parent string) []string {
	var preloads []string
	for _, column := range fields {
		if parent == "nodes" || parent == "node" {
			preloads = append(preloads, column.Name)
		}
		preloads = append(preloads, GetNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), column.Name)...)
	}
	return preloads
}
