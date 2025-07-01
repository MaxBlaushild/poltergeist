package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MaxBlaushild/authenticator/internal/config"
	"github.com/MaxBlaushild/authenticator/internal/token"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

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

	tokenClient, err := token.NewClient(cfg.Secret.AuthPrivateKey)
	if err != nil {
		panic(err)
	}

	texterClient := texter.NewClient()

	awsClient := aws.NewAWSClient("us-east-1")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	r.GET("/authenticator/users/:userID/profilePictureUploadUrl/:key", func(c *gin.Context) {
		userID := c.Query("userID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "userID is required",
			})
			return
		}

		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "key is required",
			})
			return
		}

		url, err := awsClient.GeneratePresignedUploadURL(fmt.Sprintf("users/%s", userID), key, time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"url": url,
		})
	})

	r.POST("/authenticator/users/:userID/profilePictureUrl", func(c *gin.Context) {
		var requestBody struct {
			Url string `json:"key" binding:"required"`
		}

		if err := c.Bind(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		stringUserID := c.Param("userID")
		if stringUserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "userID is required",
			})
			return
		}

		userID, err := uuid.Parse(stringUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := dbClient.User().UpdateProfilePictureUrl(c, userID, requestBody.Url); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	r.POST("/authenticator/token/verify", func(c *gin.Context) {
		var requestBody auth.VerifyTokenRequest

		if err := c.Bind(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		userID, err := tokenClient.Verify(requestBody.Token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		user, err := dbClient.User().FindByID(c, *userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, user)
	})

	r.POST("/authenticator/text/verification-code", func(c *gin.Context) {
		var requestBody struct {
			PhoneNumber string `json:"phoneNumber" binding:"required"`
			AppName     string `json:"appName" binding:"required"`
		}

		if err := c.Bind(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		user, err := dbClient.User().FindByPhoneNumber(c, requestBody.PhoneNumber)
		if err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		code, err := dbClient.TextVerificationCode().Insert(c, requestBody.PhoneNumber)
		if err != nil {
			fmt.Println("text verification code insertion error")
			fmt.Println(err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := texterClient.Text(ctx, &texter.Text{
			Body:     fmt.Sprintf("%s is your %s verification code", code.Code, requestBody.AppName),
			To:       requestBody.PhoneNumber,
			From:     cfg.Public.PhoneNumber,
			TextType: "verification-code",
		}); err != nil {
			fmt.Println("text send error")
			fmt.Println(err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, user)
	})

	r.GET("/authenticator/users", func(c *gin.Context) {
		phoneNumber := c.Query("phoneNumber")
		stringID := c.Query("id")

		if len(stringID) != 0 {
			id, err := uuid.Parse(stringID)
			if err != nil {
				c.JSON(400, gin.H{
					"error": "bad id",
				})
				return
			}

			user, err := dbClient.User().FindByID(c, id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(200, user)
			return
		}

		user, err := dbClient.User().FindByPhoneNumber(c, phoneNumber)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, user)
	})

	r.POST("/authenticator/text/login", func(c *gin.Context) {
		var requestBody auth.LoginByTextRequest

		if err := c.Bind(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		code, err := dbClient.TextVerificationCode().Find(c, requestBody.PhoneNumber, requestBody.Code)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := dbClient.TextVerificationCode().MarkUsed(ctx, code.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		user, err := dbClient.User().FindByPhoneNumber(ctx, requestBody.PhoneNumber)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		token, err := tokenClient.New(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errors.Wrap(err, "jwt creation error").Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"user":  user,
			"token": token,
		})
	})

	r.POST("/authenticator/text/register", func(c *gin.Context) {
		var requestBody auth.RegisterByTextRequest

		if err := c.Bind(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		code, err := dbClient.TextVerificationCode().Find(c, requestBody.PhoneNumber, requestBody.Code)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": errors.Wrap(err, "text verification code finding error").Error(),
			})
			return
		}

		if err := dbClient.TextVerificationCode().MarkUsed(ctx, code.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errors.Wrap(err, "text verification code marking error").Error(),
			})
			return
		}

		var userId *uuid.UUID
		if requestBody.UserID != nil {
			id, err := uuid.Parse(*requestBody.UserID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errors.Wrap(err, "shit user id").Error(),
				})
				return
			}

			userId = &id
		}

		user, err := dbClient.User().Insert(ctx, requestBody.Name, requestBody.PhoneNumber, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errors.Wrap(err, "inserting user error").Error(),
			})
			return
		}

		// Set D&D class if provided
		if requestBody.DndClassID != nil && *requestBody.DndClassID != "" {
			dndClassID, err := uuid.Parse(*requestBody.DndClassID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errors.Wrap(err, "invalid dnd class id").Error(),
				})
				return
			}

			// Verify the D&D class exists
			_, err = dbClient.DndClass().GetByID(ctx, dndClassID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errors.Wrap(err, "dnd class not found").Error(),
				})
				return
			}

			// Update user with D&D class
			if err := dbClient.User().UpdateDndClass(ctx, user.ID, dndClassID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": errors.Wrap(err, "updating user dnd class error").Error(),
				})
				return
			}

			// Reload user with D&D class information
			user, err = dbClient.User().FindByIDWithDndClass(ctx, user.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": errors.Wrap(err, "reloading user with dnd class error").Error(),
				})
				return
			}
		}

		token, err := tokenClient.New(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errors.Wrap(err, "jwt creation error").Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"user":  user,
			"token": token,
		})
	})

	r.DELETE("/authenticator/users/:userID", func(c *gin.Context) {
		stringUserID := c.Param("userID")

		userID, err := uuid.Parse(stringUserID)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := dbClient.User().Delete(c, userID); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{"message": "everything ok"})
	})

	r.GET("/authenticator/get-all-users", func(c *gin.Context) {
		users, err := dbClient.User().FindAll(c)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, users)
	})

	r.POST("/authenticator/get-users", func(c *gin.Context) {
		var getUsersRequest struct {
			UserIDs []string `json:"userIds" binding:"required"`
		}

		if err := c.Bind(&getUsersRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit get users request",
			})
			return
		}

		var userIDs []uuid.UUID
		for _, id := range getUsersRequest.UserIDs {
			userID, err := uuid.Parse(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "shit user id",
				})
				return
			}
			userIDs = append(userIDs, userID)
		}

		users, err := dbClient.User().FindUsersByIDs(c, userIDs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, users)
	})

	r.GET("/authenticator/dnd-classes", func(c *gin.Context) {
		classes, err := dbClient.DndClass().GetAll(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, classes)
	})

	r.Run(":8089")
}
