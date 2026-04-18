package processors

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ImportPointOfInterestProcessor struct {
	dbClient       db.DbClient
	locationSeeder locationseeder.Client
	asyncClient    *asynq.Client
}

func NewImportPointOfInterestProcessor(dbClient db.DbClient, locationSeeder locationseeder.Client, asyncClient *asynq.Client) *ImportPointOfInterestProcessor {
	return &ImportPointOfInterestProcessor{
		dbClient:       dbClient,
		locationSeeder: locationSeeder,
		asyncClient:    asyncClient,
	}
}

func (p *ImportPointOfInterestProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	payload := jobs.ImportPointOfInterestTaskPayload{}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	importItem, err := p.dbClient.PointOfInterestImport().FindByID(ctx, payload.ImportID)
	if err != nil {
		return err
	}
	if importItem == nil {
		return nil
	}

	importItem.Status = "in_progress"
	importItem.UpdatedAt = time.Now()
	if err := p.dbClient.PointOfInterestImport().Update(ctx, importItem); err != nil {
		return err
	}

	zone, err := p.dbClient.Zone().FindByID(ctx, importItem.ZoneID)
	if err != nil {
		msg := err.Error()
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.PointOfInterestImport().Update(ctx, importItem)
		return err
	}
	if zone == nil {
		msg := "zone not found"
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.PointOfInterestImport().Update(ctx, importItem)
		return nil
	}

	poi, err := p.locationSeeder.ImportPlace(ctx, importItem.PlaceID, *zone, importItem.Genre)
	if err != nil {
		msg := err.Error()
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.PointOfInterestImport().Update(ctx, importItem)
		return err
	}

	p.enqueueThumbnailTask(poi.ID, poi.ImageUrl)

	importItem.Status = "completed"
	importItem.PointOfInterestID = &poi.ID
	importItem.ErrorMessage = nil
	importItem.UpdatedAt = time.Now()
	return p.dbClient.PointOfInterestImport().Update(ctx, importItem)
}

func (p *ImportPointOfInterestProcessor) enqueueThumbnailTask(poiID uuid.UUID, imageURL string) {
	if p.asyncClient == nil || strings.TrimSpace(imageURL) == "" {
		return
	}
	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType: jobs.ThumbnailEntityPointOfInterest,
		EntityID:   poiID,
		SourceUrl:  imageURL,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes))
}
