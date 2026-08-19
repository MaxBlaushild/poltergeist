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

func (h *vampireHandler) CreateMystery(ctx context.Context, name, summary, fullLore string, isSubplot bool) (*models.VampireMystery, error) {
	m := models.VampireMystery{Name: name, Summary: summary, FullLore: fullLore, Active: true, IsSubplot: isSubplot}
	if err := h.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMysteries returns every mystery AND subplot (active and inactive —
// the Super Admin list shows both; the "Host a Toast" pickers filter to
// active themselves). Callers split by IsSubplot for display — see
// ListActiveMysteriesByKind for the platform-facing pickers, which filter
// server-side instead.
func (h *vampireHandler) ListMysteries(ctx context.Context) ([]models.VampireMystery, error) {
	var out []models.VampireMystery
	if err := h.db.WithContext(ctx).Order("name ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ListActiveMysteriesByKind powers the "Host a Toast" pickers: the main
// mystery picker (isSubplot=false) and the subplot multi-select
// (isSubplot=true) both read from this, each getting only their own kind.
func (h *vampireHandler) ListActiveMysteriesByKind(ctx context.Context, isSubplot bool) ([]models.VampireMystery, error) {
	var out []models.VampireMystery
	if err := h.db.WithContext(ctx).
		Where("active = ? AND is_subplot = ?", true, isSubplot).
		Order("name ASC").
		Find(&out).Error; err != nil {
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

// ReplaceMysteryBeats reconciles a mystery's beat list against what's
// submitted: beats that already have an id are updated in place (so that id
// — and any vampire_secrets.beat_id pointing at it — survives the save);
// beats no longer present are deleted (their secrets' beat_id gets
// ON DELETE SET NULL'd, same as before, since there's no ON DELETE CASCADE
// on that column); beats with no id are inserted as new.
//
// This used to be a blind delete-then-recreate, which regenerated every
// beat's id — silently orphaning every secret's beat assignment — on every
// single save, not just when a beat was actually removed. That made beat
// ids too unstable to build anything on top of (like the beat-centric
// secrets panel below), so this reconciles instead.
func (h *vampireHandler) ReplaceMysteryBeats(ctx context.Context, mysteryID uuid.UUID, beats []models.VampireMysteryBeat) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		keepIDs := make([]uuid.UUID, 0, len(beats))
		for i := range beats {
			beats[i].MysteryID = mysteryID
			if beats[i].ID != uuid.Nil {
				keepIDs = append(keepIDs, beats[i].ID)
			}
		}

		del := tx.Where("mystery_id = ?", mysteryID)
		if len(keepIDs) > 0 {
			del = del.Where("id NOT IN ?", keepIDs)
		}
		if err := del.Delete(&models.VampireMysteryBeat{}).Error; err != nil {
			return err
		}

		for i := range beats {
			if beats[i].ID == uuid.Nil {
				if err := tx.Create(&beats[i]).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Model(&models.VampireMysteryBeat{}).
				Where("id = ?", beats[i].ID).
				Updates(map[string]interface{}{
					"ordinal":     beats[i].Ordinal,
					"title":       beats[i].Title,
					"description": beats[i].Description,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- Character secrets, scoped to a mystery ----

// ListSecretsForCharacterAndMystery is used by the Super Admin character
// content editor — a character's secrets for exactly one mystery (or
// subplot) at a time, not every secret they've ever had across every
// mystery they've appeared in. Player-/GM-facing reads go through
// ListSecretsForCharacterAndMysteries (plural) instead, which combines an
// instance's main mystery with its selected subplots in one call.
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

// ListSecretsForCharacterAndMysteries is the player-/GM-facing read path
// (see me.go/gm_content.go) — a character's secrets across a set of
// mysteries, used to combine an instance's one required mystery with
// however many subplots it has selected into a single flat list. Subplots
// don't get their own display grouping; a secret from a subplot reads
// exactly like one from the main mystery.
func (h *vampireHandler) ListSecretsForCharacterAndMysteries(ctx context.Context, characterID uuid.UUID, mysteryIDs []uuid.UUID) ([]models.VampireSecret, error) {
	if len(mysteryIDs) == 0 {
		return []models.VampireSecret{}, nil
	}
	var out []models.VampireSecret
	if err := h.db.WithContext(ctx).
		Where("character_id = ? AND mystery_id IN ?", characterID, mysteryIDs).
		Order("ordinal ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ReplaceSecretsForCharacterAndMystery removes a character's existing
// secrets for one mystery (leaving their secrets for any other mystery
// untouched) and inserts the new set.
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

// ---- Character missions, scoped to a mystery ----
// Same rationale and shape as secrets above: a character can have a
// different set of missions per mystery/subplot they're cast in.

// ListMissionsForCharacterAndMystery is used by the Super Admin character
// content editor — a character's missions for exactly one mystery (or
// subplot) at a time.
func (h *vampireHandler) ListMissionsForCharacterAndMystery(ctx context.Context, characterID, mysteryID uuid.UUID) ([]models.VampireMission, error) {
	var out []models.VampireMission
	if err := h.db.WithContext(ctx).
		Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
		Order("ordinal ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ListMissionsForCharacterAndMysteries is the player-/GM-facing read path
// (see me.go/gm_content.go) — a character's missions across a set of
// mysteries, combining an instance's one required mystery with however
// many subplots it has selected into a single flat list, exactly like
// ListSecretsForCharacterAndMysteries.
func (h *vampireHandler) ListMissionsForCharacterAndMysteries(ctx context.Context, characterID uuid.UUID, mysteryIDs []uuid.UUID) ([]models.VampireMission, error) {
	if len(mysteryIDs) == 0 {
		return []models.VampireMission{}, nil
	}
	var out []models.VampireMission
	if err := h.db.WithContext(ctx).
		Where("character_id = ? AND mystery_id IN ?", characterID, mysteryIDs).
		Order("ordinal ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ReplaceMissionsForCharacterAndMystery removes a character's existing
// missions for one mystery (leaving their missions for any other mystery
// untouched) and inserts the new set.
func (h *vampireHandler) ReplaceMissionsForCharacterAndMystery(ctx context.Context, characterID, mysteryID uuid.UUID, missions []models.VampireMission) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
			Delete(&models.VampireMission{}).Error; err != nil {
			return err
		}
		for i := range missions {
			missions[i].CharacterID = characterID
			missions[i].MysteryID = &mysteryID
		}
		if len(missions) == 0 {
			return nil
		}
		return tx.Create(&missions).Error
	})
}

// ---- Beat-centric secret management ----
// Complements ListSecretsForCharacterAndMystery/ReplaceSecretsForCharacterAndMystery
// above — those are character-centric (look at one character, decide what
// they know); these are beat-centric (look at one beat, in the Story tab's
// beat panel, and decide who knows it). Individual create/update/delete,
// not wholesale replace, since this view only ever sees one beat's slice of
// a character's secrets, never their full list — replacing wholesale here
// would silently wipe out that character's secrets for every other beat.

// ListSecretsForBeat returns every secret, across every character, that
// points at one specific beat. Character names for display are resolved by
// the frontend from its already-loaded character list, not joined here.
func (h *vampireHandler) ListSecretsForBeat(ctx context.Context, beatID uuid.UUID) ([]models.VampireSecret, error) {
	var out []models.VampireSecret
	if err := h.db.WithContext(ctx).
		Where("beat_id = ?", beatID).
		Order("created_at ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// CreateSecretForCharacterMystery appends one new secret for a character in
// a mystery, tied to the given beat. Ordinal is computed as "one past
// however many secrets this character already has for this mystery" — fine
// given ordinal is display-order only, never uniqueness-enforced (see the
// model's comment).
func (h *vampireHandler) CreateSecretForCharacterMystery(ctx context.Context, characterID, mysteryID uuid.UUID, beatID *uuid.UUID, body string) (*models.VampireSecret, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&models.VampireSecret{}).
		Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	s := models.VampireSecret{
		CharacterID: characterID,
		MysteryID:   &mysteryID,
		BeatID:      beatID,
		Ordinal:     int(count) + 1,
		Body:        body,
	}
	if err := h.db.WithContext(ctx).Create(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSecretBody edits one secret's text in place — used by the beat
// panel's inline editor, which only ever touches one secret at a time.
func (h *vampireHandler) UpdateSecretBody(ctx context.Context, id uuid.UUID, body string) error {
	return h.db.WithContext(ctx).Model(&models.VampireSecret{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"body": body}).Error
}

// DeleteSecret removes one secret outright — the beat panel's "remove"
// action.
func (h *vampireHandler) DeleteSecret(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.VampireSecret{}, "id = ?", id).Error
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

// ---- Character post-Act-1 context, scoped to a mystery ----
// Same rationale as secrets — the same character cast in two different
// mysteries knows a different version of "what happens after Act One" in
// each. Unlike secrets this is a single string, not a list, so it's
// upserted rather than wholesale-replaced.

// GetCharacterMysteryContext is the player-/GM-facing read path (see me.go
// and gm_content.go) — one character's post-Act-1 context for one specific
// mystery. Returns "" (not an error) if nothing has been written yet.
func (h *vampireHandler) GetCharacterMysteryContext(ctx context.Context, characterID, mysteryID uuid.UUID) (string, error) {
	var row models.VampireCharacterMysteryContext
	if err := h.db.WithContext(ctx).
		Where("character_id = ? AND mystery_id = ?", characterID, mysteryID).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return row.PostAct1Context, nil
}

// UpsertCharacterMysteryContext creates or overwrites a character's
// post-Act-1 context for one mystery, leaving their context for any other
// mystery untouched.
func (h *vampireHandler) UpsertCharacterMysteryContext(ctx context.Context, characterID, mysteryID uuid.UUID, postAct1Context string) error {
	return h.db.WithContext(ctx).
		Where(models.VampireCharacterMysteryContext{CharacterID: characterID, MysteryID: mysteryID}).
		Assign(models.VampireCharacterMysteryContext{PostAct1Context: postAct1Context}).
		FirstOrCreate(&models.VampireCharacterMysteryContext{}).Error
}
