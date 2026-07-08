package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /gm/characters/:id — a character's full editable content (bios, secrets,
// missions, portrait) plus the real player name from its active slot.
func (s *server) gmGetCharacter(ctx *gin.Context) {
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

	playerName := ""
	if slot, _ := v.GetActivePlayerByCharacterID(ctx, id); slot != nil {
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
		"imageUrl":        c.ImageURL,
		"sigil":           c.Password,
		"playerName":      playerName,
		"secrets":         secrets,
		"missions":        missions,
	})
}

// PUT /gm/characters/:id — save the character editor: core fields, secrets and
// missions (replaced wholesale), and the player name on its active slot.
func (s *server) gmUpdateCharacter(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}

	var body struct {
		Name            string  `json:"name"`
		Title           string  `json:"title"`
		RoleType        string  `json:"roleType"`
		HouseID         *string `json:"houseId"`
		PreEventInfo    string  `json:"preEventInfo"`
		PostAct1Context string  `json:"postAct1Context"`
		ImageURL        string  `json:"imageUrl"`
		PlayerName      string  `json:"playerName"`
		Secrets         []string `json:"secrets"`
		Missions        []struct {
			Tier         string `json:"tier"`
			RewardBt     int    `json:"rewardBt"`
			Prompt       string `json:"prompt"`
			AnswerFormat string `json:"answerFormat"`
		} `json:"missions"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	v := s.dbClient.Vampire()
	fields := map[string]interface{}{
		"name":              body.Name,
		"title":             body.Title,
		"role_type":         body.RoleType,
		"pre_event_info":    body.PreEventInfo,
		"post_act1_context": body.PostAct1Context,
		"image_url":         body.ImageURL,
	}
	if body.HouseID != nil {
		if *body.HouseID == "" {
			fields["house_id"] = nil
		} else {
			hid, err := uuid.Parse(*body.HouseID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
				return
			}
			fields["house_id"] = hid
		}
	}
	if err := v.UpdateCharacter(ctx, id, fields); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	secrets := make([]models.VampireSecret, 0, len(body.Secrets))
	for _, b := range body.Secrets {
		if strings.TrimSpace(b) == "" {
			continue
		}
		secrets = append(secrets, models.VampireSecret{Ordinal: len(secrets) + 1, Body: b})
	}
	if err := v.ReplaceSecrets(ctx, id, secrets); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	missions := make([]models.VampireMission, 0, len(body.Missions))
	for _, m := range body.Missions {
		if strings.TrimSpace(m.Prompt) == "" {
			continue
		}
		tier := m.Tier
		if tier == "" {
			tier = "easy"
		}
		missions = append(missions, models.VampireMission{
			Ordinal:      len(missions) + 1,
			Tier:         tier,
			RewardBT:     m.RewardBt,
			Prompt:       m.Prompt,
			AnswerFormat: m.AnswerFormat,
		})
	}
	if err := v.ReplaceMissions(ctx, id, missions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Player name lives on the character's active login slot.
	if slot, _ := v.GetActivePlayerByCharacterID(ctx, id); slot != nil && body.PlayerName != slot.GuestLabel {
		if err := v.UpdatePlayerAssignment(ctx, slot.ID, slot.CharacterID, body.PlayerName, slot.Active); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	s.logGM(ctx, "update_character", map[string]interface{}{"characterId": id.String(), "name": body.Name})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// PUT /gm/houses/:id — edit a house's tagline.
func (s *server) gmUpdateHouse(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
		return
	}
	var body struct {
		Tagline string `json:"tagline"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Vampire().UpdateHouseTagline(ctx, id, body.Tagline); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "update_house", map[string]interface{}{"houseId": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
