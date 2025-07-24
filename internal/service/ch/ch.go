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
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	stmt, args := getLatestQuery(latestArgs)
	if latestArgs.IncludeLastSeen {
		lastSeenStmt, lastSeenArgs := getLastSeenQuery(&latestArgs.SignalArgs)
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

	stmt, args, err := getAggQuery(aggArgs)
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
	// For approximate location, we imagine expanding each element of
	// the slice model.AllFloatAggregation into two: first the
	// longitude and then the latitude. So, for example, SignalType = 3
	// and SignalIndex = 3 means latitude for the 1-th float
	// aggregation.
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
