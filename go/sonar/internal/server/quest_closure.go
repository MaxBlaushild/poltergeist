package server

import (
	"context"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type questClosureResult struct {
	ClosedAt             *time.Time
	ClosureMethod        models.QuestClosureMethod
	DebriefPending       bool
	DebriefedAt          *time.Time
	GoldAwarded          int
	RewardExperience     int
	ItemsAwarded         []models.ItemAwarded
	SpellsAwarded        []models.SpellAwarded
	BaseResourcesAwarded []models.BaseResourceDelta
}

func questClosureMethodForPolicy(
	quest *models.Quest,
) models.QuestClosureMethod {
	switch quest.ClosurePolicyNormalized() {
	case models.QuestClosurePolicyAuto:
		return models.QuestClosureMethodAuto
	case models.QuestClosurePolicyRemote:
		return models.QuestClosureMethodRemote
	default:
		return models.QuestClosureMethodInPerson
	}
}

func questNeedsSeparateDebrief(quest *models.Quest) bool {
	if quest == nil {
		return false
	}
	switch quest.DebriefPolicyNormalized() {
	case models.QuestDebriefPolicyRequiredForFollowup:
		return true
	case models.QuestDebriefPolicyOptional:
		return quest.ReturnBonusGold > 0 ||
			quest.ReturnBonusExperience > 0 ||
			!normalizeCharacterRelationshipState(quest.ReturnBonusRelationshipEffects).IsZero()
	default:
		return false
	}
}

func (s *server) ensureQuestObjectivesCompleted(
	ctx context.Context,
	acceptance *models.QuestAcceptanceV2,
	completedAt time.Time,
) error {
	if acceptance == nil || acceptance.ObjectivesCompletedAt != nil {
		return nil
	}
	if err := s.dbClient.QuestAcceptanceV2().MarkObjectivesCompleted(
		ctx,
		acceptance.ID,
		completedAt,
	); err != nil {
		return err
	}
	acceptance.ObjectivesCompletedAt = &completedAt
	return nil
}

func (s *server) awardQuestDebriefBonuses(
	ctx context.Context,
	userID uuid.UUID,
	quest *models.Quest,
) (int, int, error) {
	if quest == nil {
		return 0, 0, nil
	}

	goldAwarded := max(0, quest.ReturnBonusGold)
	if goldAwarded > 0 {
		if err := s.dbClient.User().AddGold(ctx, userID, goldAwarded); err != nil {
			return 0, 0, err
		}
	}

	rewardExperience := max(0, quest.ReturnBonusExperience)
	if rewardExperience > 0 {
		userLevel, err := s.dbClient.UserLevel().ProcessExperiencePointAdditions(
			ctx,
			userID,
			rewardExperience,
		)
		if err != nil {
			return 0, 0, err
		}
		if userLevel.LevelsGained > 0 {
			if _, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(
				ctx,
				userID,
				userLevel.Level,
			); err != nil {
				return 0, 0, err
			}
			if _, err := s.dbClient.UserCharacterStats().RestoreResourcesToFull(
				ctx,
				userID,
			); err != nil {
				return 0, 0, err
			}
		}
	}

	return goldAwarded, rewardExperience, nil
}

func (s *server) closeQuestAcceptance(
	ctx context.Context,
	userID uuid.UUID,
	quest *models.Quest,
	acceptance *models.QuestAcceptanceV2,
	closureMethod models.QuestClosureMethod,
	closedAt time.Time,
) (*questClosureResult, error) {
	if quest == nil || acceptance == nil {
		return nil, fmt.Errorf("quest acceptance not found")
	}
	if acceptance.IsClosed() {
		return &questClosureResult{
			ClosedAt:       acceptance.EffectiveClosedAt(),
			ClosureMethod:  models.NormalizeQuestClosureMethod(string(acceptance.ClosureMethod)),
			DebriefPending: acceptance.IsDebriefPending(),
			DebriefedAt:    acceptance.EffectiveDebriefedAt(),
		}, nil
	}

	if err := s.ensureQuestObjectivesCompleted(ctx, acceptance, closedAt); err != nil {
		return nil, err
	}

	goldAwarded, itemsAwarded, spellsAwarded, err := s.gameEngineClient.AwardQuestTurnInRewards(
		ctx,
		userID,
		quest.ID,
		nil,
	)
	if err != nil {
		return nil, err
	}

	baseResourcesAwarded, err := s.awardBaseResourcesToUser(
		ctx,
		userID,
		resolveBaseMaterialRewards(
			quest.RewardMode,
			quest.MaterialRewards,
			fmt.Sprintf("quest:%s:user:%s:materials", quest.ID, userID),
		),
		"quest_turn_in",
		&quest.ID,
	)
	if err != nil {
		return nil, err
	}

	method := models.NormalizeQuestClosureMethod(string(closureMethod))
	debriefPending := method != models.QuestClosureMethodInPerson && questNeedsSeparateDebrief(quest)
	var debriefedAt *time.Time
	if !debriefPending {
		debriefedAt = &closedAt
	}

	if err := s.dbClient.QuestAcceptanceV2().MarkClosed(
		ctx,
		acceptance.ID,
		closedAt,
		method,
		debriefPending,
		debriefedAt,
	); err != nil {
		return nil, err
	}

	acceptance.ClosedAt = &closedAt
	acceptance.ClosureMethod = method
	acceptance.DebriefPending = debriefPending
	acceptance.DebriefedAt = debriefedAt
	if debriefedAt != nil {
		acceptance.TurnedInAt = debriefedAt
	}

	if err := s.applyQuestStoryFlagsOnTurnIn(ctx, userID, quest); err != nil {
		return nil, err
	}
	if err := s.applyQuestGiverRelationshipEffectsOnTurnIn(ctx, userID, quest); err != nil {
		return nil, err
	}

	return &questClosureResult{
		ClosedAt:             acceptance.ClosedAt,
		ClosureMethod:        method,
		DebriefPending:       debriefPending,
		DebriefedAt:          debriefedAt,
		GoldAwarded:          goldAwarded,
		ItemsAwarded:         itemsAwarded,
		SpellsAwarded:        spellsAwarded,
		BaseResourcesAwarded: baseResourcesAwarded,
	}, nil
}

func (s *server) debriefQuestAcceptance(
	ctx context.Context,
	userID uuid.UUID,
	quest *models.Quest,
	acceptance *models.QuestAcceptanceV2,
	debriefedAt time.Time,
) (*questClosureResult, error) {
	if quest == nil || acceptance == nil {
		return nil, fmt.Errorf("quest acceptance not found")
	}
	if !acceptance.IsClosed() {
		return nil, fmt.Errorf("quest is not closed")
	}
	if !acceptance.IsDebriefPending() {
		return &questClosureResult{
			ClosedAt:         acceptance.EffectiveClosedAt(),
			ClosureMethod:    models.NormalizeQuestClosureMethod(string(acceptance.ClosureMethod)),
			DebriefPending:   false,
			DebriefedAt:      acceptance.EffectiveDebriefedAt(),
			RewardExperience: 0,
		}, nil
	}

	goldAwarded, rewardExperience, err := s.awardQuestDebriefBonuses(
		ctx,
		userID,
		quest,
	)
	if err != nil {
		return nil, err
	}
	if err := s.applyQuestGiverRelationshipDelta(
		ctx,
		userID,
		valueOrZeroUUID(quest.QuestGiverCharacterID),
		quest.ReturnBonusRelationshipEffects,
	); err != nil {
		return nil, err
	}
	if err := s.dbClient.QuestAcceptanceV2().MarkDebriefed(
		ctx,
		acceptance.ID,
		debriefedAt,
	); err != nil {
		return nil, err
	}

	acceptance.DebriefPending = false
	acceptance.DebriefedAt = &debriefedAt
	acceptance.TurnedInAt = &debriefedAt

	return &questClosureResult{
		ClosedAt:         acceptance.EffectiveClosedAt(),
		ClosureMethod:    models.NormalizeQuestClosureMethod(string(acceptance.ClosureMethod)),
		DebriefPending:   false,
		DebriefedAt:      acceptance.DebriefedAt,
		GoldAwarded:      goldAwarded,
		RewardExperience: rewardExperience,
	}, nil
}

func valueOrZeroUUID(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}

func (s *server) finalizeQuestClosureIfReady(
	ctx context.Context,
	quest *models.Quest,
	acceptance *models.QuestAcceptanceV2,
	now time.Time,
) error {
	if quest == nil || acceptance == nil || acceptance.IsClosed() {
		return nil
	}
	if acceptance.CurrentQuestNodeID != nil {
		return nil
	}
	if err := s.ensureQuestObjectivesCompleted(ctx, acceptance, now); err != nil {
		return err
	}
	if quest.ClosurePolicyNormalized() != models.QuestClosurePolicyAuto {
		return nil
	}
	_, err := s.closeQuestAcceptance(
		ctx,
		acceptance.UserID,
		quest,
		acceptance,
		models.QuestClosureMethodAuto,
		now,
	)
	return err
}
