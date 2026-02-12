//go:generate go tool mockgen -source=vc.go -destination=vc_mock_test.go -package=vc_test
package vc_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/telemetry-api/internal/config"
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

	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock services
	mockService := NewMockindexRepoService(ctrl)
	vehicleAddress := common.HexToAddress("0xfEDCBA0987654321FeDcbA0987654321fedCBA09")
	chainID := uint64(3)
	// Initialize the service with mock dependencies
	settings := config.Settings{
		VINDataVersion:        "vin_data_version",
		POMVCDataVersion:      "pomvc_data_version",
		ChainID:               chainID,
		VehicleNFTAddress:     vehicleAddress,
		StorageNodeDevLicense: common.HexToAddress("0x123"),
		VINVCDataVersion:      "vinvc_data_version",
	}
	svc := vc.New(mockService, settings)

	attestationCredData := []byte(`{
    "validFrom": "2025-01-15T10:30:00.000000Z",
    "validTo": "2025-01-20T00:00:00Z",
    "credentialSubject": {
      "id": "did:erc721:80002:0xfEDCBA0987654321FeDcbA0987654321fedCBA09:456",
      "vehicleTokenId": 456,
	  "countryCode": "US",
      "vehicleContractAddress": "eth:0xfEDCBA0987654321FeDcbA0987654321fedCBA09",
      "vehicleIdentificationNumber": "TEST12345678901234",
      "recordedBy": "did:erc721:80002:0xabcdef1234567890abcdef1234567890abcdef12:123",
      "recordedAt": "2025-01-14T15:45:30.000Z"
    }
  }`)
	attestationEvent := cloudevent.CloudEvent[json.RawMessage]{
		Data: attestationCredData,
	}
	attestationData, err := json.Marshal(attestationEvent)
	require.NoError(t, err, "failed to marshal attestationVC")

	// Test cases
	tests := []struct {
		name        string
		mockSetup   func()
		expectedVC  *model.Vinvc
		expectedErr bool
	}{
		{
			name: "Success_attestation",
			mockSetup: func() {
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), matchOpts(&grpc.AdvancedSearchOptions{
					DataVersion: &grpc.StringFilterOption{In: []string{settings.VINDataVersion}},
					Type:        &grpc.StringFilterOption{In: []string{cloudevent.TypeAttestation}},
					Subject:     &grpc.StringFilterOption{In: []string{"did:erc721:3:0xfEDCBA0987654321FeDcbA0987654321fedCBA09:123"}},
					Source:      &grpc.StringFilterOption{In: []string{common.HexToAddress("0x123").Hex()}},
				})).Return(attestationEvent, nil)
			},
			expectedVC: &model.Vinvc{
				Vin:                    ref("TEST12345678901234"),
				RecordedBy:             ref("did:erc721:80002:0xabcdef1234567890abcdef1234567890abcdef12:123"),
				RecordedAt:             ref(time.Date(2025, 1, 14, 15, 45, 30, 0, time.UTC)),
				CountryCode:            ref("US"),
				VehicleContractAddress: ref("eth:0xfEDCBA0987654321FeDcbA0987654321fedCBA09"),
				VehicleTokenID:         ref(456),
				RawVc:                  string(attestationData),
			},
		},
		{
			name: "No data found",
			mockSetup: func() {
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), gomock.Any()).Return(cloudevent.RawEvent{}, status.Error(codes.NotFound, "no data found"))
			},
			expectedVC: nil,
		},
		{
			name: "Internal error",
			mockSetup: func() {
				mockService.EXPECT().GetLatestCloudEvent(gomock.Any(), gomock.Any()).Return(cloudevent.RawEvent{}, errors.New("internal error"))
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

type optsMatcher struct {
	opts *grpc.AdvancedSearchOptions
}

func (m *optsMatcher) Matches(x interface{}) bool {
	opts, ok := x.(*grpc.AdvancedSearchOptions)
	if !ok {
		return false
	}
	if m.opts == nil {
		return opts == nil
	}
	return opts.GetAfter().AsTime().Equal(m.opts.GetAfter().AsTime()) &&
		opts.GetBefore().AsTime().Equal(m.opts.GetBefore().AsTime()) &&
		opts.GetTimestampAsc().GetValue() == m.opts.GetTimestampAsc().GetValue() &&
		slices.Equal(opts.GetType().GetIn(), m.opts.GetType().GetIn()) &&
		slices.Equal(opts.GetDataVersion().GetIn(), m.opts.GetDataVersion().GetIn()) &&
		slices.Equal(opts.GetSubject().GetIn(), m.opts.GetSubject().GetIn()) &&
		slices.Equal(opts.GetSource().GetIn(), m.opts.GetSource().GetIn()) &&
		slices.Equal(opts.GetProducer().GetIn(), m.opts.GetProducer().GetIn()) &&
		slices.Equal(opts.GetExtras().GetIn(), m.opts.GetExtras().GetIn()) &&
		slices.Equal(opts.GetId().GetIn(), m.opts.GetId().GetIn())
}

func (m *optsMatcher) String() string {
	return fmt.Sprintf("opts: %+v", m.opts)
}

func matchOpts(opts *grpc.AdvancedSearchOptions) gomock.Matcher {
	return &optsMatcher{opts: opts}
}
