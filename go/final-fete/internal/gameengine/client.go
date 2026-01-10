package gameengine

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

// UtilityClosetPuzzleClient handles the game logic for the utility closet puzzle
type UtilityClosetPuzzleClient interface {
	PressButton(ctx context.Context, slot int) (*models.UtilityClosetPuzzle, error)
	ResetPuzzle(ctx context.Context) (*models.UtilityClosetPuzzle, error)
}

type utilityClosetPuzzleClient struct {
	dbClient db.DbClient
}

func NewUtilityClosetPuzzleClient(dbClient db.DbClient) UtilityClosetPuzzleClient {
	return &utilityClosetPuzzleClient{
		dbClient: dbClient,
	}
}

// PressButton handles pressing a button in the utility closet puzzle
// Implements the V.A.C. puzzle rules:
// - Off (0) -> Red (2) -> Yellow (4) -> Blue (1) progression
// - Blue only unlocks after all lights have been red at least once
// - Purple (5) requires blue next to yellow
// - White (3) requires all 6 lights to be purple
// - Sync rule: if 5 lights are same color, clicking 6th syncs it
func (c *utilityClosetPuzzleClient) PressButton(ctx context.Context, slot int) (*models.UtilityClosetPuzzle, error) {
	// Validate slot (0-5)
	if slot < 0 || slot > 5 {
		return nil, ErrInvalidSlot
	}

	// Get puzzle instance
	puzzle, err := c.dbClient.UtilityClosetPuzzle().GetPuzzle(ctx)
	if err != nil {
		return nil, err
	}

	// Get current state of all buttons
	currentHues := make([]int, 6)
	for i := 0; i < 6; i++ {
		currentHues[i] = puzzle.GetButtonCurrentHue(i)
	}

	// Check if all lights have been red at least once
	allGreensAchieved, allPurplesAchieved := c.updateAchievementStates(puzzle, currentHues)

	// Don't allow actions if puzzle is already in green (success) state
	if c.isPuzzleSolved(currentHues) {
		return puzzle, nil
	}

	// Check sync rule: if 5 lights are the same color, clicking 6th syncs it
	// Only apply sync rule if the clicked button's color is different from the sync color
	// But don't apply sync if there's exactly one button of a different color (to allow normal progression)
	syncColor := checkSyncRule(currentHues, slot)
	if syncColor != nil {
		clickedColor := currentHues[slot]
		// Only sync if the clicked button is a different color than the sync color
		if clickedColor != *syncColor {
			// Check if there's exactly one button with the clicked color (excluding the clicked slot)
			clickedColorCount := 0
			for i := 0; i < 6; i++ {
				if i != slot && currentHues[i] == clickedColor {
					clickedColorCount++
				}
			}
			// If this is the only button of its color, allow normal progression instead of syncing
			if clickedColorCount == 0 {
				// Fall through to normal button press logic
			} else {
				// Multiple buttons of this color, so sync applies
				newHue := *syncColor
				puzzle.SetButtonCurrentHue(slot, newHue)

				// Check for success condition after sync
				if c.checkSuccessCondition(puzzle) {
					c.setAllLightsToGreen(puzzle)
				}

				// Save and return
				if err := c.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
					return nil, err
				}
				return puzzle, nil
			}
		}
		// If clicked button is already the same color, fall through to normal button press logic
	}

	// Get the color of the clicked button
	clickedColor := currentHues[slot]

	// Handle special cases based on clicked button color
	c.handleButtonPress(puzzle, slot, clickedColor, currentHues, allGreensAchieved, allPurplesAchieved)

	// Check for success condition
	if c.checkSuccessCondition(puzzle) {
		c.setAllLightsToGreen(puzzle)
	}

	// Save updated puzzle state
	if err := c.dbClient.UtilityClosetPuzzle().UpdatePuzzle(ctx, puzzle); err != nil {
		return nil, err
	}

	return puzzle, nil
}

// ResetPuzzle resets the puzzle to its base state
func (c *utilityClosetPuzzleClient) ResetPuzzle(ctx context.Context) (*models.UtilityClosetPuzzle, error) {
	puzzle, err := c.dbClient.UtilityClosetPuzzle().ResetPuzzle(ctx)
	if err != nil {
		return nil, err
	}
	return puzzle, nil
}

// updateAchievementStates updates the persistent achievement states based on current puzzle state
func (c *utilityClosetPuzzleClient) updateAchievementStates(puzzle *models.UtilityClosetPuzzle, currentHues []int) (bool, bool) {
	// Check if all lights have been red at least once
	allGreensAchieved := puzzle.AllGreensAchieved
	allCurrentlyGreen := true
	for i := 0; i < 6; i++ {
		if currentHues[i] != 2 {
			allCurrentlyGreen = false
		}
	}
	// If all are currently red, this also unlocks blue
	if allCurrentlyGreen {
		allGreensAchieved = true
	}
	// Update the persistent field if it changed
	if puzzle.AllGreensAchieved != allGreensAchieved {
		puzzle.AllGreensAchieved = allGreensAchieved
	}

	// Check if all lights have been purple at least once
	allPurplesAchieved := puzzle.AllPurplesAchieved
	allCurrentlyPurple := true
	for i := 0; i < 6; i++ {
		if currentHues[i] != 5 {
			allCurrentlyPurple = false
		}
	}
	// If all are currently purple, this also unlocks white
	if allCurrentlyPurple {
		allPurplesAchieved = true
	}
	// Update the persistent field if it changed
	if puzzle.AllPurplesAchieved != allPurplesAchieved {
		puzzle.AllPurplesAchieved = allPurplesAchieved
	}

	return allGreensAchieved, allPurplesAchieved
}

// isPuzzleSolved checks if the puzzle is already in green (success) state
func (c *utilityClosetPuzzleClient) isPuzzleSolved(currentHues []int) bool {
	for i := 0; i < 6; i++ {
		if currentHues[i] == 6 {
			return true
		}
	}
	return false
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

// handleButtonPress handles the button press logic based on the clicked button's color
func (c *utilityClosetPuzzleClient) handleButtonPress(
	puzzle *models.UtilityClosetPuzzle,
	slot int,
	clickedColor int,
	currentHues []int,
	allGreensAchieved bool,
	allPurplesAchieved bool,
) {
	switch clickedColor {
	case 0: // Off (Gray) -> Red
		puzzle.SetButtonCurrentHue(slot, 2) // Red

	case 1: // Blue -> Off (and potentially create purple if next to yellow)
		// Check neighbors for yellow
		prevSlot := (slot - 1 + 6) % 6
		nextSlot := (slot + 1) % 6

		if currentHues[prevSlot] == 2 { // Previous is red
			// Turn yellow into purple
			puzzle.SetButtonCurrentHue(prevSlot, 5) // Purple
		}
		if currentHues[nextSlot] == 2 { // Next is red
			// Turn yellow into purple
			puzzle.SetButtonCurrentHue(nextSlot, 5) // Purple
		}

		// Turn blue back to off
		puzzle.SetButtonCurrentHue(slot, 0) // Off

	case 2: // Red -> Yellow
		// If the other 5 lights are purple, turn this light purple
		purpleCount := 0
		for i := 0; i < 6; i++ {
			if i != slot && currentHues[i] == 5 {
				purpleCount++
			}
		}
		if purpleCount == 5 {
			puzzle.SetButtonCurrentHue(slot, 5) // Purple
			break
		}

		// Normal progression: Red -> Yellow
		puzzle.SetButtonCurrentHue(slot, 4) // Yellow

	case 3: // White -> Off
		// White cycles back to off
		puzzle.SetButtonCurrentHue(slot, 0) // Off

	case 4: // Yellow -> Blue (only if all reds achieved) or cycle back to Off
		if allGreensAchieved {
			puzzle.SetButtonCurrentHue(slot, 1) // Blue
		} else {
			// Can't become blue yet, cycle to off
			puzzle.SetButtonCurrentHue(slot, 0) // Off
		}

	case 5: // Purple
		// Once allPurplesAchieved is true, any purple pressed turns white
		if allPurplesAchieved {
			// White is unlocked - clicking purple turns it white
			puzzle.SetButtonCurrentHue(slot, 3) // White
		} else {
			// Purple cycles back to off
			puzzle.SetButtonCurrentHue(slot, 0) // Off
		}
	}
}

// checkSuccessCondition checks if the success condition is met
// Success condition: Gray (0) → Blue (1) → Red (2) → White (3) → Yellow (4) → Purple (5)
func (c *utilityClosetPuzzleClient) checkSuccessCondition(puzzle *models.UtilityClosetPuzzle) bool {
	successOrder := []int{0, 1, 2, 3, 4, 5} // Off, Blue, Red, White, Yellow, Purple
	for i := 0; i < 6; i++ {
		if puzzle.GetButtonCurrentHue(i) != successOrder[i] {
			return false
		}
	}
	return true
}

// setAllLightsToGreen sets all lights to green (success state)
func (c *utilityClosetPuzzleClient) setAllLightsToGreen(puzzle *models.UtilityClosetPuzzle) {
	for i := 0; i < 6; i++ {
		puzzle.SetButtonCurrentHue(i, 6) // Green
	}
}
