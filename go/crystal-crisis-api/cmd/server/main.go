package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/crystal-crisis-api/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	texterClient := texter.NewClient()
	authClient := auth.NewClient()

	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     cfg.Public.DbName,
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.POST("/crystal-crisis/neighbors", func(c *gin.Context) {
		var neighbor models.Neighbor

		if err := c.Bind(&neighbor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit neighbor create request",
			})
			return
		}

		if err := dbClient.Neighbor().Create(c, neighbor.CrystalOneID, neighbor.CrystalTwoID); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "everythings ok",
		})
	})

	router.GET("/crystal-crisis/neighbors", func(c *gin.Context) {
		neighbors, err := dbClient.Neighbor().FindAll(c)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, neighbors)
	})

	router.POST("/crystal-crisis/teams", func(c *gin.Context) {
		var createTeamsRequest struct {
			UserIDs []string `json:"userIds" binding:"required"`
			Name    string   `json:"name" binding:"required"`
		}

		if err := c.Bind(&createTeamsRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		var userIDs []uuid.UUID
		for _, id := range createTeamsRequest.UserIDs {
			userID, err := uuid.Parse(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			userIDs = append(userIDs, userID)
		}

		if err := dbClient.Team().Create(c, userIDs, createTeamsRequest.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{"message": "done"})
	})

	router.GET("/crystal-crisis/teams", func(c *gin.Context) {
		teams, err := dbClient.Team().GetAll(c)
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
			users, err := authClient.GetUsers(ctx, userIDs)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			payload["users"] = users
		}

		c.JSON(200, payload)
	})

	router.GET("/crystal-crisis/crystals/:teamID", func(c *gin.Context) {
		stringTeamID := c.Param("teamID")
		crystals, err := dbClient.Crystal().FindAll(c)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		teamID, err := uuid.Parse(stringTeamID)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		crystalUnlockings, err := dbClient.CrystalUnlocking().FindByTeamID(ctx, teamID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for i, crystal := range crystals {
			found := false
			for _, unlocking := range crystalUnlockings {
				if unlocking.CrystalID == crystal.ID {
					found = true
				}
			}

			if !found {
				crystals[i].AttuneChallenge = ""
				crystals[i].CaptureChallenge = ""
			} else {
				crystals[i].Clue = ""
			}
		}

		c.JSON(200, crystals)
	})

	router.POST("/crystal-crisis/crystals", func(c *gin.Context) {
		var createCrystalRequest struct {
			Name             string `json:"name" binding:"required"`
			Clue             string `json:"clue" binding:"required"`
			CaptureChallenge string `json:"captureChallenge" binding:"required"`
			AttuneChallenge  string `json:"attuneChallenge" binding:"required"`
			Lat              string `json:"lat" binding:"required"`
			Lng              string `json:"lng" binding:"required"`
		}

		if err := c.Bind(&createCrystalRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		if err := dbClient.Crystal().Create(ctx, models.Crystal{
			Name:             createCrystalRequest.Name,
			Clue:             createCrystalRequest.Clue,
			AttuneChallenge:  createCrystalRequest.AttuneChallenge,
			CaptureChallenge: createCrystalRequest.CaptureChallenge,
			Lat:              createCrystalRequest.Lat,
			Lng:              createCrystalRequest.Lng,
		}); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "everything cool",
		})
	})

	router.POST("/crystal-crisis/crystals/unlock", func(c *gin.Context) {
		var crystalUnlockRequest struct {
			TeamID    string `json:"teamId" binding:"required"`
			CrystalID string `json:"crystalId" binding:"required"`
		}

		if err := c.Bind(&crystalUnlockRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit crystal unlock request",
			})
			return
		}

		crystalID, err := uuid.Parse(crystalUnlockRequest.CrystalID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit crystal id",
			})
			return
		}

		teamID, err := uuid.Parse(crystalUnlockRequest.TeamID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit team id",
			})
			return
		}

		if err := dbClient.Crystal().Unlock(c, crystalID, teamID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "everything cool",
		})
	})

	router.POST("/crystal-crisis/crystals/capture", func(c *gin.Context) {
		var captureCrystalRequest struct {
			CrystalID string `json:"crystalId" binding:"required"`
			TeamID    string `json:"teamId" binding:"required"`
			Attune    bool   `json:"attune"`
		}

		if err := c.Bind(&captureCrystalRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		crystalID, err := uuid.Parse(captureCrystalRequest.CrystalID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit crystal id",
			})
			return
		}

		teamID, err := uuid.Parse(captureCrystalRequest.TeamID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit team id",
			})
			return
		}

		if err := dbClient.Crystal().Capture(ctx, crystalID, teamID, captureCrystalRequest.Attune); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		/// TEXT EVERYBODY!!!!!!!!!!!

		teams, err := dbClient.Team().GetAll(c)
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

		users, err := authClient.GetUsers(ctx, userIDs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var capturedOrAttuned string
		if captureCrystalRequest.Attune {
			capturedOrAttuned = "attuned"
		} else {
			capturedOrAttuned = "captured"
		}

		var capturingTeam models.Team
		for _, team := range teams {
			if team.ID == teamID {
				capturingTeam = team
			}
		}

		crystal, err := dbClient.Crystal().FindByID(c, crystalID)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		for _, user := range users {
			texterClient.Text(ctx, &texter.Text{
				To:   user.PhoneNumber,
				From: cfg.Public.CrystalCrisisPhoneNumber,
				Body: fmt.Sprintf("%s team has %s %s.", capturingTeam.Name, capturedOrAttuned, crystal.Name),
			})
		}

		c.JSON(200, gin.H{"messgae": "done!"})
	})

	router.Run(":8091")
}
