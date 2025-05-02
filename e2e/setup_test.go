package e2e_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	"github.com/DIMO-Network/telemetry-api/internal/app"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/ethereum/go-ethereum/common"
)

// TestServices holds all singleton service instances.
type TestServices struct {
	Identity    *mockIdentityServer
	Auth        *mockAuthServer
	FetchServer *mockFetchServer
	CH          *container.Container
	Settings    config.Settings
}

var (
	testServices *TestServices
	once         sync.Once
	cleanupOnce  sync.Once
	srvcLock     sync.Mutex
	cleanup      func()
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
	if cleanup != nil {
		cleanup()
	}
}

// GetTestServices returns singleton instances of all test services
func GetTestServices(t *testing.T) *TestServices {
	t.Helper()
	srvcLock.Lock()
	once.Do(func() {
		settings := config.Settings{
			Port:                         8080,
			MonPort:                      9090,
			IdentityAPIReqTimeoutSeconds: 5,
			TokenExchangeIssuer:          "http://127.0.0.1:3003",
			VehicleNFTAddress:            common.HexToAddress("0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF"),
			ManufacturerNFTAddress:       common.HexToAddress("0x3b07e2A2ABdd0A9B8F7878bdE6487c502164B9dd"),
			MaxRequestDuration:           "1m",
			VINVCDataVersion:             "VINVCv1.0",
			POMVCDataVersion:             "POMVCv1.0",
			ChainID:                      137,
			DeviceLastSeenBinHrs:         3,
		}

		// Setup services
		identity := setupIdentityServer()
		auth := setupAuthServer(t, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress)
		fetch := NewTestFetchAPI(t)
		ch := setupClickhouseContainer(t)
		// Create test settings

		settings.FetchAPIGRPCEndpoint = fetch.URL()
		settings.Clickhouse = ch.Config()
		settings.IdentityAPIURL = identity.URL()
		settings.TokenExchangeJWTKeySetURL = auth.URL() + "/keys"

		testServices = &TestServices{
			Identity:    identity,
			Auth:        auth,
			FetchServer: fetch,
			CH:          ch,
			Settings:    settings,
		}
		cleanup = func() {
			cleanupOnce.Do(func() {
				identity.Close()
				auth.Close()
				fetch.Close()
				ch.Terminate(context.Background())
			})
		}
	})
	srvcLock.Unlock()
	return testServices
}

func NewGraphQLServer(t *testing.T, settings config.Settings) *client.Client {
	t.Helper()

	application, err := app.New(settings)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	t.Cleanup(application.Cleanup)

	return client.New(application.Handler)
}

func WithToken(token string) client.Option {
	return client.AddHeader("Authorization", "Bearer "+token)
}
