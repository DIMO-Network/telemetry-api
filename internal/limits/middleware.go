package limits

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Limiter adds request timeouts to HTTP handlers.
type Limiter struct {
	maxRequestDuration time.Duration
}

// New creates a new Limiter, which adds the given request timeout to HTTP handlers.
// The string maxRequestDuration must be parseable as a positive time.Duration; e.g.,
// "5m", "30s".
func New(maxRequestDuration string) (*Limiter, error) {
	mrd, err := time.ParseDuration(maxRequestDuration)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse duration %q: %w", maxRequestDuration, err)
	}

	if mrd <= 0 {
		return nil, fmt.Errorf("duration %s was not positive", maxRequestDuration)
	}

	return &Limiter{maxRequestDuration: mrd}, nil
}

// AddRequestTimeout wraps the given handler in a new one that cancels the
// request context after the duration specified in the limiter.
func (l *Limiter) AddRequestTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), l.maxRequestDuration)
		defer cancel()
		next.ServeHTTP(w, r.Clone(ctx))
	})
}
