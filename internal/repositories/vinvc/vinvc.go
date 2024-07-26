package vinvc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DIMO-Network/attestation-api/pkg/verifiable"
	"github.com/DIMO-Network/nameindexer"
	"github.com/DIMO-Network/nameindexer/pkg/clickhouse/service"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/rs/zerolog"
)

type Repository struct {
	logger       *zerolog.Logger
	indexService *service.Service
	dataType     string
}

// New creates a new instance of Service.
func New(chConn clickhouse.Conn, objGetter service.ObjectGetter, bucketName, vinvcDataType string, logger *zerolog.Logger) *Repository {
	return &Repository{
		indexService: service.New(chConn, objGetter, bucketName),
		dataType:     vinvcDataType,
		logger:       logger,
	}
}

// GetLatestVC fetches the latest fingerprint message from S3.
func (s *Repository) GetLatestVC(ctx context.Context, vehicleTokenID uint32) (*model.Vinvc, error) {
	subject := nameindexer.Subject{
		TokenID: &vehicleTokenID,
	}
	data, err := s.indexService.GetLatestData(ctx, s.dataType, subject)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		s.logger.Error().Err(err).Msg("failed to get latest data")
		return nil, errors.New("internal error")
	}
	msg := verifiable.Credential{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fingerprint message: %w", err)
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
		return nil, fmt.Errorf("failed to unmarshal credential subject: %w", err)
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
		RawVc:                  string(data),
		Vin:                    vin,
		RecordedBy:             recordedBy,
		RecordedAt:             recordedAt,
		CountryCode:            countryCode,
		VehicleContractAddress: vehicleContractAddress,
		VehicleTokenID:         &tokenIDInt,
	}, nil
}
