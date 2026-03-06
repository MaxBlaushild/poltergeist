package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterBattleInviteHandler struct {
	db *gorm.DB
}

func (h *monsterBattleInviteHandler) Create(
	ctx context.Context,
	invite *models.MonsterBattleInvite,
) error {
	if invite == nil {
		return nil
	}
	now := time.Now()
	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = now
	}
	invite.UpdatedAt = now
	return h.db.WithContext(ctx).Create(invite).Error
}

func (h *monsterBattleInviteHandler) FindByID(
	ctx context.Context,
	inviteID uuid.UUID,
) (*models.MonsterBattleInvite, error) {
	invite := &models.MonsterBattleInvite{}
	if err := h.db.WithContext(ctx).
		Preload("Inviter").
		Preload("Invitee").
		Where("id = ?", inviteID).
		First(invite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return invite, nil
}

func (h *monsterBattleInviteHandler) FindByBattleID(
	ctx context.Context,
	battleID uuid.UUID,
) ([]models.MonsterBattleInvite, error) {
	invites := []models.MonsterBattleInvite{}
	if err := h.db.WithContext(ctx).
		Preload("Inviter").
		Preload("Invitee").
		Where("battle_id = ?", battleID).
		Order("created_at ASC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

func (h *monsterBattleInviteHandler) FindPendingByInvitee(
	ctx context.Context,
	inviteeID uuid.UUID,
	now time.Time,
) ([]models.MonsterBattleInvite, error) {
	invites := []models.MonsterBattleInvite{}
	if err := h.db.WithContext(ctx).
		Preload("Inviter").
		Where("invitee_user_id = ? AND status = ? AND expires_at > ?", inviteeID, string(models.MonsterBattleInviteStatusPending), now).
		Order("created_at DESC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

func (h *monsterBattleInviteHandler) UpdateStatus(
	ctx context.Context,
	inviteID uuid.UUID,
	inviteeID uuid.UUID,
	status string,
	respondedAt *time.Time,
) (int64, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	if respondedAt != nil {
		updates["responded_at"] = *respondedAt
	}
	result := h.db.WithContext(ctx).
		Model(&models.MonsterBattleInvite{}).
		Where("id = ? AND invitee_user_id = ? AND status = ?", inviteID, inviteeID, string(models.MonsterBattleInviteStatusPending)).
		Updates(updates)
	return result.RowsAffected, result.Error
}

func (h *monsterBattleInviteHandler) AutoDeclineExpiredByBattle(
	ctx context.Context,
	battleID uuid.UUID,
	now time.Time,
) (int64, error) {
	result := h.db.WithContext(ctx).
		Model(&models.MonsterBattleInvite{}).
		Where(
			"battle_id = ? AND status = ? AND expires_at <= ?",
			battleID,
			string(models.MonsterBattleInviteStatusPending),
			now,
		).
		Updates(map[string]interface{}{
			"status":       string(models.MonsterBattleInviteStatusAutoDeclined),
			"responded_at": now,
			"updated_at":   now,
		})
	return result.RowsAffected, result.Error
}

func (h *monsterBattleInviteHandler) AutoDeclineExpiredByInvitee(
	ctx context.Context,
	inviteeID uuid.UUID,
	now time.Time,
) (int64, error) {
	result := h.db.WithContext(ctx).
		Model(&models.MonsterBattleInvite{}).
		Where(
			"invitee_user_id = ? AND status = ? AND expires_at <= ?",
			inviteeID,
			string(models.MonsterBattleInviteStatusPending),
			now,
		).
		Updates(map[string]interface{}{
			"status":       string(models.MonsterBattleInviteStatusAutoDeclined),
			"responded_at": now,
			"updated_at":   now,
		})
	return result.RowsAffected, result.Error
}

func (h *monsterBattleInviteHandler) CountPendingByBattle(
	ctx context.Context,
	battleID uuid.UUID,
	now time.Time,
) (int64, error) {
	var count int64
	err := h.db.WithContext(ctx).
		Model(&models.MonsterBattleInvite{}).
		Where("battle_id = ? AND status = ? AND expires_at > ?", battleID, string(models.MonsterBattleInviteStatusPending), now).
		Count(&count).Error
	return count, err
}
