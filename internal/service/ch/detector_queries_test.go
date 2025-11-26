package ch

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	"github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/require"
)

func TestDetectorQueriesWithSampleData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup ClickHouse container
	t.Log("=== Setting up ClickHouse container ===")
	chContainer, err := container.CreateClickHouseContainer(ctx, chconfig.Settings{})
	require.NoError(t, err, "Failed to create ClickHouse container")
	defer chContainer.Terminate(ctx)

	db, err := chContainer.GetClickhouseAsDB()
	require.NoError(t, err)

	// Run migrations
	t.Log("=== Running migrations ===")
	err = migrations.RunGoose(ctx, []string{"up", "-v"}, db)
	require.NoError(t, err, "Failed to run migrations")

	conn, err := chContainer.GetClickHouseAsConn()
	require.NoError(t, err)

	// Instantiate Service
	svc := &Service{conn: conn}

	// Use last 30 days as time range for all detectors
	tokenID := uint32(162682)
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30) // 30 days ago

	t.Logf("Time range: %s to %s", from.Format(time.RFC3339), to.Format(time.RFC3339))

	// ============================================================
	// PART 1: Test IgnitionDetector (uses signal_state_changes only)
	// ============================================================
	t.Log("\n=== Loading signal_state_changes for IgnitionDetector ===")
	loadSignalStateChanges(t, conn, "../../../.sample-data/signal_state_changes_2025-11-26.csv")

	t.Log("\n=== IgnitionDetector Query ===")
	printIgnitionDetectorQuery(t, tokenID, from, to)

	t.Logf("Testing IgnitionDetector for token %d", tokenID)
	segmentsIgn, err := svc.GetSegments(ctx, tokenID, from, to, model.DetectionMechanismIgnitionDetection, nil)
	require.NoError(t, err)
	t.Logf("Found %d segments (IgnitionDetector)", len(segmentsIgn))
	printSegments(t, "IgnitionDetector", segmentsIgn)

	// ============================================================
	// PART 2: Truncate and test FrequencyDetector/ChangePointDetector
	// ============================================================
	t.Log("\n=== Truncating tables ===")
	err = conn.Exec(ctx, "TRUNCATE TABLE signal")
	require.NoError(t, err)
	err = conn.Exec(ctx, "TRUNCATE TABLE signal_state_changes")
	require.NoError(t, err)

	t.Log("\n=== Loading signals for FrequencyDetector/ChangePointDetector ===")
	loadSignals(t, conn, "../../../.sample-data/signals_token_id_ 22892_2025-11-26.csv")

	// Test FrequencyDetector (uses signal FINAL with query-time dedup)
	t.Log("\n=== FrequencyDetector Query ===")
	printFrequencyDetectorQuery(t, tokenID, from, to)

	t.Logf("Testing FrequencyDetector for token %d", tokenID)
	segmentsFreq, err := svc.GetSegments(ctx, tokenID, from, to, model.DetectionMechanismFrequencyAnalysis, nil)
	require.NoError(t, err)
	t.Logf("Found %d segments (FrequencyDetector)", len(segmentsFreq))
	printSegments(t, "FrequencyDetector", segmentsFreq)

	// Test ChangePointDetector (uses signal FINAL with query-time dedup)
	t.Log("\n=== ChangePointDetector Query ===")
	printChangePointDetectorQuery(t, tokenID, from, to)

	t.Logf("Testing ChangePointDetector for token %d", tokenID)
	segmentsCP, err := svc.GetSegments(ctx, tokenID, from, to, model.DetectionMechanismChangePointDetection, nil)
	require.NoError(t, err)
	t.Logf("Found %d segments (ChangePointDetector)", len(segmentsCP))
	printSegments(t, "ChangePointDetector", segmentsCP)
}

func printSegments(t *testing.T, detectorName string, segments []*Segment) {
	if len(segments) == 0 {
		t.Logf("%s Results: No segments found", detectorName)
		return
	}

	// Load EST timezone
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Logf("Warning: could not load EST timezone: %v, using UTC", err)
		est = time.UTC
	}

	t.Logf("\n%s Results (EST):", detectorName)
	t.Log("┌─────┬─────────────────────────┬─────────────────────────┬──────────────┬──────────┐")
	t.Log("│  #  │     Start Time (EST)    │      End Time (EST)     │ Duration (s) │ Ongoing  │")
	t.Log("├─────┼─────────────────────────┼─────────────────────────┼──────────────┼──────────┤")

	for i, seg := range segments {
		startEST := seg.StartTime.In(est)
		endTimeStr := "nil (ongoing)"
		if seg.EndTime != nil {
			endTimeStr = seg.EndTime.In(est).Format("2006-01-02 15:04:05")
		}
		t.Logf("│ %3d │ %s │ %s │ %12d │ %8v │",
			i+1,
			startEST.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%-23s", endTimeStr),
			seg.DurationSeconds,
			seg.IsOngoing,
		)
	}
	t.Log("└─────┴─────────────────────────┴─────────────────────────┴──────────────┴──────────┘")
}

func printFrequencyDetectorQuery(t *testing.T, tokenID uint32, from, to time.Time) {
	windowSize := 60             // defaultWindowSizeSeconds
	signalThreshold := 12        // defaultSignalCountThreshold
	distinctSignalThreshold := 2 // defaultDistinctSignalCountThreshold

	query := fmt.Sprintf(`
SELECT
    toStartOfInterval(timestamp, INTERVAL %d second) AS window_start,
    toStartOfInterval(timestamp, INTERVAL %d second) + INTERVAL %d second AS window_end,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal FINAL
WHERE token_id = %d
  AND timestamp >= '%s'
  AND timestamp < '%s'
GROUP BY window_start
HAVING signal_count >= %d AND distinct_signal_count >= %d
ORDER BY window_start`,
		windowSize, windowSize, windowSize,
		tokenID,
		from.Format("2006-01-02 15:04:05"),
		to.Format("2006-01-02 15:04:05"),
		signalThreshold,
		distinctSignalThreshold)

	t.Logf("Query:\n%s\n", query)
}

func printChangePointDetectorQuery(t *testing.T, tokenID uint32, from, to time.Time) {
	windowSize := 60             // defaultCUSUMWindowSeconds
	signalThreshold := 12        // defaultCUSUMSignalCountThreshold
	distinctSignalThreshold := 2 // defaultCUSUMDistinctSignalCountThreshold

	query := fmt.Sprintf(`
SELECT
    toStartOfInterval(timestamp, INTERVAL %d second) AS window_start,
    toStartOfInterval(timestamp, INTERVAL %d second) + INTERVAL %d second AS window_end,
    count() AS signal_count,
    uniq(name) AS distinct_signal_count
FROM signal FINAL
WHERE token_id = %d
  AND timestamp >= '%s'
  AND timestamp < '%s'
GROUP BY window_start
HAVING signal_count >= %d AND distinct_signal_count >= %d
ORDER BY window_start`,
		windowSize, windowSize, windowSize,
		tokenID,
		from.Format("2006-01-02 15:04:05"),
		to.Format("2006-01-02 15:04:05"),
		signalThreshold,
		distinctSignalThreshold)

	t.Logf("Query:\n%s\n", query)
}

func printIgnitionDetectorQuery(t *testing.T, tokenID uint32, from, to time.Time) {
	query := fmt.Sprintf(`
SELECT
  timestamp,
  new_state,
  prev_state
FROM signal_state_changes FINAL
WHERE token_id = %d
  AND signal_name = 'isIgnitionOn'
  AND timestamp >= '%s'
  AND timestamp < '%s'
  AND prev_state != new_state
ORDER BY timestamp`,
		tokenID,
		from.Format("2006-01-02 15:04:05"),
		to.Format("2006-01-02 15:04:05"))

	t.Logf("Query:\n%s\n", query)
}

func loadSignalStateChanges(t *testing.T, conn clickhouse.Conn, path string) {
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.NotEmpty(t, records)

	// Helper to find column index
	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO signal_state_changes")
	require.NoError(t, err)

	for i := 1; i < len(records); i++ {
		row := records[i]

		tokenID, _ := strconv.ParseUint(row[colIdx["token_id"]], 10, 32)
		timestamp, _ := parseTimestamp(row[colIdx["timestamp"]])
		newState, _ := strconv.ParseFloat(row[colIdx["new_state"]], 64)
		prevState, _ := strconv.ParseFloat(row[colIdx["prev_state"]], 64)
		timeSince, _ := strconv.ParseUint(row[colIdx["time_since_prev_seconds"]], 10, 32)
		version, _ := strconv.ParseUint(row[colIdx["version"]], 10, 64)

		err := batch.Append(
			uint32(tokenID),
			row[colIdx["signal_name"]],
			timestamp,
			newState,
			prevState,
			uint32(timeSince),
			row[colIdx["source"]],
			row[colIdx["producer"]],
			row[colIdx["cloud_event_id"]],
			version,
		)
		require.NoError(t, err, "Failed to append row %d", i)
	}
	err = batch.Send()
	require.NoError(t, err)
}

func loadSignals(t *testing.T, conn clickhouse.Conn, path string) {
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.NotEmpty(t, records)

	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[h] = i
	}

	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO signal")
	require.NoError(t, err)

	for i := 1; i < len(records); i++ {
		row := records[i]

		tokenID, _ := strconv.ParseUint(row[colIdx["token_id"]], 10, 32)
		timestamp, _ := parseTimestamp(row[colIdx["timestamp"]])
		valueNumber, _ := strconv.ParseFloat(row[colIdx["value_number"]], 64)

		// Parse value_location JSON
		var loc struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Hdop      float64 `json:"hdop"`
		}
		_ = json.Unmarshal([]byte(row[colIdx["value_location"]]), &loc)

		err := batch.Append(
			uint32(tokenID),
			timestamp,
			row[colIdx["name"]],
			row[colIdx["source"]],
			row[colIdx["producer"]],
			row[colIdx["cloud_event_id"]],
			valueNumber,
			row[colIdx["value_string"]],
			[]interface{}{loc.Latitude, loc.Longitude, loc.Hdop},
		)
		require.NoError(t, err, "Failed to append row %d", i)
	}

	err = batch.Send()
	require.NoError(t, err)
	t.Logf("Loaded %d signal rows", len(records)-1)
}

func parseTimestamp(ts string) (time.Time, error) {
	// Try with microseconds
	t, err := time.Parse("2006-01-02 15:04:05.000000", ts)
	if err == nil {
		return t, nil
	}
	// Try without
	return time.Parse("2006-01-02 15:04:05", ts)
}
