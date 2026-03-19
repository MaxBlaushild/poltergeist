package server

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	baseGridSize   = 5
	baseGridCenter = 2
)

func serializeBaseResourceBalance(balance models.BaseResourceBalance) gin.H {
	return gin.H{
		"resourceKey": balance.ResourceKey,
		"amount":      balance.Amount,
		"updatedAt":   balance.UpdatedAt,
	}
}

func serializeBaseResourceLedger(entry models.BaseResourceLedger) gin.H {
	return gin.H{
		"id":          entry.ID,
		"userId":      entry.UserID,
		"resourceKey": entry.ResourceKey,
		"delta":       entry.Delta,
		"sourceType":  entry.SourceType,
		"sourceId":    entry.SourceID,
		"notes":       entry.Notes,
		"createdAt":   entry.CreatedAt,
	}
}

func serializeBaseStructureLevelCost(cost models.BaseStructureLevelCost) gin.H {
	return gin.H{
		"id":                    cost.ID,
		"structureDefinitionId": cost.StructureDefinitionID,
		"level":                 cost.Level,
		"resourceKey":           cost.ResourceKey,
		"amount":                cost.Amount,
		"createdAt":             cost.CreatedAt,
		"updatedAt":             cost.UpdatedAt,
	}
}

func serializeBaseStructureLevelVisual(visual models.BaseStructureLevelVisual) gin.H {
	return gin.H{
		"id":                    visual.ID,
		"structureDefinitionId": visual.StructureDefinitionID,
		"level":                 visual.Level,
		"imageUrl":              visual.ImageURL,
		"thumbnailUrl":          visual.ThumbnailURL,
		"imageGenerationStatus": visual.ImageGenerationStatus,
		"imageGenerationError":  visual.ImageGenerationError,
		"createdAt":             visual.CreatedAt,
		"updatedAt":             visual.UpdatedAt,
	}
}

func buildSerializedBaseStructureLevelVisuals(definition models.BaseStructureDefinition) []gin.H {
	byLevel := make(map[int]models.BaseStructureLevelVisual, len(definition.LevelVisuals))
	for _, visual := range definition.LevelVisuals {
		byLevel[visual.Level] = visual
	}
	visuals := make([]gin.H, 0, max(definition.MaxLevel, 1))
	for level := 1; level <= max(definition.MaxLevel, 1); level++ {
		if visual, ok := byLevel[level]; ok {
			visuals = append(visuals, serializeBaseStructureLevelVisual(visual))
			continue
		}
		visuals = append(visuals, gin.H{
			"id":                    nil,
			"structureDefinitionId": definition.ID,
			"level":                 level,
			"imageUrl":              "",
			"thumbnailUrl":          "",
			"imageGenerationStatus": models.BaseStructureImageGenerationStatusNone,
			"imageGenerationError":  nil,
			"createdAt":             nil,
			"updatedAt":             nil,
		})
	}
	return visuals
}

func serializeBaseStructureDefinition(definition models.BaseStructureDefinition) gin.H {
	levelCosts := make([]gin.H, 0, len(definition.LevelCosts))
	for _, cost := range definition.LevelCosts {
		levelCosts = append(levelCosts, serializeBaseStructureLevelCost(cost))
	}

	return gin.H{
		"id":           definition.ID,
		"key":          definition.Key,
		"name":         definition.Name,
		"description":  definition.Description,
		"category":     definition.Category,
		"maxLevel":     definition.MaxLevel,
		"sortOrder":    definition.SortOrder,
		"imageUrl":     definition.ImageURL,
		"effectType":   definition.EffectType,
		"effectConfig": definition.EffectConfig,
		"prereqConfig": definition.PrereqConfig,
		"active":       definition.Active,
		"createdAt":    definition.CreatedAt,
		"updatedAt":    definition.UpdatedAt,
		"levelCosts":   levelCosts,
		"levelVisuals": buildSerializedBaseStructureLevelVisuals(definition),
	}
}

func serializeUserBaseStructure(structure models.UserBaseStructure) gin.H {
	return gin.H{
		"id":           structure.ID,
		"baseId":       structure.BaseID,
		"userId":       structure.UserID,
		"structureKey": structure.StructureKey,
		"level":        structure.Level,
		"gridX":        structure.GridX,
		"gridY":        structure.GridY,
		"createdAt":    structure.CreatedAt,
		"updatedAt":    structure.UpdatedAt,
	}
}

func serializeUserBaseDailyState(state models.UserBaseDailyState) gin.H {
	return gin.H{
		"id":        state.ID,
		"userId":    state.UserID,
		"stateKey":  state.StateKey,
		"state":     state.StateJSON,
		"resetsOn":  state.ResetsOn,
		"createdAt": state.CreatedAt,
		"updatedAt": state.UpdatedAt,
	}
}

func serializeBaseResourceBalances(balances []models.BaseResourceBalance) []gin.H {
	response := make([]gin.H, 0, len(balances))
	for _, balance := range balances {
		response = append(response, serializeBaseResourceBalance(balance))
	}
	return response
}

func serializeBaseResourceLedgerEntries(entries []models.BaseResourceLedger) []gin.H {
	response := make([]gin.H, 0, len(entries))
	for _, entry := range entries {
		response = append(response, serializeBaseResourceLedger(entry))
	}
	return response
}

func serializeUserBaseStructures(structures []models.UserBaseStructure) []gin.H {
	response := make([]gin.H, 0, len(structures))
	for _, structure := range structures {
		response = append(response, serializeUserBaseStructure(structure))
	}
	return response
}

func serializeUserBaseDailyStates(states []models.UserBaseDailyState) []gin.H {
	response := make([]gin.H, 0, len(states))
	for _, state := range states {
		response = append(response, serializeUserBaseDailyState(state))
	}
	return response
}

func serializeBaseStructureDefinitions(definitions []models.BaseStructureDefinition) []gin.H {
	response := make([]gin.H, 0, len(definitions))
	for _, definition := range definitions {
		response = append(response, serializeBaseStructureDefinition(definition))
	}
	return response
}

func buildBaseGrassTileURLs() gin.H {
	urls := gin.H{}
	for gridY := 0; gridY < baseGridSize; gridY++ {
		for gridX := 0; gridX < baseGridSize; gridX++ {
			_, destinationKey, _ := baseGrassTileConfig(gridX, gridY)
			urls[fmt.Sprintf("%d:%d", gridX, gridY)] = staticThumbnailURL(destinationKey)
		}
	}
	return urls
}

func (s *server) loadBaseSnapshot(ctx *gin.Context, base *models.Base, canManage bool) (gin.H, error) {
	structures := []models.UserBaseStructure{}
	if base != nil {
		var err error
		structures, err = s.dbClient.UserBaseStructure().FindByBaseID(ctx, base.ID)
		if err != nil {
			return nil, err
		}
	}

	balances := []models.BaseResourceBalance{}
	activeDailyStates := []models.UserBaseDailyState{}
	if base != nil && canManage {
		var err error
		balances, err = s.dbClient.BaseResourceBalance().FindByUserID(ctx, base.UserID)
		if err != nil {
			return nil, err
		}

		activeDailyStates, err = s.dbClient.UserBaseDailyState().FindActiveByUserID(ctx, base.UserID, time.Now())
		if err != nil {
			return nil, err
		}
	}

	var serializedBase interface{}
	if base != nil {
		serializedBase = serializeBase(base)
	}

	return gin.H{
		"base":               serializedBase,
		"resources":          serializeBaseResourceBalances(balances),
		"structures":         serializeUserBaseStructures(structures),
		"activeDailyEffects": serializeUserBaseDailyStates(activeDailyStates),
		"grassTileUrls":      buildBaseGrassTileURLs(),
		"canManage":          canManage,
	}, nil
}

func (s *server) loadCurrentUserBaseSnapshot(ctx *gin.Context, userID string) (gin.H, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	base, err := s.dbClient.Base().FindByUserID(ctx, parsedUserID)
	if err != nil {
		return nil, err
	}

	return s.loadBaseSnapshot(ctx, base, true)
}

func extractBaseStructurePrerequisites(config models.MetadataJSONB) map[string]int {
	required := map[string]int{}
	raw, ok := config["requiredStructures"]
	if !ok {
		return required
	}
	entries, ok := raw.([]interface{})
	if !ok {
		return required
	}
	for _, entry := range entries {
		mapped, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		key, _ := mapped["key"].(string)
		if key == "" {
			continue
		}
		level := 1
		switch rawLevel := mapped["level"].(type) {
		case float64:
			level = int(rawLevel)
		case int:
			level = rawLevel
		}
		if level < 1 {
			level = 1
		}
		required[key] = level
	}
	return required
}

func levelCostsForTargetLevel(definition *models.BaseStructureDefinition, targetLevel int) []models.BaseResourceDelta {
	costs := []models.BaseResourceDelta{}
	if definition == nil || targetLevel <= 0 {
		return costs
	}
	for _, cost := range definition.LevelCosts {
		if cost.Level != targetLevel || cost.Amount <= 0 {
			continue
		}
		costs = append(costs, models.BaseResourceDelta{
			ResourceKey: cost.ResourceKey,
			Amount:      cost.Amount,
		})
	}
	return costs
}

func isWithinBaseGrid(gridX int, gridY int) bool {
	return gridX >= 0 && gridX < baseGridSize && gridY >= 0 && gridY < baseGridSize
}

func isAdjacentBaseCell(a models.BaseGridPosition, b models.BaseGridPosition) bool {
	dx := a.GridX - b.GridX
	if dx < 0 {
		dx = -dx
	}
	dy := a.GridY - b.GridY
	if dy < 0 {
		dy = -dy
	}
	return dx+dy == 1
}

func isConnectedBaseLayout(positions map[string]models.BaseGridPosition) bool {
	if len(positions) <= 1 {
		return true
	}
	startKey := ""
	if _, ok := positions["hearth"]; ok {
		startKey = "hearth"
	} else {
		for key := range positions {
			startKey = key
			break
		}
	}
	if startKey == "" {
		return true
	}
	visited := map[string]bool{}
	queue := []string{startKey}
	visited[startKey] = true
	for len(queue) > 0 {
		currentKey := queue[0]
		queue = queue[1:]
		currentPosition := positions[currentKey]
		for otherKey, otherPosition := range positions {
			if visited[otherKey] {
				continue
			}
			if isAdjacentBaseCell(currentPosition, otherPosition) {
				visited[otherKey] = true
				queue = append(queue, otherKey)
			}
		}
	}
	return len(visited) == len(positions)
}

func canBuildBaseStructureAt(structures []models.UserBaseStructure, gridX int, gridY int) bool {
	if !isWithinBaseGrid(gridX, gridY) {
		return false
	}
	target := models.BaseGridPosition{GridX: gridX, GridY: gridY}
	hasNeighbor := false
	for _, structure := range structures {
		if structure.GridX == gridX && structure.GridY == gridY {
			return false
		}
		if isAdjacentBaseCell(
			target,
			models.BaseGridPosition{GridX: structure.GridX, GridY: structure.GridY},
		) {
			hasNeighbor = true
		}
	}
	return hasNeighbor
}

func projectedBaseMovePositions(
	structures []models.UserBaseStructure,
	selectedKeys []string,
	anchorStructureKey string,
	targetGridX int,
	targetGridY int,
) (map[string]models.BaseGridPosition, error) {
	if !isWithinBaseGrid(targetGridX, targetGridY) {
		return nil, fmt.Errorf("target position is outside the base grid")
	}
	if len(selectedKeys) == 0 {
		return nil, fmt.Errorf("at least one structure is required")
	}

	structureByKey := map[string]models.UserBaseStructure{}
	for _, structure := range structures {
		structureByKey[structure.StructureKey] = structure
	}
	anchor, ok := structureByKey[anchorStructureKey]
	if !ok {
		return nil, fmt.Errorf("anchor structure was not found")
	}

	selectedSet := map[string]bool{}
	for _, key := range selectedKeys {
		if key == "" {
			continue
		}
		if _, exists := structureByKey[key]; !exists {
			return nil, fmt.Errorf("structure %s was not found", key)
		}
		selectedSet[key] = true
	}
	selectedSet[anchorStructureKey] = true

	deltaX := targetGridX - anchor.GridX
	deltaY := targetGridY - anchor.GridY
	projected := make(map[string]models.BaseGridPosition, len(structures))
	occupied := map[string]string{}
	for _, structure := range structures {
		position := models.BaseGridPosition{GridX: structure.GridX, GridY: structure.GridY}
		if selectedSet[structure.StructureKey] {
			position = models.BaseGridPosition{
				GridX: structure.GridX + deltaX,
				GridY: structure.GridY + deltaY,
			}
		}
		if !isWithinBaseGrid(position.GridX, position.GridY) {
			return nil, fmt.Errorf("the selected rooms do not fit there")
		}
		cellKey := fmt.Sprintf("%d:%d", position.GridX, position.GridY)
		if existingKey, exists := occupied[cellKey]; exists && existingKey != structure.StructureKey {
			return nil, fmt.Errorf("the selected rooms overlap another room")
		}
		occupied[cellKey] = structure.StructureKey
		projected[structure.StructureKey] = position
	}

	if !isConnectedBaseLayout(projected) {
		return nil, fmt.Errorf("the moved rooms must stay connected to the rest of the base")
	}
	return projected, nil
}

func (s *server) mutateBaseStructure(ctx *gin.Context, isUpgrade bool) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	base, err := s.dbClient.Base().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if base == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need a base before you can build structures"})
		return
	}

	structureKey := ctx.Param("key")
	if structureKey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "structure key is required"})
		return
	}

	definition, err := s.dbClient.BaseStructureDefinition().FindActiveByKey(ctx, structureKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "base structure not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	structures, err := s.dbClient.UserBaseStructure().FindByBaseID(ctx, base.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var buildRequest struct {
		GridX *int `json:"gridX"`
		GridY *int `json:"gridY"`
	}
	if !isUpgrade {
		if err := ctx.ShouldBindJSON(&buildRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	structureLevels := map[string]int{}
	currentLevel := 0
	for _, structure := range structures {
		structureLevels[structure.StructureKey] = structure.Level
		if structure.StructureKey == structureKey {
			currentLevel = structure.Level
		}
	}

	targetLevel := 1
	if isUpgrade {
		if currentLevel <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "structure has not been built yet"})
			return
		}
		targetLevel = currentLevel + 1
	} else if currentLevel > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "structure is already built"})
		return
	} else if buildRequest.GridX == nil || buildRequest.GridY == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "gridX and gridY are required"})
		return
	} else if !canBuildBaseStructureAt(structures, *buildRequest.GridX, *buildRequest.GridY) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "new rooms can only be built on an empty tile adjacent to your existing base"})
		return
	}

	if targetLevel > definition.MaxLevel {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "structure is already at max level"})
		return
	}

	for prereqKey, prereqLevel := range extractBaseStructurePrerequisites(definition.PrereqConfig) {
		if structureLevels[prereqKey] < prereqLevel {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("requires %s level %d", prereqKey, prereqLevel),
			})
			return
		}
	}

	updatedStructure, err := s.dbClient.UserBaseStructure().UpsertLevelWithCost(
		ctx,
		base.ID,
		user.ID,
		structureKey,
		targetLevel,
		levelCostsForTargetLevel(definition, targetLevel),
		buildRequest.GridX,
		buildRequest.GridY,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	snapshot, err := s.loadCurrentUserBaseSnapshot(ctx, user.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	snapshot["structureChanged"] = serializeUserBaseStructure(*updatedStructure)
	ctx.JSON(http.StatusOK, snapshot)
}

func (s *server) getCurrentUserBase(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	snapshot, err := s.loadCurrentUserBaseSnapshot(ctx, user.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, snapshot)
}

func (s *server) getBaseProgression(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	baseID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || baseID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base ID"})
		return
	}

	base, err := s.dbClient.Base().FindByID(ctx, baseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if base == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "base not found"})
		return
	}

	canManage := base.UserID == user.ID
	if !canManage {
		areFriends, err := s.dbClient.Friend().Exists(ctx, user.ID, base.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !areFriends {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "you do not have access to this base"})
			return
		}
	}

	snapshot, err := s.loadBaseSnapshot(ctx, base, canManage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, snapshot)
}

func (s *server) moveBaseLayout(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	base, err := s.dbClient.Base().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if base == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need a base before you can move rooms"})
		return
	}

	var request struct {
		AnchorStructureKey string   `json:"anchorStructureKey"`
		StructureKeys      []string `json:"structureKeys"`
		TargetGridX        int      `json:"targetGridX"`
		TargetGridY        int      `json:"targetGridY"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.AnchorStructureKey = strings.TrimSpace(request.AnchorStructureKey)
	if request.AnchorStructureKey == "" {
		request.AnchorStructureKey = strings.TrimSpace(ctx.Query("anchorStructureKey"))
	}
	if request.AnchorStructureKey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "anchorStructureKey is required"})
		return
	}

	structureKeySet := map[string]bool{}
	selectedKeys := make([]string, 0, len(request.StructureKeys)+1)
	for _, key := range request.StructureKeys {
		key = strings.TrimSpace(key)
		if key == "" || structureKeySet[key] {
			continue
		}
		structureKeySet[key] = true
		selectedKeys = append(selectedKeys, key)
	}
	if !structureKeySet[request.AnchorStructureKey] {
		selectedKeys = append(selectedKeys, request.AnchorStructureKey)
	}
	sort.Strings(selectedKeys)

	structures, err := s.dbClient.UserBaseStructure().FindByBaseID(ctx, base.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	projected, err := projectedBaseMovePositions(
		structures,
		selectedKeys,
		request.AnchorStructureKey,
		request.TargetGridX,
		request.TargetGridY,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.UserBaseStructure().MoveMany(ctx, base.ID, user.ID, projected); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	snapshot, err := s.loadCurrentUserBaseSnapshot(ctx, user.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, snapshot)
}

func (s *server) getCurrentUserBaseResources(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	balances, err := s.dbClient.BaseResourceBalance().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ledger, err := s.dbClient.BaseResourceLedger().ListRecentByUserID(ctx, user.ID, 25)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"balances":     serializeBaseResourceBalances(balances),
		"recentLedger": serializeBaseResourceLedgerEntries(ledger),
	})
}

func (s *server) getBaseCatalog(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	definitions, err := s.dbClient.BaseStructureDefinition().FindAllActive(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"structures": serializeBaseStructureDefinitions(definitions),
	})
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *server) buildBaseStructure(ctx *gin.Context) {
	s.mutateBaseStructure(ctx, false)
}

func (s *server) upgradeBaseStructure(ctx *gin.Context) {
	s.mutateBaseStructure(ctx, true)
}
