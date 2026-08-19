package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Mysteries — the underlying story an instance's players are solving. See
// MYSTERY_REQUIREMENTS.md. Shared content, super-user-only, same
// authorization posture as the rest of this file.
//
// Subplots are a sibling of mysteries, not a separate table: same row
// shape (summary/full lore/beats/secrets all work identically), just
// IsSubplot=true. An instance picks its one required mystery plus zero or
// many subplots (see vampire_instance_subplots) — every handler below
// serves both kinds interchangeably; only the frontend splits them into
// separate tabs by isSubplot.

// GET /admin/mysteries — every mystery AND subplot, active and inactive
// (the "Host a Toast" pickers filter to active themselves; this list shows
// both so a super user can revive one). The frontend splits by isSubplot
// into the Mysteries and Sub-plots tabs.
func (s *server) adminListMysteries(ctx *gin.Context) {
	v := s.dbClient.Vampire()
	mysteries, err := v.ListMysteries(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	beatCounts, err := v.CountBeatsByMystery(ctx)
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
			"isSubplot": m.IsSubplot,
			"beatCount": beatCounts[m.ID],
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"mysteries": out})
}

// POST /admin/mysteries — create a new mystery or subplot (same row shape,
// isSubplot picks which), empty apart from a name.
func (s *server) adminCreateMystery(ctx *gin.Context) {
	var body struct {
		Name      string `json:"name"`
		IsSubplot bool   `json:"isSubplot"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "give this a name"})
		return
	}
	m, err := s.dbClient.Vampire().CreateMystery(ctx, name, "", "", body.IsSubplot)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	action := "create_mystery"
	if body.IsSubplot {
		action = "create_subplot"
	}
	s.logSuperUser(ctx, action, map[string]interface{}{"id": m.ID.String(), "name": name})
	ctx.JSON(http.StatusOK, gin.H{"id": m.ID, "name": m.Name})
}

// GET /admin/mysteries/:id — full editor payload: name, summary, full
// lore, active, isSubplot, and the ordered beat list. Same payload whether
// it's a mystery or a subplot.
func (s *server) adminGetMystery(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	v := s.dbClient.Vampire()
	m, err := v.GetMysteryByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if m == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "mystery not found"})
		return
	}
	beatRows, err := v.ListBeatsForMystery(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	beats := make([]gin.H, 0, len(beatRows))
	for _, b := range beatRows {
		beats = append(beats, gin.H{
			"id": b.ID, "ordinal": b.Ordinal, "title": b.Title, "description": b.Description,
			"linkCount": b.LinkCount,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":        m.ID,
		"name":      m.Name,
		"summary":   m.Summary,
		"fullLore":  m.FullLore,
		"active":    m.Active,
		"isSubplot": m.IsSubplot,
		"beats":     beats,
	})
}

// PUT /admin/mysteries/:id — save the mystery editor: core fields plus the
// beat list, reconciled against what's attached here (see
// ReplaceMysteryBeats — beats are shared, reusable content now, so this
// isn't a wholesale replace: an existing beat's id keeps it linked and
// edits its shared content, a missing one gets unlinked without being
// deleted, and it's also how "attach an existing beat" works). Removing a
// beat that a secret still points at un-sets that secret's beat rather
// than blocking the save (ON DELETE SET NULL) — the frontend warns before
// that happens, but doesn't hard-block it.
func (s *server) adminUpdateMystery(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mystery id"})
		return
	}
	var body struct {
		Name      string `json:"name"`
		Summary   string `json:"summary"`
		FullLore  string `json:"fullLore"`
		Active    bool   `json:"active"`
		IsSubplot bool   `json:"isSubplot"`
		Beats     []struct {
			// ID is empty for a beat being created; present for one being
			// edited. Preserving it is what keeps a beat's id — and any
			// secret's beat_id pointing at it — stable across saves (see
			// ReplaceMysteryBeats).
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
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
		"name":       name,
		"summary":    body.Summary,
		"full_lore":  body.FullLore,
		"active":     body.Active,
		"is_subplot": body.IsSubplot,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	beats := make([]db.MysteryBeat, 0, len(body.Beats))
	for i, b := range body.Beats {
		if strings.TrimSpace(b.Title) == "" && strings.TrimSpace(b.Description) == "" {
			continue
		}
		beat := db.MysteryBeat{Ordinal: i + 1, Title: b.Title, Description: b.Description}
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

// GET /admin/beats — every beat, across every mystery/subplot, with how
// many of them each is currently linked to. Powers the Story tab's "attach
// an existing beat" picker, so a super user can reuse a beat that already
// says the same thing (e.g. "the tools to kill a vampire" shared by two
// subplots) instead of duplicating it.
func (s *server) adminListBeats(ctx *gin.Context) {
	beats, err := s.dbClient.Vampire().ListAllBeats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(beats))
	for _, b := range beats {
		out = append(out, gin.H{"id": b.ID, "title": b.Title, "description": b.Description, "linkCount": b.LinkCount})
	}
	ctx.JSON(http.StatusOK, gin.H{"beats": out})
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
	missions, err := v.ListMissionsForCharacterAndMystery(ctx, characterID, mysteryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	postAct1Context, err := v.GetCharacterMysteryContext(ctx, characterID, mysteryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	secretsOut := make([]gin.H, 0, len(secrets))
	for _, sec := range secrets {
		secretsOut = append(secretsOut, gin.H{"ordinal": sec.Ordinal, "body": sec.Body, "beatId": sec.BeatID})
	}
	missionsOut := make([]gin.H, 0, len(missions))
	for _, m := range missions {
		missionsOut = append(missionsOut, gin.H{
			"ordinal":      m.Ordinal,
			"tier":         m.Tier,
			"rewardBt":     m.RewardBT,
			"prompt":       m.Prompt,
			"answerFormat": m.AnswerFormat,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"secrets": secretsOut, "missions": missionsOut, "postAct1Context": postAct1Context})
}

// PUT /admin/mysteries/:id/characters/:characterId/content — replace this
// character's secrets and missions for this mystery wholesale (an empty
// secrets list makes them ineligible for invites — see gm_players.go's
// gmListCharacters and CreatePlayerInvite; missions don't gate invites) and
// overwrite their post-Act-1 context for this mystery.
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
		Missions []struct {
			Tier         string `json:"tier"`
			RewardBt     int    `json:"rewardBt"`
			Prompt       string `json:"prompt"`
			AnswerFormat string `json:"answerFormat"`
		} `json:"missions"`
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

	missions := make([]models.VampireMission, 0, len(body.Missions))
	for _, m := range body.Missions {
		if strings.TrimSpace(m.Prompt) == "" {
			continue
		}
		tier := m.Tier
		if tier == "" {
			tier = "easy"
		}
		missions = append(missions, models.VampireMission{
			Ordinal:      len(missions) + 1,
			Tier:         tier,
			RewardBT:     m.RewardBt,
			Prompt:       m.Prompt,
			AnswerFormat: m.AnswerFormat,
		})
	}

	v := s.dbClient.Vampire()
	if err := v.ReplaceSecretsForCharacterAndMystery(ctx, characterID, mysteryID, secrets); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := v.ReplaceMissionsForCharacterAndMystery(ctx, characterID, mysteryID, missions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := v.UpsertCharacterMysteryContext(ctx, characterID, mysteryID, body.PostAct1Context); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "update_character_mystery_content", map[string]interface{}{
		"mysteryId": mysteryID.String(), "characterId": characterID.String(),
		"secretCount": len(secrets), "missionCount": len(missions),
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
