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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
}

// NewClient creates a new credit tracker client.
func NewClient(settings *config.Settings) (*Client, error) {
	conn, err := grpc.NewClient(settings.CreditTrackerEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create credit tracker client: %w", err)
	}
	ctClient := ctgrpc.NewCreditTrackerClient(conn)
	return &Client{
		conn:           conn,
		Endpoint:       settings.CreditTrackerEndpoint,
		RequestTimeout: 3 * time.Second,
		MaxRetries:     3,
		RetryTimeout:   100 * time.Millisecond,
		ctClient:       ctClient,
	}, nil
}

// DeductCredits deducts credits from the given developer license and token.
func (c *Client) DeductCredits(ctx context.Context, developerLicense string, tokenID *big.Int, amount int64) error {
	trackerCtx, cancel := context.WithTimeout(ctx, c.RequestTimeout)
	defer cancel()

	deductCredits := func() error {
		_, err := c.ctClient.DeductCredits(trackerCtx, &ctgrpc.CreditDeductRequest{
			DeveloperLicense: developerLicense,
			AssetDid: cloudevent.ERC721DID{
				ChainID:         c.chainID,
				ContractAddress: c.vehicleContractAddress,
				TokenID:         tokenID,
			}.String(),
			Amount: amount,
		})
		if err != nil {
			return fmt.Errorf("failed deduct request to credit tracker: %w", err)
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
func (c *Client) RefundCredits(ctx context.Context, developerLicense string, tokenID *big.Int, amount int64) error {
	trackerCtx, cancel := context.WithTimeout(ctx, c.RequestTimeout)
	defer cancel()

	refundCredits := func() error {
		_, err := c.ctClient.RefundCredits(trackerCtx, &ctgrpc.RefundCreditsRequest{
			DeveloperLicense: developerLicense,
			AssetDid: cloudevent.ERC721DID{
				ChainID:         c.chainID,
				ContractAddress: c.vehicleContractAddress,
				TokenID:         tokenID,
			}.String(),
			Amount: amount,
		})
		if err != nil {
			return fmt.Errorf("failed refund request to credit tracker: %w", err)
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
			time.Sleep(c.RetryTimeout)
			continue
		}
		return nil
	}
	return err
}
