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

// GetMysteryByID does NOT include beats — they're many-to-many now (a beat
// can be shared across multiple mysteries/subplots), so there's no direct
// relation to preload. Fetch them separately via ListBeatsForMystery.
func (h *vampireHandler) GetMysteryByID(ctx context.Context, id uuid.UUID) (*models.VampireMystery, error) {
	var m models.VampireMystery
	if err := h.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
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
// MysteryBeat is a beat as attached to one specific mystery: the beat's own
// shared content (id/title/description) plus its ordinal within THIS
// mystery's list, and how many mysteries/subplots it's linked to in total.
// LinkCount > 1 means editing this beat's content also changes what every
// other mystery/subplot sharing it shows — see ReplaceMysteryBeats.
type MysteryBeat struct {
	ID          uuid.UUID
	Title       string
	Description string
	Ordinal     int
	LinkCount   int
}

// ListBeatsForMystery returns the beats attached to one mystery, in that
// mystery's own order.
func (h *vampireHandler) ListBeatsForMystery(ctx context.Context, mysteryID uuid.UUID) ([]MysteryBeat, error) {
	var links []models.VampireMysteryBeatLink
	if err := h.db.WithContext(ctx).
		Where("mystery_id = ?", mysteryID).
		Order("ordinal ASC").
		Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []MysteryBeat{}, nil
	}
	ids := make([]uuid.UUID, len(links))
	for i, l := range links {
		ids[i] = l.BeatID
	}
	var beatRows []models.VampireMysteryBeat
	if err := h.db.WithContext(ctx).Where("id IN ?", ids).Find(&beatRows).Error; err != nil {
		return nil, err
	}
	beatByID := make(map[uuid.UUID]models.VampireMysteryBeat, len(beatRows))
	for _, b := range beatRows {
		beatByID[b.ID] = b
	}
	counts, err := h.countLinksByBeat(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]MysteryBeat, 0, len(links))
	for _, l := range links {
		b, ok := beatByID[l.BeatID]
		if !ok {
			continue
		}
		out = append(out, MysteryBeat{
			ID: b.ID, Title: b.Title, Description: b.Description,
			Ordinal: l.Ordinal, LinkCount: counts[b.ID],
		})
	}
	return out, nil
}

// ListAllBeats returns every beat across every mystery/subplot — for the
// "attach an existing beat" picker, so a super user can reuse a beat that
// already says the same thing instead of duplicating it.
func (h *vampireHandler) ListAllBeats(ctx context.Context) ([]MysteryBeat, error) {
	var beatRows []models.VampireMysteryBeat
	if err := h.db.WithContext(ctx).Order("title ASC").Find(&beatRows).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(beatRows))
	for i, b := range beatRows {
		ids[i] = b.ID
	}
	counts, err := h.countLinksByBeat(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]MysteryBeat, 0, len(beatRows))
	for _, b := range beatRows {
		out = append(out, MysteryBeat{ID: b.ID, Title: b.Title, Description: b.Description, LinkCount: counts[b.ID]})
	}
	return out, nil
}

// countLinksByBeat is a helper for ListBeatsForMystery/ListAllBeats — how
// many mysteries/subplots each of the given beats is currently attached to.
func (h *vampireHandler) countLinksByBeat(ctx context.Context, beatIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(beatIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	var rows []struct {
		BeatID uuid.UUID
		Count  int
	}
	if err := h.db.WithContext(ctx).Model(&models.VampireMysteryBeatLink{}).
		Select("beat_id, COUNT(*) as count").
		Where("beat_id IN ?", beatIDs).
		Group("beat_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		out[r.BeatID] = r.Count
	}
	return out, nil
}

// CountBeatsByMystery returns how many beats each mystery/subplot has
// linked, in one query — used by the Super Admin list view instead of one
// ListBeatsForMystery call per row.
func (h *vampireHandler) CountBeatsByMystery(ctx context.Context) (map[uuid.UUID]int, error) {
	var rows []struct {
		MysteryID uuid.UUID
		Count     int
	}
	if err := h.db.WithContext(ctx).Model(&models.VampireMysteryBeatLink{}).
		Select("mystery_id, COUNT(*) as count").
		Group("mystery_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		out[r.MysteryID] = r.Count
	}
	return out, nil
}

// ReplaceMysteryBeats reconciles one mystery's beat list against what's
// submitted:
//   - a submitted beat with an id updates that beat's shared content
//     (title/description) — which changes it everywhere it's linked, not
//     just here — and creates or updates this mystery's link (ordinal);
//     this is also how "attach an existing beat" works, since from this
//     function's point of view a freshly-attached existing beat looks the
//     same as one being reordered;
//   - a submitted beat with no id creates a brand-new beat and links it;
//   - a beat currently linked to this mystery but missing from the
//     submitted list is unlinked (the link row deleted) — the beat itself
//     is never deleted, since it may still be linked to other
//     mysteries/subplots.
func (h *vampireHandler) ReplaceMysteryBeats(ctx context.Context, mysteryID uuid.UUID, beats []MysteryBeat) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		keepIDs := make([]uuid.UUID, 0, len(beats))
		for i := range beats {
			if beats[i].ID != uuid.Nil {
				keepIDs = append(keepIDs, beats[i].ID)
			}
		}

		del := tx.Where("mystery_id = ?", mysteryID)
		if len(keepIDs) > 0 {
			del = del.Where("beat_id NOT IN ?", keepIDs)
		}
		if err := del.Delete(&models.VampireMysteryBeatLink{}).Error; err != nil {
			return err
		}

		for i := range beats {
			if beats[i].ID == uuid.Nil {
				nb := models.VampireMysteryBeat{Title: beats[i].Title, Description: beats[i].Description}
				if err := tx.Create(&nb).Error; err != nil {
					return err
				}
				beats[i].ID = nb.ID
			} else {
				if err := tx.Model(&models.VampireMysteryBeat{}).
					Where("id = ?", beats[i].ID).
					Updates(map[string]interface{}{
						"title":       beats[i].Title,
						"description": beats[i].Description,
					}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where(models.VampireMysteryBeatLink{MysteryID: mysteryID, BeatID: beats[i].ID}).
				Assign(models.VampireMysteryBeatLink{Ordinal: beats[i].Ordinal}).
				FirstOrCreate(&models.VampireMysteryBeatLink{}).Error; err != nil {
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
