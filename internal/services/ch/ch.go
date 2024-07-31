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
	addr := fmt.Sprintf("%s:%d", settings.CLickhouse.Host, settings.CLickhouse.Port)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.CLickhouse.User,
			Password: settings.CLickhouse.Password,
			Database: settings.CLickhouse.Database,
		},
		TLS: &tls.Config{
			RootCAs: settings.CLickhouse.RootCAs,
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
	return &Service{conn: conn, lastSeenBucketHrs: settings.ManufacturerDeviceLastSeenBucketHours}, nil
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

// GetAggregatedSignals returns a slice of aggregated signals based on the provided arguments from the ClickHouse database.
// The signals are sorted by timestamp in ascending order.
// The timestamp on each signal is for the start of the interval.
func (s *Service) GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.AggSignal, error) {
	if len(aggArgs.FloatArgs) == 0 && len(aggArgs.StringArgs) == 0 {
		return []*model.AggSignal{}, nil
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
	defer rows.Close() //nolint:errcheck // we don't care about the error here
	signals := []*vss.Signal{}
	for rows.Next() {
		var signal vss.Signal
		err := rows.Scan(&signal.Name, &signal.Timestamp, &signal.ValueNumber, &signal.ValueString)
		if err != nil {
			return nil, fmt.Errorf("failed scanning clickhouse row: %w", err)
		}
		signals = append(signals, &signal)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("clickhouse row error: %w", rows.Err())
	}
	return signals, nil
}

// TODO(elffjs): Ugly duplication.
func (s *Service) getAggSignals(ctx context.Context, stmt string, args []any) ([]*model.AggSignal, error) {
	rows, err := s.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}
	defer rows.Close() //nolint:errcheck // we don't care about the error here
	signals := []*model.AggSignal{}
	for rows.Next() {
		var signal model.AggSignal
		err := rows.Scan(&signal.Name, &signal.Agg, &signal.Timestamp, &signal.ValueNumber, &signal.ValueString)
		if err != nil {
			return nil, fmt.Errorf("failed scanning clickhouse row: %w", err)
		}
		signals = append(signals, &signal)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("clickhouse row error: %w", rows.Err())
	}
	return signals, nil
}

// GetDeviceActivity returns the start of the latest time block device activity was recorded.
func (s *Service) GetDeviceActivity(ctx context.Context, vehicleTokenID int, adManuf string) (*model.DeviceActivity, error) {
	stmt, args := getDeviceActivityQuery(s.lastSeenBucketHrs, vehicleTokenID, adManuf)
	rows := s.conn.QueryRow(ctx, stmt, args...)
	if rows.Err() != nil {
		return nil, fmt.Errorf("query failed: %w", rows.Err())
	}

	var activity model.DeviceActivity
	err := rows.Scan(&activity.LastActive)
	if err != nil {
		return nil, fmt.Errorf("failed scanning row: %w", err)
	}

	// should there be an 'ever active' that we can set to false instead of erroring here?
	if activity.LastActive.IsZero() || *activity.LastActive == time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC) {
		return nil, fmt.Errorf("no activity recorded for vehicle token ID %d", vehicleTokenID)
	}

	return &activity, nil
}
