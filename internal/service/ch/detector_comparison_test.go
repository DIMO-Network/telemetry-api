package ch

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	sigmigrations "github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/migrations"
	"github.com/stretchr/testify/require"
)

// ComparisonResult holds the results of running a detection mechanism
type ComparisonResult struct {
	Mechanism     string
	SegmentCount  int
	QueryTime     time.Duration
	TotalDuration int32 // Sum of all segment durations (seconds)
	AvgDuration   float64
	MinDuration   int32
	MaxDuration   int32
	OngoingCount  int
	Segments      []*Segment
}

// TestDetectorComparison runs all three detection mechanisms on real vehicle data
// and produces a comprehensive comparison table
func TestDetectorComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup ClickHouse container
	t.Log("=== Setting up ClickHouse container ===")
	chContainer, err := container.CreateClickHouseContainer(ctx, chconfig.Settings{})
	require.NoError(t, err, "Failed to create ClickHouse container")
	defer chContainer.Terminate(ctx)

	// Run migrations
	db, err := chContainer.GetClickhouseAsDB()
	require.NoError(t, err, "Failed to get DB connection")

	t.Log("=== Running migrations ===")
	err = sigmigrations.RunGoose(ctx, []string{"up", "-v"}, db)
	require.NoError(t, err, "Failed to run base migrations")

	err = migrations.RunGoose(ctx, []string{"up", "-v"}, db)
	require.NoError(t, err, "Failed to run telemetry-api migrations")

	conn, err := chContainer.GetClickHouseAsConn()
	require.NoError(t, err, "Failed to get ClickHouse connection")

	// Clean tables
	t.Log("=== Cleaning tables ===")
	conn.Exec(ctx, "TRUNCATE TABLE signal")
	conn.Exec(ctx, "TRUNCATE TABLE signal_state_changes")
	conn.Exec(ctx, "TRUNCATE TABLE signal_window_aggregates")

	// Load real vehicle data
	t.Log("=== Loading real vehicle data ===")
	signals, err := loadSignalsFromCSV("../../../real_vehicle_data copy.csv")
	require.NoError(t, err, "Failed to load CSV data")
	require.NotEmpty(t, signals, "CSV should contain signals")

	// Sort chronologically
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Timestamp.Before(signals[j].Timestamp)
	})

	t.Logf("Loaded %d signals from CSV", len(signals))

	// Get time range
	firstSignal := signals[0].Timestamp
	lastSignal := signals[len(signals)-1].Timestamp
	t.Logf("Data time range: %s to %s (%.1f days)",
		firstSignal.Format("2006-01-02"),
		lastSignal.Format("2006-01-02"),
		lastSignal.Sub(firstSignal).Hours()/24.0,
	)

	// Insert signals in batches
	t.Log("=== Inserting signals ===")
	batchSize := 1000
	for i := 0; i < len(signals); i += batchSize {
		end := i + batchSize
		if end > len(signals) {
			end = len(signals)
		}
		insertSignals(t, chContainer, signals[i:end])
	}
	t.Logf("Inserted %d signals", len(signals))

	// For ignition detection: manually populate state changes
	t.Log("=== Populating signal_state_changes for ignition detection ===")
	err = conn.Exec(ctx, `
		INSERT INTO signal_state_changes
		SELECT * FROM (
		  SELECT
		    token_id,
		    name as signal_name,
		    timestamp,
		    value_number as new_state,
		    lagInFrame(value_number, 1, -1) OVER (
		      PARTITION BY token_id, name 
		      ORDER BY timestamp
		    ) as prev_state,
		    dateDiff('second',
		      lagInFrame(timestamp, 1, timestamp) OVER (
		        PARTITION BY token_id, name 
		        ORDER BY timestamp
		      ),
		      timestamp
		    ) as time_since_prev_seconds,
		    source,
		    producer,
		    cloud_event_id
		  FROM signal
		  WHERE name IN ('isIgnitionOn')
		) WHERE prev_state != new_state
	`)
	require.NoError(t, err, "Failed to populate signal_state_changes")

	// Wait for frequency analysis materialized view to populate
	t.Log("=== Waiting for materialized views to populate ===")
	time.Sleep(500 * time.Millisecond)

	// Verify MV populated
	var windowCount uint64
	err = conn.QueryRow(ctx, "SELECT count() FROM signal_window_aggregates").Scan(&windowCount)
	require.NoError(t, err)
	t.Logf("Frequency analysis windows created: %d", windowCount)

	var stateChangeCount uint64
	err = conn.QueryRow(ctx, "SELECT count() FROM signal_state_changes WHERE signal_name = 'isIgnitionOn'").Scan(&stateChangeCount)
	require.NoError(t, err)
	t.Logf("Ignition state changes: %d", stateChangeCount)

	// Create service
	svc := &Service{conn: conn}

	// Define query parameters
	tokenID := uint32(186612)
	from := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)

	t.Logf("\n=== Query Parameters ===")
	t.Logf("Token ID: %d", tokenID)
	t.Logf("Time Range: %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	t.Log("Strategy: Day-by-day queries\n")

	// Run all detection mechanisms
	mechanisms := []struct {
		name      string
		mechanism model.DetectionMechanism
	}{
		{"Ignition Detection", model.DetectionMechanismIgnitionDetection},
		{"Frequency Analysis", model.DetectionMechanismFrequencyAnalysis},
		{"Change Point Detection", model.DetectionMechanismChangePointDetection},
	}

	// Print header for day-by-day output
	t.Log("╔═══════════════════════════════════════════════════════════════════════════════════════════════════╗")
	t.Log("║                              DAY-BY-DAY DETECTION RESULTS                                         ║")
	t.Log("╚═══════════════════════════════════════════════════════════════════════════════════════════════════╝")
	t.Log("")

	// Collect results per mechanism
	mechanismSegments := make(map[string][]*Segment)
	mechanismTotalTime := make(map[string]time.Duration)

	// Iterate through each day
	currentDay := from
	for currentDay.Before(to) {
		dayStart := currentDay
		dayEnd := currentDay.Add(24 * time.Hour)
		if dayEnd.After(to) {
			dayEnd = to
		}

		dayStr := currentDay.Format("2006-01-02")
		t.Logf("\n=== %s ===", dayStr)
		t.Log("  ┌─────────────────┬──────────┬──────────────┬──────────────┐")
		t.Log("  │ Mechanism       │ Segments │ Total Time   │ Query Time   │")
		t.Log("  ├─────────────────┼──────────┼──────────────┼──────────────┤")

		for _, mech := range mechanisms {
			start := time.Now()
			segments, err := svc.GetSegments(ctx, tokenID, dayStart, dayEnd, mech.mechanism, nil)
			queryTime := time.Since(start)

			require.NoError(t, err, "GetSegments should succeed for %s on %s", mech.name, dayStr)

			// Calculate daily stats
			var dayTotalDuration int32
			for _, seg := range segments {
				dayTotalDuration += seg.DurationSeconds
			}

			// Accumulate for overall results
			mechanismSegments[mech.name] = append(mechanismSegments[mech.name], segments...)
			mechanismTotalTime[mech.name] += queryTime

			mechanismShort := mech.name
			if len(mechanismShort) > 15 {
				mechanismShort = mechanismShort[:12] + "..."
			}

			t.Logf("  │ %-15s │ %8d │ %10.1fh │ %11s │",
				mechanismShort,
				len(segments),
				float64(dayTotalDuration)/3600.0,
				formatDuration(queryTime))
		}

		t.Log("  └─────────────────┴──────────┴──────────────┴──────────────┘")

		// Print detailed trip breakdown for this day
		printDayTripDetails(t, dayStr, mechanisms, mechanismSegments, dayStart, dayEnd)

		currentDay = dayEnd
	}

	// Build final aggregated results
	var results []ComparisonResult

	for _, mech := range mechanisms {
		allSegments := mechanismSegments[mech.name]

		// Calculate statistics
		var totalDuration int32
		var minDuration int32 = 999999
		var maxDuration int32
		var ongoingCount int

		for _, seg := range allSegments {
			totalDuration += seg.DurationSeconds
			if seg.DurationSeconds < minDuration {
				minDuration = seg.DurationSeconds
			}
			if seg.DurationSeconds > maxDuration {
				maxDuration = seg.DurationSeconds
			}
			if seg.IsOngoing {
				ongoingCount++
			}
		}

		avgDuration := 0.0
		if len(allSegments) > 0 {
			avgDuration = float64(totalDuration) / float64(len(allSegments)) / 60.0
		}

		results = append(results, ComparisonResult{
			Mechanism:     mech.name,
			SegmentCount:  len(allSegments),
			QueryTime:     mechanismTotalTime[mech.name],
			TotalDuration: totalDuration,
			AvgDuration:   avgDuration,
			MinDuration:   minDuration,
			MaxDuration:   maxDuration,
			OngoingCount:  ongoingCount,
			Segments:      allSegments,
		})
	}

	t.Log("\n")

	// Print comprehensive comparison table
	printComparisonTable(t, results)

	// Verify all mechanisms detected segments
	for _, result := range results {
		require.NotEmpty(t, result.Segments, "%s should detect segments", result.Mechanism)
	}
}

// printComparisonTable prints a formatted comparison table
func printComparisonTable(t *testing.T, results []ComparisonResult) {
	t.Log("\n")
	t.Log("╔═══════════════════════════════════════════════════════════════════════════════╗")
	t.Log("║                    DETECTION MECHANISM COMPARISON                         ║")
	t.Log("╠═══════════════════════════════════════════════════════════════════════════╣")
	t.Log("║ Metric                │ Ignition   │ Frequency  │ Change Pt  ║")
	t.Log("╠═══════════════════════╪════════════╪════════════╪════════════╣")

	// Helper to get value or dash if not enough results
	getValue := func(idx int, fn func(ComparisonResult) string) string {
		if idx < len(results) {
			return fn(results[idx])
		}
		return "-"
	}

	// Segments Detected
	t.Logf("║ Segments Detected     │ %-10s │ %-10s │ %-10s ║",
		getValue(0, func(r ComparisonResult) string { return fmt.Sprintf("%d", r.SegmentCount) }),
		getValue(1, func(r ComparisonResult) string { return fmt.Sprintf("%d", r.SegmentCount) }),
		getValue(2, func(r ComparisonResult) string { return fmt.Sprintf("%d", r.SegmentCount) }),
	)

	// Query Time
	t.Logf("║ Query Time            │ %-10s │ %-10s │ %-10s ║",
		getValue(0, func(r ComparisonResult) string { return formatDuration(r.QueryTime) }),
		getValue(1, func(r ComparisonResult) string { return formatDuration(r.QueryTime) }),
		getValue(2, func(r ComparisonResult) string { return formatDuration(r.QueryTime) }),
	)

	// Average Duration
	t.Logf("║ Avg Duration (min)    │ %-10.1f │ %-10.1f │ %-10.1f ║",
		getFloat(results, 0, func(r ComparisonResult) float64 { return r.AvgDuration }),
		getFloat(results, 1, func(r ComparisonResult) float64 { return r.AvgDuration }),
		getFloat(results, 2, func(r ComparisonResult) float64 { return r.AvgDuration }),
	)

	// Total Duration
	t.Logf("║ Total Time (hours)    │ %-10.1f │ %-10.1f │ %-10.1f ║",
		getFloat(results, 0, func(r ComparisonResult) float64 { return float64(r.TotalDuration) / 3600.0 }),
		getFloat(results, 1, func(r ComparisonResult) float64 { return float64(r.TotalDuration) / 3600.0 }),
		getFloat(results, 2, func(r ComparisonResult) float64 { return float64(r.TotalDuration) / 3600.0 }),
	)

	// Ongoing Segments
	t.Logf("║ Ongoing Segments      │ %-10s │ %-10s │ %-10s ║",
		getValue(0, func(r ComparisonResult) string { return fmt.Sprintf("%d", r.OngoingCount) }),
		getValue(1, func(r ComparisonResult) string { return fmt.Sprintf("%d", r.OngoingCount) }),
		getValue(2, func(r ComparisonResult) string { return fmt.Sprintf("%d", r.OngoingCount) }),
	)

	t.Log("╚═══════════════════════╧════════════╧════════════╧════════════╝")
	t.Log("")

	// Performance comparison
	baselineTime := results[0].QueryTime
	t.Log("Performance Relative to Ignition Detection:")
	for i, result := range results {
		if i == 0 {
			t.Logf("  • %s: 1.0x (baseline)", result.Mechanism)
		} else {
			ratio := float64(result.QueryTime) / float64(baselineTime)
			t.Logf("  • %s: %.1fx slower", result.Mechanism, ratio)
		}
	}

	// Accuracy comparison (assuming ignition is ground truth)
	t.Log("\nAccuracy Relative to Ignition Detection:")
	baselineCount := results[0].SegmentCount
	for i, result := range results {
		if i == 0 {
			t.Logf("  • %s: 100%% (baseline)", result.Mechanism)
		} else {
			accuracy := float64(result.SegmentCount) / float64(baselineCount) * 100.0
			diff := result.SegmentCount - baselineCount
			if diff > 0 {
				t.Logf("  • %s: %.1f%% (+%d segments)", result.Mechanism, accuracy, diff)
			} else if diff < 0 {
				t.Logf("  • %s: %.1f%% (%d segments)", result.Mechanism, accuracy, diff)
			} else {
				t.Logf("  • %s: %.1f%% (exact match)", result.Mechanism, accuracy)
			}
		}
	}
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%d ms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
	return fmt.Sprintf("%.1f s", d.Seconds())
}

// loadSignalsFromCSV loads signals from the CSV file
func loadSignalsFromCSV(path string) ([]vss.Signal, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Build column index map
	colIndex := make(map[string]int)
	for i, header := range headers {
		colIndex[header] = i
	}

	var signals []vss.Signal

	// Read data rows
	for {
		row, err := reader.Read()
		if err != nil {
			break // EOF or error
		}

		tokenIDStr := row[colIndex["token_id"]]
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			continue
		}

		timestampStr := row[colIndex["timestamp"]]
		timestamp, err := time.Parse("2006-01-02 15:04:05.000000", timestampStr)
		if err != nil {
			// Try without microseconds
			timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
			if err != nil {
				continue
			}
		}

		valueNumberStr := row[colIndex["value_number"]]
		valueNumber := 0.0
		if valueNumberStr != "" {
			valueNumber, _ = strconv.ParseFloat(valueNumberStr, 64)
		}

		// Parse location data from JSON
		var valueLocation vss.Location
		valueLocationStr := row[colIndex["value_location"]]
		if valueLocationStr != "" && valueLocationStr != "{}" {
			var locData struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
				Hdop      float64 `json:"hdop"`
			}
			// CSV reader with LazyQuotes should handle escaped quotes correctly
			if err := json.Unmarshal([]byte(valueLocationStr), &locData); err == nil {
				valueLocation = vss.Location{
					Latitude:  locData.Latitude,
					Longitude: locData.Longitude,
					HDOP:      locData.Hdop,
				}
			}
		}

		signal := vss.Signal{
			TokenID:       uint32(tokenID),
			Timestamp:     timestamp,
			Name:          row[colIndex["name"]],
			Source:        row[colIndex["source"]],
			Producer:      row[colIndex["producer"]],
			CloudEventID:  row[colIndex["cloud_event_id"]],
			ValueNumber:   valueNumber,
			ValueString:   row[colIndex["value_string"]],
			ValueLocation: valueLocation,
		}

		signals = append(signals, signal)
	}

	return signals, nil
}

// insertSignals inserts test signals into ClickHouse signal table
func insertSignals(t *testing.T, ch *container.Container, signals []vss.Signal) {
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

// getFloat helper to safely get float values from results
func getFloat(results []ComparisonResult, idx int, fn func(ComparisonResult) float64) float64 {
	if idx < len(results) {
		return fn(results[idx])
	}
	return 0.0
}

// printDayTripDetails prints detailed trip information for each mechanism on a given day
func printDayTripDetails(t *testing.T, dayStr string, mechanisms []struct {
	name      string
	mechanism model.DetectionMechanism
}, mechanismSegments map[string][]*Segment, dayStart, dayEnd time.Time) {
	t.Log("")
	t.Logf("  === Trip Details for %s ===", dayStr)
	t.Log("")

	for _, mech := range mechanisms {
		// Filter segments for this day
		var daySegments []*Segment
		for _, seg := range mechanismSegments[mech.name] {
			// Segment belongs to this day if it starts on this day
			if seg.StartTime.After(dayStart.Add(-time.Second)) && seg.StartTime.Before(dayEnd) {
				daySegments = append(daySegments, seg)
			}
		}

		if len(daySegments) == 0 {
			continue
		}

		mechanismShort := mech.name
		if len(mechanismShort) > 30 {
			mechanismShort = mechanismShort[:27] + "..."
		}

		t.Logf("  %s (%d trips):", mechanismShort, len(daySegments))
		t.Log("  ┌────┬─────────────────────┬─────────────────────┬──────────────┬──────────┐")
		t.Log("  │ #  │ Start Time          │ End Time            │ Duration     │ Status   │")
		t.Log("  ├────┼─────────────────────┼─────────────────────┼──────────────┼──────────┤")

		for i, seg := range daySegments {
			startStr := seg.StartTime.Format("15:04:05")
			endStr := "ongoing"
			status := "🔵 Active"

			if seg.EndTime != nil {
				endStr = seg.EndTime.Format("15:04:05")
				status = "✅ Done"
			}

			durationMin := float64(seg.DurationSeconds) / 60.0

			t.Logf("  │ %2d │ %19s │ %19s │ %10.1fm │ %8s │",
				i+1,
				startStr,
				endStr,
				durationMin,
				status,
			)
		}

		t.Log("  └────┴─────────────────────┴─────────────────────┴──────────────┴──────────┘")
		t.Log("")
	}
}
