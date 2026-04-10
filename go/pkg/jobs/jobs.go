package jobs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	GenerateQuestForZoneTaskType                   = "generate_quest_for_zone"
	QueueQuestGenerationsTaskType                  = "queue_quest_generations"
	CreateProfilePictureTaskType                   = "create_profile_picture"
	GenerateOutfitProfilePictureTaskType           = "generate_outfit_profile_picture"
	GenerateInventoryItemImageTaskType             = "generate_inventory_item_image"
	GenerateSpellIconTaskType                      = "generate_spell_icon"
	GenerateSpellsBulkTaskType                     = "generate_spells_bulk"
	GenerateSpellProgressionFromPromptTaskType     = "generate_spell_progression_from_prompt"
	RebalanceSpellDamageTaskType                   = "rebalance_spell_damage"
	GenerateMonsterImageTaskType                   = "generate_monster_image"
	GenerateMonsterTemplateImageTaskType           = "generate_monster_template_image"
	GenerateMonsterTemplatesBulkTaskType           = "generate_monster_templates_bulk"
	RefreshMonsterTemplateAffinitiesTaskType       = "refresh_monster_template_affinities"
	ResetMonsterTemplateProgressionsTaskType       = "reset_monster_template_progressions"
	GenerateCharacterImageTaskType                 = "generate_character_image"
	GeneratePointOfInterestImageTaskType           = "generate_point_of_interest_image"
	GenerateScenarioImageTaskType                  = "generate_scenario_image"
	GenerateExpositionImageTaskType                = "generate_exposition_image"
	GenerateTutorialImageTaskType                  = "generate_tutorial_image"
	InstantiateTutorialBaseQuestTaskType           = "instantiate_tutorial_base_quest"
	GenerateChallengeImageTaskType                 = "generate_challenge_image"
	GenerateChallengeTemplateImageTaskType         = "generate_challenge_template_image"
	GenerateInventoryItemSuggestionsTaskType       = "generate_inventory_item_suggestions"
	GenerateScenarioTaskType                       = "generate_scenario"
	GenerateChallengesTaskType                     = "generate_challenges"
	GenerateScenarioTemplatesTaskType              = "generate_scenario_templates"
	GenerateChallengeTemplatesTaskType             = "generate_challenge_templates"
	GenerateLocationArchetypesTaskType             = "generate_location_archetypes"
	GenerateQuestArchetypeSuggestionsTaskType      = "generate_quest_archetype_suggestions"
	GenerateMainStorySuggestionsTaskType           = "generate_main_story_suggestions"
	ProcessMainStoryDistrictRunTaskType            = "process_main_story_district_run"
	GenerateZoneFlavorTaskType                     = "generate_zone_flavor"
	GenerateZoneTagsTaskType                       = "generate_zone_tags"
	GenerateBaseDescriptionTaskType                = "generate_base_description"
	GenerateBaseStructureLevelImageTaskType        = "generate_base_structure_level_image"
	GenerateBaseStructureLevelTopDownImageTaskType = "generate_base_structure_level_top_down_image"
	GenerateImageThumbnailTaskType                 = "generate_image_thumbnail"
	QueueThumbnailBackfillTaskType                 = "queue_thumbnail_backfill"
	SeedTreasureChestsTaskType                     = "seed_treasure_chests"
	CalculateTrendingDestinationsTaskType          = "calculate_trending_destinations"
	ProcessRecurringQuestsTaskType                 = "process_recurring_quests"
	ProcessRecurringStandaloneContentTaskType      = "process_recurring_standalone_content"
	CleanupOrphanedQuestActionsTaskType            = "cleanup_orphaned_quest_actions"
	CheckBlockchainTransactionsTaskType            = "check_blockchain_transactions"
	ImportPointOfInterestTaskType                  = "import_point_of_interest"
	ImportZonesForMetroTaskType                    = "import_zones_for_metro"
	MonitorPolymarketTradesTaskType                = "monitor_polymarket_trades"
	SeedZoneDraftTaskType                          = "seed_zone_draft"
	SeedDistrictTaskType                           = "seed_district"
	ApplyZoneSeedDraftTaskType                     = "apply_zone_seed_draft"
	ShuffleZoneSeedChallengeTaskType               = "shuffle_zone_seed_challenge"
)

const (
	MonsterTemplateBulkStatusQueued     = "queued"
	MonsterTemplateBulkStatusInProgress = "in_progress"
	MonsterTemplateBulkStatusCompleted  = "completed"
	MonsterTemplateBulkStatusFailed     = "failed"

	MonsterTemplateBulkStatusTTL = 24 * time.Hour
)

const (
	MonsterTemplateAffinityRefreshStatusQueued     = "queued"
	MonsterTemplateAffinityRefreshStatusInProgress = "in_progress"
	MonsterTemplateAffinityRefreshStatusCompleted  = "completed"
	MonsterTemplateAffinityRefreshStatusFailed     = "failed"

	MonsterTemplateAffinityRefreshStatusTTL = 24 * time.Hour
)

const (
	MonsterTemplateProgressionResetStatusQueued     = "queued"
	MonsterTemplateProgressionResetStatusInProgress = "in_progress"
	MonsterTemplateProgressionResetStatusCompleted  = "completed"
	MonsterTemplateProgressionResetStatusFailed     = "failed"

	MonsterTemplateProgressionResetStatusTTL = 24 * time.Hour
)

const (
	SpellBulkStatusQueued     = "queued"
	SpellBulkStatusInProgress = "in_progress"
	SpellBulkStatusCompleted  = "completed"
	SpellBulkStatusFailed     = "failed"

	SpellBulkStatusTTL = 24 * time.Hour
)

const (
	SpellProgressionPromptStatusQueued     = "queued"
	SpellProgressionPromptStatusInProgress = "in_progress"
	SpellProgressionPromptStatusCompleted  = "completed"
	SpellProgressionPromptStatusFailed     = "failed"

	SpellProgressionPromptStatusTTL = 24 * time.Hour
)

const (
	SpellDamageRebalanceStatusQueued     = "queued"
	SpellDamageRebalanceStatusInProgress = "in_progress"
	SpellDamageRebalanceStatusCompleted  = "completed"
	SpellDamageRebalanceStatusFailed     = "failed"

	SpellDamageRebalanceStatusTTL = 24 * time.Hour
)

const (
	ThumbnailEntityCharacter                 = "character"
	ThumbnailEntityPointOfInterest           = "point_of_interest"
	ThumbnailEntityBase                      = "base"
	ThumbnailEntityBaseStructureLevel        = "base_structure_level"
	ThumbnailEntityBaseStructureLevelTopDown = "base_structure_level_top_down"
	ThumbnailEntityStatic                    = "static"
	ThumbnailBucket                          = "crew-profile-icons"
)

type GenerateQuestForZoneTaskPayload struct {
	ZoneID                uuid.UUID  `json:"zone_id"`
	QuestArchetypeID      uuid.UUID  `json:"quest_archetype_id"`
	QuestGiverCharacterID *uuid.UUID `json:"quest_giver_character_id,omitempty"`
	QuestGenerationJobID  *uuid.UUID `json:"quest_generation_job_id,omitempty"`
}

type InstantiateTutorialBaseQuestTaskPayload struct {
	UserID                    uuid.UUID `json:"user_id"`
	BaseLatitude              float64   `json:"base_latitude"`
	BaseLongitude             float64   `json:"base_longitude"`
	BaseQuestArchetypeID      uuid.UUID `json:"base_quest_archetype_id"`
	BaseQuestGiverCharacterID uuid.UUID `json:"base_quest_giver_character_id"`
}

type GenerateQuestArchetypeSuggestionsTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateMainStorySuggestionsTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type ProcessMainStoryDistrictRunTaskPayload struct {
	RunID uuid.UUID `json:"runId"`
}

type GenerateInventoryItemSuggestionsTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateLocationArchetypesTaskPayload struct {
	Count int    `json:"count"`
	Salt  string `json:"salt,omitempty"`
}

type GenerateZoneTagsTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type SeedDistrictTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateBaseDescriptionTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateBaseStructureLevelImageTaskPayload struct {
	StructureDefinitionID uuid.UUID `json:"structureDefinitionId"`
	Level                 int       `json:"level"`
	View                  string    `json:"view,omitempty"`
}

type CreateProfilePictureTaskPayload struct {
	UserID            uuid.UUID `json:"userId"`
	ProfilePictureUrl string    `json:"profilePictureUrl"`
}

type GenerateOutfitProfilePictureTaskPayload struct {
	GenerationID         uuid.UUID `json:"generationId"`
	UserID               uuid.UUID `json:"userId"`
	OwnedInventoryItemID uuid.UUID `json:"ownedInventoryItemId"`
	SelfieUrl            string    `json:"selfieUrl"`
	OutfitName           string    `json:"outfitName"`
}

type GenerateInventoryItemImageTaskPayload struct {
	InventoryItemID int    `json:"inventoryItemId"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	RarityTier      string `json:"rarityTier"`
}

type GenerateSpellIconTaskPayload struct {
	SpellID       uuid.UUID `json:"spellId"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	SchoolOfMagic string    `json:"schoolOfMagic"`
	EffectText    string    `json:"effectText"`
}

type SpellCreationSpec struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AbilityType   string `json:"abilityType"`
	AbilityLevel  int    `json:"abilityLevel"`
	EffectText    string `json:"effectText"`
	SchoolOfMagic string `json:"schoolOfMagic"`
	ManaCost      int    `json:"manaCost"`
}

type SpellBulkEffectCounts struct {
	DealDamage               int `json:"dealDamage"`
	DealDamageAllEnemies     int `json:"dealDamageAllEnemies"`
	RestoreLifePartyMember   int `json:"restoreLifePartyMember"`
	RestoreLifeAllParty      int `json:"restoreLifeAllPartyMembers"`
	ApplyBeneficialStatuses  int `json:"applyBeneficialStatuses"`
	RemoveDetrimentalEffects int `json:"removeDetrimentalStatuses"`
}

type GenerateSpellsBulkTaskPayload struct {
	JobID        uuid.UUID              `json:"jobId"`
	Source       string                 `json:"source"`
	AbilityType  string                 `json:"abilityType"`
	TotalCount   int                    `json:"totalCount"`
	TargetLevel  *int                   `json:"targetLevel,omitempty"`
	EffectCounts *SpellBulkEffectCounts `json:"effectCounts,omitempty"`
	// Deprecated: retained for backward compatibility with older clients.
	EffectMix *SpellBulkEffectCounts `json:"effectMix,omitempty"`
	Spells    []SpellCreationSpec    `json:"spells"`
}

type GenerateSpellProgressionFromPromptTaskPayload struct {
	JobID       uuid.UUID `json:"jobId"`
	Prompt      string    `json:"prompt"`
	AbilityType string    `json:"abilityType"`
}

type RebalanceSpellDamageTaskPayload struct {
	JobID    uuid.UUID   `json:"jobId"`
	SpellIDs []uuid.UUID `json:"spellIds,omitempty"`
}

type SpellBulkStatus struct {
	JobID        uuid.UUID              `json:"jobId"`
	Status       string                 `json:"status"`
	Source       string                 `json:"source"`
	AbilityType  string                 `json:"abilityType"`
	TotalCount   int                    `json:"totalCount"`
	CreatedCount int                    `json:"createdCount"`
	TargetLevel  *int                   `json:"targetLevel,omitempty"`
	EffectCounts *SpellBulkEffectCounts `json:"effectCounts,omitempty"`
	// Deprecated: retained for backward compatibility with older clients.
	EffectMix   *SpellBulkEffectCounts `json:"effectMix,omitempty"`
	Error       string                 `json:"error,omitempty"`
	QueuedAt    *time.Time             `json:"queuedAt,omitempty"`
	StartedAt   *time.Time             `json:"startedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type SpellProgressionPromptStatus struct {
	JobID           uuid.UUID   `json:"jobId"`
	Status          string      `json:"status"`
	Prompt          string      `json:"prompt"`
	AbilityType     string      `json:"abilityType"`
	CreatedCount    int         `json:"createdCount"`
	ProgressionID   *uuid.UUID  `json:"progressionId,omitempty"`
	SeedSpellID     *uuid.UUID  `json:"seedSpellId,omitempty"`
	CreatedSpellIDs []uuid.UUID `json:"createdSpellIds,omitempty"`
	Error           string      `json:"error,omitempty"`
	QueuedAt        *time.Time  `json:"queuedAt,omitempty"`
	StartedAt       *time.Time  `json:"startedAt,omitempty"`
	CompletedAt     *time.Time  `json:"completedAt,omitempty"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

type SpellDamageRebalanceStatus struct {
	JobID        uuid.UUID   `json:"jobId"`
	Status       string      `json:"status"`
	TotalCount   int         `json:"totalCount"`
	UpdatedCount int         `json:"updatedCount"`
	SpellIDs     []uuid.UUID `json:"spellIds,omitempty"`
	Error        string      `json:"error,omitempty"`
	QueuedAt     *time.Time  `json:"queuedAt,omitempty"`
	StartedAt    *time.Time  `json:"startedAt,omitempty"`
	CompletedAt  *time.Time  `json:"completedAt,omitempty"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type GenerateMonsterImageTaskPayload struct {
	MonsterID uuid.UUID `json:"monsterId"`
}

type GenerateMonsterTemplateImageTaskPayload struct {
	MonsterTemplateID uuid.UUID `json:"monsterTemplateId"`
}

type MonsterTemplateCreationSpec struct {
	MonsterType      string `json:"monsterType"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	BaseStrength     int    `json:"baseStrength"`
	BaseDexterity    int    `json:"baseDexterity"`
	BaseConstitution int    `json:"baseConstitution"`
	BaseIntelligence int    `json:"baseIntelligence"`
	BaseWisdom       int    `json:"baseWisdom"`
	BaseCharisma     int    `json:"baseCharisma"`
}

type RefreshMonsterTemplateAffinitiesTaskPayload struct {
	JobID              uuid.UUID   `json:"jobId"`
	MonsterTemplateIDs []uuid.UUID `json:"monsterTemplateIds,omitempty"`
}

type ResetMonsterTemplateProgressionsTaskPayload struct {
	JobID              uuid.UUID   `json:"jobId"`
	MonsterTemplateIDs []uuid.UUID `json:"monsterTemplateIds,omitempty"`
}

type GenerateMonsterTemplatesBulkTaskPayload struct {
	JobID       uuid.UUID                     `json:"jobId"`
	Source      string                        `json:"source"`
	MonsterType string                        `json:"monsterType"`
	TotalCount  int                           `json:"totalCount"`
	Templates   []MonsterTemplateCreationSpec `json:"templates"`
}

type MonsterTemplateBulkStatus struct {
	JobID        uuid.UUID  `json:"jobId"`
	Status       string     `json:"status"`
	Source       string     `json:"source"`
	MonsterType  string     `json:"monsterType"`
	TotalCount   int        `json:"totalCount"`
	CreatedCount int        `json:"createdCount"`
	Error        string     `json:"error,omitempty"`
	QueuedAt     *time.Time `json:"queuedAt,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type MonsterTemplateAffinityRefreshStatus struct {
	JobID        uuid.UUID   `json:"jobId"`
	Status       string      `json:"status"`
	TotalCount   int         `json:"totalCount"`
	UpdatedCount int         `json:"updatedCount"`
	TemplateIDs  []uuid.UUID `json:"templateIds,omitempty"`
	Error        string      `json:"error,omitempty"`
	QueuedAt     *time.Time  `json:"queuedAt,omitempty"`
	StartedAt    *time.Time  `json:"startedAt,omitempty"`
	CompletedAt  *time.Time  `json:"completedAt,omitempty"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type MonsterTemplateProgressionResetStatus struct {
	JobID        uuid.UUID   `json:"jobId"`
	Status       string      `json:"status"`
	TotalCount   int         `json:"totalCount"`
	UpdatedCount int         `json:"updatedCount"`
	TemplateIDs  []uuid.UUID `json:"templateIds,omitempty"`
	Error        string      `json:"error,omitempty"`
	QueuedAt     *time.Time  `json:"queuedAt,omitempty"`
	StartedAt    *time.Time  `json:"startedAt,omitempty"`
	CompletedAt  *time.Time  `json:"completedAt,omitempty"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type GenerateCharacterImageTaskPayload struct {
	CharacterID uuid.UUID `json:"characterId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type GeneratePointOfInterestImageTaskPayload struct {
	PointOfInterestID uuid.UUID `json:"pointOfInterestId"`
}

type GenerateScenarioImageTaskPayload struct {
	ScenarioID uuid.UUID `json:"scenarioId"`
}

type GenerateExpositionImageTaskPayload struct {
	ExpositionID uuid.UUID `json:"expositionId"`
}

type GenerateTutorialImageTaskPayload struct {
	ScenarioPrompt string `json:"scenarioPrompt"`
}

type GenerateChallengeImageTaskPayload struct {
	ChallengeID uuid.UUID `json:"challengeId"`
}

type GenerateChallengeTemplateImageTaskPayload struct {
	ChallengeTemplateID uuid.UUID `json:"challengeTemplateId"`
}

type GenerateScenarioTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateChallengesTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateScenarioTemplatesTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateChallengeTemplatesTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateZoneFlavorTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type GenerateImageThumbnailTaskPayload struct {
	EntityType     string    `json:"entityType"`
	EntityID       uuid.UUID `json:"entityId,omitempty"`
	SourceUrl      string    `json:"sourceUrl"`
	DestinationKey string    `json:"destinationKey,omitempty"`
}

type ImportPointOfInterestTaskPayload struct {
	ImportID uuid.UUID `json:"importId"`
}

type ImportZonesForMetroTaskPayload struct {
	ImportID uuid.UUID `json:"importId"`
}

type SeedZoneDraftTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type ApplyZoneSeedDraftTaskPayload struct {
	JobID uuid.UUID `json:"jobId"`
}

type ShuffleZoneSeedChallengeTaskPayload struct {
	JobID                uuid.UUID  `json:"jobId"`
	QuestDraftID         *uuid.UUID `json:"questDraftId,omitempty"`
	MainQuestNodeDraftID *uuid.UUID `json:"mainQuestNodeDraftId,omitempty"`
}

func MonsterTemplateBulkStatusKey(jobID uuid.UUID) string {
	return fmt.Sprintf("admin:monster-templates:bulk:%s", jobID.String())
}

func MonsterTemplateAffinityRefreshStatusKey(jobID uuid.UUID) string {
	return fmt.Sprintf("admin:monster-templates:affinity-refresh:%s", jobID.String())
}

func MonsterTemplateProgressionResetStatusKey(jobID uuid.UUID) string {
	return fmt.Sprintf("admin:monster-templates:progression-reset:%s", jobID.String())
}

func SpellBulkStatusKey(jobID uuid.UUID) string {
	return fmt.Sprintf("admin:spells:bulk:%s", jobID.String())
}

func SpellProgressionPromptStatusKey(jobID uuid.UUID) string {
	return fmt.Sprintf("admin:spells:progression-from-prompt:%s", jobID.String())
}

func SpellDamageRebalanceStatusKey(jobID uuid.UUID) string {
	return fmt.Sprintf("admin:spells:damage-rebalance:%s", jobID.String())
}
