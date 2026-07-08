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
// Tokens owed (defaults to the mission's reward, GM may override) but does NOT
// pay them yet: the player is notified to collect their payout from Ivara at the
// Blood Bank, where a GM marks it redeemed. Sabotage House Favor is applied once,
// on the transition into approved.
func (s *server) gmApproveSubmission(ctx *gin.Context) {
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
	if err := s.dbClient.Vampire().UpdateSubmissionStatus(ctx, subID, "approved", awarded, gmName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Side effects only on the first approval, so re-approving is idempotent.
	if sub.Status != "approved" && sub.Status != "redeemed" {
		// Sabotage missions deduct House Favor from a target house.
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

		// Direct the player to the Blood Bank for their payout.
		pid := sub.PlayerID
		if err := s.dbClient.Vampire().CreateNotification(ctx, &models.VampireNotification{
			Title:     "Payout Ready",
			Body:      "One of your missions has been approved. Visit Ivara at the Blood Bank to collect your Blood Tokens.",
			Scope:     "player",
			TargetID:  &pid,
			CreatedBy: gmName,
			Active:    true,
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	s.logGM(ctx, "approve_submission", map[string]interface{}{
		"submissionId": subID.String(),
		"awardedBt":    awarded,
	})
	ctx.JSON(http.StatusOK, gin.H{"status": "approved", "awardedBt": awarded})
}

// POST /gm/submissions/:id/redeem — mark an approved submission paid out at the
// Blood Bank. This is where the Blood Tokens are actually logged, so the digital
// total matches the physical vials handed over. Idempotent.
func (s *server) gmRedeemSubmission(ctx *gin.Context) {
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
	if sub.Status == "redeemed" {
		ctx.JSON(http.StatusOK, gin.H{"status": "redeemed", "awardedBt": sub.AwardedBT})
		return
	}
	if sub.Status != "approved" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "approve the submission before redeeming it"})
		return
	}

	gmName := gmNameFromContext(ctx)
	if err := s.dbClient.Vampire().UpdateSubmissionStatus(ctx, subID, "redeemed", sub.AwardedBT, gmName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Vampire().AddBloodTokens(ctx, &models.VampireBloodTokenLog{
		PlayerID: sub.PlayerID,
		Delta:    sub.AwardedBT,
		Reason:   "mission redeemed",
		Source:   "mission",
		GMName:   gmName,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "redeem_submission", map[string]interface{}{
		"submissionId": subID.String(),
		"awardedBt":    sub.AwardedBT,
	})
	ctx.JSON(http.StatusOK, gin.H{"status": "redeemed", "awardedBt": sub.AwardedBT})
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
