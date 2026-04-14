package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestHandle struct {
	db *gorm.DB
}

func (c *pointOfInterestHandle) preloadBase(ctx context.Context) *gorm.DB {
	return c.db.WithContext(ctx).
		Preload("Characters").
		Preload("Tags").
		Preload("PointOfInterestChallenges").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell")
}

func normalizePointOfInterestRewards(pointOfInterest *models.PointOfInterest) {
	if pointOfInterest == nil {
		return
	}
	if pointOfInterest.MaterialRewards == nil {
		pointOfInterest.MaterialRewards = models.BaseMaterialRewards{}
	}
	if strings.TrimSpace(string(pointOfInterest.RewardMode)) == "" {
		if pointOfInterest.RewardExperience > 0 ||
			pointOfInterest.RewardGold > 0 ||
			len(pointOfInterest.MaterialRewards) > 0 ||
			len(pointOfInterest.ItemRewards) > 0 ||
			len(pointOfInterest.SpellRewards) > 0 {
			pointOfInterest.RewardMode = models.RewardModeExplicit
		} else {
			pointOfInterest.RewardMode = models.RewardModeRandom
		}
	}
	pointOfInterest.RewardMode = models.NormalizeRewardMode(string(pointOfInterest.RewardMode))
	pointOfInterest.RandomRewardSize = models.NormalizeRandomRewardSize(string(pointOfInterest.RandomRewardSize))
	if pointOfInterest.RewardExperience < 0 {
		pointOfInterest.RewardExperience = 0
	}
	if pointOfInterest.RewardGold < 0 {
		pointOfInterest.RewardGold = 0
	}
}

type CreatePointOfInterestRequest struct {
	Name                         string  `binding:"required" json:"name"`
	Description                  string  `binding:"required" json:"description"`
	ImageUrl                     string  `binding:"required" json:"imageUrl"`
	Lat                          string  `binding:"required" json:"lat"`
	Lon                          string  `binding:"required" json:"lon"`
	Clue                         string  `binding:"required" json:"clue"`
	TierOne                      string  `binding:"required" json:"tierOne"`
	TierTwo                      *string `json:"tierTwo"`
	TierThree                    *string `json:"tierThree"`
	TierOneInventoryItemId       int     `json:"tierOneInventoryItemId"`
	TierTwoInventoryItemId       int     `json:"tierTwoInventoryItemId"`
	TierThreeInventoryItemId     int     `json:"tierThreeInventoryItemId"`
	PointOfInterestGroupMemberID string  `json:"pointOfInterestGroupMemberId"`
}

func (c *pointOfInterestHandle) Delete(ctx context.Context, id uuid.UUID) error {
	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}

	runDelete := func(query string, args ...interface{}) error {
		if err := tx.Exec(query, args...).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	}

	// Child rows that depend on group members/challenges for this POI.
	if err := runDelete(
		`DELETE FROM point_of_interest_children
		WHERE point_of_interest_group_member_id IN (
			SELECT id FROM point_of_interest_group_members WHERE point_of_interest_id = ?
		) OR next_point_of_interest_group_member_id IN (
			SELECT id FROM point_of_interest_group_members WHERE point_of_interest_id = ?
		) OR point_of_interest_challenge_id IN (
			SELECT id FROM point_of_interest_challenges WHERE point_of_interest_id = ?
		)`,
		id, id, id,
	); err != nil {
		return err
	}

	if err := runDelete(
		`DELETE FROM point_of_interest_challenge_submissions
		WHERE point_of_interest_challenge_id IN (
			SELECT id FROM point_of_interest_challenges WHERE point_of_interest_id = ?
		)`,
		id,
	); err != nil {
		return err
	}

	// Direct POI references.
	if err := runDelete("DELETE FROM point_of_interest_activities WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM match_points_of_interest WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM neighboring_points_of_interest WHERE point_of_interest_one_id = ? OR point_of_interest_two_id = ?", id, id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM point_of_interest_teams WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM point_of_interest_challenges WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM point_of_interest_discoveries WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM point_of_interest_group_members WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM point_of_interest_zones WHERE point_of_interest_id = ?", id); err != nil {
		return err
	}
	if err := runDelete("DELETE FROM points_of_interest WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit().Error
}

func (c *pointOfInterestHandle) UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error {
	return c.db.Model(&models.PointOfInterest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_url":     imageUrl,
		"thumbnail_url": "",
		"updated_at":    time.Now(),
	}).Error
}

func (c *pointOfInterestHandle) UpdateImageGenerationStatus(ctx context.Context, id uuid.UUID, status string, errMsg *string) error {
	return c.db.Model(&models.PointOfInterest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_generation_status": status,
		"image_generation_error":  errMsg,
		"updated_at":              time.Now(),
	}).Error
}

func (c *pointOfInterestHandle) Edit(ctx context.Context, id uuid.UUID, name string, description string, lat string, lng string, unlockTier *int, clue string, imageUrl string, originalName string, googleMapsPlaceId *string) error {
	pointOfInterest := models.PointOfInterest{
		ID:                id,
		Name:              name,
		Description:       description,
		Lat:               lat,
		Lng:               lng,
		UnlockTier:        unlockTier,
		Clue:              clue,
		ImageUrl:          imageUrl,
		OriginalName:      originalName,
		GoogleMapsPlaceID: googleMapsPlaceId,
		UpdatedAt:         time.Now(),
	}
	if err := pointOfInterest.SetGeometry(lat, lng); err != nil {
		return err
	}

	return c.db.Model(&models.PointOfInterest{}).Where("id = ?", id).Updates(pointOfInterest).Error
}

func (c *pointOfInterestHandle) FindAll(ctx context.Context) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.preloadBase(ctx).
		Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}
	for i := range pointsOfInterest {
		normalizePointOfInterestRewards(&pointsOfInterest[i])
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterest, error) {
	var pointOfInterest models.PointOfInterest

	if err := c.preloadBase(ctx).
		First(&pointOfInterest, id).Error; err != nil {
		return nil, err
	}
	normalizePointOfInterestRewards(&pointOfInterest)

	return &pointOfInterest, nil
}

func (c *pointOfInterestHandle) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.PointOfInterest, error) {
	var pointsOfInterestGroupMembers []models.PointOfInterestGroupMember

	if err := c.db.WithContext(ctx).Where("point_of_interest_group_id = ?", groupID).Find(&pointsOfInterestGroupMembers).Error; err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	for _, member := range pointsOfInterestGroupMembers {
		ids = append(ids, member.PointOfInterestID)
	}

	pointsOfInterest, err := c.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.preloadBase(ctx).
		Where("id IN (?)", ids).
		Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}
	for i := range pointsOfInterest {
		normalizePointOfInterestRewards(&pointsOfInterest[i])
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.db.WithContext(ctx).Where("match_id = ?", matchID).Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) Unlock(ctx context.Context, pointOfInterestID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID) error {
	unlock := models.PointOfInterestDiscovery{
		TeamID:            teamID,
		UserID:            userID,
		PointOfInterestID: pointOfInterestID,
	}

	return c.db.WithContext(ctx).Create(&unlock).Error
}

func (c *pointOfInterestHandle) Create(ctx context.Context, pointOfInterest models.PointOfInterest) error {
	if err := pointOfInterest.SetGeometry(pointOfInterest.Lat, pointOfInterest.Lng); err != nil {
		return err
	}
	normalizePointOfInterestRewards(&pointOfInterest)

	return c.db.WithContext(ctx).Create(&pointOfInterest).Error
}

func (c *pointOfInterestHandle) CreateForGroup(ctx context.Context, pointOfInterest *models.PointOfInterest, pointOfInterestGroupID uuid.UUID) error {
	pointOfInterest.ID = uuid.New()
	pointOfInterest.CreatedAt = time.Now()
	pointOfInterest.UpdatedAt = time.Now()
	normalizePointOfInterestRewards(pointOfInterest)

	if err := c.db.WithContext(ctx).Create(&pointOfInterest).Error; err != nil {
		return err
	}

	pointOfInterestGroupMember := models.PointOfInterestGroupMember{
		ID:                     uuid.New(),
		PointOfInterestID:      pointOfInterest.ID,
		PointOfInterestGroupID: pointOfInterestGroupID,
	}

	if err := pointOfInterest.SetGeometry(pointOfInterest.Lat, pointOfInterest.Lng); err != nil {
		return err
	}

	if err := c.db.WithContext(ctx).Create(&pointOfInterestGroupMember).Error; err != nil {
		return err
	}

	return nil
}

func (c *pointOfInterestHandle) FindAllForZone(ctx context.Context, zoneID uuid.UUID) ([]models.PointOfInterest, error) {
	var pointOfInterestZones []models.PointOfInterestZone

	if err := c.db.WithContext(ctx).Where("zone_id = ?", zoneID).Find(&pointOfInterestZones).Error; err != nil {
		return nil, err
	}

	var pointOfInterestIDs []uuid.UUID
	for _, pointOfInterestZone := range pointOfInterestZones {
		pointOfInterestIDs = append(pointOfInterestIDs, pointOfInterestZone.PointOfInterestID)
	}

	pointsOfInterest, err := c.FindByIDs(ctx, pointOfInterestIDs)
	if err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByGoogleMapsPlaceID(ctx context.Context, googleMapsPlaceID string) (*models.PointOfInterest, error) {
	var pointOfInterest models.PointOfInterest
	if err := c.preloadBase(ctx).Where("google_maps_place_id = ?", googleMapsPlaceID).First(&pointOfInterest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	normalizePointOfInterestRewards(&pointOfInterest)

	return &pointOfInterest, nil
}

func (c *pointOfInterestHandle) Update(ctx context.Context, pointOfInterestID uuid.UUID, updates *models.PointOfInterest) error {
	if updates == nil {
		return nil
	}
	updates.ID = pointOfInterestID
	updates.UpdatedAt = time.Now()
	if err := updates.SetGeometry(updates.Lat, updates.Lng); err != nil {
		return err
	}
	normalizePointOfInterestRewards(updates)

	payload := map[string]interface{}{
		"name":                    updates.Name,
		"description":             updates.Description,
		"lat":                     updates.Lat,
		"lng":                     updates.Lng,
		"image_url":               updates.ImageUrl,
		"thumbnail_url":           updates.ThumbnailURL,
		"image_generation_status": updates.ImageGenerationStatus,
		"image_generation_error":  updates.ImageGenerationError,
		"clue":                    updates.Clue,
		"original_name":           updates.OriginalName,
		"google_maps_place_id":    updates.GoogleMapsPlaceID,
		"google_maps_place_name":  updates.GoogleMapsPlaceName,
		"story_variants":          updates.StoryVariants,
		"geometry":                updates.Geometry,
		"reward_mode":             updates.RewardMode,
		"random_reward_size":      updates.RandomRewardSize,
		"reward_experience":       updates.RewardExperience,
		"reward_gold":             updates.RewardGold,
		"material_rewards_json":   updates.MaterialRewards,
		"unlock_tier":             updates.UnlockTier,
		"updated_at":              updates.UpdatedAt,
	}

	return c.db.WithContext(ctx).Model(&models.PointOfInterest{}).Where("id = ?", pointOfInterestID).Updates(payload).Error
}

func (c *pointOfInterestHandle) ReplaceItemRewards(ctx context.Context, pointOfInterestID uuid.UUID, rewards []models.PointOfInterestItemReward) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("point_of_interest_id = ?", pointOfInterestID).Delete(&models.PointOfInterestItemReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.PointOfInterestID = pointOfInterestID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *pointOfInterestHandle) ReplaceSpellRewards(ctx context.Context, pointOfInterestID uuid.UUID, rewards []models.PointOfInterestSpellReward) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("point_of_interest_id = ?", pointOfInterestID).Delete(&models.PointOfInterestSpellReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.PointOfInterestID = pointOfInterestID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *pointOfInterestHandle) FindZoneForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID) (*models.PointOfInterestZone, error) {
	var pointOfInterestZone models.PointOfInterestZone
	if err := c.db.WithContext(ctx).Where("point_of_interest_id = ?", pointOfInterestID).First(&pointOfInterestZone).Error; err != nil {
		return nil, err
	}
	return &pointOfInterestZone, nil
}

func (c *pointOfInterestHandle) UpdateLastUsedInQuest(ctx context.Context, pointOfInterestID uuid.UUID) error {
	now := time.Now()
	return c.db.WithContext(ctx).Model(&models.PointOfInterest{}).Where("id = ?", pointOfInterestID).Update("last_used_in_quest_at", now).Error
}

func (c *pointOfInterestHandle) FindRecentlyUsedInZone(ctx context.Context, zoneID uuid.UUID, since time.Time) (map[string]bool, error) {
	var pointOfInterestZones []models.PointOfInterestZone
	if err := c.db.WithContext(ctx).Where("zone_id = ?", zoneID).Find(&pointOfInterestZones).Error; err != nil {
		return nil, err
	}

	var pointOfInterestIDs []uuid.UUID
	for _, pointOfInterestZone := range pointOfInterestZones {
		pointOfInterestIDs = append(pointOfInterestIDs, pointOfInterestZone.PointOfInterestID)
	}

	var pointsOfInterest []models.PointOfInterest
	if err := c.db.WithContext(ctx).
		Where("id IN (?) AND last_used_in_quest_at > ?", pointOfInterestIDs, since).
		Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	recentlyUsed := make(map[string]bool)
	for _, poi := range pointsOfInterest {
		if poi.GoogleMapsPlaceID != nil {
			recentlyUsed[*poi.GoogleMapsPlaceID] = true
		}
	}

	return recentlyUsed, nil
}
