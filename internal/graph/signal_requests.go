package graph

import (
	"context"

	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

// reqAliasPrefix namespaces aliases derived from the signalRequests argument so
// they cannot collide with aliases built from the GraphQL selection set by
// aggregationArgsFromContext.
const reqAliasPrefix = "__req::"

func reqAlias(req *model.SignalAggregationRequest) string {
	return reqAliasPrefix + req.Name + "::" + string(req.Agg)
}

// appendSignalRequestArgs appends a FloatSignalArgs entry for every privileged
// signalRequest, returning the filtered requests and their aliases in input
// order so the resolver can map query output back to the correct entry.
func appendSignalRequestArgs(aggArgs *model.AggregatedSignalArgs, requests []*model.SignalAggregationRequest, permissions []string) ([]*model.SignalAggregationRequest, []string) {
	allowed := make([]*model.SignalAggregationRequest, 0, len(requests))
	aliases := make([]string, 0, len(requests))
	for _, req := range requests {
		if req == nil {
			continue
		}
		if !hasPrivilegesForSignal(req.Name, permissions) {
			continue
		}
		alias := reqAlias(req)
		aggArgs.FloatArgs = append(aggArgs.FloatArgs, model.FloatSignalArgs{
			Name:  req.Name,
			Agg:   req.Agg,
			Alias: alias,
		})
		allowed = append(allowed, req)
		aliases = append(aliases, alias)
	}
	return allowed, aliases
}

// populateAggregationSignals fills each bucket's Signals slice from the aliased
// FloatArgs entries that came from signalRequests. Request order is preserved.
func populateAggregationSignals(buckets []*model.SignalAggregations, requests []*model.SignalAggregationRequest, aliases []string) {
	if len(requests) == 0 {
		return
	}
	for _, bucket := range buckets {
		if bucket == nil {
			continue
		}
		signals := make([]*model.SignalAggregationValue, 0, len(requests))
		for i, req := range requests {
			v, ok := bucket.ValueNumbers[aliases[i]]
			if !ok {
				continue
			}
			signals = append(signals, &model.SignalAggregationValue{
				Name:  req.Name,
				Agg:   string(req.Agg),
				Value: v,
			})
		}
		bucket.Signals = signals
	}
}

// filterSignalNamesByPrivilege drops names the caller is not allowed to see and
// deduplicates while preserving input order.
func filterSignalNamesByPrivilege(names []string, permissions []string) []string {
	if len(names) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		if _, dup := seen[n]; dup {
			continue
		}
		if !hasPrivilegesForSignal(n, permissions) {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// filterSnapshotByName picks entries from a snapshot result whose names match
// the requested set, preserving the order of the names slice.
func filterSnapshotByName(snapshot []*model.LatestSignal, names []string) []*model.LatestSignal {
	if len(names) == 0 || len(snapshot) == 0 {
		return []*model.LatestSignal{}
	}
	byName := make(map[string]*model.LatestSignal, len(snapshot))
	for _, sig := range snapshot {
		if sig == nil {
			continue
		}
		byName[sig.Name] = sig
	}
	out := make([]*model.LatestSignal, 0, len(names))
	for _, name := range names {
		if sig, ok := byName[name]; ok {
			out = append(out, sig)
		}
	}
	return out
}

func permissionsFromContext(ctx context.Context) []string {
	claim, _ := ctx.Value(auth.TelemetryClaimContextKey{}).(*auth.TelemetryClaim)
	if claim == nil {
		return nil
	}
	return claim.Permissions
}
