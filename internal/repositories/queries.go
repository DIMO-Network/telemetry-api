package repositories

import (
	"fmt"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/volatiletech/sqlboiler/v4/drivers"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	// FloatValueCol is the column name for float values.
	FloatValueCol = "ValueNumber"

	// StringValueCol is the column name for string values.
	StringValueCol = "ValueString"

	// IntervalGroup is the column alias for the interval group.
	IntervalGroup = "group_timestamp"
)

var sourceTranslations = map[string]string{
	"macron":   "dimo/integration/2ULfuC8U9dOqRshZBAi0lMM1Rrx",
	"tesla":    "dimo/integration/26A5Dk3vvvQutjSyF0Jka2DP5lg",
	"autopi":   "dimo/integration/27qftVRWQYpVDcO5DltO5Ojbjxk",
	"smartcar": "dimo/integration/22N2xaPOq2WW2gAHBHd0Ikn4Zob",
}

var dialect = drivers.Dialect{
	LQ: ' ',
	RQ: ' ',

	UseIndexPlaceholders:    false,
	UseLastInsertID:         false,
	UseSchema:               false,
	UseDefaultKeyword:       false,
	UseAutoColumns:          false,
	UseTopClause:            false,
	UseOutputClause:         false,
	UseCaseWhenExistsClause: false,
}

// newQuery initializes a new Query using the passed in QueryMods.
func newQuery(mods ...qm.QueryMod) (string, []any) {
	q := &queries.Query{}
	queries.SetDialect(q, &dialect)
	qm.Apply(q, mods...)
	return queries.BuildQuery(q)
}

// selectValue adds a SELECT clause to the query to select a specific value column.
// Example: SELECT valueCol as value.
func selectValue(valueCol string) qm.QueryMod {
	return qm.Select(valueCol + " as value")
}

// selectTimestamp adds a SELECT clause to the query to select the Timestamp column.
// Example: 'SELECT Timestamp'.
func selectTimestamp() qm.QueryMod {
	return qm.Select("Timestamp")
}

// fromSignal adds a FROM clause to the query to select the signal table.
// Example: 'FROM signal'.
func fromSignal() qm.QueryMod {
	return qm.From("signal")
}

// withTokenID adds a WHERE clause to the query to filter by TokenID.
// Example: 'WHERE TokenID = ?'.
func withTokenID(tokenID uint32) qm.QueryMod {
	return qm.Where("TokenID = ?", tokenID)
}

// withName adds a WHERE clause to the query to filter by Name.
// Example: 'WHERE Name = ?'.
func withName(name string) qm.QueryMod {
	return qm.Where("Name = ?", name)
}

// withFromTS adds a WHERE clause to the query to filter by Timestamp greater than fromTS.
// Example: 'WHERE Timestamp > ?'.
func withFromTS(fromTS time.Time) qm.QueryMod {
	return qm.Where("Timestamp > ?", fromTS)
}

// withToTS adds a WHERE clause to the query to filter by Timestamp less than toTS.
// Example: 'WHERE Timestamp < ?'.
func withToTS(toTS time.Time) qm.QueryMod {
	return qm.Where("Timestamp < ?", toTS)
}

// withSource adds a WHERE clause to the query to filter by Source.
// Example: 'WHERE Source = ?'.
func withSource(source string) qm.QueryMod {
	if translateSource, ok := sourceTranslations[source]; ok {
		return qm.Where("Source = ?", translateSource)
	}
	return qm.Where("Source = ?", source)
}

// orderByTS adds an ORDER BY clause to the query to order by Timestamp in descending order.
// Example: 'ORDER BY Timestamp DESC'.
func orderByTS() qm.QueryMod {
	return qm.OrderBy("Timestamp DESC")
}

// orderByInterval adds an ORDER BY clause to the query to order by the interval group in ascending order.
// Example: 'ORDER BY group_timestamp ASC'.
func orderByInterval() qm.QueryMod {
	return qm.OrderBy(IntervalGroup + " ASC")
}

// groupByInterval adds a GROUP BY clause to the query to group by the interval group.
// Example: 'GROUP BY group_timestamp'.
func groupByInterval() qm.QueryMod {
	return qm.GroupBy(IntervalGroup)
}

// selectInterval adds a SELECT clause to the query to select the interval group based on the given milliSeconds.
// Example: 'SELECT toStartOfInterval(Timestamp, toIntervalMillisecond(?)) as group_timestamp'.
func selectInterval(milliSeconds int64) qm.QueryMod {
	return qm.Select(fmt.Sprintf("toStartOfInterval(Timestamp, toIntervalMillisecond(%d)) as %s", milliSeconds, IntervalGroup))
}

func getFloatAggFunc(aggType model.FloatAggregationType) string {
	var aggStr string
	switch aggType {
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
	return aggStr
}

func getStringAgg(aggType model.StringAggregationType) string {
	var aggStr string
	switch aggType {
	case model.StringAggregationTypeRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf("groupArraySample(1, %d)(ValueString)[1]", seed)
	case model.StringAggregationTypeUnique:
		aggStr = "arrayStringConcat(groupUniqArray(ValueString),',')"
	case model.StringAggregationTypeTop:
		aggStr = "arrayStringConcat(topK(1, 10)(ValueString))"
	default:
		aggStr = "arrayStringConcat(topK(1, 10)(ValueString))"
	}
	return aggStr
}

// creates a query to get the latest signal value.
/*
SELECT
  valCol as value,
  Timestamp
FROM
  signal
WHERE
  TokenID = ?
  AND Name = ?
ORDER BY
 Timestamp DESC
 LIMIT 1;
*/
func getLatestQuery(valueCol string, sigArgs *SignalArgs) (stmt string, args []any) {
	mods := []qm.QueryMod{
		selectValue(valueCol),
		selectTimestamp(),
		fromSignal(),
		withName(sigArgs.Name),
		withTokenID(sigArgs.TokenID),
		orderByTS(),
		qm.Limit(1),
	}
	mods = append(mods, getFilterMods(sigArgs.Filter)...)
	return newQuery(mods...)
}

// creates a query to get the aggregated signal values.
/*
SELECT
	aggFunc as value,
	toStartOfInterval(Timestamp, toIntervalMillisecond(intervalMS)) as group_timestamp
FROM
	signal
WHERE
	TokenID = ?
	AND Name = ?
	AND Timestamp > ?
	AND Timestamp < ?
GROUP BY
	group_timestamp
ORDER BY
	group_timestamp ASC;
*/
func getAggQuery(sigArgs SignalArgs, intervalMS int64, aggFunc string) (stmt string, args []any) {
	mods := []qm.QueryMod{
		selectValue(aggFunc),
		selectInterval(intervalMS),
		fromSignal(),
		withName(sigArgs.Name),
		withTokenID(sigArgs.TokenID),
		withFromTS(sigArgs.FromTS),
		withToTS(sigArgs.ToTS),
		groupByInterval(),
		orderByInterval(),
	}
	mods = append(mods, getFilterMods(sigArgs.Filter)...)
	return newQuery(mods...)
}

func getLastSeenQuery(sigArgs *SignalArgs) (stmt string, args []any) {
	mods := []qm.QueryMod{
		selectTimestamp(),
		fromSignal(),
		withTokenID(sigArgs.TokenID),
		orderByTS(),
		qm.Limit(1),
	}
	mods = append(mods, getFilterMods(sigArgs.Filter)...)
	return newQuery(mods...)
}

// getFilterMods returns the query mods for the filter.
func getFilterMods(filter *model.SignalFilter) []qm.QueryMod {
	if filter == nil {
		return nil
	}
	var mods []qm.QueryMod
	if filter.Source != nil {
		mods = append(mods, withSource(*filter.Source))
	}
	return mods
}
