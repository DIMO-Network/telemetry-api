package graph

//go:generate go run github.com/DIMO-Network/server-garage/cmd/mcpgen -schema ../../schema/ -prefix telemetry -out mcp_tools_gen.go -package graph

import (
	"github.com/DIMO-Network/telemetry-api/internal/proxy"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	BaseRepo        *repositories.Repository
	VCRepo          *vc.Repository
	AttestationRepo *attestation.Repository
	// ProxyClient, when non-nil, forwards signal/event/segment queries to dq instead of ClickHouse.
	ProxyClient *proxy.Client
	// ProxySubjectFunc converts a tokenID to a DID subject for the proxy.
	ProxySubjectFunc func(tokenID int) string
}
