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
	ReefFilamentDensityGCm3            float64 `mapstructure:"REEF_FILAMENT_DENSITY_G_CM3"`
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
	ReefMaxSupportMaterialPct          float64 `mapstructure:"REEF_MAX_SUPPORT_MATERIAL_PCT"`

	// bgi-site generation/slicing — same shape as the Reef* block above,
	// own env var names/values so the two products tune independently. No
	// Bgi*MinDrainPathMm: bgi trays run with SealedVoidRuleEnabled=false
	// unconditionally (open-top wells have no cavity story at all), so
	// there's nothing for a drain-path threshold to gate.
	BgiOpenSCADBin                    string  `mapstructure:"BGI_OPENSCAD_BIN"`
	BgiSlicerBin                      string  `mapstructure:"BGI_SLICER_BIN"`
	BgiFilamentDensityGCm3            float64 `mapstructure:"BGI_FILAMENT_DENSITY_G_CM3"`
	BgiSubprocessTimeoutSec           int     `mapstructure:"BGI_SUBPROCESS_TIMEOUT_SEC"`
	BgiSubprocessMemoryMB             int     `mapstructure:"BGI_SUBPROCESS_MEMORY_MB"`
	BgiS3Bucket                       string  `mapstructure:"BGI_S3_BUCKET"`
	BgiAwsRegion                      string  `mapstructure:"BGI_AWS_REGION"`
	BgiPriceSetupFeeCents             int64   `mapstructure:"BGI_PRICE_SETUP_FEE_CENTS"`
	BgiPriceMaterialRateCentsPerGram  float64 `mapstructure:"BGI_PRICE_MATERIAL_RATE_CENTS_PER_GRAM"`
	BgiPriceMachineRateCentsPerMinute float64 `mapstructure:"BGI_PRICE_MACHINE_RATE_CENTS_PER_MINUTE"`
	BgiPriceFulfillmentFeeCents       int64   `mapstructure:"BGI_PRICE_FULFILLMENT_FEE_CENTS"`
	BgiPriceMarginMultiplier          float64 `mapstructure:"BGI_PRICE_MARGIN_MULTIPLIER"`
	BgiSetAssemblyFeeCents            int64   `mapstructure:"BGI_SET_ASSEMBLY_FEE_CENTS"`
	BgiMaxBboxMm                      float64 `mapstructure:"BGI_MAX_BBOX_MM"`
	BgiMinWallMm                      float64 `mapstructure:"BGI_MIN_WALL_MM"`
	BgiMaxPrintTimeS                  int64   `mapstructure:"BGI_MAX_PRINT_TIME_S"`
	BgiMaxWeightG                     float64 `mapstructure:"BGI_MAX_WEIGHT_G"`
	BgiMaxSupportMaterialPct          float64 `mapstructure:"BGI_MAX_SUPPORT_MATERIAL_PCT"`
	// BgiMaxSetPrintTimeS is R-6.2 rule 4's throughput/lead-time ceiling on
	// the TOTAL set (all resolved trays' print time summed), distinct from
	// BgiMaxPrintTimeS above which still gates each individual tray.
	BgiMaxSetPrintTimeS int64 `mapstructure:"BGI_MAX_SET_PRINT_TIME_S"`
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

	// AutomaticEnv() + Unmarshal only picks up a field's real env var if
	// viper already knows the key exists (via SetDefault/BindEnv/a config
	// file entry) — none of these REEF_* keys appear in live.env, so
	// without these defaults every one of them silently used its hardcoded
	// Go fallback below regardless of what ecs.tf actually set, with no
	// error or warning. (Confirmed: this is exactly why REEF_MAX_BBOX_MM=250
	// on the core container had no effect on job-runner's own bbox check —
	// core picked up the override, job-runner silently stayed at 210.)
	viper.SetDefault("REEF_OPENSCAD_BIN", "openscad")
	viper.SetDefault("REEF_SLICER_BIN", "prusa-slicer")
	// Without this, PrusaSlicer reports every part's weight as exactly
	// 0.00g regardless of actual size — confirmed against a real 2.7.4
	// binary — which zeroes out pricing.Price's material-cost component on
	// every order. 1.27 g/cm3 is a standard PETG density; override per
	// spool if the actual filament differs meaningfully.
	viper.SetDefault("REEF_FILAMENT_DENSITY_G_CM3", 1.27)
	viper.SetDefault("REEF_SUBPROCESS_TIMEOUT_SEC", 300)
	viper.SetDefault("REEF_SUBPROCESS_MEMORY_MB", 1536)
	viper.SetDefault("REEF_S3_BUCKET", "reef-site-artifacts")
	viper.SetDefault("REEF_AWS_REGION", "us-east-1")
	viper.SetDefault("REEF_PRICE_SETUP_FEE_CENTS", 300)
	viper.SetDefault("REEF_PRICE_MATERIAL_RATE_CENTS_PER_GRAM", 8.0)
	viper.SetDefault("REEF_PRICE_MACHINE_RATE_CENTS_PER_MINUTE", 4.0)
	viper.SetDefault("REEF_PRICE_FULFILLMENT_FEE_CENTS", 250)
	viper.SetDefault("REEF_PRICE_MARGIN_MULTIPLIER", 1.8)
	viper.SetDefault("REEF_MAX_BBOX_MM", 210.0)
	viper.SetDefault("REEF_MIN_WALL_MM", 2.0)
	viper.SetDefault("REEF_MAX_PRINT_TIME_S", 4*60*60)
	viper.SetDefault("REEF_MAX_WEIGHT_G", 250.0)
	viper.SetDefault("REEF_MIN_DRAIN_PATH_MM", 4.0)
	// Minor/localized support (e.g. frag_rack's small vent-channel overhang,
	// measured ~2.4% on a real print) should still pass; only a design
	// needing scaffolding through a substantial fraction of the print
	// should reject. See go/pkg/reef/validate's checkExcessiveSupport.
	viper.SetDefault("REEF_MAX_SUPPORT_MATERIAL_PCT", 10.0)

	viper.SetDefault("BGI_OPENSCAD_BIN", "openscad")
	viper.SetDefault("BGI_SLICER_BIN", "prusa-slicer")
	viper.SetDefault("BGI_FILAMENT_DENSITY_G_CM3", 1.27)
	viper.SetDefault("BGI_SUBPROCESS_TIMEOUT_SEC", 300)
	viper.SetDefault("BGI_SUBPROCESS_MEMORY_MB", 1536)
	viper.SetDefault("BGI_S3_BUCKET", "bgi-site-artifacts")
	viper.SetDefault("BGI_AWS_REGION", "us-east-1")
	viper.SetDefault("BGI_PRICE_SETUP_FEE_CENTS", 300)
	viper.SetDefault("BGI_PRICE_MATERIAL_RATE_CENTS_PER_GRAM", 8.0)
	viper.SetDefault("BGI_PRICE_MACHINE_RATE_CENTS_PER_MINUTE", 4.0)
	viper.SetDefault("BGI_PRICE_FULFILLMENT_FEE_CENTS", 250)
	viper.SetDefault("BGI_PRICE_MARGIN_MULTIPLIER", 1.8)
	viper.SetDefault("BGI_SET_ASSEMBLY_FEE_CENTS", 500)
	viper.SetDefault("BGI_MAX_BBOX_MM", 210.0)
	viper.SetDefault("BGI_MIN_WALL_MM", 2.0)
	viper.SetDefault("BGI_MAX_PRINT_TIME_S", 4*60*60)
	viper.SetDefault("BGI_MAX_WEIGHT_G", 250.0)
	viper.SetDefault("BGI_MAX_SUPPORT_MATERIAL_PCT", 10.0)
	viper.SetDefault("BGI_MAX_SET_PRINT_TIME_S", 30*60*60)

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
