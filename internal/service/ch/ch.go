// Package ch is used to interact with ClickHouse servers.
package ch

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

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
	conn clickhouse.Conn
}

// NewService creates a new ClickHouse service.
func NewService(settings config.Settings, rootCAs *x509.CertPool) (*Service, error) {
	maxExecutionTime := settings.CLickHouseMaxExecutionTimeSec
	if maxExecutionTime == 0 {
		settings.CLickHouseMaxExecutionTimeSec = defaultMaxExecutionTime
	}
	addr := fmt.Sprintf("%s:%d", settings.ClickHouseHost, settings.ClickHouseTCPPort)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.ClickHouseUser,
			Password: settings.ClickHousePassword,
			Database: settings.ClickHouseDatabase,
		},
		TLS: &tls.Config{
			RootCAs: rootCAs,
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
	return &Service{conn: conn}, nil
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
func (s *Service) GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*vss.Signal, error) {
	stmt, args := getAggQuery(aggArgs)
	signals, err := s.getSignals(ctx, stmt, args)
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
