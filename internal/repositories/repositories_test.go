package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

type Mocks struct {
	CHService *MockCHService
}

func setupMocks(t *testing.T) *Mocks {
	mockCtrl := gomock.NewController(t)
	return &Mocks{
		CHService: NewMockCHService(mockCtrl),
	}
}

func TestGetSignal(t *testing.T) {
	logger := zerolog.New(nil)
	defaultArgs := &model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			TokenID: 1,
		},
		FromTS:   time.Now(),
		ToTS:     time.Now().Add(time.Hour),
		Interval: 1,
	}
	// idSvc := identity.NewMockIdentityService(gomock.NewController(t))
	tests := []struct {
		name           string
		aggArgs        *model.AggregatedSignalArgs
		mockSetup      func(m *Mocks)
		expectedResult []*model.SignalAggregations
		expectError    bool
	}{
		{
			name:    "Success case - No signals",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), defaultArgs).
					Return([]*model.AggSignal{}, nil)
			},
			expectedResult: []*model.SignalAggregations{},
			expectError:    false,
		},
		{
			name:    "Success case - One signal",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*model.AggSignal{
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 1.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: []*model.SignalAggregations{
				{
					Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC),
				},
			},
			expectError: false,
		},
		{
			name:    "Success case - Combine signals last signal has different timestamp",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*model.AggSignal{
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 1.0},
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 2.0},
					{Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 3.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: []*model.SignalAggregations{
				{
					Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC),
				},
				{
					Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC),
				},
			},
			expectError: false,
		},
		{
			name:    "Success case - Combine signals all signal have the same timestamp",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*model.AggSignal{
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 1.0},
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 2.0},
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 3.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: []*model.SignalAggregations{
				{
					Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC),
				},
			},
			expectError: false,
		},
		{
			name:    "Success case - Combine signals first signal has different timestamp",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*model.AggSignal{
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 1.0},
					{Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 2.0},
					{Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 3.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: []*model.SignalAggregations{
				{
					Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC),
				},
				{
					Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC),
				},
			},
			expectError: false,
		},
		{
			name:    "Invalid arguments",
			aggArgs: nil, // Example of invalid argument case
			mockSetup: func(m *Mocks) {
				// No expectations as validateAggSigArgs will fail
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:    "CHService error",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), defaultArgs).
					Return(nil, errors.New("service error"))
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := setupMocks(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mocks)
			}

			repo := repositories.NewRepository(&logger, mocks.CHService, 3)
			result, err := repo.GetSignal(context.Background(), tt.aggArgs)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				// Ensure the number of results is as expected
				require.Len(t, result, len(tt.expectedResult))
				// Compare the timestamps of the results
				for i, res := range result {
					require.Equal(t, tt.expectedResult[i].Timestamp, res.Timestamp)
				}
			}
		})
	}
}
func TestGetSignalLatest(t *testing.T) {
	logger := zerolog.New(nil)
	defaultArgs := &model.LatestSignalsArgs{
		SignalArgs: model.SignalArgs{
			TokenID: 1,
		},
		IncludeLastSeen: true,
	}
	tests := []struct {
		name           string
		latestArgs     *model.LatestSignalsArgs
		mockSetup      func(m *Mocks)
		expectedResult *model.SignalCollection
		expectError    bool
	}{
		{
			name:       "Success case - No signals",
			latestArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), defaultArgs).
					Return([]*vss.Signal{}, nil)
			},
			expectedResult: &model.SignalCollection{},
			expectError:    false,
		},
		{
			name:       "Success case - One signal",
			latestArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*vss.Signal{
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 1.0},
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: model.LastSeenField},
				}
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: &model.SignalCollection{
				LastSeen: ref(time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC)),
				Speed: &model.SignalFloat{
					Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC),
					Value:     1.0,
				},
			},
			expectError: false,
		},
		{
			name:       "Success case - Multiple signals",
			latestArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*vss.Signal{
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: "speed", ValueNumber: 1.0},
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: "speed", ValueNumber: 2.0},
					{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: "speed", ValueNumber: 3.0},
					{Timestamp: time.Date(2024, 6, 11, 2, 0, 0, 0, time.UTC), Name: model.LastSeenField},
				}
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: &model.SignalCollection{
				LastSeen: ref(time.Date(2024, 6, 11, 2, 0, 0, 0, time.UTC)),
				Speed: &model.SignalFloat{
					Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC),
					Value:     3.0,
				},
			},
			expectError: false,
		},
		{
			name:       "Invalid arguments",
			latestArgs: nil, // Example of invalid argument case
			mockSetup: func(m *Mocks) {
				// No expectations as validateSignalArgs will fail
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:       "CHService error",
			latestArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), defaultArgs).
					Return(nil, errors.New("service error"))
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	// idSvc := identity.NewMockIdentityService(gomock.NewController(t))
	// var idSvc *identity.MockIdentityService
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := setupMocks(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mocks)
			}

			repo := repositories.NewRepository(&logger, mocks.CHService, 2)
			result, err := repo.GetSignalLatest(context.Background(), tt.latestArgs)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func ref[T any](t T) *T {
	return &t
}
