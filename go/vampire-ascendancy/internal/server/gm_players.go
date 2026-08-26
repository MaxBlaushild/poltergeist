package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /gm/players — the accepted roster: every player who has accepted
// their invite, with their current Blood Token total. Pending/declined
// invites live on the Invites tab instead (see player_invites.go).
func (s *server) gmListPlayers(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	v := s.dbClient.Vampire()
	players, err := v.ListPlayers(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totals, err := v.BloodTokenTotalsByPlayer(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	btByPlayer := map[string]int{}
	for _, t := range totals {
		btByPlayer[t.PlayerID.String()] = t.Total
	}

	out := make([]gin.H, 0, len(players))
	for _, p := range players {
		row := gin.H{
			"id":         p.ID,
			"guestLabel": p.GuestLabel,
			"active":     p.Active,
			"btTotal":    btByPlayer[p.ID.String()],
		}
		if p.Character != nil {
			ch := gin.H{
				"id":       p.Character.ID,
				"name":     p.Character.Name,
				"roleType": p.Character.RoleType,
			}
			if p.Character.House != nil {
				ch["house"] = p.Character.House.Name
			}
			row["character"] = ch
		} else {
			row["character"] = nil
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"players": out})
}

// PUT /gm/players/:id — admin correction to an already-accepted player: the
// guest label, active flag, or (rarely) reassign their character. Creating
// a player is now only done by accepting an invite (see player_invites.go).
func (s *server) gmUpdatePlayer(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	playerID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	var body struct {
		CharacterID *string `json:"characterId"`
		GuestLabel  string  `json:"guestLabel"`
		Active      bool    `json:"active"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var characterID *uuid.UUID
	if body.CharacterID != nil && *body.CharacterID != "" {
		id, err := uuid.Parse(*body.CharacterID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
			return
		}
		characterID = &id
	}

	if err := s.dbClient.Vampire().UpdatePlayerAssignment(ctx, instanceID, playerID, characterID, body.GuestLabel, body.Active); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "update_player", map[string]interface{}{
		"playerId":    playerID.String(),
		"characterId": body.CharacterID,
		"active":      body.Active,
	})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /gm/characters — the shared library's playable-role characters, for
// the Invites tab's "who is this invite for" picker. Not filtered to
// "included" (that concept is retired — see MULTI_TENANT_REQUIREMENTS.md);
// the frontend cross-references the current roster + pending invites to
// grey out ones already spoken for. Also filtered to characters with at
// least one secret authored for this instance's mystery — a character
// with none can't be invited (see MYSTERY_REQUIREMENTS.md's eligibility
// gating; CreatePlayerInvite enforces the same rule server-side as
// defense-in-depth, this filtering is just so the picker doesn't offer
// them in the first place). Carries bio and tags so the picker can
// search/filter/preview without a second round trip per character.
func (s *server) gmListCharacters(ctx *gin.Context) {
	v := s.dbClient.Vampire()
	chars, err := v.ListCharacters(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	hasSecrets, err := v.ListCharacterIDsWithSecretsForMystery(ctx, mysteryIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(chars))
	for _, c := range chars {
		if c.RoleType != "player" {
			continue
		}
		if !hasSecrets[c.ID] {
			continue
		}
		tags := []string(c.Tags)
		if tags == nil {
			tags = []string{}
		}
		row := gin.H{
			"id":         c.ID,
			"name":       c.Name,
			"title":      c.Title,
			"roleType":   c.RoleType,
			"isOptional": c.IsOptional,
			"bio":        c.Bio,
			"tags":       tags,
		}
		if c.House != nil {
			row["house"] = c.House.Name
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"characters": out})
}
