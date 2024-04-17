package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
)

const signalQuery = `
SELECT
  %s as value, 
  toStartOfInterval(Timestamp, toIntervalSecond(?)) as group_timestamp 
FROM 
  signal
WHERE 
  TokenID = ? 
  AND Name = ?
  And Timestamp > ?
  And Timestamp < ?
GROUP BY 
  group_timestamp 
ORDER BY 
  group_timestamp ASC;
`
const latestSignalQuery = `
SELECT
  %s as value,
  Timestamp
FROM
  signal
WHERE
  TokenID = ?
  AND Name = ?
ORDER BY
  Timestamp DESC
LIMIT 1;
`

// SignalArgs is the base arguments for querying signals.
type SignalArgs struct {
	FromTS  time.Time
	ToTS    time.Time
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

func getFloatAggFunc(aggType *model.FloatAggregationType) string {
	var agg model.FloatAggregationType
	if aggType != nil {
		agg = *aggType
	}
	var aggStr string
	switch agg {
	case model.FloatAggregationTypeAvg:
		aggStr = "avg(ValueNumber)"
	case model.FloatAggregationTypeRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf("groupArraySample(1, %d)(ValueNumber)[1]", seed)
	case model.FloatAggregationTypeMin:
		aggStr = "min(ValueNumber)"
	case model.FloatAggregationTypeMax:
		aggStr = "max(ValueNumber)"
	case model.FloatAggregationTypeMed:
		aggStr = "median(ValueNumber)"
	default:
		aggStr = "avg(ValueNumber)"
	}
	return fmt.Sprintf(signalQuery, aggStr)
}

func getStringAgg(aggType *model.StringAggregationType) string {
	var agg model.StringAggregationType
	if aggType != nil {
		agg = *aggType
	}
	var aggStr string
	switch agg {
	case model.StringAggregationTypeRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf("groupArraySample(1, %d)(ValueString)[1]", seed)
	case model.StringAggregationTypeUnique:
		aggStr = "groupUniqArray(ValueString)"
	case model.StringAggregationTypeTop:
		aggStr = "topK(1, 10)(ValueString)"
	default:
		aggStr = "topK(1, 10)(ValueString)"
	}
	return fmt.Sprintf(signalQuery, aggStr)
}

// GetSignalFloats returns the float signals based on the provided arguments.
func (r *Repository) GetSignalFloats(ctx context.Context, sigArgs FloatSignalArgs) ([]*model.SignalFloat, error) {
	if err := validateSigArgs(sigArgs.SignalArgs); err != nil {
		return nil, err
	}
	query := getFloatAggFunc(sigArgs.Agg.Type)
	rows, err := r.conn.Query(ctx, query, sigArgs.Agg.Interval, sigArgs.TokenID, sigArgs.Name, sigArgs.FromTS, sigArgs.ToTS)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}

	defer rows.Close() //nolint:errcheck // rows.Close() is called in the caller function
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
func (r *Repository) GetSignalString(ctx context.Context, sigArgs StringSignalArgs) ([]*model.SignalString, error) {
	if err := validateSigArgs(sigArgs.SignalArgs); err != nil {
		return nil, err
	}
	query := getStringAgg(sigArgs.Agg.Type)
	rows, err := r.conn.Query(ctx, query, sigArgs.Agg.Interval, sigArgs.TokenID, sigArgs.Name, sigArgs.FromTS, sigArgs.ToTS)
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
func (r *Repository) GetLatestSignalFloat(ctx context.Context, sigArgs SignalArgs) (*model.SignalFloat, error) {
	query := fmt.Sprintf(latestSignalQuery, "ValueNumber")
	row := r.conn.QueryRow(ctx, query, sigArgs.TokenID, sigArgs.Name)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", row.Err())
	}
	var signal model.SignalFloat
	err := row.Scan(&signal.Value, &signal.Timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed scanning clickhouse rows: %w", err)
	}
	return &signal, nil
}

// GetLatestSignalString returns the latest string signal based on the provided arguments.
func (r *Repository) GetLatestSignalString(ctx context.Context, sigArgs SignalArgs) (*model.SignalString, error) {
	query := fmt.Sprintf(latestSignalQuery, "ValueString")
	row := r.conn.QueryRow(ctx, query, sigArgs.TokenID, sigArgs.Name)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", row.Err())
	}
	var signal model.SignalString
	err := row.Scan(&signal.Value, &signal.Timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed scanning clickhouse rows: %w", err)
	}
	return &signal, nil
}

func validateSigArgs(args SignalArgs) error {
	if args.FromTS.IsZero() {
		return fmt.Errorf("from timestamp is zero")
	}
	if args.ToTS.IsZero() {
		return fmt.Errorf("to timestamp is zero")
	}

	// check if time range is greater than 2 weeks
	if args.ToTS.Sub(args.FromTS) > 14*24*time.Hour {
		return fmt.Errorf("time range is greater than 2 weeks")
	}

	if args.TokenID == 0 {
		return fmt.Errorf("token id is zero")
	}
	if args.Name == "" {
		return fmt.Errorf("name is empty")
	}
	return nil
}
