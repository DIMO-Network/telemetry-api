package proxy

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSubject(t *testing.T) {
	got := ToSubject(137, common.HexToAddress("0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF"), 42)
	assert.Equal(t, "did:erc721:137:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF:42", got)
}

func TestToDQMechanism(t *testing.T) {
	cases := []struct {
		in   model.DetectionMechanism
		want string
	}{
		{model.DetectionMechanismIgnitionDetection, "IGNITION_DETECTION"},
		{model.DetectionMechanismFrequencyAnalysis, "FREQUENCY_ANALYSIS"},
		{model.DetectionMechanismChangePointDetection, "CHANGE_POINT_DETECTION"},
		{model.DetectionMechanismIdling, "IDLING"},
		{model.DetectionMechanismRefuel, "REFUEL"},
		{model.DetectionMechanismRecharge, "RECHARGE"},
	}
	for _, tc := range cases {
		got, err := ToDQMechanism(tc.in)
		require.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}

	_, err := ToDQMechanism("bogus")
	require.Error(t, err)
}

func TestBuildSignalsQuery_NoFilters(t *testing.T) {
	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{
			{Name: "speed", Agg: model.FloatAggregationLast, Alias: "speed"},
		},
		StringArgs: []model.StringSignalArgs{
			{Name: "obdFuelTypeName", Agg: model.StringAggregationLast, Alias: "obdFuelTypeName"},
		},
		LocationArgs: []model.LocationSignalArgs{
			{Name: "currentLocationCoordinates", Agg: model.LocationAggregationLast, Alias: "currentLocationCoordinates"},
		},
	}

	query, filterVars := BuildSignalsQuery(aggArgs)

	assert.Empty(t, filterVars)
	assert.Contains(t, query, `speed: speed(agg: LAST)`)
	assert.Contains(t, query, `obdFuelTypeName: obdFuelTypeName(agg: LAST)`)
	assert.Contains(t, query, `currentLocationCoordinates: currentLocationCoordinates(agg: LAST) { latitude longitude hdop }`)
	assert.Contains(t, query, `timestamp`)
	// No extra variable declarations beyond the base ones.
	assert.NotContains(t, query, `SignalFloatFilter`)
	assert.NotContains(t, query, `SignalLocationFilter`)
}

func TestBuildSignalsQuery_AliasPreserved(t *testing.T) {
	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{
			{Name: "speed", Agg: model.FloatAggregationLast, Alias: "mySpeed"},
		},
	}

	query, _ := BuildSignalsQuery(aggArgs)

	assert.Contains(t, query, `mySpeed: speed(agg: LAST)`)
	assert.NotContains(t, query, `speed: speed`)
}

func TestBuildSignalsQuery_WithFloatFilter(t *testing.T) {
	gt := 50.0
	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{
			{Name: "speed", Agg: model.FloatAggregationAvg, Alias: "speed", Filter: &model.SignalFloatFilter{Gt: &gt}},
			{Name: "rpm", Agg: model.FloatAggregationAvg, Alias: "rpm"},
		},
	}

	query, filterVars := BuildSignalsQuery(aggArgs)

	require.Len(t, filterVars, 1)
	assert.Contains(t, filterVars, "ff0")
	assert.Contains(t, query, `$ff0: SignalFloatFilter`)
	assert.Contains(t, query, `speed: speed(agg: AVG, filter: $ff0)`)
	assert.Contains(t, query, `rpm: rpm(agg: AVG)`)
	assert.NotContains(t, query, `$ff1`) // second arg has no filter
}

func TestBuildSignalsQuery_ApproxCoordinatesAlias(t *testing.T) {
	// Approximate location: Name = raw coords field, Alias = approximate field name.
	aggArgs := &model.AggregatedSignalArgs{
		LocationArgs: []model.LocationSignalArgs{
			{
				Name:  vss.FieldCurrentLocationCoordinates,
				Agg:   model.LocationAggregationLast,
				Alias: model.ApproximateCoordinatesField,
			},
		},
	}

	query, _ := BuildSignalsQuery(aggArgs)

	// Alias is the approximate field, Name is the raw coordinates field.
	assert.Contains(t, query, model.ApproximateCoordinatesField+`: `+vss.FieldCurrentLocationCoordinates+`(agg: LAST) { latitude longitude hdop }`)
}

func TestUnmarshalSignalsResponse_Basic(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"signals":[{"timestamp":"2024-01-01T00:00:00Z","speed":55.5,"gear":"D","currentLocationCoordinates":{"latitude":37.0,"longitude":-122.0,"hdop":1.5}}]}`)

	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs:  []model.FloatSignalArgs{{Name: "speed", Alias: "speed"}},
		StringArgs: []model.StringSignalArgs{{Name: "gear", Alias: "gear"}},
		LocationArgs: []model.LocationSignalArgs{
			{Name: "currentLocationCoordinates", Alias: "currentLocationCoordinates"},
		},
	}

	result, err := UnmarshalSignalsResponse(data, aggArgs)
	require.NoError(t, err)
	require.Len(t, result, 1)

	row := result[0]
	assert.Equal(t, ts, row.Timestamp)
	assert.Equal(t, 55.5, row.ValueNumbers["speed"])
	assert.Equal(t, "D", row.ValueStrings["gear"])
	loc := row.ValueLocations["currentLocationCoordinates"]
	assert.InDelta(t, 37.0, loc.Latitude, 1e-9)
	assert.InDelta(t, -122.0, loc.Longitude, 1e-9)
	assert.InDelta(t, 1.5, loc.HDOP, 1e-9)
}

func TestUnmarshalSignalsResponse_AliasMapping(t *testing.T) {
	data := json.RawMessage(`{"signals":[{"timestamp":"2024-01-01T00:00:00Z","mySpeed":88.0}]}`)

	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{{Name: "speed", Alias: "mySpeed"}},
	}

	result, err := UnmarshalSignalsResponse(data, aggArgs)
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, 88.0, result[0].ValueNumbers["mySpeed"])
	assert.Empty(t, result[0].ValueNumbers["speed"]) // original name not used as key
}

func TestUnmarshalSignalsResponse_NullFieldSkipped(t *testing.T) {
	data := json.RawMessage(`{"signals":[{"timestamp":"2024-01-01T00:00:00Z","speed":null}]}`)

	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{{Name: "speed", Alias: "speed"}},
	}

	result, err := UnmarshalSignalsResponse(data, aggArgs)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Empty(t, result[0].ValueNumbers) // null → not added to map
}

func TestUnmarshalSignalsResponse_MultipleRows(t *testing.T) {
	data := json.RawMessage(`{"signals":[
		{"timestamp":"2024-01-01T00:00:00Z","speed":10.0},
		{"timestamp":"2024-01-01T01:00:00Z","speed":20.0}
	]}`)

	aggArgs := &model.AggregatedSignalArgs{
		FloatArgs: []model.FloatSignalArgs{{Name: "speed", Alias: "speed"}},
	}

	result, err := UnmarshalSignalsResponse(data, aggArgs)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, 10.0, result[0].ValueNumbers["speed"])
	assert.Equal(t, 20.0, result[1].ValueNumbers["speed"])
}

func TestBuildSignalsLatestQuery_Fields(t *testing.T) {
	latestArgs := &model.LatestSignalsArgs{
		IncludeLastSeen: true,
		SignalNames:     map[string]struct{}{"speed": {}, "obdFuelTypeName": {}},
		LocationSignalNames: map[string]struct{}{
			vss.FieldCurrentLocationCoordinates: {},
		},
	}

	query := BuildSignalsLatestQuery(latestArgs)

	assert.Contains(t, query, `lastSeen`)
	assert.Contains(t, query, `speed { timestamp value }`)
	assert.Contains(t, query, `obdFuelTypeName { timestamp value }`)
	assert.Contains(t, query, vss.FieldCurrentLocationCoordinates+` { timestamp value { latitude longitude hdop } }`)
}

func TestBuildSignalsLatestQuery_NoLastSeen(t *testing.T) {
	latestArgs := &model.LatestSignalsArgs{
		IncludeLastSeen: false,
		SignalNames:     map[string]struct{}{"speed": {}},
	}

	query := BuildSignalsLatestQuery(latestArgs)
	assert.NotContains(t, query, `lastSeen`)
}

func TestUnmarshalSignalsLatestResponse(t *testing.T) {
	ts := "2024-06-01T12:00:00Z"
	data := json.RawMessage(`{"signalsLatest":{"speed":{"timestamp":"` + ts + `","value":60.0}}}`)

	coll, err := UnmarshalSignalsLatestResponse(data)
	require.NoError(t, err)
	require.NotNil(t, coll)
	require.NotNil(t, coll.Speed)
	assert.Equal(t, 60.0, coll.Speed.Value)
}

func TestUnmarshalSignalsLatestResponse_Nil(t *testing.T) {
	data := json.RawMessage(`{"signalsLatest":null}`)
	coll, err := UnmarshalSignalsLatestResponse(data)
	require.NoError(t, err)
	assert.Nil(t, coll)
}

func TestBuildSignalsQuery_ContainsRequiredVars(t *testing.T) {
	query, _ := BuildSignalsQuery(&model.AggregatedSignalArgs{})
	for _, v := range []string{"$subject", "$interval", "$from", "$to", "$filter"} {
		assert.True(t, strings.Contains(query, v), "missing variable %s", v)
	}
}
