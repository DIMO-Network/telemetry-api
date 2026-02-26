package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	sigmigrations "github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSignal represents minimal signal data without location fields
type testSignal struct {
	TokenID      uint32
	Timestamp    time.Time
	Name         string
	ValueNumber  float64
	ValueString  string
	Source       string
	Producer     string
	CloudEventID string
}

// testStateChange represents ignition state change data
type testStateChange struct {
	TokenID              uint32
	SignalName           string
	Timestamp            time.Time
	NewState             float64
	PrevState            float64
	TimeSincePrevSeconds uint32
	Source               string
	Producer             string
	CloudEventID         string
	Version              uint64
}

const (
	testTokenID  = uint32(12345)
	testSource   = "test-source"
	testProducer = "test-producer"
	// tripDuration must be >= 240s (defaultMinSegmentDurationSeconds in detectors)
	tripDuration = 300 // 5 minutes
	// tripGap must be > 300s (defaultMaxGapSeconds) to ensure separate trips
	tripGap = 720 // 12 minutes
)

// baseTime is a fixed timestamp for repeatable tests
var baseTime = time.Date(2025, 11, 25, 12, 0, 0, 0, time.UTC)

// generateTripSignals creates signal data for a trip of the specified duration.
// Uses non-location signals: speed, obdStatusDTCCount, lowVoltageBatteryCurrentVoltage
func generateTripSignals(startTime time.Time, durationSeconds int, eventIDPrefix string) []testSignal {
	signals := []testSignal{}
	signalNames := []string{"speed", "obdStatusDTCCount", "lowVoltageBatteryCurrentVoltage"}

	// Generate signals every 5 seconds for each signal type
	// This gives us 3 signals * 12 per minute = 36 signals per minute
	for offset := 0; offset < durationSeconds; offset += 5 {
		ts := startTime.Add(time.Duration(offset) * time.Second)
		for i, name := range signalNames {
			signals = append(signals, testSignal{
				TokenID:      testTokenID,
				Timestamp:    ts,
				Name:         name,
				ValueNumber:  float64(offset + i), // Varying values
				ValueString:  "",
				Source:       testSource,
				Producer:     testProducer,
				CloudEventID: fmt.Sprintf("%s%s%s", eventIDPrefix, ts.Format("150405"), name[:3]),
			})
		}
	}
	return signals
}

// generateTestData creates test data for exactly 2 trips.
// Trip 1: tripDuration seconds of signals starting at baseTime
// Gap: tripGap seconds (> 10 minutes to ensure separate trips)
// Trip 2: tripDuration seconds of signals starting after the gap
func generateTestData() ([]testSignal, []testStateChange) {
	var signals []testSignal
	var stateChanges []testStateChange

	// Trip 1: tripDuration seconds of data
	trip1Start := baseTime
	trip1End := trip1Start.Add(time.Duration(tripDuration) * time.Second)
	signals = append(signals, generateTripSignals(trip1Start, tripDuration, "trip1-")...)

	// Trip 2: tripDuration seconds starting tripGap seconds after trip 1 end
	// This creates a gap > maxGapSeconds (300s default) between trips
	trip2Start := trip1End.Add(time.Duration(tripGap) * time.Second)
	trip2End := trip2Start.Add(time.Duration(tripDuration) * time.Second)
	signals = append(signals, generateTripSignals(trip2Start, tripDuration, "trip2-")...)

	// State changes for ignition detection
	stateChanges = []testStateChange{
		// Trip 1: Ignition ON
		{
			TokenID:              testTokenID,
			SignalName:           "isIgnitionOn",
			Timestamp:            trip1Start,
			NewState:             1,
			PrevState:            0,
			TimeSincePrevSeconds: 3600, // 1 hour since last
			Source:               testSource,
			Producer:             testProducer,
			CloudEventID:         "trip1-ign-on",
			Version:              1000001,
		},
		// Trip 1: Ignition OFF
		{
			TokenID:              testTokenID,
			SignalName:           "isIgnitionOn",
			Timestamp:            trip1End,
			NewState:             0,
			PrevState:            1,
			TimeSincePrevSeconds: uint32(tripDuration),
			Source:               testSource,
			Producer:             testProducer,
			CloudEventID:         "trip1-ign-off",
			Version:              1000002,
		},
		// Trip 2: Ignition ON
		{
			TokenID:              testTokenID,
			SignalName:           "isIgnitionOn",
			Timestamp:            trip2Start,
			NewState:             1,
			PrevState:            0,
			TimeSincePrevSeconds: uint32(tripGap),
			Source:               testSource,
			Producer:             testProducer,
			CloudEventID:         "trip2-ign-on",
			Version:              1000003,
		},
		// Trip 2: Ignition OFF
		{
			TokenID:              testTokenID,
			SignalName:           "isIgnitionOn",
			Timestamp:            trip2End,
			NewState:             0,
			PrevState:            1,
			TimeSincePrevSeconds: uint32(tripDuration),
			Source:               testSource,
			Producer:             testProducer,
			CloudEventID:         "trip2-ign-off",
			Version:              1000004,
		},
	}

	return signals, stateChanges
}

func insertTestSignals(t *testing.T, conn clickhouse.Conn, signals []testSignal) {
	t.Helper()
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO signal")
	require.NoError(t, err)

	for _, sig := range signals {
		err := batch.Append(
			sig.TokenID,
			sig.Timestamp,
			sig.Name,
			sig.Source,
			sig.Producer,
			sig.CloudEventID,
			sig.ValueNumber,
			sig.ValueString,
			[]interface{}{float64(0), float64(0), float64(0)}, // Empty location tuple
		)
		require.NoError(t, err, "Failed to append signal")
	}
	err = batch.Send()
	require.NoError(t, err, "Failed to send signal batch")
	t.Logf("Inserted %d test signals", len(signals))
}

func insertTestStateChanges(t *testing.T, conn clickhouse.Conn, stateChanges []testStateChange) {
	t.Helper()
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO signal_state_changes")
	require.NoError(t, err)

	for _, sc := range stateChanges {
		err := batch.Append(
			sc.TokenID,
			sc.SignalName,
			sc.Timestamp,
			sc.NewState,
			sc.PrevState,
			sc.TimeSincePrevSeconds,
			sc.Source,
			sc.Producer,
			sc.CloudEventID,
			sc.Version,
		)
		require.NoError(t, err, "Failed to append state change")
	}
	err = batch.Send()
	require.NoError(t, err, "Failed to send state change batch")
	t.Logf("Inserted %d test state changes", len(stateChanges))
}

// TestSegmentDetectors validates that all segment detection mechanisms
// correctly identify trips from the test data.
func TestSegmentDetectors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup ClickHouse container
	chContainer, err := container.CreateClickHouseContainer(ctx, chconfig.Settings{})
	require.NoError(t, err, "Failed to create ClickHouse container")
	t.Cleanup(func() { chContainer.Terminate(ctx) })

	db, err := chContainer.GetClickhouseAsDB()
	require.NoError(t, err)

	// Run migrations
	err = sigmigrations.RunGoose(ctx, []string{"up", "-v"}, db)
	require.NoError(t, err, "Failed to run migrations")

	conn, err := chContainer.GetClickHouseAsConn()
	require.NoError(t, err)

	// Generate and insert test data
	signals, stateChanges := generateTestData()
	insertTestSignals(t, conn, signals)
	insertTestStateChanges(t, conn, stateChanges)

	t.Logf("Test data: %d signals, %d state changes", len(signals), len(stateChanges))

	// Time range that covers both trips (with margin)
	from := baseTime.Add(-1 * time.Hour)
	to := baseTime.Add(2 * time.Hour)

	// Expected trip times
	trip1Start := baseTime
	trip2Start := trip1Start.Add(time.Duration(tripDuration)*time.Second + time.Duration(tripGap)*time.Second)

	t.Run("IgnitionDetector", func(t *testing.T) {
		detector := ch.NewIgnitionDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, from, to, nil)
		require.NoError(t, err)

		assert.Len(t, segments, 2, "Expected 2 trips from IgnitionDetector")
		if len(segments) >= 2 {
			assert.Equal(t, trip1Start, segments[0].Start.Timestamp)
			assert.NotNil(t, segments[0].End)
			assert.False(t, segments[0].IsOngoing)

			assert.Equal(t, trip2Start, segments[1].Start.Timestamp)
			assert.NotNil(t, segments[1].End)
			assert.False(t, segments[1].IsOngoing)
		}
		for i, seg := range segments {
			t.Logf("  Segment %d: %s (duration: %ds, ongoing: %v)",
				i+1, seg.Start.Timestamp.Format(time.RFC3339), seg.Duration, seg.IsOngoing)
		}
	})

	t.Run("FrequencyDetector", func(t *testing.T) {
		detector := ch.NewFrequencyDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, from, to, nil)
		require.NoError(t, err)

		assert.Len(t, segments, 2, "Expected 2 trips from FrequencyDetector")
		for i, seg := range segments {
			t.Logf("  Segment %d: %s (duration: %ds, ongoing: %v)",
				i+1, seg.Start.Timestamp.Format(time.RFC3339), seg.Duration, seg.IsOngoing)
		}
	})

	t.Run("ChangePointDetector", func(t *testing.T) {
		detector := ch.NewChangePointDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, from, to, nil)
		require.NoError(t, err)

		assert.Len(t, segments, 2, "Expected 2 trips from ChangePointDetector")
		for i, seg := range segments {
			t.Logf("  Segment %d: %s (duration: %ds, ongoing: %v)",
				i+1, seg.Start.Timestamp.Format(time.RFC3339), seg.Duration, seg.IsOngoing)
		}
	})

	// Test StartedBeforeRange flag
	fromMidTrip1 := baseTime.Add(90 * time.Second)

	t.Run("IgnitionDetector_StartedBeforeRange", func(t *testing.T) {
		detector := ch.NewIgnitionDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, fromMidTrip1, to, nil)
		require.NoError(t, err)
		require.Len(t, segments, 2, "Expected 2 trips")
		assert.True(t, segments[0].StartedBeforeRange, "Trip 1 should have StartedBeforeRange=true")
		assert.False(t, segments[1].StartedBeforeRange, "Trip 2 should have StartedBeforeRange=false")
	})

	t.Run("FrequencyDetector_StartedBeforeRange", func(t *testing.T) {
		detector := ch.NewFrequencyDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, fromMidTrip1, to, nil)
		require.NoError(t, err)
		require.Len(t, segments, 2, "Expected 2 trips")
		assert.True(t, segments[0].StartedBeforeRange, "Trip 1 should have StartedBeforeRange=true")
		assert.False(t, segments[1].StartedBeforeRange, "Trip 2 should have StartedBeforeRange=false")
	})

	t.Run("ChangePointDetector_StartedBeforeRange", func(t *testing.T) {
		detector := ch.NewChangePointDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, fromMidTrip1, to, nil)
		require.NoError(t, err)
		require.Len(t, segments, 2, "Expected 2 trips")
		assert.True(t, segments[0].StartedBeforeRange, "Trip 1 should have StartedBeforeRange=true")
		assert.False(t, segments[1].StartedBeforeRange, "Trip 2 should have StartedBeforeRange=false")
	})

	// Idling: engine speed (RPM) in idle range for a contiguous period
	idleStart := baseTime.Add(48 * time.Hour)
	idleDurationSec := 15 * 60 // 15 minutes
	t.Run("IdlingDetector", func(t *testing.T) {
		idleSignals := generateIdleRpmSignals(idleStart, idleDurationSec)
		insertTestSignals(t, conn, idleSignals)

		fromIdle := idleStart.Add(-1 * time.Hour)
		toIdle := idleStart.Add(time.Duration(idleDurationSec)*time.Second + 1*time.Hour)

		detector := ch.NewIdlingDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, fromIdle, toIdle, nil)
		require.NoError(t, err)

		require.Len(t, segments, 1, "Expected 1 idling segment")
		seg := segments[0]
		assert.False(t, seg.IsOngoing)
		assert.NotNil(t, seg.End)
		assert.GreaterOrEqual(t, seg.Duration, 240)
		t.Logf("Idling segment: %s (duration: %ds)", seg.Start.Timestamp.Format(time.RFC3339), seg.Duration)
	})

	// Refuel: fuel level rises sharply
	refuelStart := baseTime.Add(72 * time.Hour)
	t.Run("RefuelDetector", func(t *testing.T) {
		refuelSignals := generateRefuelSignals(refuelStart)
		insertTestSignals(t, conn, refuelSignals)

		fromRefuel := refuelStart.Add(-1 * time.Hour)
		toRefuel := refuelStart.Add(2 * time.Hour)

		detector := ch.NewRefuelDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, fromRefuel, toRefuel, nil)
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(segments), 1, "Expected at least 1 refuel segment")
		for i, seg := range segments {
			assert.False(t, seg.IsOngoing)
			assert.NotNil(t, seg.End)
			t.Logf("  Refuel segment %d: %s (duration: %ds)", i+1, seg.Start.Timestamp.Format(time.RFC3339), seg.Duration)
		}
	})

	// Recharge: SoC rises over time
	rechargeStart := baseTime.Add(96 * time.Hour)
	t.Run("RechargeDetector", func(t *testing.T) {
		rechargeSignals := generateRechargeSignals(rechargeStart)
		insertTestSignals(t, conn, rechargeSignals)

		fromRecharge := rechargeStart.Add(-1 * time.Hour)
		toRecharge := rechargeStart.Add(4 * time.Hour)

		detector := ch.NewRechargeDetector(conn)
		segments, err := detector.DetectSegments(ctx, testTokenID, fromRecharge, toRecharge, nil)
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(segments), 1, "Expected at least 1 recharge segment")
		for i, seg := range segments {
			assert.False(t, seg.IsOngoing)
			assert.NotNil(t, seg.End)
			t.Logf("  Recharge segment %d: %s (duration: %ds)", i+1, seg.Start.Timestamp.Format(time.RFC3339), seg.Duration)
		}
	})
}

// generateIdleRpmSignals creates powertrainCombustionEngineSpeed signals in idle range (e.g. 800 rpm)
// at 10s intervals for the given duration so 60s windows have enough samples and max(rpm) <= 1000.
func generateIdleRpmSignals(startTime time.Time, durationSeconds int) []testSignal {
	const engineSpeedName = "powertrainCombustionEngineSpeed"
	const idleRpm = 800.0
	signals := []testSignal{}
	for offset := 0; offset < durationSeconds; offset += 10 {
		ts := startTime.Add(time.Duration(offset) * time.Second)
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         engineSpeedName,
			ValueNumber:  idleRpm,
			ValueString:  "",
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("idle-%s-%d", ts.Format("150405"), offset),
		})
	}
	return signals
}

// generateRefuelSignals creates fuel level signals that simulate a refuel:
// - 30 min at ~20% fuel (low, pre-refuel)
// - sharp jump to ~80% (refuel event)
// - 30 min at ~80% (post-refuel)
func generateRefuelSignals(startTime time.Time) []testSignal {
	const fuelName = "powertrainFuelSystemRelativeLevel"
	signals := []testSignal{}

	// Pre-refuel: 30 min at low fuel, every 60s
	for offset := 0; offset < 30*60; offset += 60 {
		ts := startTime.Add(time.Duration(offset) * time.Second)
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         fuelName,
			ValueNumber:  20.0,
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("refuel-pre-%d", offset),
		})
	}

	// Refuel event: jump to 80% (within a 5-minute window, rise is 300% relative)
	refuelTime := startTime.Add(30 * time.Minute)
	signals = append(signals, testSignal{
		TokenID:      testTokenID,
		Timestamp:    refuelTime,
		Name:         fuelName,
		ValueNumber:  80.0,
		Source:       testSource,
		Producer:     testProducer,
		CloudEventID: "refuel-jump",
	})

	// Post-refuel: 30 min at high fuel, every 60s
	for offset := 1; offset <= 30; offset++ {
		ts := refuelTime.Add(time.Duration(offset) * time.Minute)
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         fuelName,
			ValueNumber:  80.0,
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("refuel-post-%d", offset),
		})
	}

	return signals
}

// generateRechargeSignals creates SoC signals that simulate a recharge session:
// - 30 min declining to 20% (driving, pre-charge)
// - 2 hours rising from 20% to 90% (charging)
// - Odometer stays constant during charge period
func generateRechargeSignals(startTime time.Time) []testSignal {
	const socName = "powertrainTractionBatteryStateOfChargeCurrent"
	const odoName = "powertrainTransmissionTravelledDistance"
	signals := []testSignal{}

	// Pre-charge: declining SoC, every 60s for 30 min
	for offset := 0; offset < 30; offset++ {
		ts := startTime.Add(time.Duration(offset) * time.Minute)
		soc := 35.0 - float64(offset)*0.5 // 35% down to 20%
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         socName,
			ValueNumber:  soc,
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("recharge-pre-%d", offset),
		})
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         odoName,
			ValueNumber:  50000.0 + float64(offset)*0.1, // moving
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("recharge-odo-pre-%d", offset),
		})
	}

	// Charging: rising SoC, every 60s for 2 hours. Odometer constant.
	chargeStart := startTime.Add(30 * time.Minute)
	for offset := 0; offset < 120; offset++ {
		ts := chargeStart.Add(time.Duration(offset) * time.Minute)
		soc := 20.0 + float64(offset)*0.58 // ~20% to ~90% over 2 hours
		if soc > 90.0 {
			soc = 90.0
		}
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         socName,
			ValueNumber:  soc,
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("recharge-charge-%d", offset),
		})
		signals = append(signals, testSignal{
			TokenID:      testTokenID,
			Timestamp:    ts,
			Name:         odoName,
			ValueNumber:  50003.0, // stationary during charge
			Source:       testSource,
			Producer:     testProducer,
			CloudEventID: fmt.Sprintf("recharge-odo-charge-%d", offset),
		})
	}

	return signals
}
