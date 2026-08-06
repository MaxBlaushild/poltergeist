package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// bgiEventHandle deliberately only implements Create for v1 — no
// CountByType/CountRejectionsByRule, since the operator dashboard is out of
// scope for this vertical slice (see go/bgi-site/PLATFORM_FINDINGS.md). Add
// them when that work actually starts; the raw event stream is already
// being captured either way.
type bgiEventHandle struct {
	db *gorm.DB
}

func (h *bgiEventHandle) Create(ctx context.Context, event *models.BgiEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	return h.db.WithContext(ctx).Create(event).Error
}
