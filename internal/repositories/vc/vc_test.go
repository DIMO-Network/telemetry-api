//go:generate mockgen -destination=mock_service_test.go -package=vc_test github.com/DIMO-Network/nameindexer/pkg/clickhouse/indexrepo ObjectGetter
//go:generate mockgen -destination=mock_clickhouse_test.go -package=vc_test github.com/ClickHouse/clickhouse-go/v2 Conn
package vc_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/DIMO-Network/attestation-api/pkg/verifiable"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/vc"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
	logger := zerolog.New(nil)
	ctx := context.Background()
	vehicleTokenID := uint32(123)
	dataType := "vinvc"
	bucketName := "bucket-name"

	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock services
	mockChConn := NewMockConn(ctrl)
	mockObjGetter := NewMockObjectGetter(ctrl)

	// Initialize the service with mock dependencies
	svc := vc.New(mockChConn, mockObjGetter, bucketName, dataType, "", "", &logger)

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
	defaultData, err := json.Marshal(defaultVC)
	require.NoError(t, err, "failed to marshal defaultVC")

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
				mockChConn.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&MockRow{data: "filename"})
				mockObjGetter.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(defaultData))}, nil)
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
				mockChConn.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&MockRow{err: sql.ErrNoRows})
			},
			expectedVC: nil,
		},
		{
			name: "Internal error",
			mockSetup: func() {
				mockChConn.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&MockRow{err: errors.New("internal error")})
			},
			expectedVC:  nil,
			expectedErr: true,
		},
		{
			name: "Invalid data format",
			mockSetup: func() {
				mockChConn.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&MockRow{data: "filename"})
				mockObjGetter.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("invalid data")))}, nil)
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
