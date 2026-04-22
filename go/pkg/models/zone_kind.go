package models

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultZoneKindRatio = 1.0

const (
	ZoneKindPatternTileGenerationStatusNone       = "none"
	ZoneKindPatternTileGenerationStatusQueued     = "queued"
	ZoneKindPatternTileGenerationStatusInProgress = "in_progress"
	ZoneKindPatternTileGenerationStatusComplete   = "complete"
	ZoneKindPatternTileGenerationStatusFailed     = "failed"
)

type ZoneKind struct {
	ID                          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                   time.Time `json:"createdAt"`
	UpdatedAt                   time.Time `json:"updatedAt"`
	Slug                        string    `json:"slug" gorm:"column:slug"`
	Name                        string    `json:"name"`
	Description                 string    `json:"description"`
	OverlayColor                string    `json:"overlayColor" gorm:"column:overlay_color"`
	PatternTileURL              string    `json:"patternTileUrl" gorm:"column:pattern_tile_url"`
	PatternTilePrompt           string    `json:"patternTilePrompt" gorm:"column:pattern_tile_prompt"`
	PatternTileGenerationStatus string    `json:"patternTileGenerationStatus" gorm:"column:pattern_tile_generation_status"`
	PatternTileGenerationError  string    `json:"patternTileGenerationError" gorm:"column:pattern_tile_generation_error"`
	PlaceCountRatio             float64   `json:"placeCountRatio" gorm:"column:place_count_ratio"`
	MonsterCountRatio           float64   `json:"monsterCountRatio" gorm:"column:monster_count_ratio"`
	BossEncounterCountRatio     float64   `json:"bossEncounterCountRatio" gorm:"column:boss_encounter_count_ratio"`
	RaidEncounterCountRatio     float64   `json:"raidEncounterCountRatio" gorm:"column:raid_encounter_count_ratio"`
	InputEncounterCountRatio    float64   `json:"inputEncounterCountRatio" gorm:"column:input_encounter_count_ratio"`
	OptionEncounterCountRatio   float64   `json:"optionEncounterCountRatio" gorm:"column:option_encounter_count_ratio"`
	TreasureChestCountRatio     float64   `json:"treasureChestCountRatio" gorm:"column:treasure_chest_count_ratio"`
	HealingFountainCountRatio   float64   `json:"healingFountainCountRatio" gorm:"column:healing_fountain_count_ratio"`
	HerbalismResourceCountRatio float64   `json:"herbalismResourceCountRatio" gorm:"column:herbalism_resource_count_ratio"`
	MiningResourceCountRatio    float64   `json:"miningResourceCountRatio" gorm:"column:mining_resource_count_ratio"`
	ResourceCountRatio          float64   `json:"resourceCountRatio" gorm:"column:resource_count_ratio"`
}

func (ZoneKind) TableName() string {
	return "zone_kinds"
}

func normalizeZoneKindRatio(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return defaultZoneKindRatio
	}
	if value < 0 {
		return 0
	}
	return value
}

func NormalizeHexColor(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "#") {
		trimmed = trimmed[1:]
	}
	switch len(trimmed) {
	case 3:
		var expanded strings.Builder
		expanded.Grow(6)
		for _, char := range trimmed {
			if !isHexColorRune(char) {
				return ""
			}
			expanded.WriteRune(char)
			expanded.WriteRune(char)
		}
		trimmed = expanded.String()
	case 6:
		for _, char := range trimmed {
			if !isHexColorRune(char) {
				return ""
			}
		}
	default:
		return ""
	}
	return "#" + trimmed
}

func isHexColorRune(char rune) bool {
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')
}

func normalizeZoneKindPatternTileGenerationStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case ZoneKindPatternTileGenerationStatusQueued:
		return ZoneKindPatternTileGenerationStatusQueued
	case ZoneKindPatternTileGenerationStatusInProgress:
		return ZoneKindPatternTileGenerationStatusInProgress
	case ZoneKindPatternTileGenerationStatusComplete:
		return ZoneKindPatternTileGenerationStatusComplete
	case ZoneKindPatternTileGenerationStatusFailed:
		return ZoneKindPatternTileGenerationStatusFailed
	default:
		return ZoneKindPatternTileGenerationStatusNone
	}
}

func (z *ZoneKind) BeforeSave(tx *gorm.DB) error {
	z.Name = strings.TrimSpace(z.Name)
	z.Description = strings.TrimSpace(z.Description)
	z.Slug = NormalizeZoneKind(z.Slug)
	if z.Slug == "" {
		z.Slug = NormalizeZoneKind(z.Name)
	}
	z.OverlayColor = NormalizeHexColor(z.OverlayColor)
	z.PatternTileURL = strings.TrimSpace(z.PatternTileURL)
	z.PatternTilePrompt = strings.TrimSpace(z.PatternTilePrompt)
	z.PatternTileGenerationError = strings.TrimSpace(z.PatternTileGenerationError)
	z.PatternTileGenerationStatus = normalizeZoneKindPatternTileGenerationStatus(z.PatternTileGenerationStatus)
	z.PlaceCountRatio = normalizeZoneKindRatio(z.PlaceCountRatio)
	z.MonsterCountRatio = normalizeZoneKindRatio(z.MonsterCountRatio)
	z.BossEncounterCountRatio = normalizeZoneKindRatio(z.BossEncounterCountRatio)
	z.RaidEncounterCountRatio = normalizeZoneKindRatio(z.RaidEncounterCountRatio)
	z.InputEncounterCountRatio = normalizeZoneKindRatio(z.InputEncounterCountRatio)
	z.OptionEncounterCountRatio = normalizeZoneKindRatio(z.OptionEncounterCountRatio)
	z.TreasureChestCountRatio = normalizeZoneKindRatio(z.TreasureChestCountRatio)
	z.HealingFountainCountRatio = normalizeZoneKindRatio(z.HealingFountainCountRatio)
	z.HerbalismResourceCountRatio = normalizeZoneKindRatio(z.HerbalismResourceCountRatio)
	z.MiningResourceCountRatio = normalizeZoneKindRatio(z.MiningResourceCountRatio)
	z.ResourceCountRatio = normalizeZoneKindRatio(z.ResourceCountRatio)
	return nil
}

func (z ZoneKind) ApplyToCounts(counts ZoneSeedResolvedCounts) ZoneSeedResolvedCounts {
	apply := func(value int, ratio float64) int {
		if value == 0 {
			return 0
		}
		return int(math.Round(float64(value) * normalizeZoneKindRatio(ratio)))
	}

	return ZoneSeedResolvedCounts{
		PlaceCount:             apply(counts.PlaceCount, z.PlaceCountRatio),
		MonsterCount:           apply(counts.MonsterCount, z.MonsterCountRatio),
		BossEncounterCount:     apply(counts.BossEncounterCount, z.BossEncounterCountRatio),
		RaidEncounterCount:     apply(counts.RaidEncounterCount, z.RaidEncounterCountRatio),
		InputEncounterCount:    apply(counts.InputEncounterCount, z.InputEncounterCountRatio),
		OptionEncounterCount:   apply(counts.OptionEncounterCount, z.OptionEncounterCountRatio),
		TreasureChestCount:     apply(counts.TreasureChestCount, z.TreasureChestCountRatio),
		HealingFountainCount:   apply(counts.HealingFountainCount, z.HealingFountainCountRatio),
		HerbalismResourceCount: apply(counts.HerbalismResourceCount, z.HerbalismResourceCountRatio),
		MiningResourceCount:    apply(counts.MiningResourceCount, z.MiningResourceCountRatio),
	}.WithLegacyResourceCount()
}
