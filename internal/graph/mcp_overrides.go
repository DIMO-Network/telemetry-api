package graph

import (
	"fmt"

	"github.com/DIMO-Network/server-garage/pkg/mcpserver"
)

// OverrideMCPTools patches the mcpgen-generated MCPTools slice so the two
// shortcut tools that wrap `signals` and `signalsLatest` actually work.
//
// Neither GraphQL field declares per-signal arguments — callers name signals
// by field-selection (e.g. `speed(agg: AVG)`). mcpgen, which builds a single
// static query string per tool, can't express that shape. The two tools
// therefore shipped with a selection of only `timestamp` / `lastSeen` and
// returned empty data.
//
// Rather than change the GraphQL schema, this override uses the templated
// selection + tool-only-argument features from server-garage (v0.3.0) to
// reshape the two ToolDefinitions at startup:
//
//   - SelectionTemplate renders the concrete selection set at call time
//     from the supplied signal list.
//   - A ToolOnly ArgDefinition lets the model pass that list in, without
//     the argument flowing through to the GraphQL executor.
//
// The GraphQL schema is untouched.
//
// Returns an error if either tool is missing from the generated slice, so a
// rename in mcpgen output fails startup instead of silently shipping the
// original empty-selection tools.
func OverrideMCPTools(tools []mcpserver.ToolDefinition) ([]mcpserver.ToolDefinition, error) {
	out := make([]mcpserver.ToolDefinition, len(tools))
	copy(out, tools)
	overridden := make(map[string]bool, 2)
	for i := range out {
		switch out[i].Name {
		case "telemetry_get_signals_time_series":
			overrideSignalsTimeSeries(&out[i])
			overridden[out[i].Name] = true
		case "telemetry_get_latest_signals":
			overrideLatestSignals(&out[i])
			overridden[out[i].Name] = true
		}
	}
	for _, name := range []string{"telemetry_get_signals_time_series", "telemetry_get_latest_signals"} {
		if !overridden[name] {
			return nil, fmt.Errorf("graph: OverrideMCPTools did not find tool %q in generated MCPTools; mcpgen output may have renamed it", name)
		}
	}
	return out, nil
}

func overrideSignalsTimeSeries(t *mcpserver.ToolDefinition) {
	t.Description = "Get aggregated time series for a named list of float signals. Pass signalRequests as [{name, agg}] (e.g. [{name:\"speed\",agg:\"AVG\"},{name:\"powertrainTractionBatteryStateOfChargeCurrent\",agg:\"LAST\"}]). Returns buckets of {timestamp, <signal>: <value>, ...}. Signal names come from get_available_signals or get_data_summary. Aggregations: AVG, MED, MAX, MIN, RAND, FIRST, LAST."
	t.Query = `query($tokenId: Int!, $interval: String!, $from: Time!, $to: Time!, $filter: SignalFilter) { signals(tokenId: $tokenId, interval: $interval, from: $from, to: $to, filter: $filter) { __MCPGEN_SELECTION__ } }`
	t.SelectionTemplate = "timestamp{{range .signalRequests}} {{.name}}(agg: {{.agg}}){{end}}"
	t.Args = append(t.Args, mcpserver.ArgDefinition{
		Name:        "signalRequests",
		Type:        "array",
		ItemsType:   "object",
		Required:    true,
		ToolOnly:    true,
		Description: "List of {name, agg} pairs specifying which float signals to aggregate. Each `name` is a signal field name; each `agg` is one of AVG, MED, MAX, MIN, RAND, FIRST, LAST.",
	})
}

func overrideLatestSignals(t *mcpserver.ToolDefinition) {
	t.Description = "Get the most recent value for a named list of float signals. Pass signalNames as an array of strings (e.g. [\"speed\",\"powertrainTractionBatteryStateOfChargeCurrent\"]). For non-float signals (strings, locations) use get_signals_snapshot. Signal names come from get_available_signals or get_data_summary."
	t.Query = `query($tokenId: Int!, $filter: SignalFilter) { signalsLatest(tokenId: $tokenId, filter: $filter) { __MCPGEN_SELECTION__ } }`
	t.SelectionTemplate = "lastSeen{{range .signalNames}} {{.}} {timestamp value}{{end}}"
	t.Args = append(t.Args, mcpserver.ArgDefinition{
		Name:        "signalNames",
		Type:        "array",
		ItemsType:   "string",
		Required:    true,
		ToolOnly:    true,
		Description: "List of float-signal field names to return the latest value for.",
	})
}
