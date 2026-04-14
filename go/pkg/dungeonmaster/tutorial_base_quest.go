package dungeonmaster

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
)

const TutorialGeneratedBaseQuestGiverTag = "tutorial_base_quest_giver"

func PurgeTutorialReplayArtifacts(
	ctx context.Context,
	dbClient db.DbClient,
	userID uuid.UUID,
	baseQuestArchetypeID *uuid.UUID,
) error {
	characters, err := dbClient.Character().FindAll(ctx)
	if err != nil {
		return err
	}
	questGiverIDs := map[uuid.UUID]struct{}{}
	for _, character := range characters {
		if character == nil || character.OwnerUserID == nil || *character.OwnerUserID != userID || !character.Ephemeral {
			continue
		}
		if !hasTutorialReplayArtifactTag(character) {
			continue
		}
		questGiverIDs[character.ID] = struct{}{}
	}

	quests, err := dbClient.Quest().FindAll(ctx)
	if err != nil {
		return err
	}
	for i := range quests {
		quest := &quests[i]
		if quest.OwnerUserID == nil || *quest.OwnerUserID != userID || !quest.Ephemeral {
			continue
		}
		if !isTutorialReplayQuestArtifact(quest, questGiverIDs, baseQuestArchetypeID) {
			continue
		}
		if quest.QuestGiverCharacterID != nil && *quest.QuestGiverCharacterID != uuid.Nil {
			questGiverIDs[*quest.QuestGiverCharacterID] = struct{}{}
		}
		if err := dbClient.Quest().Delete(ctx, quest.ID); err != nil {
			return err
		}
	}

	for _, character := range characters {
		if character == nil || character.OwnerUserID == nil || *character.OwnerUserID != userID || !character.Ephemeral {
			continue
		}
		if !isTutorialReplayCharacterArtifact(character, questGiverIDs) {
			continue
		}
		if err := dbClient.Character().Delete(ctx, character.ID); err != nil {
			return err
		}
	}

	return nil
}

func InstantiateTutorialBaseQuest(
	ctx context.Context,
	dbClient db.DbClient,
	dungeonmasterClient Client,
	userID uuid.UUID,
	base *models.Base,
	baseQuestArchetypeID uuid.UUID,
	baseQuestGiverCharacterID *uuid.UUID,
	baseQuestGiverCharacterTemplateID *uuid.UUID,
) error {
	if base == nil || baseQuestArchetypeID == uuid.Nil {
		return nil
	}

	if err := PurgeTutorialReplayArtifacts(ctx, dbClient, userID, &baseQuestArchetypeID); err != nil {
		return err
	}

	sourceTemplate, err := loadTutorialBaseQuestGiverTemplateData(
		ctx,
		dbClient,
		baseQuestGiverCharacterID,
		baseQuestGiverCharacterTemplateID,
	)
	if err != nil {
		return err
	}

	zones, err := dbClient.Zone().FindAll(ctx)
	if err != nil {
		return err
	}
	zone, err := selectTutorialZoneForCoordinates(zones, base.Latitude, base.Longitude)
	if err != nil {
		return err
	}

	cloneLatitude, cloneLongitude := offsetTutorialCharacterNearBase(base.Latitude, base.Longitude, 60)
	clonedCharacterID := uuid.New()
	now := time.Now()
	var generatedQuestID *uuid.UUID
	cleanupArtifacts := true
	defer func() {
		if !cleanupArtifacts {
			return
		}
		if generatedQuestID != nil && *generatedQuestID != uuid.Nil {
			if err := dbClient.Quest().Delete(ctx, *generatedQuestID); err != nil {
				log.Printf(
					"[tutorial] failed to clean up generated base quest quest=%s err=%v",
					generatedQuestID.String(),
					err,
				)
			}
		}
		if err := dbClient.Character().Delete(ctx, clonedCharacterID); err != nil {
			log.Printf(
				"[tutorial] failed to clean up generated base quest giver character=%s err=%v",
				clonedCharacterID.String(),
				err,
			)
		}
	}()

	clonedCharacter := sourceTemplate.Instantiate(
		models.CharacterTemplateInstanceOptions{
			ID:           clonedCharacterID,
			CreatedAt:    now,
			UpdatedAt:    now,
			OwnerUserID:  &userID,
			Ephemeral:    true,
			InternalTags: cloneTutorialQuestGiverInternalTags(sourceTemplate.InternalTags),
		},
	)
	if err := dbClient.Character().Create(ctx, clonedCharacter); err != nil {
		return err
	}
	if err := dbClient.CharacterLocation().ReplaceForCharacter(ctx, clonedCharacterID, []models.CharacterLocation{
		{
			ID:          uuid.New(),
			CreatedAt:   now,
			UpdatedAt:   now,
			CharacterID: clonedCharacterID,
			Latitude:    cloneLatitude,
			Longitude:   cloneLongitude,
		},
	}); err != nil {
		return err
	}

	quest, err := dungeonmasterClient.GenerateQuest(ctx, zone, baseQuestArchetypeID, &clonedCharacterID)
	if err != nil {
		return err
	}
	if quest == nil {
		return fmt.Errorf("failed to generate tutorial base quest")
	}
	generatedQuestID = &quest.ID
	quest.OwnerUserID = &userID
	quest.Ephemeral = true
	quest.QuestGiverCharacterID = &clonedCharacterID
	quest.UpdatedAt = time.Now()
	if err := dbClient.Quest().Update(ctx, quest.ID, quest); err != nil {
		return err
	}

	cleanupArtifacts = false
	return nil
}

func loadTutorialBaseQuestGiverTemplateData(
	ctx context.Context,
	dbClient db.DbClient,
	baseQuestGiverCharacterID *uuid.UUID,
	baseQuestGiverCharacterTemplateID *uuid.UUID,
) (models.CharacterTemplateData, error) {
	switch {
	case baseQuestGiverCharacterID != nil && *baseQuestGiverCharacterID != uuid.Nil:
		sourceCharacter, err := dbClient.Character().FindByID(ctx, *baseQuestGiverCharacterID)
		if err != nil {
			return models.CharacterTemplateData{}, err
		}
		if sourceCharacter == nil {
			return models.CharacterTemplateData{}, fmt.Errorf("tutorial base quest giver character not found")
		}
		return models.CharacterTemplateDataFromCharacter(sourceCharacter), nil
	case baseQuestGiverCharacterTemplateID != nil && *baseQuestGiverCharacterTemplateID != uuid.Nil:
		sourceTemplate, err := dbClient.CharacterTemplate().FindByID(ctx, *baseQuestGiverCharacterTemplateID)
		if err != nil {
			return models.CharacterTemplateData{}, err
		}
		if sourceTemplate == nil {
			return models.CharacterTemplateData{}, fmt.Errorf("tutorial base quest giver character template not found")
		}
		return models.CharacterTemplateDataFromCharacterTemplate(sourceTemplate), nil
	default:
		return models.CharacterTemplateData{}, fmt.Errorf("tutorial base quest giver character source is required")
	}
}

func cloneTutorialQuestGiverInternalTags(input models.StringArray) models.StringArray {
	tags := append(models.StringArray{}, input...)
	if !models.CharacterHasInternalTag(
		&models.Character{InternalTags: tags},
		TutorialGeneratedBaseQuestGiverTag,
	) {
		tags = append(tags, TutorialGeneratedBaseQuestGiverTag)
	}
	return tags
}

func isTutorialReplayCharacterArtifact(
	character *models.Character,
	questGiverIDs map[uuid.UUID]struct{},
) bool {
	if character == nil {
		return false
	}
	if _, ok := questGiverIDs[character.ID]; ok {
		return true
	}
	return hasTutorialReplayArtifactTag(character)
}

func hasTutorialReplayArtifactTag(character *models.Character) bool {
	if character == nil {
		return false
	}
	for _, rawTag := range character.InternalTags {
		if strings.EqualFold(strings.TrimSpace(rawTag), TutorialGeneratedBaseQuestGiverTag) {
			return true
		}
	}
	return false
}

func isTutorialReplayQuestArtifact(
	quest *models.Quest,
	questGiverIDs map[uuid.UUID]struct{},
	baseQuestArchetypeID *uuid.UUID,
) bool {
	if quest == nil {
		return false
	}
	if quest.QuestGiverCharacterID != nil {
		if _, ok := questGiverIDs[*quest.QuestGiverCharacterID]; ok {
			return true
		}
	}
	return baseQuestArchetypeID != nil &&
		quest.QuestArchetypeID != nil &&
		*quest.QuestArchetypeID == *baseQuestArchetypeID
}

func selectTutorialZoneForCoordinates(zones []*models.Zone, latitude float64, longitude float64) (*models.Zone, error) {
	var nearest *models.Zone
	nearestDistance := math.MaxFloat64

	for _, zone := range zones {
		if zone == nil {
			continue
		}
		if zone.IsPointInBoundary(latitude, longitude) {
			return zone, nil
		}
		distance := util.HaversineDistance(latitude, longitude, zone.Latitude, zone.Longitude)
		if distance < nearestDistance {
			nearest = zone
			nearestDistance = distance
		}
	}

	if nearest == nil {
		return nil, fmt.Errorf("no zones available")
	}
	return nearest, nil
}

func offsetTutorialCharacterNearBase(latitude float64, longitude float64, distanceMeters float64) (float64, float64) {
	if distanceMeters <= 0 {
		return latitude, longitude
	}
	latOffset := distanceMeters / 111111.0
	cosLat := math.Cos(latitude * math.Pi / 180.0)
	if math.Abs(cosLat) < 0.00001 {
		cosLat = 0.00001
	}
	lngOffset := (distanceMeters * 0.35) / (111111.0 * cosLat)
	return latitude + latOffset, longitude + lngOffset
}
