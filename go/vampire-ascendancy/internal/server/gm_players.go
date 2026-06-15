package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func genPlayerToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GET /gm/players — every player slot with its assignment and current (recorded)
// Blood Token total, for the assignment editor and BT-award panel.
func (s *server) gmListPlayers(ctx *gin.Context) {
	players, err := s.dbClient.Vampire().ListPlayers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totals, err := s.dbClient.Vampire().BloodTokenTotalsByPlayer(ctx)
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
			"token":      p.Token,
			"guestLabel": p.GuestLabel,
			"active":     p.Active,
			"btTotal":    btByPlayer[p.ID.String()],
		}
		if p.Character != nil {
			ch := gin.H{
				"id":       p.Character.ID,
				"name":     p.Character.Name,
				"roleType": p.Character.RoleType,
				"sigil":    p.Character.Password, // GM needs this to hand out
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

// POST /gm/players — create a new player slot with a fresh token.
func (s *server) gmCreatePlayer(ctx *gin.Context) {
	var body struct {
		GuestLabel  string  `json:"guestLabel"`
		CharacterID *string `json:"characterId"`
	}
	_ = ctx.ShouldBindJSON(&body)

	token, err := genPlayerToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	player := &models.VampirePlayer{
		Token:       token,
		CharacterID: characterID,
		GuestLabel:  body.GuestLabel,
		Active:      true,
	}
	if err := s.dbClient.Vampire().CreatePlayer(ctx, player); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "create_player", map[string]interface{}{"playerId": player.ID.String(), "guestLabel": body.GuestLabel})
	ctx.JSON(http.StatusOK, gin.H{"id": player.ID, "token": player.Token})
}

// PUT /gm/players/:id — edit a player's character assignment, label, or active flag.
func (s *server) gmUpdatePlayer(ctx *gin.Context) {
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

	if err := s.dbClient.Vampire().UpdatePlayerAssignment(ctx, playerID, characterID, body.GuestLabel, body.Active); err != nil {
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

// GET /gm/characters — character roster for the assignment dropdown.
func (s *server) gmListCharacters(ctx *gin.Context) {
	chars, err := s.dbClient.Vampire().ListCharacters(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(chars))
	for _, c := range chars {
		row := gin.H{
			"id":         c.ID,
			"name":       c.Name,
			"title":      c.Title,
			"roleType":   c.RoleType,
			"isOptional": c.IsOptional,
		}
		if c.House != nil {
			row["house"] = c.House.Name
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"characters": out})
}
