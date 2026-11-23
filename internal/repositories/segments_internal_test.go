package repositories

import (
	"testing"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/require"
)

func TestBuildAggArgsPreferFix(t *testing.T) {
	_, loc := buildAggArgs(nil)
	require.Len(t, loc, 2)
	require.Equal(t, model.LocationSignalArgs{
		Name:      vss.FieldCurrentLocationCoordinates,
		Agg:       model.LocationAggregationFirst,
		Alias:     "startLoc",
		PreferFix: true,
	}, loc[0])
	require.Equal(t, model.LocationSignalArgs{
		Name:      vss.FieldCurrentLocationCoordinates,
		Agg:       model.LocationAggregationLast,
		Alias:     "endLoc",
		PreferFix: true,
	}, loc[1])
}
