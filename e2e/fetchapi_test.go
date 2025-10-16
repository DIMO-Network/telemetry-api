// Package rpc provides the gRPC server implementation for the index repo service.
package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	pb "github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockFetchServer wraps the gRPC server and contains test configuration
type mockFetchServer struct {
	grpcServer       *grpc.Server
	listener         net.Listener
	port             int
	mutex            sync.Mutex
	cloudeventReturn []cloudevent.CloudEvent[json.RawMessage]
	pb.UnimplementedFetchServiceServer
	t *testing.T
}

// NewTestFetchAPI creates and starts a gRPC server on a random available port
func NewTestFetchAPI(t *testing.T) *mockFetchServer {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	// Create the gRPC server
	grpcServer := grpc.NewServer()
	testServer := &mockFetchServer{
		grpcServer: grpcServer,
		t:          t,
		listener:   listener,
		port:       listener.Addr().(*net.TCPAddr).Port,
	}

	pb.RegisterFetchServiceServer(grpcServer, testServer)

	// Start the server
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()

	// Wait a moment for the server to start
	time.Sleep(100 * time.Millisecond)

	return testServer

}

// Stop gracefully stops the test server
func (ts *mockFetchServer) Close() {
	ts.grpcServer.GracefulStop()

	if ts.listener != nil {
		ts.listener.Close() //nolint:errcheck
	}
}

func (ts *mockFetchServer) SetCloudEventReturn(ce ...cloudevent.CloudEvent[json.RawMessage]) {
	ts.mutex.Lock()
	ts.cloudeventReturn = ce
	ts.mutex.Unlock()
}

// GetAddress returns the full address of the server
func (ts *mockFetchServer) URL() string {
	return fmt.Sprintf("localhost:%d", ts.port)
}

// GetLatestIndex translates the gRPC call to the indexrepo type and returns the latest index for the given options.
func (s *mockFetchServer) GetLatestIndex(ctx context.Context, req *pb.GetLatestIndexRequest) (*pb.GetLatestIndexResponse, error) {
	return nil, nil
}

// ListIndex translates the pb call to the indexrepo type and fetches index keys for the given options.
func (s *mockFetchServer) ListIndex(ctx context.Context, req *pb.ListIndexesRequest) (*pb.ListIndexesResponse, error) {
	return nil, nil
}

// ListCloudEvents translates the pb call to the indexrepo type and fetches data for the given options.
func (s *mockFetchServer) ListCloudEvents(ctx context.Context, req *pb.ListCloudEventsRequest) (*pb.ListCloudEventsResponse, error) {
	respEvts := []*pb.CloudEvent{}
	for _, ce := range s.cloudeventReturn {
		respEvts = append(respEvts, pb.CloudEventToProto(ce))
	}
	return &pb.ListCloudEventsResponse{
		CloudEvents: respEvts,
	}, nil
}

// GetLatestCloudEvent translates the pb call to the indexrepo type and fetches the latest data for the given options.
func (s *mockFetchServer) GetLatestCloudEvent(ctx context.Context, req *pb.GetLatestCloudEventRequest) (*pb.GetLatestCloudEventResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return &pb.GetLatestCloudEventResponse{
		CloudEvent: pb.CloudEventToProto(s.cloudeventReturn[0]),
	}, nil

}

// ListCloudEventsFromIndex translates the pb call to the indexrepo type and fetches data for the given index keys.
func (s *mockFetchServer) ListCloudEventsFromIndex(ctx context.Context, req *pb.ListCloudEventsFromKeysRequest) (*pb.ListCloudEventsFromKeysResponse, error) {
	return nil, nil
}
