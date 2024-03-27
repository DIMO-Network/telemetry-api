package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

// MarshalUint16 marshals an uint16 to a graphql.Marshaler.
func MarshalUint16(i uint16) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatUint(uint64(i), 10))
	})
}

// UnmarshalUint16 unmarshals an uint16 from a graphql response.
func UnmarshalUint16(v interface{}) (uint16, error) {
	switch v := v.(type) {
	case string:
		iv, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(iv), nil
	case int:
		if v < 0 {
			return 0, errors.New("cannot convert negative numbers to uint16")
		}

		return uint16(v), nil
	case int64:
		if v < 0 {
			return 0, errors.New("cannot convert negative numbers to uint16")
		}

		return uint16(v), nil
	case json.Number:
		iv, err := strconv.ParseUint(string(v), 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(iv), nil
	default:
		return 0, fmt.Errorf("%T is not an uint", v)
	}
}

// MarshalUint8 marshals an uint8 to a graphql.Marshaler.
func MarshalUint8(i uint8) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.FormatUint(uint64(i), 10))
	})
}

// UnmarshalUint8 unmarshals an uint8 from a graphql response.
func UnmarshalUint8(v interface{}) (uint8, error) {
	switch v := v.(type) {
	case string:
		iv, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(iv), nil
	case int:
		if v < 0 {
			return 0, errors.New("cannot convert negative numbers to uint8")
		}

		return uint8(v), nil
	case int64:
		if v < 0 {
			return 0, errors.New("cannot convert negative numbers to uint8")
		}

		return uint8(v), nil
	case json.Number:
		iv, err := strconv.ParseUint(string(v), 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(iv), nil
	default:
		return 0, fmt.Errorf("%T is not an uint", v)
	}
}
