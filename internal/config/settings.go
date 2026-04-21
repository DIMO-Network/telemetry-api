package config

import (
	"github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/ethereum/go-ethereum/common"
)

// Settings contains the application config.
type Settings struct {
	LogLevel                  string          `yaml:"LOG_LEVEL"`
	Port                      int             `yaml:"PORT"`
	MonPort                   int             `yaml:"MON_PORT"`
	EnablePprof               bool            `yaml:"ENABLE_PPROF"`
	Clickhouse                config.Settings `yaml:",inline"`
	TokenExchangeJWTKeySetURL string          `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	TokenExchangeIssuer       string          `yaml:"TOKEN_EXCHANGE_ISSUER_URL"`
	VehicleNFTAddress         common.Address  `yaml:"VEHICLE_NFT_ADDRESS"`
	MaxRequestDuration        string          `yaml:"MAX_REQUEST_DURATION"`
	LatestSignalsLookbackDays int64           `yaml:"LATEST_SIGNALS_LOOKBACK_DAYS"`
	ChainID                   uint64          `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	FetchAPIGRPCEndpoint      string          `yaml:"FETCH_API_GRPC_ENDPOINT"`
	CreditTrackerEndpoint     string          `yaml:"CREDIT_TRACKER_ENDPOINT"`
	StorageNodeDevLicense     common.Address  `yaml:"STORAGE_NODE_DEV_LICENSE"`
	VINDataVersion            string          `yaml:"VIN_DATA_VERSION"`

	// DQEndpoint is the URL of the dq GraphQL endpoint (e.g. http://dq:3000/query).
	// When set, signal/event/segment queries are proxied to dq instead of ClickHouse.
	DQEndpoint string `yaml:"DQ_ENDPOINT"`
}
