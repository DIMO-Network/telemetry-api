package errorhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ErrorPresenter is a custom error presenter that logs the error and returns a gqlerror.Error.
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	if err == nil {
		return nil
	}
	var gqlErr *gqlerror.Error
	if !errors.As(err, &gqlErr) {
		// If someone incorrectly returns a raw error, do not expose the error message.
		gqlErr = gqlerror.WrapPath(graphql.GetPath(ctx), err)
		gqlErr.Message = "internal server error"
	}
	zerolog.Ctx(ctx).Error().
		Err(gqlErr.Err).
		Str("gqlPath", gqlErr.Path.String()).
		Fields(gqlErr.Extensions).
		Msg(gqlErr.Message)
	return gqlErr
}

// NewInternalErrorWithMsg creates a new internal server error with a message.
func NewInternalErrorWithMsg(ctx context.Context, err error, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Err:     err,
		Message: message,
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]interface{}{
			"reason": http.StatusText(http.StatusInternalServerError),
			"code":   http.StatusInternalServerError,
		},
	}
}

// NewBadRequestErrorWithMsg creates a new bad request error with a message.
func NewBadRequestErrorWithMsg(ctx context.Context, err error, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Err:     err,
		Message: message,
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]interface{}{
			"reason": http.StatusText(http.StatusBadRequest),
			"code":   http.StatusBadRequest,
		},
	}
}

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(ctx context.Context, err error) *gqlerror.Error {
	return NewBadRequestErrorWithMsg(ctx, err, err.Error())
}

// NewUnauthorizedErrorWithMsg creates a new unauthorized error with a message.
func NewUnauthorizedErrorWithMsg(ctx context.Context, err error, message string) *gqlerror.Error {
	return &gqlerror.Error{
		Err:     err,
		Message: message,
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]interface{}{
			"reason": http.StatusText(http.StatusUnauthorized),
			"code":   http.StatusUnauthorized,
		},
	}
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(ctx context.Context, err error) *gqlerror.Error {
	return NewUnauthorizedErrorWithMsg(ctx, err, err.Error())
}
