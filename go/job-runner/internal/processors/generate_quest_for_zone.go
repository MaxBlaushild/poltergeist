package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	log.Println("Initializing GenerateQuestForZoneProcessor")
	return GenerateQuestForZoneProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmaster,
	}
}

func (p *GenerateQuestForZoneProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate quest for zone task: %v", task.Type())

	var payload jobs.GenerateQuestForZoneTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Generating quest for zone ID: %v with quest archetype ID: %v", payload.ZoneID, payload.QuestArchetypeID)
	return p.generateQuestForZone(ctx, payload.ZoneID, payload.QuestArchetypeID)
}

func (p *GenerateQuestForZoneProcessor) generateQuestForZone(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID) error {
	log.Printf("Finding zone with ID: %v", zoneID)
	zone, err := p.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		log.Printf("Failed to find zone: %v", err)
		return fmt.Errorf("failed to find zone: %w", err)
	}

	log.Printf("Found zone: %v, generating quest...", zone.Name)
	_, err = p.dungeonmaster.GenerateQuest(ctx, zone, questArchetypeID)
	if err != nil {
		log.Printf("Failed to generate quest: %v", err)
		return fmt.Errorf("failed to generate quest: %w", err)
	}
	log.Printf("Successfully generated quest for zone %v", zone.Name)

	return nil
}
