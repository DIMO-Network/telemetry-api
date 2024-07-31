package auth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/services/identity"

	"github.com/ethereum/go-ethereum/common"
)

var vehilcePrivMap = map[privileges.Privilege]model.Privilege{
	privileges.VehicleNonLocationData: model.PrivilegeVehicleNonLocationData,
	privileges.VehicleCommands:        model.PrivilegeVehicleCommands,
	privileges.VehicleCurrentLocation: model.PrivilegeVehicleCurrentLocation,
	privileges.VehicleAllTimeLocation: model.PrivilegeVehicleAllTimeLocation,
	privileges.VehicleVinCredential:   model.PrivilegeVehicleVinCredential,
}

var manufacturerPrivMap = map[privileges.Privilege]model.Privilege{
	// privileges.ManufacturerDeviceLastSeen: model.PrivilegeManufacturerDeviceLastSeen,
}

type PrivilegeValidator struct {
	VehicleNFTAddress      common.Address
	ManufacturerNFTAddress common.Address
}

type TokenValidator struct {
	IdentitySvc identity.IdentityService
}

// VehicleNFTPrivCheck checks if the claim set in the context includes the correct address the required privileges for the VehicleNFT contract.
func (pcv *PrivilegeValidator) VehicleNFTPrivCheck(ctx context.Context, _ any, next graphql.Resolver, privs []model.Privilege) (any, error) {
	if err := pcv.checkPrivWithAddress(ctx, privs, pcv.VehicleNFTAddress); err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return next(ctx)
}

// ManufacturerNFTPrivCheck checks if the claim set in the context includes the correct address the required privileges for the ManufacturerNFT contract.
func (pcv *PrivilegeValidator) ManufacturerNFTPrivCheck(ctx context.Context, obj interface{}, next graphql.Resolver, privs []model.Privilege) (res interface{}, err error) {
	if err := pcv.checkPrivWithAddress(ctx, privs, pcv.ManufacturerNFTAddress); err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return next(ctx)
}

// VehicleTokenCheck checks if the vehicle tokenID in the context matches the tokenID in the claim.
func (tv *TokenValidator) VehicleTokenCheck(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	vehicleTokenID, err := getArg[int](ctx, "tokenId")
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	if err := headerTokenMatchesQuery(ctx, func() (string, error) {
		return strconv.Itoa(vehicleTokenID), nil
	}); err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return next(ctx)
}

func (tv *TokenValidator) ManufacturerTokenCheck(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	// TODO(ae) allow user to search by addr, tokenID, or serial
	aftermarketDeviceAddr, err := getArg[common.Address](ctx, "address")
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	if err := headerTokenMatchesQuery(ctx, func() (string, error) {
		resp, err := tv.IdentitySvc.AftermarketDevice(ctx, &aftermarketDeviceAddr, nil, nil)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(resp.ManufacturerTokenID), nil
	}); err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return next(ctx)
}

func (pcv *PrivilegeValidator) checkPrivWithAddress(ctx context.Context, privs []model.Privilege, expectedAddr common.Address) error {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return err
	}

	if claim.ContractAddress != expectedAddr {
		return fmt.Errorf("expected contract %s but recieved: %s", expectedAddr, claim.ContractAddress.Hex())
	}

	for _, priv := range privs {
		if _, ok := claim.privileges[priv]; !ok {
			return fmt.Errorf("missing required privilege: %s", priv)
		}
	}

	return nil
}

func headerTokenMatchesQuery(ctx context.Context, getTokenStrFromArgs func() (string, error)) error {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return err
	}

	tknStr, err := getTokenStrFromArgs()
	if err != nil {
		return err
	}

	if tknStr != claim.TokenID {
		return fmt.Errorf("token id does not match")
	}

	return nil
}

func getArg[T any](ctx context.Context, key string) (T, error) {
	var resp T
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return resp, fmt.Errorf("no field context found")
	}

	resp, ok := fCtx.Args[key].(T)
	if !ok {
		return resp, fmt.Errorf("failed to get %s of type %T from args", key, resp)
	}

	return resp, nil
}
