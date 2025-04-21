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
	return QueueQuestGenerationsProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmaster,
		asyncClient:   asyncClient,
	}
}

func (p *QueueQuestGenerationsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	zoneQuestArchetypes, err := p.dbClient.ZoneQuestArchetype().FindAll(ctx)
	if err != nil {
		return err
	}

	for _, zoneQuestArchetype := range zoneQuestArchetypes {
		for i := 0; i < zoneQuestArchetype.NumberOfQuests; i++ {
			payload, err := json.Marshal(jobs.GenerateQuestForZoneTaskPayload{
				ZoneID:           zoneQuestArchetype.ZoneID,
				QuestArchetypeID: zoneQuestArchetype.QuestArchetypeID,
			})
			if err != nil {
				log.Printf("error marshalling generate quest for zone task payload: %v", err)
				continue
			}
			if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateQuestForZoneTaskType, payload)); err != nil {
				log.Printf("error enqueuing generate quest for zone task: %v", err)
				continue
			}
		}
	}

	return nil
}
