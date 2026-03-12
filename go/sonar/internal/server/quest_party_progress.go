package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
)

const sharedQuestObjectiveRangeMeters = 100.0

type questNodeCompletionTarget struct {
	Quest      *models.Quest
	Acceptance *models.QuestAcceptanceV2
	Node       *models.QuestNode
}

func (s *server) findMatchingCurrentQuestNodeTargets(
	ctx context.Context,
	userID uuid.UUID,
	matches func(*models.QuestNode) bool,
) ([]questNodeCompletionTarget, error) {
	acceptances, err := s.dbClient.QuestAcceptanceV2().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	targets := make([]questNodeCompletionTarget, 0)
	for _, acceptance := range acceptances {
		if acceptance.TurnedInAt != nil {
			continue
		}

		quest, err := s.dbClient.Quest().FindByID(ctx, acceptance.QuestID)
		if err != nil {
			return nil, err
		}
		if quest == nil {
			continue
		}

		currentNode, err := s.currentQuestNode(ctx, quest, acceptance.ID)
		if err != nil {
			return nil, err
		}
		if currentNode == nil || !matches(currentNode) {
			continue
		}

		acceptanceCopy := acceptance
		currentNodeCopy := *currentNode
		targets = append(targets, questNodeCompletionTarget{
			Quest:      quest,
			Acceptance: &acceptanceCopy,
			Node:       &currentNodeCopy,
		})
	}

	return targets, nil
}

func (s *server) markQuestNodeCompleteForAcceptance(
	ctx context.Context,
	acceptance *models.QuestAcceptanceV2,
	nodeID uuid.UUID,
	completedAt time.Time,
) (bool, error) {
	if acceptance == nil {
		return false, nil
	}

	progress, err := s.dbClient.QuestNodeProgress().FindByAcceptanceAndNode(
		ctx,
		acceptance.ID,
		nodeID,
	)
	if err != nil {
		return false, err
	}
	if progress == nil {
		progress = &models.QuestNodeProgress{
			ID:                uuid.New(),
			CreatedAt:         completedAt,
			UpdatedAt:         completedAt,
			QuestAcceptanceID: acceptance.ID,
			QuestNodeID:       nodeID,
			CompletedAt:       &completedAt,
		}
		if err := s.dbClient.QuestNodeProgress().Create(ctx, progress); err != nil {
			return false, err
		}
		return true, nil
	}
	if progress.CompletedAt != nil {
		return false, nil
	}
	if err := s.dbClient.QuestNodeProgress().MarkCompleted(ctx, progress.ID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *server) shareQuestNodeCompletionWithEligiblePartyMembers(
	ctx context.Context,
	sourceUser *models.User,
	quest *models.Quest,
	node *models.QuestNode,
) {
	if sourceUser == nil || sourceUser.PartyID == nil || quest == nil || node == nil {
		return
	}

	partyMembers, err := s.dbClient.User().FindPartyMembers(ctx, sourceUser.ID)
	if err != nil {
		log.Printf(
			"[quest-share][progress] failed to load party members source=%s quest=%s node=%s err=%v",
			sourceUser.ID,
			quest.ID,
			node.ID,
			err,
		)
		return
	}

	completedAt := time.Now()
	for _, member := range partyMembers {
		isActive, err := s.livenessClient.IsActive(ctx, member.ID)
		if err != nil {
			log.Printf(
				"[quest-share][progress] active check failed source=%s member=%s quest=%s node=%s err=%v",
				sourceUser.ID,
				member.ID,
				quest.ID,
				node.ID,
				err,
			)
			continue
		}
		if !isActive {
			continue
		}

		inRange, err := s.userIsInRangeForQuestNode(ctx, member.ID, node)
		if err != nil {
			log.Printf(
				"[quest-share][progress] range check failed source=%s member=%s quest=%s node=%s err=%v",
				sourceUser.ID,
				member.ID,
				quest.ID,
				node.ID,
				err,
			)
			continue
		}
		if !inRange {
			continue
		}

		acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, member.ID, quest.ID)
		if err != nil {
			log.Printf(
				"[quest-share][progress] acceptance lookup failed source=%s member=%s quest=%s node=%s err=%v",
				sourceUser.ID,
				member.ID,
				quest.ID,
				node.ID,
				err,
			)
			continue
		}
		if acceptance == nil || acceptance.TurnedInAt != nil {
			continue
		}

		currentNode, err := s.currentQuestNode(ctx, quest, acceptance.ID)
		if err != nil {
			log.Printf(
				"[quest-share][progress] current node lookup failed source=%s member=%s quest=%s node=%s err=%v",
				sourceUser.ID,
				member.ID,
				quest.ID,
				node.ID,
				err,
			)
			continue
		}
		if currentNode == nil || currentNode.ID != node.ID {
			continue
		}

		completed, err := s.markQuestNodeCompleteForAcceptance(
			ctx,
			acceptance,
			node.ID,
			completedAt,
		)
		if err != nil {
			log.Printf(
				"[quest-share][progress] mark complete failed source=%s member=%s quest=%s node=%s err=%v",
				sourceUser.ID,
				member.ID,
				quest.ID,
				node.ID,
				err,
			)
			continue
		}
		if !completed {
			continue
		}

		s.sendQuestObjectiveSharedPush(ctx, member.ID, sourceUser, quest, node)
	}
}

func (s *server) sendQuestObjectiveSharedPush(
	ctx context.Context,
	recipientUserID uuid.UUID,
	completedBy *models.User,
	quest *models.Quest,
	node *models.QuestNode,
) {
	if quest == nil || node == nil {
		return
	}

	body := fmt.Sprintf(
		"A party member completed an objective for %s. You received credit.",
		quest.Name,
	)
	completedByUserID := ""
	if completedBy != nil {
		completedByUserID = completedBy.ID.String()
	}

	s.sendSocialPushToUser(
		ctx,
		"quest-objective-shared",
		recipientUserID,
		"Quest Objective Complete",
		body,
		map[string]string{
			"type":              "quest_objective_shared",
			"questId":           quest.ID.String(),
			"questNodeId":       node.ID.String(),
			"completedByUserId": completedByUserID,
			"sentAt":            time.Now().UTC().Format(time.RFC3339),
		},
	)
}

func (s *server) userIsInRangeForQuestNode(
	ctx context.Context,
	userID uuid.UUID,
	node *models.QuestNode,
) (bool, error) {
	if node == nil {
		return false, nil
	}

	userLat, userLng, err := s.getUserLatLng(ctx, userID)
	if err != nil {
		return false, err
	}

	switch {
	case node.ScenarioID != nil:
		scenario, err := s.dbClient.Scenario().FindByID(ctx, *node.ScenarioID)
		if err != nil {
			return false, err
		}
		if scenario == nil {
			return false, nil
		}
		return util.HaversineDistance(
			userLat,
			userLng,
			scenario.Latitude,
			scenario.Longitude,
		) <= sharedQuestObjectiveRangeMeters, nil
	case node.MonsterEncounterID != nil:
		encounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, *node.MonsterEncounterID)
		if err != nil {
			return false, err
		}
		if encounter == nil {
			return false, nil
		}
		return util.HaversineDistance(
			userLat,
			userLng,
			encounter.Latitude,
			encounter.Longitude,
		) <= sharedQuestObjectiveRangeMeters, nil
	case node.MonsterID != nil:
		monster, err := s.dbClient.Monster().FindByID(ctx, *node.MonsterID)
		if err != nil {
			return false, err
		}
		if monster == nil {
			return false, nil
		}
		return util.HaversineDistance(
			userLat,
			userLng,
			monster.Latitude,
			monster.Longitude,
		) <= sharedQuestObjectiveRangeMeters, nil
	case node.ChallengeID != nil:
		challenge, err := s.dbClient.Challenge().FindByID(ctx, *node.ChallengeID)
		if err != nil {
			return false, err
		}
		if challenge == nil {
			return false, nil
		}
		if challenge.HasPolygon() {
			return challenge.ContainsPoint(userLat, userLng), nil
		}
		return util.HaversineDistance(
			userLat,
			userLng,
			challenge.Latitude,
			challenge.Longitude,
		) <= sharedQuestObjectiveRangeMeters, nil
	default:
		return false, nil
	}
}
