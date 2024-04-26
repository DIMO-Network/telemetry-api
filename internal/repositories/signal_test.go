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

const day = time.Hour * 24

type RepositoryTestSuite struct {
	suite.Suite
	dataStartTime time.Time
	repo          *repositories.Repository
	container     *clickhouseinfra.Container
}

func (r *RepositoryTestSuite) SetupTest() {
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
func (r *RepositoryTestSuite) TearDownTest() {
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
	}
	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the GetSignalFloats method
			result, err := r.repo.GetSignalFloats(ctx, tc.sigArgs)
			r.Require().NoError(err)

			for i, sig := range result {
				r.Require().Equal(tc.expected[i], *sig)
			}
		})
	}
}

func (r *RepositoryTestSuite) insertTestData() {
	ctx := context.Background()
	conn, err := clickhouseinfra.GetClickHouseAsConn(r.container.ClickHouseContainer)
	r.Require().NoError(err, "Failed to get clickhouse connection")
	testSignal := []vss.Signal{}
	for i := range 10 {
		sig := vss.Signal{
			Name:        vss.FieldSpeed,
			Timestamp:   r.dataStartTime.Add(time.Second * time.Duration(30*i)),
			TokenID:     1,
			ValueNumber: float64(i),
		}
		testSignal = append(testSignal, sig)
	}
	for i := range 10 {
		sig := vss.Signal{
			Name:        vss.FieldPowertrainType,
			Timestamp:   r.dataStartTime.Add(time.Second * time.Duration(30*i)),
			TokenID:     1,
			ValueString: fmt.Sprintf("value%d", i%3+1),
		}
		testSignal = append(testSignal, sig)
	}

	batch, err := conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", vss.TableName))
	r.Require().NoError(err, "Failed to prepare batch")
	for _, sig := range testSignal {
		err := batch.AppendStruct(&sig)
		r.Require().NoError(err, "Failed to append struct")
	}
	err = batch.Send()
	r.Require().NoError(err, "Failed to send batch")
}
func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func ref[T any](t T) *T {
	return &t
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
	}

	for _, tc := range testCases {
		r.Run(tc.name, func() {
			// Call the GetSignalString method
			result, err := r.repo.GetSignalString(ctx, tc.sigArgs)
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
