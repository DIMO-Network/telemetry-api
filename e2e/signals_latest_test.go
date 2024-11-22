package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
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
			Source:      ch.SourceTranslations["smartcar"],
			Timestamp:   smartCarTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 65.5,
			TokenID:     39718,
		},
		{
			Source:      ch.SourceTranslations["autopi"],
			Timestamp:   autopiTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 14,
			TokenID:     39718,
		},
		{
			Source:      ch.SourceTranslations["macaron"],
			Timestamp:   macaronTime,
			Name:        vss.FieldSpeed,
			ValueNumber: 3,
			TokenID:     39718,
		},
	}

	insertSignal(t, services.CH, signals)

	// Create and set up GraphQL server
	server := NewGraphQLServer(t, services.Settings)
	defer server.Close()

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, "39718", []int{1}) // Assuming 1 is the privilege needed

	// Execute the query
	query := `
	query Latest_all {
		smartcar: signalsLatest(filter: {source: "smartcar"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
		}
		autopi: signalsLatest(filter: {source: "autopi"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
		}
		macaron: signalsLatest(filter: {source: "macaron"}, tokenId: 39718) {
			lastSeen
			speed {
				timestamp
				value
			}
		}
		tesla: signalsLatest(filter: {source: "tesla"}, tokenId: 39718) {
			lastSeen
		}
	}`

	// Create request
	body, err := json.Marshal(map[string]interface{}{
		"query": query,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var result struct {
		Data struct {
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
		} `json:"data"`
	}
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	fmt.Println(string(data))
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Assert the results
	assert.Equal(t, smartCarTime.Format(time.RFC3339), result.Data.Smartcar.LastSeen)
	assert.Equal(t, smartCarTime.Format(time.RFC3339), result.Data.Smartcar.Speed.Timestamp)
	require.NotNil(t, result.Data.Autopi.Speed)

	assert.Equal(t, autopiTime.Format(time.RFC3339), result.Data.Autopi.LastSeen)
	require.NotNil(t, result.Data.Autopi.Speed)
	assert.Equal(t, autopiTime.Format(time.RFC3339), result.Data.Autopi.Speed.Timestamp)
	assert.Equal(t, float64(14), result.Data.Autopi.Speed.Value)

	assert.Equal(t, macaronTime.Format(time.RFC3339), result.Data.Macaron.LastSeen)
	require.NotNil(t, result.Data.Macaron.Speed)
	assert.Equal(t, macaronTime.Format(time.RFC3339), result.Data.Macaron.Speed.Timestamp)
	assert.Equal(t, float64(3), result.Data.Macaron.Speed.Value)

	assert.Nil(t, result.Data.Tesla.LastSeen)
}

type SignalWithTime struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}
