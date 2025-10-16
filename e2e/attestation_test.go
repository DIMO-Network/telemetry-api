package e2e_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttestations(t *testing.T) {
	testAttestation := cloudevent.CloudEvent[json.RawMessage]{
		CloudEventHeader: cloudevent.CloudEventHeader{
			ID:          "2oYEQJDQFRYCI1UIf6QTy92MS3G",
			Source:      "0x0000000000000000000000000000000000000000",
			Producer:    "did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA:31653",
			Subject:     "did:nft:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:39718",
			Time:        time.Date(2024, 11, 8, 3, 54, 51, 0, time.UTC),
			Type:        "dimo.attestation",
			DataVersion: "test-data-version",
			Tags:        []string{"tag1"},
		},
		Data: json.RawMessage(`{"test":"test"}`),
	}

	services := GetTestServices(t)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	services.FetchServer.SetCloudEventReturn(testAttestation)

	// Create auth token
	token := services.Auth.CreateVehicleToken(t, 39718, nil, tokenclaims.Event{
		EventType: "dimo.attestation",
		Source:    tokenclaims.GlobalIdentifier,
		IDs:       []string{tokenclaims.GlobalIdentifier},
		Tags:      []string{tokenclaims.GlobalIdentifier},
	})

	query := `
	query attestations {
		attestations(tokenId: 39718) {
			id
			time
			attestation
			type
			source
			dataVersion
			tags
		}
	}`

	var result struct {
		Attestations []struct {
			ID          string   `json:"id"`
			Time        string   `json:"time"`
			Attestation string   `json:"attestation"`
			Type        string   `json:"type"`
			Source      string   `json:"source"`
			DataVersion string   `json:"dataVersion"`
			Tags        []string `json:"tags"`
		} `json:"attestations"`
	}

	err := telemetryClient.Post(query, &result, WithToken(token))
	require.NoError(t, err)

	actual := result.Attestations
	expectedJSON, err := json.Marshal(testAttestation)
	require.NoError(t, err)
	require.Equal(t, 1, len(result.Attestations))

	assert.JSONEq(t, string(expectedJSON), actual[0].Attestation)
	assert.Equal(t, "2oYEQJDQFRYCI1UIf6QTy92MS3G", actual[0].ID)
	assert.Equal(t, "2024-11-08T03:54:51Z", actual[0].Time)
	assert.Equal(t, "dimo.attestation", actual[0].Type)
	assert.Equal(t, "0x0000000000000000000000000000000000000000", actual[0].Source)
	assert.Equal(t, "test-data-version", actual[0].DataVersion)
	assert.Equal(t, []string{"tag1"}, actual[0].Tags)
}

func TestAttestationAddressSubject(t *testing.T) {

	testAttestation := cloudevent.CloudEvent[json.RawMessage]{
		CloudEventHeader: cloudevent.CloudEventHeader{
			ID:          "2oYEQJDQFRYCI1UIf6QTy92MS3G",
			Source:      "0x0000000000000000000000000000000000000000",
			Producer:    "did:nft:137:0x9c94C395cBcBDe662235E0A9d3bB87Ad708561BA:31653",
			Subject:     "did:ethr:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF",
			Time:        time.Date(2024, 11, 8, 3, 54, 51, 0, time.UTC),
			Type:        "dimo.attestation",
			DataVersion: "test-data-version",
			Tags:        []string{"tag1"},
		},
		Data: json.RawMessage(`{"test":"test"}`),
	}

	services := GetTestServices(t)

	// Create and set up GraphQL server
	telemetryClient := NewGraphQLServer(t, services.Settings)

	services.FetchServer.SetCloudEventReturn(testAttestation)
	asset := cloudevent.EthrDID{
		ChainID:         uint64(137),
		ContractAddress: common.HexToAddress("0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF"),
	}
	// Create auth token
	token := services.Auth.CreateUserToken(t, asset.ContractAddress, nil, tokenclaims.Event{
		EventType: "dimo.attestation",
		Source:    tokenclaims.GlobalIdentifier,
		IDs:       []string{tokenclaims.GlobalIdentifier},
		Tags:      []string{tokenclaims.GlobalIdentifier},
	})

	query := `
	query attestations {
		attestations(subject: "` + asset.String() + `") {
			id
			time
			attestation
			type
			source
			dataVersion
			tags
		}
	}`

	var result struct {
		Attestations []struct {
			ID          string   `json:"id"`
			Time        string   `json:"time"`
			Attestation string   `json:"attestation"`
			Type        string   `json:"type"`
			Source      string   `json:"source"`
			DataVersion string   `json:"dataVersion"`
			Tags        []string `json:"tags"`
		} `json:"attestations"`
	}

	err := telemetryClient.Post(query, &result, WithToken(token))
	require.NoError(t, err)

	actual := result.Attestations
	expectedJSON, err := json.Marshal(testAttestation)
	require.NoError(t, err)
	require.Equal(t, 1, len(result.Attestations))

	assert.JSONEq(t, string(expectedJSON), actual[0].Attestation)
	assert.Equal(t, "2oYEQJDQFRYCI1UIf6QTy92MS3G", actual[0].ID)
	assert.Equal(t, "2024-11-08T03:54:51Z", actual[0].Time)
	assert.Equal(t, "dimo.attestation", actual[0].Type)
	assert.Equal(t, "0x0000000000000000000000000000000000000000", actual[0].Source)
	assert.Equal(t, "test-data-version", actual[0].DataVersion)
	assert.Equal(t, []string{"tag1"}, actual[0].Tags)
}
