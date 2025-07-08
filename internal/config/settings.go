package config

import (
	"github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"
	"github.com/ethereum/go-ethereum/common"
)

// Settings contains the application config.
type Settings struct {
	Port                         int             `yaml:"PORT"`
	MonPort                      int             `yaml:"MON_PORT"`
	EnablePprof                  bool            `yaml:"ENABLE_PPROF"`
	Clickhouse                   config.Settings `yaml:",inline"`
	TokenExchangeJWTKeySetURL    string          `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	TokenExchangeIssuer          string          `yaml:"TOKEN_EXCHANGE_ISSUER_URL"`
	VehicleNFTAddress            common.Address  `yaml:"VEHICLE_NFT_ADDRESS"`
	MaxRequestDuration           string          `yaml:"MAX_REQUEST_DURATION"`
	POMVCDataVersion             string          `yaml:"POMVC_DATA_VERSION"`
	ManufacturerNFTAddress       common.Address  `yaml:"MANUFACTURER_NFT_ADDRESS"`
	IdentityAPIURL               string          `yaml:"IDENTITY_API_URL"`
	IdentityAPIReqTimeoutSeconds int             `yaml:"IDENTITY_API_REQUEST_TIMEOUT_SECONDS"`
	DeviceLastSeenBinHrs         int64           `yaml:"DEVICE_LAST_SEEN_BIN_HOURS"`
	ChainID                      uint64          `yaml:"DIMO_REGISTRY_CHAIN_ID"`
	FetchAPIGRPCEndpoint         string          `yaml:"FETCH_API_GRPC_ENDPOINT"`
	CreditTrackerEndpoint        string          `yaml:"CREDIT_TRACKER_ENDPOINT"`
	StorageNodeDevLicense        common.Address  `yaml:"STORAGE_NODE_DEV_LICENSE"`
	VINDataVersion               string          `yaml:"VIN_DATA_VERSION"`

	// TODO: remove these once manual vinvc are migrated to attestation
	VINVCDataVersion string `yaml:"VINVC_DATA_VERSION"`
}
