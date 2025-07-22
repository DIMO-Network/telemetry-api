package vc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
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
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type indexRepoService interface {
	GetLatestCloudEvent(ctx context.Context, filter *grpc.SearchOptions) (cloudevent.CloudEvent[json.RawMessage], error)
}
type Repository struct {
	indexService          indexRepoService
	vinDataVersion        string
	pomVCDataVersion      string
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
		pomVCDataVersion:      settings.POMVCDataVersion,
		chainID:               settings.ChainID,
		vehicleAddress:        settings.VehicleNFTAddress,
		storageNodeDevLicense: settings.StorageNodeDevLicense,
		vinVCDataVersion:      settings.VINVCDataVersion,
	}
}

// GetLatestVINVC fetches the latest VIN VC for the given vehicle.
func (r *Repository) GetLatestVINVC(ctx context.Context, vehicleTokenID uint32) (*model.Vinvc, error) {
	vehicleDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(vehicleTokenID)),
	}.String()
	dataObj, err := r.getVINVC(ctx, vehicleDID)
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

// GetLatestPOMVC fetches the latest POM VC for the given vehicle.
func (r *Repository) GetLatestPOMVC(ctx context.Context, vehicleTokenID uint32) (*model.Pomvc, error) {
	vehicleDID := cloudevent.ERC721DID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         new(big.Int).SetUint64(uint64(vehicleTokenID)),
	}.String()
	opts := &grpc.SearchOptions{
		DataVersion: &wrapperspb.StringValue{Value: r.pomVCDataVersion},
		Type:        &wrapperspb.StringValue{Value: cloudevent.TypeVerifableCredential},
		Subject:     &wrapperspb.StringValue{Value: vehicleDID},
	}
	dataObj, err := r.indexService.GetLatestCloudEvent(ctx, opts)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil //nolint // we nil is a valid response
		}
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to get latest POM VC data: %w", err), "internal errors")
	}
	cred := types.Credential{}
	if err := json.Unmarshal(dataObj.Data, &cred); err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to unmarshal POM VC: %w", err), "internal error")
	}

	var createdAt *time.Time
	if !cred.ValidFrom.IsZero() {
		createdAt = &cred.ValidFrom
	}

	credSubject := types.POMSubject{}
	if err := json.Unmarshal(cred.CredentialSubject, &credSubject); err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, fmt.Errorf("failed to unmarshal POM credential subject: %w", err), "internal error")
	}
	var recordedBy *string
	if credSubject.RecordedBy != "" {
		recordedBy = &credSubject.RecordedBy
	}
	var vehicleContractAddress *string
	if credSubject.VehicleContractAddress != "" {
		vehicleContractAddress = &credSubject.VehicleContractAddress
	}
	tokenIDInt := int(credSubject.VehicleTokenID)

	return &model.Pomvc{
		ValidFrom:              createdAt,
		RawVc:                  string(dataObj.Data),
		RecordedBy:             recordedBy,
		VehicleContractAddress: vehicleContractAddress,
		VehicleTokenID:         &tokenIDInt,
	}, nil
}

func (r *Repository) getVINVC(ctx context.Context, vehicleDID string) (cloudevent.RawEvent, error) {
	opts := &grpc.SearchOptions{
		DataVersion: &wrapperspb.StringValue{Value: r.vinDataVersion},
		Type:        &wrapperspb.StringValue{Value: cloudevent.TypeAttestation},
		Subject:     &wrapperspb.StringValue{Value: vehicleDID},
		Source:      &wrapperspb.StringValue{Value: r.storageNodeDevLicense.Hex()},
	}
	dataObj, err := r.indexService.GetLatestCloudEvent(ctx, opts)
	if err == nil {
		return dataObj, nil
	} else if status.Code(err) != codes.NotFound {
		return cloudevent.RawEvent{}, fmt.Errorf("failed to get latest VIN attestation data: %w", err)
	}

	opts = &grpc.SearchOptions{
		DataVersion: &wrapperspb.StringValue{Value: r.vinVCDataVersion},
		Type:        &wrapperspb.StringValue{Value: cloudevent.TypeVerifableCredential},
		Subject:     &wrapperspb.StringValue{Value: vehicleDID},
	}
	dataObj, err = r.indexService.GetLatestCloudEvent(ctx, opts)
	if err != nil {
		return cloudevent.RawEvent{}, fmt.Errorf("failed to get latest VIN VC data: %w", err)
	}
	return dataObj, nil
}
