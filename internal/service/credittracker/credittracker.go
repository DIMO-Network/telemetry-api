package credittracker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/cloudevent"
	ctgrpc "github.com/DIMO-Network/credit-tracker/pkg/grpc"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Client implements the Client interface.
type Client struct {
	conn                   *grpc.ClientConn
	Endpoint               string
	RequestTimeout         time.Duration
	MaxRetries             int
	RetryTimeout           time.Duration
	ctClient               ctgrpc.CreditTrackerClient
	chainID                uint64
	vehicleContractAddress common.Address
	appName                string
}

// NewClient creates a new credit tracker client.
func NewClient(settings *config.Settings, appName string) (*Client, error) {
	conn, err := grpc.NewClient(settings.CreditTrackerEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create credit tracker client: %w", err)
	}
	ctClient := ctgrpc.NewCreditTrackerClient(conn)
	return &Client{
		conn:                   conn,
		Endpoint:               settings.CreditTrackerEndpoint,
		RequestTimeout:         3 * time.Second,
		MaxRetries:             3,
		RetryTimeout:           100 * time.Millisecond,
		ctClient:               ctClient,
		appName:                appName,
		chainID:                settings.ChainID,
		vehicleContractAddress: settings.VehicleNFTAddress,
	}, nil
}

// DeductCredits deducts credits from the given developer license and token.
func (c *Client) DeductCredits(ctx context.Context, referenceID string, developerLicense string, tokenID *big.Int, amount uint64) error {
	trackerCtx, cancel := context.WithTimeout(ctx, c.RequestTimeout)
	defer cancel()

	deductCredits := func() error {
		_, err := c.ctClient.DeductCredits(trackerCtx, &ctgrpc.CreditDeductRequest{
			ReferenceId:      referenceID,
			AppName:          c.appName,
			DeveloperLicense: developerLicense,
			AssetDid: cloudevent.ERC721DID{
				ChainID:         c.chainID,
				ContractAddress: c.vehicleContractAddress,
				TokenID:         tokenID,
			}.String(),
			Amount: amount,
		})
		if err != nil {
			return fmt.Errorf("failed to send deduct request to credit tracker: %w", err)
		}
		return nil
	}
	err := c.runWithRetry(ctx, deductCredits)
	if err != nil {
		return err
	}
	return nil
}

// RefundCredits refunds credits from the given developer license and token.
func (c *Client) RefundCredits(ctx context.Context, referenceID string) error {
	trackerCtx, cancel := context.WithTimeout(ctx, c.RequestTimeout)
	defer cancel()

	refundCredits := func() error {
		_, err := c.ctClient.RefundCredits(trackerCtx, &ctgrpc.RefundCreditsRequest{
			ReferenceId: referenceID,
			AppName:     c.appName,
		})
		if err != nil {
			return fmt.Errorf("failed to send refund request to credit tracker: %w", err)
		}
		return nil
	}
	err := c.runWithRetry(ctx, refundCredits)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) runWithRetry(ctx context.Context, f func() error) error {
	var err error
	for i := 0; i < c.MaxRetries; i++ {
		if err = f(); err != nil {
			// Check if this is a bad request error that shouldn't be retried
			if isBadRequestError(err) {
				return err
			}
			time.Sleep(c.RetryTimeout)
			continue
		}
		return nil
	}
	return err
}

// isBadRequestError checks if the error is a bad request that shouldn't be retried
func isBadRequestError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// Check gRPC status code for bad request errors
	if st.Code() == codes.InvalidArgument || st.Code() == codes.NotFound || st.Code() == codes.PermissionDenied {
		return true
	}

	// Check error details for specific bad request reasons
	for _, detail := range st.Details() {
		if errorInfo, ok := detail.(*errdetails.ErrorInfo); ok {
			switch errorInfo.Reason {
			case ctgrpc.ErrorReason_ERROR_REASON_INVALID_ASSET_DID.String():
				return true
			case ctgrpc.ErrorReason_ERROR_REASON_INSUFFICIENT_CREDITS.String():
				return true
			}
		}
	}

	return false
}
