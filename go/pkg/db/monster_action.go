package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterActionHandler struct {
	db *gorm.DB
}

func (h *monsterActionHandler) GetAll(ctx context.Context) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where("active = ?", true).Order("monster_id, action_type, order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) GetByID(ctx context.Context, id uuid.UUID) (*models.MonsterAction, error) {
	var action models.MonsterAction
	err := h.db.WithContext(ctx).Where("id = ? AND active = ?", id, true).First(&action).Error
	if err != nil {
		return nil, err
	}
	return &action, nil
}

func (h *monsterActionHandler) GetByMonsterID(ctx context.Context, monsterID uuid.UUID) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where("monster_id = ? AND active = ?", monsterID, true).Order("action_type, order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) GetByMonsterIDAndType(ctx context.Context, monsterID uuid.UUID, actionType string) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where("monster_id = ? AND action_type = ? AND active = ?", monsterID, actionType, true).Order("order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) Create(ctx context.Context, action *models.MonsterAction) error {
	return h.db.WithContext(ctx).Create(action).Error
}

func (h *monsterActionHandler) Update(ctx context.Context, action *models.MonsterAction) error {
	return h.db.WithContext(ctx).Save(action).Error
}

func (h *monsterActionHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.MonsterAction{}).Where("id = ?", id).Update("active", false).Error
}

func (h *monsterActionHandler) DeleteByMonsterID(ctx context.Context, monsterID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.MonsterAction{}).Where("monster_id = ?", monsterID).Update("active", false).Error
}

func (h *monsterActionHandler) CreateBatch(ctx context.Context, actions []models.MonsterAction) error {
	if len(actions) == 0 {
		return nil
	}
	return h.db.WithContext(ctx).Create(&actions).Error
}

func (h *monsterActionHandler) UpdateOrderIndexes(ctx context.Context, monsterID uuid.UUID, actionType string, actionIDs []uuid.UUID) error {
	// Start a transaction to update all order indexes atomically
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, actionID := range actionIDs {
			err := tx.Model(&models.MonsterAction{}).
				Where("id = ? AND monster_id = ? AND action_type = ?", actionID, monsterID, actionType).
				Update("order_index", i).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *monsterActionHandler) GetNextOrderIndex(ctx context.Context, monsterID uuid.UUID, actionType string) (int, error) {
	var maxOrder int
	err := h.db.WithContext(ctx).Model(&models.MonsterAction{}).
		Where("monster_id = ? AND action_type = ? AND active = ?", monsterID, actionType, true).
		Select("COALESCE(MAX(order_index), -1)").
		Scan(&maxOrder).Error
	return maxOrder + 1, err
}

func (h *monsterActionHandler) Search(ctx context.Context, query string) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	searchTerm := "%" + query + "%"
	err := h.db.WithContext(ctx).Where(
		"active = ? AND (name ILIKE ? OR description ILIKE ? OR damage_type ILIKE ?)",
		true, searchTerm, searchTerm, searchTerm,
	).Order("monster_id, action_type, order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) GetByDamageType(ctx context.Context, damageType string) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where(
		"active = ? AND (damage_type = ? OR additional_damage_type = ?)",
		true, damageType, damageType,
	).Order("monster_id, action_type, order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) GetAttacks(ctx context.Context) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where(
		"active = ? AND attack_bonus IS NOT NULL",
		true,
	).Order("monster_id, action_type, order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) GetSaveAbilities(ctx context.Context) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where(
		"active = ? AND save_dc IS NOT NULL AND save_ability IS NOT NULL",
		true,
	).Order("monster_id, action_type, order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) GetLegendaryActions(ctx context.Context, monsterID uuid.UUID) ([]models.MonsterAction, error) {
	var actions []models.MonsterAction
	err := h.db.WithContext(ctx).Where(
		"monster_id = ? AND action_type = ? AND active = ?",
		monsterID, models.ActionTypeLegendaryAction, true,
	).Order("order_index").Find(&actions).Error
	return actions, err
}

func (h *monsterActionHandler) CloneActionsToMonster(ctx context.Context, sourceMonsterID, targetMonsterID uuid.UUID) error {
	// Get all actions from the source monster
	sourceActions, err := h.GetByMonsterID(ctx, sourceMonsterID)
	if err != nil {
		return err
	}

	// Create new actions for the target monster
	var newActions []models.MonsterAction
	for _, action := range sourceActions {
		newAction := models.MonsterAction{
			MonsterID:             targetMonsterID,
			ActionType:            action.ActionType,
			OrderIndex:            action.OrderIndex,
			Name:                  action.Name,
			Description:           action.Description,
			AttackBonus:           action.AttackBonus,
			DamageDice:            action.DamageDice,
			DamageType:            action.DamageType,
			AdditionalDamageDice:  action.AdditionalDamageDice,
			AdditionalDamageType:  action.AdditionalDamageType,
			SaveDC:                action.SaveDC,
			SaveAbility:           action.SaveAbility,
			SaveEffectHalfDamage:  action.SaveEffectHalfDamage,
			RangeReach:            action.RangeReach,
			RangeLong:             action.RangeLong,
			AreaType:              action.AreaType,
			AreaSize:              action.AreaSize,
			Recharge:              action.Recharge,
			UsesPerDay:            action.UsesPerDay,
			SpecialEffects:        action.SpecialEffects,
			LegendaryCost:         action.LegendaryCost,
			Active:                true,
		}
		newActions = append(newActions, newAction)
	}

	return h.CreateBatch(ctx, newActions)
}