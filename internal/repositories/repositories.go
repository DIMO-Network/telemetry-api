package repositories

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/service/deviceapi"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// MaxPageSize is the maximum page size for paginated results
	MaxPageSize = 100
)

// Repository is the base repository for all repositories.
type Repository struct {
	conn      clickhouse.Conn
	Log       *zerolog.Logger
	deviceAPI *deviceapi.Service
}

// NewRepository creates a new base repository.
func NewRepository(logger *zerolog.Logger, settings config.Settings) (*Repository, error) {
	addr := fmt.Sprintf("%s:%d", settings.ClickHouseHost, settings.ClickHouseTCPPort)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.ClickHouseUser,
			Password: settings.ClickHousePassword,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}
	devicesConn, err := grpc.Dial(settings.DevicesAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial devices api: %w", err)
	}
	deviceAPI := deviceapi.NewService(devicesConn)
	return &Repository{
		conn:      conn,
		Log:       logger,
		deviceAPI: deviceAPI,
	}, nil
}
