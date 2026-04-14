package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

type InstantiateTutorialBaseQuestProcessor struct {
	dbClient      db.DbClient
	dungeonmaster dungeonmaster.Client
}

func NewInstantiateTutorialBaseQuestProcessor(
	dbClient db.DbClient,
	dungeonmasterClient dungeonmaster.Client,
) InstantiateTutorialBaseQuestProcessor {
	log.Println("Initializing InstantiateTutorialBaseQuestProcessor")
	return InstantiateTutorialBaseQuestProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmasterClient,
	}
}

func (p *InstantiateTutorialBaseQuestProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing instantiate tutorial base quest task: %v", task.Type())

	var payload jobs.InstantiateTutorialBaseQuestTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal tutorial base quest task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	base := &models.Base{
		Latitude:  payload.BaseLatitude,
		Longitude: payload.BaseLongitude,
	}
	if err := dungeonmaster.InstantiateTutorialBaseQuest(
		ctx,
		p.dbClient,
		p.dungeonmaster,
		payload.UserID,
		base,
		payload.BaseQuestArchetypeID,
		payload.BaseQuestGiverCharacterID,
		payload.BaseQuestGiverCharacterTemplateID,
	); err != nil {
		return fmt.Errorf("failed to instantiate tutorial base quest: %w", err)
	}

	return nil
}
