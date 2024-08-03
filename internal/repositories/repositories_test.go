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
				signals := []*vss.Signal{
					{Timestamp: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), Name: model.LastSeenField},
				}
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), defaultArgs).
					Return(signals, nil)
			},
			expectedResult: &model.SignalCollection{
				LastSeen: nil,
			},
			expectError: false,
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

func TestDeviceActivity(t *testing.T) {
	logger := zerolog.New(nil)
	vehicleTokenID := int64(1)
	hashdog := "Hashdog"
	source := "macaron"
	lastSeen := time.Date(2024, 6, 11, 1, 2, 3, 3, time.UTC)
	lastSeenBin := time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC)

	latestArgs := &model.LatestSignalsArgs{
		IncludeLastSeen: true,
		SignalArgs: model.SignalArgs{
			TokenID: uint32(vehicleTokenID),
			Filter: &model.SignalFilter{
				Source: &source,
			},
		},
	}

	tests := []struct {
		name           string
		mockSetup      func(m *Mocks)
		manufacturer   string
		expectedResult *model.DeviceActivity
		expectError    bool
	}{
		{
			name: "Success case - valid last seen",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), latestArgs).
					Return([]*vss.Signal{
						{Timestamp: lastSeen, Name: model.LastSeenField},
					}, nil)
			},
			manufacturer: hashdog,
			expectedResult: &model.DeviceActivity{
				LastActive: &lastSeenBin,
			},
			expectError: false,
		},
		{
			name: "vehicle has not transmitted any signals",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), latestArgs).
					Return([]*vss.Signal{
						{Timestamp: time.Unix(0, 0).UTC(), Name: model.LastSeenField},
					}, nil)
			},
			manufacturer:   hashdog,
			expectedResult: &model.DeviceActivity{},
			expectError:    false,
		},
		{
			name:         "unrecognized aftermarket manufacturer",
			manufacturer: "Zaphod",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := setupMocks(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mocks)
			}

			repo := repositories.NewRepository(&logger, mocks.CHService, 2)
			result, err := repo.GetDeviceActivity(context.Background(), int(vehicleTokenID), tt.manufacturer)
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
