package config

// Settings contains the application config.
type Settings struct {
	Port                          int    `yaml:"PORT"`
	MonPort                       int    `yaml:"MON_PORT"`
	ClickHouseHost                string `yaml:"CLICKHOUSE_HOST"`
	ClickHouseTCPPort             int    `yaml:"CLICKHOUSE_TCP_PORT"`
	ClickHouseUser                string `yaml:"CLICKHOUSE_USER"`
	ClickHousePassword            string `yaml:"CLICKHOUSE_PASSWORD"`
	ClickHouseDatabase            string `yaml:"CLICKHOUSE_DATABASE"`
	CLickHouseMaxExecutionTimeSec int    `yaml:"CLICKHOUSE_MAX_EXECUTION_TIME_SECONDS"`
	TokenExchangeJWTKeySetURL     string `yaml:"TOKEN_EXCHANGE_JWK_KEY_SET_URL"`
	TokenExchangeIssuer           string `yaml:"TOKEN_EXCHANGE_ISSUER_URL"`
	VehicleNFTAddress             string `yaml:"VEHICLE_NFT_ADDRESS"`
	MaxRequestDuration            string `yaml:"MAX_REQUEST_DURATION"`
}
