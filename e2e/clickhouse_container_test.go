package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/clickhouse-infra/pkg/connect"
	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	sigmigrations "github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	indexmigrations "github.com/DIMO-Network/nameindexer/pkg/clickhouse/migrations"
	"github.com/stretchr/testify/require"
)

func setupClickhouseContainer(t *testing.T, indexDB string) *container.Container {
	t.Helper()
	ctx := context.Background()

	chContainer, err := container.CreateClickHouseContainer(ctx, chconfig.Settings{})
	if err != nil {
		t.Fatalf("Failed to create clickhouse container: %v", err)
	}

	db, err := chContainer.GetClickhouseAsDB()
	if err != nil {
		t.Fatalf("Failed to get clickhouse connection: %v", err)
	}

	err = sigmigrations.RunGoose(ctx, []string{"up", "-v"}, db)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	// Create index database
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + indexDB)
	require.NoError(t, err, "Failed to create index database")
	chConfig := chContainer.Config()
	chConfig.Database = indexDB
	fileDB := connect.GetClickhouseDB(&chConfig)
	if err != nil {
		t.Fatalf("Failed to get clickhouse connection: %v", err)
	}
	err = indexmigrations.RunGoose(ctx, []string{"up", "-v"}, fileDB)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return chContainer
}

// insertSignal inserts a test signal into Clickhouse
func insertSignal(t *testing.T, ch *container.Container, signals []vss.Signal) {
	t.Helper()

	conn, err := ch.GetClickHouseAsConn()
	require.NoError(t, err)
	batch, err := conn.PrepareBatch(context.Background(), fmt.Sprintf("INSERT INTO %s", vss.TableName))
	require.NoError(t, err)

	for _, sig := range signals {
		err := batch.AppendStruct(&sig)
		require.NoError(t, err, "Failed to append signal to batch")
	}
	err = batch.Send()
	require.NoError(t, err, "Failed to send batch")
}
