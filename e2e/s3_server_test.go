package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
)

type mockS3Server struct {
	container    testcontainers.Container
	client       *s3.Client
	vcBucket     string
	baseEndpoint *string
}

func setupS3Server(t *testing.T, vcBucket string) *mockS3Server {
	t.Helper()
	ctx := context.Background()

	minioContainer, err := minio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z",
		testcontainers.WithEnv(map[string]string{
			"MINIO_ACCESS_KEY": "minioadmin",
			"MINIO_SECRET":     "minioadmin",
		}),
		testcontainers.WithHostPortAccess(9000),
	)

	if err != nil {
		require.NoError(t, err)
	}

	mappedPort, err := minioContainer.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}
	endpont := fmt.Sprintf("http://localhost:%s", mappedPort.Port())

	// Create S3 client
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", "")),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("Failed to create AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = &endpont
	})
	s3srv := &mockS3Server{
		container:    minioContainer,
		client:       client,
		vcBucket:     vcBucket,
		baseEndpoint: &endpont,
	}

	// Create VC bucket
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &vcBucket,
	})
	if err != nil {
		t.Fatalf("Failed to create VC bucket: %v", err)
	}
	return s3srv
}

func (s *mockS3Server) GetClient() *s3.Client {
	return s.client
}

func (s *mockS3Server) BaseEndpoint() *string {
	return s.baseEndpoint
}

func (s *mockS3Server) GetVCBucket() string {
	return s.vcBucket
}

func (s *mockS3Server) Cleanup(t *testing.T) {
	t.Helper()
	if err := s.container.Terminate(context.Background()); err != nil {
		t.Logf("Failed to terminate S3 container: %v", err)
	}
}
