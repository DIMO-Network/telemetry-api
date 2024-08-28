package model

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ethereum/go-ethereum/common"
)

var zeroAddress common.Address

func MarshalAddress(addr common.Address) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(addr.Hex()))
	})
}

func UnmarshalAddress(v interface{}) (common.Address, error) {
	s, ok := v.(string)
	if !ok {
		return zeroAddress, fmt.Errorf("type %T not a string", v)
	}

	if !common.IsHexAddress(s) {
		return zeroAddress, errors.New("not a valid hex address")
	}

	return common.HexToAddress(s), nil
}
