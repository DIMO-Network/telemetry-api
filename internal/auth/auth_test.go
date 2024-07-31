package auth

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/middleware/privilegetoken"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/services/identity"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var expectedReturn = struct{}{}

func emptyResolver(_ context.Context) (any, error) {
	return expectedReturn, nil
}

func TestRequiresVehicleTokenCheck(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		args          map[string]any
		telmetryClaim *TelemetryClaim
		expectedError bool
	}{
		{
			name: "valid_token",
			args: map[string]any{
				"tokenId": 123,
			},
			telmetryClaim: &TelemetryClaim{
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
			telmetryClaim: &TelemetryClaim{
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
			telmetryClaim: &TelemetryClaim{
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

	tknValidator := &TokenValidator{
		IdentitySvc: &identity.MockIdentityService{},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testCtx := graphql.WithFieldContext(context.Background(), &graphql.FieldContext{
				Args: tc.args,
			})
			testCtx = context.WithValue(testCtx, TelemetryClaimContextKey{}, tc.telmetryClaim)
			result, err := tknValidator.VehicleTokenCheck(testCtx, nil, graphql.Resolver(emptyResolver))
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedReturn, result)
		})
	}
}
func TestRequiresVehiclePrivilegeCheck(t *testing.T) {
	t.Parallel()
	vehicleNFTAddr := common.BigToAddress(big.NewInt(10))

	testCases := []struct {
		name           string
		privs          []model.Privilege
		telemetryClaim *TelemetryClaim
		expectedError  error
	}{
		{
			name: "valid_privileges",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				contractPrivMap: map[common.Address]map[privileges.Privilege]model.Privilege{
					vehicleNFTAddr: vehilcePrivMap,
				},
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
						privileges.VehicleNonLocationData,
					},
					ContractAddress: vehicleNFTAddr,
				},
			},
		},
		{
			name: "missing_all_privilege",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs:    nil,
					ContractAddress: vehicleNFTAddr,
				},
			},
			expectedError: fmt.Errorf("unauthorized: missing required privilege: %s", model.PrivilegeVehicleAllTimeLocation),
		},
		{
			name: "missing_one_privilege",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
					},
					ContractAddress: vehicleNFTAddr,
				},
			},
			expectedError: fmt.Errorf("unauthorized: missing required privilege: %s", model.PrivilegeVehicleAllTimeLocation),
		},
		{
			name:           "missing_claim",
			privs:          []model.Privilege{},
			telemetryClaim: nil,
			expectedError:  fmt.Errorf("unauthorized: %w", jwtmiddleware.ErrJWTMissing),
		},
		{
			name: "wrongAddr",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
					},
					ContractAddress: common.BigToAddress(big.NewInt(20)),
				},
			},
			expectedError: fmt.Errorf("unauthorized: expected contract %s but recieved: %s", vehicleNFTAddr, common.BigToAddress(big.NewInt(20)).Hex()),
		},
	}

	privValidator := &PrivilegeValidator{
		VehicleNFTAddress: vehicleNFTAddr,
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.telemetryClaim != nil {
				tc.telemetryClaim.SetPrivileges()
			}
			testCtx := context.WithValue(context.Background(), TelemetryClaimContextKey{}, tc.telemetryClaim)
			next, err := privValidator.VehicleNFTPrivCheck(testCtx, nil, emptyResolver, tc.privs)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedReturn, next)

		})
	}
}
