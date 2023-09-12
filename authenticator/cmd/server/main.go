package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MaxBlaushild/authenticator/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/encoding"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
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

	if err := dbClient.Migrate(
		ctx,
		&models.User{},
		&models.Challenge{},
		&models.Credential{},
	); err != nil {
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

	r.POST("/authenticator/nuke", func(c *gin.Context) {
		if err := dbClient.Credential().DeleteAll(c); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := dbClient.User().DeleteAll(c); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{"message": "nuked"})
	})

	r.DELETE("/authenticator/users/:userID", func(c *gin.Context) {
		userID := c.Param("userID")

		uint64Val, err := strconv.ParseUint(userID, 10, 64) // Base 10, BitSize 64
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := dbClient.User().Delete(c, uint(uint64Val)); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{"message": "everything ok"})
	})

	r.DELETE("/authenticator/credentials/:credentialID", func(c *gin.Context) {
		credentialID := c.Param("credentialID")

		uint64Val, err := strconv.ParseUint(credentialID, 10, 64) // Base 10, BitSize 64
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := dbClient.Credential().Delete(c, uint(uint64Val)); err != nil {
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
			UserIDs []uint `json:"userIds" binding:"required"`
		}

		if err := c.Bind(&getUsersRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "shit get users request",
			})
			return
		}

		fmt.Println(getUsersRequest)

		users, err := dbClient.User().FindUsersByIDs(c, getUsersRequest.UserIDs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, users)
	})

	r.POST("/authenticator/registration-options", func(ctx *gin.Context) {
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
		user, err := dbClient.User().Insert(ctx, registerOptionsRequest.Name, registerOptionsRequest.PhoneNumber)
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

	r.POST("/authenticator/register", func(c *gin.Context) {
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

		user, err := dbClient.User().FindByID(c, challenge.AuthUserID)
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

		c.JSON(200, user)
	})

	r.POST("/authenticator/login-options", func(c *gin.Context) {
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

	r.POST("/authenticator/login", func(c *gin.Context) {
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

		user, err := dbClient.User().FindByID(c, challenge.AuthUserID)
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

		c.JSON(200, user)
	})

	r.Run(":8089")
}

// func LoginHandler(c *gin.Context) {
// 	// Typically, retrieve the user first
// 	user := &User{
// 		ID:       1,
// 		Username: "exampleUser",
// 		// ... and other fields
// 	}

// 	// Start authentication
// 	options, sessionData, err := w.BeginLogin(user)
// 	if err != nil {
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Store sessionData, e.g., in a session store or a database.
// 	// ...

// 	c.JSON(200, options)
// }
