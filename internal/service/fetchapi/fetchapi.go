package fetchapi

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/DIMO-Network/cloudevent"
	pb "github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// FetchAPIService implements the FetchAPIService interface using gRPC
type FetchAPIService struct {
	fetchGRPCAddr string
	client        pb.FetchServiceClient
	vehicleAddr   common.Address
	chainID       uint64
	once          sync.Once
}

// New creates a new instance of FetchAPIService with the specified server address,
// vehicle contract address, aftermarket contract address, and chain ID
func New(settings *config.Settings) *FetchAPIService {
	return &FetchAPIService{
		fetchGRPCAddr: settings.FetchAPIGRPCEndpoint,
		vehicleAddr:   settings.VehicleNFTAddress,
		chainID:       uint64(settings.ChainID),
	}
}

// GetLatestCloudEvent retrieves the most recent file content matching the provided search criteria
func (c *FetchAPIService) GetLatestCloudEvent(ctx context.Context, filter *pb.SearchOptions) (cloudevent.CloudEvent[json.RawMessage], error) {
	client, err := c.getClient()
	if err != nil {
		return cloudevent.CloudEvent[json.RawMessage]{}, err
	}
	resp, err := client.GetLatestCloudEvent(ctx, &pb.GetLatestCloudEventRequest{
		Options: filter,
	})
	if err != nil {
		return cloudevent.CloudEvent[json.RawMessage]{}, fmt.Errorf("failed to get latest file: %w", err)
	}
	return resp.GetCloudEvent().AsCloudEvent(), nil
}

// GetAllCloudEvents retrieves the most recent file content matching the provided search criteria
func (c *FetchAPIService) GetAllCloudEvents(ctx context.Context, filter *pb.SearchOptions, limit int32) ([]cloudevent.CloudEvent[json.RawMessage], error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.ListCloudEvents(ctx, &pb.ListCloudEventsRequest{
		Options: filter,
		Limit:   limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	cldEvts := []cloudevent.CloudEvent[json.RawMessage]{}
	for _, ce := range resp.GetCloudEvents() {
		cldEvts = append(cldEvts, ce.AsCloudEvent())
	}

	return cldEvts, nil
}

func (c *FetchAPIService) getClient() (pb.FetchServiceClient, error) {
	if c.client != nil {
		return c.client, nil
	}
	var err error
	c.once.Do(func() {
		var conn *grpc.ClientConn
		conn, err = grpc.NewClient(c.fetchGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			err = fmt.Errorf("failed to connect to Fetch API gRPC server: %w", err)
			return
		}
		c.client = pb.NewFetchServiceClient(conn)
	})
	return c.client, err
}
