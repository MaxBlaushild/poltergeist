package db

import (
	"context"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

var zoneKindReferenceTables = []string{
	"points_of_interest",
	"quests",
	"challenges",
	"scenarios",
	"expositions",
	"monsters",
	"monster_encounters",
	"treasure_chests",
	"healing_fountains",
	"resources",
	"movement_patterns",
	"inventory_items",
}

var zoneKindDirectZoneTables = []string{
	"quests",
	"challenges",
	"scenarios",
	"expositions",
	"monsters",
	"monster_encounters",
	"treasure_chests",
	"healing_fountains",
	"resources",
	"movement_patterns",
}

const inventoryItemZoneKindEvidenceCTE = `
WITH evidence AS (
  SELECT qir.inventory_item_id AS inventory_item_id, q.zone_kind AS zone_kind
  FROM quest_item_rewards qir
  JOIN quests q ON q.id = qir.quest_id
  WHERE COALESCE(q.zone_kind, '') <> ''

  UNION ALL

  SELECT poir.inventory_item_id AS inventory_item_id, poi.zone_kind AS zone_kind
  FROM point_of_interest_item_rewards poir
  JOIN points_of_interest poi ON poi.id = poir.point_of_interest_id
  WHERE COALESCE(poi.zone_kind, '') <> ''

  UNION ALL

  SELECT poc.inventory_item_id AS inventory_item_id, poi.zone_kind AS zone_kind
  FROM point_of_interest_challenges poc
  JOIN points_of_interest poi ON poi.id = poc.point_of_interest_id
  WHERE poc.inventory_item_id > 0 AND COALESCE(poi.zone_kind, '') <> ''

  UNION ALL

  SELECT c.inventory_item_id AS inventory_item_id, c.zone_kind AS zone_kind
  FROM challenges c
  WHERE c.inventory_item_id IS NOT NULL AND COALESCE(c.zone_kind, '') <> ''

  UNION ALL

  SELECT cicr.inventory_item_id AS inventory_item_id, c.zone_kind AS zone_kind
  FROM challenge_item_choice_rewards cicr
  JOIN challenges c ON c.id = cicr.challenge_id
  WHERE COALESCE(c.zone_kind, '') <> ''

  UNION ALL

  SELECT sir.inventory_item_id AS inventory_item_id, s.zone_kind AS zone_kind
  FROM scenario_item_rewards sir
  JOIN scenarios s ON s.id = sir.scenario_id
  WHERE COALESCE(s.zone_kind, '') <> ''

  UNION ALL

  SELECT sicr.inventory_item_id AS inventory_item_id, s.zone_kind AS zone_kind
  FROM scenario_item_choice_rewards sicr
  JOIN scenarios s ON s.id = sicr.scenario_id
  WHERE COALESCE(s.zone_kind, '') <> ''

  UNION ALL

  SELECT soir.inventory_item_id AS inventory_item_id, s.zone_kind AS zone_kind
  FROM scenario_option_item_rewards soir
  JOIN scenario_options so ON so.id = soir.scenario_option_id
  JOIN scenarios s ON s.id = so.scenario_id
  WHERE COALESCE(s.zone_kind, '') <> ''

  UNION ALL

  SELECT soicr.inventory_item_id AS inventory_item_id, s.zone_kind AS zone_kind
  FROM scenario_option_item_choice_rewards soicr
  JOIN scenario_options so ON so.id = soicr.scenario_option_id
  JOIN scenarios s ON s.id = so.scenario_id
  WHERE COALESCE(s.zone_kind, '') <> ''

  UNION ALL

  SELECT eir.inventory_item_id AS inventory_item_id, e.zone_kind AS zone_kind
  FROM exposition_item_rewards eir
  JOIN expositions e ON e.id = eir.exposition_id
  WHERE COALESCE(e.zone_kind, '') <> ''

  UNION ALL

  SELECT mir.inventory_item_id AS inventory_item_id, m.zone_kind AS zone_kind
  FROM monster_item_rewards mir
  JOIN monsters m ON m.id = mir.monster_id
  WHERE COALESCE(m.zone_kind, '') <> ''

  UNION ALL

  SELECT CAST(reward->>'inventoryItemId' AS INTEGER) AS inventory_item_id, me.zone_kind AS zone_kind
  FROM monster_encounters me
  CROSS JOIN LATERAL jsonb_array_elements(COALESCE(me.item_rewards_json, '[]'::jsonb)) reward
  WHERE COALESCE(me.zone_kind, '') <> ''
    AND COALESCE(reward->>'inventoryItemId', '') ~ '^[0-9]+$'

  UNION ALL

  SELECT tci.inventory_item_id AS inventory_item_id, tc.zone_kind AS zone_kind
  FROM treasure_chest_items tci
  JOIN treasure_chests tc ON tc.id = tci.treasure_chest_id
  WHERE COALESCE(tc.zone_kind, '') <> ''

  UNION ALL

  SELECT rgr.required_inventory_item_id AS inventory_item_id, r.zone_kind AS zone_kind
  FROM resource_gather_requirements rgr
  JOIN resources r ON r.resource_type_id = rgr.resource_type_id
  WHERE COALESCE(r.zone_kind, '') <> ''

  UNION ALL

  SELECT ii.id AS inventory_item_id, r.zone_kind AS zone_kind
  FROM inventory_items ii
  JOIN resources r ON r.resource_type_id = ii.resource_type_id
  WHERE COALESCE(ii.zone_kind, '') = '' AND COALESCE(r.zone_kind, '') <> ''

  UNION ALL

  SELECT COALESCE(m.weapon_inventory_item_id, m.dominant_hand_inventory_item_id) AS inventory_item_id, m.zone_kind AS zone_kind
  FROM monsters m
  WHERE COALESCE(m.zone_kind, '') <> ''
    AND COALESCE(m.weapon_inventory_item_id, m.dominant_hand_inventory_item_id) IS NOT NULL

  UNION ALL

  SELECT m.off_hand_inventory_item_id AS inventory_item_id, m.zone_kind AS zone_kind
  FROM monsters m
  WHERE COALESCE(m.zone_kind, '') <> ''
    AND m.off_hand_inventory_item_id IS NOT NULL
),
counts AS (
  SELECT e.inventory_item_id, e.zone_kind, COUNT(*) AS evidence_count
  FROM evidence e
  JOIN inventory_items ii ON ii.id = e.inventory_item_id
  WHERE e.inventory_item_id IS NOT NULL
    AND COALESCE(e.zone_kind, '') <> ''
    AND COALESCE(ii.zone_kind, '') = ''
  GROUP BY inventory_item_id, zone_kind
),
ranked AS (
  SELECT
    inventory_item_id,
    zone_kind,
    evidence_count,
    ROW_NUMBER() OVER (
      PARTITION BY inventory_item_id
      ORDER BY evidence_count DESC, zone_kind ASC
    ) AS rank,
    LEAD(evidence_count) OVER (
      PARTITION BY inventory_item_id
      ORDER BY evidence_count DESC, zone_kind ASC
    ) AS next_evidence_count
  FROM counts
)
`

func (h *zoneKindHandle) ReplaceReferences(ctx context.Context, currentKind string, nextKind string) (int, error) {
	normalizedCurrent := models.NormalizeZoneKind(currentKind)
	if normalizedCurrent == "" {
		return 0, nil
	}
	normalizedNext := models.NormalizeZoneKind(nextKind)

	totalUpdated := 0
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tableName := range zoneKindReferenceTables {
			result := tx.Table(tableName).
				Where("zone_kind = ?", normalizedCurrent).
				Update("zone_kind", normalizedNext)
			if result.Error != nil {
				return result.Error
			}
			totalUpdated += int(result.RowsAffected)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalUpdated, nil
}

func (h *zoneKindHandle) BackfillMissingContentKinds(ctx context.Context) (*jobs.ZoneKindBackfillSummary, error) {
	summary := &jobs.ZoneKindBackfillSummary{
		Results: make([]jobs.ZoneKindBackfillResult, 0, len(zoneKindDirectZoneTables)+2),
	}

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tableName := range zoneKindDirectZoneTables {
			result, err := backfillDirectZoneKindTable(tx, tableName)
			if err != nil {
				return err
			}
			appendZoneKindBackfillResult(summary, result)
		}

		poiResult, err := backfillPointOfInterestZoneKinds(tx)
		if err != nil {
			return err
		}
		appendZoneKindBackfillResult(summary, poiResult)

		inventoryResult, err := backfillInventoryItemZoneKinds(tx)
		if err != nil {
			return err
		}
		appendZoneKindBackfillResult(summary, inventoryResult)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func appendZoneKindBackfillResult(
	summary *jobs.ZoneKindBackfillSummary,
	result jobs.ZoneKindBackfillResult,
) {
	summary.Results = append(summary.Results, result)
	summary.MissingCount += result.MissingCount
	summary.AssignedCount += result.AssignedCount
	summary.AmbiguousCount += result.AmbiguousCount
	summary.SkippedCount += result.SkippedCount
}

func backfillDirectZoneKindTable(
	tx *gorm.DB,
	tableName string,
) (jobs.ZoneKindBackfillResult, error) {
	missingCount, err := countByRawQuery(tx, fmt.Sprintf(
		`SELECT COUNT(*) FROM %s WHERE COALESCE(zone_kind, '') = ''`,
		tableName,
	))
	if err != nil {
		return jobs.ZoneKindBackfillResult{}, err
	}

	result := tx.Exec(fmt.Sprintf(
		`UPDATE %s AS target
		SET zone_kind = zones.kind
		FROM zones
		WHERE target.zone_id = zones.id
			AND COALESCE(target.zone_kind, '') = ''
			AND COALESCE(zones.kind, '') <> ''`,
		tableName,
	))
	if result.Error != nil {
		return jobs.ZoneKindBackfillResult{}, result.Error
	}

	assignedCount := int(result.RowsAffected)
	return jobs.ZoneKindBackfillResult{
		ContentType:   tableName,
		MissingCount:  missingCount,
		AssignedCount: assignedCount,
		SkippedCount:  clampBackfillSkippedCount(missingCount - assignedCount),
	}, nil
}

func backfillPointOfInterestZoneKinds(tx *gorm.DB) (jobs.ZoneKindBackfillResult, error) {
	missingCount, err := countByRawQuery(
		tx,
		`SELECT COUNT(*) FROM points_of_interest WHERE COALESCE(zone_kind, '') = ''`,
	)
	if err != nil {
		return jobs.ZoneKindBackfillResult{}, err
	}

	ambiguousCount, err := countByRawQuery(tx, `
SELECT COUNT(*)
FROM (
  SELECT poi.id
  FROM points_of_interest poi
  JOIN point_of_interest_zones piz ON piz.point_of_interest_id = poi.id
  JOIN zones z ON z.id = piz.zone_id
  WHERE COALESCE(poi.zone_kind, '') = ''
    AND COALESCE(z.kind, '') <> ''
  GROUP BY poi.id
  HAVING COUNT(DISTINCT z.kind) > 1
) AS ambiguous_points
`)
	if err != nil {
		return jobs.ZoneKindBackfillResult{}, err
	}

	updateResult := tx.Exec(`
WITH resolved AS (
  SELECT poi.id, MIN(z.kind) AS zone_kind
  FROM points_of_interest poi
  JOIN point_of_interest_zones piz ON piz.point_of_interest_id = poi.id
  JOIN zones z ON z.id = piz.zone_id
  WHERE COALESCE(poi.zone_kind, '') = ''
    AND COALESCE(z.kind, '') <> ''
  GROUP BY poi.id
  HAVING COUNT(DISTINCT z.kind) = 1
)
UPDATE points_of_interest poi
SET zone_kind = resolved.zone_kind
FROM resolved
WHERE poi.id = resolved.id
`)
	if updateResult.Error != nil {
		return jobs.ZoneKindBackfillResult{}, updateResult.Error
	}

	assignedCount := int(updateResult.RowsAffected)
	return jobs.ZoneKindBackfillResult{
		ContentType:    "points_of_interest",
		MissingCount:   missingCount,
		AssignedCount:  assignedCount,
		AmbiguousCount: ambiguousCount,
		SkippedCount:   clampBackfillSkippedCount(missingCount - assignedCount - ambiguousCount),
	}, nil
}

func backfillInventoryItemZoneKinds(tx *gorm.DB) (jobs.ZoneKindBackfillResult, error) {
	missingCount, err := countByRawQuery(
		tx,
		`SELECT COUNT(*) FROM inventory_items WHERE COALESCE(zone_kind, '') = ''`,
	)
	if err != nil {
		return jobs.ZoneKindBackfillResult{}, err
	}

	ambiguousCount, err := countByRawQuery(tx, inventoryItemZoneKindEvidenceCTE+`
SELECT COUNT(*)
FROM ranked
WHERE rank = 1
  AND next_evidence_count IS NOT NULL
  AND evidence_count = next_evidence_count
`)
	if err != nil {
		return jobs.ZoneKindBackfillResult{}, err
	}

	updateResult := tx.Exec(inventoryItemZoneKindEvidenceCTE + `
, assignments AS (
  SELECT inventory_item_id, zone_kind
  FROM ranked
  WHERE rank = 1
    AND (next_evidence_count IS NULL OR evidence_count > next_evidence_count)
)
UPDATE inventory_items ii
SET zone_kind = assignments.zone_kind
FROM assignments
WHERE ii.id = assignments.inventory_item_id
  AND COALESCE(ii.zone_kind, '') = ''
`)
	if updateResult.Error != nil {
		return jobs.ZoneKindBackfillResult{}, updateResult.Error
	}

	assignedCount := int(updateResult.RowsAffected)
	return jobs.ZoneKindBackfillResult{
		ContentType:    "inventory_items",
		MissingCount:   missingCount,
		AssignedCount:  assignedCount,
		AmbiguousCount: ambiguousCount,
		SkippedCount:   clampBackfillSkippedCount(missingCount - assignedCount - ambiguousCount),
	}, nil
}

func countByRawQuery(tx *gorm.DB, query string) (int, error) {
	var count int64
	if err := tx.Raw(query).Scan(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func clampBackfillSkippedCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
