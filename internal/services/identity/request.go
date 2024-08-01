package identity

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=./request.go -destination=request_mocks.go -package=identity
type IdentityService interface {
	AftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*DeviceInfos, error)
}

type identityAPI struct {
	client graphql.Client
}

func NewService(url string, timeout int) IdentityService {
	graphqlClient := graphql.NewClient(url, &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	})
	return &identityAPI{
		client: graphqlClient,
	}
}

func (i *identityAPI) AftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*DeviceInfos, error) {
	resp, err := aftermarketDevice(ctx, i.client, AftermarketDeviceBy{
		TokenId: tokenID,
		Address: address,
		Serial:  serial,
	})
	if err != nil {
		return nil, err
	}

	if resp.AftermarketDevice.Vehicle == nil {
		return nil, fmt.Errorf("no vehicle attached to device")
	}

	return &DeviceInfos{
		ManufacturerTokenID:      resp.AftermarketDevice.Manufacturer.TokenId,
		VehicleTokenID:           resp.AftermarketDevice.Vehicle.TokenId,
		AftermarketDeviceTokenID: resp.AftermarketDevice.TokenId,
		ManufacturerName:         resp.AftermarketDevice.Manufacturer.Name,
	}, nil
}

type DeviceInfos struct {
	ManufacturerTokenID      int    `json:"manufacturerTokenId"`
	VehicleTokenID           int    `json:"vehicleTokenId"`
	AftermarketDeviceTokenID int    `json:"aftermarketDeviceTokenId"`
	ManufacturerName         string `json:"manufacturerName"`
}
