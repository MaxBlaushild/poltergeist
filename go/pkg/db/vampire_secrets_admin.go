package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

// Cross-cutting secret views for the Super Admin Secrets tab — view every
// secret in the system in one place (across every character and every
// mystery/subplot), instead of walking the cast per-mystery
// (superadmin_mysteries.go) or per-character (superadmin.go/
// SuperAdminCharacters.tsx). Both of those write through the same
// character-mystery-scoped methods already in vampire_mystery.go; this
// file only adds read/create surface for the flat, sortable/filterable
// list and its "pick a character + a beat, type the text" creation form.

// SecretRow is one secret plus enough denormalized context (character
// name, mystery/subplot name, beat title) to render, sort, and filter it
// without an N+1 lookup per row.
type SecretRow struct {
	ID            uuid.UUID
	CharacterID   uuid.UUID
	CharacterName string
	MysteryID     uuid.UUID
	MysteryName   string
	IsSubplot     bool
	BeatID        *uuid.UUID
	BeatTitle     string
	Body          string
	Ordinal       int
}

// ListAllSecrets returns every secret in the system, newest first, with
// its character/mystery/beat names resolved in four queries total
// regardless of how many secrets there are (one for the secrets, three
// bulk lookups), not one query per secret.
func (h *vampireHandler) ListAllSecrets(ctx context.Context) ([]SecretRow, error) {
	var secrets []models.VampireSecret
	if err := h.db.WithContext(ctx).Order("created_at DESC").Find(&secrets).Error; err != nil {
		return nil, err
	}
	if len(secrets) == 0 {
		return []SecretRow{}, nil
	}

	charIDSet := make(map[uuid.UUID]bool)
	mysteryIDSet := make(map[uuid.UUID]bool)
	beatIDSet := make(map[uuid.UUID]bool)
	for _, s := range secrets {
		charIDSet[s.CharacterID] = true
		if s.MysteryID != nil {
			mysteryIDSet[*s.MysteryID] = true
		}
		if s.BeatID != nil {
			beatIDSet[*s.BeatID] = true
		}
	}

	charNameByID, err := h.namesByID(ctx, "vampire_characters", uuidSetToSlice(charIDSet))
	if err != nil {
		return nil, err
	}

	var mysteries []models.VampireMystery
	if len(mysteryIDSet) > 0 {
		if err := h.db.WithContext(ctx).Where("id IN ?", uuidSetToSlice(mysteryIDSet)).Find(&mysteries).Error; err != nil {
			return nil, err
		}
	}
	mysteryByID := make(map[uuid.UUID]models.VampireMystery, len(mysteries))
	for _, m := range mysteries {
		mysteryByID[m.ID] = m
	}

	var beats []models.VampireMysteryBeat
	if len(beatIDSet) > 0 {
		if err := h.db.WithContext(ctx).Where("id IN ?", uuidSetToSlice(beatIDSet)).Find(&beats).Error; err != nil {
			return nil, err
		}
	}
	beatTitleByID := make(map[uuid.UUID]string, len(beats))
	for _, b := range beats {
		beatTitleByID[b.ID] = b.Title
	}

	out := make([]SecretRow, 0, len(secrets))
	for _, s := range secrets {
		row := SecretRow{
			ID:            s.ID,
			CharacterID:   s.CharacterID,
			CharacterName: charNameByID[s.CharacterID],
			Body:          s.Body,
			Ordinal:       s.Ordinal,
		}
		if s.MysteryID != nil {
			row.MysteryID = *s.MysteryID
			if m, ok := mysteryByID[*s.MysteryID]; ok {
				row.MysteryName = m.Name
				row.IsSubplot = m.IsSubplot
			}
		}
		if s.BeatID != nil {
			row.BeatID = s.BeatID
			row.BeatTitle = beatTitleByID[*s.BeatID]
		}
		out = append(out, row)
	}
	return out, nil
}

// MysteryBeatOption is one (mystery, beat) pairing — the Secrets tab's
// "create a new secret" form picks a beat this way (not a bare beat id)
// because a beat can be shared across several mysteries/subplots (see
// ReplaceMysteryBeats) and a secret must belong to exactly one story; a
// shared beat appears once per mystery it's attached to.
type MysteryBeatOption struct {
	MysteryID   uuid.UUID
	MysteryName string
	IsSubplot   bool
	BeatID      uuid.UUID
	BeatTitle   string
}

// ListAllMysteryBeatOptions returns every (mystery, beat) link in the
// system with display names resolved — the Secrets tab's beat picker.
func (h *vampireHandler) ListAllMysteryBeatOptions(ctx context.Context) ([]MysteryBeatOption, error) {
	var links []models.VampireMysteryBeatLink
	if err := h.db.WithContext(ctx).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []MysteryBeatOption{}, nil
	}

	mysteryIDSet := make(map[uuid.UUID]bool, len(links))
	beatIDSet := make(map[uuid.UUID]bool, len(links))
	for _, l := range links {
		mysteryIDSet[l.MysteryID] = true
		beatIDSet[l.BeatID] = true
	}

	var mysteries []models.VampireMystery
	if err := h.db.WithContext(ctx).Where("id IN ?", uuidSetToSlice(mysteryIDSet)).Find(&mysteries).Error; err != nil {
		return nil, err
	}
	mysteryByID := make(map[uuid.UUID]models.VampireMystery, len(mysteries))
	for _, m := range mysteries {
		mysteryByID[m.ID] = m
	}

	var beats []models.VampireMysteryBeat
	if err := h.db.WithContext(ctx).Where("id IN ?", uuidSetToSlice(beatIDSet)).Find(&beats).Error; err != nil {
		return nil, err
	}
	beatByID := make(map[uuid.UUID]models.VampireMysteryBeat, len(beats))
	for _, b := range beats {
		beatByID[b.ID] = b
	}

	out := make([]MysteryBeatOption, 0, len(links))
	for _, l := range links {
		m, mok := mysteryByID[l.MysteryID]
		b, bok := beatByID[l.BeatID]
		if !mok || !bok {
			continue
		}
		out = append(out, MysteryBeatOption{
			MysteryID: m.ID, MysteryName: m.Name, IsSubplot: m.IsSubplot,
			BeatID: b.ID, BeatTitle: b.Title,
		})
	}
	return out, nil
}

// namesByID is a small helper for bulk-resolving a "name" column for a set
// of ids in one query — used for characters here; mysteries/beats above do
// their own lookup since they also need more than just the name
// (isSubplot, etc.).
func (h *vampireHandler) namesByID(ctx context.Context, table string, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	out := make(map[uuid.UUID]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ID   uuid.UUID
		Name string
	}
	if err := h.db.WithContext(ctx).Table(table).Select("id, name").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.ID] = r.Name
	}
	return out, nil
}

func uuidSetToSlice(set map[uuid.UUID]bool) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}
