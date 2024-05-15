package repositories

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/rs/zerolog"
)

const (
	// MaxPageSize is the maximum page size for paginated results
	MaxPageSize = 100
)

// Repository is the base repository for all repositories.
type Repository struct {
	conn clickhouse.Conn
	Log  *zerolog.Logger
}

// NewRepository creates a new base repository.
// clientCAs is optional and can be nil.
func NewRepository(logger *zerolog.Logger, settings config.Settings, rootCAs *x509.CertPool) (*Repository, error) {
	addr := fmt.Sprintf("%s:%d", settings.ClickHouseHost, settings.ClickHouseTCPPort)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.ClickHouseUser,
			Password: settings.ClickHousePassword,
			Database: settings.ClickHouseDatabase,
		},
		TLS: &tls.Config{
			RootCAs: rootCAs,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}
	return &Repository{
		conn: conn,
		Log:  logger,
	}, nil
}
