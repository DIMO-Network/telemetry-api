package dtcmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type contextKey string

const (
	// EstimateCostKey is a context key for estimating cost
	EstimateCostKey contextKey = "estimateCost"
)

var ErrOperationNotSet = errors.New("operation not set")

type EstimateCostResponse struct {
	EstimatedCredits uint64 `json:"estimatedCredits"`
	Message          string `json:"message"`
}

// EstimateCostHeaderMiddleware injects estimate cost header into the context
func EstimateCostHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get source ip from request could be cloudflare proxy
		if r.Header.Get("X-Estimate-Cost") == "true" {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), EstimateCostKey, true)))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), EstimateCostKey, false)))
	})
}

// isEstimationRequest checks if the request has the cost estimation header
func (d DCT) isEstimationRequest(ctx context.Context) bool {
	if ok, isTrue := ctx.Value(EstimateCostKey).(bool); ok && isTrue {
		return true
	}
	return false
}

// handleCostEstimation calculates and returns cost without executing the query
func (d DCT) handleCostEstimation(ctx context.Context) *graphql.Response {
	// Calculate the cost
	credits, gqlErr := d.calculateCredits(ctx)
	if gqlErr != nil {
		return &graphql.Response{
			Errors: gqlerror.List{gqlErr},
		}
	}
	estimate := EstimateCostResponse{
		EstimatedCredits: credits,
		Message:          fmt.Sprintf("This query would cost %d credits", credits),
	}

	// Marshal to JSON
	data, err := json.Marshal(estimate)
	if err != nil {
		return &graphql.Response{
			Errors: []*gqlerror.Error{
				errorhandler.NewInternalErrorWithMsg(ctx, err, "Failed to format cost estimation"),
			},
		}
	}

	zerolog.Ctx(ctx).Info().
		Uint64("estimatedCredits", credits).
		Msg("Cost estimation requested")

	return &graphql.Response{
		Data: data,
	}
}
