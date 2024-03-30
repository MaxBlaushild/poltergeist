package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MaxBlaushild/authenticator/internal/config"
	"github.com/MaxBlaushild/authenticator/internal/token"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/encoding"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
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

	wconfig := &webauthn.Config{
		RPDisplayName: cfg.Public.RpDisplayName,
		RPID:          cfg.Public.RpID,
		RPOrigins:     []string{cfg.Public.RpOrigin},
	}

	webauthnClient, err := webauthn.New(wconfig)
	if err != nil {
		panic(err)
	}

	tokenClient, err := token.NewClient(cfg.Secret.AuthPrivateKey)
	if err != nil {
		panic(err)
	}

	texterClient := texter.NewClient()

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

	r.GET("/authenticator/credentials", func(c *gin.Context) {
		credentials, err := dbClient.Credential().FindAll(c)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, credentials)
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

	r.POST("/authenticator/webauthn/registration-options", func(ctx *gin.Context) {
		var registerOptionsRequest struct {
			Name        string `json:"name" binding:"required"`
			PhoneNumber string `json:"phoneNumber" binding:"required"`
		}

		if err := ctx.Bind(&registerOptionsRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "shit register options request",
			})
			return
		}
		user, err := dbClient.User().Insert(ctx, registerOptionsRequest.Name, registerOptionsRequest.PhoneNumber, nil)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		options, sessionData, err := webauthnClient.BeginRegistration(
			user,
			webauthn.WithCredentialParameters([]protocol.CredentialParameter{
				{
					Type:      "public-key",
					Algorithm: -7,
				},
				{
					Type:      "public-key",
					Algorithm: -257,
				},
			}),
		)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if err := dbClient.Challenge().Insert(ctx, sessionData.Challenge, user.ID); err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, options)
	})

	r.POST("/authenticator/webauthn/register", func(c *gin.Context) {
		response, err := protocol.ParseCredentialCreationResponseBody(c.Request.Body)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		chal := response.Response.CollectedClientData.Challenge

		challenge, err := dbClient.Challenge().Find(c, chal)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user, err := dbClient.User().FindByID(c, challenge.UserID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		credential, err := webauthnClient.CreateCredential(user, webauthn.SessionData{
			Challenge: challenge.Challenge,
			UserID:    user.WebAuthnID(),
		}, response)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			return
		}

		_, err = dbClient.Credential().Insert(
			c,
			encoding.BytesToBase64UrlEncoded(credential.ID),
			encoding.BytesToBase64UrlEncoded(credential.PublicKey),
			user.ID,
		)
		if err != nil {

			c.JSON(500, gin.H{"error": err.Error()})
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

	r.POST("/authenticator/webauthn/login-options", func(c *gin.Context) {
		var loginOptions struct {
			PhoneNumber string `json:"phoneNumber" binding:"required"`
		}

		if err := c.Bind(&loginOptions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "login options request",
			})
			return
		}

		user, err := dbClient.User().FindByPhoneNumber(c, loginOptions.PhoneNumber)
		if err != nil {
			c.JSON(404, gin.H{
				"error": err.Error(),
			})
			return
		}

		options, session, err := webauthnClient.BeginLogin(user)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err = dbClient.Challenge().Insert(c, session.Challenge, user.ID); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, options)
	})

	r.POST("/authenticator/webauthn/login", func(c *gin.Context) {
		response, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		chal := response.Response.CollectedClientData.Challenge

		challenge, err := dbClient.Challenge().Find(c, chal)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user, err := dbClient.User().FindByID(c, challenge.UserID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		_, err = webauthnClient.ValidateLogin(user, webauthn.SessionData{
			Challenge: challenge.Challenge,
			UserID:    user.WebAuthnID(),
		}, response)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
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

	r.Run(":8089")
}
