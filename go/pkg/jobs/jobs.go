package jobs

import (
	"github.com/google/uuid"
)

const (
	GenerateQuestForZoneTaskType          = "generate_quest_for_zone"
	QueueQuestGenerationsTaskType         = "queue_quest_generations"
	CreateProfilePictureTaskType          = "create_profile_picture"
	GenerateInventoryItemImageTaskType    = "generate_inventory_item_image"
	GenerateCharacterImageTaskType        = "generate_character_image"
	SeedTreasureChestsTaskType            = "seed_treasure_chests"
	CalculateTrendingDestinationsTaskType = "calculate_trending_destinations"
	CheckBlockchainTransactionsTaskType   = "check_blockchain_transactions"
	ImportPointOfInterestTaskType         = "import_point_of_interest"
)

type GenerateQuestForZoneTaskPayload struct {
	ZoneID                uuid.UUID  `json:"zone_id"`
	QuestArchetypeID      uuid.UUID  `json:"quest_archetype_id"`
	QuestGiverCharacterID *uuid.UUID `json:"quest_giver_character_id,omitempty"`
}

type CreateProfilePictureTaskPayload struct {
	UserID            uuid.UUID `json:"userId"`
	ProfilePictureUrl string    `json:"profilePictureUrl"`
}

type GenerateInventoryItemImageTaskPayload struct {
	InventoryItemID int    `json:"inventoryItemId"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	RarityTier      string `json:"rarityTier"`
}

type GenerateCharacterImageTaskPayload struct {
	CharacterID uuid.UUID `json:"characterId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type ImportPointOfInterestTaskPayload struct {
	ImportID uuid.UUID `json:"importId"`
}
