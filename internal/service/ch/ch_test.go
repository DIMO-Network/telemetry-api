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
	c.dataStartTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
			name: "first string",
			aggArgs: model.AggregatedSignalArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				FromTS:   c.dataStartTime,
				ToTS:     endTs,
				Interval: day.Milliseconds(),
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
				Interval: day.Milliseconds(),
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
		MaxRequestDuration: "1s",
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
		Interval: 28 * day.Milliseconds(),
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
