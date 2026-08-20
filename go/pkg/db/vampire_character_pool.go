package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The character pool — which of an instance's mystery-eligible characters
// a Host has made selectable by players. Players pick their own character
// (from the pool, minus whoever's already claimed one) after accepting
// their invite, instead of a Host assigning one up front — see
// MYSTERY_REQUIREMENTS.md and player_invites.go/character_self_select.go.

// SetCharacterPool wholesale-replaces the set of characters selectable by
// players in this instance — the Character Pool tab's save action.
func (h *vampireHandler) SetCharacterPool(ctx context.Context, instanceID uuid.UUID, characterIDs []uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("instance_id = ?", instanceID).Delete(&models.VampireInstanceCharacterPool{}).Error; err != nil {
			return err
		}
		if len(characterIDs) == 0 {
			return nil
		}
		rows := make([]models.VampireInstanceCharacterPool, 0, len(characterIDs))
		for _, cid := range characterIDs {
			rows = append(rows, models.VampireInstanceCharacterPool{InstanceID: instanceID, CharacterID: cid})
		}
		return tx.Create(&rows).Error
	})
}

// ListCharacterPoolIDs returns the set of character ids in an instance's
// pool.
func (h *vampireHandler) ListCharacterPoolIDs(ctx context.Context, instanceID uuid.UUID) (map[uuid.UUID]bool, error) {
	var ids []uuid.UUID
	if err := h.db.WithContext(ctx).Model(&models.VampireInstanceCharacterPool{}).
		Where("instance_id = ?", instanceID).
		Pluck("character_id", &ids).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// ClaimCharacterForPlayer lets a signed-in player pick their own character
// from the instance's pool — the self-select replacement for a Host
// assigning one at invite time. Row-locks the player during the check so
// two rapid claims by the same player can't both succeed. Returns a
// *ConflictError if the player already holds a character, the character
// isn't in the pool, or it's already claimed by another active player (a
// race between two players tapping the same one).
func (h *vampireHandler) ClaimCharacterForPlayer(ctx context.Context, instanceID, playerID, characterID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var player models.VampirePlayer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&player, "id = ? AND instance_id = ?", playerID, instanceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &ConflictError{Message: "player not found"}
			}
			return err
		}
		if player.CharacterID != nil {
			return &ConflictError{Message: "you've already chosen a character"}
		}

		var inPool int64
		if err := tx.Model(&models.VampireInstanceCharacterPool{}).
			Where("instance_id = ? AND character_id = ?", instanceID, characterID).
			Count(&inPool).Error; err != nil {
			return err
		}
		if inPool == 0 {
			return &ConflictError{Message: "this character isn't available to choose"}
		}

		var taken int64
		if err := tx.Model(&models.VampirePlayer{}).
			Where("instance_id = ? AND character_id = ? AND active = ?", instanceID, characterID, true).
			Count(&taken).Error; err != nil {
			return err
		}
		if taken > 0 {
			return &ConflictError{Message: "someone else just chose this character — pick another"}
		}

		return tx.Model(&player).Update("character_id", characterID).Error
	})
}
