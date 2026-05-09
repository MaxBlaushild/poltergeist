package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func proximityBypassEnabled(ctx context.Context) bool {
	return middleware.DebugProximityBypassEnabled(ctx)
}

func (s *server) requireProximityWithin(
	ctx *gin.Context,
	distanceMeters float64,
	maxDistanceMeters float64,
	subject string,
) bool {
	if proximityBypassEnabled(ctx.Request.Context()) {
		return true
	}
	if distanceMeters <= maxDistanceMeters {
		return true
	}
	ctx.JSON(http.StatusBadRequest, gin.H{
		"error": fmt.Sprintf(
			"you must be within %.0f meters of %s. Currently %.0f meters away",
			maxDistanceMeters,
			subject,
			distanceMeters,
		),
	})
	return false
}

func (s *server) requireChallengeAreaAccess(
	ctx *gin.Context,
	insideArea bool,
) bool {
	if proximityBypassEnabled(ctx.Request.Context()) || insideArea {
		return true
	}
	ctx.JSON(
		http.StatusBadRequest,
		gin.H{"error": "you must be inside the challenge area to submit an answer"},
	)
	return false
}
