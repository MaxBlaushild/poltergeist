package processors

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

func loadOptionalZoneKind(
	ctx context.Context,
	dbClient db.DbClient,
	raw string,
) (*models.ZoneKind, error) {
	normalized := models.NormalizeZoneKind(raw)
	if normalized == "" {
		return nil, nil
	}
	if dbClient == nil {
		return &models.ZoneKind{Slug: normalized}, nil
	}
	zoneKind, err := dbClient.ZoneKind().FindBySlug(ctx, normalized)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.ZoneKind{Slug: normalized}, nil
		}
		return nil, err
	}
	return zoneKind, nil
}

func zoneKindInstructionBlock(zoneKind *models.ZoneKind) string {
	if zoneKind == nil {
		return ""
	}
	label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	slug := strings.TrimSpace(models.ZoneKindPromptSlug(zoneKind))
	seed := strings.TrimSpace(models.ZoneKindPromptSeed(zoneKind))
	if label == "" && slug == "" && seed == "" {
		return ""
	}
	if label == "" {
		label = slug
	}
	if seed == "" {
		seed = fmt.Sprintf(
			"Keep the content naturally suited to %s zones.",
			strings.ToLower(label),
		)
	}
	return fmt.Sprintf(
		`Zone kind direction:
- zone kind: %s
- slug: %s
- creative seed: %s

Additional rules:
- The content should feel naturally suited to %s zones while remaining reusable across many places of that kind.
- Let the environment influence props, hazards, traversal, factions, and scene logic.
`,
		label,
		slug,
		seed,
		strings.ToLower(label),
	)
}
