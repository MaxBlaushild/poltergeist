package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// The Secrets tab: a flat, cross-cutting view of every secret in the
// system — across every character and every mystery/subplot — instead of
// walking the cast per-mystery (superadmin_mysteries.go's
// CharacterContentEditor) or per-mystery per-character
// (SuperAdminCharacters.tsx's CharacterSecretsByMystery). Sorting/
// filtering happens client-side against this one list; creation reuses
// the same CreateSecretForCharacterMystery every other secret-authoring
// surface already goes through.

// GET /admin/secrets — every secret, with its character/mystery/beat
// names resolved so the frontend can render, sort, and filter without a
// lookup per row.
func (s *server) adminListAllSecrets(ctx *gin.Context) {
	rows, err := s.dbClient.Vampire().ListAllSecrets(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":            r.ID,
			"characterId":   r.CharacterID,
			"characterName": r.CharacterName,
			"mysteryId":     r.MysteryID,
			"mysteryName":   r.MysteryName,
			"isSubplot":     r.IsSubplot,
			"beatId":        r.BeatID,
			"beatTitle":     r.BeatTitle,
			"body":          r.Body,
			"ordinal":       r.Ordinal,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"secrets": out})
}

// GET /admin/mystery-beat-options — every (mystery, beat) pairing in the
// system, for the "create a new secret" form's beat picker. A beat shared
// across several mysteries/subplots appears once per mystery it's
// attached to, since a secret must belong to exactly one story.
func (s *server) adminListMysteryBeatOptions(ctx *gin.Context) {
	options, err := s.dbClient.Vampire().ListAllMysteryBeatOptions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(options))
	for _, o := range options {
		out = append(out, gin.H{
			"mysteryId":   o.MysteryID,
			"mysteryName": o.MysteryName,
			"isSubplot":   o.IsSubplot,
			"beatId":      o.BeatID,
			"beatTitle":   o.BeatTitle,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"options": out})
}

// POST /admin/secrets { characterId, mysteryId, beatId, body } — the
// Secrets tab's creation form: pick who, pick which (mystery, beat), type
// what they know. Same underlying write as the mystery-first and
// character-first editors (CreateSecretForCharacterMystery); this is just
// a third way to reach it. This is also what makes the character eligible
// for invites to an instance running this mystery, same as adding a
// secret from either of the other two editors.
func (s *server) adminCreateSecretFlat(ctx *gin.Context) {
	var body struct {
		CharacterID string `json:"characterId"`
		MysteryID   string `json:"mysteryId"`
		BeatID      string `json:"beatId"`
		Body        string `json:"body"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	characterID, err := uuid.Parse(body.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "choose a character"})
		return
	}
	mysteryID, err := uuid.Parse(body.MysteryID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "choose a beat"})
		return
	}
	var beatID *uuid.UUID
	if strings.TrimSpace(body.BeatID) != "" {
		bid, err := uuid.Parse(body.BeatID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid beat id"})
			return
		}
		beatID = &bid
	}
	if strings.TrimSpace(body.Body) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "secret text is required"})
		return
	}

	sec, err := s.dbClient.Vampire().CreateSecretForCharacterMystery(ctx, characterID, mysteryID, beatID, body.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logSuperUser(ctx, "create_secret", map[string]interface{}{
		"characterId": characterID.String(), "mysteryId": mysteryID.String(),
	})
	ctx.JSON(http.StatusOK, gin.H{"id": sec.ID})
}
