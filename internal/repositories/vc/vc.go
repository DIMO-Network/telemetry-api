package vc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/attestation-api/pkg/types"
	"github.com/DIMO-Network/cloudevent"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type indexRepoService interface {
	GetLatestCloudEvent(ctx context.Context, filter *grpc.AdvancedSearchOptions) (cloudevent.CloudEvent[json.RawMessage], error)
}
type Repository struct {
	indexService          indexRepoService
	vinDataVersion        string
	chainID               uint64
	vehicleAddress        common.Address
	storageNodeDevLicense common.Address

	vinVCDataVersion string
}

// New creates a new instance of Service.
func New(indexService indexRepoService, settings config.Settings) *Repository {
	return &Repository{
		indexService:          indexService,
		vinDataVersion:        settings.VINDataVersion,
		chainID:               settings.ChainID,
		vehicleAddress:        settings.VehicleNFTAddress,
		storageNodeDevLicense: settings.StorageNodeDevLicense,
		vinVCDataVersion:      settings.VINVCDataVersion,
	}
}

// GetLatestVINVC fetches the latest VIN VC for the given vehicle.
func (r *Repository) GetLatestVINVC(ctx context.Context, subject string) (*model.Vinvc, error) {
	dataObj, err := r.getVINVC(ctx, subject)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil //nolint // we nil is a valid response
		}
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, err, "internal error")
	}

	cred := types.Credential{}
	if err := json.Unmarshal(dataObj.Data, &cred); err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to unmarshal VIN VC: %w", err), "internal error")
	}

	var expiresAt *time.Time
	if !cred.ValidTo.IsZero() {
		expiresAt = &cred.ValidTo
	}
	var createdAt *time.Time
	if !cred.ValidFrom.IsZero() {
		createdAt = &cred.ValidFrom
	}
	credSubject := types.VINSubject{}
	if err := json.Unmarshal(cred.CredentialSubject, &credSubject); err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to unmarshal VIN credential subject: %w", err), "internal error")
	}
	var vin *string
	if credSubject.VehicleIdentificationNumber != "" {
		vin = &credSubject.VehicleIdentificationNumber
	}
	var recordedBy *string
	if credSubject.RecordedBy != "" {
		recordedBy = &credSubject.RecordedBy
	}
	var recordedAt *time.Time
	if !credSubject.RecordedAt.IsZero() {
		recordedAt = &credSubject.RecordedAt
	}
	var countryCode *string
	if credSubject.CountryCode != "" {
		countryCode = &credSubject.CountryCode
	}
	var vehicleContractAddress *string
	if credSubject.VehicleContractAddress != "" {
		vehicleContractAddress = &credSubject.VehicleContractAddress
	}
	rawVc, err := json.Marshal(dataObj)
	if err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to marshal VIN VC: %w", err), "internal error")
	}
	tokenIDInt := int(credSubject.VehicleTokenID)
	return &model.Vinvc{
		ValidFrom:              createdAt,
		ValidTo:                expiresAt,
		RawVc:                  string(rawVc),
		Vin:                    vin,
		RecordedBy:             recordedBy,
		RecordedAt:             recordedAt,
		CountryCode:            countryCode,
		VehicleContractAddress: vehicleContractAddress,
		VehicleTokenID:         &tokenIDInt,
	}, nil
}

func (r *Repository) getVINVC(ctx context.Context, vehicleDID string) (cloudevent.RawEvent, error) {
	opts := &grpc.AdvancedSearchOptions{
		DataVersion: &grpc.StringFilterOption{
			In: []string{r.vinDataVersion},
		},
		Type: &grpc.StringFilterOption{
			In: []string{cloudevent.TypeAttestation},
		},
		Subject: &grpc.StringFilterOption{
			In: []string{vehicleDID},
		},
		Source: &grpc.StringFilterOption{
			In: []string{r.storageNodeDevLicense.Hex()},
		},
	}
	dataObj, err := r.indexService.GetLatestCloudEvent(ctx, opts)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return cloudevent.RawEvent{}, fmt.Errorf("failed to get latest VIN attestation data: %w", err)
		}
		return cloudevent.RawEvent{}, fmt.Errorf("failed to get latest VIN attestation data: %w", err)
	}
	return dataObj, nil
}
