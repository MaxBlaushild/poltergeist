package dungeonmaster

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

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
	_ = baseQuestGiverCharacterID
	_ = baseQuestGiverCharacterTemplateID
	if base == nil || baseQuestArchetypeID == uuid.Nil {
		return nil
	}

	if err := PurgeTutorialReplayArtifacts(ctx, dbClient, userID, &baseQuestArchetypeID); err != nil {
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

	explicitNoGiver := uuid.Nil
	var generatedQuestID uuid.UUID
	cleanupQuest := false
	defer func() {
		if !cleanupQuest || generatedQuestID == uuid.Nil {
			return
		}
		if err := dbClient.Quest().Delete(ctx, generatedQuestID); err != nil {
			log.Printf(
				"[tutorial] failed to clean up generated base quest quest=%s err=%v",
				generatedQuestID.String(),
				err,
			)
		}
	}()

	quest, err := dungeonmasterClient.GenerateQuest(ctx, zone, baseQuestArchetypeID, &explicitNoGiver)
	if err != nil {
		return err
	}
	if quest == nil {
		return fmt.Errorf("failed to generate tutorial base quest")
	}
	generatedQuestID = quest.ID
	cleanupQuest = true
	quest.OwnerUserID = &userID
	quest.Ephemeral = true
	quest.QuestGiverCharacterID = nil
	quest.ClosurePolicy = models.QuestClosurePolicyAuto
	quest.DebriefPolicy = models.QuestDebriefPolicyNone
	if err := dbClient.Quest().Update(ctx, quest.ID, quest); err != nil {
		return err
	}

	cleanupQuest = false
	return nil
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
