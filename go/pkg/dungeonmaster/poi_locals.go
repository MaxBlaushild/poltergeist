package dungeonmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster/poilocals"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func (c *client) ensurePointOfInterestLocals(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
) error {
	if zone == nil || poi == nil {
		return nil
	}

	existing, err := c.dbClient.Character().FindByPointOfInterestID(ctx, poi.ID)
	if err != nil {
		return err
	}

	desiredCount := poilocals.DesiredLocalCount(pointOfInterestLocalSeedKey(poi))
	existingGeneratedCount := 0
	existingNames := map[string]struct{}{}
	for _, character := range existing {
		if character == nil {
			continue
		}
		if nameKey := normalizedCharacterName(character.Name); nameKey != "" {
			existingNames[nameKey] = struct{}{}
		}
		if models.CharacterHasInternalTag(character, models.CharacterInternalTagGeneratedPOILocal) {
			existingGeneratedCount++
		}
	}
	if existingGeneratedCount >= desiredCount {
		return nil
	}

	zoneContext := poilocals.ZoneContext{
		Name:        strings.TrimSpace(zone.Name),
		Description: strings.TrimSpace(zone.Description),
	}
	placeContext := poilocals.PlaceContext{
		ID:               pointOfInterestLocalSeedKey(poi),
		Name:             strings.TrimSpace(poi.Name),
		OriginalName:     strings.TrimSpace(poi.OriginalName),
		Description:      strings.TrimSpace(poi.Description),
		EditorialSummary: strings.TrimSpace(poi.Clue),
		Types:            []string{},
	}

	drafts := poilocals.GenerateDrafts(c.deepPriest, zoneContext, []poilocals.PlaceContext{placeContext})
	createdCount, err := c.createPointOfInterestLocalsFromDrafts(
		ctx,
		poi,
		drafts,
		existingNames,
		desiredCount-existingGeneratedCount,
	)
	if err != nil {
		return err
	}
	remaining := desiredCount - existingGeneratedCount - createdCount
	if remaining <= 0 {
		return nil
	}

	_, err = c.createPointOfInterestLocalsFromDrafts(
		ctx,
		poi,
		poilocals.FallbackDrafts(zoneContext, placeContext),
		existingNames,
		remaining,
	)
	return err
}

func (c *client) createPointOfInterestLocalsFromDrafts(
	ctx context.Context,
	poi *models.PointOfInterest,
	drafts []poilocals.CharacterDraft,
	existingNames map[string]struct{},
	needed int,
) (int, error) {
	if needed <= 0 {
		return 0, nil
	}
	created := 0
	for _, draft := range drafts {
		if created >= needed {
			break
		}
		nameKey := normalizedCharacterName(draft.Name)
		if nameKey == "" {
			continue
		}
		if _, exists := existingNames[nameKey]; exists {
			continue
		}
		character, err := c.createPointOfInterestLocalCharacter(ctx, poi, draft)
		if err != nil {
			return created, err
		}
		if err := c.ensurePointOfInterestLocalTalkAction(ctx, character, draft.Dialogue, poi); err != nil {
			return created, err
		}
		existingNames[nameKey] = struct{}{}
		created++
	}
	return created, nil
}

func (c *client) createPointOfInterestLocalCharacter(
	ctx context.Context,
	poi *models.PointOfInterest,
	draft poilocals.CharacterDraft,
) (*models.Character, error) {
	imageStatus := models.CharacterImageGenerationStatusNone
	if c.asyncClient != nil {
		imageStatus = models.CharacterImageGenerationStatusQueued
	}

	character := &models.Character{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Name:                  strings.TrimSpace(draft.Name),
		Description:           strings.TrimSpace(draft.Description),
		PointOfInterestID:     &poi.ID,
		InternalTags:          models.StringArray{models.CharacterInternalTagGeneratedPOILocal},
		ImageGenerationStatus: imageStatus,
	}
	if err := c.dbClient.Character().Create(ctx, character); err != nil {
		return nil, err
	}
	if c.asyncClient == nil {
		return character, nil
	}
	if err := c.enqueueCharacterImageTask(character); err != nil {
		errMsg := err.Error()
		_ = c.dbClient.Character().Update(ctx, character.ID, &models.Character{
			ImageGenerationStatus: models.CharacterImageGenerationStatusFailed,
			ImageGenerationError:  &errMsg,
		})
		return nil, err
	}
	return character, nil
}

func (c *client) enqueueCharacterImageTask(character *models.Character) error {
	if c.asyncClient == nil || character == nil {
		return nil
	}
	payloadBytes, err := json.Marshal(jobs.GenerateCharacterImageTaskPayload{
		CharacterID: character.ID,
		Name:        character.Name,
		Description: character.Description,
	})
	if err != nil {
		return err
	}
	_, err = c.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateCharacterImageTaskType, payloadBytes))
	return err
}

func (c *client) ensurePointOfInterestLocalTalkAction(
	ctx context.Context,
	character *models.Character,
	dialogueLines []string,
	poi *models.PointOfInterest,
) error {
	if character == nil {
		return fmt.Errorf("character is nil")
	}
	lines := sanitizePointOfInterestLocalDialogue(dialogueLines)
	if len(lines) == 0 {
		return nil
	}

	actions, err := c.dbClient.CharacterAction().FindByCharacterID(ctx, character.ID)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if action == nil || action.ActionType != models.ActionTypeTalk || action.Metadata == nil {
			continue
		}
		if strings.ToLower(strings.TrimSpace(fmt.Sprint(action.Metadata["source"]))) == "poilocal" {
			return nil
		}
	}

	metadata := map[string]interface{}{"source": "poiLocal"}
	if poi != nil {
		metadata["pointOfInterestId"] = poi.ID.String()
	}
	return c.dbClient.CharacterAction().Create(ctx, &models.CharacterAction{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CharacterID: character.ID,
		ActionType:  models.ActionTypeTalk,
		Dialogue:    models.DialogueSequenceFromStringLines(lines),
		Metadata:    metadata,
	})
}

func sanitizePointOfInterestLocalDialogue(lines []string) []string {
	seen := map[string]struct{}{}
	sanitized := make([]string, 0, 3)
	for _, raw := range lines {
		line := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
		if line == "" {
			continue
		}
		if len(line) > 180 {
			line = strings.TrimSpace(line[:180])
		}
		key := strings.ToLower(line)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		sanitized = append(sanitized, line)
		if len(sanitized) >= 3 {
			break
		}
	}
	return sanitized
}

func pointOfInterestLocalSeedKey(poi *models.PointOfInterest) string {
	if poi == nil {
		return "poi-local"
	}
	if poi.GoogleMapsPlaceID != nil {
		if placeID := strings.TrimSpace(*poi.GoogleMapsPlaceID); placeID != "" {
			return placeID
		}
	}
	return poi.ID.String()
}

func normalizedCharacterName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
