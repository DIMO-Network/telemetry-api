package repositories

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	// MaxPageSize is the maximum page size for paginated results
	MaxPageSize = 100
)

var (
	errInvalidToken = fmt.Errorf("invalid token")

	// InternalError is a generic error message for internal errors.
	InternalError = gqlerror.Errorf("Internal error")
)

type primaryKey struct {
	TokenID int `json:"primaryKeys"`
}

// EncodeGlobalTokenID encodes a global token form and ID by prefixing it with a string and encoding it to base64.
func EncodeGlobalTokenID(prefix string, id int) (string, error) {
	var buf bytes.Buffer
	e := msgpack.NewEncoder(&buf)
	e.UseArrayEncodedStructs(true)
	err := e.Encode(primaryKey{TokenID: id})
	if err != nil {
		return "", fmt.Errorf("error encoding token id: %w", err)
	}
	return encodeGlobalToken(prefix, buf.Bytes()), nil
}

// encodeGlobalToken encodes a global token by prefixing it with a string and encoding it to base64.
func encodeGlobalToken(prefix string, data []byte) string {
	return fmt.Sprintf("%s_%s", prefix, base64.StdEncoding.EncodeToString(data))
}

// DecodeGlobalTokenID decodes a global token and returns the prefix and token id.
func DecodeGlobalTokenID(token string) (string, int, error) {
	prefix, data, err := decodeGlobalToken(token)
	if err != nil {
		return "", 0, err
	}
	var pk primaryKey
	d := msgpack.NewDecoder(bytes.NewBuffer(data))
	if err := d.Decode(&pk); err != nil {
		return "", 0, fmt.Errorf("error decoding token id: %w", err)
	}
	return prefix, pk.TokenID, nil
}

// decodeGlobalToken decodes a global token by removing the prefix and decoding it from base64.
func decodeGlobalToken(token string) (string, []byte, error) {
	parts := strings.SplitN(token, "_", 2)
	if len(parts) != 2 {
		return "", nil, errInvalidToken
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, errInvalidToken
	}
	return parts[0], data, nil
}

func validateFirstLast(first, last *int, maxPageSize int) error {
	if first != nil {
		if last != nil {
			return errors.New("pass `first` or `last`, but not both")
		}
		if *first < 0 {
			return errors.New("the value for `first` cannot be negative")
		}
		if *first > maxPageSize {
			return fmt.Errorf("the value %d for `first` exceeds the limit %d", *first, maxPageSize)
		}
		return nil
	}

	if last == nil {
		return errors.New("provide `first` or `last`")
	}
	if *last < 0 {
		return errors.New("the value for `last` cannot be negative")
	}
	if *last > maxPageSize {
		return fmt.Errorf("the value %d for `last` exceeds the limit %d", *last, maxPageSize)
	}

	return nil
}
