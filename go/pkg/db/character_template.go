package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type characterTemplateHandle struct {
	db *gorm.DB
}

func (h *characterTemplateHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).Preload("Genre")
}

func (h *characterTemplateHandle) Create(ctx context.Context, template *models.CharacterTemplate) error {
	if template == nil {
		return nil
	}
	resolvedGenreID, err := resolveCharacterTemplateGenreID(ctx, h.db, template)
	if err != nil {
		return err
	}
	template.GenreID = resolvedGenreID
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *characterTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.CharacterTemplate, error) {
	var template models.CharacterTemplate
	if err := h.preloadBase(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *characterTemplateHandle) FindAll(ctx context.Context) ([]models.CharacterTemplate, error) {
	var templates []models.CharacterTemplate
	if err := h.preloadBase(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *characterTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.CharacterTemplate) error {
	if updates == nil {
		return nil
	}
	resolvedGenreID, err := resolveCharacterTemplateGenreID(ctx, h.db, updates)
	if err != nil {
		return err
	}
	updates.GenreID = resolvedGenreID
	updates.UpdatedAt = time.Now()
	payload := map[string]interface{}{
		"name":                    updates.Name,
		"description":             updates.Description,
		"internal_tags":           updates.InternalTags,
		"map_icon_url":            updates.MapIconURL,
		"dialogue_image_url":      updates.DialogueImageURL,
		"thumbnail_url":           updates.ThumbnailURL,
		"genre_id":                updates.GenreID,
		"story_variants":          updates.StoryVariants,
		"image_generation_status": updates.ImageGenerationStatus,
		"image_generation_error":  updates.ImageGenerationError,
		"updated_at":              updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.CharacterTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *characterTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.CharacterTemplate{}, "id = ?", id).Error
}
