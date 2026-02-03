package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRecentPostTagHandle struct {
	db *gorm.DB
}

func (h *userRecentPostTagHandle) Upsert(ctx context.Context, userID uuid.UUID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	now := time.Now()
	records := make([]models.UserRecentPostTag, 0, len(tags))
	seen := make(map[string]bool)
	for _, tag := range tags {
		t := trimTag(tag)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		records = append(records, models.UserRecentPostTag{
			UserID:       userID,
			Tag:          t,
			LastPostedAt: now,
		})
	}
	if len(records) == 0 {
		return nil
	}
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "tag"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_posted_at"}),
	}).Create(&records).Error
}

func (h *userRecentPostTagHandle) FindRecentByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.UserRecentPostTag
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_posted_at DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	tags := make([]string, len(rows))
	for i := range rows {
		tags[i] = rows[i].Tag
	}
	return tags, nil
}
