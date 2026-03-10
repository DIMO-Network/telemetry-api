package e2e_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
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
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "0x1234567890abcdef1234567890abcdef12345678",
				Subject: cloudevent.ERC721DID{
					ChainID:         services.Settings.ChainID,
					ContractAddress: services.Settings.VehicleNFTAddress,
					TokenID:         big.NewInt(39718),
				}.String(),
			},
			Data: vss.EventData{
				Name:       "behavior.harshBraking",
				Timestamp:  event1Time,
				DurationNs: 1000000, // 1ms
				Metadata:   `{"counter": 3}`,
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				Subject: cloudevent.ERC721DID{
					ChainID:         services.Settings.ChainID,
					ContractAddress: services.Settings.VehicleNFTAddress,
					TokenID:         big.NewInt(39718),
				}.String(),
			},
			Data: vss.EventData{
				Name:       "behavior.extremeBraking",
				Timestamp:  event2Time,
				DurationNs: 2000000, // 2ms
				Metadata:   `{"counter": 1}`,
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
				Subject: cloudevent.ERC721DID{
					ChainID:         services.Settings.ChainID,
					ContractAddress: services.Settings.VehicleNFTAddress,
					TokenID:         big.NewInt(39718),
				}.String(),
			},
			Data: vss.EventData{
				Name:       "energy.chargingStart",
				Timestamp:  event3Time,
				DurationNs: 500000, // 0.5ms
				Metadata:   `{"station": "tesla_supercharger"}`,
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
				Subject: cloudevent.ERC721DID{
					ChainID:         services.Settings.ChainID,
					ContractAddress: services.Settings.VehicleNFTAddress,
					TokenID:         big.NewInt(39718),
				}.String(),
			},
			Data: vss.EventData{
				Name:       "energy.chargingStart",
				Timestamp:  event4Time,
				DurationNs: 3000000, // 3ms
				Metadata:   "",      // Empty metadata
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				Subject: cloudevent.ERC721DID{
					ChainID:         services.Settings.ChainID,
					ContractAddress: services.Settings.VehicleNFTAddress,
					TokenID:         big.NewInt(39718),
				}.String(),
			},
			Data: vss.EventData{
				Name:       "behavior.harshBraking",
				Timestamp:  event1Time.Add(1 * time.Hour),
				DurationNs: 750000, // 0.75ms
				Metadata:   `{"counter": 2}`,
			},
		},
		// Event for different token (should not appear in results)
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "0x1234567890abcdef1234567890abcdef12345678",
				Subject: cloudevent.ERC721DID{
					ChainID:         services.Settings.ChainID,
					ContractAddress: services.Settings.VehicleNFTAddress,
					TokenID:         big.NewInt(99999),
				}.String(),
			},
			Data: vss.EventData{
				Name:       "behavior.harshBraking",
				Timestamp:  event1Time,
				DurationNs: 1000000,
				Metadata:   `{"counter": 99}`,
			},
		},
	}

	insertEvent(t, services.CH, events)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	// Create auth token for vehicle
	token := services.Auth.CreateVehicleToken(t, 39718, []string{tokenclaims.PermissionGetNonLocationHistory, tokenclaims.PermissionGetLocationHistory})

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
		assert.Equal(t, "behavior.harshBraking", result.Events[0].Name)
		assert.Equal(t, "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", result.Events[0].Source)
		assert.Equal(t, event1Time.Add(1*time.Hour).Format(time.RFC3339), result.Events[0].Timestamp)
		assert.NotNil(t, result.Events[0].Metadata)
		assert.Equal(t, `{"counter": 2}`, *result.Events[0].Metadata)

		assert.Equal(t, "energy.chargingStart", result.Events[1].Name)
		assert.Equal(t, "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed", result.Events[1].Source)
		assert.Equal(t, event4Time.Format(time.RFC3339), result.Events[1].Timestamp)
		assert.Equal(t, 3000000, result.Events[1].DurationNs)
		assert.Nil(t, result.Events[1].Metadata) // Empty metadata should be nil

		assert.Equal(t, "energy.chargingStart", result.Events[2].Name)
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
			assert.Equal(t, "energy.chargingStart", event.Name)
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
				filter: {name: {eq: "behavior.harshBraking"}}
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
			assert.Equal(t, "behavior.harshBraking", event.Name)
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
				filter: {name: {eq: "behavior.extremeBraking"}, source: {eq: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"}}
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
		assert.Equal(t, "behavior.extremeBraking", event.Name)
		assert.Equal(t, "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd", event.Source)
		assert.NotNil(t, event.Metadata)
		assert.Equal(t, `{"counter": 1}`, *event.Metadata)
	})

}

func TestEventTags(t *testing.T) {
	services := GetTestServices(t)

	baseTime := time.Date(2024, 11, 20, 10, 0, 0, 0, time.UTC)
	subject := cloudevent.ERC721DID{
		ChainID:         services.Settings.ChainID,
		ContractAddress: services.Settings.VehicleNFTAddress,
		TokenID:         big.NewInt(39718),
	}.String()

	events := []vss.Event{
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source:  "0x1111111111111111111111111111111111111111",
				Subject: subject,
			},
			Data: vss.EventData{
				Name:       "behavior.harshBraking",
				Timestamp:  baseTime,
				DurationNs: 1000,
				Tags:       []string{"behavior.harshAcceleration", "behavior.harshBraking"},
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source:  "0x2222222222222222222222222222222222222222",
				Subject: subject,
			},
			Data: vss.EventData{
				Name:       "safety.collision",
				Timestamp:  baseTime.Add(5 * time.Minute),
				DurationNs: 2000,
				Tags:       []string{"safety.collision"},
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source:  "0x3333333333333333333333333333333333333333",
				Subject: subject,
			},
			Data: vss.EventData{
				Name:       "behavior.harshAcceleration",
				Timestamp:  baseTime.Add(10 * time.Minute),
				DurationNs: 3000,
				Tags:       []string{"behavior.harshAcceleration"},
			},
		},
	}

	insertEvent(t, services.CH, events)

	telemetryClient := NewGraphQLServer(t, services.Settings)
	token := services.Auth.CreateVehicleToken(t, 39718, []string{tokenclaims.PermissionGetNonLocationHistory, tokenclaims.PermissionGetLocationHistory})

	t.Run("filter by tags containsAny returns matching events", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-20T09:00:00Z",
				to: "2024-11-20T11:00:00Z",
				filter: {tags: {containsAny: ["safety.collision"]}}
			) {
				name
				source
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)
		require.Len(t, result.Events, 1)
		assert.Equal(t, "safety.collision", result.Events[0].Name)
		assert.Equal(t, "0x2222222222222222222222222222222222222222", result.Events[0].Source)
	})

	t.Run("filter by tags containsAny matches multiple events", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-20T09:00:00Z",
				to: "2024-11-20T11:00:00Z",
				filter: {tags: {containsAny: ["behavior.harshAcceleration"]}}
			) {
				name
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)
		require.Len(t, result.Events, 2)
		// Ordered DESC by timestamp
		assert.Equal(t, "behavior.harshAcceleration", result.Events[0].Name)
		assert.Equal(t, "behavior.harshBraking", result.Events[1].Name)
	})

	t.Run("filter by tags containsAll requires all tags present", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-20T09:00:00Z",
				to: "2024-11-20T11:00:00Z",
				filter: {tags: {containsAll: ["behavior.harshAcceleration", "behavior.harshBraking"]}}
			) {
				name
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)
		require.Len(t, result.Events, 1)
		assert.Equal(t, "behavior.harshBraking", result.Events[0].Name)
	})

	t.Run("filter by tags no match returns empty", func(t *testing.T) {
		query := `
		query {
			events(
				tokenId: 39718,
				from: "2024-11-20T09:00:00Z",
				to: "2024-11-20T11:00:00Z",
				filter: {tags: {containsAll: ["behavior.harshAcceleration", "safety.collision"]}}
			) {
				name
			}
		}`

		result := EventsResult{}
		err := telemetryClient.Post(query, &result, WithToken(token))
		require.NoError(t, err)
		require.Len(t, result.Events, 0)
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
