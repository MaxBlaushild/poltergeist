package db

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterEncounterHandle struct {
	db *gorm.DB
}

type monsterEncounterAdminListRow struct {
	ID   uuid.UUID `gorm:"column:id"`
	Name string    `gorm:"column:name"`
}

func (h *monsterEncounterHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("PointOfInterest").
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("slot ASC").Order("created_at ASC")
		}).
		Preload("Members.Monster").
		Preload("Members.Monster.Genre").
		Preload("Members.Monster.Zone").
		Preload("Members.Monster.Template").
		Preload("Members.Monster.Template.Genre").
		Preload("Members.Monster.Template.Spells").
		Preload("Members.Monster.Template.Spells.Spell").
		Preload("Members.Monster.Template.Progressions").
		Preload("Members.Monster.Template.Progressions.Progression").
		Preload("Members.Monster.Template.Progressions.Progression.Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("level_band ASC")
		}).
		Preload("Members.Monster.Template.Progressions.Progression.Members.Spell").
		Preload("Members.Monster.DominantHandInventoryItem").
		Preload("Members.Monster.OffHandInventoryItem").
		Preload("Members.Monster.WeaponInventoryItem").
		Preload("Members.Monster.ItemRewards").
		Preload("Members.Monster.ItemRewards.InventoryItem")
}

func (h *monsterEncounterHandle) visibleQuery(ctx context.Context) *gorm.DB {
	return h.preloadBase(ctx).Where("retired_at IS NULL")
}

func (h *monsterEncounterHandle) Create(ctx context.Context, encounter *models.MonsterEncounter) error {
	now := time.Now()
	if encounter.ID == uuid.Nil {
		encounter.ID = uuid.New()
	}
	if encounter.CreatedAt.IsZero() {
		encounter.CreatedAt = now
	}
	encounter.UpdatedAt = now
	encounter.RequiredStoryFlags = normalizeJSONStringArray(encounter.RequiredStoryFlags)
	if err := encounter.SetGeometry(encounter.Latitude, encounter.Longitude); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(encounter).Error
}

func (h *monsterEncounterHandle) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.MonsterEncounter, error) {
	var encounter models.MonsterEncounter
	if err := h.preloadBase(ctx).Where("id = ?", id).First(&encounter).Error; err != nil {
		return nil, err
	}
	return &encounter, nil
}

func (h *monsterEncounterHandle) FindAll(ctx context.Context) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	if err := h.visibleQuery(ctx).Order("name ASC").Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) adminListBaseQuery(
	ctx context.Context,
	params MonsterEncounterAdminListParams,
) *gorm.DB {
	query := h.db.WithContext(ctx).
		Model(&models.MonsterEncounter{}).
		Where("monster_encounters.retired_at IS NULL").
		Joins("LEFT JOIN zones ON zones.id = monster_encounters.zone_id").
		Joins("LEFT JOIN monster_encounter_members ON monster_encounter_members.monster_encounter_id = monster_encounters.id").
		Joins("LEFT JOIN monsters member_monsters ON member_monsters.id = monster_encounter_members.monster_id")

	if normalizedQuery := strings.TrimSpace(strings.ToLower(params.Query)); normalizedQuery != "" {
		searchTerm := "%" + normalizedQuery + "%"
		query = query.Where(
			`(
				LOWER(monster_encounters.name) LIKE ?
				OR LOWER(monster_encounters.description) LIKE ?
				OR LOWER(COALESCE(zones.name, '')) LIKE ?
				OR LOWER(COALESCE(member_monsters.name, '')) LIKE ?
			)`,
			searchTerm,
			searchTerm,
			searchTerm,
			searchTerm,
		)
	}

	if normalizedZoneQuery := strings.TrimSpace(strings.ToLower(params.ZoneQuery)); normalizedZoneQuery != "" {
		zoneSearchTerm := "%" + normalizedZoneQuery + "%"
		query = query.Where("LOWER(COALESCE(zones.name, '')) LIKE ?", zoneSearchTerm)
	}

	if params.GenreID != nil && *params.GenreID != uuid.Nil {
		query = query.Where("member_monsters.genre_id = ?", *params.GenreID)
	}

	return query
}

func (h *monsterEncounterHandle) ListAdmin(
	ctx context.Context,
	params MonsterEncounterAdminListParams,
) (*MonsterEncounterAdminListResult, error) {
	var total int64
	if err := h.adminListBaseQuery(ctx, params).
		Distinct("monster_encounters.id").
		Count(&total).Error; err != nil {
		return nil, err
	}

	rows := []monsterEncounterAdminListRow{}
	if err := h.adminListBaseQuery(ctx, params).
		Select("monster_encounters.id, monster_encounters.name").
		Distinct().
		Order("monster_encounters.name ASC").
		Limit(params.PageSize).
		Offset((params.Page - 1) * params.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	encounters := make([]models.MonsterEncounter, 0, len(ids))
	if len(ids) > 0 {
		loaded := []models.MonsterEncounter{}
		if err := h.preloadBase(ctx).
			Where("monster_encounters.id IN ?", ids).
			Where("monster_encounters.retired_at IS NULL").
			Find(&loaded).Error; err != nil {
			return nil, err
		}
		encountersByID := make(map[uuid.UUID]models.MonsterEncounter, len(loaded))
		for _, encounter := range loaded {
			encountersByID[encounter.ID] = encounter
		}
		for _, id := range ids {
			encounter, ok := encountersByID[id]
			if ok {
				encounters = append(encounters, encounter)
			}
		}
	}

	return &MonsterEncounterAdminListResult{
		Encounters: encounters,
		Total:      total,
	}, nil
}

func (h *monsterEncounterHandle) FindByZoneID(
	ctx context.Context,
	zoneID uuid.UUID,
) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Order("name ASC").
		Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) FindByZoneIDExcludingQuestNodes(
	ctx context.Context,
	zoneID uuid.UUID,
) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.monster_encounter_id = monster_encounters.id)").
		Order("name ASC").
		Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) FindDueRecurring(
	ctx context.Context,
	asOf time.Time,
	limit int,
) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	query := h.db.WithContext(ctx).
		Where("retired_at IS NULL").
		Where("recurrence_frequency IS NOT NULL AND recurrence_frequency <> ''").
		Where("next_recurrence_at IS NOT NULL AND next_recurrence_at <= ?", asOf).
		Order("next_recurrence_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) FindFirstByMonsterID(
	ctx context.Context,
	monsterID uuid.UUID,
) (*models.MonsterEncounter, error) {
	var encounter models.MonsterEncounter
	err := h.preloadBase(ctx).
		Joins("JOIN monster_encounter_members mem ON mem.monster_encounter_id = monster_encounters.id").
		Where("mem.monster_id = ?", monsterID).
		Order("monster_encounters.created_at ASC").
		First(&encounter).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &encounter, nil
}

func (h *monsterEncounterHandle) Update(
	ctx context.Context,
	id uuid.UUID,
	updates *models.MonsterEncounter,
) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	updates.RequiredStoryFlags = normalizeJSONStringArray(updates.RequiredStoryFlags)
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"name":                           updates.Name,
		"description":                    updates.Description,
		"image_url":                      updates.ImageURL,
		"thumbnail_url":                  updates.ThumbnailURL,
		"encounter_type":                 updates.EncounterType,
		"reward_mode":                    updates.RewardMode,
		"random_reward_size":             updates.RandomRewardSize,
		"reward_experience":              updates.RewardExperience,
		"reward_gold":                    updates.RewardGold,
		"item_rewards_json":              updates.ItemRewards,
		"scale_with_user_level":          updates.ScaleWithUserLevel,
		"recurring_monster_encounter_id": updates.RecurringMonsterEncounterID,
		"recurrence_frequency":           updates.RecurrenceFrequency,
		"next_recurrence_at":             updates.NextRecurrenceAt,
		"retired_at":                     updates.RetiredAt,
		"zone_id":                        updates.ZoneID,
		"required_story_flags":           updates.RequiredStoryFlags,
		"point_of_interest_id":           updates.PointOfInterestID,
		"latitude":                       updates.Latitude,
		"longitude":                      updates.Longitude,
		"geometry":                       updates.Geometry,
		"updated_at":                     updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).
		Model(&models.MonsterEncounter{}).
		Where("id = ?", id).
		Updates(payload).Error
}

func (h *monsterEncounterHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.MonsterEncounter{}, "id = ?", id).Error
}

func (h *monsterEncounterHandle) ReplaceMembers(
	ctx context.Context,
	encounterID uuid.UUID,
	members []models.MonsterEncounterMember,
) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monster_encounter_id = ?", encounterID).
			Delete(&models.MonsterEncounterMember{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, member := range members {
			member.ID = uuid.New()
			member.CreatedAt = now
			member.UpdatedAt = now
			member.MonsterEncounterID = encounterID
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
