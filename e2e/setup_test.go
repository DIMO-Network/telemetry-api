package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	"github.com/DIMO-Network/model-garage/pkg/cloudevent"
	"github.com/DIMO-Network/nameindexer"
	"github.com/DIMO-Network/nameindexer/pkg/clickhouse/indexrepo"
	"github.com/DIMO-Network/telemetry-api/internal/app"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/rs/zerolog"
)

// TestServices holds all singleton service instances
type TestServices struct {
	Identity  *mockIdentityServer
	Auth      *mockAuthServer
	IndexRepo *indexrepo.Service
	CH        *container.Container
	Settings  config.Settings
}

var (
	testServices *TestServices
	once         sync.Once
	cleanupOnce  sync.Once
	cleanup      func()
)

// GetTestServices returns singleton instances of all test services
func GetTestServices(t *testing.T) *TestServices {
	t.Helper()

	once.Do(func() {
		// Setup services
		identity := setupIdentityServer()
		auth := setupAuthServer(t)
		// s3 := setupS3Server(t, "test-vc-bucket", "test-pom-bucket")
		ch := setupClickhouseContainer(t)
		// chConn, err := ch.GetClickHouseAsConn()
		// require.NoError(t, err)
		// indexService := indexrepo.New(chConn, s3.GetClient())
		// Create test settings
		settings := config.Settings{
			Port:                         8080,
			MonPort:                      9090,
			IdentityAPIURL:               identity.URL(),
			IdentityAPIReqTimeoutSeconds: 5,
			TokenExchangeJWTKeySetURL:    auth.URL() + "/keys",
			TokenExchangeIssuer:          "http://127.0.0.1:3003",
			VehicleNFTAddress:            "0x45fbCD3ef7361d156e8b16F5538AE36DEdf61Da8",
			ManufacturerNFTAddress:       "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF",
			MaxRequestDuration:           "1m",
			S3AWSRegion:                  "us-east-1",
			S3AWSAccessKeyID:             "minioadmin",
			S3AWSSecretAccessKey:         "minioadmin",
			// VCBucket:                     s3.GetVCBucket(), // Use bucket from S3 server
			VINVCDataType:        "vin",
			POMVCDataType:        "pom",
			ChainID:              1,
			CLickhouse:           ch.Config(),
			DeviceLastSeenBinHrs: 3,
		}
		// storeSampleVC(context.Background(), indexService, s3.GetVCBucket())
		testServices = &TestServices{
			Identity: identity,
			Auth:     auth,
			// IndexRepo: indexService,
			CH:       ch,
			Settings: settings,
		}

		// Setup cleanup function
		cleanup = func() {
			cleanupOnce.Do(func() {
				identity.Close()
				auth.Close()
				// s3.Cleanup(t)
				ch.Terminate(context.Background())
			})
		}
	})

	// Register cleanup to run after all tests
	t.Cleanup(func() {
		cleanup()
	})

	return testServices

}

const testVC = `{"id":"2oYEQJDQFRYCI1UIf6QTy92MS3G","source":"0x0000000000000000000000000000000000000000","producer":"did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA_31653","specversion":"1.0","subject":"did:nft:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF_39718","time":"2024-11-08T03:54:51.291563165Z","type":"dimo.verifiablecredential","datacontenttype":"application/json","dataversion":"VINVCv1.0","data":{"@context":["https://www.w3.org/ns/credentials/v2",{"vehicleIdentificationNumber":"https://schema.org/vehicleIdentificationNumber"},"https://attestation-api.dimo.zone/v1/vc/context"],"id":"urn:uuid:0b74cd3b-5998-4436-ada3-22dd6cfe2b3c","type":["VerifiableCredential","Vehicle"],"issuer":"https://attestation-api.dimo.zone/v1/vc/keys","validFrom":"2024-11-08T03:54:51Z","validTo":"2024-11-10T00:00:00Z","credentialSubject":{"id":"did:nft:137_erc721:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF_39718","vehicleTokenId":39718,"vehicleContractAddress":"eth:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF","vehicleIdentificationNumber":"3eMCZ5AN5PM647548","recordedBy":"did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA_31653","recordedAt":"2024-11-08T00:40:29Z"},"credentialStatus":{"id":"https://attestation-api.dimo.zone/v1/vc/status/39718","type":"BitstringStatusListEntry","statusPurpose":"revocation","statusListIndex":0,"statusListCredential":"https://attestation-api.dimo.zone/v1/vc/status"},"proof":{"type":"DataIntegrityProof","cryptosuite":"ecdsa-rdfc-2019","verificationMethod":"https://attestation-api.dimo.zone/v1/vc/keys#key1","created":"2024-11-08T03:54:51Z","proofPurpose":"assertionMethod","proofValue":"381yXZFRShe2rrr9A3VFGeXS9izouz7Gor1GTb6Mwjkpge8eEn814QivzssEogoLKrzGN6WPKWBQrFLgfcsUhuaAWhS421Dn"}}}`

func storeSampleVC(ctx context.Context, idxSrv *indexrepo.Service, bucket string) error {
	hdr := cloudevent.CloudEventHeader{}
	json.Unmarshal([]byte(testVC), &hdr)
	cloudIdx, err := nameindexer.CloudEventToCloudIndex(&hdr, nameindexer.DefaultSecondaryFiller)
	if err != nil {
		return fmt.Errorf("failed to convert VC to cloud index: %w", err)
	}

	err = idxSrv.StoreCloudEventObject(ctx, cloudIdx, bucket, []byte(testVC))
	if err != nil {
		return fmt.Errorf("failed to store VC: %w", err)
	}
	return nil
}

func NewGraphQLServer(t *testing.T, settings config.Settings) *httptest.Server {
	t.Helper()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	application, err := app.New(settings, &logger)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	t.Cleanup(application.Cleanup)

	return httptest.NewServer(application.Handler)
}
