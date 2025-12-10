package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

func (s *server) GetTrendingDestinations(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	trendingHandler := s.dbClient.TrendingDestination()

	// Get top 5 cities
	cities, err := trendingHandler.FindByType(ctx, models.LocationTypeCity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch trending cities: " + err.Error(),
		})
		return
	}

	// Get top 5 countries
	countries, err := trendingHandler.FindByType(ctx, models.LocationTypeCountry)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch trending countries: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"cities":    cities,
		"countries": countries,
	})
}
