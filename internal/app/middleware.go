package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// errorPresenter wraps the server-garage error handler in order to lower the log level
// for client errors. The default behavior logs everything at error level, which is too
// noisy for invalid queries and bad request inputs.
func errorPresenter(ctx context.Context, err error) *gqlerror.Error {
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		code := errorhandler.ErrCode(gqlErr)
		switch code {
		case errorhandler.CodeGraphQLValidationFailed, errorhandler.CodeGraphQLParseFailed:
			zerolog.Ctx(ctx).Debug().
				Err(gqlErr.Err).
				Str("gqlPath", gqlErr.Path.String()).
				Fields(gqlErr.Extensions).
				Msg(gqlErr.Message)
			return gqlErr
		case errorhandler.CodeBadRequest, errorhandler.CodeBadUserInput:
			zerolog.Ctx(ctx).Warn().
				Err(gqlErr.Err).
				Str("gqlPath", gqlErr.Path.String()).
				Fields(gqlErr.Extensions).
				Msg(gqlErr.Message)
			return gqlErr
		}
	}
	return errorhandler.ErrorPresenter(ctx, err)
}

// LoggerMiddleware adds the source IP to the logger.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get source ip from request could be cloudflare proxy
		sourceIP := r.Header.Get("X-Forwarded-For")
		if sourceIP == "" {
			sourceIP = r.Header.Get("X-Real-IP")
		}
		if sourceIP == "" {
			sourceIP = r.RemoteAddr
		}
		loggerCtx := zerolog.Ctx(r.Context()).With().Str("method", r.Method).Str("path", r.URL.Path).Str("sourceIp", sourceIP).Logger().WithContext(r.Context())
		r = r.WithContext(loggerCtx)
		next.ServeHTTP(w, r)
	})
}

// authLoggerMiddleware adds the authenticated user to the logger
func authLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validateClaims, ok := auth.GetValidatedClaims(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		loggerCtx := zerolog.Ctx(r.Context()).With().Str("jwtSubject", validateClaims.RegisteredClaims.Subject).Logger()
		r = r.WithContext(loggerCtx.WithContext(r.Context()))
		next.ServeHTTP(w, r)
	})
}

// PanicRecoveryMiddleware recovers from panics and logs them.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "panic: %v\n%s\n", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
