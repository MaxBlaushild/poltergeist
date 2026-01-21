package config

import (
	"os"
)

// Config holds all configuration for core and its integrated services
type Config struct {
	// Public configuration
	Public PublicConfig
	// Secret configuration
	Secret SecretConfig
}

// PublicConfig holds public configuration values
type PublicConfig struct {
	// Database configuration
	DbHost string
	DbUser string
	DbPort string
	DbName string

	// Common configuration
	PhoneNumber string
	RedisUrl    string

	// Hue OAuth configuration
	HueRedirectURI string

	// Travel Angels configuration
	GoogleDriveRedirectURI string
	DropboxRedirectURI     string
	BaseURL                string
}

// SecretConfig holds secret configuration values
type SecretConfig struct {
	// Database configuration
	DbPassword string

	// Hue configuration
	HueClientID       string
	HueClientSecret   string
	HueApplicationKey string
	HueBridgeHostname string
	HueBridgeUsername string

	// Sonar-specific API keys
	ImagineApiKey    string
	UseApiKey        string
	MapboxApiKey     string
	GoogleMapsApiKey string

	// Travel Angels-specific configuration
	GoogleDriveClientID     string
	GoogleDriveClientSecret string
	DropboxClientID         string
	DropboxClientSecret     string
}

// NewConfigFromEnv creates a Config from environment variables
func NewConfigFromEnv() *Config {
	return &Config{
		Public: PublicConfig{
			DbHost:                 os.Getenv("DB_HOST"),
			DbUser:                 os.Getenv("DB_USER"),
			DbPort:                 os.Getenv("DB_PORT"),
			DbName:                 os.Getenv("DB_NAME"),
			PhoneNumber:            os.Getenv("PHONE_NUMBER"),
			RedisUrl:               os.Getenv("REDIS_URL"),
			HueRedirectURI:         os.Getenv("HUE_REDIRECT_URI"),
			GoogleDriveRedirectURI: os.Getenv("GOOGLE_DRIVE_REDIRECT_URI"),
			DropboxRedirectURI:     os.Getenv("DROPBOX_REDIRECT_URI"),
			BaseURL:                os.Getenv("BASE_URL"),
		},
		Secret: SecretConfig{
			DbPassword:              os.Getenv("DB_PASSWORD"),
			HueClientID:             os.Getenv("HUE_CLIENT_ID"),
			HueClientSecret:         os.Getenv("HUE_CLIENT_SECRET"),
			HueApplicationKey:       os.Getenv("HUE_APPLICATION_KEY"),
			HueBridgeHostname:       os.Getenv("HUE_BRIDGE_HOSTNAME"),
			HueBridgeUsername:       os.Getenv("HUE_BRIDGE_USERNAME"),
			ImagineApiKey:           os.Getenv("IMAGINE_API_KEY"),
			UseApiKey:               os.Getenv("USE_API_KEY"),
			MapboxApiKey:            os.Getenv("MAPBOX_API_KEY"),
			GoogleMapsApiKey:        os.Getenv("GOOGLE_MAPS_API_KEY"),
			GoogleDriveClientID:     os.Getenv("GOOGLE_DRIVE_CLIENT_ID"),
			GoogleDriveClientSecret: os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET"),
			DropboxClientID:         os.Getenv("DROPBOX_CLIENT_ID"),
			DropboxClientSecret:     os.Getenv("DROPBOX_CLIENT_SECRET"),
		},
	}
}

// GetPublic returns the public configuration
func (c *Config) GetPublic() PublicConfig {
	return c.Public
}

// GetSecret returns the secret configuration
func (c *Config) GetSecret() SecretConfig {
	return c.Secret
}

// CoreConfig interface methods - allows Config to satisfy sonar/pkg.CoreConfig
func (c *Config) GetDbHost() string                  { return c.Public.DbHost }
func (c *Config) GetDbUser() string                  { return c.Public.DbUser }
func (c *Config) GetDbPort() string                  { return c.Public.DbPort }
func (c *Config) GetDbName() string                  { return c.Public.DbName }
func (c *Config) GetPhoneNumber() string             { return c.Public.PhoneNumber }
func (c *Config) GetRedisUrl() string                { return c.Public.RedisUrl }
func (c *Config) GetDbPassword() string              { return c.Secret.DbPassword }
func (c *Config) GetImagineApiKey() string           { return c.Secret.ImagineApiKey }
func (c *Config) GetUseApiKey() string               { return c.Secret.UseApiKey }
func (c *Config) GetMapboxApiKey() string            { return c.Secret.MapboxApiKey }
func (c *Config) GetGoogleMapsApiKey() string        { return c.Secret.GoogleMapsApiKey }
func (c *Config) GetGoogleDriveClientID() string     { return c.Secret.GoogleDriveClientID }
func (c *Config) GetGoogleDriveClientSecret() string { return c.Secret.GoogleDriveClientSecret }
func (c *Config) GetGoogleDriveRedirectURI() string  { return c.Public.GoogleDriveRedirectURI }
func (c *Config) GetDropboxClientID() string         { return c.Secret.DropboxClientID }
func (c *Config) GetDropboxClientSecret() string     { return c.Secret.DropboxClientSecret }
func (c *Config) GetDropboxRedirectURI() string      { return c.Public.DropboxRedirectURI }
func (c *Config) GetBaseURL() string                 { return c.Public.BaseURL }
