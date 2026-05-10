package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PointOfInterestExpositionSeedProfile struct {
	Category                     PointOfInterestMarkerCategory `json:"category"`
	FirstSpawnChanceBasisPoints  int                           `json:"firstSpawnChanceBasisPoints"`
	SecondSpawnChanceBasisPoints int                           `json:"secondSpawnChanceBasisPoints"`
}

type PointOfInterestExpositionSeedConfig struct {
	ID           int                                    `gorm:"primaryKey" json:"id"`
	ProfilesJSON datatypes.JSON                         `gorm:"column:profiles_json;type:jsonb;default:'[]'" json:"-"`
	Profiles     []PointOfInterestExpositionSeedProfile `gorm:"-" json:"profiles"`
	CreatedAt    time.Time                              `json:"createdAt"`
	UpdatedAt    time.Time                              `json:"updatedAt"`
}

func (PointOfInterestExpositionSeedConfig) TableName() string {
	return "point_of_interest_exposition_seed_configs"
}

func (c *PointOfInterestExpositionSeedConfig) BeforeSave(tx *gorm.DB) (err error) {
	c.Profiles = ResolvePointOfInterestExpositionSeedProfiles(c.Profiles)
	raw, err := json.Marshal(c.Profiles)
	if err != nil {
		return err
	}
	c.ProfilesJSON = datatypes.JSON(raw)
	return nil
}

func (c *PointOfInterestExpositionSeedConfig) AfterFind(tx *gorm.DB) (err error) {
	if len(c.ProfilesJSON) == 0 {
		c.Profiles = ResolvePointOfInterestExpositionSeedProfiles(nil)
		return nil
	}

	var profiles []PointOfInterestExpositionSeedProfile
	if err := json.Unmarshal(c.ProfilesJSON, &profiles); err != nil {
		return err
	}
	c.Profiles = ResolvePointOfInterestExpositionSeedProfiles(profiles)
	return nil
}

func DefaultPointOfInterestExpositionSeedProfiles() []PointOfInterestExpositionSeedProfile {
	return []PointOfInterestExpositionSeedProfile{
		{
			Category:                    PointOfInterestMarkerCategoryGeneric,
			FirstSpawnChanceBasisPoints: 1400,
		},
		{
			Category:                     PointOfInterestMarkerCategoryCoffeehouse,
			FirstSpawnChanceBasisPoints:  6600,
			SecondSpawnChanceBasisPoints: 1800,
		},
		{
			Category:                     PointOfInterestMarkerCategoryTavern,
			FirstSpawnChanceBasisPoints:  7400,
			SecondSpawnChanceBasisPoints: 2600,
		},
		{
			Category:                     PointOfInterestMarkerCategoryEatery,
			FirstSpawnChanceBasisPoints:  5800,
			SecondSpawnChanceBasisPoints: 1500,
		},
		{
			Category:                     PointOfInterestMarkerCategoryMarket,
			FirstSpawnChanceBasisPoints:  8200,
			SecondSpawnChanceBasisPoints: 3200,
		},
		{
			Category:                     PointOfInterestMarkerCategoryArchive,
			FirstSpawnChanceBasisPoints:  7000,
			SecondSpawnChanceBasisPoints: 2400,
		},
		{
			Category:                     PointOfInterestMarkerCategoryPark,
			FirstSpawnChanceBasisPoints:  3000,
			SecondSpawnChanceBasisPoints: 900,
		},
		{
			Category:                     PointOfInterestMarkerCategoryWaterfront,
			FirstSpawnChanceBasisPoints:  5200,
			SecondSpawnChanceBasisPoints: 1800,
		},
		{
			Category:                     PointOfInterestMarkerCategoryMuseum,
			FirstSpawnChanceBasisPoints:  6200,
			SecondSpawnChanceBasisPoints: 2100,
		},
		{
			Category:                     PointOfInterestMarkerCategoryTheater,
			FirstSpawnChanceBasisPoints:  6600,
			SecondSpawnChanceBasisPoints: 2200,
		},
		{
			Category:                     PointOfInterestMarkerCategoryLandmark,
			FirstSpawnChanceBasisPoints:  2600,
			SecondSpawnChanceBasisPoints: 700,
		},
		{
			Category:                     PointOfInterestMarkerCategoryCivic,
			FirstSpawnChanceBasisPoints:  2200,
			SecondSpawnChanceBasisPoints: 500,
		},
		{
			Category:                     PointOfInterestMarkerCategoryArena,
			FirstSpawnChanceBasisPoints:  4800,
			SecondSpawnChanceBasisPoints: 1500,
		},
	}
}

func PointOfInterestExpositionSeedProfileForCategory(
	profiles []PointOfInterestExpositionSeedProfile,
	category PointOfInterestMarkerCategory,
) PointOfInterestExpositionSeedProfile {
	resolved := ResolvePointOfInterestExpositionSeedProfiles(profiles)
	if normalized, ok := ParsePointOfInterestMarkerCategory(string(category)); ok {
		category = normalized
	}
	for _, profile := range resolved {
		if profile.Category == category {
			return profile
		}
	}
	return PointOfInterestExpositionSeedProfile{Category: category}
}

func ResolvePointOfInterestExpositionSeedProfiles(
	profiles []PointOfInterestExpositionSeedProfile,
) []PointOfInterestExpositionSeedProfile {
	defaults := DefaultPointOfInterestExpositionSeedProfiles()
	byCategory := make(map[PointOfInterestMarkerCategory]PointOfInterestExpositionSeedProfile, len(defaults))
	for _, profile := range defaults {
		byCategory[profile.Category] = profile
	}
	for _, rawProfile := range profiles {
		profile, ok := normalizePointOfInterestExpositionSeedProfile(rawProfile)
		if !ok {
			continue
		}
		byCategory[profile.Category] = profile
	}

	resolved := make([]PointOfInterestExpositionSeedProfile, 0, len(AllPointOfInterestMarkerCategories()))
	for _, category := range AllPointOfInterestMarkerCategories() {
		profile, ok := byCategory[category]
		if !ok {
			profile = PointOfInterestExpositionSeedProfile{Category: category}
		}
		resolved = append(resolved, profile)
	}
	return resolved
}

func normalizePointOfInterestExpositionSeedProfile(
	profile PointOfInterestExpositionSeedProfile,
) (PointOfInterestExpositionSeedProfile, bool) {
	category, ok := ParsePointOfInterestMarkerCategory(string(profile.Category))
	if !ok {
		return PointOfInterestExpositionSeedProfile{}, false
	}
	return PointOfInterestExpositionSeedProfile{
		Category:                     category,
		FirstSpawnChanceBasisPoints:  clampPointOfInterestShopkeeperSpawnChance(profile.FirstSpawnChanceBasisPoints),
		SecondSpawnChanceBasisPoints: clampPointOfInterestShopkeeperSpawnChance(profile.SecondSpawnChanceBasisPoints),
	}, true
}
