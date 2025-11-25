package e2e_test

import (
	"context"
	"fmt"
	"testing"

	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	sigmigrations "github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/stretchr/testify/require"
)

func setupClickhouseContainer(t *testing.T) *container.Container {
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

// insertEvent inserts test events into Clickhouse
func insertEvent(t *testing.T, ch *container.Container, events []vss.Event) {
	t.Helper()

	conn, err := ch.GetClickHouseAsConn()
	require.NoError(t, err)
	batch, err := conn.PrepareBatch(context.Background(), fmt.Sprintf("INSERT INTO %s", vss.EventTableName))
	require.NoError(t, err)

	for _, event := range events {
		err := batch.AppendStruct(&event)
		require.NoError(t, err, "Failed to append event to batch")
	}
	err = batch.Send()
	require.NoError(t, err, "Failed to send batch")
}
