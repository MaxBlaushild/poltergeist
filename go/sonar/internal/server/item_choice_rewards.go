package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type chooseItemChoiceRewardRequest struct {
	InventoryItemID int `json:"inventoryItemId"`
}

func (s *server) chooseScenarioItemChoiceReward(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	scenarioID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario ID"})
		return
	}

	var requestBody chooseItemChoiceRewardRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.InventoryItemID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventoryItemId is required"})
		return
	}

	pending, err := s.dbClient.Scenario().FindItemChoicePendingByUserAndScenario(ctx, user.ID, scenarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if pending == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "no pending scenario item choice reward"})
		return
	}

	scenario, err := s.dbClient.Scenario().FindByID(ctx, scenarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if scenario == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	choiceRewards := []scenarioRewardItem{}
	if pending.ScenarioOptionID != nil {
		option := findScenarioOption(scenario, *pending.ScenarioOptionID)
		if option == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "pending scenario option no longer exists"})
			return
		}
		choiceRewards = scenarioRewardItemChoicesFromOption(option.ItemChoiceRewards)
	} else {
		choiceRewards = scenarioRewardItemChoicesFromScenario(scenario.ItemChoiceRewards)
	}
	if len(choiceRewards) == 0 {
		_ = s.dbClient.Scenario().DeleteItemChoicePending(ctx, pending.ID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no item choice rewards configured for this scenario result"})
		return
	}

	quantity := 0
	for _, reward := range choiceRewards {
		if reward.InventoryItemID == requestBody.InventoryItemID {
			quantity = reward.Quantity
			break
		}
	}
	if quantity <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "selected item is not a valid choice"})
		return
	}

	if err := s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(
		ctx,
		nil,
		&user.ID,
		requestBody.InventoryItemID,
		quantity,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().DeleteItemChoicePending(ctx, pending.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	awarded := s.hydrateItemAwarded(ctx, []models.ItemAwarded{{ID: requestBody.InventoryItemID, Quantity: quantity}})
	if len(awarded) == 0 {
		awarded = []models.ItemAwarded{{ID: requestBody.InventoryItemID, Quantity: quantity}}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"itemAwarded":       awarded[0],
		"itemsAwarded":      awarded,
		"itemChoiceRewards": []models.ItemAwarded{},
	})
}

func (s *server) chooseChallengeItemChoiceReward(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	challengeID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge ID"})
		return
	}

	var requestBody chooseItemChoiceRewardRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.InventoryItemID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventoryItemId is required"})
		return
	}

	pending, err := s.dbClient.Challenge().FindItemChoicePendingByUserAndChallenge(ctx, user.ID, challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if pending == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "no pending challenge item choice reward"})
		return
	}

	challenge, err := s.dbClient.Challenge().FindByID(ctx, challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "challenge not found"})
		return
	}

	choiceRewards := scenarioRewardItemChoicesFromChallenge(challenge.ItemChoiceRewards)
	if len(choiceRewards) == 0 {
		_ = s.dbClient.Challenge().DeleteItemChoicePending(ctx, pending.ID)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no item choice rewards configured for this challenge result"})
		return
	}

	quantity := 0
	for _, reward := range choiceRewards {
		if reward.InventoryItemID == requestBody.InventoryItemID {
			quantity = reward.Quantity
			break
		}
	}
	if quantity <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "selected item is not a valid choice"})
		return
	}

	if err := s.dbClient.InventoryItem().CreateOrIncrementInventoryItem(
		ctx,
		nil,
		&user.ID,
		requestBody.InventoryItemID,
		quantity,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Challenge().DeleteItemChoicePending(ctx, pending.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	awarded := s.hydrateItemAwarded(ctx, []models.ItemAwarded{{ID: requestBody.InventoryItemID, Quantity: quantity}})
	if len(awarded) == 0 {
		awarded = []models.ItemAwarded{{ID: requestBody.InventoryItemID, Quantity: quantity}}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"itemAwarded":       awarded[0],
		"itemsAwarded":      awarded,
		"itemChoiceRewards": []models.ItemAwarded{},
	})
}
