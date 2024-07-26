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

var privToAPI = map[privileges.Privilege]model.Privilege{
	privileges.VehicleNonLocationData: model.PrivilegeVehicleNonLocationData,
	privileges.VehicleCommands:        model.PrivilegeVehicleCommands,
	privileges.VehicleCurrentLocation: model.PrivilegeVehicleCurrentLocation,
	privileges.VehicleAllTimeLocation: model.PrivilegeVehicleAllTimeLocation,
	privileges.VehicleVinCredential:   model.PrivilegeVehicleVinCredential,
}

// RequiresTokenCheck checks if the tokenID in the context matches the tokenID in the claim.
func RequiresTokenCheck(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return nil, fmt.Errorf("%w: no field context found", errUnauthorized)
	}
	tokenID, ok := fCtx.Args["tokenId"].(int)
	if !ok {
		return nil, fmt.Errorf("%w: failed to get tokenID from args", errUnauthorized)
	}
	if strconv.Itoa(tokenID) != claim.TokenID {
		return nil, fmt.Errorf("%w: tokenID mismatch", errUnauthorized)
	}
	return next(ctx)
}

type PrivilegeContractValidator struct {
	VehicleNFTAddress common.Address
	// ManufAddr         string
}

func (pcv *PrivilegeContractValidator) VehicleNFTPrivCheck(ctx context.Context, _ any, next graphql.Resolver, privs []model.Privilege) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
	}

	if claim.ContractAddress != pcv.VehicleNFTAddress {
		return nil, fmt.Errorf("%w: contract address mismatch", errUnauthorized)
	}

	for _, priv := range privs {
		if _, ok := claim.privileges[priv]; !ok {
			return nil, fmt.Errorf("%w: missing required privilege %s", errUnauthorized, priv)
		}
	}

	return next(ctx)
}

// func (pcv *PrivilegeContractValidator) ManufacturerNFTPrivCheck(ctx context.Context, _ any, next graphql.Resolver, privs []model.Privilege) (any, error) {
// 	claim, err := getTelemetryClaim(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: %w", errUnauthorized.Error(), err)
// 	}

// 	if claim.ContractAddress.Hex() != pcv.ManufAddr {
// 		return nil, fmt.Errorf("%w: contract address mismatch", errUnauthorized)
// 	}

// 	for _, priv := range privs {
// 		if _, ok := claim.privileges[priv]; !ok {
// 			return nil, fmt.Errorf("%w: missing required privilege %s", errUnauthorized, priv)
// 		}
// 	}

// 	return next(ctx)
// }
