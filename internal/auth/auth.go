package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	services "github.com/DIMO-Network/telemetry-api/internal/service/identity_api"
	"github.com/ethereum/go-ethereum/common"
)

var errUnauthorized = errors.New("unauthorized")

var privToAPI = map[privileges.Privilege]model.Privilege{
	privileges.VehicleNonLocationData: model.PrivilegeVehicleNonLocationData,
	privileges.VehicleCommands:        model.PrivilegeVehicleCommands,
	privileges.VehicleCurrentLocation: model.PrivilegeVehicleCurrentLocation,
	privileges.VehicleAllTimeLocation: model.PrivilegeVehicleAllTimeLocation,
	privileges.VehicleVinCredential:   model.PrivilegeVehicleVinCredential,
}

type PrivilegeContractValidator struct {
	vehicleNFTAddress      common.Address
	manufacturerNFTAddress common.Address
	identityService        services.IdentityService
}

func NewPrivilegeContractValidator(settings config.Settings) *PrivilegeContractValidator {
	vNFTAddr := common.HexToAddress(settings.VehicleNFTAddress)
	mNFTAddr := common.HexToAddress(settings.ManufacturerNFTAddress)
	return &PrivilegeContractValidator{
		vehicleNFTAddress:      vNFTAddr,
		manufacturerNFTAddress: mNFTAddr,
		identityService:        services.NewIdentityService(&settings),
	}
}

// VehicleNFTPrivCheck checks if the claim set in the context includes the correct address the required privileges for the VehicleNFT contract.
func (pcv *PrivilegeContractValidator) VehicleNFTPrivCheck(ctx context.Context, _ any, next graphql.Resolver, privs []model.Privilege) (any, error) {
	if err := pcv.checkPrivWithAddress(ctx, privs, pcv.vehicleNFTAddress); err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}

	return next(ctx)
}

// ManufacturerNFTPrivCheck checks if the claim set in the context includes the correct address the required privileges for the ManufacturerNFT contract.
func (pcv *PrivilegeContractValidator) ManufacturerNFTPrivCheck(ctx context.Context, obj interface{}, next graphql.Resolver, priv model.Privilege) (res interface{}, err error) {
	if err := pcv.checkPrivWithAddress(ctx, []model.Privilege{priv}, pcv.manufacturerNFTAddress); err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}

	return next(ctx)
}

func (pcv *PrivilegeContractValidator) checkPrivWithAddress(ctx context.Context, privs []model.Privilege, contractAddr common.Address) error {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return err
	}

	if claim.ContractAddress != contractAddr {
		return fmt.Errorf("contract address mismatch")
	}

	for _, priv := range privs {
		if _, ok := claim.privileges[priv]; !ok {
			return fmt.Errorf("missing required privilege %s", priv)
		}
	}

	return nil
}

// VehicleTokenCheck checks if the vehicle tokenID in the context matches the tokenID in the claim.
func (pcv *PrivilegeContractValidator) VehicleTokenCheck(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, err
	}

	vehicleTokenID, err := getRequestValues[int](ctx, "tokenId")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}

	if strconv.Itoa(vehicleTokenID) != claim.TokenID {
		return nil, fmt.Errorf("%w: tokenID mismatch", errUnauthorized)
	}

	return next(ctx)
}

func (pcv *PrivilegeContractValidator) ManufacturerTokenCheck(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, err
	}

	// TODO allow user to search by addr, tokenID, or serial
	aftermarketDeviceAddr, err := getRequestValues[common.Address](ctx, "address")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}

	resp, err := pcv.identityService.AftermarketDevice(ctx, services.AftermarketDeviceBy{Address: &aftermarketDeviceAddr})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get aftermarket device", errUnauthorized)
	}

	if strconv.Itoa(resp.AftermarketDevice.Manufacturer.TokenId) != claim.TokenID {
		return nil, fmt.Errorf("%w: tokenID mismatch", errUnauthorized)
	}

	return next(ctx)
}

func getRequestValues[T any](ctx context.Context, key string) (T, error) {
	var resp T
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return resp, fmt.Errorf("no field context found")
	}

	resp, ok := fCtx.Args[key].(T)
	if !ok {
		return resp, fmt.Errorf("failed to get %s from args", key)
	}

	return resp, nil
}
