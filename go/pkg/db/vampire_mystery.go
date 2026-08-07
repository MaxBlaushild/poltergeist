package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Mysteries — the underlying story an instance's players are solving. See
// go/vampire-ascendancy/docs/MYSTERY_REQUIREMENTS.md. Shared content,
// created/edited only by super users.

func (h *vampireHandler) CreateMystery(ctx context.Context, name, summary, fullLore string) (*models.VampireMystery, error) {
	m := models.VampireMystery{Name: name, Summary: summary, FullLore: fullLore, Active: true}
	if err := h.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMysteries returns every mystery (active and inactive — the Super
// Admin list shows both; the "Host a Toast" picker filters to active
// itself).
func (h *vampireHandler) ListMysteries(ctx context.Context) ([]models.VampireMystery, error) {
	var out []models.VampireMystery
	if err := h.db.WithContext(ctx).Order("name ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (h *vampireHandler) GetMysteryByID(ctx context.Context, id uuid.UUID) (*models.VampireMystery, error) {
	var m models.VampireMystery
	if err := h.db.WithContext(ctx).
		Preload("Beats", func(db *gorm.DB) *gorm.DB { return db.Order("ordinal ASC") }).
		First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// GetMysteryByName is used by cmd/seed, whose seed/*.json files are all
// authored for "The Crimson Toast" specifically — it looks that mystery up
// by name rather than taking a --mystery-id flag, and fails loudly if it's
// missing rather than silently creating a duplicate.
func (h *vampireHandler) GetMysteryByName(ctx context.Context, name string) (*models.VampireMystery, error) {
	var m models.VampireMystery
	if err := h.db.WithContext(ctx).First(&m, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (h *vampireHandler) UpdateMystery(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return h.db.WithContext(ctx).Model(&models.VampireMystery{}).Where("id = ?", id).Updates(fields).Error
}

// ReplaceMysteryBeats removes a mystery's existing beats and inserts the
// new set — same wholesale-replace pattern as ReplaceSecrets/ReplaceMissions.
// Existing vampire_secrets.beat_id references to the removed beats are set
// to NULL by the FK (no ON DELETE CASCADE on that column), not deleted —
// see the Super Admin mystery editor, which warns before removing a beat
// that's still referenced.
func (h *vampireHandler) ReplaceMysteryBeats(ctx context.Context, mysteryID uuid.UUID, beats []models.VampireMysteryBeat) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("mystery_id = ?", mysteryID).Delete(&models.VampireMysteryBeat{}).Error; err != nil {
			return err
		}
		for i := range beats {
			beats[i].MysteryID = mysteryID
		}
		if len(beats) == 0 {
			return nil
		}
		return tx.Create(&beats).Error
	})
}

// ---- Character secrets, scoped to a mystery ----

// ListSecretsForCharacterAndMystery is the player-facing read path (see
// me.go) — a character's secrets for one specific mystery, not every
// secret they've ever had across every mystery they've appeared in.
func (h *vampireHandler) ListSecretsForCharacterAndMystery(ctx context.Context, characterID, mysteryID uuid.UUID) ([]models.VampireSecret, error) {
	var out []models.VampireSecret
	if err := h.db.WithContext(ctx).
		Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
		Order("ordinal ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ReplaceSecretsForCharacterAndMystery removes a character's existing
// secrets for one mystery (leaving their secrets for any other mystery
// untouched) and inserts the new set — the mystery-scoped equivalent of
// the old ReplaceSecrets.
func (h *vampireHandler) ReplaceSecretsForCharacterAndMystery(ctx context.Context, characterID, mysteryID uuid.UUID, secrets []models.VampireSecret) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
			Delete(&models.VampireSecret{}).Error; err != nil {
			return err
		}
		for i := range secrets {
			secrets[i].CharacterID = characterID
			secrets[i].MysteryID = &mysteryID
		}
		if len(secrets) == 0 {
			return nil
		}
		return tx.Create(&secrets).Error
	})
}

// CharacterHasSecretsForMystery is the eligibility check invites rely on:
// a character with zero secrets authored for a mystery can't be invited to
// an instance running it. Used both by the Invites tab's picker (via
// ListCharacterIDsWithSecretsForMystery, for filtering a whole list) and
// as a defense-in-depth check inside CreatePlayerInvite itself.
func (h *vampireHandler) CharacterHasSecretsForMystery(ctx context.Context, characterID, mysteryID uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&models.VampireSecret{}).
		Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListCharacterIDsWithSecretsForMystery powers the Invites tab picker's
// eligibility filter in one query instead of one CharacterHasSecretsForMystery
// call per character.
func (h *vampireHandler) ListCharacterIDsWithSecretsForMystery(ctx context.Context, mysteryID uuid.UUID) (map[uuid.UUID]bool, error) {
	var ids []uuid.UUID
	if err := h.db.WithContext(ctx).Model(&models.VampireSecret{}).
		Where("mystery_id = ?", mysteryID).
		Distinct().
		Pluck("character_id", &ids).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}
