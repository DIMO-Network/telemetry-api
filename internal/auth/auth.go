package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/shared/privileges"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
)

var errUnauthorized = errors.New("unauthorized")

var vehiclePrivToAPI = map[privileges.Privilege]model.Privilege{
	privileges.VehicleNonLocationData: model.PrivilegeVehicleNonLocationData,
	privileges.VehicleCommands:        model.PrivilegeVehicleCommands,
	privileges.VehicleCurrentLocation: model.PrivilegeVehicleCurrentLocation,
	privileges.VehicleAllTimeLocation: model.PrivilegeVehicleAllTimeLocation,
	privileges.VehicleVinCredential:   model.PrivilegeVehicleVinCredential,
}

var manufacturerPrivToAPI = map[privileges.Privilege]model.Privilege{}

const tokenIdArgName = "tokenId"

func CreateVehicleTokenCheck(contractAddr string) func(context.Context, any, graphql.Resolver) (any, error) {
	requiredAddr := common.HexToAddress(contractAddr)

	return func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		claim, err := getTelemetryClaim(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
		}

		if claim.ContractAddress != requiredAddr {
			return nil, fmt.Errorf("%w: contract did not match", errUnauthorized)
		}

		fCtx := graphql.GetFieldContext(ctx)
		if fCtx == nil {
			return nil, fmt.Errorf("%w: no field context found", errUnauthorized)
		}
		tokenID, ok := fCtx.Args[tokenIdArgName].(int)
		if !ok {
			return nil, fmt.Errorf("%w: failed to get %s from args", errUnauthorized, tokenIdArgName)
		}
		if strconv.Itoa(tokenID) != claim.TokenID {
			return nil, fmt.Errorf("%w: %s mismatch", errUnauthorized, tokenIdArgName)
		}

		return next(ctx)
	}
}

// PrivilegeCheck checks if the claim set in the context includes the required privileges.
func PrivilegeCheck(ctx context.Context, _ any, next graphql.Resolver, privs []model.Privilege) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}

	for _, priv := range privs {
		if _, ok := claim.privileges[priv]; !ok {
			return nil, fmt.Errorf("%w: missing required privilege %s", errUnauthorized, priv)
		}
	}

	return next(ctx)
}
