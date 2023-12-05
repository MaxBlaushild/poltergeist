package middleware

import (
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/http"
	"github.com/gin-gonic/gin"
)

const (
	bearer = "Bearer"
)

func WithAuthentication(authClient auth.Client, next gin.HandlerFunc) gin.HandlerFunc {
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
