package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNoStatAllocations     = errors.New("no stat allocations")
	ErrInvalidStatAllocation = errors.New("invalid stat allocation")
	ErrInsufficientStatPoints = errors.New("insufficient stat points")
)

type userCharacterStatsHandler struct {
	db *gorm.DB
}

func (h *userCharacterStatsHandler) Create(ctx context.Context, stats *models.UserCharacterStats) error {
	return h.db.WithContext(ctx).Create(stats).Error
}

func (h *userCharacterStatsHandler) FindOrCreateForUser(ctx context.Context, userID uuid.UUID) (*models.UserCharacterStats, error) {
	stats, err := h.FindByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if stats != nil {
		return stats, nil
	}

	stats = defaultCharacterStats(userID)
	return stats, h.Create(ctx, stats)
}

func (h *userCharacterStatsHandler) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserCharacterStats, error) {
	var stats models.UserCharacterStats
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

func (h *userCharacterStatsHandler) EnsureLevelPoints(ctx context.Context, userID uuid.UUID, currentLevel int) (*models.UserCharacterStats, error) {
	return h.applyUpdate(ctx, userID, currentLevel, nil)
}

func (h *userCharacterStatsHandler) ApplyAllocations(ctx context.Context, userID uuid.UUID, currentLevel int, allocations map[string]int) (*models.UserCharacterStats, error) {
	if len(allocations) == 0 {
		return nil, ErrNoStatAllocations
	}
	return h.applyUpdate(ctx, userID, currentLevel, allocations)
}

func (h *userCharacterStatsHandler) AddStatPoints(ctx context.Context, userID uuid.UUID, additions map[string]int) (*models.UserCharacterStats, error) {
	if len(additions) == 0 {
		return h.FindOrCreateForUser(ctx, userID)
	}
	var result *models.UserCharacterStats
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stats, err := h.findOrCreateForUserTx(tx, userID)
		if err != nil {
			return err
		}
		if err := applyStatAdditions(stats, additions); err != nil {
			return err
		}
		stats.UpdatedAt = time.Now()
		if err := tx.Save(stats).Error; err != nil {
			return err
		}
		result = stats
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *userCharacterStatsHandler) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserCharacterStats{}).Error
}

func (h *userCharacterStatsHandler) applyUpdate(ctx context.Context, userID uuid.UUID, currentLevel int, allocations map[string]int) (*models.UserCharacterStats, error) {
	var result *models.UserCharacterStats
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stats, err := h.findOrCreateForUserTx(tx, userID)
		if err != nil {
			return err
		}

		applyLevelAwards(stats, currentLevel)

		if allocations != nil {
			total, err := applyAllocations(stats, allocations)
			if err != nil {
				return err
			}
			if total > stats.UnspentPoints {
				return ErrInsufficientStatPoints
			}
			stats.UnspentPoints -= total
		}

		stats.UpdatedAt = time.Now()
		if err := tx.Save(stats).Error; err != nil {
			return err
		}
		result = stats
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *userCharacterStatsHandler) findOrCreateForUserTx(tx *gorm.DB, userID uuid.UUID) (*models.UserCharacterStats, error) {
	var stats models.UserCharacterStats
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&stats).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		created := defaultCharacterStats(userID)
		if err := tx.Create(created).Error; err != nil {
			return nil, err
		}
		return created, nil
	}
	return &stats, nil
}

func defaultCharacterStats(userID uuid.UUID) *models.UserCharacterStats {
	now := time.Now()
	return &models.UserCharacterStats{
		ID:               uuid.New(),
		UserID:           userID,
		CreatedAt:        now,
		UpdatedAt:        now,
		Strength:         models.CharacterStatBaseValue,
		Dexterity:        models.CharacterStatBaseValue,
		Constitution:     models.CharacterStatBaseValue,
		Intelligence:     models.CharacterStatBaseValue,
		Wisdom:           models.CharacterStatBaseValue,
		Charisma:         models.CharacterStatBaseValue,
		UnspentPoints:    0,
		LastLevelAwarded: 1,
	}
}

func applyLevelAwards(stats *models.UserCharacterStats, currentLevel int) {
	if currentLevel <= 0 {
		return
	}
	if stats.LastLevelAwarded < 1 {
		stats.LastLevelAwarded = 1
	}
	if currentLevel <= stats.LastLevelAwarded {
		return
	}
	gainedLevels := currentLevel - stats.LastLevelAwarded
	stats.UnspentPoints += gainedLevels * models.CharacterStatPointsPerLevel
	stats.LastLevelAwarded = currentLevel
}

func applyAllocations(stats *models.UserCharacterStats, allocations map[string]int) (int, error) {
	total := 0
	for key, value := range allocations {
		if value < 0 {
			return 0, ErrInvalidStatAllocation
		}
		if value == 0 {
			continue
		}
		switch key {
		case "strength":
			stats.Strength += value
		case "dexterity":
			stats.Dexterity += value
		case "constitution":
			stats.Constitution += value
		case "intelligence":
			stats.Intelligence += value
		case "wisdom":
			stats.Wisdom += value
		case "charisma":
			stats.Charisma += value
		default:
			return 0, ErrInvalidStatAllocation
		}
		total += value
	}
	if total == 0 {
		return 0, ErrNoStatAllocations
	}
	return total, nil
}

func applyStatAdditions(stats *models.UserCharacterStats, additions map[string]int) error {
	for key, value := range additions {
		if value < 0 {
			return ErrInvalidStatAllocation
		}
		if value == 0 {
			continue
		}
		switch key {
		case "strength":
			stats.Strength += value
		case "dexterity":
			stats.Dexterity += value
		case "constitution":
			stats.Constitution += value
		case "intelligence":
			stats.Intelligence += value
		case "wisdom":
			stats.Wisdom += value
		case "charisma":
			stats.Charisma += value
		default:
			return ErrInvalidStatAllocation
		}
	}
	return nil
}
