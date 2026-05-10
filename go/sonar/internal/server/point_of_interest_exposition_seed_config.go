package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type pointOfInterestExpositionSeedConfigResponse struct {
	ID       int                                            `json:"id"`
	Profiles []pointOfInterestExpositionSeedProfileResponse `json:"profiles"`
}

type pointOfInterestExpositionSeedProfileResponse struct {
	Category                     string `json:"category"`
	Label                        string `json:"label"`
	FirstSpawnChanceBasisPoints  int    `json:"firstSpawnChanceBasisPoints"`
	SecondSpawnChanceBasisPoints int    `json:"secondSpawnChanceBasisPoints"`
}

func buildPointOfInterestExpositionSeedConfigResponse(
	config *models.PointOfInterestExpositionSeedConfig,
) pointOfInterestExpositionSeedConfigResponse {
	if config == nil {
		config = &models.PointOfInterestExpositionSeedConfig{
			ID:       1,
			Profiles: models.DefaultPointOfInterestExpositionSeedProfiles(),
		}
	}

	resolvedProfiles := models.ResolvePointOfInterestExpositionSeedProfiles(config.Profiles)
	profiles := make([]pointOfInterestExpositionSeedProfileResponse, 0, len(resolvedProfiles))
	for _, profile := range resolvedProfiles {
		profiles = append(profiles, pointOfInterestExpositionSeedProfileResponse{
			Category:                     string(profile.Category),
			Label:                        models.PointOfInterestMarkerCategoryLabel(profile.Category),
			FirstSpawnChanceBasisPoints:  profile.FirstSpawnChanceBasisPoints,
			SecondSpawnChanceBasisPoints: profile.SecondSpawnChanceBasisPoints,
		})
	}

	return pointOfInterestExpositionSeedConfigResponse{
		ID:       config.ID,
		Profiles: profiles,
	}
}

func (s *server) getPointOfInterestExpositionSeedConfig(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	config, err := s.dbClient.PointOfInterestExpositionSeedConfig().Get(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, buildPointOfInterestExpositionSeedConfigResponse(config))
}

func (s *server) updatePointOfInterestExpositionSeedConfig(ctx *gin.Context) {
	_, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Profiles []models.PointOfInterestExpositionSeedProfile `json:"profiles"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := &models.PointOfInterestExpositionSeedConfig{
		Profiles: requestBody.Profiles,
	}
	updated, err := s.dbClient.PointOfInterestExpositionSeedConfig().Upsert(ctx, config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, buildPointOfInterestExpositionSeedConfigResponse(updated))
}
