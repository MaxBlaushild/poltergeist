package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /inventory — the player's owned items, plus the roster to target and
// whether targeting is locked (the quiz has begun).
func (s *server) getInventory(ctx *gin.Context) {
	player := playerFromContext(ctx)
	v := s.dbClient.Vampire()

	state, err := v.GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	locked := state.QuizPart1Open || state.QuizPart2Open

	pis, err := v.ListPlayerItems(ctx, player.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]gin.H, 0, len(pis))
	for _, pi := range pis {
		row := gin.H{"id": pi.ID, "targetPlayerId": pi.TargetPlayerID}
		if pi.Item != nil {
			row["name"] = pi.Item.Name
			row["description"] = pi.Item.Description
			row["effect"] = pi.Item.Effect
			row["targetsPlayer"] = pi.Item.TargetsPlayer
		}
		items = append(items, row)
	}

	// Targetable players — everyone with a character, excluding the viewer.
	players, err := v.ListPlayers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	targets := make([]gin.H, 0, len(players))
	for _, p := range players {
		if p.ID == player.ID || p.Character == nil || p.Character.Name == "" {
			continue
		}
		targets = append(targets, gin.H{"playerId": p.ID, "name": p.Character.Name})
	}

	ctx.JSON(http.StatusOK, gin.H{"items": items, "targets": targets, "locked": locked})
}

// POST /inventory/:id/target — set (or clear) an owned item's target player.
// Refused once the quiz has begun.
func (s *server) setInventoryTarget(ctx *gin.Context) {
	player := playerFromContext(ctx)
	v := s.dbClient.Vampire()

	state, err := v.GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if state.QuizPart1Open || state.QuizPart2Open {
		ctx.JSON(http.StatusConflict, gin.H{"error": "item targets are locked once the quiz begins"})
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	// Ownership check: the item must belong to the requesting player.
	pis, err := v.ListPlayerItems(ctx, player.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	owned := false
	for _, pi := range pis {
		if pi.ID == id {
			owned = true
			break
		}
	}
	if !owned {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	var body struct {
		TargetPlayerID string `json:"targetPlayerId"`
	}
	_ = ctx.ShouldBindJSON(&body)
	var target *uuid.UUID
	if body.TargetPlayerID != "" {
		tid, err := uuid.Parse(body.TargetPlayerID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid target id"})
			return
		}
		target = &tid
	}
	if err := v.SetPlayerItemTarget(ctx, id, target); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
