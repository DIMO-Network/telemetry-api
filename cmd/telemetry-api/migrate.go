package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/pressly/goose/v3"
)

func RunGoose(ctx context.Context, settings *config.Settings, command string) error {
	db := getClickhouseDB(settings)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping clickhouse: %v", err)
	}

	cmd := os.Args[2]
	var args []string
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}
	if err := goose.SetDialect("clickhouse"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}
	err := goose.RunContext(ctx, cmd, getClickhouseDB(settings), ".", args...)
	if err != nil {
		return fmt.Errorf("failed to run goose command: %v", err)
	}
	return nil
}

func getClickhouseDB(settings *config.Settings) *sql.DB {
	addr := fmt.Sprintf("%s:%d", settings.ClickHouseHost, settings.ClickHouseTCPPort)
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.ClickHouseUser,
			Password: settings.ClickHousePassword,
		},
		DialTimeout: time.Minute * 30,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	return conn
}
