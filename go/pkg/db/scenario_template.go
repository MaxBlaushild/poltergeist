package db

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scenarioTemplateHandle struct {
	db *gorm.DB
}

type scenarioTemplateAdminListRow struct {
	ID        uuid.UUID `gorm:"column:id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (h *scenarioTemplateHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Genre")
}

func (h *scenarioTemplateHandle) Create(ctx context.Context, template *models.ScenarioTemplate) error {
	if template == nil {
		return nil
	}
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt
	resolvedGenreID, err := resolveScenarioTemplateGenreID(ctx, h.db, template)
	if err != nil {
		return err
	}
	template.GenreID = resolvedGenreID
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *scenarioTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ScenarioTemplate, error) {
	var template models.ScenarioTemplate
	if err := h.preloadBase(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *scenarioTemplateHandle) FindAll(ctx context.Context) ([]models.ScenarioTemplate, error) {
	var templates []models.ScenarioTemplate
	if err := h.preloadBase(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *scenarioTemplateHandle) FindRecent(ctx context.Context, limit int) ([]models.ScenarioTemplate, error) {
	return h.findRecentWithGenre(ctx, nil, limit)
}

func (h *scenarioTemplateHandle) FindRecentByGenre(ctx context.Context, genreID uuid.UUID, limit int) ([]models.ScenarioTemplate, error) {
	return h.findRecentWithGenre(ctx, &genreID, limit)
}

func (h *scenarioTemplateHandle) findRecentWithGenre(ctx context.Context, genreID *uuid.UUID, limit int) ([]models.ScenarioTemplate, error) {
	var templates []models.ScenarioTemplate
	q := h.preloadBase(ctx).Order("created_at DESC")
	if genreID != nil && *genreID != uuid.Nil {
		q = q.Where("genre_id = ?", *genreID)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *scenarioTemplateHandle) adminListBaseQuery(
	ctx context.Context,
	params ScenarioTemplateAdminListParams,
) *gorm.DB {
	query := h.db.WithContext(ctx).Model(&models.ScenarioTemplate{})

	if normalizedQuery := strings.TrimSpace(strings.ToLower(params.Query)); normalizedQuery != "" {
		searchTerm := "%" + normalizedQuery + "%"
		query = query.Where(
			`(
				LOWER(scenario_templates.prompt) LIKE ?
				OR LOWER(CAST(scenario_templates.id AS text)) LIKE ?
			)`,
			searchTerm,
			searchTerm,
		)
	}

	if params.GenreID != nil && *params.GenreID != uuid.Nil {
		query = query.Where("scenario_templates.genre_id = ?", *params.GenreID)
	}

	return query
}

func (h *scenarioTemplateHandle) ListAdmin(
	ctx context.Context,
	params ScenarioTemplateAdminListParams,
) (*ScenarioTemplateAdminListResult, error) {
	var total int64
	if err := h.adminListBaseQuery(ctx, params).Count(&total).Error; err != nil {
		return nil, err
	}

	rows := []scenarioTemplateAdminListRow{}
	if err := h.adminListBaseQuery(ctx, params).
		Select("scenario_templates.id, scenario_templates.created_at").
		Order("scenario_templates.created_at DESC").
		Limit(params.PageSize).
		Offset((params.Page - 1) * params.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	templates := make([]models.ScenarioTemplate, 0, len(ids))
	if len(ids) > 0 {
		loaded := []models.ScenarioTemplate{}
		if err := h.preloadBase(ctx).
			Where("id IN ?", ids).
			Find(&loaded).Error; err != nil {
			return nil, err
		}
		byID := make(map[uuid.UUID]models.ScenarioTemplate, len(loaded))
		for _, template := range loaded {
			byID[template.ID] = template
		}
		for _, id := range ids {
			template, ok := byID[id]
			if ok {
				templates = append(templates, template)
			}
		}
	}

	return &ScenarioTemplateAdminListResult{
		Templates: templates,
		Total:     total,
	}, nil
}

func (h *scenarioTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.ScenarioTemplate) error {
	if updates == nil {
		return nil
	}
	updates.UpdatedAt = time.Now()
	resolvedGenreID, err := resolveScenarioTemplateGenreIDForUpdate(ctx, h.db, id, updates)
	if err != nil {
		return err
	}
	updates.GenreID = resolvedGenreID
	payload := map[string]interface{}{
		"genre_id":                     updates.GenreID,
		"prompt":                       updates.Prompt,
		"image_url":                    updates.ImageURL,
		"thumbnail_url":                updates.ThumbnailURL,
		"scale_with_user_level":        updates.ScaleWithUserLevel,
		"reward_mode":                  updates.RewardMode,
		"random_reward_size":           updates.RandomRewardSize,
		"difficulty":                   updates.Difficulty,
		"reward_experience":            updates.RewardExperience,
		"reward_gold":                  updates.RewardGold,
		"open_ended":                   updates.OpenEnded,
		"success_handoff_text":         updates.SuccessHandoffText,
		"failure_handoff_text":         updates.FailureHandoffText,
		"failure_penalty_mode":         updates.FailurePenaltyMode,
		"failure_health_drain_type":    updates.FailureHealthDrainType,
		"failure_health_drain_value":   updates.FailureHealthDrainValue,
		"failure_mana_drain_type":      updates.FailureManaDrainType,
		"failure_mana_drain_value":     updates.FailureManaDrainValue,
		"failure_statuses":             updates.FailureStatuses,
		"success_reward_mode":          updates.SuccessRewardMode,
		"success_health_restore_type":  updates.SuccessHealthRestoreType,
		"success_health_restore_value": updates.SuccessHealthRestoreValue,
		"success_mana_restore_type":    updates.SuccessManaRestoreType,
		"success_mana_restore_value":   updates.SuccessManaRestoreValue,
		"success_statuses":             updates.SuccessStatuses,
		"options":                      updates.Options,
		"item_rewards":                 updates.ItemRewards,
		"item_choice_rewards":          updates.ItemChoiceRewards,
		"spell_rewards":                updates.SpellRewards,
		"updated_at":                   updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.ScenarioTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *scenarioTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ScenarioTemplate{}, "id = ?", id).Error
}
