package repositories

import (
	"context"
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
	query := getFloatAggFunc(sigArgs.Agg.Type)
	rows, err := r.conn.Query(ctx, query, sigArgs.Agg.Interval, sigArgs.TokenID, sigArgs.Name, sigArgs.FromTS, sigArgs.ToTS)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}
	defer rows.Close()
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
