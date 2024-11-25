package graph

import (
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/uber/h3-go/v4"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*repositories.Repository
	IdentityService *identity.APIClient
	VCRepo          *vc.Repository
}

func approximateLocationSignalAggregations(obj *model.SignalAggregations, agg model.FloatAggregation) *h3.LatLng {
	latVal, ok := obj.ValueNumbers[model.AliasKey{Name: vss.FieldCurrentLocationLatitude, Agg: agg.String()}]
	if !ok {
		return nil
	}
	lngVal, ok := obj.ValueNumbers[model.AliasKey{Name: vss.FieldCurrentLocationLongitude, Agg: agg.String()}]
	if !ok {
		return nil
	}
	return repositories.GetApproximateLoc(latVal, lngVal)
}
