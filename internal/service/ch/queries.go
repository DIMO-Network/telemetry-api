package ch

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/aarondl/sqlboiler/v4/drivers"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/queries/qmhelper"
)

const (
	// IntervalGroup is the column alias for the interval group.
	IntervalGroup     = "group_timestamp"
	AggNumberCol      = "agg_number"
	AggStringCol      = "agg_string"
	AggLocationCol    = "agg_location"
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

// variables for the last seen signal query.
const (
	lastSeenName = "'" + model.LastSeenField + "' AS name"
	numValAsNull = "NULL AS " + vss.ValueNumberCol
	strValAsNull = "NULL AS " + vss.ValueStringCol
	locValAsZero = "CAST(tuple(0, 0, 0), 'Tuple(latitude Float64, longitude Float64, hdop Float64)') AS " + vss.ValueLocationCol

	lastSeenTS = "max(" + vss.TimestampCol + ") AS ts"
)

// Aggregation functions for latest signals.
const (
	latestString    = "argMax(" + vss.ValueStringCol + ", " + vss.TimestampCol + ") as " + vss.ValueStringCol
	latestNumber    = "argMax(" + vss.ValueNumberCol + ", " + vss.TimestampCol + ") as " + vss.ValueNumberCol
	latestLocation  = "argMax(" + vss.ValueLocationCol + ", " + vss.TimestampCol + ") as " + AggLocationCol
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

const (
	locationTupleType = "Tuple(latitude Float64, longitude Float64, hdop Float64)"
	locationZeroTuple = "CAST(tuple(0, 0, 0), '" + locationTupleType + "')"
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
	LocType    FieldType = 4
)

func (t *FieldType) Scan(value any) error {
	w, ok := value.(uint8)
	if !ok {
		return fmt.Errorf("expected value of type uint8, but got type %T", value)
	}

	if w == 0 || w > 4 {
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

func selectNumberAggs(numberAggs []model.FloatSignalArgs, appLocAggs map[model.FloatAggregation]struct{}) qm.QueryMod {
	if len(numberAggs) == 0 && len(appLocAggs) == 0 {
		return qm.Select("NULL AS " + AggNumberCol)
	}
	// Add a CASE statement for each name and its corresponding aggregation function
	caseStmts := make([]string, 0, len(numberAggs)+2*len(appLocAggs))
	for i, agg := range numberAggs {
		caseStmts = append(caseStmts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, FloatType, signalIndexCol, i, getFloatAggFunc(agg.Agg)))
	}
	for i, agg := range model.AllFloatAggregation {
		if _, ok := appLocAggs[agg]; ok {
			caseStmts = append(caseStmts,
				fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, AppLocType, signalIndexCol, 2*i, getFloatAggFunc(agg)),
				fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, AppLocType, signalIndexCol, 2*i+1, getFloatAggFunc(agg)))
		}
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(caseStmts, " "), AggNumberCol)
	return qm.Select(caseStmt)
}

func selectStringAggs(stringAggs []model.StringSignalArgs) qm.QueryMod {
	if len(stringAggs) == 0 {
		return qm.Select("NULL AS " + AggStringCol)
	}
	// Add a CASE statement for each name and its corresponding aggregation function
	caseStmts := make([]string, 0, len(stringAggs))
	for i, agg := range stringAggs {
		caseStmts = append(caseStmts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, StringType, signalIndexCol, i, getStringAgg(agg.Agg)))
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(caseStmts, " "), AggStringCol)
	return qm.Select(caseStmt)
}

func selectLocationAggs(stringAggs []model.LocationSignalArgs) qm.QueryMod {
	if len(stringAggs) == 0 {
		return qm.Select(locationZeroTuple + " AS " + AggLocationCol)
	}
	caseStmts := make([]string, 0, len(stringAggs))
	for i, agg := range stringAggs {
		caseStmts = append(caseStmts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", signalTypeCol, LocType, signalIndexCol, i, getLocationAgg(agg.Agg)))
	}
	caseStmt := fmt.Sprintf("CASE %s ELSE %s END AS %s", strings.Join(caseStmts, " "), locationZeroTuple, AggLocationCol)
	return qm.Select(caseStmt)
}

func getFloatAggFunc(aggType model.FloatAggregation) string {
	return floatAggExpr(vss.ValueNumberCol, vss.TimestampCol, aggType)
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

func getLocationAgg(aggType model.LocationAggregation) string {
	return locationAggExpr(vss.ValueLocationCol, vss.TimestampCol, aggType)
}

// floatAggExpr returns the aggregation expression for a float agg type using the given column exprs (e.g. value_number and timestamp, optionally qualified).
func floatAggExpr(valueNumberExpr, timestampExpr string, aggType model.FloatAggregation) string {
	switch aggType {
	case model.FloatAggregationAvg:
		return "avg(" + valueNumberExpr + ")"
	case model.FloatAggregationRand:
		return fmt.Sprintf("groupArraySample(1, %d)("+valueNumberExpr+")[1]", time.Now().UnixMilli())
	case model.FloatAggregationMin:
		return "min(" + valueNumberExpr + ")"
	case model.FloatAggregationMax:
		return "max(" + valueNumberExpr + ")"
	case model.FloatAggregationMed:
		return "median(" + valueNumberExpr + ")"
	case model.FloatAggregationFirst:
		return "argMin(" + valueNumberExpr + ", " + timestampExpr + ")"
	case model.FloatAggregationLast:
		return "argMax(" + valueNumberExpr + ", " + timestampExpr + ")"
	default:
		return "avg(" + valueNumberExpr + ")"
	}
}

// locationAggExpr returns the aggregation expression for a location agg type using the given column exprs.
func locationAggExpr(valueLocationExpr, timestampExpr string, aggType model.LocationAggregation) string {
	switch aggType {
	case model.LocationAggregationAvg:
		return "CAST(tuple(avg(" + valueLocationExpr + ".latitude), avg(" + valueLocationExpr + ".longitude), avg(" + valueLocationExpr + ".hdop)), '" + locationTupleType + "')"
	case model.LocationAggregationRand:
		return fmt.Sprintf("groupArraySample(1, %d)("+valueLocationExpr+")[1]", time.Now().UnixMilli())
	case model.LocationAggregationFirst:
		return "argMin(" + valueLocationExpr + ", " + timestampExpr + ")"
	case model.LocationAggregationLast:
		return "argMax(" + valueLocationExpr + ", " + timestampExpr + ")"
	default:
		return "argMin(" + valueLocationExpr + ", " + timestampExpr + ")"
	}
}

// batchFloatCaseExprQualified is like batchFloatCaseExpr but with qualified column names (alias.value_number, alias.timestamp).
func batchFloatCaseExprQualified(alias string, floatArgs []model.FloatSignalArgs) string {
	return batchFloatCaseExprWithAlias(alias, floatArgs)
}

func batchFloatCaseExprWithAlias(alias string, floatArgs []model.FloatSignalArgs) string {
	if len(floatArgs) == 0 {
		return "NULL AS " + AggNumberCol
	}
	typeCol, indexCol := signalTypeCol, signalIndexCol
	vNum, ts := vss.ValueNumberCol, vss.TimestampCol
	if alias != "" {
		typeCol, indexCol = alias+"."+signalTypeCol, alias+"."+signalIndexCol
		vNum, ts = alias+"."+vss.ValueNumberCol, alias+"."+vss.TimestampCol
	}
	parts := make([]string, 0, len(floatArgs))
	for i, agg := range floatArgs {
		expr := floatAggExpr(vNum, ts, agg.Agg)
		parts = append(parts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", typeCol, FloatType, indexCol, i, expr))
	}
	return fmt.Sprintf("CASE %s ELSE NULL END AS %s", strings.Join(parts, " "), AggNumberCol)
}

// batchLocationCaseExprQualified is like batchLocationCaseExpr with qualified column names.
func batchLocationCaseExprQualified(alias string, locationArgs []model.LocationSignalArgs) string {
	return batchLocationCaseExprWithAlias(alias, locationArgs)
}

func batchLocationCaseExprWithAlias(alias string, locationArgs []model.LocationSignalArgs) string {
	zeroLoc := locationZeroTuple + " AS " + AggLocationCol
	if len(locationArgs) == 0 {
		return zeroLoc
	}
	typeCol, indexCol := signalTypeCol, signalIndexCol
	vLoc, ts := vss.ValueLocationCol, vss.TimestampCol
	if alias != "" {
		typeCol, indexCol = alias+"."+signalTypeCol, alias+"."+signalIndexCol
		vLoc, ts = alias+"."+vss.ValueLocationCol, alias+"."+vss.TimestampCol
	}
	parts := make([]string, 0, len(locationArgs))
	for i, agg := range locationArgs {
		expr := locationAggExpr(vLoc, ts, agg.Agg)
		parts = append(parts, fmt.Sprintf("WHEN %s = %d AND %s = %d THEN %s", typeCol, LocType, indexCol, i, expr))
	}
	return fmt.Sprintf("CASE %s ELSE %s END AS %s", strings.Join(parts, " "), locationZeroTuple, AggLocationCol)
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

	locationSignalNames := make([]string, 0, len(latestArgs.LocationSignalNames))
	for name := range latestArgs.LocationSignalNames {
		locationSignalNames = append(locationSignalNames, name)
	}

	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		qm.Select(latestTimestamp),
		qm.Select(latestNumber),
		qm.Select(latestString),
		qm.Select(latestLocation),
		qm.From(vss.TableName),
		qm.Where(tokenIDWhere, latestArgs.TokenID),
		qm.Expr(
			qm.WhereIn(nameIn, signalNames),
			qm.Or2(
				qm.Expr(
					qm.WhereIn(nameIn, locationSignalNames),
					qm.Expr(
						qmhelper.Where(vss.ValueLocationCol+".latitude", qmhelper.NEQ, 0),
						qm.Or2(qmhelper.Where(vss.ValueLocationCol+".longitude", qmhelper.NEQ, 0)),
					),
				),
			),
		),
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
		qm.Select(locValAsZero),
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

	numAggs := len(aggArgs.FloatArgs) + len(aggArgs.StringArgs) + 2*len(aggArgs.ApproxLocArgs) + len(aggArgs.LocationArgs)
	if numAggs == 0 {
		return "", nil, errors.New("no aggregations requested")
	}

	// I can't find documentation for this VALUES syntax anywhere besides GitHub
	// https://github.com/ClickHouse/ClickHouse/issues/5984#issuecomment-513411725
	// You can see the alternatives in the issue and they are ugly.
	valuesArgs := make([]string, 0, numAggs)
	for i, agg := range aggArgs.FloatArgs {
		valuesArgs = append(valuesArgs, aggTableEntry(FloatType, i, agg.Name))
	}
	for i, agg := range aggArgs.StringArgs {
		valuesArgs = append(valuesArgs, aggTableEntry(StringType, i, agg.Name))
	}
	for i, agg := range model.AllFloatAggregation {
		if _, ok := aggArgs.ApproxLocArgs[agg]; ok {
			valuesArgs = append(valuesArgs,
				aggTableEntry(AppLocType, 2*i, vss.FieldCurrentLocationLatitude),
				aggTableEntry(AppLocType, 2*i+1, vss.FieldCurrentLocationLongitude))
		}
	}
	for i, agg := range aggArgs.LocationArgs {
		valuesArgs = append(valuesArgs, aggTableEntry(LocType, i, agg.Name))
	}
	valueTable := fmt.Sprintf("VALUES('%s', %s) as %s ON %s.%s = %s.%s", valueTableDef, strings.Join(valuesArgs, ", "), aggTableName, vss.TableName, vss.NameCol, aggTableName, vss.NameCol)

	var perSignalFilters []qm.QueryMod

	if len(aggArgs.FloatArgs) != 0 {
		// These are for float fields. One sub-Expr per field.
		var innerFloatFilters []qm.QueryMod

		for i, agg := range aggArgs.FloatArgs {
			fieldFilters := []qm.QueryMod{
				qmhelper.Where(signalIndexCol, qmhelper.EQ, i),
			}
			fieldFilters = append(fieldFilters, buildFloatConditionList(agg.Filter)...)

			// It's okay to also use Or2 for the first entry: it's simply ignored.
			innerFloatFilters = append(innerFloatFilters, qm.Or2(qm.Expr(fieldFilters...)))
		}

		perSignalFilters = append(perSignalFilters, qm.Or2(
			qm.Expr(
				qmhelper.Where(signalTypeCol, qmhelper.EQ, FloatType),
				qm.Expr(innerFloatFilters...),
			),
		))
	}

	if len(aggArgs.StringArgs) != 0 {
		perSignalFilters = append(perSignalFilters, qm.Or2(qmhelper.Where(signalTypeCol, qmhelper.EQ, StringType)))
	}

	if len(aggArgs.ApproxLocArgs) != 0 {
		perSignalFilters = append(perSignalFilters, qm.Or2(qmhelper.Where(signalTypeCol, qmhelper.EQ, AppLocType)))
	}

	if len(aggArgs.LocationArgs) != 0 {
		var innerLocationFilters []qm.QueryMod

		for i, agg := range aggArgs.LocationArgs {
			fieldFilters := []qm.QueryMod{
				qmhelper.Where(signalIndexCol, qmhelper.EQ, i),
			}
			fieldFilters = append(fieldFilters, buildLocationConditionList(agg.Filter)...)

			innerLocationFilters = append(innerLocationFilters, qm.Or2(qm.Expr(fieldFilters...)))
		}

		perSignalFilters = append(perSignalFilters, qm.Or2(
			qm.Expr(
				qmhelper.Where(signalTypeCol, qmhelper.EQ, LocType),
				qm.Expr(
					qmhelper.Where(vss.ValueLocationCol+".latitude", qmhelper.NEQ, 0),
					qm.Or2(qmhelper.Where(vss.ValueLocationCol+".longitude", qmhelper.NEQ, 0)),
				),
				qm.Expr(innerLocationFilters...),
			),
		))
	}

	mods := []qm.QueryMod{
		qm.Select(signalTypeCol),
		qm.Select(signalIndexCol),
		selectInterval(aggArgs.Interval, aggArgs.FromTS),
		selectNumberAggs(aggArgs.FloatArgs, aggArgs.ApproxLocArgs),
		selectStringAggs(aggArgs.StringArgs),
		selectLocationAggs(aggArgs.LocationArgs),
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
	mods = append(mods, qm.Expr(perSignalFilters...)) // Parenthesization is very important here!

	stmt, args := newQuery(mods...)
	return stmt, args, nil
}

// getBatchAggQuery returns a query that computes the same aggregations as getAggQuery for multiple
// time ranges (segments) in one round-trip. Only FloatArgs and LocationArgs are supported.
// Result columns: seg_idx (Int32), signal_type, signal_index, value_number, value_string, value_location.
func getBatchAggQuery(tokenID uint32, ranges []TimeRange, globalFrom, globalTo time.Time, floatArgs []model.FloatSignalArgs, locationArgs []model.LocationSignalArgs) (string, []any, error) {
	if len(ranges) == 0 {
		return "", nil, errors.New("no ranges for batch agg")
	}
	if len(floatArgs) == 0 && len(locationArgs) == 0 {
		return "", nil, errors.New("no aggregations for batch agg")
	}
	valueTable := buildBatchAggValueTable(floatArgs, locationArgs)
	multiIf := buildSegmentIndexMultiIf(vss.TimestampCol, len(ranges))
	args := make([]any, 0, 2*len(ranges)+3)
	for _, r := range ranges {
		args = append(args, r.From, r.To)
	}
	args = append(args, tokenID, globalFrom, globalTo)
	inner := buildBatchAggInner(valueTable, multiIf)
	outer := buildBatchAggOuter(inner, floatArgs, locationArgs)
	return outer, args, nil
}

func buildBatchAggValueTable(floatArgs []model.FloatSignalArgs, locationArgs []model.LocationSignalArgs) string {
	valuesArgs := make([]string, 0, len(floatArgs)+len(locationArgs))
	for i, agg := range floatArgs {
		valuesArgs = append(valuesArgs, aggTableEntry(FloatType, i, agg.Name))
	}
	for i, agg := range locationArgs {
		valuesArgs = append(valuesArgs, aggTableEntry(LocType, i, agg.Name))
	}
	return fmt.Sprintf("VALUES('%s', %s) as %s ON %s.%s = %s.%s", valueTableDef, strings.Join(valuesArgs, ", "), aggTableName, vss.TableName, vss.NameCol, aggTableName, vss.NameCol)
}

func buildBatchAggInner(valueTable, multiIf string) string {
	selectList := multiIf + ", " + signalTypeCol + ", " + signalIndexCol + ", " + vss.TimestampCol + ", " + vss.ValueNumberCol + ", " + vss.ValueStringCol + ", " + vss.ValueLocationCol
	return "SELECT " + selectList + " FROM " + vss.TableName + " INNER JOIN " + valueTable +
		" WHERE " + tokenIDWhere + " AND " + vss.TimestampCol + " >= ? AND " + vss.TimestampCol + " < ?"
}

func buildBatchAggOuter(inner string, floatArgs []model.FloatSignalArgs, locationArgs []model.LocationSignalArgs) string {
	const alias = "batch_inner"
	selectList := alias + ".seg_idx, " + alias + "." + signalTypeCol + ", " + alias + "." + signalIndexCol + ", " +
		batchFloatCaseExprQualified(alias, floatArgs) + ", NULL AS " + AggStringCol + ", " + batchLocationCaseExprQualified(alias, locationArgs)
	groupBy := alias + ".seg_idx, " + alias + "." + signalTypeCol + ", " + alias + "." + signalIndexCol
	return "SELECT " + selectList + " FROM (" + inner + ") AS " + alias + " WHERE " + alias + ".seg_idx >= 0 GROUP BY " + groupBy
}

func buildFloatConditionList(fil *model.SignalFloatFilter) []qm.QueryMod {
	if fil == nil {
		return nil
	}

	var mods []qm.QueryMod

	if fil.Eq != nil {
		mods = append(mods, qmhelper.Where(vss.ValueNumberCol, qmhelper.EQ, *fil.Eq))
	}
	if fil.Neq != nil {
		mods = append(mods, qmhelper.Where(vss.ValueNumberCol, qmhelper.NEQ, *fil.Neq))
	}
	if fil.Gt != nil {
		mods = append(mods, qmhelper.Where(vss.ValueNumberCol, qmhelper.GT, *fil.Gt))
	}
	if fil.Lt != nil {
		mods = append(mods, qmhelper.Where(vss.ValueNumberCol, qmhelper.LT, *fil.Lt))
	}
	if fil.Gte != nil {
		mods = append(mods, qmhelper.Where(vss.ValueNumberCol, qmhelper.GTE, *fil.Gte))
	}
	if fil.Lte != nil {
		mods = append(mods, qmhelper.Where(vss.ValueNumberCol, qmhelper.LTE, *fil.Lte))
	}
	if len(fil.NotIn) != 0 {
		mods = append(mods, qm.WhereNotIn(vss.ValueNumberCol+" NOT IN ?", fil.NotIn))
	}
	if len(fil.In) != 0 {
		mods = append(mods, qm.WhereIn(vss.ValueNumberCol+" IN ?", fil.In))
	}

	var orMods []qm.QueryMod
	for _, cond := range fil.Or {
		clauseMods := buildFloatConditionList(cond)
		if len(clauseMods) != 0 {
			orMods = append(orMods, qm.Or2(qm.Expr(clauseMods...)))
		}
	}

	if len(orMods) != 0 {
		mods = append(mods, qm.Expr(orMods...))
	}

	return mods
}

func buildLocationConditionList(fil *model.SignalLocationFilter) []qm.QueryMod {
	if fil == nil {
		return nil
	}

	var mods []qm.QueryMod

	// This will not work well if points at at the edges of the coordinate system:
	// for example, around the antimeridian.
	if len(fil.InPolygon) != 0 {
		// TODO(elffjs): Can the ClickHouse driver handle this list assembly for us?
		var interp []any
		for _, pt := range fil.InPolygon {
			// Important: ClickHouse thinks of these as (x, y), so longitude goes first.
			interp = append(interp, pt.Longitude, pt.Latitude)
		}

		// ClickHouse function:
		// https://clickhouse.com/docs/sql-reference/functions/geo/coordinates#pointinpolygon
		mods = append(mods, qm.Where(
			"pointInPolygon(("+vss.ValueLocationCol+".longitude, "+vss.ValueLocationCol+".latitude), ["+repeatWithSep("(?, ?)", len(fil.InPolygon), ", ")+"])",
			interp...,
		))
	}

	// ClickHouse function, which returns meters:
	// https://clickhouse.com/docs/sql-reference/functions/geo/coordinates#geodistance
	if fil.InCircle != nil {
		mods = append(mods, qm.Where(
			"geoDistance(?, ?, "+vss.ValueLocationCol+".longitude, "+vss.ValueLocationCol+".latitude) <= ?",
			fil.InCircle.Center.Longitude, fil.InCircle.Center.Latitude, kilometersToMeters(fil.InCircle.Radius),
		))
	}

	return mods
}

func kilometersToMeters(d float64) float64 {
	return 1000 * d
}

func repeatWithSep(s string, count int, sep string) string {
	if count == 0 {
		return ""
	}
	// Don't actually need to special case this, since strings.Repeat(s, 0) is "".
	// We do avoid a concatenation, though.
	if count == 1 {
		return s
	}
	return strings.Repeat(s+sep, count-1) + s
}

func aggTableEntry(ft FieldType, index int, name string) string {
	escaped := strings.ReplaceAll(name, "'", "\\'")
	return fmt.Sprintf("(%d, %d, '%s')", ft, index, escaped)
}

// buildSegmentIndexMultiIf returns "multiIf( (tsCol >= ? AND tsCol < ?), 0, ..., -1) AS seg_idx" for n ranges.
func buildSegmentIndexMultiIf(timestampCol string, nRanges int) string {
	if nRanges == 0 {
		return "toInt32(-1) AS seg_idx"
	}
	cond := "(" + timestampCol + " >= ? AND " + timestampCol + " < ?)"
	parts := make([]string, 0, nRanges)
	for i := 0; i < nRanges; i++ {
		parts = append(parts, cond+", "+fmt.Sprintf("%d", i))
	}
	return "multiIf(" + strings.Join(parts, ", ") + ", -1) AS seg_idx"
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

// select Count(*), max(timestamp), min(timestamp), name, from signal where token_id = '39718' GROUP BY name
func getSignalSummariesQuery(tokenId uint32, filter *model.SignalFilter) (string, []any) {
	mods := []qm.QueryMod{
		qm.Select(vss.NameCol),
		qm.Select("COUNT(*)"),
		qm.Select("MIN(" + vss.TimestampCol + ")"),
		qm.Select("MAX(" + vss.TimestampCol + ")"),
		qm.From(vss.TableName),
		qm.Where(tokenIDWhere, tokenId),
		qm.GroupBy(vss.NameCol),
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
		mods = append(mods, stringFilterMod(vss.EventNameCol, filter.Name)...)
	}
	if filter.Source != nil {
		mods = append(mods, stringFilterMod(vss.EventSourceCol, filter.Source)...)
	}
	if filter.Tags != nil {
		mods = append(mods, stringArrayFilterMod(filter.Tags, vss.EventTagsCol)...)
	}
	return mods
}

// getEventSummariesQuery returns a query that summarizes events by name for a subject (all time).
func getEventSummariesQuery(subject string) (string, []any) {
	mods := []qm.QueryMod{
		qm.Select(vss.EventNameCol + " AS name"),
		qm.Select("count(*) AS count"),
		qm.Select("MIN(" + vss.EventTimestampCol + ") AS first_seen"),
		qm.Select("MAX(" + vss.EventTimestampCol + ") AS last_seen"),
		qm.From(vss.EventTableName),
		qm.Where(eventSubjectWhere, subject),
		qm.GroupBy(vss.EventNameCol),
		qm.OrderBy(vss.EventNameCol),
	}
	return newQuery(mods...)
}

// getEventCountsQuery returns a query that counts events by name in the given time range.
// If eventNames is non-nil and non-empty, only those names are included; otherwise all names.
func getEventCountsQuery(subject string, from, to time.Time, eventNames []string) (string, []any) {
	mods := []qm.QueryMod{
		qm.Select(vss.EventNameCol + " AS name"),
		qm.Select("count(*) AS count"),
		qm.From(vss.EventTableName),
		qm.Where(eventSubjectWhere, subject),
		qm.Where(vss.EventTimestampCol+" >= ?", from),
		qm.Where(vss.EventTimestampCol+" < ?", to),
		qm.GroupBy(vss.EventNameCol),
	}
	if len(eventNames) > 0 {
		mods = append(mods, qm.WhereIn(vss.EventNameCol+" IN ?", eventNames))
	}
	return newQuery(mods...)
}

// TimeRange is a [From, To) interval for batch event count queries.
type TimeRange struct {
	From, To time.Time
}

// getEventCountsForRangesQuery returns a query that counts events by name per segment index
// for multiple time ranges in one round-trip. Ranges are (from, to) exclusive on to.
// Result columns: seg_idx (Int32), name (String), count (UInt64).
func getEventCountsForRangesQuery(subject string, ranges []TimeRange, eventNames []string) (string, []any) {
	if len(ranges) == 0 {
		return eventCountsForRangesEmptyQuery(), nil
	}
	multiIf := buildSegmentIndexMultiIf(vss.EventTimestampCol, len(ranges))
	args := buildEventCountsForRangesArgs(ranges, subject, eventNames)
	innerSelect := buildEventCountsForRangesInner(multiIf, eventNames)
	stmt := "SELECT seg_idx, name, count(*) AS count FROM (" + innerSelect + ") WHERE seg_idx >= 0 GROUP BY seg_idx, name"
	return stmt, args
}

func eventCountsForRangesEmptyQuery() string {
	return "SELECT toInt32(-1) AS seg_idx, '' AS name, toUInt64(0) AS count FROM " + vss.EventTableName + " WHERE 0"
}

func buildEventCountsForRangesArgs(ranges []TimeRange, subject string, eventNames []string) []any {
	args := make([]any, 0, 2*len(ranges)+1+len(eventNames))
	for _, r := range ranges {
		args = append(args, r.From, r.To)
	}
	args = append(args, subject)
	for _, n := range eventNames {
		args = append(args, n)
	}
	return args
}

func buildEventCountsForRangesInner(multiIf string, eventNames []string) string {
	inner := "SELECT " + multiIf + ", " + vss.EventNameCol + " AS name FROM " + vss.EventTableName + " PREWHERE " + eventSubjectWhere
	if len(eventNames) > 0 {
		placeholders := make([]string, len(eventNames))
		for i := range eventNames {
			placeholders[i] = "?"
		}
		inner += " AND " + vss.EventNameCol + " IN (" + strings.Join(placeholders, ", ") + ")"
	}
	return inner
}

func stringFilterMod(field string, filter *model.StringValueFilter) []qm.QueryMod {
	var newMods []qm.QueryMod
	if filter == nil {
		return nil
	}
	if filter.Eq != nil {
		newMods = append(newMods, qm.Where(field+" = ?", *filter.Eq))
	}
	if filter.Neq != nil {
		newMods = append(newMods, qm.Where(field+" != ?", *filter.Neq))
	}
	if filter.NotIn != nil {
		newMods = append(newMods, qm.WhereNotIn(field+" NOT IN (?)", filter.NotIn))
	}
	if filter.In != nil {
		newMods = append(newMods, qm.WhereIn(field+" IN (?)", filter.In))
	}

	for _, cond := range filter.Or {
		clauseMods := stringFilterMod(field, cond)
		if len(clauseMods) != 0 {
			newMods = append(newMods, qm.Or2(qm.Expr(clauseMods...)))
		}
	}

	if len(filter.Or) != 0 {
		// if we have an Or wrap the expression so it doesn't get ordered with the parent where clause
		newMods = []qm.QueryMod{qm.Expr(newMods...)}
	}

	return newMods
}

func stringArrayFilterMod(filter *model.StringArrayFilter, field string) []qm.QueryMod {
	var newMods []qm.QueryMod
	if filter == nil {
		return newMods
	}

	if len(filter.ContainsAny) != 0 {
		newMods = append(newMods, qm.Where("hasAny("+field+", ?)", filter.ContainsAny))
	}
	if len(filter.ContainsAll) != 0 {
		newMods = append(newMods, qm.Where("hasAll("+field+", ?)", filter.ContainsAll))
	}
	if len(filter.NotContainsAny) != 0 {
		newMods = append(newMods, qm.Where("NOT hasAny("+field+", ?)", filter.NotContainsAny))
	}
	if len(filter.NotContainsAll) != 0 {
		newMods = append(newMods, qm.Where("NOT hasAll("+field+", ?)", filter.NotContainsAll))
	}

	for _, cond := range filter.Or {
		clauseMods := stringArrayFilterMod(cond, field)
		if len(clauseMods) != 0 {
			newMods = append(newMods, qm.Or2(qm.Expr(clauseMods...)))
		}
	}

	if len(filter.Or) != 0 {
		newMods = []qm.QueryMod{qm.Expr(newMods...)}
	}

	return newMods
}
