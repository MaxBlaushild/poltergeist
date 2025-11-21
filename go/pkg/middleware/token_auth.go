package middleware

import (
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/http"
	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/gin-gonic/gin"
)

const (
	bearer = "Bearer"
)

func WithAuthentication(authClient auth.Client, livenessClient liveness.LivenessClient, next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.Request.Header.Get("Authorization")
		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != bearer {
			ctx.JSON(401, http.ErrorResponse{
				Error: "invalid authorization header",
			})
			return
		}

		user, err := authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{
			Token: headerParts[1],
		})
		if err != nil {
			ctx.JSON(401, http.ErrorResponse{
				Error: "authorization header not valid",
			})
			return
		}

		// Extract and save user location if provided
		locationHeader := ctx.Request.Header.Get("X-User-Location")
		log.Printf("[DEBUG] Location header: %s", locationHeader)
		if locationHeader != "" {
			if err = livenessClient.SetUserLocation(ctx, user.ID, locationHeader); err != nil {
				log.Println("error setting user location", err)
			}
		}

		ctx.Set("user", user)

		next(ctx)
	}
}

func WithAuthenticationWithoutLocation(authClient auth.Client, next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.Request.Header.Get("Authorization")
		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != bearer {
			ctx.JSON(401, http.ErrorResponse{
				Error: "invalid authorization header",
			})
			return
		}

		user, err := authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{
			Token: headerParts[1],
		})
		if err != nil {
			ctx.JSON(401, http.ErrorResponse{
				Error: "authorization header not valid",
			})
			return
		}

		ctx.Set("user", user)

		next(ctx)
	}
}
