package e2e_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	ctgrpc "github.com/DIMO-Network/credit-tracker/pkg/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockCreditTrackerServer wraps the gRPC server and contains test configuration
type mockCreditTrackerServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
	port       int
	mutex      sync.Mutex
	responses  map[string]map[string]any // method -> request key -> response
	ctgrpc.UnimplementedCreditTrackerServer
	t *testing.T
}

// setupCreditTrackerContainer creates and starts a gRPC server on a random available port
func setupCreditTrackerContainer(t *testing.T) *mockCreditTrackerServer {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	// Create the gRPC server
	grpcServer := grpc.NewServer()
	testServer := &mockCreditTrackerServer{
		grpcServer: grpcServer,
		t:          t,
		listener:   listener,
		port:       listener.Addr().(*net.TCPAddr).Port,
		responses:  make(map[string]map[string]any),
	}

	ctgrpc.RegisterCreditTrackerServer(grpcServer, testServer)

	// Start the server
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()

	// Wait for server to be ready by attempting to connect
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for server to start")
		default:
			conn, err := net.Dial("tcp", testServer.URL())
			if err == nil {
				_ = conn.Close()
				return testServer
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Close gracefully stops the test server
func (ts *mockCreditTrackerServer) Close() {
	ts.grpcServer.GracefulStop()

	if ts.listener != nil {
		ts.listener.Close() //nolint:errcheck
	}
}

// URL returns the full address of the server
func (ts *mockCreditTrackerServer) URL() string {
	return ts.listener.Addr().String()
}

// SetResponse sets a response for a given method and request parameters
func (ts *mockCreditTrackerServer) SetResponse(method string, requestKey string, response any) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if ts.responses[method] == nil {
		ts.responses[method] = make(map[string]any)
	}
	ts.responses[method][requestKey] = response
}

// getRequestKey generates a unique key for a request based on its parameters
func getRequestKey(req any) string {
	switch r := req.(type) {
	case *ctgrpc.GetBalanceRequest:
		return fmt.Sprintf("%s:%s", r.DeveloperLicense, r.AssetDid)
	case *ctgrpc.CreditDeductRequest:
		return fmt.Sprintf("%s:%s", r.GetReferenceId(), r.GetAppName())
	case *ctgrpc.RefundCreditsRequest:
		return fmt.Sprintf("%s:%s", r.GetReferenceId(), r.GetAppName())
	default:
		return ""
	}
}

// CheckCredits implements the gRPC CheckCredits method
func (s *mockCreditTrackerServer) GetBalance(ctx context.Context, req *ctgrpc.GetBalanceRequest) (*ctgrpc.GetBalanceResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	requestKey := getRequestKey(req)
	if responses, ok := s.responses["CheckCredits"]; ok {
		if response, ok := responses[requestKey]; ok {
			if resp, ok := response.(*ctgrpc.GetBalanceResponse); ok {
				return resp, nil
			}
		}
	}

	// Default response if no custom response is set
	return &ctgrpc.GetBalanceResponse{
		RemainingCredits: 100,
	}, nil
}

// DeductCredits implements the gRPC DeductCredits method
func (s *mockCreditTrackerServer) DeductCredits(ctx context.Context, req *ctgrpc.CreditDeductRequest) (*ctgrpc.CreditDeductResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	requestKey := getRequestKey(req)
	if responses, ok := s.responses["DeductCredits"]; ok {
		if response, ok := responses[requestKey]; ok {
			if resp, ok := response.(*ctgrpc.CreditDeductResponse); ok {
				return resp, nil
			}
		}
	}

	// Default response if no custom response is set
	return &ctgrpc.CreditDeductResponse{}, nil
}

// RefundCredits implements the gRPC RefundCredits method
func (s *mockCreditTrackerServer) RefundCredits(ctx context.Context, req *ctgrpc.RefundCreditsRequest) (*ctgrpc.RefundCreditsResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	requestKey := getRequestKey(req)
	if responses, ok := s.responses["RefundCredits"]; ok {
		if response, ok := responses[requestKey]; ok {
			if resp, ok := response.(*ctgrpc.RefundCreditsResponse); ok {
				return resp, nil
			}
		}
	}

	// Default response if no custom response is set
	return &ctgrpc.RefundCreditsResponse{}, nil
}
