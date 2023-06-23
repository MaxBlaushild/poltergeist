package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	OpenAIKey      string
	DbPassword     string
	SendgridApiKey string
}

type PublicConfig struct {
	DbHost           string `mapstructure:"DB_HOST"`
	DbUser           string `mapstructure:"DB_USER"`
	DbPort           string `mapstructure:"DB_PORT"`
	DbName           string `mapstructure:"DB_NAME"`
	EmailFromAddress string `mapstructure:"EMAIL_FROM_ADDRESS"`
	ApiHost          string `mapstructure:"API_HOST"`
	WebHost          string `mapstructure:"WEB_HOST"`
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
	flag.StringVar(&params.Name, "config-name", "local", "The name of the config file.")
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
			OpenAIKey:      os.Getenv("OPEN_AI_KEY"),
			DbPassword:     os.Getenv("ROBOT_DB_PASSWORD"),
			SendgridApiKey: os.Getenv("SENDGRID_API_KEY"),
		},
		Public: publicCfg,
	}, nil
}
