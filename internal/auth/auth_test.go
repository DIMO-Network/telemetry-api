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
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var expectedReturn = struct{}{}

func emptyResolver(_ context.Context) (any, error) {
	return expectedReturn, nil
}

func TestRequiresVehicleTokenCheck(t *testing.T) {
	t.Parallel()

	vehicleNFTAddrRaw := "0x1"
	vehicleNFTAddr := common.HexToAddress(vehicleNFTAddrRaw)

	testCases := []struct {
		name           string
		args           map[string]any
		telemetryClaim *TelemetryClaim
		expectedError  bool
	}{
		{
			name: "valid_token",
			args: map[string]any{
				"tokenId": 123,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: vehicleNFTAddr,
					TokenID:         "123",
				},
			},
		},
		{
			name: "invalid_token",
			args: map[string]any{
				"tokenId": 456,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: vehicleNFTAddr,
					TokenID:         "123",
				},
			},
			expectedError: true,
		},
		{
			name:          "missing_tokenId",
			args:          map[string]any{},
			expectedError: true,
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: vehicleNFTAddr,
					TokenID:         "123",
				},
			},
		},
		{
			name:          "wrong_contract",
			args:          map[string]any{},
			expectedError: true,
			telemetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: common.HexToAddress("0x4"),
					TokenID:         "123",
				},
			},
		},
		{
			name: "missing claim",
			args: map[string]any{
				"tokenId": 123,
			},
			expectedError:  true,
			telemetryClaim: nil,
		},
	}

	vehicleCheck := CreateVehicleTokenCheck(vehicleNFTAddrRaw)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testCtx := graphql.WithFieldContext(context.Background(), &graphql.FieldContext{
				Args: tc.args,
			})
			testCtx = context.WithValue(testCtx, TelemetryClaimContextKey{}, tc.telemetryClaim)
			result, err := vehicleCheck(testCtx, nil, graphql.Resolver(emptyResolver))
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedReturn, result)
		})
	}
}

func TestRequiresManufacturerTokenCheck(t *testing.T) {
	t.Parallel()

	mftContract := common.HexToAddress("0x1")

	validAutoPiAddr := common.BigToAddress(big.NewInt(123))
	invalidAutoPiAddr := common.BigToAddress(big.NewInt(456))
	id := identity.NewMockIdentityService(gomock.NewController(t))
	id.EXPECT().AftermarketDevice(gomock.Any(), &validAutoPiAddr, gomock.Any(), gomock.Any()).Return(
		&identity.ManufacturerTokenID{
			ManufacturerTokenID: 137,
		}, nil).AnyTimes()
	id.EXPECT().AftermarketDevice(gomock.Any(), &invalidAutoPiAddr, gomock.Any(), gomock.Any()).Return(
		nil, fmt.Errorf("")).AnyTimes()

	tokenValidator := CreateManufacturerTokenCheck(mftContract.Hex(), id)

	testCases := []struct {
		name          string
		args          map[string]any // address of autopi device
		telmetryClaim *TelemetryClaim
		expectedError bool
	}{
		{
			name: "valid_manufacturer_token",
			args: map[string]any{
				"by": model.AftermarketDeviceBy{
					Address: ref(validAutoPiAddr),
				},
			},
			telmetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: mftContract,
					TokenID:         "137",
				},
			},
			expectedError: false,
		},
		{
			name: "wrong aftermarket device manufacturer",
			args: map[string]any{
				"by": model.AftermarketDeviceBy{
					Address: ref(validAutoPiAddr),
				},
			},
			telmetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: mftContract,
					TokenID:         "138",
				},
			},
			expectedError: true,
		},
		{
			name: "wrong address",
			args: map[string]any{
				"by": model.AftermarketDeviceBy{
					Address: ref(invalidAutoPiAddr),
				},
			},
			telmetryClaim: &TelemetryClaim{
				CustomClaims: privilegetoken.CustomClaims{
					ContractAddress: mftContract,
					TokenID:         "137",
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testCtx := graphql.WithFieldContext(context.Background(), &graphql.FieldContext{
				Args: tc.args,
			})
			testCtx = context.WithValue(testCtx, TelemetryClaimContextKey{}, tc.telmetryClaim)
			result, err := tokenValidator(testCtx, nil, graphql.Resolver(emptyResolver))
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
	manufNFTAddr := common.BigToAddress(big.NewInt(11))

	privMaps := map[common.Address]map[privileges.Privilege]model.Privilege{
		vehicleNFTAddr: vehiclePrivToAPI,
		manufNFTAddr:   manufacturerPrivToAPI,
	}

	testCases := []struct {
		name           string
		privs          []model.Privilege
		telemetryClaim *TelemetryClaim
		expectedError  bool
	}{
		{
			name: "valid_privileges",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
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
			expectedError: true,
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
			expectedError: true,
		},
		{
			name:           "missing_claim",
			privs:          []model.Privilege{},
			telemetryClaim: nil,
			expectedError:  true,
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
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.telemetryClaim != nil {
				tc.telemetryClaim.SetPrivileges(privMaps)
			}
			testCtx := context.WithValue(context.Background(), TelemetryClaimContextKey{}, tc.telemetryClaim)
			next, err := PrivilegeCheck(testCtx, nil, emptyResolver, tc.privs)
			if tc.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedReturn, next)

		})
	}
}

// func TestRequiresManufacturerPrivilegeCheck(t *testing.T) {
// 	t.Parallel()
// 	manufacturerNFTAddr := common.BigToAddress(big.NewInt(10))

// 	testCases := []struct {
// 		name           string
// 		privs          []model.Privilege
// 		telemetryClaim *TelemetryClaim
// 		expectedError  error
// 	}{
// 		{
// 			name: "valid_privileges",
// 			privs: []model.Privilege{
// 				model.PrivilegeManufacturerDeviceLastSeen,
// 			},
// 			telemetryClaim: &TelemetryClaim{
// 				contractPrivMap: map[common.Address]map[privileges.Privilege]model.Privilege{
// 					manufacturerNFTAddr: {
// 						privileges.ManufacturerDeviceLastSeen: model.PrivilegeManufacturerDeviceLastSeen,
// 					},
// 				},
// 				CustomClaims: privilegetoken.CustomClaims{
// 					PrivilegeIDs: []privileges.Privilege{
// 						privileges.ManufacturerDeviceLastSeen,
// 					},
// 					ContractAddress: manufacturerNFTAddr,
// 				},
// 			},
// 		},
// 		{
// 			name: "missing_all_privilege",
// 			privs: []model.Privilege{
// 				model.PrivilegeManufacturerDeviceLastSeen,
// 			},
// 			telemetryClaim: &TelemetryClaim{
// 				CustomClaims: privilegetoken.CustomClaims{
// 					PrivilegeIDs:    nil,
// 					ContractAddress: manufacturerNFTAddr,
// 				},
// 			},
// 			expectedError: fmt.Errorf("unauthorized: missing required privilege: %s", model.PrivilegeManufacturerDeviceLastSeen),
// 		},
// 		{
// 			name: "missing_one_privilege",
// 			privs: []model.Privilege{
// 				model.PrivilegeManufacturerDeviceLastSeen,
// 			},
// 			telemetryClaim: &TelemetryClaim{
// 				CustomClaims: privilegetoken.CustomClaims{
// 					PrivilegeIDs: []privileges.Privilege{
// 						privileges.ManufacturerDeviceLastSeen,
// 					},
// 					ContractAddress: manufacturerNFTAddr,
// 				},
// 			},
// 			expectedError: fmt.Errorf("unauthorized: missing required privilege: %s", model.PrivilegeManufacturerDeviceLastSeen),
// 		},
// 		{
// 			name:           "missing_claim",
// 			privs:          []model.Privilege{},
// 			telemetryClaim: nil,
// 			expectedError:  fmt.Errorf("unauthorized: %w", jwtmiddleware.ErrJWTMissing),
// 		},
// 		{
// 			name: "wrongAddr",
// 			privs: []model.Privilege{
// 				model.PrivilegeManufacturerDeviceLastSeen,
// 			},
// 			telemetryClaim: &TelemetryClaim{
// 				CustomClaims: privilegetoken.CustomClaims{
// 					PrivilegeIDs: []privileges.Privilege{
// 						privileges.ManufacturerDeviceLastSeen,
// 					},
// 					ContractAddress: common.BigToAddress(big.NewInt(20)),
// 				},
// 			},
// 			expectedError: fmt.Errorf("unauthorized: expected contract %s but recieved: %s", manufacturerNFTAddr, common.BigToAddress(big.NewInt(20)).Hex()),
// 		},
// 	}

// 	privValidator := &PrivilegeValidator{
// 		ManufacturerNFTAddress: manufacturerNFTAddr,
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()
// 			if tc.telemetryClaim != nil {
// 				tc.telemetryClaim.SetPrivileges()
// 			}
// 			testCtx := context.WithValue(context.Background(), TelemetryClaimContextKey{}, tc.telemetryClaim)
// 			next, err := privValidator.ManufacturerNFTPrivCheck(testCtx, nil, emptyResolver, tc.privs)
// 			if tc.expectedError != nil {
// 				require.Equal(t, tc.expectedError.Error(), err.Error())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.Equal(t, expectedReturn, next)

// 		})
// 	}
// }

func ref[A any](a A) *A {
	return &a
}
