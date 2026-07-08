package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getMe returns the authenticated player's view. Pre-event info is always
// included; post-Act-1 context, secrets, and missions are only included once a
// GM has flipped content_unlocked. The gating happens here on the server — the
// locked fields are never serialized into the response while content is locked.
func (s *server) getMe(ctx *gin.Context) {
	player := playerFromContext(ctx)

	state, err := s.dbClient.Vampire().GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{
		// guestLabel is intentionally omitted: it holds the real-world player name
		// (a GM-only roster field) and must never reach the player app.
		"player": gin.H{
			"id": player.ID,
		},
		"gameState": gin.H{
			"currentAct":           state.CurrentAct,
			"contentUnlocked":      state.ContentUnlocked,
			"quizPart1Open":        state.QuizPart1Open,
			"quizPart2Open":        state.QuizPart2Open,
			"quizPart1OpenedAt":    state.QuizPart1OpenedAt,
			"activeNotificationId": state.ActiveNotificationID,
		},
	}

	// Attach the active broadcast that applies to this player, if any. The client
	// renders it as a full-screen takeover.
	var houseID *uuid.UUID
	if player.Character != nil {
		houseID = player.Character.HouseID
	}
	notif, err := s.dbClient.Vampire().GetActiveNotificationForPlayer(ctx, player.ID, houseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if notif != nil {
		resp["notification"] = gin.H{
			"id":    notif.ID,
			"title": notif.Title,
			"body":  notif.Body,
		}
	} else {
		resp["notification"] = nil
	}

	// No character assigned yet -> holding screen on the client.
	if player.CharacterID == nil {
		resp["character"] = nil
		ctx.JSON(http.StatusOK, resp)
		return
	}

	character, err := s.dbClient.Vampire().GetCharacterByID(ctx, *player.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if character == nil {
		resp["character"] = nil
		ctx.JSON(http.StatusOK, resp)
		return
	}

	charResp := gin.H{
		"id":           character.ID,
		"name":         character.Name,
		"title":        character.Title,
		"roleType":     character.RoleType,
		"preEventInfo": character.PreEventInfo,
		"imageUrl":     character.ImageURL,
	}
	if character.House != nil {
		charResp["house"] = gin.H{"id": character.House.ID, "name": character.House.Name, "tagline": character.House.Tagline}
	}

	// Gated content — only revealed after the host opens the evening.
	if state.ContentUnlocked {
		charResp["postAct1Context"] = character.PostAct1Context
		charResp["secrets"] = character.Secrets

		// Attach each mission's submission state for this player.
		subs, err := s.dbClient.Vampire().ListSubmissionsForPlayer(ctx, player.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		photoMap, err := s.photoIDsBySubmission(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		subByMission := map[string]gin.H{}
		for _, sub := range subs {
			ids := photoMap[sub.ID.String()]
			if ids == nil {
				ids = []string{}
			}
			subByMission[sub.MissionID.String()] = gin.H{
				"status":       sub.Status,
				"playerAnswer": sub.PlayerAnswer,
				"awardedBt":    sub.AwardedBT,
				"photoIds":     ids,
			}
		}

		missions := make([]gin.H, 0, len(character.Missions))
		for _, m := range character.Missions {
			missions = append(missions, gin.H{
				"id":           m.ID,
				"ordinal":      m.Ordinal,
				"tier":         m.Tier,
				"rewardBt":     m.RewardBT,
				"prompt":       m.Prompt,
				"answerFormat": m.AnswerFormat,
				"submission":   subByMission[m.ID.String()], // nil if not started
			})
		}
		charResp["missions"] = missions
	}

	resp["character"] = charResp
	ctx.JSON(http.StatusOK, resp)
}
