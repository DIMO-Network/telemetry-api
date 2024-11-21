package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// DeviceActivity is the resolver for the deviceActivity field.
func (r *queryResolver) DeviceActivity(ctx context.Context, by model.AftermarketDeviceBy) (*model.DeviceActivity, error) {
	adInfo, err := r.IdentityService.GetAftermarketDevice(ctx, by.Address, by.TokenID, by.Serial)
	if err != nil {
		return nil, err
	}

	return r.Repository.GetDeviceActivity(ctx, adInfo.VehicleTokenID, adInfo.ManufacturerName)
}
