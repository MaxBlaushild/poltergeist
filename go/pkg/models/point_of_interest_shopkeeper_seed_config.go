package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PointOfInterestShopkeeperSeedCandidate struct {
	Tag    string `json:"tag"`
	Weight int    `json:"weight"`
}

type PointOfInterestShopkeeperSeedProfile struct {
	Category               PointOfInterestMarkerCategory            `json:"category"`
	SpawnChanceBasisPoints int                                      `json:"spawnChanceBasisPoints"`
	Candidates             []PointOfInterestShopkeeperSeedCandidate `json:"candidates"`
}

type PointOfInterestShopkeeperSeedConfig struct {
	ID           int                                    `gorm:"primaryKey" json:"id"`
	ProfilesJSON datatypes.JSON                         `gorm:"column:profiles_json;type:jsonb;default:'[]'" json:"-"`
	Profiles     []PointOfInterestShopkeeperSeedProfile `gorm:"-" json:"profiles"`
	CreatedAt    time.Time                              `json:"createdAt"`
	UpdatedAt    time.Time                              `json:"updatedAt"`
}

func (PointOfInterestShopkeeperSeedConfig) TableName() string {
	return "point_of_interest_shopkeeper_seed_configs"
}

func (c *PointOfInterestShopkeeperSeedConfig) BeforeSave(tx *gorm.DB) (err error) {
	c.Profiles = ResolvePointOfInterestShopkeeperSeedProfiles(c.Profiles)
	raw, err := json.Marshal(c.Profiles)
	if err != nil {
		return err
	}
	c.ProfilesJSON = datatypes.JSON(raw)
	return nil
}

func (c *PointOfInterestShopkeeperSeedConfig) AfterFind(tx *gorm.DB) (err error) {
	if len(c.ProfilesJSON) == 0 {
		c.Profiles = ResolvePointOfInterestShopkeeperSeedProfiles(nil)
		return nil
	}

	var profiles []PointOfInterestShopkeeperSeedProfile
	if err := json.Unmarshal(c.ProfilesJSON, &profiles); err != nil {
		return err
	}
	c.Profiles = ResolvePointOfInterestShopkeeperSeedProfiles(profiles)
	return nil
}

func DefaultPointOfInterestShopkeeperSeedProfiles() []PointOfInterestShopkeeperSeedProfile {
	return []PointOfInterestShopkeeperSeedProfile{
		{
			Category:               PointOfInterestMarkerCategoryGeneric,
			SpawnChanceBasisPoints: 0,
			Candidates:             []PointOfInterestShopkeeperSeedCandidate{},
		},
		{
			Category:               PointOfInterestMarkerCategoryCoffeehouse,
			SpawnChanceBasisPoints: 8200,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "social", Weight: 6},
				{Tag: "guide", Weight: 4},
				{Tag: "arcane", Weight: 2},
				{Tag: "potion", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryTavern,
			SpawnChanceBasisPoints: 8600,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "martial", Weight: 5},
				{Tag: "social", Weight: 4},
				{Tag: "potion", Weight: 3},
				{Tag: "hunter", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryEatery,
			SpawnChanceBasisPoints: 7800,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "social", Weight: 5},
				{Tag: "nature", Weight: 3},
				{Tag: "guide", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryMarket,
			SpawnChanceBasisPoints: 9300,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "social", Weight: 5},
				{Tag: "guide", Weight: 5},
				{Tag: "martial", Weight: 4},
				{Tag: "arcane", Weight: 3},
				{Tag: "relic", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryArchive,
			SpawnChanceBasisPoints: 7600,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "arcane", Weight: 6},
				{Tag: "guide", Weight: 5},
				{Tag: "relic", Weight: 3},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryPark,
			SpawnChanceBasisPoints: 3400,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "nature", Weight: 6},
				{Tag: "wild", Weight: 5},
				{Tag: "guide", Weight: 3},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryWaterfront,
			SpawnChanceBasisPoints: 4800,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "sea", Weight: 6},
				{Tag: "guide", Weight: 4},
				{Tag: "nature", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryMuseum,
			SpawnChanceBasisPoints: 6400,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "relic", Weight: 6},
				{Tag: "arcane", Weight: 5},
				{Tag: "guide", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryTheater,
			SpawnChanceBasisPoints: 7000,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "social", Weight: 6},
				{Tag: "court", Weight: 4},
				{Tag: "arcane", Weight: 2},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryLandmark,
			SpawnChanceBasisPoints: 2800,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "guide", Weight: 5},
				{Tag: "relic", Weight: 4},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryCivic,
			SpawnChanceBasisPoints: 2200,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "guide", Weight: 4},
				{Tag: "holy", Weight: 3},
				{Tag: "relic", Weight: 1},
			},
		},
		{
			Category:               PointOfInterestMarkerCategoryArena,
			SpawnChanceBasisPoints: 6600,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "martial", Weight: 6},
				{Tag: "hunter", Weight: 3},
				{Tag: "potion", Weight: 2},
			},
		},
	}
}

func PointOfInterestShopkeeperSeedProfileForCategory(
	profiles []PointOfInterestShopkeeperSeedProfile,
	category PointOfInterestMarkerCategory,
) PointOfInterestShopkeeperSeedProfile {
	resolved := ResolvePointOfInterestShopkeeperSeedProfiles(profiles)
	if normalized, ok := ParsePointOfInterestMarkerCategory(string(category)); ok {
		category = normalized
	}
	for _, profile := range resolved {
		if profile.Category == category {
			return clonePointOfInterestShopkeeperSeedProfile(profile)
		}
	}
	return PointOfInterestShopkeeperSeedProfile{
		Category:   category,
		Candidates: []PointOfInterestShopkeeperSeedCandidate{},
	}
}

func ResolvePointOfInterestShopkeeperSeedProfiles(
	profiles []PointOfInterestShopkeeperSeedProfile,
) []PointOfInterestShopkeeperSeedProfile {
	defaults := DefaultPointOfInterestShopkeeperSeedProfiles()
	byCategory := make(map[PointOfInterestMarkerCategory]PointOfInterestShopkeeperSeedProfile, len(defaults))
	for _, profile := range defaults {
		byCategory[profile.Category] = clonePointOfInterestShopkeeperSeedProfile(profile)
	}
	for _, rawProfile := range profiles {
		profile, ok := normalizePointOfInterestShopkeeperSeedProfile(rawProfile)
		if !ok {
			continue
		}
		byCategory[profile.Category] = profile
	}

	resolved := make([]PointOfInterestShopkeeperSeedProfile, 0, len(AllPointOfInterestMarkerCategories()))
	for _, category := range AllPointOfInterestMarkerCategories() {
		profile, ok := byCategory[category]
		if !ok {
			profile = PointOfInterestShopkeeperSeedProfile{
				Category:   category,
				Candidates: []PointOfInterestShopkeeperSeedCandidate{},
			}
		}
		resolved = append(resolved, clonePointOfInterestShopkeeperSeedProfile(profile))
	}
	return resolved
}

func normalizePointOfInterestShopkeeperSeedProfile(
	profile PointOfInterestShopkeeperSeedProfile,
) (PointOfInterestShopkeeperSeedProfile, bool) {
	category, ok := ParsePointOfInterestMarkerCategory(string(profile.Category))
	if !ok {
		return PointOfInterestShopkeeperSeedProfile{}, false
	}
	return PointOfInterestShopkeeperSeedProfile{
		Category:               category,
		SpawnChanceBasisPoints: clampPointOfInterestShopkeeperSpawnChance(profile.SpawnChanceBasisPoints),
		Candidates:             normalizePointOfInterestShopkeeperSeedCandidates(profile.Candidates),
	}, true
}

func normalizePointOfInterestShopkeeperSeedCandidates(
	candidates []PointOfInterestShopkeeperSeedCandidate,
) []PointOfInterestShopkeeperSeedCandidate {
	if len(candidates) == 0 {
		return []PointOfInterestShopkeeperSeedCandidate{}
	}

	normalized := make([]PointOfInterestShopkeeperSeedCandidate, 0, len(candidates))
	indexByTag := map[string]int{}
	for _, candidate := range candidates {
		tags := NormalizeTagList([]string{candidate.Tag})
		if len(tags) == 0 || candidate.Weight <= 0 {
			continue
		}
		tag := tags[0]
		if index, ok := indexByTag[tag]; ok {
			normalized[index].Weight += candidate.Weight
			continue
		}
		indexByTag[tag] = len(normalized)
		normalized = append(normalized, PointOfInterestShopkeeperSeedCandidate{
			Tag:    tag,
			Weight: candidate.Weight,
		})
	}

	return normalized
}

func clampPointOfInterestShopkeeperSpawnChance(value int) int {
	if value < 0 {
		return 0
	}
	if value > 10_000 {
		return 10_000
	}
	return value
}

func clonePointOfInterestShopkeeperSeedProfile(
	profile PointOfInterestShopkeeperSeedProfile,
) PointOfInterestShopkeeperSeedProfile {
	clonedCandidates := make([]PointOfInterestShopkeeperSeedCandidate, 0, len(profile.Candidates))
	clonedCandidates = append(clonedCandidates, profile.Candidates...)
	return PointOfInterestShopkeeperSeedProfile{
		Category:               profile.Category,
		SpawnChanceBasisPoints: profile.SpawnChanceBasisPoints,
		Candidates:             clonedCandidates,
	}
}
