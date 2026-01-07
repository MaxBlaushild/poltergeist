package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/MaxBlaushild/poltergeist/final-fete/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

// PressButton handles pressing a button in the utility closet puzzle
func (s *server) PressButton(ctx *gin.Context) {
	slotParam := ctx.Param("slot")
	slot, err := strconv.Atoi(slotParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid slot parameter: " + err.Error()})
		return
	}

	// Use game engine client to handle the button press
	puzzle, err := s.puzzleGameEngineClient.PressButton(ctx, slot)
	if err != nil {
		if err == gameengine.ErrInvalidSlot {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to press button: " + err.Error()})
		return
	}

	// Update Hue lights to match the new puzzle state
	s.updateHueLightsForPuzzle(ctx, puzzle)

	ctx.JSON(http.StatusOK, puzzle)
}

// ResetPuzzle resets the puzzle to its base state
func (s *server) ResetPuzzle(ctx *gin.Context) {
	// Use game engine client to reset the puzzle
	puzzle, err := s.puzzleGameEngineClient.ResetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset puzzle: " + err.Error()})
		return
	}

	// Update all Hue lights to base state colors
	s.updateHueLightsForPuzzle(ctx, puzzle)

	ctx.JSON(http.StatusOK, puzzle)
}

// updateHueLightsForPuzzle updates all Hue lights to match the current puzzle state
func (s *server) updateHueLightsForPuzzle(ctx *gin.Context, puzzle *models.UtilityClosetPuzzle) {
	if s.hueClient == nil {
		return
	}

	for slot := 0; slot < 6; slot++ {
		lightID := puzzle.GetButtonHueLightID(slot)
		if lightID != nil {
			currentHue := puzzle.GetButtonCurrentHue(slot)
			r, g, b := models.ColorIndexToRGB(currentHue)
			if err := s.hueClient.SetColorRGB(ctx, *lightID, r, g, b); err != nil {
				log.Printf("Warning: Failed to set light %d to color %d for button %d: %v", *lightID, currentHue, slot, err)
				// Don't fail the request if light update fails
			}
		}
	}
}

// GetPuzzleState returns the current state of the puzzle
func (s *server) GetPuzzleState(ctx *gin.Context) {
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, puzzle)
}

type UpdatePuzzleRequest struct {
	Buttons []ButtonConfig `json:"buttons" binding:"required,dive"`
}

type ButtonConfig struct {
	Slot       int  `json:"slot" binding:"min=0,max=5"`
	HueLightID *int `json:"hueLightId"`
	BaseHue    int  `json:"baseHue" binding:"min=0,max=5"`
}

// UpdatePuzzle updates the puzzle configuration (light IDs and base hues)
func (s *server) UpdatePuzzle(ctx *gin.Context) {
	var req UpdatePuzzleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that we have exactly 6 buttons
	if len(req.Buttons) != 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must provide exactly 6 button configurations"})
		return
	}

	// Validate that slots are unique and 0-5
	slotMap := make(map[int]bool)
	for _, button := range req.Buttons {
		if slotMap[button.Slot] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "duplicate slot found"})
			return
		}
		if button.Slot < 0 || button.Slot > 5 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "slot must be between 0 and 5"})
			return
		}
		slotMap[button.Slot] = true
	}

	// Get puzzle instance
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	// Update each button configuration
	for _, button := range req.Buttons {
		puzzle.SetButtonHueLightID(button.Slot, button.HueLightID)
		puzzle.SetButtonBaseHue(button.Slot, button.BaseHue)
	}

	// Save updated puzzle
	if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update puzzle: " + err.Error()})
		return
	}

	// Update Hue lights to base state colors
	if s.hueClient != nil {
		for _, button := range req.Buttons {
			if button.HueLightID != nil {
				r, g, b := models.ColorIndexToRGB(button.BaseHue)
				if err := s.hueClient.SetColorRGB(ctx, *button.HueLightID, r, g, b); err != nil {
					log.Printf("Warning: Failed to set light %d to base color %d for button %d: %v", *button.HueLightID, button.BaseHue, button.Slot, err)
					// Don't fail the request if light update fails
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, puzzle)
}

// Admin CRUD endpoints

// AdminGetPuzzleState returns the current state of the puzzle (admin endpoint)
func (s *server) AdminGetPuzzleState(ctx *gin.Context) {
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, puzzle)
}

type AdminCreatePuzzleRequest struct {
	Buttons []AdminButtonConfig `json:"buttons" binding:"required,dive"`
}

type AdminButtonConfig struct {
	Slot       int  `json:"slot" binding:"min=0,max=5"`
	HueLightID *int `json:"hueLightId"`
	BaseHue    int  `json:"baseHue" binding:"min=0,max=6"`
	CurrentHue int  `json:"currentHue" binding:"min=0,max=6"`
}

// AdminCreatePuzzle creates a new puzzle instance (admin endpoint)
func (s *server) AdminCreatePuzzle(ctx *gin.Context) {
	var req AdminCreatePuzzleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that we have exactly 6 buttons
	if len(req.Buttons) != 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must provide exactly 6 button configurations"})
		return
	}

	// Validate that slots are unique and 0-5
	slotMap := make(map[int]bool)
	for _, button := range req.Buttons {
		if slotMap[button.Slot] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "duplicate slot found"})
			return
		}
		if button.Slot < 0 || button.Slot > 5 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "slot must be between 0 and 5"})
			return
		}
		slotMap[button.Slot] = true
	}

	// Check if puzzle already exists
	existingPuzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err == nil && existingPuzzle != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "puzzle already exists. Use PUT to update or DELETE to remove first"})
		return
	}

	// Create new puzzle
	puzzle := &models.UtilityClosetPuzzle{}
	for _, button := range req.Buttons {
		puzzle.SetButtonHueLightID(button.Slot, button.HueLightID)
		puzzle.SetButtonBaseHue(button.Slot, button.BaseHue)
		puzzle.SetButtonCurrentHue(button.Slot, button.CurrentHue)
	}

	// Save puzzle
	if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, puzzle)
}

type AdminUpdatePuzzleRequest struct {
	Buttons            []AdminButtonConfig `json:"buttons" binding:"required,dive"`
	AllGreensAchieved  *bool               `json:"allGreensAchieved"`
	AllPurplesAchieved *bool               `json:"allPurplesAchieved"`
}

// AdminUpdatePuzzle updates the puzzle with comprehensive admin controls (admin endpoint)
func (s *server) AdminUpdatePuzzle(ctx *gin.Context) {
	var req AdminUpdatePuzzleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that we have exactly 6 buttons
	if len(req.Buttons) != 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must provide exactly 6 button configurations"})
		return
	}

	// Validate that slots are unique and 0-5
	slotMap := make(map[int]bool)
	for _, button := range req.Buttons {
		if slotMap[button.Slot] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "duplicate slot found"})
			return
		}
		if button.Slot < 0 || button.Slot > 5 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "slot must be between 0 and 5"})
			return
		}
		slotMap[button.Slot] = true
	}

	// Get puzzle instance
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	// Update each button configuration
	for _, button := range req.Buttons {
		puzzle.SetButtonHueLightID(button.Slot, button.HueLightID)
		puzzle.SetButtonBaseHue(button.Slot, button.BaseHue)
		puzzle.SetButtonCurrentHue(button.Slot, button.CurrentHue)
	}

	// Update allGreensAchieved if provided
	if req.AllGreensAchieved != nil {
		puzzle.AllGreensAchieved = *req.AllGreensAchieved
	}

	// Update allPurplesAchieved if provided
	if req.AllPurplesAchieved != nil {
		puzzle.AllPurplesAchieved = *req.AllPurplesAchieved
	}

	// Save updated puzzle
	if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update puzzle: " + err.Error()})
		return
	}

	// Update Hue lights to current colors
	if s.hueClient != nil {
		for _, button := range req.Buttons {
			if button.HueLightID != nil {
				r, g, b := models.ColorIndexToRGB(button.CurrentHue)
				if err := s.hueClient.SetColorRGB(ctx, *button.HueLightID, r, g, b); err != nil {
					log.Printf("Warning: Failed to set light %d to color %d for button %d: %v", *button.HueLightID, button.CurrentHue, button.Slot, err)
					// Don't fail the request if light update fails
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, puzzle)
}

// AdminDeletePuzzle deletes the puzzle (admin endpoint)
func (s *server) AdminDeletePuzzle(ctx *gin.Context) {
	// Get puzzle instance
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	// Delete puzzle from database
	if err := s.dbClient.UtilityClosetPuzzle().DeletePuzzle(ctx, puzzle.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "puzzle deleted successfully"})
}

type ToggleAchievementRequest struct {
	AchievementType string `json:"achievementType" binding:"required,oneof=allGreens allPurples"`
}

// ToggleAchievement toggles the allGreensAchieved or allPurplesAchieved state
func (s *server) ToggleAchievement(ctx *gin.Context) {
	var req ToggleAchievementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get puzzle instance
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	// Toggle the appropriate achievement state
	if req.AchievementType == "allGreens" {
		puzzle.AllGreensAchieved = !puzzle.AllGreensAchieved
	} else if req.AchievementType == "allPurples" {
		puzzle.AllPurplesAchieved = !puzzle.AllPurplesAchieved
	}

	// Save updated puzzle
	if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, puzzle)
}
