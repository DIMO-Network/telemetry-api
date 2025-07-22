package identity

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/DIMO-Network/server-garage/pkg/gql/errorhandler"
	"github.com/Khan/genqlient/graphql"
	"github.com/ethereum/go-ethereum/common"
)

type APIClient struct {
	client graphql.Client
}

func NewService(url string, timeout int) *APIClient {
	graphqlClient := graphql.NewClient(url, &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	})
	return &APIClient{
		client: graphqlClient,
	}
}

func (i *APIClient) GetAftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*DeviceInfos, error) {
	resp, err := aftermarketDevice(ctx, i.client, AftermarketDeviceBy{
		TokenId: tokenID,
		Address: address,
		Serial:  serial,
	})
	if err != nil {
		return nil, errorhandler.NewInternalErrorWithMsg(ctx, err, "internal identity service error")
	}

	if resp.AftermarketDevice.Vehicle == nil {
		return nil, errorhandler.NewBadRequestError(ctx, errors.New("no vehicle attached to device"))
	}

	return &DeviceInfos{
		ManufacturerTokenID:      resp.AftermarketDevice.Manufacturer.TokenId,
		VehicleTokenID:           resp.AftermarketDevice.Vehicle.TokenId,
		AftermarketDeviceTokenID: resp.AftermarketDevice.TokenId,
		ManufacturerName:         resp.AftermarketDevice.Manufacturer.Name,
	}, nil
}

type DeviceInfos struct {
	ManufacturerTokenID      int
	VehicleTokenID           int
	AftermarketDeviceTokenID int
	ManufacturerName         string
}
