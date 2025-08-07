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
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
)

const (
	// IntervalGroup is the column alias for the interval group.
	IntervalGroup     = "group_timestamp"
	AggNumberCol      = "agg_number"
	AggStringCol      = "agg_string"
	aggTableName      = "agg_table"
	tokenIDWhere      = vss.TokenIDCol + " = ?"
	eventSubjectWhere = vss.EventSubjectCol + " = ?"
	nameIn            = vss.NameCol + " IN ?"
	timestampFrom     = vss.TimestampCol + " >= ?"
	timestampTo       = vss.TimestampCol + " < ?"
	sourceWhere       = vss.SourceCol + " = ?"
	sourceIn          = vss.SourceCol + " IN ?"
	groupAsc          = IntervalGroup + " ASC"
	signalTypeCol     = "signal_type"
	signalIndexCol    = "signal_index"

	valueTableDef = signalTypeCol + " UInt8, " + signalIndexCol + " UInt16, " + vss.NameCol + " String"
)

// varibles for the last seen signal query.
const (
	lastSeenName  = "'" + model.LastSeenField + "' AS name"
	numValAsNull  = "NULL AS " + vss.ValueNumberCol
	strValAsNull  = "NULL AS " + vss.ValueStringCol
	locValAsZeros = "CAST((0, 0, 0) AS Tuple(latitude Float64, longitude Float64, hdop Float64)) AS " + vss.ValueLocationCol
	lastSeenTS    = "max(" + vss.TimestampCol + ") AS ts"
)

// Aggregation functions for latest signals.
const (
	latestString    = "argMax(" + vss.ValueStringCol + ", " + vss.TimestampCol + ") as " + vss.ValueStringCol
	latestNumber    = "argMax(" + vss.ValueNumberCol + ", " + vss.TimestampCol + ") as " + vss.ValueNumberCol
	latestLocation  = "argMax(" + vss.ValueLocationCol + ", " + vss.TimestampCol + ") as " + vss.ValueLocationCol
	latestTimestamp = "max(" + vss.TimestampCol + ") as ts"
)

// Aggregation functions for string signals.
const (
	randStringGroup  = "groupArraySample(1, %d)(" + vss.ValueStringCol + ")[1]"
	uniqueGroup      = "arrayStringConcat(groupUniqArray(" + vss.ValueStringCol + "),',')"
	topGroup         = "arrayStringConcat(topK(1, 10)(" + vss.ValueStringCol + "))"
	firstStringGroup = "argMin(" + vss.ValueStringCol + ", " + vss.TimestampCol + ")"
	lastStringGroup  = "argMax(" + vss.ValueStringCol + ", " + vss.TimestampCol + ")"
)

var SourceTranslations = map[string][]string{
	"macaron":  {"dimo/integration/2ULfuC8U9dOqRshZBAi0lMM1Rrx", "0x4c674ddE8189aEF6e3b58F5a36d7438b2b1f6Bc2"},
	"tesla":    {"dimo/integration/26A5Dk3vvvQutjSyF0Jka2DP5lg", "0xc4035Fecb1cc906130423EF05f9C20977F643722"},
	"autopi":   {"dimo/integration/27qftVRWQYpVDcO5DltO5Ojbjxk", "0x5e31bBc786D7bEd95216383787deA1ab0f1c1897"},
	"smartcar": {"dimo/integration/22N2xaPOq2WW2gAHBHd0Ikn4Zob", "0xcd445F4c6bDAD32b68a2939b912150Fe3C88803E"},
	"ruptela":  {"0xF26421509Efe92861a587482100c6d728aBf1CD0"},
	"compass":  {"0x55BF1c27d468314Ea119CF74979E2b59F962295c"},
	"motorq":   {"0x5879B43D88Fa93CE8072d6612cBc8dE93E98CE5d"},
}

// FieldType indicates the type of values in the aggregation. Currently
// there are three types: normal float values, string values, and
// "approximate location" values that are computed from the precise
// location values, in Go.
type FieldType uint8

const (
	// FloatType is the type for rows with numeric values that are in
	// the VSS spec.
	FloatType FieldType = 1
	// StringType is the type for rows with string values.
	StringType FieldType = 2
	// AppLocType is the type for rows needed to compute approximate
	// locations.
	AppLocType FieldType = 3
)

func (t *FieldType) Scan(value any) error {
	w, ok := value.(uint8)
	if !ok {
		return fmt.Errorf("expected value of type uint8, but got type %T", value)
	}

	if w == 0 || w > 3 {
		return fmt.Errorf("invalid value %d for field type", w)
	}

	*t = FieldType(w)
	return nil
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
	if translateSources, ok := SourceTranslations[source]; ok {
		return qm.WhereIn(sourceIn, translateSources)
	}
	return qm.Where(sourceWhere, source)
}

// selectInterval adds a SELECT clause to the query to select the interval group based on the given milliSeconds.
func selectInterval(microSeconds int64, origin time.Time) qm.QueryMod {
	// Newer version of toStartOfInterval with "origin".
	// Requires ClickHouse Cloud 24.10.
	//
	// Note that this new overload seems to have a bug when the interval
	// is an IntervalMilliseconds.
	return qm.Select(fmt.Sprintf("toStartOfInterval(%s, toIntervalMicrosecond(%d), fromUnixTimestamp64Micro(%d)) as %s",
		vss.TimestampCol, microSeconds, origin.UnixMicro(), IntervalGroup))
}

func getFloatSelectExpr(field string) string {
	switch field {
	case vss.FieldCurrentLocationLatitude:
		return vss.ValueLocationCol + ".latitude"
	case vss.FieldCurrentLocationLongitude:
		return vss.ValueLocationCol + ".longitude"
	default:
		return vss.ValueNumberCol
	}
}

func selectNumberAggs(numberAggs []model.FloatSignalArgs, appLocAggs map[model.FloatAggregation]struct{}) qm.QueryMod {
	if len(numberAggs) == 0 && len(appLocAggs) == 0 {
		return qm.Select("NULL AS " + vss.ValueNumberCol)
	}
	// Add a CASE statement for each name and its corresponding aggregation function
	caseStmts := make([]string, 0, len(numberAggs)+2*len(appLocAggs))
	for i, agg := range numberAggs {
		caseStmts = append(caseStmts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, FloatType, signalIndexCol, i, getFloatAggFunc(agg.Agg, getFloatSelectExpr(agg.Name))))
	}
	for i, agg := range model.AllFloatAggregation {
		if _, ok := appLocAggs[agg]; ok {
			caseStmts = append(caseStmts,
				fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, AppLocType, signalIndexCol, 2*i, getFloatAggFunc(agg, getFloatSelectExpr(vss.FieldCurrentLocationLatitude))),
				fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, AppLocType, signalIndexCol, 2*i+1, getFloatAggFunc(agg, getFloatSelectExpr(vss.FieldCurrentLocationLongitude))),
			)
		}
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(caseStmts, " "), AggNumberCol)
	return qm.Select(caseStmt)
}

func selectStringAggs(stringAggs []model.StringSignalArgs) qm.QueryMod {
	if len(stringAggs) == 0 {
		return qm.Select("NULL AS " + vss.ValueStringCol)
	}
	// Add a CASE statement for each name and its corresponding aggregation function
	caseStmts := make([]string, 0, len(stringAggs))
	for i, agg := range stringAggs {
		caseStmts = append(caseStmts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, StringType, signalIndexCol, i, getStringAgg(agg.Agg)))
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(caseStmts, " "), AggStringCol)
	return qm.Select(caseStmt)
}

// returns a string representation of the aggregation function based on the aggregation type.
func getFloatAggFunc(aggType model.FloatAggregation, selectExpr string) string {
	aggStr := "avg(" + selectExpr + ")"
	switch aggType {
	case model.FloatAggregationAvg:
		aggStr = "avg(" + selectExpr + ")"
	case model.FloatAggregationRand:
		seed := time.Now().UnixMilli()
		aggStr = fmt.Sprintf("groupArraySample(1, %d)("+selectExpr+")[1]", seed)
	case model.FloatAggregationMin:
		aggStr = "min(" + selectExpr + ")"
	case model.FloatAggregationMax:
		aggStr = "max(" + selectExpr + ")"
	case model.FloatAggregationMed:
		aggStr = "median(" + selectExpr + ")"
	case model.FloatAggregationFirst:
		aggStr = "argMin(" + selectExpr + ", " + vss.TimestampCol + ")"
	case model.FloatAggregationLast:
		aggStr = "argMax(" + selectExpr + ", " + vss.TimestampCol + ")"
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
	case model.StringAggregationFirst:
		aggStr = firstStringGroup
	case model.StringAggregationLast:
		aggStr = lastStringGroup
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
	signalNames := make([]string, 0, len(latestArgs.SignalNames))
	for name := range latestArgs.SignalNames {
		signalNames = append(signalNames, name)
	}
	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		qm.Select(latestTimestamp),
		qm.Select(latestNumber),
		qm.Select(latestString),
		qm.Select(latestLocation),
		qm.From(vss.TableName),
		qm.Where(tokenIDWhere, latestArgs.TokenID),
		qm.WhereIn(nameIn, signalNames),
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
		qm.Select(locValAsZeros),
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

const FieldCurrentLocation = "currentLocation"

// getAggQuery creates a single query to perform multiple aggregations on the signal data in the same time range and interval.
// This function returns an error if no aggregations are provided.
/*
SELECT
    signal_type,
	signal_id,
	toStartOfInterval(timestamp, toIntervalMicrosecond(60000000), fromUnixTimestamp64Micro(1751274600000000)) AS group_timestamp,
    CASE
        WHEN signal_type = 1 AND signal_index = 0 THEN max(value_number)
        WHEN signal_type = 1 AND signal_index = 1 THEN median(value_number)
        ELSE NULL
    END AS value_number,
    CASE
        WHEN signal_type = 2 AND signal_index = 0 THEN arrayStringConcat(groupUniqArray(value_string),',')
        WHEN signal_type = 2 AND signal_index = 1 THEN groupArraySample(1, 1716404995385)(value_string)[1]
        ELSE NULL
    END AS value_string
FROM
    signal
JOIN
	VALUES(
		'signal_type UInt8, signal_index UInt8, name String',
		(1, 0, 'speed'),
		(1, 1, 'obdRunTime'),
		(2, 0, 'powertrainType'),
		(2, 1, 'powertrainFuelSystemSupportedFuelTypes')
	) AS agg_table
ON
	signal.name = agg_table.name
WHERE
    token_id = 15
    AND timestamp > toDateTime('2024-04-15 09:21:19')
    AND timestamp < toDateTime('2024-04-27 09:21:19')
GROUP BY
    group_timestamp,
    signal_type,
    signal_index
ORDER BY
    group_timestamp ASC,
	signal_type ASC,
	signal_index ASC;
*/
func getAggQuery(aggArgs *model.AggregatedSignalArgs) (string, []any, error) {
	if aggArgs == nil {
		return "", nil, nil
	}

	numAggs := len(aggArgs.FloatArgs) + len(aggArgs.StringArgs) + 2*len(aggArgs.ApproxLocArgs)
	if numAggs == 0 {
		return "", nil, errors.New("no aggregations requested")
	}

	// I can't find documentation for this VALUES syntax anywhere besides GitHub
	// https://github.com/ClickHouse/ClickHouse/issues/5984#issuecomment-513411725
	// You can see the alternatives in the issue and they are ugly.
	valuesArgs := make([]string, 0, numAggs)
	for i, agg := range aggArgs.FloatArgs {
		name := agg.Name
		if name == vss.FieldCurrentLocationLatitude || name == vss.FieldCurrentLocationLongitude {
			name = FieldCurrentLocation
		}
		valuesArgs = append(valuesArgs, aggTableEntry(FloatType, i, name))
	}
	for i, agg := range aggArgs.StringArgs {
		valuesArgs = append(valuesArgs, aggTableEntry(StringType, i, agg.Name))
	}
	for i, agg := range model.AllFloatAggregation {
		if _, ok := aggArgs.ApproxLocArgs[agg]; ok {
			valuesArgs = append(valuesArgs,
				aggTableEntry(AppLocType, 2*i, FieldCurrentLocation),
				aggTableEntry(AppLocType, 2*i+1, FieldCurrentLocation))
		}
	}
	valueTable := fmt.Sprintf("VALUES('%s', %s) as %s ON %s.%s = %s.%s", valueTableDef, strings.Join(valuesArgs, ", "), aggTableName, vss.TableName, vss.NameCol, aggTableName, vss.NameCol)

	floatFilters := []qm.QueryMod{
		// Make sure non-float rows can still get returned.
		qmhelper.Where(signalTypeCol, qmhelper.NEQ, FloatType),
	}

	for i, agg := range aggArgs.FloatArgs {
		// TODO(elffjs): Some duplication here. Also a bit wasteful if
		// there are no filters at all.
		fieldFilters := []qm.QueryMod{
			qmhelper.Where(signalTypeCol, qmhelper.EQ, FloatType),
			qmhelper.Where(signalIndexCol, qmhelper.EQ, i),
		}
		fieldFilters = append(fieldFilters, buildConditionList(agg.Filter, getFloatSelectExpr(agg.Name))...)

		floatFilters = append(floatFilters, qm.Or2(qm.Expr(fieldFilters...)))
	}

	mods := []qm.QueryMod{
		qm.Select(signalTypeCol),
		qm.Select(signalIndexCol),
		selectInterval(aggArgs.Interval, aggArgs.FromTS),
		selectNumberAggs(aggArgs.FloatArgs, aggArgs.ApproxLocArgs),
		selectStringAggs(aggArgs.StringArgs),
		qm.Where(tokenIDWhere, aggArgs.TokenID),
		qm.Where(timestampFrom, aggArgs.FromTS),
		qm.Where(timestampTo, aggArgs.ToTS),
		qm.From(vss.TableName),
		qm.InnerJoin(valueTable),
		qm.GroupBy(IntervalGroup),
		qm.GroupBy(signalTypeCol),
		qm.GroupBy(signalIndexCol),
		qm.OrderBy(groupAsc),
	}
	mods = append(mods, getFilterMods(aggArgs.Filter)...)
	mods = append(mods, qm.Expr(floatFilters...))

	stmt, args := newQuery(mods...)
	return stmt, args, nil
}

func buildConditionList(fil *model.SignalFloatFilter, selectExpr string) []qm.QueryMod {
	if fil == nil {
		return nil
	}

	var mods []qm.QueryMod

	if fil.Eq != nil {
		mods = append(mods, qmhelper.Where(selectExpr, qmhelper.EQ, *fil.Eq))
	}
	if fil.Neq != nil {
		mods = append(mods, qmhelper.Where(selectExpr, qmhelper.NEQ, *fil.Neq))
	}
	if fil.Gt != nil {
		mods = append(mods, qmhelper.Where(selectExpr, qmhelper.GT, *fil.Gt))
	}
	if fil.Lt != nil {
		mods = append(mods, qmhelper.Where(selectExpr, qmhelper.LT, *fil.Lt))
	}
	if fil.Gte != nil {
		mods = append(mods, qmhelper.Where(selectExpr, qmhelper.GTE, *fil.Gte))
	}
	if fil.Lte != nil {
		mods = append(mods, qmhelper.Where(selectExpr, qmhelper.LTE, *fil.Lte))
	}
	if len(fil.NotIn) != 0 {
		mods = append(mods, qm.WhereNotIn(selectExpr+" NOT IN ?", fil.NotIn))
	}
	if len(fil.In) != 0 {
		mods = append(mods, qm.WhereIn(selectExpr+" IN ?", fil.In))
	}

	var orMods []qm.QueryMod
	for _, cond := range fil.Or {
		clauseMods := buildConditionList(cond, selectExpr)
		if len(clauseMods) != 0 {
			if len(orMods) == 0 {
				orMods = append(orMods, qm.Expr(clauseMods...))
			} else {
				orMods = append(orMods, qm.Or2(qm.Expr(clauseMods...)))
			}
		}
	}
	if len(orMods) != 0 {
		mods = append(mods, qm.Expr(orMods...))
	}

	return mods
}

func aggTableEntry(ft FieldType, index int, name string) string {
	return fmt.Sprintf("(%d, %d, '%s')", ft, index, name)
}

func getDistinctQuery(tokenId uint32, filter *model.SignalFilter) (string, []any) {
	mods := []qm.QueryMod{
		qm.Distinct(vss.NameCol),
		qm.From(vss.TableName),
		qm.Where(tokenIDWhere, tokenId),
		qm.OrderBy(vss.NameCol),
	}
	mods = append(mods, getFilterMods(filter)...)
	stmt, args := newQuery(mods...)
	return stmt, args
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

func appendEventFilterMods(mods []qm.QueryMod, filter *model.EventFilter) []qm.QueryMod {
	if filter == nil {
		return mods
	}
	if filter.Name != nil {
		mods = appendStringFilterMod(mods, vss.EventNameCol, filter.Name)
	}
	if filter.Source != nil {
		mods = appendStringFilterMod(mods, vss.EventSourceCol, filter.Source)
	}
	return mods
}

func appendStringFilterMod(mods []qm.QueryMod, field string, filter *model.StringValueFilter) []qm.QueryMod {
	if filter == nil {
		return mods
	}
	if filter.Eq != nil {
		mods = append(mods, qm.Where(field+" = ?", *filter.Eq))
	}
	if filter.Neq != nil {
		mods = append(mods, qm.Where(field+" != ?", *filter.Neq))
	}
	if filter.NotIn != nil {
		mods = append(mods, qm.WhereNotIn(field+" NOT IN (?)", filter.NotIn))
	}
	if filter.In != nil {
		mods = append(mods, qm.WhereIn(field+" IN (?)", filter.In))
	}
	return mods
}
