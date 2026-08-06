package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

// PublicConfig mirrors go/reef-site/internal/config's shape exactly, with a
// BGI_ prefix throughout.
type PublicConfig struct {
	DbHost   string `mapstructure:"DB_HOST"`
	DbUser   string `mapstructure:"DB_USER"`
	DbPort   string `mapstructure:"DB_PORT"`
	DbName   string `mapstructure:"DB_NAME"`
	RedisUrl string `mapstructure:"REDIS_URL"`

	BaseURL string `mapstructure:"BGI_BASE_URL"`
	SiteURL string `mapstructure:"BGI_SITE_URL"`

	// R-2.5-equivalent subprocess contract, reusing go/pkg/reef/generate's
	// shared runner — same config shape as reef, own env var names so the
	// two products can be tuned independently.
	OpenSCADBin          string `mapstructure:"BGI_OPENSCAD_BIN"`
	SlicerBin            string `mapstructure:"BGI_SLICER_BIN"`
	SubprocessTimeoutSec int    `mapstructure:"BGI_SUBPROCESS_TIMEOUT_SEC"`
	SubprocessMemoryMB   int    `mapstructure:"BGI_SUBPROCESS_MEMORY_MB"`
	PreviewTimeoutSec    int    `mapstructure:"BGI_PREVIEW_TIMEOUT_SEC"`

	S3Bucket  string `mapstructure:"BGI_S3_BUCKET"`
	AwsRegion string `mapstructure:"BGI_AWS_REGION"`

	// R-7.2 pricing rates, same shape as reef's REEF_PRICE_* block, plus
	// R-7.1/R-7.2's set assembly fee — the one genuinely bgi-specific
	// pricing knob (see go/pkg/reef/set and R-7 in the requirements doc).
	SetupFeeCents              int64   `mapstructure:"BGI_PRICE_SETUP_FEE_CENTS"`
	MaterialRateCentsPerGram   float64 `mapstructure:"BGI_PRICE_MATERIAL_RATE_CENTS_PER_GRAM"`
	MachineRateCentsPerMinute  float64 `mapstructure:"BGI_PRICE_MACHINE_RATE_CENTS_PER_MINUTE"`
	FulfillmentFeeCents        int64   `mapstructure:"BGI_PRICE_FULFILLMENT_FEE_CENTS"`
	MarginMultiplier           float64 `mapstructure:"BGI_PRICE_MARGIN_MULTIPLIER"`
	SetAssemblyFeeCents        int64   `mapstructure:"BGI_SET_ASSEMBLY_FEE_CENTS"`
	// R-7.3: shipping is dimensional here, not mass-driven, and AOV clears
	// the free-shipping threshold trivially — bake shipping into price and
	// offer free shipping outright rather than reusing reef's threshold
	// mechanic. FreeShippingThresholdCents=0 means "always free" under
	// go/pkg/reef/pricing.Shipping's existing semantics.
	FreeShippingThresholdCents int64 `mapstructure:"BGI_FREE_SHIPPING_THRESHOLD_CENTS"`
	FlatShippingCents          int64 `mapstructure:"BGI_FLAT_SHIPPING_CENTS"`

	// R-5.2-equivalent rejection thresholds — config, not literals, same
	// shape as reef's REEF_MAX_*/REEF_MIN_* block. No MinDrainPathMm/
	// SealedVoidRuleEnabled equivalent here: bgi trays run with
	// SealedVoidRuleEnabled=false unconditionally (see go/pkg/reef/validate
	// and generate_bgi_set.go), so there's nothing for a drain-path
	// threshold to gate.
	MaxBboxMm     float64 `mapstructure:"BGI_MAX_BBOX_MM"`
	MinWallMm     float64 `mapstructure:"BGI_MIN_WALL_MM"`
	MaxPrintTimeS int64   `mapstructure:"BGI_MAX_PRINT_TIME_S"`
	MaxWeightG    float64 `mapstructure:"BGI_MAX_WEIGHT_G"`
	// MaxSupportMaterialPct mirrors reef's checkExcessiveSupport threshold.
	MaxSupportMaterialPct float64 `mapstructure:"BGI_MAX_SUPPORT_MATERIAL_PCT"`

	// R-6.2 rule 4: a throughput/lead-time ceiling on total SET print time
	// (not a single tray's), distinct from MaxPrintTimeS above (which still
	// gates each individual tray the same way reef gates a single part).
	MaxSetPrintTimeS int64 `mapstructure:"BGI_MAX_SET_PRINT_TIME_S"`

	FulfillmentProvider string `mapstructure:"BGI_FULFILLMENT_PROVIDER"`
	OperatorEmail        string `mapstructure:"BGI_OPERATOR_EMAIL"`
	EmailFromAddress     string `mapstructure:"EMAIL_FROM_ADDRESS"`
}

type SecretConfig struct {
	DbPassword       string
	TwilioAccountSid string
	TwilioAuthToken  string
	AdminToken       string
}

type Config struct {
	Public PublicConfig
	Secret SecretConfig
}

func defaults(v *viper.Viper) {
	v.SetDefault("REDIS_URL", "redis://localhost:6379")
	v.SetDefault("BGI_BASE_URL", "http://localhost:3000")
	v.SetDefault("BGI_SITE_URL", "http://localhost:5182")
	v.SetDefault("BGI_OPERATOR_EMAIL", "")
	v.SetDefault("EMAIL_FROM_ADDRESS", "")
	v.SetDefault("BGI_OPENSCAD_BIN", "openscad")
	v.SetDefault("BGI_SLICER_BIN", "prusa-slicer")
	v.SetDefault("BGI_SUBPROCESS_TIMEOUT_SEC", 60)
	v.SetDefault("BGI_SUBPROCESS_MEMORY_MB", 1024)
	v.SetDefault("BGI_PREVIEW_TIMEOUT_SEC", 90)
	v.SetDefault("BGI_S3_BUCKET", "bgi-site-artifacts")
	v.SetDefault("BGI_AWS_REGION", "us-east-1")
	v.SetDefault("BGI_PRICE_SETUP_FEE_CENTS", 300)
	v.SetDefault("BGI_PRICE_MATERIAL_RATE_CENTS_PER_GRAM", 8.0)
	v.SetDefault("BGI_PRICE_MACHINE_RATE_CENTS_PER_MINUTE", 4.0)
	v.SetDefault("BGI_PRICE_FULFILLMENT_FEE_CENTS", 250)
	v.SetDefault("BGI_PRICE_MARGIN_MULTIPLIER", 1.8)
	// R-7.1: sets should land at $45-90+ from real slice data — this fee is
	// a starting placeholder pending real fulfillment quotes, same "[DECIDE]"
	// posture reef's own pricing constants shipped with.
	v.SetDefault("BGI_SET_ASSEMBLY_FEE_CENTS", 500)
	// R-7.3: bake shipping into price, offer free shipping outright.
	v.SetDefault("BGI_FREE_SHIPPING_THRESHOLD_CENTS", 0)
	v.SetDefault("BGI_FLAT_SHIPPING_CENTS", 0)
	v.SetDefault("BGI_MAX_BBOX_MM", 210.0)
	v.SetDefault("BGI_MIN_WALL_MM", 2.0)
	v.SetDefault("BGI_MAX_PRINT_TIME_S", 4*60*60)
	v.SetDefault("BGI_MAX_WEIGHT_G", 250.0)
	v.SetDefault("BGI_MAX_SUPPORT_MATERIAL_PCT", 10.0)
	// R-6.2 rule 4's own example ceiling: 30 machine-hours per set.
	v.SetDefault("BGI_MAX_SET_PRINT_TIME_S", 30*60*60)
	v.SetDefault("BGI_FULFILLMENT_PROVIDER", "manual")
}

func ParseFlagsAndGetConfig() (*Config, error) {
	name := flag.String("config-name", "local", "The name of the config file.")
	fileType := flag.String("config-type", "env", "The type of the config file.")
	path := flag.String("config-path", ".", "The path of the config file.")
	flag.Parse()

	return load(*name, *fileType, *path)
}

// NewConfigFromEnv supports being composed into go/core — same reasoning as
// go/reef-site/internal/config.NewConfigFromEnv.
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
			AdminToken:       os.Getenv("BGI_ADMIN_TOKEN"),
		},
	}, nil
}
