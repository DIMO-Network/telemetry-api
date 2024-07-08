package ch

import (
	"errors"
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
	AggCol        = "agg"
	aggTableName  = "agg_table"
	tokenIDWhere  = vss.TokenIDCol + " = ?"
	nameIn        = vss.NameCol + " IN ?"
	timestampFrom = vss.TimestampCol + " >= ?"
	timestampTo   = vss.TimestampCol + " < ?"
	sourceWhere   = vss.SourceCol + " = ?"
	groupAsc      = IntervalGroup + " ASC"
	valueTableDef = "name String, agg String"
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
	avgGroup       = "avg(" + vss.ValueNumberCol + ")"
	randFloatGroup = "groupArraySample(1, %d)(" + vss.ValueNumberCol + ")[1]"
	minGroup       = "min(" + vss.ValueNumberCol + ")"
	maxGroup       = "max(" + vss.ValueNumberCol + ")"
	medGroup       = "median(" + vss.ValueNumberCol + ")"
	firstGroup     = "argMin(" + vss.ValueNumberCol + ", " + vss.TimestampCol + ")"
	lastGroup      = "argMax(" + vss.ValueNumberCol + ", " + vss.TimestampCol + ")"
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

// withSource adds a WHERE clause to the query to filter by Source.
// Example: 'WHERE Source = ?'.
func withSource(source string) qm.QueryMod {
	// TODO: remove this logic when we move to storing the device address as source.
	if translateSource, ok := SourceTranslations[source]; ok {
		return qm.Where(sourceWhere, translateSource)
	}
	return qm.Where(sourceWhere, source)
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
		caseStmts[i] = fmt.Sprintf("WHEN %s = '%s' AND %s = '%s' THEN %s", vss.NameCol, numberAggs[i].Name, AggCol, numberAggs[i].Agg, getFloatAggFunc(numberAggs[i].Agg))
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
		caseStmts[i] = fmt.Sprintf("WHEN %s = '%s' AND %s = '%s' THEN %s", vss.NameCol, stringAggs[i].Name, AggCol, stringAggs[i].Agg, getStringAgg(stringAggs[i].Agg))
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
		aggStr = fmt.Sprintf(randFloatGroup, seed)
	case model.FloatAggregationMin:
		aggStr = minGroup
	case model.FloatAggregationMax:
		aggStr = maxGroup
	case model.FloatAggregationMed:
		aggStr = medGroup
	case model.FloatAggregationFirst:
		aggStr = firstGroup
	case model.FloatAggregationLast:
		aggStr = lastGroup
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

// getLatestQuery creates a query to get the latest signal value for each signal names
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
GROUP BY
  name
*/
func getLatestQuery(latestArgs *model.LatestSignalsArgs) (string, []any) {
	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		qm.Select(latestTimestamp),
		qm.Select(latestNumber),
		qm.Select(latestString),
		qm.From(vss.TableName),
		qm.Where(tokenIDWhere, latestArgs.TokenID),
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
		qm.Where(tokenIDWhere, sigArgs.TokenID),
	}
	mods = append(mods, getFilterMods(sigArgs.Filter)...)
	return newQuery(mods...)
}

// unionAll creates a UNION ALL statement from the given statements and arguments.
func unionAll(allStatements []string, allArgs [][]any) (string, []any) {
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
// This function returns an error if no aggregations are provided.
/*
SELECT
    `name`,
    toStartOfInterval(timestamp, toIntervalMillisecond(30000)) AS group_timestamp,
    CASE
        WHEN name = 'speed' AND agg = 'MAX' THEN max(value_number)
        WHEN name = 'obdRunTime' AND agg = 'MEDIAN' THEN median(value_number)
        ELSE NULL
    END AS value_number,
    CASE
        WHEN name = 'powertrainType' AND agg = 'UNIQUE' THEN arrayStringConcat(groupUniqArray(value_string),',')
        WHEN name = 'powertrainFuelSystemSupportedFuelTypes' AND agg = 'RAND' THEN groupArraySample(1, 1716404995385)(value_string)[1]
        ELSE NULL
    END AS value_string
FROM
    `signal`
JOIN
	VALUES(
		'name String, agg String',
		('speed, 'MAX'),
		('obdRunTime', 'MEDIAN'),
		('powertrainType', 'UNIQUE'),
		('powertrainFuelSystemSupportedFuelTypes', 'RAND')
	) AS agg_table
ON
	signal.name = agg_table.name
WHERE
    token_id = 15
    AND timestamp > toDateTime('2024-04-15 09:21:19')
    AND timestamp < toDateTime('2024-04-27 09:21:19')
GROUP BY
    group_timestamp,
    name
ORDER BY
    group_timestamp ASC,
	name ASC,
	agg ASC;
*/
func getAggQuery(aggArgs *model.AggregatedSignalArgs) (string, []any, error) {
	if aggArgs == nil {
		return "", nil, nil
	}

	numAggs := len(aggArgs.FloatArgs) + len(aggArgs.StringArgs)
	if numAggs == 0 {
		return "", nil, errors.New("no aggregations requested")
	}

	// I can't find documentation for this VALUES syntax anywhere besides GitHub
	// https://github.com/ClickHouse/ClickHouse/issues/5984#issuecomment-513411725
	// You can see the alternatives in the issue and they are ugly.
	valuesArgs := make([]string, 0, numAggs)
	for _, agg := range aggArgs.FloatArgs {
		valuesArgs = append(valuesArgs, fmt.Sprintf("('%s', '%s')", agg.Name, agg.Agg))
	}
	for _, agg := range aggArgs.StringArgs {
		valuesArgs = append(valuesArgs, fmt.Sprintf("('%s', '%s')", agg.Name, agg.Agg))
	}
	valueTable := fmt.Sprintf("VALUES('%s', %s) as %s ON %s.%s = %s.%s", valueTableDef, strings.Join(valuesArgs, ", "), aggTableName, vss.TableName, vss.NameCol, aggTableName, vss.NameCol)

	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		qm.Select(AggCol),
		selectInterval(aggArgs.Interval),
		selectNumberAggs(aggArgs.FloatArgs),
		selectStringAggs(aggArgs.StringArgs),
		qm.Where(tokenIDWhere, aggArgs.TokenID),
		qm.Where(timestampFrom, aggArgs.FromTS),
		qm.Where(timestampTo, aggArgs.ToTS),
		qm.From(vss.TableName),
		qm.InnerJoin(valueTable),
		qm.GroupBy(IntervalGroup),
		qm.GroupBy(vss.NameCol),
		qm.GroupBy(AggCol),
		qm.OrderBy(groupAsc),
	}
	mods = append(mods, getFilterMods(aggArgs.Filter)...)
	stmt, args := newQuery(mods...)
	return stmt, args, nil
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
