package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	DbPassword       string
	ImagineApiKey    string
	UseApiKey        string
	GoogleMapsApiKey string
	PolymarketAPIKey string
}

type PublicConfig struct {
	DbHost   string `mapstructure:"DB_HOST"`
	DbUser   string `mapstructure:"DB_USER"`
	DbPort   string `mapstructure:"DB_PORT"`
	DbName   string `mapstructure:"DB_NAME"`
	RedisUrl string `mapstructure:"REDIS_URL"`
	ChainID  int64  `mapstructure:"CHAIN_ID"`
	RPCURL   string `mapstructure:"RPC_URL"`

	PolymarketTradesURL                    string  `mapstructure:"POLYMARKET_TRADES_URL"`
	PolymarketBaseURL                      string  `mapstructure:"POLYMARKET_BASE_URL"`
	PolymarketTradesPath                   string  `mapstructure:"POLYMARKET_TRADES_PATH"`
	PolymarketAlertToNumber                string  `mapstructure:"POLYMARKET_ALERT_TO_NUMBER"`
	PolymarketAlertFromNumber              string  `mapstructure:"POLYMARKET_ALERT_FROM_NUMBER"`
	PolymarketSuspiciousNotionalThreshold  float64 `mapstructure:"POLYMARKET_SUSPICIOUS_NOTIONAL_THRESHOLD"`
	PolymarketSuspiciousSizeThreshold      float64 `mapstructure:"POLYMARKET_SUSPICIOUS_SIZE_THRESHOLD"`
	PolymarketTradesLimit                  int     `mapstructure:"POLYMARKET_TRADES_LIMIT"`
}

type Config struct {
	Public PublicConfig
	Secret SecretConfig
}

type loadConfigParams struct {
	Name string
	Type string
	Path string
}

func ParseFlagsAndGetConfig() (*Config, error) {
	var params loadConfigParams
	flag.StringVar(&params.Name, "config-name", "live", "The name of the config file.")
	flag.StringVar(&params.Type, "config-type", "env", "The type of the config file.")
	flag.StringVar(&params.Path, "config-path", ".", "The path of the config file.")
	flag.Parse()

	viper.AddConfigPath(params.Path)
	viper.SetConfigName(params.Name)
	viper.SetConfigType(params.Type)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	publicCfg := PublicConfig{}

	if err := viper.Unmarshal(&publicCfg); err != nil {
		return nil, err
	}

	// Parse CHAIN_ID from environment if not set via viper
	if publicCfg.ChainID == 0 {
		if chainIDStr := os.Getenv("CHAIN_ID"); chainIDStr != "" {
			chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
			if err == nil {
				publicCfg.ChainID = chainID
			}
		}
	}

	// Get RPC_URL from environment if not set via viper
	if publicCfg.RPCURL == "" {
		publicCfg.RPCURL = os.Getenv("RPC_URL")
	}

	if publicCfg.PolymarketAlertToNumber == "" {
		publicCfg.PolymarketAlertToNumber = "14407858475"
	}
	if publicCfg.PolymarketSuspiciousNotionalThreshold == 0 {
		publicCfg.PolymarketSuspiciousNotionalThreshold = 1000
	}
	if publicCfg.PolymarketTradesLimit == 0 {
		publicCfg.PolymarketTradesLimit = 100
	}

	return &Config{
		Secret: SecretConfig{
			DbPassword:       os.Getenv("DB_PASSWORD"),
			ImagineApiKey:    os.Getenv("IMAGINE_API_KEY"),
			UseApiKey:        os.Getenv("USE_API_KEY"),
			GoogleMapsApiKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
			PolymarketAPIKey: os.Getenv("POLYMARKET_API_KEY"),
		},
		Public: publicCfg,
	}, nil
}
