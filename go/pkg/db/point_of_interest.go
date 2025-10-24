package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestHandle struct {
	db *gorm.DB
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
	// Start a transaction since we'll be deleting multiple related records
	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete related PointOfInterestChallenge records
	if err := tx.Where("point_of_interest_id = ?", id).Delete(&models.PointOfInterestChallenge{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("point_of_interest_id = ?", id).Delete(&models.PointOfInterestDiscovery{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete related PointOfInterestGroupMember records
	if err := tx.Where("point_of_interest_id = ?", id).Delete(&models.PointOfInterestGroupMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("point_of_interest_id = ?", id).Delete(&models.PointOfInterestZone{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the PointOfInterest itself
	if err := tx.Delete(&models.PointOfInterest{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (c *pointOfInterestHandle) UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error {
	return c.db.Model(&models.PointOfInterest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_url":  imageUrl,
		"updated_at": time.Now(),
	}).Error
}

func (c *pointOfInterestHandle) Edit(ctx context.Context, id uuid.UUID, name string, description string, lat string, lng string) error {
	pointOfInterest := models.PointOfInterest{
		ID:          id,
		Name:        name,
		Description: description,
		Lat:         lat,
		Lng:         lng,
		UpdatedAt:   time.Now(),
	}
	if err := pointOfInterest.SetGeometry(lat, lng); err != nil {
		return err
	}

	return c.db.Model(&models.PointOfInterest{}).Where("id = ?", id).Updates(pointOfInterest).Error
}

func (c *pointOfInterestHandle) FindAll(ctx context.Context) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.db.WithContext(ctx).Preload("PointOfInterestChallenges").Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterest, error) {
	var pointOfInterest models.PointOfInterest

	if err := c.db.WithContext(ctx).First(&pointOfInterest, id).Error; err != nil {
		return nil, err
	}

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

	if err := c.db.WithContext(ctx).
		Preload("Tags").
		Preload("PointOfInterestChallenges").
		Where("id IN (?)", ids).
		Find(&pointsOfInterest).Error; err != nil {
		return nil, err
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

	return c.db.WithContext(ctx).Create(&pointOfInterest).Error
}

func (c *pointOfInterestHandle) CreateForGroup(ctx context.Context, pointOfInterest *models.PointOfInterest, pointOfInterestGroupID uuid.UUID) error {
	pointOfInterest.ID = uuid.New()
	pointOfInterest.CreatedAt = time.Now()
	pointOfInterest.UpdatedAt = time.Now()

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
	if err := c.db.WithContext(ctx).Where("google_maps_place_id = ?", googleMapsPlaceID).First(&pointOfInterest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &pointOfInterest, nil
}

func (c *pointOfInterestHandle) Update(ctx context.Context, pointOfInterestID uuid.UUID, updates *models.PointOfInterest) error {
	return c.db.WithContext(ctx).Model(&models.PointOfInterest{}).Where("id = ?", pointOfInterestID).Updates(updates).Error
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
