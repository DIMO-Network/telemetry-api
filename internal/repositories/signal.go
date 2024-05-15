package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

var (
	twoWeeks       = 14 * 24 * time.Hour
	errInvalidArgs = errors.New("invalid arguments")
)

// GetSignalFloats returns the float signals based on the provided arguments.
func (r *Repository) GetSignalFloats(ctx context.Context, sigArgs *model.FloatSignalArgs) ([]*model.SignalFloat, error) {
	if err := validateAggSigArgs(&sigArgs.SignalArgs); err != nil {
		return nil, err
	}
	interval, err := getIntervalMS(sigArgs.Agg.Interval)
	if err != nil {
		graphql.AddError(ctx, err)
		return nil, err
	}
	stmt, args := getAggQuery(sigArgs.SignalArgs, interval, sigArgs.Name, getFloatAggFunc(sigArgs.Agg.Type))
	rows, err := r.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}

	defer rows.Close() //nolint:errcheck // we don't care about the error here
	signals := []*model.SignalFloat{}
	for rows.Next() {
		var signal model.SignalFloat
		err := rows.Scan(&signal.Value, &signal.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed scanning clickhouse rows: %w", err)
		}
		signals = append(signals, &signal)
	}
	return signals, nil
}

// GetSignalString returns the string signals based on the provided arguments.
func (r *Repository) GetSignalString(ctx context.Context, sigArgs *model.StringSignalArgs) ([]*model.SignalString, error) {
	if err := validateAggSigArgs(&sigArgs.SignalArgs); err != nil {
		return nil, err
	}
	interval, err := getIntervalMS(sigArgs.Agg.Interval)
	if err != nil {
		graphql.AddError(ctx, err)
		return nil, err
	}
	stmt, args := getAggQuery(sigArgs.SignalArgs, interval, sigArgs.Name, getStringAgg(sigArgs.Agg.Type))
	rows, err := r.conn.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}
	defer rows.Close()
	signals := []*model.SignalString{}
	for rows.Next() {
		var signal model.SignalString
		err := rows.Scan(&signal.Value, &signal.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed scanning clickhouse rows: %w", err)
		}
		signals = append(signals, &signal)
	}
	return signals, nil
}

// GetLatestSignalFloat returns the latest float signal based on the provided arguments.
func (r *Repository) GetLatestSignalFloat(ctx context.Context, sigArgs *model.FloatSignalArgs) (*model.SignalFloat, error) {
	var signal model.SignalFloat
	err := r.getLatestSignal(ctx, &sigArgs.SignalArgs, sigArgs.Name, vss.ValueNumberCol, &signal.Value, &signal.Timestamp)
	return &signal, err
}

// GetLatestSignalString returns the latest string signal based on the provided arguments.
func (r *Repository) GetLatestSignalString(ctx context.Context, sigArgs *model.StringSignalArgs) (*model.SignalString, error) {
	var signal model.SignalString
	err := validateLastest(&sigArgs.SignalArgs)
	if err != nil {
		return nil, err
	}
	err = r.getLatestSignal(ctx, &sigArgs.SignalArgs, sigArgs.Name, vss.ValueStringCol, &signal.Value, &signal.Timestamp)
	return &signal, err
}

func (r *Repository) getLatestSignal(ctx context.Context, sigArgs *model.SignalArgs, name, valueCol string, dest ...any) error {
	stmt, args := getLatestQuery(valueCol, name, sigArgs)
	row := r.conn.QueryRow(ctx, stmt, args...)
	if row.Err() != nil {
		return fmt.Errorf("failed querying clickhouse: %w", row.Err())
	}
	err := row.Scan(dest...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed scanning clickhouse rows: %w", err)
	}
	return nil
}

// GetLastSeen returns the last seen timestamp of a token.
func (r *Repository) GetLastSeen(ctx context.Context, sigArgs *model.SignalArgs) (*time.Time, error) {
	err := validateLastest(sigArgs)
	if err != nil {
		return nil, err
	}
	stmt, args := getLastSeenQuery(sigArgs)
	row := r.conn.QueryRow(ctx, stmt, args...)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", row.Err())
	}
	var timestamp time.Time
	err = row.Scan(&timestamp)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed scanning clickhouse rows: %w", err)
	}
	return &timestamp, nil
}

func validateAggSigArgs(args *model.SignalArgs) error {
	if args.FromTS.IsZero() {
		return fmt.Errorf("%w from timestamp is zero", errInvalidArgs)
	}
	if args.ToTS.IsZero() {
		return fmt.Errorf("%w to timestamp is zero", errInvalidArgs)
	}

	// check if time range is greater than 2 weeks
	if args.ToTS.Sub(args.FromTS) > twoWeeks {
		return fmt.Errorf("%w time range is greater than 2 weeks", errInvalidArgs)
	}
	return validateLastest(args)
}

func validateLastest(args *model.SignalArgs) error {
	if args.TokenID < 1 {
		return fmt.Errorf("%w tokenId is not a non-zero uint32", errInvalidArgs)
	}

	return validateFilter(args.Filter)
}

func validateFilter(filter *model.SignalFilter) error {
	if filter == nil {
		return nil
	}
	// TODO: remove this check when we move to storing the device address as source
	if filter.Source != nil {
		if _, ok := sourceTranslations[*filter.Source]; !ok {
			return fmt.Errorf("%w source '%s', is not a valid value", errInvalidArgs, *filter.Source)
		}
	}
	return nil
}

// getIntervalMS returns the interval in milliseconds.
func getIntervalMS(interval string) (int64, error) {
	dur, err := time.ParseDuration(interval)
	if err != nil {
		return 0, fmt.Errorf("failed parsing interval: %w", err)
	}
	if dur < time.Millisecond {
		return 0, fmt.Errorf("%w interval less than 1 millisecond are not supported", errInvalidArgs)
	}
	return dur.Milliseconds(), nil
}
