package repositories

import (
	"fmt"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/volatiletech/sqlboiler/v4/drivers"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	// IntervalGroup is the column alias for the interval group.
	IntervalGroup = "group_timestamp"
	tokenIDWhere  = vss.TokenIDCol + " = ?"
	nameWhere     = vss.NameCol + " = ?"
	timestampFrom = vss.TimestampCol + " > ?"
	timestampTo   = vss.TimestampCol + " < ?"
	sourceWhere   = vss.SourceCol + " = ?"
	timestampDesc = vss.TimestampCol + " DESC"
	groupAsc      = IntervalGroup + " ASC"
)

// Aggregation functions for float signals.
const (
	avgGroup  = "avg(" + vss.ValueNumberCol + ")"
	randGroup = "groupArraySample(1, %d)(" + vss.ValueNumberCol + ")[1]"
	minGroup  = "min(" + vss.ValueNumberCol + ")"
	maxGroup  = "max(" + vss.ValueNumberCol + ")"
	medGroup  = "median(" + vss.ValueNumberCol + ")"
)

const (
	randStringGroup = "groupArraySample(1, %d)(" + vss.ValueStringCol + ")[1]"
	uniqueGroup     = "arrayStringConcat(groupUniqArray(" + vss.ValueStringCol + "),',')"
	topGroup        = "arrayStringConcat(topK(1, 10)(" + vss.ValueStringCol + "))"
)

// TODO: remove this map when we move to storing the device address
var sourceTranslations = map[string]string{
	"macaron":  "dimo/integration/2ULfuC8U9dOqRshZBAi0lMM1Rrx",
	"tesla":    "dimo/integration/26A5Dk3vvvQutjSyF0Jka2DP5lg",
	"autopi":   "dimo/integration/27qftVRWQYpVDcO5DltO5Ojbjxk",
	"smartcar": "dimo/integration/22N2xaPOq2WW2gAHBHd0Ikn4Zob",
}

var dialect = drivers.Dialect{
	LQ: '`',
	RQ: '`',

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
	return qm.Select(vss.TimestampCol)
}

// fromSignal adds a FROM clause to the query to select the signal table.
// Example: 'FROM signal'.
func fromSignal() qm.QueryMod {
	return qm.From(vss.TableName)
}

// withTokenID adds a WHERE clause to the query to filter by TokenID.
// Example: 'WHERE TokenID = ?'.
func withTokenID(tokenID uint32) qm.QueryMod {
	return qm.Where(tokenIDWhere, tokenID)
}

// withName adds a WHERE clause to the query to filter by Name.
// Example: 'WHERE Name = ?'.
func withName(name string) qm.QueryMod {
	return qm.Where(nameWhere, name)
}

// withFromTS adds a WHERE clause to the query to filter by Timestamp greater than fromTS.
// Example: 'WHERE Timestamp > ?'.
func withFromTS(fromTS time.Time) qm.QueryMod {
	return qm.Where(timestampFrom, fromTS)
}

// withToTS adds a WHERE clause to the query to filter by Timestamp less than toTS.
// Example: 'WHERE Timestamp < ?'.
func withToTS(toTS time.Time) qm.QueryMod {
	return qm.Where(timestampTo, toTS)
}

// withSource adds a WHERE clause to the query to filter by Source.
// Example: 'WHERE Source = ?'.
func withSource(source string) qm.QueryMod {
	// TODO: remove this logic when we move to storing the device address as source.
	if translateSource, ok := sourceTranslations[source]; ok {
		return qm.Where(sourceWhere, translateSource)
	}
	return qm.Where(sourceWhere, source)
}

// orderByTimeStampDESC adds an ORDER BY clause to the query to order by Timestamp in descending order.
// Example: 'ORDER BY Timestamp DESC'.
func orderByTimeStampDESC() qm.QueryMod {
	return qm.OrderBy(timestampDesc)
}

// orderByIntervalASC adds an ORDER BY clause to the query to order by the interval group in ascending order.
// Example: 'ORDER BY group_timestamp ASC'.
func orderByIntervalASC() qm.QueryMod {
	return qm.OrderBy(groupAsc)
}

// groupByInterval adds a GROUP BY clause to the query to group by the interval group.
// Example: 'GROUP BY group_timestamp'.
func groupByInterval() qm.QueryMod {
	return qm.GroupBy(IntervalGroup)
}

// selectInterval adds a SELECT clause to the query to select the interval group based on the given milliSeconds.
// Example: 'SELECT toStartOfInterval(Timestamp, toIntervalMillisecond(?)) as group_timestamp'.
func selectInterval(milliSeconds int64) qm.QueryMod {
	return qm.Select(fmt.Sprintf("toStartOfInterval(%s, toIntervalMillisecond(%d)) as %s", vss.TimestampCol, milliSeconds, IntervalGroup))
}

// returns a string representation of the aggregation function based on the aggregation type.
func getFloatAggFunc(aggType model.FloatAggregationType) string {
	aggStr := avgGroup
	switch aggType {
	case model.FloatAggregationTypeAvg:
		aggStr = avgGroup
	case model.FloatAggregationTypeRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf(randGroup, seed)
	case model.FloatAggregationTypeMin:
		aggStr = minGroup
	case model.FloatAggregationTypeMax:
		aggStr = maxGroup
	case model.FloatAggregationTypeMed:
		aggStr = medGroup
	}
	return aggStr
}

// returns a string representation of the aggregation function based on the aggregation type.
func getStringAgg(aggType model.StringAggregationType) string {
	aggStr := topGroup
	switch aggType {
	case model.StringAggregationTypeRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf(randStringGroup, seed)
	case model.StringAggregationTypeUnique:
		aggStr = uniqueGroup
	case model.StringAggregationTypeTop:
		aggStr = topGroup
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
func getLatestQuery(valueCol, name string, sigArgs *model.SignalArgs) (stmt string, args []any) {
	mods := []qm.QueryMod{
		selectValue(valueCol),
		selectTimestamp(),
		fromSignal(),
		withName(name),
		withTokenID(sigArgs.TokenID),
		orderByTimeStampDESC(),
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
func getAggQuery(sigArgs model.SignalArgs, intervalMS int64, name, aggFunc string) (stmt string, args []any) {
	mods := []qm.QueryMod{
		selectValue(aggFunc),
		selectInterval(intervalMS),
		fromSignal(),
		withName(name),
		withTokenID(sigArgs.TokenID),
		withFromTS(sigArgs.FromTS),
		withToTS(sigArgs.ToTS),
		groupByInterval(),
		orderByIntervalASC(),
	}
	mods = append(mods, getFilterMods(sigArgs.Filter)...)
	return newQuery(mods...)
}

// creates a query to get the last seen timestamp of a token.
/*
SELECT
	Timestamp
FROM
	signal
WHERE
	TokenID = ?
ORDER BY
	Timestamp DESC
LIMIT 1;
*/
func getLastSeenQuery(sigArgs *model.SignalArgs) (stmt string, args []any) {
	mods := []qm.QueryMod{
		selectTimestamp(),
		fromSignal(),
		withTokenID(sigArgs.TokenID),
		orderByTimeStampDESC(),
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
