package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	"github.com/DIMO-Network/model-garage/pkg/cloudevent"
	"github.com/DIMO-Network/nameindexer/pkg/clickhouse/indexrepo"
	"github.com/DIMO-Network/telemetry-api/internal/app"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/rs/zerolog"
)

// TestServices holds all singleton service instances.
type TestServices struct {
	Identity *mockIdentityServer
	Auth     *mockAuthServer
	S3Server *mockS3Server
	CH       *container.Container
	Settings config.Settings
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
			VehicleNFTAddress:            "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF",
			ManufacturerNFTAddress:       "0x3b07e2A2ABdd0A9B8F7878bdE6487c502164B9dd",
			MaxRequestDuration:           "1m",
			S3AWSRegion:                  "us-east-1",
			S3AWSAccessKeyID:             "minioadmin",
			S3AWSSecretAccessKey:         "minioadmin",
			VCBucket:                     "test.vc.bucket", // TLDR keep the dots; If we don't use a non DNS resolved bucket name then the bucket lookup will attempt to use BUCKET_NAME.baseEndpoint
			VINVCDataType:                "VINVCv1.0",
			POMVCDataType:                "POMVCv1.0",
			ChainID:                      137,
			ClickhouseFileIndexDatabase:  "file_index",
			DeviceLastSeenBinHrs:         3,
		}

		// Setup services
		identity := setupIdentityServer()
		auth := setupAuthServer(t, settings.VehicleNFTAddress, settings.ManufacturerNFTAddress)
		s3 := setupS3Server(t, settings.VCBucket)
		ch := setupClickhouseContainer(t, settings.ClickhouseFileIndexDatabase)
		// Create test settings

		settings.CLickhouse = ch.Config()
		settings.S3BaseEndpoint = s3.BaseEndpoint()
		settings.IdentityAPIURL = identity.URL()
		settings.TokenExchangeJWTKeySetURL = auth.URL() + "/keys"

		testServices = &TestServices{
			Identity: identity,
			Auth:     auth,
			S3Server: s3,
			CH:       ch,
			Settings: settings,
		}
		cleanup = func() {
			cleanupOnce.Do(func() {
				identity.Close()
				auth.Close()
				s3.Cleanup(t)
				ch.Terminate(context.Background())
			})
		}
	})
	srvcLock.Unlock()
	return testServices
}

func StoreSampleVC(ctx context.Context, idxSrv *indexrepo.Service, bucket string, testVC string) error {
	hdr := cloudevent.CloudEventHeader{}
	err := json.Unmarshal([]byte(testVC), &hdr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal VC: %w", err)
	}

	err = idxSrv.StoreObject(ctx, bucket, &hdr, []byte(testVC))
	if err != nil {
		return fmt.Errorf("failed to store VC: %w", err)
	}
	return nil
}

func NewGraphQLServer(t *testing.T, settings config.Settings) *client.Client {
	t.Helper()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	application, err := app.New(settings, &logger)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	t.Cleanup(application.Cleanup)

	return client.New(application.Handler)
}

func WithToken(token string) client.Option {
	return client.AddHeader("Authorization", "Bearer "+token)
}
