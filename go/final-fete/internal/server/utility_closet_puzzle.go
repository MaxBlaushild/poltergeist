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
// Implements the V.A.C. puzzle rules:
// - Off (0) -> Green (2) -> Red (4) -> Blue (1) progression
// - Blue only unlocks after all lights have been green at least once
// - Purple (5) requires blue next to red
// - White (3) requires all 6 lights to be purple
// - Sync rule: if 5 lights are same color, clicking 6th syncs it
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

	// Get current state of all buttons
	currentHues := make([]int, 6)
	for i := 0; i < 6; i++ {
		currentHues[i] = puzzle.GetButtonCurrentHue(i)
	}

	// Check if all lights have been green at least once
	// This is true if any light has progressed beyond green (is red, blue, purple, or white)
	// OR if all 6 lights are currently green (which unlocks blue)
	allGreensAchieved := false
	allCurrentlyGreen := true
	for i := 0; i < 6; i++ {
		if currentHues[i] != 2 {
			allCurrentlyGreen = false
		}
		// If any light is red (4), blue (1), white (3), or purple (5), we've had all greens unlocked
		if currentHues[i] == 1 || currentHues[i] == 3 || currentHues[i] == 4 || currentHues[i] == 5 {
			allGreensAchieved = true
		}
	}
	// If all are currently green, this also unlocks blue
	if allCurrentlyGreen {
		allGreensAchieved = true
	}

	// Don't allow actions if puzzle is already in gold (success) state
	if currentHues[0] == 6 || currentHues[1] == 6 || currentHues[2] == 6 ||
		currentHues[3] == 6 || currentHues[4] == 6 || currentHues[5] == 6 {
		// Puzzle is already solved, no further actions allowed
		ctx.JSON(http.StatusOK, puzzle)
		return
	}

	// Check sync rule: if 5 lights are the same color, clicking 6th syncs it
	syncColor := checkSyncRule(currentHues, req.Slot)
	if syncColor != nil {
		newHue := *syncColor
		puzzle.SetButtonCurrentHue(req.Slot, newHue)
		updateLightColor(ctx, s, puzzle, req.Slot, newHue)

		// Check for success condition after sync
		successOrder := []int{0, 1, 2, 3, 4, 5}
		successAchieved := true
		for i := 0; i < 6; i++ {
			if puzzle.GetButtonCurrentHue(i) != successOrder[i] {
				successAchieved = false
				break
			}
		}
		if successAchieved {
			for i := 0; i < 6; i++ {
				puzzle.SetButtonCurrentHue(i, 6) // Gold
				updateLightColor(ctx, s, puzzle, i, 6)
			}
		}

		// Save and return
		if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update puzzle: " + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, puzzle)
		return
	}

	// Get the color of the clicked button
	clickedColor := currentHues[req.Slot]

	// Handle special cases based on clicked button color
	switch clickedColor {
	case 0: // Off (Gray) -> Green
		newHue := 2 // Green
		puzzle.SetButtonCurrentHue(req.Slot, newHue)
		updateLightColor(ctx, s, puzzle, req.Slot, newHue)

	case 1: // Blue -> Off (and potentially create purple if next to red)
		// Check neighbors for red
		prevSlot := (req.Slot - 1 + 6) % 6
		nextSlot := (req.Slot + 1) % 6

		if currentHues[prevSlot] == 4 { // Previous is red
			// Turn red into purple
			puzzle.SetButtonCurrentHue(prevSlot, 5) // Purple
			updateLightColor(ctx, s, puzzle, prevSlot, 5)
		}
		if currentHues[nextSlot] == 4 { // Next is red
			// Turn red into purple
			puzzle.SetButtonCurrentHue(nextSlot, 5) // Purple
			updateLightColor(ctx, s, puzzle, nextSlot, 5)
		}

		// Turn blue back to off
		puzzle.SetButtonCurrentHue(req.Slot, 0) // Off
		updateLightColor(ctx, s, puzzle, req.Slot, 0)

	case 2: // Green
		// Normal progression: Green -> Red
		// When all are green, clicking any one unlocks blue (handled above in allGreensAchieved)
		puzzle.SetButtonCurrentHue(req.Slot, 4) // Red
		updateLightColor(ctx, s, puzzle, req.Slot, 4)

	case 3: // White
		// White can only be clicked if all 6 are purple - clicking turns it back
		// But per rules, white is only achievable after all purples, so clicking white might do nothing
		// or cycle it. Let's make it cycle: White -> Off
		puzzle.SetButtonCurrentHue(req.Slot, 0) // Off
		updateLightColor(ctx, s, puzzle, req.Slot, 0)

	case 4: // Red -> Blue (only if all greens achieved) or cycle back to Off
		if allGreensAchieved {
			puzzle.SetButtonCurrentHue(req.Slot, 1) // Blue
			updateLightColor(ctx, s, puzzle, req.Slot, 1)
		} else {
			// Can't become blue yet, stay red or cycle to off?
			// Per rules, red comes after green, so if we can't have blue, maybe cycle to off
			puzzle.SetButtonCurrentHue(req.Slot, 0) // Off
			updateLightColor(ctx, s, puzzle, req.Slot, 0)
		}

	case 5: // Purple
		// Check if all 6 lights are purple - if so, clicking unlocks white
		allPurple := true
		for i := 0; i < 6; i++ {
			if currentHues[i] != 5 {
				allPurple = false
				break
			}
		}

		if allPurple {
			// All purple - clicking one turns it white
			puzzle.SetButtonCurrentHue(req.Slot, 3) // White
			updateLightColor(ctx, s, puzzle, req.Slot, 3)
		} else {
			// Purple can't be clicked directly, so maybe cycle or do nothing
			// Per rules, purple is special, so let's make clicking it cycle back to off
			puzzle.SetButtonCurrentHue(req.Slot, 0) // Off
			updateLightColor(ctx, s, puzzle, req.Slot, 0)
		}
	}

	// Check for success condition: Gray (0) → Blue (1) → Green (2) → White (3) → Red (4) → Purple (5)
	successOrder := []int{0, 1, 2, 3, 4, 5} // Off, Blue, Green, White, Red, Purple
	successAchieved := true
	for i := 0; i < 6; i++ {
		currentHue := puzzle.GetButtonCurrentHue(i)
		if currentHue != successOrder[i] {
			successAchieved = false
			break
		}
	}

	// If success condition is met, turn all lights to gold
	if successAchieved {
		for i := 0; i < 6; i++ {
			puzzle.SetButtonCurrentHue(i, 6) // Gold
			updateLightColor(ctx, s, puzzle, i, 6)
		}
	}

	// Save updated puzzle state
	if err := s.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update puzzle: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, puzzle)
}

// checkSyncRule checks if 5 of 6 lights are the same color
// If so, returns the color that the 6th should sync to
func checkSyncRule(hues []int, clickedSlot int) *int {
	// Count colors excluding the clicked slot
	colorCounts := make(map[int]int)
	for i := 0; i < 6; i++ {
		if i != clickedSlot {
			colorCounts[hues[i]]++
		}
	}

	// Check if any color appears 5 times
	for color, count := range colorCounts {
		if count == 5 {
			return &color
		}
	}

	return nil
}

// updateLightColor updates the Hue light color for a given slot
func updateLightColor(ctx *gin.Context, s *server, puzzle *models.UtilityClosetPuzzle, slot int, hue int) {
	lightID := puzzle.GetButtonHueLightID(slot)
	if lightID != nil && s.hueClient != nil {
		r, g, b := models.ColorIndexToRGB(hue)
		if err := s.hueClient.SetColorRGB(ctx, *lightID, r, g, b); err != nil {
			log.Printf("Warning: Failed to set light %d to color %d for button %d: %v", *lightID, hue, slot, err)
			// Don't fail the request if light update fails
		}
	}
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
