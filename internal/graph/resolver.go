package graph

import (
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vinvc"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity_api"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*repositories.Repository
	VINVCRepo       *vinvc.Repository
	IdentityService identity_api.IdentityService
}
