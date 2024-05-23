package ch

import (
	"fmt"
	"strings"
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
	nameIn        = vss.NameCol + " IN ?"
	timestampFrom = vss.TimestampCol + " > ?"
	timestampTo   = vss.TimestampCol + " < ?"
	sourceWhere   = vss.SourceCol + " = ?"
	timestampDesc = vss.TimestampCol + " DESC"
	groupAsc      = IntervalGroup + " ASC"
)

// varibles for the last seen signal query.
const (
	lastSeenName = "'" + model.LastSeenField + "' AS name"
	numValAsNull = "NULL AS " + vss.ValueNumberCol
	strValAsNull = "NULL AS " + vss.ValueStringCol
	lastSeenTS   = "max(" + vss.TimestampCol + ") AS ts"
)

// Aggregation functions for latest signals.
const (
	latestString    = "argMax(" + vss.ValueStringCol + ", " + vss.TimestampCol + ") as " + vss.ValueStringCol
	latestNumber    = "argMax(" + vss.ValueNumberCol + ", " + vss.TimestampCol + ") as " + vss.ValueNumberCol
	latestTimestamp = "max(" + vss.TimestampCol + ") as ts"
)

// Aggregation functions for float signals.
const (
	avgGroup  = "avg(" + vss.ValueNumberCol + ")"
	randGroup = "groupArraySample(1, %d)(" + vss.ValueNumberCol + ")[1]"
	minGroup  = "min(" + vss.ValueNumberCol + ")"
	maxGroup  = "max(" + vss.ValueNumberCol + ")"
	medGroup  = "median(" + vss.ValueNumberCol + ")"
)

// Aggregation functions for string signals.
const (
	randStringGroup = "groupArraySample(1, %d)(" + vss.ValueStringCol + ")[1]"
	uniqueGroup     = "arrayStringConcat(groupUniqArray(" + vss.ValueStringCol + "),',')"
	topGroup        = "arrayStringConcat(topK(1, 10)(" + vss.ValueStringCol + "))"
)

// TODO: remove this map when we move to storing the device address
var SourceTranslations = map[string]string{
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
	if translateSource, ok := SourceTranslations[source]; ok {
		return qm.Where(sourceWhere, translateSource)
	}
	return qm.Where(sourceWhere, source)
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

func selectNumberAggs(numberAggs []model.FloatSignalArgs) qm.QueryMod {
	if len(numberAggs) == 0 {
		return qm.Select("NULL AS " + vss.ValueNumberCol)
	}
	// Add a CASE statement for each name and its corresponding aggregation function
	caseStmts := make([]string, len(numberAggs))
	for i := range numberAggs {
		caseStmts[i] = fmt.Sprintf("WHEN %s = '%s' THEN %s", vss.NameCol, numberAggs[i].Name, getFloatAggFunc(numberAggs[i].Agg))
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(caseStmts, " "), vss.ValueNumberCol)
	return qm.Select(caseStmt)
}

func selectStringAggs(stringAggs []model.StringSignalArgs) qm.QueryMod {
	if len(stringAggs) == 0 {
		return qm.Select("NULL AS " + vss.ValueStringCol)
	}
	// Add a CASE statement for each name and its corresponding aggregation function
	caseStmts := make([]string, len(stringAggs))
	for i := range stringAggs {
		caseStmts[i] = fmt.Sprintf("WHEN %s = '%s' THEN %s", vss.NameCol, stringAggs[i].Name, getStringAgg(stringAggs[i].Agg))
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(caseStmts, " "), vss.ValueStringCol)
	return qm.Select(caseStmt)
}

// returns a string representation of the aggregation function based on the aggregation type.
func getFloatAggFunc(aggType model.FloatAggregation) string {
	aggStr := avgGroup
	switch aggType {
	case model.FloatAggregationAvg:
		aggStr = avgGroup
	case model.FloatAggregationRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf(randGroup, seed)
	case model.FloatAggregationMin:
		aggStr = minGroup
	case model.FloatAggregationMax:
		aggStr = maxGroup
	case model.FloatAggregationMed:
		aggStr = medGroup
	}
	return aggStr
}

// returns a string representation of the aggregation function based on the aggregation type.
func getStringAgg(aggType model.StringAggregation) string {
	aggStr := topGroup
	switch aggType {
	case model.StringAggregationRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf(randStringGroup, seed)
	case model.StringAggregationUnique:
		aggStr = uniqueGroup
	case model.StringAggregationTop:
		aggStr = topGroup
	}
	return aggStr
}

// GetLatestQuery creates a query to get the latest signal value for each signal names
// returns the query statement and the arguments list,
/*
SELECT
  name,
  max(timestamp),
  argMax(value_string, timestamp) as value_string,
  argMax(value_number, timestamp) as value_float
FROM
  signal
WHERE
  token_id = 15 AND
  (name = 'speed' OR name = 'currentLocationLatitude' OR name = 'currentLocationLongitude' OR name = 'powertrainFuelSystemSupportedFuelTypes' OR name = 'none')
Group BY
  name
*/
func GetLatestQuery(latestArgs *model.LatestSignalsArgs) (string, []any) {
	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		qm.Select(latestTimestamp),
		qm.Select(latestNumber),
		qm.Select(latestString),
		qm.From(vss.TableName),
		withTokenID(latestArgs.TokenID),
		qm.WhereIn(nameIn, latestArgs.SignalNames),
		qm.GroupBy(vss.NameCol),
	}
	mods = append(mods, getFilterMods(latestArgs.Filter)...)
	return newQuery(mods...)
}

// GetLastSeenQuery creates a query to get the last seen timestamp of any signal.
// returns the query statement and the arguments list,
/*
SELECT
	'lastSeen' AS name,
	max(timestamp) AS ts,
	NULL AS value_string,
	NULL AS value_float
FROM
	signal
WHERE
	token_id = 15
*/
func getLastSeenQuery(sigArgs *model.SignalArgs) (string, []any) {
	if sigArgs == nil {
		return "", nil
	}
	mods := []qm.QueryMod{
		qm.Select(lastSeenName),
		qm.Select(lastSeenTS),
		qm.Select(numValAsNull),
		qm.Select(strValAsNull),
		qm.From(vss.TableName),
		withTokenID(sigArgs.TokenID),
	}
	mods = append(mods, getFilterMods(sigArgs.Filter)...)
	return newQuery(mods...)
}

// UnionAll creates a UNION ALL statement from the given statements and arguments.
func UnionAll(allStatements []string, allArgs [][]any) (string, []any) {
	var args []any
	for i := range allStatements {
		allStatements[i] = strings.TrimSuffix(allStatements[i], ";")
	}
	unionStmt := strings.Join(allStatements, " UNION ALL ")
	for _, arg := range allArgs {
		args = append(args, arg...)
	}
	return unionStmt, args
}

// getAggQuery creates a single query to perform multiple aggregations on the signal data in the same time range and interval.
/*
SELECT
    `name`,
    toStartOfInterval(timestamp, toIntervalMillisecond(30000)) AS group_timestamp,
    CASE
        WHEN name = 'speed' THEN max(value_number)
        WHEN name = 'oBDRunTime' THEN median(value_number)
        ELSE NULL
    END AS value_number,
    CASE
        WHEN name = 'powertrainType' THEN arrayStringConcat(groupUniqArray(value_string),',')
        WHEN name = 'powertrainFuelSystemSupportedFuelTypes' THEN groupArraySample(1, 1716404995385)(value_string)[1]
        ELSE NULL
    END AS value_string
FROM
    `signal`
WHERE
    `name` IN ['speed', 'oBDRunTime', 'powertrainType', 'powertrainFuelSystemSupportedFuelTypes']
    AND token_id = 15
    AND timestamp > toDateTime('2024-04-15 09:21:19')
    AND timestamp < toDateTime('2024-04-27 09:21:19')
GROUP BY
    group_timestamp,
    name
ORDER BY
    group_timestamp ASC;
*/
func getAggQuery(aggArgs *model.AggregatedSignalArgs) (string, []any) {
	if aggArgs == nil {
		return "", nil
	}
	names := make([]string, 0, len(aggArgs.FloatArgs)+len(aggArgs.StringArgs))
	for _, agg := range aggArgs.FloatArgs {
		names = append(names, agg.Name)
	}
	for _, agg := range aggArgs.StringArgs {
		names = append(names, agg.Name)
	}
	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		selectInterval(aggArgs.Interval),
		selectNumberAggs(aggArgs.FloatArgs),
		selectStringAggs(aggArgs.StringArgs),
		qm.WhereIn(nameIn, names),
		withTokenID(aggArgs.TokenID),
		withFromTS(aggArgs.FromTS),
		withToTS(aggArgs.ToTS),
		qm.From(vss.TableName),
		qm.GroupBy(IntervalGroup),
		qm.GroupBy(vss.NameCol),
		qm.OrderBy(groupAsc),
	}
	mods = append(mods, getFilterMods(aggArgs.Filter)...)
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
