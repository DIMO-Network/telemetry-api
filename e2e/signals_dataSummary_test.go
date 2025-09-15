package e2e_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalsMetadata(t *testing.T) {
	services := GetTestServices(t)
	tokenID := uint32(rand.Intn(1000000))
	// Define test timestamps
	smartCarTime1 := time.Date(2024, 11, 20, 22, 28, 17, 0, time.UTC)
	smartCarTime2 := time.Date(2024, 11, 21, 10, 15, 30, 0, time.UTC)
	autopiTime1 := time.Date(2024, 11, 1, 20, 1, 29, 0, time.UTC)
	autopiTime2 := time.Date(2024, 11, 15, 14, 45, 0, 0, time.UTC)
	macaronTime := time.Date(2024, 3, 20, 20, 13, 57, 0, time.UTC)
	teslaTime := time.Date(2024, 10, 5, 8, 30, 45, 0, time.UTC)

	// Set up test data in Clickhouse with multiple signals and sources
	signals := []vss.Signal{
		// SmartCar signals - speed and location
		{
			Source:      ch.SourceTranslations["smartcar"][0],
			Timestamp:   smartCarTime1,
			Name:        vss.FieldSpeed,
			ValueNumber: 65.5,
			TokenID:     tokenID,
		},
		{
			Source:      ch.SourceTranslations["smartcar"][0],
			Timestamp:   smartCarTime2,
			Name:        vss.FieldSpeed,
			ValueNumber: 70.2,
			TokenID:     tokenID,
		},
		{
			Source:      ch.SourceTranslations["smartcar"][0],
			Timestamp:   smartCarTime1,
			Name:        vss.FieldCurrentLocationLatitude,
			ValueNumber: 40.73899538333504,
			TokenID:     tokenID,
		},
		{
			Source:      ch.SourceTranslations["smartcar"][0],
			Timestamp:   smartCarTime1,
			Name:        vss.FieldCurrentLocationLongitude,
			ValueNumber: 73.99386110247163,
			TokenID:     tokenID,
		},

		// AutoPi signals - speed and battery
		{
			Source:      ch.SourceTranslations["autopi"][0],
			Timestamp:   autopiTime1,
			Name:        vss.FieldSpeed,
			ValueNumber: 14,
			TokenID:     tokenID,
		},
		{
			Source:      ch.SourceTranslations["autopi"][0],
			Timestamp:   autopiTime2,
			Name:        vss.FieldSpeed,
			ValueNumber: 25.8,
			TokenID:     tokenID,
		},
		{
			Source:      ch.SourceTranslations["autopi"][0],
			Timestamp:   autopiTime1,
			Name:        vss.FieldPowertrainTractionBatteryStateOfChargeCurrent,
			ValueNumber: 75.5,
			TokenID:     tokenID,
		},

		// Macaron signals - just speed
		{
			Source:      ch.SourceTranslations["macaron"][0],
			Timestamp:   macaronTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 3,
			TokenID:     tokenID,
		},

		// Tesla signals - speed and different battery field
		{
			Source:      ch.SourceTranslations["tesla"][0],
			Timestamp:   teslaTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 88.5,
			TokenID:     tokenID,
		},
		{
			Source:      ch.SourceTranslations["tesla"][0],
			Timestamp:   teslaTime,
			Name:        vss.FieldPowertrainTractionBatteryChargingChargeCurrentAC,
			ValueNumber: 82.3,
			TokenID:     tokenID,
		},
	}

	insertSignal(t, services.CH, signals)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, fmt.Sprintf("%d", tokenID), []int{1, 4})

	t.Run("All signals metadata", func(t *testing.T) {
		// Execute the query for all signals
		query := `
		query SignalsMetadataTest {
			dataSummary(tokenId: %d) {
				numberOfSignals
				availableSignals
				firstSeen
				lastSeen
				signalDataSummary {
					name
					numberOfSignals
					firstSeen
					lastSeen
				}
			}
		}`

		// Execute request
		result := SignalsMetadataResult{}
		err := telemetryClient.Post(fmt.Sprintf(query, tokenID), &result, WithToken(token))
		require.NoError(t, err)

		// Assert the overall metadata results
		assert.Equal(t, uint64(10), result.DataSummary.NumberOfSignals)
		assert.Equal(t, macaronTime.Format(time.RFC3339), result.DataSummary.FirstSeen)
		assert.Equal(t, smartCarTime2.Format(time.RFC3339), result.DataSummary.LastSeen)

		// Assert available signals (should be sorted)
		expectedAvailableSignals := []string{
			"currentLocationLatitude",
			"currentLocationLongitude",
			"powertrainTractionBatteryChargingChargeCurrentAC",
			"powertrainTractionBatteryStateOfChargeCurrent",
			"speed",
		}
		assert.Equal(t, expectedAvailableSignals, result.DataSummary.AvailableSignals)

		// Assert signal metadata - should have 5 different signal types
		require.Len(t, result.DataSummary.SignalDataSummary, 5)

		// Find and validate speed signal metadata (most common signal)
		var speedMetadata *DataSummaryTest
		for _, sm := range result.DataSummary.SignalDataSummary {
			if sm.Name == "speed" {
				speedMetadata = sm
				break
			}
		}
		require.NotNil(t, speedMetadata, "Speed signal metadata should be present")
		assert.Equal(t, uint64(6), speedMetadata.NumberOfSignals) // 6 speed signals total
		assert.Equal(t, macaronTime.Format(time.RFC3339), speedMetadata.FirstSeen)
		assert.Equal(t, smartCarTime2.Format(time.RFC3339), speedMetadata.LastSeen)

		// Find and validate battery signal metadata
		var batteryCurrentMetadata *DataSummaryTest
		for _, sm := range result.DataSummary.SignalDataSummary {
			if sm.Name == "powertrainTractionBatteryStateOfChargeCurrent" {
				batteryCurrentMetadata = sm
				break
			}
		}
		require.NotNil(t, batteryCurrentMetadata, "Battery current signal metadata should be present")
		assert.Equal(t, uint64(1), batteryCurrentMetadata.NumberOfSignals)
		assert.Equal(t, autopiTime1.Format(time.RFC3339), batteryCurrentMetadata.FirstSeen)
		assert.Equal(t, autopiTime1.Format(time.RFC3339), batteryCurrentMetadata.LastSeen)
	})

	t.Run("Filtered by source", func(t *testing.T) {
		// Execute the query filtered by smartcar source
		query := `
		query SignalsMetadataFiltered {
			dataSummary(tokenId: %d, filter: {source: "smartcar"}) {
				numberOfSignals
				availableSignals
				firstSeen
				lastSeen
				signalDataSummary {
					name
					numberOfSignals
					firstSeen
					lastSeen
				}
			}
		}`

		// Execute request
		result := SignalsMetadataResult{}
		err := telemetryClient.Post(fmt.Sprintf(query, tokenID), &result, WithToken(token))
		require.NoError(t, err)

		// Assert filtered results - should only include smartcar signals
		assert.Equal(t, uint64(4), result.DataSummary.NumberOfSignals) // 2 speed + 1 lat + 1 lon
		assert.Equal(t, smartCarTime1.Format(time.RFC3339), result.DataSummary.FirstSeen)
		assert.Equal(t, smartCarTime2.Format(time.RFC3339), result.DataSummary.LastSeen)

		// Assert available signals for smartcar only
		expectedSmartcarSignals := []string{
			"currentLocationLatitude",
			"currentLocationLongitude",
			"speed",
		}
		assert.Equal(t, expectedSmartcarSignals, result.DataSummary.AvailableSignals)

		// Assert signal metadata - should have 3 different signal types for smartcar
		require.Len(t, result.DataSummary.SignalDataSummary, 3)

		// Validate speed signal metadata for smartcar only
		var speedMetadata *DataSummaryTest
		for _, sm := range result.DataSummary.SignalDataSummary {
			if sm.Name == "speed" {
				speedMetadata = sm
				break
			}
		}
		require.NotNil(t, speedMetadata, "Speed signal metadata should be present")
		assert.Equal(t, uint64(2), speedMetadata.NumberOfSignals) // 2 smartcar speed signals
		assert.Equal(t, smartCarTime1.Format(time.RFC3339), speedMetadata.FirstSeen)
		assert.Equal(t, smartCarTime2.Format(time.RFC3339), speedMetadata.LastSeen)
	})

	t.Run("No signals for non-existent source", func(t *testing.T) {
		// Execute the query filtered by non-existent source
		query := `
		query SignalsMetadataEmpty {
			dataSummary(tokenId: %d, filter: {source: "nonexistent"}) {
				numberOfSignals
				availableSignals
				firstSeen
				lastSeen
				signalDataSummary {
					name
					numberOfSignals
					firstSeen
					lastSeen
				}
			}
		}`

		// Execute request
		result := SignalsMetadataResult{}
		err := telemetryClient.Post(fmt.Sprintf(query, tokenID), &result, WithToken(token))
		require.NoError(t, err)

		// Assert empty results
		assert.Equal(t, uint64(0), result.DataSummary.NumberOfSignals)
		assert.Empty(t, result.DataSummary.AvailableSignals)
		assert.Empty(t, result.DataSummary.SignalDataSummary)
		// firstSeen and lastSeen should be set to default values when no signals exist
	})
}

type DataSummaryTest struct {
	Name            string `json:"name"`
	NumberOfSignals uint64 `json:"numberOfSignals"`
	FirstSeen       string `json:"firstSeen"`
	LastSeen        string `json:"lastSeen"`
}

type SignalsMetadataResult struct {
	DataSummary struct {
		NumberOfSignals   uint64             `json:"numberOfSignals"`
		AvailableSignals  []string           `json:"availableSignals"`
		FirstSeen         string             `json:"firstSeen"`
		LastSeen          string             `json:"lastSeen"`
		SignalDataSummary []*DataSummaryTest `json:"signalDataSummary"`
	} `json:"dataSummary"`
}
