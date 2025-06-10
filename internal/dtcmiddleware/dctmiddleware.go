// Package dtcmiddleware provides a middleware for the Developer Credit Tracker.
package dtcmiddleware

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/credit-tracker/pkg/grpc"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/service/credittracker"
	"github.com/DIMO-Network/telemetry-api/pkg/errorhandler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

var defaultCreditAmount = int64(1)

// DCT provides a GraphQL middleware for the Developer Credit Tracker.
type DCT struct {
	Tracker *credittracker.Client
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
func NewDCT(tracker *credittracker.Client) *DCT {
	return &DCT{
		Tracker: tracker,
	}
}

// InterceptResponse intercepts GraphQL responses to handle errors from the credit tracker.
func (d DCT) InterceptResponse(
	ctx context.Context,
	next graphql.ResponseHandler,
) *graphql.Response {
	// Start timing the entire middleware operation
	middlewareTimer := prometheus.NewTimer(MiddlewareLatency)
	defer middlewareTimer.ObserveDuration()

	if d.Tracker == nil {
		zerolog.Ctx(ctx).Error().Msg("DCT is not enabled")
		// return graphql.ErrorResponse(ctx, "DCT is not enabled")
	}

	// Determine who to charge
	developerID, tokenID, gqlError := d.getSubjectAndTokenID(ctx)
	if gqlError != nil {
		zerolog.Ctx(ctx).Error().Err(gqlError).Msg("Failed to get subject and token ID")
		// return &graphql.Response{
		// 	Errors: gqlerror.List{gqlError},
		// }
		return next(ctx)
	}

	// Determine how many credits to charge
	credits, gqlError := d.calculateCredits(ctx)
	if gqlError != nil {
		zerolog.Ctx(ctx).Error().Err(gqlError).Msg("Failed to calculate credits")
		// return &graphql.Response{
		// 	Errors: gqlerror.List{gqlError},
		// }
		return next(ctx)
	}

	// Start timing the DCT request
	dctTimer := prometheus.NewTimer(DCTRequestLatency.WithLabelValues("deduct"))
	// Deduct the credits
	err := d.Tracker.DeductCredits(ctx, developerID, tokenID, credits)
	dctTimer.ObserveDuration()

	if err != nil {
		gqlError := processDCTErrorToGraphqlError(ctx, err)
		zerolog.Ctx(ctx).Error().Err(gqlError.Err).Msg("Failed to deduct credits")
		// return &graphql.Response{
		// 	Errors: gqlerror.List{gqlError},
		// }
		return next(ctx)
	}

	// Complete the request and get the response
	response := next(ctx)

	// If it's our fault the request failed, refund the credits
	if errorhandler.HasInternalError(&response.Errors) {
		// Start timing the refund operation
		refundTimer := prometheus.NewTimer(DCTRequestLatency.WithLabelValues("refund"))
		err := d.Tracker.RefundCredits(ctx, developerID, tokenID, credits)
		refundTimer.ObserveDuration()

		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to refund credits")
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

func (d DCT) getSubjectAndTokenID(ctx context.Context) (string, *big.Int, *gqlerror.Error) {
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

func (d DCT) calculateCredits(ctx context.Context) (int64, *gqlerror.Error) {
	// TODO: We can add logic here to determine what the base cost for a given operations should be
	return defaultCreditAmount, nil

}
