package graph

import (
	"github.com/DIMO-Network/model-garage/pkg/vss"
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
	lngVal, ok := obj.AppLocNumbers[model.AppLocKey{Aggregation: agg, Name: vss.FieldCurrentLocationLongitude}]
	if !ok {
		return nil, false
	}
	latVal, ok := obj.AppLocNumbers[model.AppLocKey{Aggregation: agg, Name: vss.FieldCurrentLocationLatitude}]
	if !ok {
		return nil, false
	}

	return repositories.GetApproximateLoc(latVal, lngVal), true
}
