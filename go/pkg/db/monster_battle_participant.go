package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterBattleParticipantHandler struct {
	db *gorm.DB
}

func (h *monsterBattleParticipantHandler) CreateOrUpdate(
	ctx context.Context,
	participant *models.MonsterBattleParticipant,
) error {
	if participant == nil {
		return nil
	}

	now := time.Now()
	existing := &models.MonsterBattleParticipant{}
	err := h.db.WithContext(ctx).
		Where("battle_id = ? AND user_id = ?", participant.BattleID, participant.UserID).
		First(existing).
		Error
	if err == nil {
		updates := map[string]interface{}{
			"is_initiator": participant.IsInitiator,
			"joined_at":    participant.JoinedAt,
			"updated_at":   now,
		}
		if participant.JoinedAt.IsZero() {
			updates["joined_at"] = existing.JoinedAt
		}
		return h.db.WithContext(ctx).Model(existing).Updates(updates).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if participant.ID == uuid.Nil {
		participant.ID = uuid.New()
	}
	if participant.CreatedAt.IsZero() {
		participant.CreatedAt = now
	}
	participant.UpdatedAt = now
	if participant.JoinedAt.IsZero() {
		participant.JoinedAt = now
	}
	return h.db.WithContext(ctx).Create(participant).Error
}

func (h *monsterBattleParticipantHandler) FindByBattleID(
	ctx context.Context,
	battleID uuid.UUID,
) ([]models.MonsterBattleParticipant, error) {
	participants := []models.MonsterBattleParticipant{}
	if err := h.db.WithContext(ctx).
		Preload("User").
		Where("battle_id = ?", battleID).
		Order("joined_at ASC").
		Find(&participants).Error; err != nil {
		return nil, err
	}
	return participants, nil
}

func (h *monsterBattleParticipantHandler) UpdateRewards(
	ctx context.Context,
	battleID uuid.UUID,
	userID uuid.UUID,
	rewardExperience int,
	rewardGold int,
	itemsAwarded []models.ItemAwarded,
) error {
	items := models.MonsterBattleItemAwards{}
	if len(itemsAwarded) > 0 {
		items = append(items, itemsAwarded...)
	}
	return h.db.WithContext(ctx).
		Model(&models.MonsterBattleParticipant{}).
		Where("battle_id = ? AND user_id = ?", battleID, userID).
		Updates(map[string]interface{}{
			"reward_experience": rewardExperience,
			"reward_gold":       rewardGold,
			"items_awarded":     items,
			"updated_at":        time.Now(),
		}).Error
}

func (h *monsterBattleParticipantHandler) DeleteByBattleAndUser(
	ctx context.Context,
	battleID uuid.UUID,
	userID uuid.UUID,
) error {
	return h.db.WithContext(ctx).
		Where("battle_id = ? AND user_id = ?", battleID, userID).
		Delete(&models.MonsterBattleParticipant{}).Error
}

func (h *monsterBattleParticipantHandler) DeleteAllForBattleID(
	ctx context.Context,
	battleID uuid.UUID,
) error {
	return h.db.WithContext(ctx).
		Where("battle_id = ?", battleID).
		Delete(&models.MonsterBattleParticipant{}).Error
}
