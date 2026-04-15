package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const pointOfInterestMarkerCategoryBackfillPreviewLimit = 50

type pointOfInterestMarkerCategoryBackfillUpdate struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	PreviousCategory string    `json:"previousCategory"`
	MarkerCategory   string    `json:"markerCategory"`
	WasDerived       bool      `json:"wasDerived"`
	Source           string    `json:"source"`
}

func pointOfInterestMarkerCategoryFromGooglePlace(
	place *googlemaps.Place,
) models.PointOfInterestMarkerCategory {
	if place == nil {
		return models.PointOfInterestMarkerCategoryGeneric
	}
	return models.InferPointOfInterestMarkerCategory(
		place.PrimaryType,
		place.Types,
		place.PrimaryTypeDisplayName.Text,
		place.ServesCoffee != nil && *place.ServesCoffee,
		place.ServesBeer != nil && *place.ServesBeer,
		place.ServesWine != nil && *place.ServesWine,
		place.ServesCocktails != nil && *place.ServesCocktails,
		place.LiveMusic != nil && *place.LiveMusic,
	)
}

func (s *server) resolvePointOfInterestMarkerCategoryForBackfill(
	pointOfInterest *models.PointOfInterest,
) (models.PointOfInterestMarkerCategory, string, string) {
	if pointOfInterest == nil {
		return models.PointOfInterestMarkerCategoryGeneric, "local", "point of interest was nil"
	}

	if pointOfInterest.GoogleMapsPlaceID != nil {
		googleMapsPlaceID := strings.TrimSpace(*pointOfInterest.GoogleMapsPlaceID)
		if googleMapsPlaceID != "" {
			place, err := s.googlemapsClient.FindPlaceByID(googleMapsPlaceID)
			if err == nil && place != nil {
				return pointOfInterestMarkerCategoryFromGooglePlace(place), "google", ""
			}

			warning := "failed to fetch Google place for marker category backfill"
			if err != nil {
				warning += ": " + err.Error()
			} else {
				warning += ": place not found"
			}
			return models.InferPointOfInterestMarkerCategoryFromPointOfInterest(pointOfInterest), "local_fallback", warning
		}
	}

	return models.InferPointOfInterestMarkerCategoryFromPointOfInterest(pointOfInterest), "local", ""
}

func (s *server) backfillPointOfInterestMarkerCategories(ctx *gin.Context) {
	var requestBody struct {
		ZoneID *uuid.UUID `json:"zoneId"`
		Limit  int        `json:"limit"`
		Force  bool       `json:"force"`
		DryRun bool       `json:"dryRun"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Limit < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit must be zero or greater"})
		return
	}

	var (
		pointsOfInterest []models.PointOfInterest
		err              error
	)
	zoneIDProvided := requestBody.ZoneID != nil && *requestBody.ZoneID != uuid.Nil
	if zoneIDProvided {
		pointsOfInterest, err = s.dbClient.PointOfInterest().FindAllForZone(ctx, *requestBody.ZoneID)
	} else {
		pointsOfInterest, err = s.dbClient.PointOfInterest().FindAll(ctx)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	processedCount := 0
	skippedAlreadySetCount := 0
	updatedCount := 0
	unchangedCount := 0
	googleDerivedCount := 0
	localDerivedCount := 0
	googleFallbackCount := 0
	warningCount := 0
	limitReached := false

	updates := make([]pointOfInterestMarkerCategoryBackfillUpdate, 0, pointOfInterestMarkerCategoryBackfillPreviewLimit)
	warnings := make([]string, 0, pointOfInterestMarkerCategoryBackfillPreviewLimit)

	addWarning := func(message string) {
		if strings.TrimSpace(message) == "" {
			return
		}
		warningCount++
		if len(warnings) < pointOfInterestMarkerCategoryBackfillPreviewLimit {
			warnings = append(warnings, message)
		}
	}

	for i := range pointsOfInterest {
		pointOfInterest := &pointsOfInterest[i]
		if !requestBody.Force && !pointOfInterest.MarkerCategoryDerived {
			skippedAlreadySetCount++
			continue
		}
		if requestBody.Limit > 0 && processedCount >= requestBody.Limit {
			limitReached = true
			break
		}
		processedCount++

		previousCategory := models.NormalizePointOfInterestMarkerCategory(string(pointOfInterest.MarkerCategory))
		markerCategory, source, warning := s.resolvePointOfInterestMarkerCategoryForBackfill(pointOfInterest)
		addWarning(warning)

		if !pointOfInterest.MarkerCategoryDerived && previousCategory == markerCategory {
			unchangedCount++
			continue
		}

		if !requestBody.DryRun {
			if err := s.dbClient.PointOfInterest().UpdateMarkerCategory(ctx, pointOfInterest.ID, markerCategory); err != nil {
				addWarning("failed to update marker category for " + pointOfInterest.ID.String() + ": " + err.Error())
				continue
			}
		}

		updatedCount++
		switch source {
		case "google":
			googleDerivedCount++
		case "local_fallback":
			googleFallbackCount++
			localDerivedCount++
		default:
			localDerivedCount++
		}

		if len(updates) < pointOfInterestMarkerCategoryBackfillPreviewLimit {
			updates = append(updates, pointOfInterestMarkerCategoryBackfillUpdate{
				ID:               pointOfInterest.ID,
				Name:             pointOfInterest.Name,
				PreviousCategory: string(previousCategory),
				MarkerCategory:   string(markerCategory),
				WasDerived:       pointOfInterest.MarkerCategoryDerived,
				Source:           source,
			})
		}
	}

	response := gin.H{
		"dryRun":                 requestBody.DryRun,
		"force":                  requestBody.Force,
		"limit":                  requestBody.Limit,
		"limitReached":           limitReached,
		"totalPointsOfInterest":  len(pointsOfInterest),
		"processedCount":         processedCount,
		"skippedAlreadySetCount": skippedAlreadySetCount,
		"updatedCount":           updatedCount,
		"unchangedCount":         unchangedCount,
		"googleDerivedCount":     googleDerivedCount,
		"localDerivedCount":      localDerivedCount,
		"googleFallbackCount":    googleFallbackCount,
		"warningCount":           warningCount,
	}
	if zoneIDProvided {
		response["zoneId"] = requestBody.ZoneID.String()
	}
	if len(updates) > 0 {
		response["updates"] = updates
	}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	ctx.JSON(http.StatusOK, response)
}
