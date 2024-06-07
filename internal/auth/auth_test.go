package auth_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/require"
)

var expectedReturn = struct{}{}

func emptyResolver(_ context.Context) (any, error) {
	return expectedReturn, nil
}

func TestRequiresTokenCheck(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		args          map[string]any
		telmetryClaim *auth.TelemetryClaim
		expectedError bool
	}{
		{
			name: "valid_token",
			args: map[string]any{
				"tokenId": 123,
			},
			telmetryClaim: &auth.TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					TokenID: "123",
				},
			},
		},
		{
			name: "invalid_token",
			args: map[string]any{
				"tokenId": 456,
			},
			telmetryClaim: &auth.TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					TokenID: "123",
				},
			},
			expectedError: true,
		},
		{
			name:          "missing_tokenId",
			args:          map[string]any{},
			expectedError: true,
			telmetryClaim: &auth.TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					TokenID: "123",
				},
			},
		},
		{
			name: "missing claim",
			args: map[string]any{
				"tokenId": 123,
			},
			expectedError: true,
			telmetryClaim: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testCtx := graphql.WithFieldContext(context.Background(), &graphql.FieldContext{
				Args: tc.args,
			})
			testCtx = context.WithValue(testCtx, auth.TelemetryClaimContextKey{}, tc.telmetryClaim)
			result, err := auth.RequiresTokenCheck(testCtx, nil, graphql.Resolver(emptyResolver))
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedReturn, result)
		})
	}
}
func TestRequiresPrivilegeCheck(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		privs          []model.Privilege
		telemetryClaim *auth.TelemetryClaim
		expectedError  bool
	}{
		{
			name: "valid_privileges",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &auth.TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
						privileges.VehicleNonLocationData,
					},
				},
			},
		},
		{
			name: "missing_all_privilege",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &auth.TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs: nil,
				},
			},
			expectedError: true,
		},
		{
			name: "missing_one_privilege",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &auth.TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
					},
				},
			},
			expectedError: true,
		},
		{
			name:           "missing_claim",
			privs:          []model.Privilege{},
			telemetryClaim: nil,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.telemetryClaim != nil {
				tc.telemetryClaim.SetPrivileges()
			}
			testCtx := context.WithValue(context.Background(), auth.TelemetryClaimContextKey{}, tc.telemetryClaim)
			next, err := auth.RequiresPrivilegeCheck(testCtx, nil, emptyResolver, tc.privs)
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedReturn, next)
		})
	}
}
