package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

var (
	twoWeeks       = 14 * 24 * time.Hour
	sourceValues   = []string{"autopi", "macaron", "smartcar", "tesla"}
	errInvalidArgs = errors.New("invalid arguments")
)

// SignalArgs is the base arguments for querying signals.
type SignalArgs struct {
	FromTS  time.Time
	ToTS    time.Time
	Filter  *model.SignalFilter
	Name    string
	TokenID uint32
}

// FloatSignalArgs is the arguments for querying a float signals.
type FloatSignalArgs struct {
	Agg model.FloatAggregation
	SignalArgs
}

// StringSignalArgs is the arguments for querying a string signals.
type StringSignalArgs struct {
	Agg model.StringAggregation
	SignalArgs
}

// GetSignalFloats returns the float signals based on the provided arguments.
func (r *Repository) GetSignalFloats(ctx context.Context, sigArgs *FloatSignalArgs) ([]*model.SignalFloat, error) {
	if err := validateAggSigArgs(&sigArgs.SignalArgs); err != nil {
		return nil, err
	}
	interval, err := getIntervalMS(sigArgs.Agg.Interval)
	if err != nil {
		graphql.AddError(ctx, err)
		return nil, err
	}
	stmt, args := getAggQuery(sigArgs.SignalArgs, interval, getFloatAggFunc(sigArgs.Agg.Type))
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
func (r *Repository) GetSignalString(ctx context.Context, sigArgs *StringSignalArgs) ([]*model.SignalString, error) {
	if err := validateAggSigArgs(&sigArgs.SignalArgs); err != nil {
		return nil, err
	}
	interval, err := getIntervalMS(sigArgs.Agg.Interval)
	if err != nil {
		graphql.AddError(ctx, err)
		return nil, err
	}
	stmt, args := getAggQuery(sigArgs.SignalArgs, interval, getStringAgg(sigArgs.Agg.Type))
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
func (r *Repository) GetLatestSignalFloat(ctx context.Context, sigArgs *SignalArgs) (*model.SignalFloat, error) {
	var signal model.SignalFloat
	err := r.getLatestSignal(ctx, sigArgs, FloatValueCol, &signal.Value, &signal.Timestamp)
	return &signal, err
}

// GetLatestSignalString returns the latest string signal based on the provided arguments.
func (r *Repository) GetLatestSignalString(ctx context.Context, sigArgs *SignalArgs) (*model.SignalString, error) {
	var signal model.SignalString
	err := validateLastestSigArgs(sigArgs)
	if err != nil {
		return nil, err
	}
	err = r.getLatestSignal(ctx, sigArgs, StringValueCol, &signal.Value, &signal.Timestamp)
	return &signal, err
}

func (r *Repository) getLatestSignal(ctx context.Context, sigArgs *SignalArgs, valueCol string, dest ...any) error {
	stmt, args := getLatestQuery(valueCol, sigArgs)
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
func (r *Repository) GetLastSeen(ctx context.Context, sigArgs *SignalArgs) (time.Time, error) {
	err := validateLastSeenSigArgs(sigArgs)
	if err != nil {
		return time.Time{}, err
	}
	stmt, args := getLastSeenQuery(sigArgs)
	row := r.conn.QueryRow(ctx, stmt, args...)
	if row.Err() != nil {
		return time.Time{}, fmt.Errorf("failed querying clickhouse: %w", row.Err())
	}
	var timestamp time.Time
	err = row.Scan(&timestamp)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, fmt.Errorf("failed scanning clickhouse rows: %w", err)
	}
	return timestamp, nil
}

func validateAggSigArgs(args *SignalArgs) error {
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
	return validateLastestSigArgs(args)
}

func validateLastestSigArgs(args *SignalArgs) error {
	if args.Name == "" {
		return fmt.Errorf("%w name is empty", errInvalidArgs)
	}
	return validateLastSeenSigArgs(args)
}

func validateLastSeenSigArgs(args *SignalArgs) error {
	if args.TokenID == 0 {
		return fmt.Errorf("%w token id is zero", errInvalidArgs)
	}

	return validateFilter(args.Filter)
}

func validateFilter(filter *model.SignalFilter) error {
	if filter == nil {
		return nil
	}
	// if we move to storing the device address as source we can remove this check.
	if filter.Source != nil && !slices.Contains(sourceValues, *filter.Source) {
		return fmt.Errorf("%w source '%s', expected one of %v", errInvalidArgs, *filter.Source, sourceValues)
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
