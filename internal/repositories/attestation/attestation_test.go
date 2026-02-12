//go:generate go tool mockgen -source=attestation.go -destination=attestation_mock_test.go -package=attestation_test
package attestation_test

import (
	context "context"
	json "encoding/json"
	"math/big"
	"testing"
	"time"

	cloudevent "github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/auth"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/repositories/attestation"
	"github.com/DIMO-Network/token-exchange-api/pkg/tokenclaims"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var vehicleAddress = common.HexToAddress("0x123")
var chainID = uint64(1)

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

func TestGetAttestations(t *testing.T) {
	// Initialize variables
	ctx := context.Background()
	validVehTknID := int(123)
	invalidVehTknID := int(321)

	validSigner := common.BigToAddress(big.NewInt(1))

	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock services
	mockService := NewMockindexRepoService(ctrl)
	// Initialize the service with mock dependencies
	att := attestation.New(mockService, chainID, vehicleAddress)

	vehicleDID := cloudevent.ERC721DID{
		ChainID:         chainID,
		ContractAddress: vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(validVehTknID)),
	}.String()

	dataStr := `{"goodTires": true}`
	defaultEvent := cloudevent.CloudEvent[json.RawMessage]{
		Data: json.RawMessage(dataStr),
	}
	defaultEvent.Time = time.Now()
	defaultEvent.Source = validSigner.Hex()
	defaultEvent.Subject = vehicleDID
	defaultEvent.Signature = "signature"
	defaultJSONB, err := json.Marshal(defaultEvent)
	defaultJSON := string(defaultJSONB)
	require.NoError(t, err)
	time := time.Now()
	id := ksuid.New().String()
	producer := "did:nft:153:0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF_123"
	dataVersion := "1.0"
	limit := 10

	// Test cases
	tests := []struct {
		name         string
		mockSetup    func() context.Context
		vehTknID     int
		filters      *model.AttestationFilter
		expectedAtts []*model.Attestation
		err          error
	}{
		{
			name: "success: search for all attestations for token id",
			mockSetup: func() context.Context {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
				return populateClaimMap(ctx, []string{tokenclaims.GlobalIdentifier}, []string{tokenclaims.GlobalIdentifier}, [][]string{[]string{tokenclaims.GlobalIdentifier}}, validVehTknID)
			},
			vehTknID: validVehTknID,
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					ID:             id,
					VehicleTokenID: validVehTknID,
					Time:           defaultEvent.Time,
					Attestation:    defaultJSON,
					Type:           cloudevent.TypeAttestation,
					Source:         validSigner,
					Producer:       &producer,
					DataVersion:    dataVersion,
				},
			},
		},
		{
			name: "success: search for all attestations for token id, test all filters",
			mockSetup: func() context.Context {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]cloudevent.CloudEvent[json.RawMessage]{
					defaultEvent,
				}, nil)
				return populateClaimMap(ctx, []string{tokenclaims.GlobalIdentifier}, []string{tokenclaims.GlobalIdentifier}, [][]string{[]string{tokenclaims.GlobalIdentifier}}, validVehTknID)
			},
			filters: &model.AttestationFilter{
				Before:      &time,
				After:       &time,
				DataVersion: &dataVersion,
				Producer:    &producer,
				Source:      &validSigner,
				Limit:       &limit,
				ID:          &id,
			},
			vehTknID: validVehTknID,
			expectedAtts: []*model.Attestation{
				&model.Attestation{
					ID:             id,
					VehicleTokenID: validVehTknID,
					Time:           defaultEvent.Time,
					Attestation:    defaultJSON,
					Type:           cloudevent.TypeAttestation,
					Source:         validSigner,
					Producer:       &producer,
					DataVersion:    dataVersion,
				},
			},
		},
		{
			name: "success: no attestations for token id",
			mockSetup: func() context.Context {
				mockService.EXPECT().GetAllCloudEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
				return populateClaimMap(ctx, []string{tokenclaims.GlobalIdentifier}, []string{tokenclaims.GlobalIdentifier}, [][]string{[]string{tokenclaims.GlobalIdentifier}}, invalidVehTknID)
			},
			vehTknID: invalidVehTknID,
		},
		{
			name: "fail: no claims for source",
			mockSetup: func() context.Context {
				return populateClaimMap(ctx,
					[]string{
						tokenclaims.GlobalIdentifier,
					},
					[]string{
						common.BigToAddress(big.NewInt(999)).Hex(),
					},
					[][]string{
						[]string{tokenclaims.GlobalIdentifier},
					}, validVehTknID)
			},
			filters: &model.AttestationFilter{
				Before:      &time,
				After:       &time,
				DataVersion: &dataVersion,
				Producer:    &producer,
				Source:      &validSigner,
				Limit:       &limit,
			},
			vehTknID:     validVehTknID,
			expectedAtts: []*model.Attestation{},
			err:          errorhandler.NewInternalErrorWithMsg(ctx, jwtmiddleware.ErrJWTInvalid, "invalid claims"),
		},
		{
			name: "fail: no attestation claims in jwt",
			mockSetup: func() context.Context {
				return ctx
			},
			filters: &model.AttestationFilter{
				Before:      &time,
				After:       &time,
				DataVersion: &dataVersion,
				Producer:    &producer,
				Source:      &validSigner,
				Limit:       &limit,
			},
			vehTknID:     validVehTknID,
			expectedAtts: []*model.Attestation{},
			err:          errorhandler.NewInternalErrorWithMsg(ctx, jwtmiddleware.ErrJWTInvalid, "invalid claims"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			enrichedCtx := tt.mockSetup()
			// Call the method

			attestations, err := att.GetAttestations(enrichedCtx, att.DefaultDID(tt.vehTknID), tt.filters)
			if tt.err != nil {
				require.Error(t, err)
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
				require.EqualValues(t, tt.expectedAtts[idx].Time, att.Time)
				require.EqualValues(t, tt.expectedAtts[idx].VehicleTokenID, att.VehicleTokenID)
			}
		})
	}
}

func populateClaimMap(ctx context.Context, ce, source []string, ids [][]string, tokenID int) context.Context {
	var claims auth.TelemetryClaim
	claims.CloudEvents = &tokenclaims.CloudEvents{}

	for idx := range ce {
		claims.CloudEvents.Events = append(claims.CloudEvents.Events, tokenclaims.Event{
			EventType: ce[idx],
			Source:    source[idx],
			IDs:       ids[idx],
		})
	}
	claims.Asset = cloudevent.ERC721DID{
		ChainID:         uint64(1),
		ContractAddress: vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(tokenID)),
	}.String()

	return context.WithValue(ctx, auth.TelemetryClaimContextKey{}, &claims)
}
