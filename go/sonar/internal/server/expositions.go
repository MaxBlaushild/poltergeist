package server

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type expositionUpsertRequest struct {
	ZoneID            string                       `json:"zoneId"`
	ZoneKind          *string                      `json:"zoneKind"`
	PointOfInterestID string                       `json:"pointOfInterestId"`
	Latitude          float64                      `json:"latitude"`
	Longitude         float64                      `json:"longitude"`
	Title             string                       `json:"title"`
	Description       string                       `json:"description"`
	Dialogue          []models.DialogueMessage     `json:"dialogue"`
	ImageURL          string                       `json:"imageUrl"`
	ThumbnailURL      string                       `json:"thumbnailUrl"`
	RewardMode        string                       `json:"rewardMode"`
	RandomRewardSize  string                       `json:"randomRewardSize"`
	RewardExperience  int                          `json:"rewardExperience"`
	RewardGold        int                          `json:"rewardGold"`
	MaterialRewards   []baseMaterialRewardPayload  `json:"materialRewards"`
	ItemRewards       []scenarioRewardItemPayload  `json:"itemRewards"`
	SpellRewards      []scenarioRewardSpellPayload `json:"spellRewards"`
}

func (s *server) parseExpositionDialogue(
	ctx context.Context,
	input []models.DialogueMessage,
) (models.DialogueSequence, error) {
	dialogue := make(models.DialogueSequence, 0, len(input))
	for _, raw := range input {
		text := strings.TrimSpace(raw.Text)
		if text == "" {
			continue
		}
		if raw.CharacterID == nil || *raw.CharacterID == uuid.Nil {
			return nil, fmt.Errorf("each dialogue line requires a characterId")
		}
		character, err := s.dbClient.Character().FindByID(ctx, *raw.CharacterID)
		if err != nil {
			return nil, err
		}
		if character == nil {
			return nil, fmt.Errorf("dialogue references an unknown characterId")
		}
		characterID := *raw.CharacterID
		dialogue = append(dialogue, models.DialogueMessage{
			Speaker:     "character",
			Text:        text,
			Order:       len(dialogue),
			Effect:      models.NormalizeDialogueEffect(string(raw.Effect)),
			CharacterID: &characterID,
		})
	}
	if len(dialogue) == 0 {
		return nil, fmt.Errorf("dialogue is required")
	}
	return dialogue, nil
}

func (s *server) parseExpositionItemRewards(
	input []scenarioRewardItemPayload,
) ([]models.ExpositionItemReward, error) {
	rewards := make([]models.ExpositionItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			return nil, fmt.Errorf("itemRewards require inventoryItemId and positive quantity")
		}
		rewards = append(rewards, models.ExpositionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards, nil
}

func (s *server) parseExpositionSpellRewards(
	ctx context.Context,
	input []scenarioRewardSpellPayload,
) ([]models.ExpositionSpellReward, error) {
	rewards := make([]models.ExpositionSpellReward, 0, len(input))
	for _, reward := range input {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil {
			return nil, fmt.Errorf("invalid spellId")
		}
		spell, err := s.dbClient.Spell().FindByID(ctx, spellID)
		if err != nil {
			return nil, err
		}
		if spell == nil {
			return nil, fmt.Errorf("spellId not found")
		}
		rewards = append(rewards, models.ExpositionSpellReward{SpellID: spellID})
	}
	return rewards, nil
}

func (s *server) parseExpositionUpsertRequest(
	ctx context.Context,
	body expositionUpsertRequest,
) (*models.Exposition, []models.ExpositionItemReward, []models.ExpositionSpellReward, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid zoneId")
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		return nil, nil, nil, fmt.Errorf("title is required")
	}
	dialogue, err := s.parseExpositionDialogue(ctx, body.Dialogue)
	if err != nil {
		return nil, nil, nil, err
	}
	pointOfInterestID, err := parseStandalonePointOfInterestID(body.PointOfInterestID)
	if err != nil {
		return nil, nil, nil, err
	}
	resolvedPointOfInterestID, resolvedLatitude, resolvedLongitude, err := s.resolveStandaloneLocation(
		ctx,
		&zoneID,
		pointOfInterestID,
		body.Latitude,
		body.Longitude,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	if body.RewardExperience < 0 || body.RewardGold < 0 {
		return nil, nil, nil, fmt.Errorf("reward values must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(body.MaterialRewards, "materialRewards")
	if err != nil {
		return nil, nil, nil, err
	}
	itemRewards, err := s.parseExpositionItemRewards(body.ItemRewards)
	if err != nil {
		return nil, nil, nil, err
	}
	spellRewards, err := s.parseExpositionSpellRewards(ctx, body.SpellRewards)
	if err != nil {
		return nil, nil, nil, err
	}
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 ||
			body.RewardGold > 0 ||
			len(materialRewards) > 0 ||
			len(itemRewards) > 0 ||
			len(spellRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	imageURL := strings.TrimSpace(body.ImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if thumbnailURL == "" && imageURL != "" {
		thumbnailURL = imageURL
	}

	exposition := &models.Exposition{
		ZoneID:            zoneID,
		ZoneKind:          normalizeZoneKindRequest(body.ZoneKind),
		PointOfInterestID: resolvedPointOfInterestID,
		Latitude:          resolvedLatitude,
		Longitude:         resolvedLongitude,
		Title:             title,
		Description:       strings.TrimSpace(body.Description),
		Dialogue:          dialogue,
		ImageURL:          imageURL,
		ThumbnailURL:      thumbnailURL,
		RewardMode:        rewardMode,
		RandomRewardSize:  models.NormalizeRandomRewardSize(body.RandomRewardSize),
		RewardExperience:  body.RewardExperience,
		RewardGold:        body.RewardGold,
		MaterialRewards:   materialRewards,
	}
	return exposition, itemRewards, spellRewards, nil
}

func expositionRewardItemsFromExposition(
	rewards []models.ExpositionItemReward,
) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func expositionRewardSpellsFromExposition(
	rewards []models.ExpositionSpellReward,
) []scenarioRewardSpell {
	out := make([]scenarioRewardSpell, 0, len(rewards))
	for _, reward := range rewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		out = append(out, scenarioRewardSpell{SpellID: reward.SpellID})
	}
	return out
}

func (s *server) getAdminExpositions(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	expositions, err := s.dbClient.Exposition().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": expositions})
}

func (s *server) getExposition(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	expositionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition ID"})
		return
	}
	exposition, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exposition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !expositionAvailableForStoryFlags(exposition, activeStoryFlags) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}
	if markerURL := s.resolveSharedContentMapMarkerURL(
		ctx.Request.Context(),
		sharedContentMapMarkerDefinitions[2],
		effectiveContentMapMarkerZoneKind(exposition.ZoneKind, &exposition.Zone),
		exposition.ThumbnailURL,
		contentMapMarkerExistenceCache{},
	); markerURL != "" {
		exposition.ThumbnailURL = markerURL
	}
	ctx.JSON(http.StatusOK, exposition)
}

func (s *server) getExpositionsForZone(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	expositions, err := s.dbClient.Exposition().FindByZoneIDExcludingQuestNodes(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	expositionIDs := make([]uuid.UUID, 0, len(expositions))
	for _, exposition := range expositions {
		if !expositionAvailableForStoryFlags(&exposition, activeStoryFlags) {
			continue
		}
		expositionIDs = append(expositionIDs, exposition.ID)
	}
	completedIDs, err := s.dbClient.Exposition().FindCompletedExpositionIDsByUser(ctx, user.ID, expositionIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	completedSet := make(map[uuid.UUID]struct{}, len(completedIDs))
	for _, id := range completedIDs {
		completedSet[id] = struct{}{}
	}
	response := make([]models.Exposition, 0, len(expositions))
	markerCache := contentMapMarkerExistenceCache{}
	for i := range expositions {
		if !expositionAvailableForStoryFlags(&expositions[i], activeStoryFlags) {
			continue
		}
		if _, completed := completedSet[expositions[i].ID]; completed {
			continue
		}
		if markerURL := s.resolveSharedContentMapMarkerURL(
			ctx.Request.Context(),
			sharedContentMapMarkerDefinitions[2],
			effectiveContentMapMarkerZoneKind(expositions[i].ZoneKind, zone),
			expositions[i].ThumbnailURL,
			markerCache,
		); markerURL != "" {
			expositions[i].ThumbnailURL = markerURL
		}
		response = append(response, expositions[i])
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createExposition(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody expositionUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exposition, itemRewards, spellRewards, err := s.parseExpositionUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exposition.ID = uuid.New()
	exposition.CreatedAt = time.Now()
	exposition.UpdatedAt = exposition.CreatedAt
	if err := s.dbClient.Exposition().Create(ctx, exposition); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Exposition().ReplaceItemRewards(ctx, exposition.ID, itemRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Exposition().ReplaceSpellRewards(ctx, exposition.ID, spellRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.Exposition().FindByID(ctx, exposition.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateExposition(ctx *gin.Context) {
	expositionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition ID"})
		return
	}
	existing, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}

	var requestBody expositionUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exposition, itemRewards, spellRewards, err := s.parseExpositionUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exposition.UpdatedAt = time.Now()
	if err := s.dbClient.Exposition().Update(ctx, expositionID, exposition); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Exposition().ReplaceItemRewards(ctx, expositionID, itemRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Exposition().ReplaceSpellRewards(ctx, expositionID, spellRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteExposition(ctx *gin.Context) {
	expositionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition ID"})
		return
	}
	existing, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}
	linkedToQuestNode, err := s.dbClient.Exposition().IsLinkedToQuestNode(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if linkedToQuestNode {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "exposition is referenced by a quest node and cannot be deleted directly",
		})
		return
	}
	if err := s.dbClient.Exposition().Delete(ctx, expositionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "exposition deleted"})
}

func (s *server) generateExpositionImage(ctx *gin.Context) {
	expositionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition ID"})
		return
	}
	exposition, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exposition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateExpositionImageTaskPayload{
		ExpositionID: expositionID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateExpositionImageTaskType, payloadBytes)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{
		"status":     "queued",
		"exposition": exposition,
	})
}

func (s *server) performExposition(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	expositionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition ID"})
		return
	}
	exposition, err := s.dbClient.Exposition().FindByID(ctx, expositionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exposition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !expositionAvailableForStoryFlags(exposition, activeStoryFlags) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition not found"})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	distance := util.HaversineDistance(
		userLat,
		userLng,
		exposition.Latitude,
		exposition.Longitude,
	)
	if distance > scenarioInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"you must be within %.0f meters of the exposition. Currently %.0f meters away",
				scenarioInteractRadiusMeters,
				distance,
			),
		})
		return
	}

	questTargets, err := s.findMatchingCurrentQuestNodeTargets(
		ctx,
		user.ID,
		func(node *models.QuestNode) bool {
			return node != nil &&
				node.ExpositionID != nil &&
				*node.ExpositionID == exposition.ID
		},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	linkedToQuestNode, err := s.dbClient.Exposition().IsLinkedToQuestNode(ctx, exposition.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if linkedToQuestNode && len(questTargets) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "exposition is linked to a quest node and can only be completed when it is the current objective",
		})
		return
	}

	existingCompletion, err := s.dbClient.Exposition().FindCompletionByUserAndExposition(ctx, user.ID, exposition.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingCompletion != nil && len(questTargets) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exposition already completed"})
		return
	}

	rewardExperience := 0
	rewardGold := 0
	itemsAwarded := []models.ItemAwarded{}
	baseResourcesAwarded := []models.BaseResourceDelta{}
	spellsAwarded := []models.SpellAwarded{}
	awardedRewards := existingCompletion == nil
	if awardedRewards {
		rewardMode := models.NormalizeRewardMode(string(exposition.RewardMode))
		rewardItems := []scenarioRewardItem{}
		rewardSpells := []scenarioRewardSpell{}
		if rewardMode == models.RewardModeRandom {
			plan, _, _, err := s.randomRewardPlanForUser(
				ctx,
				user.ID,
				exposition.RandomRewardSize,
				fmt.Sprintf("exposition:%s:user:%s", exposition.ID, user.ID),
			)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			rewardExperience = plan.Experience
			rewardGold = plan.Gold
			rewardItems = mergeScenarioRewardItems(
				randomRewardPlanToScenarioItems(plan),
				expositionRewardItemsFromExposition(exposition.ItemRewards),
			)
		} else {
			rewardExperience = exposition.RewardExperience
			rewardGold = exposition.RewardGold
			rewardItems = expositionRewardItemsFromExposition(exposition.ItemRewards)
			rewardSpells = expositionRewardSpellsFromExposition(exposition.SpellRewards)
		}
		itemsAwarded, spellsAwarded, err = s.awardScenarioRewards(
			ctx,
			user.ID,
			rewardExperience,
			rewardGold,
			rewardItems,
			rewardSpells,
			nil,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		baseResourcesAwarded, err = s.awardBaseResourcesToUser(
			ctx,
			user.ID,
			resolveBaseMaterialRewards(
				exposition.RewardMode,
				exposition.MaterialRewards,
				fmt.Sprintf("exposition:%s:user:%s:materials", exposition.ID, user.ID),
			),
			"exposition",
			&exposition.ID,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := s.dbClient.Exposition().UpsertCompletion(ctx, user.ID, exposition.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	sharedQuestNodeIDs := map[uuid.UUID]struct{}{}
	completedAt := time.Now()
	for _, target := range questTargets {
		completedNode, err := s.markQuestNodeCompleteForAcceptance(
			ctx,
			target.Quest,
			target.Acceptance,
			target.Node.ID,
			completedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !completedNode {
			continue
		}
		if _, exists := sharedQuestNodeIDs[target.Node.ID]; exists {
			continue
		}
		sharedQuestNodeIDs[target.Node.ID] = struct{}{}
		s.shareQuestNodeCompletionWithEligiblePartyMembers(
			ctx,
			user,
			target.Quest,
			target.Node,
		)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"expositionId":         exposition.ID.String(),
		"successful":           true,
		"title":                exposition.Title,
		"rewardExperience":     rewardExperience,
		"rewardGold":           rewardGold,
		"baseResourcesAwarded": serializeBaseResourceDeltas(baseResourcesAwarded),
		"itemsAwarded":         itemsAwarded,
		"spellsAwarded":        spellsAwarded,
		"awardedRewards":       awardedRewards,
	})
}
