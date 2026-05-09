package server

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const scrapworksRoomKey = "scrapworks"

type inventorySalvageOutputPayload struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type inventorySalvageRecipePayload struct {
	Tier    int                             `json:"tier"`
	Outputs []inventorySalvageOutputPayload `json:"outputs"`
}

type resolvedScrapworksSalvageable struct {
	OwnedItem           models.OwnedInventoryItem
	Item                models.InventoryItem
	Recipe              models.InventorySalvageRecipe
	EquippedQuantity    int
	AvailableQuantity   int
	OwnedByOutputItemID map[int]int
}

func inventorySalvageRecipePayloadsFromModels(
	recipes models.InventorySalvageRecipes,
) []inventorySalvageRecipePayload {
	if len(recipes) == 0 {
		return []inventorySalvageRecipePayload{}
	}

	payloads := make([]inventorySalvageRecipePayload, 0, len(recipes))
	for _, recipe := range recipes {
		outputs := make([]inventorySalvageOutputPayload, 0, len(recipe.Outputs))
		for _, output := range recipe.Outputs {
			outputs = append(outputs, inventorySalvageOutputPayload{
				ItemID:   output.ItemID,
				Quantity: output.Quantity,
			})
		}
		payloads = append(payloads, inventorySalvageRecipePayload{
			Tier:    recipe.Tier,
			Outputs: outputs,
		})
	}
	return payloads
}

func (s *server) parseInventorySalvageConfiguration(
	ctx *gin.Context,
	payloads []inventorySalvageRecipePayload,
	currentItemID *int,
) (models.InventorySalvageRecipes, error) {
	recipes := make(models.InventorySalvageRecipes, 0, len(payloads))
	if len(payloads) == 0 {
		return recipes, nil
	}

	itemExistsCache := map[int]bool{}
	seenTiers := map[int]struct{}{}
	for recipeIndex, payload := range payloads {
		if payload.Tier < 1 {
			return nil, fmt.Errorf("scrapworksRecipes[%d].tier must be 1 or greater", recipeIndex)
		}
		if _, exists := seenTiers[payload.Tier]; exists {
			return nil, fmt.Errorf("scrapworksRecipes[%d].tier is duplicated", recipeIndex)
		}
		seenTiers[payload.Tier] = struct{}{}
		if len(payload.Outputs) == 0 {
			return nil, fmt.Errorf("scrapworksRecipes[%d].outputs must include at least one output", recipeIndex)
		}

		outputByItemID := map[int]int{}
		for outputIndex, output := range payload.Outputs {
			if output.ItemID <= 0 {
				return nil, fmt.Errorf("scrapworksRecipes[%d].outputs[%d].itemId is required", recipeIndex, outputIndex)
			}
			if output.Quantity <= 0 {
				return nil, fmt.Errorf("scrapworksRecipes[%d].outputs[%d].quantity must be 1 or greater", recipeIndex, outputIndex)
			}
			if currentItemID != nil && output.ItemID == *currentItemID {
				return nil, fmt.Errorf("scrapworksRecipes[%d].outputs[%d].itemId cannot match the salvaged item", recipeIndex, outputIndex)
			}
			if !itemExistsCache[output.ItemID] {
				if _, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, output.ItemID); err != nil {
					return nil, fmt.Errorf("scrapworksRecipes[%d].outputs[%d].itemId does not exist", recipeIndex, outputIndex)
				}
				itemExistsCache[output.ItemID] = true
			}
			outputByItemID[output.ItemID] += output.Quantity
		}

		itemIDs := make([]int, 0, len(outputByItemID))
		for itemID := range outputByItemID {
			itemIDs = append(itemIDs, itemID)
		}
		sort.Ints(itemIDs)

		outputs := make([]models.InventorySalvageOutput, 0, len(itemIDs))
		for _, itemID := range itemIDs {
			outputs = append(outputs, models.InventorySalvageOutput{
				ItemID:   itemID,
				Quantity: outputByItemID[itemID],
			})
		}

		recipes = append(recipes, models.InventorySalvageRecipe{
			Tier:    payload.Tier,
			Outputs: outputs,
		})
	}

	sort.SliceStable(recipes, func(i, j int) bool {
		return recipes[i].Tier < recipes[j].Tier
	})

	return recipes, nil
}

func bestScrapworksRecipeForTier(
	recipes models.InventorySalvageRecipes,
	roomTier int,
) *models.InventorySalvageRecipe {
	if roomTier <= 0 || len(recipes) == 0 {
		return nil
	}

	var selected *models.InventorySalvageRecipe
	for idx := range recipes {
		recipe := &recipes[idx]
		if recipe.Tier <= 0 || recipe.Tier > roomTier || len(recipe.Outputs) == 0 {
			continue
		}
		if selected == nil || recipe.Tier > selected.Tier {
			selected = recipe
		}
	}
	return selected
}

func (s *server) resolvedScrapworksSalvageablesForUser(
	ctx *gin.Context,
	userID uuid.UUID,
) (int, []resolvedScrapworksSalvageable, error) {
	base, err := s.dbClient.Base().FindByUserID(ctx, userID)
	if err != nil {
		return 0, nil, err
	}
	if base == nil {
		return 0, nil, fmt.Errorf("you need a base before you can salvage items here")
	}

	structures, err := s.dbClient.UserBaseStructure().FindByBaseID(ctx, base.ID)
	if err != nil {
		return 0, nil, err
	}

	roomTier := 0
	for _, structure := range structures {
		if structure.StructureKey == scrapworksRoomKey {
			roomTier = structure.Level
			break
		}
	}

	items, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		return roomTier, nil, err
	}
	itemByID := make(map[int]models.InventoryItem, len(items))
	for _, item := range items {
		itemByID[item.ID] = item
	}

	ownedItems, err := s.dbClient.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{
		UserID: &userID,
	})
	if err != nil {
		return roomTier, nil, err
	}
	ownedByInventoryItemID := map[int]int{}
	for _, owned := range ownedItems {
		ownedByInventoryItemID[owned.InventoryItemID] += owned.Quantity
	}

	equipment, err := s.dbClient.UserEquipment().FindByUserID(ctx, userID)
	if err != nil {
		return roomTier, nil, err
	}
	equippedCountByOwnedItemID := map[uuid.UUID]int{}
	for _, entry := range equipment {
		equippedCountByOwnedItemID[entry.OwnedInventoryItemID]++
	}

	resolved := make([]resolvedScrapworksSalvageable, 0)
	for _, owned := range ownedItems {
		item, ok := itemByID[owned.InventoryItemID]
		if !ok {
			continue
		}
		recipe := bestScrapworksRecipeForTier(item.ScrapworksRecipes, roomTier)
		if recipe == nil {
			continue
		}
		equippedQuantity := equippedCountByOwnedItemID[owned.ID]
		availableQuantity := owned.Quantity - equippedQuantity
		if availableQuantity <= 0 {
			continue
		}

		resolved = append(resolved, resolvedScrapworksSalvageable{
			OwnedItem:           owned,
			Item:                item,
			Recipe:              *recipe,
			EquippedQuantity:    equippedQuantity,
			AvailableQuantity:   availableQuantity,
			OwnedByOutputItemID: ownedByInventoryItemID,
		})
	}

	sort.SliceStable(resolved, func(i, j int) bool {
		left := resolved[i]
		right := resolved[j]
		if left.Item.ItemLevel != right.Item.ItemLevel {
			return left.Item.ItemLevel > right.Item.ItemLevel
		}
		return strings.ToLower(left.Item.Name) < strings.ToLower(right.Item.Name)
	})

	return roomTier, resolved, nil
}

func serializeResolvedScrapworksSalvageable(
	salvageable resolvedScrapworksSalvageable,
	itemsByID map[int]*models.InventoryItem,
) gin.H {
	outputs := make([]gin.H, 0, len(salvageable.Recipe.Outputs))
	for _, output := range salvageable.Recipe.Outputs {
		outputs = append(outputs, gin.H{
			"itemId":        output.ItemID,
			"quantity":      output.Quantity,
			"ownedQuantity": salvageable.OwnedByOutputItemID[output.ItemID],
			"item":          itemsByID[output.ItemID],
		})
	}

	return gin.H{
		"ownedInventoryItemId": salvageable.OwnedItem.ID,
		"ownedQuantity":        salvageable.OwnedItem.Quantity,
		"equippedQuantity":     salvageable.EquippedQuantity,
		"availableQuantity":    salvageable.AvailableQuantity,
		"tier":                 salvageable.Recipe.Tier,
		"item":                 salvageable.Item,
		"outputs":              outputs,
	}
}

func (s *server) getBaseScrapworksSalvageables(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roomTier, salvageables, err := s.resolvedScrapworksSalvageablesForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if roomTier <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need a scrapworks before you can salvage items here"})
		return
	}

	itemsByID := map[int]*models.InventoryItem{}
	for idx := range salvageables {
		itemCopy := salvageables[idx].Item
		itemsByID[itemCopy.ID] = &itemCopy
		for _, output := range salvageables[idx].Recipe.Outputs {
			if _, exists := itemsByID[output.ItemID]; exists {
				continue
			}
			item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, output.ItemID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			itemsByID[output.ItemID] = item
		}
	}

	serialized := make([]gin.H, 0, len(salvageables))
	for _, salvageable := range salvageables {
		serialized = append(serialized, serializeResolvedScrapworksSalvageable(salvageable, itemsByID))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"roomKey":      scrapworksRoomKey,
		"roomTier":     roomTier,
		"salvageables": serialized,
	})
}

func (s *server) salvageBaseScrapworksItem(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ownedInventoryItemID, err := uuid.Parse(strings.TrimSpace(ctx.Param("ownedInventoryItemID")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid owned inventory item ID"})
		return
	}

	roomTier, salvageables, err := s.resolvedScrapworksSalvageablesForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if roomTier <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need a scrapworks before you can salvage items here"})
		return
	}

	var selected *resolvedScrapworksSalvageable
	for idx := range salvageables {
		if salvageables[idx].OwnedItem.ID == ownedInventoryItemID {
			selected = &salvageables[idx]
			break
		}
	}
	if selected == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "salvageable item not found"})
		return
	}

	updatedOwned, updatedOutputQuantities, err := s.dbClient.InventoryItem().SalvageUserInventoryItem(
		ctx,
		user.ID,
		ownedInventoryItemID,
		selected.Recipe.Outputs,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	yielded := make([]gin.H, 0, len(selected.Recipe.Outputs))
	for _, output := range selected.Recipe.Outputs {
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, output.ItemID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		yielded = append(yielded, gin.H{
			"itemId":        output.ItemID,
			"quantity":      output.Quantity,
			"ownedQuantity": updatedOutputQuantities[output.ItemID],
			"item":          item,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":              fmt.Sprintf("salvaged %s", selected.Item.Name),
		"roomKey":              scrapworksRoomKey,
		"roomTier":             roomTier,
		"ownedInventoryItemId": selected.OwnedItem.ID,
		"remainingQuantity": func() int {
			if updatedOwned == nil {
				return 0
			}
			return updatedOwned.Quantity
		}(),
		"salvagedItem": selected.Item,
		"yielded":      yielded,
	})
}
