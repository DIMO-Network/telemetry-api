package vinvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
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
	msg := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fingerprint message: %w", err)
	}

	var proof *string
	if proofVal, ok := msg["proof"]; ok {
		proof = new(string)
		*proof = string(proofVal)
	}

	var createdAt *time.Time
	if err := json.Unmarshal(msg["issuanceDate"], &createdAt); err != nil {
		createdAt = nil
	}

	vc := model.Vinvc{
		CreatedAt: createdAt,
		RawVc:     string(data),
		RawProof:  proof,
	}
	return &vc, nil
}
