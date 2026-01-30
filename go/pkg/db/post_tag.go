package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postTagHandle struct {
	db *gorm.DB
}

func (h *postTagHandle) CreateForPost(ctx context.Context, postID uuid.UUID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	records := make([]models.PostTag, 0, len(tags))
	seen := make(map[string]bool)
	for _, tag := range tags {
		t := trimTag(tag)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		records = append(records, models.PostTag{PostID: postID, Tag: t})
	}
	if len(records) == 0 {
		return nil
	}
	return h.db.WithContext(ctx).Create(&records).Error
}

func (h *postTagHandle) FindByPostIDs(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	if len(postIDs) == 0 {
		return make(map[uuid.UUID][]string), nil
	}
	var rows []models.PostTag
	if err := h.db.WithContext(ctx).Where("post_id IN ?", postIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]string)
	for _, r := range rows {
		result[r.PostID] = append(result[r.PostID], r.Tag)
	}
	return result, nil
}

func trimTag(s string) string {
	// Trim whitespace; limit length
	b := 0
	for b < len(s) && (s[b] == ' ' || s[b] == '\t') {
		b++
	}
	e := len(s)
	for e > b && (s[e-1] == ' ' || s[e-1] == '\t') {
		e--
	}
	t := s[b:e]
	if len(t) > 64 {
		t = t[:64]
	}
	return t
}
