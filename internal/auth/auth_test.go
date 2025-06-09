package auth

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/pkg/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/identity"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
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

	vehicleNFTAddr := common.HexToAddress("0x1")

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
				CustomClaims: tokenclaims.CustomClaims{
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
				CustomClaims: tokenclaims.CustomClaims{
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
				CustomClaims: tokenclaims.CustomClaims{
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
				CustomClaims: tokenclaims.CustomClaims{
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

	vehicleCheck := NewVehicleTokenCheck(vehicleNFTAddr)
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
	mtrNFTAddr := common.HexToAddress("0x1")

	autopiAddr := common.BigToAddress(big.NewInt(123))
	autopiTknID := 123
	autopiSerial := "serial"

	testCases := []struct {
		name             string
		args             model.AftermarketDeviceBy
		telmetryClaim    *TelemetryClaim
		identityResponse *identity.DeviceInfos
		identityError    error
		expectedError    error
	}{
		{
			name: "valid_manufacturer_token_by_address",
			args: model.AftermarketDeviceBy{
				Address: &autopiAddr,
			},
			telmetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					ContractAddress: mtrNFTAddr,
					TokenID:         "137",
				},
			},
			identityResponse: &identity.DeviceInfos{
				ManufacturerTokenID: 137,
			},
		},
		{
			name: "valid_manufacturer_token_by_token_id",
			args: model.AftermarketDeviceBy{
				TokenID: &autopiTknID,
			},
			telmetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					ContractAddress: mtrNFTAddr,
					TokenID:         "137",
				},
			},
			identityResponse: &identity.DeviceInfos{
				ManufacturerTokenID: 137,
			},
		},
		{
			name: "valid_manufacturer_token_by_serial",
			args: model.AftermarketDeviceBy{
				Serial: &autopiSerial,
			},
			telmetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					ContractAddress: mtrNFTAddr,
					TokenID:         "137",
				},
			},
			identityResponse: &identity.DeviceInfos{
				ManufacturerTokenID: 137,
			},
		},
		{
			name: "wrong aftermarket device manufacturer",
			telmetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					ContractAddress: mtrNFTAddr,
					TokenID:         "138",
				},
			},
			args: model.AftermarketDeviceBy{
				Address: &autopiAddr,
			},
			identityError: fmt.Errorf("invalid autopi address"),
			expectedError: fmt.Errorf("unauthorized: token id does not match"),
		},
		{
			name: "invalid autopi address",
			telmetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					ContractAddress: mtrNFTAddr,
					TokenID:         "137",
				},
			},
			args: model.AftermarketDeviceBy{
				Address: &autopiAddr,
				TokenID: &autopiTknID,
				Serial:  &autopiSerial,
			},
			identityError: fmt.Errorf("invalid autopi address"),
			expectedError: fmt.Errorf("unauthorized: token id does not match"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testCtx := graphql.WithFieldContext(context.Background(), &graphql.FieldContext{
				Args: map[string]any{
					"by": tc.args,
				},
			})

			id := NewMockIdentityService(gomock.NewController(t))
			id.EXPECT().GetAftermarketDevice(context.WithValue(testCtx, TelemetryClaimContextKey{}, tc.telmetryClaim), tc.args.Address, tc.args.TokenID, tc.args.Serial).Return(
				tc.identityResponse, tc.identityError).AnyTimes()

			mfrValidator := NewManufacturerTokenCheck(mtrNFTAddr, id)

			testCtx = context.WithValue(testCtx, TelemetryClaimContextKey{}, tc.telmetryClaim)
			result, err := mfrValidator(testCtx, nil, graphql.Resolver(emptyResolver))
			if tc.expectedError != nil {
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
				CustomClaims: tokenclaims.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
						privileges.VehicleNonLocationData,
					},
					ContractAddress: vehicleNFTAddr,
				},
			},
			expectedError: false,
		},
		{
			name: "missing_all_privilege",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
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
				CustomClaims: tokenclaims.CustomClaims{
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
			name: "wrong contract for privilege",
			// this will be the same as no privilege bc we dont have this priv for the passed contract
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						// this is the same number priv but from a different contract
						privileges.ManufacturerDeviceDefinitionInsert,
					},
					ContractAddress: manufNFTAddr,
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
			next, err := AllOfPrivilegeCheck(testCtx, nil, emptyResolver, tc.privs)
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedReturn, next)
			}
		})
	}
}

func TestRequiresOneOfPrivilegeCheck(t *testing.T) {
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
			name: "has_one_of_required_privileges",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
					},
					ContractAddress: vehicleNFTAddr,
				},
			},
			expectedError: false,
		},
		{
			name: "has_all_required_privileges",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.VehicleAllTimeLocation,
						privileges.VehicleNonLocationData,
					},
					ContractAddress: vehicleNFTAddr,
				},
			},
			expectedError: false,
		},
		{
			name: "missing_all_privileges",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
				model.PrivilegeVehicleNonLocationData,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					PrivilegeIDs:    nil,
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
			name: "wrong_contract_for_privilege",
			privs: []model.Privilege{
				model.PrivilegeVehicleAllTimeLocation,
			},
			telemetryClaim: &TelemetryClaim{
				CustomClaims: tokenclaims.CustomClaims{
					PrivilegeIDs: []privileges.Privilege{
						privileges.ManufacturerDeviceDefinitionInsert,
					},
					ContractAddress: manufNFTAddr,
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
			next, err := OneOfPrivilegeCheck(testCtx, nil, emptyResolver, tc.privs)
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedReturn, next)
			}
		})
	}
}

func TestGetAttestationClaimMap(t *testing.T) {

	testCases := []struct {
		name        string
		getContext  func(context.Context) context.Context
		expectedErr error
	}{
		{
			name: "no claims found",
			getContext: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedErr: errors.New("no cloudevent claims found"),
		},
		{
			name: "no cloud events in claim",
			getContext: func(ctx context.Context) context.Context {
				var claims TelemetryClaim
				return context.WithValue(ctx, TelemetryClaimContextKey{}, &claims)
			},
			expectedErr: errors.New("no cloudevent claims found"),
		},
		{
			name: "no events, still valid",
			getContext: func(ctx context.Context) context.Context {
				var claims TelemetryClaim
				claims.CloudEvents = &tokenclaims.CloudEvents{
					Events: []tokenclaims.Event{},
				}
				return context.WithValue(ctx, TelemetryClaimContextKey{}, &claims)
			},
		},
		{
			name: "valid events set",
			getContext: func(ctx context.Context) context.Context {
				var claims TelemetryClaim
				claims.CloudEvents = &tokenclaims.CloudEvents{
					Events: []tokenclaims.Event{
						{
							EventType: tokenclaims.GlobalIdentifier,
							Source:    tokenclaims.GlobalIdentifier,
							IDs:       []string{tokenclaims.GlobalIdentifier},
						},
					},
				}
				return context.WithValue(ctx, TelemetryClaimContextKey{}, &claims)
			},
		},
	}

	for _, tc := range testCases {

		ctx := context.Background()
		ctx = tc.getContext(ctx)
		_, err := GetAttestationClaimMap(ctx)
		if tc.expectedErr != nil {
			require.Equal(t, err.Error(), tc.expectedErr.Error())
			continue
		}
		require.Nil(t, err)
	}
}
