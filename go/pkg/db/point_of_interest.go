package db

import (
	"context"
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
	return c.db.Model(&models.PointOfInterest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"lat":         lat,
		"lng":         lng,
		"updated_at":  time.Now(),
	}).Error
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

	if err := c.db.WithContext(ctx).Where("id IN (?)", ids).Find(&pointsOfInterest).Error; err != nil {
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

	if err := c.db.WithContext(ctx).Create(&pointOfInterestGroupMember).Error; err != nil {
		return err
	}

	return nil
}

func (c *pointOfInterestHandle) CreateWithChallenges(ctx context.Context, request *CreatePointOfInterestRequest) error {
	// Start a transaction since we need to create multiple related records
	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}

	updatedAt := time.Now()
	createdAt := updatedAt
	pointOfInterestID := uuid.New()

	// Create the point of interest first
	pointOfInterest := models.PointOfInterest{
		ID:          pointOfInterestID,
		Name:        request.Name,
		Description: request.Description,
		Lat:         request.Lat,
		Lng:         request.Lon,
		ImageUrl:    request.ImageUrl,
		Clue:        request.Clue,
		UpdatedAt:   updatedAt,
		CreatedAt:   createdAt,
	}

	if err := tx.Create(&pointOfInterest).Error; err != nil {
		tx.Rollback()
		return err
	}

	pointOfInterestGroupID, err := uuid.Parse(request.PointOfInterestGroupMemberID)
	if err != nil {
		tx.Rollback()
		return err
	}

	pointOfInterestGroupMember := models.PointOfInterestGroupMember{
		ID:                     uuid.New(),
		PointOfInterestID:      pointOfInterestID,
		PointOfInterestGroupID: pointOfInterestGroupID,
	}

	if err := tx.Create(&pointOfInterestGroupMember).Error; err != nil {
		tx.Rollback()
		return err
	}

	tierOneChallenge := models.PointOfInterestChallenge{
		ID:                uuid.New(),
		PointOfInterestID: pointOfInterestID,
		Question:          request.TierOne,
		Tier:              1,
		InventoryItemID:   request.TierOneInventoryItemId,
	}

	if err := tx.Create(&tierOneChallenge).Error; err != nil {
		tx.Rollback()
		return err
	}

	if request.TierTwo != nil && *request.TierTwo != "" {
		tierTwoChallenge := models.PointOfInterestChallenge{
			ID:                uuid.New(),
			PointOfInterestID: pointOfInterestID,
			Question:          *request.TierTwo,
			Tier:              2,
			InventoryItemID:   request.TierTwoInventoryItemId,
		}

		if err := tx.Create(&tierTwoChallenge).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if request.TierThree != nil && *request.TierThree != "" {
		tierThreeChallenge := models.PointOfInterestChallenge{
			ID:                uuid.New(),
			PointOfInterestID: pointOfInterestID,
			Question:          *request.TierThree,
			Tier:              3,
			InventoryItemID:   request.TierThreeInventoryItemId,
		}

		if err := tx.Create(&tierThreeChallenge).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction if everything succeeded
	return tx.Commit().Error
}
