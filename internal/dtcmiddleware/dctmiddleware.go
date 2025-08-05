// Package dtcmiddleware provides a middleware for the Developer Credit Tracker.
package dtcmiddleware

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/credit-tracker/pkg/grpc"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/pricing"
	"github.com/DIMO-Network/telemetry-api/internal/service/credittracker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

// DCT provides a GraphQL middleware for the Developer Credit Tracker.
type DCT struct {
	Tracker        *credittracker.Client
	CostCalculator *pricing.CostCalculator
}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
} = DCT{}

// ExtensionName returns the name of this extension.
func (DCT) ExtensionName() string {
	return "DCT"
}

// Validate validates the GraphQL schema.
func (DCT) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// NewDCT creates a new DCT middleware with default values.
func NewDCT(tracker *credittracker.Client, costCalculator *pricing.CostCalculator) *DCT {
	return &DCT{
		Tracker:        tracker,
		CostCalculator: costCalculator,
	}
}

// InterceptResponse intercepts GraphQL responses to handle errors from the credit tracker.
func (d DCT) InterceptResponse(
	ctx context.Context,
	next graphql.ResponseHandler,
) *graphql.Response {
	if d.isEstimationRequest(ctx) {
		resp := d.handleCostEstimation(ctx)
		if errors.Is(resp.Errors, ErrOperationNotSet) {
			return next(ctx)
		}
		return resp
	}

	if d.Tracker == nil {
		zerolog.Ctx(ctx).Error().Msg("DCT is not enabled")
		// return graphql.ErrorResponse(ctx, "DCT is not enabled")
	}

	// Determine who to charge
	developerID, tokenID, gqlError := GetSubjectAndTokenID(ctx)
	if gqlError != nil {
		zerolog.Ctx(ctx).Warn().Err(gqlError).Msg("Failed to get subject and token ID")
		// return &graphql.Response{
		// 	Errors: gqlerror.List{gqlError},
		// }
		return next(ctx)
	}

	// Determine how many credits to charge
	credits, gqlError := d.calculateCredits(ctx)
	if gqlError != nil {
		// if errors.Is(gqlError.Err, OperationNotSetError) {
		// 	return next(ctx)
		// }
		// return &graphql.Response{
		// 	Errors: gqlerror.List{gqlError},
		// }
		zerolog.Ctx(ctx).Warn().Err(gqlError).Msg("Failed to calculate credits")
		return next(ctx)
	}

	// Start timing the DCT request
	dctTimer := prometheus.NewTimer(DCTRequestLatency.WithLabelValues("deduct"))
	// Deduct the credits
	referenceID := ksuid.New().String()
	err := d.Tracker.DeductCredits(ctx, referenceID, developerID, tokenID, credits)
	dctTimer.ObserveDuration()

	if err != nil {
		gqlError := processDCTErrorToGraphqlError(ctx, err)
		zerolog.Ctx(ctx).Warn().Err(gqlError.Err).Msg("Failed to deduct credits")
		// return &graphql.Response{
		// 	Errors: gqlerror.List{gqlError},
		// }
		return next(ctx)
	}

	// Complete the request and get the response
	response := next(ctx)

	// If it's our fault the request failed, refund the credits
	if errorhandler.HasErrCode(&response.Errors, errorhandler.CodeInternalServerError) {
		// Start timing the refund operation
		refundTimer := prometheus.NewTimer(DCTRequestLatency.WithLabelValues("refund"))
		err := d.Tracker.RefundCredits(ctx, referenceID)
		refundTimer.ObserveDuration()

		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("Failed to refund credits")
		}
		return response
	}

	return response
}

// processDCTError extracts and processes error details from a gRPC error
func processDCTErrorToGraphqlError(ctx context.Context, err error) *gqlerror.Error {
	st, ok := status.FromError(err)
	if !ok {
		return graphql.DefaultErrorPresenter(ctx, err)
	}

	for _, detail := range st.Details() {
		if errorInfo, ok := detail.(*errdetails.ErrorInfo); ok {
			if err := handleErrorDetails(ctx, errorInfo, err); err != nil {
				return err
			}
		}
	}

	return errorhandler.NewInternalErrorWithMsg(ctx, err, "Failed to process credit operation")
}

// handleErrorDetails processes the error details from a gRPC status error
func handleErrorDetails(ctx context.Context, errorInfo *errdetails.ErrorInfo, originalError error) *gqlerror.Error {
	switch errorInfo.Reason {
	case grpc.ErrorReason_ERROR_REASON_INVALID_ASSET_DID.String():
		err := fmt.Errorf("invalid asset DID: %s", errorInfo.Metadata[grpc.MetadataKey_METADATA_KEY_ASSET_DID.String()])
		return errorhandler.NewInternalErrorWithMsg(ctx, err, "Failed to process credit operation")
	case grpc.ErrorReason_ERROR_REASON_INSUFFICIENT_CREDITS.String():
		if txHash, ok := errorInfo.Metadata[grpc.MetadataKey_METADATA_KEY_TRANSACTION_HASH.String()]; ok {
			return &gqlerror.Error{
				Message: fmt.Sprintf("insufficient credits, burn transaction initiated: %s", txHash),
				Err:     originalError,
				Extensions: map[string]any{
					"reason": errorInfo.Reason,
					"code":   http.StatusPaymentRequired,
				},
			}
		}
		return &gqlerror.Error{
			Message: fmt.Sprintf("insufficient credits for asset: %s", errorInfo.Metadata[grpc.MetadataKey_METADATA_KEY_ASSET_DID.String()]),
			Extensions: map[string]any{
				"reason": errorInfo.Reason,
				"code":   http.StatusPaymentRequired,
			},
		}
	default:
		return nil
	}
}

// GetSubjectAndTokenID gets the subject and token ID from the context.
func GetSubjectAndTokenID(ctx context.Context) (string, *big.Int, *gqlerror.Error) {
	validateClaims, ok := auth.GetValidatedClaims(ctx)
	if !ok || validateClaims.CustomClaims == nil {
		return "", nil, errorhandler.NewUnauthorizedErrorWithMsg(ctx, fmt.Errorf("failed to get validated claims"), "Unauthorized")
	}
	telemClaim, ok := validateClaims.CustomClaims.(*auth.TelemetryClaim)
	if !ok {
		return "", nil, errorhandler.NewUnauthorizedErrorWithMsg(ctx, fmt.Errorf("failed to get cast exchange custom claims"), "Unauthorized")
	}
	tokenIDBig, ok := new(big.Int).SetString(telemClaim.TokenID, 10)
	if !ok {
		return "", nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to parse token ID"), "Failed to parse token ID")
	}

	return validateClaims.RegisteredClaims.Subject, tokenIDBig, nil
}

func (d DCT) calculateCreditsBreakdown(ctx context.Context) (*pricing.CostBreakdown, *gqlerror.Error) {
	// If cost calculator is not available, fall back to default
	if d.CostCalculator == nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("cost calculator not available"), "Cost calculator not available")
	}

	// Get the GraphQL operation context
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil || opCtx.Operation == nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, ErrOperationNotSet, "No GraphQL operation context found")
	}

	// Calculate the cost with breakdown based on the query
	breakdown, err := d.CostCalculator.CalculateQueryCost(ctx, opCtx.Operation, opCtx.Variables)
	if err != nil {
		// Fallback to default cost on error
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, err, "Failed to calculate query cost")
	}

	return breakdown, nil
}

func (d DCT) calculateCredits(ctx context.Context) (uint64, *gqlerror.Error) {
	breakdown, gqlErr := d.calculateCreditsBreakdown(ctx)
	if gqlErr != nil {
		return 0, gqlErr
	}
	return breakdown.Cost, nil
}
