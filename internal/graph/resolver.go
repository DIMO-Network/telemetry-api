package graph

import (
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
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
	AttestationRepo *attestation.Repository
}

func approximateLocationSignalAggregations(obj *model.SignalAggregations, agg model.FloatAggregation) (*h3.LatLng, bool) {
	return nil, false
	// latVal, ok := obj.ValueNumbers[model.AliasKey{Name: vss.FieldCurrentLocationLatitude, Agg: agg.String()}]
	// if !ok {
	// 	return nil, false
	// }
	// lngVal, ok := obj.ValueNumbers[model.AliasKey{Name: vss.FieldCurrentLocationLongitude, Agg: agg.String()}]
	// if !ok {
	// 	return nil, false
	// }
	// return repositories.GetApproximateLoc(latVal, lngVal), true
}
