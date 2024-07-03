package vinvc

import (
	"context"
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
		s.logger.Error().Err(err).Msg("failed to get latest data")
		return nil, errors.New("internal error")
	}
	msg := verifiable.Credential{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fingerprint message: %w", err)
	}

	var expiresAt *time.Time
	if expirationDate, err := time.Parse(time.RFC3339, msg.ExpirationDate); err == nil {
		expiresAt = &expirationDate
	}
	var createdAt *time.Time
	if issuanceDate, err := time.Parse(time.RFC3339, msg.IssuanceDate); err == nil {
		createdAt = &issuanceDate
	}
	credSubject := verifiable.VINSubject{}
	var vin *string
	if err := json.Unmarshal(msg.CredentialSubject, &credSubject); err == nil {
		vin = &credSubject.VehicleIdentificationNumber
	}

	vc := model.Vinvc{
		IssuanceDate:   createdAt,
		ExpirationDate: expiresAt,
		RawVc:          string(data),
		Vin:            vin,
	}
	return &vc, nil
}
