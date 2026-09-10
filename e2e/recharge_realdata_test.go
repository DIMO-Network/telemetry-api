package e2e_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/stretchr/testify/require"
)

// TestRechargeRealDataEndToEnd loads a local export of a vehicle's SoC and odometer series into the
// ClickHouse container and runs the full HTTP stack (JWT auth, resolvers, repository summary signals,
// MCP tool) over it. Skips unless RECHARGE_REALDATA_JSON points at an export; the file is never committed.
// Same format as internal/service/ch/recharge_realdata_test.go, plus a "tokenId" field.
func TestRechargeRealDataEndToEnd(t *testing.T) {
	path := os.Getenv("RECHARGE_REALDATA_JSON")
	if path == "" {
		t.Skip("RECHARGE_REALDATA_JSON not set")
	}
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var export struct {
		TokenID          int       `json:"tokenId"`
		From             time.Time `json:"from"`
		To               time.Time `json:"to"`
		Soc, Odo         []struct {
			Ts    time.Time `json:"ts"`
			Value float64   `json:"value"`
		}
		ExpectedStarts   []time.Time `json:"expectedStarts"`
		UnexpectedStarts []time.Time `json:"unexpectedStarts"`
		LaterWindowFrom  *time.Time  `json:"laterWindowFrom"`
		ExpectedSoc      []struct {
			Start time.Time `json:"start"`
			From  float64   `json:"from"`
			To    float64   `json:"to"`
		} `json:"expectedSoc"`
	}
	require.NoError(t, json.Unmarshal(raw, &export))
	require.NotZero(t, export.TokenID)

	services := GetTestServices(t)
	subject := fmt.Sprintf("did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:%d", export.TokenID)
	var signals []vss.Signal
	for name, series := range map[string][]struct {
		Ts    time.Time `json:"ts"`
		Value float64   `json:"value"`
	}{vss.FieldPowertrainTractionBatteryStateOfChargeCurrent: export.Soc, vss.FieldPowertrainTransmissionTravelledDistance: export.Odo} {
		for _, s := range series {
			signals = append(signals, vss.Signal{
				CloudEventHeader: cloudevent.CloudEventHeader{Source: "0x0000000000000000000000000000000000000001", Subject: subject},
				Data:             vss.SignalData{Timestamp: s.Ts, Name: name, ValueNumber: s.Value},
			})
		}
	}
	insertSignal(t, services.CH, signals)
	t.Logf("inserted %d signal rows for token %d", len(signals), export.TokenID)

	client := NewGraphQLServer(t, services.Settings)
	token := services.Auth.CreateVehicleToken(t, export.TokenID, []string{tokenclaims.PermissionGetNonLocationHistory, tokenclaims.PermissionGetLocationHistory})

	type segment struct {
		Start              struct{ Timestamp string }
		End                *struct{ Timestamp string }
		Duration           int
		IsOngoing          bool
		StartedBeforeRange bool
		MaxSampleGapSeconds *int
		Signals            []struct {
			Name  string
			Agg   string
			Value float64
		}
	}
	const selection = `start { timestamp } end { timestamp } duration isOngoing startedBeforeRange maxSampleGapSeconds signals { name agg value }`
	query := func(t *testing.T, mechanism string, from, to time.Time, config string) []segment {
		t.Helper()
		var res struct{ Segments []segment }
		if !strings.Contains(config, "limit:") {
			config += ", limit: 200"
		}
		q := fmt.Sprintf(`query { segments(tokenId: %d, from: %q, to: %q, mechanism: %s%s) { %s } }`,
			export.TokenID, from.Format(time.RFC3339), to.Format(time.RFC3339), mechanism, config, selection)
		require.NoError(t, client.Post(q, &res, WithToken(token)))
		return res.Segments
	}
	socAgg := func(s segment, agg string) float64 {
		for _, sig := range s.Signals {
			if sig.Name == vss.FieldPowertrainTractionBatteryStateOfChargeCurrent && sig.Agg == agg {
				return sig.Value
			}
		}
		return -1
	}
	ts := func(s string) time.Time {
		t.Helper()
		parsed, err := time.Parse(time.RFC3339Nano, s)
		require.NoError(t, err)
		return parsed
	}
	near := func(a, b time.Time) bool { d := a.Sub(b); return d > -15*time.Minute && d < 15*time.Minute }
	findStart := func(segs []segment, want time.Time) (segment, bool) {
		for _, s := range segs {
			if near(ts(s.Start.Timestamp), want) {
				return s, true
			}
		}
		return segment{}, false
	}

	// Segments are only detected up to now; the export window may extend past that.
	to := export.To
	if now := time.Now().UTC(); to.After(now) {
		to = now
	}

	segs := query(t, "recharge", export.From, to, "")
	t.Logf("recharge over full window: %d segments", len(segs))
	for _, s := range segs {
		gap := -1
		if s.MaxSampleGapSeconds != nil {
			gap = *s.MaxSampleGapSeconds
		}
		t.Logf("  %s -> %s dur=%6d gap=%6d soc %.0f->%.0f", ts(s.Start.Timestamp).Format("01-02 15:04"), ts(s.End.Timestamp).Format("01-02 15:04"), s.Duration, gap, socAgg(s, "FIRST"), socAgg(s, "LAST"))
	}

	t.Run("every recharge segment carries maxSampleGapSeconds within duration and a positive SoC delta", func(t *testing.T) {
		require.NotEmpty(t, segs)
		for _, s := range segs {
			require.NotNil(t, s.MaxSampleGapSeconds, "segment at %s", s.Start.Timestamp)
			require.LessOrEqual(t, *s.MaxSampleGapSeconds, s.Duration)
			require.False(t, s.IsOngoing)
			require.NotNil(t, s.End)
			require.Equal(t, int(ts(s.End.Timestamp).Sub(ts(s.Start.Timestamp)).Seconds()), s.Duration)
			require.Greater(t, socAgg(s, "LAST"), socAgg(s, "FIRST"), "segment at %s", s.Start.Timestamp)
		}
	})

	t.Run("expected sessions are returned with the expected summary SoC", func(t *testing.T) {
		for _, want := range export.ExpectedStarts {
			_, ok := findStart(segs, want)
			require.Truef(t, ok, "expected a segment starting near %s", want)
		}
		for _, unwanted := range export.UnexpectedStarts {
			_, ok := findStart(segs, unwanted)
			require.Falsef(t, ok, "expected no segment starting near %s", unwanted)
		}
		for _, e := range export.ExpectedSoc {
			s, ok := findStart(segs, e.Start)
			require.True(t, ok)
			require.InDelta(t, e.From, socAgg(s, "FIRST"), 0.5, "FIRST SoC at %s", e.Start)
			require.InDelta(t, e.To, socAgg(s, "LAST"), 0.5, "LAST SoC at %s", e.Start)
		}
	})

	t.Run("a narrower window still yields sessions", func(t *testing.T) {
		if export.LaterWindowFrom == nil {
			t.Skip("laterWindowFrom not set")
		}
		later := query(t, "recharge", *export.LaterWindowFrom, to, "")
		t.Logf("recharge from %s: %d segments", export.LaterWindowFrom.Format(time.RFC3339), len(later))
		require.NotEmpty(t, later)
		for _, s := range later {
			require.False(t, ts(s.Start.Timestamp).Before(*export.LaterWindowFrom))
		}
	})

	t.Run("minIncreasePercent override filters small sessions", func(t *testing.T) {
		big := query(t, "recharge", export.From, to, ", config: {minIncreasePercent: 30}")
		t.Logf("recharge with minIncreasePercent 30: %d segments", len(big))
		require.Less(t, len(big), len(segs))
		for _, s := range big {
			require.GreaterOrEqual(t, socAgg(s, "LAST")-socAgg(s, "FIRST"), 30.0, "segment at %s", s.Start.Timestamp)
		}
	})

	t.Run("pagination with after and limit", func(t *testing.T) {
		first := query(t, "recharge", export.From, to, ", limit: 3")
		require.Len(t, first, 3)
		rest := query(t, "recharge", export.From, to, fmt.Sprintf(", after: %q", first[2].Start.Timestamp))
		require.Len(t, rest, len(segs)-3)
		require.True(t, ts(rest[0].Start.Timestamp).After(ts(first[2].Start.Timestamp)))
	})

	t.Run("maxSampleGapSeconds is null for other mechanisms", func(t *testing.T) {
		other := query(t, "frequencyAnalysis", export.From, to, ", config: {signalCountThreshold: 1}")
		t.Logf("frequencyAnalysis: %d segments", len(other))
		for _, s := range other {
			require.Nil(t, s.MaxSampleGapSeconds)
		}
		require.Empty(t, query(t, "refuel", export.From, to, ""))
	})

	t.Run("MCP get_trip_segments returns the recharge sessions", func(t *testing.T) {
		mcp := newMCPServer(t, services.Settings)
		text, isErr := callTool(t, mcp.URL, token, "telemetry_get_trip_segments", map[string]any{
			"tokenId": export.TokenID, "from": export.From.Format(time.RFC3339), "to": to.Format(time.RFC3339), "mechanism": "recharge",
		})
		require.False(t, isErr, text)
		require.Equal(t, len(segs), strings.Count(text, `"duration"`), "MCP should list the same sessions as GraphQL")
	})
}
