package graph

import (
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*repositories.Repository
}
