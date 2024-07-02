package config

import "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"

// Settings contains the application config.
type Settings struct {
	Port                      int             `yaml:"PORT"`
	MonPort                   int             `yaml:"MON_PORT"`
	CLickhouse                config.Settings `yaml:",inline"`
	TokenExchangeJWTKeySetURL string          `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	TokenExchangeIssuer       string          `yaml:"TOKEN_EXCHANGE_ISSUER_URL"`
	VehicleNFTAddress         string          `yaml:"VEHICLE_NFT_ADDRESS"`
	MaxRequestDuration        string          `yaml:"MAX_REQUEST_DURATION"`
}
