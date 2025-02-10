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

func (c *pointOfInterestGroupHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return c.db.Delete(&models.PointOfInterestGroup{}, "id = ?", id).Error
}

func (c *pointOfInterestGroupHandle) UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error {
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_url":  imageUrl,
		"updated_at": time.Now(),
	}).Error
}

func (c *pointOfInterestGroupHandle) Edit(ctx context.Context, id uuid.UUID, name string, description string) error {
	group := models.PointOfInterestGroup{
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now(),
	}
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", id).Updates(group).Error
}

func (c *pointOfInterestGroupHandle) Create(ctx context.Context, name string, description string, imageUrl string) (*models.PointOfInterestGroup, error) {

	pointOfInterestGroup := models.PointOfInterestGroup{
		Name:        name,
		Description: description,
		ImageUrl:    imageUrl,
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
