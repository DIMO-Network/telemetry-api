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
		CLickhouse:                            cfg,
		MaxRequestDuration:                    "1s",
		ManufacturerDeviceLastSeenBucketHours: 3,
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
		expected []model.AggSignal
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
			expected: []model.AggSignal{},
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
						Name: vss.FieldSpeed,
						Agg:  model.FloatAggregationAvg,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime,
					ValueNumber: 4.5,
					Agg:         model.FloatAggregationAvg.String(),
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
						Name: vss.FieldSpeed,
						Agg:  model.FloatAggregationMax,
					},
					{
						Name: vss.FieldSpeed,
						Agg:  model.FloatAggregationMin,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime,
					ValueNumber: 9,
					Agg:         model.FloatAggregationMax.String(),
				},
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime,
					ValueNumber: 0,
					Agg:         model.FloatAggregationMin.String(),
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
						Name: vss.FieldSpeed,
						Agg:  model.FloatAggregationMax,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime,
					ValueNumber: 8.0,
					Agg:         model.FloatAggregationMax.String(),
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
						Name: vss.FieldPowertrainType,
						Agg:  model.StringAggregationUnique,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldPowertrainType,
					Timestamp:   c.dataStartTime,
					ValueString: "value10,value3,value2,value9,value7,value5,value4,value8,value1,value6",
					Agg:         model.StringAggregationUnique.String(),
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
						Name: vss.FieldPowertrainType,
						Agg:  model.StringAggregationTop,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldPowertrainType,
					Timestamp:   c.dataStartTime,
					ValueString: "value2",
					Agg:         model.StringAggregationTop.String(),
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
						Name: vss.FieldSpeed,
						Agg:  model.FloatAggregationFirst,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime,
					ValueNumber: 0,
					Agg:         model.FloatAggregationFirst.String(),
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
						Name: vss.FieldSpeed,
						Agg:  model.FloatAggregationLast,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldSpeed,
					Timestamp:   c.dataStartTime,
					ValueNumber: dataPoints - 1,
					Agg:         model.FloatAggregationLast.String(),
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
						Name: vss.FieldPowertrainType,
						Agg:  model.StringAggregationFirst,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldPowertrainType,
					Timestamp:   c.dataStartTime,
					ValueString: "value1",
					Agg:         model.StringAggregationFirst.String(),
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
						Name: vss.FieldPowertrainType,
						Agg:  model.StringAggregationLast,
					},
				},
			},
			expected: []model.AggSignal{
				{
					Name:        vss.FieldPowertrainType,
					Timestamp:   c.dataStartTime,
					ValueString: fmt.Sprintf("value%d", dataPoints),
					Agg:         model.StringAggregationLast.String(),
				},
			},
		},
	}
	for _, tc := range testCases {
		c.Run(tc.name, func() {
			// Call the GetSignalFloats method
			result, err := c.chService.GetAggregatedSignals(ctx, &tc.aggArgs)
			c.Require().NoError(err)

			c.Require().Len(result, len(tc.expected))

			// Standardize slice order so we can compare position by position.
			slices.SortFunc(result, func(a, b *model.AggSignal) int {
				if cmpName := cmp.Compare(a.Name, a.Name); cmpName != 0 {
					return cmpName
				}
				return cmp.Compare(a.Agg, b.Agg)
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
		sigNames   []string
		latestArgs model.LatestSignalsArgs
		expected   []vss.Signal
	}{
		{
			name: "latest",
			latestArgs: model.LatestSignalsArgs{
				SignalArgs: model.SignalArgs{
					TokenID: 1,
				},
				SignalNames: []string{vss.FieldSpeed},
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
				SignalNames: []string{vss.FieldSpeed},
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
				SignalNames:     []string{},
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
			// Call the GetLatestSignalFloat method
			result, err := c.chService.GetLatestSignals(ctx, &tc.latestArgs)
			c.Require().NoError(err)
			for i, sig := range result {
				c.Require().Equal(tc.expected[i], *sig)
			}
		})
	}
}
func (c *CHServiceTestSuite) TestExecutionTimeout() {
	ctx := context.Background()

	cfg := c.container.Config()

	settings := config.Settings{
		CLickhouse:         cfg,
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

func (c *CHServiceTestSuite) TestGetDeviceActivity() {
	ctx := context.Background()
	lastActive := c.dataStartTime
	testCases := []struct {
		name           string
		vehicleTokenID int
		adManuf        string
		expected       model.DeviceActivity
		expectedError  error
	}{
		{
			name:           "get-bucket-for-last-device-activity",
			vehicleTokenID: 1,
			adManuf:        "AutoPi",
			expected: model.DeviceActivity{
				LastActive: &lastActive,
			},
		},
		{
			name:           "what if that device didnt transmit any data",
			vehicleTokenID: 2,
			adManuf:        "AutoPi",
			expected: model.DeviceActivity{
				LastActive: &lastActive,
			},
			expectedError: fmt.Errorf("no activity recorded for vehicle token ID %d", 2),
		},
	}
	for _, tc := range testCases {
		c.Run(tc.name, func() {
			result, err := c.chService.GetDeviceActivity(ctx, tc.vehicleTokenID, tc.adManuf)
			if tc.expectedError != nil {
				c.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}
			c.Require().Equal(tc.expected, *result)
		})
	}
}

func ref[T any](t T) *T {
	return &t
}