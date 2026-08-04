package graph

import (
	"strings"
	"testing"
	"text/template"

	"github.com/DIMO-Network/server-garage/pkg/mcpserver"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

func TestOverrideMCPTools_PatchesBothTools(t *testing.T) {
	out, err := OverrideMCPTools(MCPTools)
	require.NoError(t, err)
	require.Len(t, out, len(MCPTools))

	for _, name := range []string{"telemetry_get_signals_time_series", "telemetry_get_latest_signals"} {
		tool := findTool(t, out, name)
		require.Contains(t, tool.Query, mcpserver.SelectionPlaceholder, "%s query must contain selection placeholder", name)
		require.NotEmpty(t, tool.SelectionTemplate, "%s must have a selection template", name)
		last := tool.Args[len(tool.Args)-1]
		require.True(t, last.ToolOnly, "%s appended arg must be ToolOnly", name)
		require.True(t, last.Required, "%s appended arg must be required", name)
	}
}

func TestOverrideMCPTools_DoesNotMutateInput(t *testing.T) {
	_, err := OverrideMCPTools(MCPTools)
	require.NoError(t, err)

	for _, name := range []string{"telemetry_get_signals_time_series", "telemetry_get_latest_signals"} {
		orig := findTool(t, MCPTools, name)
		require.NotContains(t, orig.Query, mcpserver.SelectionPlaceholder, "generated %s must be untouched", name)
		require.Empty(t, orig.SelectionTemplate)
		for _, a := range orig.Args {
			require.False(t, a.ToolOnly)
		}
	}
}

func TestOverrideMCPTools_ErrorsWhenToolMissing(t *testing.T) {
	var trimmed []mcpserver.ToolDefinition
	for _, tool := range MCPTools {
		if tool.Name == "telemetry_get_signals_time_series" {
			continue
		}
		trimmed = append(trimmed, tool)
	}

	_, err := OverrideMCPTools(trimmed)
	require.Error(t, err, "must fail when an expected tool is missing from generated output")

	_, err = OverrideMCPTools(nil)
	require.Error(t, err)
}

// TestOverrideMCPTools_TemplatesRenderValidGraphQL renders each override's
// selection template with representative arguments, splices it into the query,
// and checks the result parses as GraphQL.
func TestOverrideMCPTools_TemplatesRenderValidGraphQL(t *testing.T) {
	out, err := OverrideMCPTools(MCPTools)
	require.NoError(t, err)

	cases := []struct {
		toolName string
		args     map[string]any
	}{
		{
			toolName: "telemetry_get_signals_time_series",
			args: map[string]any{
				"signalRequests": []any{
					map[string]any{"name": "speed", "agg": "AVG"},
					map[string]any{"name": "powertrainTractionBatteryStateOfChargeCurrent", "agg": "LAST"},
				},
			},
		},
		{
			toolName: "telemetry_get_latest_signals",
			args: map[string]any{
				"signalNames": []any{"speed", "powertrainTractionBatteryStateOfChargeCurrent"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.toolName, func(t *testing.T) {
			tool := findTool(t, out, tc.toolName)

			tmpl, err := template.New(tc.toolName).Parse(tool.SelectionTemplate)
			require.NoError(t, err)
			var buf strings.Builder
			require.NoError(t, tmpl.Execute(&buf, tc.args))

			query := strings.Replace(tool.Query, mcpserver.SelectionPlaceholder, buf.String(), 1)
			_, parseErr := parser.ParseQuery(&ast.Source{Input: query})
			require.Nil(t, parseErr, "rendered query must parse: %s", query)

			require.Contains(t, buf.String(), "speed")
			require.Contains(t, buf.String(), "powertrainTractionBatteryStateOfChargeCurrent")
		})
	}
}

// TestOverrideMCPTools_LocationSignalsValidateAgainstSchema reproduces the
// 2026-08-03 production failures: location signals passed to the time-series
// and latest tools rendered selections without subfields, and the executor
// rejected them with "must have a selection of subfields". Parsing alone
// can't catch that, so rendered queries must validate against the schema.
func TestOverrideMCPTools_LocationSignalsValidateAgainstSchema(t *testing.T) {
	out, err := OverrideMCPTools(MCPTools)
	require.NoError(t, err)

	cases := []struct {
		toolName string
		args     map[string]any
	}{
		{
			toolName: "telemetry_get_signals_time_series",
			args: map[string]any{
				"signalRequests": []any{
					map[string]any{"name": "speed", "agg": "AVG"},
					map[string]any{"name": "currentLocationCoordinates", "agg": "LAST"},
					map[string]any{"name": "currentLocationApproximateCoordinates", "agg": "FIRST"},
				},
			},
		},
		{
			toolName: "telemetry_get_latest_signals",
			args: map[string]any{
				"signalNames": []any{"speed", "currentLocationCoordinates", "currentLocationApproximateCoordinates"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.toolName, func(t *testing.T) {
			tool := findTool(t, out, tc.toolName)

			tmpl, err := template.New(tc.toolName).Option("missingkey=error").Parse(tool.SelectionTemplate)
			require.NoError(t, err)
			var buf strings.Builder
			require.NoError(t, tmpl.Execute(&buf, tc.args))

			query := strings.Replace(tool.Query, mcpserver.SelectionPlaceholder, buf.String(), 1)
			doc, parseErr := parser.ParseQuery(&ast.Source{Input: query})
			require.Nil(t, parseErr, "rendered query must parse: %s", query)

			errs := validator.Validate(parsedSchema, doc)
			require.Empty(t, errs, "rendered query must validate against the schema: %s", query)
		})
	}
}

func findTool(t *testing.T, tools []mcpserver.ToolDefinition, name string) mcpserver.ToolDefinition {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not found", name)
	return mcpserver.ToolDefinition{}
}
