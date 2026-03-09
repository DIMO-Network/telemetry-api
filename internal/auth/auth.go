package auth

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ethereum/go-ethereum/common"
)

const (
	tokenIdArg = "tokenId"
	subjectArg = "subject"
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

func NewVehicleTokenCheck(requiredAddr common.Address) func(context.Context, any, graphql.Resolver) (any, error) {
	return func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
		tokenID, _ := getArg[*int](ctx, tokenIdArg)
		subject, _ := getArg[*string](ctx, subjectArg)

		switch {
		case tokenID != nil && subject != nil:
			return nil, UnauthorizedError{message: "provide either tokenId or subject, not both"}
		case tokenID != nil:
			if err := validateHeader(ctx, requiredAddr, *tokenID); err != nil {
				return nil, UnauthorizedError{err: err}
			}
		case subject != nil:
			if err := validateSubject(ctx, *subject); err != nil {
				return nil, UnauthorizedError{err: err}
			}
		default:
			return nil, UnauthorizedError{message: "tokenId or subject is required"}
		}

		return next(ctx)
	}
}

func validateSubject(ctx context.Context, subject string) error {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return err
	}
	if subject != claim.Asset {
		return newError("subject does not match token claim")
	}
	return nil
}

func validateHeader(ctx context.Context, requiredAddr common.Address, tokenID int) error {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return err
	}

	if claim.AssetDID.ContractAddress != requiredAddr {
		return newError("contract in claim is %s instead of the required %s", claim.AssetDID.ContractAddress, requiredAddr)
	}

	if claim.AssetDID.TokenID.Int64() != int64(tokenID) {
		return newError("token id does not match")
	}

	return nil
}

// AllOfPrivilegeCheck checks if the claim set in the context includes the required privileges.
func AllOfPrivilegeCheck(ctx context.Context, _ any, next graphql.Resolver, requiredPrivs []string) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, UnauthorizedError{err: err}
	}

	for _, priv := range requiredPrivs {
		if !slices.Contains(claim.Permissions, priv) {
			return nil, newError("missing required privilege(s) %s", priv)
		}
	}

	return next(ctx)
}

// OneOfPrivilegeCheck checks if the claim set in the context includes at least one of the required privileges.
func OneOfPrivilegeCheck(ctx context.Context, _ any, next graphql.Resolver, requiredPrivs []string) (any, error) {
	claim, err := getTelemetryClaim(ctx)
	if err != nil {
		return nil, UnauthorizedError{err: err}
	}

	for _, priv := range requiredPrivs {
		if slices.Contains(claim.Permissions, priv) {
			return next(ctx)
		}
	}

	return nil, newError("requires at least one of the following privileges %v", requiredPrivs)
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
