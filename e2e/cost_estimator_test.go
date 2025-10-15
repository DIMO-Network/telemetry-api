package e2e_test

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/stretchr/testify/require"
)

func TestEstimateCost(t *testing.T) {
	services := GetTestServices(t)
	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, 39718, []string{tokenclaims.PermissionGetLocationHistory})

	// Execute the query
	query := `
	query Latest_all {
		smartcar: signalsLatest(filter: {source: "smartcar"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
			s1: speed {
				timestamp
				value
			}
			dimoAftermarketWPAState {
				timestamp
				value
			}
			currentLocationHeading {
				timestamp
				value
			}
			currentLocationIsRedacted {
				timestamp
				value
			}
			loc2: currentLocationIsRedacted {
				timestamp
				value
			}
				

		}
	
	}`

	// Execute request
	_, err := telemetryClient.RawPost(query, WithToken(token), WithEstimateCost())
	require.NoError(t, err)
}

func WithEstimateCost() client.Option {
	return client.AddHeader("X-Estimate-Cost", "true")
}
