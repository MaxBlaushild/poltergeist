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

	// Deliberately its own key rather than the repo-wide BASE_URL: that one
	// is shared by go/core and go/travel-angels for api.poltergeist.gg
	// (their real domain), but reef-site is served from a different domain
	// (api.unclaimedstreets.com) and reusing BASE_URL sent Stripe's payment
	// completion callback to a host that doesn't resolve, silently stranding
	// every order at "awaiting payment" forever.
	BaseURL string `mapstructure:"REEF_BASE_URL"`

	// The customer-facing storefront domain — distinct from BaseURL/REEF_BASE_URL
	// above (that one is reef-site's own API domain, used for Stripe's
	// server-to-server callback). Order confirmation emails link here.
	SiteURL string `mapstructure:"REEF_SITE_URL"`

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

	// R-7.3 SlantAdapter — reached only via an explicit per-order operator
	// action (see internal/server/operator_orders.go), never the checkout-time
	// default (FulfillmentProvider above stays "manual"). PlatformID isn't a
	// secret (just a UUID identifying this integration in Slant's system),
	// unlike the API key.
	SlantPlatformID string `mapstructure:"SLANT_PLATFORM_ID"`
}

type SecretConfig struct {
	DbPassword       string
	TwilioAccountSid string
	TwilioAuthToken  string
	// AdminToken gates /api/reef/operator/* (HTTP Basic Auth) — that surface
	// now includes real customer names/addresses via the print queue, not
	// just aggregate metrics, so it's no longer left as an unlisted-URL-only
	// endpoint.
	AdminToken  string
	SlantAPIKey string
}

type Config struct {
	Public PublicConfig
	Secret SecretConfig
}

func defaults(v *viper.Viper) {
	// AutomaticEnv() + Unmarshal only picks up a field's real env var if
	// viper already knows the key exists (via SetDefault/BindEnv/a config
	// file) — NewConfigFromEnv's path (composed into go/core) skips the
	// config-file step entirely, so every field below silently zeroed out
	// in production despite the real env var being set correctly, until
	// these defaults existed to make viper aware of the keys.
	v.SetDefault("REDIS_URL", "redis://localhost:6379")
	v.SetDefault("REEF_BASE_URL", "http://localhost:3000")
	v.SetDefault("REEF_SITE_URL", "http://localhost:5181")
	v.SetDefault("REEF_OPERATOR_EMAIL", "")
	v.SetDefault("EMAIL_FROM_ADDRESS", "")
	v.SetDefault("REEF_OPENSCAD_BIN", "openscad")
	v.SetDefault("REEF_SLICER_BIN", "prusa-slicer")
	v.SetDefault("REEF_SUBPROCESS_TIMEOUT_SEC", 60)
	v.SetDefault("REEF_SUBPROCESS_MEMORY_MB", 1024)
	// Worst case (4 tiers x 12 holes at $fn=24, R-4.2's schema maxima)
	// measured at ~65s end-to-end; 90s leaves headroom without making every
	// ordinary preview wait that long — this is a per-request timeout, not
	// a fixed delay.
	v.SetDefault("REEF_PREVIEW_TIMEOUT_SEC", 90)
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
	v.SetDefault("SLANT_PLATFORM_ID", "")
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
			DbPassword:       os.Getenv("DB_PASSWORD"),
			TwilioAccountSid: os.Getenv("TWILIO_ACCOUNT_SID"),
			TwilioAuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
			AdminToken:       os.Getenv("REEF_ADMIN_TOKEN"),
			SlantAPIKey:      os.Getenv("SLANT_API_KEY"),
		},
	}, nil
}
