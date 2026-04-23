package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

func (s *server) resolveOptionalZoneKind(
	ctx context.Context,
	raw string,
) (*models.ZoneKind, error) {
	normalized := models.NormalizeZoneKind(raw)
	if normalized == "" {
		return nil, nil
	}
	if s == nil || s.dbClient == nil {
		return &models.ZoneKind{Slug: normalized}, nil
	}
	zoneKind, err := s.dbClient.ZoneKind().FindBySlug(ctx, normalized)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("zoneKind %q was not found", normalized)
		}
		return nil, err
	}
	return zoneKind, nil
}
