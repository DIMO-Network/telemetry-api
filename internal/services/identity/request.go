package identity

import (
	"context"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=./request.go -destination=request_mocks.go -package=identity
type IdentityService interface {
	AftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*aftermarketDeviceResponse, error)
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

func (i *identityAPI) AftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*aftermarketDeviceResponse, error) {
	return aftermarketDevice(ctx, i.client, AftermarketDeviceBy{
		TokenId: tokenID,
		Address: address,
		Serial:  serial,
	})
}
