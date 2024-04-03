package types

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

// MarshalInt16 marshals an int16 to a graphql.Marshaler.
func MarshalInt16(i int16) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatInt(int64(i), 10))
	})
}

// UnmarshalInt16 unmarshals an int16 from a graphql response.
func UnmarshalInt16(v any) (int16, error) {
	switch typedV := v.(type) {
	case string:
		val, err := strconv.ParseInt(typedV, 10, 16)
		return int16(val), err
	case int:
		return int16(typedV), nil
	case int64:
		return int16(typedV), nil
	case json.Number:
		val, err := strconv.ParseInt(string(typedV), 10, 16)
		return int16(val), err
	default:
		return 0, fmt.Errorf("%T is not an int", v)
	}
}

// MarshalInt8 marshals an int8 to a graphql.Marshaler.
func MarshalInt8(i int8) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatInt(int64(i), 10))
	})
}

// UnmarshalInt8 unmarshals an int8 from a graphql response.
func UnmarshalInt8(v any) (int8, error) {
	switch typedV := v.(type) {
	case string:
		val, err := strconv.ParseInt(typedV, 10, 8)
		return int8(val), err
	case int:
		return int8(typedV), nil
	case int64:
		return int8(typedV), nil
	case json.Number:
		val, err := strconv.ParseInt(string(typedV), 10, 8)
		return int8(val), err
	default:
		return 0, fmt.Errorf("%T is not an int", v)
	}
}
