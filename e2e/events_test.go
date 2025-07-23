package e2e_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	services := GetTestServices(t)

	// Define test times
	baseTime := time.Date(2024, 11, 15, 12, 0, 0, 0, time.UTC)
	event1Time := baseTime
	event2Time := baseTime.Add(5 * time.Minute)
	event3Time := baseTime.Add(10 * time.Minute)
	event4Time := baseTime.Add(15 * time.Minute)

	// Set up test event data with different sources and names
	// Use real-world event names and 0x addresses for sources
	events := []vss.Event{
		{
			Name:       "harshBraking",
			Source:     "0x1234567890abcdef1234567890abcdef12345678",
			Timestamp:  event1Time,
			DurationNs: 1000000, // 1ms
			Subject: cloudevent.ERC721DID{
				ChainID:         services.Settings.ChainID,
				ContractAddress: services.Settings.VehicleNFTAddress,
				TokenID:         big.NewInt(39718),
			}.String(),
			Metadata: `{"counter": 3}`,
		},
		{
			Name:       "extremeBraking",
			Source:     "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			Timestamp:  event2Time,
			DurationNs: 2000000, // 2ms
			Subject: cloudevent.ERC721DID{
				ChainID:         services.Settings.ChainID,
				ContractAddress: services.Settings.VehicleNFTAddress,
				TokenID:         big.NewInt(39718),
			}.String(),
			Metadata: `{"counter": 1}`,
		},
		{
			Name:       "chargingStart",
			Source:     "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
			Timestamp:  event3Time,
			DurationNs: 500000, // 0.5ms
			Subject: cloudevent.ERC721DID{
				ChainID:         services.Settings.ChainID,
				ContractAddress: services.Settings.VehicleNFTAddress,
				TokenID:         big.NewInt(39718),
			}.String(),
			Metadata: `{"station": "tesla_supercharger"}`,
		},
		{
			Name:       "chargingStart",
			Source:     "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
			Timestamp:  event4Time,
			DurationNs: 3000000, // 3ms
			Subject: cloudevent.ERC721DID{
				ChainID:         services.Settings.ChainID,
				ContractAddress: services.Settings.VehicleNFTAddress,
				TokenID:         big.NewInt(39718),
			}.String(),
			Metadata: "", // Empty metadata
		},
		{
			Name:       "harshBraking",
			Source:     "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Timestamp:  event1Time.Add(1 * time.Hour),
			DurationNs: 750000, // 0.75ms
			Subject: cloudevent.ERC721DID{
				ChainID:         services.Settings.ChainID,
				ContractAddress: services.Settings.VehicleNFTAddress,
				TokenID:         big.NewInt(39718),
			}.String(),
			Metadata: `{"counter": 2}`,
		},
		// Event for different token (should not appear in results)
		{
			Name:       "harshBraking",
			Source:     "0x1234567890abcdef1234567890abcdef12345678",
			Timestamp:  event1Time,
			DurationNs: 1000000,
			Subject: cloudevent.ERC721DID{
				ChainID:         services.Settings.ChainID,
				ContractAddress: services.Settings.VehicleNFTAddress,
				TokenID:         big.NewInt(99999),
			}.String(),
			Metadata: `{"counter": 99}`,
		},
	}

	insertEvent(t, services.CH, events)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, "39718", []int{1})

	t.Run("Basic events query without filter", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718, 
				from: "2024-11-15T11:00:00Z", 
				to: "2024-11-15T14:00:00Z"
			) {
				timestamp
				name
				source
				durationNs
				metadata
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)

		// Should return all events for this token in the time range (5 events)
		require.Len(t, result.Events, 5)

		// Events should be ordered by timestamp DESC (newest first)
		assert.Equal(t, "harshBraking", result.Events[0].Name)
		assert.Equal(t, "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", result.Events[0].Source)
		assert.Equal(t, event1Time.Add(1*time.Hour).Format(time.RFC3339), result.Events[0].Timestamp)
		assert.NotNil(t, result.Events[0].Metadata)
		assert.Equal(t, `{"counter": 2}`, *result.Events[0].Metadata)

		assert.Equal(t, "chargingStart", result.Events[1].Name)
		assert.Equal(t, "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed", result.Events[1].Source)
		assert.Equal(t, event4Time.Format(time.RFC3339), result.Events[1].Timestamp)
		assert.Equal(t, 3000000, result.Events[1].DurationNs)
		assert.Nil(t, result.Events[1].Metadata) // Empty metadata should be nil

		assert.Equal(t, "chargingStart", result.Events[2].Name)
		assert.Equal(t, "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed", result.Events[2].Source)
		assert.Equal(t, event3Time.Format(time.RFC3339), result.Events[2].Timestamp)
		assert.Equal(t, 500000, result.Events[2].DurationNs)
		assert.NotNil(t, result.Events[2].Metadata)
		assert.Equal(t, `{"station": "tesla_supercharger"}`, *result.Events[2].Metadata)
	})

	// Filter by source (0xfeedfeed...)
	t.Run("Filter by source address", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-15T11:00:00Z",
				to: "2024-11-15T14:00:00Z",
				filter: {source: {eq: "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed"}}
			) {
				timestamp
				name
				source
				durationNs
				metadata
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)

		// Should return only chargingStart events from 0xfeedfeed... (2 events)
		require.Len(t, result.Events, 2)
		for _, event := range result.Events {
			assert.Equal(t, "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed", event.Source)
			assert.Equal(t, "chargingStart", event.Name)
		}
	})

	// Filter by name (harshBraking)
	t.Run("Filter by name", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-15T11:00:00Z",
				to: "2024-11-15T14:00:00Z",
				filter: {name: {eq: "harshBraking"}}
			) {
				timestamp
				name
				source
				durationNs
				metadata
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)

		// Should return only harshBraking events (2 events)
		require.Len(t, result.Events, 2)
		for _, event := range result.Events {
			assert.Equal(t, "harshBraking", event.Name)
			assert.NotNil(t, event.Metadata)
		}
		// Check that both sources are represented
		sources := []string{result.Events[0].Source, result.Events[1].Source}
		assert.Contains(t, sources, "0x1234567890abcdef1234567890abcdef12345678")
		assert.Contains(t, sources, "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	})

	// Filter by both name and source
	t.Run("Filter by both name and source", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-15T11:00:00Z",
				to: "2024-11-15T14:00:00Z",
				filter: {name: {eq: "extremeBraking"}, source: {eq: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"}}
			) {
				timestamp
				name
				source
				durationNs
				metadata
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)

		// Should return only the one matching event
		require.Len(t, result.Events, 1)
		event := result.Events[0]
		assert.Equal(t, "extremeBraking", event.Name)
		assert.Equal(t, "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd", event.Source)
		assert.NotNil(t, event.Metadata)
		assert.Equal(t, `{"counter": 1}`, *event.Metadata)
	})

}

type EventsResult struct {
	Events []EventData `json:"events"`
}

type EventData struct {
	Timestamp  string  `json:"timestamp"`
	Name       string  `json:"name"`
	Source     string  `json:"source"`
	DurationNs int     `json:"durationNs"`
	Metadata   *string `json:"metadata"`
}
