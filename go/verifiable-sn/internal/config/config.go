package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	DbPassword            string
	CAPrivateKey          string
	InstagramClientSecret string
	TwitterClientSecret   string
}

type PublicConfig struct {
	DbHost                string `mapstructure:"DB_HOST"`
	DbUser                string `mapstructure:"DB_USER"`
	DbPort                string `mapstructure:"DB_PORT"`
	DbName                string `mapstructure:"DB_NAME"`
	PhoneNumber           string `mapstructure:"PHONE_NUMBER"`
	RedisUrl              string `mapstructure:"REDIS_URL"`
	EthereumTransactorURL string `mapstructure:"ETHEREUM_TRANSACTOR_URL"`
	C2PAContractAddress   string `mapstructure:"C2PA_CONTRACT_ADDRESS"`
	InstagramClientID     string `mapstructure:"INSTAGRAM_CLIENT_ID"`
	InstagramRedirectURL  string `mapstructure:"INSTAGRAM_REDIRECT_URL"`
	InstagramAuthURL      string `mapstructure:"INSTAGRAM_AUTH_URL"`
	InstagramTokenURL     string `mapstructure:"INSTAGRAM_TOKEN_URL"`
	InstagramScopes       string `mapstructure:"INSTAGRAM_SCOPES"`
	TwitterClientID       string `mapstructure:"TWITTER_CLIENT_ID"`
	TwitterRedirectURL    string `mapstructure:"TWITTER_REDIRECT_URL"`
	TwitterAuthURL        string `mapstructure:"TWITTER_AUTH_URL"`
	TwitterTokenURL       string `mapstructure:"TWITTER_TOKEN_URL"`
	TwitterScopes         string `mapstructure:"TWITTER_SCOPES"`
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
			DbPassword:            os.Getenv("DB_PASSWORD"),
			CAPrivateKey:          os.Getenv("CA_PRIVATE_KEY"), // Optional - if empty, CA will be generated
			InstagramClientSecret: os.Getenv("INSTAGRAM_CLIENT_SECRET"),
			TwitterClientSecret:   os.Getenv("TWITTER_CLIENT_SECRET"),
		},
		Public: publicCfg,
	}, nil
}
