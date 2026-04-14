package server

import (
	stdErrors "errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type locationUpdateRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func bindLocationUpdate(ctx *gin.Context) (float64, float64, bool) {
	var body locationUpdateRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 0, 0, false
	}
	if body.Latitude == nil || body.Longitude == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return 0, 0, false
	}
	return *body.Latitude, *body.Longitude, true
}

func (s *server) updatePointOfInterestLocation(ctx *gin.Context) {
	pointOfInterestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest ID"})
		return
	}

	existing, err := s.dbClient.PointOfInterest().FindByID(ctx, pointOfInterestID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "point of interest not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "point of interest not found"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Lat = strconv.FormatFloat(latitude, 'f', 6, 64)
	existing.Lng = strconv.FormatFloat(longitude, 'f', 6, 64)
	if err := s.dbClient.PointOfInterest().Update(ctx, pointOfInterestID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.PointOfInterest().FindByID(ctx, pointOfInterestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateTreasureChestLocation(ctx *gin.Context) {
	treasureChestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid treasure chest ID"})
		return
	}

	existing, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "treasure chest not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "treasure chest not found"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Latitude = latitude
	existing.Longitude = longitude
	if err := s.dbClient.TreasureChest().Update(ctx, treasureChestID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update treasure chest location: " + err.Error()})
		return
	}

	updated, err := s.dbClient.TreasureChest().FindByID(ctx, treasureChestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateHealingFountainLocation(ctx *gin.Context) {
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}

	existing, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Latitude = latitude
	existing.Longitude = longitude
	if err := s.dbClient.HealingFountain().Update(ctx, fountainID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update healing fountain location: " + err.Error()})
		return
	}

	updated, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateMonsterEncounterLocation(ctx *gin.Context) {
	encounterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster encounter ID"})
		return
	}

	existing, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounterID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "monster encounter not found"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Latitude = latitude
	existing.Longitude = longitude
	if err := s.dbClient.MonsterEncounter().Update(ctx, encounterID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update monster encounter location: " + err.Error()})
		return
	}

	updated, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateScenarioLocation(ctx *gin.Context) {
	scenarioID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	existing, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Latitude = latitude
	existing.Longitude = longitude
	if err := s.dbClient.Scenario().Update(ctx, scenarioID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update scenario location: " + err.Error()})
		return
	}

	updated, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateExpositionLocation(ctx *gin.Context) {
	expositionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition ID"})
		return
	}

	existing, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Latitude = latitude
	existing.Longitude = longitude
	if err := s.dbClient.Exposition().Update(ctx, expositionID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update exposition location: " + err.Error()})
		return
	}

	updated, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateChallengeLocation(ctx *gin.Context) {
	challengeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge ID"})
		return
	}

	existing, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}
	if existing.HasPolygon() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "polygon challenges must be repositioned by editing their polygon"})
		return
	}

	latitude, longitude, ok := bindLocationUpdate(ctx)
	if !ok {
		return
	}

	existing.Latitude = latitude
	existing.Longitude = longitude
	if err := s.dbClient.Challenge().Update(ctx, challengeID, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update challenge location: " + err.Error()})
		return
	}

	updated, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}
