package repositories_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/clickhouseinfra"
	"github.com/DIMO-Network/model-garage/pkg/migrations"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

const (
	day        = time.Hour * 24
	dataPoints = 10
)

type RepositoryTestSuite struct {
	suite.Suite
	dataStartTime time.Time
	repo          *repositories.Repository
	container     *clickhouseinfra.Container
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (r *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	var err error
	r.container, err = clickhouseinfra.CreateClickHouseContainer(ctx, "", "")
	r.Require().NoError(err, "Failed to create clickhouse container")

	db, err := clickhouseinfra.GetClickhouseAsDB(ctx, r.container.ClickHouseContainer)
	r.Require().NoError(err, "Failed to get clickhouse connection")

	err = migrations.RunGoose(ctx, []string{"up", "-v"}, db)
	r.Require().NoError(err, "Failed to run migrations")

	testLogger := zerolog.New(os.Stderr)

	host, err := r.container.Host(ctx)
	r.Require().NoError(err, "Failed to get clickhouse host")

	port, err := r.container.MappedPort(ctx, nat.Port("9000/tcp"))
	r.Require().NoError(err, "Failed to get clickhouse port")

	settings := config.Settings{
		ClickHouseHost:     host,
		ClickHouseTCPPort:  port.Int(),
		ClickHouseUser:     r.container.User,
		ClickHousePassword: r.container.Password,
	}
	r.repo, err = repositories.NewRepository(&testLogger, settings)
	r.Require().NoError(err, "Failed to create repository")
	r.dataStartTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r.insertTestData()
}

func (r *RepositoryTestSuite) TearDownSuite() {
	r.container.Terminate(context.Background())
}

func (r *RepositoryTestSuite) TestGetSignalFloats() {
	endTs := r.dataStartTime.Add(time.Hour * 24 * 14)
	ctx := context.Background()
	testCases := []struct {
		name     string
		sigArgs  repositories.FloatSignalArgs
		expected []model.SignalFloat
	}{
		{
			name: "average",
			sigArgs: repositories.FloatSignalArgs{
				SignalArgs: repositories.SignalArgs{
					TokenID: 1,
					FromTS:  r.dataStartTime,
					ToTS:    endTs,
					Name:    vss.FieldSpeed,
				},
				Agg: model.FloatAggregation{
					Type:     model.FloatAggregationTypeAvg,
					Interval: day.String(),
				},
			},
			expected: []model.SignalFloat{
				{
					Timestamp: ref(r.dataStartTime),
					Value:     ref(5.0),
				},
			},
		},
		{
			name: "max smartcar",
			sigArgs: repositories.FloatSignalArgs{
				SignalArgs: repositories.SignalArgs{
					TokenID: 1,
					FromTS:  r.dataStartTime,
					ToTS:    endTs,
					Name:    vss.FieldSpeed,
					Filter: &model.SignalFilter{
						Source: ref("autopi"),
					},
				},
				Agg: model.FloatAggregation{
					Type:     model.FloatAggregationTypeMax,
					Interval: day.String(),
				},
			},

			expected: []model.SignalFloat{
				{
					Timestamp: ref(r.dataStartTime),
					Value:     ref(7.0),
				},
			},
		},
	}
	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the GetSignalFloats method
			result, err := r.repo.GetSignalFloats(ctx, &tc.sigArgs)
			r.Require().NoError(err)

			for i, sig := range result {
				r.Require().Equal(tc.expected[i], *sig)
			}
		})
	}
}

func (r *RepositoryTestSuite) TestGetSignalString() {
	ctx := context.Background()
	testCases := []struct {
		name     string
		sigArgs  repositories.StringSignalArgs
		expected []*model.SignalString
	}{
		{
			name: "unique",
			sigArgs: repositories.StringSignalArgs{
				SignalArgs: repositories.SignalArgs{
					TokenID: 1,
					FromTS:  r.dataStartTime,
					ToTS:    r.dataStartTime.Add(time.Hour),
					Name:    vss.FieldPowertrainType,
				},
				Agg: model.StringAggregation{
					Type:     model.StringAggregationTypeUnique,
					Interval: day.String(),
				},
			},
			expected: []*model.SignalString{
				{
					Timestamp: ref(r.dataStartTime),
					Value:     ref("value2,value1,value3"),
				},
			},
		},
		{
			name: "Top autopi",
			sigArgs: repositories.StringSignalArgs{
				SignalArgs: repositories.SignalArgs{
					TokenID: 1,
					FromTS:  r.dataStartTime,
					ToTS:    r.dataStartTime.Add(time.Hour),
					Name:    vss.FieldPowertrainType,
					Filter: &model.SignalFilter{
						Source: ref("autopi"),
					},
				},
				Agg: model.StringAggregation{
					Type:     model.StringAggregationTypeTop,
					Interval: day.String(),
				},
			},
			expected: []*model.SignalString{
				{
					Timestamp: ref(r.dataStartTime),
					Value:     ref("value2"),
				},
			},
		},
	}

	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the GetSignalString method
			result, err := r.repo.GetSignalString(ctx, &tc.sigArgs)
			r.Require().NoError(err)

			for i, sig := range result {
				if tc.sigArgs.Agg.Type == model.StringAggregationTypeUnique {
					// split the expected value and the actual value by comma
					expected := strings.Split(*tc.expected[i].Value, ",")
					actual := strings.Split(*sig.Value, ",")
					r.Require().Equal(expected, actual)
				} else {
					r.Require().Equal(tc.expected[i], sig)
				}
			}
		})
	}
}
func (r *RepositoryTestSuite) TestGetLatestSignalFloat() {
	ctx := context.Background()
	testCases := []struct {
		name     string
		sigArgs  repositories.SignalArgs
		expected *model.SignalFloat
	}{
		{
			name: "latest",
			sigArgs: repositories.SignalArgs{
				TokenID: 1,
				Name:    vss.FieldSpeed,
			},
			expected: &model.SignalFloat{
				Timestamp: ref(r.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-1)))),
				Value:     ref(9.0),
			},
		},
		{
			name: "latest smartcar",
			sigArgs: repositories.SignalArgs{
				TokenID: 1,
				Name:    vss.FieldSpeed,
				Filter: &model.SignalFilter{
					Source: ref("smartcar"),
				},
			},
			expected: &model.SignalFloat{
				Timestamp: ref(r.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-2)))),
				Value:     ref(8.0),
			},
		},
	}
	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the GetLatestSignalFloat method
			result, err := r.repo.GetLatestSignalFloat(ctx, &tc.sigArgs)
			r.Require().NoError(err)
			r.Require().Equal(tc.expected, result)
		})
	}
}

func (r *RepositoryTestSuite) TestGetLatestSignalString() {
	ctx := context.Background()
	testCases := []struct {
		name     string
		sigArgs  repositories.SignalArgs
		expected *model.SignalString
	}{
		{
			name: "latest",
			sigArgs: repositories.SignalArgs{
				TokenID: 1,
				Name:    vss.FieldPowertrainType,
			},
			expected: &model.SignalString{
				Timestamp: ref(r.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-1)))),
				Value:     ref("value1"),
			},
		},
		{
			name: "latest smartcar",
			sigArgs: repositories.SignalArgs{
				TokenID: 1,
				Name:    vss.FieldPowertrainType,
				Filter: &model.SignalFilter{
					Source: ref("smartcar"),
				},
			},
			expected: &model.SignalString{
				Timestamp: ref(r.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-2)))),
				Value:     ref("value3"),
			},
		},
	}
	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the GetLatestSignalString method
			result, err := r.repo.GetLatestSignalString(ctx, &tc.sigArgs)
			r.Require().NoError(err)
			r.Require().Equal(tc.expected, result)
		})
	}
}

func (r *RepositoryTestSuite) TestLastSeen() {
	ctx := context.Background()
	testCases := []struct {
		name     string
		sigArgs  repositories.SignalArgs
		expected time.Time
	}{
		{
			name:     "last seen",
			expected: r.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-1))),
			sigArgs: repositories.SignalArgs{
				TokenID: 1,
			},
		},
		{
			name:     "last seen smartcar",
			expected: r.dataStartTime.Add(time.Second * time.Duration(30*(dataPoints-2))),
			sigArgs: repositories.SignalArgs{
				TokenID: 1,
				Filter: &model.SignalFilter{
					Source: ref("smartcar"),
				},
			},
		},
	}
	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the LastSeen method
			result, err := r.repo.GetLastSeen(ctx, &tc.sigArgs)
			r.Require().NoError(err)
			r.Require().Equal(tc.expected, result)
		})
	}

}

// insertTestData inserts test data into the clickhouse database.
// it loops for 10 iterations and inserts a 2 signals  with each iteration that have a value of i and a powertrain type of "value"+ n%3+1
// The source is selected from a list of sources in a round robin fashion of sources[i%3].
// The timestamp is incremented by 30 seconds for each iteration.
func (r *RepositoryTestSuite) insertTestData() {
	ctx := context.Background()
	conn, err := clickhouseinfra.GetClickHouseAsConn(r.container.ClickHouseContainer)
	r.Require().NoError(err, "Failed to get clickhouse connection")
	testSignal := []vss.Signal{}
	var sources = []string{"dimo/integration/2ULfuC8U9dOqRshZBAi0lMM1Rrx", "dimo/integration/27qftVRWQYpVDcO5DltO5Ojbjxk", "dimo/integration/22N2xaPOq2WW2gAHBHd0Ikn4Zob"}
	for i := range dataPoints {

		numSig := vss.Signal{
			Name:        vss.FieldSpeed,
			Timestamp:   r.dataStartTime.Add(time.Second * time.Duration(30*i)),
			Source:      sources[i%3],
			TokenID:     1,
			ValueNumber: float64(i),
		}

		strSig := vss.Signal{
			Name:        vss.FieldPowertrainType,
			Timestamp:   r.dataStartTime.Add(time.Second * time.Duration(30*i)),
			Source:      sources[i%3],
			TokenID:     1,
			ValueString: fmt.Sprintf("value%d", i%3+1),
		}

		testSignal = append(testSignal, numSig, strSig)
	}

	// insert the test data into the clickhouse database
	batch, err := conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", vss.TableName))
	r.Require().NoError(err, "Failed to prepare batch")
	for _, sig := range testSignal {
		err := batch.AppendStruct(&sig)
		r.Require().NoError(err, "Failed to append struct")
	}
	err = batch.Send()
	r.Require().NoError(err, "Failed to send batch")
}

func ref[T any](t T) *T {
	return &t
}
