package vc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/attestation-api/pkg/verifiable"
	"github.com/DIMO-Network/model-garage/pkg/cloudevent"
	"github.com/DIMO-Network/nameindexer"
	"github.com/DIMO-Network/nameindexer/pkg/clickhouse/indexrepo"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
)

type Repository struct {
	logger         *zerolog.Logger
	indexService   *indexrepo.Service
	vcBucket       string
	vinVCDataType  string
	pomVCDataType  string
	chainID        uint64
	vehicleAddress common.Address
}

// New creates a new instance of Service.
func New(chConn clickhouse.Conn, objGetter indexrepo.ObjectGetter, vcBucketName, vinVCDataType, pomVCDataType string, chainID uint64, vehicleAddress common.Address, logger *zerolog.Logger) *Repository {
	return &Repository{
		indexService:   indexrepo.New(chConn, objGetter),
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
	filler := nameindexer.FillerVerifiableCredential
	vehicleDID := cloudevent.NFTDID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         vehicleTokenID,
	}
	opts := indexrepo.CloudEventSearchOptions{
		DataType:      &r.vinVCDataType,
		PrimaryFiller: &filler,
		Subject:       &vehicleDID,
	}
	dataObj, err := r.indexService.GetLatestCloudEventData(ctx, r.vcBucket, opts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("failed to get latest VIN VC data")
		return nil, errors.New("internal error")
	}
	msg := verifiable.Credential{}
	if err := json.Unmarshal(dataObj.Data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VIN VC: %w", err)
	}

	var expiresAt *time.Time
	if expirationDate, err := time.Parse(time.RFC3339, msg.ValidTo); err == nil {
		expiresAt = &expirationDate
	}
	var createdAt *time.Time
	if issuanceDate, err := time.Parse(time.RFC3339, msg.ValidFrom); err == nil {
		createdAt = &issuanceDate
	}
	credSubject := verifiable.VINSubject{}
	if err := json.Unmarshal(msg.CredentialSubject, &credSubject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VIN credential subject: %w", err)
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
	tokenIDInt := int(credSubject.VehicleTokenID)

	return &model.Vinvc{
		ValidFrom:              createdAt,
		ValidTo:                expiresAt,
		RawVc:                  string(dataObj.Data),
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
	filler := nameindexer.FillerVerifiableCredential
	vehicleDID := cloudevent.NFTDID{
		ChainID:         r.chainID,
		ContractAddress: r.vehicleAddress,
		TokenID:         vehicleTokenID,
	}
	opts := indexrepo.CloudEventSearchOptions{
		DataType:      &r.pomVCDataType,
		PrimaryFiller: &filler,
		Subject:       &vehicleDID,
	}
	dataObj, err := r.indexService.GetLatestCloudEventData(ctx, r.vcBucket, opts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("failed to get latest POM VC data")
		return nil, errors.New("internal error")
	}
	msg := verifiable.Credential{}
	if err := json.Unmarshal(dataObj.Data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal POM VC: %w", err)
	}

	var createdAt *time.Time
	if issuanceDate, err := time.Parse(time.RFC3339, msg.ValidFrom); err == nil {
		createdAt = &issuanceDate
	}
	credSubject := verifiable.POMSubject{}
	if err := json.Unmarshal(msg.CredentialSubject, &credSubject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal POM credential subject: %w", err)
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
