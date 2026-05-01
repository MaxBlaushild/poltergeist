package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func TestQueueMonsterTemplateImageGenerationQueuesTaskAndMarksTemplateQueued(t *testing.T) {
	templateID := uuid.MustParse("00000000-0000-0000-0000-000000000321")
	template := &models.MonsterTemplate{
		ID:                    templateID,
		ImageGenerationStatus: models.MonsterTemplateImageGenerationStatusNone,
	}

	updatedStatuses := make([]string, 0, 1)
	var queuedPayload jobs.GenerateMonsterTemplateImageTaskPayload
	var queuedTaskType string

	err := queueMonsterTemplateImageGeneration(
		context.Background(),
		template,
		func(_ context.Context, id uuid.UUID, updates *models.MonsterTemplate) error {
			if id != templateID {
				t.Fatalf("expected update for %s, got %s", templateID, id)
			}
			updatedStatuses = append(updatedStatuses, updates.ImageGenerationStatus)
			return nil
		},
		func(task *asynq.Task) error {
			queuedTaskType = task.Type()
			return json.Unmarshal(task.Payload(), &queuedPayload)
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(updatedStatuses) != 1 {
		t.Fatalf("expected one template update, got %d", len(updatedStatuses))
	}
	if updatedStatuses[0] != models.MonsterTemplateImageGenerationStatusQueued {
		t.Fatalf("expected queued status update, got %q", updatedStatuses[0])
	}
	if template.ImageGenerationStatus != models.MonsterTemplateImageGenerationStatusQueued {
		t.Fatalf("expected final template status queued, got %q", template.ImageGenerationStatus)
	}
	if template.ImageGenerationError == nil || *template.ImageGenerationError != "" {
		t.Fatalf("expected cleared image generation error, got %#v", template.ImageGenerationError)
	}
	if queuedTaskType != jobs.GenerateMonsterTemplateImageTaskType {
		t.Fatalf("expected task type %q, got %q", jobs.GenerateMonsterTemplateImageTaskType, queuedTaskType)
	}
	if queuedPayload.MonsterTemplateID != templateID {
		t.Fatalf("expected queued payload to target %s, got %s", templateID, queuedPayload.MonsterTemplateID)
	}
}

func TestQueueMonsterTemplateImageGenerationMarksTemplateFailedOnEnqueueError(t *testing.T) {
	template := &models.MonsterTemplate{
		ID:                    uuid.MustParse("00000000-0000-0000-0000-000000000654"),
		ImageGenerationStatus: models.MonsterTemplateImageGenerationStatusNone,
	}

	updatedStatuses := make([]string, 0, 2)
	queueErr := errors.New("queue offline")

	err := queueMonsterTemplateImageGeneration(
		context.Background(),
		template,
		func(_ context.Context, _ uuid.UUID, updates *models.MonsterTemplate) error {
			updatedStatuses = append(updatedStatuses, updates.ImageGenerationStatus)
			return nil
		},
		func(_ *asynq.Task) error {
			return queueErr
		},
	)
	if !errors.Is(err, queueErr) {
		t.Fatalf("expected queue error %v, got %v", queueErr, err)
	}

	if len(updatedStatuses) != 2 {
		t.Fatalf("expected queued and failed updates, got %d", len(updatedStatuses))
	}
	if updatedStatuses[0] != models.MonsterTemplateImageGenerationStatusQueued {
		t.Fatalf("expected first status update queued, got %q", updatedStatuses[0])
	}
	if updatedStatuses[1] != models.MonsterTemplateImageGenerationStatusFailed {
		t.Fatalf("expected second status update failed, got %q", updatedStatuses[1])
	}
	if template.ImageGenerationStatus != models.MonsterTemplateImageGenerationStatusFailed {
		t.Fatalf("expected final template status failed, got %q", template.ImageGenerationStatus)
	}
	if template.ImageGenerationError == nil || *template.ImageGenerationError != queueErr.Error() {
		t.Fatalf("expected final image generation error %q, got %#v", queueErr.Error(), template.ImageGenerationError)
	}
}
