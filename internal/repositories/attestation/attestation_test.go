//go:generate go tool mockgen -source=attestation.go -destination=attestation_mock_test.go -package=attestation_test
package attestation_test

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/ethereum/go-ethereum/common"
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
	validVehTknID := int(123)
	invalidVehTknID := int(321)

	validSigner := common.BigToAddress(big.NewInt(1))
	invalidSigner := "invalidSigner"

	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock services
	mockService := NewMockindexRepoService(ctrl)
	vehicleAddress := common.HexToAddress("0x123")
	chainID := uint64(3)

	// Initialize the service with mock dependencies
	att := attestation.New(mockService, chainID, vehicleAddress, &logger)

	vehicleDID := cloudevent.NFTDID{
		ChainID:         1,
		ContractAddress: vehicleAddress,
		TokenID:         uint32(validVehTknID),
	}.String()

	dataStr := `{"goodTires": true}`
	defaultEvent := cloudevent.CloudEvent[json.RawMessage]{
		Data: json.RawMessage(dataStr),
	}
	defaultEvent.Time = time.Now()
	defaultEvent.Source = validSigner.Hex()
	defaultEvent.Subject = vehicleDID

	// Test cases
	tests := []struct {
		name         string
		mockSetup    func()
		vehTknID     uint32
		signer       *string
		expectedAtts []*model.Attestation
		expectedErr  bool
		err          error
	}{
		{
			name: "Success with signer",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any()).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
			},
			vehTknID: uint32(validVehTknID),
			signer:   &defaultEvent.Source,
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					VehicleTokenID: &validVehTknID,
					RecordedAt:     &defaultEvent.Time,
					Attestation:    dataStr,
				},
			},
		},
		{
			name: "Success without signer",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any()).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
			},
			vehTknID: uint32(validVehTknID),
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					VehicleTokenID: &validVehTknID,
					RecordedAt:     &defaultEvent.Time,
					Attestation:    dataStr,
				},
			},
		},
		{
			name: "no cloud events returned (not an error)",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			vehTknID: uint32(invalidVehTknID),
		},
		{
			name:        "Invalid attestation signer",
			mockSetup:   func() {},
			signer:      &invalidSigner,
			vehTknID:    uint32(invalidVehTknID),
			expectedErr: true,
			err:         errors.New("invalid attestation signer"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.mockSetup()
			// Call the method
			attestations, err := att.GetAttestations(ctx, tt.vehTknID, tt.signer)

			// Assert the results
			if tt.expectedErr {
				require.Error(t, err)
				require.Equal(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
			if tt.expectedAtts == nil {
				require.Nil(t, attestations)
				return
			}

			for idx, att := range attestations {
				require.JSONEq(t, tt.expectedAtts[idx].Attestation, att.Attestation)
				require.EqualValues(t, tt.expectedAtts[idx].RecordedAt, att.RecordedAt)
				require.EqualValues(t, tt.expectedAtts[idx].VehicleTokenID, att.VehicleTokenID)
			}
		})
	}
}

func ref[T any](v T) *T {
	return &v
}
