// Package ch is used to interact with ClickHouse servers.
package ch

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

const (
	defaultMaxExecutionTime                    = 5
	defaultTimeoutBeforeCheckingExecutionSpeed = 2
	// TimeoutErrCode is the error code returned by ClickHouse when a query is interrupted due to exceeding the max_execution_time.
	TimeoutErrCode = int32(159)
)

// Service is a ClickHouse service that interacts with the ClickHouse database.
type Service struct {
	conn              clickhouse.Conn
	lastSeenBucketHrs int64
}

// NewService creates a new ClickHouse service.
func NewService(settings config.Settings) (*Service, error) {
	maxExecutionTime, err := getMaxExecutionTime(settings.MaxRequestDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to get max execution time: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", settings.Clickhouse.Host, settings.Clickhouse.Port)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.Clickhouse.User,
			Password: settings.Clickhouse.Password,
			Database: settings.Clickhouse.Database,
		},
		TLS: &tls.Config{
			RootCAs: settings.Clickhouse.RootCAs,
		},
		Settings: map[string]any{
			// ClickHouse will interrupt a query if the projected execution time exceeds the specified max_execution_time.
			// The estimated execution time is calculated after `timeout_before_checking_execution_speed`
			// More info: https://clickhouse.com/docs/en/operations/settings/query-complexity#max-execution-time
			"max_execution_time":                      maxExecutionTime,
			"timeout_before_checking_execution_speed": defaultTimeoutBeforeCheckingExecutionSpeed,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}
	return &Service{conn: conn, lastSeenBucketHrs: settings.DeviceLastSeenBinHrs}, nil
}

func getMaxExecutionTime(maxRequestDuration string) (int, error) {
	if maxRequestDuration == "" {
		return defaultMaxExecutionTime, nil
	}
	maxExecutionTime, err := time.ParseDuration(maxRequestDuration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse max request duration: %w", err)
	}
	return int(maxExecutionTime.Seconds()), nil
}

// GetLatestSignals returns the latest signals based on the provided arguments from the ClickHouse database.
func (s *Service) GetLatestSignals(ctx context.Context, latestArgs *model.LatestSignalsArgs) ([]*vss.Signal, error) {
	filteredArgs, err := s.getFilteredLatestSignalsArgs(ctx, latestArgs)
	if err != nil {
		return nil, err
	}

	// If no signals are available but lastSeen is requested, return only lastSeen
	if len(filteredArgs.SignalNames) == 0 {
		if filteredArgs.IncludeLastSeen {
			lastSeenStmt, lastSeenArgs := getLastSeenQuery(&filteredArgs.SignalArgs)
			signals, err := s.getSignals(ctx, lastSeenStmt, lastSeenArgs)
			if err != nil {
				return nil, err
			}
			return signals, nil
		}
		return []*vss.Signal{}, nil
	}

	stmt, args := getLatestQuery(filteredArgs)
	if filteredArgs.IncludeLastSeen {
		lastSeenStmt, lastSeenArgs := getLastSeenQuery(&filteredArgs.SignalArgs)
		stmt, args = unionAll([]string{stmt, lastSeenStmt}, [][]any{args, lastSeenArgs})
	}

	signals, err := s.getSignals(ctx, stmt, args)
	if err != nil {
		return nil, err
	}
	return signals, nil
}

type AliasHandleMapper struct {
	aliasToHandle, handleToAlias map[string]string
}

func NewAliasHandleMapper() *AliasHandleMapper {
	return &AliasHandleMapper{
		aliasToHandle: make(map[string]string),
		handleToAlias: make(map[string]string),
	}
}

func (m *AliasHandleMapper) Add(alias, handle string) {
	m.aliasToHandle[alias] = handle
	m.handleToAlias[handle] = alias
}

func (m *AliasHandleMapper) Handle(alias string) string {
	return m.aliasToHandle[alias]
}

func (m *AliasHandleMapper) Alias(handle string) string {
	return m.handleToAlias[handle]
}

// GetAggregatedSignals returns a slice of aggregated signals based on the provided arguments from the ClickHouse database.
// The signals are sorted by timestamp in ascending order.
// The timestamp on each signal is for the start of the interval.
func (s *Service) GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*AggSignal, error) {
	if len(aggArgs.FloatArgs) == 0 && len(aggArgs.StringArgs) == 0 && len(aggArgs.ApproxLocArgs) == 0 {
		return []*AggSignal{}, nil
	}

	filteredArgs, err := s.getFilteredAggregatedSignalArgs(ctx, aggArgs)
	if err != nil {
		return nil, err
	}
	if len(filteredArgs.FloatArgs) == 0 && len(filteredArgs.StringArgs) == 0 && len(filteredArgs.ApproxLocArgs) == 0 {
		return []*AggSignal{}, nil
	}

	stmt, args, err := getAggQuery(filteredArgs)
	if err != nil {
		return nil, err
	}

	signals, err := s.getAggSignals(ctx, stmt, args)
	if err != nil {
		return nil, err
	}

	return signals, nil
}

func (s *Service) getSignals(ctx context.Context, stmt string, args []any) ([]*vss.Signal, error) {
	rows, err := s.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}
	signals := []*vss.Signal{}
	for rows.Next() {
		var signal vss.Signal
		err := rows.Scan(&signal.Name, &signal.Timestamp, &signal.ValueNumber, &signal.ValueString)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("failed scanning clickhouse row: %w", err)
		}
		signals = append(signals, &signal)
	}
	_ = rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("clickhouse row error: %w", rows.Err())
	}
	return signals, nil
}

type AggSignal struct {
	// SignalType describes the type of values in the aggregation:
	// float, string, or approximate location.
	SignalType FieldType
	// SignalIndex is an identifier for the aggregation within its
	// SignalType.
	//
	// For float and string aggregations this is simply an index
	// into the corresponding argument array.
	//
	// For approximate location (SignalType = AppLocType = 3), we
	// imagine expanding each element of the slice
	// model.AllFloatAggregation into two: first the latitude and then
	// the longitude. So, for example, SignalType = 3 and
	// SignalIndex = 4 means we want approximate latitude (4 % 2 = 0)
	// for the index 2 (4 / 2 = 2) float aggregation.
	//
	// We could get away with a single number, since we know how many
	// arguments of each type there are, but it appears to us that this
	// would make adding new types riskier.
	SignalIndex uint16
	// Timestamp is the timestamp for the bucket, the leftmost point.
	Timestamp time.Time
	// ValueNumber is the value for this row if it is of float or
	// approximate location type.
	ValueNumber float64
	// ValueNumber is the value for this row if it is of float or
	// approximate location type.
	ValueString string
}

func (s *Service) getAggSignals(ctx context.Context, stmt string, args []any) ([]*AggSignal, error) {
	rows, err := s.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}
	signals := []*AggSignal{}
	for rows.Next() {
		var signal AggSignal
		err := rows.Scan(&signal.SignalType, &signal.SignalIndex, &signal.Timestamp, &signal.ValueNumber, &signal.ValueString)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("failed scanning clickhouse row: %w", err)
		}
		signals = append(signals, &signal)
	}
	_ = rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("clickhouse row error: %w", rows.Err())
	}
	return signals, nil
}

// GetAvailableSignals returns a slice of available signals from the ClickHouse database.
// if no signals are available, a nil slice is returned.
func (s *Service) GetAvailableSignals(ctx context.Context, tokenId uint32, filter *model.SignalFilter) ([]string, error) {
	stmt, args := getDistinctQuery(tokenId, filter)
	rows, err := s.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}
	var signals []string
	for rows.Next() {
		var signal string
		err := rows.Scan(&signal)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("failed scanning clickhouse row: %w", err)
		}
		signals = append(signals, signal)
	}

	_ = rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("clickhouse row error: %w", rows.Err())
	}
	return signals, nil
}

// getFilteredLatestSignalsArgs gets available signals and filters the LatestSignalsArgs to only include available signals.
func (s *Service) getFilteredLatestSignalsArgs(ctx context.Context, args *model.LatestSignalsArgs) (*model.LatestSignalsArgs, error) {
	availableSignals, err := s.GetAvailableSignals(ctx, args.TokenID, args.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get available signals: %w", err)
	}
	return s.filterLatestSignalsArgs(args, availableSignals), nil
}

// getFilteredAggregatedSignalArgs gets available signals and filters the AggregatedSignalArgs to only include available signals.
func (s *Service) getFilteredAggregatedSignalArgs(ctx context.Context, args *model.AggregatedSignalArgs) (*model.AggregatedSignalArgs, error) {
	availableSignals, err := s.GetAvailableSignals(ctx, args.TokenID, args.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get available signals: %w", err)
	}
	return s.filterAggregatedSignalArgs(args, availableSignals), nil
}

// filterLatestSignalsArgs filters the LatestSignalsArgs to only include signals that are available in the database.
func (s *Service) filterLatestSignalsArgs(args *model.LatestSignalsArgs, availableSignals []string) *model.LatestSignalsArgs {
	availableSet := make(map[string]struct{}, len(availableSignals))
	for _, signal := range availableSignals {
		availableSet[signal] = struct{}{}
	}

	filteredSignalNames := make(map[string]struct{})
	for signalName := range args.SignalNames {
		if _, exists := availableSet[signalName]; exists {
			filteredSignalNames[signalName] = struct{}{}
		}
	}

	return &model.LatestSignalsArgs{
		SignalArgs:      args.SignalArgs,
		SignalNames:     filteredSignalNames,
		IncludeLastSeen: args.IncludeLastSeen,
	}
}

// filterAggregatedSignalArgs filters the AggregatedSignalArgs to only include signals that are available in the database.
func (s *Service) filterAggregatedSignalArgs(args *model.AggregatedSignalArgs, availableSignals []string) *model.AggregatedSignalArgs {
	availableSet := make(map[string]struct{}, len(availableSignals))
	for _, signal := range availableSignals {
		availableSet[signal] = struct{}{}
	}

	// Filter FloatArgs
	filteredFloatArgs := make([]model.FloatSignalArgs, 0, len(args.FloatArgs))
	for _, floatArg := range args.FloatArgs {
		if _, exists := availableSet[floatArg.Name]; exists {
			filteredFloatArgs = append(filteredFloatArgs, floatArg)
		}
	}

	// Filter StringArgs
	filteredStringArgs := make([]model.StringSignalArgs, 0, len(args.StringArgs))
	for _, stringArg := range args.StringArgs {
		if _, exists := availableSet[stringArg.Name]; exists {
			filteredStringArgs = append(filteredStringArgs, stringArg)
		}
	}

	// Filter ApproxLocArgs - check if latitude and longitude fields are available
	filteredApproxLocArgs := make(map[model.FloatAggregation]struct{})
	_, hasLat := availableSet["currentLocationLatitude"]
	_, hasLong := availableSet["currentLocationLongitude"]
	if hasLat && hasLong {
		// Only include approximate location args if both lat and long are available
		for agg := range args.ApproxLocArgs {
			filteredApproxLocArgs[agg] = struct{}{}
		}
	}

	return &model.AggregatedSignalArgs{
		SignalArgs:    args.SignalArgs,
		FromTS:        args.FromTS,
		ToTS:          args.ToTS,
		Interval:      args.Interval,
		FloatArgs:     filteredFloatArgs,
		StringArgs:    filteredStringArgs,
		ApproxLocArgs: filteredApproxLocArgs,
	}
}

func (s *Service) GetEvents(ctx context.Context, subject string, from, to time.Time, filter *model.EventFilter) ([]*vss.Event, error) {
	mods := []qm.QueryMod{
		qm.Select(vss.EventNameCol, vss.EventSourceCol, vss.EventTimestampCol, vss.EventDurationNsCol, vss.EventMetadataCol),
		qm.From(vss.EventTableName),
		qm.Where(eventSubjectWhere, subject),
		qm.Where(timestampFrom, from),
		qm.Where(timestampTo, to),
		qm.OrderBy(vss.EventTimestampCol + " DESC"),
	}
	mods = appendEventFilterMods(mods, filter)
	stmt, args := newQuery(mods...)

	rows, err := s.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse for events: %w", err)
	}
	events := []*vss.Event{}
	for rows.Next() {
		var event vss.Event
		err := rows.Scan(&event.Name, &event.Source, &event.Timestamp, &event.DurationNs, &event.Metadata)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("failed scanning clickhouse event row: %w", err)
		}
		events = append(events, &event)
	}
	_ = rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("clickhouse event row error: %w", rows.Err())
	}
	return events, nil
}
