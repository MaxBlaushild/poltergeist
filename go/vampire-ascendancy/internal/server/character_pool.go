package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// The character pool: which of this instance's mystery-eligible characters
// (see gmListCharacters) a Host has made selectable by players. Curated
// separately from sending invites — an invite no longer names a character,
// so a Host decides "who's eligible to be chosen" here instead, and the
// Invites tab shows the pool's size next to how many people have been
// invited (see gmListInvites' counts on the frontend).

// GET /gm/character-pool — the current pool, as a plain set of character
// ids. Pair with GET /gm/characters (the full mystery-eligible candidate
// list, tags and all) to render the picker.
func (s *server) gmGetCharacterPool(ctx *gin.Context) {
	ids, err := s.dbClient.Vampire().ListCharacterPoolIDs(ctx, instanceIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]uuid.UUID, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	ctx.JSON(http.StatusOK, gin.H{"characterIds": out})
}

// PUT /gm/character-pool { characterIds } — wholesale-replace the pool.
// Only mystery-eligible characters are accepted (defense-in-depth mirror
// of gmListCharacters' own filtering — the frontend picker only offers
// eligible ones in the first place).
func (s *server) gmSetCharacterPool(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	var body struct {
		CharacterIDs []string `json:"characterIds"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v := s.dbClient.Vampire()
	eligible, err := v.ListCharacterIDsWithSecretsForMystery(ctx, mysteryIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ids := make([]uuid.UUID, 0, len(body.CharacterIDs))
	for _, raw := range body.CharacterIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
			return
		}
		if !eligible[id] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "a character in the pool must have secrets written for this Toast's mystery"})
			return
		}
		ids = append(ids, id)
	}

	if err := v.SetCharacterPool(ctx, instanceID, ids); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "set_character_pool", map[string]interface{}{"count": len(ids)})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
