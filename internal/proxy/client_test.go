package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientExecute_ForwardsAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{}}`)
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), AuthHeaderKey{}, "Bearer test-token")
	client := NewClient(srv.URL)
	_, err := client.Execute(ctx, `{__typename}`, nil)
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token", gotAuth)
}

func TestClientExecute_NoAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{}}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.Execute(context.Background(), `{__typename}`, nil)
	require.NoError(t, err)
	assert.Empty(t, gotAuth)
}

func TestClientExecute_SendsVariables(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"foo":1}}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	vars := map[string]any{"subject": "did:erc721:137:0xabc:1"}
	_, err := client.Execute(context.Background(), `query($subject: String!) { foo }`, vars)
	require.NoError(t, err)

	assert.Equal(t, `query($subject: String!) { foo }`, gotBody["query"])
	assert.Equal(t, "did:erc721:137:0xabc:1", gotBody["variables"].(map[string]any)["subject"])
}

func TestClientExecute_GraphQLErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errors":[{"message":"unauthorized"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.Execute(context.Background(), `{__typename}`, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestClientExecute_ReturnsDataJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"speed":42.0}}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	data, err := client.Execute(context.Background(), `{speed}`, nil)
	require.NoError(t, err)

	var got map[string]float64
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 42.0, got["speed"])
}

func TestProxySignals_RoundTrip(t *testing.T) {
	ts := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request contains the right subject and includes timestamp + field selection.
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		vars := body["variables"].(map[string]any)
		assert.Equal(t, "did:erc721:137:0xabc:1", vars["subject"])
		assert.Equal(t, "1h", vars["interval"])

		query := body["query"].(string)
		assert.Contains(t, query, `mySpeed: speed(agg: LAST)`)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"signals":[{"timestamp":"2024-01-01T12:00:00Z","mySpeed":77.5}]}}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{
			{Name: "speed", Agg: model.FloatAggregationLast, Alias: "mySpeed"},
		},
	}

	result, err := client.ProxySignals(
		context.Background(),
		"did:erc721:137:0xabc:1", "1h",
		ts, ts.Add(time.Hour), nil,
		aggArgs,
	)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, ts, result[0].Timestamp)
	assert.Equal(t, 77.5, result[0].ValueNumbers["mySpeed"])
}

func TestProxySegments_MechanismMapped(t *testing.T) {
	var gotMechanism string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		gotMechanism = body["variables"].(map[string]any)["mechanism"].(string)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"segments":[]}}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	now := time.Now()
	_, err := client.ProxySegments(
		context.Background(),
		"did:erc721:137:0xabc:1",
		now, now.Add(time.Hour),
		model.DetectionMechanismIgnitionDetection,
		nil, nil, nil, nil, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "IGNITION_DETECTION", gotMechanism)
}
