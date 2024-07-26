package services

import (
	"context"
	"net/http"
	"time"

	"github.com/DIMO-Network/telemetry-api/internal/config"
	"github.com/Khan/genqlient/graphql"
)

type IdentityService interface {
	AftermarketDevice(ctx context.Context, by AftermarketDeviceBy) (*aftermarketDeviceResponse, error)
}

type identityAPI struct {
	client graphql.Client
}

func NewIdentityService(settings *config.Settings) IdentityService {
	graphqlClient := graphql.NewClient(settings.IdentityAPIURL, &http.Client{
		Timeout: time.Duration(settings.IdentityAPIRequestTimeout) * time.Second,
	})
	return &identityAPI{
		client: graphqlClient,
	}
}

func (i *identityAPI) AftermarketDevice(ctx context.Context, by AftermarketDeviceBy) (*aftermarketDeviceResponse, error) {
	return aftermarketDevice(ctx, i.client, by)
}
