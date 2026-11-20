package graph

import (
	"context"
	"testing"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	"github.com/stretchr/testify/require"
)

var approxOnly = []string{tokenclaims.PermissionGetNonLocationHistory, tokenclaims.PermissionGetApproximateLocation}

func TestCallerPermissions(t *testing.T) {
	require.Nil(t, callerPermissions(context.Background()))

	claim := &auth.TelemetryClaim{}
	claim.Permissions = approxOnly
	ctx := context.WithValue(context.Background(), auth.TelemetryClaimContextKey{}, claim)
	require.Equal(t, approxOnly, callerPermissions(ctx))
}

func TestCheckSegmentSignalRequests(t *testing.T) {
	ctx := context.Background()
	speed := &model.SegmentSignalRequest{Name: "speed", Agg: model.FloatAggregationMax}
	altitude := &model.SegmentSignalRequest{Name: "currentLocationAltitude", Agg: model.FloatAggregationAvg}
	unknown := &model.SegmentSignalRequest{Name: "notASignal", Agg: model.FloatAggregationAvg}

	require.NoError(t, checkSegmentSignalRequests(ctx, nil, approxOnly))
	require.NoError(t, checkSegmentSignalRequests(ctx, []*model.SegmentSignalRequest{speed, unknown, nil}, approxOnly))

	err := checkSegmentSignalRequests(ctx, []*model.SegmentSignalRequest{speed, altitude}, approxOnly)
	require.EqualError(t, err, "input: unauthorized: missing required privilege(s) privilege:GetLocationHistory for signal currentLocationAltitude in signalRequests")

	full := append([]string{tokenclaims.PermissionGetLocationHistory}, approxOnly...)
	require.NoError(t, checkSegmentSignalRequests(ctx, []*model.SegmentSignalRequest{speed, altitude}, full))
}

func TestCheckSegmentEventRequests(t *testing.T) {
	ctx := context.Background()
	reqs := []*model.SegmentEventRequest{{Name: "behavior.harshBraking"}}

	require.NoError(t, checkSegmentEventRequests(ctx, nil, approxOnly))
	require.EqualError(t, checkSegmentEventRequests(ctx, reqs, approxOnly),
		"input: unauthorized: missing required privilege(s) privilege:GetLocationHistory for eventRequests")
	require.NoError(t, checkSegmentEventRequests(ctx, reqs, append([]string{tokenclaims.PermissionGetLocationHistory}, approxOnly...)))
}

func TestHasExactLocation(t *testing.T) {
	require.False(t, hasExactLocation(nil))
	require.False(t, hasExactLocation(approxOnly))
	require.True(t, hasExactLocation([]string{tokenclaims.PermissionGetLocationHistory}))
}

func TestApproximateLocation(t *testing.T) {
	require.Nil(t, approximateLocation(nil))

	// The repository's "no location data" placeholder stays a placeholder.
	require.Equal(t, &model.Location{}, approximateLocation(&model.Location{Hdop: 1}))

	exact := &model.Location{Latitude: 42.331429, Longitude: -83.045753, Hdop: 1.5}
	want := repositories.GetApproximateLoc(exact.Latitude, exact.Longitude)
	require.NotNil(t, want)

	got := approximateLocation(exact)
	require.Equal(t, &model.Location{Latitude: want.Lat, Longitude: want.Lng, Hdop: 1.5}, got)
	require.NotEqual(t, exact.Latitude, got.Latitude)
	require.NotEqual(t, exact.Longitude, got.Longitude)

	// Nearby points in the same cell are indistinguishable.
	near := approximateLocation(&model.Location{Latitude: exact.Latitude + 0.0005, Longitude: exact.Longitude + 0.0005})
	require.Equal(t, got.Latitude, near.Latitude)
	require.Equal(t, got.Longitude, near.Longitude)
}

func TestApproximateSegmentAndDailyActivityLocations(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	exact := func() *model.Location { return &model.Location{Latitude: 42.331429, Longitude: -83.045753} }
	want := approximateLocation(exact())

	segments := []*model.Segment{
		{Start: &model.SignalLocation{Timestamp: ts, Value: exact()}, End: &model.SignalLocation{Timestamp: ts.Add(time.Hour), Value: exact()}},
		{Start: &model.SignalLocation{Timestamp: ts, Value: exact()}, IsOngoing: true}, // ongoing: no end
		nil,
	}
	approximateSegmentLocations(segments)
	require.Equal(t, want, segments[0].Start.Value)
	require.Equal(t, want, segments[0].End.Value)
	require.Equal(t, ts, segments[0].Start.Timestamp, "timestamps are kept")
	require.Equal(t, ts.Add(time.Hour), segments[0].End.Timestamp)
	require.Equal(t, want, segments[1].Start.Value)
	require.Nil(t, segments[1].End)

	days := []*model.DailyActivity{
		{Start: &model.SignalLocation{Timestamp: ts, Value: exact()}, End: &model.SignalLocation{Timestamp: ts, Value: exact()}},
		{SegmentCount: 0}, // no start/end
		nil,
	}
	approximateDailyActivityLocations(days)
	require.Equal(t, want, days[0].Start.Value)
	require.Equal(t, want, days[0].End.Value)
	require.Nil(t, days[1].Start)
}
