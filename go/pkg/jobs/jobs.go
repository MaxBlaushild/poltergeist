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
	GenerateCharacterImageTaskType        = "generate_character_image"
	GeneratePointOfInterestImageTaskType  = "generate_point_of_interest_image"
	SeedTreasureChestsTaskType            = "seed_treasure_chests"
	CalculateTrendingDestinationsTaskType = "calculate_trending_destinations"
	ProcessRecurringQuestsTaskType        = "process_recurring_quests"
	CheckBlockchainTransactionsTaskType   = "check_blockchain_transactions"
	ImportPointOfInterestTaskType         = "import_point_of_interest"
	ImportZonesForMetroTaskType           = "import_zones_for_metro"
	MonitorPolymarketTradesTaskType       = "monitor_polymarket_trades"
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

type GenerateCharacterImageTaskPayload struct {
	CharacterID uuid.UUID `json:"characterId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type GeneratePointOfInterestImageTaskPayload struct {
	PointOfInterestID uuid.UUID `json:"pointOfInterestId"`
}

type ImportPointOfInterestTaskPayload struct {
	ImportID uuid.UUID `json:"importId"`
}

type ImportZonesForMetroTaskPayload struct {
	ImportID uuid.UUID `json:"importId"`
}
