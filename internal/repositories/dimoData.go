package repositories

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/DIMO-Network/model-garage/pkg/vss"
	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/DIMO-Network/telemetry-api/internal/graph/model"
	"github.com/DIMO-Network/telemetry-api/internal/service/deviceapi"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	tableName = "dimo"
)

var (
	bySubjectClause = vss.FieldSubject + " = ?"
	bySinceClause   = vss.FieldTimestamp + " >= ?"
	byUntilClause   = vss.FieldTimestamp + " <= ?"
	defaultTimeout  = 30 * time.Second
)

// Repository is the base repository for all repositories.
type Repository struct {
	conn      clickhouse.Conn
	Log       *zerolog.Logger
	deviceAPI *deviceapi.Service
}

// NewRepository creates a new base repository.
func NewRepository(logger *zerolog.Logger, settings config.Settings) (*Repository, error) {
	addr := fmt.Sprintf("%s:%d", settings.ClickHouseHost, settings.ClickHouseTCPPort)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: settings.ClickHouseUser,
			Password: settings.ClickHousePassword,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}
	devicesConn, err := grpc.Dial(settings.DevicesAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial devices api: %w", err)
	}
	deviceAPI := deviceapi.NewService(devicesConn)
	return &Repository{
		conn:      conn,
		Log:       logger,
		deviceAPI: deviceAPI,
	}, nil
}

// GetDIMOData returns the DIMO data from the database.
func (r *Repository) GetDIMOData(ctx context.Context, cols []string, page model.PageSelection, filterBy *model.DimosFilter) (*model.DimoConnection, error) {
	err := validateFirstLast(page.First, page.Last, MaxPageSize)
	if err != nil {
		return nil, err
	}
	var dimoConn model.DimoConnection
	// create where clause
	where, args, err := r.buildWhereClause(ctx, filterBy, page)
	if err != nil {
		return nil, err
	}

	dimoConn.TotalCount, err = r.getTotalCount(where, args)
	if err != nil {
		return nil, err
	}

	limit, order := getLimitAndOrder(page)

	// always get the timestamp for pagination
	cols = append(cols, vss.FieldTimestamp)

	colStr := strings.Join(cols, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY %s %s LIMIT %d ", colStr, tableName, where, vss.FieldTimestamp, order, limit)

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying clickhouse: %w", err)
	}

	err = populateDimoConn(rows, page, &dimoConn)
	if err != nil {
		return nil, err
	}

	return &dimoConn, nil
}

// buildWhereClause builds the where clause for the query. returns the where clause and the arguments.
func (r *Repository) buildWhereClause(ctx context.Context, filterBy *model.DimosFilter, page model.PageSelection) (string, []any, error) {
	var where []string
	var args []any
	if filterBy == nil {
		return "1", args, nil
	}
	subj, err := r.SubjectFromTokenID(ctx, filterBy.TokenID)
	if err != nil {
		return "", nil, err
	}
	where = append(where, bySubjectClause)
	args = append(args, subj)

	// filter since
	where = append(where, bySinceClause)
	since := filterBy.Since
	if page.After != nil {
		afterCurs, err := decodeCusor(*page.After)
		if err != nil {
			return "", nil, err
		}
		if afterCurs.After(since) {
			since = afterCurs
		}
	}
	args = append(args, since)

	// filter until
	where = append(where, byUntilClause)
	until := filterBy.Until
	if page.Before != nil {
		beforeCurs, err := decodeCusor(*page.Before)
		if err != nil {
			return "", nil, err
		}
		if beforeCurs.Before(until) {
			until = beforeCurs
		}
	}
	args = append(args, until)

	return strings.Join(where, " AND "), args, nil
}

// SubjectFromTokenID returns the subject from a token ID.
func (r *Repository) SubjectFromTokenID(ctx context.Context, tokenID int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	userDevice, err := r.deviceAPI.GetUserDeviceByTokenID(ctx, int64(tokenID))
	if err != nil {
		return "", fmt.Errorf("failed to get user device by token id: %w", err)
	}
	id := strings.Trim(userDevice.Id, " ")
	return id, nil
}

func (r *Repository) getTotalCount(where string, args []any) (int, error) {
	// get total count
	var count uint64
	err := r.conn.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", tableName, where), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}
	return int(count), nil
}

// ToAPI converts a DIMO model to a DIMO API model.
func ToAPI(dimo *vss.Dimo, id string, tokenID int) *model.DIMOData {
	return &model.DIMOData{
		ID:      id,
		TokenID: tokenID,
		Dimo:    *dimo,
	}
}

func decodeCusor(cursor string) (time.Time, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, fmt.Errorf("error decoding cursor: %w", err)
	}
	bigInt := new(big.Int).SetBytes(data)
	return time.Unix(bigInt.Int64(), 0).UTC(), nil
}

func encodeCursor(t time.Time) (string, error) {
	bigInt := big.NewInt(t.Unix())
	return base64.StdEncoding.EncodeToString(bigInt.Bytes()), nil
}

func getLimitAndOrder(page model.PageSelection) (int, string) {
	limit := 0
	var order string
	if page.First != nil {
		limit = int(*page.First) + 1
		order = "ASC"
	} else {
		limit = int(*page.Last) + 1
		order = "DESC"
	}
	return limit, order
}

func populateDimoConn(rows driver.Rows, page model.PageSelection, dimoConn *model.DimoConnection) error {
	for rows.Next() {
		dimoData := vss.Dimo{}
		if err := rows.ScanStruct(&dimoData); err != nil {
			return fmt.Errorf("error scanning struct: %w", err)
		}
		dimoAPI := ToAPI(&dimoData, "", 0)
		curs, err := encodeCursor(*dimoData.Timestamp)
		if err != nil {
			return fmt.Errorf("error encoding global token id: %w", err)
		}
		dimoConn.Edges = append(dimoConn.Edges, &model.DimoEdge{
			Node:   dimoAPI,
			Cursor: curs,
		})
		dimoConn.Nodes = append(dimoConn.Nodes, dimoAPI)
	}
	dimoConn.PageInfo = &model.PageInfo{}
	_ = rows.Close()
	if page.First != nil {
		if len(dimoConn.Nodes) > *page.First {
			dimoConn.PageInfo.HasNextPage = true
			dimoConn.Nodes = dimoConn.Nodes[:*page.First]
			dimoConn.Edges = dimoConn.Edges[:*page.First]
		}
	} else {
		if len(dimoConn.Nodes) > *page.Last {
			dimoConn.PageInfo.HasPreviousPage = true
			dimoConn.Nodes = dimoConn.Nodes[:*page.Last]
			dimoConn.Edges = dimoConn.Edges[:*page.Last]
			slices.Reverse(dimoConn.Nodes)
			slices.Reverse(dimoConn.Edges)
		}
	}
	if len(dimoConn.Nodes) > 0 {
		dimoConn.PageInfo.StartCursor = &dimoConn.Edges[0].Cursor
		dimoConn.PageInfo.EndCursor = &dimoConn.Edges[len(dimoConn.Edges)-1].Cursor
	}

	return rows.Err()
}
