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

const tokenIdArg = "tokenId"

var (
	vehiclePrivToAPI = map[privileges.Privilege]model.Privilege{
		privileges.VehicleNonLocationData: model.PrivilegeVehicleNonLocationData,
		privileges.VehicleCommands:        model.PrivilegeVehicleCommands,
		privileges.VehicleCurrentLocation: model.PrivilegeVehicleCurrentLocation,
		privileges.VehicleAllTimeLocation: model.PrivilegeVehicleAllTimeLocation,
		privileges.VehicleVinCredential:   model.PrivilegeVehicleVinCredential,
	}

	manufacturerPrivToAPI = map[privileges.Privilege]model.Privilege{
		privileges.ManufacturerDeviceLastSeen: model.PrivilegeManufacturerDeviceLastSeen,
	}
)

type UnauthorizedError struct {
	message string
	err     error
}

func (e UnauthorizedError) Error() string {
	if e.message != "" {
		if e.err != nil {
			return fmt.Sprintf("unauthorized: %s: %s", e.message, e.err)
		}
		return fmt.Sprintf("unauthorized: %s", e.message)
	}
	if e.err != nil {
		return fmt.Sprintf("unauthorized: %s", e.err)
	}
	return "unauthorized"
}

func (e UnauthorizedError) Unwrap() error {
	return e.err
}

func newError(msg string, args ...any) error {
	return UnauthorizedError{message: fmt.Sprintf(msg, args...)}
}

func CreateVehicleTokenCheck(contractAddr string) func(context.Context, any, graphql.Resolver) (any, error) {
	requiredAddr := common.HexToAddress(contractAddr)

	return func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		claim, err := getTelemetryClaim(ctx)
		if err != nil {
			return nil, UnauthorizedError{err: err}
		}

		if claim.ContractAddress != requiredAddr {
			return nil, newError("contract in claim is %s instead of the required %s", claim.ContractAddress, requiredAddr)
		}

		fCtx := graphql.GetFieldContext(ctx)
		if fCtx == nil {
			return nil, newError("no field context")
		}
		tokenIDAny, ok := fCtx.Args[tokenIdArg]
		if !ok {
			return nil, newError("no argument named %s", tokenIdArg)
		}
		tokenID, ok := tokenIDAny.(int)
		if !ok {
			return nil, newError("argument %s has type %T instead of the expected %T", tokenIdArg, tokenIDAny, tokenID)

		}
		if strconv.Itoa(tokenID) != claim.TokenID {
			return nil, newError("token id %s in the claim does not match token id %d in the query", claim.TokenID, tokenID)
		}

		return next(ctx)
	}
}

type IdentityService interface {
	AftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*identity.ManufacturerTokenID, error)
}

type TokenValidator struct {
	IdentitySvc IdentityService
}

const byArg = "by"

func CreateManufacturerTokenCheck(contractAddr string, idSvc IdentityService) func(context.Context, any, graphql.Resolver) (any, error) {
	requiredAddr := common.HexToAddress(contractAddr)

	return func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		claim, err := getTelemetryClaim(ctx)
		if err != nil {
			return nil, UnauthorizedError{err: err}
		}

		if claim.ContractAddress != requiredAddr {
			return nil, newError("contract in claim is %s instead of the required %s", claim.ContractAddress, requiredAddr)
		}

		fCtx := graphql.GetFieldContext(ctx)
		if fCtx == nil {
			return nil, newError("no field context")
		}
		byAny, ok := fCtx.Args[byArg]
		if !ok {
			return nil, newError("no argument named %s", byArg)
		}
		by, ok := byAny.(model.AftermarketDeviceBy)
		if !ok {
			return nil, newError("argument %s has type %T instead of the expected %T", byArg, byAny, by)
		}

		ad, err := idSvc.AftermarketDevice(ctx, by.Address, by.TokenID, by.Serial)
		if err != nil {
			return nil, err
		}

		if strconv.Itoa(ad.ManufacturerTokenID) != claim.TokenID {
			return nil, newError("token id %s in the claim does not match device manufacturer token id %d", claim.TokenID, ad.ManufacturerTokenID)
		}

		return next(ctx)
	}
}

// PrivilegeCheck checks if the claim set in the context includes the required privileges.
func PrivilegeCheck(ctx context.Context, _ any, next graphql.Resolver, privs []model.Privilege) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, UnauthorizedError{err: err}
	}

	for _, priv := range privs {
		if !claim.privileges.Contains(priv) {
			return nil, newError("missing required privilege %s", priv)
		}
	}

	return next(ctx)
}
