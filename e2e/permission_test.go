package e2e_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPermission(t *testing.T) {
	services := GetTestServices(t)
	const unusedTokenID = 999999
	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	tests := []struct {
		name        string
		tokenID     int
		query       string
		permissions []int
		expectedErr string // checking error strings since that is what the server returns
	}{
		{
			name:    "No permissions",
			tokenID: unusedTokenID,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
				}
			}`,
			permissions: []int{},
			expectedErr: "unauthorized: token id does not match",
		},
		{
			name:    "Token permissions",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
				}
			}`,
			permissions: []int{},
		},
		{
			name:    "Partial permissions",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
					speed {
						value
					}
				}
			}`,
			permissions: []int{},
			expectedErr: "unauthorized: missing required privilege(s) VEHICLE_NON_LOCATION_DATA",
		},
		{
			name:    "Non Location permissions",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
					speed {
						value
					}
				}
			}`,
			permissions: []int{1},
		},
		{
			name:    "Location permissions",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
					currentLocationLatitude {
						value
					}
				}
			}`,
			permissions: []int{4},
		},
		{
			name:    "Approximate Location permissions",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
					currentLocationApproximateLatitude {
						value
					}
				}
			}`,
			permissions: []int{8},
		},
		{
			name:    "Approximate Location with ALL_TIME permission",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
					currentLocationApproximateLatitude {
						value
					}
				}
			}`,
			permissions: []int{4},
		},

		{
			name:    "Neither Location nor Approximate Location permissions",
			tokenID: 39718,
			query: `query {
				signalsLatest(tokenId: 39718) {
					lastSeen
					currentLocationApproximateLatitude {
						value
					}
				}
			}`,
			permissions: []int{1},
			expectedErr: "unauthorized: requires at least one of the following privileges [VEHICLE_APPROXIMATE_LOCATION VEHICLE_ALL_TIME_LOCATION]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := services.Auth.CreateVehicleToken(t, strconv.Itoa(tt.tokenID), tt.permissions)
			// Execute request
			result := map[string]any{}
			err := telemetryClient.Post(tt.query, &result, WithToken(token))
			if tt.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
