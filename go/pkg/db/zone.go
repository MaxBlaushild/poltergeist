package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneHandler struct {
	db *gorm.DB
}

func (h *zoneHandler) Create(ctx context.Context, zone *models.Zone) error {
	if zone.Boundary == "" {
		zone.Boundary = "010300000000000000" // Empty polygon in WKB hex format
	}
	if zone.InternalTags == nil {
		zone.InternalTags = models.StringArray{}
	}

	return h.db.WithContext(ctx).Create(zone).Error
}

func (h *zoneHandler) FindAll(ctx context.Context) ([]*models.Zone, error) {
	var zones []*models.Zone
	if err := h.db.WithContext(ctx).Preload("Points").Find(&zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

func (h *zoneHandler) FindAdminSummaries(ctx context.Context) ([]models.ZoneAdminSummary, error) {
	const query = `
SELECT
	zones.id,
	zones.created_at,
	zones.updated_at,
	zones.name,
	zones.description,
	zones.internal_tags,
	zones.latitude,
	zones.longitude,
	zones.zone_import_id,
	zone_imports.metro_name AS import_metro_name,
	COALESCE(boundary_points.boundary_point_count, 0) AS boundary_point_count,
	COALESCE(point_of_interest_zones.point_of_interest_count, 0) AS point_of_interest_count,
	COALESCE(quests.quest_count, 0) AS quest_count,
	COALESCE(zone_quest_archetypes.zone_quest_archetype_count, 0) AS zone_quest_archetype_count,
	COALESCE(challenges.challenge_count, 0) AS challenge_count,
	COALESCE(scenarios.scenario_count, 0) AS scenario_count,
	COALESCE(monsters.monster_count, 0) AS monster_count,
	COALESCE(monster_encounters.monster_encounter_count, 0) AS monster_encounter_count,
	COALESCE(treasure_chests.treasure_chest_count, 0) AS treasure_chest_count,
	COALESCE(healing_fountains.healing_fountain_count, 0) AS healing_fountain_count
FROM zones
LEFT JOIN zone_imports
	ON zone_imports.id = zones.zone_import_id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS boundary_point_count
	FROM boundary_points
	GROUP BY zone_id
) AS boundary_points
	ON boundary_points.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS point_of_interest_count
	FROM point_of_interest_zones
	WHERE deleted_at IS NULL
	GROUP BY zone_id
) AS point_of_interest_zones
	ON point_of_interest_zones.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS quest_count
	FROM quests
	GROUP BY zone_id
) AS quests
	ON quests.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS zone_quest_archetype_count
	FROM zone_quest_archetypes
	WHERE deleted_at IS NULL
	GROUP BY zone_id
) AS zone_quest_archetypes
	ON zone_quest_archetypes.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS challenge_count
	FROM challenges
	GROUP BY zone_id
) AS challenges
	ON challenges.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS scenario_count
	FROM scenarios
	GROUP BY zone_id
) AS scenarios
	ON scenarios.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS monster_count
	FROM monsters
	GROUP BY zone_id
) AS monsters
	ON monsters.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS monster_encounter_count
	FROM monster_encounters
	GROUP BY zone_id
) AS monster_encounters
	ON monster_encounters.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS treasure_chest_count
	FROM treasure_chests
	GROUP BY zone_id
) AS treasure_chests
	ON treasure_chests.zone_id = zones.id
LEFT JOIN (
	SELECT zone_id, COUNT(*) AS healing_fountain_count
	FROM healing_fountains
	GROUP BY zone_id
) AS healing_fountains
	ON healing_fountains.zone_id = zones.id
ORDER BY zones.name ASC, zones.created_at DESC
`

	var summaries []models.ZoneAdminSummary
	if err := h.db.WithContext(ctx).Raw(query).Scan(&summaries).Error; err != nil {
		return nil, err
	}
	return summaries, nil
}

func (h *zoneHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.Zone, error) {
	var zone models.Zone
	if err := h.db.WithContext(ctx).Preload("Points").Where("id = ?", id).First(&zone).Error; err != nil {
		return nil, err
	}
	return &zone, nil
}

func (h *zoneHandler) Update(ctx context.Context, zone *models.Zone) error {
	return h.db.WithContext(ctx).Save(zone).Error
}

func (h *zoneHandler) Delete(ctx context.Context, zoneID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.BoundaryPoint{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Unscoped().Where("zone_id = ?", zoneID).Delete(&models.PointOfInterestZone{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.ZoneQuestArchetype{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.UserZoneReputation{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.TreasureChest{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.MovementPattern{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.Quest{}).Error; err != nil {
			return err
		}
		return tx.WithContext(ctx).Delete(&models.Zone{}, "id = ?", zoneID).Error
	})
}

func (h *zoneHandler) DeleteByImportID(ctx context.Context, importID uuid.UUID) (int, error) {
	var zoneIDs []uuid.UUID
	if err := h.db.WithContext(ctx).
		Model(&models.Zone{}).
		Where("zone_import_id = ?", importID).
		Pluck("id", &zoneIDs).Error; err != nil {
		return 0, err
	}
	if len(zoneIDs) == 0 {
		return 0, nil
	}

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("zone_id IN ?", zoneIDs).Delete(&models.BoundaryPoint{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Unscoped().Where("zone_id IN ?", zoneIDs).Delete(&models.PointOfInterestZone{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id IN ?", zoneIDs).Delete(&models.ZoneQuestArchetype{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id IN ?", zoneIDs).Delete(&models.UserZoneReputation{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id IN ?", zoneIDs).Delete(&models.TreasureChest{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id IN ?", zoneIDs).Delete(&models.MovementPattern{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("zone_id IN ?", zoneIDs).Delete(&models.Quest{}).Error; err != nil {
			return err
		}
		return tx.WithContext(ctx).Where("id IN ?", zoneIDs).Delete(&models.Zone{}).Error
	})
	if err != nil {
		return 0, err
	}

	return len(zoneIDs), nil
}

func (h *zoneHandler) AddPointOfInterestToZone(ctx context.Context, zoneID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).Create(&models.PointOfInterestZone{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		ZoneID:            zoneID,
		PointOfInterestID: pointOfInterestID,
	}).Error
}

func (h *zoneHandler) RemovePointOfInterestFromZone(ctx context.Context, zoneID uuid.UUID, pointOfInterestID uuid.UUID) error {
	return h.db.WithContext(ctx).
		Where("zone_id = ? AND point_of_interest_id = ?", zoneID, pointOfInterestID).
		Delete(&models.PointOfInterestZone{}).Error
}

func (h *zoneHandler) UpdateBoundary(ctx context.Context, zoneID uuid.UUID, boundary [][]float64) error {
	// Create a PostGIS Polygon WKT string
	// Format: POLYGON((lng1 lat1, lng2 lat2, ...))
	var coords []string
	for _, point := range boundary {
		// Format coordinates with 6 decimal places and ensure proper spacing
		coords = append(coords, fmt.Sprintf("%.6f %.6f", point[0], point[1]))
	}
	// Close the polygon by repeating the first point
	if len(boundary) > 0 {
		coords = append(coords, fmt.Sprintf("%.6f %.6f", boundary[0][0], boundary[0][1]))
	}

	// Format the WKT string with proper spacing and parentheses
	wkt := fmt.Sprintf("POLYGON((%s))", strings.Join(coords, ", "))

	// Use ST_GeomFromText to properly parse the WKT
	updates := map[string]interface{}{
		"boundary": gorm.Expr("ST_GeomFromText(?, 4326)", wkt),
	}

	if err := h.db.WithContext(ctx).Model(&models.Zone{}).Where("id = ?", zoneID).Updates(updates).Error; err != nil {
		return err
	}

	pointHandle := pointHandler{db: h.db}

	if err := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Delete(&models.BoundaryPoint{}).Error; err != nil {
		return err
	}

	for _, point := range boundary {
		// boundary is [lng, lat]
		point, err := pointHandle.CreatePoint(ctx, point[1], point[0])
		if err != nil {
			return err
		}

		if err := h.db.WithContext(ctx).Create(&models.BoundaryPoint{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ZoneID:    zoneID,
			PointID:   point.ID,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *zoneHandler) FindByPointOfInterestID(ctx context.Context, pointOfInterestID uuid.UUID) (*models.Zone, error) {
	var pointOfInterestZone models.PointOfInterestZone
	if err := h.db.WithContext(ctx).Where("point_of_interest_id = ?", pointOfInterestID).First(&pointOfInterestZone).Error; err != nil {
		return nil, err
	}

	var zone models.Zone
	if err := h.db.WithContext(ctx).Where("id = ?", pointOfInterestZone.ZoneID).First(&zone).Error; err != nil {
		return nil, err
	}

	return &zone, nil
}

func (h *zoneHandler) UpdateNameAndDescription(ctx context.Context, zoneID uuid.UUID, name string, description string) error {
	return h.db.WithContext(ctx).Model(&models.Zone{}).Where("id = ?", zoneID).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).Error
}

func (h *zoneHandler) UpdateMetadata(ctx context.Context, zoneID uuid.UUID, name string, description string, internalTags models.StringArray) (*models.Zone, error) {
	if internalTags == nil {
		internalTags = models.StringArray{}
	}
	if err := h.db.WithContext(ctx).Model(&models.Zone{}).Where("id = ?", zoneID).Updates(map[string]interface{}{
		"name":          name,
		"description":   description,
		"internal_tags": internalTags,
	}).Error; err != nil {
		return nil, err
	}

	return h.FindByID(ctx, zoneID)
}
