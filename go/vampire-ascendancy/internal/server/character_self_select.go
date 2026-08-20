package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// The player-facing side of the character pool: once an invite is
// accepted, the player has a VampirePlayer row with no character yet (see
// player_invites.go's acceptPlayerInvite). This is where they browse the
// Host's curated pool (see character_pool.go) — filtered, tag and all,
// same CharacterBrowser UI the Host uses to curate it — and claim one.

// GET /i/:instanceId/selectable-characters — the pool, minus whoever's
// already claimed a character, with enough detail (bio, tags, house) for
// the tag-filtered browser. Same shape as gmListCharacters' rows.
func (s *server) listSelectableCharacters(ctx *gin.Context) {
	player := playerFromContext(ctx)
	v := s.dbClient.Vampire()

	poolIDs, err := v.ListCharacterPoolIDs(ctx, player.InstanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	players, err := v.ListPlayers(ctx, player.InstanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	taken := make(map[uuid.UUID]bool, len(players))
	for _, p := range players {
		if p.Active && p.CharacterID != nil {
			taken[*p.CharacterID] = true
		}
	}
	chars, err := v.ListCharacters(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(poolIDs))
	for _, c := range chars {
		if !poolIDs[c.ID] || taken[c.ID] {
			continue
		}
		tags := []string(c.Tags)
		if tags == nil {
			tags = []string{}
		}
		row := gin.H{
			"id":           c.ID,
			"name":         c.Name,
			"title":        c.Title,
			"preEventInfo": c.PreEventInfo,
			"tags":         tags,
		}
		if c.House != nil {
			row["house"] = c.House.Name
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"characters": out})
}

// POST /i/:instanceId/me/character { characterId } — claim a character
// from the pool. One-shot: once set, it can't be re-picked through this
// endpoint (a Host can still reassign via PUT /gm/players/:id if needed).
func (s *server) chooseCharacter(ctx *gin.Context) {
	player := playerFromContext(ctx)
	var body struct {
		CharacterID string `json:"characterId"`
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
	if err := s.dbClient.Vampire().ClaimCharacterForPlayer(ctx, player.InstanceID, player.ID, characterID); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
