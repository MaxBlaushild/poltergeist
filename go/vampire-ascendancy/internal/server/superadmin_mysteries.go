package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Mysteries — the underlying story an instance's players are solving. See
// MYSTERY_REQUIREMENTS.md. Shared content, super-user-only, same
// authorization posture as the rest of this file.

// GET /admin/mysteries — every mystery, active and inactive (the "Host a
// Toast" picker filters to active itself; this list shows both so a super
// user can revive one).
func (s *server) adminListMysteries(ctx *gin.Context) {
	mysteries, err := s.dbClient.Vampire().ListMysteries(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(mysteries))
	for _, m := range mysteries {
		out = append(out, gin.H{
			"id":        m.ID,
			"name":      m.Name,
			"summary":   m.Summary,
			"active":    m.Active,
			"beatCount": len(m.Beats),
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"mysteries": out})
}

// POST /admin/mysteries — create a new mystery, empty apart from a name.
func (s *server) adminCreateMystery(ctx *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "give this mystery a name"})
		return
	}
	m, err := s.dbClient.Vampire().CreateMystery(ctx, name, "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "create_mystery", map[string]interface{}{"id": m.ID.String(), "name": name})
	ctx.JSON(http.StatusOK, gin.H{"id": m.ID, "name": m.Name})
}

// GET /admin/mysteries/:id — full editor payload: name, summary, full
// lore, active, and the ordered beat list.
func (s *server) adminGetMystery(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	m, err := s.dbClient.Vampire().GetMysteryByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if m == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "mystery not found"})
		return
	}
	beats := make([]gin.H, 0, len(m.Beats))
	for _, b := range m.Beats {
		beats = append(beats, gin.H{"id": b.ID, "ordinal": b.Ordinal, "body": b.Body})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":       m.ID,
		"name":     m.Name,
		"summary":  m.Summary,
		"fullLore": m.FullLore,
		"active":   m.Active,
		"beats":    beats,
	})
}

// PUT /admin/mysteries/:id — save the mystery editor: core fields plus the
// beat list (replaced wholesale, same pattern secrets/missions already
// use). Removing a beat that a secret still points at un-sets that
// secret's beat rather than blocking the save (ON DELETE SET NULL) — the
// frontend warns before that happens, but doesn't hard-block it.
func (s *server) adminUpdateMystery(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	var body struct {
		Name     string `json:"name"`
		Summary  string `json:"summary"`
		FullLore string `json:"fullLore"`
		Active   bool   `json:"active"`
		Beats    []struct {
			// ID is empty for a beat being created; present for one being
			// edited. Preserving it is what keeps a beat's id — and any
			// secret's beat_id pointing at it — stable across saves (see
			// ReplaceMysteryBeats).
			ID   string `json:"id"`
			Body string `json:"body"`
		} `json:"beats"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	v := s.dbClient.Vampire()
	if err := v.UpdateMystery(ctx, id, map[string]interface{}{
		"name":      name,
		"summary":   body.Summary,
		"full_lore": body.FullLore,
		"active":    body.Active,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	beats := make([]models.VampireMysteryBeat, 0, len(body.Beats))
	for i, b := range body.Beats {
		if strings.TrimSpace(b.Body) == "" {
			continue
		}
		beat := models.VampireMysteryBeat{Ordinal: i + 1, Body: b.Body}
		if b.ID != "" {
			bid, err := uuid.Parse(b.ID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid beat id"})
				return
			}
			beat.ID = bid
		}
		beats = append(beats, beat)
	}
	if err := v.ReplaceMysteryBeats(ctx, id, beats); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logSuperUser(ctx, "update_mystery", map[string]interface{}{"id": id.String(), "beatCount": len(beats)})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Per-character content, scoped to a mystery: secrets + post-Act-1
// context ----
// Deliberately reached through the mystery, not the character (see
// MYSTERY_REQUIREMENTS.md's "Super Admin UI" section) — authoring a
// mystery means walking its cast and deciding what each of them knows, and
// what happens to them, in that mystery specifically.

// GET /admin/mysteries/:id/characters/:characterId/content
func (s *server) adminGetCharacterContentForMystery(ctx *gin.Context) {
	mysteryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	characterID, err := uuid.Parse(ctx.Param("characterId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	v := s.dbClient.Vampire()
	secrets, err := v.ListSecretsForCharacterAndMystery(ctx, characterID, mysteryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	postAct1Context, err := v.GetCharacterMysteryContext(ctx, characterID, mysteryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(secrets))
	for _, sec := range secrets {
		out = append(out, gin.H{"ordinal": sec.Ordinal, "body": sec.Body, "beatId": sec.BeatID})
	}
	ctx.JSON(http.StatusOK, gin.H{"secrets": out, "postAct1Context": postAct1Context})
}

// PUT /admin/mysteries/:id/characters/:characterId/content — replace this
// character's secrets for this mystery wholesale (an empty list makes them
// ineligible for invites — see gm_players.go's gmListCharacters and
// CreatePlayerInvite) and overwrite their post-Act-1 context for this
// mystery.
func (s *server) adminUpdateCharacterContentForMystery(ctx *gin.Context) {
	mysteryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	characterID, err := uuid.Parse(ctx.Param("characterId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	var body struct {
		Secrets []struct {
			Body   string  `json:"body"`
			BeatID *string `json:"beatId"`
		} `json:"secrets"`
		PostAct1Context string `json:"postAct1Context"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secrets := make([]models.VampireSecret, 0, len(body.Secrets))
	for i, sec := range body.Secrets {
		if strings.TrimSpace(sec.Body) == "" {
			continue
		}
		s := models.VampireSecret{Ordinal: i + 1, Body: sec.Body}
		if sec.BeatID != nil && *sec.BeatID != "" {
			bid, err := uuid.Parse(*sec.BeatID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid beat id"})
				return
			}
			s.BeatID = &bid
		}
		secrets = append(secrets, s)
	}

	v := s.dbClient.Vampire()
	if err := v.ReplaceSecretsForCharacterAndMystery(ctx, characterID, mysteryID, secrets); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := v.UpsertCharacterMysteryContext(ctx, characterID, mysteryID, body.PostAct1Context); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "update_character_mystery_content", map[string]interface{}{
		"mysteryId": mysteryID.String(), "characterId": characterID.String(), "secretCount": len(secrets),
	})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Beat-centric secret management ----
// Complements the character-centric editor above — lets a super user, while
// looking at one beat in the Story tab, decide who knows it, instead of
// having to leave and walk the cast one character at a time. See
// ListSecretsForBeat's comment for why this is individual CRUD rather than
// wholesale-replace.

// GET /admin/mysteries/:id/beats/:beatId/secrets
func (s *server) adminListBeatSecrets(ctx *gin.Context) {
	beatID, err := uuid.Parse(ctx.Param("beatId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid beat id"})
		return
	}
	secrets, err := s.dbClient.Vampire().ListSecretsForBeat(ctx, beatID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(secrets))
	for _, sec := range secrets {
		out = append(out, gin.H{"id": sec.ID, "characterId": sec.CharacterID, "body": sec.Body})
	}
	ctx.JSON(http.StatusOK, gin.H{"secrets": out})
}

// POST /admin/mysteries/:id/beats/:beatId/secrets — add one secret for one
// character, tied to this beat. This is also what makes that character
// eligible for invites to an instance running this mystery, same as adding
// a secret from the character-centric editor (see gm_players.go's
// gmListCharacters and CreatePlayerInvite).
func (s *server) adminCreateBeatSecret(ctx *gin.Context) {
	mysteryID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	beatID, err := uuid.Parse(ctx.Param("beatId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid beat id"})
		return
	}
	var body struct {
		CharacterID string `json:"characterId"`
		Body        string `json:"body"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	characterID, err := uuid.Parse(body.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	if strings.TrimSpace(body.Body) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "secret body is required"})
		return
	}
	sec, err := s.dbClient.Vampire().CreateSecretForCharacterMystery(ctx, characterID, mysteryID, &beatID, body.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "create_beat_secret", map[string]interface{}{
		"mysteryId": mysteryID.String(), "beatId": beatID.String(), "characterId": characterID.String(),
	})
	ctx.JSON(http.StatusOK, gin.H{"id": sec.ID})
}

// PUT /admin/secrets/:secretId — edit one secret's text in place. Not
// nested under a mystery/beat in the URL since a secret id is already
// globally unique and this only ever touches that one row.
func (s *server) adminUpdateSecretBody(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("secretId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret id"})
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Body) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "secret body is required"})
		return
	}
	if err := s.dbClient.Vampire().UpdateSecretBody(ctx, id, body.Body); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /admin/secrets/:secretId — the beat panel's "remove" action. If
// this was a character's only secret for the mystery, they become
// ineligible for invites to it again (see CreatePlayerInvite).
func (s *server) adminDeleteSecret(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("secretId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid secret id"})
		return
	}
	if err := s.dbClient.Vampire().DeleteSecret(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
