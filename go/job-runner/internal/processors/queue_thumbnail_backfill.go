package processors

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// QueueThumbnailBackfillProcessor enqueues thumbnail generation for entities missing thumbnails.
type QueueThumbnailBackfillProcessor struct {
	dbClient    db.DbClient
	asyncClient *asynq.Client
}

func NewQueueThumbnailBackfillProcessor(dbClient db.DbClient, asyncClient *asynq.Client) QueueThumbnailBackfillProcessor {
	log.Println("Initializing QueueThumbnailBackfillProcessor")
	return QueueThumbnailBackfillProcessor{
		dbClient:    dbClient,
		asyncClient: asyncClient,
	}
}

func (p *QueueThumbnailBackfillProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if p.asyncClient == nil {
		log.Printf("QueueThumbnailBackfill: async client unavailable")
		return nil
	}

	characters, err := p.dbClient.Character().FindAll(ctx)
	if err != nil {
		return err
	}
	pointsOfInterest, err := p.dbClient.PointOfInterest().FindAll(ctx)
	if err != nil {
		return err
	}

	var queuedCharacters int
	for _, character := range characters {
		if character == nil {
			continue
		}
		if strings.TrimSpace(character.ThumbnailURL) != "" {
			continue
		}
		sourceUrl := strings.TrimSpace(character.DialogueImageURL)
		if sourceUrl == "" {
			continue
		}
		if p.enqueueThumbnailTask(jobs.ThumbnailEntityCharacter, character.ID, sourceUrl) {
			queuedCharacters++
		}
	}

	var queuedPois int
	for _, poi := range pointsOfInterest {
		if strings.TrimSpace(poi.ThumbnailURL) != "" {
			continue
		}
		sourceUrl := strings.TrimSpace(poi.ImageUrl)
		if sourceUrl == "" {
			continue
		}
		if p.enqueueThumbnailTask(jobs.ThumbnailEntityPointOfInterest, poi.ID, sourceUrl) {
			queuedPois++
		}
	}

	log.Printf("QueueThumbnailBackfill: queued %d characters, %d points of interest", queuedCharacters, queuedPois)
	return nil
}

func (p *QueueThumbnailBackfillProcessor) enqueueThumbnailTask(entityType string, entityID uuid.UUID, sourceUrl string) bool {
	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType: entityType,
		EntityID:   entityID,
		SourceUrl:  sourceUrl,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("QueueThumbnailBackfill: failed to marshal payload: %v", err)
		return false
	}
	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		log.Printf("QueueThumbnailBackfill: failed to enqueue task: %v", err)
		return false
	}
	return true
}
