package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /gm/submissions?status=submitted — the queue GMs work from.
func (s *server) gmListSubmissions(ctx *gin.Context) {
	status := ctx.Query("status")
	details, err := s.dbClient.Vampire().ListSubmissionsDetailed(ctx, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	photoMap, err := s.photoIDsBySubmission(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(details))
	for _, d := range details {
		ids := photoMap[d.ID.String()]
		if ids == nil {
			ids = []string{}
		}
		out = append(out, gin.H{
			"id":                  d.ID,
			"status":              d.Status,
			"playerAnswer":        d.PlayerAnswer,
			"awardedBt":           d.AwardedBT,
			"guestLabel":          d.GuestLabel,
			"characterName":       d.CharacterName,
			"houseName":           d.HouseName,
			"missionTier":         d.MissionTier,
			"missionPrompt":       d.MissionPrompt,
			"missionAnswerFormat": d.MissionAnswerFormat,
			"rewardBt":            d.RewardBT,
			"photoIds":            ids,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"submissions": out})
}

// POST /gm/submissions/:id/verify — approve a submission. Records the Blood
// Tokens (defaults to the mission's reward, GM may override). BT is only logged
// on the transition into verified, so re-verifying does not double-award.
func (s *server) gmVerifySubmission(ctx *gin.Context) {
	subID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	sub, err := s.dbClient.Vampire().GetSubmissionByID(ctx, subID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}

	mission, err := s.dbClient.Vampire().GetMissionByID(ctx, sub.MissionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if mission == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "mission not found"})
		return
	}

	var body struct {
		AwardedBT *int `json:"awardedBt"`
	}
	_ = ctx.ShouldBindJSON(&body)

	awarded := mission.RewardBT
	if body.AwardedBT != nil {
		awarded = *body.AwardedBT
	}

	gmName := gmNameFromContext(ctx)
	if err := s.dbClient.Vampire().UpdateSubmissionStatus(ctx, subID, "verified", awarded, gmName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record the Blood Token award only on the transition into verified.
	if sub.Status != "verified" {
		if err := s.dbClient.Vampire().AddBloodTokens(ctx, &models.VampireBloodTokenLog{
			PlayerID: sub.PlayerID,
			Delta:    awarded,
			Reason:   "mission verified",
			Source:   "mission",
			GMName:   gmName,
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Sabotage missions also deduct House Favor from a target house, once,
		// on the transition into verified.
		if mission.SabotageHouseID != nil && mission.SabotageHF > 0 {
			if err := s.dbClient.Vampire().AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
				HouseID: *mission.SabotageHouseID,
				Delta:   -float64(mission.SabotageHF),
				Reason:  "Sabotage mission confirmed",
				GMName:  gmName,
				Source:  "mission",
			}); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	s.logGM(ctx, "verify_submission", map[string]interface{}{
		"submissionId": subID.String(),
		"awardedBt":    awarded,
	})
	ctx.JSON(http.StatusOK, gin.H{"status": "verified", "awardedBt": awarded})
}

// POST /gm/submissions/:id/reject — send a submission back. Does not touch BT.
func (s *server) gmRejectSubmission(ctx *gin.Context) {
	subID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	sub, err := s.dbClient.Vampire().GetSubmissionByID(ctx, subID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}

	gmName := gmNameFromContext(ctx)
	if err := s.dbClient.Vampire().UpdateSubmissionStatus(ctx, subID, "rejected", 0, gmName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "reject_submission", map[string]interface{}{"submissionId": subID.String()})
	ctx.JSON(http.StatusOK, gin.H{"status": "rejected"})
}
