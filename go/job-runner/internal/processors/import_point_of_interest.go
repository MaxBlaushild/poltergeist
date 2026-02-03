package processors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/hibiken/asynq"
)

type ImportPointOfInterestProcessor struct {
	dbClient       db.DbClient
	locationSeeder locationseeder.Client
}

func NewImportPointOfInterestProcessor(dbClient db.DbClient, locationSeeder locationseeder.Client) *ImportPointOfInterestProcessor {
	return &ImportPointOfInterestProcessor{
		dbClient:       dbClient,
		locationSeeder: locationSeeder,
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

	poi, err := p.locationSeeder.ImportPlace(ctx, importItem.PlaceID, *zone)
	if err != nil {
		msg := err.Error()
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.PointOfInterestImport().Update(ctx, importItem)
		return err
	}

	importItem.Status = "completed"
	importItem.PointOfInterestID = &poi.ID
	importItem.ErrorMessage = nil
	importItem.UpdatedAt = time.Now()
	return p.dbClient.PointOfInterestImport().Update(ctx, importItem)
}
