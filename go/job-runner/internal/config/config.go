package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	DbPassword              string
	ImagineApiKey           string
	UseApiKey               string
	GoogleMapsApiKey        string
	PolymarketAPIKey        string
	PolymarketAPISecret     string
	PolymarketAPIPassphrase string
	PolymarketAddress       string
}

type PublicConfig struct {
	DbHost   string `mapstructure:"DB_HOST"`
	DbUser   string `mapstructure:"DB_USER"`
	DbPort   string `mapstructure:"DB_PORT"`
	DbName   string `mapstructure:"DB_NAME"`
	RedisUrl string `mapstructure:"REDIS_URL"`
	ChainID  int64  `mapstructure:"CHAIN_ID"`
	RPCURL   string `mapstructure:"RPC_URL"`

	PolymarketTradesURL                   string  `mapstructure:"POLYMARKET_TRADES_URL"`
	PolymarketBaseURL                     string  `mapstructure:"POLYMARKET_BASE_URL"`
	PolymarketTradesPath                  string  `mapstructure:"POLYMARKET_TRADES_PATH"`
	PolymarketAlertToNumber               string  `mapstructure:"POLYMARKET_ALERT_TO_NUMBER"`
	PolymarketAlertFromNumber             string  `mapstructure:"POLYMARKET_ALERT_FROM_NUMBER"`
	PolymarketSuspiciousNotionalThreshold float64 `mapstructure:"POLYMARKET_SUSPICIOUS_NOTIONAL_THRESHOLD"`
	PolymarketSuspiciousSizeThreshold     float64 `mapstructure:"POLYMARKET_SUSPICIOUS_SIZE_THRESHOLD"`
	PolymarketTradesLimit                 int     `mapstructure:"POLYMARKET_TRADES_LIMIT"`

	// reef-site generation/slicing (R-2.4/R-2.5/R-2.7) — same env var names
	// as go/reef-site/internal/config so one terraform env block configures
	// both processes consistently.
	ReefOpenSCADBin                    string  `mapstructure:"REEF_OPENSCAD_BIN"`
	ReefSlicerBin                      string  `mapstructure:"REEF_SLICER_BIN"`
	ReefSubprocessTimeoutSec           int     `mapstructure:"REEF_SUBPROCESS_TIMEOUT_SEC"`
	ReefSubprocessMemoryMB             int     `mapstructure:"REEF_SUBPROCESS_MEMORY_MB"`
	ReefS3Bucket                       string  `mapstructure:"REEF_S3_BUCKET"`
	ReefAwsRegion                      string  `mapstructure:"REEF_AWS_REGION"`
	ReefPriceSetupFeeCents             int64   `mapstructure:"REEF_PRICE_SETUP_FEE_CENTS"`
	ReefPriceMaterialRateCentsPerGram  float64 `mapstructure:"REEF_PRICE_MATERIAL_RATE_CENTS_PER_GRAM"`
	ReefPriceMachineRateCentsPerMinute float64 `mapstructure:"REEF_PRICE_MACHINE_RATE_CENTS_PER_MINUTE"`
	ReefPriceFulfillmentFeeCents       int64   `mapstructure:"REEF_PRICE_FULFILLMENT_FEE_CENTS"`
	ReefPriceMarginMultiplier          float64 `mapstructure:"REEF_PRICE_MARGIN_MULTIPLIER"`
	ReefMaxBboxMm                      float64 `mapstructure:"REEF_MAX_BBOX_MM"`
	ReefMinWallMm                      float64 `mapstructure:"REEF_MIN_WALL_MM"`
	ReefMaxPrintTimeS                  int64   `mapstructure:"REEF_MAX_PRINT_TIME_S"`
	ReefMaxWeightG                     float64 `mapstructure:"REEF_MAX_WEIGHT_G"`
	ReefMinDrainPathMm                 float64 `mapstructure:"REEF_MIN_DRAIN_PATH_MM"`
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

	if publicCfg.ReefOpenSCADBin == "" {
		publicCfg.ReefOpenSCADBin = "openscad"
	}
	if publicCfg.ReefSlicerBin == "" {
		publicCfg.ReefSlicerBin = "prusa-slicer"
	}
	if publicCfg.ReefSubprocessTimeoutSec == 0 {
		publicCfg.ReefSubprocessTimeoutSec = 120
	}
	if publicCfg.ReefSubprocessMemoryMB == 0 {
		publicCfg.ReefSubprocessMemoryMB = 1536
	}
	if publicCfg.ReefS3Bucket == "" {
		publicCfg.ReefS3Bucket = "reef-site-artifacts"
	}
	if publicCfg.ReefAwsRegion == "" {
		publicCfg.ReefAwsRegion = "us-east-1"
	}
	if publicCfg.ReefPriceSetupFeeCents == 0 {
		publicCfg.ReefPriceSetupFeeCents = 300
	}
	if publicCfg.ReefPriceMaterialRateCentsPerGram == 0 {
		publicCfg.ReefPriceMaterialRateCentsPerGram = 8.0
	}
	if publicCfg.ReefPriceMachineRateCentsPerMinute == 0 {
		publicCfg.ReefPriceMachineRateCentsPerMinute = 4.0
	}
	if publicCfg.ReefPriceFulfillmentFeeCents == 0 {
		publicCfg.ReefPriceFulfillmentFeeCents = 250
	}
	if publicCfg.ReefPriceMarginMultiplier == 0 {
		publicCfg.ReefPriceMarginMultiplier = 1.8
	}
	if publicCfg.ReefMaxBboxMm == 0 {
		publicCfg.ReefMaxBboxMm = 210.0
	}
	if publicCfg.ReefMinWallMm == 0 {
		publicCfg.ReefMinWallMm = 2.0
	}
	if publicCfg.ReefMaxPrintTimeS == 0 {
		publicCfg.ReefMaxPrintTimeS = 4 * 60 * 60
	}
	if publicCfg.ReefMaxWeightG == 0 {
		publicCfg.ReefMaxWeightG = 250.0
	}
	if publicCfg.ReefMinDrainPathMm == 0 {
		publicCfg.ReefMinDrainPathMm = 4.0
	}

	return &Config{
		Secret: SecretConfig{
			DbPassword:              os.Getenv("DB_PASSWORD"),
			ImagineApiKey:           os.Getenv("IMAGINE_API_KEY"),
			UseApiKey:               os.Getenv("USE_API_KEY"),
			GoogleMapsApiKey:        os.Getenv("GOOGLE_MAPS_API_KEY"),
			PolymarketAPIKey:        os.Getenv("POLYMARKET_API_KEY"),
			PolymarketAPISecret:     os.Getenv("POLYMARKET_API_SECRET"),
			PolymarketAPIPassphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
			PolymarketAddress:       os.Getenv("POLYMARKET_ADDRESS"),
		},
		Public: publicCfg,
	}, nil
}
