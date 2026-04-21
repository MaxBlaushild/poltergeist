package models

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultZoneKindRatio = 1.0

type ZoneKind struct {
	ID                        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
	Slug                      string    `json:"slug" gorm:"column:slug"`
	Name                      string    `json:"name"`
	Description               string    `json:"description"`
	PlaceCountRatio           float64   `json:"placeCountRatio" gorm:"column:place_count_ratio"`
	MonsterCountRatio         float64   `json:"monsterCountRatio" gorm:"column:monster_count_ratio"`
	BossEncounterCountRatio   float64   `json:"bossEncounterCountRatio" gorm:"column:boss_encounter_count_ratio"`
	RaidEncounterCountRatio   float64   `json:"raidEncounterCountRatio" gorm:"column:raid_encounter_count_ratio"`
	InputEncounterCountRatio  float64   `json:"inputEncounterCountRatio" gorm:"column:input_encounter_count_ratio"`
	OptionEncounterCountRatio float64   `json:"optionEncounterCountRatio" gorm:"column:option_encounter_count_ratio"`
	TreasureChestCountRatio   float64   `json:"treasureChestCountRatio" gorm:"column:treasure_chest_count_ratio"`
	HealingFountainCountRatio float64   `json:"healingFountainCountRatio" gorm:"column:healing_fountain_count_ratio"`
	ResourceCountRatio        float64   `json:"resourceCountRatio" gorm:"column:resource_count_ratio"`
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

func (z *ZoneKind) BeforeSave(tx *gorm.DB) error {
	z.Name = strings.TrimSpace(z.Name)
	z.Description = strings.TrimSpace(z.Description)
	z.Slug = NormalizeZoneKind(z.Slug)
	if z.Slug == "" {
		z.Slug = NormalizeZoneKind(z.Name)
	}
	z.PlaceCountRatio = normalizeZoneKindRatio(z.PlaceCountRatio)
	z.MonsterCountRatio = normalizeZoneKindRatio(z.MonsterCountRatio)
	z.BossEncounterCountRatio = normalizeZoneKindRatio(z.BossEncounterCountRatio)
	z.RaidEncounterCountRatio = normalizeZoneKindRatio(z.RaidEncounterCountRatio)
	z.InputEncounterCountRatio = normalizeZoneKindRatio(z.InputEncounterCountRatio)
	z.OptionEncounterCountRatio = normalizeZoneKindRatio(z.OptionEncounterCountRatio)
	z.TreasureChestCountRatio = normalizeZoneKindRatio(z.TreasureChestCountRatio)
	z.HealingFountainCountRatio = normalizeZoneKindRatio(z.HealingFountainCountRatio)
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
		PlaceCount:           apply(counts.PlaceCount, z.PlaceCountRatio),
		MonsterCount:         apply(counts.MonsterCount, z.MonsterCountRatio),
		BossEncounterCount:   apply(counts.BossEncounterCount, z.BossEncounterCountRatio),
		RaidEncounterCount:   apply(counts.RaidEncounterCount, z.RaidEncounterCountRatio),
		InputEncounterCount:  apply(counts.InputEncounterCount, z.InputEncounterCountRatio),
		OptionEncounterCount: apply(counts.OptionEncounterCount, z.OptionEncounterCountRatio),
		TreasureChestCount:   apply(counts.TreasureChestCount, z.TreasureChestCountRatio),
		HealingFountainCount: apply(counts.HealingFountainCount, z.HealingFountainCountRatio),
		ResourceCount:        apply(counts.ResourceCount, z.ResourceCountRatio),
	}
}
