package processors

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/hibiken/asynq"
)

type QueueQuestGenerationsProcessor struct {
	dbClient      db.DbClient
	dungeonmaster dungeonmaster.Client
	asyncClient   *asynq.Client
}

func NewQueueQuestGenerationsProcessor(dbClient db.DbClient, dungeonmaster dungeonmaster.Client, asyncClient *asynq.Client) QueueQuestGenerationsProcessor {
	log.Println("Initializing QueueQuestGenerationsProcessor")
	return QueueQuestGenerationsProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmaster,
		asyncClient:   asyncClient,
	}
}

func (p *QueueQuestGenerationsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing queue quest generations task: %v", task.Type())

	log.Println("Fetching all zone quest archetypes")
	zoneQuestArchetypes, err := p.dbClient.ZoneQuestArchetype().FindAll(ctx)
	if err != nil {
		log.Printf("Failed to fetch zone quest archetypes: %v", err)
		return err
	}
	log.Printf("Found %d zone quest archetypes", len(zoneQuestArchetypes))

	for _, zoneQuestArchetype := range zoneQuestArchetypes {
		log.Printf("Processing zone quest archetype for zone ID: %v with quest archetype ID: %v, number of quests: %d",
			zoneQuestArchetype.ZoneID, zoneQuestArchetype.QuestArchetypeID, zoneQuestArchetype.NumberOfQuests)

		for i := 0; i < zoneQuestArchetype.NumberOfQuests; i++ {
			log.Printf("Creating quest %d/%d for zone quest archetype", i+1, zoneQuestArchetype.NumberOfQuests)

			payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
				ZoneID:           zoneQuestArchetype.ZoneID,
				QuestArchetypeID: zoneQuestArchetype.QuestArchetypeID,
			})
			if err != nil {
				log.Printf("error marshalling generate quest for zone task payload: %v", err)
				continue
			}

			log.Printf("Enqueueing generate quest task for zone ID: %v", zoneQuestArchetype.ZoneID)
			if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateQuestForZoneTaskType, payload)); err != nil {
				log.Printf("error enqueuing generate quest for zone task: %v", err)
				continue
			}
			log.Printf("Successfully enqueued generate quest task for zone ID: %v", zoneQuestArchetype.ZoneID)
		}
	}

	log.Println("Completed processing queue quest generations task")
	return nil
}
