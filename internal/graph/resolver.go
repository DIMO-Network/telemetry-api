package graph

import (
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/uber/h3-go/v4"
)

const (
	approximateLocationResolution = 6
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*repositories.Repository
	IdentityService *identity.APIClient
	VCRepo          *vc.Repository
}

func (r *signalAggregationsResolver) aproximateLocationSignalAggregations(obj *model.SignalAggregations, agg model.FloatAggregation) (*h3.LatLng, error) {
	latVal, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationLatitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	logVal, ok := obj.ValueNumbers[model.AliasKey{Name: "currentLocationLongitude", Agg: agg.String()}]
	if !ok {
		return nil, nil
	}
	h3LatLng := h3.NewLatLng(latVal, logVal)
	cell := h3.LatLngToCell(h3LatLng, approximateLocationResolution)
	latLong := h3.CellToLatLng(cell)
	return &latLong, nil
}
