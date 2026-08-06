package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /gm/characters/:id — a character's full content (bios, secrets,
// missions — shared, edited only by super users, see GET
// /admin/characters/:id) plus this Toast's portrait/sigil and the real
// player name from its active slot, both of which ARE per-instance and
// editable here.
func (s *server) gmGetCharacter(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	v := s.dbClient.Vampire()
	c, err := v.GetCharacterByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if c == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	ic, err := v.GetInstanceCharacter(ctx, instanceID, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sigil, imageURL := "", ""
	if ic != nil {
		sigil, imageURL = ic.Sigil, ic.ImageURL
	}

	playerName := ""
	if slot, _ := v.GetActivePlayerByCharacterID(ctx, instanceID, id); slot != nil {
		playerName = slot.GuestLabel
	}

	secrets := make([]gin.H, 0, len(c.Secrets))
	for _, sec := range c.Secrets {
		secrets = append(secrets, gin.H{"ordinal": sec.Ordinal, "body": sec.Body})
	}
	missions := make([]gin.H, 0, len(c.Missions))
	for _, m := range c.Missions {
		missions = append(missions, gin.H{
			"ordinal":      m.Ordinal,
			"tier":         m.Tier,
			"rewardBt":     m.RewardBT,
			"prompt":       m.Prompt,
			"answerFormat": m.AnswerFormat,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":              c.ID,
		"name":            c.Name,
		"title":           c.Title,
		"roleType":        c.RoleType,
		"isOptional":      c.IsOptional,
		"houseId":         c.HouseID,
		"preEventInfo":    c.PreEventInfo,
		"postAct1Context": c.PostAct1Context,
		"imageUrl":        imageURL,
		"sigil":           sigil,
		"playerName":      playerName,
		"secrets":         secrets,
		"missions":        missions,
	})
}

// PUT /gm/characters/:id/portrait — set this Toast's portrait for a
// character. The only character edit a Host/Co-Host can make directly; bios,
// secrets, and missions are shared content edited only by super users (see
// PUT /admin/characters/:id).
func (s *server) gmSetCharacterPortrait(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}
	var body struct {
		ImageURL string `json:"imageUrl"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Vampire().SetInstanceCharacterImageURL(ctx, instanceID, id, body.ImageURL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "set_character_portrait", map[string]interface{}{"characterId": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
