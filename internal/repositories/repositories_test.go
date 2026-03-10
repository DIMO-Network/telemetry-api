//go:generate go tool mockgen -source=./repositories.go -destination=repositories_mocks_test.go -package=repositories_test
package repositories_test

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories"
	"github.com/DIMO-Network/telemetry-api/internal/service/ch"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

var baseSettings = config.Settings{
	DeviceLastSeenBinHrs: 3,
	ChainID:              80002,
	VehicleNFTAddress:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
}

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
	testSubject := cloudevent.ERC721DID{
		ChainID:         baseSettings.ChainID,
		ContractAddress: baseSettings.VehicleNFTAddress,
		TokenID:         big.NewInt(1),
	}.String()
	defaultArgs := &model.AggregatedSignalArgs{
		SignalArgs: model.SignalArgs{
			TokenID: 1,
		},
		FromTS:   time.Now(),
		ToTS:     time.Now().Add(time.Hour),
		Interval: 1,
		FloatArgs: []model.FloatSignalArgs{
			{
				Name:  vss.FieldSpeed,
				Agg:   model.FloatAggregationAvg,
				Alias: vss.FieldSpeed,
			},
		},
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
					GetAggregatedSignals(gomock.Any(), testSubject, defaultArgs).
					Return([]*ch.AggSignal{}, nil)
			},
			expectedResult: []*model.SignalAggregations{},
			expectError:    false,
		},
		{
			name:    "Success case - One signal",
			aggArgs: defaultArgs,
			mockSetup: func(m *Mocks) {
				signals := []*ch.AggSignal{
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 1.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), testSubject, defaultArgs).
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
				signals := []*ch.AggSignal{
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 1.0},
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 2.0},
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC), ValueNumber: 3.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), testSubject, defaultArgs).
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
				signals := []*ch.AggSignal{
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 1.0},
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 2.0},
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 3.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), testSubject, defaultArgs).
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
				signals := []*ch.AggSignal{
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), ValueNumber: 1.0},
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC), ValueNumber: 2.0},
					{SignalType: ch.FloatType, SignalIndex: 0, Timestamp: time.Date(2024, 6, 11, 1, 0, 0, 0, time.UTC), ValueNumber: 3.0},
				}
				m.CHService.EXPECT().
					GetAggregatedSignals(gomock.Any(), testSubject, defaultArgs).
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
					GetAggregatedSignals(gomock.Any(), testSubject, defaultArgs).
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

			repo, err := repositories.NewRepository(mocks.CHService, baseSettings)
			require.NoError(t, err)
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
	testSubject := cloudevent.ERC721DID{
		ChainID:         baseSettings.ChainID,
		ContractAddress: baseSettings.VehicleNFTAddress,
		TokenID:         big.NewInt(1),
	}.String()
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
					{Data: vss.SignalData{Timestamp: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), Name: model.LastSeenField}},
				}
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), testSubject, defaultArgs).
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
					{Data: vss.SignalData{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: vss.FieldSpeed, ValueNumber: 1.0}},
					{Data: vss.SignalData{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: model.LastSeenField}},
				}
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), testSubject, defaultArgs).
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
					{Data: vss.SignalData{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: "speed", ValueNumber: 1.0}},
					{Data: vss.SignalData{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: "speed", ValueNumber: 2.0}},
					{Data: vss.SignalData{Timestamp: time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC), Name: "speed", ValueNumber: 3.0}},
					{Data: vss.SignalData{Timestamp: time.Date(2024, 6, 11, 2, 0, 0, 0, time.UTC), Name: model.LastSeenField}},
				}
				m.CHService.EXPECT().
					GetLatestSignals(gomock.Any(), testSubject, defaultArgs).
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
					GetLatestSignals(gomock.Any(), testSubject, defaultArgs).
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

			repo, err := repositories.NewRepository(mocks.CHService, baseSettings)
			require.NoError(t, err)
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

func TestGetAvailableSignals(t *testing.T) {
	testSubject := cloudevent.ERC721DID{
		ChainID:         baseSettings.ChainID,
		ContractAddress: baseSettings.VehicleNFTAddress,
		TokenID:         big.NewInt(1),
	}.String()
	tests := []struct {
		name           string
		mockSetup      func(m *Mocks)
		expectedResult []string
		expectError    bool
	}{
		{
			name: "No signals",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
					Return(nil, nil)
			},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "One signal",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
					Return([]string{"speed"}, nil)
			},
			expectedResult: []string{"speed"},
			expectError:    false,
		},
		{
			name: "Multiple signals",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
					Return([]string{"speed", "powertrainTractionBatteryStateOfChargeCurrent"}, nil)
			},
			expectedResult: []string{"speed", "powertrainTractionBatteryStateOfChargeCurrent"},
			expectError:    false,
		},
		{
			name: "Mix Unknown signals",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
					Return([]string{"speed", "newSignalName"}, nil)
			},
			expectedResult: []string{"speed"},
			expectError:    false,
		},
		{
			name: "one unknown signals",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
					Return([]string{"newSignalName"}, nil)
			},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "multiple unknown signals",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
					Return([]string{"newSignalName", "newSignalName2"}, nil)
			},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "CHService error",
			mockSetup: func(m *Mocks) {
				m.CHService.EXPECT().
					GetAvailableSignals(gomock.Any(), testSubject, nil).
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

			repo, err := repositories.NewRepository(mocks.CHService, baseSettings)
			require.NoError(t, err)
			result, err := repo.GetAvailableSignals(context.Background(), 1, nil)
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

func TestGetEvents(t *testing.T) {
	tokenID := 12345
	from := time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC)
	to := from.Add(1 * time.Hour)
	filter := &model.EventFilter{
		Name:   nil,
		Source: nil,
	}
	subject := cloudevent.ERC721DID{
		ChainID:         baseSettings.ChainID,
		ContractAddress: baseSettings.VehicleNFTAddress,
		TokenID:         big.NewInt(int64(tokenID)),
	}.String()

	eventMeta := "{\"foo\":\"bar\"}"
	vssEvents := []*vss.Event{
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "source1",
			},
			Data: vss.EventData{
				Timestamp:  from.Add(10 * time.Minute),
				Name:       "event1",
				DurationNs: 123,
				Metadata:   eventMeta,
			},
		},
		{
			CloudEventHeader: cloudevent.CloudEventHeader{
				Source: "source2",
			},
			Data: vss.EventData{
				Timestamp:  from.Add(20 * time.Minute),
				Name:       "event2",
				DurationNs: 456,
				Metadata:   "",
			},
		},
	}

	mocks := setupMocks(t)
	repo, err := repositories.NewRepository(mocks.CHService, baseSettings)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		mocks.CHService.EXPECT().
			GetEvents(gomock.Any(), subject, from, to, filter).
			Return(vssEvents, nil)
		result, err := repo.GetEvents(context.Background(), tokenID, from, to, filter)
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, vssEvents[0].Data.Name, result[0].Name)
		require.Equal(t, vssEvents[0].Source, result[0].Source)
		require.Equal(t, vssEvents[0].Data.Timestamp, result[0].Timestamp)
		require.Equal(t, int(vssEvents[0].Data.DurationNs), result[0].DurationNs)
		if vssEvents[0].Data.Metadata != "" {
			require.NotNil(t, result[0].Metadata)
			require.Equal(t, vssEvents[0].Data.Metadata, *result[0].Metadata)
		} else {
			require.Nil(t, result[0].Metadata)
		}
	})

	t.Run("error from service", func(t *testing.T) {
		mocks.CHService.EXPECT().
			GetEvents(gomock.Any(), subject, from, to, filter).
			Return(nil, errors.New("service error"))
		result, err := repo.GetEvents(context.Background(), tokenID, from, to, filter)
		require.Error(t, err)
		require.Nil(t, result)
	})

}

func ref[T any](t T) *T {
	return &t
}
