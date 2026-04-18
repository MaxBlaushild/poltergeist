package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type zoneGenreHandler struct {
	db *gorm.DB
}

func (h *zoneGenreHandler) Create(ctx context.Context, genre *models.ZoneGenre) error {
	if genre == nil {
		return fmt.Errorf("zone genre is required")
	}
	genre.Name = strings.TrimSpace(genre.Name)
	if genre.Name == "" {
		return fmt.Errorf("zone genre name is required")
	}
	if genre.ID == uuid.Nil {
		genre.ID = uuid.New()
	}
	now := time.Now()
	if genre.CreatedAt.IsZero() {
		genre.CreatedAt = now
	}
	genre.UpdatedAt = now
	genre.PromptSeed = strings.TrimSpace(genre.PromptSeed)
	if genre.PromptSeed == "" && models.IsFantasyZoneGenreName(genre.Name) {
		genre.PromptSeed = models.DefaultFantasyZoneGenrePromptSeed()
	}
	return h.db.WithContext(ctx).Create(genre).Error
}

func (h *zoneGenreHandler) FindAll(ctx context.Context, includeInactive bool) ([]models.ZoneGenre, error) {
	var genres []models.ZoneGenre
	query := h.db.WithContext(ctx).Order("sort_order ASC").Order("name ASC")
	if !includeInactive {
		query = query.Where("active = ?", true)
	}
	if err := query.Find(&genres).Error; err != nil {
		return nil, err
	}
	return genres, nil
}

func (h *zoneGenreHandler) FindActive(ctx context.Context) ([]models.ZoneGenre, error) {
	return h.FindAll(ctx, false)
}

func (h *zoneGenreHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.ZoneGenre, error) {
	var genre models.ZoneGenre
	if err := h.db.WithContext(ctx).First(&genre, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &genre, nil
}

func (h *zoneGenreHandler) FindByName(ctx context.Context, name string) (*models.ZoneGenre, error) {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return nil, fmt.Errorf("zone genre name is required")
	}
	var genre models.ZoneGenre
	if err := h.db.WithContext(ctx).
		Where("LOWER(name) = LOWER(?)", normalizedName).
		Order("sort_order ASC").
		Order("created_at ASC").
		First(&genre).Error; err != nil {
		return nil, err
	}
	return &genre, nil
}

func (h *zoneGenreHandler) Update(ctx context.Context, genre *models.ZoneGenre) error {
	if genre == nil || genre.ID == uuid.Nil {
		return fmt.Errorf("zone genre is required")
	}
	genre.Name = strings.TrimSpace(genre.Name)
	if genre.Name == "" {
		return fmt.Errorf("zone genre name is required")
	}
	genre.PromptSeed = strings.TrimSpace(genre.PromptSeed)
	if genre.PromptSeed == "" && models.IsFantasyZoneGenreName(genre.Name) {
		genre.PromptSeed = models.DefaultFantasyZoneGenrePromptSeed()
	}
	genre.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Model(&models.ZoneGenre{}).Where("id = ?", genre.ID).Updates(map[string]interface{}{
		"name":        genre.Name,
		"sort_order":  genre.SortOrder,
		"active":      genre.Active,
		"prompt_seed": genre.PromptSeed,
		"updated_at":  genre.UpdatedAt,
	}).Error
}

func (h *zoneGenreHandler) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("zone genre ID is required")
	}
	result := h.db.WithContext(ctx).Delete(&models.ZoneGenre{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type zoneGenreScoreHandler struct {
	db *gorm.DB
}

func (h *zoneGenreScoreHandler) FindByZoneIDs(ctx context.Context, zoneIDs []uuid.UUID, includeInactiveGenres bool) ([]models.ZoneGenreScore, error) {
	if len(zoneIDs) == 0 {
		return []models.ZoneGenreScore{}, nil
	}
	query := h.db.WithContext(ctx).
		Preload("Genre").
		Where("zone_id IN ?", zoneIDs).
		Order("score DESC").
		Order("updated_at DESC")
	if !includeInactiveGenres {
		query = query.Joins("JOIN zone_genres ON zone_genres.id = zone_genre_scores.genre_id").Where("zone_genres.active = ?", true)
	}
	var scores []models.ZoneGenreScore
	if err := query.Find(&scores).Error; err != nil {
		return nil, err
	}
	return scores, nil
}

func (h *zoneGenreScoreHandler) Increment(ctx context.Context, zoneID uuid.UUID, genreID uuid.UUID, delta int) (*models.ZoneGenreScore, error) {
	if zoneID == uuid.Nil || genreID == uuid.Nil || delta == 0 {
		return nil, fmt.Errorf("invalid zone genre score increment")
	}

	now := time.Now()
	score := &models.ZoneGenreScore{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		ZoneID:    zoneID,
		GenreID:   genreID,
		Score:     delta,
	}
	if err := h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "zone_id"}, {Name: "genre_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"score":      gorm.Expr("zone_genre_scores.score + EXCLUDED.score"),
			"updated_at": now,
		}),
	}).Create(score).Error; err != nil {
		return nil, err
	}

	var updated models.ZoneGenreScore
	if err := h.db.WithContext(ctx).
		Preload("Genre").
		Where("zone_id = ? AND genre_id = ?", zoneID, genreID).
		First(&updated).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

func (h *zoneGenreScoreHandler) ConsumeUserItemAndIncrement(ctx context.Context, userID uuid.UUID, inventoryItemID int, zoneID uuid.UUID, genreID uuid.UUID, delta int) (*models.ZoneGenreScore, error) {
	if userID == uuid.Nil || inventoryItemID <= 0 || zoneID == uuid.Nil || genreID == uuid.Nil || delta == 0 {
		return nil, fmt.Errorf("invalid chaos engine increment")
	}

	var updated models.ZoneGenreScore
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var owned models.OwnedInventoryItem
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND inventory_item_id = ?", userID, inventoryItemID).
			First(&owned).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("insufficient quantity")
			}
			return err
		}

		if owned.Quantity < 1 {
			return errors.New("insufficient quantity")
		}

		owned.Quantity -= 1
		if owned.Quantity <= 0 {
			if err := tx.Delete(&owned).Error; err != nil {
				return err
			}
			if err := tx.Where("owned_inventory_item_id = ?", owned.ID).Delete(&models.UserEquipment{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&owned).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		score := &models.ZoneGenreScore{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			ZoneID:    zoneID,
			GenreID:   genreID,
			Score:     delta,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "zone_id"}, {Name: "genre_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"score":      gorm.Expr("zone_genre_scores.score + EXCLUDED.score"),
				"updated_at": now,
			}),
		}).Create(score).Error; err != nil {
			return err
		}

		return tx.
			Preload("Genre").
			Where("zone_id = ? AND genre_id = ?", zoneID, genreID).
			First(&updated).Error
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
