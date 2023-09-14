package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type crystalHandle struct {
	db *gorm.DB
}

func (c *crystalHandle) FindAll(ctx context.Context) ([]models.Crystal, error) {
	var crystals []models.Crystal

	if err := c.db.WithContext(ctx).Find(&crystals).Error; err != nil {
		return nil, err
	}

	return crystals, nil
}

func (c *crystalHandle) FindByID(ctx context.Context, id uint) (*models.Crystal, error) {
	var crystal models.Crystal

	if err := c.db.WithContext(ctx).First(&crystal, id).Error; err != nil {
		return nil, err
	}

	return &crystal, nil
}

func (c *crystalHandle) Capture(ctx context.Context, crystalID uint, teamID uint, attune bool) error {
	updates := models.Crystal{
		Attuned:       attune,
		Captured:      !attune,
		CaptureTeamID: teamID,
	}

	return c.db.WithContext(ctx).Model(&models.Crystal{}).Where("id = ?", crystalID).Updates(&updates).Error
}

func (c *crystalHandle) Unlock(ctx context.Context, crystalID uint, teamID uint) error {
	unlock := models.CrystalUnlocking{
		TeamID:    teamID,
		CrystalID: crystalID,
	}

	return c.db.WithContext(ctx).Create(&unlock).Error
}

func (c *crystalHandle) Create(ctx context.Context, crystal models.Crystal) error {
	return c.db.WithContext(ctx).Create(&crystal).Error
}
