package jobs

import (
	"github.com/google/uuid"
)

const (
	PollImageGenerationTaskType      = "poll_image_generation"
	PollImageUpscaleTaskType         = "upscale_image"
	QueuePollImageGenerationTaskType = "queue_poll_image_generation"
	GenerateQuestForZoneTaskType     = "generate_quest_for_zone"
	QueueQuestGenerationsTaskType    = "queue_quest_generations"
	CreateProfilePictureTaskType     = "create_profile_picture"
)

type GenerateQuestForZoneTaskPayload struct {
	ZoneID           uuid.UUID `json:"zone_id"`
	QuestArchetypeID uuid.UUID `json:"quest_archetype_id"`
}

type CreateProfilePictureTaskPayload struct {
	UserID            uuid.UUID `json:"user_id"`
	ProfilePictureUrl string    `json:"profile_picture_url"`
}
