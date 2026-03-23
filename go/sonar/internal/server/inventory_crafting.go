package server

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryRecipeIngredientPayload struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type inventoryRecipePayload struct {
	ID          string                             `json:"id"`
	Tier        int                                `json:"tier"`
	IsPublic    bool                               `json:"isPublic"`
	Ingredients []inventoryRecipeIngredientPayload `json:"ingredients"`
}

type inventoryRecipeDefinition struct {
	Station    models.CraftingStation
	ResultItem models.InventoryItem
	Recipe     models.InventoryRecipe
}

type resolvedCraftingRecipe struct {
	Definition        inventoryRecipeDefinition
	OwnedByIngredient map[int]int
	CanCraft          bool
	Known             bool
}

func roomKeyForCraftingStation(station models.CraftingStation) string {
	switch station {
	case models.CraftingStationAlchemy:
		return "alchemy_lab"
	case models.CraftingStationWorkshop:
		return "workshop"
	default:
		return ""
	}
}

func (s *server) allInventoryRecipeDefinitions(
	ctx *gin.Context,
	includeArchived bool,
) ([]inventoryRecipeDefinition, error) {
	items, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		return nil, err
	}

	definitions := make([]inventoryRecipeDefinition, 0)
	for _, item := range items {
		if item.Archived && !includeArchived {
			continue
		}
		for _, recipe := range item.AlchemyRecipes {
			definitions = append(definitions, inventoryRecipeDefinition{
				Station:    models.CraftingStationAlchemy,
				ResultItem: item,
				Recipe:     recipe,
			})
		}
		for _, recipe := range item.WorkshopRecipes {
			definitions = append(definitions, inventoryRecipeDefinition{
				Station:    models.CraftingStationWorkshop,
				ResultItem: item,
				Recipe:     recipe,
			})
		}
	}

	return definitions, nil
}

func (s *server) parseInventoryRecipeConfiguration(
	ctx *gin.Context,
	alchemyPayloads []inventoryRecipePayload,
	workshopPayloads []inventoryRecipePayload,
	consumeTeachRecipeIDs []string,
	currentItemID *int,
) (models.InventoryRecipes, models.InventoryRecipes, models.StringArray, error) {
	existingDefinitions, err := s.allInventoryRecipeDefinitions(ctx, true)
	if err != nil {
		return nil, nil, nil, err
	}

	existingRecipeIDs := map[string]struct{}{}
	validTeachRecipeIDs := map[string]struct{}{}
	for _, definition := range existingDefinitions {
		if currentItemID != nil && definition.ResultItem.ID == *currentItemID {
			continue
		}
		if definition.Recipe.ID == "" {
			continue
		}
		existingRecipeIDs[definition.Recipe.ID] = struct{}{}
		validTeachRecipeIDs[definition.Recipe.ID] = struct{}{}
	}

	itemExistsCache := map[int]bool{}
	localRecipeIDs := map[string]struct{}{}

	parseRecipes := func(
		fieldName string,
		payloads []inventoryRecipePayload,
	) (models.InventoryRecipes, error) {
		recipes := make(models.InventoryRecipes, 0, len(payloads))
		for recipeIndex, payload := range payloads {
			recipeID := strings.TrimSpace(payload.ID)
			if recipeID == "" {
				recipeID = uuid.NewString()
			}
			if _, exists := existingRecipeIDs[recipeID]; exists {
				return nil, fmt.Errorf("%s[%d].id is already used by another recipe", fieldName, recipeIndex)
			}
			if _, exists := localRecipeIDs[recipeID]; exists {
				return nil, fmt.Errorf("%s[%d].id is duplicated", fieldName, recipeIndex)
			}
			localRecipeIDs[recipeID] = struct{}{}
			if payload.Tier < 1 {
				return nil, fmt.Errorf("%s[%d].tier must be 1 or greater", fieldName, recipeIndex)
			}
			if len(payload.Ingredients) == 0 {
				return nil, fmt.Errorf("%s[%d].ingredients must include at least one ingredient", fieldName, recipeIndex)
			}

			ingredientByItemID := map[int]int{}
			for ingredientIndex, ingredient := range payload.Ingredients {
				if ingredient.ItemID <= 0 {
					return nil, fmt.Errorf("%s[%d].ingredients[%d].itemId is required", fieldName, recipeIndex, ingredientIndex)
				}
				if ingredient.Quantity <= 0 {
					return nil, fmt.Errorf("%s[%d].ingredients[%d].quantity must be 1 or greater", fieldName, recipeIndex, ingredientIndex)
				}
				if !itemExistsCache[ingredient.ItemID] {
					if currentItemID != nil && ingredient.ItemID == *currentItemID {
						itemExistsCache[ingredient.ItemID] = true
					} else {
						if _, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, ingredient.ItemID); err != nil {
							if err == gorm.ErrRecordNotFound {
								return nil, fmt.Errorf("%s[%d].ingredients[%d].itemId does not exist", fieldName, recipeIndex, ingredientIndex)
							}
							return nil, err
						}
						itemExistsCache[ingredient.ItemID] = true
					}
				}
				ingredientByItemID[ingredient.ItemID] += ingredient.Quantity
			}

			itemIDs := make([]int, 0, len(ingredientByItemID))
			for itemID := range ingredientByItemID {
				itemIDs = append(itemIDs, itemID)
			}
			sort.Ints(itemIDs)

			ingredients := make([]models.InventoryRecipeIngredient, 0, len(itemIDs))
			for _, itemID := range itemIDs {
				ingredients = append(ingredients, models.InventoryRecipeIngredient{
					ItemID:   itemID,
					Quantity: ingredientByItemID[itemID],
				})
			}

			recipes = append(recipes, models.InventoryRecipe{
				ID:          recipeID,
				Tier:        payload.Tier,
				IsPublic:    payload.IsPublic,
				Ingredients: ingredients,
			})
			validTeachRecipeIDs[recipeID] = struct{}{}
		}
		return recipes, nil
	}

	alchemyRecipes, err := parseRecipes("alchemyRecipes", alchemyPayloads)
	if err != nil {
		return nil, nil, nil, err
	}
	workshopRecipes, err := parseRecipes("workshopRecipes", workshopPayloads)
	if err != nil {
		return nil, nil, nil, err
	}

	teachRecipeIDs := make(models.StringArray, 0, len(consumeTeachRecipeIDs))
	seenTeachIDs := map[string]struct{}{}
	for idx, rawID := range consumeTeachRecipeIDs {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}
		if _, ok := validTeachRecipeIDs[trimmed]; !ok {
			return nil, nil, nil, fmt.Errorf("consumeTeachRecipeIds[%d] does not reference a valid recipe", idx)
		}
		if _, exists := seenTeachIDs[trimmed]; exists {
			continue
		}
		seenTeachIDs[trimmed] = struct{}{}
		teachRecipeIDs = append(teachRecipeIDs, trimmed)
	}

	return alchemyRecipes, workshopRecipes, teachRecipeIDs, nil
}

func (s *server) learnedRecipeDefinitionSummaries(
	ctx *gin.Context,
	userID uuid.UUID,
	sourceItemID *int,
	ownedItemID *uuid.UUID,
	recipeIDs []string,
) ([]gin.H, error) {
	if len(recipeIDs) == 0 {
		return nil, nil
	}

	definitions, err := s.allInventoryRecipeDefinitions(ctx, true)
	if err != nil {
		return nil, err
	}
	definitionsByID := map[string]inventoryRecipeDefinition{}
	for _, definition := range definitions {
		if definition.Recipe.ID == "" {
			continue
		}
		definitionsByID[definition.Recipe.ID] = definition
	}

	learned := make([]gin.H, 0, len(recipeIDs))
	existingLearned, err := s.dbClient.UserLearnedRecipe().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	existingLearnedByID := map[string]struct{}{}
	for _, recipe := range existingLearned {
		existingLearnedByID[recipe.RecipeID] = struct{}{}
	}
	for _, recipeID := range recipeIDs {
		definition, ok := definitionsByID[recipeID]
		if !ok {
			continue
		}
		record := &models.UserLearnedRecipe{
			UserID:                     userID,
			RecipeID:                   recipeID,
			LearnedFromInventoryItemID: sourceItemID,
			LearnedFromOwnedItemID:     ownedItemID,
		}
		if err := s.dbClient.UserLearnedRecipe().Upsert(ctx, record); err != nil {
			return nil, err
		}
		if _, alreadyKnown := existingLearnedByID[recipeID]; alreadyKnown {
			continue
		}
		learned = append(learned, gin.H{
			"recipeId":     definition.Recipe.ID,
			"station":      definition.Station,
			"itemId":       definition.ResultItem.ID,
			"itemName":     definition.ResultItem.Name,
			"itemImageUrl": definition.ResultItem.ImageURL,
			"tier":         definition.Recipe.Tier,
		})
	}

	return learned, nil
}

func (s *server) resolvedCraftingRecipesForUser(
	ctx *gin.Context,
	userID uuid.UUID,
	station models.CraftingStation,
) (int, []resolvedCraftingRecipe, error) {
	roomKey := roomKeyForCraftingStation(station)
	if roomKey == "" {
		return 0, nil, fmt.Errorf("invalid crafting station")
	}

	base, err := s.dbClient.Base().FindByUserID(ctx, userID)
	if err != nil {
		return 0, nil, err
	}
	if base == nil {
		return 0, nil, fmt.Errorf("you need a base before you can craft here")
	}

	structures, err := s.dbClient.UserBaseStructure().FindByBaseID(ctx, base.ID)
	if err != nil {
		return 0, nil, err
	}

	roomTier := 0
	for _, structure := range structures {
		if structure.StructureKey == roomKey {
			roomTier = structure.Level
			break
		}
	}

	definitions, err := s.allInventoryRecipeDefinitions(ctx, false)
	if err != nil {
		return roomTier, nil, err
	}

	learnedRecipes, err := s.dbClient.UserLearnedRecipe().FindByUserID(ctx, userID)
	if err != nil {
		return roomTier, nil, err
	}
	learnedByID := map[string]bool{}
	for _, recipe := range learnedRecipes {
		learnedByID[recipe.RecipeID] = true
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

	resolved := make([]resolvedCraftingRecipe, 0)
	for _, definition := range definitions {
		if definition.Station != station {
			continue
		}
		if definition.Recipe.Tier > roomTier {
			continue
		}
		known := definition.Recipe.IsPublic || learnedByID[definition.Recipe.ID]
		if !known {
			continue
		}
		canCraft := true
		for _, ingredient := range definition.Recipe.Ingredients {
			if ownedByInventoryItemID[ingredient.ItemID] < ingredient.Quantity {
				canCraft = false
				break
			}
		}
		resolved = append(resolved, resolvedCraftingRecipe{
			Definition:        definition,
			OwnedByIngredient: ownedByInventoryItemID,
			CanCraft:          canCraft,
			Known:             known,
		})
	}

	sort.SliceStable(resolved, func(i, j int) bool {
		left := resolved[i]
		right := resolved[j]
		if left.Definition.Recipe.Tier != right.Definition.Recipe.Tier {
			return left.Definition.Recipe.Tier < right.Definition.Recipe.Tier
		}
		return strings.ToLower(left.Definition.ResultItem.Name) < strings.ToLower(right.Definition.ResultItem.Name)
	})

	return roomTier, resolved, nil
}

func serializeResolvedCraftingRecipe(recipe resolvedCraftingRecipe) gin.H {
	ingredients := make([]gin.H, 0, len(recipe.Definition.Recipe.Ingredients))
	for _, ingredient := range recipe.Definition.Recipe.Ingredients {
		ingredients = append(ingredients, gin.H{
			"itemId":        ingredient.ItemID,
			"quantity":      ingredient.Quantity,
			"ownedQuantity": recipe.OwnedByIngredient[ingredient.ItemID],
			"item":          gin.H{"id": ingredient.ItemID},
		})
	}

	return gin.H{
		"id":          recipe.Definition.Recipe.ID,
		"station":     recipe.Definition.Station,
		"tier":        recipe.Definition.Recipe.Tier,
		"isPublic":    recipe.Definition.Recipe.IsPublic,
		"known":       recipe.Known,
		"canCraft":    recipe.CanCraft,
		"resultItem":  recipe.Definition.ResultItem,
		"ingredients": ingredients,
	}
}

func (s *server) hydrateCraftingIngredientItems(
	ctx *gin.Context,
	recipes []resolvedCraftingRecipe,
) ([]gin.H, error) {
	itemIDs := map[int]struct{}{}
	for _, recipe := range recipes {
		for _, ingredient := range recipe.Definition.Recipe.Ingredients {
			itemIDs[ingredient.ItemID] = struct{}{}
		}
	}

	itemsByID := map[int]*models.InventoryItem{}
	for itemID := range itemIDs {
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, itemID)
		if err != nil {
			return nil, err
		}
		itemsByID[itemID] = item
	}

	serialized := make([]gin.H, 0, len(recipes))
	for _, recipe := range recipes {
		entry := serializeResolvedCraftingRecipe(recipe)
		rawIngredients := entry["ingredients"].([]gin.H)
		for idx, ingredient := range recipe.Definition.Recipe.Ingredients {
			rawIngredients[idx]["item"] = itemsByID[ingredient.ItemID]
		}
		entry["ingredients"] = rawIngredients
		serialized = append(serialized, entry)
	}

	return serialized, nil
}

func (s *server) getBaseCraftingRecipes(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	station := models.NormalizeCraftingStation(strings.TrimSpace(ctx.Param("station")))
	if station == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid crafting station"})
		return
	}

	roomTier, recipes, err := s.resolvedCraftingRecipesForUser(ctx, user.ID, station)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if roomTier <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you need a %s before you can craft here", strings.ReplaceAll(roomKeyForCraftingStation(station), "_", " "))})
		return
	}

	serializedRecipes, err := s.hydrateCraftingIngredientItems(ctx, recipes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"station":  station,
		"roomKey":  roomKeyForCraftingStation(station),
		"roomTier": roomTier,
		"recipes":  serializedRecipes,
	})
}

func (s *server) craftBaseRecipe(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	station := models.NormalizeCraftingStation(strings.TrimSpace(ctx.Param("station")))
	if station == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid crafting station"})
		return
	}
	recipeID := strings.TrimSpace(ctx.Param("recipeID"))
	if recipeID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "recipe ID is required"})
		return
	}

	roomTier, recipes, err := s.resolvedCraftingRecipesForUser(ctx, user.ID, station)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if roomTier <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you need a %s before you can craft here", strings.ReplaceAll(roomKeyForCraftingStation(station), "_", " "))})
		return
	}

	var selected *resolvedCraftingRecipe
	for idx := range recipes {
		if recipes[idx].Definition.Recipe.ID == recipeID {
			selected = &recipes[idx]
			break
		}
	}
	if selected == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	if !selected.CanCraft {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you do not have the required ingredients"})
		return
	}

	if err := s.dbClient.InventoryItem().CraftUserInventoryItem(
		ctx,
		user.ID,
		selected.Definition.ResultItem.ID,
		selected.Definition.Recipe.Ingredients,
	); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     fmt.Sprintf("crafted %s", selected.Definition.ResultItem.Name),
		"station":     station,
		"recipeId":    selected.Definition.Recipe.ID,
		"craftedItem": selected.Definition.ResultItem,
	})
}
