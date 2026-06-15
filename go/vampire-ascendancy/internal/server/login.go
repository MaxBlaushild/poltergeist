package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /characters — public roster for the "select your character" dropdown.
// Names only, no content, no sigils. Only standard playable characters appear.
func (s *server) listCharactersPublic(ctx *gin.Context) {
	chars, err := s.dbClient.Vampire().ListCharacters(ctx)
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

// GET /characters/:id — public name/house for one character, for the confirm
// screen reached from a /c/<characterId> link. No content, no sigil.
func (s *server) getCharacterPublic(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	c, err := s.dbClient.Vampire().GetCharacterByID(ctx, id)
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

// POST /login — { characterId, password }. Validates the sigil for that
// character and returns the session token of the active player holding the seat.
// This is the only way to obtain a token; the link in the URL grants nothing.
func (s *server) login(ctx *gin.Context) {
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

	character, err := s.dbClient.Vampire().GetCharacterByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if character == nil || character.RoleType != "player" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if character.Password == "" || body.Password != character.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect sigil"})
		return
	}

	player, err := s.dbClient.Vampire().GetActivePlayerByCharacterID(ctx, id)
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
