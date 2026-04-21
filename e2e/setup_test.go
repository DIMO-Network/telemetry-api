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
	"github.com/rs/zerolog"
)

// TestServices holds all singleton service instances.
type TestServices struct {
	Auth        *mockAuthServer
	FetchServer *mockFetchServer
	CH          *container.Container
	CT          *mockCreditTrackerServer
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
			Port:                8080,
			MonPort:             9090,
			TokenExchangeIssuer: "http://127.0.0.1:3003",
			VehicleNFTAddress:   common.HexToAddress("0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF"),
			MaxRequestDuration:  "1m",
			ChainID:             137,
		}

		// Setup services
		auth := setupAuthServer(t, settings)
		fetch := NewTestFetchAPI(t)
		ch := setupClickhouseContainer(t)
		ct := setupCreditTrackerContainer(t)

		// Create test settings
		settings.FetchAPIGRPCEndpoint = fetch.URL()
		settings.Clickhouse = ch.Config()
		settings.TokenExchangeJWTKeySetURL = auth.URL() + "/keys"
		settings.CreditTrackerEndpoint = ct.URL()

		testServices = &TestServices{
			Auth:        auth,
			FetchServer: fetch,
			CH:          ch,
			CT:          ct,
			Settings:    settings,
		}
		cleanup = func() {
			cleanupOnce.Do(func() {
				auth.Close()
				fetch.Close()
				ch.Terminate(context.Background())
				ct.Close()
			})
		}
	})
	srvcLock.Unlock()
	return testServices
}

func NewGraphQLServer(t *testing.T, settings config.Settings) *client.Client {
	t.Helper()

	zerolog.SetGlobalLevel(zerolog.PanicLevel)
	testLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &testLogger

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
