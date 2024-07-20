package db

import (
	"context"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sonarUserHandle struct {
	db *gorm.DB
}

const (
	totalNumberOfProfileIcons = 29
)

func (h *sonarUserHandle) FindOrCreateSonarUser(ctx context.Context, viewerID uuid.UUID, vieweeID uuid.UUID) error {
	sonarUser, err := h.FindUserByViewerAndViewee(ctx, viewerID, vieweeID)
	if err == gorm.ErrRecordNotFound {
		profileIcon, err := h.GetSonarUserProfileIcon(ctx, viewerID)
		if err != nil {
			return err
		}

		sonarUser = &models.SonarUser{
			ViewerID:          viewerID,
			VieweeID:          vieweeID,
			ProfilePictureUrl: profileIcon,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ID:                uuid.New(),
		}
	}
	if err != nil {
		return err
	}

	return h.db.WithContext(ctx).Create(sonarUser).Error
}

func (h *sonarUserHandle) GetSonarUserCount(ctx context.Context, viewerID uuid.UUID) (int64, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&models.SonarUser{}).Where("viewer_id = ?", viewerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (h *sonarUserHandle) FindUserByViewerAndViewee(ctx context.Context, viewerID uuid.UUID, vieweeID uuid.UUID) (*models.SonarUser, error) {
	sonarUser := &models.SonarUser{}
	if err := h.db.WithContext(ctx).Where("viewer_id = ? AND viewee_id = ?", viewerID, vieweeID).First(sonarUser).Error; err != nil {
		return nil, err
	}
	return sonarUser, nil
}

func (h *sonarUserHandle) GetSonarUserProfileIcon(ctx context.Context, viewerID uuid.UUID) (string, error) {
	count, err := h.GetSonarUserCount(ctx, viewerID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://crew-profile-icons.s3.amazonaws.com/%d.png", count%totalNumberOfProfileIcons+1), nil
}
