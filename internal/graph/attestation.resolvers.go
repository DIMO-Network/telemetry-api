package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// Attestations is the resolver for the attestations field.
func (r *queryResolver) Attestations(ctx context.Context, tokenID int, signer *string) ([]*model.Attestation, error) {
	return r.AttestationRepo.GetAttestations(ctx, uint32(tokenID), signer)
}
