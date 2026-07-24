package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

// PublicConfig mirrors the repo-wide viper + local.env convention (see
// go/vampire-ascendancy/internal/config, go/migrate/internal/config).
type PublicConfig struct {
	DbHost   string `mapstructure:"DB_HOST"`
	DbUser   string `mapstructure:"DB_USER"`
	DbPort   string `mapstructure:"DB_PORT"`
	DbName   string `mapstructure:"DB_NAME"`
	RedisUrl string `mapstructure:"REDIS_URL"`
	BaseURL  string `mapstructure:"BASE_URL"`

	// R-2.5 subprocess contract: binary paths + resource limits, all
	// config-driven rather than hardcoded so they can be tuned per
	// environment without a code change.
	OpenSCADBin          string `mapstructure:"REEF_OPENSCAD_BIN"`
	SlicerBin            string `mapstructure:"REEF_SLICER_BIN"`
	SubprocessTimeoutSec int    `mapstructure:"REEF_SUBPROCESS_TIMEOUT_SEC"`
	SubprocessMemoryMB   int    `mapstructure:"REEF_SUBPROCESS_MEMORY_MB"`
	PreviewTimeoutSec    int    `mapstructure:"REEF_PREVIEW_TIMEOUT_SEC"`

	// R-2.9 object storage (reusing go/pkg/aws).
	S3Bucket  string `mapstructure:"REEF_S3_BUCKET"`
	AwsRegion string `mapstructure:"REEF_AWS_REGION"`

	// R-6.1 pricing rates — [DECIDE]: seeded with placeholder values pending
	// real fulfillment quotes (see internal/reef/pricing).
	SetupFeeCents              int64   `mapstructure:"REEF_PRICE_SETUP_FEE_CENTS"`
	MaterialRateCentsPerGram   float64 `mapstructure:"REEF_PRICE_MATERIAL_RATE_CENTS_PER_GRAM"`
	MachineRateCentsPerMinute  float64 `mapstructure:"REEF_PRICE_MACHINE_RATE_CENTS_PER_MINUTE"`
	FulfillmentFeeCents        int64   `mapstructure:"REEF_PRICE_FULFILLMENT_FEE_CENTS"`
	MarginMultiplier           float64 `mapstructure:"REEF_PRICE_MARGIN_MULTIPLIER"`
	FreeShippingThresholdCents int64   `mapstructure:"REEF_FREE_SHIPPING_THRESHOLD_CENTS"`
	FlatShippingCents          int64   `mapstructure:"REEF_FLAT_SHIPPING_CENTS"`

	// R-5.2 rejection thresholds — config, not literals.
	MaxBboxMm      float64 `mapstructure:"REEF_MAX_BBOX_MM"`
	MinWallMm      float64 `mapstructure:"REEF_MIN_WALL_MM"`
	MaxPrintTimeS  int64   `mapstructure:"REEF_MAX_PRINT_TIME_S"`
	MaxWeightG     float64 `mapstructure:"REEF_MAX_WEIGHT_G"`
	MinDrainPathMm float64 `mapstructure:"REEF_MIN_DRAIN_PATH_MM"`

	// R-7.2 ManualAdapter fulfillment notification.
	FulfillmentProvider string `mapstructure:"REEF_FULFILLMENT_PROVIDER"`
	OperatorEmail       string `mapstructure:"REEF_OPERATOR_EMAIL"`
	EmailFromAddress    string `mapstructure:"EMAIL_FROM_ADDRESS"`
}

type SecretConfig struct {
	DbPassword  string
	EmailApiKey string
}

type Config struct {
	Public PublicConfig
	Secret SecretConfig
}

func defaults(v *viper.Viper) {
	v.SetDefault("REEF_OPENSCAD_BIN", "openscad")
	v.SetDefault("REEF_SLICER_BIN", "prusa-slicer")
	v.SetDefault("REEF_SUBPROCESS_TIMEOUT_SEC", 60)
	v.SetDefault("REEF_SUBPROCESS_MEMORY_MB", 1024)
	v.SetDefault("REEF_PREVIEW_TIMEOUT_SEC", 10)
	v.SetDefault("REEF_S3_BUCKET", "reef-site-artifacts")
	v.SetDefault("REEF_AWS_REGION", "us-east-1")
	v.SetDefault("REEF_PRICE_SETUP_FEE_CENTS", 300)
	v.SetDefault("REEF_PRICE_MATERIAL_RATE_CENTS_PER_GRAM", 8.0)
	v.SetDefault("REEF_PRICE_MACHINE_RATE_CENTS_PER_MINUTE", 4.0)
	v.SetDefault("REEF_PRICE_FULFILLMENT_FEE_CENTS", 250)
	v.SetDefault("REEF_PRICE_MARGIN_MULTIPLIER", 1.8)
	v.SetDefault("REEF_FREE_SHIPPING_THRESHOLD_CENTS", 4500)
	v.SetDefault("REEF_FLAT_SHIPPING_CENTS", 795)
	v.SetDefault("REEF_MAX_BBOX_MM", 210.0)
	v.SetDefault("REEF_MIN_WALL_MM", 2.0)
	v.SetDefault("REEF_MAX_PRINT_TIME_S", 4*60*60)
	v.SetDefault("REEF_MAX_WEIGHT_G", 250.0)
	v.SetDefault("REEF_MIN_DRAIN_PATH_MM", 4.0)
	v.SetDefault("REEF_FULFILLMENT_PROVIDER", "manual")
}

func ParseFlagsAndGetConfig() (*Config, error) {
	name := flag.String("config-name", "local", "The name of the config file.")
	fileType := flag.String("config-type", "env", "The type of the config file.")
	path := flag.String("config-path", ".", "The path of the config file.")
	flag.Parse()

	return load(*name, *fileType, *path)
}

// NewConfigFromEnv supports being composed into go/core, where configuration
// arrives as process environment variables rather than a local.env file.
func NewConfigFromEnv() (*Config, error) {
	return load("", "env", "")
}

func load(name, fileType, path string) (*Config, error) {
	v := viper.New()
	defaults(v)
	v.AutomaticEnv()

	if name != "" {
		v.AddConfigPath(path)
		v.SetConfigName(name)
		v.SetConfigType(fileType)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, err
			}
		}
	}

	publicCfg := PublicConfig{}
	if err := v.Unmarshal(&publicCfg); err != nil {
		return nil, err
	}

	return &Config{
		Public: publicCfg,
		Secret: SecretConfig{
			DbPassword:  os.Getenv("DB_PASSWORD"),
			EmailApiKey: os.Getenv("SENDGRID_API_KEY"),
		},
	}, nil
}
