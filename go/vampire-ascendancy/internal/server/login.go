package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /i/:instanceId/characters — public roster for the "select your
// character" dropdown, scoped to this Toast's included roster. Names only,
// no content, no sigils.
func (s *server) listCharactersPublic(ctx *gin.Context) {
	instanceID, err := instanceIDParam(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	chars, err := s.dbClient.Vampire().ListIncludedCharacters(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(chars))
	for _, c := range chars {
		if c.RoleType != "player" {
			continue
		}
		row := gin.H{"id": c.ID, "name": c.Name, "title": c.Title}
		if c.House != nil {
			row["house"] = c.House.Name
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"characters": out})
}

// GET /i/:instanceId/characters/:id — public name/house for one character in
// this Toast, for the confirm screen reached from a /e/<instanceId>/c/<characterId>
// link. No content, no sigil.
func (s *server) getCharacterPublic(ctx *gin.Context) {
	instanceID, err := instanceIDParam(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	c, err := s.dbClient.Vampire().GetIncludedCharacterByID(ctx, instanceID, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if c == nil || c.RoleType != "player" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	resp := gin.H{"id": c.ID, "name": c.Name, "title": c.Title}
	if c.House != nil {
		resp["house"] = c.House.Name
	}
	ctx.JSON(http.StatusOK, resp)
}

// POST /i/:instanceId/login — { characterId, password }. Validates the
// sigil for that character within this Toast and returns the session token
// of the active player holding the seat. This is the only way to obtain a
// token; the link in the URL grants nothing.
func (s *server) login(ctx *gin.Context) {
	instanceID, err := instanceIDParam(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	var body struct {
		CharacterID string `json:"characterId"`
		Password    string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(body.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}

	v := s.dbClient.Vampire()
	character, err := v.GetIncludedCharacterByID(ctx, instanceID, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if character == nil || character.RoleType != "player" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	ic, err := v.GetInstanceCharacter(ctx, instanceID, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ic == nil || ic.Sigil == "" || body.Password != ic.Sigil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect sigil"})
		return
	}

	player, err := v.GetActivePlayerByCharacterID(ctx, instanceID, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if player == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "no active seat for this character"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": player.Token})
}
