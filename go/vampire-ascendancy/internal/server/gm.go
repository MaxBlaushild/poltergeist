package server

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// logGM writes an audit entry for the instance in the URL, attributed to the
// signed-in admin. Failures to log are swallowed so they never block the
// action the admin was performing.
func (s *server) logGM(ctx *gin.Context, action string, payload map[string]interface{}) {
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}
	_ = s.dbClient.Vampire().LogGMAction(ctx, instanceIDFromContext(ctx), gmNameFromContext(ctx), action, raw)
}

// GET /gm/state — current game state for the admin controls.
func (s *server) gmGetState(ctx *gin.Context) {
	state, err := s.dbClient.Vampire().GetGameState(ctx, instanceIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gameStateResponse(state))
}

// POST /gm/unlock — the master switch that reveals post-Act-1 content, secrets,
// and missions to all players.
func (s *server) gmSetUnlock(ctx *gin.Context) {
	var body struct {
		Unlocked bool `json:"unlocked"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state, err := s.dbClient.Vampire().UpdateGameState(ctx, instanceIDFromContext(ctx), map[string]interface{}{
		"content_unlocked": body.Unlocked,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "set_unlock", map[string]interface{}{"unlocked": body.Unlocked})
	ctx.JSON(http.StatusOK, gameStateResponse(state))
}

// POST /gm/reset — wipe this instance's play progress for a clean playtest.
// Requires an explicit confirmation so it can't be triggered by accident.
func (s *server) gmResetGame(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	var body struct {
		Confirm string `json:"confirm"`
		Force   bool   `json:"force"`
	}
	_ = ctx.ShouldBindJSON(&body)
	if body.Confirm != "RESET" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "confirmation required"})
		return
	}

	// Live-lock: once the game is past pre-event, refuse a reset unless the caller
	// explicitly forces it. This stops a live game's scores from being wiped by an
	// accidental click; a deliberate reset must roll back to pre-event or pass force.
	// (Either way, ResetGameProgress archives the score ledgers first.)
	state, err := s.dbClient.Vampire().GetGameState(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if state.CurrentAct != "pre_event" && !body.Force {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "the game is live (" + state.CurrentAct + "); reset is locked. Roll back to pre-event or pass force to override.",
		})
		return
	}

	if err := s.dbClient.Vampire().ResetGameProgress(ctx, instanceID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// The reset cleared the audit log; record that the reset happened.
	s.logGM(ctx, "reset_game", map[string]interface{}{"forced": body.Force, "fromAct": state.CurrentAct})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

var validActs = map[string]bool{
	"pre_event": true, "act1": true, "act2": true, "act3": true,
	"quiz": true, "quiz_part1": true, "quiz_part2": true, "resolved": true,
}

// POST /gm/act — advance the night (act1 -> act2 -> act3 -> quiz -> resolved).
func (s *server) gmSetAct(ctx *gin.Context) {
	var body struct {
		Act string `json:"act"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validActs[body.Act] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid act"})
		return
	}

	state, err := s.dbClient.Vampire().UpdateGameState(ctx, instanceIDFromContext(ctx), map[string]interface{}{
		"current_act": body.Act,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Item House Favor is now a live "+X" overlay on the standings (see
	// houseItemFavor), so nothing needs to be folded into the ledger at reveal.

	s.logGM(ctx, "set_act", map[string]interface{}{"act": body.Act})
	ctx.JSON(http.StatusOK, gameStateResponse(state))
}
