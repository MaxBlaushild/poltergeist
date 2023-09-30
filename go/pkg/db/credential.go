package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type credentialHandle struct {
	db *gorm.DB
}

func (h *credentialHandle) Insert(ctx context.Context, credentialID string, publicKey string, userID uint) (*models.Credential, error) {
	credential := models.Credential{
		CredentialID: credentialID,
		PublicKey:    publicKey,
		UserID:       userID,
	}

	if err := h.db.WithContext(ctx).Create(&credential).Error; err != nil {
		return nil, err
	}

	return &credential, nil
}

func (h *credentialHandle) FindAll(ctx context.Context) ([]models.Credential, error) {
	var credentials []models.Credential

	if err := h.db.WithContext(ctx).Find(&credentials).Error; err != nil {
		return nil, err
	}

	return credentials, nil
}

func (h *credentialHandle) Delete(ctx context.Context, credentialID uint) error {
	return h.db.WithContext(ctx).Delete(&models.Credential{}, credentialID).Error
}

func (h *credentialHandle) DeleteAll(ctx context.Context) error {
	return h.db.WithContext(ctx).Where("1 = 1").Delete(&models.Credential{}).Error
}
