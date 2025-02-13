package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// VinVCLatest is the resolver for the vinVCLatest field.
func (r *queryResolver) VinVCLatest(ctx context.Context, tokenID int) (*model.Vinvc, error) {
	return r.VCRepo.GetLatestVINVC(ctx, uint32(tokenID))
}

// PomVCLatest is the resolver for the pomVCLatest field.
func (r *queryResolver) PomVCLatest(ctx context.Context, tokenID int) (*model.Pomvc, error) {
	return r.VCRepo.GetLatestPOMVC(ctx, uint32(tokenID))
}
