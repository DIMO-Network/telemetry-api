package e2e_test

import (
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApproximateLocation(t *testing.T) {
	services := GetTestServices(t)
	locationTime := time.Date(2024, 11, 20, 22, 28, 17, 0, time.UTC)
	// Set up test data in Clickhouse
	signals := []vss.Signal{
		{
			Source:      ch.SourceTranslations["smartcar"],
			Timestamp:   locationTime.Add(-time.Hour * 24),
			Name:        vss.FieldCurrentLocationLatitude,
			ValueNumber: 40.73899538333504,
			TokenID:     39718,
		},
		{
			Source:      ch.SourceTranslations["smartcar"],
			Timestamp:   locationTime.Add(-time.Hour * 24),
			Name:        vss.FieldCurrentLocationLongitude,
			ValueNumber: 73.99386110247163,
			TokenID:     39718,
		},
		{
			Source:      ch.SourceTranslations["smartcar"],
			Timestamp:   locationTime,
			Name:        vss.FieldCurrentLocationLatitude,
			ValueNumber: 40.73899538333504,
			TokenID:     39718,
		},
		{
			Source:      ch.SourceTranslations["smartcar"],
			Timestamp:   locationTime,
			Name:        vss.FieldCurrentLocationLongitude,
			ValueNumber: 73.99386110247163,
			TokenID:     39718,
		},
	}

	insertSignal(t, services.CH, signals)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Execute the query
	query := `
	query {
		signalsLatest(tokenId: 39718) {
			lastSeen
			currentLocationApproximateLatitude{
				timestamp
				value
			}
			currentLocationApproximateLongitude{
				timestamp
				value
			}
		}
	}`

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, "39718", []int{8})

	// Execute request
	result := ApproxResult{}
	err := telemetryClient.Post(query, &result, WithToken(token))
	require.NoError(t, err)

	expectedLatLong := repositories.GetApproximateLoc(40.73899538333504, 73.99386110247163)
	// Assert the results
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.LastSeen)
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.ApproxLat.Timestamp)
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.ApproxLong.Timestamp)
	assert.Equal(t, expectedLatLong.Lat, result.SignalLatest.ApproxLat.Value)
	assert.Equal(t, expectedLatLong.Lng, result.SignalLatest.ApproxLong.Value)

	// verify we do not leak the exact location
	query = `query {
		signalsLatest(tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
			currentLocationApproximateLatitude{
				timestamp
				value
			}
			currentLocationApproximateLongitude{
				timestamp
				value
			}
			currentLocationLatitude{
				timestamp
				value
			}
			currentLocationLongitude{
				timestamp
				value
			}
		}
	}`

	// Execute request
	result = ApproxResult{}
	err = telemetryClient.Post(query, &result, WithToken(token))
	require.Error(t, err)

	// Assert the results
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.LastSeen)
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.ApproxLat.Timestamp)
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.ApproxLong.Timestamp)
	assert.Equal(t, expectedLatLong.Lat, result.SignalLatest.ApproxLat.Value)
	assert.Equal(t, expectedLatLong.Lng, result.SignalLatest.ApproxLong.Value)
	assert.Nil(t, result.SignalLatest.Lat)
	assert.Nil(t, result.SignalLatest.Long)

	query = `query {
		signals(tokenId:39718, from: "2020-04-15T09:21:19Z", to: "2025-04-27T09:21:19Z", interval:"24h"){
			timestamp
			currentLocationApproximateLatitude(agg: FIRST)
			currentLocationApproximateLongitude(agg: FIRST)
		}
	}`
	// Execute request
	aggResult := ApproxAgg{}
	err = telemetryClient.Post(query, &aggResult, WithToken(token))
	require.NoError(t, err)

	require.Len(t, aggResult.Signals, 2)
	// Assert the results
	assert.Equal(t, locationTime.Add(-time.Hour*24).Truncate(time.Hour*24).Format(time.RFC3339), aggResult.Signals[0].Timestamp)
	assert.Equal(t, expectedLatLong.Lat, *aggResult.Signals[0].ApproxLat)
	assert.Equal(t, expectedLatLong.Lng, *aggResult.Signals[0].ApproxLong)

	assert.Equal(t, locationTime.Truncate(time.Hour*24).Format(time.RFC3339), aggResult.Signals[1].Timestamp)
	assert.Equal(t, expectedLatLong.Lat, *aggResult.Signals[1].ApproxLat)
	assert.Equal(t, expectedLatLong.Lng, *aggResult.Signals[1].ApproxLong)

	query = `query {
		signals(tokenId:39718, from: "2020-04-15T09:21:19Z", to: "2025-04-27T09:21:19Z", interval:"24h"){
			timestamp
			currentLocationApproximateLatitude(agg: FIRST)
			currentLocationApproximateLongitude(agg: FIRST)
			currentLocationLatitude(agg: FIRST)
			currentLocationLongitude(agg: FIRST)
		}
	}`
	// Execute request
	aggResult = ApproxAgg{}
	err = telemetryClient.Post(query, &aggResult, WithToken(token))
	require.Error(t, err)

	// Assert the results
	require.Len(t, aggResult.Signals, 2)
	assert.Equal(t, locationTime.Add(-time.Hour*24).Truncate(time.Hour*24).Format(time.RFC3339), aggResult.Signals[0].Timestamp)
	assert.Equal(t, expectedLatLong.Lat, *aggResult.Signals[0].ApproxLat)
	assert.Equal(t, expectedLatLong.Lng, *aggResult.Signals[0].ApproxLong)
	assert.Nil(t, aggResult.Signals[0].Lat)
	assert.Nil(t, aggResult.Signals[0].Long)

	assert.Equal(t, locationTime.Truncate(time.Hour*24).Format(time.RFC3339), aggResult.Signals[1].Timestamp)
	assert.Equal(t, expectedLatLong.Lat, *aggResult.Signals[1].ApproxLat)
	assert.Equal(t, expectedLatLong.Lng, *aggResult.Signals[1].ApproxLong)
	assert.Nil(t, aggResult.Signals[1].Lat)
	assert.Nil(t, aggResult.Signals[1].Long)

}

type ApproxResult struct {
	SignalLatest struct {
		LastSeen   string          `json:"lastSeen"`
		ApproxLat  *SignalWithTime `json:"currentLocationApproximateLatitude"`
		ApproxLong *SignalWithTime `json:"currentLocationApproximateLongitude"`
		Lat        *SignalWithTime `json:"currentLocationLatitude"`
		Long       *SignalWithTime `json:"currentLocationLongitude"`
	} `json:"signalsLatest"`
}

type ApproxAgg struct {
	Signals []struct {
		Timestamp  string   `json:"timestamp"`
		ApproxLat  *float64 `json:"currentLocationApproximateLatitude"`
		ApproxLong *float64 `json:"currentLocationApproximateLongitude"`
		Lat        *float64 `json:"currentLocationLatitude"`
		Long       *float64 `json:"currentLocationLongitude"`
	} `json:"signals"`
}
