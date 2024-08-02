package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	identity "github.com/DIMO-Network/telemetry-api/internal/services/identity"

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

//go:generate mockgen -source=./auth.go -destination=auth_mocks.go -package=auth
type IdentityService interface {
	GetAftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*identity.DeviceInfos, error)
}

type TokenValidator struct {
	IdentitySvc IdentityService
}

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

func (tv *TokenValidator) VehicleTokenCheck(contractAddr string) func(context.Context, any, graphql.Resolver) (any, error) {
	requiredAddr := common.HexToAddress(contractAddr)

	return func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		vehicleTokenID, err := getArg[int](ctx, tokenIdArg)
		if err != nil {
			return nil, UnauthorizedError{err: err}
		}

		if err := headerTokenMatchesQuery(ctx, requiredAddr, func() (string, error) {
			return strconv.Itoa(vehicleTokenID), nil
		}); err != nil {
			return nil, UnauthorizedError{err: err}
		}

		return next(ctx)
	}
}

func (tv *TokenValidator) ManufacturerTokenCheck(contractAddr string) func(context.Context, any, graphql.Resolver) (any, error) {
	requiredAddr := common.HexToAddress(contractAddr)

	return func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		adFilter, err := getArg[model.AftermarketDeviceBy](ctx, "by")
		if err != nil {
			return nil, fmt.Errorf("unauthorized: %w", err)
		}

		if err := headerTokenMatchesQuery(ctx, requiredAddr, func() (string, error) {
			resp, err := tv.IdentitySvc.GetAftermarketDevice(ctx, adFilter.Address, adFilter.TokenID, adFilter.Serial)
			if err != nil {
				return "", err
			}
			return strconv.Itoa(resp.ManufacturerTokenID), nil
		}); err != nil {
			return nil, UnauthorizedError{err: err}
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

	return nil, nil
}

func headerTokenMatchesQuery(ctx context.Context, requiredAddr common.Address, getTokenStrFromArgs func() (string, error)) error {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return err
	}

	if claim.ContractAddress != requiredAddr {
		return newError("contract in claim is %s instead of the required %s", claim.ContractAddress, requiredAddr)
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

func getArg[T any](ctx context.Context, name string) (T, error) {
	var resp T
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return resp, errors.New("no field context found")
	}

	val, ok := fCtx.Args[name]
	if !ok {
		return resp, fmt.Errorf("no argument named %s", name)
	}

	resp, ok = val.(T)
	if !ok {
		return resp, fmt.Errorf("argument %s had type %T instead of the expected %T", name, val, resp)
	}

	return resp, nil
}
