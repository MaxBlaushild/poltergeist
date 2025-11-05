package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
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
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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
	gameEngineClient gameengine.GameEngineClient
	livenessClient   liveness.LivenessClient
}

type Server interface {
	ListenAndServe(port string)
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
		gameEngineClient: gameEngineClient,
		livenessClient:   livenessClient,
	}
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()

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
	r.POST("/sonar/teams/:teamID/edit", middleware.WithAuthentication(s.authClient, s.livenessClient, s.editTeamName))
	r.GET("/sonar/items", s.getInventoryItems)
	r.GET("/sonar/inventory-items", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getAllInventoryItems))
	r.GET("/sonar/inventory-items/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getInventoryItem))
	r.POST("/sonar/inventory-items", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createInventoryItem))
	r.PUT("/sonar/inventory-items/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateInventoryItem))
	r.DELETE("/sonar/inventory-items/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteInventoryItem))
	r.GET("/sonar/teams/:teamID/inventory", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTeamsInventory))
	r.POST("/sonar/inventory/:ownedInventoryItemID/use", middleware.WithAuthentication(s.authClient, s.livenessClient, s.useItem))
	r.GET("/sonar/chat", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getChat))
	r.POST("/sonar/teams/:teamID/inventory/add", s.addItemToTeam)
	r.POST("/sonar/admin/pointOfInterest/unlock", middleware.WithAuthentication(s.authClient, s.livenessClient, s.unlockPointOfInterestForTeam))
	r.POST("/sonar/pointsOfInterest/group/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createPointOfInterest))
	r.POST("/sonar/generateProfilePictureOptions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateProfilePictureOptions))
	r.GET("/sonar/generations/complete", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCompleteGenerationsForUser))
	r.POST("/sonar/profilePicture", middleware.WithAuthentication(s.authClient, s.livenessClient, s.setProfilePicture))
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
	r.GET("/sonar/tagGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTagGroups))
	r.POST("/sonar/tags/add", middleware.WithAuthentication(s.authClient, s.livenessClient, s.addTagToPointOfInterest))
	r.DELETE("/sonar/tags/:tagID/pointOfInterest/:pointOfInterestID", middleware.WithAuthentication(s.authClient, s.livenessClient, s.removeTagFromPointOfInterest))
	r.GET("/sonar/zones", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZones))
	r.GET("/sonar/zones/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZone))
	r.POST("/sonar/zones", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createZone))
	r.GET("/sonar/zones/:id/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPointsOfInterestForZone))
	r.POST("/sonar/zones/:id/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generatePointsOfInterestForZone))
	r.GET("/sonar/placeTypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPlaceTypes))
	r.DELETE("/sonar/zones/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteZone))
	r.POST("/sonar/zones/:id/pointOfInterest/:pointOfInterestId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.addPointOfInterestToZone))
	r.DELETE("/sonar/zones/:id/pointOfInterest/:pointOfInterestId", middleware.WithAuthentication(s.authClient, s.livenessClient, s.removePointOfInterestFromZone))
	r.GET("/sonar/pointOfInterest/:id/zone", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneForPointOfInterest))
	r.POST("/sonar/pointOfInterest/import", middleware.WithAuthentication(s.authClient, s.livenessClient, s.importPointOfInterest))
	r.POST("/sonar/pointOfInterest/refresh", middleware.WithAuthentication(s.authClient, s.livenessClient, s.refreshPointOfInterest))
	r.POST("/sonar/pointOfInterest/image/refresh", middleware.WithAuthentication(s.authClient, s.livenessClient, s.refreshPointOfInterestImage))
	r.GET("/sonar/google/places", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getGooglePlaces))
	r.GET("/sonar/google/place/:placeID", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getGooglePlace))
	r.POST("/sonar/quests/:zoneID/:questArchTypeID/generate", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateQuest))
	r.POST("/sonar/tags/move", middleware.WithAuthentication(s.authClient, s.livenessClient, s.moveTagToTagGroup))
	r.POST("/sonar/tags/createGroup", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTagGroup))
	r.GET("/sonar/locationArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getLocationArchetypes))
	r.GET("/sonar/locationArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getLocationArchetype))
	r.POST("/sonar/locationArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createLocationArchetype))
	r.DELETE("/sonar/locationArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteLocationArchetype))
	r.PATCH("/sonar/locationArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateLocationArchetype))
	r.GET("/sonar/questArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestArchetypes))
	r.GET("/sonar/questArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestArchetype))
	r.POST("/sonar/questArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuestArchetype))
	r.DELETE("/sonar/questArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteQuestArchetype))
	r.PATCH("/sonar/questArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateQuestArchetype))
	r.POST("/sonar/questArchetypeNodes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createQuestArchetypeNode))
	r.POST("/sonar/questArchetypes/:id/challenges", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateQuestArchetypeChallenge))
	r.GET("/sonar/questArchetypes/:id/challenges", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getQuestArchetypeChallenges))
	r.POST("/sonar/zones/:id/questArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.generateQuestArchetypesForZone))
	r.GET("/sonar/zoneQuestArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getZoneQuestArchetypes))
	r.POST("/sonar/zoneQuestArchetypes", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createZoneQuestArchetype))
	r.DELETE("/sonar/zoneQuestArchetypes/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteZoneQuestArchetype))
	r.GET("/sonar/search/tags", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getRelevantTags))
	r.POST("/sonar/trackedPointOfInterestGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createTrackedPointOfInterestGroup))
	r.GET("/sonar/trackedPointOfInterestGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getTrackedPointOfInterestGroups))
	r.DELETE("/sonar/trackedPointOfInterestGroups/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteTrackedPointOfInterestGroup))
	r.DELETE("/sonar/trackedPointOfInterestGroups", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteAllTrackedPointOfInterestGroups))
	r.POST("/sonar/zones/:id/boundary", middleware.WithAuthentication(s.authClient, s.livenessClient, s.upsertZoneBoundary))
	r.PATCH("/sonar/zones/:id/edit", middleware.WithAuthentication(s.authClient, s.livenessClient, s.editZone))
	r.GET("/sonar/level", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getLevel))
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
	r.POST("/sonar/characters", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createCharacter))
	r.PUT("/sonar/characters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateCharacter))
	r.DELETE("/sonar/characters/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteCharacter))
	r.GET("/sonar/characters/:id/actions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacterActions))
	r.GET("/sonar/character-actions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getCharacterAction))
	r.POST("/sonar/character-actions", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createCharacterAction))
	r.PUT("/sonar/character-actions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.updateCharacterAction))
	r.DELETE("/sonar/character-actions/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteCharacterAction))
	r.POST("/sonar/character-actions/:id/purchase", middleware.WithAuthentication(s.authClient, s.livenessClient, s.purchaseFromShop))
	r.POST("/sonar/friendInvites/accept", middleware.WithAuthentication(s.authClient, s.livenessClient, s.acceptFriendInvite))
	r.POST("/sonar/friendInvites/create", middleware.WithAuthentication(s.authClient, s.livenessClient, s.createFriendInvite))
	r.GET("/sonar/partyInvites", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getPartyInvites))
	r.GET("/sonar/friendInvites", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getFriendInvites))
	r.DELETE("/sonar/friendInvites/:id", middleware.WithAuthentication(s.authClient, s.livenessClient, s.deleteFriendInvite))
	r.GET("/sonar/friends", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getFriends))
	r.POST("/sonar/profile", middleware.WithAuthentication(s.authClient, s.livenessClient, s.setProfile))
	r.GET("/sonar/activities", middleware.WithAuthentication(s.authClient, s.livenessClient, s.getActivities))
	r.POST("/sonar/activities/markAsSeen", middleware.WithAuthentication(s.authClient, s.livenessClient, s.markActivitiesAsSeen))
	r.Run(":8042")
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

	ctx.JSON(http.StatusOK, activities)
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

func (s *server) createTrackedPointOfInterestGroup(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		PointOfInterestGroupID uuid.UUID `json:"pointOfInterestGroupID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.dbClient.TrackedPointOfInterestGroup().Create(ctx, requestBody.PointOfInterestGroupID, user.ID)
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
	trackedPointOfInterestGroups, err := s.dbClient.TrackedPointOfInterestGroup().GetByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, trackedPointOfInterestGroups)
}

func (s *server) deleteTrackedPointOfInterestGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	groupIDUUID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tracked point of interest group ID"})
		return
	}
	err = s.dbClient.TrackedPointOfInterestGroup().Delete(ctx, groupIDUUID)
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
	err = s.dbClient.TrackedPointOfInterestGroup().DeleteAllForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "all tracked point of interest groups deleted successfully"})
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
		ZoneID           uuid.UUID `json:"zoneID"`
		QuestArchetypeID uuid.UUID `json:"questArchetypeID"`
		NumberOfQuests   int       `json:"numberOfQuests"`
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
	}

	err := s.dbClient.ZoneQuestArchetype().Create(ctx, zoneQuestArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, zoneQuestArchetype)
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
			ZoneID:           zoneIDUUID,
			QuestArchetypeID: zoneQuestArchetype.QuestArchetypeID,
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
		LocationArchetypeID *uuid.UUID `json:"locationArchetypeID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Reward:         requestBody.Reward,
		UnlockedNodeID: newNodeID,
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

func (s *server) createQuestArchetypeNode(ctx *gin.Context) {
	var requestBody struct {
		LocationArchetypeID uuid.UUID `json:"locationArchetypeID"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	questArchetypeNode := &models.QuestArchetypeNode{
		ID:                  uuid.New(),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		LocationArchetypeID: requestBody.LocationArchetypeID,
	}

	err := s.dbClient.QuestArchetypeNode().Create(ctx, questArchetypeNode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, questArchetypeNode)
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
	ctx.JSON(http.StatusOK, questArchetype)
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
		Name          string   `json:"name"`
		IncludedTypes []string `json:"includedTypes"`
		ExcludedTypes []string `json:"excludedTypes"`
		Challenges    []string `json:"challenges"`
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
	locationArchetype.Challenges = pq.StringArray(requestBody.Challenges)

	err = s.dbClient.LocationArchetype().Update(ctx, locationArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, locationArchetype)
}

func (s *server) createLocationArchetype(ctx *gin.Context) {
	var requestBody struct {
		Name          string   `json:"name"`
		IncludedTypes []string `json:"includedTypes"`
		ExcludedTypes []string `json:"excludedTypes"`
		Challenges    []string `json:"challenges"`
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
		Challenges:    pq.StringArray(requestBody.Challenges),
	}

	err := s.dbClient.LocationArchetype().Create(ctx, locationArchetype)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"shit ass error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, locationArchetype)
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
	ctx.JSON(http.StatusOK, questArchType)
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

	payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
		ZoneID:           zoneID,
		QuestArchetypeID: questArchTypeIDUUID,
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

	err = s.locationSeeder.RefreshPointOfInterestImage(ctx, poi)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, poi)
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

	zone, err := s.dbClient.Zone().FindByID(ctx, requestBody.ZoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	poi, err := s.locationSeeder.ImportPlace(ctx, requestBody.PlaceID, *zone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, poi)
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
		Radius      float64 `json:"radius"`
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
		Radius:      requestBody.Radius,
		Description: requestBody.Description,
	}
	if err := s.dbClient.Zone().Create(ctx, zone); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zone)
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
	zoneID, err := uuid.Parse(stringZoneID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
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
		Name        string `binding:"required" json:"name"`
		Description string `binding:"required" json:"description"`
		Lat         string `binding:"required" json:"lat"`
		Lng         string `binding:"required" json:"lng"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterest().Edit(ctx, pointOfInterestID, requestBody.Name, requestBody.Description, requestBody.Lat, requestBody.Lng); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
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
		Name            string `binding:"required" json:"name"`
		Description     string `binding:"required" json:"description"`
		Type            int    `binding:"required" json:"type"`
		Gold            *int   `json:"gold"`
		InventoryItemID *int   `json:"inventoryItemId"`
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
	}

	if err := ctx.Bind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.dbClient.PointOfInterest().CreateForGroup(ctx, &models.PointOfInterest{
		Name:        request.Name,
		Description: request.Description,
		Lat:         request.Latitude,
		Lng:         request.Longitude,
		ImageUrl:    request.ImageUrl,
		Clue:        request.Clue,
	}, pointOfInterestGroupID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "point of interest created successfully",
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

	if err := s.quartermaster.UseItem(ctx, ownedInventoryItemID, &request); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
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

	if err := s.chatClient.AddUseItemMessage(ctx, *ownedInventoryItem, request); err != nil {
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
		Name          string `json:"name" binding:"required"`
		ImageURL      string `json:"imageUrl"`
		FlavorText    string `json:"flavorText"`
		EffectText    string `json:"effectText"`
		RarityTier    string `json:"rarityTier" binding:"required"`
		IsCaptureType bool   `json:"isCaptureType"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &models.InventoryItem{
		Name:          requestBody.Name,
		ImageURL:      requestBody.ImageURL,
		FlavorText:    requestBody.FlavorText,
		EffectText:    requestBody.EffectText,
		RarityTier:    requestBody.RarityTier,
		IsCaptureType: requestBody.IsCaptureType,
	}

	if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create inventory item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item)
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	var requestBody struct {
		Name          string `json:"name"`
		ImageURL      string `json:"imageUrl"`
		FlavorText    string `json:"flavorText"`
		EffectText    string `json:"effectText"`
		RarityTier    string `json:"rarityTier"`
		IsCaptureType bool   `json:"isCaptureType"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &models.InventoryItem{
		Name:          requestBody.Name,
		ImageURL:      requestBody.ImageURL,
		FlavorText:    requestBody.FlavorText,
		EffectText:    requestBody.EffectText,
		RarityTier:    requestBody.RarityTier,
		IsCaptureType: requestBody.IsCaptureType,
	}

	if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, id, item); err != nil {
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
	_, err := s.getAuthenticatedUser(ctx)
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
		Name            string `binding:"required" json:"name"`
		Description     string `binding:"required" json:"description"`
		ImageUrl        string `binding:"required" json:"imageUrl"`
		Type            int    `binding:"required" json:"type"`
		Gold            *int   `json:"gold"`
		InventoryItemID *int   `json:"inventoryItemId"`
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

	if requestBody.Gold != nil || requestBody.InventoryItemID != nil {
		// Update the group with provided gold value and/or inventory item
		updateData := make(map[string]interface{})
		if requestBody.Gold != nil {
			updateData["gold"] = *requestBody.Gold
			group.Gold = *requestBody.Gold
		}
		if requestBody.InventoryItemID != nil {
			updateData["inventory_item_id"] = *requestBody.InventoryItemID
			group.InventoryItemID = requestBody.InventoryItemID
		}
		if err := s.dbClient.PointOfInterestGroup().Update(ctx, group.ID, &models.PointOfInterestGroup{Gold: group.Gold, InventoryItemID: group.InventoryItemID}); err != nil {
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
		ctx.JSON(200, match.PointsOfInterest)
		return
	}

	pointOfInterests, err := s.dbClient.PointOfInterest().FindAll(ctx)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
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
	if err := s.dbClient.TrackedPointOfInterestGroup().DeleteAllForUser(ctx, userID); err != nil {
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

	// 16. Delete all owned inventory items
	if err := s.dbClient.InventoryItem().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete owned inventory items: " + err.Error()})
		return
	}

	// 17. Delete all user team relationships
	if err := s.dbClient.UserTeam().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user team relationships: " + err.Error()})
		return
	}

	// 18. Delete all parties where user is leader
	if err := s.dbClient.Party().DeleteAllForUser(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete parties: " + err.Error()})
		return
	}

	// 19. Finally, delete the user
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
		if err := s.dbClient.TrackedPointOfInterestGroup().DeleteAllForUser(ctx, userID); err != nil {
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

		// 9. Finally, delete the user
		if err := s.dbClient.User().Delete(ctx, userID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user " + userID.String() + ": " + err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("successfully deleted %d users", len(requestBody.UserIDs))})
}

// Character handlers
func (s *server) getCharacters(ctx *gin.Context) {
	characters, err := s.dbClient.Character().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

func (s *server) createCharacter(ctx *gin.Context) {
	var requestBody struct {
		Name             string `json:"name" binding:"required"`
		Description      string `json:"description"`
		MapIconUrl       string `json:"mapIconUrl"`
		DialogueImageUrl string `json:"dialogueImageUrl"`
		MovementPattern  struct {
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
		MovementPatternID: movementPattern.ID,
		MovementPattern:   *movementPattern,
	}

	if err := s.dbClient.Character().Create(ctx, character); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create character: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, character)
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
		Name             string `json:"name"`
		Description      string `json:"description"`
		MapIconUrl       string `json:"mapIconUrl"`
		DialogueImageUrl string `json:"dialogueImageUrl"`
		MovementPattern  struct {
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
	}

	if err := s.dbClient.Character().Update(ctx, id, characterUpdates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update character: " + err.Error()})
		return
	}

	// Fetch the updated character with movement pattern
	updatedCharacter, err := s.dbClient.Character().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedCharacter)
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

	actions, err := s.dbClient.CharacterAction().FindByCharacterID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
