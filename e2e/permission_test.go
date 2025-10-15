package e2e_test

import (
	"testing"

	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
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
		permissions []string
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
			permissions: []string{},
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
			permissions: []string{},
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
			permissions: []string{},
			expectedErr: "unauthorized: missing required privilege(s) privilege:GetNonLocationHistory",
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
			permissions: []string{tokenclaims.PermissionGetNonLocationHistory},
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
			permissions: []string{tokenclaims.PermissionGetLocationHistory},
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
			permissions: []string{tokenclaims.PermissionGetApproximateLocation},
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
			permissions: []string{tokenclaims.PermissionGetLocationHistory},
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
			permissions: []string{tokenclaims.PermissionGetNonLocationHistory},
			expectedErr: "unauthorized: requires at least one of the following privileges [privilege:GetApproximateLocation privilege:GetLocationHistory]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := services.Auth.CreateVehicleToken(t, tt.tokenID, tt.permissions)
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
