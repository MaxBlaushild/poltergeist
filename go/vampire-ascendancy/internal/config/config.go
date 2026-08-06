package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	DbPassword string
}

type PublicConfig struct {
	DbHost      string `mapstructure:"DB_HOST"`
	DbUser      string `mapstructure:"DB_USER"`
	DbPort      string `mapstructure:"DB_PORT"`
	DbName      string `mapstructure:"DB_NAME"`
	PhoneNumber string `mapstructure:"PHONE_NUMBER"` // texter "From" number for player-invite SMS
	RedisUrl    string `mapstructure:"REDIS_URL"`
	// The player-facing frontend origin, used to build the RSVP link sent by
	// SMS when a player is invited (see gm_invites.go). Deliberately its own
	// key rather than a repo-wide BASE_URL — see reef-site's config for the
	// same reasoning.
	SiteURL string `mapstructure:"VAMPIRE_SITE_URL"`
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
	viper.SetDefault("VAMPIRE_SITE_URL", "http://localhost:5180")

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
			DbPassword: os.Getenv("DB_PASSWORD"),
		},
		Public: publicCfg,
	}, nil
}
