package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/h3-go/v4"
)

func TestApproximateLocation(t *testing.T) {
	services := GetTestServices(t)
	locationTime := time.Date(2024, 11, 20, 22, 28, 17, 0, time.UTC)
	// Set up test data in Clickhouse
	startLoc := h3.LatLng{Lat: 40.75005062700222, Lng: -74.0094415688571}
	endLoc := h3.LatLng{Lat: 40.73899538333504, Lng: -73.99386110247163}
	const startHDOP = 1.0
	const endHDOP = 2.0
	signals := []vss.Signal{
		{
			Source:    ch.SourceTranslations["smartcar"][0],
			Timestamp: locationTime.Add(-time.Hour * 24),
			Name:      vss.FieldCurrentLocationCoordinates,
			ValueLocation: vss.Location{
				Latitude:  startLoc.Lat,
				Longitude: startLoc.Lng,
				HDOP:      startHDOP,
			},
			Subject: "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
		{
			Source:    ch.SourceTranslations["smartcar"][0],
			Timestamp: locationTime,
			Name:      vss.FieldCurrentLocationCoordinates,
			ValueLocation: vss.Location{
				Latitude:  endLoc.Lat,
				Longitude: endLoc.Lng,
				HDOP:      endHDOP,
			},
			Subject: "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
	}

	insertSignal(t, services.CH, signals)

	expectedStartLatLong := repositories.GetApproximateLoc(startLoc.Lat, startLoc.Lng)
	expectedEndLatLong := repositories.GetApproximateLoc(endLoc.Lat, endLoc.Lng)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, 39718, []string{tokenclaims.PermissionGetApproximateLocation})

	// Execute the latest query
	query := `
	query {
		signalsLatest(tokenId: 39718) {
			lastSeen
			currentLocationApproximateCoordinates {
				timestamp
				value {
					latitude
					longitude
					hdop
				}
			}
		}
	}`

	result := ApproxResult{}
	err := telemetryClient.Post(query, &result, WithToken(token))
	require.NoError(t, err)

	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.LastSeen)
	require.NotNil(t, result.SignalLatest.ApproxCoords)
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.ApproxCoords.Timestamp)
	require.NotNil(t, result.SignalLatest.ApproxCoords.Value)
	assert.Equal(t, expectedEndLatLong.Lat, result.SignalLatest.ApproxCoords.Value.Latitude)
	assert.Equal(t, expectedEndLatLong.Lng, result.SignalLatest.ApproxCoords.Value.Longitude)
	assert.Equal(t, endHDOP, result.SignalLatest.ApproxCoords.Value.Hdop)

	// verify we do not leak the exact location
	query = `query {
		signalsLatest(tokenId: 39718) {
			lastSeen
			currentLocationApproximateCoordinates {
				timestamp
				value {
					latitude
					longitude
					hdop
				}
			}
			currentLocationCoordinates {
				timestamp
				value {
					latitude
					longitude
					hdop
				}
			}
		}
	}`

	result = ApproxResult{}
	err = telemetryClient.Post(query, &result, WithToken(token))
	require.Error(t, err)

	// partial results are still returned for the permitted field
	assert.Equal(t, locationTime.Format(time.RFC3339), result.SignalLatest.LastSeen)
	require.NotNil(t, result.SignalLatest.ApproxCoords)
	assert.Equal(t, expectedEndLatLong.Lat, result.SignalLatest.ApproxCoords.Value.Latitude)
	assert.Equal(t, expectedEndLatLong.Lng, result.SignalLatest.ApproxCoords.Value.Longitude)
	assert.Nil(t, result.SignalLatest.ExactCoords)

	fromTime := "2024-11-19T09:21:19Z"
	fromtTimePlus24 := "2024-11-20T09:21:19Z"
	query = fmt.Sprintf(`query {
		signals(tokenId:39718, from: "%s", to: "2025-04-27T09:21:19Z", interval:"24h"){
			timestamp
			currentLocationApproximateCoordinates(agg: FIRST) {
				latitude
				longitude
				hdop
			}
		}
	}`, fromTime)

	aggResult := ApproxAgg{}
	err = telemetryClient.Post(query, &aggResult, WithToken(token))
	require.NoError(t, err)

	require.Len(t, aggResult.Signals, 2)
	assert.Equal(t, fromTime, aggResult.Signals[0].Timestamp)
	require.NotNil(t, aggResult.Signals[0].ApproxCoords)
	assert.Equal(t, expectedStartLatLong.Lat, aggResult.Signals[0].ApproxCoords.Latitude)
	assert.Equal(t, expectedStartLatLong.Lng, aggResult.Signals[0].ApproxCoords.Longitude)
	assert.Equal(t, startHDOP, aggResult.Signals[0].ApproxCoords.Hdop)

	assert.Equal(t, fromtTimePlus24, aggResult.Signals[1].Timestamp)
	require.NotNil(t, aggResult.Signals[1].ApproxCoords)
	assert.Equal(t, expectedEndLatLong.Lat, aggResult.Signals[1].ApproxCoords.Latitude)
	assert.Equal(t, expectedEndLatLong.Lng, aggResult.Signals[1].ApproxCoords.Longitude)
	assert.Equal(t, endHDOP, aggResult.Signals[1].ApproxCoords.Hdop)

	// verify aggregation does not leak the exact location
	query = fmt.Sprintf(`query {
		signals(tokenId:39718, from: "%s", to: "2025-04-27T09:21:19Z", interval:"24h"){
			timestamp
			currentLocationApproximateCoordinates(agg: FIRST) {
				latitude
				longitude
				hdop
			}
			currentLocationCoordinates(agg: FIRST) {
				latitude
				longitude
				hdop
			}
		}
	}`, fromTime)

	aggResult = ApproxAgg{}
	err = telemetryClient.Post(query, &aggResult, WithToken(token))
	require.Error(t, err)

	// partial results are still returned for the permitted field
	require.Len(t, aggResult.Signals, 2)
	assert.Equal(t, fromTime, aggResult.Signals[0].Timestamp)
	require.NotNil(t, aggResult.Signals[0].ApproxCoords)
	assert.Equal(t, expectedStartLatLong.Lat, aggResult.Signals[0].ApproxCoords.Latitude)
	assert.Equal(t, expectedStartLatLong.Lng, aggResult.Signals[0].ApproxCoords.Longitude)
	assert.Nil(t, aggResult.Signals[0].ExactCoords)

	assert.Equal(t, fromtTimePlus24, aggResult.Signals[1].Timestamp)
	require.NotNil(t, aggResult.Signals[1].ApproxCoords)
	assert.Equal(t, expectedEndLatLong.Lat, aggResult.Signals[1].ApproxCoords.Latitude)
	assert.Equal(t, expectedEndLatLong.Lng, aggResult.Signals[1].ApproxCoords.Longitude)
	assert.Nil(t, aggResult.Signals[1].ExactCoords)
}

type approxCoordsValue struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Hdop      float64 `json:"hdop"`
}

type ApproxResult struct {
	SignalLatest struct {
		LastSeen     string `json:"lastSeen"`
		ApproxCoords *struct {
			Timestamp string             `json:"timestamp"`
			Value     *approxCoordsValue `json:"value"`
		} `json:"currentLocationApproximateCoordinates"`
		ExactCoords *struct {
			Timestamp string             `json:"timestamp"`
			Value     *approxCoordsValue `json:"value"`
		} `json:"currentLocationCoordinates"`
	} `json:"signalsLatest"`
}

type ApproxAgg struct {
	Signals []struct {
		Timestamp    string             `json:"timestamp"`
		ApproxCoords *approxCoordsValue `json:"currentLocationApproximateCoordinates"`
		ExactCoords  *approxCoordsValue `json:"currentLocationCoordinates"`
	} `json:"signals"`
}
