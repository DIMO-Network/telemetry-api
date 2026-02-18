package e2e_test

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	sigmigrations "github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/stretchr/testify/require"
)

const loadBatchSize = 5000

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

// LoadSampleDataInto loads signal and event CSVs into the ClickHouse container.
func LoadSampleDataInto(t *testing.T, ch *container.Container, signalPath, eventPath string) {
	t.Helper()

	sf, err := os.Open(signalPath)
	require.NoError(t, err)
	defer func() { _ = sf.Close() }()
	br := bufio.NewReader(sf)
	if peek, _ := br.Peek(3); len(peek) == 3 && peek[0] == 0xef && peek[1] == 0xbb && peek[2] == 0xbf {
		_, _ = br.Discard(3)
	}
	sr := csv.NewReader(br)
	sr.FieldsPerRecord = -1
	sigRows, err := sr.ReadAll()
	require.NoError(t, err)
	if len(sigRows) < 2 {
		t.Fatalf("signal CSV has no data rows")
	}
	signals := make([]vss.Signal, 0, len(sigRows)-1)
	for _, row := range sigRows[1:] {
		if len(row) < 9 {
			continue
		}
		tokenID, _ := strconv.ParseUint(row[0], 10, 32)
		ts, _ := time.Parse("2006-01-02 15:04:05.000000", row[1])
		var loc vss.Location
		if len(row[8]) > 0 {
			_ = json.Unmarshal([]byte(row[8]), &loc)
		}
		valNum, _ := strconv.ParseFloat(row[6], 64)
		signals = append(signals, vss.Signal{
			TokenID:       uint32(tokenID),
			Timestamp:     ts,
			Name:          row[2],
			Source:        row[3],
			Producer:      row[4],
			CloudEventID:  row[5],
			ValueNumber:   valNum,
			ValueString:   row[7],
			ValueLocation: loc,
		})
	}
	for i := 0; i < len(signals); i += loadBatchSize {
		end := i + loadBatchSize
		if end > len(signals) {
			end = len(signals)
		}
		insertSignal(t, ch, signals[i:end])
	}

	ef, err := os.Open(eventPath)
	require.NoError(t, err)
	defer func() { _ = ef.Close() }()
	erBr := bufio.NewReader(ef)
	if peek, _ := erBr.Peek(3); len(peek) == 3 && peek[0] == 0xef && peek[1] == 0xbb && peek[2] == 0xbf {
		_, _ = erBr.Discard(3)
	}
	er := csv.NewReader(erBr)
	er.FieldsPerRecord = -1
	evRows, err := er.ReadAll()
	require.NoError(t, err)
	if len(evRows) < 2 {
		t.Fatalf("event CSV has no data rows")
	}
	events := make([]vss.Event, 0, len(evRows)-1)
	for _, row := range evRows[1:] {
		if len(row) < 9 {
			continue
		}
		ts, _ := time.Parse("2006-01-02 15:04:05.000000", row[5])
		durNs, _ := strconv.ParseUint(row[6], 10, 64)
		var tags []string
		if len(row[8]) > 0 {
			_ = json.Unmarshal([]byte(row[8]), &tags)
		}
		events = append(events, vss.Event{
			Subject:      row[0],
			Source:       row[1],
			Producer:     row[2],
			CloudEventID: row[3],
			Name:         row[4],
			Timestamp:    ts,
			DurationNs:   durNs,
			Metadata:     row[7],
			Tags:         tags,
		})
	}
	for i := 0; i < len(events); i += loadBatchSize {
		end := i + loadBatchSize
		if end > len(events) {
			end = len(events)
		}
		insertEvent(t, ch, events[i:end])
	}
}
