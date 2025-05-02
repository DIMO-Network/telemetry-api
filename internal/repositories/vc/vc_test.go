//go:generate go tool mockgen -source=vc.go -destination=vc_mock_test.go -package=vc_test
package vc_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DIMO-Network/attestation-api/pkg/verifiable"
	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockRow implements sql.Row and returns a string when scanned.
type MockRow struct {
	data string
	err  error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	if len(dest) > 0 {
		if s, ok := dest[0].(*string); ok {
			*s = m.data
		}
	}
	return nil
}

func (m *MockRow) Err() error {
	return nil
}

func (m *MockRow) ScanStruct(any) error {
	return nil
}

func TestGetLatestVC(t *testing.T) {
	// Initialize variables
	ctx := context.Background()
	vehicleTokenID := uint32(123)
	dataVersion := "vinvc"
	bucketName := "bucket-name"

	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock services
	mockService := NewMockindexRepoService(ctrl)
	vehicleAddress := common.HexToAddress("0x123")
	chainID := uint64(3)
	// Initialize the service with mock dependencies
	svc := vc.New(mockService, bucketName, dataVersion, chainID, vehicleAddress)

	defaultVC := verifiable.Credential{
		ValidTo:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		ValidFrom: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		CredentialSubject: json.RawMessage(`{
			"vehicleIdentificationNumber": "VIN123",
			"recordedBy": "Recorder",
			"recordedAt": "2024-01-01T00:00:00Z",
			"countryCode": "US",
			"vehicleContractAddress": "0xAddress",
			"vehicleTokenID": 123
		}`),
	}
	credData, err := json.Marshal(defaultVC)
	require.NoError(t, err, "failed to marshal defaultVC")
	defaultEvent := cloudevent.CloudEvent[json.RawMessage]{
		Data: credData,
	}
	defaultData, err := json.Marshal(defaultEvent)
	require.NoError(t, err, "failed to marshal defaultVC")

	emptyEvent := cloudevent.CloudEvent[json.RawMessage]{}
	// Test cases
	tests := []struct {
		name        string
		mockSetup   func()
		expectedVC  *model.Vinvc
		expectedErr bool
	}{
		{
			name: "Success",
			mockSetup: func() {
				// Create a mock verifiable credential
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), gomock.Any()).Return(defaultEvent, nil)
			},
			expectedVC: &model.Vinvc{
				Vin:                    ref("VIN123"),
				RecordedBy:             ref("Recorder"),
				RecordedAt:             ref(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				CountryCode:            ref("US"),
				VehicleContractAddress: ref("0xAddress"),
				VehicleTokenID:         ref(123),
				RawVc:                  string(defaultData),
			},
		},
		{
			name: "No data found",
			mockSetup: func() {
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), gomock.Any()).Return(emptyEvent, status.Error(codes.NotFound, "no data found"))
			},
			expectedVC: nil,
		},
		{
			name: "Internal error",
			mockSetup: func() {
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), gomock.Any()).Return(emptyEvent, errors.New("internal error"))
			},
			expectedVC:  nil,
			expectedErr: true,
		},
		{
			name: "Invalid data format",
			mockSetup: func() {
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), gomock.Any()).Return(cloudevent.CloudEvent[json.RawMessage]{Data: json.RawMessage("invalid data")}, nil)
			},
			expectedVC:  nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.mockSetup()

			// Call the method
			vc, err := svc.GetLatestVINVC(ctx, vehicleTokenID)

			// Assert the results
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.expectedVC == nil {
				require.Nil(t, vc)
				return
			}
			require.EqualValues(t, tt.expectedVC.Vin, vc.Vin)
			require.EqualValues(t, tt.expectedVC.RecordedBy, vc.RecordedBy)
			require.EqualValues(t, tt.expectedVC.RecordedAt, vc.RecordedAt)
			require.EqualValues(t, tt.expectedVC.CountryCode, vc.CountryCode)
			require.EqualValues(t, tt.expectedVC.VehicleContractAddress, vc.VehicleContractAddress)
			require.EqualValues(t, tt.expectedVC.VehicleTokenID, vc.VehicleTokenID)
			require.JSONEq(t, tt.expectedVC.RawVc, vc.RawVc)
		})
	}
}

func ref[T any](v T) *T {
	return &v
}
