package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type SecretConfig struct {
	TwilioAccountSid        string
	TwilioAuthToken         string
	GuessHowManyPhoneNumber string
}

type Config struct {
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

	return &Config{
		Secret: SecretConfig{
			TwilioAccountSid:        os.Getenv("TWILIO_ACCOUNT_SID"),
			TwilioAuthToken:         os.Getenv("TWILIO_AUTH_TOKEN"),
			GuessHowManyPhoneNumber: os.Getenv("GUESS_HOW_MANY_PHONE_NUMBER"),
		},
	}, nil
}
