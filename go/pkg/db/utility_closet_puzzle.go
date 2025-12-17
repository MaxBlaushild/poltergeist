package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type utilityClosetPuzzleHandler struct {
	db *gorm.DB
}

func (h *utilityClosetPuzzleHandler) GetPuzzle(ctx context.Context) (*models.UtilityClosetPuzzle, error) {
	var puzzle models.UtilityClosetPuzzle

	// Try to find existing puzzle
	err := h.db.WithContext(ctx).First(&puzzle).Error
	if err == gorm.ErrRecordNotFound {
		// Create new puzzle with default state if it doesn't exist
		puzzle = models.UtilityClosetPuzzle{
			ID:                uuid.New(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			Button0CurrentHue: 0,
			Button1CurrentHue: 1,
			Button2CurrentHue: 2,
			Button3CurrentHue: 3,
			Button4CurrentHue: 4,
			Button5CurrentHue: 5,
			Button0BaseHue:    0,
			Button1BaseHue:    1,
			Button2BaseHue:    2,
			Button3BaseHue:    3,
			Button4BaseHue:    4,
			Button5BaseHue:    5,
			AllGreensAchieved: false,
		}
		if err := h.db.WithContext(ctx).Create(&puzzle).Error; err != nil {
			return nil, err
		}
		return &puzzle, nil
	}
	if err != nil {
		return nil, err
	}

	return &puzzle, nil
}

func (h *utilityClosetPuzzleHandler) UpdatePuzzle(ctx context.Context, puzzle *models.UtilityClosetPuzzle) error {
	puzzle.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Save(puzzle).Error
}

func (h *utilityClosetPuzzleHandler) ResetPuzzle(ctx context.Context) (*models.UtilityClosetPuzzle, error) {
	puzzle, err := h.GetPuzzle(ctx)
	if err != nil {
		return nil, err
	}

	// Reset all buttons to base state
	puzzle.ResetToBaseState()

	// Update in database
	if err := h.UpdatePuzzle(ctx, puzzle); err != nil {
		return nil, err
	}

	return puzzle, nil
}

func (h *utilityClosetPuzzleHandler) DeletePuzzle(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.UtilityClosetPuzzle{}, id).Error
}
