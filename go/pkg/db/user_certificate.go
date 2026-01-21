package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userCertificateHandle struct {
	db *gorm.DB
}

func (h *userCertificateHandle) Create(ctx context.Context, userID uuid.UUID, certificateDER []byte, certificatePEM string, publicKeyPEM string, fingerprint []byte) (*models.UserCertificate, error) {
	cert := &models.UserCertificate{
		UserID:        userID,
		Certificate:   certificateDER,
		CertificatePEM: certificatePEM,
		PublicKey:     publicKeyPEM,
		Fingerprint:   fingerprint,
		Active:        false, // Certificates are created as inactive by default
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.db.WithContext(ctx).Create(cert).Error; err != nil {
		return nil, err
	}

	return cert, nil
}

func (h *userCertificateHandle) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserCertificate, error) {
	var cert models.UserCertificate
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).First(&cert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (h *userCertificateHandle) FindByFingerprint(ctx context.Context, fingerprint []byte) (*models.UserCertificate, error) {
	var cert models.UserCertificate
	if err := h.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&cert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (h *userCertificateHandle) UpdateActive(ctx context.Context, userID uuid.UUID, active bool) error {
	return h.db.WithContext(ctx).
		Model(&models.UserCertificate{}).
		Where("user_id = ?", userID).
		Update("active", active).Error
}

func (h *userCertificateHandle) Delete(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.UserCertificate{}, "user_id = ?", userID).Error
}
