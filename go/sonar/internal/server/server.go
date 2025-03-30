package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/mapbox"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/charicturist"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/config"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/questlog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrNotAuthenticated = errors.New("no authenticated user found")
)

type server struct {
	authClient     auth.Client
	texterClient   texter.Client
	dbClient       db.DbClient
	config         *config.Config
	awsClient      aws.AWSClient
	judgeClient    judge.Client
	quartermaster  quartermaster.Quartermaster
	chatClient     chat.Client
	charicturist   charicturist.Client
	mapboxClient   mapbox.Client
	questlogClient questlog.QuestlogClient
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
) Server {
	return &server{
		authClient:     authClient,
		texterClient:   texterClient,
		dbClient:       dbClient,
		config:         config,
		awsClient:      awsClient,
		judgeClient:    judgeClient,
		quartermaster:  quartermaster,
		chatClient:     chatClient,
		charicturist:   charicturist,
		mapboxClient:   mapboxClient,
		questlogClient: questlogClient,
	}
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()

	r.POST("/sonar/register", s.register)
	r.POST("/sonar/login", s.login)

	r.GET("/sonar/surveys", middleware.WithAuthentication(s.authClient, s.getSurverys))
	r.POST("/sonar/surveys", middleware.WithAuthentication(s.authClient, s.newSurvey))
	r.GET("sonar/surveys/:id/submissions", middleware.WithAuthentication(s.authClient, s.getSubmissionForSurvey))
	r.GET("/sonar/submissions/:id", middleware.WithAuthentication(s.authClient, s.getSubmission))
	r.GET("/sonar/whoami", middleware.WithAuthentication(s.authClient, s.whoami))
	r.POST("/sonar/categories", middleware.WithAuthentication(s.authClient, s.createCategory))
	r.POST("/sonar/activities", middleware.WithAuthentication(s.authClient, s.createActivity))
	r.DELETE("/sonar/categories/:id", middleware.WithAuthentication(s.authClient, s.deleteCategory))
	r.DELETE("/sonar/activities/:id", middleware.WithAuthentication(s.authClient, s.deleteActivity))
	r.GET("/sonar/teams", middleware.WithAuthentication(s.authClient, s.getTeams))
	r.POST("/sonar/pointsOfInterest", s.createPointOfInterest)
	r.GET("/sonar/pointsOfInterest", middleware.WithAuthentication(s.authClient, s.getPointsOfInterest))
	r.POST("/sonar/pointOfInterest/unlock", middleware.WithAuthentication(s.authClient, s.unlockPointOfInterest))
	r.POST("/sonar/neighbors", middleware.WithAuthentication(s.authClient, s.createNeighbor))
	r.GET("/sonar/neighbors", middleware.WithAuthentication(s.authClient, s.getNeighbors))
	r.POST("/sonar/matches/:id/start", middleware.WithAuthentication(s.authClient, s.startMatch))
	r.POST("/sonar/matches/:id/end", middleware.WithAuthentication(s.authClient, s.endMatch))
	r.POST("/sonar/matches", middleware.WithAuthentication(s.authClient, s.createMatch))
	r.GET("/sonar/matchesById/:id", middleware.WithAuthentication(s.authClient, s.getMatch))
	r.POST("/sonar/matches/:id/leave", middleware.WithAuthentication(s.authClient, s.leaveMatch))
	r.POST("/sonar/matches/:id/teams", middleware.WithAuthentication(s.authClient, s.createTeamForMatch))
	r.POST("/sonar/teams/:teamID", middleware.WithAuthentication(s.authClient, s.addUserToTeam))
	r.GET("/sonar/pointsOfInterest/group/:id", s.getPointOfInterestGroup)
	r.POST("/sonar/pointsOfInterest/group", s.createPointOfInterestGroup)
	r.GET("/sonar/pointsOfInterest/groups", s.getPointsOfInterestGroups)
	r.GET("/sonar/matches/current", middleware.WithAuthentication(s.authClient, s.getCurrentMatch))
	r.POST("/sonar/media/uploadUrl", middleware.WithAuthentication(s.authClient, s.getPresignedUploadUrl))
	r.POST("/sonar/pointOfInterest/challenge", middleware.WithAuthentication(s.authClient, s.submitAnswerPointOfInterestChallenge))
	r.POST("/sonar/teams/:teamID/edit", middleware.WithAuthentication(s.authClient, s.editTeamName))
	r.GET("/sonar/items", s.getInventoryItems)
	r.GET("/sonar/teams/:teamID/inventory", middleware.WithAuthentication(s.authClient, s.getTeamsInventory))
	r.POST("/sonar/inventory/:ownedInventoryItemID/use", middleware.WithAuthentication(s.authClient, s.useItem))
	r.GET("/sonar/chat", middleware.WithAuthentication(s.authClient, s.getChat))
	r.POST("/sonar/teams/:teamID/inventory/add", s.addItemToTeam)
	r.POST("/sonar/admin/pointOfInterest/unlock", middleware.WithAuthentication(s.authClient, s.unlockPointOfInterestForTeam))
	r.POST("/sonar/admin/pointOfInterest/capture", middleware.WithAuthentication(s.authClient, s.capturePointOfInterestForTeam))
	r.POST("/sonar/pointsOfInterest/group/:id", middleware.WithAuthentication(s.authClient, s.createPointOfInterest))
	r.POST("/sonar/generateProfilePictureOptions", middleware.WithAuthentication(s.authClient, s.generateProfilePictureOptions))
	r.GET("/sonar/generations/complete", middleware.WithAuthentication(s.authClient, s.getCompleteGenerationsForUser))
	r.POST("/sonar/profilePicture", middleware.WithAuthentication(s.authClient, s.setProfilePicture))
	r.PATCH("/sonar/pointsOfInterest/group/:id", s.editPointOfInterestGroup)
	r.DELETE("/sonar/pointsOfInterest/group/:id", s.deletePointOfInterestGroup)
	r.DELETE("/sonar/pointsOfInterest/challenge/:id", s.deletePointOfInterestChallenge)
	r.PATCH("/sonar/pointsOfInterest/challenge/:id", s.editPointOfInterestChallenge)
	r.POST("/sonar/pointsOfInterest/challenge", s.createPointOfInterestChallenge)
	r.PATCH("/sonar/pointsOfInterest/:id", s.editPointOfInterest)
	r.DELETE("/sonar/pointsOfInterest/:id", s.deletePointOfInterest)
	r.PATCH("/sonar/pointsofInterest/group/imageUrl/:id", s.editPointOfInterestGroupImageUrl)
	r.PATCH("/sonar/pointsofInterest/imageUrl/:id", s.editPointOfInterestImageUrl)
	r.POST("/sonar/pointOfInterest/children", middleware.WithAuthentication(s.authClient, s.createPointOfInterestChildren))
	r.DELETE("/sonar/pointOfInterest/children/:id", middleware.WithAuthentication(s.authClient, s.deletePointOfInterestChildren))
	r.GET("/sonar/pointsOfInterest/discoveries", middleware.WithAuthentication(s.authClient, s.getPointOfInterestDiscoveries))
	r.GET("/sonar/pointsOfInterest/challenges/submissions", middleware.WithAuthentication(s.authClient, s.getPointOfInterestChallengeSubmissions))
	r.GET("/sonar/ownedInventoryItems", middleware.WithAuthentication(s.authClient, s.getOwnedInventoryItems))
	r.POST("/sonar/matches/:id/invite", middleware.WithAuthentication(s.authClient, s.inviteToMatch))
	r.GET("/sonar/matches/:id/users", middleware.WithAuthentication(s.authClient, s.getMatch))
	r.GET("/sonar/mapbox/places", s.getMapboxPlaces)
	r.GET("/sonar/questlog", middleware.WithAuthentication(s.authClient, s.getQuestLog))
	r.GET("/sonar/matches/hasCurrentMatch", middleware.WithAuthentication(s.authClient, s.hasCurrentMatch))
	r.GET("/sonar/users", middleware.WithAuthentication(s.authClient, s.getAllUsers))
	r.POST("/sonar/users/giveItem", middleware.WithAuthentication(s.authClient, s.giveItem))
	r.Run(":8042")
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

	stringLat := ctx.Query("lat")
	stringLng := ctx.Query("lng")
	lat, err := strconv.ParseFloat(stringLat, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
		return
	}
	lng, err := strconv.ParseFloat(stringLng, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng"})
		return
	}

	questLog, err := s.questlogClient.GetQuestLog(ctx, user.ID, lat, lng)
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

	if err := s.dbClient.PointOfInterestChildren().Create(ctx, requestBody.PointOfInterestGroupMemberID, requestBody.PointOfInterestID, requestBody.PointOfInterestChallengeID); err != nil {
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
		PointOfInterestID uuid.UUID `binding:"required" json:"pointOfInterestId"`
		Tier              int       `binding:"required" json:"tier"`
		Question          string    `binding:"required" json:"question"`
		InventoryItemID   int       `binding:"required" json:"inventoryItemId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := s.dbClient.PointOfInterestChallenge().Create(ctx, requestBody.PointOfInterestID, requestBody.Tier, requestBody.Question, requestBody.InventoryItemID); err != nil {
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
		Name        string `binding:"required" json:"name"`
		Description string `binding:"required" json:"description"`
		Type        int    `binding:"required" json:"type"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	typeValue := models.PointOfInterestGroupType(requestBody.Type)

	if err := s.dbClient.PointOfInterestGroup().Edit(ctx, pointOfInterestGroupID, requestBody.Name, requestBody.Description, typeValue); err != nil {
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
		Gender            string `binding:"required" json:"gender"`
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

	if err := s.charicturist.CreateCharacter(ctx, charicturist.CreateCharacterRequest{
		ProfilePictureUrl: requestBody.ProfilePictureUrl,
		UserId:            user.ID,
		Gender:            requestBody.Gender,
	}); err != nil {
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

func (s *server) capturePointOfInterestForTeam(ctx *gin.Context) {
	var requestBody struct {
		PointOfInterestID uuid.UUID  `binding:"required" json:"pointOfInterestId"`
		TeamID            *uuid.UUID `binding:"required" json:"teamId"`
		UserID            *uuid.UUID `json:"userId,omitempty"`
		Tier              int        `binding:"required" json:"tier"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	challenge, err := s.dbClient.PointOfInterestChallenge().GetChallengeForPointOfInterest(ctx, requestBody.PointOfInterestID, requestBody.Tier)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := s.dbClient.PointOfInterestChallenge().SubmitAnswerForChallenge(ctx, challenge.ID, requestBody.TeamID, requestBody.UserID, "", "", true); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := s.chatClient.AddCaptureMessage(ctx, requestBody.TeamID, requestBody.UserID, challenge.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "challenge submitted successfully",
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

	submission, err := s.judgeClient.JudgeSubmission(ctx, judge.JudgeSubmissionRequest{
		ChallengeID:        requestBody.ChallengeID,
		TeamID:             requestBody.TeamID,
		UserID:             requestBody.UserID,
		TextSubmission:     requestBody.TextSubmission,
		ImageSubmissionUrl: requestBody.ImageSubmissionUrl,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	challenge, err := s.dbClient.PointOfInterestChallenge().FindByID(ctx, requestBody.ChallengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var item quartermaster.InventoryItem
	if submission.Judgement.Judgement {
		if challenge.InventoryItemID == 0 {
			item, err = s.quartermaster.GetItem(ctx, requestBody.TeamID, requestBody.UserID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		} else {
			item, err = s.quartermaster.GetItemSpecificItem(ctx, requestBody.TeamID, requestBody.UserID, challenge.InventoryItemID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		if err := s.chatClient.AddCaptureMessage(ctx, requestBody.TeamID, requestBody.UserID, requestBody.ChallengeID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"judgement": submission,
		"item":      item,
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
		Name        string `binding:"required" json:"name"`
		Description string `binding:"required" json:"description"`
		ImageUrl    string `binding:"required" json:"imageUrl"`
		Type        int    `binding:"required" json:"type"`
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
