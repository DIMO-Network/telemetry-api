//go:generate go tool mockgen -source=attestation.go -destination=attestation_mock_test.go -package=attestation_test
package attestation_test

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
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

func TestAttestation(t *testing.T) {
	// Initialize variables
	logger := zerolog.New(nil)
	ctx := context.Background()
	validVehTknID := int(123)
	invalidVehTknID := int(321)

	validSigner := common.BigToAddress(big.NewInt(1))
	invalidSigner := common.BigToAddress(big.NewInt(100))

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
	time := time.Now()
	id := ksuid.New().String()
	producer := "did:nft:153:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF_123"
	dataVersion := "1.0"
	limit := int32(10)

	// Test cases
	tests := []struct {
		name         string
		mockSetup    func()
		vehTknID     uint32
		signer       *common.Address
		filters      *model.AttestationFilter
		expectedAtts []*model.Attestation
		expectedErr  bool
		err          error
	}{
		{
			name: "successful query, search for all attestations for token id",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), limit).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
			},
			vehTknID: uint32(validVehTknID),
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					ID:             id,
					VehicleTokenID: validVehTknID,
					RecordedAt:     defaultEvent.Time,
					Attestation:    dataStr,
					Type:           cloudevent.TypeAttestation,
					Source:         validSigner,
					Producer:       &producer,
					DataVersion:    dataVersion,
				},
			},
		},
		{
			name: "successful query, search for all attestations for token id by signer",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), limit).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
			},
			vehTknID: uint32(validVehTknID),
			signer:   &validSigner,
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					ID:             id,
					VehicleTokenID: validVehTknID,
					RecordedAt:     defaultEvent.Time,
					Attestation:    dataStr,
					Type:           cloudevent.TypeAttestation,
					Source:         validSigner,
					Producer:       &producer,
					DataVersion:    dataVersion,
				},
			},
		},
		{
			name: "successful query, search for all attestations for token id by signer, test all filters",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), limit).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
			},
			filters: &model.AttestationFilter{
				EffectiveAt: &time,
				ExpiresAt:   &time,
				DataVersion: &dataVersion,
				Producer:    &producer,
			},
			vehTknID: uint32(validVehTknID),
			signer:   &validSigner,
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					ID:             id,
					VehicleTokenID: validVehTknID,
					RecordedAt:     defaultEvent.Time,
					Attestation:    dataStr,
					Type:           cloudevent.TypeAttestation,
					Source:         validSigner,
					Producer:       &producer,
					DataVersion:    dataVersion,
				},
			},
		},
		{
			name: "successful query, no attestations for token id",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), limit).Return(nil, nil)
			},
			vehTknID: uint32(invalidVehTknID),
		},
		{
			name: "successful query, no attestations for token id by signer",
			mockSetup: func() {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), limit).Return(nil, nil)
			},
			signer:   &invalidSigner,
			vehTknID: uint32(invalidVehTknID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.mockSetup()
			// Call the method
			attestations, err := att.GetAttestations(ctx, tt.vehTknID, tt.signer, tt.filters)

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

			require.EqualValues(t, len(tt.expectedAtts), len(attestations))

			for idx, att := range attestations {
				require.JSONEq(t, tt.expectedAtts[idx].Attestation, att.Attestation)
				require.EqualValues(t, tt.expectedAtts[idx].RecordedAt, att.RecordedAt)
				require.EqualValues(t, tt.expectedAtts[idx].VehicleTokenID, att.VehicleTokenID)
			}
		})
	}
}
