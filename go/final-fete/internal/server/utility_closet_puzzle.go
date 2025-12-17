package server

import (
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type PressButtonRequest struct {
	Slot int `json:"slot" binding:"required"`
}

// PressButton handles pressing a button in the utility closet puzzle
// Pressing button N affects buttons (N-1 mod 6), N, and (N+1 mod 6)
// Each affected button cycles to the next color (0->1->2->3->4->5->0)
func (s *server) PressButton(ctx *gin.Context) {
	var req PressButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate slot (0-5)
	if req.Slot < 0 || req.Slot > 5 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "slot must be between 0 and 5"})
		return
	}

	// Get puzzle instance
	puzzle, err := s.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get puzzle: " + err.Error()})
		return
	}

	// Determine which buttons are affected (N-1, N, N+1, wrapping around)
	affectedSlots := []int{
		(req.Slot - 1 + 6) % 6, // Previous button (wrapping)
		req.Slot,               // Current button
		(req.Slot + 1) % 6,     // Next button (wrapping)
	}

	// Update affected buttons: cycle to next color (0->1->2->3->4->5->0)
	for _, slot := range affectedSlots {
		currentHue := puzzle.GetButtonCurrentHue(slot)
		nextHue := (currentHue + 1) % 6
		puzzle.SetButtonCurrentHue(slot, nextHue)

		// Update Hue light if configured
		lightID := puzzle.GetButtonHueLightID(slot)
		if lightID != nil && s.hueClient != nil {
			r, g, b := models.ColorIndexToRGB(nextHue)
			if err := s.hueClient.SetColorRGB(ctx, *lightID, r, g, b); err != nil {
				log.Printf("Warning: Failed to set light %d to color %d for button %d: %v", *lightID, nextHue, slot, err)
				// Don't fail the request if light update fails
			}
		}
	}

	// Save updated puzzle state
	if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, puzzle)
}

// ResetPuzzle resets the puzzle to its base state
func (s *server) ResetPuzzle(ctx *gin.Context) {
	// Reset puzzle in database
	puzzle, err := s.dbClient.UtilityClosetPuzzle().ResetPuzzle(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset puzzle: " + err.Error()})
		return
	}

	// Update all Hue lights to base state colors
	if s.hueClient != nil {
		for slot := 0; slot < 6; slot++ {
			lightID := puzzle.GetButtonHueLightID(slot)
			if lightID != nil {
				baseHue := puzzle.GetButtonBaseHue(slot)
				r, g, b := models.ColorIndexToRGB(baseHue)
				if err := s.hueClient.SetColorRGB(ctx, *lightID, r, g, b); err != nil {
					log.Printf("Warning: Failed to set light %d to base color %d for button %d: %v", *lightID, baseHue, slot, err)
					// Don't fail the request if light update fails
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, puzzle)
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
	Slot       int  `json:"slot" binding:"required,min=0,max=5"`
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
