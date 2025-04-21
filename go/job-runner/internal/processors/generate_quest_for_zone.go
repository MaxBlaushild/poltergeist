package processors

import (
	"context"
	"encoding/json"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type GenerateQuestForZoneProcessor struct {
	dbClient      db.DbClient
	dungeonmaster dungeonmaster.Client
}

func NewGenerateQuestForZoneProcessor(dbClient db.DbClient, dungeonmaster dungeonmaster.Client) GenerateQuestForZoneProcessor {
	return GenerateQuestForZoneProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmaster,
	}
}

func (p *GenerateQuestForZoneProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.GenerateQuestForZoneTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	return p.generateQuestForZone(ctx, payload.ZoneID, payload.QuestArchetypeID)
}

func (p *GenerateQuestForZoneProcessor) generateQuestForZone(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID) error {
	zone, err := p.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		return err
	}

	if _, err := p.dungeonmaster.GenerateQuest(ctx, zone, questArchetypeID); err != nil {
		return err
	}

	return nil
}
