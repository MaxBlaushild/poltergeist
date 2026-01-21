package jobs

import (
	"github.com/google/uuid"
)

const (
	GenerateQuestForZoneTaskType          = "generate_quest_for_zone"
	QueueQuestGenerationsTaskType         = "queue_quest_generations"
	CreateProfilePictureTaskType          = "create_profile_picture"
	SeedTreasureChestsTaskType            = "seed_treasure_chests"
	CalculateTrendingDestinationsTaskType = "calculate_trending_destinations"
	CheckBlockchainTransactionsTaskType   = "check_blockchain_transactions"
)

type GenerateQuestForZoneTaskPayload struct {
	ZoneID           uuid.UUID `json:"zone_id"`
	QuestArchetypeID uuid.UUID `json:"quest_archetype_id"`
}

type CreateProfilePictureTaskPayload struct {
	UserID            uuid.UUID `json:"userId"`
	ProfilePictureUrl string    `json:"profilePictureUrl"`
}
