package pkg

import (
	"os"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/mapbox"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/charicturist"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/config"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/questlog"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/search"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Server interface for sonar server
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// Config holds configuration for sonar server
type Config struct {
	Public PublicConfig
	Secret SecretConfig
}

// PublicConfig holds public configuration
type PublicConfig struct {
	DbHost      string
	DbUser      string
	DbPort      string
	DbName      string
	PhoneNumber string
	RedisUrl    string
}

// SecretConfig holds secret configuration
type SecretConfig struct {
	DbPassword       string
	ImagineApiKey    string
	UseApiKey        string
	MapboxApiKey     string
	GoogleMapsApiKey string
}

// NewConfigFromEnv creates a Config from environment variables
func NewConfigFromEnv() *Config {
	return &Config{
		Public: PublicConfig{
			DbHost:      os.Getenv("DB_HOST"),
			DbUser:      os.Getenv("DB_USER"),
			DbPort:      os.Getenv("DB_PORT"),
			DbName:      os.Getenv("DB_NAME"),
			PhoneNumber: os.Getenv("PHONE_NUMBER"),
			RedisUrl:    os.Getenv("REDIS_URL"),
		},
		Secret: SecretConfig{
			DbPassword:       os.Getenv("DB_PASSWORD"),
			ImagineApiKey:    os.Getenv("IMAGINE_API_KEY"),
			UseApiKey:        os.Getenv("USE_API_KEY"),
			MapboxApiKey:     os.Getenv("MAPBOX_API_KEY"),
			GoogleMapsApiKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		},
	}
}

// toInternalConfig converts public Config to internal config
func (c *Config) toInternalConfig() *config.Config {
	return &config.Config{
		Public: config.PublicConfig{
			DbHost:      c.Public.DbHost,
			DbUser:      c.Public.DbUser,
			DbPort:      c.Public.DbPort,
			DbName:      c.Public.DbName,
			PhoneNumber: c.Public.PhoneNumber,
			RedisUrl:    c.Public.RedisUrl,
		},
		Secret: config.SecretConfig{
			DbPassword:       c.Secret.DbPassword,
			ImagineApiKey:    c.Secret.ImagineApiKey,
			UseApiKey:        c.Secret.UseApiKey,
			MapboxApiKey:     c.Secret.MapboxApiKey,
			GoogleMapsApiKey: c.Secret.GoogleMapsApiKey,
		},
	}
}

// CoreConfig is an interface for core configuration
// It allows sonar to accept any config implementation that provides these getters
type CoreConfig interface {
	GetDbHost() string
	GetDbUser() string
	GetDbPort() string
	GetDbName() string
	GetPhoneNumber() string
	GetRedisUrl() string
	GetDbPassword() string
	GetImagineApiKey() string
	GetUseApiKey() string
	GetMapboxApiKey() string
	GetGoogleMapsApiKey() string
}

// NewServerFromDependencies creates a new sonar server with minimal dependencies
// It initializes all internal dependencies internally using the provided core config
func NewServerFromDependencies(
	authClient auth.Client,
	texterClient texter.Client,
	dbClient db.DbClient,
	coreConfig CoreConfig,
) Server {
	// Convert core config to sonar config
	cfg := &Config{
		Public: PublicConfig{
			DbHost:      coreConfig.GetDbHost(),
			DbUser:      coreConfig.GetDbUser(),
			DbPort:      coreConfig.GetDbPort(),
			DbName:      coreConfig.GetDbName(),
			PhoneNumber: coreConfig.GetPhoneNumber(),
			RedisUrl:    coreConfig.GetRedisUrl(),
		},
		Secret: SecretConfig{
			DbPassword:       coreConfig.GetDbPassword(),
			ImagineApiKey:    coreConfig.GetImagineApiKey(),
			UseApiKey:        coreConfig.GetUseApiKey(),
			MapboxApiKey:     coreConfig.GetMapboxApiKey(),
			GoogleMapsApiKey: coreConfig.GetGoogleMapsApiKey(),
		},
	}

	awsClient := aws.NewAWSClient("us-east-1")
	deepPriest := deep_priest.SummonDeepPriest()
	judgeClient := judge.NewClient(awsClient, dbClient, deepPriest)
	quartermaster := quartermaster.NewClient(dbClient)
	chatClient := chat.NewClient(dbClient, quartermaster)
	useApiClient := useapi.NewClient(cfg.Secret.UseApiKey)
	charicturistClient := charicturist.NewClient(useApiClient, dbClient)
	mapboxClient := mapbox.NewClient(cfg.Secret.MapboxApiKey)
	questlogClient := questlog.NewClient(dbClient)
	googlemapsClient := googlemaps.NewClient(cfg.Secret.GoogleMapsApiKey)
	locationSeeder := locationseeder.NewClient(googlemapsClient, dbClient, deepPriest, awsClient)
	dungeonmasterClient := dungeonmaster.NewClient(googlemapsClient, dbClient, deepPriest, locationSeeder, awsClient)

	var redisClient *redis.Client
	var asyncClient *asynq.Client
	if cfg.Public.RedisUrl != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Public.RedisUrl,
			Password: "",
			DB:       0,
		})
		asyncClient = asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Public.RedisUrl})
	}

	searchClient := search.NewSearchClient(dbClient, deepPriest)
	var livenessClient liveness.LivenessClient
	if redisClient != nil {
		livenessClient = liveness.NewClient(redisClient)
	}
	gameEngineClient := gameengine.NewGameEngineClient(dbClient, judgeClient, quartermaster, chatClient, livenessClient)

	return server.NewServer(
		authClient,
		texterClient,
		dbClient,
		cfg.toInternalConfig(),
		awsClient,
		judgeClient,
		quartermaster,
		chatClient,
		charicturistClient,
		mapboxClient,
		questlogClient,
		locationSeeder,
		googlemapsClient,
		dungeonmasterClient,
		asyncClient,
		redisClient,
		searchClient,
		gameEngineClient,
		livenessClient,
	)
}

// NewServer creates a new sonar server with all dependencies provided
func NewServer(
	authClient auth.Client,
	texterClient texter.Client,
	dbClient db.DbClient,
	cfg *Config,
	awsClient aws.AWSClient,
	judgeClient judge.Client,
	quartermaster quartermaster.Quartermaster,
	chatClient chat.Client,
	charicturist charicturist.Client,
	mapboxClient mapbox.Client,
	questlogClient questlog.QuestlogClient,
	locationSeeder locationseeder.Client,
	googlemapsClient googlemaps.Client,
	dungeonmaster dungeonmaster.Client,
	asyncClient *asynq.Client,
	redisClient *redis.Client,
	searchClient search.SearchClient,
	gameEngineClient gameengine.GameEngineClient,
	livenessClient liveness.LivenessClient,
) Server {
	return server.NewServer(
		authClient,
		texterClient,
		dbClient,
		cfg.toInternalConfig(),
		awsClient,
		judgeClient,
		quartermaster,
		chatClient,
		charicturist,
		mapboxClient,
		questlogClient,
		locationSeeder,
		googlemapsClient,
		dungeonmaster,
		asyncClient,
		redisClient,
		searchClient,
		gameEngineClient,
		livenessClient,
	)
}
