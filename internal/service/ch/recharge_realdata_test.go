package ch

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// rechargeRealDataEnv points at a local JSON export of a vehicle's SoC and odometer series.
// The file is never committed; the test skips when the variable is unset.
//
// Format:
//
//	{
//	  "from": "...", "to": "...",
//	  "soc": [{"ts": "RFC3339", "value": 44}, ...],
//	  "odo": [{"ts": "RFC3339", "value": 12345.6}, ...],
//	  "expectedStarts":   ["RFC3339", ...],   // a session must start within expectTolerance of each
//	  "unexpectedStarts": ["RFC3339", ...],   // no session may start within expectTolerance of any
//	  "laterWindowFrom": "RFC3339",           // optional: detection over [this, end) must not be empty
//	  "expectedSoc": [{"start": "RFC3339", "from": 23, "to": 56}, ...]
//	}
const rechargeRealDataEnv = "RECHARGE_REALDATA_JSON"

// Starts may shift by the flat lead-in (SoC sitting at its start value before the rise), so allow some slack.
const rechargeRealDataTolerance = 15 * time.Minute

type realDataSample struct {
	Ts    time.Time `json:"ts"`
	Value float64   `json:"value"`
}

type realDataExport struct {
	From             time.Time        `json:"from"`
	To               time.Time        `json:"to"`
	Soc              []realDataSample `json:"soc"`
	Odo              []realDataSample `json:"odo"`
	ExpectedStarts   []time.Time      `json:"expectedStarts"`
	UnexpectedStarts []time.Time      `json:"unexpectedStarts"`
	LaterWindowFrom  *time.Time       `json:"laterWindowFrom"` // optional: a narrower window that must still yield sessions
	ExpectedSoc      []struct {
		Start time.Time `json:"start"`
		From  float64   `json:"from"`
		To    float64   `json:"to"`
	} `json:"expectedSoc"`
}

func toLevelSamples(in []realDataSample) []levelSample {
	out := make([]levelSample, 0, len(in))
	for _, s := range in {
		out = append(out, levelSample{ts: s.Ts, value: s.Value})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ts.Before(out[j].ts) })
	return out
}

func TestRechargeRealData(t *testing.T) {
	path := os.Getenv(rechargeRealDataEnv)
	if path == "" {
		t.Skipf("%s not set", rechargeRealDataEnv)
	}
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var export realDataExport
	require.NoError(t, json.Unmarshal(raw, &export))

	soc := toLevelSamples(export.Soc)
	odo := toLevelSamples(export.Odo)
	sessions := detectRechargeSessions(soc, odo, rechargeDefaultMinDurationSeconds, rechargeMinRisePct)

	t.Logf("%d sessions over %s -> %s", len(sessions), export.From.Format(time.RFC3339), export.To.Format(time.RFC3339))
	for _, s := range sessions {
		startSoc, endSoc, _ := levelFirstLastInRange(soc, s.start, s.end)
		t.Logf("  %s -> %s  dur=%6.0f min  gap=%6.0f min  soc %.0f->%.0f",
			s.start.Format("01-02 15:04"), s.end.Format("01-02 15:04"),
			s.end.Sub(s.start).Minutes(), float64(s.maxSampleGapSeconds)/60, startSoc, endSoc)
	}

	findStart := func(want time.Time) (rechargeSession, bool) {
		for _, s := range sessions {
			if d := s.start.Sub(want); d > -rechargeRealDataTolerance && d < rechargeRealDataTolerance {
				return s, true
			}
		}
		return rechargeSession{}, false
	}
	for _, want := range export.ExpectedStarts {
		_, ok := findStart(want)
		require.Truef(t, ok, "expected a session starting near %s", want.Format(time.RFC3339))
	}
	for _, unwanted := range export.UnexpectedStarts {
		_, ok := findStart(unwanted)
		require.Falsef(t, ok, "expected no session starting near %s", unwanted.Format(time.RFC3339))
	}
	for _, e := range export.ExpectedSoc {
		s, ok := findStart(e.Start)
		require.Truef(t, ok, "expected a session starting near %s", e.Start.Format(time.RFC3339))
		startSoc, endSoc, ok := levelFirstLastInRange(soc, s.start, s.end)
		require.True(t, ok)
		require.InDeltaf(t, e.From, startSoc, 0.5, "start SoC for session at %s", e.Start.Format(time.RFC3339))
		require.InDeltaf(t, e.To, endSoc, 0.5, "end SoC for session at %s", e.Start.Format(time.RFC3339))
	}

	// A window that starts after the first sessions must still see the later ones.
	if len(export.ExpectedStarts) > 0 {
		last := export.ExpectedStarts[0]
		for _, s := range export.ExpectedStarts {
			if s.After(last) {
				last = s
			}
		}
		cut := last.Add(-24 * time.Hour)
		if export.LaterWindowFrom != nil {
			cut = *export.LaterWindowFrom
		}
		cutSoc := soc[sort.Search(len(soc), func(i int) bool { return !soc[i].ts.Before(cut) }):]
		cutOdo := odo[sort.Search(len(odo), func(i int) bool { return !odo[i].ts.Before(cut) }):]
		require.NotEmpty(t, detectRechargeSessions(cutSoc, cutOdo, rechargeDefaultMinDurationSeconds, rechargeMinRisePct),
			"window from %s must not be empty", cut.Format(time.RFC3339))
	}
}
