package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// AddSuperUser grants shared-content-library edit access. createdBy is nil
// for the ops-only bootstrap grant (cmd/grant-super-user).
func (h *vampireHandler) AddSuperUser(ctx context.Context, userID uuid.UUID, createdBy *uuid.UUID) error {
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoNothing: true}).
		Create(&models.VampireSuperUser{UserID: userID, CreatedBy: createdBy}).Error
}

func (h *vampireHandler) RemoveSuperUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireSuperUser{}, "user_id = ?", userID).Error
}

func (h *vampireHandler) IsSuperUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&models.VampireSuperUser{}).
		Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *vampireHandler) ListSuperUsers(ctx context.Context) ([]models.VampireSuperUser, error) {
	var rows []models.VampireSuperUser
	if err := h.db.WithContext(ctx).Preload("User").Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// LogSuperUserAction audits an edit to the shared content library.
func (h *vampireHandler) LogSuperUserAction(ctx context.Context, userID uuid.UUID, action string, payload []byte) error {
	entry := models.VampireSuperUserActionLog{UserID: &userID, Action: action, Payload: payload}
	return h.db.WithContext(ctx).Create(&entry).Error
}
