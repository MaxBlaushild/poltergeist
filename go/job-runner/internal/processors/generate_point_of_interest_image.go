package processors

import (
  "context"
  "encoding/json"
  "fmt"
  "log"

  "github.com/MaxBlaushild/poltergeist/pkg/db"
  "github.com/MaxBlaushild/poltergeist/pkg/jobs"
  "github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
  "github.com/MaxBlaushild/poltergeist/pkg/models"

  "github.com/google/uuid"
  "github.com/hibiken/asynq"
)

// GeneratePointOfInterestImageProcessor refreshes a point of interest image in the background.
type GeneratePointOfInterestImageProcessor struct {
  dbClient      db.DbClient
  locationSeeder locationseeder.Client
}

func NewGeneratePointOfInterestImageProcessor(dbClient db.DbClient, locationSeeder locationseeder.Client) GeneratePointOfInterestImageProcessor {
  log.Println("Initializing GeneratePointOfInterestImageProcessor")
  return GeneratePointOfInterestImageProcessor{dbClient: dbClient, locationSeeder: locationSeeder}
}

func (p *GeneratePointOfInterestImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
  log.Printf("Processing generate point of interest image task: %v", task.Type())

  var payload jobs.GeneratePointOfInterestImageTaskPayload
  if err := json.Unmarshal(task.Payload(), &payload); err != nil {
    log.Printf("Failed to unmarshal task payload: %v", err)
    return fmt.Errorf("failed to unmarshal payload: %w", err)
  }

  poi, err := p.dbClient.PointOfInterest().FindByID(ctx, payload.PointOfInterestID)
  if err != nil {
    log.Printf("Failed to find point of interest: %v", err)
    return fmt.Errorf("failed to find point of interest: %w", err)
  }
  if poi == nil {
    return fmt.Errorf("point of interest not found")
  }

  if err := p.dbClient.PointOfInterest().UpdateImageGenerationStatus(
    ctx,
    payload.PointOfInterestID,
    models.PointOfInterestImageGenerationStatusInProgress,
    nil,
  ); err != nil {
    log.Printf("Failed to update point of interest status: %v", err)
    return fmt.Errorf("failed to update point of interest status: %w", err)
  }

  if err := p.locationSeeder.RefreshPointOfInterestImage(ctx, poi); err != nil {
    log.Printf("Failed to refresh point of interest image: %v", err)
    return p.markFailed(ctx, payload.PointOfInterestID, err)
  }

  clearedErr := ""
  if err := p.dbClient.PointOfInterest().UpdateImageGenerationStatus(
    ctx,
    payload.PointOfInterestID,
    models.PointOfInterestImageGenerationStatusComplete,
    &clearedErr,
  ); err != nil {
    log.Printf("Failed to update point of interest image status: %v", err)
    return fmt.Errorf("failed to update point of interest status: %w", err)
  }

  log.Printf("Point of interest image generated successfully for ID: %s", payload.PointOfInterestID)
  return nil
}

func (p *GeneratePointOfInterestImageProcessor) markFailed(ctx context.Context, poiID uuid.UUID, err error) error {
  errMsg := err.Error()
  if dbErr := p.dbClient.PointOfInterest().UpdateImageGenerationStatus(
    ctx,
    poiID,
    models.PointOfInterestImageGenerationStatusFailed,
    &errMsg,
  ); dbErr != nil {
    log.Printf("Failed to mark point of interest generation failed: %v", dbErr)
  }
  return err
}
