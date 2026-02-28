package jobs

import (
	"github.com/google/uuid"
)

const (
	GenerateQuestForZoneTaskType          = "generate_quest_for_zone"
	QueueQuestGenerationsTaskType         = "queue_quest_generations"
	CreateProfilePictureTaskType          = "create_profile_picture"
	GenerateOutfitProfilePictureTaskType  = "generate_outfit_profile_picture"
	GenerateInventoryItemImageTaskType    = "generate_inventory_item_image"
	GenerateSpellIconTaskType             = "generate_spell_icon"
	GenerateCharacterImageTaskType        = "generate_character_image"
	GeneratePointOfInterestImageTaskType  = "generate_point_of_interest_image"
	GenerateScenarioImageTaskType         = "generate_scenario_image"
	GenerateScenarioTaskType              = "generate_scenario"
	GenerateImageThumbnailTaskType        = "generate_image_thumbnail"
	QueueThumbnailBackfillTaskType        = "queue_thumbnail_backfill"
	SeedTreasureChestsTaskType            = "seed_treasure_chests"
	CalculateTrendingDestinationsTaskType = "calculate_trending_destinations"
	ProcessRecurringQuestsTaskType        = "process_recurring_quests"
	CleanupOrphanedQuestActionsTaskType   = "cleanup_orphaned_quest_actions"
	CheckBlockchainTransactionsTaskType   = "check_blockchain_transactions"
	ImportPointOfInterestTaskType         = "import_point_of_interest"
	ImportZonesForMetroTaskType           = "import_zones_for_metro"
	MonitorPolymarketTradesTaskType       = "monitor_polymarket_trades"
	SeedZoneDraftTaskType                 = "seed_zone_draft"
	ApplyZoneSeedDraftTaskType            = "apply_zone_seed_draft"
	ShuffleZoneSeedChallengeTaskType      = "shuffle_zone_seed_challenge"
	ShuffleQuestNodeChallengeTaskType     = "shuffle_quest_node_challenge"
)

const (
	ThumbnailEntityCharacter       = "character"
	ThumbnailEntityPointOfInterest = "point_of_interest"
	ThumbnailEntityStatic          = "static"
	ThumbnailBucket                = "crew-profile-icons"
)

type GenerateQuestForZoneTaskPayload struct {
	ZoneID                uuid.UUID  `json:"zone_id"`
	QuestArchetypeID      uuid.UUID  `json:"quest_archetype_id"`
	QuestGiverCharacterID *uuid.UUID `json:"quest_giver_character_id,omitempty"`
	QuestGenerationJobID  *uuid.UUID `json:"quest_generation_job_id,omitempty"`
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

type GenerateScenarioTaskPayload struct {
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

type ShuffleQuestNodeChallengeTaskPayload struct {
	QuestNodeChallengeID uuid.UUID `json:"questNodeChallengeId"`
}
