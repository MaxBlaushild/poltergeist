package models

type ZoneContentFlushSummary struct {
	ZoneCount                          int `json:"zoneCount"`
	DeletedPointOfInterestCount        int `json:"deletedPointOfInterestCount"`
	DetachedSharedPointOfInterestCount int `json:"detachedSharedPointOfInterestCount"`
	DeletedCharacterCount              int `json:"deletedCharacterCount"`
	DeletedQuestCount                  int `json:"deletedQuestCount"`
	DeletedZoneQuestArchetypeCount     int `json:"deletedZoneQuestArchetypeCount"`
	DeletedChallengeCount              int `json:"deletedChallengeCount"`
	DeletedScenarioCount               int `json:"deletedScenarioCount"`
	DeletedExpositionCount             int `json:"deletedExpositionCount"`
	DeletedMonsterEncounterCount       int `json:"deletedMonsterEncounterCount"`
	DeletedMonsterCount                int `json:"deletedMonsterCount"`
	DeletedTreasureChestCount          int `json:"deletedTreasureChestCount"`
	DeletedHealingFountainCount        int `json:"deletedHealingFountainCount"`
	DeletedResourceCount               int `json:"deletedResourceCount"`
	DeletedMovementPatternCount        int `json:"deletedMovementPatternCount"`
	DeletedZoneSeedJobCount            int `json:"deletedZoneSeedJobCount"`
	DeletedQuestGenerationJobCount     int `json:"deletedQuestGenerationJobCount"`
	DeletedScenarioGenerationJobCount  int `json:"deletedScenarioGenerationJobCount"`
	DeletedChallengeGenerationJobCount int `json:"deletedChallengeGenerationJobCount"`
}
