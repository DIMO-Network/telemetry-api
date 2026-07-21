package queryRecorder

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/telemetry-api/internal/dtcmiddleware"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

// ReplayLogMessage is the message on every replay record line, for log pipeline filtering.
const ReplayLogMessage = "query replay record"

// ReplayLogger is a GraphQL extension that emits one structured log line per
// operation submitted by an allowlisted developer. Each line carries the raw
// query, its variables, and the caller's identity so the request can be
// replayed against another environment.
type ReplayLogger struct {
	devs map[common.Address]struct{}
}

// NewReplayLogger parses a comma-separated list of developer license addresses
// into a ReplayLogger. Returns nil if the list is empty, meaning recording is
// disabled.
func NewReplayLogger(developers string) (*ReplayLogger, error) {
	devs := make(map[common.Address]struct{})
	for dev := range strings.SplitSeq(developers, ",") {
		dev = strings.TrimSpace(dev)
		if dev == "" {
			continue
		}
		if !common.IsHexAddress(dev) {
			return nil, fmt.Errorf("invalid developer license address %q", dev)
		}
		devs[common.HexToAddress(dev)] = struct{}{}
	}
	if len(devs) == 0 {
		return nil, nil
	}
	return &ReplayLogger{devs: devs}, nil
}

func (*ReplayLogger) ExtensionName() string {
	return "ReplayLogger"
}

func (*ReplayLogger) Validate(graphql.ExecutableSchema) error {
	return nil
}

func (r *ReplayLogger) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	r.record(ctx)
	return next(ctx)
}

func (r *ReplayLogger) record(ctx context.Context) {
	op := graphql.GetOperationContext(ctx)
	if op == nil || op.RawQuery == "" {
		return
	}
	subject, tokenID, _ := dtcmiddleware.GetSubjectAndTokenID(ctx)
	if !common.IsHexAddress(subject) {
		return
	}
	if _, ok := r.devs[common.HexToAddress(subject)]; !ok {
		return
	}
	evt := zerolog.Ctx(ctx).Info().
		Str("developer", subject).
		Str("operationName", op.OperationName).
		Str("query", op.RawQuery).
		Interface("variables", op.Variables).
		Time("operationStart", op.Stats.OperationStart)
	if tokenID != nil {
		evt = evt.Str("vehicleTokenId", tokenID.String())
	}
	evt.Msg(ReplayLogMessage)
}
