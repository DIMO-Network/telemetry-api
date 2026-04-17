package graph

import (
	"slices"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
)

// privilegeEnumToPermission maps GraphQL Privilege enum values to tokenclaims permission strings.
var privilegeEnumToPermission = map[string]string{
	"VEHICLE_NON_LOCATION_DATA":    tokenclaims.PermissionGetNonLocationHistory,
	"VEHICLE_ALL_TIME_LOCATION":    tokenclaims.PermissionGetLocationHistory,
	"VEHICLE_APPROXIMATE_LOCATION": tokenclaims.PermissionGetApproximateLocation,
}

// hasPrivilegesForSignal checks if the caller has all required privileges for a signal.
func hasPrivilegesForSignal(signalName string, permissions []string) bool {
	required, ok := model.SignalPrivileges[signalName]
	if !ok {
		// Approximate location is a derived signal not in the generated map.
		// Require at least one of approximate or all-time location privileges.
		if signalName == model.ApproximateCoordinatesField {
			return slices.Contains(permissions, tokenclaims.PermissionGetApproximateLocation) ||
				slices.Contains(permissions, tokenclaims.PermissionGetLocationHistory)
		}
		return false
	}
	for _, priv := range required {
		perm, mapped := privilegeEnumToPermission[priv]
		if !mapped {
			return false
		}
		if !slices.Contains(permissions, perm) {
			return false
		}
	}
	return true
}
