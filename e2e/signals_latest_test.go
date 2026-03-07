package e2e_test

import (
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalsLatest(t *testing.T) {
	services := GetTestServices(t)
	smartCarTime := time.Date(2024, 11, 20, 22, 28, 17, 0, time.UTC)
	autopiTime := time.Date(2024, 11, 1, 20, 1, 29, 0, time.UTC)
	macaronTime := time.Date(2024, 3, 20, 20, 13, 57, 0, time.UTC)
	// Set up test data in Clickhouse
	signals := []vss.Signal{
		{
			Source:      "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E",
			Timestamp:   smartCarTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 65.5,
			Subject:     "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
		{
			Source:    "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E",
			Timestamp: smartCarTime,
			Name:      vss.FieldCurrentLocationCoordinates,
			ValueLocation: vss.Location{
				Latitude:  40.73899538333504,
				Longitude: 73.99386110247163,
			},
			Subject: "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
		{
			Source:    "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E",
			Timestamp: smartCarTime.Add(time.Hour),
			Name:      vss.FieldCurrentLocationCoordinates,
			ValueLocation: vss.Location{
				HDOP: 7,
			},
			Subject: "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
		{
			Source:      "0x5e31bBc786D7bEd95216383787deA1ab0f1c1897",
			Timestamp:   autopiTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 14,
			Subject:     "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
		{
			Source:      "0x4c674ddE8189aEF6e3b58F5a36d7438b2b1f6Bc2",
			Timestamp:   macaronTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 3,
			Subject:     "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
		},
	}

	insertSignal(t, services.CH, signals)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, 39718, []string{tokenclaims.PermissionGetNonLocationHistory, tokenclaims.PermissionGetLocationHistory})

	// Execute the query
	query := `
	query Latest_all {
		smartcar: signalsLatest(filter: {source: "did:ethr:137:0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
		}
		autopi: signalsLatest(filter: {source: "did:ethr:137:0x5e31bBc786D7bEd95216383787deA1ab0f1c1897"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
		}
		macaron: signalsLatest(filter: {source: "did:ethr:137:0x4c674ddE8189aEF6e3b58F5a36d7438b2b1f6Bc2"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
		}
		tesla: signalsLatest(filter: {source: "did:ethr:137:0xc4035Fecb1cc906130423EF05f9C20977F643722"}, tokenId: 39718) {
			lastSeen
		}
		location: signalsLatest(tokenId: 39718) {
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

	// Execute request
	result := LatestResult{}
	err := telemetryClient.Post(query, &result, WithToken(token))
	require.NoError(t, err)

	// Assert the results
	assert.Equal(t, smartCarTime.Add(time.Hour).Format(time.RFC3339), result.Smartcar.LastSeen)
	assert.Equal(t, smartCarTime.Format(time.RFC3339), result.Smartcar.Speed.Timestamp)
	require.NotNil(t, result.Autopi.Speed)

	assert.Equal(t, autopiTime.Format(time.RFC3339), result.Autopi.LastSeen)
	require.NotNil(t, result.Autopi.Speed)
	assert.Equal(t, autopiTime.Format(time.RFC3339), result.Autopi.Speed.Timestamp)
	assert.Equal(t, float64(14), result.Autopi.Speed.Value)

	assert.Equal(t, macaronTime.Format(time.RFC3339), result.Macaron.LastSeen)
	require.NotNil(t, result.Macaron.Speed)
	assert.Equal(t, macaronTime.Format(time.RFC3339), result.Macaron.Speed.Timestamp)
	assert.Equal(t, float64(3), result.Macaron.Speed.Value)

	require.NotNil(t, result.Location.CurrentLocationCoordinates)
	assert.Equal(t, smartCarTime.Format(time.RFC3339), result.Location.CurrentLocationCoordinates.Timestamp)
	assert.Equal(t, 40.73899538333504, result.Location.CurrentLocationCoordinates.Value.Latitude)
	assert.Equal(t, 73.99386110247163, result.Location.CurrentLocationCoordinates.Value.Longitude)

	assert.Nil(t, result.Tesla.LastSeen)
}

type SignalWithTime struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type LocationSignalWithTime struct {
	Timestamp string         `json:"timestamp"`
	Value     model.Location `json:"value"`
}

type LatestResult struct {
	Smartcar struct {
		LastSeen string          `json:"lastSeen"`
		Speed    *SignalWithTime `json:"speed"`
	} `json:"smartcar"`
	Autopi struct {
		LastSeen string          `json:"lastSeen"`
		Speed    *SignalWithTime `json:"speed"`
	} `json:"autopi"`
	Macaron struct {
		LastSeen string          `json:"lastSeen"`
		Speed    *SignalWithTime `json:"speed"`
	} `json:"macaron"`
	Tesla struct {
		LastSeen *string `json:"lastSeen"`
	} `json:"tesla"`
	Location struct {
		LastSeen                   string                  `json:"lastSeen"`
		CurrentLocationCoordinates *LocationSignalWithTime `json:"currentLocationCoordinates"`
	} `json:"location"`
}
