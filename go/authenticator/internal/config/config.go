package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	DbPassword     string
	AuthPrivateKey string
	// GoogleClientID is also the audience checked when verifying "Sign in
	// with Google" ID tokens — kept alongside the secret (rather than in
	// PublicConfig/the baked-in live.env like PHONE_NUMBER) so both values
	// live in the same place and neither requires an image rebuild to
	// rotate.
	GoogleClientID     string
	GoogleClientSecret string
}

type PublicConfig struct {
	DbHost        string `mapstructure:"DB_HOST"`
	DbUser        string `mapstructure:"DB_USER"`
	DbPort        string `mapstructure:"DB_PORT"`
	DbName        string `mapstructure:"DB_NAME"`
	RpID          string `mapstructure:"RP_ID"`
	RpOrigin      string `mapstructure:"RP_ORIGIN"`
	RpDisplayName string `mapstructure:"RP_DISPLAY_NAME"`
	PhoneNumber   string `mapstructure:"PHONE_NUMBER"`
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

	return &Config{
		Secret: SecretConfig{
			DbPassword:         os.Getenv("DB_PASSWORD"),
			AuthPrivateKey:     os.Getenv("AUTH_PRIVATE_KEY"),
			GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		},
		Public: publicCfg,
	}, nil
}
