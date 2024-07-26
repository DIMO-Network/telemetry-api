package config

import "github.com/DIMO-Network/clickhouse-infra/pkg/connect/config"

// Settings contains the application config.
type Settings struct {
	Port                        int             `yaml:"PORT"`
	MonPort                     int             `yaml:"MON_PORT"`
	CLickhouse                  config.Settings `yaml:",inline"`
	ClickhouseFileIndexDatabase string          `yaml:"CLICKHOUSE_FILE_INDEX_DATABASE"`
	TokenExchangeJWTKeySetURL   string          `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	TokenExchangeIssuer         string          `yaml:"TOKEN_EXCHANGE_ISSUER_URL"`
	VehicleNFTAddress           string          `yaml:"VEHICLE_NFT_ADDRESS"`
	MaxRequestDuration          string          `yaml:"MAX_REQUEST_DURATION"`
	S3AWSRegion                 string          `yaml:"S3_AWS_REGION"`
	S3AWSAccessKeyID            string          `yaml:"S3_AWS_ACCESS_KEY_ID"`
	S3AWSSecretAccessKey        string          `yaml:"S3_AWS_SECRET_ACCESS_KEY"`
	VINVCBucket                 string          `yaml:"VINVC_BUCKET"`
	VINVCDataType               string          `yaml:"VINVC_DATA_TYPE"`
	ManufacturerNFTAddress      string          `yaml:"MANUFACTURER_NFT_ADDRESS"`
	IdentityAPIURL              string          `yaml:"IDENTITY_API_URL"`
	IdentityAPIRequestTimeout   int             `yaml:"IDENTITY_API_REQUEST_TIMEOUT"`
}
