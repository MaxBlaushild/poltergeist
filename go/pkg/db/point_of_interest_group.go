package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestGroupHandle struct {
	db *gorm.DB
}

func (c *pointOfInterestGroupHandle) GetNearbyQuests(ctx context.Context, userID uuid.UUID, lat float64, lng float64, radiusInMeters float64, tags []string) ([]models.PointOfInterestGroup, error) {
	var groups []models.PointOfInterestGroup
	query := c.db.WithContext(ctx).
		Distinct("pog.*").
		Table("point_of_interest_groups pog").
		Joins("JOIN point_of_interest_group_members pogm ON pogm.point_of_interest_group_id = pog.id").
		Joins("JOIN points_of_interest poi ON poi.id = pogm.point_of_interest_id").
		Joins("JOIN tag_entities te ON te.point_of_interest_id = poi.id").
		Joins("JOIN tags t ON t.id = te.tag_id").
		Where("ST_DWithin(poi.geometry, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			lng, lat, radiusInMeters).
		// Left join with children to check if point is a child
		Joins("LEFT JOIN point_of_interest_children poic ON poic.point_of_interest_id = poi.id").
		Where("pog.type = ?", models.PointOfInterestGroupTypeQuest).
		Where("poic.point_of_interest_id IS NULL"). // Only get points that don't appear as children
		Where("t.value IN ?", tags).
		Preload("GroupMembers").
		Preload("GroupMembers.PointOfInterest").
		Preload("GroupMembers.PointOfInterest.PointOfInterestChallenges").
		Preload("GroupMembers.Children").
		Preload("GroupMembers.Children.PointOfInterest").
		Find(&groups)

	if err := query.Error; err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *pointOfInterestGroupHandle) GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) {
	var groups []models.PointOfInterestGroup
	query := c.db.WithContext(ctx).
		Distinct("pog.*").
		Table("point_of_interest_groups pog").
		Joins("JOIN point_of_interest_group_members pogm ON pogm.point_of_interest_group_id = pog.id").
		Joins("JOIN point_of_interest_challenges poc ON poc.point_of_interest_id = pogm.point_of_interest_id").
		Joins("JOIN point_of_interest_challenge_submissions pocs ON pocs.point_of_interest_challenge_id = poc.id").
		Preload("GroupMembers").
		Preload("GroupMembers.PointOfInterest").
		Preload("GroupMembers.PointOfInterest.PointOfInterestChallenges").
		Preload("GroupMembers.Children").
		Preload("GroupMembers.Children.PointOfInterest").
		Where("pocs.user_id = ?", userID).
		Where("pog.type = ?", models.PointOfInterestGroupTypeQuest).
		Find(&groups)

	if err := query.Error; err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *pointOfInterestGroupHandle) Delete(ctx context.Context, id uuid.UUID) error {
	// Start a transaction since we'll be deleting multiple related records
	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}

	// Delete related PointOfInterestGroupMember records
	if err := tx.Where("point_of_interest_group_id = ?", id).Delete(&models.PointOfInterestGroupMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the PointOfInterestGroup itself
	if err := tx.Delete(&models.PointOfInterestGroup{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (c *pointOfInterestGroupHandle) UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error {
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_url":  imageUrl,
		"updated_at": time.Now(),
	}).Error
}

func (c *pointOfInterestGroupHandle) Edit(ctx context.Context, id uuid.UUID, name string, description string, typeValue models.PointOfInterestGroupType) error {
	group := models.PointOfInterestGroup{
		Name:        name,
		Description: description,
		Type:        typeValue,
		UpdatedAt:   time.Now(),
	}
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", id).Updates(group).Error
}

func (c *pointOfInterestGroupHandle) Create(ctx context.Context, name string, description string, imageUrl string, typeValue models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) {

	pointOfInterestGroup := models.PointOfInterestGroup{
		Name:        name,
		Description: description,
		ImageUrl:    imageUrl,
		Type:        typeValue,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := c.db.Create(&pointOfInterestGroup).Error; err != nil {
		return nil, err
	}

	return &pointOfInterestGroup, nil
}

func (c *pointOfInterestGroupHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error) {
	var pointOfInterestGroup models.PointOfInterestGroup
	if err := c.db.Preload("PointsOfInterest.PointOfInterestChallenges").Preload("GroupMembers.Children").First(&pointOfInterestGroup, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pointOfInterestGroup, nil
}

func (c *pointOfInterestGroupHandle) FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error) {
	var pointOfInterestGroups []*models.PointOfInterestGroup
	if err := c.db.Preload("PointsOfInterest.PointOfInterestChallenges").Preload("GroupMembers.Children").Find(&pointOfInterestGroups).Error; err != nil {
		return nil, err
	}
	return pointOfInterestGroups, nil
}

func (c *pointOfInterestGroupHandle) FindByType(ctx context.Context, typeValue models.PointOfInterestGroupType) ([]*models.PointOfInterestGroup, error) {
	var pointOfInterestGroups []*models.PointOfInterestGroup
	if err := c.db.Preload("PointsOfInterest.PointOfInterestChallenges").Preload("GroupMembers.Children").Where("type = ?", typeValue).Find(&pointOfInterestGroups).Error; err != nil {
		return nil, err
	}
	return pointOfInterestGroups, nil
}
