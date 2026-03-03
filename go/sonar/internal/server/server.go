package server

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/mapbox"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/charicturist"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/config"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/questlog"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/search"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/planar"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	poiPlaceholderImageURL        = "https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp"
	poiPlaceholderThumbnailKey    = "thumbnails/placeholders/poi-undiscovered.png"
	scenarioUndiscoveredIconKey   = "thumbnails/placeholders/scenario-undiscovered.png"
	monsterUndiscoveredIconKey    = "thumbnails/placeholders/monster-undiscovered.png"
	scenarioUndiscoveredStatusKey = "admin:thumbnails:scenario-undiscovered:requested-at"
	monsterUndiscoveredStatusKey  = "admin:thumbnails:monster-undiscovered:requested-at"
	scenarioUndiscoveredIconText  = "A retro 16-bit RPG map marker icon for an undiscovered scenario. Mysterious parchment sigil, subtle compass motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	monsterUndiscoveredIconText   = "A retro 16-bit RPG map marker icon for an undiscovered monster. Hidden beast silhouette and warning rune motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	staticThumbnailJobTimeout     = 10 * time.Minute
	staticThumbnailStatusTTL      = 2 * time.Hour
	questAcceptRadiusMeters       = 50.0
	scenarioInteractRadiusMeters  = 50.0
	scenarioDefaultDifficulty     = 24
	scenarioRollSides             = 20
)

var (
	ErrNotAuthenticated = errors.New("no authenticated user found")
)

type server struct {
	authClient       auth.Client
	texterClient     texter.Client
	dbClient         db.DbClient
	config           *config.Config
	awsClient        aws.AWSClient
	judgeClient      judge.Client
	quartermaster    quartermaster.Quartermaster
	chatClient       chat.Client
	charicturist     charicturist.Client
	mapboxClient     mapbox.Client
	questlogClient   questlog.QuestlogClient
	locationSeeder   locationseeder.Client
	googlemapsClient googlemaps.Client
	dungeonmaster    dungeonmaster.Client
	asyncClient      *asynq.Client
	redisClient      *redis.Client
	searchClient     search.SearchClient
	deepPriest       deep_priest.DeepPriest
	gameEngineClient gameengine.GameEngineClient
	livenessClient   liveness.LivenessClient
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(
	authClient auth.Client,
	texterClient texter.Client,
	dbClient db.DbClient,
	config *config.Config,
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
	deepPriest deep_priest.DeepPriest,
	gameEngineClient gameengine.GameEngineClient,
	livenessClient liveness.LivenessClient,
) Server {
	return &server{
		authClient:       authClient,
		texterClient:     texterClient,
		dbClient:         dbClient,
		config:           config,
		awsClient:        awsClient,
		judgeClient:      judgeClient,
		quartermaster:    quartermaster,
		chatClient:       chatClient,
		charicturist:     charicturist,
		mapboxClient:     mapboxClient,
		questlogClient:   questlogClient,
		locationSeeder:   locationSeeder,
		googlemapsClient: googlemapsClient,
		dungeonmaster:    dungeonmaster,
		asyncClient:      asyncClient,
		redisClient:      redisClient,
		searchClient:     searchClient,
		deepPriest:       deepPriest,
		gameEngineClient: gameEngineClient,
		livenessClient:   livenessClient,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.POST("/sonar/register", s.register)
	r.POST("/sonar/login", s.login)

	r.GET("/sonar/surveys", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getSurverys))
	r.POST("/sonar/surveys", middleware.WithAuthentication(s.authClient, s.livenessClient, s.newSurvey))
	r.GET("sonar/surveys/:id/submissions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getSubmissionForSurvey))
	r.GET("/sonar/submissions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getSubmission))
	r.GET("/sonar/whoami", middleware.WithAuthentication(s.authClient, s.livenessClient, s.whoami))
	r.POST("/sonar/categories", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createCategory))
	r.POST("/sonar/activities", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createActivity))
	r.DELETE("/sonar/categories/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteCategory))
	r.DELETE("/sonar/activities/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteActivity))
	r.GET("/sonar/teams", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTeams))
	r.POST("/sonar/pointsOfInterest", s.createPointOfInterest)
	r.GET("/sonar/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointsOfInterest))
	r.POST("/sonar/pointOfInterest/unlock", middleware.WithAuthentication(s.authClient, s.livenessClient, s.unlockPointOfInterest))
	r.POST("/sonar/neighbors", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createNeighbor))
	r.GET("/sonar/neighbors", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getNeighbors))
	r.POST("/sonar/matches/:id/start", middleware.WithAuthentication(s.authClient, s.livenessClient, s.startMatch))
	r.POST("/sonar/matches/:id/end", middleware.WithAuthentication(s.authClient, s.livenessClient, s.endMatch))
	r.POST("/sonar/matches", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createMatch))
	r.GET("/sonar/matchesById/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMatch))
	r.POST("/sonar/matches/:id/leave", middleware.WithAuthentication(s.authClient, s.livenessClient, s.leaveMatch))
	r.POST("/sonar/matches/:id/teams", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTeamForMatch))
	r.POST("/sonar/teams/:teamID", middleware.WithAuthentication(s.authClient, s.livenessClient, s.addUserToTeam))
	r.GET("/sonar/pointsOfInterest/group/:id", s.getPointOfInterestGroup)
	r.POST("/sonar/pointsOfInterest/group", s.createPointOfInterestGroup)
	r.GET("/sonar/pointsOfInterest/groups", s.getPointsOfInterestGroups)
	r.GET("/sonar/matches/current", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCurrentMatch))
	r.POST("/sonar/media/uploadUrl", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPresignedUploadUrl))
	r.POST("/sonar/pointOfInterest/challenge", middleware.WithAuthentication(s.authClient, s.livenessClient, s.submitAnswerPointOfInterestChallenge))
	r.POST("/sonar/questNodes/:id/submit", middleware.WithAuthentication(s.authClient, s.livenessClient, s.submitQuestNodeChallenge))
	r.POST("/sonar/teams/:teamID/edit", middleware.WithAuthentication(s.authClient, s.livenessClient, s.editTeamName))
	r.GET("/sonar/items", s.getInventoryItems)
	r.GET("/sonar/inventory-items", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getAllInventoryItems))
	r.GET("/sonar/inventory-items/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getInventoryItem))
	r.POST("/sonar/inventory-items", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createInventoryItem))
	r.POST("/sonar/inventory-items/generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateInventoryItem))
	r.POST("/sonar/inventory-items/generate-equippable-set", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateEquippableInventorySet))
	r.POST("/sonar/inventory-items/:id/generate-consumable-qualities", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateConsumableQualities))
	r.POST("/sonar/inventory-items/:id/generate-set", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateInventoryItemSet))
	r.POST("/sonar/inventory-items/:id/regenerate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.regenerateInventoryItemImage))
	r.PUT("/sonar/inventory-items/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateInventoryItem))
	r.DELETE("/sonar/inventory-items/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteInventoryItem))
	r.POST("/sonar/inventory-items/bulk-delete", middleware.WithAuthentication(s.authClient, s.livenessClient, s.bulkDeleteInventoryItems))
	r.GET("/sonar/spells", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getSpells))
	r.POST("/sonar/spells/bulk-generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.bulkGenerateSpells))
	r.GET("/sonar/spells/bulk-generate/:jobId/status", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getBulkGenerateSpellsStatus))
	r.GET("/sonar/spells/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getSpell))
	r.POST("/sonar/spells", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createSpell))
	r.POST("/sonar/spells/:id/generate-icon", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateSpellIcon))
	r.POST("/sonar/spells/:id/cast", middleware.WithAuthentication(s.authClient, s.livenessClient, s.castSpell))
	r.PUT("/sonar/spells/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateSpell))
	r.DELETE("/sonar/spells/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteSpell))
	r.GET("/sonar/techniques", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTechniques))
	r.POST("/sonar/techniques/bulk-generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.bulkGenerateTechniques))
	r.GET("/sonar/techniques/bulk-generate/:jobId/status", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getBulkGenerateSpellsStatus))
	r.GET("/sonar/techniques/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTechnique))
	r.POST("/sonar/techniques", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTechnique))
	r.POST("/sonar/techniques/:id/generate-icon", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateTechniqueIcon))
	r.POST("/sonar/techniques/:id/cast", middleware.WithAuthentication(s.authClient, s.livenessClient, s.castTechnique))
	r.PUT("/sonar/techniques/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateTechnique))
	r.DELETE("/sonar/techniques/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteTechnique))
	r.GET("/sonar/user-spells", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCurrentUserSpells))
	r.GET("/sonar/user-techniques", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCurrentUserTechniques))
	r.GET("/sonar/teams/:teamID/inventory", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTeamsInventory))
	r.POST("/sonar/inventory/:ownedInventoryItemID/use", middleware.WithAuthentication(s.authClient, s.livenessClient, s.useItem))
	r.POST("/sonar/inventory/:ownedInventoryItemID/use-outfit", middleware.WithAuthentication(s.authClient, s.livenessClient, s.useOutfitItem))
	r.GET("/sonar/inventory/:ownedInventoryItemID/outfit-generation", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getOutfitGeneration))
	r.GET("/sonar/equipment", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getUserEquipment))
	r.POST("/sonar/equipment/equip", middleware.WithAuthentication(s.authClient, s.livenessClient, s.equipInventoryItem))
	r.POST("/sonar/equipment/unequip", middleware.WithAuthentication(s.authClient, s.livenessClient, s.unequipInventoryItem))
	r.GET("/sonar/chat", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getChat))
	r.POST("/sonar/teams/:teamID/inventory/add", s.addItemToTeam)
	r.POST("/sonar/admin/pointOfInterest/unlock", middleware.WithAuthentication(s.authClient, s.livenessClient, s.unlockPointOfInterestForTeam))
	r.POST("/sonar/pointsOfInterest/group/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createPointOfInterest))
	r.POST("/sonar/generateProfilePictureOptions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateProfilePictureOptions))
	r.GET("/sonar/generations/complete", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCompleteGenerationsForUser))
	r.POST("/sonar/profilePicture", middleware.WithAuthentication(s.authClient, s.livenessClient, s.setProfilePicture))
	r.GET("/sonar/admin/insider-trades", middleware.WithAuthentication(s.authClient, s.livenessClient, s.listInsiderTrades))
	r.GET("/sonar/admin/parties", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminListParties))
	r.POST("/sonar/admin/parties", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminCreateParty))
	r.GET("/sonar/admin/parties/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminGetParty))
	r.PATCH("/sonar/admin/parties/:id/leader", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminSetPartyLeader))
	r.POST("/sonar/admin/parties/:id/members", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminAddPartyMember))
	r.DELETE("/sonar/admin/parties/:id/members/:userId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminRemovePartyMember))
	r.DELETE("/sonar/admin/parties/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminDeleteParty))
	r.PATCH("/sonar/pointsOfInterest/group/:id", s.editPointOfInterestGroup)
	r.DELETE("/sonar/pointsOfInterest/group/:id", s.deletePointOfInterestGroup)
	r.POST("/sonar/pointsOfInterest/group/bulk-delete", s.bulkDeletePointOfInterestGroups)
	r.DELETE("/sonar/pointsOfInterest/challenge/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deletePointOfInterestChallenge))
	r.PATCH("/sonar/pointsOfInterest/challenge/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.editPointOfInterestChallenge))
	r.POST("/sonar/pointsOfInterest/challenge", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createPointOfInterestChallenge))
	r.PATCH("/sonar/pointsOfInterest/:id", s.editPointOfInterest)
	r.DELETE("/sonar/pointsOfInterest/:id", s.deletePointOfInterest)
	r.PATCH("/sonar/pointsofInterest/group/imageUrl/:id", s.editPointOfInterestGroupImageUrl)
	r.PATCH("/sonar/pointsofInterest/imageUrl/:id", s.editPointOfInterestImageUrl)
	r.POST("/sonar/pointOfInterest/children", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createPointOfInterestChildren))
	r.DELETE("/sonar/pointOfInterest/children/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deletePointOfInterestChildren))
	r.GET("/sonar/pointsOfInterest/discoveries", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointOfInterestDiscoveries))
	r.GET("/sonar/pointsOfInterest/challenges/submissions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointOfInterestChallengeSubmissions))
	r.GET("/sonar/ownedInventoryItems", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getOwnedInventoryItems))
	r.POST("/sonar/matches/:id/invite", middleware.WithAuthentication(s.authClient, s.livenessClient, s.inviteToMatch))
	r.GET("/sonar/matches/:id/users", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMatch))
	r.GET("/sonar/mapbox/places", s.getMapboxPlaces)
	r.GET("/sonar/questlog", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestLog))
	r.GET("/sonar/matches/hasCurrentMatch", middleware.WithAuthentication(s.authClient, s.livenessClient, s.hasCurrentMatch))
	r.GET("/sonar/users", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getAllUsers))
	r.POST("/sonar/users/giveItem", middleware.WithAuthentication(s.authClient, s.livenessClient, s.giveItem))
	r.GET("/sonar/admin/new-user-starter-config", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getNewUserStarterConfig))
	r.PUT("/sonar/admin/new-user-starter-config", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateNewUserStarterConfig))
	r.POST("/sonar/admin/useOutfitItem", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminUseOutfitItem))
	r.POST("/sonar/admin/users/:id/statuses", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminCreateUserStatus))
	r.POST("/sonar/admin/users/:id/resources", middleware.WithAuthentication(s.authClient, s.livenessClient, s.adminAdjustUserResources))
	r.PATCH("/sonar/users/:id/gold", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateUserGold))
	r.DELETE("/sonar/users/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteUser))
	r.DELETE("/sonar/users", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteUsers))
	r.GET("/sonar/users/:id/discoveries", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getUserDiscoveries))
	r.POST("/sonar/users/:id/discoveries", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createUserDiscoveries))
	r.DELETE("/sonar/users/:id/discoveries/:discoveryId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteUserDiscovery))
	r.DELETE("/sonar/users/:id/discoveries", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteAllUserDiscoveries))
	r.GET("/sonar/users/:id/submissions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getUserSubmissions))
	r.DELETE("/sonar/submissions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteSubmission))
	r.DELETE("/sonar/users/:id/submissions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteAllUserSubmissions))
	r.GET("/sonar/users/:id/activities", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getUserActivities))
	r.DELETE("/sonar/users/:id/activities", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteAllUserActivities))
	r.GET("/sonar/tags", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTags))
	r.GET("/sonar/proficiencies", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getProficiencies))
	r.GET("/sonar/tagGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTagGroups))
	r.POST("/sonar/tags/add", middleware.WithAuthentication(s.authClient, s.livenessClient, s.addTagToPointOfInterest))
	r.DELETE("/sonar/tags/:tagID/pointOfInterest/:pointOfInterestID", middleware.WithAuthentication(s.authClient, s.livenessClient, s.removeTagFromPointOfInterest))
	r.GET("/sonar/zones", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZones))
	r.GET("/sonar/zones/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZone))
	r.POST("/sonar/zones", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createZone))
	r.POST("/sonar/zones/import", middleware.WithAuthentication(s.authClient, s.livenessClient, s.importZonesForMetro))
	r.GET("/sonar/zones/imports", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneImports))
	r.GET("/sonar/zones/imports/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneImport))
	r.DELETE("/sonar/zones/imports/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteZoneImport))
	r.POST("/sonar/admin/zones/:id/seed-draft", middleware.WithAuthentication(s.authClient, s.livenessClient, s.seedZoneDraft))
	r.GET("/sonar/admin/zone-seed-jobs", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneSeedJobs))
	r.GET("/sonar/admin/zone-seed-jobs/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneSeedJob))
	r.POST("/sonar/admin/zone-seed-jobs/:id/approve", middleware.WithAuthentication(s.authClient, s.livenessClient, s.approveZoneSeedJob))
	r.POST("/sonar/admin/zone-seed-jobs/:id/retry", middleware.WithAuthentication(s.authClient, s.livenessClient, s.retryZoneSeedJob))
	r.POST("/sonar/admin/zone-seed-jobs/:id/shuffle-challenge", middleware.WithAuthentication(s.authClient, s.livenessClient, s.shuffleZoneSeedJobChallenge))
	r.DELETE("/sonar/admin/zone-seed-jobs/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteZoneSeedJob))
	r.POST("/sonar/admin/scenario-generation-jobs", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createScenarioGenerationJob))
	r.GET("/sonar/admin/scenario-generation-jobs", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getScenarioGenerationJobs))
	r.GET("/sonar/admin/scenario-generation-jobs/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getScenarioGenerationJob))
	r.POST("/sonar/admin/thumbnails/poi-placeholder", middleware.WithAuthentication(s.authClient, s.livenessClient, s.queuePoiPlaceholderThumbnail))
	r.POST("/sonar/admin/thumbnails/scenario-undiscovered", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateScenarioUndiscoveredIcon))
	r.POST("/sonar/admin/thumbnails/monster-undiscovered", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateMonsterUndiscoveredIcon))
	r.GET("/sonar/admin/thumbnails/scenario-undiscovered/status", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getScenarioUndiscoveredIconStatus))
	r.GET("/sonar/admin/thumbnails/monster-undiscovered/status", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMonsterUndiscoveredIconStatus))
	r.DELETE("/sonar/admin/thumbnails/scenario-undiscovered", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteScenarioUndiscoveredIcon))
	r.DELETE("/sonar/admin/thumbnails/monster-undiscovered", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteMonsterUndiscoveredIcon))
	r.GET("/sonar/zones/:id/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointsOfInterestForZone))
	r.POST("/sonar/zones/:id/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generatePointsOfInterestForZone))
	r.GET("/sonar/placeTypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPlaceTypes))
	r.DELETE("/sonar/zones/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteZone))
	r.POST("/sonar/zones/:id/pointOfInterest/:pointOfInterestId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.addPointOfInterestToZone))
	r.DELETE("/sonar/zones/:id/pointOfInterest/:pointOfInterestId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.removePointOfInterestFromZone))
	r.GET("/sonar/pointOfInterest/:id/zone", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneForPointOfInterest))
	r.POST("/sonar/pointOfInterest/import", middleware.WithAuthentication(s.authClient, s.livenessClient, s.importPointOfInterest))
	r.GET("/sonar/pointOfInterest/imports", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointOfInterestImports))
	r.GET("/sonar/pointOfInterest/imports/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointOfInterestImport))
	r.POST("/sonar/pointOfInterest/refresh", middleware.WithAuthentication(s.authClient, s.livenessClient, s.refreshPointOfInterest))
	r.POST("/sonar/pointOfInterest/image/refresh", middleware.WithAuthentication(s.authClient, s.livenessClient, s.refreshPointOfInterestImage))
	r.GET("/sonar/google/places", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getGooglePlaces))
	r.GET("/sonar/google/place/:placeID", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getGooglePlace))
	r.POST("/sonar/quests/:zoneID/:questArchTypeID/generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateQuest))
	r.GET("/sonar/quests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuests))
	r.GET("/sonar/quests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuest))
	r.POST("/sonar/quests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuest))
	r.PATCH("/sonar/quests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateQuest))
	r.DELETE("/sonar/quests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteQuest))
	r.POST("/sonar/questNodes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuestNode))
	r.POST("/sonar/questNodes/:id/challenges", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuestNodeChallenge))
	r.PATCH("/sonar/questNodes/:nodeId/challenges/:challengeId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateQuestNodeChallenge))
	r.POST("/sonar/questNodeChallenges/:challengeId/shuffle", middleware.WithAuthentication(s.authClient, s.livenessClient, s.shuffleQuestNodeChallenge))
	r.DELETE("/sonar/questNodes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteQuestNode))
	r.POST("/sonar/tags/move", middleware.WithAuthentication(s.authClient, s.livenessClient, s.moveTagToTagGroup))
	r.POST("/sonar/tags/createGroup", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTagGroup))
	r.GET("/sonar/locationArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getLocationArchetypes))
	r.GET("/sonar/locationArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getLocationArchetype))
	r.POST("/sonar/locationArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createLocationArchetype))
	r.POST("/sonar/locationArchetypes/challenges/generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateLocationArchetypeChallenges))
	r.DELETE("/sonar/locationArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteLocationArchetype))
	r.PATCH("/sonar/locationArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateLocationArchetype))
	r.GET("/sonar/questArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestArchetypes))
	r.GET("/sonar/questArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestArchetype))
	r.POST("/sonar/questArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuestArchetype))
	r.DELETE("/sonar/questArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteQuestArchetype))
	r.PATCH("/sonar/questArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateQuestArchetype))
	r.POST("/sonar/questArchetypeNodes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuestArchetypeNode))
	r.PATCH("/sonar/questArchetypeNodes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateQuestArchetypeNode))
	r.POST("/sonar/questArchetypes/:id/challenges", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateQuestArchetypeChallenge))
	r.GET("/sonar/questArchetypes/:id/challenges", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestArchetypeChallenges))
	r.PATCH("/sonar/questArchetypeChallenges/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateQuestArchetypeChallenge))
	r.DELETE("/sonar/questArchetypeChallenges/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteQuestArchetypeChallenge))
	r.POST("/sonar/zones/:id/questArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateQuestArchetypesForZone))
	r.GET("/sonar/zoneQuestArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneQuestArchetypes))
	r.POST("/sonar/zoneQuestArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createZoneQuestArchetype))
	r.PATCH("/sonar/zoneQuestArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateZoneQuestArchetype))
	r.DELETE("/sonar/zoneQuestArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteZoneQuestArchetype))
	r.POST("/sonar/zoneQuestArchetypes/:id/generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateZoneQuestArchetypeQuests))
	r.GET("/sonar/zoneQuestArchetypes/:id/questGenerations", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestGenerationJobsForZoneQuestArchetype))
	r.GET("/sonar/search/tags", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getRelevantTags))
	r.POST("/sonar/trackedPointOfInterestGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTrackedPointOfInterestGroup))
	r.GET("/sonar/trackedPointOfInterestGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTrackedPointOfInterestGroups))
	r.DELETE("/sonar/trackedPointOfInterestGroups/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteTrackedPointOfInterestGroup))
	r.DELETE("/sonar/trackedPointOfInterestGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteAllTrackedPointOfInterestGroups))
	r.POST("/sonar/trackedQuests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTrackedPointOfInterestGroup))
	r.GET("/sonar/trackedQuests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTrackedPointOfInterestGroups))
	r.DELETE("/sonar/trackedQuests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteTrackedPointOfInterestGroup))
	r.DELETE("/sonar/trackedQuests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteAllTrackedPointOfInterestGroups))
	r.POST("/sonar/quests/accept", middleware.WithAuthentication(s.authClient, s.livenessClient, s.acceptQuest))
	r.POST("/sonar/quests/turnIn/:questId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.turnInQuest))
	r.POST("/sonar/zones/:id/boundary", middleware.WithAuthentication(s.authClient, s.livenessClient, s.upsertZoneBoundary))
	r.PATCH("/sonar/zones/:id/edit", middleware.WithAuthentication(s.authClient, s.livenessClient, s.editZone))
	r.GET("/sonar/level", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getLevel))
	r.GET("/sonar/character-stats", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacterStats))
	r.GET("/sonar/users/:id/character", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getUserCharacterProfile))
	r.PUT("/sonar/character-stats/allocate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.allocateCharacterStats))
	r.GET("/sonar/zones/:id/reputation", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneReputation))
	r.GET("/sonar/reputations", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getReputations))
	r.POST("/sonar/partyInvites", middleware.WithAuthentication(s.authClient, s.livenessClient, s.inviteToParty))
	r.GET("/sonar/party", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getParty))
	r.POST("/sonar/party/leave", middleware.WithAuthentication(s.authClient, s.livenessClient, s.leaveParty))
	r.POST("/sonar/party/setLeader", middleware.WithAuthentication(s.authClient, s.livenessClient, s.setPartyLeader))
	r.POST("/sonar/partyInvites/accept", middleware.WithAuthentication(s.authClient, s.livenessClient, s.acceptPartyInvite))
	r.POST("/sonar/partyInvites/reject", middleware.WithAuthentication(s.authClient, s.livenessClient, s.rejectPartyInvite))
	r.GET("/sonar/username/validate", s.validateUsername)
	r.GET("/sonar/users/search", s.searchUsers)
	r.GET("/sonar/users/byUsername/:username", s.getUserByUsername)
	r.GET("/sonar/characters", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacters))
	r.GET("/sonar/characters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacter))
	r.GET("/sonar/characters/:id/locations", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacterLocations))
	r.POST("/sonar/characters", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createCharacter))
	r.POST("/sonar/characters/generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateCharacter))
	r.POST("/sonar/characters/:id/regenerate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.regenerateCharacterImage))
	r.PUT("/sonar/characters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateCharacter))
	r.PUT("/sonar/characters/:id/locations", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateCharacterLocations))
	r.DELETE("/sonar/characters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteCharacter))
	r.GET("/sonar/characters/:id/actions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacterActions))
	r.GET("/sonar/character-actions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacterAction))
	r.POST("/sonar/character-actions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createCharacterAction))
	r.PUT("/sonar/character-actions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateCharacterAction))
	r.DELETE("/sonar/character-actions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteCharacterAction))
	r.POST("/sonar/character-actions/:id/purchase", middleware.WithAuthentication(s.authClient, s.livenessClient, s.purchaseFromShop))
	r.POST("/sonar/character-actions/:id/sell", middleware.WithAuthentication(s.authClient, s.livenessClient, s.sellToShop))
	r.POST("/sonar/friendInvites/accept", middleware.WithAuthentication(s.authClient, s.livenessClient, s.acceptFriendInvite))
	r.POST("/sonar/friendInvites/create", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createFriendInvite))
	r.GET("/sonar/partyInvites", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPartyInvites))
	r.GET("/sonar/friendInvites", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getFriendInvites))
	r.DELETE("/sonar/friendInvites/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteFriendInvite))
	r.GET("/sonar/friends", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getFriends))
	r.POST("/sonar/profile", middleware.WithAuthentication(s.authClient, s.livenessClient, s.setProfile))
	r.GET("/sonar/activities", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getActivities))
	r.POST("/sonar/activities/markAsSeen", middleware.WithAuthentication(s.authClient, s.livenessClient, s.markActivitiesAsSeen))
	r.GET("/sonar/treasure-chests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTreasureChests))
	r.GET("/sonar/treasure-chests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTreasureChest))
	r.GET("/sonar/zones/:id/treasure-chests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTreasureChestsForZone))
	r.POST("/sonar/treasure-chests", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTreasureChest))
	r.PUT("/sonar/treasure-chests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateTreasureChest))
	r.DELETE("/sonar/treasure-chests/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteTreasureChest))
	r.POST("/sonar/treasure-chests/:id/open", middleware.WithAuthentication(s.authClient, s.livenessClient, s.openTreasureChest))
	r.GET("/sonar/monster-templates", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMonsterTemplates))
	r.GET("/sonar/monster-templates/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMonsterTemplate))
	r.POST("/sonar/monster-templates", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createMonsterTemplate))
	r.POST("/sonar/monster-templates/bulk-generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.bulkGenerateMonsterTemplates))
	r.GET("/sonar/monster-templates/bulk-generate/:jobId/status", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getBulkGenerateMonsterTemplatesStatus))
	r.PUT("/sonar/monster-templates/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateMonsterTemplate))
	r.POST("/sonar/monster-templates/:id/generate-image", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateMonsterTemplateImage))
	r.DELETE("/sonar/monster-templates/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteMonsterTemplate))
	r.GET("/sonar/monsters", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMonsters))
	r.GET("/sonar/monsters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMonster))
	r.GET("/sonar/zones/:id/monsters", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getMonstersForZone))
	r.POST("/sonar/monsters/:id/battle/start", middleware.WithAuthentication(s.authClient, s.livenessClient, s.startMonsterBattle))
	r.POST("/sonar/monsters/:id/battle/turn", middleware.WithAuthentication(s.authClient, s.livenessClient, s.advanceMonsterBattleTurn))
	r.POST("/sonar/monsters/:id/battle/end", middleware.WithAuthentication(s.authClient, s.livenessClient, s.endMonsterBattle))
	r.POST("/sonar/monsters", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createMonster))
	r.PUT("/sonar/monsters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateMonster))
	r.POST("/sonar/monsters/:id/generate-image", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateMonsterImage))
	r.DELETE("/sonar/monsters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteMonster))
	r.GET("/sonar/scenarios", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getScenarios))
	r.GET("/sonar/scenarios/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getScenario))
	r.GET("/sonar/zones/:id/scenarios", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getScenariosForZone))
	r.POST("/sonar/scenarios", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createScenario))
	r.PUT("/sonar/scenarios/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateScenario))
	r.POST("/sonar/scenarios/:id/generate-image", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateScenarioImage))
	r.DELETE("/sonar/scenarios/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteScenario))
	r.POST("/sonar/scenarios/:id/perform", middleware.WithAuthentication(s.authClient, s.livenessClient, s.performScenario))
	r.POST("/sonar/admin/treasure-chests/seed", middleware.WithAuthentication(s.authClient, s.livenessClient, s.seedTreasureChests))
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}

func (s *server) getActivities(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	activities, err := s.dbClient.Activity().GetFeed(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enriched, err := s.enrichActivities(ctx, activities)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, enriched)
}

type activityFeedItem struct {
	ID           uuid.UUID              `json:"id"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	UserID       uuid.UUID              `json:"userId"`
	ActivityType models.ActivityType    `json:"activityType"`
	Data         map[string]interface{} `json:"data"`
	Seen         bool                   `json:"seen"`
}

func (s *server) enrichActivities(ctx context.Context, activities []models.Activity) ([]activityFeedItem, error) {
	out := make([]activityFeedItem, 0, len(activities))
	for _, activity := range activities {
		data := map[string]interface{}{}
		if len(activity.Data) > 0 {
			_ = json.Unmarshal(activity.Data, &data)
		}

		entities := map[string]interface{}{}
		switch activity.ActivityType {
		case models.ActivityTypeQuestCompleted:
			var payload models.QuestCompletedActivity
			if err := json.Unmarshal(activity.Data, &payload); err == nil {
				if payload.QuestID != uuid.Nil {
					if quest, err := s.dbClient.Quest().FindByID(ctx, payload.QuestID); err == nil && quest != nil {
						entities["quest"] = map[string]interface{}{
							"id":                    quest.ID,
							"name":                  quest.Name,
							"imageUrl":              quest.ImageURL,
							"zoneId":                quest.ZoneID,
							"questGiverCharacterId": quest.QuestGiverCharacterID,
						}
					}
				}
			}
		case models.ActivityTypeChallengeCompleted:
			var payload models.ChallengeCompletedActivity
			if err := json.Unmarshal(activity.Data, &payload); err == nil {
				if payload.ChallengeID != uuid.Nil {
					if challenge, err := s.dbClient.PointOfInterestChallenge().FindByID(ctx, payload.ChallengeID); err == nil && challenge != nil {
						entities["challenge"] = map[string]interface{}{
							"id":                     challenge.ID,
							"question":               challenge.Question,
							"tier":                   challenge.Tier,
							"pointOfInterestId":      challenge.PointOfInterestID,
							"pointOfInterestGroupId": challenge.PointOfInterestGroupID,
						}
					}
				}
				if payload.QuestID != uuid.Nil || payload.QuestName != "" {
					entities["quest"] = map[string]interface{}{
						"id":   payload.QuestID,
						"name": payload.QuestName,
					}
				}
				if payload.ZoneID != uuid.Nil || payload.ZoneName != "" {
					entities["zone"] = map[string]interface{}{
						"id":   payload.ZoneID,
						"name": payload.ZoneName,
					}
				}
				entities["currentPoi"] = map[string]interface{}{
					"id":       payload.CurrentPOI.ID,
					"name":     payload.CurrentPOI.Name,
					"imageUrl": payload.CurrentPOI.ImageURL,
				}
				if payload.NextPOI != nil {
					entities["nextPoi"] = map[string]interface{}{
						"id":       payload.NextPOI.ID,
						"name":     payload.NextPOI.Name,
						"imageUrl": payload.NextPOI.ImageURL,
					}
				}
			}
		case models.ActivityTypeItemReceived:
			var payload models.ItemReceivedActivity
			if err := json.Unmarshal(activity.Data, &payload); err == nil {
				itemInfo := map[string]interface{}{
					"id":   payload.ItemID,
					"name": payload.ItemName,
				}
				if payload.ItemID != 0 {
					if item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, payload.ItemID); err == nil && item != nil {
						itemInfo["imageUrl"] = item.ImageURL
					}
				}
				entities["item"] = itemInfo
			}
		case models.ActivityTypeReputationUp:
			var payload models.ReputationUpActivity
			if err := json.Unmarshal(activity.Data, &payload); err == nil {
				entities["zone"] = map[string]interface{}{
					"id":   payload.ZoneID,
					"name": payload.ZoneName,
				}
			}
		case models.ActivityTypeLevelUp:
			var payload models.LevelUpActivity
			if err := json.Unmarshal(activity.Data, &payload); err == nil {
				entities["level"] = map[string]interface{}{
					"newLevel": payload.NewLevel,
				}
			}
		}

		if len(entities) > 0 {
			data["entities"] = entities
		}

		out = append(out, activityFeedItem{
			ID:           activity.ID,
			CreatedAt:    activity.CreatedAt,
			UpdatedAt:    activity.UpdatedAt,
			UserID:       activity.UserID,
			ActivityType: activity.ActivityType,
			Data:         data,
			Seen:         activity.Seen,
		})
	}
	return out, nil
}

func (s *server) getReputations(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	reputations, err := s.dbClient.UserZoneReputation().FindAllForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, reputations)
}

func (s *server) markActivitiesAsSeen(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		ActivityIDs []uuid.UUID `json:"activityIDs"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.dbClient.Activity().MarkAsSeen(ctx, requestBody.ActivityIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "activities marked as seen successfully"})
}

func (s *server) setPartyLeader(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		LeaderID string `json:"leaderID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderID, err := uuid.Parse(requestBody.LeaderID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid leader ID"})
		return
	}
	if leaderID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "leader ID cannot be empty"})
		return
	}

	err = s.dbClient.Party().SetLeader(ctx, *user.PartyID, leaderID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party leader set successfully"})
}

func (s *server) getParty(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	if user.PartyID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user is not in a party"})
		return
	}

	party, err := s.dbClient.Party().FindUsersParty(ctx, *user.PartyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range party.Members {
		isActive, err := s.livenessClient.HasRecentLocation(ctx, party.Members[i].ID)
		if err == nil {
			party.Members[i].IsActive = &isActive
		}
	}

	if party.Leader.ID != uuid.Nil {
		isActive, err := s.livenessClient.HasRecentLocation(ctx, party.Leader.ID)
		if err == nil {
			party.Leader.IsActive = &isActive
		}
	}

	ctx.JSON(http.StatusOK, party)
}

func (s *server) leaveParty(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	err = s.dbClient.Party().LeaveParty(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party left successfully"})
}

func (s *server) adminListParties(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	parties, err := s.dbClient.Party().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, parties)
}

func (s *server) adminGetParty(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	partyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid party ID"})
		return
	}

	party, err := s.dbClient.Party().FindByID(ctx, partyID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "party not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, party)
}

func (s *server) adminCreateParty(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		LeaderID  string   `json:"leaderId" binding:"required"`
		MemberIDs []string `json:"memberIds"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderID, err := uuid.Parse(requestBody.LeaderID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid leader ID"})
		return
	}
	if leaderID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "leader ID cannot be empty"})
		return
	}

	memberIDs := make([]uuid.UUID, 0, len(requestBody.MemberIDs))
	for _, id := range requestBody.MemberIDs {
		memberID, err := uuid.Parse(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
			return
		}
		memberIDs = append(memberIDs, memberID)
	}

	party, err := s.dbClient.Party().CreateWithMembers(ctx, leaderID, memberIDs)
	if err != nil {
		if stdErrors.Is(err, db.ErrMaxPartySizeReached) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, party)
}

func (s *server) adminSetPartyLeader(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	partyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid party ID"})
		return
	}

	var requestBody struct {
		LeaderID string `json:"leaderId" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderID, err := uuid.Parse(requestBody.LeaderID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid leader ID"})
		return
	}

	if err := s.dbClient.Party().SetLeaderAdmin(ctx, partyID, leaderID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "party or user not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party leader set successfully"})
}

func (s *server) adminAddPartyMember(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	partyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid party ID"})
		return
	}

	var requestBody struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(requestBody.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.Party().AddMember(ctx, partyID, userID); err != nil {
		if stdErrors.Is(err, db.ErrMaxPartySizeReached) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "party or user not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party member added successfully"})
}

func (s *server) adminRemovePartyMember(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	partyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid party ID"})
		return
	}

	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.Party().RemoveMember(ctx, partyID, userID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "party or user not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party member removed successfully"})
}

func (s *server) adminDeleteParty(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	partyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid party ID"})
		return
	}

	if _, err := s.dbClient.Party().FindByID(ctx, partyID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "party not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Party().Delete(ctx, partyID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party deleted successfully"})
}

func (s *server) getPartyInvites(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	invites, err := s.dbClient.PartyInvite().FindAllInvites(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, invites)
}

func (s *server) acceptPartyInvite(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		InviteID string `json:"inviteID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviteID, err := uuid.Parse(requestBody.InviteID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite ID"})
		return
	}

	invite, err := s.dbClient.PartyInvite().Accept(ctx, inviteID, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, invite)
}

func (s *server) rejectPartyInvite(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		InviteID string `json:"inviteID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviteID, err := uuid.Parse(requestBody.InviteID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite ID"})
		return
	}

	err = s.dbClient.PartyInvite().Reject(ctx, inviteID, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party invite rejected successfully"})
}

func (s *server) inviteToParty(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		InviteeID string `json:"inviteeID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviteeID, err := uuid.Parse(requestBody.InviteeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitee ID"})
		return
	}

	invite, err := s.dbClient.PartyInvite().Create(ctx, user, inviteeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, invite)
}

func (s *server) deleteFriendInvite(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	id := ctx.Param("id")
	inviteID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite ID"})
		return
	}

	invite, err := s.dbClient.FriendInvite().FindByID(ctx, inviteID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if invite.InviteeID != user.ID && invite.InviterID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you are not the invitee or inviter"})
		return
	}

	err = s.dbClient.FriendInvite().Delete(ctx, inviteID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "friend invite deleted successfully"})
}

func (s *server) setProfile(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		Username          string `json:"username"`
		ProfilePictureUrl string `json:"profilePictureUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.CreateProfilePictureTaskPayload{
		UserID:            user.ID,
		ProfilePictureUrl: requestBody.ProfilePictureUrl,
	})

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.CreateProfilePictureTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.User().Update(ctx, user.ID, models.User{
		Username:          &requestBody.Username,
		ProfilePictureUrl: models.LoadingProfilePictureUrl,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "profile set successfully"})
}

func (s *server) searchUsers(ctx *gin.Context) {
	query := ctx.Query("query")

	users, err := s.dbClient.User().FindLikeByUsername(ctx, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (s *server) getUserByUsername(ctx *gin.Context) {
	usernameQuery := ctx.Param("username")

	user, err := s.dbClient.User().FindByUsername(ctx, usernameQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	isActive, err := s.livenessClient.IsActive(ctx, user.ID)
	if err == nil {
		user.IsActive = &isActive
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *server) validateUsername(ctx *gin.Context) {
	usernameQuery := ctx.Query("username")

	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := s.dbClient.User().FindLikeByUsername(ctx, usernameQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user != nil {
		ctx.JSON(http.StatusOK, gin.H{"valid": false, "message": "Username already taken."})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"valid": true})
}

func (s *server) getFriends(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	friends, err := s.dbClient.Friend().FindAllFriends(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, friend := range friends {
		isActive, err := s.livenessClient.IsActive(ctx, friend.ID)
		if err == nil {
			friends[i].IsActive = &isActive
		}
	}

	ctx.JSON(http.StatusOK, friends)
}

func (s *server) getFriendInvites(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	invites, err := s.dbClient.FriendInvite().FindAllInvites(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, invites)
}

func (s *server) createFriendInvite(ctx *gin.Context) {
	var requestBody struct {
		InviteeID uuid.UUID `json:"inviteeID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	if _, err = s.dbClient.FriendInvite().Create(ctx, user.ID, requestBody.InviteeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "friend invite created successfully"})
}

func (s *server) acceptFriendInvite(ctx *gin.Context) {
	var requestBody struct {
		InviteID uuid.UUID `json:"inviteId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	invite, err := s.dbClient.FriendInvite().FindByID(ctx, requestBody.InviteID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if invite.InviteeID != user.ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite"})
		return
	}

	if _, err = s.dbClient.Friend().Create(ctx, invite.InviterID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = s.dbClient.FriendInvite().Delete(ctx, requestBody.InviteID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "friend invite accepted successfully"})
}

func (s *server) getPartyMembers(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	members, err := s.dbClient.User().FindPartyMembers(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, members)
}

func (s *server) joinParty(ctx *gin.Context) {
	var requestBody struct {
		InviterID uuid.UUID `json:"inviterID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	err = s.dbClient.User().JoinParty(ctx, requestBody.InviterID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "party created successfully"})
}

func (s *server) getLevel(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	level, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	level.ExperienceToNextLevel = level.XPToNextLevel()

	fmt.Println("hjksdhjksadhjkahdsjkahdjkshjdhaksj")
	fmt.Println(level.XPToNextLevel())
	fmt.Println(level.ExperienceToNextLevel)

	ctx.JSON(http.StatusOK, level)
}

func (s *server) getZoneReputation(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	zoneID := ctx.Param("id")
	zoneIDUUID, err := uuid.Parse(zoneID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	reputation, err := s.dbClient.UserZoneReputation().FindOrCreateForUserAndZone(ctx, user.ID, zoneIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, reputation)
}

func (s *server) editZone(ctx *gin.Context) {
	zoneID := ctx.Param("id")
	zoneIDUUID, err := uuid.Parse(zoneID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	var requestBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.dbClient.Zone().UpdateNameAndDescription(ctx, zoneIDUUID, requestBody.Name, requestBody.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "zone updated successfully"})
}

func (s *server) upsertZoneBoundary(ctx *gin.Context) {
	zoneID := ctx.Param("id")
	zoneIDUUID, err := uuid.Parse(zoneID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	var requestBody struct {
		Boundary [][]float64 `json:"boundary"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.dbClient.Zone().UpdateBoundary(ctx, zoneIDUUID, requestBody.Boundary)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "zone boundary updated successfully"})
}

func (s *server) userZoneReputationLevel(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID) (int, error) {
	reputation, err := s.dbClient.UserZoneReputation().FindOrCreateForUserAndZone(ctx, userID, zoneID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 1, nil
		}
		return 0, err
	}
	if reputation.Level <= 0 {
		return 1, nil
	}
	return reputation.Level, nil
}

func (s *server) userMeetsQuestReputation(ctx context.Context, userID uuid.UUID, questID uuid.UUID) (bool, int, int, error) {
	quest, err := s.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		return false, 0, 0, err
	}
	if quest == nil {
		return false, 0, 0, fmt.Errorf("quest not found")
	}
	return s.userMeetsQuestReputationForQuest(ctx, userID, quest)
}

func (s *server) userMeetsQuestReputationForQuest(ctx context.Context, userID uuid.UUID, quest *models.Quest) (bool, int, int, error) {
	if quest.ZoneID == nil {
		return true, 0, 0, nil
	}
	userLevel, err := s.userZoneReputationLevel(ctx, userID, *quest.ZoneID)
	if err != nil {
		return false, 0, 0, err
	}
	// Quests currently do not have required reputation levels; treat as unlocked.
	return true, userLevel, 0, nil
}

func extractActionQuestID(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	keys := []string{"questId", "pointOfInterestGroupId"}
	for _, key := range keys {
		if val, ok := metadata[key]; ok {
			switch v := val.(type) {
			case string:
				return v
			case uuid.UUID:
				return v.String()
			default:
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}

func (s *server) questAvailabilityByCharacter(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	actions, err := s.dbClient.CharacterAction().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	questsByCharacter := map[uuid.UUID][]uuid.UUID{}
	questIDsSet := map[uuid.UUID]struct{}{}
	for _, action := range actions {
		if action == nil || action.ActionType != models.ActionTypeGiveQuest {
			continue
		}
		questIDStr := extractActionQuestID(action.Metadata)
		if questIDStr == "" {
			continue
		}
		questID, err := uuid.Parse(questIDStr)
		if err != nil || questID == uuid.Nil {
			continue
		}
		questsByCharacter[action.CharacterID] = append(
			questsByCharacter[action.CharacterID],
			questID,
		)
		questIDsSet[questID] = struct{}{}
	}

	if len(questIDsSet) == 0 {
		return map[uuid.UUID]bool{}, nil
	}

	acceptances, err := s.dbClient.QuestAcceptanceV2().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	acceptedByQuest := map[uuid.UUID]models.QuestAcceptanceV2{}
	for _, acc := range acceptances {
		acceptedByQuest[acc.QuestID] = acc
	}

	questIDs := make([]uuid.UUID, 0, len(questIDsSet))
	for questID := range questIDsSet {
		questIDs = append(questIDs, questID)
	}

	quests, err := s.dbClient.Quest().FindByIDs(ctx, questIDs)
	if err != nil {
		return nil, err
	}
	questByID := map[uuid.UUID]*models.Quest{}
	for i := range quests {
		questByID[quests[i].ID] = &quests[i]
	}

	availableQuestIDs := map[uuid.UUID]bool{}
	for _, questID := range questIDs {
		if acc, ok := acceptedByQuest[questID]; ok {
			if acc.TurnedInAt == nil {
				continue
			}
			continue
		}
		quest := questByID[questID]
		if quest == nil {
			continue
		}
		meets, _, _, err := s.userMeetsQuestReputationForQuest(ctx, userID, quest)
		if err != nil || !meets {
			continue
		}
		availableQuestIDs[questID] = true
	}

	hasAvailable := map[uuid.UUID]bool{}
	for characterID, questIDs := range questsByCharacter {
		for _, questID := range questIDs {
			if availableQuestIDs[questID] {
				hasAvailable[characterID] = true
				break
			}
		}
	}

	return hasAvailable, nil
}

func (s *server) currentQuestNode(ctx context.Context, quest *models.Quest, acceptanceID uuid.UUID) (*models.QuestNode, error) {
	if quest == nil {
		return nil, nil
	}
	nodes, err := s.dbClient.QuestNode().FindByQuestID(ctx, quest.ID)
	if err != nil {
		return nil, err
	}
	progressEntries, err := s.dbClient.QuestNodeProgress().FindByAcceptanceID(ctx, acceptanceID)
	if err != nil {
		return nil, err
	}
	completed := map[uuid.UUID]bool{}
	for _, p := range progressEntries {
		if p.CompletedAt != nil {
			completed[p.QuestNodeID] = true
		}
	}
	for _, node := range nodes {
		if !completed[node.ID] {
			return &node, nil
		}
	}
	return nil, nil
}

func (s *server) getUserLatLng(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
	locationStr, err := s.livenessClient.GetUserLocation(ctx, userID)
	if err != nil || locationStr == "" {
		return 0, 0, fmt.Errorf("user location not available")
	}

	parts := strings.Split(locationStr, ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid location format")
	}

	userLat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude in user location")
	}

	userLng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude in user location")
	}

	return userLat, userLng, nil
}

func parseQuestNodePolygon(raw string) (orb.Polygon, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToUpper(trimmed), "SRID=") {
		if parts := strings.SplitN(trimmed, ";", 2); len(parts) == 2 {
			trimmed = parts[1]
		}
	}
	geom, err := wkt.Unmarshal(trimmed)
	if err != nil {
		return nil, err
	}
	polygon, ok := geom.(orb.Polygon)
	if !ok {
		return nil, fmt.Errorf("invalid polygon geometry")
	}
	return polygon, nil
}

func selectQuestNodeChallenge(node *models.QuestNode, challengeID *uuid.UUID) (*models.QuestNodeChallenge, error) {
	if node == nil || len(node.Challenges) == 0 {
		return nil, fmt.Errorf("quest node has no challenges")
	}
	if challengeID != nil && *challengeID != uuid.Nil {
		for _, ch := range node.Challenges {
			if ch.ID == *challengeID {
				return &ch, nil
			}
		}
		return nil, fmt.Errorf("quest node challenge not found")
	}
	if len(node.Challenges) == 1 {
		return &node.Challenges[0], nil
	}
	return nil, fmt.Errorf("questNodeChallengeId is required")
}

func (s *server) createTrackedPointOfInterestGroup(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		PointOfInterestGroupID uuid.UUID `json:"pointOfInterestGroupID"`
		QuestID                uuid.UUID `json:"questId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	questID := requestBody.QuestID
	if questID == uuid.Nil {
		questID = requestBody.PointOfInterestGroupID
	}
	if questID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "questId is required"})
		return
	}

	acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, user.ID, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if acceptance == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "quest must be accepted before tracking"})
		return
	}
	meetsReputation, _, requiredLevel, err := s.userMeetsQuestReputation(ctx, user.ID, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !meetsReputation {
		ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("requires zone reputation level %d", requiredLevel)})
		return
	}
	err = s.dbClient.TrackedQuest().Create(ctx, questID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "tracked point of interest group created successfully"})
}

func (s *server) getTrackedPointOfInterestGroups(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	trackedPointOfInterestGroups, err := s.dbClient.TrackedQuest().GetByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filtered := make([]models.TrackedQuest, 0, len(trackedPointOfInterestGroups))
	for _, tracked := range trackedPointOfInterestGroups {
		acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, user.ID, tracked.QuestID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if acceptance == nil {
			continue
		}
		meetsReputation, _, _, err := s.userMeetsQuestReputation(ctx, user.ID, tracked.QuestID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if meetsReputation {
			filtered = append(filtered, tracked)
		}
	}
	ctx.JSON(http.StatusOK, filtered)
}

func (s *server) deleteTrackedPointOfInterestGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	groupIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tracked point of interest group ID"})
		return
	}
	err = s.dbClient.TrackedQuest().Delete(ctx, groupIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "tracked point of interest group deleted successfully"})
}

func (s *server) deleteAllTrackedPointOfInterestGroups(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	err = s.dbClient.TrackedQuest().DeleteAllForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "all tracked point of interest groups deleted successfully"})
}

func (s *server) acceptQuest(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		CharacterID            uuid.UUID `json:"characterId" binding:"required"`
		PointOfInterestGroupID uuid.UUID `json:"pointOfInterestGroupId"`
		QuestID                uuid.UUID `json:"questId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify character exists
	character, err := s.dbClient.Character().FindByID(ctx, requestBody.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if character == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	// Require proximity to the nearest known quest giver location.
	type coordinate struct {
		lat float64
		lng float64
	}
	isValidCoordinate := func(lat float64, lng float64) bool {
		if math.IsNaN(lat) || math.IsNaN(lng) {
			return false
		}
		if math.IsInf(lat, 0) || math.IsInf(lng, 0) {
			return false
		}
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			return false
		}
		return lat != 0 || lng != 0
	}
	candidates := make([]coordinate, 0, len(character.Locations)+1)
	if character.PointOfInterest != nil {
		poiLat, latErr := strconv.ParseFloat(strings.TrimSpace(character.PointOfInterest.Lat), 64)
		poiLng, lngErr := strconv.ParseFloat(strings.TrimSpace(character.PointOfInterest.Lng), 64)
		if latErr == nil && lngErr == nil && isValidCoordinate(poiLat, poiLng) {
			candidates = append(candidates, coordinate{lat: poiLat, lng: poiLng})
		}
	}
	if len(candidates) == 0 {
		for _, loc := range character.Locations {
			if !isValidCoordinate(loc.Latitude, loc.Longitude) {
				continue
			}
			candidates = append(candidates, coordinate{lat: loc.Latitude, lng: loc.Longitude})
		}
	}
	startLat := character.MovementPattern.StartingLatitude
	startLng := character.MovementPattern.StartingLongitude
	if len(candidates) == 0 && isValidCoordinate(startLat, startLng) {
		candidates = append(candidates, coordinate{lat: startLat, lng: startLng})
	}
	if len(candidates) > 0 {
		locationStr, err := s.livenessClient.GetUserLocation(ctx, user.ID)
		if err != nil || strings.TrimSpace(locationStr) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user location not available"})
			return
		}
		parts := strings.Split(locationStr, ",")
		if len(parts) < 2 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid location format"})
			return
		}
		userLat, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude in user location"})
			return
		}
		userLng, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude in user location"})
			return
		}

		minDistance := math.MaxFloat64
		for _, point := range candidates {
			distance := util.HaversineDistance(userLat, userLng, point.lat, point.lng)
			if distance < minDistance {
				minDistance = distance
			}
		}

		if minDistance > questAcceptRadiusMeters {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"you must be within %.0f meters of the quest giver. Currently %.0f meters away",
					questAcceptRadiusMeters,
					minDistance,
				),
			})
			return
		}
	}

	questID := requestBody.QuestID
	if questID == uuid.Nil {
		questID = requestBody.PointOfInterestGroupID
	}
	if questID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "questId is required"})
		return
	}

	// Verify quest exists
	quest, err := s.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}

	meetsReputation, _, requiredLevel, err := s.userMeetsQuestReputationForQuest(ctx, user.ID, quest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !meetsReputation {
		ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("requires zone reputation level %d", requiredLevel)})
		return
	}

	// Verify quest has this character as quest giver
	if quest.QuestGiverCharacterID == nil || *quest.QuestGiverCharacterID != requestBody.CharacterID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "character is not the quest giver for this quest"})
		return
	}

	// Verify character has a giveQuest action for this quest
	characterActions, err := s.dbClient.CharacterAction().FindByCharacterID(ctx, requestBody.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasGiveQuestAction := false
	for _, action := range characterActions {
		if action.ActionType == models.ActionTypeGiveQuest {
			// Check if metadata contains the quest ID (can be string or UUID)
			if questIDStr := extractActionQuestID(action.Metadata); questIDStr != "" {
				if questIDStr == questID.String() {
					hasGiveQuestAction = true
					break
				}
			}
		}
	}

	if !hasGiveQuestAction {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "character does not have a giveQuest action for this quest"})
		return
	}

	// Check if user has already accepted this quest
	existingAcceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, user.ID, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingAcceptance != nil {
		// Ensure accepted quests stay tracked even if acceptance predates tracking.
		if err := s.dbClient.TrackedQuest().Create(ctx, questID, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "quest already accepted"})
		return
	}

	// Create quest acceptance record
	questAcceptance := &models.QuestAcceptanceV2{
		ID:         uuid.New(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UserID:     user.ID,
		QuestID:    questID,
		AcceptedAt: time.Now(),
	}

	if err := s.dbClient.QuestAcceptanceV2().Create(ctx, questAcceptance); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Automatically track the quest
	if err := s.dbClient.TrackedQuest().Create(ctx, questID, user.ID); err != nil {
		// Log error but don't fail the request
		log.Printf("Error tracking quest after acceptance: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "quest accepted successfully"})
}

func (s *server) turnInQuest(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil || user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	questIDStr := ctx.Param("questId")
	if questIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "questId is required"})
		return
	}
	questID, err := uuid.Parse(questIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid questId"})
		return
	}
	log.Printf("turnInQuest: userId=%s questId=%s", user.ID, questID.String())

	acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, user.ID, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if acceptance == nil {
		log.Printf("turnInQuest: no acceptance found userId=%s questId=%s", user.ID, questID.String())
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest not accepted"})
		return
	}
	if acceptance.TurnedInAt != nil {
		log.Printf("turnInQuest: already turned in userId=%s questId=%s", user.ID, questID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest already turned in"})
		return
	}

	objectivesComplete, err := s.questlogClient.AreQuestObjectivesComplete(ctx, user.ID, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !objectivesComplete {
		log.Printf("turnInQuest: objectives incomplete userId=%s questId=%s", user.ID, questID.String())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest objectives not complete"})
		return
	}

	// Award gold and items (teamID nil for single-player / user inventory)
	log.Printf("turnInQuest: awarding rewards userId=%s questId=%s", user.ID, questID.String())
	goldAwarded, itemsAwarded, spellsAwarded, err := s.gameEngineClient.AwardQuestTurnInRewards(ctx, user.ID, questID, nil)
	if err != nil {
		log.Printf("turnInQuest: reward error userId=%s questId=%s err=%v", user.ID, questID.String(), err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("turnInQuest: rewards awarded userId=%s questId=%s gold=%d items=%d spells=%d", user.ID, questID.String(), goldAwarded, len(itemsAwarded), len(spellsAwarded))

	if err := s.dbClient.QuestAcceptanceV2().MarkTurnedIn(ctx, acceptance.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{
		"goldAwarded": goldAwarded,
	}
	if len(itemsAwarded) > 0 {
		resp["itemsAwarded"] = itemsAwarded
	}
	if len(spellsAwarded) > 0 {
		resp["spellsAwarded"] = spellsAwarded
	}
	ctx.JSON(http.StatusOK, resp)
}

func (s *server) getRelevantTags(ctx *gin.Context) {
	query := ctx.Query("query")
	relevantTags, err := s.searchClient.FindRelevantTags(ctx, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, relevantTags)
}

func (s *server) getZoneQuestArchetypes(ctx *gin.Context) {
	zoneQuestArchetypes, err := s.dbClient.ZoneQuestArchetype().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zoneQuestArchetypes)
}

func (s *server) createZoneQuestArchetype(ctx *gin.Context) {
	var requestBody struct {
		ZoneID           uuid.UUID  `json:"zoneID"`
		QuestArchetypeID uuid.UUID  `json:"questArchetypeID"`
		NumberOfQuests   int        `json:"numberOfQuests"`
		CharacterID      *uuid.UUID `json:"characterId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneQuestArchetype := &models.ZoneQuestArchetype{
		ID:               uuid.New(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ZoneID:           requestBody.ZoneID,
		QuestArchetypeID: requestBody.QuestArchetypeID,
		NumberOfQuests:   requestBody.NumberOfQuests,
		CharacterID:      requestBody.CharacterID,
	}

	err := s.dbClient.ZoneQuestArchetype().Create(ctx, zoneQuestArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, zoneQuestArchetype)
}

func (s *server) updateZoneQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneQuestArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone quest archetype ID"})
		return
	}

	var payload map[string]interface{}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if value, ok := payload["characterId"]; ok {
		if value == nil {
			updates["character_id"] = nil
		} else if characterIDStr, ok := value.(string); ok {
			characterIDUUID, err := uuid.Parse(characterIDStr)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
				return
			}
			updates["character_id"] = &characterIDUUID
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
			return
		}
	}

	if value, ok := payload["numberOfQuests"]; ok {
		switch v := value.(type) {
		case float64:
			updates["number_of_quests"] = int(v)
		case int:
			updates["number_of_quests"] = v
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid numberOfQuests"})
			return
		}
	}

	if len(updates) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	if err := s.dbClient.ZoneQuestArchetype().Update(ctx, zoneQuestArchetypeIDUUID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedZoneQuestArchetype, err := s.dbClient.ZoneQuestArchetype().FindByID(ctx, zoneQuestArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if updatedZoneQuestArchetype == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone quest archetype not found"})
		return
	}

	ctx.JSON(http.StatusOK, updatedZoneQuestArchetype)
}

func (s *server) deleteZoneQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneQuestArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone quest archetype ID"})
		return
	}

	err = s.dbClient.ZoneQuestArchetype().Delete(ctx, zoneQuestArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "zone quest archetype deleted successfully"})
}

func (s *server) generateZoneQuestArchetypeQuests(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneQuestArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone quest archetype ID"})
		return
	}

	zoneQuestArchetype, err := s.dbClient.ZoneQuestArchetype().FindByID(ctx, zoneQuestArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if zoneQuestArchetype == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone quest archetype not found"})
		return
	}

	if zoneQuestArchetype.NumberOfQuests <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "number of quests must be greater than zero"})
		return
	}

	job := &models.QuestGenerationJob{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		ZoneQuestArchetypeID:  zoneQuestArchetype.ID,
		ZoneID:                zoneQuestArchetype.ZoneID,
		QuestArchetypeID:      zoneQuestArchetype.QuestArchetypeID,
		QuestGiverCharacterID: zoneQuestArchetype.CharacterID,
		Status:                models.QuestGenerationStatusQueued,
		TotalCount:            zoneQuestArchetype.NumberOfQuests,
		CompletedCount:        0,
		FailedCount:           0,
		QuestIDs:              models.StringArray{},
	}

	if err := s.dbClient.QuestGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := 0; i < zoneQuestArchetype.NumberOfQuests; i++ {
		payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
			ZoneID:                zoneQuestArchetype.ZoneID,
			QuestArchetypeID:      zoneQuestArchetype.QuestArchetypeID,
			QuestGiverCharacterID: zoneQuestArchetype.CharacterID,
			QuestGenerationJobID:  &job.ID,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateQuestForZoneTaskType, payload)); err != nil {
			msg := err.Error()
			job.Status = models.QuestGenerationStatusFailed
			job.ErrorMessage = &msg
			job.UpdatedAt = time.Now()
			_ = s.dbClient.QuestGenerationJob().Update(ctx, job)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *server) getQuestGenerationJobsForZoneQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneQuestArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone quest archetype ID"})
		return
	}

	limit := 6
	if limitParam := ctx.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	jobsList, err := s.dbClient.QuestGenerationJob().FindByZoneQuestArchetypeID(ctx, zoneQuestArchetypeIDUUID, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	questIDs := make([]uuid.UUID, 0)
	seen := map[uuid.UUID]struct{}{}
	for _, job := range jobsList {
		for _, idStr := range job.QuestIDs {
			questID, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			if _, ok := seen[questID]; ok {
				continue
			}
			seen[questID] = struct{}{}
			questIDs = append(questIDs, questID)
		}
	}

	questsByID := map[uuid.UUID]models.Quest{}
	if len(questIDs) > 0 {
		quests, err := s.dbClient.Quest().FindByIDs(ctx, questIDs)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, quest := range quests {
			questsByID[quest.ID] = quest
		}
	}

	for _, job := range jobsList {
		if len(job.QuestIDs) == 0 {
			continue
		}
		job.Quests = make([]models.Quest, 0, len(job.QuestIDs))
		for _, idStr := range job.QuestIDs {
			questID, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			if quest, ok := questsByID[questID]; ok {
				job.Quests = append(job.Quests, quest)
			}
		}
	}

	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) generateQuestArchetypesForZone(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	zoneQuestArchetypes, err := s.dbClient.ZoneQuestArchetype().FindByZoneID(ctx, zoneIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, zoneQuestArchetype := range zoneQuestArchetypes {
		payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
			ZoneID:                zoneIDUUID,
			QuestArchetypeID:      zoneQuestArchetype.QuestArchetypeID,
			QuestGiverCharacterID: zoneQuestArchetype.CharacterID,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateQuestForZoneTaskType, payload)); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "quest archetypes generated successfully"})
}

func (s *server) getQuestArchetypeChallenges(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype ID"})
		return
	}

	questArchetypeChallenges, err := s.dbClient.QuestArchetypeChallenge().FindAllByNodeID(ctx, questArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, questArchetypeChallenges)
}

func (s *server) generateQuestArchetypeChallenge(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype ID"})
		return
	}

	var requestBody struct {
		Reward              int        `json:"reward"`
		InventoryItemID     *int       `json:"inventoryItemId"`
		Proficiency         *string    `json:"proficiency"`
		Difficulty          *int       `json:"difficulty"`
		LocationArchetypeID *uuid.UUID `json:"locationArchetypeID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Proficiency != nil {
		trimmed := strings.TrimSpace(*requestBody.Proficiency)
		if trimmed == "" {
			requestBody.Proficiency = nil
		} else {
			requestBody.Proficiency = &trimmed
		}
	}

	if requestBody.Difficulty != nil && *requestBody.Difficulty < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be zero or greater"})
		return
	}

	var newNodeID *uuid.UUID
	if requestBody.LocationArchetypeID != nil {
		id := uuid.New()
		newNodeID = &id
		questArchetypeNode := &models.QuestArchetypeNode{
			ID:                  *newNodeID,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
			LocationArchetypeID: *requestBody.LocationArchetypeID,
		}

		if err := s.dbClient.QuestArchetypeNode().Create(ctx, questArchetypeNode); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	questArchetypeChallenge := &models.QuestArchetypeChallenge{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Reward:          requestBody.Reward,
		InventoryItemID: requestBody.InventoryItemID,
		Proficiency:     requestBody.Proficiency,
		Difficulty:      0,
		UnlockedNodeID:  newNodeID,
	}
	if requestBody.Difficulty != nil {
		questArchetypeChallenge.Difficulty = *requestBody.Difficulty
	}

	err = s.dbClient.QuestArchetypeChallenge().Create(ctx, questArchetypeChallenge)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.QuestArchetypeNodeChallenge().Create(ctx, &models.QuestArchetypeNodeChallenge{
		ID:                        uuid.New(),
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		QuestArchetypeChallengeID: questArchetypeChallenge.ID,
		QuestArchetypeNodeID:      questArchetypeIDUUID,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, questArchetypeChallenge)
}

func (s *server) updateQuestArchetypeChallenge(ctx *gin.Context) {
	id := ctx.Param("id")
	challengeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype challenge ID"})
		return
	}

	var requestBody struct {
		Reward          *int    `json:"reward"`
		InventoryItemID *int    `json:"inventoryItemId"`
		Proficiency     *string `json:"proficiency"`
		Difficulty      *int    `json:"difficulty"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := s.dbClient.QuestArchetypeChallenge().FindByID(ctx, challengeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Reward != nil {
		existing.Reward = *requestBody.Reward
	}
	if requestBody.InventoryItemID != nil {
		if *requestBody.InventoryItemID <= 0 {
			existing.InventoryItemID = nil
		} else {
			value := *requestBody.InventoryItemID
			existing.InventoryItemID = &value
		}
	}
	if requestBody.Proficiency != nil {
		trimmed := strings.TrimSpace(*requestBody.Proficiency)
		if trimmed == "" {
			existing.Proficiency = nil
		} else {
			existing.Proficiency = &trimmed
		}
	}
	if requestBody.Difficulty != nil {
		if *requestBody.Difficulty < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be zero or greater"})
			return
		}
		existing.Difficulty = *requestBody.Difficulty
	}
	existing.UpdatedAt = time.Now()

	if err := s.dbClient.QuestArchetypeChallenge().Update(ctx, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.QuestArchetypeChallenge().FindByID(ctx, existing.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteQuestArchetypeChallenge(ctx *gin.Context) {
	id := ctx.Param("id")
	challengeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype challenge ID"})
		return
	}

	existing, err := s.dbClient.QuestArchetypeChallenge().FindByID(ctx, challengeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype challenge not found"})
		return
	}

	if err := s.dbClient.QuestArchetypeNodeChallenge().DeleteByChallengeID(ctx, challengeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.QuestArchetypeChallenge().Delete(ctx, challengeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "quest archetype challenge deleted successfully"})
}

func (s *server) createQuestArchetypeNode(ctx *gin.Context) {
	var requestBody struct {
		LocationArchetypeID uuid.UUID `json:"locationArchetypeID"`
		Difficulty          *int      `json:"difficulty"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Difficulty != nil && *requestBody.Difficulty < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be zero or greater"})
		return
	}
	questArchetypeNode := &models.QuestArchetypeNode{
		ID:                  uuid.New(),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		LocationArchetypeID: requestBody.LocationArchetypeID,
		Difficulty:          0,
	}
	if requestBody.Difficulty != nil {
		questArchetypeNode.Difficulty = *requestBody.Difficulty
	}

	err := s.dbClient.QuestArchetypeNode().Create(ctx, questArchetypeNode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, questArchetypeNode)
}

func (s *server) updateQuestArchetypeNode(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchetypeNodeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype node ID"})
		return
	}

	var requestBody struct {
		Difficulty *int `json:"difficulty"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questArchetypeNode, err := s.dbClient.QuestArchetypeNode().FindByID(ctx, questArchetypeNodeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype node not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Difficulty != nil {
		if *requestBody.Difficulty < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be zero or greater"})
			return
		}
		questArchetypeNode.Difficulty = *requestBody.Difficulty
	}
	questArchetypeNode.UpdatedAt = time.Now()

	if err := s.dbClient.QuestArchetypeNode().Update(ctx, questArchetypeNode); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.QuestArchetypeNode().FindByID(ctx, questArchetypeNodeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype ID"})
		return
	}
	err = s.dbClient.QuestArchetype().Delete(ctx, questArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "quest archetype deleted successfully"})
}

func (s *server) updateQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype ID"})
		return
	}
	var requestBody struct {
		Name        string `json:"name"`
		DefaultGold *int   `json:"defaultGold"`
		ItemRewards *[]struct {
			InventoryItemID int `json:"inventoryItemId"`
			Quantity        int `json:"quantity"`
		} `json:"itemRewards"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questArchetype, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	questArchetype.Name = requestBody.Name
	if requestBody.DefaultGold != nil {
		questArchetype.DefaultGold = *requestBody.DefaultGold
	}

	err = s.dbClient.QuestArchetype().Update(ctx, questArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requestBody.ItemRewards != nil {
		rewards := []models.QuestArchetypeItemReward{}
		for _, reward := range *requestBody.ItemRewards {
			if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
				continue
			}
			rewards = append(rewards, models.QuestArchetypeItemReward{
				ID:               uuid.New(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				QuestArchetypeID: questArchetype.ID,
				InventoryItemID:  reward.InventoryItemID,
				Quantity:         reward.Quantity,
			})
		}
		if err := s.dbClient.QuestArchetypeItemReward().ReplaceForQuestArchetype(ctx, questArchetype.ID, rewards); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	updated, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchetype.ID)
	if err != nil || updated == nil {
		ctx.JSON(http.StatusOK, questArchetype)
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteLocationArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	locationArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid location archetype ID"})
		return
	}
	err = s.dbClient.LocationArchetype().Delete(ctx, locationArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "location archetype deleted successfully"})
}

func (s *server) updateLocationArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	locationArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid location archetype ID"})
		return
	}

	var requestBody struct {
		Name          string                              `json:"name"`
		IncludedTypes []string                            `json:"includedTypes"`
		ExcludedTypes []string                            `json:"excludedTypes"`
		Challenges    []locationArchetypeChallengePayload `json:"challenges"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, locationArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	locationArchetype.Name = requestBody.Name
	locationArchetype.IncludedTypes = googlemaps.NewPlaceTypeSlice(requestBody.IncludedTypes)
	locationArchetype.ExcludedTypes = googlemaps.NewPlaceTypeSlice(requestBody.ExcludedTypes)
	normalizedChallenges, err := normalizeLocationArchetypeChallenges(requestBody.Challenges)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	locationArchetype.Challenges = normalizedChallenges

	err = s.dbClient.LocationArchetype().Update(ctx, locationArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, locationArchetype)
}

func (s *server) createLocationArchetype(ctx *gin.Context) {
	var requestBody struct {
		Name          string                              `json:"name"`
		IncludedTypes []string                            `json:"includedTypes"`
		ExcludedTypes []string                            `json:"excludedTypes"`
		Challenges    []locationArchetypeChallengePayload `json:"challenges"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	locationArchetype := &models.LocationArchetype{
		Name:          requestBody.Name,
		ID:            uuid.New(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IncludedTypes: googlemaps.NewPlaceTypeSlice(requestBody.IncludedTypes),
		ExcludedTypes: googlemaps.NewPlaceTypeSlice(requestBody.ExcludedTypes),
	}
	normalizedChallenges, err := normalizeLocationArchetypeChallenges(requestBody.Challenges)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	locationArchetype.Challenges = normalizedChallenges

	err = s.dbClient.LocationArchetype().Create(ctx, locationArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"shit ass error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, locationArchetype)
}

const generateLocationArchetypeChallengesPromptTemplate = `
You are a game designer creating quest challenges for real-world locations in a fantasy RPG.

Generate %d distinct challenge prompts for players to complete at a location archetype.

Location archetype name: %s
Included place types: %v
Excluded place types: %v
Allowed submission types: %v

Rules:
- Each challenge must be safe, respectful, and appropriate for public locations.
- Challenges should be doable at many locations in the included types.
- Keep each question under 140 characters.
- Do not reference specific brands or businesses by name.
- Choose a submission type from the allowed list for each challenge.
- Provide a short proficiency (1-3 words) that represents the skill being tested or trained.
- Provide a difficulty rating from 0 to 100 (integer) for each challenge.
- Try to include a mix of submission types when appropriate.

Return JSON ONLY in the following format:
{
  "challenges": [
    { "question": "string", "submissionType": "text|photo|video", "proficiency": "string", "difficulty": 0 }
  ]
}
`

type locationArchetypeChallengePayload struct {
	Question       string  `json:"question"`
	SubmissionType string  `json:"submissionType"`
	Proficiency    *string `json:"proficiency"`
	Difficulty     *int    `json:"difficulty"`
}

func normalizeLocationArchetypeChallenges(challenges []locationArchetypeChallengePayload) (models.LocationArchetypeChallenges, error) {
	normalized := models.LocationArchetypeChallenges{}
	seen := map[string]struct{}{}
	for _, challenge := range challenges {
		question := strings.TrimSpace(challenge.Question)
		if question == "" {
			continue
		}
		submissionType := strings.TrimSpace(challenge.SubmissionType)
		if submissionType == "" {
			submissionType = string(models.DefaultQuestNodeSubmissionType())
		}
		parsed := models.QuestNodeSubmissionType(submissionType)
		if !parsed.IsValid() {
			return nil, fmt.Errorf("invalid submission type")
		}
		var proficiency *string
		if challenge.Proficiency != nil {
			trimmed := strings.TrimSpace(*challenge.Proficiency)
			if trimmed != "" {
				proficiency = &trimmed
			}
		}
		difficulty := 0
		if challenge.Difficulty != nil {
			if *challenge.Difficulty < 0 {
				return nil, fmt.Errorf("difficulty must be zero or greater")
			}
			difficulty = *challenge.Difficulty
		}
		proficiencyKey := ""
		if proficiency != nil {
			proficiencyKey = *proficiency
		}
		key := strings.ToLower(question) + "|" + strings.ToLower(string(parsed)) + "|" + strings.ToLower(proficiencyKey) + "|" + fmt.Sprintf("%d", difficulty)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, models.LocationArchetypeChallenge{
			Question:       question,
			SubmissionType: parsed,
			Proficiency:    proficiency,
			Difficulty:     difficulty,
		})
	}
	return normalized, nil
}

func (s *server) generateLocationArchetypeChallenges(ctx *gin.Context) {
	var requestBody struct {
		Name                   string   `json:"name"`
		IncludedTypes          []string `json:"includedTypes"`
		ExcludedTypes          []string `json:"excludedTypes"`
		AllowedSubmissionTypes []string `json:"allowedSubmissionTypes"`
		Count                  int      `json:"count"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count := requestBody.Count
	if count <= 0 || count > 20 {
		count = 10
	}

	allowedTypes := []string{}
	for _, candidate := range requestBody.AllowedSubmissionTypes {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		parsed := models.QuestNodeSubmissionType(trimmed)
		if !parsed.IsValid() {
			continue
		}
		allowedTypes = append(allowedTypes, string(parsed))
	}
	if len(allowedTypes) == 0 {
		allowedTypes = []string{
			string(models.QuestNodeSubmissionTypeText),
			string(models.QuestNodeSubmissionTypePhoto),
			string(models.QuestNodeSubmissionTypeVideo),
		}
	}

	prompt := fmt.Sprintf(
		generateLocationArchetypeChallengesPromptTemplate,
		count,
		requestBody.Name,
		requestBody.IncludedTypes,
		requestBody.ExcludedTypes,
		allowedTypes,
	)

	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response struct {
		Challenges []locationArchetypeChallengePayload `json:"challenges"`
	}

	if err := json.Unmarshal([]byte(answer.Answer), &response); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	seen := make(map[string]bool)
	cleaned := make([]locationArchetypeChallengePayload, 0, count)
	for _, challenge := range response.Challenges {
		question := strings.TrimSpace(challenge.Question)
		if question == "" {
			continue
		}
		submissionType := strings.TrimSpace(challenge.SubmissionType)
		parsed := models.QuestNodeSubmissionType(submissionType)
		if !parsed.IsValid() {
			if len(allowedTypes) > 0 {
				submissionType = allowedTypes[0]
				parsed = models.QuestNodeSubmissionType(submissionType)
			} else {
				submissionType = string(models.DefaultQuestNodeSubmissionType())
				parsed = models.QuestNodeSubmissionType(submissionType)
			}
		}
		if len(allowedTypes) > 0 {
			allowed := false
			for _, allowedType := range allowedTypes {
				if allowedType == string(parsed) {
					allowed = true
					break
				}
			}
			if !allowed {
				submissionType = allowedTypes[0]
				parsed = models.QuestNodeSubmissionType(submissionType)
			}
		}
		proficiency := ""
		if challenge.Proficiency != nil {
			proficiency = strings.TrimSpace(*challenge.Proficiency)
		}
		difficulty := 0
		if challenge.Difficulty != nil {
			difficulty = *challenge.Difficulty
		}
		if difficulty < 0 {
			difficulty = 0
		}
		key := strings.ToLower(question) + "|" + strings.ToLower(string(parsed)) + "|" + strings.ToLower(proficiency) + "|" + fmt.Sprintf("%d", difficulty)
		if seen[key] {
			continue
		}
		seen[key] = true

		var cleanedProficiency *string
		if proficiency != "" {
			cleanedProficiency = &proficiency
		}
		difficultyValue := difficulty
		cleaned = append(cleaned, locationArchetypeChallengePayload{
			Question:       question,
			SubmissionType: string(parsed),
			Proficiency:    cleanedProficiency,
			Difficulty:     &difficultyValue,
		})

		if len(cleaned) >= count {
			break
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"challenges": cleaned})
}

func (s *server) getLocationArchetypes(ctx *gin.Context) {
	locationArchetypes, err := s.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, locationArchetypes)
}

func (s *server) getLocationArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	locationArchetypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid location archetype ID"})
		return
	}
	locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, locationArchetypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, locationArchetype)
}

func (s *server) getQuestArchetypes(ctx *gin.Context) {
	questArchTypes, err := s.dbClient.QuestArchetype().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, questArchTypes)
}

func (s *server) getQuests(ctx *gin.Context) {
	quests, err := s.dbClient.Quest().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, quests)
}

func (s *server) getQuest(ctx *gin.Context) {
	id := ctx.Param("id")
	questID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest ID"})
		return
	}
	quest, err := s.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}
	ctx.JSON(http.StatusOK, quest)
}

func (s *server) createQuest(ctx *gin.Context) {
	var requestBody struct {
		Name                  string     `json:"name"`
		Description           string     `json:"description"`
		AcceptanceDialogue    []string   `json:"acceptanceDialogue"`
		ImageURL              string     `json:"imageUrl"`
		ZoneID                *uuid.UUID `json:"zoneId"`
		QuestArchetypeID      *uuid.UUID `json:"questArchetypeId"`
		QuestGiverCharacterID *uuid.UUID `json:"questGiverCharacterId"`
		RecurrenceFrequency   *string    `json:"recurrenceFrequency"`
		Gold                  *int       `json:"gold"`
		ItemRewards           *[]struct {
			InventoryItemID int `json:"inventoryItemId"`
			Quantity        int `json:"quantity"`
		} `json:"itemRewards"`
		SpellRewards *[]struct {
			SpellID string `json:"spellId"`
		} `json:"spellRewards"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(requestBody.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest name is required"})
		return
	}

	acceptanceDialogue := models.StringArray(requestBody.AcceptanceDialogue)
	if acceptanceDialogue == nil {
		acceptanceDialogue = models.StringArray{}
	}

	now := time.Now()
	quest := &models.Quest{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		Name:                  requestBody.Name,
		Description:           requestBody.Description,
		AcceptanceDialogue:    acceptanceDialogue,
		ImageURL:              requestBody.ImageURL,
		ZoneID:                requestBody.ZoneID,
		QuestArchetypeID:      requestBody.QuestArchetypeID,
		QuestGiverCharacterID: requestBody.QuestGiverCharacterID,
		Gold:                  0,
	}
	if requestBody.Gold != nil {
		quest.Gold = *requestBody.Gold
	}
	if requestBody.RecurrenceFrequency != nil {
		recurrence := models.NormalizeQuestRecurrenceFrequency(*requestBody.RecurrenceFrequency)
		if recurrence != "" {
			if !models.IsValidQuestRecurrenceFrequency(recurrence) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid recurrence frequency"})
				return
			}
			nextAt, ok := models.NextQuestRecurrenceAt(now, recurrence)
			if !ok {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid recurrence frequency"})
				return
			}
			recurringID := uuid.New()
			quest.RecurringQuestID = &recurringID
			quest.RecurrenceFrequency = &recurrence
			quest.NextRecurrenceAt = &nextAt
		}
	}

	if err := s.dbClient.Quest().Create(ctx, quest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quest.QuestGiverCharacterID != nil {
		_ = s.ensureQuestActionForCharacter(ctx, quest.ID, *quest.QuestGiverCharacterID)
	}
	if requestBody.ItemRewards != nil {
		rewards := []models.QuestItemReward{}
		for _, reward := range *requestBody.ItemRewards {
			if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
				continue
			}
			rewards = append(rewards, models.QuestItemReward{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				QuestID:         quest.ID,
				InventoryItemID: reward.InventoryItemID,
				Quantity:        reward.Quantity,
			})
		}
		if err := s.dbClient.QuestItemReward().ReplaceForQuest(ctx, quest.ID, rewards); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if requestBody.SpellRewards != nil {
		rewards := []models.QuestSpellReward{}
		for _, reward := range *requestBody.SpellRewards {
			spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
			if err != nil || spellID == uuid.Nil {
				continue
			}
			rewards = append(rewards, models.QuestSpellReward{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				QuestID:   quest.ID,
				SpellID:   spellID,
			})
		}
		if err := s.dbClient.QuestSpellReward().ReplaceForQuest(ctx, quest.ID, rewards); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	createdQuest, err := s.dbClient.Quest().FindByID(ctx, quest.ID)
	if err != nil || createdQuest == nil {
		ctx.JSON(http.StatusOK, quest)
		return
	}
	ctx.JSON(http.StatusOK, createdQuest)
}

func (s *server) updateQuest(ctx *gin.Context) {
	id := ctx.Param("id")
	questID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest ID"})
		return
	}

	var requestBody struct {
		Name                  string     `json:"name"`
		Description           string     `json:"description"`
		AcceptanceDialogue    *[]string  `json:"acceptanceDialogue"`
		ImageURL              string     `json:"imageUrl"`
		ZoneID                *uuid.UUID `json:"zoneId"`
		QuestArchetypeID      *uuid.UUID `json:"questArchetypeId"`
		QuestGiverCharacterID *uuid.UUID `json:"questGiverCharacterId"`
		RecurrenceFrequency   *string    `json:"recurrenceFrequency"`
		Gold                  *int       `json:"gold"`
		ItemRewards           *[]struct {
			InventoryItemID int `json:"inventoryItemId"`
			Quantity        int `json:"quantity"`
		} `json:"itemRewards"`
		SpellRewards *[]struct {
			SpellID string `json:"spellId"`
		} `json:"spellRewards"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quest, err := s.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}

	previousQuestGiver := quest.QuestGiverCharacterID
	quest.Name = requestBody.Name
	quest.Description = requestBody.Description
	if requestBody.AcceptanceDialogue != nil {
		quest.AcceptanceDialogue = models.StringArray(*requestBody.AcceptanceDialogue)
	}
	quest.ImageURL = requestBody.ImageURL
	quest.ZoneID = requestBody.ZoneID
	quest.QuestArchetypeID = requestBody.QuestArchetypeID
	quest.QuestGiverCharacterID = requestBody.QuestGiverCharacterID
	if requestBody.Gold != nil {
		quest.Gold = *requestBody.Gold
	}
	if requestBody.RecurrenceFrequency != nil {
		recurrence := models.NormalizeQuestRecurrenceFrequency(*requestBody.RecurrenceFrequency)
		if recurrence == "" {
			quest.RecurrenceFrequency = nil
			quest.NextRecurrenceAt = nil
		} else {
			if !models.IsValidQuestRecurrenceFrequency(recurrence) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid recurrence frequency"})
				return
			}
			if quest.RecurringQuestID == nil {
				recurringID := uuid.New()
				quest.RecurringQuestID = &recurringID
			}
			nextAt, ok := models.NextQuestRecurrenceAt(time.Now(), recurrence)
			if !ok {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid recurrence frequency"})
				return
			}
			quest.RecurrenceFrequency = &recurrence
			quest.NextRecurrenceAt = &nextAt
		}
	}
	quest.UpdatedAt = time.Now()

	if err := s.dbClient.Quest().Update(ctx, questID, quest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if previousQuestGiver != nil && (quest.QuestGiverCharacterID == nil || *previousQuestGiver != *quest.QuestGiverCharacterID) {
		_ = s.removeQuestActionForCharacter(ctx, quest.ID, *previousQuestGiver)
	}
	if quest.QuestGiverCharacterID != nil {
		_ = s.ensureQuestActionForCharacter(ctx, quest.ID, *quest.QuestGiverCharacterID)
	}
	if requestBody.ItemRewards != nil {
		rewards := []models.QuestItemReward{}
		for _, reward := range *requestBody.ItemRewards {
			if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
				continue
			}
			rewards = append(rewards, models.QuestItemReward{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				QuestID:         quest.ID,
				InventoryItemID: reward.InventoryItemID,
				Quantity:        reward.Quantity,
			})
		}
		if err := s.dbClient.QuestItemReward().ReplaceForQuest(ctx, quest.ID, rewards); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if requestBody.SpellRewards != nil {
		rewards := []models.QuestSpellReward{}
		for _, reward := range *requestBody.SpellRewards {
			spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
			if err != nil || spellID == uuid.Nil {
				continue
			}
			rewards = append(rewards, models.QuestSpellReward{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				QuestID:   quest.ID,
				SpellID:   spellID,
			})
		}
		if err := s.dbClient.QuestSpellReward().ReplaceForQuest(ctx, quest.ID, rewards); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	updatedQuest, err := s.dbClient.Quest().FindByID(ctx, quest.ID)
	if err != nil || updatedQuest == nil {
		ctx.JSON(http.StatusOK, quest)
		return
	}
	ctx.JSON(http.StatusOK, updatedQuest)
}

func (s *server) deleteQuest(ctx *gin.Context) {
	id := ctx.Param("id")
	questID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest ID"})
		return
	}

	if err := s.dbClient.Quest().Delete(ctx, questID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.CleanupOrphanedQuestActionsTaskType, nil)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "quest deleted successfully"})
}

func (s *server) ensureQuestActionForCharacter(ctx *gin.Context, questID uuid.UUID, characterID uuid.UUID) error {
	actions, err := s.dbClient.CharacterAction().FindByCharacterID(ctx, characterID)
	if err != nil {
		return err
	}
	questIDStr := questID.String()
	for _, action := range actions {
		if action.ActionType != models.ActionTypeGiveQuest {
			continue
		}
		if action.Metadata == nil {
			continue
		}
		if value, ok := action.Metadata["questId"]; ok && fmt.Sprint(value) == questIDStr {
			return nil
		}
	}

	action := &models.CharacterAction{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CharacterID: characterID,
		ActionType:  models.ActionTypeGiveQuest,
		Dialogue:    []models.DialogueMessage{},
		Metadata:    map[string]interface{}{"questId": questIDStr},
	}
	return s.dbClient.CharacterAction().Create(ctx, action)
}

func (s *server) removeQuestActionForCharacter(ctx *gin.Context, questID uuid.UUID, characterID uuid.UUID) error {
	actions, err := s.dbClient.CharacterAction().FindByCharacterID(ctx, characterID)
	if err != nil {
		return err
	}
	questIDStr := questID.String()
	for _, action := range actions {
		if action.ActionType != models.ActionTypeGiveQuest {
			continue
		}
		if action.Metadata == nil {
			continue
		}
		if value, ok := action.Metadata["questId"]; ok && fmt.Sprint(value) == questIDStr {
			_ = s.dbClient.CharacterAction().Delete(ctx, action.ID)
		}
	}
	return nil
}

func (s *server) createQuestNode(ctx *gin.Context) {
	var requestBody struct {
		QuestID           uuid.UUID    `json:"questId"`
		OrderIndex        int          `json:"orderIndex"`
		PointOfInterestID *uuid.UUID   `json:"pointOfInterestId"`
		ScenarioID        *uuid.UUID   `json:"scenarioId"`
		MonsterID         *uuid.UUID   `json:"monsterId"`
		Polygon           string       `json:"polygon"`
		PolygonPoints     [][2]float64 `json:"polygonPoints"`
		SubmissionType    string       `json:"submissionType"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.QuestID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "questId is required"})
		return
	}

	hasPolygon := strings.TrimSpace(requestBody.Polygon) != "" || len(requestBody.PolygonPoints) > 0
	targetCount := 0
	if requestBody.PointOfInterestID != nil {
		targetCount++
	}
	if requestBody.ScenarioID != nil {
		targetCount++
	}
	if requestBody.MonsterID != nil {
		targetCount++
	}
	if hasPolygon {
		targetCount++
	}
	if targetCount == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest node must have a pointOfInterestId, scenarioId, monsterId, or polygon"})
		return
	}
	if targetCount > 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest node must have exactly one target: pointOfInterestId, scenarioId, monsterId, or polygon"})
		return
	}

	node := &models.QuestNode{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		QuestID:           requestBody.QuestID,
		OrderIndex:        requestBody.OrderIndex,
		PointOfInterestID: requestBody.PointOfInterestID,
		ScenarioID:        requestBody.ScenarioID,
		MonsterID:         requestBody.MonsterID,
	}
	if strings.TrimSpace(requestBody.SubmissionType) == "" {
		node.SubmissionType = models.DefaultQuestNodeSubmissionType()
	} else {
		node.SubmissionType = models.QuestNodeSubmissionType(strings.TrimSpace(requestBody.SubmissionType))
		if !node.SubmissionType.IsValid() {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submissionType"})
			return
		}
	}
	if len(requestBody.PolygonPoints) > 0 {
		node.SetPolygonFromPoints(requestBody.PolygonPoints)
	} else {
		node.Polygon = requestBody.Polygon
	}

	if err := s.dbClient.QuestNode().Create(ctx, node); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, node)
}

func (s *server) createQuestNodeChallenge(ctx *gin.Context) {
	id := ctx.Param("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest node ID"})
		return
	}

	var requestBody struct {
		Tier            int      `json:"tier"`
		Question        string   `json:"question"`
		Reward          int      `json:"reward"`
		InventoryItemID *int     `json:"inventoryItemId"`
		StatTags        []string `json:"statTags"`
		Difficulty      int      `json:"difficulty"`
		Proficiency     string   `json:"proficiency"`
		SubmissionType  string   `json:"submissionType"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(requestBody.Question) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	if requestBody.Difficulty < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be zero or greater"})
		return
	}

	statTags := models.StringArray(requestBody.StatTags)
	if statTags == nil {
		statTags = models.StringArray{}
	}
	proficiency := strings.TrimSpace(requestBody.Proficiency)
	var proficiencyPtr *string
	if proficiency != "" {
		proficiencyPtr = &proficiency
	}

	submissionType := strings.TrimSpace(requestBody.SubmissionType)
	if submissionType == "" {
		node, err := s.dbClient.QuestNode().FindByID(ctx, nodeID)
		if err == nil && node != nil {
			submissionType = strings.TrimSpace(string(node.SubmissionType))
		}
	}
	if submissionType == "" {
		submissionType = string(models.DefaultQuestNodeSubmissionType())
	}
	parsedSubmissionType := models.QuestNodeSubmissionType(submissionType)
	if !parsedSubmissionType.IsValid() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submissionType"})
		return
	}

	challenge := &models.QuestNodeChallenge{
		ID:                     uuid.New(),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		QuestNodeID:            nodeID,
		Tier:                   requestBody.Tier,
		Question:               requestBody.Question,
		Reward:                 requestBody.Reward,
		InventoryItemID:        requestBody.InventoryItemID,
		SubmissionType:         parsedSubmissionType,
		Difficulty:             requestBody.Difficulty,
		StatTags:               statTags,
		Proficiency:            proficiencyPtr,
		ChallengeShuffleStatus: models.QuestNodeChallengeShuffleStatusIdle,
		ChallengeShuffleError:  nil,
	}

	if err := s.dbClient.QuestNodeChallenge().Create(ctx, challenge); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, challenge)
}

func (s *server) updateQuestNodeChallenge(ctx *gin.Context) {
	nodeIDParam := ctx.Param("nodeId")
	challengeIDParam := ctx.Param("challengeId")

	nodeID, err := uuid.Parse(nodeIDParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest node ID"})
		return
	}

	challengeID, err := uuid.Parse(challengeIDParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest node challenge ID"})
		return
	}

	existing, err := s.dbClient.QuestNodeChallenge().FindByID(ctx, challengeID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest node challenge not found"})
		return
	}
	if existing.QuestNodeID != nodeID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "challenge does not belong to quest node"})
		return
	}

	var requestBody struct {
		Tier            int      `json:"tier"`
		Question        string   `json:"question"`
		Reward          int      `json:"reward"`
		InventoryItemID *int     `json:"inventoryItemId"`
		StatTags        []string `json:"statTags"`
		Difficulty      int      `json:"difficulty"`
		Proficiency     string   `json:"proficiency"`
		SubmissionType  *string  `json:"submissionType"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(requestBody.Question) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	if requestBody.Difficulty < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be zero or greater"})
		return
	}

	statTags := models.StringArray(requestBody.StatTags)
	if statTags == nil {
		statTags = models.StringArray{}
	}
	proficiency := strings.TrimSpace(requestBody.Proficiency)
	var proficiencyPtr *string
	if proficiency != "" {
		proficiencyPtr = &proficiency
	}

	updates := &models.QuestNodeChallenge{
		Tier:                   requestBody.Tier,
		Question:               requestBody.Question,
		Reward:                 requestBody.Reward,
		InventoryItemID:        requestBody.InventoryItemID,
		Difficulty:             requestBody.Difficulty,
		StatTags:               statTags,
		Proficiency:            proficiencyPtr,
		SubmissionType:         existing.SubmissionType,
		ChallengeShuffleStatus: existing.ChallengeShuffleStatus,
		ChallengeShuffleError:  existing.ChallengeShuffleError,
		UpdatedAt:              time.Now(),
	}

	if requestBody.SubmissionType != nil {
		trimmed := strings.TrimSpace(*requestBody.SubmissionType)
		if trimmed == "" {
			trimmed = string(models.DefaultQuestNodeSubmissionType())
		}
		parsed := models.QuestNodeSubmissionType(trimmed)
		if !parsed.IsValid() {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submissionType"})
			return
		}
		updates.SubmissionType = parsed
	}

	updated, err := s.dbClient.QuestNodeChallenge().Update(ctx, challengeID, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (s *server) shuffleQuestNodeChallenge(ctx *gin.Context) {
	challengeIDParam := ctx.Param("challengeId")

	challengeID, err := uuid.Parse(challengeIDParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest node challenge ID"})
		return
	}

	existing, err := s.dbClient.QuestNodeChallenge().FindByID(ctx, challengeID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest node challenge not found"})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest node challenge not found"})
		return
	}
	if existing.ChallengeShuffleStatus == models.QuestNodeChallengeShuffleStatusQueued ||
		existing.ChallengeShuffleStatus == models.QuestNodeChallengeShuffleStatusInProgress {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "challenge shuffle already in progress"})
		return
	}

	queued := *existing
	queued.ChallengeShuffleStatus = models.QuestNodeChallengeShuffleStatusQueued
	queued.ChallengeShuffleError = nil
	queued.UpdatedAt = time.Now()

	updated, err := s.dbClient.QuestNodeChallenge().Update(ctx, challengeID, &queued)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.ShuffleQuestNodeChallengeTaskPayload{
		QuestNodeChallengeID: challengeID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ShuffleQuestNodeChallengeTaskType, payloadBytes)); err != nil {
		failed := *updated
		errMsg := err.Error()
		failed.ChallengeShuffleStatus = models.QuestNodeChallengeShuffleStatusFailed
		failed.ChallengeShuffleError = &errMsg
		failed.UpdatedAt = time.Now()
		_, _ = s.dbClient.QuestNodeChallenge().Update(ctx, challengeID, &failed)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, updated)
}

func (s *server) deleteQuestNode(ctx *gin.Context) {
	id := ctx.Param("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest node ID"})
		return
	}

	if err := s.dbClient.QuestNodeChild().DeleteByNodeID(ctx, nodeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.QuestNodeChild().DeleteByNextNodeID(ctx, nodeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.QuestNodeChallenge().DeleteByNodeID(ctx, nodeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.QuestNodeProgress().DeleteByNodeID(ctx, nodeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.QuestNode().DeleteByID(ctx, nodeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "quest node deleted"})
}

func (s *server) getQuestArchetype(ctx *gin.Context) {
	id := ctx.Param("id")
	questArchTypeIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype ID"})
		return
	}
	questArchType, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchTypeIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, questArchType)
}

func (s *server) createQuestArchetype(ctx *gin.Context) {
	var requestBody struct {
		Name        string    `json:"name"`
		RootID      uuid.UUID `json:"rootID"`
		DefaultGold *int      `json:"defaultGold"`
		ItemRewards *[]struct {
			InventoryItemID int `json:"inventoryItemId"`
			Quantity        int `json:"quantity"`
		} `json:"itemRewards"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	questArchType := &models.QuestArchetype{
		Name:      requestBody.Name,
		RootID:    requestBody.RootID,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if requestBody.DefaultGold != nil {
		questArchType.DefaultGold = *requestBody.DefaultGold
	}

	err := s.dbClient.QuestArchetype().Create(ctx, questArchType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requestBody.ItemRewards != nil {
		rewards := []models.QuestArchetypeItemReward{}
		for _, reward := range *requestBody.ItemRewards {
			if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
				continue
			}
			rewards = append(rewards, models.QuestArchetypeItemReward{
				ID:               uuid.New(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				QuestArchetypeID: questArchType.ID,
				InventoryItemID:  reward.InventoryItemID,
				Quantity:         reward.Quantity,
			})
		}
		if err := s.dbClient.QuestArchetypeItemReward().ReplaceForQuestArchetype(ctx, questArchType.ID, rewards); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	created, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchType.ID)
	if err != nil || created == nil {
		ctx.JSON(http.StatusOK, questArchType)
		return
	}
	ctx.JSON(http.StatusOK, created)
}

func (s *server) moveTagToTagGroup(ctx *gin.Context) {
	var requestBody struct {
		TagID      uuid.UUID `json:"tagID"`
		TagGroupID uuid.UUID `json:"tagGroupID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.dbClient.Tag().MoveTagToTagGroup(ctx, requestBody.TagID, requestBody.TagGroupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "tag moved to tag group successfully"})
}

func (s *server) createTagGroup(ctx *gin.Context) {
	var requestBody struct {
		Name     string `json:"name"`
		IconUrl  string `json:"iconUrl"`
		ImageUrl string `json:"imageUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tagGroup := &models.TagGroup{
		Name:      requestBody.Name,
		ID:        uuid.New(),
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
		IconUrl:   requestBody.IconUrl,
		ImageUrl:  requestBody.ImageUrl,
	}

	err := s.dbClient.Tag().CreateTagGroup(ctx, tagGroup)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tagGroup)
}

func (s *server) generateQuest(ctx *gin.Context) {
	id := ctx.Param("zoneID")
	questArchTypeID := ctx.Param("questArchTypeID")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	questArchTypeIDUUID, err := uuid.Parse(questArchTypeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest arch type ID"})
		return
	}

	var questGiverCharacterID *uuid.UUID
	if zoneQuestArchetype, err := s.dbClient.ZoneQuestArchetype().FindByZoneIDAndQuestArchetypeID(ctx, zoneID, questArchTypeIDUUID); err == nil {
		questGiverCharacterID = zoneQuestArchetype.CharacterID
	}

	payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
		ZoneID:                zoneID,
		QuestArchetypeID:      questArchTypeIDUUID,
		QuestGiverCharacterID: questGiverCharacterID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateQuestForZoneTaskType, payload)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "quest generation started"})
}

func (s *server) seedTreasureChests(ctx *gin.Context) {
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.SeedTreasureChestsTaskType, nil)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "treasure chest seeding job queued"})
}

func (s *server) getGooglePlaces(ctx *gin.Context) {
	query := ctx.Query("query")
	places, err := s.googlemapsClient.FindCandidatesByQuery(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, places)
}

func (s *server) getGooglePlace(ctx *gin.Context) {
	placeID := ctx.Param("placeID")
	place, err := s.googlemapsClient.FindPlaceByID(placeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, place)
}

func (s *server) refreshPointOfInterestImage(ctx *gin.Context) {
	var requestBody struct {
		PointOfInterestID uuid.UUID `json:"pointOfInterestID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poi, err := s.dbClient.PointOfInterest().FindByID(ctx, requestBody.PointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if poi == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "point of interest not found"})
		return
	}

	clearErr := ""
	if err := s.dbClient.PointOfInterest().UpdateImageGenerationStatus(
		ctx,
		requestBody.PointOfInterestID,
		models.PointOfInterestImageGenerationStatusQueued,
		&clearErr,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update point of interest: " + err.Error()})
		return
	}

	payload := jobs.GeneratePointOfInterestImageTaskPayload{
		PointOfInterestID: requestBody.PointOfInterestID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GeneratePointOfInterestImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = s.dbClient.PointOfInterest().UpdateImageGenerationStatus(
			ctx,
			requestBody.PointOfInterestID,
			models.PointOfInterestImageGenerationStatusFailed,
			&errMsg,
		)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedPoi, err := s.dbClient.PointOfInterest().FindByID(ctx, requestBody.PointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedPoi)
}

func (s *server) importPointOfInterest(ctx *gin.Context) {
	var requestBody struct {
		PlaceID string    `json:"placeID"`
		ZoneID  uuid.UUID `json:"zoneID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.PlaceID == "" || requestBody.ZoneID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "placeID and zoneID are required"})
		return
	}

	importItem := &models.PointOfInterestImport{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PlaceID:   requestBody.PlaceID,
		ZoneID:    requestBody.ZoneID,
		Status:    "queued",
	}
	if err := s.dbClient.PointOfInterestImport().Create(ctx, importItem); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.ImportPointOfInterestTaskPayload{
		ImportID: importItem.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ImportPointOfInterestTaskType, payload)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, importItem)
}

func (s *server) getPointOfInterestImports(ctx *gin.Context) {
	zoneIDParam := ctx.Query("zoneId")
	limit := 50
	var (
		items []models.PointOfInterestImport
		err   error
	)
	if zoneIDParam != "" {
		zoneID, parseErr := uuid.Parse(zoneIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
			return
		}
		items, err = s.dbClient.PointOfInterestImport().FindByZoneID(ctx, zoneID, limit)
	} else {
		items, err = s.dbClient.PointOfInterestImport().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getPointOfInterestImport(ctx *gin.Context) {
	id := ctx.Param("id")
	importID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid import ID"})
		return
	}
	item, err := s.dbClient.PointOfInterestImport().FindByID(ctx, importID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "import not found"})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (s *server) refreshPointOfInterest(ctx *gin.Context) {
	var requestBody struct {
		PointOfInterestID uuid.UUID `json:"pointOfInterestID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poi, err := s.dbClient.PointOfInterest().FindByID(ctx, requestBody.PointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = s.locationSeeder.RefreshPointOfInterest(ctx, poi)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, poi)
}

func (s *server) deleteZone(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	err = s.dbClient.Zone().Delete(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "zone deleted successfully"})
}

func (s *server) getPlaceTypes(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, googlemaps.GetAllPlaceTypes())
}

func (s *server) generatePointsOfInterestForZone(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		IncludedTypes  []googlemaps.PlaceType `json:"includedTypes"`
		ExcludedTypes  []googlemaps.PlaceType `json:"excludedTypes"`
		NumberOfPlaces int32                  `json:"numberOfPlaces"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pointsOfInterest, err := s.locationSeeder.SeedPointsOfInterest(ctx, *zone, requestBody.IncludedTypes, requestBody.ExcludedTypes, requestBody.NumberOfPlaces)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, poi := range pointsOfInterest {
		if poi == nil || strings.TrimSpace(poi.ImageUrl) == "" {
			continue
		}
		s.enqueueThumbnailTask(jobs.ThumbnailEntityPointOfInterest, poi.ID, poi.ImageUrl)
	}
	ctx.JSON(http.StatusOK, pointsOfInterest)
}

func (s *server) getPointsOfInterestForZone(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	pointsOfInterest, err := s.dbClient.PointOfInterest().FindAllForZone(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, pointsOfInterest)
}
func (s *server) getZone(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zone)
}

func (s *server) createZone(ctx *gin.Context) {
	var requestBody struct {
		Name        string  `json:"name"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Description string  `json:"description"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone := &models.Zone{
		Name:        requestBody.Name,
		Latitude:    requestBody.Latitude,
		Longitude:   requestBody.Longitude,
		Description: requestBody.Description,
	}
	if err := s.dbClient.Zone().Create(ctx, zone); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zone)
}

func (s *server) importZonesForMetro(ctx *gin.Context) {
	var requestBody struct {
		MetroName string `json:"metroName"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	metroName := strings.TrimSpace(requestBody.MetroName)
	if metroName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "metroName is required"})
		return
	}

	importItem := &models.ZoneImport{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MetroName: metroName,
		Status:    "queued",
		ZoneCount: 0,
	}
	if err := s.dbClient.ZoneImport().Create(ctx, importItem); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.ImportZonesForMetroTaskPayload{
		ImportID: importItem.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ImportZonesForMetroTaskType, payload)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, importItem)
}

func (s *server) getZoneImports(ctx *gin.Context) {
	metroName := strings.TrimSpace(ctx.Query("metroName"))
	limit := 50
	var (
		items []models.ZoneImport
		err   error
	)
	if metroName != "" {
		items, err = s.dbClient.ZoneImport().FindByMetroName(ctx, metroName, limit)
	} else {
		items, err = s.dbClient.ZoneImport().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getZoneImport(ctx *gin.Context) {
	id := ctx.Param("id")
	importID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid import ID"})
		return
	}
	item, err := s.dbClient.ZoneImport().FindByID(ctx, importID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "import not found"})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (s *server) deleteZoneImport(ctx *gin.Context) {
	id := ctx.Param("id")
	importID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid import ID"})
		return
	}

	item, err := s.dbClient.ZoneImport().FindByID(ctx, importID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "import not found"})
		return
	}

	deletedCount, err := s.dbClient.Zone().DeleteByImportID(ctx, importID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	item.Status = "deleted"
	item.UpdatedAt = time.Now()
	_ = s.dbClient.ZoneImport().Update(ctx, item)

	ctx.JSON(http.StatusOK, gin.H{"deletedCount": deletedCount})
}

func (s *server) seedZoneDraft(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if zone == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
		return
	}

	var requestBody struct {
		PlaceCount           *int     `json:"placeCount"`
		CharacterCount       *int     `json:"characterCount"`
		QuestCount           *int     `json:"questCount"`
		MainQuestCount       *int     `json:"mainQuestCount"`
		MonsterCount         *int     `json:"monsterCount"`
		InputEncounterCount  *int     `json:"inputEncounterCount"`
		OptionEncounterCount *int     `json:"optionEncounterCount"`
		RequiredPlaceTags    []string `json:"requiredPlaceTags"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	placeCount := 8
	if requestBody.PlaceCount != nil {
		placeCount = *requestBody.PlaceCount
	}
	characterCount := 4
	if requestBody.CharacterCount != nil {
		characterCount = *requestBody.CharacterCount
	}
	questCount := 4
	if requestBody.QuestCount != nil {
		questCount = *requestBody.QuestCount
	}
	mainQuestCount := 1
	if requestBody.MainQuestCount != nil {
		mainQuestCount = *requestBody.MainQuestCount
	}
	monsterCount := 6
	if requestBody.MonsterCount != nil {
		monsterCount = *requestBody.MonsterCount
	}
	inputEncounterCount := 0
	if requestBody.InputEncounterCount != nil {
		inputEncounterCount = *requestBody.InputEncounterCount
	}
	optionEncounterCount := 0
	if requestBody.OptionEncounterCount != nil {
		optionEncounterCount = *requestBody.OptionEncounterCount
	}
	requiredPlaceTags := make([]string, 0)
	if len(requestBody.RequiredPlaceTags) > 0 {
		seenTags := make(map[string]struct{})
		for _, tag := range requestBody.RequiredPlaceTags {
			normalized := strings.ToLower(strings.TrimSpace(tag))
			if normalized == "" {
				continue
			}
			if _, ok := seenTags[normalized]; ok {
				continue
			}
			seenTags[normalized] = struct{}{}
			requiredPlaceTags = append(requiredPlaceTags, normalized)
		}
	}

	if placeCount <= 0 || characterCount <= 0 || questCount <= 0 || mainQuestCount < 0 || monsterCount <= 0 || inputEncounterCount < 0 || optionEncounterCount < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "placeCount, characterCount, questCount, and monsterCount must be greater than zero; mainQuestCount, inputEncounterCount, and optionEncounterCount must be zero or greater"})
		return
	}
	if len(requiredPlaceTags) > placeCount {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "requiredPlaceTags cannot exceed placeCount"})
		return
	}

	job := &models.ZoneSeedJob{
		ID:                   uuid.New(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		ZoneID:               zoneID,
		Status:               models.ZoneSeedStatusQueued,
		PlaceCount:           placeCount,
		CharacterCount:       characterCount,
		QuestCount:           questCount,
		MainQuestCount:       mainQuestCount,
		MonsterCount:         monsterCount,
		InputEncounterCount:  inputEncounterCount,
		OptionEncounterCount: optionEncounterCount,
		RequiredPlaceTags:    models.StringArray(requiredPlaceTags),
	}
	if err := s.dbClient.ZoneSeedJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.SeedZoneDraftTaskPayload{
		JobID: job.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.SeedZoneDraftTaskType, payload)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *server) getZoneSeedJobs(ctx *gin.Context) {
	zoneIDParam := strings.TrimSpace(ctx.Query("zoneId"))
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		jobsList []models.ZoneSeedJob
		err      error
	)
	if zoneIDParam != "" {
		zoneID, parseErr := uuid.Parse(zoneIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
			return
		}
		jobsList, err = s.dbClient.ZoneSeedJob().FindByZoneID(ctx, zoneID, limit)
	} else {
		jobsList, err = s.dbClient.ZoneSeedJob().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getZoneSeedJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone seed job ID"})
		return
	}

	job, err := s.dbClient.ZoneSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone seed job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) approveZoneSeedJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone seed job ID"})
		return
	}

	job, err := s.dbClient.ZoneSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone seed job not found"})
		return
	}
	if job.Status != models.ZoneSeedStatusAwaitingApproval {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "zone seed job is not awaiting approval"})
		return
	}

	job.Status = models.ZoneSeedStatusApproved
	job.UpdatedAt = time.Now()
	if err := s.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.ApplyZoneSeedDraftTaskPayload{
		JobID: job.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ApplyZoneSeedDraftTaskType, payload)); err != nil {
		job.Status = models.ZoneSeedStatusAwaitingApproval
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ZoneSeedJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *server) retryZoneSeedJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone seed job ID"})
		return
	}

	job, err := s.dbClient.ZoneSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone seed job not found"})
		return
	}
	if job.Status == models.ZoneSeedStatusInProgress || job.Status == models.ZoneSeedStatusApplying {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot retry a job that is in progress or applying"})
		return
	}
	if job.Status != models.ZoneSeedStatusFailed {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "zone seed job is not in a failed state"})
		return
	}

	if zoneSeedDraftHasContent(job.Draft) {
		job.Status = models.ZoneSeedStatusAwaitingApproval
		job.ErrorMessage = nil
		job.UpdatedAt = time.Now()
		if err := s.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, job)
		return
	}

	job.Status = models.ZoneSeedStatusQueued
	job.ErrorMessage = nil
	job.Draft = models.ZoneSeedDraft{}
	job.UpdatedAt = time.Now()
	if err := s.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.SeedZoneDraftTaskPayload{
		JobID: job.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.SeedZoneDraftTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.ZoneSeedStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ZoneSeedJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}

func (s *server) shuffleZoneSeedJobChallenge(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone seed job ID"})
		return
	}

	var requestBody struct {
		QuestDraftID         *string `json:"questDraftId"`
		MainQuestNodeDraftID *string `json:"mainQuestNodeDraftId"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hasQuest := requestBody.QuestDraftID != nil && strings.TrimSpace(*requestBody.QuestDraftID) != ""
	hasNode := requestBody.MainQuestNodeDraftID != nil && strings.TrimSpace(*requestBody.MainQuestNodeDraftID) != ""
	if hasQuest == hasNode {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "provide exactly one target: questDraftId or mainQuestNodeDraftId"})
		return
	}

	job, err := s.dbClient.ZoneSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone seed job not found"})
		return
	}
	if job.Status == models.ZoneSeedStatusInProgress || job.Status == models.ZoneSeedStatusApplying || job.Status == models.ZoneSeedStatusApplied {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot shuffle challenge for this job status"})
		return
	}
	if !zoneSeedDraftHasContent(job.Draft) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "zone seed job has no draft content"})
		return
	}

	payload := jobs.ShuffleZoneSeedChallengeTaskPayload{JobID: jobID}
	if hasQuest {
		questDraftID, parseErr := uuid.Parse(strings.TrimSpace(*requestBody.QuestDraftID))
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid questDraftId"})
			return
		}
		if !zoneSeedDraftHasQuest(job.Draft, questDraftID) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "questDraftId not found in draft"})
			return
		}
		payload.QuestDraftID = &questDraftID
		if !job.Draft.SetQuestChallengeShuffleStatus(questDraftID, models.ZoneSeedChallengeShuffleStatusQueued, nil) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "questDraftId not found in draft"})
			return
		}
	} else {
		nodeDraftID, parseErr := uuid.Parse(strings.TrimSpace(*requestBody.MainQuestNodeDraftID))
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mainQuestNodeDraftId"})
			return
		}
		if !zoneSeedDraftHasMainQuestNode(job.Draft, nodeDraftID) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "mainQuestNodeDraftId not found in draft"})
			return
		}
		payload.MainQuestNodeDraftID = &nodeDraftID
		if !job.Draft.SetMainQuestNodeChallengeShuffleStatus(nodeDraftID, models.ZoneSeedChallengeShuffleStatusQueued, nil) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "mainQuestNodeDraftId not found in draft"})
			return
		}
	}

	job.UpdatedAt = time.Now()
	if err := s.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ShuffleZoneSeedChallengeTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		if payload.QuestDraftID != nil {
			_ = job.Draft.SetQuestChallengeShuffleStatus(*payload.QuestDraftID, models.ZoneSeedChallengeShuffleStatusFailed, &errMsg)
		}
		if payload.MainQuestNodeDraftID != nil {
			_ = job.Draft.SetMainQuestNodeChallengeShuffleStatus(*payload.MainQuestNodeDraftID, models.ZoneSeedChallengeShuffleStatusFailed, &errMsg)
		}
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ZoneSeedJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"queued": true, "jobId": jobID})
}

func zoneSeedDraftHasContent(d models.ZoneSeedDraft) bool {
	if strings.TrimSpace(d.FantasyName) != "" {
		return true
	}
	if strings.TrimSpace(d.ZoneDescription) != "" {
		return true
	}
	if len(d.PointsOfInterest) > 0 {
		return true
	}
	if len(d.Characters) > 0 {
		return true
	}
	if len(d.Quests) > 0 {
		return true
	}
	if len(d.MainQuests) > 0 {
		return true
	}
	return false
}

func zoneSeedDraftHasQuest(d models.ZoneSeedDraft, questDraftID uuid.UUID) bool {
	for _, quest := range d.Quests {
		if quest.DraftID == questDraftID {
			return true
		}
	}
	return false
}

func zoneSeedDraftHasMainQuestNode(d models.ZoneSeedDraft, nodeDraftID uuid.UUID) bool {
	for _, mainQuest := range d.MainQuests {
		for _, node := range mainQuest.Nodes {
			if node.DraftID == nodeDraftID {
				return true
			}
		}
	}
	return false
}

func (s *server) deleteZoneSeedJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone seed job ID"})
		return
	}

	job, err := s.dbClient.ZoneSeedJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone seed job not found"})
		return
	}

	if job.Status == models.ZoneSeedStatusInProgress || job.Status == models.ZoneSeedStatusApplying {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a job that is in progress or applying"})
		return
	}

	if err := s.dbClient.ZoneSeedJob().DeleteByID(ctx, job.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *server) createScenarioGenerationJob(ctx *gin.Context) {
	var requestBody scenarioGenerationJobRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(strings.TrimSpace(requestBody.ZoneID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
		return
	}

	if (requestBody.Latitude == nil) != (requestBody.Longitude == nil) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude must be provided together"})
		return
	}
	if requestBody.Latitude != nil && requestBody.Longitude != nil {
		if math.IsNaN(*requestBody.Latitude) || math.IsInf(*requestBody.Latitude, 0) || *requestBody.Latitude < -90 || *requestBody.Latitude > 90 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "latitude must be between -90 and 90"})
			return
		}
		if math.IsNaN(*requestBody.Longitude) || math.IsInf(*requestBody.Longitude, 0) || *requestBody.Longitude < -180 || *requestBody.Longitude > 180 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "longitude must be between -180 and 180"})
			return
		}
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if zone == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
		return
	}

	job := &models.ScenarioGenerationJob{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ZoneID:    zoneID,
		Status:    models.ScenarioGenerationStatusQueued,
		OpenEnded: requestBody.OpenEnded,
		Latitude:  requestBody.Latitude,
		Longitude: requestBody.Longitude,
	}
	if err := s.dbClient.ScenarioGenerationJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateScenarioTaskPayload{
		JobID: job.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateScenarioTaskType, payload)); err != nil {
		errMsg := err.Error()
		job.Status = models.ScenarioGenerationStatusFailed
		job.ErrorMessage = &errMsg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.ScenarioGenerationJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getScenarioGenerationJobs(ctx *gin.Context) {
	zoneIDParam := strings.TrimSpace(ctx.Query("zoneId"))
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		jobsList []models.ScenarioGenerationJob
		err      error
	)
	if zoneIDParam != "" {
		zoneID, parseErr := uuid.Parse(zoneIDParam)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
			return
		}
		jobsList, err = s.dbClient.ScenarioGenerationJob().FindByZoneID(ctx, zoneID, limit)
	} else {
		jobsList, err = s.dbClient.ScenarioGenerationJob().FindRecent(ctx, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getScenarioGenerationJob(ctx *gin.Context) {
	id := ctx.Param("id")
	jobID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario generation job ID"})
		return
	}

	job, err := s.dbClient.ScenarioGenerationJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario generation job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) getZones(ctx *gin.Context) {
	zones, err := s.dbClient.Zone().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zones)
}

func (s *server) addPointOfInterestToZone(ctx *gin.Context) {
	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	pointOfInterestID, err := uuid.Parse(ctx.Param("pointOfInterestId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest ID"})
		return
	}

	err = s.dbClient.Zone().AddPointOfInterestToZone(ctx, zoneID, pointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "point of interest added to zone successfully"})
}

func (s *server) removePointOfInterestFromZone(ctx *gin.Context) {
	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	pointOfInterestID, err := uuid.Parse(ctx.Param("pointOfInterestId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest ID"})
		return
	}

	err = s.dbClient.Zone().RemovePointOfInterestFromZone(ctx, zoneID, pointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "point of interest removed from zone successfully"})
}

func (s *server) getZoneForPointOfInterest(ctx *gin.Context) {
	pointOfInterestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest ID"})
		return
	}

	zone, err := s.dbClient.Zone().FindByPointOfInterestID(ctx, pointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found for this point of interest"})
		return
	}
	ctx.JSON(http.StatusOK, zone)
}

func (s *server) addTagToPointOfInterest(ctx *gin.Context) {
	var requestBody struct {
		TagID             uuid.UUID `json:"tagID"`
		PointOfInterestID uuid.UUID `json:"pointOfInterestID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.dbClient.Tag().AddTagToPointOfInterest(ctx, requestBody.TagID, requestBody.PointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "tag added to point of interest successfully"})
}

func (s *server) removeTagFromPointOfInterest(ctx *gin.Context) {
	tagID := ctx.Param("tagID")
	pointOfInterestID := ctx.Param("pointOfInterestID")
	tagIDUUID, err := uuid.Parse(tagID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag ID"})
		return
	}
	pointOfInterestIDUUID, err := uuid.Parse(pointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest ID"})
		return
	}

	err = s.dbClient.Tag().RemoveTagFromPointOfInterest(ctx, tagIDUUID, pointOfInterestIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "tag removed from point of interest successfully"})
}

func (s *server) getTags(ctx *gin.Context) {
	tags, err := s.dbClient.Tag().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tags)
}

func (s *server) getProficiencies(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("query"))
	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	proficiencies, err := s.dbClient.UserProficiency().FindDistinctProficiencies(ctx, query, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, proficiencies)
}

func (s *server) getTagGroups(ctx *gin.Context) {
	tagGroups, err := s.dbClient.TagGroup().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tagGroups)
}

func (s *server) getAllUsers(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	users, err := s.dbClient.User().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (s *server) giveItem(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		UserID   *uuid.UUID `json:"userID"`
		TeamID   *uuid.UUID `json:"teamID"`
		ItemID   int        `binding:"required" json:"itemID"`
		Quantity int        `binding:"required" json:"quantity"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(
		ctx,
		requestBody.TeamID,
		requestBody.UserID,
		requestBody.ItemID,
		requestBody.Quantity,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "item given to user successfully"})
}

func (s *server) getNewUserStarterConfig(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config, err := s.dbClient.NewUserStarterConfig().Get(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, config)
}

func (s *server) updateNewUserStarterConfig(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Gold  int                         `json:"gold"`
		Items []models.NewUserStarterItem `json:"items"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Gold < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "gold must be non-negative"})
		return
	}

	// Deduplicate and validate items.
	itemMap := map[int]int{}
	for _, item := range requestBody.Items {
		if item.InventoryItemID <= 0 || item.Quantity <= 0 {
			continue
		}
		itemMap[item.InventoryItemID] += item.Quantity
	}

	items := make([]models.NewUserStarterItem, 0, len(itemMap))
	for itemID, qty := range itemMap {
		if _, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, itemID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("inventory item %d not found", itemID)})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, models.NewUserStarterItem{
			InventoryItemID: itemID,
			Quantity:        qty,
		})
	}

	config := &models.NewUserStarterConfig{
		Gold:  requestBody.Gold,
		Items: items,
	}
	updated, err := s.dbClient.NewUserStarterConfig().Upsert(ctx, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (s *server) hasCurrentMatch(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	matchID, err := s.dbClient.Match().FindCurrentMatchIDForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"hasCurrentMatch": matchID != nil,
		"matchID":         matchID,
	})
}

func (s *server) getQuestLog(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stringZoneID := ctx.Query("zoneId")
	zoneID := uuid.UUID{}
	if stringZoneID != "" {
		parsed, err := uuid.Parse(stringZoneID)
		if err != nil {
			log.Printf("getQuestLog: invalid zoneId '%s', will try location-based fallback", stringZoneID)
		} else {
			zoneID = parsed
		}
	}
	if zoneID == uuid.Nil {
		userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "zone ID is required when user location is unavailable"})
			return
		}
		zones, err := s.dbClient.Zone().FindAll(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, zone := range zones {
			if zone == nil {
				continue
			}
			if zone.IsPointInBoundary(userLat, userLng) {
				zoneID = zone.ID
				break
			}
		}
		if zoneID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "no zone found for user location"})
			return
		}
	}
	var tags []string
	if tagsQuery := ctx.Query("tags"); tagsQuery != "" {
		tags = strings.Split(tagsQuery, ",")
	}

	questLog, err := s.questlogClient.GetQuestLog(ctx, user.ID, zoneID, tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, questLog)
}

func (s *server) inviteToMatch(ctx *gin.Context) {
	matchId := ctx.Param("id")
	matchID, err := uuid.Parse(matchId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	var requestBody struct {
		UserID uuid.UUID `binding:"required" json:"userId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	matchUser := &models.MatchUser{
		MatchID: matchID,
		UserID:  requestBody.UserID,
	}

	if err := s.dbClient.MatchUser().Create(ctx, matchUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user invited to match successfully"})
}

func (s *server) getOwnedInventoryItems(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userOrTeam := models.OwnedInventoryItem{UserID: &user.ID}

	matchID, err := s.dbClient.Match().FindCurrentMatchIDForUser(ctx, user.ID)
	if matchID != nil && err == nil {
		teams, err := s.dbClient.Team().GetByMatchID(ctx, *matchID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, team := range teams {
			for _, user := range team.Users {
				if user.ID == user.ID {
					userOrTeam.TeamID = &team.ID
					break
				}
			}
		}
	}

	items, err := s.dbClient.InventoryItem().GetItems(ctx, userOrTeam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getPointOfInterestDiscoveries(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	matchID, err := s.dbClient.Match().FindCurrentMatchIDForUser(ctx, user.ID)
	if matchID == nil || err != nil {
		discoveries, err := s.dbClient.PointOfInterestDiscovery().GetDiscoveriesForUser(user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, discoveries)
		return
	}

	teams, err := s.dbClient.Team().GetByMatchID(ctx, *matchID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var teamID uuid.UUID
	for _, team := range teams {
		for _, user := range team.Users {
			if user.ID == user.ID {
				teamID = team.ID
				break
			}
		}
	}

	discoveries, err := s.dbClient.PointOfInterestDiscovery().GetDiscoveriesForTeam(teamID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, discoveries)
}

func (s *server) getPointOfInterestChallengeSubmissions(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	matchID, err := s.dbClient.Match().FindCurrentMatchIDForUser(ctx, user.ID)
	if matchID == nil || err != nil {
		submissions, err := s.dbClient.PointOfInterestChallenge().GetSubmissionsForUser(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, submissions)
		return
	}

	submissions, err := s.dbClient.PointOfInterestChallenge().GetSubmissionsForMatch(ctx, *matchID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, submissions)
}

func (s *server) createPointOfInterestChildren(ctx *gin.Context) {
	var requestBody struct {
		PointOfInterestGroupMemberID uuid.UUID `binding:"required" json:"pointOfInterestGroupMemberId"`
		PointOfInterestID            uuid.UUID `binding:"required" json:"pointOfInterestId"`
		PointOfInterestChallengeID   uuid.UUID `binding:"required" json:"pointOfInterestChallengeId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the current group member to find the group ID
	currentGroupMember, err := s.dbClient.PointOfInterestGroupMember().FindByID(ctx, requestBody.PointOfInterestGroupMemberID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find current group member: " + err.Error()})
		return
	}

	// Find the group member for the next point of interest
	nextGroupMember, err := s.dbClient.PointOfInterestGroupMember().FindByPointOfInterestAndGroup(ctx, requestBody.PointOfInterestID, currentGroupMember.PointOfInterestGroupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find next point group member: " + err.Error()})
		return
	}

	if err := s.dbClient.PointOfInterestChildren().Create(ctx, requestBody.PointOfInterestGroupMemberID, nextGroupMember.ID, requestBody.PointOfInterestChallengeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "point of interest children created successfully"})
}

func (s *server) deletePointOfInterestChildren(ctx *gin.Context) {
	stringPointOfInterestChildrenID := ctx.Param("id")
	if stringPointOfInterestChildrenID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "point of interest children ID is required"})
		return
	}

	pointOfInterestChildrenID, err := uuid.Parse(stringPointOfInterestChildrenID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest children ID"})
		return
	}

	if err := s.dbClient.PointOfInterestChildren().Delete(ctx, pointOfInterestChildrenID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "point of interest children deleted successfully"})
}

func (s *server) getMapboxPlaces(ctx *gin.Context) {
	address := ctx.Query("address")
	places, err := s.mapboxClient.GetPlaces(ctx, address)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, places)
}

func (s *server) editPointOfInterestGroupImageUrl(ctx *gin.Context) {
	stringPointOfInterestGroupID := ctx.Param("id")
	if stringPointOfInterestGroupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest group ID is required",
		})
		return
	}

	pointOfInterestGroupID, err := uuid.Parse(stringPointOfInterestGroupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest group ID",
		})
		return
	}

	var requestBody struct {
		ImageUrl string `binding:"required" json:"imageUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterestGroup().UpdateImageUrl(ctx, pointOfInterestGroupID, requestBody.ImageUrl); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest group image URL edited successfully",
	})
}

func (s *server) editPointOfInterestImageUrl(ctx *gin.Context) {
	stringPointOfInterestID := ctx.Param("id")
	if stringPointOfInterestID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest ID is required",
		})
		return
	}

	pointOfInterestID, err := uuid.Parse(stringPointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest ID",
		})
		return
	}

	var requestBody struct {
		ImageUrl string `binding:"required" json:"imageUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterest().UpdateImageUrl(ctx, pointOfInterestID, requestBody.ImageUrl); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if strings.TrimSpace(requestBody.ImageUrl) != "" {
		s.enqueueThumbnailTask(jobs.ThumbnailEntityPointOfInterest, pointOfInterestID, requestBody.ImageUrl)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest image URL edited successfully",
	})
}

func (s *server) editPointOfInterest(ctx *gin.Context) {
	stringPointOfInterestID := ctx.Param("id")
	if stringPointOfInterestID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest ID is required",
		})
		return
	}

	pointOfInterestID, err := uuid.Parse(stringPointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest ID",
		})
		return
	}

	var requestBody struct {
		Name              string  `binding:"required" json:"name"`
		Description       string  `binding:"required" json:"description"`
		Lat               string  `binding:"required" json:"lat"`
		Lng               string  `binding:"required" json:"lng"`
		UnlockTier        *int    `json:"unlockTier"`
		Clue              string  `json:"clue"`
		ImageUrl          string  `json:"imageUrl"`
		OriginalName      string  `json:"originalName"`
		GoogleMapsPlaceID *string `json:"googleMapsPlaceId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	existingPoi, err := s.dbClient.PointOfInterest().FindByID(ctx, pointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Edit(
		ctx,
		pointOfInterestID,
		requestBody.Name,
		requestBody.Description,
		requestBody.Lat,
		requestBody.Lng,
		requestBody.UnlockTier,
		requestBody.Clue,
		requestBody.ImageUrl,
		requestBody.OriginalName,
		requestBody.GoogleMapsPlaceID,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	imageChanged := existingPoi == nil || requestBody.ImageUrl != existingPoi.ImageUrl
	thumbnailMissing := existingPoi == nil || strings.TrimSpace(existingPoi.ThumbnailURL) == ""
	if strings.TrimSpace(requestBody.ImageUrl) != "" && imageChanged {
		if err := s.dbClient.PointOfInterest().UpdateImageUrl(ctx, pointOfInterestID, requestBody.ImageUrl); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	if strings.TrimSpace(requestBody.ImageUrl) != "" && (imageChanged || thumbnailMissing) {
		s.enqueueThumbnailTask(jobs.ThumbnailEntityPointOfInterest, pointOfInterestID, requestBody.ImageUrl)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest edited successfully",
	})
}

func (s *server) createPointOfInterestChallenge(ctx *gin.Context) {
	var requestBody struct {
		PointOfInterestID      uuid.UUID `binding:"required" json:"pointOfInterestId"`
		Tier                   int       `binding:"required" json:"tier"`
		Question               string    `binding:"required" json:"question"`
		InventoryItemID        int       `binding:"required" json:"inventoryItemId"`
		PointOfInterestGroupID uuid.UUID `json:"pointOfInterestGroupId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := s.dbClient.PointOfInterestChallenge().Create(ctx, requestBody.PointOfInterestID, requestBody.Tier, requestBody.Question, requestBody.InventoryItemID, &requestBody.PointOfInterestGroupID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest challenge created successfully",
	})
}

func (s *server) editPointOfInterestChallenge(ctx *gin.Context) {
	stringPointOfInterestChallengeID := ctx.Param("id")
	if stringPointOfInterestChallengeID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest challenge ID is required",
		})
		return
	}

	pointOfInterestChallengeID, err := uuid.Parse(stringPointOfInterestChallengeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest challenge ID",
		})
		return
	}

	var requestBody struct {
		Question        string `binding:"required" json:"question"`
		InventoryItemID int    `binding:"required" json:"inventoryItemId"`
		Tier            int    `binding:"required" json:"tier"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := s.dbClient.PointOfInterestChallenge().Edit(ctx, pointOfInterestChallengeID, requestBody.Question, requestBody.InventoryItemID, requestBody.Tier); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest challenge edited successfully",
	})
}

func (s *server) deletePointOfInterestChallenge(ctx *gin.Context) {
	stringPointOfInterestChallengeID := ctx.Param("id")
	if stringPointOfInterestChallengeID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest challenge ID is required",
		})
		return
	}

	pointOfInterestChallengeID, err := uuid.Parse(stringPointOfInterestChallengeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest challenge ID",
		})
		return
	}

	if err := s.dbClient.PointOfInterestChallenge().Delete(ctx, pointOfInterestChallengeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest challenge deleted successfully",
	})
}

func (s *server) deletePointOfInterest(ctx *gin.Context) {
	stringPointOfInterestID := ctx.Param("id")
	if stringPointOfInterestID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest ID is required",
		})
		return
	}

	pointOfInterestID, err := uuid.Parse(stringPointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest ID",
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Delete(ctx, pointOfInterestID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest deleted successfully",
	})
}

func (s *server) deletePointOfInterestGroup(ctx *gin.Context) {
	stringPointOfInterestGroupID := ctx.Param("id")
	if stringPointOfInterestGroupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest group ID is required",
		})
		return
	}

	pointOfInterestGroupID, err := uuid.Parse(stringPointOfInterestGroupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest group ID",
		})
		return
	}

	if err := s.dbClient.PointOfInterestGroup().Delete(ctx, pointOfInterestGroupID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest group deleted successfully",
	})
}

func (s *server) bulkDeletePointOfInterestGroups(ctx *gin.Context) {
	var request struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ids array is required",
		})
		return
	}

	if len(request.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ids array cannot be empty",
		})
		return
	}

	ids := make([]uuid.UUID, 0, len(request.IDs))
	for _, idStr := range request.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid UUID: %s", idStr),
			})
			return
		}
		ids = append(ids, id)
	}

	if err := s.dbClient.PointOfInterestGroup().DeleteByIDs(ctx, ids); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("successfully deleted %d point of interest group(s)", len(ids)),
		"deleted": len(ids),
	})
}

func (s *server) editPointOfInterestGroup(ctx *gin.Context) {
	stringPointOfInterestGroupID := ctx.Param("id")
	if stringPointOfInterestGroupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest group ID is required",
		})
		return
	}

	pointOfInterestGroupID, err := uuid.Parse(stringPointOfInterestGroupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest group ID",
		})
		return
	}

	var requestBody struct {
		Name                    string     `binding:"required" json:"name"`
		Description             string     `binding:"required" json:"description"`
		Type                    int        `binding:"required" json:"type"`
		Gold                    *int       `json:"gold"`
		InventoryItemID         *int       `json:"inventoryItemId"`
		RequiredReputationLevel *int       `json:"requiredReputationLevel"`
		QuestGiverCharacterID   *uuid.UUID `json:"questGiverCharacterId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	typeValue := models.PointOfInterestGroupType(requestBody.Type)

	updates := &models.PointOfInterestGroup{
		Name:        requestBody.Name,
		Description: requestBody.Description,
		Type:        typeValue,
	}
	if requestBody.Gold != nil {
		updates.Gold = *requestBody.Gold
	}
	if requestBody.InventoryItemID != nil {
		updates.InventoryItemID = requestBody.InventoryItemID
	}
	if requestBody.RequiredReputationLevel != nil {
		updates.RequiredReputationLevel = requestBody.RequiredReputationLevel
	}
	if requestBody.QuestGiverCharacterID != nil {
		character, err := s.dbClient.Character().FindByID(ctx, *requestBody.QuestGiverCharacterID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if character == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest giver character not found"})
			return
		}
		updates.QuestGiverCharacterID = requestBody.QuestGiverCharacterID
	}
	if err := s.dbClient.PointOfInterestGroup().Update(ctx, pointOfInterestGroupID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
}

func (s *server) createPointOfInterest(ctx *gin.Context) {
	stringPointOfInterestGroupID := ctx.Param("id")
	if stringPointOfInterestGroupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "point of interest group ID is required",
		})
		return
	}

	pointOfInterestGroupID, err := uuid.Parse(stringPointOfInterestGroupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid point of interest group ID",
		})
		return
	}

	var request struct {
		Name        string `binding:"required" json:"name"`
		Description string `binding:"required" json:"description"`
		Latitude    string `binding:"required" json:"latitude"`
		Longitude   string `binding:"required" json:"longitude"`
		ImageUrl    string `binding:"required" json:"imageUrl"`
		Clue        string `binding:"required" json:"clue"`
		UnlockTier  *int   `json:"unlockTier"`
	}

	if err := ctx.Bind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	poi := &models.PointOfInterest{
		Name:        request.Name,
		Description: request.Description,
		Lat:         request.Latitude,
		Lng:         request.Longitude,
		ImageUrl:    request.ImageUrl,
		Clue:        request.Clue,
		UnlockTier:  request.UnlockTier,
	}

	if err := s.dbClient.PointOfInterest().CreateForGroup(ctx, poi, pointOfInterestGroupID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if strings.TrimSpace(poi.ImageUrl) != "" {
		s.enqueueThumbnailTask(jobs.ThumbnailEntityPointOfInterest, poi.ID, poi.ImageUrl)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest created successfully",
	})
}

func (s *server) enqueueThumbnailTask(entityType string, entityID uuid.UUID, sourceUrl string) {
	if s.asyncClient == nil || strings.TrimSpace(sourceUrl) == "" {
		return
	}
	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType: entityType,
		EntityID:   entityID,
		SourceUrl:  sourceUrl,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal thumbnail task payload: %v", err)
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		log.Printf("Failed to enqueue thumbnail task: %v", err)
	}
}

func staticThumbnailURL(destinationKey string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", jobs.ThumbnailBucket, destinationKey)
}

func (s *server) setStaticThumbnailRequestedAt(ctx *gin.Context, statusKey string, requestedAt time.Time) {
	if s.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	_ = s.redisClient.Set(
		ctx.Request.Context(),
		statusKey,
		requestedAt.UTC().Format(time.RFC3339Nano),
		staticThumbnailStatusTTL,
	).Err()
}

func (s *server) clearStaticThumbnailRequestedAt(ctx *gin.Context, statusKey string) {
	if s.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	_ = s.redisClient.Del(ctx.Request.Context(), statusKey).Err()
}

func (s *server) getStaticThumbnailRequestedAt(ctx *gin.Context, statusKey string) (*time.Time, error) {
	if s.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return nil, nil
	}
	value, err := s.redisClient.Get(ctx.Request.Context(), statusKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	parsed, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if parseErr != nil {
		return nil, parseErr
	}
	return &parsed, nil
}

func (s *server) queueGeneratedStaticThumbnail(ctx *gin.Context, defaultPrompt string, destinationKey string, statusKey string) {
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}

	var requestBody struct {
		Prompt *string `json:"prompt"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := strings.TrimSpace(defaultPrompt)
	if requestBody.Prompt != nil {
		prompt = strings.TrimSpace(*requestBody.Prompt)
		if prompt == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt cannot be blank"})
			return
		}
	}

	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)

	sourceImageURL, err := s.deepPriest.GenerateImage(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType:     jobs.ThumbnailEntityStatic,
		SourceUrl:      sourceImageURL,
		DestinationKey: destinationKey,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	requestedAt := time.Now().UTC()
	s.setStaticThumbnailRequestedAt(ctx, statusKey, requestedAt)

	ctx.JSON(http.StatusOK, gin.H{
		"status":       "queued",
		"prompt":       prompt,
		"sourceImage":  sourceImageURL,
		"thumbnailUrl": staticThumbnailURL(destinationKey),
		"requestedAt":  requestedAt.Format(time.RFC3339Nano),
	})
}

func (s *server) generateScenarioUndiscoveredIcon(ctx *gin.Context) {
	s.queueGeneratedStaticThumbnail(ctx, scenarioUndiscoveredIconText, scenarioUndiscoveredIconKey, scenarioUndiscoveredStatusKey)
}

func (s *server) generateMonsterUndiscoveredIcon(ctx *gin.Context) {
	s.queueGeneratedStaticThumbnail(ctx, monsterUndiscoveredIconText, monsterUndiscoveredIconKey, monsterUndiscoveredStatusKey)
}

func (s *server) getStaticThumbnailStatus(ctx *gin.Context, destinationKey string, statusKey string) {
	if strings.TrimSpace(destinationKey) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing thumbnail destination key"})
		return
	}
	lastModified, err := s.awsClient.GetObjectLastModified(jobs.ThumbnailBucket, destinationKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	requestedAt, requestedAtErr := s.getStaticThumbnailRequestedAt(ctx, statusKey)
	if requestedAtErr != nil {
		log.Printf("Failed to read static thumbnail status key %s: %v", statusKey, requestedAtErr)
	}

	status := "missing"
	exists := lastModified != nil
	if requestedAt != nil {
		if exists && !lastModified.Before(*requestedAt) {
			status = "completed"
			s.clearStaticThumbnailRequestedAt(ctx, statusKey)
		} else if time.Since(*requestedAt) > staticThumbnailJobTimeout {
			status = "failed"
		} else if exists {
			status = "in_progress"
		} else {
			status = "queued"
		}
	} else if exists {
		status = "completed"
	}

	response := gin.H{
		"status":       status,
		"exists":       exists,
		"thumbnailUrl": staticThumbnailURL(destinationKey),
	}
	if lastModified != nil {
		response["lastModified"] = lastModified.UTC().Format(time.RFC3339Nano)
	}
	if requestedAt != nil {
		response["requestedAt"] = requestedAt.UTC().Format(time.RFC3339Nano)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getScenarioUndiscoveredIconStatus(ctx *gin.Context) {
	s.getStaticThumbnailStatus(ctx, scenarioUndiscoveredIconKey, scenarioUndiscoveredStatusKey)
}

func (s *server) getMonsterUndiscoveredIconStatus(ctx *gin.Context) {
	s.getStaticThumbnailStatus(ctx, monsterUndiscoveredIconKey, monsterUndiscoveredStatusKey)
}

func (s *server) deleteStaticThumbnail(ctx *gin.Context, destinationKey string, statusKey string) {
	if strings.TrimSpace(destinationKey) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing thumbnail destination key"})
		return
	}
	if err := s.awsClient.DeleteObjectFromS3(jobs.ThumbnailBucket, destinationKey); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.clearStaticThumbnailRequestedAt(ctx, statusKey)
	ctx.JSON(http.StatusOK, gin.H{
		"status":       "deleted",
		"thumbnailUrl": staticThumbnailURL(destinationKey),
	})
}

func (s *server) deleteScenarioUndiscoveredIcon(ctx *gin.Context) {
	s.deleteStaticThumbnail(ctx, scenarioUndiscoveredIconKey, scenarioUndiscoveredStatusKey)
}

func (s *server) deleteMonsterUndiscoveredIcon(ctx *gin.Context) {
	s.deleteStaticThumbnail(ctx, monsterUndiscoveredIconKey, monsterUndiscoveredStatusKey)
}

func (s *server) queuePoiPlaceholderThumbnail(ctx *gin.Context) {
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}

	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType:     jobs.ThumbnailEntityStatic,
		SourceUrl:      poiPlaceholderImageURL,
		DestinationKey: poiPlaceholderThumbnailKey,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"thumbnailUrl": staticThumbnailURL(poiPlaceholderThumbnailKey),
		"status":       "queued",
	})
}

func (s *server) setProfilePicture(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		ProfilePictureUrl string `binding:"required" json:"profilePictureUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.User().UpdateProfilePictureUrl(ctx, user.ID, requestBody.ProfilePictureUrl); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "profile picture set successfully",
	})
}

func (s *server) getCompleteGenerationsForUser(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	generations, err := s.dbClient.ImageGeneration().GetCompleteGenerationsForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, generations)
}

func (s *server) generateProfilePictureOptions(ctx *gin.Context) {
	var requestBody struct {
		ProfilePictureUrl string `binding:"required" json:"profilePictureUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.User().UpdateProfilePictureUrl(ctx, user.ID, models.LoadingProfilePictureUrl); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	createProfilePayload := jobs.CreateProfilePictureTaskPayload{
		UserID:            user.ID,
		ProfilePictureUrl: requestBody.ProfilePictureUrl,
	}

	payloadBytes, err := json.Marshal(createProfilePayload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.CreateProfilePictureTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "success!",
	})
}

func (s *server) unlockPointOfInterestForTeam(ctx *gin.Context) {
	var requestBody struct {
		TeamID            *uuid.UUID `binding:"required" json:"teamId"`
		PointOfInterestID uuid.UUID  `json:"pointOfInterestId,omitempty"`
		UserID            *uuid.UUID `json:"userId,omitempty"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Unlock(ctx, requestBody.PointOfInterestID, requestBody.TeamID, requestBody.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.chatClient.AddUnlockMessage(ctx, requestBody.TeamID, requestBody.UserID, requestBody.PointOfInterestID); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest unlocked successfully",
	})
}

func (s *server) getChat(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	matchID, err := s.dbClient.Match().FindCurrentMatchIDForUser(ctx, user.ID)
	if matchID == nil || err != nil {
		auditItems, err := s.dbClient.AuditItem().GetAuditItemsForUser(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, auditItems)
		return
	}

	auditItems, err := s.dbClient.AuditItem().GetAuditItemsForMatch(ctx, *matchID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, auditItems)
}

func (s *server) addItemToTeam(ctx *gin.Context) {
	stringTeamID := ctx.Param("teamID")
	if stringTeamID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "team ID is required",
		})
		return
	}

	teamID, err := uuid.Parse(stringTeamID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid team ID",
		})
		return
	}

	item, err := s.quartermaster.GetItem(ctx, &teamID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (s *server) useItem(ctx *gin.Context) {
	stringOwnedInventoryItemID := ctx.Param("ownedInventoryItemID")
	if stringOwnedInventoryItemID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "owned inventory item ID is required",
		})
		return
	}

	ownedInventoryItemID, err := uuid.Parse(stringOwnedInventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid owned inventory item ID",
		})
		return
	}

	var request quartermaster.UseItemMetadata
	if err := ctx.Bind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ownedInventoryItem, err := s.dbClient.InventoryItem().FindByID(ctx, ownedInventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	inventoryItem, err := s.quartermaster.FindItemForItemID(ownedInventoryItem.InventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if isOutfitItemName(inventoryItem.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "outfit items require a selfie. Use the outfit endpoint.",
		})
		return
	}

	if err := s.quartermaster.UseItem(ctx, ownedInventoryItemID, &request); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.chatClient.AddUseItemMessage(ctx, *ownedInventoryItem, request); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if inventoryItem.IsCaptureType {
		challenge, err := s.dbClient.PointOfInterestChallenge().FindByID(ctx, request.ChallengeID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		result, err := s.gameEngineClient.ProcessSuccessfulSubmission(ctx, gameengine.Submission{
			TeamID:      ownedInventoryItem.TeamID,
			UserID:      ownedInventoryItem.UserID,
			ChallengeID: challenge.ID,
			Text:        "",
			ImageURL:    "",
		}, challenge)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, result)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "item used successfully",
	})
}

func (s *server) useOutfitItem(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stringOwnedInventoryItemID := ctx.Param("ownedInventoryItemID")
	if stringOwnedInventoryItemID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "owned inventory item ID is required"})
		return
	}

	ownedInventoryItemID, err := uuid.Parse(stringOwnedInventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid owned inventory item ID"})
		return
	}

	var requestBody struct {
		SelfieUrl string `json:"selfieUrl" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(requestBody.SelfieUrl) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "selfieUrl is required"})
		return
	}

	ownedInventoryItem, err := s.dbClient.InventoryItem().FindByID(ctx, ownedInventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ownedInventoryItem == nil || ownedInventoryItem.UserID == nil || *ownedInventoryItem.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "item does not belong to user"})
		return
	}
	if ownedInventoryItem.Quantity <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item is no longer available"})
		return
	}

	inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, ownedInventoryItem.InventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inventoryItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}
	if !isOutfitItemName(inventoryItem.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item is not an outfit"})
		return
	}

	generation, err := s.startOutfitGeneration(
		ctx,
		user.ID,
		ownedInventoryItemID,
		ownedInventoryItem.InventoryItemID,
		inventoryItem.Name,
		requestBody.SelfieUrl,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, generation)
}

func (s *server) getOutfitGeneration(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stringOwnedInventoryItemID := ctx.Param("ownedInventoryItemID")
	if stringOwnedInventoryItemID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "owned inventory item ID is required"})
		return
	}

	ownedInventoryItemID, err := uuid.Parse(stringOwnedInventoryItemID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid owned inventory item ID"})
		return
	}

	gen, err := s.dbClient.OutfitProfileGeneration().FindByOwnedInventoryItemID(ctx, ownedInventoryItemID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "outfit generation not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if gen.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	ctx.JSON(http.StatusOK, gen)
}

func isOutfitItemName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	b.Grow(len(normalized))
	for _, r := range normalized {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return strings.Contains(b.String(), "outfit")
}

func (s *server) startOutfitGeneration(
	ctx context.Context,
	userID uuid.UUID,
	ownedInventoryItemID uuid.UUID,
	inventoryItemID int,
	outfitName string,
	selfieUrl string,
) (*models.OutfitProfileGeneration, error) {
	var existing *models.OutfitProfileGeneration
	existing, err := s.dbClient.OutfitProfileGeneration().FindByOwnedInventoryItemID(ctx, ownedInventoryItemID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var generation *models.OutfitProfileGeneration
	if existing != nil && (existing.Status == models.OutfitGenerationStatusQueued || existing.Status == models.OutfitGenerationStatusInProgress) {
		return existing, nil
	}

	if existing != nil && existing.Status == models.OutfitGenerationStatusFailed {
		clearErr := ""
		update := &models.OutfitProfileGeneration{
			SelfieUrl:         selfieUrl,
			Status:            models.OutfitGenerationStatusQueued,
			ErrorMessage:      &clearErr,
			ProfilePictureUrl: nil,
		}
		if err := s.dbClient.OutfitProfileGeneration().Update(ctx, existing.ID, update); err != nil {
			return nil, err
		}
		updated, err := s.dbClient.OutfitProfileGeneration().FindByID(ctx, existing.ID)
		if err != nil {
			return nil, err
		}
		generation = updated
	} else {
		generation = &models.OutfitProfileGeneration{
			ID:                   uuid.New(),
			UserID:               userID,
			OwnedInventoryItemID: ownedInventoryItemID,
			InventoryItemID:      inventoryItemID,
			OutfitName:           outfitName,
			SelfieUrl:            selfieUrl,
			Status:               models.OutfitGenerationStatusQueued,
		}
		if err := s.dbClient.OutfitProfileGeneration().Create(ctx, generation); err != nil {
			return nil, err
		}
	}

	payload := jobs.GenerateOutfitProfilePictureTaskPayload{
		GenerationID:         generation.ID,
		UserID:               userID,
		OwnedInventoryItemID: ownedInventoryItemID,
		SelfieUrl:            selfieUrl,
		OutfitName:           outfitName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateOutfitProfilePictureTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		update := &models.OutfitProfileGeneration{
			Status:       models.OutfitGenerationStatusFailed,
			ErrorMessage: &errMsg,
		}
		_ = s.dbClient.OutfitProfileGeneration().Update(ctx, generation.ID, update)
		return nil, err
	}

	return generation, nil
}

func (s *server) adminUseOutfitItem(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		UserID    uuid.UUID `json:"userID" binding:"required"`
		ItemID    int       `json:"itemID" binding:"required"`
		SelfieUrl string    `json:"selfieUrl" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(requestBody.SelfieUrl) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "selfieUrl is required"})
		return
	}

	inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, requestBody.ItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inventoryItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}
	if !isOutfitItemName(inventoryItem.Name) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item is not an outfit"})
		return
	}

	items, err := s.dbClient.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{UserID: &requestBody.UserID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var owned *models.OwnedInventoryItem
	for i := range items {
		item := items[i]
		if item.InventoryItemID == requestBody.ItemID && item.Quantity > 0 {
			owned = &item
			break
		}
	}
	if owned == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user does not own this item"})
		return
	}

	generation, err := s.startOutfitGeneration(
		ctx,
		requestBody.UserID,
		owned.ID,
		owned.InventoryItemID,
		inventoryItem.Name,
		requestBody.SelfieUrl,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, generation)
}

func (s *server) getTeamsInventory(ctx *gin.Context) {
	stringTeamID := ctx.Param("teamID")
	if stringTeamID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "team ID is required",
		})
		return
	}

	teamID, err := uuid.Parse(stringTeamID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid team ID",
		})
		return
	}

	inventory, err := s.dbClient.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{TeamID: &teamID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, inventory)
}

func (s *server) getInventoryItems(ctx *gin.Context) {
	items := s.quartermaster.GetInventoryItems()
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getAllInventoryItems(ctx *gin.Context) {
	items, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getInventoryItem(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (s *server) createInventoryItem(ctx *gin.Context) {
	var requestBody struct {
		Name                    string                         `json:"name" binding:"required"`
		ImageURL                string                         `json:"imageUrl"`
		FlavorText              string                         `json:"flavorText"`
		EffectText              string                         `json:"effectText"`
		RarityTier              string                         `json:"rarityTier" binding:"required"`
		IsCaptureType           bool                           `json:"isCaptureType"`
		UnlockTier              *int                           `json:"unlockTier"`
		EquipSlot               *string                        `json:"equipSlot"`
		StrengthMod             int                            `json:"strengthMod"`
		DexterityMod            int                            `json:"dexterityMod"`
		ConstitutionMod         int                            `json:"constitutionMod"`
		IntelligenceMod         int                            `json:"intelligenceMod"`
		WisdomMod               int                            `json:"wisdomMod"`
		CharismaMod             int                            `json:"charismaMod"`
		HandItemCategory        *string                        `json:"handItemCategory"`
		Handedness              *string                        `json:"handedness"`
		DamageMin               *int                           `json:"damageMin"`
		DamageMax               *int                           `json:"damageMax"`
		DamageAffinity          *string                        `json:"damageAffinity"`
		SwipesPerAttack         *int                           `json:"swipesPerAttack"`
		BlockPercentage         *int                           `json:"blockPercentage"`
		DamageBlocked           *int                           `json:"damageBlocked"`
		SpellDamageBonusPercent *int                           `json:"spellDamageBonusPercent"`
		ConsumeHealthDelta      int                            `json:"consumeHealthDelta"`
		ConsumeManaDelta        int                            `json:"consumeManaDelta"`
		ConsumeStatusesToAdd    []scenarioFailureStatusPayload `json:"consumeStatusesToAdd"`
		ConsumeStatusesToRemove []string                       `json:"consumeStatusesToRemove"`
		ConsumeSpellIDs         []string                       `json:"consumeSpellIds"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var equipSlot *string
	if requestBody.EquipSlot != nil {
		trimmed := strings.TrimSpace(*requestBody.EquipSlot)
		if trimmed != "" {
			if !models.IsValidInventoryEquipSlot(trimmed) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid equip slot"})
				return
			}
			equipSlot = &trimmed
		}
	}
	handAttrs, err := models.NormalizeAndValidateHandEquipment(equipSlot, models.HandEquipmentAttributes{
		HandItemCategory:        requestBody.HandItemCategory,
		Handedness:              requestBody.Handedness,
		DamageMin:               requestBody.DamageMin,
		DamageMax:               requestBody.DamageMax,
		DamageAffinity:          requestBody.DamageAffinity,
		SwipesPerAttack:         requestBody.SwipesPerAttack,
		BlockPercentage:         requestBody.BlockPercentage,
		DamageBlocked:           requestBody.DamageBlocked,
		SpellDamageBonusPercent: requestBody.SpellDamageBonusPercent,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	consumeStatusesToAdd, err := parseScenarioFailureStatusTemplates(requestBody.ConsumeStatusesToAdd, "consumeStatusesToAdd")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	consumeStatusesToRemove := parseInventoryConsumeStatusNames(requestBody.ConsumeStatusesToRemove)
	consumeSpellIDs, err := parseInventoryConsumeSpellIDs(requestBody.ConsumeSpellIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for idx, rawSpellID := range consumeSpellIDs {
		spellID, _ := uuid.Parse(rawSpellID)
		if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("consumeSpellIds[%d] not found", idx)})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	item := &models.InventoryItem{
		Name:                    requestBody.Name,
		ImageURL:                requestBody.ImageURL,
		FlavorText:              requestBody.FlavorText,
		EffectText:              requestBody.EffectText,
		RarityTier:              requestBody.RarityTier,
		IsCaptureType:           requestBody.IsCaptureType,
		UnlockTier:              requestBody.UnlockTier,
		EquipSlot:               equipSlot,
		StrengthMod:             requestBody.StrengthMod,
		DexterityMod:            requestBody.DexterityMod,
		ConstitutionMod:         requestBody.ConstitutionMod,
		IntelligenceMod:         requestBody.IntelligenceMod,
		WisdomMod:               requestBody.WisdomMod,
		CharismaMod:             requestBody.CharismaMod,
		HandItemCategory:        handAttrs.HandItemCategory,
		Handedness:              handAttrs.Handedness,
		DamageMin:               handAttrs.DamageMin,
		DamageMax:               handAttrs.DamageMax,
		DamageAffinity:          handAttrs.DamageAffinity,
		SwipesPerAttack:         handAttrs.SwipesPerAttack,
		BlockPercentage:         handAttrs.BlockPercentage,
		DamageBlocked:           handAttrs.DamageBlocked,
		SpellDamageBonusPercent: handAttrs.SpellDamageBonusPercent,
		ConsumeHealthDelta:      requestBody.ConsumeHealthDelta,
		ConsumeManaDelta:        requestBody.ConsumeManaDelta,
		ConsumeStatusesToAdd:    consumeStatusesToAdd,
		ConsumeStatusesToRemove: consumeStatusesToRemove,
		ConsumeSpellIDs:         consumeSpellIDs,
		ImageGenerationStatus: func() string {
			if requestBody.ImageURL != "" {
				return models.InventoryImageGenerationStatusComplete
			}
			return models.InventoryImageGenerationStatusNone
		}(),
	}

	if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create inventory item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

func (s *server) generateInventoryItem(ctx *gin.Context) {
	var requestBody struct {
		Name             string  `json:"name" binding:"required"`
		Description      string  `json:"description"`
		RarityTier       string  `json:"rarityTier" binding:"required"`
		EquipSlot        *string `json:"equipSlot"`
		HandItemCategory *string `json:"handItemCategory"`
		Handedness       *string `json:"handedness"`
		DamageAffinity   *string `json:"damageAffinity"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var equipSlot *string
	if requestBody.EquipSlot != nil {
		trimmed := strings.TrimSpace(*requestBody.EquipSlot)
		if trimmed != "" {
			if !models.IsValidInventoryEquipSlot(trimmed) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid equip slot"})
				return
			}
			equipSlot = &trimmed
		}
	}

	handAttrs := models.HandEquipmentAttributes{
		HandItemCategory: requestBody.HandItemCategory,
		Handedness:       requestBody.Handedness,
		DamageAffinity:   requestBody.DamageAffinity,
	}
	if equipSlot != nil && models.IsHandEquipSlot(*equipSlot) {
		if handAttrs.HandItemCategory == nil || strings.TrimSpace(*handAttrs.HandItemCategory) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "handItemCategory is required for hand equipment generation"})
			return
		}
		if handAttrs.Handedness == nil || strings.TrimSpace(*handAttrs.Handedness) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "handedness is required for hand equipment generation"})
			return
		}
		handAttrs = generateInventoryItemHandAttributes(requestBody.RarityTier, *handAttrs.HandItemCategory, *handAttrs.Handedness)
	}
	validatedHandAttrs, err := models.NormalizeAndValidateHandEquipment(equipSlot, handAttrs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &models.InventoryItem{
		Name:                    requestBody.Name,
		FlavorText:              requestBody.Description,
		RarityTier:              requestBody.RarityTier,
		IsCaptureType:           false,
		EquipSlot:               equipSlot,
		HandItemCategory:        validatedHandAttrs.HandItemCategory,
		Handedness:              validatedHandAttrs.Handedness,
		DamageMin:               validatedHandAttrs.DamageMin,
		DamageMax:               validatedHandAttrs.DamageMax,
		DamageAffinity:          validatedHandAttrs.DamageAffinity,
		SwipesPerAttack:         validatedHandAttrs.SwipesPerAttack,
		BlockPercentage:         validatedHandAttrs.BlockPercentage,
		DamageBlocked:           validatedHandAttrs.DamageBlocked,
		SpellDamageBonusPercent: validatedHandAttrs.SpellDamageBonusPercent,
		ConsumeStatusesToAdd:    models.ScenarioFailureStatusTemplates{},
		ConsumeStatusesToRemove: models.StringArray{},
		ConsumeSpellIDs:         models.StringArray{},
		ImageGenerationStatus:   models.InventoryImageGenerationStatusQueued,
	}

	if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create inventory item: " + err.Error()})
		return
	}

	payload := jobs.GenerateInventoryItemImageTaskPayload{
		InventoryItemID: item.ID,
		Name:            requestBody.Name,
		Description:     requestBody.Description,
		RarityTier:      requestBody.RarityTier,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateInventoryItemImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = s.dbClient.InventoryItem().UpdateInventoryItem(ctx, item.ID, map[string]interface{}{
			"image_generation_status": models.InventoryImageGenerationStatusFailed,
			"image_generation_error":  errMsg,
		})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

func (s *server) generateInventoryItemSet(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	sourceItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sourceItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}
	if sourceItem.EquipSlot == nil || strings.TrimSpace(*sourceItem.EquipSlot) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "source item must be equippable"})
		return
	}

	sourceSlot := normalizeInventorySetSlot(*sourceItem.EquipSlot)
	if !models.IsValidInventoryEquipSlot(sourceSlot) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "source item has invalid equip slot"})
		return
	}

	setTheme := inventorySetThemeFromName(sourceItem.Name)
	profile := deriveInventorySetProfile(sourceItem)
	targetSlots := inventorySetTargetSlots(sourceSlot)

	existingItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	existingKeys := make(map[string]struct{}, len(existingItems))
	for _, existing := range existingItems {
		if existing.EquipSlot == nil || strings.TrimSpace(*existing.EquipSlot) == "" {
			continue
		}
		existingKeys[inventorySetItemKey(normalizeInventorySetSlot(*existing.EquipSlot), existing.Name)] = struct{}{}
	}

	createdItems := make([]models.InventoryItem, 0, len(targetSlots))
	skippedSlots := make([]string, 0)
	enqueueWarnings := make([]string, 0)
	for _, slot := range targetSlots {
		name, handCategory := inventorySetItemName(sourceItem, setTheme, slot, profile)
		itemKey := inventorySetItemKey(slot, name)
		if _, exists := existingKeys[itemKey]; exists {
			skippedSlots = append(skippedSlots, slot)
			continue
		}

		strengthMod, dexterityMod, constitutionMod, intelligenceMod, wisdomMod, charismaMod := inventorySetScaledStats(sourceItem, slot, profile)

		item := &models.InventoryItem{
			Name:                    name,
			FlavorText:              inventorySetFlavorText(setTheme, slot, handCategory),
			EffectText:              "",
			RarityTier:              sourceItem.RarityTier,
			IsCaptureType:           false,
			SellValue:               cloneIntPtr(sourceItem.SellValue),
			UnlockTier:              cloneIntPtr(sourceItem.UnlockTier),
			EquipSlot:               stringPtr(slot),
			StrengthMod:             strengthMod,
			DexterityMod:            dexterityMod,
			ConstitutionMod:         constitutionMod,
			IntelligenceMod:         intelligenceMod,
			WisdomMod:               wisdomMod,
			CharismaMod:             charismaMod,
			ConsumeStatusesToAdd:    models.ScenarioFailureStatusTemplates{},
			ConsumeStatusesToRemove: models.StringArray{},
			ConsumeSpellIDs:         models.StringArray{},
			ImageGenerationStatus:   models.InventoryImageGenerationStatusQueued,
		}

		if models.IsHandEquipSlot(slot) {
			attrs, handErr := inventorySetHandAttributesForSlot(sourceItem, slot, profile)
			if handErr != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": handErr.Error()})
				return
			}
			item.HandItemCategory = attrs.HandItemCategory
			item.Handedness = attrs.Handedness
			item.DamageMin = attrs.DamageMin
			item.DamageMax = attrs.DamageMax
			item.DamageAffinity = attrs.DamageAffinity
			item.SwipesPerAttack = attrs.SwipesPerAttack
			item.BlockPercentage = attrs.BlockPercentage
			item.DamageBlocked = attrs.DamageBlocked
			item.SpellDamageBonusPercent = attrs.SpellDamageBonusPercent
		}

		resolvedHandCategory := handCategory
		if item.HandItemCategory != nil {
			resolvedHandCategory = *item.HandItemCategory
		}
		item.EffectText = inventorySetEffectText(item, resolvedHandCategory)

		if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create set item: " + err.Error()})
			return
		}

		if err := s.enqueueInventoryItemImageGeneration(ctx, item.ID, item.Name, item.FlavorText, item.RarityTier); err != nil {
			enqueueWarnings = append(enqueueWarnings, fmt.Sprintf("%s: %s", item.Name, err.Error()))
		}

		createdItems = append(createdItems, *item)
		existingKeys[itemKey] = struct{}{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"sourceItemId":    sourceItem.ID,
		"setTheme":        setTheme,
		"createdItems":    createdItems,
		"skippedSlots":    skippedSlots,
		"enqueueWarnings": enqueueWarnings,
		"message":         fmt.Sprintf("created %d set item(s)", len(createdItems)),
	})
}

func normalizeInventorySetStatKey(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strength":
		return "strength"
	case "dexterity":
		return "dexterity"
	case "constitution":
		return "constitution"
	case "intelligence":
		return "intelligence"
	case "wisdom":
		return "wisdom"
	case "charisma":
		return "charisma"
	default:
		return ""
	}
}

func inventorySetStatDisplayName(key string) string {
	switch normalizeInventorySetStatKey(key) {
	case "strength":
		return "Strength"
	case "dexterity":
		return "Dexterity"
	case "constitution":
		return "Constitution"
	case "intelligence":
		return "Intelligence"
	case "wisdom":
		return "Wisdom"
	case "charisma":
		return "Charisma"
	default:
		return "Unknown"
	}
}

func inventorySetThemeFromStats(majorStat string, minorStat string, targetLevel int) string {
	rank := "Wayfarer"
	switch {
	case targetLevel >= 80:
		rank = "Ascendant"
	case targetLevel >= 55:
		rank = "Paragon"
	case targetLevel >= 30:
		rank = "Vanguard"
	}

	majorTitle := "Wayfarer"
	switch normalizeInventorySetStatKey(majorStat) {
	case "strength":
		majorTitle = "Warbrand"
	case "dexterity":
		majorTitle = "Nightstep"
	case "constitution":
		majorTitle = "Stoneward"
	case "intelligence":
		majorTitle = "Spellweaver"
	case "wisdom":
		majorTitle = "Dawnseer"
	case "charisma":
		majorTitle = "Crownsworn"
	}

	minorAspect := ""
	switch normalizeInventorySetStatKey(minorStat) {
	case "strength":
		minorAspect = "Ember"
	case "dexterity":
		minorAspect = "Shade"
	case "constitution":
		minorAspect = "Bulwark"
	case "intelligence":
		minorAspect = "Rune"
	case "wisdom":
		minorAspect = "Grove"
	case "charisma":
		minorAspect = "Crown"
	}

	if minorAspect == "" || normalizeInventorySetStatKey(majorStat) == normalizeInventorySetStatKey(minorStat) {
		return fmt.Sprintf("%s %s", rank, majorTitle)
	}
	return fmt.Sprintf("%s %s of the %s", rank, majorTitle, minorAspect)
}

func inventorySetRarityForTargetLevel(targetLevel int) string {
	switch {
	case targetLevel >= 80:
		return "Mythic"
	case targetLevel >= 55:
		return "Epic"
	case targetLevel >= 30:
		return "Uncommon"
	default:
		return "Common"
	}
}

func normalizeInventorySetRarityTier(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "common":
		return "Common"
	case "uncommon":
		return "Uncommon"
	case "epic":
		return "Epic"
	case "mythic":
		return "Mythic"
	default:
		return ""
	}
}

func inventorySetPrimaryStatPointsForTargetLevel(targetLevel int, rarity string) int {
	level := targetLevel
	if level < 1 {
		level = 1
	}
	if level > 100 {
		level = 100
	}

	basePoints := int(math.Round(2 + (float64(level) * 0.17)))
	rarityRank := 1
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "uncommon":
		rarityRank = 2
	case "epic":
		rarityRank = 3
	case "mythic":
		rarityRank = 4
	}

	rarityBonus := int(math.Round(float64(rarityRank-1) * (0.5 + (float64(level) * 0.015))))
	points := basePoints + maxInt(0, rarityBonus)

	levelCap := int(math.Round(3 + (float64(level) * 0.27)))
	if points > levelCap {
		points = levelCap
	}
	return maxInt(2, points)
}

func setInventoryStatValue(item *models.InventoryItem, stat string, value int) {
	if item == nil {
		return
	}
	switch normalizeInventorySetStatKey(stat) {
	case "strength":
		item.StrengthMod = value
	case "dexterity":
		item.DexterityMod = value
	case "constitution":
		item.ConstitutionMod = value
	case "intelligence":
		item.IntelligenceMod = value
	case "wisdom":
		item.WisdomMod = value
	case "charisma":
		item.CharismaMod = value
	}
}

func inventorySetAllEquippableSlots() []string {
	return []string{
		string(models.EquipmentSlotHat),
		string(models.EquipmentSlotNecklace),
		string(models.EquipmentSlotChest),
		string(models.EquipmentSlotLegs),
		string(models.EquipmentSlotShoes),
		string(models.EquipmentSlotGloves),
		string(models.EquipmentSlotDominantHand),
		string(models.EquipmentSlotOffHand),
		string(models.EquipmentSlotRing),
	}
}

func (s *server) generateEquippableInventorySet(ctx *gin.Context) {
	var requestBody struct {
		TargetLevel int    `json:"targetLevel" binding:"required"`
		MajorStat   string `json:"majorStat" binding:"required"`
		MinorStat   string `json:"minorStat" binding:"required"`
		RarityTier  string `json:"rarityTier"`
		SetTheme    string `json:"setTheme"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.TargetLevel < 1 || requestBody.TargetLevel > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetLevel must be between 1 and 100"})
		return
	}

	majorStat := normalizeInventorySetStatKey(requestBody.MajorStat)
	minorStat := normalizeInventorySetStatKey(requestBody.MinorStat)
	if majorStat == "" || minorStat == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "majorStat and minorStat must be one of: strength, dexterity, constitution, intelligence, wisdom, charisma",
		})
		return
	}
	if majorStat == minorStat {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "majorStat and minorStat must be different"})
		return
	}

	rarity := normalizeInventorySetRarityTier(requestBody.RarityTier)
	if strings.TrimSpace(requestBody.RarityTier) != "" && rarity == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "rarityTier must be one of: Common, Uncommon, Epic, Mythic",
		})
		return
	}
	if rarity == "" {
		rarity = inventorySetRarityForTargetLevel(requestBody.TargetLevel)
	}

	setTheme := strings.TrimSpace(requestBody.SetTheme)
	if setTheme == "" {
		setTheme = inventorySetThemeFromStats(majorStat, minorStat, requestBody.TargetLevel)
	}
	majorPoints := inventorySetPrimaryStatPointsForTargetLevel(requestBody.TargetLevel, rarity)
	minorPoints := maxInt(1, int(math.Round(float64(majorPoints)*0.6)))

	sourceItem := &models.InventoryItem{
		Name:                    fmt.Sprintf("%s Core", setTheme),
		RarityTier:              rarity,
		IsCaptureType:           false,
		UnlockTier:              intPtr(requestBody.TargetLevel),
		EquipSlot:               stringPtr(string(models.EquipmentSlotChest)),
		ConsumeStatusesToAdd:    models.ScenarioFailureStatusTemplates{},
		ConsumeStatusesToRemove: models.StringArray{},
		ConsumeSpellIDs:         models.StringArray{},
	}
	setInventoryStatValue(sourceItem, majorStat, majorPoints)
	setInventoryStatValue(sourceItem, minorStat, minorPoints)

	profile := deriveInventorySetProfile(sourceItem)
	targetSlots := inventorySetAllEquippableSlots()

	existingItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	existingKeys := make(map[string]struct{}, len(existingItems))
	for _, existing := range existingItems {
		if existing.EquipSlot == nil || strings.TrimSpace(*existing.EquipSlot) == "" {
			continue
		}
		existingKeys[inventorySetItemKey(normalizeInventorySetSlot(*existing.EquipSlot), existing.Name)] = struct{}{}
	}

	createdItems := make([]models.InventoryItem, 0, len(targetSlots))
	skippedSlots := make([]string, 0)
	enqueueWarnings := make([]string, 0)
	for _, slot := range targetSlots {
		name, handCategory := inventorySetItemName(sourceItem, setTheme, slot, profile)
		itemKey := inventorySetItemKey(slot, name)
		if _, exists := existingKeys[itemKey]; exists {
			skippedSlots = append(skippedSlots, slot)
			continue
		}

		strengthMod, dexterityMod, constitutionMod, intelligenceMod, wisdomMod, charismaMod := inventorySetScaledStats(sourceItem, slot, profile)

		item := &models.InventoryItem{
			Name:                    name,
			FlavorText:              inventorySetFlavorText(setTheme, slot, handCategory),
			EffectText:              "",
			RarityTier:              rarity,
			IsCaptureType:           false,
			UnlockTier:              intPtr(requestBody.TargetLevel),
			EquipSlot:               stringPtr(slot),
			StrengthMod:             strengthMod,
			DexterityMod:            dexterityMod,
			ConstitutionMod:         constitutionMod,
			IntelligenceMod:         intelligenceMod,
			WisdomMod:               wisdomMod,
			CharismaMod:             charismaMod,
			ConsumeStatusesToAdd:    models.ScenarioFailureStatusTemplates{},
			ConsumeStatusesToRemove: models.StringArray{},
			ConsumeSpellIDs:         models.StringArray{},
			ImageGenerationStatus:   models.InventoryImageGenerationStatusQueued,
		}

		if models.IsHandEquipSlot(slot) {
			attrs, handErr := inventorySetHandAttributesForSlot(sourceItem, slot, profile)
			if handErr != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": handErr.Error()})
				return
			}
			item.HandItemCategory = attrs.HandItemCategory
			item.Handedness = attrs.Handedness
			item.DamageMin = attrs.DamageMin
			item.DamageMax = attrs.DamageMax
			item.DamageAffinity = attrs.DamageAffinity
			item.SwipesPerAttack = attrs.SwipesPerAttack
			item.BlockPercentage = attrs.BlockPercentage
			item.DamageBlocked = attrs.DamageBlocked
			item.SpellDamageBonusPercent = attrs.SpellDamageBonusPercent
		}

		resolvedHandCategory := handCategory
		if item.HandItemCategory != nil {
			resolvedHandCategory = *item.HandItemCategory
		}
		item.EffectText = inventorySetEffectText(item, resolvedHandCategory)

		if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create set item: " + err.Error()})
			return
		}

		if err := s.enqueueInventoryItemImageGeneration(ctx, item.ID, item.Name, item.FlavorText, item.RarityTier); err != nil {
			enqueueWarnings = append(enqueueWarnings, fmt.Sprintf("%s: %s", item.Name, err.Error()))
		}

		createdItems = append(createdItems, *item)
		existingKeys[itemKey] = struct{}{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"setTheme":        setTheme,
		"targetLevel":     requestBody.TargetLevel,
		"majorStat":       majorStat,
		"minorStat":       minorStat,
		"rarityTier":      rarity,
		"createdItems":    createdItems,
		"skippedSlots":    skippedSlots,
		"enqueueWarnings": enqueueWarnings,
		"message":         fmt.Sprintf("created %d equippable set item(s)", len(createdItems)),
	})
}

type consumableQualityConfig struct {
	Label              string
	LevelMin           int
	LevelMax           int
	PowerMultiplier    float64
	DurationMultiplier float64
	DefaultRarity      string
}

var consumableQualityConfigs = []consumableQualityConfig{
	{Label: "Minor", LevelMin: 1, LevelMax: 15, PowerMultiplier: 1.0, DurationMultiplier: 1.0, DefaultRarity: "Common"},
	{Label: "Lesser", LevelMin: 16, LevelMax: 30, PowerMultiplier: 1.8, DurationMultiplier: 1.4, DefaultRarity: "Common"},
	{Label: "Greater", LevelMin: 31, LevelMax: 50, PowerMultiplier: 3.0, DurationMultiplier: 2.0, DefaultRarity: "Uncommon"},
	{Label: "Major", LevelMin: 51, LevelMax: 70, PowerMultiplier: 5.0, DurationMultiplier: 3.0, DefaultRarity: "Epic"},
	{Label: "Superior", LevelMin: 71, LevelMax: 85, PowerMultiplier: 7.5, DurationMultiplier: 4.2, DefaultRarity: "Epic"},
	{Label: "Superb", LevelMin: 86, LevelMax: 100, PowerMultiplier: 10.5, DurationMultiplier: 5.8, DefaultRarity: "Mythic"},
}

func (s *server) generateConsumableQualities(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	sourceItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sourceItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}
	if sourceItem.EquipSlot != nil && strings.TrimSpace(*sourceItem.EquipSlot) != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "source item must be a non-equippable consumable"})
		return
	}
	if !inventoryItemHasConsumableEffects(sourceItem) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "source item has no consumable effects to scale"})
		return
	}

	sourceQualityLabel, baseName, hasQuality := parseConsumableQualityAndBaseName(sourceItem.Name)
	if !hasQuality {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "source item name must include a quality prefix like 'Minor'",
		})
		return
	}
	if strings.ToLower(sourceQualityLabel) != "minor" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "source consumable must start with 'Minor' to generate full quality progression",
		})
		return
	}

	sourceQuality, ok := findConsumableQualityConfig(sourceQualityLabel)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported source consumable quality"})
		return
	}

	existingItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	existingNames := make(map[string]struct{}, len(existingItems))
	for _, existing := range existingItems {
		existingNames[strings.ToLower(strings.TrimSpace(existing.Name))] = struct{}{}
	}

	createdItems := make([]models.InventoryItem, 0, len(consumableQualityConfigs))
	skippedQualities := make([]string, 0)
	enqueueWarnings := make([]string, 0)

	for _, quality := range consumableQualityConfigs {
		itemName := fmt.Sprintf("%s %s", quality.Label, baseName)
		normalizedName := strings.ToLower(strings.TrimSpace(itemName))
		if _, exists := existingNames[normalizedName]; exists {
			skippedQualities = append(skippedQualities, quality.Label)
			continue
		}

		powerScale := quality.PowerMultiplier / sourceQuality.PowerMultiplier
		durationScale := quality.DurationMultiplier / sourceQuality.DurationMultiplier

		statusesToAdd := scaleConsumableStatuses(sourceItem.ConsumeStatusesToAdd, powerScale, durationScale)
		consumeHealthDelta := scaleConsumableValue(sourceItem.ConsumeHealthDelta, powerScale)
		consumeManaDelta := scaleConsumableValue(sourceItem.ConsumeManaDelta, powerScale)

		item := &models.InventoryItem{
			Name:                    itemName,
			ImageURL:                "",
			FlavorText:              buildConsumableQualityFlavorText(baseName, quality),
			EffectText:              buildConsumableQualityEffectText(baseName, quality, consumeHealthDelta, consumeManaDelta, statusesToAdd, sourceItem.ConsumeStatusesToRemove),
			RarityTier:              selectConsumableQualityRarity(sourceItem.RarityTier, quality.DefaultRarity),
			IsCaptureType:           false,
			SellValue:               scaleConsumableOptionalInt(sourceItem.SellValue, powerScale),
			UnlockTier:              scaleConsumableOptionalInt(sourceItem.UnlockTier, math.Max(1.0, powerScale*0.5)),
			EquipSlot:               nil,
			StrengthMod:             0,
			DexterityMod:            0,
			ConstitutionMod:         0,
			IntelligenceMod:         0,
			WisdomMod:               0,
			CharismaMod:             0,
			HandItemCategory:        nil,
			Handedness:              nil,
			DamageMin:               nil,
			DamageMax:               nil,
			SwipesPerAttack:         nil,
			BlockPercentage:         nil,
			DamageBlocked:           nil,
			SpellDamageBonusPercent: nil,
			ConsumeHealthDelta:      consumeHealthDelta,
			ConsumeManaDelta:        consumeManaDelta,
			ConsumeStatusesToAdd:    statusesToAdd,
			ConsumeStatusesToRemove: sourceItem.ConsumeStatusesToRemove,
			ConsumeSpellIDs:         sourceItem.ConsumeSpellIDs,
			ImageGenerationStatus:   models.InventoryImageGenerationStatusQueued,
		}

		if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create consumable quality item: " + err.Error()})
			return
		}

		if err := s.enqueueInventoryItemImageGeneration(ctx, item.ID, item.Name, item.FlavorText, item.RarityTier); err != nil {
			enqueueWarnings = append(enqueueWarnings, fmt.Sprintf("%s: %s", item.Name, err.Error()))
		}

		createdItems = append(createdItems, *item)
		existingNames[normalizedName] = struct{}{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"sourceItemId":     sourceItem.ID,
		"baseName":         baseName,
		"createdItems":     createdItems,
		"skippedQualities": skippedQualities,
		"enqueueWarnings":  enqueueWarnings,
		"message":          fmt.Sprintf("created %d consumable quality item(s)", len(createdItems)),
	})
}

func inventoryItemHasConsumableEffects(item *models.InventoryItem) bool {
	if item == nil {
		return false
	}
	if item.ConsumeHealthDelta != 0 || item.ConsumeManaDelta != 0 {
		return true
	}
	if len(item.ConsumeStatusesToAdd) > 0 || len(item.ConsumeStatusesToRemove) > 0 || len(item.ConsumeSpellIDs) > 0 {
		return true
	}
	return false
}

func parseConsumableQualityAndBaseName(name string) (string, string, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", "", false
	}

	for _, quality := range consumableQualityConfigs {
		prefix := strings.ToLower(quality.Label + " ")
		if strings.HasPrefix(strings.ToLower(trimmed), prefix) {
			baseName := strings.TrimSpace(trimmed[len(prefix):])
			if baseName == "" {
				return "", "", false
			}
			return quality.Label, baseName, true
		}
	}

	return "", "", false
}

func findConsumableQualityConfig(label string) (consumableQualityConfig, bool) {
	normalized := strings.ToLower(strings.TrimSpace(label))
	for _, quality := range consumableQualityConfigs {
		if strings.ToLower(quality.Label) == normalized {
			return quality, true
		}
	}
	return consumableQualityConfig{}, false
}

func scaleConsumableValue(value int, multiplier float64) int {
	if value == 0 {
		return 0
	}
	scaled := int(math.Round(float64(value) * multiplier))
	if scaled == 0 {
		if value > 0 {
			return 1
		}
		return -1
	}
	return scaled
}

func scaleConsumableOptionalInt(value *int, multiplier float64) *int {
	if value == nil {
		return nil
	}
	scaled := scaleConsumableValue(*value, multiplier)
	if scaled < 0 {
		scaled = 0
	}
	return &scaled
}

func scaleConsumableStatuses(
	statuses models.ScenarioFailureStatusTemplates,
	powerMultiplier float64,
	durationMultiplier float64,
) models.ScenarioFailureStatusTemplates {
	if len(statuses) == 0 {
		return models.ScenarioFailureStatusTemplates{}
	}

	scaled := make(models.ScenarioFailureStatusTemplates, 0, len(statuses))
	for _, status := range statuses {
		next := status
		if next.DurationSeconds > 0 {
			next.DurationSeconds = maxInt(1, int(math.Round(float64(next.DurationSeconds)*durationMultiplier)))
		}
		next.DamagePerTick = scaleConsumableValue(next.DamagePerTick, powerMultiplier)
		next.StrengthMod = scaleConsumableValue(next.StrengthMod, powerMultiplier)
		next.DexterityMod = scaleConsumableValue(next.DexterityMod, powerMultiplier)
		next.ConstitutionMod = scaleConsumableValue(next.ConstitutionMod, powerMultiplier)
		next.IntelligenceMod = scaleConsumableValue(next.IntelligenceMod, powerMultiplier)
		next.WisdomMod = scaleConsumableValue(next.WisdomMod, powerMultiplier)
		next.CharismaMod = scaleConsumableValue(next.CharismaMod, powerMultiplier)
		scaled = append(scaled, next)
	}
	return scaled
}

func buildConsumableQualityFlavorText(baseName string, quality consumableQualityConfig) string {
	return fmt.Sprintf(
		"A %s-grade %s, expertly refined for demanding journeys.",
		strings.ToLower(quality.Label),
		strings.ToLower(strings.TrimSpace(baseName)),
	)
}

func buildConsumableQualityEffectText(
	baseName string,
	quality consumableQualityConfig,
	healthDelta int,
	manaDelta int,
	statuses models.ScenarioFailureStatusTemplates,
	statusesToRemove []string,
) string {
	effects := make([]string, 0, 5)
	if healthDelta != 0 {
		effects = append(effects, fmt.Sprintf("Health %+d", healthDelta))
	}
	if manaDelta != 0 {
		effects = append(effects, fmt.Sprintf("Mana %+d", manaDelta))
	}
	if len(statuses) > 0 {
		effects = append(effects, fmt.Sprintf("Bestows %d scaled status effect(s)", len(statuses)))
	}
	if len(statusesToRemove) > 0 {
		effects = append(effects, fmt.Sprintf("Removes %d status effect(s)", len(statusesToRemove)))
	}
	if len(effects) == 0 {
		effects = append(effects, "General consumable utility")
	}

	return fmt.Sprintf(
		"%s quality. %s.",
		quality.Label,
		strings.Join(effects, ". "),
	)
}

func selectConsumableQualityRarity(sourceRarity string, targetDefault string) string {
	normalize := func(value string) string {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "common":
			return "Common"
		case "uncommon":
			return "Uncommon"
		case "epic":
			return "Epic"
		case "mythic":
			return "Mythic"
		case "not droppable":
			return "Not Droppable"
		default:
			return "Common"
		}
	}
	rank := map[string]int{
		"Common":        1,
		"Uncommon":      2,
		"Epic":          3,
		"Mythic":        4,
		"Not Droppable": 5,
	}

	normalizedSource := normalize(sourceRarity)
	normalizedTarget := normalize(targetDefault)
	if rank[normalizedSource] > rank[normalizedTarget] {
		return normalizedSource
	}
	return normalizedTarget
}

func enqueueInventoryItemImageTaskPayload(inventoryItemID int, name string, description string, rarityTier string) ([]byte, error) {
	payload := jobs.GenerateInventoryItemImageTaskPayload{
		InventoryItemID: inventoryItemID,
		Name:            name,
		Description:     description,
		RarityTier:      rarityTier,
	}
	return json.Marshal(payload)
}

func (s *server) enqueueInventoryItemImageGeneration(
	ctx context.Context,
	inventoryItemID int,
	name string,
	description string,
	rarityTier string,
) error {
	payloadBytes, err := enqueueInventoryItemImageTaskPayload(inventoryItemID, name, description, rarityTier)
	if err != nil {
		return err
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateInventoryItemImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = s.dbClient.InventoryItem().UpdateInventoryItem(ctx, inventoryItemID, map[string]interface{}{
			"image_generation_status": models.InventoryImageGenerationStatusFailed,
			"image_generation_error":  errMsg,
		})
		return err
	}
	return nil
}

func inventorySetTargetSlots(sourceSlot string) []string {
	allSlots := []string{
		string(models.EquipmentSlotHat),
		string(models.EquipmentSlotNecklace),
		string(models.EquipmentSlotChest),
		string(models.EquipmentSlotLegs),
		string(models.EquipmentSlotShoes),
		string(models.EquipmentSlotGloves),
		string(models.EquipmentSlotDominantHand),
		string(models.EquipmentSlotOffHand),
		string(models.EquipmentSlotRing),
	}
	targets := make([]string, 0, len(allSlots))
	for _, slot := range allSlots {
		if slot == sourceSlot {
			continue
		}
		targets = append(targets, slot)
	}
	return targets
}

func normalizeInventorySetSlot(slot string) string {
	normalized := strings.ToLower(strings.TrimSpace(slot))
	switch normalized {
	case string(models.EquipmentSlotRingLeft), string(models.EquipmentSlotRingRight):
		return string(models.EquipmentSlotRing)
	default:
		return normalized
	}
}

func inventorySetItemKey(slot string, name string) string {
	return strings.ToLower(strings.TrimSpace(slot)) + "|" + strings.ToLower(strings.TrimSpace(name))
}

func inventorySetThemeFromName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "Forged"
	}

	parts := strings.Fields(trimmed)
	if len(parts) <= 1 {
		return trimmed
	}

	descriptors := map[string]struct{}{
		"sword": {}, "blade": {}, "axe": {}, "mace": {}, "hammer": {}, "staff": {}, "wand": {}, "spear": {}, "dagger": {},
		"shield": {}, "orb": {}, "helm": {}, "hat": {}, "cowl": {}, "necklace": {}, "amulet": {}, "pendant": {},
		"chestplate": {}, "armor": {}, "plate": {}, "mail": {}, "tunic": {}, "vest": {}, "greaves": {}, "leggings": {},
		"boots": {}, "shoes": {}, "gauntlets": {}, "gloves": {}, "ring": {},
	}

	last := strings.ToLower(parts[len(parts)-1])
	if _, ok := descriptors[last]; ok {
		candidate := strings.TrimSpace(strings.Join(parts[:len(parts)-1], " "))
		if candidate != "" {
			return candidate
		}
	}

	candidate := strings.TrimSpace(strings.Join(parts[:len(parts)-1], " "))
	if candidate == "" {
		return trimmed
	}
	return candidate
}

func deriveInventorySetProfile(sourceItem *models.InventoryItem) string {
	if sourceItem == nil {
		return "martial"
	}

	if sourceItem.HandItemCategory != nil {
		switch strings.ToLower(strings.TrimSpace(*sourceItem.HandItemCategory)) {
		case string(models.HandItemCategoryStaff), string(models.HandItemCategoryOrb):
			return "caster"
		case string(models.HandItemCategoryShield):
			return "tank"
		case string(models.HandItemCategoryWeapon):
			return "martial"
		}
	}

	casterScore := sourceItem.IntelligenceMod + sourceItem.WisdomMod
	if sourceItem.SpellDamageBonusPercent != nil {
		casterScore += *sourceItem.SpellDamageBonusPercent / 8
	}

	martialScore := sourceItem.StrengthMod + sourceItem.DexterityMod
	if sourceItem.DamageMin != nil {
		martialScore += *sourceItem.DamageMin / 2
	}
	if sourceItem.DamageMax != nil {
		martialScore += *sourceItem.DamageMax / 3
	}

	tankScore := sourceItem.ConstitutionMod
	if sourceItem.BlockPercentage != nil {
		tankScore += *sourceItem.BlockPercentage / 12
	}
	if sourceItem.DamageBlocked != nil {
		tankScore += *sourceItem.DamageBlocked / 3
	}

	if casterScore >= martialScore && casterScore >= tankScore {
		return "caster"
	}
	if tankScore >= martialScore && tankScore >= casterScore {
		return "tank"
	}
	return "martial"
}

func inventorySetItemName(sourceItem *models.InventoryItem, theme string, slot string, profile string) (string, string) {
	switch slot {
	case string(models.EquipmentSlotHat):
		return fmt.Sprintf("%s Helm", theme), ""
	case string(models.EquipmentSlotNecklace):
		return fmt.Sprintf("%s Amulet", theme), ""
	case string(models.EquipmentSlotChest):
		return fmt.Sprintf("%s Chestplate", theme), ""
	case string(models.EquipmentSlotLegs):
		return fmt.Sprintf("%s Greaves", theme), ""
	case string(models.EquipmentSlotShoes):
		return fmt.Sprintf("%s Boots", theme), ""
	case string(models.EquipmentSlotGloves):
		return fmt.Sprintf("%s Gauntlets", theme), ""
	case string(models.EquipmentSlotRing):
		return fmt.Sprintf("%s Ring", theme), ""
	case string(models.EquipmentSlotDominantHand):
		dominantCategory := inventorySetDominantCategory(sourceItem, profile)
		if dominantCategory == string(models.HandItemCategoryStaff) {
			return fmt.Sprintf("%s Staff", theme), dominantCategory
		}
		weaponNoun := inventorySetDominantWeaponNoun(sourceItem)
		return fmt.Sprintf("%s %s", theme, weaponNoun), dominantCategory
	case string(models.EquipmentSlotOffHand):
		offHandCategory := inventorySetOffHandCategory(sourceItem, profile)
		if offHandCategory == string(models.HandItemCategoryOrb) {
			return fmt.Sprintf("%s Orb", theme), offHandCategory
		}
		return fmt.Sprintf("%s Shield", theme), offHandCategory
	default:
		return fmt.Sprintf("%s Gear", theme), ""
	}
}

func inventorySetDominantWeaponNoun(sourceItem *models.InventoryItem) string {
	if sourceItem == nil || strings.TrimSpace(sourceItem.Name) == "" {
		return "Blade"
	}
	if sourceItem.EquipSlot != nil &&
		normalizeInventorySetSlot(*sourceItem.EquipSlot) == string(models.EquipmentSlotDominantHand) &&
		sourceItem.HandItemCategory != nil &&
		strings.ToLower(strings.TrimSpace(*sourceItem.HandItemCategory)) == string(models.HandItemCategoryWeapon) {
		parts := strings.Fields(strings.TrimSpace(sourceItem.Name))
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
	}
	return "Blade"
}

func inventorySetDominantCategory(sourceItem *models.InventoryItem, profile string) string {
	if sourceItem != nil && sourceItem.HandItemCategory != nil {
		switch strings.ToLower(strings.TrimSpace(*sourceItem.HandItemCategory)) {
		case string(models.HandItemCategoryWeapon), string(models.HandItemCategoryStaff):
			return strings.ToLower(strings.TrimSpace(*sourceItem.HandItemCategory))
		case string(models.HandItemCategoryOrb):
			return string(models.HandItemCategoryStaff)
		case string(models.HandItemCategoryShield):
			return string(models.HandItemCategoryWeapon)
		}
	}
	if profile == "caster" {
		return string(models.HandItemCategoryStaff)
	}
	return string(models.HandItemCategoryWeapon)
}

func inventorySetOffHandCategory(sourceItem *models.InventoryItem, profile string) string {
	if sourceItem != nil && sourceItem.HandItemCategory != nil {
		switch strings.ToLower(strings.TrimSpace(*sourceItem.HandItemCategory)) {
		case string(models.HandItemCategoryShield), string(models.HandItemCategoryOrb):
			return strings.ToLower(strings.TrimSpace(*sourceItem.HandItemCategory))
		case string(models.HandItemCategoryStaff):
			return string(models.HandItemCategoryOrb)
		case string(models.HandItemCategoryWeapon):
			return string(models.HandItemCategoryShield)
		}
	}
	if profile == "caster" {
		return string(models.HandItemCategoryOrb)
	}
	return string(models.HandItemCategoryShield)
}

func inventorySetHandAttributesForSlot(sourceItem *models.InventoryItem, targetSlot string, profile string) (models.HandEquipmentAttributes, error) {
	slotPtr := stringPtr(targetSlot)
	if targetSlot == string(models.EquipmentSlotDominantHand) {
		category := inventorySetDominantCategory(sourceItem, profile)
		handedness := string(models.HandednessOneHanded)
		if category == string(models.HandItemCategoryStaff) {
			handedness = string(models.HandednessTwoHanded)
		} else if sourceItem != nil && sourceItem.Handedness != nil && strings.TrimSpace(*sourceItem.Handedness) != "" {
			handedness = strings.ToLower(strings.TrimSpace(*sourceItem.Handedness))
		}
		attrs := generateInventoryItemHandAttributes(sourceItem.RarityTier, category, handedness)
		return models.NormalizeAndValidateHandEquipment(slotPtr, attrs)
	}

	category := inventorySetOffHandCategory(sourceItem, profile)
	attrs := generateInventoryItemHandAttributes(sourceItem.RarityTier, category, string(models.HandednessOneHanded))
	return models.NormalizeAndValidateHandEquipment(slotPtr, attrs)
}

func inventorySetFlavorText(theme string, slot string, handCategory string) string {
	normalizedTheme := strings.TrimSpace(theme)
	if normalizedTheme == "" {
		normalizedTheme = "Nameless"
	}

	switch slot {
	case string(models.EquipmentSlotDominantHand):
		if handCategory == string(models.HandItemCategoryStaff) {
			return fmt.Sprintf("An old %s stave etched with spiraling sigils that answer to steady hands.", normalizedTheme)
		}
		return fmt.Sprintf("A battle-forged %s weapon tempered to sing the moment steel meets intent.", normalizedTheme)
	case string(models.EquipmentSlotOffHand):
		if handCategory == string(models.HandItemCategoryOrb) {
			return fmt.Sprintf("A polished %s orb that hums with restrained power between each breath.", normalizedTheme)
		}
		return fmt.Sprintf("A weighted %s shield built to turn chaos into rhythm.", normalizedTheme)
	case string(models.EquipmentSlotHat):
		return fmt.Sprintf("A crested %s helm worn by scouts who prefer foresight over luck.", normalizedTheme)
	case string(models.EquipmentSlotNecklace):
		return fmt.Sprintf("A %s amulet threaded with a quiet ward and a long memory.", normalizedTheme)
	case string(models.EquipmentSlotChest):
		return fmt.Sprintf("Layered %s plating shaped to carry impact without losing momentum.", normalizedTheme)
	case string(models.EquipmentSlotLegs):
		return fmt.Sprintf("%s greaves articulated for long marches and sudden reversals.", normalizedTheme)
	case string(models.EquipmentSlotShoes):
		return fmt.Sprintf("%s boots that favor silent footing and stubborn endurance.", normalizedTheme)
	case string(models.EquipmentSlotGloves):
		return fmt.Sprintf("%s gauntlets cut for precision grip when the fight gets messy.", normalizedTheme)
	case string(models.EquipmentSlotRing):
		return fmt.Sprintf("A %s ring whose inner rune flickers when resolve is tested.", normalizedTheme)
	default:
		return fmt.Sprintf("A %s relic carried by those who finish what they start.", normalizedTheme)
	}
}

func inventorySetEffectText(item *models.InventoryItem, handCategory string) string {
	slot := ""
	if item != nil && item.EquipSlot != nil {
		slot = normalizeInventorySetSlot(*item.EquipSlot)
	}

	parts := []string{inventorySetEffectLead(slot, strings.ToLower(strings.TrimSpace(handCategory)))}
	if statLean := inventorySetStatLeanText(item); statLean != "" {
		parts = append(parts, statLean)
	}
	return strings.Join(parts, " ")
}

func inventorySetEffectLead(slot string, handCategory string) string {
	switch slot {
	case string(models.EquipmentSlotDominantHand):
		if handCategory == string(models.HandItemCategoryStaff) {
			return "Arcane channels converge through its core, turning patient casting into sharper bursts."
		}
		return "Its edge favors committed timing, rewarding pressure and clean follow-through."
	case string(models.EquipmentSlotOffHand):
		if handCategory == string(models.HandItemCategoryOrb) {
			return "Ambient mana gathers at its surface, then releases in disciplined surges."
		}
		return "Its guard geometry catches incoming force and keeps your stance anchored."
	case string(models.EquipmentSlotHat):
		return "Keeps your head clear when the field turns noisy."
	case string(models.EquipmentSlotNecklace):
		return "A steady ward that smooths the rough edges of drawn-out encounters."
	case string(models.EquipmentSlotChest):
		return "Built to absorb punishment while preserving your forward tempo."
	case string(models.EquipmentSlotLegs):
		return "Encourages balanced footwork through pivots, rushes, and recoveries."
	case string(models.EquipmentSlotShoes):
		return "Shortens hesitation between decisions and movement."
	case string(models.EquipmentSlotGloves):
		return "Improves control at the moment intent becomes action."
	case string(models.EquipmentSlotRing):
		return "Subtle runes reinforce whatever fighting style you already trust."
	default:
		return "Its craftsmanship favors consistency over spectacle."
	}
}

func inventorySetStatLeanText(item *models.InventoryItem) string {
	if item == nil {
		return ""
	}

	type statLean struct {
		label string
		value int
	}

	leans := []statLean{
		{label: "strength", value: item.StrengthMod},
		{label: "dexterity", value: item.DexterityMod},
		{label: "constitution", value: item.ConstitutionMod},
		{label: "intelligence", value: item.IntelligenceMod},
		{label: "wisdom", value: item.WisdomMod},
		{label: "charisma", value: item.CharismaMod},
	}

	filtered := make([]statLean, 0, len(leans))
	for _, lean := range leans {
		if lean.value > 0 {
			filtered = append(filtered, lean)
		}
	}
	if len(filtered) == 0 {
		return ""
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].value == filtered[j].value {
			return filtered[i].label < filtered[j].label
		}
		return filtered[i].value > filtered[j].value
	})

	if len(filtered) == 1 {
		return fmt.Sprintf("Its enchantment leans into %s.", filtered[0].label)
	}

	return fmt.Sprintf(
		"Its enchantment leans into %s and %s.",
		filtered[0].label,
		filtered[1].label,
	)
}

func inventorySetSlotDisplayName(slot string) string {
	switch slot {
	case string(models.EquipmentSlotHat):
		return "the head"
	case string(models.EquipmentSlotNecklace):
		return "the neck"
	case string(models.EquipmentSlotChest):
		return "the chest"
	case string(models.EquipmentSlotLegs):
		return "the legs"
	case string(models.EquipmentSlotShoes):
		return "the feet"
	case string(models.EquipmentSlotGloves):
		return "the hands"
	case string(models.EquipmentSlotRing):
		return "the ring slot"
	default:
		return "its slot"
	}
}

func inventorySetScaledStats(sourceItem *models.InventoryItem, targetSlot string, profile string) (int, int, int, int, int, int) {
	sourceStats := []int{
		sourceItem.StrengthMod,
		sourceItem.DexterityMod,
		sourceItem.ConstitutionMod,
		sourceItem.IntelligenceMod,
		sourceItem.WisdomMod,
		sourceItem.CharismaMod,
	}
	sourceTotal := 0
	for _, value := range sourceStats {
		if value < 0 {
			sourceTotal += -value
		} else {
			sourceTotal += value
		}
	}

	scale := inventorySetSlotStatScale(targetSlot)
	if sourceTotal > 0 {
		scaled := make([]int, len(sourceStats))
		for idx, value := range sourceStats {
			scaled[idx] = int(math.Round(float64(value) * scale))
		}
		nonZero := false
		for _, value := range scaled {
			if value != 0 {
				nonZero = true
				break
			}
		}
		if !nonZero {
			largestIdx := 0
			largestAbs := 0
			for idx, value := range sourceStats {
				absValue := value
				if absValue < 0 {
					absValue = -absValue
				}
				if absValue > largestAbs {
					largestAbs = absValue
					largestIdx = idx
				}
			}
			if sourceStats[largestIdx] < 0 {
				scaled[largestIdx] = -1
			} else {
				scaled[largestIdx] = 1
			}
		}
		return scaled[0], scaled[1], scaled[2], scaled[3], scaled[4], scaled[5]
	}

	points := inventorySetDefaultStatPoints(sourceItem.RarityTier, targetSlot)
	if points <= 0 {
		return 0, 0, 0, 0, 0, 0
	}

	switch profile {
	case "caster":
		intelligence := maxInt(1, int(math.Ceil(float64(points)*0.55)))
		wisdom := maxInt(0, points-intelligence)
		return 0, 0, 0, intelligence, wisdom, 0
	case "tank":
		constitution := maxInt(1, int(math.Ceil(float64(points)*0.55)))
		strength := maxInt(0, points-constitution)
		return strength, 0, constitution, 0, 0, 0
	default:
		strength := maxInt(1, int(math.Ceil(float64(points)*0.55)))
		dexterity := maxInt(0, points-strength)
		return strength, dexterity, 0, 0, 0, 0
	}
}

func inventorySetSlotStatScale(slot string) float64 {
	switch slot {
	case string(models.EquipmentSlotChest):
		return 1.25
	case string(models.EquipmentSlotLegs):
		return 1.05
	case string(models.EquipmentSlotHat), string(models.EquipmentSlotGloves), string(models.EquipmentSlotShoes):
		return 0.85
	case string(models.EquipmentSlotNecklace):
		return 0.8
	case string(models.EquipmentSlotRing):
		return 0.72
	case string(models.EquipmentSlotDominantHand):
		return 0.95
	case string(models.EquipmentSlotOffHand):
		return 0.88
	default:
		return 1.0
	}
}

func inventorySetDefaultStatPoints(rarity string, slot string) int {
	base := 2
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		base = 2
	case "uncommon":
		base = 4
	case "epic":
		base = 7
	case "mythic":
		base = 10
	}
	return maxInt(1, int(math.Round(float64(base)*inventorySetSlotStatScale(slot))))
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func generateInventoryItemHandAttributes(rarity string, category string, handedness string) models.HandEquipmentAttributes {
	normalizedCategory := strings.ToLower(strings.TrimSpace(category))
	normalizedHandedness := strings.ToLower(strings.TrimSpace(handedness))
	attrs := models.HandEquipmentAttributes{
		HandItemCategory: stringPtr(normalizedCategory),
		Handedness:       stringPtr(normalizedHandedness),
	}

	switch normalizedCategory {
	case string(models.HandItemCategoryWeapon), string(models.HandItemCategoryStaff):
		damageMin, damageMax := generatedInventoryDamageRangeByRarity(rarity)
		if normalizedHandedness == string(models.HandednessTwoHanded) {
			damageMin = int(float64(damageMin) * 1.35)
			damageMax = int(float64(damageMax) * 1.35)
		}
		if normalizedCategory == string(models.HandItemCategoryStaff) {
			damageMin = int(float64(damageMin) * 0.9)
			damageMax = int(float64(damageMax) * 0.9)
			spellMin, spellMax := generatedInventorySpellBonusRangeByRarity(rarity)
			attrs.SpellDamageBonusPercent = intPtr(secureRandomIntBetween(spellMin, spellMax))
		}
		attrs.DamageMin = intPtr(maxInt(1, damageMin))
		attrs.DamageMax = intPtr(maxInt(*attrs.DamageMin, damageMax))
		attrs.DamageAffinity = stringPtr(generatedInventoryDamageAffinity(normalizedCategory))
		attrs.SwipesPerAttack = intPtr(generatedInventorySwipesPerAttack(normalizedCategory, normalizedHandedness))
	case string(models.HandItemCategoryShield):
		blockPctMin, blockPctMax := generatedInventoryBlockPercentRangeByRarity(rarity)
		blockMin, blockMax := generatedInventoryBlockedDamageRangeByRarity(rarity)
		attrs.BlockPercentage = intPtr(secureRandomIntBetween(blockPctMin, blockPctMax))
		attrs.DamageBlocked = intPtr(secureRandomIntBetween(blockMin, blockMax))
	case string(models.HandItemCategoryOrb):
		spellMin, spellMax := generatedInventorySpellBonusRangeByRarity(rarity)
		attrs.SpellDamageBonusPercent = intPtr(secureRandomIntBetween(spellMin, spellMax))
	}
	return attrs
}

func generatedInventoryDamageAffinity(category string) string {
	var options []models.DamageAffinity
	switch strings.ToLower(strings.TrimSpace(category)) {
	case string(models.HandItemCategoryStaff):
		options = []models.DamageAffinity{
			models.DamageAffinityArcane,
			models.DamageAffinityFire,
			models.DamageAffinityIce,
			models.DamageAffinityLightning,
			models.DamageAffinityShadow,
			models.DamageAffinityHoly,
		}
	default:
		options = []models.DamageAffinity{
			models.DamageAffinityPhysical,
			models.DamageAffinityFire,
			models.DamageAffinityIce,
			models.DamageAffinityLightning,
			models.DamageAffinityPoison,
		}
	}
	if len(options) == 0 {
		return string(models.DamageAffinityPhysical)
	}
	return string(options[secureRandomIntBetween(0, len(options)-1)])
}

func generatedInventoryDamageRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 3, 6
	case "uncommon":
		return 5, 10
	case "epic":
		return 9, 16
	case "mythic":
		return 14, 24
	default:
		return 3, 6
	}
}

func generatedInventorySwipesPerAttack(category string, handedness string) int {
	if category == string(models.HandItemCategoryStaff) {
		return secureRandomIntBetween(1, 2)
	}
	if handedness == string(models.HandednessTwoHanded) {
		return secureRandomIntBetween(1, 2)
	}
	return secureRandomIntBetween(2, 4)
}

func generatedInventoryBlockPercentRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 10, 18
	case "uncommon":
		return 18, 28
	case "epic":
		return 28, 42
	case "mythic":
		return 40, 58
	default:
		return 10, 18
	}
}

func generatedInventoryBlockedDamageRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 2, 5
	case "uncommon":
		return 4, 8
	case "epic":
		return 8, 14
	case "mythic":
		return 13, 20
	default:
		return 2, 5
	}
}

func generatedInventorySpellBonusRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 8, 14
	case "uncommon":
		return 14, 22
	case "epic":
		return 22, 34
	case "mythic":
		return 34, 50
	default:
		return 8, 14
	}
}

func secureRandomIntBetween(minValue int, maxValue int) int {
	if maxValue < minValue {
		maxValue = minValue
	}
	if minValue == maxValue {
		return minValue
	}
	diff := maxValue - minValue + 1
	n, err := crand.Int(crand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return minValue
	}
	return minValue + int(n.Int64())
}

func intPtr(value int) *int {
	v := value
	return &v
}

func stringPtr(value string) *string {
	v := value
	return &v
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *server) regenerateInventoryItemImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, id, map[string]interface{}{
		"image_generation_status": models.InventoryImageGenerationStatusQueued,
		"image_generation_error":  "",
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update inventory item: " + err.Error()})
		return
	}

	payload := jobs.GenerateInventoryItemImageTaskPayload{
		InventoryItemID: item.ID,
		Name:            item.Name,
		Description:     item.FlavorText,
		RarityTier:      item.RarityTier,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateInventoryItemImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = s.dbClient.InventoryItem().UpdateInventoryItem(ctx, id, map[string]interface{}{
			"image_generation_status": models.InventoryImageGenerationStatusFailed,
			"image_generation_error":  errMsg,
		})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedItem)
}

func (s *server) updateInventoryItem(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	// Check if item exists
	existingItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	var requestBody struct {
		Name                    string                         `json:"name"`
		ImageURL                string                         `json:"imageUrl"`
		FlavorText              string                         `json:"flavorText"`
		EffectText              string                         `json:"effectText"`
		RarityTier              string                         `json:"rarityTier"`
		IsCaptureType           bool                           `json:"isCaptureType"`
		UnlockTier              *int                           `json:"unlockTier"`
		EquipSlot               *string                        `json:"equipSlot"`
		StrengthMod             int                            `json:"strengthMod"`
		DexterityMod            int                            `json:"dexterityMod"`
		ConstitutionMod         int                            `json:"constitutionMod"`
		IntelligenceMod         int                            `json:"intelligenceMod"`
		WisdomMod               int                            `json:"wisdomMod"`
		CharismaMod             int                            `json:"charismaMod"`
		HandItemCategory        *string                        `json:"handItemCategory"`
		Handedness              *string                        `json:"handedness"`
		DamageMin               *int                           `json:"damageMin"`
		DamageMax               *int                           `json:"damageMax"`
		DamageAffinity          *string                        `json:"damageAffinity"`
		SwipesPerAttack         *int                           `json:"swipesPerAttack"`
		BlockPercentage         *int                           `json:"blockPercentage"`
		DamageBlocked           *int                           `json:"damageBlocked"`
		SpellDamageBonusPercent *int                           `json:"spellDamageBonusPercent"`
		ConsumeHealthDelta      int                            `json:"consumeHealthDelta"`
		ConsumeManaDelta        int                            `json:"consumeManaDelta"`
		ConsumeStatusesToAdd    []scenarioFailureStatusPayload `json:"consumeStatusesToAdd"`
		ConsumeStatusesToRemove []string                       `json:"consumeStatusesToRemove"`
		ConsumeSpellIDs         []string                       `json:"consumeSpellIds"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var equipSlot *string
	if requestBody.EquipSlot != nil {
		trimmed := strings.TrimSpace(*requestBody.EquipSlot)
		if trimmed != "" {
			if !models.IsValidInventoryEquipSlot(trimmed) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid equip slot"})
				return
			}
			equipSlot = &trimmed
		}
	}
	handAttrs, err := models.NormalizeAndValidateHandEquipment(equipSlot, models.HandEquipmentAttributes{
		HandItemCategory:        requestBody.HandItemCategory,
		Handedness:              requestBody.Handedness,
		DamageMin:               requestBody.DamageMin,
		DamageMax:               requestBody.DamageMax,
		DamageAffinity:          requestBody.DamageAffinity,
		SwipesPerAttack:         requestBody.SwipesPerAttack,
		BlockPercentage:         requestBody.BlockPercentage,
		DamageBlocked:           requestBody.DamageBlocked,
		SpellDamageBonusPercent: requestBody.SpellDamageBonusPercent,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	consumeStatusesToAdd, err := parseScenarioFailureStatusTemplates(requestBody.ConsumeStatusesToAdd, "consumeStatusesToAdd")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	consumeStatusesToRemove := parseInventoryConsumeStatusNames(requestBody.ConsumeStatusesToRemove)
	consumeSpellIDs, err := parseInventoryConsumeSpellIDs(requestBody.ConsumeSpellIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for idx, rawSpellID := range consumeSpellIDs {
		spellID, _ := uuid.Parse(rawSpellID)
		if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("consumeSpellIds[%d] not found", idx)})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updates := map[string]interface{}{
		"name":                       requestBody.Name,
		"image_url":                  requestBody.ImageURL,
		"flavor_text":                requestBody.FlavorText,
		"effect_text":                requestBody.EffectText,
		"rarity_tier":                requestBody.RarityTier,
		"is_capture_type":            requestBody.IsCaptureType,
		"unlock_tier":                requestBody.UnlockTier,
		"equip_slot":                 equipSlot,
		"strength_mod":               requestBody.StrengthMod,
		"dexterity_mod":              requestBody.DexterityMod,
		"constitution_mod":           requestBody.ConstitutionMod,
		"intelligence_mod":           requestBody.IntelligenceMod,
		"wisdom_mod":                 requestBody.WisdomMod,
		"charisma_mod":               requestBody.CharismaMod,
		"hand_item_category":         handAttrs.HandItemCategory,
		"handedness":                 handAttrs.Handedness,
		"damage_min":                 handAttrs.DamageMin,
		"damage_max":                 handAttrs.DamageMax,
		"damage_affinity":            handAttrs.DamageAffinity,
		"swipes_per_attack":          handAttrs.SwipesPerAttack,
		"block_percentage":           handAttrs.BlockPercentage,
		"damage_blocked":             handAttrs.DamageBlocked,
		"spell_damage_bonus_percent": handAttrs.SpellDamageBonusPercent,
		"consume_health_delta":       requestBody.ConsumeHealthDelta,
		"consume_mana_delta":         requestBody.ConsumeManaDelta,
		"consume_statuses_to_add":    consumeStatusesToAdd,
		"consume_statuses_to_remove": consumeStatusesToRemove,
		"consume_spell_ids":          consumeSpellIDs,
	}

	if requestBody.ImageURL != "" {
		updates["image_generation_status"] = models.InventoryImageGenerationStatusComplete
		updates["image_generation_error"] = ""
	}

	if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update inventory item: " + err.Error()})
		return
	}

	// Fetch the updated item
	updatedItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedItem)
}

func (s *server) deleteInventoryItem(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	// Check if item exists
	existingItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	if err := s.dbClient.InventoryItem().DeleteInventoryItem(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete inventory item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "inventory item deleted successfully"})
}

func (s *server) bulkDeleteInventoryItems(ctx *gin.Context) {
	var requestBody struct {
		IDs []int `json:"ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ids array is required"})
		return
	}

	if len(requestBody.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ids array cannot be empty"})
		return
	}

	seen := make(map[int]struct{}, len(requestBody.IDs))
	uniqueIDs := make([]int, 0, len(requestBody.IDs))
	for _, id := range requestBody.IDs {
		if id <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid inventory item ID: %d", id),
			})
			return
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	for _, id := range uniqueIDs {
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
		if err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("inventory item not found: %d", id),
				})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if item == nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("inventory item not found: %d", id),
			})
			return
		}
	}

	for _, id := range uniqueIDs {
		if err := s.dbClient.InventoryItem().DeleteInventoryItem(ctx, id); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to delete inventory item %d: %s", id, err.Error()),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("deleted %d inventory item(s)", len(uniqueIDs)),
		"deleted": len(uniqueIDs),
		"ids":     uniqueIDs,
	})
}

func (s *server) editTeamName(ctx *gin.Context) {
	stringTeamID := ctx.Param("teamID")
	if stringTeamID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "team ID is required",
		})
		return
	}

	teamID, err := uuid.Parse(stringTeamID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid team ID",
		})
		return
	}

	var requestBody struct {
		Name string `binding:"required" json:"name"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	team, err := s.dbClient.Team().UpdateTeamName(ctx, teamID, requestBody.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, team)
}

func (s *server) submitAnswerPointOfInterestChallenge(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		ChallengeID        uuid.UUID  `binding:"required" json:"challengeID"`
		TeamID             *uuid.UUID `json:"teamID"`
		UserID             *uuid.UUID `json:"userID"`
		TextSubmission     string     `json:"textSubmission"`
		ImageSubmissionUrl string     `json:"imageSubmissionUrl"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	challenge, err := s.dbClient.PointOfInterestChallenge().FindByID(ctx, requestBody.ChallengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge.PointOfInterestGroupID != nil {
		acceptanceV2, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, user.ID, *challenge.PointOfInterestGroupID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if acceptanceV2 == nil {
			acceptanceLegacy, err := s.dbClient.QuestAcceptance().FindByUserAndQuest(ctx, user.ID, *challenge.PointOfInterestGroupID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if acceptanceLegacy == nil {
				ctx.JSON(http.StatusForbidden, gin.H{"error": "quest must be accepted before completing challenges"})
				return
			}
		}
		meetsReputation, _, requiredLevel, err := s.userMeetsQuestReputation(ctx, user.ID, *challenge.PointOfInterestGroupID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !meetsReputation {
			ctx.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("requires zone reputation level %d", requiredLevel)})
			return
		}
	}

	submissionResult, err := s.gameEngineClient.ProcessSubmission(ctx, gameengine.Submission{
		ChallengeID: requestBody.ChallengeID,
		TeamID:      requestBody.TeamID,
		UserID:      requestBody.UserID,
		Text:        requestBody.TextSubmission,
		ImageURL:    requestBody.ImageSubmissionUrl,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, submissionResult)
}

func (s *server) submitQuestNodeChallenge(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	nodeIDStr := ctx.Param("id")
	if nodeIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest node id is required"})
		return
	}
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest node id"})
		return
	}

	var requestBody struct {
		QuestNodeChallengeID *uuid.UUID `json:"questNodeChallengeId"`
		TextSubmission       string     `json:"textSubmission"`
		ImageSubmissionUrl   string     `json:"imageSubmissionUrl"`
		VideoSubmissionUrl   string     `json:"videoSubmissionUrl"`
		TeamID               *uuid.UUID `json:"teamID"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := s.dbClient.QuestNode().FindByID(ctx, nodeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if node == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest node not found"})
		return
	}

	quest, err := s.dbClient.Quest().FindByID(ctx, node.QuestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}

	acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, user.ID, quest.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if acceptance == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "quest must be accepted before completing challenges"})
		return
	}
	if acceptance.TurnedInAt != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest already turned in"})
		return
	}

	currentNode, err := s.currentQuestNode(ctx, quest, acceptance.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if currentNode == nil || currentNode.ID != node.ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest node is not the current objective"})
		return
	}

	challenge, err := selectQuestNodeChallenge(node, requestBody.QuestNodeChallengeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submissionType := challenge.SubmissionType
	if strings.TrimSpace(string(submissionType)) == "" {
		submissionType = node.SubmissionType
	}
	if strings.TrimSpace(string(submissionType)) == "" {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	switch submissionType {
	case models.QuestNodeSubmissionTypeText:
		if strings.TrimSpace(requestBody.TextSubmission) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "text submission is required"})
			return
		}
	case models.QuestNodeSubmissionTypePhoto:
		if strings.TrimSpace(requestBody.ImageSubmissionUrl) == "" && strings.TrimSpace(requestBody.TextSubmission) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "photo submission is required"})
			return
		}
	case models.QuestNodeSubmissionTypeVideo:
		if strings.TrimSpace(requestBody.VideoSubmissionUrl) == "" && strings.TrimSpace(requestBody.TextSubmission) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "video submission is required"})
			return
		}
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if node.PointOfInterestID != nil {
		poi, err := s.dbClient.PointOfInterest().FindByID(ctx, *node.PointOfInterestID)
		if err != nil || poi == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load point of interest"})
			return
		}
		poiLat, err := strconv.ParseFloat(poi.Lat, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest latitude"})
			return
		}
		poiLng, err := strconv.ParseFloat(poi.Lng, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest longitude"})
			return
		}
		distance := util.HaversineDistance(userLat, userLng, poiLat, poiLng)
		if distance > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you must be within 100 meters of the location to submit an answer. Currently %.0f meters away", distance)})
			return
		}
	} else if node.ScenarioID != nil {
		scenario, err := s.dbClient.Scenario().FindByID(ctx, *node.ScenarioID)
		if err != nil || scenario == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load scenario"})
			return
		}
		distance := util.HaversineDistance(userLat, userLng, scenario.Latitude, scenario.Longitude)
		if distance > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you must be within 100 meters of the location to submit an answer. Currently %.0f meters away", distance)})
			return
		}
	} else if node.MonsterID != nil {
		monster, err := s.dbClient.Monster().FindByID(ctx, *node.MonsterID)
		if err != nil || monster == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load monster"})
			return
		}
		distance := util.HaversineDistance(userLat, userLng, monster.Latitude, monster.Longitude)
		if distance > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you must be within 100 meters of the location to submit an answer. Currently %.0f meters away", distance)})
			return
		}
	} else if node.Polygon != "" {
		polygon, err := parseQuestNodePolygon(node.Polygon)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest polygon"})
			return
		}
		if !planar.PolygonContains(polygon, orb.Point{userLng, userLat}) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "you must be inside the quest area to submit"})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest node has no location"})
		return
	}

	textSubmission := requestBody.TextSubmission
	if strings.TrimSpace(textSubmission) == "" && strings.TrimSpace(requestBody.VideoSubmissionUrl) != "" {
		textSubmission = fmt.Sprintf("Video submission URL: %s", requestBody.VideoSubmissionUrl)
	}

	judgement, err := s.judgeClient.JudgeFreeform(ctx, judge.FreeformJudgeSubmissionRequest{
		Question:           challenge.Question,
		ImageSubmissionUrl: requestBody.ImageSubmissionUrl,
		TextSubmission:     textSubmission,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	score := int(math.Round(judgement.Judgement.Score))
	if score < 0 {
		score = 0
	} else if score > 50 {
		score = 50
	}

	validStats := map[string]struct{}{
		"strength":     {},
		"dexterity":    {},
		"constitution": {},
		"intelligence": {},
		"wisdom":       {},
		"charisma":     {},
	}
	statTags := []string{}
	seenTags := map[string]struct{}{}
	for _, tag := range []string(challenge.StatTags) {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := validStats[normalized]; !ok {
			continue
		}
		if _, ok := seenTags[normalized]; ok {
			continue
		}
		seenTags[normalized] = struct{}{}
		statTags = append(statTags, normalized)
	}

	combinedScore := score
	var statValues map[string]int
	if len(statTags) > 0 {
		stats, err := s.dbClient.UserCharacterStats().FindOrCreateForUser(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		statusBonuses, err := s.dbClient.UserStatus().GetActiveStatBonuses(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		totalBonuses := equipmentBonuses.Add(statusBonuses)
		statValues = map[string]int{}
		statValueSum := 0
		for _, tag := range statTags {
			value := 0
			switch tag {
			case "strength":
				value = stats.Strength + totalBonuses.Strength
			case "dexterity":
				value = stats.Dexterity + totalBonuses.Dexterity
			case "constitution":
				value = stats.Constitution + totalBonuses.Constitution
			case "intelligence":
				value = stats.Intelligence + totalBonuses.Intelligence
			case "wisdom":
				value = stats.Wisdom + totalBonuses.Wisdom
			case "charisma":
				value = stats.Charisma + totalBonuses.Charisma
			}
			statValues[tag] = value
			statValueSum += value
		}
		combinedScore = score + statValueSum
	}

	difficulty := challenge.Difficulty
	if combinedScore < difficulty {
		reason := strings.TrimSpace(judgement.Judgement.Reason)
		if reason == "" {
			reason = "Submission did not meet the difficulty threshold."
		}
		scoreValue := score
		difficultyValue := difficulty
		combinedValue := combinedScore
		ctx.JSON(http.StatusOK, gameengine.SubmissionResult{
			Successful:     false,
			Reason:         reason,
			QuestCompleted: false,
			Score:          &scoreValue,
			Difficulty:     &difficultyValue,
			CombinedScore:  &combinedValue,
			StatTags:       statTags,
			StatValues:     statValues,
		})
		return
	}

	progress, err := s.dbClient.QuestNodeProgress().FindByAcceptanceAndNode(ctx, acceptance.ID, node.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shouldAward := progress == nil || progress.CompletedAt == nil
	if progress == nil {
		now := time.Now()
		progress = &models.QuestNodeProgress{
			ID:                uuid.New(),
			CreatedAt:         now,
			UpdatedAt:         now,
			QuestAcceptanceID: acceptance.ID,
			QuestNodeID:       node.ID,
			CompletedAt:       &now,
		}
		if err := s.dbClient.QuestNodeProgress().Create(ctx, progress); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if progress.CompletedAt == nil {
		if err := s.dbClient.QuestNodeProgress().MarkCompleted(ctx, progress.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	completed, err := s.questlogClient.AreQuestObjectivesComplete(ctx, user.ID, quest.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if shouldAward {
		if err := s.gameEngineClient.AwardQuestNodeSubmissionRewards(ctx, user.ID, requestBody.TeamID, quest, node, challenge, completed); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	scoreValue := score
	difficultyValue := difficulty
	combinedValue := combinedScore
	successReason := strings.TrimSpace(judgement.Judgement.Reason)
	if successReason == "" {
		successReason = "Challenge completed successfully!"
	}
	ctx.JSON(http.StatusOK, gameengine.SubmissionResult{
		Successful:     true,
		Reason:         successReason,
		QuestCompleted: completed,
		Score:          &scoreValue,
		Difficulty:     &difficultyValue,
		CombinedScore:  &combinedValue,
		StatTags:       statTags,
		StatValues:     statValues,
	})
}

func (s *server) getPresignedUploadUrl(ctx *gin.Context) {
	var requestBody struct {
		Bucket string `binding:"required" json:"bucket"`
		Key    string `binding:"required" json:"key"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	url, err := s.awsClient.GeneratePresignedUploadURL(requestBody.Bucket, requestBody.Key, time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

func (s *server) leaveMatch(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	stringMatchID := ctx.Param("id")
	if stringMatchID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "match ID is required",
		})
		return
	}

	matchID, err := uuid.Parse(stringMatchID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid match ID",
		})
		return
	}

	if err := s.dbClient.Team().RemoveUserFromMatch(ctx, matchID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user left match successfully",
	})
}

func (s *server) getCurrentMatch(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	match, err := s.dbClient.Match().FindCurrentMatchForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if match == nil || match.EndedAt != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no current match",
		})
		return
	}

	ctx.JSON(http.StatusOK, match)
}

func (s *server) getPointsOfInterestGroups(ctx *gin.Context) {
	intTypeAsString := ctx.Query("type")
	var typeValue models.PointOfInterestGroupType
	if intTypeAsString != "" {
		intType, err := strconv.Atoi(intTypeAsString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid type value",
			})
			return
		}
		typeValue = models.PointOfInterestGroupType(intType)
		groups, err := s.dbClient.PointOfInterestGroup().FindByType(ctx, typeValue)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, groups)
		return
	}

	groups, err := s.dbClient.PointOfInterestGroup().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (s *server) createPointOfInterestGroup(ctx *gin.Context) {
	var requestBody struct {
		Name                    string     `binding:"required" json:"name"`
		Description             string     `binding:"required" json:"description"`
		ImageUrl                string     `binding:"required" json:"imageUrl"`
		Type                    int        `binding:"required" json:"type"`
		Gold                    *int       `json:"gold"`
		InventoryItemID         *int       `json:"inventoryItemId"`
		RequiredReputationLevel *int       `json:"requiredReputationLevel"`
		QuestGiverCharacterID   *uuid.UUID `json:"questGiverCharacterId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	group, err := s.dbClient.PointOfInterestGroup().Create(ctx, requestBody.Name, requestBody.Description, requestBody.ImageUrl, models.PointOfInterestGroupType(requestBody.Type))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if requestBody.QuestGiverCharacterID != nil {
		character, err := s.dbClient.Character().FindByID(ctx, *requestBody.QuestGiverCharacterID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if character == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "quest giver character not found"})
			return
		}
	}

	if requestBody.Gold != nil || requestBody.InventoryItemID != nil || requestBody.QuestGiverCharacterID != nil || requestBody.RequiredReputationLevel != nil {
		// Update the group with provided gold value and/or inventory item
		if requestBody.Gold != nil {
			group.Gold = *requestBody.Gold
		}
		if requestBody.InventoryItemID != nil {
			group.InventoryItemID = requestBody.InventoryItemID
		}
		if requestBody.RequiredReputationLevel != nil {
			group.RequiredReputationLevel = requestBody.RequiredReputationLevel
		}
		if requestBody.QuestGiverCharacterID != nil {
			group.QuestGiverCharacterID = requestBody.QuestGiverCharacterID
		}
		if err := s.dbClient.PointOfInterestGroup().Update(ctx, group.ID, &models.PointOfInterestGroup{
			Gold:                    group.Gold,
			InventoryItemID:         group.InventoryItemID,
			RequiredReputationLevel: group.RequiredReputationLevel,
			QuestGiverCharacterID:   group.QuestGiverCharacterID,
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, group)
}

func (s *server) GetPointsWithinRadius(ctx *gin.Context) {

}

func (s *server) getPointOfInterestGroup(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "group ID is required",
		})
		return
	}

	uuidGroupID, err := uuid.Parse(groupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid group ID",
		})
		return
	}

	group, err := s.dbClient.PointOfInterestGroup().FindByID(ctx, uuidGroupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, group)
}

func (s *server) deleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("id")
	if categoryID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "category ID is required",
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid category ID",
		})
		return
	}

	if err := s.dbClient.SonarCategory().DeleteCategory(ctx, categoryUUID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "category deleted successfully",
	})
}

func (s *server) deleteActivity(ctx *gin.Context) {
	activityID := ctx.Param("id")
	if activityID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "activity ID is required",
		})
		return
	}

	id, err := uuid.Parse(activityID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity ID"})
		return
	}

	if err := s.dbClient.Activity().DeleteByID(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "activity deleted successfully"})
}

func (s *server) createCategory(ctx *gin.Context) {
	var requestBody struct {
		Title string `binding:"required" json:"name"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	category, err := s.dbClient.SonarCategory().CreateCategory(ctx, models.SonarCategory{
		Title:  requestBody.Title,
		UserID: &user.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

func (s *server) createActivity(ctx *gin.Context) {
	var requestBody struct {
		Title      string    `binding:"required" json:"title"`
		CategoryID uuid.UUID `binding:"required" json:"categoryId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	activity, err := s.dbClient.SonarActivity().CreateActivity(ctx, models.SonarActivity{
		Title:           requestBody.Title,
		SonarCategoryID: requestBody.CategoryID,
		UserID:          &user.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, activity)
}

func (s *server) whoami(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *server) getSubmission(ctx *gin.Context) {
	submissionID := ctx.Param("id")
	if submissionID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "submission ID is required",
		})
		return
	}

	submissionUUID, err := uuid.Parse(submissionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid survey ID",
		})
		return
	}

	submission, err := s.dbClient.SonarSurveySubmission().GetSubmissionByID(ctx, submissionUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch submission",
		})
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (s *server) getSubmissionForSurvey(ctx *gin.Context) {
	surveyID := ctx.Param("id")
	if surveyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "survey ID is required",
		})
		return
	}

	surveyUUID, err := uuid.Parse(surveyID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid survey ID",
		})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	submission, err := s.dbClient.SonarSurveySubmission().GetUserSubmissionForSurvey(ctx, user.ID, surveyUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch submission for survey",
		})
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

func (s *server) getAuthenticatedUser(ctx *gin.Context) (*models.User, error) {
	u, ok := ctx.Get("user")
	if !ok {
		return nil, ErrNotAuthenticated
	}

	user, ok := u.(*models.User)
	if !ok {
		return nil, ErrNotAuthenticated
	}

	return user, nil
}

func (s *server) getSurverys(ctx *gin.Context) {
	user, ok := ctx.Get("user")

	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "no user found in request",
		})
		return
	}

	surveys, err := s.dbClient.SonarSurvey().GetSurveys(ctx, user.(*models.User).ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "survey fetch error").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, surveys)
}

func (s *server) login(ctx *gin.Context) {
	var requestBody auth.LoginByTextRequest

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	authenticateResponse, err := s.authClient.LoginByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payload := gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	}

	ctx.JSON(200, payload)
}

func (s *server) newSurvey(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var newSurveyRequest struct {
		ActivityIDs []uuid.UUID `binding:"required" json:"activityIds"`
		Name        string      `binding:"required" json:"name"`
	}

	if err := ctx.Bind(&newSurveyRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	survey, err := s.dbClient.SonarSurvey().CreateSurvey(ctx, user.ID, newSurveyRequest.Name, newSurveyRequest.ActivityIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, survey)
}

func (s *server) register(ctx *gin.Context) {
	var requestBody auth.RegisterByTextRequest

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if requestBody.Username != nil && !util.ValidateUsername(*requestBody.Username) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid username",
		})
		return
	}

	authenticateResponse, err := s.authClient.RegisterByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.NewUserStarterConfig().ApplyToUser(ctx, authenticateResponse.User.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to apply starter config: %s", err.Error()),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	})
}

func (s *server) createNeighbor(c *gin.Context) {
	var neighbor models.NeighboringPointsOfInterest

	if err := c.Bind(&neighbor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "shit neighbor create request",
		})
		return
	}

	if err := s.dbClient.NeighboringPointsOfInterest().Create(c, neighbor.PointOfInterestOneID, neighbor.PointOfInterestTwoID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "everythings ok",
	})
}

func (s *server) getNeighbors(c *gin.Context) {
	neighbors, err := s.dbClient.NeighboringPointsOfInterest().FindAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, neighbors)
}

func (s *server) getTeams(c *gin.Context) {
	teams, err := s.dbClient.Team().GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var userIDs []uuid.UUID
	for _, team := range teams {
		for _, userTeam := range team.UserTeams {
			userIDs = append(userIDs, userTeam.UserID)
		}
	}

	payload := gin.H{
		"teams": teams,
	}

	if len(teams) > 0 {
		users, err := s.authClient.GetUsers(c, userIDs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		payload["users"] = users
	}

	c.JSON(200, payload)
}

func (s *server) getPointsOfInterest(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	matchID, err := s.dbClient.Match().FindCurrentMatchIDForUser(ctx, user.ID)
	if err != nil && err != sql.ErrNoRows && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if matchID != nil {
		match, err := s.dbClient.Match().FindByID(ctx, *matchID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		availability, err := s.questAvailabilityByCharacter(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for i := range match.PointsOfInterest {
			poi := match.PointsOfInterest[i]
			hasAvailable := false
			for j := range poi.Characters {
				ch := poi.Characters[j]
				if availability[ch.ID] {
					hasAvailable = true
				}
				poi.Characters[j].HasAvailableQuest = availability[ch.ID]
			}
			match.PointsOfInterest[i].HasAvailableQuest = hasAvailable
		}
		ctx.JSON(200, match.PointsOfInterest)
		return
	}

	pointOfInterests, err := s.dbClient.PointOfInterest().FindAll(ctx)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	availability, err := s.questAvailabilityByCharacter(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range pointOfInterests {
		poi := pointOfInterests[i]
		hasAvailable := false
		for j := range poi.Characters {
			ch := poi.Characters[j]
			if availability[ch.ID] {
				hasAvailable = true
			}
			poi.Characters[j].HasAvailableQuest = availability[ch.ID]
		}
		pointOfInterests[i].HasAvailableQuest = hasAvailable
	}
	ctx.JSON(200, pointOfInterests)
}

func (s *server) createMatch(c *gin.Context) {
	var createMatchRequest struct {
		PointsOfInterestIDs []uuid.UUID `json:"pointsOfInterestIds" binding:"required"`
	}

	if err := c.Bind(&createMatchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := s.getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	match, err := s.dbClient.Match().Create(c, user.ID, createMatchRequest.PointsOfInterestIDs)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, match)
}

func (s *server) createTeamForMatch(c *gin.Context) {
	stringMatchID := c.Param("id")

	matchID, err := uuid.Parse(stringMatchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error() + "for match id"})
		return
	}

	var createTeamForMatchRequest struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.Bind(&createTeamForMatchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(createTeamForMatchRequest.UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error() + "for user id"})
		return
	}

	team, err := s.dbClient.Team().Create(c, []uuid.UUID{userID}, util.GenerateTeamName(), matchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, team)
}

func (s *server) addUserToTeam(c *gin.Context) {
	stringTeamID := c.Param("teamID")

	teamID, err := uuid.Parse(stringTeamID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var addUserToTeamRequest struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.Bind(&addUserToTeamRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(addUserToTeamRequest.UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Team().AddUserToTeam(c, teamID, userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "done"})
}

func (s *server) getMatch(c *gin.Context) {
	matchID := c.Param("id")

	uuidMatchID, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	match, err := s.dbClient.Match().FindByID(c, uuidMatchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, match)
}

func (s *server) startMatch(c *gin.Context) {
	user, err := s.getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	matchID := c.Param("id")

	uuidMatchID, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	match, err := s.dbClient.Match().FindByID(c, uuidMatchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if match.StartedAt != nil {
		c.JSON(400, gin.H{"error": "match already started"})
		return
	}

	if match.CreatorID != user.ID {
		c.JSON(401, gin.H{"error": "you are not the creator of this match"})
		return
	}

	if err := s.dbClient.Match().StartMatch(c, uuidMatchID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "done"})
}

func (s *server) endMatch(c *gin.Context) {
	matchID := c.Param("id")

	uuidMatchID, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Match().EndMatch(c, uuidMatchID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "done"})
}

func (s *server) unlockPointOfInterest(c *gin.Context) {
	user, err := s.getAuthenticatedUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var pointOfInterestUnlockRequest struct {
		TeamID            *uuid.UUID `json:"teamId"`
		UserID            *uuid.UUID `json:"userId"`
		PointOfInterestID uuid.UUID  `json:"pointOfInterestId" binding:"required"`
		Lat               string     `json:"lat" binding:"required"`
		Lng               string     `json:"lng" binding:"required"`
	}

	if err := c.Bind(&pointOfInterestUnlockRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	pointOfInterest, err := s.dbClient.PointOfInterest().FindByID(c, pointOfInterestUnlockRequest.PointOfInterestID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	latPOI, err := strconv.ParseFloat(pointOfInterest.Lat, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid latitude format"})
		return
	}
	lngPOI, err := strconv.ParseFloat(pointOfInterest.Lng, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid longitude format"})
		return
	}
	latReq, err := strconv.ParseFloat(pointOfInterestUnlockRequest.Lat, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request latitude format"})
		return
	}
	lngReq, err := strconv.ParseFloat(pointOfInterestUnlockRequest.Lng, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request longitude format"})
		return
	}

	distanceFromPOI := util.HaversineDistance(latPOI, lngPOI, latReq, lngReq)

	if distanceFromPOI > 200 {
		c.JSON(400, gin.H{"error": fmt.Sprintf("point of interest is not within 200 meters: %f", distanceFromPOI)})
		return
	}

	// Check if POI is locked
	var unlockItemID *uuid.UUID
	if pointOfInterest.UnlockTier != nil {
		// Find user's owned inventory items with unlock tier
		ownedItems, err := s.dbClient.InventoryItem().GetItems(c, models.OwnedInventoryItem{UserID: &user.ID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user inventory"})
			return
		}

		// Find inventory items with unlock tier >= POI unlock tier
		var validUnlockItems []models.OwnedInventoryItem
		for _, ownedItem := range ownedItems {
			if ownedItem.Quantity > 0 {
				// Get the inventory item to check unlock tier
				inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(c, ownedItem.InventoryItemID)
				if err == nil && inventoryItem != nil && inventoryItem.UnlockTier != nil {
					if *inventoryItem.UnlockTier >= *pointOfInterest.UnlockTier {
						validUnlockItems = append(validUnlockItems, ownedItem)
					}
				}
			}
		}

		if len(validUnlockItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "you do not have an item with sufficient unlock tier to unlock this point of interest"})
			return
		}

		// Find the item with the lowest unlock tier
		var lowestTierItem *models.OwnedInventoryItem
		var lowestTier int
		for i := range validUnlockItems {
			inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(c, validUnlockItems[i].InventoryItemID)
			if err == nil && inventoryItem != nil && inventoryItem.UnlockTier != nil {
				if lowestTierItem == nil || *inventoryItem.UnlockTier < lowestTier {
					lowestTier = *inventoryItem.UnlockTier
					lowestTierItem = &validUnlockItems[i]
				}
			}
		}

		if lowestTierItem == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not find valid unlock item"})
			return
		}

		unlockItemID = &lowestTierItem.ID
	}

	// Consume unlock item if needed
	if unlockItemID != nil {
		if err := s.dbClient.InventoryItem().UseInventoryItem(c, *unlockItemID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to consume unlock item: " + err.Error()})
			return
		}
	}

	if err := s.dbClient.PointOfInterest().Unlock(c, pointOfInterestUnlockRequest.PointOfInterestID, pointOfInterestUnlockRequest.TeamID, pointOfInterestUnlockRequest.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if pointOfInterestUnlockRequest.TeamID != nil {
		if err := s.chatClient.AddUnlockMessage(c, pointOfInterestUnlockRequest.TeamID, pointOfInterestUnlockRequest.UserID, pointOfInterestUnlockRequest.PointOfInterestID); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{
		"message": "everything cool",
	})
}

// User management endpoints
func (s *server) deleteUser(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Delete all dependent data before deleting the user
	// 1. Delete all discoveries
	if err := s.dbClient.PointOfInterestDiscovery().DeleteByUserID(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete discoveries: " + err.Error()})
		return
	}

	// 2. Delete all submissions
	if err := s.dbClient.PointOfInterestChallenge().DeleteAllSubmissionsForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete submissions: " + err.Error()})
		return
	}

	// 3. Delete all activities
	if err := s.dbClient.Activity().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete activities: " + err.Error()})
		return
	}

	// 4. Delete all tracked point of interest groups
	if err := s.dbClient.TrackedQuest().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tracked groups: " + err.Error()})
		return
	}

	// 5. Delete all friend relationships
	if err := s.dbClient.Friend().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete friends: " + err.Error()})
		return
	}

	// 6. Delete all friend invites
	if err := s.dbClient.FriendInvite().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete friend invites: " + err.Error()})
		return
	}

	// 7. Delete all party invites
	if err := s.dbClient.PartyInvite().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete party invites: " + err.Error()})
		return
	}

	// 8. Delete all image generations
	if err := s.dbClient.ImageGeneration().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete image generations: " + err.Error()})
		return
	}

	// 9. Delete all audit items
	if err := s.dbClient.AuditItem().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete audit items: " + err.Error()})
		return
	}

	// 10. Delete all how many answers
	if err := s.dbClient.HowManyAnswer().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete how many answers: " + err.Error()})
		return
	}

	// 11. Delete all how many subscriptions
	if err := s.dbClient.HowManySubscription().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete how many subscriptions: " + err.Error()})
		return
	}

	// 12. Delete all sonar survey submissions
	if err := s.dbClient.SonarSurveySubmission().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete sonar survey submissions: " + err.Error()})
		return
	}

	// 13. Delete all match users
	if err := s.dbClient.MatchUser().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete match users: " + err.Error()})
		return
	}

	// 14. Delete all user levels
	if err := s.dbClient.UserLevel().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user levels: " + err.Error()})
		return
	}

	// 15. Delete all user zone reputations
	if err := s.dbClient.UserZoneReputation().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user zone reputations: " + err.Error()})
		return
	}

	// 16. Delete all user proficiencies
	if err := s.dbClient.UserProficiency().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user proficiencies: " + err.Error()})
		return
	}

	// 17. Delete all user statuses
	if err := s.dbClient.UserStatus().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user statuses: " + err.Error()})
		return
	}

	// 18. Delete all owned inventory items
	if err := s.dbClient.InventoryItem().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete owned inventory items: " + err.Error()})
		return
	}

	// 19. Delete all user team relationships
	if err := s.dbClient.UserTeam().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user team relationships: " + err.Error()})
		return
	}

	// 20. Delete all parties where user is leader
	if err := s.dbClient.Party().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete parties: " + err.Error()})
		return
	}

	// 21. Finally, delete the user
	if err := s.dbClient.User().Delete(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (s *server) getUserDiscoveries(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	discoveries, err := s.dbClient.PointOfInterestDiscovery().GetDiscoveriesForUser(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, discoveries)
}

func (s *server) createUserDiscoveries(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var requestBody struct {
		PointOfInterestIDs []uuid.UUID `json:"pointOfInterestIds" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create discoveries for each POI
	for _, poiID := range requestBody.PointOfInterestIDs {
		discovery := &models.PointOfInterestDiscovery{
			UserID:            &userID,
			PointOfInterestID: poiID,
		}
		if err := s.dbClient.PointOfInterestDiscovery().Create(ctx, discovery); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "discoveries created successfully"})
}

func (s *server) deleteUserDiscovery(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	discoveryID, err := uuid.Parse(ctx.Param("discoveryId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid discovery ID"})
		return
	}

	if err := s.dbClient.PointOfInterestDiscovery().DeleteByID(ctx, discoveryID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "discovery deleted successfully"})
}

func (s *server) deleteAllUserDiscoveries(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.PointOfInterestDiscovery().DeleteByUserID(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "all discoveries deleted successfully"})
}

func (s *server) getUserSubmissions(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	submissions, err := s.dbClient.PointOfInterestChallenge().GetSubmissionsForUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, submissions)
}

func (s *server) deleteSubmission(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	submissionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission ID"})
		return
	}

	if err := s.dbClient.PointOfInterestChallenge().DeleteSubmission(ctx, submissionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "submission deleted successfully"})
}

func (s *server) deleteAllUserSubmissions(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.PointOfInterestChallenge().DeleteAllSubmissionsForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "all submissions deleted successfully"})
}

func (s *server) getUserActivities(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	activities, err := s.dbClient.Activity().GetFeed(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, activities)
}

func (s *server) deleteAllUserActivities(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.Activity().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "all activities deleted successfully"})
}

func (s *server) updateUserGold(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var requestBody struct {
		Gold int `json:"gold" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Gold < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "gold amount must be >= 0"})
		return
	}

	if err := s.dbClient.User().SetGold(ctx, userID, requestBody.Gold); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update gold: " + err.Error()})
		return
	}

	// Fetch and return updated user
	user, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (s *server) adminCreateUserStatus(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var requestBody struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		Effect          string `json:"effect"`
		EffectType      string `json:"effectType"`
		Positive        *bool  `json:"positive"`
		DamagePerTick   int    `json:"damagePerTick"`
		DurationSeconds int    `json:"durationSeconds" binding:"required"`
		StrengthMod     int    `json:"strengthMod"`
		DexterityMod    int    `json:"dexterityMod"`
		ConstitutionMod int    `json:"constitutionMod"`
		IntelligenceMod int    `json:"intelligenceMod"`
		WisdomMod       int    `json:"wisdomMod"`
		CharismaMod     int    `json:"charismaMod"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestBody.Name = strings.TrimSpace(requestBody.Name)
	if requestBody.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if requestBody.DurationSeconds <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "durationSeconds must be > 0"})
		return
	}
	effectType := normalizeUserStatusEffectType(requestBody.EffectType)
	if effectType == models.UserStatusEffectTypeDamageOverTime && requestBody.DamagePerTick <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "damagePerTick must be > 0 for damage_over_time statuses"})
		return
	}

	user, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user: " + err.Error()})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	isPositive := true
	if requestBody.Positive != nil {
		isPositive = *requestBody.Positive
	}

	now := time.Now()
	status := &models.UserStatus{
		UserID:          userID,
		Name:            requestBody.Name,
		Description:     strings.TrimSpace(requestBody.Description),
		Effect:          strings.TrimSpace(requestBody.Effect),
		Positive:        isPositive,
		EffectType:      effectType,
		DamagePerTick:   requestBody.DamagePerTick,
		StrengthMod:     requestBody.StrengthMod,
		DexterityMod:    requestBody.DexterityMod,
		ConstitutionMod: requestBody.ConstitutionMod,
		IntelligenceMod: requestBody.IntelligenceMod,
		WisdomMod:       requestBody.WisdomMod,
		CharismaMod:     requestBody.CharismaMod,
		StartedAt:       now,
		ExpiresAt:       now.Add(time.Duration(requestBody.DurationSeconds) * time.Second),
	}

	if err := s.dbClient.UserStatus().Create(ctx, status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user status: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, status)
}

func (s *server) adminAdjustUserResources(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var requestBody struct {
		HealthDelta int `json:"healthDelta"`
		ManaDelta   int `json:"manaDelta"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.HealthDelta == 0 && requestBody.ManaDelta == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one of healthDelta or manaDelta must be non-zero"})
		return
	}

	user, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user: " + err.Error()})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, userID, userLevel.Level); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(
		ctx,
		userID,
		-requestBody.HealthDelta,
		-requestBody.ManaDelta,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	spells, err := s.dbClient.UserSpell().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses, spells))
}

func (s *server) deleteUsers(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		UserIDs []uuid.UUID `json:"userIds" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(requestBody.UserIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no user IDs provided"})
		return
	}

	// Delete each user with all their dependencies
	for _, userID := range requestBody.UserIDs {
		// Delete all dependent data for this user
		// 1. Delete all discoveries
		if err := s.dbClient.PointOfInterestDiscovery().DeleteByUserID(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete discoveries for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 2. Delete all submissions
		if err := s.dbClient.PointOfInterestChallenge().DeleteAllSubmissionsForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete submissions for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 3. Delete all activities
		if err := s.dbClient.Activity().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete activities for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 4. Delete all tracked point of interest groups
		if err := s.dbClient.TrackedQuest().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tracked groups for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 5. Delete all friend relationships
		if err := s.dbClient.Friend().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete friends for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 6. Delete all friend invites
		if err := s.dbClient.FriendInvite().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete friend invites for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 7. Delete all party invites
		if err := s.dbClient.PartyInvite().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete party invites for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 8. Delete all image generations
		if err := s.dbClient.ImageGeneration().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete image generations for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 9. Delete all user proficiencies
		if err := s.dbClient.UserProficiency().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user proficiencies for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 10. Delete all user statuses
		if err := s.dbClient.UserStatus().DeleteAllForUser(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user statuses for user " + userID.String() + ": " + err.Error()})
			return
		}

		// 11. Finally, delete the user
		if err := s.dbClient.User().Delete(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user " + userID.String() + ": " + err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("successfully deleted %d users", len(requestBody.UserIDs))})
}

// Character handlers
func (s *server) getCharacters(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	characters, err := s.dbClient.Character().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	availability, err := s.questAvailabilityByCharacter(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range characters {
		ch := characters[i]
		if ch != nil {
			ch.HasAvailableQuest = availability[ch.ID]
		}
	}

	ctx.JSON(http.StatusOK, characters)
}

func (s *server) getCharacter(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	character, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if character == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	ctx.JSON(http.StatusOK, character)
}

func (s *server) getCharacterLocations(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	locations, err := s.dbClient.CharacterLocation().FindByCharacterID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, locations)
}

func (s *server) createCharacter(ctx *gin.Context) {
	var requestBody struct {
		Name              string     `json:"name" binding:"required"`
		Description       string     `json:"description"`
		MapIconUrl        string     `json:"mapIconUrl"`
		DialogueImageUrl  string     `json:"dialogueImageUrl"`
		ThumbnailUrl      string     `json:"thumbnailUrl"`
		PointOfInterestID *uuid.UUID `json:"pointOfInterestId"`
		MovementPattern   struct {
			MovementPatternType models.MovementPatternType `json:"movementPatternType" binding:"required"`
			ZoneID              *uuid.UUID                 `json:"zoneId"`
			StartingLatitude    float64                    `json:"startingLatitude"`
			StartingLongitude   float64                    `json:"startingLongitude"`
			Path                []models.Location          `json:"path"`
		} `json:"movementPattern" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.PointOfInterestID != nil {
		_, err := s.dbClient.PointOfInterest().FindByID(ctx, *requestBody.PointOfInterestID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "point of interest not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load point of interest: " + err.Error()})
			return
		}
	}

	// First create the movement pattern
	movementPattern := &models.MovementPattern{
		MovementPatternType: requestBody.MovementPattern.MovementPatternType,
		ZoneID:              requestBody.MovementPattern.ZoneID,
		StartingLatitude:    requestBody.MovementPattern.StartingLatitude,
		StartingLongitude:   requestBody.MovementPattern.StartingLongitude,
		Path:                models.LocationPath(requestBody.MovementPattern.Path),
	}

	if err := s.dbClient.MovementPattern().Create(ctx, movementPattern); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create movement pattern: " + err.Error()})
		return
	}

	// Then create the character with geometry auto-populated from movement pattern
	character := &models.Character{
		Name:              requestBody.Name,
		Description:       requestBody.Description,
		MapIconURL:        requestBody.MapIconUrl,
		DialogueImageURL:  requestBody.DialogueImageUrl,
		ThumbnailURL:      requestBody.ThumbnailUrl,
		PointOfInterestID: requestBody.PointOfInterestID,
		MovementPatternID: movementPattern.ID,
		MovementPattern:   *movementPattern,
		ImageGenerationStatus: func() string {
			if strings.TrimSpace(requestBody.DialogueImageUrl) != "" {
				return models.CharacterImageGenerationStatusComplete
			}
			return models.CharacterImageGenerationStatusNone
		}(),
	}

	if err := s.dbClient.Character().Create(ctx, character); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create character: " + err.Error()})
		return
	}

	if strings.TrimSpace(character.ThumbnailURL) == "" && strings.TrimSpace(character.DialogueImageURL) != "" {
		s.enqueueThumbnailTask(jobs.ThumbnailEntityCharacter, character.ID, character.DialogueImageURL)
	}

	ctx.JSON(http.StatusCreated, character)
}

func (s *server) generateCharacter(ctx *gin.Context) {
	var requestBody struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description"`
		MovementPattern *struct {
			MovementPatternType models.MovementPatternType `json:"movementPatternType"`
			ZoneID              *uuid.UUID                 `json:"zoneId"`
			StartingLatitude    float64                    `json:"startingLatitude"`
			StartingLongitude   float64                    `json:"startingLongitude"`
			Path                []models.Location          `json:"path"`
		} `json:"movementPattern"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movementType := models.MovementPatternStatic
	var zoneID *uuid.UUID
	var startingLatitude float64
	var startingLongitude float64
	var path []models.Location
	if requestBody.MovementPattern != nil {
		if requestBody.MovementPattern.MovementPatternType != "" {
			movementType = requestBody.MovementPattern.MovementPatternType
		}
		zoneID = requestBody.MovementPattern.ZoneID
		startingLatitude = requestBody.MovementPattern.StartingLatitude
		startingLongitude = requestBody.MovementPattern.StartingLongitude
		path = requestBody.MovementPattern.Path
	}

	movementPattern := &models.MovementPattern{
		MovementPatternType: movementType,
		ZoneID:              zoneID,
		StartingLatitude:    startingLatitude,
		StartingLongitude:   startingLongitude,
		Path:                models.LocationPath(path),
	}

	if err := s.dbClient.MovementPattern().Create(ctx, movementPattern); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create movement pattern: " + err.Error()})
		return
	}

	character := &models.Character{
		Name:                  requestBody.Name,
		Description:           requestBody.Description,
		MovementPatternID:     movementPattern.ID,
		MovementPattern:       *movementPattern,
		ImageGenerationStatus: models.CharacterImageGenerationStatusQueued,
	}

	if err := s.dbClient.Character().Create(ctx, character); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create character: " + err.Error()})
		return
	}

	payload := jobs.GenerateCharacterImageTaskPayload{
		CharacterID: character.ID,
		Name:        requestBody.Name,
		Description: requestBody.Description,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateCharacterImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		update := &models.Character{
			ImageGenerationStatus: models.CharacterImageGenerationStatusFailed,
			ImageGenerationError:  &errMsg,
		}
		_ = s.dbClient.Character().Update(ctx, character.ID, update)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, character)
}

func (s *server) regenerateCharacterImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	character, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if character == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	clearErr := ""
	update := &models.Character{
		ImageGenerationStatus: models.CharacterImageGenerationStatusQueued,
		ImageGenerationError:  &clearErr,
	}
	if err := s.dbClient.Character().Update(ctx, id, update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update character: " + err.Error()})
		return
	}

	payload := jobs.GenerateCharacterImageTaskPayload{
		CharacterID: character.ID,
		Name:        character.Name,
		Description: character.Description,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateCharacterImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		failUpdate := &models.Character{
			ImageGenerationStatus: models.CharacterImageGenerationStatusFailed,
			ImageGenerationError:  &errMsg,
		}
		_ = s.dbClient.Character().Update(ctx, id, failUpdate)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedCharacter, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updatedCharacter)
}

func (s *server) updateCharacter(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	// Check if character exists
	existingCharacter, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingCharacter == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	var requestBody struct {
		Name              string     `json:"name"`
		Description       string     `json:"description"`
		MapIconUrl        string     `json:"mapIconUrl"`
		DialogueImageUrl  string     `json:"dialogueImageUrl"`
		ThumbnailUrl      string     `json:"thumbnailUrl"`
		PointOfInterestID *uuid.UUID `json:"pointOfInterestId"`
		MovementPattern   struct {
			MovementPatternType models.MovementPatternType `json:"movementPatternType"`
			ZoneID              *uuid.UUID                 `json:"zoneId"`
			StartingLatitude    float64                    `json:"startingLatitude"`
			StartingLongitude   float64                    `json:"startingLongitude"`
			Path                []models.Location          `json:"path"`
		} `json:"movementPattern"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the movement pattern first
	movementPatternUpdates := &models.MovementPattern{
		MovementPatternType: requestBody.MovementPattern.MovementPatternType,
		ZoneID:              requestBody.MovementPattern.ZoneID,
		StartingLatitude:    requestBody.MovementPattern.StartingLatitude,
		StartingLongitude:   requestBody.MovementPattern.StartingLongitude,
		Path:                models.LocationPath(requestBody.MovementPattern.Path),
	}

	if err := s.dbClient.MovementPattern().Update(ctx, existingCharacter.MovementPatternID, movementPatternUpdates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update movement pattern: " + err.Error()})
		return
	}

	// Update the character
	characterUpdates := &models.Character{
		Name:             requestBody.Name,
		Description:      requestBody.Description,
		MapIconURL:       requestBody.MapIconUrl,
		DialogueImageURL: requestBody.DialogueImageUrl,
		ThumbnailURL:     requestBody.ThumbnailUrl,
	}
	if requestBody.PointOfInterestID != nil {
		_, err := s.dbClient.PointOfInterest().FindByID(ctx, *requestBody.PointOfInterestID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "point of interest not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load point of interest: " + err.Error()})
			return
		}
		characterUpdates.PointOfInterestID = requestBody.PointOfInterestID
	}

	if err := s.dbClient.Character().Update(ctx, id, characterUpdates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update character: " + err.Error()})
		return
	}

	if strings.TrimSpace(requestBody.DialogueImageUrl) != "" &&
		(strings.TrimSpace(requestBody.ThumbnailUrl) == "" ||
			strings.TrimSpace(existingCharacter.ThumbnailURL) == "" ||
			requestBody.DialogueImageUrl != existingCharacter.DialogueImageURL) {
		s.enqueueThumbnailTask(jobs.ThumbnailEntityCharacter, id, requestBody.DialogueImageUrl)
	}

	// Fetch the updated character with movement pattern
	updatedCharacter, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedCharacter)
}

func (s *server) updateCharacterLocations(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	var requestBody struct {
		Locations []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"locations"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	locations := make([]models.CharacterLocation, 0, len(requestBody.Locations))
	for _, loc := range requestBody.Locations {
		locations = append(locations, models.CharacterLocation{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	if err := s.dbClient.CharacterLocation().ReplaceForCharacter(ctx, id, locations); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "character locations updated successfully"})
}

func (s *server) deleteCharacter(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	// Check if character exists and get its movement pattern ID
	existingCharacter, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingCharacter == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	// Delete the character first (due to foreign key constraint)
	if err := s.dbClient.Character().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete character: " + err.Error()})
		return
	}

	// Then delete the movement pattern
	if err := s.dbClient.MovementPattern().Delete(ctx, existingCharacter.MovementPatternID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete movement pattern: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "character deleted successfully"})
}

// Character action handlers
func (s *server) getCharacterActions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}
	log.Printf("getCharacterActions: characterId=%s", id.String())

	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	quests, err := s.dbClient.Quest().FindByQuestGiverCharacterID(ctx, id)
	if err == nil {
		log.Printf("getCharacterActions: found %d quests for characterId=%s", len(quests), id.String())
		for _, quest := range quests {
			_ = s.ensureQuestActionForCharacter(ctx, quest.ID, id)
		}
	}

	actions, err := s.dbClient.CharacterAction().FindByCharacterID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("getCharacterActions: found %d actions for characterId=%s", len(actions), id.String())

	acceptedV2 := map[uuid.UUID]models.QuestAcceptanceV2{}
	if acceptances, accErr := s.dbClient.QuestAcceptanceV2().FindByUserID(ctx, user.ID); accErr == nil {
		for _, acc := range acceptances {
			acceptedV2[acc.QuestID] = acc
		}
	}
	acceptedLegacy := map[uuid.UUID]struct{}{}
	if legacy, accErr := s.dbClient.QuestAcceptance().FindByUserID(ctx, user.ID); accErr == nil {
		for _, acc := range legacy {
			acceptedLegacy[acc.PointOfInterestGroupID] = struct{}{}
		}
	}

	if len(acceptedV2) > 0 || len(acceptedLegacy) > 0 {
		filtered := actions[:0]
		for _, action := range actions {
			if action == nil || action.ActionType != models.ActionTypeGiveQuest {
				filtered = append(filtered, action)
				continue
			}
			if action.Metadata == nil {
				filtered = append(filtered, action)
				continue
			}
			questIDStr := extractActionQuestID(action.Metadata)
			if questIDStr == "" {
				filtered = append(filtered, action)
				continue
			}
			questID, err := uuid.Parse(questIDStr)
			if err != nil || questID == uuid.Nil {
				filtered = append(filtered, action)
				continue
			}
			if acc, exists := acceptedV2[questID]; exists {
				if acc.TurnedInAt != nil {
					log.Printf("getCharacterActions: hiding turned-in quest action=%s questId=%v userId=%s", action.ID, questID, user.ID)
					continue
				}
				filtered = append(filtered, action)
				continue
			}
			if _, exists := acceptedLegacy[questID]; exists {
				log.Printf("getCharacterActions: hiding legacy accepted quest action=%s questId=%v userId=%s", action.ID, questID, user.ID)
				continue
			}
			filtered = append(filtered, action)
		}
		actions = filtered
	}

	if len(actions) > 0 && len(quests) > 0 {
		questByID := make(map[string]*models.Quest, len(quests))
		for i := range quests {
			quest := quests[i]
			questByID[quest.ID.String()] = &quest
		}
		validStats := map[string]struct{}{
			"strength":     {},
			"dexterity":    {},
			"constitution": {},
			"intelligence": {},
			"wisdom":       {},
			"charisma":     {},
		}
		for _, action := range actions {
			if action == nil || action.ActionType != models.ActionTypeGiveQuest {
				continue
			}
			if action.Metadata == nil {
				action.Metadata = models.MetadataJSONB{}
			}
			questID, ok := action.Metadata["questId"]
			if !ok {
				log.Printf("getCharacterActions: giveQuest action=%s missing questId metadata", action.ID)
				continue
			}
			quest, ok := questByID[fmt.Sprint(questID)]
			if !ok || quest == nil {
				log.Printf("getCharacterActions: giveQuest action=%s questId=%v not found in questByID", action.ID, questID)
				continue
			}
			if quest.Name != "" {
				action.Metadata["questName"] = quest.Name
			}
			if quest.Description != "" {
				action.Metadata["questDescription"] = quest.Description
			}
			if len(quest.AcceptanceDialogue) > 0 {
				action.Metadata["acceptanceDialogue"] = quest.AcceptanceDialogue
			}
			var nodeAvgs []float64
			tags := map[string]struct{}{}
			for _, node := range quest.Nodes {
				if len(node.Challenges) == 0 {
					continue
				}
				sum := 0.0
				for _, challenge := range node.Challenges {
					sum += float64(challenge.Difficulty)
					for _, tag := range []string(challenge.StatTags) {
						normalized := strings.ToLower(strings.TrimSpace(tag))
						if normalized == "" {
							continue
						}
						if _, ok := validStats[normalized]; !ok {
							continue
						}
						tags[normalized] = struct{}{}
					}
				}
				nodeAvgs = append(nodeAvgs, sum/float64(len(node.Challenges)))
			}
			avgDifficulty := 0.0
			if len(nodeAvgs) > 0 {
				sum := 0.0
				for _, value := range nodeAvgs {
					sum += value
				}
				avgDifficulty = sum / float64(len(nodeAvgs))
			}
			tagList := make([]string, 0, len(tags))
			for tag := range tags {
				tagList = append(tagList, tag)
			}
			sort.Strings(tagList)
			action.Metadata["questAverageDifficulty"] = avgDifficulty
			action.Metadata["questStatTags"] = tagList
			log.Printf(
				"getCharacterActions: giveQuest action=%s questId=%s questName=%s acceptanceDialogue=%d",
				action.ID,
				quest.ID.String(),
				quest.Name,
				len(quest.AcceptanceDialogue),
			)
		}
	}

	ctx.JSON(http.StatusOK, actions)
}

func (s *server) getCharacterAction(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character action ID"})
		return
	}

	action, err := s.dbClient.CharacterAction().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if action == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character action not found"})
		return
	}

	ctx.JSON(http.StatusOK, action)
}

func (s *server) createCharacterAction(ctx *gin.Context) {
	var requestBody struct {
		CharacterID uuid.UUID                `json:"characterId" binding:"required"`
		ActionType  models.ActionType        `json:"actionType" binding:"required"`
		Dialogue    []models.DialogueMessage `json:"dialogue"`
		Metadata    map[string]interface{}   `json:"metadata"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify character exists
	character, err := s.dbClient.Character().FindByID(ctx, requestBody.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if character == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	action := &models.CharacterAction{
		CharacterID: requestBody.CharacterID,
		ActionType:  requestBody.ActionType,
		Dialogue:    models.DialogueSequence(requestBody.Dialogue),
		Metadata:    models.MetadataJSONB(requestBody.Metadata),
	}

	if err := s.dbClient.CharacterAction().Create(ctx, action); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, action)
}

func (s *server) updateCharacterAction(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character action ID"})
		return
	}

	// Verify action exists
	existingAction, err := s.dbClient.CharacterAction().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingAction == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character action not found"})
		return
	}

	var requestBody struct {
		ActionType models.ActionType        `json:"actionType"`
		Dialogue   []models.DialogueMessage `json:"dialogue"`
		Metadata   map[string]interface{}   `json:"metadata"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only update fields that are present in the request
	updates := &models.CharacterAction{}
	if requestBody.ActionType != "" {
		updates.ActionType = requestBody.ActionType
	}
	if requestBody.Dialogue != nil {
		updates.Dialogue = models.DialogueSequence(requestBody.Dialogue)
	}
	if requestBody.Metadata != nil {
		updates.Metadata = models.MetadataJSONB(requestBody.Metadata)
	}

	if err := s.dbClient.CharacterAction().Update(ctx, id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch updated action
	updatedAction, err := s.dbClient.CharacterAction().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedAction)
}

func (s *server) deleteCharacterAction(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character action ID"})
		return
	}

	// Verify action exists
	existingAction, err := s.dbClient.CharacterAction().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingAction == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character action not found"})
		return
	}

	if err := s.dbClient.CharacterAction().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "character action deleted successfully"})
}

func (s *server) purchaseFromShop(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	actionIDStr := ctx.Param("id")
	actionID, err := uuid.Parse(actionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character action ID"})
		return
	}

	// Verify action exists and is shop type
	action, err := s.dbClient.CharacterAction().FindByID(ctx, actionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if action == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character action not found"})
		return
	}

	if action.ActionType != models.ActionTypeShop {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "action is not a shop action"})
		return
	}

	// Parse request body
	var requestBody struct {
		ItemID   int `json:"itemId" binding:"required"`
		Quantity int `json:"quantity"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Quantity <= 0 {
		requestBody.Quantity = 1
	}

	// Extract inventory from metadata
	metadata := action.Metadata
	if metadata == nil {
		metadata = models.MetadataJSONB{}
	}

	inventoryInterface, ok := metadata["inventory"]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "shop has no inventory"})
		return
	}

	inventoryArray, ok := inventoryInterface.([]interface{})
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory format"})
		return
	}

	// Find item in inventory and get price
	var itemPrice float64
	var found bool
	for _, itemInterface := range inventoryArray {
		item, ok := itemInterface.(map[string]interface{})
		if !ok {
			continue
		}

		itemIDInterface, ok := item["itemId"]
		if !ok {
			continue
		}

		var itemID int
		switch v := itemIDInterface.(type) {
		case float64:
			itemID = int(v)
		case int:
			itemID = v
		default:
			continue
		}

		if itemID == requestBody.ItemID {
			priceInterface, ok := item["price"]
			if !ok {
				continue
			}

			switch v := priceInterface.(type) {
			case float64:
				itemPrice = v
			case int:
				itemPrice = float64(v)
			default:
				continue
			}

			found = true
			break
		}
	}

	if !found {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item not found in shop inventory"})
		return
	}

	// Calculate total price
	totalPrice := int(itemPrice) * requestBody.Quantity

	// Verify user has sufficient gold
	if user.Gold < totalPrice {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "insufficient gold"})
		return
	}

	// Use transaction to deduct gold and add item
	// Since we need transaction support, we'll use GORM's transaction
	// First, let's get the user again with lock to check gold one more time
	currentUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user: " + err.Error()})
		return
	}

	if currentUser.Gold < totalPrice {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "insufficient gold"})
		return
	}

	// Deduct gold
	if err := s.dbClient.User().SubtractGold(ctx, user.ID, totalPrice); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct gold: " + err.Error()})
		return
	}

	// Add item to inventory
	if err := s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(
		ctx,
		nil, // teamID
		&user.ID,
		requestBody.ItemID,
		requestBody.Quantity,
	); err != nil {
		// If adding item fails, we should try to refund gold
		// For now, just return error (in production, consider transaction or compensation)
		_ = s.dbClient.User().AddGold(ctx, user.ID, totalPrice)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to inventory: " + err.Error()})
		return
	}

	// Fetch and return updated user
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":       updatedUser,
		"itemId":     requestBody.ItemID,
		"quantity":   requestBody.Quantity,
		"totalPrice": totalPrice,
	})
}

func (s *server) sellToShop(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	actionIDStr := ctx.Param("id")
	actionID, err := uuid.Parse(actionIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character action ID"})
		return
	}

	// Verify action exists and is shop type
	action, err := s.dbClient.CharacterAction().FindByID(ctx, actionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if action == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character action not found"})
		return
	}

	if action.ActionType != models.ActionTypeShop {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "action is not a shop action"})
		return
	}

	// Parse request body
	var requestBody struct {
		ItemID   int `json:"itemId" binding:"required"`
		Quantity int `json:"quantity"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestBody.Quantity <= 0 {
		requestBody.Quantity = 1
	}

	// Get the inventory item to check sell value
	inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, requestBody.ItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find inventory item: " + err.Error()})
		return
	}

	if inventoryItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	// Check if item has a sell value
	if inventoryItem.SellValue == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "item cannot be sold"})
		return
	}

	// Calculate total sell value
	totalSellValue := *inventoryItem.SellValue * requestBody.Quantity

	// Verify user owns the item and has sufficient quantity
	ownedItems, err := s.dbClient.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{UserID: &user.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user items: " + err.Error()})
		return
	}

	var ownedItem *models.OwnedInventoryItem
	for _, item := range ownedItems {
		if item.InventoryItemID == requestBody.ItemID {
			ownedItem = &item
			break
		}
	}

	if ownedItem == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user does not own this item"})
		return
	}

	if ownedItem.Quantity < requestBody.Quantity {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "insufficient quantity"})
		return
	}

	// Decrement item quantity
	if err := s.dbClient.InventoryItem().DecrementUserInventoryItem(ctx, user.ID, requestBody.ItemID, requestBody.Quantity); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove item from inventory: " + err.Error()})
		return
	}

	// Add gold to user
	if err := s.dbClient.User().AddGold(ctx, user.ID, totalSellValue); err != nil {
		// If adding gold fails, try to restore the item
		// For now, just return error (in production, consider transaction or compensation)
		_ = s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(ctx, nil, &user.ID, requestBody.ItemID, requestBody.Quantity)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add gold: " + err.Error()})
		return
	}

	// Fetch and return updated user
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":           updatedUser,
		"itemId":         requestBody.ItemID,
		"quantity":       requestBody.Quantity,
		"totalSellValue": totalSellValue,
	})
}

func (s *server) getTreasureChests(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if err == nil {
		userID = &user.ID
	}

	treasureChests, openedMap, err := s.dbClient.TreasureChest().FindAllWithUserStatus(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add openedByUser field to each chest
	type TreasureChestResponse struct {
		models.TreasureChest
		OpenedByUser bool `json:"openedByUser"`
	}
	response := make([]TreasureChestResponse, len(treasureChests))
	for i, chest := range treasureChests {
		response[i] = TreasureChestResponse{
			TreasureChest: chest,
			OpenedByUser:  openedMap[chest.ID],
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) getTreasureChest(ctx *gin.Context) {
	id := ctx.Param("id")
	treasureChestID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid treasure chest ID"})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if err == nil {
		userID = &user.ID
	}

	treasureChest, openedByUser, err := s.dbClient.TreasureChest().FindByIDWithUserStatus(ctx, treasureChestID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if treasureChest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "treasure chest not found"})
		return
	}

	type TreasureChestResponse struct {
		models.TreasureChest
		OpenedByUser bool `json:"openedByUser"`
	}
	response := TreasureChestResponse{
		TreasureChest: *treasureChest,
		OpenedByUser:  openedByUser,
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) getTreasureChestsForZone(ctx *gin.Context) {
	id := ctx.Param("id")
	zoneID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	user, err := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if err == nil {
		userID = &user.ID
	}

	treasureChests, openedMap, err := s.dbClient.TreasureChest().FindByZoneIDWithUserStatus(ctx, zoneID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add openedByUser field to each chest
	type TreasureChestResponse struct {
		models.TreasureChest
		OpenedByUser bool `json:"openedByUser"`
	}
	response := make([]TreasureChestResponse, len(treasureChests))
	for i, chest := range treasureChests {
		response[i] = TreasureChestResponse{
			TreasureChest: chest,
			OpenedByUser:  openedMap[chest.ID],
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) createTreasureChest(ctx *gin.Context) {
	var requestBody struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		ZoneID    string  `json:"zoneId" binding:"required"`
		Gold      *int    `json:"gold"`
		Items     []struct {
			InventoryItemID int `json:"inventoryItemId"`
			Quantity        int `json:"quantity"`
		} `json:"items"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(requestBody.ZoneID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	treasureChest := &models.TreasureChest{
		Latitude:  requestBody.Latitude,
		Longitude: requestBody.Longitude,
		ZoneID:    zoneID,
		Gold:      requestBody.Gold,
	}

	if err := s.dbClient.TreasureChest().Create(ctx, treasureChest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create treasure chest: " + err.Error()})
		return
	}

	// Add items if provided
	for _, item := range requestBody.Items {
		if err := s.dbClient.TreasureChest().AddItem(ctx, treasureChest.ID, item.InventoryItemID, item.Quantity); err != nil {
			// Log error but don't fail the request
			// In production, you might want to rollback the treasure chest creation
			continue
		}
	}

	// Fetch the created treasure chest with items
	createdChest, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChest.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch created treasure chest: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdChest)
}

func (s *server) updateTreasureChest(ctx *gin.Context) {
	id := ctx.Param("id")
	treasureChestID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid treasure chest ID"})
		return
	}

	// Check if treasure chest exists
	existingChest, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingChest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "treasure chest not found"})
		return
	}

	var requestBody struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		ZoneID    *string  `json:"zoneId"`
		Gold      *int     `json:"gold"`
		Items     []struct {
			InventoryItemID int `json:"inventoryItemId"`
			Quantity        int `json:"quantity"`
		} `json:"items"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := &models.TreasureChest{
		ID: treasureChestID,
	}

	if requestBody.Latitude != nil && requestBody.Longitude != nil {
		updates.Latitude = *requestBody.Latitude
		updates.Longitude = *requestBody.Longitude
	} else {
		updates.Latitude = existingChest.Latitude
		updates.Longitude = existingChest.Longitude
	}

	if requestBody.ZoneID != nil {
		zoneID, err := uuid.Parse(*requestBody.ZoneID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
			return
		}
		updates.ZoneID = zoneID
	} else {
		updates.ZoneID = existingChest.ZoneID
	}

	if requestBody.Gold != nil {
		updates.Gold = requestBody.Gold
	} else {
		updates.Gold = existingChest.Gold
	}

	if err := s.dbClient.TreasureChest().Update(ctx, treasureChestID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update treasure chest: " + err.Error()})
		return
	}

	// Update items if provided
	if requestBody.Items != nil {
		// Remove all existing items
		existingChest, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
		if err == nil && existingChest != nil {
			for _, item := range existingChest.Items {
				_ = s.dbClient.TreasureChest().RemoveItem(ctx, treasureChestID, item.InventoryItemID)
			}
		}

		// Add new items
		for _, item := range requestBody.Items {
			if err := s.dbClient.TreasureChest().AddItem(ctx, treasureChestID, item.InventoryItemID, item.Quantity); err != nil {
				// Log error but continue
				continue
			}
		}
	}

	// Fetch the updated treasure chest
	updatedChest, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated treasure chest: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedChest)
}

func (s *server) deleteTreasureChest(ctx *gin.Context) {
	id := ctx.Param("id")
	treasureChestID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid treasure chest ID"})
		return
	}

	// Check if treasure chest exists
	existingChest, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingChest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "treasure chest not found"})
		return
	}

	if err := s.dbClient.TreasureChest().Delete(ctx, treasureChestID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete treasure chest: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "treasure chest deleted successfully"})
}

func (s *server) openTreasureChest(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	id := ctx.Param("id")
	treasureChestID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid treasure chest ID"})
		return
	}

	// Get treasure chest
	treasureChest, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if treasureChest == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "treasure chest not found"})
		return
	}

	// Check if user already opened this chest
	hasOpened, err := s.dbClient.TreasureChest().HasUserOpenedChest(ctx, user.ID, treasureChestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasOpened {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "treasure chest already opened"})
		return
	}

	// Validate user proximity (10 meters)
	locationStr, err := s.livenessClient.GetUserLocation(ctx, user.ID)
	if err != nil || locationStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user location not available"})
		return
	}

	parts := strings.Split(locationStr, ",")
	if len(parts) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid location format"})
		return
	}

	userLat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude in user location"})
		return
	}

	userLng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude in user location"})
		return
	}

	distance := util.HaversineDistance(userLat, userLng, treasureChest.Latitude, treasureChest.Longitude)
	if distance > 10 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you must be within 10 meters of the treasure chest. Currently %.0f meters away", distance)})
		return
	}

	// Check if chest is locked
	var unlockItemID *uuid.UUID
	if treasureChest.UnlockTier != nil {
		// Find user's owned inventory items with unlock tier
		ownedItems, err := s.dbClient.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{UserID: &user.ID})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user inventory"})
			return
		}

		// Find inventory items with unlock tier >= chest unlock tier
		var validUnlockItems []models.OwnedInventoryItem
		for _, ownedItem := range ownedItems {
			if ownedItem.Quantity > 0 {
				// Get the inventory item to check unlock tier
				inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, ownedItem.InventoryItemID)
				if err == nil && inventoryItem != nil && inventoryItem.UnlockTier != nil {
					if *inventoryItem.UnlockTier >= *treasureChest.UnlockTier {
						validUnlockItems = append(validUnlockItems, ownedItem)
					}
				}
			}
		}

		if len(validUnlockItems) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "you do not have an item with sufficient unlock tier to open this chest"})
			return
		}

		// Find the item with the lowest unlock tier
		var lowestTierItem *models.OwnedInventoryItem
		var lowestTier int
		for i := range validUnlockItems {
			inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, validUnlockItems[i].InventoryItemID)
			if err == nil && inventoryItem != nil && inventoryItem.UnlockTier != nil {
				if lowestTierItem == nil || *inventoryItem.UnlockTier < lowestTier {
					lowestTier = *inventoryItem.UnlockTier
					lowestTierItem = &validUnlockItems[i]
				}
			}
		}

		if lowestTierItem == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not find valid unlock item"})
			return
		}

		unlockItemID = &lowestTierItem.ID
	}

	// Consume unlock item if needed
	if unlockItemID != nil {
		if err := s.dbClient.InventoryItem().UseInventoryItem(ctx, *unlockItemID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to consume unlock item: " + err.Error()})
			return
		}
	}

	// Give gold if chest has gold
	if treasureChest.Gold != nil && *treasureChest.Gold > 0 {
		if err := s.dbClient.User().AddGold(ctx, user.ID, *treasureChest.Gold); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add gold: " + err.Error()})
			return
		}
	}

	// Give items
	for _, chestItem := range treasureChest.Items {
		if err := s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(ctx, nil, &user.ID, chestItem.InventoryItemID, chestItem.Quantity); err != nil {
			// Log error but continue
			continue
		}
	}

	// Record opening
	opening := &models.UserTreasureChestOpening{
		UserID:          user.ID,
		TreasureChestID: treasureChestID,
	}
	if err := s.dbClient.TreasureChest().CreateUserTreasureChestOpening(ctx, opening); err != nil {
		// Log error but don't fail the request
	}

	// Fetch updated user
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "treasure chest opened successfully",
		"user":    updatedUser,
	})
}

var scenarioValidStatTags = map[string]struct{}{
	"strength":     {},
	"dexterity":    {},
	"constitution": {},
	"intelligence": {},
	"wisdom":       {},
	"charisma":     {},
}

type scenarioRewardItemPayload struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type scenarioRewardSpellPayload struct {
	SpellID string `json:"spellId"`
}

type scenarioOptionPayload struct {
	OptionText                string                         `json:"optionText"`
	SuccessText               string                         `json:"successText"`
	FailureText               string                         `json:"failureText"`
	StatTag                   string                         `json:"statTag"`
	Proficiencies             []string                       `json:"proficiencies"`
	Difficulty                *int                           `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience"`
	RewardGold                int                            `json:"rewardGold"`
	FailureHealthDrainType    string                         `json:"failureHealthDrainType"`
	FailureHealthDrainValue   int                            `json:"failureHealthDrainValue"`
	FailureManaDrainType      string                         `json:"failureManaDrainType"`
	FailureManaDrainValue     int                            `json:"failureManaDrainValue"`
	FailureStatuses           []scenarioFailureStatusPayload `json:"failureStatuses"`
	SuccessHealthRestoreType  string                         `json:"successHealthRestoreType"`
	SuccessHealthRestoreValue int                            `json:"successHealthRestoreValue"`
	SuccessManaRestoreType    string                         `json:"successManaRestoreType"`
	SuccessManaRestoreValue   int                            `json:"successManaRestoreValue"`
	SuccessStatuses           []scenarioFailureStatusPayload `json:"successStatuses"`
	ItemRewards               []scenarioRewardItemPayload    `json:"itemRewards"`
	SpellRewards              []scenarioRewardSpellPayload   `json:"spellRewards"`
}

type scenarioFailureStatusPayload struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Effect          string `json:"effect"`
	EffectType      string `json:"effectType"`
	Positive        *bool  `json:"positive"`
	DamagePerTick   int    `json:"damagePerTick"`
	DurationSeconds int    `json:"durationSeconds"`
	StrengthMod     int    `json:"strengthMod"`
	DexterityMod    int    `json:"dexterityMod"`
	ConstitutionMod int    `json:"constitutionMod"`
	IntelligenceMod int    `json:"intelligenceMod"`
	WisdomMod       int    `json:"wisdomMod"`
	CharismaMod     int    `json:"charismaMod"`
}

type scenarioUpsertRequest struct {
	ZoneID                    string                         `json:"zoneId"`
	Latitude                  float64                        `json:"latitude"`
	Longitude                 float64                        `json:"longitude"`
	Prompt                    string                         `json:"prompt"`
	ImageURL                  string                         `json:"imageUrl"`
	ThumbnailURL              string                         `json:"thumbnailUrl"`
	Difficulty                *int                           `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience"`
	RewardGold                int                            `json:"rewardGold"`
	OpenEnded                 bool                           `json:"openEnded"`
	FailurePenaltyMode        string                         `json:"failurePenaltyMode"`
	FailureHealthDrainType    string                         `json:"failureHealthDrainType"`
	FailureHealthDrainValue   int                            `json:"failureHealthDrainValue"`
	FailureManaDrainType      string                         `json:"failureManaDrainType"`
	FailureManaDrainValue     int                            `json:"failureManaDrainValue"`
	FailureStatuses           []scenarioFailureStatusPayload `json:"failureStatuses"`
	SuccessRewardMode         string                         `json:"successRewardMode"`
	SuccessHealthRestoreType  string                         `json:"successHealthRestoreType"`
	SuccessHealthRestoreValue int                            `json:"successHealthRestoreValue"`
	SuccessManaRestoreType    string                         `json:"successManaRestoreType"`
	SuccessManaRestoreValue   int                            `json:"successManaRestoreValue"`
	SuccessStatuses           []scenarioFailureStatusPayload `json:"successStatuses"`
	Options                   []scenarioOptionPayload        `json:"options"`
	ItemRewards               []scenarioRewardItemPayload    `json:"itemRewards"`
	SpellRewards              []scenarioRewardSpellPayload   `json:"spellRewards"`
}

type scenarioGenerationJobRequest struct {
	ZoneID    string   `json:"zoneId"`
	OpenEnded bool     `json:"openEnded"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type scenarioWithUserStatus struct {
	models.Scenario
	AttemptedByUser bool `json:"attemptedByUser"`
}

type scenarioPerformRequest struct {
	ScenarioOptionID *uuid.UUID `json:"scenarioOptionId"`
	ResponseText     string     `json:"responseText"`
}

type scenarioPerformResponse struct {
	Successful             bool                           `json:"successful"`
	Reason                 string                         `json:"reason"`
	OutcomeText            string                         `json:"outcomeText"`
	ScenarioID             uuid.UUID                      `json:"scenarioId"`
	ScenarioOptionID       *uuid.UUID                     `json:"scenarioOptionId,omitempty"`
	Roll                   int                            `json:"roll"`
	StatTag                string                         `json:"statTag"`
	StatValue              int                            `json:"statValue"`
	Proficiencies          []string                       `json:"proficiencies"`
	ProficiencyBonus       int                            `json:"proficiencyBonus"`
	CreativityBonus        int                            `json:"creativityBonus"`
	Threshold              int                            `json:"threshold"`
	TotalScore             int                            `json:"totalScore"`
	FailureHealthDrained   int                            `json:"failureHealthDrained"`
	FailureManaDrained     int                            `json:"failureManaDrained"`
	FailureStatusesApplied []scenarioAppliedFailureStatus `json:"failureStatusesApplied"`
	SuccessHealthRestored  int                            `json:"successHealthRestored"`
	SuccessManaRestored    int                            `json:"successManaRestored"`
	SuccessStatusesApplied []scenarioAppliedFailureStatus `json:"successStatusesApplied"`
	RewardExperience       int                            `json:"rewardExperience"`
	RewardGold             int                            `json:"rewardGold"`
	ItemsAwarded           []models.ItemAwarded           `json:"itemsAwarded"`
	SpellsAwarded          []models.SpellAwarded          `json:"spellsAwarded"`
}

type scenarioFreeformAssessment struct {
	StatTag         string   `json:"statTag"`
	Proficiencies   []string `json:"proficiencies"`
	CreativityBonus int      `json:"creativityBonus"`
	Reasoning       string   `json:"reasoning"`
	SuccessText     string   `json:"successText"`
	FailureText     string   `json:"failureText"`
}

const scenarioFreeformAssessmentPromptTemplate = `
You are a tabletop RPG game master evaluating how a player responds to a fantasy scenario.

Scenario:
%s

Player response:
%s

Pick the best primary DnD stat this response uses:
- strength
- dexterity
- constitution
- intelligence
- wisdom
- charisma

Pick 0 to 3 short proficiencies (1-3 words each) this response demonstrates.
Assign a creativityBonus integer from 0 to 10:
- 0 means generic or weak
- 10 means unusually creative and compelling

Return JSON only:
{
  "statTag": "strength|dexterity|constitution|intelligence|wisdom|charisma",
  "proficiencies": ["string"],
  "creativityBonus": 0,
  "reasoning": "short string",
  "successText": "one or two short sentences describing what success looks like for this response",
  "failureText": "one or two short sentences describing what failure looks like for this response"
}
`

type scenarioRewardItem struct {
	InventoryItemID int
	Quantity        int
}

type scenarioRewardSpell struct {
	SpellID uuid.UUID
}

type scenarioFailurePenalty struct {
	HealthDrainType  models.ScenarioFailureDrainType
	HealthDrainValue int
	ManaDrainType    models.ScenarioFailureDrainType
	ManaDrainValue   int
	Statuses         models.ScenarioFailureStatusTemplates
}

type scenarioSuccessReward struct {
	HealthRestoreType  models.ScenarioFailureDrainType
	HealthRestoreValue int
	ManaRestoreType    models.ScenarioFailureDrainType
	ManaRestoreValue   int
	Statuses           models.ScenarioFailureStatusTemplates
}

type scenarioAppliedFailureStatus struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Effect          string `json:"effect"`
	EffectType      string `json:"effectType"`
	Positive        bool   `json:"positive"`
	DamagePerTick   int    `json:"damagePerTick"`
	DurationSeconds int    `json:"durationSeconds"`
}

type scenarioAppliedFailurePenalty struct {
	HealthDrained int
	ManaDrained   int
	Statuses      []scenarioAppliedFailureStatus
}

type scenarioAppliedSuccessReward struct {
	HealthRestored int
	ManaRestored   int
	Statuses       []scenarioAppliedFailureStatus
}

func normalizeScenarioStatTag(raw string) (string, bool) {
	tag := strings.ToLower(strings.TrimSpace(raw))
	_, ok := scenarioValidStatTags[tag]
	return tag, ok
}

func normalizeScenarioFailurePenaltyMode(raw string, openEnded bool) (models.ScenarioFailurePenaltyMode, error) {
	if openEnded {
		return models.ScenarioFailurePenaltyModeShared, nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(models.ScenarioFailurePenaltyModeShared):
		return models.ScenarioFailurePenaltyModeShared, nil
	case string(models.ScenarioFailurePenaltyModeIndividual):
		return models.ScenarioFailurePenaltyModeIndividual, nil
	default:
		return "", fmt.Errorf("invalid failurePenaltyMode")
	}
}

func normalizeScenarioSuccessRewardMode(raw string, openEnded bool) (models.ScenarioSuccessRewardMode, error) {
	if openEnded {
		return models.ScenarioSuccessRewardModeShared, nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(models.ScenarioSuccessRewardModeShared):
		return models.ScenarioSuccessRewardModeShared, nil
	case string(models.ScenarioSuccessRewardModeIndividual):
		return models.ScenarioSuccessRewardModeIndividual, nil
	default:
		return "", fmt.Errorf("invalid successRewardMode")
	}
}

func normalizeScenarioFailureDrainType(raw string) (models.ScenarioFailureDrainType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(models.ScenarioFailureDrainTypeNone):
		return models.ScenarioFailureDrainTypeNone, nil
	case string(models.ScenarioFailureDrainTypeFlat):
		return models.ScenarioFailureDrainTypeFlat, nil
	case string(models.ScenarioFailureDrainTypePercent):
		return models.ScenarioFailureDrainTypePercent, nil
	default:
		return "", fmt.Errorf("invalid failure drain type")
	}
}

func normalizeScenarioFailureDrainValue(
	drainType models.ScenarioFailureDrainType,
	value int,
	fieldName string,
) (int, error) {
	if drainType == models.ScenarioFailureDrainTypeNone {
		return 0, nil
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must be zero or greater", fieldName)
	}
	if drainType == models.ScenarioFailureDrainTypePercent && value > 100 {
		return 0, fmt.Errorf("%s percent must be 100 or less", fieldName)
	}
	return value, nil
}

func parseScenarioFailureStatusTemplates(
	input []scenarioFailureStatusPayload,
	fieldName string,
) (models.ScenarioFailureStatusTemplates, error) {
	templates := make(models.ScenarioFailureStatusTemplates, 0, len(input))
	for idx, status := range input {
		name := strings.TrimSpace(status.Name)
		if name == "" {
			return nil, fmt.Errorf("%s[%d].name is required", fieldName, idx)
		}
		if status.DurationSeconds <= 0 {
			return nil, fmt.Errorf("%s[%d].durationSeconds must be > 0", fieldName, idx)
		}
		effectType := normalizeUserStatusEffectType(status.EffectType)
		if effectType == models.UserStatusEffectTypeDamageOverTime && status.DamagePerTick <= 0 {
			return nil, fmt.Errorf("%s[%d].damagePerTick must be > 0 for damage_over_time statuses", fieldName, idx)
		}
		positive := true
		if status.Positive != nil {
			positive = *status.Positive
		}
		templates = append(templates, models.ScenarioFailureStatusTemplate{
			Name:            name,
			Description:     strings.TrimSpace(status.Description),
			Effect:          strings.TrimSpace(status.Effect),
			EffectType:      string(effectType),
			Positive:        positive,
			DamagePerTick:   status.DamagePerTick,
			DurationSeconds: status.DurationSeconds,
			StrengthMod:     status.StrengthMod,
			DexterityMod:    status.DexterityMod,
			ConstitutionMod: status.ConstitutionMod,
			IntelligenceMod: status.IntelligenceMod,
			WisdomMod:       status.WisdomMod,
			CharismaMod:     status.CharismaMod,
		})
	}
	return templates, nil
}

func parseInventoryConsumeStatusNames(input []string) models.StringArray {
	statusNames := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, rawName := range input {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		statusNames = append(statusNames, name)
	}
	return statusNames
}

func parseInventoryConsumeSpellIDs(input []string) (models.StringArray, error) {
	spellIDs := make(models.StringArray, 0, len(input))
	seen := map[uuid.UUID]struct{}{}
	for idx, rawID := range input {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}
		spellID, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, fmt.Errorf("consumeSpellIds[%d] must be a valid UUID", idx)
		}
		if _, exists := seen[spellID]; exists {
			continue
		}
		seen[spellID] = struct{}{}
		spellIDs = append(spellIDs, spellID.String())
	}
	return spellIDs, nil
}

func scenarioFailurePenaltyFromScenario(scenario *models.Scenario) scenarioFailurePenalty {
	if scenario == nil {
		return scenarioFailurePenalty{
			HealthDrainType: models.ScenarioFailureDrainTypeNone,
			ManaDrainType:   models.ScenarioFailureDrainTypeNone,
			Statuses:        models.ScenarioFailureStatusTemplates{},
		}
	}
	statuses := scenario.FailureStatuses
	if statuses == nil {
		statuses = models.ScenarioFailureStatusTemplates{}
	}
	return scenarioFailurePenalty{
		HealthDrainType:  scenario.FailureHealthDrainType,
		HealthDrainValue: scenario.FailureHealthDrainValue,
		ManaDrainType:    scenario.FailureManaDrainType,
		ManaDrainValue:   scenario.FailureManaDrainValue,
		Statuses:         statuses,
	}
}

func scenarioFailurePenaltyFromOption(option *models.ScenarioOption) scenarioFailurePenalty {
	if option == nil {
		return scenarioFailurePenalty{
			HealthDrainType: models.ScenarioFailureDrainTypeNone,
			ManaDrainType:   models.ScenarioFailureDrainTypeNone,
			Statuses:        models.ScenarioFailureStatusTemplates{},
		}
	}
	statuses := option.FailureStatuses
	if statuses == nil {
		statuses = models.ScenarioFailureStatusTemplates{}
	}
	return scenarioFailurePenalty{
		HealthDrainType:  option.FailureHealthDrainType,
		HealthDrainValue: option.FailureHealthDrainValue,
		ManaDrainType:    option.FailureManaDrainType,
		ManaDrainValue:   option.FailureManaDrainValue,
		Statuses:         statuses,
	}
}

func scenarioSuccessRewardFromScenario(scenario *models.Scenario) scenarioSuccessReward {
	if scenario == nil {
		return scenarioSuccessReward{
			HealthRestoreType: models.ScenarioFailureDrainTypeNone,
			ManaRestoreType:   models.ScenarioFailureDrainTypeNone,
			Statuses:          models.ScenarioFailureStatusTemplates{},
		}
	}
	statuses := scenario.SuccessStatuses
	if statuses == nil {
		statuses = models.ScenarioFailureStatusTemplates{}
	}
	return scenarioSuccessReward{
		HealthRestoreType:  scenario.SuccessHealthRestoreType,
		HealthRestoreValue: scenario.SuccessHealthRestoreValue,
		ManaRestoreType:    scenario.SuccessManaRestoreType,
		ManaRestoreValue:   scenario.SuccessManaRestoreValue,
		Statuses:           statuses,
	}
}

func scenarioSuccessRewardFromOption(option *models.ScenarioOption) scenarioSuccessReward {
	if option == nil {
		return scenarioSuccessReward{
			HealthRestoreType: models.ScenarioFailureDrainTypeNone,
			ManaRestoreType:   models.ScenarioFailureDrainTypeNone,
			Statuses:          models.ScenarioFailureStatusTemplates{},
		}
	}
	statuses := option.SuccessStatuses
	if statuses == nil {
		statuses = models.ScenarioFailureStatusTemplates{}
	}
	return scenarioSuccessReward{
		HealthRestoreType:  option.SuccessHealthRestoreType,
		HealthRestoreValue: option.SuccessHealthRestoreValue,
		ManaRestoreType:    option.SuccessManaRestoreType,
		ManaRestoreValue:   option.SuccessManaRestoreValue,
		Statuses:           statuses,
	}
}

func resolveScenarioFailurePenalty(scenario *models.Scenario, selectedOption *models.ScenarioOption) scenarioFailurePenalty {
	if scenario == nil {
		return scenarioFailurePenaltyFromScenario(nil)
	}
	if scenario.OpenEnded {
		return scenarioFailurePenaltyFromScenario(scenario)
	}
	if scenario.FailurePenaltyMode == models.ScenarioFailurePenaltyModeIndividual {
		return scenarioFailurePenaltyFromOption(selectedOption)
	}
	return scenarioFailurePenaltyFromScenario(scenario)
}

func resolveScenarioSuccessReward(scenario *models.Scenario, selectedOption *models.ScenarioOption) scenarioSuccessReward {
	if scenario == nil {
		return scenarioSuccessRewardFromScenario(nil)
	}
	if scenario.OpenEnded {
		return scenarioSuccessRewardFromScenario(scenario)
	}
	if scenario.SuccessRewardMode == models.ScenarioSuccessRewardModeIndividual {
		return scenarioSuccessRewardFromOption(selectedOption)
	}
	return scenarioSuccessRewardFromScenario(scenario)
}

func scenarioDrainAmount(drainType models.ScenarioFailureDrainType, value int, maxValue int) int {
	if value <= 0 || maxValue <= 0 {
		return 0
	}
	switch drainType {
	case models.ScenarioFailureDrainTypeFlat:
		return value
	case models.ScenarioFailureDrainTypePercent:
		drain := (maxValue * value) / 100
		if drain == 0 {
			drain = 1
		}
		return drain
	default:
		return 0
	}
}

func normalizeScenarioProficiencies(input []string) []string {
	normalized := []string{}
	seen := map[string]struct{}{}
	for _, candidate := range input {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func sanitizeScenarioOutcomeText(input string, fallback string) string {
	text := strings.TrimSpace(input)
	if text == "" {
		text = strings.TrimSpace(fallback)
	}
	if len(text) > 320 {
		text = strings.TrimSpace(text[:320])
	}
	if text == "" {
		return "The outcome resolves."
	}
	return text
}

func normalizeScenarioThumbnailURL(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("thumbnailUrl is required")
	}
	return trimmed, nil
}

func scenarioDifficultyValue(input *int) (int, error) {
	if input == nil {
		return scenarioDefaultDifficulty, nil
	}
	if *input < 0 {
		return 0, fmt.Errorf("difficulty must be zero or greater")
	}
	return *input, nil
}

func rollScenarioDie() (int, error) {
	value, err := crand.Int(crand.Reader, big.NewInt(scenarioRollSides))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()) + 1, nil
}

func extractLLMJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```JSON")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func (s *server) assessScenarioFreeform(ctx context.Context, scenarioPrompt, responseText string) (*scenarioFreeformAssessment, error) {
	prompt := fmt.Sprintf(scenarioFreeformAssessmentPromptTemplate, scenarioPrompt, responseText)
	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, err
	}

	assessment := &scenarioFreeformAssessment{}
	if err := json.Unmarshal([]byte(extractLLMJSONObject(answer.Answer)), assessment); err != nil {
		return nil, fmt.Errorf("failed to parse scenario assessment: %w", err)
	}
	if statTag, ok := normalizeScenarioStatTag(assessment.StatTag); ok {
		assessment.StatTag = statTag
	} else {
		assessment.StatTag = "charisma"
	}
	assessment.Proficiencies = normalizeScenarioProficiencies(assessment.Proficiencies)
	if len(assessment.Proficiencies) > 3 {
		assessment.Proficiencies = assessment.Proficiencies[:3]
	}
	if assessment.CreativityBonus < 0 {
		assessment.CreativityBonus = 0
	} else if assessment.CreativityBonus > 10 {
		assessment.CreativityBonus = 10
	}
	assessment.Reasoning = strings.TrimSpace(assessment.Reasoning)
	assessment.SuccessText = strings.TrimSpace(assessment.SuccessText)
	assessment.FailureText = strings.TrimSpace(assessment.FailureText)
	return assessment, nil
}

func (s *server) getScenarioStatAndProficiencyBonuses(ctx context.Context, userID uuid.UUID, statTag string, proficiencies []string) (int, int, error) {
	stats, err := s.dbClient.UserCharacterStats().FindOrCreateForUser(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	statusBonuses, err := s.dbClient.UserStatus().GetActiveStatBonuses(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	totalBonuses := equipmentBonuses.Add(statusBonuses)

	statValue := 0
	switch statTag {
	case "strength":
		statValue = stats.Strength + totalBonuses.Strength
	case "dexterity":
		statValue = stats.Dexterity + totalBonuses.Dexterity
	case "constitution":
		statValue = stats.Constitution + totalBonuses.Constitution
	case "intelligence":
		statValue = stats.Intelligence + totalBonuses.Intelligence
	case "wisdom":
		statValue = stats.Wisdom + totalBonuses.Wisdom
	case "charisma":
		statValue = stats.Charisma + totalBonuses.Charisma
	default:
		return 0, 0, fmt.Errorf("invalid stat tag")
	}

	rows, err := s.dbClient.UserProficiency().FindByUserID(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	levelByProficiency := map[string]int{}
	for _, row := range rows {
		key := strings.ToLower(strings.TrimSpace(row.Proficiency))
		if key == "" {
			continue
		}
		levelByProficiency[key] = row.Level
	}

	proficiencyBonus := 0
	for _, proficiency := range normalizeScenarioProficiencies(proficiencies) {
		level := levelByProficiency[strings.ToLower(proficiency)]
		if level <= 0 {
			continue
		}
		bonus := 1 + (level-1)/5
		if bonus > 6 {
			bonus = 6
		}
		proficiencyBonus += bonus
	}

	return statValue, proficiencyBonus, nil
}

func (s *server) getScenarioResourceState(
	ctx context.Context,
	userID uuid.UUID,
) (*models.UserCharacterStats, int, int, int, int, error) {
	if err := s.applyOutOfBattleUserDamageOverTime(ctx, userID); err != nil {
		return nil, 0, 0, 0, 0, err
	}
	stats, err := s.dbClient.UserCharacterStats().FindOrCreateForUser(ctx, userID)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	statusBonuses, err := s.dbClient.UserStatus().GetActiveStatBonuses(ctx, userID)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	totalBonuses := equipmentBonuses.Add(statusBonuses)
	maxHealth, maxMana, currentHealth, currentMana := deriveCharacterResources(
		stats,
		totalBonuses,
	)
	return stats, maxHealth, maxMana, currentHealth, currentMana, nil
}

func (s *server) applyScenarioFailurePenalty(
	ctx context.Context,
	userID uuid.UUID,
	penalty scenarioFailurePenalty,
) (scenarioAppliedFailurePenalty, error) {
	applied := scenarioAppliedFailurePenalty{
		Statuses: []scenarioAppliedFailureStatus{},
	}
	if penalty.HealthDrainType == "" {
		penalty.HealthDrainType = models.ScenarioFailureDrainTypeNone
	}
	if penalty.ManaDrainType == "" {
		penalty.ManaDrainType = models.ScenarioFailureDrainTypeNone
	}

	healthDrain := 0
	manaDrain := 0
	if penalty.HealthDrainType != models.ScenarioFailureDrainTypeNone || penalty.ManaDrainType != models.ScenarioFailureDrainTypeNone {
		_, maxHealth, maxMana, currentHealth, currentMana, err := s.getScenarioResourceState(ctx, userID)
		if err != nil {
			return applied, err
		}
		healthDrain = min(
			scenarioDrainAmount(penalty.HealthDrainType, penalty.HealthDrainValue, maxHealth),
			currentHealth,
		)
		manaDrain = min(
			scenarioDrainAmount(penalty.ManaDrainType, penalty.ManaDrainValue, maxMana),
			currentMana,
		)
	}
	if healthDrain > 0 || manaDrain > 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, healthDrain, manaDrain); err != nil {
			return applied, err
		}
	}
	applied.HealthDrained = healthDrain
	applied.ManaDrained = manaDrain

	now := time.Now()
	for _, statusTemplate := range penalty.Statuses {
		name := strings.TrimSpace(statusTemplate.Name)
		if name == "" || statusTemplate.DurationSeconds <= 0 {
			continue
		}
		status := &models.UserStatus{
			UserID:          userID,
			Name:            name,
			Description:     strings.TrimSpace(statusTemplate.Description),
			Effect:          strings.TrimSpace(statusTemplate.Effect),
			Positive:        statusTemplate.Positive,
			EffectType:      normalizeUserStatusEffectType(statusTemplate.EffectType),
			DamagePerTick:   statusTemplate.DamagePerTick,
			StrengthMod:     statusTemplate.StrengthMod,
			DexterityMod:    statusTemplate.DexterityMod,
			ConstitutionMod: statusTemplate.ConstitutionMod,
			IntelligenceMod: statusTemplate.IntelligenceMod,
			WisdomMod:       statusTemplate.WisdomMod,
			CharismaMod:     statusTemplate.CharismaMod,
			StartedAt:       now,
			ExpiresAt:       now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
		}
		if err := s.dbClient.UserStatus().Create(ctx, status); err != nil {
			return applied, err
		}
		applied.Statuses = append(applied.Statuses, scenarioAppliedFailureStatus{
			Name:            name,
			Description:     strings.TrimSpace(statusTemplate.Description),
			Effect:          strings.TrimSpace(statusTemplate.Effect),
			EffectType:      string(normalizeUserStatusEffectType(statusTemplate.EffectType)),
			Positive:        statusTemplate.Positive,
			DamagePerTick:   statusTemplate.DamagePerTick,
			DurationSeconds: statusTemplate.DurationSeconds,
		})
	}
	return applied, nil
}

func (s *server) applyScenarioSuccessReward(
	ctx context.Context,
	userID uuid.UUID,
	reward scenarioSuccessReward,
) (scenarioAppliedSuccessReward, error) {
	applied := scenarioAppliedSuccessReward{
		Statuses: []scenarioAppliedFailureStatus{},
	}
	if reward.HealthRestoreType == "" {
		reward.HealthRestoreType = models.ScenarioFailureDrainTypeNone
	}
	if reward.ManaRestoreType == "" {
		reward.ManaRestoreType = models.ScenarioFailureDrainTypeNone
	}

	healthRestore := 0
	manaRestore := 0
	if reward.HealthRestoreType != models.ScenarioFailureDrainTypeNone || reward.ManaRestoreType != models.ScenarioFailureDrainTypeNone {
		stats, maxHealth, maxMana, _, _, err := s.getScenarioResourceState(ctx, userID)
		if err != nil {
			return applied, err
		}
		healthRestore = min(
			scenarioDrainAmount(reward.HealthRestoreType, reward.HealthRestoreValue, maxHealth),
			stats.HealthDeficit,
		)
		manaRestore = min(
			scenarioDrainAmount(reward.ManaRestoreType, reward.ManaRestoreValue, maxMana),
			stats.ManaDeficit,
		)
	}
	if healthRestore > 0 || manaRestore > 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, -healthRestore, -manaRestore); err != nil {
			return applied, err
		}
	}
	applied.HealthRestored = healthRestore
	applied.ManaRestored = manaRestore

	now := time.Now()
	for _, statusTemplate := range reward.Statuses {
		name := strings.TrimSpace(statusTemplate.Name)
		if name == "" || statusTemplate.DurationSeconds <= 0 {
			continue
		}
		status := &models.UserStatus{
			UserID:          userID,
			Name:            name,
			Description:     strings.TrimSpace(statusTemplate.Description),
			Effect:          strings.TrimSpace(statusTemplate.Effect),
			Positive:        statusTemplate.Positive,
			EffectType:      normalizeUserStatusEffectType(statusTemplate.EffectType),
			DamagePerTick:   statusTemplate.DamagePerTick,
			StrengthMod:     statusTemplate.StrengthMod,
			DexterityMod:    statusTemplate.DexterityMod,
			ConstitutionMod: statusTemplate.ConstitutionMod,
			IntelligenceMod: statusTemplate.IntelligenceMod,
			WisdomMod:       statusTemplate.WisdomMod,
			CharismaMod:     statusTemplate.CharismaMod,
			StartedAt:       now,
			ExpiresAt:       now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
		}
		if err := s.dbClient.UserStatus().Create(ctx, status); err != nil {
			return applied, err
		}
		applied.Statuses = append(applied.Statuses, scenarioAppliedFailureStatus{
			Name:            name,
			Description:     strings.TrimSpace(statusTemplate.Description),
			Effect:          strings.TrimSpace(statusTemplate.Effect),
			EffectType:      string(normalizeUserStatusEffectType(statusTemplate.EffectType)),
			Positive:        statusTemplate.Positive,
			DamagePerTick:   statusTemplate.DamagePerTick,
			DurationSeconds: statusTemplate.DurationSeconds,
		})
	}
	return applied, nil
}

func (s *server) parseScenarioUpsertRequest(body scenarioUpsertRequest) (*models.Scenario, []models.ScenarioOption, []models.ScenarioItemReward, []models.ScenarioSpellReward, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid zone ID")
	}
	if strings.TrimSpace(body.Prompt) == "" {
		return nil, nil, nil, nil, fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(body.ImageURL) == "" {
		return nil, nil, nil, nil, fmt.Errorf("imageUrl is required")
	}
	thumbnailURL, err := normalizeScenarioThumbnailURL(body.ThumbnailURL)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if body.RewardExperience < 0 || body.RewardGold < 0 {
		return nil, nil, nil, nil, fmt.Errorf("reward values must be zero or greater")
	}
	difficulty, err := scenarioDifficultyValue(body.Difficulty)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if body.OpenEnded && len(body.Options) > 0 {
		return nil, nil, nil, nil, fmt.Errorf("open-ended scenarios cannot include options")
	}
	if !body.OpenEnded && len(body.Options) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("non-open-ended scenarios require at least one option")
	}
	if !body.OpenEnded && (body.RewardExperience > 0 || body.RewardGold > 0 || len(body.ItemRewards) > 0 || len(body.SpellRewards) > 0) {
		return nil, nil, nil, nil, fmt.Errorf("scenario-level rewards are only for open-ended scenarios")
	}
	failurePenaltyMode, err := normalizeScenarioFailurePenaltyMode(body.FailurePenaltyMode, body.OpenEnded)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	successRewardMode, err := normalizeScenarioSuccessRewardMode(body.SuccessRewardMode, body.OpenEnded)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	failureHealthDrainType, err := normalizeScenarioFailureDrainType(body.FailureHealthDrainType)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid failureHealthDrainType")
	}
	failureHealthDrainValue, err := normalizeScenarioFailureDrainValue(
		failureHealthDrainType,
		body.FailureHealthDrainValue,
		"failureHealthDrainValue",
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	failureManaDrainType, err := normalizeScenarioFailureDrainType(body.FailureManaDrainType)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid failureManaDrainType")
	}
	failureManaDrainValue, err := normalizeScenarioFailureDrainValue(
		failureManaDrainType,
		body.FailureManaDrainValue,
		"failureManaDrainValue",
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	failureStatuses, err := parseScenarioFailureStatusTemplates(body.FailureStatuses, "failureStatuses")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	successHealthRestoreType, err := normalizeScenarioFailureDrainType(body.SuccessHealthRestoreType)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid successHealthRestoreType")
	}
	successHealthRestoreValue, err := normalizeScenarioFailureDrainValue(
		successHealthRestoreType,
		body.SuccessHealthRestoreValue,
		"successHealthRestoreValue",
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	successManaRestoreType, err := normalizeScenarioFailureDrainType(body.SuccessManaRestoreType)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid successManaRestoreType")
	}
	successManaRestoreValue, err := normalizeScenarioFailureDrainValue(
		successManaRestoreType,
		body.SuccessManaRestoreValue,
		"successManaRestoreValue",
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	successStatuses, err := parseScenarioFailureStatusTemplates(body.SuccessStatuses, "successStatuses")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	scenario := &models.Scenario{
		ZoneID:                    zoneID,
		Latitude:                  body.Latitude,
		Longitude:                 body.Longitude,
		Prompt:                    strings.TrimSpace(body.Prompt),
		ImageURL:                  strings.TrimSpace(body.ImageURL),
		ThumbnailURL:              thumbnailURL,
		Difficulty:                difficulty,
		RewardExperience:          body.RewardExperience,
		RewardGold:                body.RewardGold,
		OpenEnded:                 body.OpenEnded,
		FailurePenaltyMode:        failurePenaltyMode,
		FailureHealthDrainType:    failureHealthDrainType,
		FailureHealthDrainValue:   failureHealthDrainValue,
		FailureManaDrainType:      failureManaDrainType,
		FailureManaDrainValue:     failureManaDrainValue,
		FailureStatuses:           failureStatuses,
		SuccessRewardMode:         successRewardMode,
		SuccessHealthRestoreType:  successHealthRestoreType,
		SuccessHealthRestoreValue: successHealthRestoreValue,
		SuccessManaRestoreType:    successManaRestoreType,
		SuccessManaRestoreValue:   successManaRestoreValue,
		SuccessStatuses:           successStatuses,
	}

	options := []models.ScenarioOption{}
	for _, optionPayload := range body.Options {
		optionText := strings.TrimSpace(optionPayload.OptionText)
		if optionText == "" {
			return nil, nil, nil, nil, fmt.Errorf("optionText is required")
		}
		successText := strings.TrimSpace(optionPayload.SuccessText)
		if successText == "" {
			successText = "Your approach works, and momentum turns in your favor."
		}
		failureText := strings.TrimSpace(optionPayload.FailureText)
		if failureText == "" {
			failureText = "The attempt falls short, and the moment slips away."
		}
		statTag, ok := normalizeScenarioStatTag(optionPayload.StatTag)
		if !ok {
			return nil, nil, nil, nil, fmt.Errorf("invalid option statTag")
		}
		if optionPayload.RewardExperience < 0 || optionPayload.RewardGold < 0 {
			return nil, nil, nil, nil, fmt.Errorf("option reward values must be zero or greater")
		}
		optionFailureHealthDrainType, err := normalizeScenarioFailureDrainType(optionPayload.FailureHealthDrainType)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("invalid option failureHealthDrainType")
		}
		optionFailureHealthDrainValue, err := normalizeScenarioFailureDrainValue(
			optionFailureHealthDrainType,
			optionPayload.FailureHealthDrainValue,
			"option failureHealthDrainValue",
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		optionFailureManaDrainType, err := normalizeScenarioFailureDrainType(optionPayload.FailureManaDrainType)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("invalid option failureManaDrainType")
		}
		optionFailureManaDrainValue, err := normalizeScenarioFailureDrainValue(
			optionFailureManaDrainType,
			optionPayload.FailureManaDrainValue,
			"option failureManaDrainValue",
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		optionFailureStatuses, err := parseScenarioFailureStatusTemplates(
			optionPayload.FailureStatuses,
			"option failureStatuses",
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		optionSuccessHealthRestoreType, err := normalizeScenarioFailureDrainType(optionPayload.SuccessHealthRestoreType)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("invalid option successHealthRestoreType")
		}
		optionSuccessHealthRestoreValue, err := normalizeScenarioFailureDrainValue(
			optionSuccessHealthRestoreType,
			optionPayload.SuccessHealthRestoreValue,
			"option successHealthRestoreValue",
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		optionSuccessManaRestoreType, err := normalizeScenarioFailureDrainType(optionPayload.SuccessManaRestoreType)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("invalid option successManaRestoreType")
		}
		optionSuccessManaRestoreValue, err := normalizeScenarioFailureDrainValue(
			optionSuccessManaRestoreType,
			optionPayload.SuccessManaRestoreValue,
			"option successManaRestoreValue",
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		optionSuccessStatuses, err := parseScenarioFailureStatusTemplates(
			optionPayload.SuccessStatuses,
			"option successStatuses",
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		var optionDifficulty *int
		if optionPayload.Difficulty != nil {
			if *optionPayload.Difficulty < 0 {
				return nil, nil, nil, nil, fmt.Errorf("option difficulty must be zero or greater")
			}
			value := *optionPayload.Difficulty
			optionDifficulty = &value
		}

		itemRewards := []models.ScenarioOptionItemReward{}
		for _, reward := range optionPayload.ItemRewards {
			if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
				return nil, nil, nil, nil, fmt.Errorf("option item rewards require inventoryItemId and positive quantity")
			}
			itemRewards = append(itemRewards, models.ScenarioOptionItemReward{
				InventoryItemID: reward.InventoryItemID,
				Quantity:        reward.Quantity,
			})
		}
		spellRewards := []models.ScenarioOptionSpellReward{}
		for _, reward := range optionPayload.SpellRewards {
			spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
			if err != nil || spellID == uuid.Nil {
				return nil, nil, nil, nil, fmt.Errorf("option spell rewards require a valid spellId")
			}
			spellRewards = append(spellRewards, models.ScenarioOptionSpellReward{
				SpellID: spellID,
			})
		}

		options = append(options, models.ScenarioOption{
			OptionText:                optionText,
			SuccessText:               successText,
			FailureText:               failureText,
			StatTag:                   statTag,
			Proficiencies:             models.StringArray(normalizeScenarioProficiencies(optionPayload.Proficiencies)),
			Difficulty:                optionDifficulty,
			RewardExperience:          optionPayload.RewardExperience,
			RewardGold:                optionPayload.RewardGold,
			FailureHealthDrainType:    optionFailureHealthDrainType,
			FailureHealthDrainValue:   optionFailureHealthDrainValue,
			FailureManaDrainType:      optionFailureManaDrainType,
			FailureManaDrainValue:     optionFailureManaDrainValue,
			FailureStatuses:           optionFailureStatuses,
			SuccessHealthRestoreType:  optionSuccessHealthRestoreType,
			SuccessHealthRestoreValue: optionSuccessHealthRestoreValue,
			SuccessManaRestoreType:    optionSuccessManaRestoreType,
			SuccessManaRestoreValue:   optionSuccessManaRestoreValue,
			SuccessStatuses:           optionSuccessStatuses,
			ItemRewards:               itemRewards,
			SpellRewards:              spellRewards,
		})
	}

	scenarioRewards := []models.ScenarioItemReward{}
	for _, reward := range body.ItemRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			return nil, nil, nil, nil, fmt.Errorf("scenario item rewards require inventoryItemId and positive quantity")
		}
		scenarioRewards = append(scenarioRewards, models.ScenarioItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	scenarioSpellRewards := []models.ScenarioSpellReward{}
	for _, reward := range body.SpellRewards {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil || spellID == uuid.Nil {
			return nil, nil, nil, nil, fmt.Errorf("scenario spell rewards require a valid spellId")
		}
		scenarioSpellRewards = append(scenarioSpellRewards, models.ScenarioSpellReward{
			SpellID: spellID,
		})
	}
	return scenario, options, scenarioRewards, scenarioSpellRewards, nil
}

func findScenarioOption(scenario *models.Scenario, optionID uuid.UUID) *models.ScenarioOption {
	if scenario == nil {
		return nil
	}
	for i := range scenario.Options {
		if scenario.Options[i].ID == optionID {
			return &scenario.Options[i]
		}
	}
	return nil
}

func scenarioRewardItemsFromOption(rewards []models.ScenarioOptionItemReward) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func scenarioRewardItemsFromScenario(rewards []models.ScenarioItemReward) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func scenarioRewardSpellsFromOption(rewards []models.ScenarioOptionSpellReward) []scenarioRewardSpell {
	out := make([]scenarioRewardSpell, 0, len(rewards))
	for _, reward := range rewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		out = append(out, scenarioRewardSpell{SpellID: reward.SpellID})
	}
	return out
}

func scenarioRewardSpellsFromScenario(rewards []models.ScenarioSpellReward) []scenarioRewardSpell {
	out := make([]scenarioRewardSpell, 0, len(rewards))
	for _, reward := range rewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		out = append(out, scenarioRewardSpell{SpellID: reward.SpellID})
	}
	return out
}

func (s *server) awardScenarioRewards(
	ctx context.Context,
	userID uuid.UUID,
	rewardExperience int,
	rewardGold int,
	rewardItems []scenarioRewardItem,
	rewardSpells []scenarioRewardSpell,
	proficiencies []string,
) ([]models.ItemAwarded, []models.SpellAwarded, error) {
	if rewardGold > 0 {
		if err := s.dbClient.User().AddGold(ctx, userID, rewardGold); err != nil {
			return nil, nil, err
		}
	}
	if rewardExperience > 0 {
		userLevel, err := s.dbClient.UserLevel().ProcessExperiencePointAdditions(ctx, userID, rewardExperience)
		if err != nil {
			return nil, nil, err
		}
		if userLevel.LevelsGained > 0 {
			if _, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, userID, userLevel.Level); err != nil {
				return nil, nil, err
			}
		}
	}

	itemsAwarded := []models.ItemAwarded{}
	for _, reward := range rewardItems {
		if err := s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(ctx, nil, &userID, reward.InventoryItemID, reward.Quantity); err != nil {
			return nil, nil, err
		}
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID)
		if err != nil || item == nil {
			continue
		}
		itemsAwarded = append(itemsAwarded, models.ItemAwarded{
			ID:       item.ID,
			Name:     item.Name,
			ImageURL: item.ImageURL,
			Quantity: reward.Quantity,
		})
	}

	spellsAwarded := []models.SpellAwarded{}
	seenSpells := map[uuid.UUID]bool{}
	for _, reward := range rewardSpells {
		if reward.SpellID == uuid.Nil {
			continue
		}
		if err := s.dbClient.UserSpell().GrantToUser(ctx, userID, reward.SpellID); err != nil {
			return nil, nil, err
		}
		if seenSpells[reward.SpellID] {
			continue
		}
		seenSpells[reward.SpellID] = true
		spell, err := s.dbClient.Spell().FindByID(ctx, reward.SpellID)
		if err != nil || spell == nil {
			continue
		}
		spellsAwarded = append(spellsAwarded, models.SpellAwarded{
			ID:      spell.ID,
			Name:    spell.Name,
			IconURL: spell.IconURL,
		})
	}

	for _, proficiency := range normalizeScenarioProficiencies(proficiencies) {
		if err := s.dbClient.UserProficiency().Increment(ctx, userID, proficiency, 1); err != nil {
			return nil, nil, err
		}
	}

	return itemsAwarded, spellsAwarded, nil
}

func (s *server) generateScenarioImage(ctx *gin.Context) {
	id := ctx.Param("id")
	scenarioID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	scenario, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if scenario == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	payload := jobs.GenerateScenarioImageTaskPayload{
		ScenarioID: scenarioID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateScenarioImageTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"status":   "queued",
		"scenario": scenario,
	})
}

func (s *server) getScenarios(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	scenarios, attemptedMap, err := s.dbClient.Scenario().FindAllWithUserStatus(ctx, &user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]scenarioWithUserStatus, len(scenarios))
	for i, scenario := range scenarios {
		response[i] = scenarioWithUserStatus{
			Scenario:        scenario,
			AttemptedByUser: attemptedMap[scenario.ID],
		}
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getScenario(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id := ctx.Param("id")
	scenarioID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	scenario, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	attempt, err := s.dbClient.Scenario().FindAttemptByUserAndScenario(ctx, user.ID, scenarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, scenarioWithUserStatus{
		Scenario:        *scenario,
		AttemptedByUser: attempt != nil,
	})
}

func (s *server) getScenariosForZone(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	scenarios, attemptedMap, err := s.dbClient.Scenario().FindByZoneIDWithUserStatus(ctx, zoneID, &user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]scenarioWithUserStatus, 0, len(scenarios))
	for _, scenario := range scenarios {
		if attemptedMap[scenario.ID] {
			// Hide resolved scenarios for this user so they no longer appear on map.
			continue
		}
		response = append(response, scenarioWithUserStatus{
			Scenario:        scenario,
			AttemptedByUser: false,
		})
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createScenario(ctx *gin.Context) {
	var requestBody scenarioUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scenario, options, scenarioRewards, scenarioSpellRewards, err := s.parseScenarioUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Scenario().Create(ctx, scenario); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceOptions(ctx, scenario.ID, options); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceItemRewards(ctx, scenario.ID, scenarioRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceSpellRewards(ctx, scenario.ID, scenarioSpellRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.Scenario().FindByID(ctx, scenario.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, scenarioWithUserStatus{
		Scenario:        *created,
		AttemptedByUser: false,
	})
}

func (s *server) updateScenario(ctx *gin.Context) {
	id := ctx.Param("id")
	scenarioID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	existing, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	var requestBody scenarioUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scenario, options, scenarioRewards, scenarioSpellRewards, err := s.parseScenarioUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Scenario().Update(ctx, scenarioID, scenario); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceOptions(ctx, scenarioID, options); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceItemRewards(ctx, scenarioID, scenarioRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceSpellRewards(ctx, scenarioID, scenarioSpellRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, scenarioWithUserStatus{
		Scenario:        *updated,
		AttemptedByUser: false,
	})
}

func (s *server) deleteScenario(ctx *gin.Context) {
	id := ctx.Param("id")
	scenarioID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	if _, err := s.dbClient.Scenario().FindByID(ctx, scenarioID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Scenario().Delete(ctx, scenarioID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "scenario deleted successfully"})
}

func (s *server) performScenario(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	scenarioID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	scenario, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if scenario == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	existingAttempt, err := s.dbClient.Scenario().FindAttemptByUserAndScenario(ctx, user.ID, scenarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingAttempt != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "scenario already resolved"})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	distance := util.HaversineDistance(userLat, userLng, scenario.Latitude, scenario.Longitude)
	if distance > scenarioInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("you must be within %.0f meters of the scenario. Currently %.0f meters away", scenarioInteractRadiusMeters, distance),
		})
		return
	}

	var requestBody scenarioPerformRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roll, err := rollScenarioDie()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to roll scenario die"})
		return
	}

	statTag := "charisma"
	proficiencies := []string{}
	creativityBonus := 0
	rewardExperience := 0
	rewardGold := 0
	rewardItems := []scenarioRewardItem{}
	rewardSpells := []scenarioRewardSpell{}
	threshold := scenario.Difficulty
	reason := "The outcome was uncertain."
	outcomeText := ""
	openEndedSuccessText := ""
	openEndedFailureText := ""
	var scenarioOptionID *uuid.UUID
	var selectedOption *models.ScenarioOption
	var freeformResponse *string

	if scenario.OpenEnded {
		responseText := strings.TrimSpace(requestBody.ResponseText)
		if responseText == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "responseText is required for open-ended scenarios"})
			return
		}
		assessment, err := s.assessScenarioFreeform(ctx, scenario.Prompt, responseText)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		statTag = assessment.StatTag
		proficiencies = assessment.Proficiencies
		creativityBonus = assessment.CreativityBonus
		if assessment.Reasoning != "" {
			reason = assessment.Reasoning
		}
		openEndedSuccessText = assessment.SuccessText
		openEndedFailureText = assessment.FailureText
		rewardExperience = scenario.RewardExperience
		rewardGold = scenario.RewardGold
		rewardItems = scenarioRewardItemsFromScenario(scenario.ItemRewards)
		rewardSpells = scenarioRewardSpellsFromScenario(scenario.SpellRewards)
		freeformResponse = &responseText
	} else {
		if requestBody.ScenarioOptionID == nil || *requestBody.ScenarioOptionID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "scenarioOptionId is required for choice scenarios"})
			return
		}
		option := findScenarioOption(scenario, *requestBody.ScenarioOptionID)
		if option == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "scenario option not found"})
			return
		}
		normalizedStatTag, ok := normalizeScenarioStatTag(option.StatTag)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "scenario option has invalid stat tag"})
			return
		}
		statTag = normalizedStatTag
		proficiencies = normalizeScenarioProficiencies([]string(option.Proficiencies))
		if option.Difficulty != nil {
			threshold = *option.Difficulty
		}
		rewardExperience = option.RewardExperience
		rewardGold = option.RewardGold
		rewardItems = scenarioRewardItemsFromOption(option.ItemRewards)
		rewardSpells = scenarioRewardSpellsFromOption(option.SpellRewards)
		reason = "The chosen approach shaped the outcome."
		if successCandidate := strings.TrimSpace(option.SuccessText); successCandidate != "" {
			outcomeText = successCandidate
		}
		scenarioOptionID = requestBody.ScenarioOptionID
		selectedOption = option
	}

	if threshold < 0 {
		threshold = 0
	}

	statValue, proficiencyBonus, err := s.getScenarioStatAndProficiencyBonuses(ctx, user.ID, statTag, proficiencies)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalScore := roll + statValue + proficiencyBonus + creativityBonus
	success := totalScore >= threshold

	itemsAwarded := []models.ItemAwarded{}
	spellsAwarded := []models.SpellAwarded{}
	appliedFailurePenalty := scenarioAppliedFailurePenalty{
		Statuses: []scenarioAppliedFailureStatus{},
	}
	appliedSuccessReward := scenarioAppliedSuccessReward{
		Statuses: []scenarioAppliedFailureStatus{},
	}
	if success {
		itemsAwarded, spellsAwarded, err = s.awardScenarioRewards(
			ctx,
			user.ID,
			rewardExperience,
			rewardGold,
			rewardItems,
			rewardSpells,
			proficiencies,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		successReward := resolveScenarioSuccessReward(scenario, selectedOption)
		appliedSuccessReward, err = s.applyScenarioSuccessReward(ctx, user.ID, successReward)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if scenario.OpenEnded {
			outcomeText = sanitizeScenarioOutcomeText(
				openEndedSuccessText,
				"Your response works, and the situation turns in your favor.",
			)
		}
	} else {
		rewardExperience = 0
		rewardGold = 0
		if scenario.OpenEnded {
			if strings.TrimSpace(reason) == "" {
				reason = "The roll did not beat the required threshold."
			}
			outcomeText = sanitizeScenarioOutcomeText(
				openEndedFailureText,
				"Your response falls short, and the moment slips away.",
			)
		} else {
			reason = "The roll did not beat the required threshold."
			if scenarioOptionID != nil {
				option := findScenarioOption(scenario, *scenarioOptionID)
				if option != nil {
					if failureCandidate := strings.TrimSpace(option.FailureText); failureCandidate != "" {
						outcomeText = failureCandidate
					}
				}
			}
			if outcomeText == "" {
				outcomeText = "The attempt falls short."
			}
		}
		penalty := resolveScenarioFailurePenalty(scenario, selectedOption)
		appliedFailurePenalty, err = s.applyScenarioFailurePenalty(ctx, user.ID, penalty)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if success && outcomeText == "" {
		outcomeText = "Success. Your plan holds."
	}

	attempt := &models.UserScenarioAttempt{
		UserID:            user.ID,
		ScenarioID:        scenario.ID,
		ScenarioOptionID:  scenarioOptionID,
		FreeformResponse:  freeformResponse,
		Roll:              roll,
		StatTag:           statTag,
		StatValue:         statValue,
		ProficienciesUsed: models.StringArray(proficiencies),
		ProficiencyBonus:  proficiencyBonus,
		CreativityBonus:   creativityBonus,
		Threshold:         threshold,
		TotalScore:        totalScore,
		Successful:        success,
		Reasoning:         &reason,
		RewardExperience:  rewardExperience,
		RewardGold:        rewardGold,
	}
	if err := s.dbClient.Scenario().CreateAttempt(ctx, attempt); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, scenarioPerformResponse{
		Successful:             success,
		Reason:                 reason,
		OutcomeText:            outcomeText,
		ScenarioID:             scenario.ID,
		ScenarioOptionID:       scenarioOptionID,
		Roll:                   roll,
		StatTag:                statTag,
		StatValue:              statValue,
		Proficiencies:          proficiencies,
		ProficiencyBonus:       proficiencyBonus,
		CreativityBonus:        creativityBonus,
		Threshold:              threshold,
		TotalScore:             totalScore,
		FailureHealthDrained:   appliedFailurePenalty.HealthDrained,
		FailureManaDrained:     appliedFailurePenalty.ManaDrained,
		FailureStatusesApplied: appliedFailurePenalty.Statuses,
		SuccessHealthRestored:  appliedSuccessReward.HealthRestored,
		SuccessManaRestored:    appliedSuccessReward.ManaRestored,
		SuccessStatusesApplied: appliedSuccessReward.Statuses,
		RewardExperience:       rewardExperience,
		RewardGold:             rewardGold,
		ItemsAwarded:           itemsAwarded,
		SpellsAwarded:          spellsAwarded,
	})
}
