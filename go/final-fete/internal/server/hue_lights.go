package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HueLightResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (s *server) getAllHueLights(ctx *gin.Context) {
	if s.hueClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Hue client not configured"})
		return
	}

	lights, err := s.hueClient.GetLights(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lights: " + err.Error()})
		return
	}

	// Convert to simplified response
	response := make([]HueLightResponse, 0, len(lights))
	for _, light := range lights {
		response = append(response, HueLightResponse{
			ID:   light.ID,
			Name: light.Name,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

