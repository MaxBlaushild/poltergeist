package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneHandler struct {
	db *gorm.DB
}

func normalizeZoneIDs(zoneIDs []uuid.UUID) []uuid.UUID {
	if len(zoneIDs) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]struct{}, len(zoneIDs))
	normalized := make([]uuid.UUID, 0, len(zoneIDs))
	for _, zoneID := range zoneIDs {
		if zoneID == uuid.Nil {
			continue
		}
		if _, exists := seen[zoneID]; exists {
			continue
		}
		seen[zoneID] = struct{}{}
		normalized = append(normalized, zoneID)
	}
	return normalized
}

func zoneCharacterCoordinateVisible(lat, lng float64) bool {
	if lat > 90 || lat < -90 || lng > 180 || lng < -180 {
		return false
	}
	return lat != 0 || lng != 0
}

func zoneContainsCharacter(zone *models.Zone, zonePoiIDs map[uuid.UUID]struct{}, character *models.Character) bool {
	if zone == nil || character == nil {
		return false
	}
	if character.PointOfInterestID != nil {
		if _, ok := zonePoiIDs[*character.PointOfInterestID]; ok {
			return true
		}
	}
	if character.PointOfInterest != nil {
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(character.PointOfInterest.Lat), 64)
		lng, lngErr := strconv.ParseFloat(strings.TrimSpace(character.PointOfInterest.Lng), 64)
		if latErr == nil && lngErr == nil && zoneCharacterCoordinateVisible(lat, lng) {
			return zone.IsPointInBoundary(lat, lng)
		}
	}
	for _, location := range character.Locations {
		if !zoneCharacterCoordinateVisible(location.Latitude, location.Longitude) {
			continue
		}
		if zone.IsPointInBoundary(location.Latitude, location.Longitude) {
			return true
		}
	}
	return false
}

func deleteZoneCharacters(tx *gorm.DB, characterIDs []uuid.UUID) (int, error) {
	if len(characterIDs) == 0 {
		return 0, nil
	}

	if err := tx.Model(&models.ZoneQuestArchetype{}).
		Where("character_id IN ?", characterIDs).
		Update("character_id", nil).Error; err != nil {
		return 0, err
	}

	if err := tx.Model(&models.Quest{}).
		Where("quest_giver_character_id IN ?", characterIDs).
		Update("quest_giver_character_id", nil).Error; err != nil {
		return 0, err
	}

	if err := tx.Model(&models.QuestArchetype{}).
		Where("quest_giver_character_id IN ?", characterIDs).
		Update("quest_giver_character_id", nil).Error; err != nil {
		return 0, err
	}

	if err := tx.Model(&models.PointOfInterestGroup{}).
		Where("quest_giver_character_id IN ?", characterIDs).
		Update("quest_giver_character_id", nil).Error; err != nil {
		return 0, err
	}

	if err := tx.Where("character_id IN ?", characterIDs).
		Delete(&models.QuestAcceptance{}).Error; err != nil {
		return 0, err
	}

	result := tx.Where("id IN ?", characterIDs).Delete(&models.Character{})
	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
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
	COALESCE(monster_encounters.standard_encounter_count, 0) AS standard_encounter_count,
	COALESCE(monster_encounters.boss_encounter_count, 0) AS boss_encounter_count,
	COALESCE(monster_encounters.raid_encounter_count, 0) AS raid_encounter_count,
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
	SELECT
		zone_id,
		COUNT(*) AS monster_encounter_count,
		SUM(CASE WHEN encounter_type = 'monster' OR encounter_type IS NULL OR encounter_type = '' THEN 1 ELSE 0 END) AS standard_encounter_count,
		SUM(CASE WHEN encounter_type = 'boss' THEN 1 ELSE 0 END) AS boss_encounter_count,
		SUM(CASE WHEN encounter_type = 'raid' THEN 1 ELSE 0 END) AS raid_encounter_count
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

func (h *zoneHandler) FlushContent(ctx context.Context, zoneIDs []uuid.UUID, options models.ZoneContentFlushOptions) (*models.ZoneContentFlushSummary, error) {
	normalizedZoneIDs := normalizeZoneIDs(zoneIDs)
	if len(normalizedZoneIDs) == 0 {
		return &models.ZoneContentFlushSummary{}, nil
	}
	if len(options.ContentTypes) == 0 {
		options = models.DefaultZoneContentFlushOptions()
	}

	summary := &models.ZoneContentFlushSummary{
		ZoneCount: len(normalizedZoneIDs),
	}

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var zones []models.Zone
		if err := tx.WithContext(ctx).
			Preload("Points").
			Where("id IN ?", normalizedZoneIDs).
			Find(&zones).Error; err != nil {
			return err
		}
		if len(zones) != len(normalizedZoneIDs) {
			return fmt.Errorf("one or more zones were not found")
		}

		zoneByID := make(map[uuid.UUID]*models.Zone, len(zones))
		for idx := range zones {
			zone := &zones[idx]
			zoneByID[zone.ID] = zone
		}

		shouldFlushPointsOfInterest := options.Includes(models.ZoneContentFlushTypePointsOfInterest)
		shouldFlushQuests := options.Includes(models.ZoneContentFlushTypeQuests)
		shouldFlushChallenges := options.Includes(models.ZoneContentFlushTypeChallenges)
		shouldFlushScenarios := options.Includes(models.ZoneContentFlushTypeScenarios)
		shouldFlushExpositions := options.Includes(models.ZoneContentFlushTypeExpositions)
		shouldFlushMonsters := options.Includes(models.ZoneContentFlushTypeMonsters)
		shouldFlushTreasureChests := options.Includes(models.ZoneContentFlushTypeTreasureChests)
		shouldFlushHealingFountains := options.Includes(models.ZoneContentFlushTypeHealingFountains)
		shouldFlushResources := options.Includes(models.ZoneContentFlushTypeResources)
		shouldFlushMovementPatterns := options.Includes(models.ZoneContentFlushTypeMovementPatterns)
		shouldFlushJobs := options.Includes(models.ZoneContentFlushTypeJobs)

		orphanPointOfInterestIDs := make([]uuid.UUID, 0)
		sharedPointOfInterestIDs := make([]uuid.UUID, 0)
		characterIDsToDelete := make([]uuid.UUID, 0)
		if shouldFlushPointsOfInterest {
			type zonePointOfInterestRow struct {
				ZoneID            uuid.UUID `gorm:"column:zone_id"`
				PointOfInterestID uuid.UUID `gorm:"column:point_of_interest_id"`
			}

			var pointOfInterestRows []zonePointOfInterestRow
			if err := tx.WithContext(ctx).
				Model(&models.PointOfInterestZone{}).
				Select("zone_id, point_of_interest_id").
				Where("deleted_at IS NULL").
				Where("zone_id IN ?", normalizedZoneIDs).
				Find(&pointOfInterestRows).Error; err != nil {
				return err
			}

			allZonePointOfInterestIDs := make([]uuid.UUID, 0, len(pointOfInterestRows))
			zonePointOfInterestIDs := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(normalizedZoneIDs))
			pointOfInterestSet := make(map[uuid.UUID]struct{}, len(pointOfInterestRows))
			for _, row := range pointOfInterestRows {
				if _, exists := zonePointOfInterestIDs[row.ZoneID]; !exists {
					zonePointOfInterestIDs[row.ZoneID] = map[uuid.UUID]struct{}{}
				}
				zonePointOfInterestIDs[row.ZoneID][row.PointOfInterestID] = struct{}{}
				if _, exists := pointOfInterestSet[row.PointOfInterestID]; exists {
					continue
				}
				pointOfInterestSet[row.PointOfInterestID] = struct{}{}
				allZonePointOfInterestIDs = append(allZonePointOfInterestIDs, row.PointOfInterestID)
			}

			if len(allZonePointOfInterestIDs) > 0 {
				var pointOfInterestLinkCounts []struct {
					PointOfInterestID uuid.UUID `gorm:"column:point_of_interest_id"`
					RemainingCount    int64     `gorm:"column:remaining_count"`
				}
				if err := tx.WithContext(ctx).
					Model(&models.PointOfInterestZone{}).
					Select("point_of_interest_id, COUNT(*) AS remaining_count").
					Where("deleted_at IS NULL").
					Where("point_of_interest_id IN ?", allZonePointOfInterestIDs).
					Where("zone_id NOT IN ?", normalizedZoneIDs).
					Group("point_of_interest_id").
					Find(&pointOfInterestLinkCounts).Error; err != nil {
					return err
				}

				remainingLinkCountByPOI := make(map[uuid.UUID]int64, len(pointOfInterestLinkCounts))
				for _, row := range pointOfInterestLinkCounts {
					remainingLinkCountByPOI[row.PointOfInterestID] = row.RemainingCount
				}

				for _, pointOfInterestID := range allZonePointOfInterestIDs {
					if remainingLinkCountByPOI[pointOfInterestID] > 0 {
						sharedPointOfInterestIDs = append(sharedPointOfInterestIDs, pointOfInterestID)
						continue
					}
					orphanPointOfInterestIDs = append(orphanPointOfInterestIDs, pointOfInterestID)
				}
			}

			sharedPointOfInterestSet := make(map[uuid.UUID]struct{}, len(sharedPointOfInterestIDs))
			for _, pointOfInterestID := range sharedPointOfInterestIDs {
				sharedPointOfInterestSet[pointOfInterestID] = struct{}{}
			}

			characterQuery := tx.WithContext(ctx).
				Preload("PointOfInterest").
				Preload("Locations")
			if len(allZonePointOfInterestIDs) > 0 {
				characterQuery = characterQuery.Where(
					"point_of_interest_id IN ? OR id IN (SELECT DISTINCT character_id FROM character_locations)",
					allZonePointOfInterestIDs,
				)
			} else {
				characterQuery = characterQuery.Where("id IN (SELECT DISTINCT character_id FROM character_locations)")
			}

			var characters []models.Character
			if err := characterQuery.Find(&characters).Error; err != nil {
				return err
			}

			characterIDSet := make(map[uuid.UUID]struct{}, len(characters))
			for idx := range characters {
				character := &characters[idx]
				if character.PointOfInterestID != nil {
					if _, shared := sharedPointOfInterestSet[*character.PointOfInterestID]; shared {
						continue
					}
				}
				for _, zoneID := range normalizedZoneIDs {
					zone := zoneByID[zoneID]
					if zone == nil {
						continue
					}
					if !zoneContainsCharacter(zone, zonePointOfInterestIDs[zoneID], character) {
						continue
					}
					if _, exists := characterIDSet[character.ID]; exists {
						break
					}
					characterIDSet[character.ID] = struct{}{}
					characterIDsToDelete = append(characterIDsToDelete, character.ID)
					break
				}
			}
		}

		result := tx.WithContext(ctx)
		if shouldFlushJobs {
			result = result.Where("zone_id IN ?", normalizedZoneIDs).Delete(&models.QuestGenerationJob{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedQuestGenerationJobCount = int(result.RowsAffected)

			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.ScenarioGenerationJob{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedScenarioGenerationJobCount = int(result.RowsAffected)

			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.ChallengeGenerationJob{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedChallengeGenerationJobCount = int(result.RowsAffected)

			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.ZoneSeedJob{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedZoneSeedJobCount = int(result.RowsAffected)
		}

		if shouldFlushQuests {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.ZoneQuestArchetype{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedZoneQuestArchetypeCount = int(result.RowsAffected)

			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.Quest{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedQuestCount = int(result.RowsAffected)
		}

		if len(characterIDsToDelete) > 0 {
			deletedCharacterCount, err := deleteZoneCharacters(tx.WithContext(ctx), characterIDsToDelete)
			if err != nil {
				return err
			}
			summary.DeletedCharacterCount = deletedCharacterCount
		}

		if shouldFlushChallenges {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.Challenge{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedChallengeCount = int(result.RowsAffected)
		}

		if shouldFlushScenarios {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.Scenario{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedScenarioCount = int(result.RowsAffected)
		}

		if shouldFlushExpositions {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.Exposition{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedExpositionCount = int(result.RowsAffected)
		}

		if shouldFlushMonsters {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.MonsterEncounter{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedMonsterEncounterCount = int(result.RowsAffected)

			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.Monster{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedMonsterCount = int(result.RowsAffected)
		}

		if shouldFlushTreasureChests {
			var treasureChestIDs []uuid.UUID
			if err := tx.WithContext(ctx).
				Model(&models.TreasureChest{}).
				Where("zone_id IN ?", normalizedZoneIDs).
				Pluck("id", &treasureChestIDs).Error; err != nil {
				return err
			}
			if len(treasureChestIDs) > 0 {
				if err := deleteTreasureChests(tx, treasureChestIDs); err != nil {
					return err
				}
			}
			summary.DeletedTreasureChestCount = len(treasureChestIDs)
		}

		if shouldFlushHealingFountains {
			var healingFountainIDs []uuid.UUID
			if err := tx.WithContext(ctx).
				Model(&models.HealingFountain{}).
				Where("zone_id IN ?", normalizedZoneIDs).
				Pluck("id", &healingFountainIDs).Error; err != nil {
				return err
			}
			if len(healingFountainIDs) > 0 {
				if err := deleteHealingFountains(tx, healingFountainIDs); err != nil {
					return err
				}
			}
			summary.DeletedHealingFountainCount = len(healingFountainIDs)
		}

		if shouldFlushResources {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.Resource{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedResourceCount = int(result.RowsAffected)
		}

		if shouldFlushMovementPatterns {
			result = tx.WithContext(ctx).
				Where("zone_id IN ?", normalizedZoneIDs).
				Delete(&models.MovementPattern{})
			if result.Error != nil {
				return result.Error
			}
			summary.DeletedMovementPatternCount = int(result.RowsAffected)
		}

		if shouldFlushPointsOfInterest {
			if len(sharedPointOfInterestIDs) > 0 {
				result = tx.WithContext(ctx).
					Unscoped().
					Where("zone_id IN ?", normalizedZoneIDs).
					Where("point_of_interest_id IN ?", sharedPointOfInterestIDs).
					Delete(&models.PointOfInterestZone{})
				if result.Error != nil {
					return result.Error
				}
			}
			summary.DetachedSharedPointOfInterestCount = len(sharedPointOfInterestIDs)

			if len(orphanPointOfInterestIDs) > 0 {
				if err := deletePointsOfInterest(tx, orphanPointOfInterestIDs); err != nil {
					return err
				}
			}
			summary.DeletedPointOfInterestCount = len(orphanPointOfInterestIDs)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return summary, nil
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
	var existingCount int64
	if err := h.db.WithContext(ctx).
		Model(&models.PointOfInterestZone{}).
		Where("zone_id = ? AND point_of_interest_id = ?", zoneID, pointOfInterestID).
		Where("deleted_at IS NULL").
		Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount > 0 {
		return nil
	}

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
