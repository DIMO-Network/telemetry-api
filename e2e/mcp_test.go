package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/app"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpTestTokenID = 77001

// newMCPServer builds the application and serves its MCP handler over HTTP.
func newMCPServer(t *testing.T, settings config.Settings) *httptest.Server {
	t.Helper()

	zerolog.SetGlobalLevel(zerolog.PanicLevel)
	testLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &testLogger

	application, err := app.New(settings)
	require.NoError(t, err)
	t.Cleanup(application.Cleanup)

	server := httptest.NewServer(application.MCPHandler)
	t.Cleanup(server.Close)
	return server
}

// mcpRequest posts a single JSON-RPC request to the stateless MCP endpoint and
// returns the decoded "result" object. Responses may arrive as SSE
// ("data: {...}") or plain JSON depending on transport negotiation.
func mcpRequest(t *testing.T, serverURL, token, method string, params any) map[string]any {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck

	raw := new(bytes.Buffer)
	_, err = raw.ReadFrom(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "MCP response body: %s", raw.String())

	payload := raw.String()
	if idx := strings.Index(payload, "data: "); idx >= 0 {
		payload = payload[idx+len("data: "):]
		if end := strings.IndexByte(payload, '\n'); end >= 0 {
			payload = payload[:end]
		}
	}

	var rpc struct {
		Result map[string]any `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal([]byte(payload), &rpc), "raw payload: %s", payload)
	require.Nil(t, rpc.Error, "JSON-RPC error from MCP server")
	return rpc.Result
}

// callTool invokes an MCP tool and returns (text content, isError).
func callTool(t *testing.T, serverURL, token, name string, args map[string]any) (string, bool) {
	t.Helper()
	result := mcpRequest(t, serverURL, token, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})

	content, ok := result["content"].([]any)
	require.True(t, ok && len(content) > 0, "tool result missing content: %v", result)
	first, ok := content[0].(map[string]any)
	require.True(t, ok)
	text, _ := first["text"].(string)
	isError, _ := result["isError"].(bool)
	return text, isError
}

func TestMCPToolsList(t *testing.T) {
	services := GetTestServices(t)
	server := newMCPServer(t, services.Settings)

	result := mcpRequest(t, server.URL, "", "tools/list", map[string]any{})
	toolsRaw, ok := result["tools"].([]any)
	require.True(t, ok, "tools/list result missing tools: %v", result)

	tools := make(map[string]map[string]any, len(toolsRaw))
	for _, raw := range toolsRaw {
		tool, ok := raw.(map[string]any)
		require.True(t, ok)
		tools[tool["name"].(string)] = tool
	}

	// Builtin tools.
	require.Contains(t, tools, "telemetry_get_schema")
	require.Contains(t, tools, "telemetry_query")

	// The two overridden shortcut tools must require their ToolOnly signal args.
	timeSeries, ok := tools["telemetry_get_signals_time_series"]
	require.True(t, ok, "overridden time-series tool missing")
	assert.Contains(t, requiredArgs(t, timeSeries), "signalRequests")

	latest, ok := tools["telemetry_get_latest_signals"]
	require.True(t, ok, "overridden latest-signals tool missing")
	assert.Contains(t, requiredArgs(t, latest), "signalNames")
}

func requiredArgs(t *testing.T, tool map[string]any) []string {
	t.Helper()
	schema, ok := tool["inputSchema"].(map[string]any)
	require.True(t, ok, "tool %v missing inputSchema", tool["name"])
	rawRequired, _ := schema["required"].([]any)
	required := make([]string, 0, len(rawRequired))
	for _, r := range rawRequired {
		required = append(required, r.(string))
	}
	return required
}

func TestMCPSignalTools(t *testing.T) {
	services := GetTestServices(t)

	baseTime := time.Date(2024, 11, 20, 10, 0, 0, 0, time.UTC)
	subject := "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:77001"
	source := "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"
	signals := []vss.Signal{
		{
			CloudEventHeader: cloudevent.CloudEventHeader{Source: source, Subject: subject},
			Data:             vss.SignalData{Timestamp: baseTime.Add(15 * time.Minute), Name: vss.FieldSpeed, ValueNumber: 30},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{Source: source, Subject: subject},
			Data:             vss.SignalData{Timestamp: baseTime.Add(45 * time.Minute), Name: vss.FieldSpeed, ValueNumber: 80},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{Source: source, Subject: subject},
			Data:             vss.SignalData{Timestamp: baseTime.Add(30 * time.Minute), Name: vss.FieldPowertrainTractionBatteryStateOfChargeCurrent, ValueNumber: 50},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{Source: source, Subject: subject},
			Data:             vss.SignalData{Timestamp: baseTime.Add(90 * time.Minute), Name: vss.FieldPowertrainTractionBatteryStateOfChargeCurrent, ValueNumber: 42.5},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{Source: source, Subject: subject},
			Data: vss.SignalData{
				Timestamp:     baseTime.Add(60 * time.Minute),
				Name:          vss.FieldCurrentLocationCoordinates,
				ValueLocation: vss.Location{Latitude: 42.615208, Longitude: -83.029093, HDOP: 5},
			},
		},
	}
	insertSignal(t, services.CH, signals)

	server := newMCPServer(t, services.Settings)
	token := services.Auth.CreateVehicleToken(t, mcpTestTokenID, []string{tokenclaims.PermissionGetNonLocationHistory, tokenclaims.PermissionGetLocationHistory})

	t.Run("latest signals returns data", func(t *testing.T) {
		text, isError := callTool(t, server.URL, token, "telemetry_get_latest_signals", map[string]any{
			"tokenId":     mcpTestTokenID,
			"signalNames": []string{"speed", "powertrainTractionBatteryStateOfChargeCurrent"},
		})
		require.False(t, isError, "tool error: %s", text)

		var resp struct {
			Data struct {
				SignalsLatest struct {
					LastSeen string `json:"lastSeen"`
					Speed    struct {
						Timestamp string  `json:"timestamp"`
						Value     float64 `json:"value"`
					} `json:"speed"`
					SoC struct {
						Value float64 `json:"value"`
					} `json:"powertrainTractionBatteryStateOfChargeCurrent"`
				} `json:"signalsLatest"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		require.NoError(t, json.Unmarshal([]byte(text), &resp))
		require.Empty(t, resp.Errors, "GraphQL errors: %s", text)

		assert.Equal(t, 80.0, resp.Data.SignalsLatest.Speed.Value)
		assert.Equal(t, baseTime.Add(45*time.Minute).Format(time.RFC3339), resp.Data.SignalsLatest.Speed.Timestamp)
		assert.Equal(t, 42.5, resp.Data.SignalsLatest.SoC.Value)
		assert.Equal(t, baseTime.Add(90*time.Minute).Format(time.RFC3339), resp.Data.SignalsLatest.LastSeen)
	})

	t.Run("time series aggregates named signals", func(t *testing.T) {
		text, isError := callTool(t, server.URL, token, "telemetry_get_signals_time_series", map[string]any{
			"tokenId":  mcpTestTokenID,
			"interval": "1h",
			"from":     baseTime.Format(time.RFC3339),
			"to":       baseTime.Add(2 * time.Hour).Format(time.RFC3339),
			"signalRequests": []map[string]any{
				{"name": "speed", "agg": "MAX"},
				{"name": "powertrainTractionBatteryStateOfChargeCurrent", "agg": "LAST"},
			},
		})
		require.False(t, isError, "tool error: %s", text)

		var resp struct {
			Data struct {
				Signals []struct {
					Timestamp string   `json:"timestamp"`
					Speed     *float64 `json:"speed"`
					SoC       *float64 `json:"powertrainTractionBatteryStateOfChargeCurrent"`
				} `json:"signals"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		require.NoError(t, json.Unmarshal([]byte(text), &resp))
		require.Empty(t, resp.Errors, "GraphQL errors: %s", text)
		require.Len(t, resp.Data.Signals, 2)

		first := resp.Data.Signals[0]
		require.NotNil(t, first.Speed)
		assert.Equal(t, 80.0, *first.Speed, "MAX of speed in first bucket")
		require.NotNil(t, first.SoC)
		assert.Equal(t, 50.0, *first.SoC)

		second := resp.Data.Signals[1]
		assert.Nil(t, second.Speed, "no speed data in second bucket")
		require.NotNil(t, second.SoC)
		assert.Equal(t, 42.5, *second.SoC, "LAST of SoC in second bucket")
	})

	t.Run("latest signals returns location values with subfields", func(t *testing.T) {
		text, isError := callTool(t, server.URL, token, "telemetry_get_latest_signals", map[string]any{
			"tokenId":     mcpTestTokenID,
			"signalNames": []string{"speed", "currentLocationCoordinates"},
		})
		require.False(t, isError, "tool error: %s", text)

		var resp struct {
			Data struct {
				SignalsLatest struct {
					Coordinates struct {
						Timestamp string `json:"timestamp"`
						Value     struct {
							Latitude  float64 `json:"latitude"`
							Longitude float64 `json:"longitude"`
							HDOP      float64 `json:"hdop"`
						} `json:"value"`
					} `json:"currentLocationCoordinates"`
				} `json:"signalsLatest"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		require.NoError(t, json.Unmarshal([]byte(text), &resp))
		require.Empty(t, resp.Errors, "GraphQL errors: %s", text)

		coords := resp.Data.SignalsLatest.Coordinates
		assert.Equal(t, baseTime.Add(60*time.Minute).Format(time.RFC3339), coords.Timestamp)
		assert.Equal(t, 42.615208, coords.Value.Latitude)
		assert.Equal(t, -83.029093, coords.Value.Longitude)
		assert.Equal(t, 5.0, coords.Value.HDOP)
	})

	t.Run("time series aggregates location signals", func(t *testing.T) {
		text, isError := callTool(t, server.URL, token, "telemetry_get_signals_time_series", map[string]any{
			"tokenId":  mcpTestTokenID,
			"interval": "1h",
			"from":     baseTime.Format(time.RFC3339),
			"to":       baseTime.Add(2 * time.Hour).Format(time.RFC3339),
			"signalRequests": []map[string]any{
				{"name": "currentLocationCoordinates", "agg": "LAST"},
			},
		})
		require.False(t, isError, "tool error: %s", text)

		var resp struct {
			Data struct {
				Signals []struct {
					Coordinates *struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
					} `json:"currentLocationCoordinates"`
				} `json:"signals"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		require.NoError(t, json.Unmarshal([]byte(text), &resp))
		require.Empty(t, resp.Errors, "GraphQL errors: %s", text)
		require.NotEmpty(t, resp.Data.Signals)

		require.NotNil(t, resp.Data.Signals[0].Coordinates)
		assert.Equal(t, 42.615208, resp.Data.Signals[0].Coordinates.Latitude)
		assert.Equal(t, -83.029093, resp.Data.Signals[0].Coordinates.Longitude)
	})

	t.Run("missing agg key fails template render before execution", func(t *testing.T) {
		text, isError := callTool(t, server.URL, token, "telemetry_get_signals_time_series", map[string]any{
			"tokenId":  mcpTestTokenID,
			"interval": "1h",
			"from":     baseTime.Format(time.RFC3339),
			"to":       baseTime.Add(2 * time.Hour).Format(time.RFC3339),
			"signalRequests": []map[string]any{
				{"name": "speed"},
			},
		})
		require.True(t, isError)
		assert.Contains(t, text, "failed to render selection template")
		assert.NotContains(t, text, "<no value>")
	})

	t.Run("unauthenticated call is rejected at the resolver", func(t *testing.T) {
		text, isError := callTool(t, server.URL, "", "telemetry_get_latest_signals", map[string]any{
			"tokenId":     mcpTestTokenID,
			"signalNames": []string{"speed"},
		})
		require.False(t, isError, "auth failures surface as GraphQL errors, not tool errors: %s", text)
		assert.Contains(t, text, "unauthorized")
	})

	t.Run("token without privilege cannot read signals", func(t *testing.T) {
		noPrivToken := services.Auth.CreateVehicleToken(t, mcpTestTokenID, []string{})
		text, _ := callTool(t, server.URL, noPrivToken, "telemetry_get_latest_signals", map[string]any{
			"tokenId":     mcpTestTokenID,
			"signalNames": []string{"speed"},
		})
		assert.Contains(t, text, "unauthorized")
	})
}
