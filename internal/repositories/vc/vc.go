package vc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/DIMO-Network/attestation-api/pkg/verifiable"
	"github.com/DIMO-Network/fetch-api/pkg/grpc"
	"github.com/DIMO-Network/model-garage/pkg/cloudevent"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type indexRepoService interface {
	GetLatestCloudEvent(ctx context.Context, filter *grpc.SearchOptions) (cloudevent.CloudEvent[json.RawMessage], error)
}
type Repository struct {
	logger         *zerolog.Logger
	indexService   indexRepoService
	vcBucket       string
	vinVCDataType  string
	pomVCDataType  string
	chainID        uint64
	vehicleAddress common.Address
}

// New creates a new instance of Service.
func New(indexService indexRepoService, vcBucketName, vinVCDataType, pomVCDataType string, chainID uint64, vehicleAddress common.Address, logger *zerolog.Logger) *Repository {
	return &Repository{
		indexService:   indexService,
		vcBucket:       vcBucketName,
		vinVCDataType:  vinVCDataType,
		pomVCDataType:  pomVCDataType,
		chainID:        chainID,
		vehicleAddress: vehicleAddress,
		logger:         logger,
	}
}

// GetLatestVINVC fetches the latest VIN VC for the given vehicle.
func (r *Repository) GetLatestVINVC(ctx context.Context, vehicleTokenID uint32) (*model.Vinvc, error) {
	vehicleDID := cloudevent.NFTDID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         vehicleTokenID,
	}.String()
	opts := &grpc.SearchOptions{
		DataVersion: &wrapperspb.StringValue{Value: r.vinVCDataType},
		Type:        &wrapperspb.StringValue{Value: cloudevent.TypeVerifableCredential},
		Subject:     &wrapperspb.StringValue{Value: vehicleDID},
	}
	dataObj, err := r.indexService.GetLatestCloudEvent(ctx, opts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("failed to get latest VIN VC data")
		return nil, errors.New("internal error")
	}
	cred := verifiable.Credential{}
	if err := json.Unmarshal(dataObj.Data, &cred); err != nil {
		r.logger.Error().Err(err).Msg("failed to unmarshal VIN VC")
		return nil, errors.New("internal error")
	}

	var expiresAt *time.Time
	if expirationDate, err := time.Parse(time.RFC3339, cred.ValidTo); err == nil {
		expiresAt = &expirationDate
	}
	var createdAt *time.Time
	if issuanceDate, err := time.Parse(time.RFC3339, cred.ValidFrom); err == nil {
		createdAt = &issuanceDate
	}
	credSubject := verifiable.VINSubject{}
	if err := json.Unmarshal(cred.CredentialSubject, &credSubject); err != nil {
		r.logger.Error().Err(err).Msg("failed to unmarshal VIN credential subject")
		return nil, errors.New("internal error")
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
		r.logger.Error().Err(err).Msg("failed to marshal VIN VC")
		return nil, errors.New("internal error")
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
	vehicleDID := cloudevent.NFTDID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         vehicleTokenID,
	}.String()
	opts := &grpc.SearchOptions{
		DataVersion: &wrapperspb.StringValue{Value: r.pomVCDataType},
		Type:        &wrapperspb.StringValue{Value: cloudevent.TypeVerifableCredential},
		Subject:     &wrapperspb.StringValue{Value: vehicleDID},
	}
	dataObj, err := r.indexService.GetLatestCloudEvent(ctx, opts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("failed to get latest POM VC data")
		return nil, errors.New("internal error")
	}
	cred := verifiable.Credential{}
	if err := json.Unmarshal(dataObj.Data, &cred); err != nil {
		r.logger.Error().Err(err).Msg("failed to unmarshal POM VC")
		return nil, errors.New("internal error")
	}

	var createdAt *time.Time
	if issuanceDate, err := time.Parse(time.RFC3339, cred.ValidFrom); err == nil {
		createdAt = &issuanceDate
	}
	credSubject := verifiable.POMSubject{}
	if err := json.Unmarshal(cred.CredentialSubject, &credSubject); err != nil {
		r.logger.Error().Err(err).Msg("failed to unmarshal POM credential subject")
		return nil, errors.New("internal error")
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
