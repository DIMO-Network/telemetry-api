package graph

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
)

// segments and dailyActivity only require VEHICLE_NON_LOCATION_DATA plus one of
// VEHICLE_ALL_TIME_LOCATION or VEHICLE_APPROXIMATE_LOCATION, so the resolvers enforce the rest:
//   - signalRequests may not name a signal the caller cannot read (e.g. currentLocationAltitude),
//   - eventRequests need VEHICLE_ALL_TIME_LOCATION, as the events query does,
//   - without VEHICLE_ALL_TIME_LOCATION, start/end coordinates are replaced by their approximation.

// callerPermissions returns the permissions in the caller's telemetry claim, or nil if there is none.
func callerPermissions(ctx context.Context) []string {
	claim, _ := ctx.Value(auth.TelemetryClaimContextKey{}).(*auth.TelemetryClaim)
	if claim == nil {
		return nil
	}
	return claim.Permissions
}

// checkSegmentSignalRequests rejects a signal request for a known signal the caller lacks privileges for.
// Unknown names are left to the repository, as before: they match no stored signal.
func checkSegmentSignalRequests(ctx context.Context, reqs []*model.SegmentSignalRequest, permissions []string) error {
	for _, req := range reqs {
		if req == nil {
			continue
		}
		required, known := model.SignalPrivileges[req.Name]
		if !known || hasPrivilegesForSignal(req.Name, permissions) {
			continue
		}
		var missing []string
		for _, priv := range required {
			perm, mapped := privilegeEnumToPermission[priv]
			if !mapped {
				missing = append(missing, priv)
			} else if !slices.Contains(permissions, perm) {
				missing = append(missing, perm)
			}
		}
		msg := fmt.Sprintf("unauthorized: missing required privilege(s) %s for signal %s in signalRequests", strings.Join(missing, " "), req.Name)
		return errorhandler.NewUnauthorizedErrorWithMsg(ctx, nil, msg)
	}
	return nil
}

// checkSegmentEventRequests applies the events query's privileges to eventRequests: the events query requires
// VEHICLE_ALL_TIME_LOCATION, so event counts do too.
func checkSegmentEventRequests(ctx context.Context, reqs []*model.SegmentEventRequest, permissions []string) error {
	if len(reqs) == 0 || hasExactLocation(permissions) {
		return nil
	}
	msg := fmt.Sprintf("unauthorized: missing required privilege(s) %s for eventRequests", tokenclaims.PermissionGetLocationHistory)
	return errorhandler.NewUnauthorizedErrorWithMsg(ctx, nil, msg)
}

// hasExactLocation reports whether the caller may see exact coordinates.
func hasExactLocation(permissions []string) bool {
	return slices.Contains(permissions, tokenclaims.PermissionGetLocationHistory)
}

// approximateLocation returns loc snapped to the center of its H3 cell, as for currentLocationApproximateCoordinates.
// It never returns the exact value: the repository's (0, 0) "no location data" placeholder stays (0, 0), and so does
// a location the H3 library cannot convert.
func approximateLocation(loc *model.Location) *model.Location {
	if loc == nil {
		return nil
	}
	if loc.Latitude == 0 && loc.Longitude == 0 {
		return &model.Location{}
	}
	latLng := repositories.GetApproximateLoc(loc.Latitude, loc.Longitude)
	if latLng == nil {
		return &model.Location{}
	}
	return &model.Location{Latitude: latLng.Lat, Longitude: latLng.Lng, Hdop: loc.Hdop}
}

func approximateSignalLocation(sl *model.SignalLocation) {
	if sl != nil {
		sl.Value = approximateLocation(sl.Value)
	}
}

func approximateSegmentLocations(segments []*model.Segment) {
	for _, seg := range segments {
		if seg != nil {
			approximateSignalLocation(seg.Start)
			approximateSignalLocation(seg.End)
		}
	}
}

func approximateDailyActivityLocations(days []*model.DailyActivity) {
	for _, day := range days {
		if day != nil {
			approximateSignalLocation(day.Start)
			approximateSignalLocation(day.End)
		}
	}
}
