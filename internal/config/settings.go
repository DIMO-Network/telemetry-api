package config

import (
	"github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/ethereum/go-ethereum/common"
)

// Settings contains the application config.
type Settings struct {
	Port                         int             `yaml:"PORT"`
	MonPort                      int             `yaml:"MON_PORT"`
	CLickhouse                   config.Settings `yaml:",inline"`
	TokenExchangeJWTKeySetURL    string          `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	TokenExchangeIssuer          string          `yaml:"TOKEN_EXCHANGE_ISSUER_URL"`
	VehicleNFTAddress            common.Address  `yaml:"VEHICLE_NFT_ADDRESS"`
	MaxRequestDuration           string          `yaml:"MAX_REQUEST_DURATION"`
	VCBucket                     string          `yaml:"VC_BUCKET"`
	VINVCDataType                string          `yaml:"VINVC_DATA_TYPE"`
	POMVCDataType                string          `yaml:"POMVC_DATA_TYPE"`
	ManufacturerNFTAddress       common.Address  `yaml:"MANUFACTURER_NFT_ADDRESS"`
	IdentityAPIURL               string          `yaml:"IDENTITY_API_URL"`
	IdentityAPIReqTimeoutSeconds int             `yaml:"IDENTITY_API_REQUEST_TIMEOUT_SECONDS"`
	DeviceLastSeenBinHrs         int64           `yaml:"DEVICE_LAST_SEEN_BIN_HOURS"`
	ChainID                      int             `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	FetchAPIGRPCEndpoint         string          `yaml:"FETCH_API_GRPC_ENDPOINT"`
}
