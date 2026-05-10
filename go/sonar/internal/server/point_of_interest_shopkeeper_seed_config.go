package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type pointOfInterestShopkeeperSeedConfigResponse struct {
	ID       int                                            `json:"id"`
	Profiles []pointOfInterestShopkeeperSeedProfileResponse `json:"profiles"`
}

type pointOfInterestShopkeeperSeedProfileResponse struct {
	Category               string                                          `json:"category"`
	Label                  string                                          `json:"label"`
	SpawnChanceBasisPoints int                                             `json:"spawnChanceBasisPoints"`
	Candidates             []models.PointOfInterestShopkeeperSeedCandidate `json:"candidates"`
}

func buildPointOfInterestShopkeeperSeedConfigResponse(
	config *models.PointOfInterestShopkeeperSeedConfig,
) pointOfInterestShopkeeperSeedConfigResponse {
	if config == nil {
		config = &models.PointOfInterestShopkeeperSeedConfig{
			ID:       1,
			Profiles: models.DefaultPointOfInterestShopkeeperSeedProfiles(),
		}
	}

	resolvedProfiles := models.ResolvePointOfInterestShopkeeperSeedProfiles(config.Profiles)
	profiles := make([]pointOfInterestShopkeeperSeedProfileResponse, 0, len(resolvedProfiles))
	for _, profile := range resolvedProfiles {
		profiles = append(profiles, pointOfInterestShopkeeperSeedProfileResponse{
			Category:               string(profile.Category),
			Label:                  models.PointOfInterestMarkerCategoryLabel(profile.Category),
			SpawnChanceBasisPoints: profile.SpawnChanceBasisPoints,
			Candidates:             append([]models.PointOfInterestShopkeeperSeedCandidate{}, profile.Candidates...),
		})
	}

	return pointOfInterestShopkeeperSeedConfigResponse{
		ID:       config.ID,
		Profiles: profiles,
	}
}

func (s *server) getPointOfInterestShopkeeperSeedConfig(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config, err := s.dbClient.PointOfInterestShopkeeperSeedConfig().Get(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, buildPointOfInterestShopkeeperSeedConfigResponse(config))
}

func (s *server) updatePointOfInterestShopkeeperSeedConfig(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Profiles []models.PointOfInterestShopkeeperSeedProfile `json:"profiles"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := &models.PointOfInterestShopkeeperSeedConfig{
		Profiles: requestBody.Profiles,
	}
	updated, err := s.dbClient.PointOfInterestShopkeeperSeedConfig().Upsert(ctx, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, buildPointOfInterestShopkeeperSeedConfigResponse(updated))
}
