package types

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalFloat32(f float32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf("%g", f))
	})
}

func UnmarshalFloat32(v interface{}) (float32, error) {
	switch v := v.(type) {
	case string:
		val, err := strconv.ParseFloat(v, 32)
		return float32(val), err
	case int:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case float32:
		return v, nil
	case json.Number:
		val, err := strconv.ParseFloat(string(v), 32)
		return float32(val), err
	default:
		return 0, fmt.Errorf("%T is not an float", v)
	}
}

func MarshalFloat32Context(f float32) graphql.ContextMarshaler {
	return graphql.ContextWriterFunc(func(ctx context.Context, w io.Writer) error {
		if math.IsInf(float64(f), 0) || math.IsNaN(float64(f)) {
			return fmt.Errorf("cannot marshal infinite no NaN float values")
		}
		io.WriteString(w, fmt.Sprintf("%g", f))
		return nil
	})
}

func UnmarshalFloatContext(ctx context.Context, v interface{}) (float32, error) {
	return UnmarshalFloat32(v)
}
