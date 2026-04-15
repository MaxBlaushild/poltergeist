package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fetchQuestTurnInRequirement struct {
	InventoryItemID int                   `json:"inventoryItemId"`
	Quantity        int                   `json:"quantity"`
	OwnedQuantity   int                   `json:"ownedQuantity"`
	InventoryItem   *models.InventoryItem `json:"inventoryItem,omitempty"`
}

type fetchQuestTurnInResponse struct {
	QuestID          string                        `json:"questId"`
	QuestName        string                        `json:"questName"`
	QuestDescription string                        `json:"questDescription"`
	QuestNodeID      string                        `json:"questNodeId"`
	CharacterID      string                        `json:"characterId"`
	CharacterName    string                        `json:"characterName"`
	Requirements     []fetchQuestTurnInRequirement `json:"requirements"`
	CanDeliver       bool                          `json:"canDeliver"`
}

func isGeneratedFetchQuestCharacter(character *models.Character) bool {
	return models.CharacterHasInternalTag(
		character,
		models.CharacterInternalTagGeneratedFetchQuest,
	)
}

func fetchQuestCharacterVisibleToUser(
	character *models.Character,
	activeFetchCharacterIDs map[uuid.UUID]struct{},
) bool {
	if !isGeneratedFetchQuestCharacter(character) {
		return true
	}
	if character == nil {
		return false
	}
	_, ok := activeFetchCharacterIDs[character.ID]
	return ok
}

func (s *server) activeFetchQuestCharacterIDsForUser(
	ctx context.Context,
	userID uuid.UUID,
) (map[uuid.UUID]struct{}, error) {
	acceptances, err := s.dbClient.QuestAcceptanceV2().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ids := make(map[uuid.UUID]struct{})
	for _, acceptance := range acceptances {
		if acceptance.IsClosed() {
			continue
		}
		quest, err := s.dbClient.Quest().FindByID(ctx, acceptance.QuestID)
		if err != nil {
			return nil, err
		}
		if quest == nil || !questVisibleToUser(userID, quest) {
			continue
		}
		currentNode, err := s.currentQuestNode(ctx, quest, &acceptance)
		if err != nil {
			return nil, err
		}
		if currentNode == nil || !currentNode.IsFetchQuestNode() {
			continue
		}
		if currentNode.FetchCharacterID == nil || *currentNode.FetchCharacterID == uuid.Nil {
			continue
		}
		ids[*currentNode.FetchCharacterID] = struct{}{}
	}

	return ids, nil
}

func (s *server) currentFetchQuestNodeForUser(
	ctx *gin.Context,
	userID uuid.UUID,
	questID uuid.UUID,
) (*models.Quest, *models.QuestAcceptanceV2, *models.QuestNode, error) {
	quest, err := s.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		return nil, nil, nil, err
	}
	if quest == nil || !questVisibleToUser(userID, quest) {
		return nil, nil, nil, fmt.Errorf("quest not found")
	}

	acceptance, err := s.dbClient.QuestAcceptanceV2().FindByUserAndQuest(
		ctx,
		userID,
		questID,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	if acceptance == nil {
		return nil, nil, nil, fmt.Errorf("quest not accepted")
	}
	if acceptance.IsClosed() {
		return nil, nil, nil, fmt.Errorf("quest already turned in")
	}

	currentNode, err := s.currentQuestNode(ctx, quest, acceptance)
	if err != nil {
		return nil, nil, nil, err
	}
	if currentNode == nil || !currentNode.IsFetchQuestNode() {
		return nil, nil, nil, fmt.Errorf("quest has no active fetch objective")
	}

	return quest, acceptance, currentNode, nil
}

func (s *server) resolveFetchQuestTurnInState(
	ctx *gin.Context,
	userID uuid.UUID,
	quest *models.Quest,
	node *models.QuestNode,
) (*fetchQuestTurnInResponse, *models.Character, error) {
	if quest == nil || node == nil || !node.IsFetchQuestNode() {
		return nil, nil, fmt.Errorf("quest has no active fetch objective")
	}
	if node.FetchCharacterID == nil || *node.FetchCharacterID == uuid.Nil {
		return nil, nil, fmt.Errorf("fetch quest character missing")
	}

	character := node.FetchCharacter
	if character == nil || character.ID != *node.FetchCharacterID {
		loadedCharacter, err := s.dbClient.Character().FindByID(
			ctx,
			*node.FetchCharacterID,
		)
		if err != nil {
			return nil, nil, err
		}
		if loadedCharacter == nil {
			return nil, nil, fmt.Errorf("fetch quest character not found")
		}
		character = loadedCharacter
	}

	ownedItems, err := s.dbClient.InventoryItem().GetUsersItems(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	ownedQuantityByItemID := map[int]int{}
	for _, ownedItem := range ownedItems {
		if ownedItem.InventoryItemID <= 0 || ownedItem.Quantity <= 0 {
			continue
		}
		ownedQuantityByItemID[ownedItem.InventoryItemID] += ownedItem.Quantity
	}

	requirements := make([]fetchQuestTurnInRequirement, 0, len(node.FetchRequirements))
	canDeliver := len(node.FetchRequirements) > 0
	for _, requirement := range node.FetchRequirements {
		if requirement.InventoryItemID <= 0 || requirement.Quantity <= 0 {
			continue
		}
		inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(
			ctx,
			requirement.InventoryItemID,
		)
		if err != nil {
			return nil, nil, err
		}
		ownedQuantity := ownedQuantityByItemID[requirement.InventoryItemID]
		if ownedQuantity < requirement.Quantity {
			canDeliver = false
		}
		requirements = append(requirements, fetchQuestTurnInRequirement{
			InventoryItemID: requirement.InventoryItemID,
			Quantity:        requirement.Quantity,
			OwnedQuantity:   ownedQuantity,
			InventoryItem:   inventoryItem,
		})
	}

	characterName := strings.TrimSpace(character.Name)
	if characterName == "" {
		characterName = "Character"
	}

	return &fetchQuestTurnInResponse{
		QuestID:          quest.ID.String(),
		QuestName:        strings.TrimSpace(quest.Name),
		QuestDescription: strings.TrimSpace(quest.Description),
		QuestNodeID:      node.ID.String(),
		CharacterID:      character.ID.String(),
		CharacterName:    characterName,
		Requirements:     requirements,
		CanDeliver:       canDeliver,
	}, character, nil
}

func (s *server) buildFetchQuestTurnInActionsForCharacter(
	ctx *gin.Context,
	user *models.User,
	characterID uuid.UUID,
) ([]*models.CharacterAction, error) {
	if user == nil {
		return nil, nil
	}

	acceptances, err := s.dbClient.QuestAcceptanceV2().FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	actions := make([]*models.CharacterAction, 0)
	for _, acceptance := range acceptances {
		if acceptance.IsClosed() {
			continue
		}

		quest, err := s.dbClient.Quest().FindByID(ctx, acceptance.QuestID)
		if err != nil {
			return nil, err
		}
		if quest == nil || !questVisibleToUser(user.ID, quest) {
			continue
		}

		currentNode, err := s.currentQuestNode(ctx, quest, &acceptance)
		if err != nil {
			return nil, err
		}
		if currentNode == nil || !currentNode.IsFetchQuestNode() {
			continue
		}
		if currentNode.FetchCharacterID == nil || *currentNode.FetchCharacterID != characterID {
			continue
		}

		actions = append(actions, &models.CharacterAction{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CharacterID: characterID,
			ActionType:  models.ActionTypeReceiveQuestItems,
			Dialogue:    models.DialogueSequence{},
			Metadata: models.MetadataJSONB{
				"questId":          quest.ID.String(),
				"questNodeId":      currentNode.ID.String(),
				"questName":        strings.TrimSpace(quest.Name),
				"questDescription": strings.TrimSpace(quest.Description),
				"questCategory":    quest.Category,
			},
		})
	}

	return actions, nil
}

func (s *server) getQuestFetchTurnIn(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	questID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest id"})
		return
	}

	quest, _, node, err := s.currentFetchQuestNodeForUser(ctx, user.ID, questID)
	if err != nil {
		switch err.Error() {
		case "quest not found":
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "quest not accepted":
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	response, _, err := s.resolveFetchQuestTurnInState(ctx, user.ID, quest, node)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) submitQuestFetchTurnIn(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	questID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest id"})
		return
	}

	quest, acceptance, node, err := s.currentFetchQuestNodeForUser(
		ctx,
		user.ID,
		questID,
	)
	if err != nil {
		switch err.Error() {
		case "quest not found":
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "quest not accepted":
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	response, character, err := s.resolveFetchQuestTurnInState(
		ctx,
		user.ID,
		quest,
		node,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !response.CanDeliver {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing required items"})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	distance, ok := nearestCharacterDistanceMeters(character, userLat, userLng)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "character location unavailable"})
		return
	}
	if distance > scenarioInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"you must be within %.0f meters of %s. Currently %.0f meters away",
				scenarioInteractRadiusMeters,
				response.CharacterName,
				distance,
			),
		})
		return
	}

	requirements := make([]models.FetchQuestRequirement, 0, len(node.FetchRequirements))
	for _, requirement := range node.FetchRequirements {
		requirements = append(requirements, requirement)
	}
	if err := s.dbClient.InventoryItem().ConsumeUserInventoryItems(
		ctx,
		user.ID,
		requirements,
	); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.markQuestNodeCompleteForAcceptance(
		ctx,
		quest,
		acceptance,
		node.ID,
		time.Now(),
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.shareQuestNodeCompletionWithEligiblePartyMembers(ctx, user, quest, node)

	objectivesComplete, err := s.questlogClient.AreQuestObjectivesComplete(
		ctx,
		user.ID,
		quest.ID,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":        "items delivered",
		"questId":        quest.ID.String(),
		"questNodeId":    node.ID.String(),
		"questCompleted": objectivesComplete,
	})
}
