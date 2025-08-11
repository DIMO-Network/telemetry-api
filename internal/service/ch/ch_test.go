package ch

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	chconfig "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/DIMO-Network/clickhouse-infra/pkg/container"
	"github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/stretchr/testify/suite"
)

const (
	day        = time.Hour * 24
	dataPoints = 10
)

type CHServiceTestSuite struct {
	suite.Suite
	dataStartTime time.Time
	chService     *Service
	container     *container.Container
}

func TestCHService(t *testing.T) {
	suite.Run(t, new(CHServiceTestSuite))
}

func (c *CHServiceTestSuite) SetupSuite() {
	ctx := context.Background()
	var err error
	c.container, err = container.CreateClickHouseContainer(ctx, chconfig.Settings{})
	c.Require().NoError(err, "Failed to create clickhouse container")

	db, err := c.container.GetClickhouseAsDB()
	c.Require().NoError(err, "Failed to get clickhouse connection")

	cfg := c.container.Config()

	err = migrations.RunGoose(ctx, []string{"up", "-v"}, db)
	c.Require().NoError(err, "Failed to run migrations")

	settings := config.Settings{
		Clickhouse:           cfg,
		MaxRequestDuration:   "1s",
		DeviceLastSeenBinHrs: 3,
	}
	c.chService, err = NewService(settings)
	c.Require().NoError(err, "Failed to create repository")
	c.dataStartTime = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	c.insertTestData()
}

func (c *CHServiceTestSuite) TearDownSuite() {
	c.container.Terminate(context.Background())
}

func (c *CHServiceTestSuite) TestGetAggSignal() {
	endTs := c.dataStartTime.Add(time.Second * time.Duration(30*dataPoints))
	ctx := context.Background()
	testCases := []struct {
		name     string
		aggArgs  model.AggregatedSignalArgs
		expected []AggSignal
	}{
		{
			name: "no aggs",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
			},
			expected: []AggSignal{},
		},
		{
			name: "average",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 4.5,
				},
			},
		},
		{
			name: "max and min",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationMax,
						Alias: vss.FieldSpeed,
					},
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationMin,
						Alias: vss.FieldSpeed,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 9,
				},
				{
					SignalType:  FloatType,
					SignalIndex: 1,
					Timestamp:   c.dataStartTime,
					ValueNumber: 0,
				},
			},
		},
		{
			name: "max smartcar",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
					Filter: &model.SignalFilter{
						Source: ref("smartcar"),
					},
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationMax,
						Alias: vss.FieldSpeed,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 8.0,
				},
			},
		},
		{
			name: "unique",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				StringArgs: []model.StringSignalArgs{
					{
						Name:  vss.FieldPowertrainType,
						Agg:   model.StringAggregationUnique,
						Alias: vss.FieldPowertrainType,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  StringType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueString: "value10,value3,value2,value9,value7,value5,value4,value8,value1,value6",
				},
			},
		},
		{
			name: "Top autopi",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
					Filter: &model.SignalFilter{
						Source: ref("autopi"),
					},
				},
				FromTS:   c.dataStartTime,
				ToTS:     c.dataStartTime.Add(time.Hour),
				Interval: day.Microseconds(),
				StringArgs: []model.StringSignalArgs{
					{
						Name:  vss.FieldPowertrainType,
						Agg:   model.StringAggregationTop,
						Alias: vss.FieldPowertrainType,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  StringType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueString: "value2",
				},
			},
		},
		{
			name: "first float",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationFirst,
						Alias: vss.FieldSpeed,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 0,
				},
			},
		},
		{
			name: "last float",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationLast,
						Alias: vss.FieldSpeed,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: dataPoints - 1,
				},
			},
		},
		{
			name: "lt filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Lt: ref(float64(5)),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 2,
				},
			},
		},
		{
			name: "gt filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Gt: ref(float64(5)),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 7.5,
				},
			},
		},
		{
			name: "gte filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Gte: ref(float64(5)),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 7,
				},
			},
		},
		{
			name: "lte filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Lte: ref(float64(7)),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 3.5,
				},
			},
		},
		{
			name: "filter for numeric values in set",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							In: []float64{3, 9},
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 6,
				},
			},
		},
		{
			name: "float neq filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Neq: ref(0.0),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 5,
				},
			},
		},
		{
			name: "float neq filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Eq: ref(3.0),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 3,
				},
			},
		},
		{
			name: "float not in filter",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							NotIn: []float64{3, 5},
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 4.625,
				},
			},
		},
		{
			name: "float filters and-ed",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Gte: ref(3.0),
							Neq: ref(6.0),
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 6.0,
				},
			},
		},
		{
			name: "float filters or-ed",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationAvg,
						Alias: vss.FieldSpeed,
						Filter: &model.SignalFloatFilter{
							Or: []*model.SignalFloatFilter{
								{Lt: ref(2.0)},
								{Gte: ref(8.0)},
							},
						},
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 4.5,
				},
			},
		},
		{
			name: "first string",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				StringArgs: []model.StringSignalArgs{
					{
						Name:  vss.FieldPowertrainType,
						Agg:   model.StringAggregationFirst,
						Alias: vss.FieldPowertrainType,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  StringType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueString: "value1",
				},
			},
		},
		{
			name: "last string",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				StringArgs: []model.StringSignalArgs{
					{
						Name:  vss.FieldPowertrainType,
						Agg:   model.StringAggregationLast,
						Alias: vss.FieldPowertrainType,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  StringType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueString: fmt.Sprintf("value%d", dataPoints),
				},
			},
		},
		{
			name: "multiple",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Microseconds(),
				FloatArgs: []model.FloatSignalArgs{
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationMed,
						Alias: "speed1",
						Filter: &model.SignalFloatFilter{
							Or: []*model.SignalFloatFilter{
								{Lte: ref(1.0)},
								{Gte: ref(9.0)},
							},
						},
					},
					{
						Name:  vss.FieldSpeed,
						Agg:   model.FloatAggregationLast,
						Alias: "speed2",
						Filter: &model.SignalFloatFilter{
							Gt:  ref(1.0),
							Lte: ref(6.0),
						},
					},
				},
				StringArgs: []model.StringSignalArgs{
					{
						Name:  vss.FieldPowertrainType,
						Agg:   model.StringAggregationLast,
						Alias: vss.FieldPowertrainType,
					},
				},
			},
			expected: []AggSignal{
				{
					SignalType:  FloatType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueNumber: 1,
				},
				{
					SignalType:  FloatType,
					SignalIndex: 1,
					Timestamp:   c.dataStartTime,
					ValueNumber: 6,
				},
				{
					SignalType:  StringType,
					SignalIndex: 0,
					Timestamp:   c.dataStartTime,
					ValueString: "value10",
				},
			},
		},
	}
	for _, tc := range testCases {
		c.Run(tc.name, func() {
			result, err := c.chService.GetAggregatedSignals(ctx, &tc.aggArgs)
			c.Require().NoError(err)

			c.Require().Len(result, len(tc.expected))

			slices.SortFunc(result, func(a, b *AggSignal) int {
				return cmp.Or(cmp.Compare(a.SignalType, b.SignalType), cmp.Compare(a.SignalIndex, b.SignalIndex))
			})

			for i, sig := range result {
				c.Require().Equal(tc.expected[i], *sig)
			}
		})
	}
}

func (c *CHServiceTestSuite) TestGetLatestSignal() {
	ctx := context.Background()
	testCases := []struct {
		name       string
		latestArgs model.LatestSignalsArgs
		expected   []vss.Signal
	}{
		{
			name: "latest",
			latestArgs: model.LatestSignalsArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				SignalNames: map[string]struct{}{
					vss.FieldSpeed: {},
				},
			},
			expected: []vss.Signal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-1))),
					ValueNumber: 9.0,
				},
			},
		},
		{
			name: "latest smartcar",
			latestArgs: model.LatestSignalsArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
					Filter: &model.SignalFilter{
						Source: ref("smartcar"),
					},
				},
				SignalNames: map[string]struct{}{
					vss.FieldSpeed: {},
				},
			},
			expected: []vss.Signal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-2))),
					ValueNumber: 8.0,
				},
			},
		},
		{
			name: "lastSeen",
			latestArgs: model.LatestSignalsArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				IncludeLastSeen: true,
				SignalNames:     map[string]struct{}{},
			},
			expected: []vss.Signal{
				{
					Name:      model.LastSeenField,
					Timestamp: c.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-1))),
				},
			},
		},
	}
	for _, tc := range testCases {
		c.Run(tc.name, func() {
			result, err := c.chService.GetLatestSignals(ctx, &tc.latestArgs)
			c.Require().NoError(err)
			for i, sig := range result {
				c.Require().Equal(tc.expected[i], *sig)
			}
		})
	}
}

func (c *CHServiceTestSuite) TestGetAvailableSignals() {
	ctx := context.Background()
	c.Run("has signals", func() {
		result, err := c.chService.GetAvailableSignals(ctx, 1, nil)
		c.Require().NoError(err)
		c.Require().Len(result, 2)
		c.Require().Equal([]string{vss.FieldPowertrainType, vss.FieldSpeed}, result)
	})

	c.Run("no signals", func() {
		result, err := c.chService.GetAvailableSignals(ctx, 2, nil)
		c.Require().NoError(err)
		c.Require().Nil(result)
	})

	c.Run("filter signals", func() {
		result, err := c.chService.GetAvailableSignals(ctx, 1, &model.SignalFilter{Source: ref("Unknown")})
		c.Require().NoError(err)
		c.Require().Nil(result)
	})
}

func (c *CHServiceTestSuite) TestExecutionTimeout() {
	ctx := context.Background()

	cfg := c.container.Config()

	settings := config.Settings{
		Clickhouse:         cfg,
		MaxRequestDuration: "1s500ms",
	}
	chService, err := NewService(settings)
	c.Require().NoError(err, "Failed to create repository")

	var delay bool
	err = chService.conn.QueryRow(ctx, "SELECT sleep(3) as delay").Scan(&delay)
	c.Require().Error(err, "Query returned without an error")
	protoErr := &proto.Exception{}
	c.Require().ErrorAs(err, &protoErr, "Query returned without timeout error type: %T", err)
	c.Require().Equalf(TimeoutErrCode, protoErr.Code, "Expected error code %d, got %d, err: %v ", TimeoutErrCode, protoErr.Code, protoErr)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = chService.conn.QueryRow(ctx, "SELECT sleep(2) as delay").Scan(&delay)
	c.Require().Error(err, "Query returned without timeout error")
	c.Require().True(errors.Is(err, context.DeadlineExceeded), "Expected error to be DeadlineExceeded, got %v", err)
}

func (c *CHServiceTestSuite) TestOrginGrouping() {
	ctx := context.Background()
	conn, err := c.container.GetClickHouseAsConn()
	c.Require().NoError(err, "Failed to get clickhouse connection")

	// Set up test data for February 2024
	startTime := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC)

	// Create test signals - one per day in February
	var signals []vss.Signal
	currentTime := startTime
	for currentTime.Before(endTime) {
		signal := vss.Signal{
			Name:        vss.FieldSpeed,
			Timestamp:   currentTime,
			Source:      "test/origin",
			TokenID:     100,
			ValueNumber: 100.0,
		}
		signals = append(signals, signal)
		currentTime = currentTime.Add(24 * time.Hour)
	}

	// Insert signals
	batch, err := conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", vss.TableName))
	c.Require().NoError(err, "Failed to prepare batch")

	for _, sig := range signals {
		err := batch.AppendStruct(&sig)
		c.Require().NoError(err, "Failed to append struct")
	}
	err = batch.Send()
	c.Require().NoError(err, "Failed to send batch")

	// Create aggregation query args
	aggArgs := &model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			TokenID: 100,
		},
		FromTS:   startTime,
		ToTS:     endTime,
		Interval: 28 * day.Microseconds(),
		FloatArgs: []model.FloatSignalArgs{
			{
				Name:  vss.FieldSpeed,
				Agg:   model.FloatAggregationAvg,
				Alias: vss.FieldSpeed,
			},
		},
	}

	// Query signals
	result, err := c.chService.GetAggregatedSignals(ctx, aggArgs)
	c.Require().NoError(err, "Failed to get aggregated signals")

	// We expect exactly one group since we're using a 30-day interval
	c.Require().Len(result, 1, "Expected exactly one group")

	// Verify the group's timestamp matches the start time
	c.Require().Equal(startTime, result[0].Timestamp, "Group timestamp should match start time")

	// Verify the average value (should be 100.0 since all values are 100.0)
	c.Require().Equal(100.0, result[0].ValueNumber, "Unexpected average value")
}

func (c *CHServiceTestSuite) TestGetEvents() {
	ctx := context.Background()
	conn, err := c.container.GetClickHouseAsConn()
	c.Require().NoError(err, "Failed to get clickhouse connection")

	subject := "did:erc721:1:0x0000000000000000000000000000000000000001:42"
	baseTime := time.Date(2024, 6, 12, 12, 0, 0, 0, time.UTC)
	events := []vss.Event{
		{
			Name:       "event.a",
			Source:     "source1",
			Timestamp:  baseTime,
			DurationNs: 1000,
			Subject:    subject,
			Metadata:   `{"foo":"bar"}`,
		},
		{
			Name:       "event.b",
			Source:     "source2",
			Timestamp:  baseTime.Add(5 * time.Minute),
			DurationNs: 2000,
			Subject:    subject,
			Metadata:   "",
		},
		{
			Name:       "event.a",
			Source:     "source2",
			Timestamp:  baseTime.Add(10 * time.Minute),
			DurationNs: 3000,
			Subject:    subject,
			Metadata:   `{"baz":123}`,
		},
		// Event for a different subject (should not be returned)
		{
			Name:       "event.a",
			Source:     "source1",
			Timestamp:  baseTime,
			DurationNs: 999,
			Subject:    "did:erc721:1:0x0000000000000000000000000000000000000001:99",
			Metadata:   `{"should":"not_appear"}`,
		},
	}

	batch, err := conn.PrepareBatch(ctx, "INSERT INTO "+vss.EventTableName)
	c.Require().NoError(err, "Failed to prepare batch")
	for _, event := range events {
		err := batch.AppendStruct(&event)
		c.Require().NoError(err, "Failed to append event struct")
	}
	err = batch.Send()
	c.Require().NoError(err, "Failed to send batch")

	from := baseTime.Add(-time.Minute)
	to := baseTime.Add(15 * time.Minute)

	c.Run("all events for subject and time range", func() {
		result, err := c.chService.GetEvents(ctx, subject, from, to, nil)
		c.Require().NoError(err)
		c.Require().Len(result, 3)
		// Should be ordered by timestamp DESC
		c.Require().Equal("event.a", result[0].Name)
		c.Require().Equal(baseTime.Add(10*time.Minute), result[0].Timestamp)
		c.Require().Equal("source2", result[0].Source)
		c.Require().Equal(uint64(3000), result[0].DurationNs)
		c.Require().Equal(`{"baz":123}`, result[0].Metadata)

		c.Require().Equal("event.b", result[1].Name)
		c.Require().Equal(baseTime.Add(5*time.Minute), result[1].Timestamp)
		c.Require().Equal("source2", result[1].Source)
		c.Require().Equal(uint64(2000), result[1].DurationNs)
		c.Require().Equal("", result[1].Metadata)

		c.Require().Equal("event.a", result[2].Name)
		c.Require().Equal(baseTime, result[2].Timestamp)
		c.Require().Equal("source1", result[2].Source)
		c.Require().Equal(uint64(1000), result[2].DurationNs)
		c.Require().Equal(`{"foo":"bar"}`, result[2].Metadata)
	})

	c.Run("filter by name", func() {
		filter := &model.EventFilter{Name: &model.StringValueFilter{Eq: ref("event.a")}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 2)
		for _, ev := range result {
			c.Require().Equal("event.a", ev.Name)
		}
	})

	c.Run("filter by source", func() {
		filter := &model.EventFilter{Source: &model.StringValueFilter{Eq: ref("source2")}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 2)
		for _, ev := range result {
			c.Require().Equal("source2", ev.Source)
		}
	})

	c.Run("no events in range", func() {
		result, err := c.chService.GetEvents(ctx, subject, baseTime.Add(-2*time.Hour), baseTime.Add(-time.Hour), nil)
		c.Require().NoError(err)
		c.Require().Len(result, 0)
	})

	c.Run("filter by name neq", func() {
		filter := &model.EventFilter{Name: &model.StringValueFilter{Neq: ref("event.a")}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 1)
		c.Require().Equal("event.b", result[0].Name)
	})

	c.Run("filter by name in", func() {
		filter := &model.EventFilter{Name: &model.StringValueFilter{In: []string{"event.a", "event.b"}}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 3)
	})

	c.Run("filter by name notin", func() {
		filter := &model.EventFilter{Name: &model.StringValueFilter{NotIn: []string{"event.a"}}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 1)
		c.Require().Equal("event.b", result[0].Name)
	})

	c.Run("filter by source neq", func() {
		filter := &model.EventFilter{Source: &model.StringValueFilter{Neq: ref("source2")}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 1)
		c.Require().Equal("source1", result[0].Source)
	})

	c.Run("filter by source in", func() {
		filter := &model.EventFilter{Source: &model.StringValueFilter{In: []string{"source1", "source2"}}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 3)
	})

	c.Run("filter by source notin", func() {
		filter := &model.EventFilter{Source: &model.StringValueFilter{NotIn: []string{"source2"}}}
		result, err := c.chService.GetEvents(ctx, subject, from, to, filter)
		c.Require().NoError(err)
		c.Require().Len(result, 1)
		c.Require().Equal("source1", result[0].Source)
	})
}

// insertTestData inserts test data into the clickhouse database.
// it loops for 10 iterations and inserts a 2 signals  with each iteration that have a value of i and a powertrain type of "value"+ n%3+1
// The source is selected from a list of sources in a round robin fashion of sources[i%3].
// The timestamp is incremented by 30 seconds for each iteration.
func (c *CHServiceTestSuite) insertTestData() {
	ctx := context.Background()
	conn, err := c.container.GetClickHouseAsConn()
	c.Require().NoError(err, "Failed to get clickhouse connection")
	testSignal := []vss.Signal{}
	var sources = []string{"dimo/integration/2ULfuC8U9dOqRshZBAi0lMM1Rrx", "dimo/integration/27qftVRWQYpVDcO5DltO5Ojbjxk", "dimo/integration/22N2xaPOq2WW2gAHBHd0Ikn4Zob"}
	for i := range dataPoints {
		numSig := vss.Signal{
			Name:        vss.FieldSpeed,
			Timestamp:   c.dataStartTime.Add(time.Second * time.Duration(30*i)),
			Source:      sources[i%3],
			TokenID:     1,
			ValueNumber: float64(i),
		}

		strSig := vss.Signal{
			Name:        vss.FieldPowertrainType,
			Timestamp:   c.dataStartTime.Add(time.Second * time.Duration(30*i)),
			Source:      sources[i%3],
			TokenID:     1,
			ValueString: fmt.Sprintf("value%d", i+1),
		}
		testSignal = append(testSignal, numSig, strSig)
	}
	// insert the test data into the clickhouse database
	batch, err := conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", vss.TableName))
	c.Require().NoError(err, "Failed to prepare batch")

	for _, sig := range testSignal {
		err := batch.AppendStruct(&sig)
		c.Require().NoError(err, "Failed to append struct")
	}
	err = batch.Send()
	c.Require().NoError(err, "Failed to send batch")
}

func ref[T any](t T) *T {
	return &t
}
