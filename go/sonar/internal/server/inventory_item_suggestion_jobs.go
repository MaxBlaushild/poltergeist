package server

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type inventoryProgressionBand struct {
	Label       string
	TargetLevel int
	RarityTier  string
	RecipeTier  int
}

var inventoryProgressionBands = []inventoryProgressionBand{
	{Label: "Minor", TargetLevel: 8, RarityTier: "Common", RecipeTier: 1},
	{Label: "Lesser", TargetLevel: 23, RarityTier: "Common", RecipeTier: 1},
	{Label: "Greater", TargetLevel: 40, RarityTier: "Uncommon", RecipeTier: 2},
	{Label: "Major", TargetLevel: 60, RarityTier: "Epic", RecipeTier: 3},
	{Label: "Superior", TargetLevel: 78, RarityTier: "Epic", RecipeTier: 4},
	{Label: "Superb", TargetLevel: 93, RarityTier: "Mythic", RecipeTier: 5},
}

type inventoryItemUpsertRequest struct {
	Name                                     string                         `json:"name"`
	Archived                                 *bool                          `json:"archived"`
	GenreID                                  string                         `json:"genreId"`
	ImageURL                                 string                         `json:"imageUrl"`
	FlavorText                               string                         `json:"flavorText"`
	EffectText                               string                         `json:"effectText"`
	RarityTier                               string                         `json:"rarityTier"`
	ResourceTypeID                           *string                        `json:"resourceTypeId"`
	ResourceType                             *string                        `json:"resourceType"`
	IsCaptureType                            bool                           `json:"isCaptureType"`
	BuyPrice                                 *int                           `json:"buyPrice"`
	UnlockTier                               *int                           `json:"unlockTier"`
	UnlockLocksStrength                      *int                           `json:"unlockLocksStrength"`
	ItemLevel                                *int                           `json:"itemLevel"`
	EquipSlot                                *string                        `json:"equipSlot"`
	StrengthMod                              int                            `json:"strengthMod"`
	DexterityMod                             int                            `json:"dexterityMod"`
	ConstitutionMod                          int                            `json:"constitutionMod"`
	IntelligenceMod                          int                            `json:"intelligenceMod"`
	WisdomMod                                int                            `json:"wisdomMod"`
	CharismaMod                              int                            `json:"charismaMod"`
	PhysicalDamageBonusPercent               int                            `json:"physicalDamageBonusPercent"`
	PiercingDamageBonusPercent               int                            `json:"piercingDamageBonusPercent"`
	SlashingDamageBonusPercent               int                            `json:"slashingDamageBonusPercent"`
	BludgeoningDamageBonusPercent            int                            `json:"bludgeoningDamageBonusPercent"`
	FireDamageBonusPercent                   int                            `json:"fireDamageBonusPercent"`
	IceDamageBonusPercent                    int                            `json:"iceDamageBonusPercent"`
	LightningDamageBonusPercent              int                            `json:"lightningDamageBonusPercent"`
	PoisonDamageBonusPercent                 int                            `json:"poisonDamageBonusPercent"`
	ArcaneDamageBonusPercent                 int                            `json:"arcaneDamageBonusPercent"`
	HolyDamageBonusPercent                   int                            `json:"holyDamageBonusPercent"`
	ShadowDamageBonusPercent                 int                            `json:"shadowDamageBonusPercent"`
	PhysicalResistancePercent                int                            `json:"physicalResistancePercent"`
	PiercingResistancePercent                int                            `json:"piercingResistancePercent"`
	SlashingResistancePercent                int                            `json:"slashingResistancePercent"`
	BludgeoningResistancePercent             int                            `json:"bludgeoningResistancePercent"`
	FireResistancePercent                    int                            `json:"fireResistancePercent"`
	IceResistancePercent                     int                            `json:"iceResistancePercent"`
	LightningResistancePercent               int                            `json:"lightningResistancePercent"`
	PoisonResistancePercent                  int                            `json:"poisonResistancePercent"`
	ArcaneResistancePercent                  int                            `json:"arcaneResistancePercent"`
	HolyResistancePercent                    int                            `json:"holyResistancePercent"`
	ShadowResistancePercent                  int                            `json:"shadowResistancePercent"`
	HandItemCategory                         *string                        `json:"handItemCategory"`
	Handedness                               *string                        `json:"handedness"`
	DamageMin                                *int                           `json:"damageMin"`
	DamageMax                                *int                           `json:"damageMax"`
	DamageAffinity                           *string                        `json:"damageAffinity"`
	SwipesPerAttack                          *int                           `json:"swipesPerAttack"`
	BlockPercentage                          *int                           `json:"blockPercentage"`
	DamageBlocked                            *int                           `json:"damageBlocked"`
	SpellDamageBonusPercent                  *int                           `json:"spellDamageBonusPercent"`
	ConsumeHealthDelta                       int                            `json:"consumeHealthDelta"`
	ConsumeManaDelta                         int                            `json:"consumeManaDelta"`
	ConsumeRevivePartyMemberHealth           int                            `json:"consumeRevivePartyMemberHealth"`
	ConsumeReviveAllDownedPartyMembersHealth int                            `json:"consumeReviveAllDownedPartyMembersHealth"`
	ConsumeDealDamage                        int                            `json:"consumeDealDamage"`
	ConsumeDealDamageHits                    *int                           `json:"consumeDealDamageHits"`
	ConsumeDealDamageAllEnemies              int                            `json:"consumeDealDamageAllEnemies"`
	ConsumeDealDamageAllEnemiesHits          *int                           `json:"consumeDealDamageAllEnemiesHits"`
	ConsumeCreateBase                        bool                           `json:"consumeCreateBase"`
	ConsumeStatusesToAdd                     []scenarioFailureStatusPayload `json:"consumeStatusesToAdd"`
	ConsumeStatusesToRemove                  []string                       `json:"consumeStatusesToRemove"`
	ConsumeSpellIDs                          []string                       `json:"consumeSpellIds"`
	ConsumeTeachRecipeIDs                    []string                       `json:"consumeTeachRecipeIds"`
	AlchemyRecipes                           []inventoryRecipePayload       `json:"alchemyRecipes"`
	WorkshopRecipes                          []inventoryRecipePayload       `json:"workshopRecipes"`
	InternalTags                             []string                       `json:"internalTags"`
}

type inventoryItemSuggestionJobRequest struct {
	Count        int      `json:"count"`
	GenreID      string   `json:"genreId"`
	ThemePrompt  string   `json:"themePrompt"`
	Categories   []string `json:"categories"`
	RarityTiers  []string `json:"rarityTiers"`
	EquipSlots   []string `json:"equipSlots"`
	StatTags     []string `json:"statTags"`
	BenefitTags  []string `json:"benefitTags"`
	StatusNames  []string `json:"statusNames"`
	InternalTags []string `json:"internalTags"`
	MinItemLevel *int     `json:"minItemLevel"`
	MaxItemLevel *int     `json:"maxItemLevel"`
}

func (s *server) createInventoryItemSuggestionJob(ctx *gin.Context) {
	var body inventoryItemSuggestionJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Count <= 0 {
		body.Count = 12
	}
	if body.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}

	minLevel := 1
	if body.MinItemLevel != nil && *body.MinItemLevel > 0 {
		minLevel = *body.MinItemLevel
	}
	maxLevel := 100
	if body.MaxItemLevel != nil && *body.MaxItemLevel > 0 {
		maxLevel = *body.MaxItemLevel
	}
	if maxLevel < minLevel {
		maxLevel = minLevel
	}
	genre, err := s.resolveZoneGenre(ctx, body.GenreID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &models.InventoryItemSuggestionJob{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GenreID:      genre.ID,
		Genre:        genre,
		Status:       models.InventoryItemSuggestionJobStatusQueued,
		Count:        body.Count,
		ThemePrompt:  strings.TrimSpace(body.ThemePrompt),
		Categories:   normalizeInventorySuggestionCategories(body.Categories),
		RarityTiers:  normalizeInventorySuggestionRarityList(body.RarityTiers),
		EquipSlots:   normalizeInventorySuggestionEquipSlots(body.EquipSlots),
		StatTags:     normalizeInventorySuggestionStatTags(body.StatTags),
		BenefitTags:  normalizeInventorySuggestionBenefitTags(body.BenefitTags),
		StatusNames:  normalizeInventorySuggestionStatusNames(body.StatusNames),
		InternalTags: parseInventoryInternalTags(body.InternalTags),
		MinItemLevel: minLevel,
		MaxItemLevel: maxLevel,
		CreatedCount: 0,
	}
	if err := s.dbClient.InventoryItemSuggestionJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateInventoryItemSuggestionsTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateInventoryItemSuggestionsTaskType, payload)); err != nil {
		msg := err.Error()
		job.Status = models.InventoryItemSuggestionJobStatusFailed
		job.ErrorMessage = &msg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.InventoryItemSuggestionJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getInventoryItemSuggestionJobs(ctx *gin.Context) {
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	jobsList, err := s.dbClient.InventoryItemSuggestionJob().FindRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getInventoryItemSuggestionJob(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item suggestion job ID"})
		return
	}
	job, err := s.dbClient.InventoryItemSuggestionJob().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item suggestion job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) getInventoryItemSuggestionDrafts(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item suggestion job ID"})
		return
	}
	drafts, err := s.dbClient.InventoryItemSuggestionDraft().FindByJobID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, drafts)
}

func (s *server) deleteInventoryItemSuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.InventoryItemSuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item suggestion draft not found"})
		return
	}
	if draft.InventoryItemID != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "converted drafts cannot be deleted"})
		return
	}
	if err := s.dbClient.InventoryItemSuggestionDraft().Delete(ctx, draftID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "inventory item suggestion draft deleted"})
}

func (s *server) convertInventoryItemSuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.InventoryItemSuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item suggestion draft not found"})
		return
	}
	if draft.InventoryItemID != nil {
		existing, findErr := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, *draft.InventoryItemID)
		if findErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": findErr.Error()})
			return
		}
		ctx.JSON(http.StatusOK, existing)
		return
	}

	item, err := s.materializeInventoryItemSuggestionDraft(ctx, draft)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (s *server) generateInventoryItemProgressionDrafts(ctx *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid inventory item ID"})
		return
	}

	sourceItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sourceItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	}

	sourceCategory := inferInventorySuggestionCategory(sourceItem)
	minLevel := inventoryProgressionBands[0].TargetLevel
	maxLevel := inventoryProgressionBands[len(inventoryProgressionBands)-1].TargetLevel
	job := &models.InventoryItemSuggestionJob{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GenreID:      sourceItem.GenreID,
		Genre:        sourceItem.Genre,
		Status:       models.InventoryItemSuggestionJobStatusCompleted,
		Count:        maxInt(0, len(inventoryProgressionBands)-1),
		ThemePrompt:  fmt.Sprintf("Progression of %s", sourceItem.Name),
		Categories:   models.StringArray{sourceCategory},
		RarityTiers:  models.StringArray{},
		EquipSlots:   normalizeInventorySuggestionEquipSlots([]string{valueOrEmpty(sourceItem.EquipSlot)}),
		InternalTags: parseInventoryInternalTags(sourceItem.InternalTags),
		MinItemLevel: minLevel,
		MaxItemLevel: maxLevel,
		CreatedCount: 0,
	}
	if err := s.dbClient.InventoryItemSuggestionJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sourceBandIndex := nearestInventoryProgressionBandIndex(sourceItem.ItemLevel)
	createdCount := 0
	for bandIndex, band := range inventoryProgressionBands {
		if bandIndex == sourceBandIndex {
			continue
		}
		draftItem, warnings, buildErr := s.buildInventoryProgressionDraftItem(ctx, sourceItem, band)
		if buildErr != nil {
			msg := buildErr.Error()
			job.Status = models.InventoryItemSuggestionJobStatusFailed
			job.ErrorMessage = &msg
			job.CreatedCount = createdCount
			job.UpdatedAt = time.Now()
			_ = s.dbClient.InventoryItemSuggestionJob().Update(ctx, job)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": buildErr.Error()})
			return
		}
		draft := &models.InventoryItemSuggestionDraft{
			ID:           uuid.New(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			JobID:        job.ID,
			Status:       models.InventoryItemSuggestionDraftStatusSuggested,
			Name:         draftItem.Name,
			Category:     sourceCategory,
			RarityTier:   draftItem.RarityTier,
			ItemLevel:    draftItem.ItemLevel,
			EquipSlot:    draftItem.EquipSlot,
			WhyItFits:    fmt.Sprintf("Scaled from %s for the %s level band.", sourceItem.Name, strings.ToLower(band.Label)),
			InternalTags: draftItem.InternalTags,
			Warnings:     dedupeInventorySuggestionWarnings(warnings),
			Payload: models.InventoryItemSuggestionPayloadValue{
				Category:  sourceCategory,
				WhyItFits: fmt.Sprintf("Scaled from %s for the %s level band.", sourceItem.Name, strings.ToLower(band.Label)),
				Item:      *draftItem,
			},
		}
		if err := s.dbClient.InventoryItemSuggestionDraft().Create(ctx, draft); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		createdCount++
	}

	job.CreatedCount = createdCount
	job.UpdatedAt = time.Now()
	if err := s.dbClient.InventoryItemSuggestionJob().Update(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) materializeInventoryItemSuggestionDraft(
	ctx *gin.Context,
	draft *models.InventoryItemSuggestionDraft,
) (*models.InventoryItem, error) {
	if draft == nil {
		return nil, fmt.Errorf("draft is required")
	}

	request := inventoryItemUpsertRequestFromDraftPayload(models.InventoryItemSuggestionPayload(draft.Payload).Item)
	item, err := s.normalizeInventoryItemUpsertRequest(ctx, request, nil)
	if err != nil {
		return nil, err
	}
	if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}
	if item.ImageURL == "" {
		if err := s.enqueueInventoryItemImageGeneration(ctx, item.ID, item.Name, item.FlavorText, item.RarityTier); err != nil {
			return nil, fmt.Errorf("failed to queue item image generation: %w", err)
		}
	}

	now := time.Now()
	draft.Status = models.InventoryItemSuggestionDraftStatusConverted
	draft.InventoryItemID = &item.ID
	draft.ConvertedAt = &now
	draft.UpdatedAt = now
	draft.InventoryItem = item
	if err := s.dbClient.InventoryItemSuggestionDraft().Update(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to update inventory item suggestion draft: %w", err)
	}
	return item, nil
}

func normalizeInventorySuggestionCategories(input []string) models.StringArray {
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		normalized := strings.ToLower(strings.TrimSpace(raw))
		switch normalized {
		case "equippable", "consumable", "material", "utility":
		default:
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeInventorySuggestionStatTags(input []string) models.StringArray {
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		normalized := strings.ToLower(strings.TrimSpace(raw))
		switch normalized {
		case "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma":
		default:
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeInventorySuggestionBenefitTags(input []string) models.StringArray {
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		normalized := strings.ToLower(strings.TrimSpace(raw))
		normalized = strings.ReplaceAll(normalized, " ", "_")
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeInventorySuggestionStatusNames(input []string) models.StringArray {
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeInventorySuggestionRarityList(input []string) models.StringArray {
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		var normalized string
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "common":
			normalized = "Common"
		case "uncommon":
			normalized = "Uncommon"
		case "epic":
			normalized = "Epic"
		case "mythic":
			normalized = "Mythic"
		case "not droppable":
			normalized = "Not Droppable"
		default:
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeInventorySuggestionEquipSlots(input []string) models.StringArray {
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		normalized := strings.TrimSpace(raw)
		if normalized == "" || !models.IsValidInventoryEquipSlot(normalized) {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func clearGeneratedInventoryUnlockLocksStrength(item *models.InventoryItem) {
	if item == nil {
		return
	}
	item.UnlockLocksStrength = nil
}

func inventoryItemUpsertRequestFromDraftPayload(item models.InventoryItem) inventoryItemUpsertRequest {
	var resourceTypeID *string
	var genreID string
	if item.ResourceTypeID != nil {
		resourceTypeID = stringPtr(item.ResourceTypeID.String())
	}
	if item.GenreID != uuid.Nil {
		genreID = item.GenreID.String()
	}
	clearGeneratedInventoryUnlockLocksStrength(&item)
	return inventoryItemUpsertRequest{
		Name:                                     item.Name,
		GenreID:                                  genreID,
		ImageURL:                                 item.ImageURL,
		FlavorText:                               item.FlavorText,
		EffectText:                               item.EffectText,
		RarityTier:                               item.RarityTier,
		ResourceTypeID:                           resourceTypeID,
		IsCaptureType:                            item.IsCaptureType,
		BuyPrice:                                 item.BuyPrice,
		UnlockTier:                               item.UnlockTier,
		UnlockLocksStrength:                      nil,
		ItemLevel:                                intPtr(item.ItemLevel),
		EquipSlot:                                item.EquipSlot,
		StrengthMod:                              item.StrengthMod,
		DexterityMod:                             item.DexterityMod,
		ConstitutionMod:                          item.ConstitutionMod,
		IntelligenceMod:                          item.IntelligenceMod,
		WisdomMod:                                item.WisdomMod,
		CharismaMod:                              item.CharismaMod,
		PhysicalDamageBonusPercent:               item.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:               item.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:               item.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent:            item.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:                   item.FireDamageBonusPercent,
		IceDamageBonusPercent:                    item.IceDamageBonusPercent,
		LightningDamageBonusPercent:              item.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:                 item.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:                 item.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:                   item.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:                 item.ShadowDamageBonusPercent,
		PhysicalResistancePercent:                item.PhysicalResistancePercent,
		PiercingResistancePercent:                item.PiercingResistancePercent,
		SlashingResistancePercent:                item.SlashingResistancePercent,
		BludgeoningResistancePercent:             item.BludgeoningResistancePercent,
		FireResistancePercent:                    item.FireResistancePercent,
		IceResistancePercent:                     item.IceResistancePercent,
		LightningResistancePercent:               item.LightningResistancePercent,
		PoisonResistancePercent:                  item.PoisonResistancePercent,
		ArcaneResistancePercent:                  item.ArcaneResistancePercent,
		HolyResistancePercent:                    item.HolyResistancePercent,
		ShadowResistancePercent:                  item.ShadowResistancePercent,
		HandItemCategory:                         item.HandItemCategory,
		Handedness:                               item.Handedness,
		DamageMin:                                item.DamageMin,
		DamageMax:                                item.DamageMax,
		DamageAffinity:                           item.DamageAffinity,
		SwipesPerAttack:                          item.SwipesPerAttack,
		BlockPercentage:                          item.BlockPercentage,
		DamageBlocked:                            item.DamageBlocked,
		SpellDamageBonusPercent:                  item.SpellDamageBonusPercent,
		ConsumeHealthDelta:                       item.ConsumeHealthDelta,
		ConsumeManaDelta:                         item.ConsumeManaDelta,
		ConsumeRevivePartyMemberHealth:           item.ConsumeRevivePartyMemberHealth,
		ConsumeReviveAllDownedPartyMembersHealth: item.ConsumeReviveAllDownedPartyMembersHealth,
		ConsumeDealDamage:                        item.ConsumeDealDamage,
		ConsumeDealDamageHits:                    zeroIntToNil(item.ConsumeDealDamageHits),
		ConsumeDealDamageAllEnemies:              item.ConsumeDealDamageAllEnemies,
		ConsumeDealDamageAllEnemiesHits:          zeroIntToNil(item.ConsumeDealDamageAllEnemiesHits),
		ConsumeCreateBase:                        item.ConsumeCreateBase,
		ConsumeStatusesToAdd:                     inventoryItemStatusPayloadsFromTemplates(item.ConsumeStatusesToAdd),
		ConsumeStatusesToRemove:                  append([]string{}, item.ConsumeStatusesToRemove...),
		ConsumeSpellIDs:                          append([]string{}, item.ConsumeSpellIDs...),
		ConsumeTeachRecipeIDs:                    append([]string{}, item.ConsumeTeachRecipeIDs...),
		AlchemyRecipes:                           inventoryRecipePayloadsFromModels(item.AlchemyRecipes),
		WorkshopRecipes:                          inventoryRecipePayloadsFromModels(item.WorkshopRecipes),
		InternalTags:                             append([]string{}, item.InternalTags...),
	}
}

func (s *server) buildInventoryProgressionDraftItem(
	ctx *gin.Context,
	sourceItem *models.InventoryItem,
	band inventoryProgressionBand,
) (*models.InventoryItem, []string, error) {
	if sourceItem == nil {
		return nil, nil, fmt.Errorf("source item is required")
	}
	ratio := inventoryProgressionScaleRatio(sourceItem.ItemLevel, band.TargetLevel)
	recipeIDMap := map[string]string{}
	item := *sourceItem
	item.ID = 0
	item.CreatedAt = time.Time{}
	item.UpdatedAt = time.Time{}
	item.Archived = false
	item.ImageURL = ""
	item.ImageGenerationStatus = models.InventoryImageGenerationStatusNone
	item.ImageGenerationError = nil
	item.Name = inventoryProgressionBandName(sourceItem.Name, band.Label)
	item.RarityTier = band.RarityTier
	item.ItemLevel = band.TargetLevel
	item.UnlockTier = intPtr(band.TargetLevel)
	sourceBuyPrice := 0
	if item.BuyPrice != nil {
		sourceBuyPrice = *item.BuyPrice
	}
	item.BuyPrice = intPtr(scalePositiveValue(sourceBuyPrice, ratio, 0, 1000000))
	clearGeneratedInventoryUnlockLocksStrength(&item)
	item.StrengthMod = scaleSignedValue(item.StrengthMod, ratio)
	item.DexterityMod = scaleSignedValue(item.DexterityMod, ratio)
	item.ConstitutionMod = scaleSignedValue(item.ConstitutionMod, ratio)
	item.IntelligenceMod = scaleSignedValue(item.IntelligenceMod, ratio)
	item.WisdomMod = scaleSignedValue(item.WisdomMod, ratio)
	item.CharismaMod = scaleSignedValue(item.CharismaMod, ratio)
	item.PhysicalDamageBonusPercent = scaleSignedValue(item.PhysicalDamageBonusPercent, ratio)
	item.PiercingDamageBonusPercent = scaleSignedValue(item.PiercingDamageBonusPercent, ratio)
	item.SlashingDamageBonusPercent = scaleSignedValue(item.SlashingDamageBonusPercent, ratio)
	item.BludgeoningDamageBonusPercent = scaleSignedValue(item.BludgeoningDamageBonusPercent, ratio)
	item.FireDamageBonusPercent = scaleSignedValue(item.FireDamageBonusPercent, ratio)
	item.IceDamageBonusPercent = scaleSignedValue(item.IceDamageBonusPercent, ratio)
	item.LightningDamageBonusPercent = scaleSignedValue(item.LightningDamageBonusPercent, ratio)
	item.PoisonDamageBonusPercent = scaleSignedValue(item.PoisonDamageBonusPercent, ratio)
	item.ArcaneDamageBonusPercent = scaleSignedValue(item.ArcaneDamageBonusPercent, ratio)
	item.HolyDamageBonusPercent = scaleSignedValue(item.HolyDamageBonusPercent, ratio)
	item.ShadowDamageBonusPercent = scaleSignedValue(item.ShadowDamageBonusPercent, ratio)
	item.PhysicalResistancePercent = scaleSignedValue(item.PhysicalResistancePercent, ratio)
	item.PiercingResistancePercent = scaleSignedValue(item.PiercingResistancePercent, ratio)
	item.SlashingResistancePercent = scaleSignedValue(item.SlashingResistancePercent, ratio)
	item.BludgeoningResistancePercent = scaleSignedValue(item.BludgeoningResistancePercent, ratio)
	item.FireResistancePercent = scaleSignedValue(item.FireResistancePercent, ratio)
	item.IceResistancePercent = scaleSignedValue(item.IceResistancePercent, ratio)
	item.LightningResistancePercent = scaleSignedValue(item.LightningResistancePercent, ratio)
	item.PoisonResistancePercent = scaleSignedValue(item.PoisonResistancePercent, ratio)
	item.ArcaneResistancePercent = scaleSignedValue(item.ArcaneResistancePercent, ratio)
	item.HolyResistancePercent = scaleSignedValue(item.HolyResistancePercent, ratio)
	item.ShadowResistancePercent = scaleSignedValue(item.ShadowResistancePercent, ratio)
	item.DamageMin = scaleOptionalInt(item.DamageMin, ratio, 1, 1000000)
	item.DamageMax = scaleOptionalInt(item.DamageMax, ratio, 1, 1000000)
	if item.DamageMin != nil && item.DamageMax != nil && *item.DamageMax < *item.DamageMin {
		item.DamageMax = intPtr(*item.DamageMin)
	}
	item.SwipesPerAttack = scaleOptionalInt(item.SwipesPerAttack, math.Sqrt(ratio), 1, 8)
	item.BlockPercentage = scaleOptionalInt(item.BlockPercentage, ratio, 1, 100)
	item.DamageBlocked = scaleOptionalInt(item.DamageBlocked, ratio, 1, 1000000)
	item.SpellDamageBonusPercent = scaleOptionalInt(item.SpellDamageBonusPercent, ratio, 1, 1000000)
	item.ConsumeHealthDelta = scaleSignedValue(item.ConsumeHealthDelta, ratio)
	item.ConsumeManaDelta = scaleSignedValue(item.ConsumeManaDelta, ratio)
	item.ConsumeRevivePartyMemberHealth = scalePositiveValue(item.ConsumeRevivePartyMemberHealth, ratio, 1, 1000000)
	item.ConsumeReviveAllDownedPartyMembersHealth = scalePositiveValue(item.ConsumeReviveAllDownedPartyMembersHealth, ratio, 1, 1000000)
	item.ConsumeDealDamage = scalePositiveValue(item.ConsumeDealDamage, ratio, 1, 1000000)
	item.ConsumeDealDamageHits = scalePositiveValue(item.ConsumeDealDamageHits, math.Sqrt(ratio), 1, 12)
	item.ConsumeDealDamageAllEnemies = scalePositiveValue(item.ConsumeDealDamageAllEnemies, ratio, 1, 1000000)
	item.ConsumeDealDamageAllEnemiesHits = scalePositiveValue(item.ConsumeDealDamageAllEnemiesHits, math.Sqrt(ratio), 1, 12)
	item.ConsumeStatusesToAdd = scaleInventorySuggestionStatuses(item.ConsumeStatusesToAdd, ratio)
	item.AlchemyRecipes, recipeIDMap = cloneAndScaleInventoryRecipes(item.AlchemyRecipes, ratio, band.RecipeTier)
	workshopRecipes, workshopMap := cloneAndScaleInventoryRecipes(item.WorkshopRecipes, ratio, band.RecipeTier)
	item.WorkshopRecipes = workshopRecipes
	for oldID, newID := range workshopMap {
		recipeIDMap[oldID] = newID
	}
	item.ConsumeTeachRecipeIDs = remapTeachRecipeIDs(item.ConsumeTeachRecipeIDs, recipeIDMap)
	item.InternalTags = parseInventoryInternalTags(append(item.InternalTags, "progression_draft", strings.ToLower(band.Label)))
	item.FlavorText = inventoryProgressionFlavorText(sourceItem.FlavorText, band)
	item.EffectText = inventoryProgressionEffectText(sourceItem.EffectText, band)

	request := inventoryItemUpsertRequestFromDraftPayload(item)
	normalized, err := s.normalizeInventoryItemUpsertRequest(ctx, request, nil)
	if err != nil {
		return nil, nil, err
	}
	warnings := []string{fmt.Sprintf("Generated from %s for target level %d.", sourceItem.Name, band.TargetLevel)}
	return normalized, warnings, nil
}

func nearestInventoryProgressionBandIndex(level int) int {
	if len(inventoryProgressionBands) == 0 {
		return 0
	}
	bestIndex := 0
	bestDistance := inventoryProgressionAbsInt(level - inventoryProgressionBands[0].TargetLevel)
	for idx := 1; idx < len(inventoryProgressionBands); idx++ {
		distance := inventoryProgressionAbsInt(level - inventoryProgressionBands[idx].TargetLevel)
		if distance < bestDistance {
			bestDistance = distance
			bestIndex = idx
		}
	}
	return bestIndex
}

func inventoryProgressionScaleRatio(sourceLevel int, targetLevel int) float64 {
	base := maxInt(1, sourceLevel)
	target := maxInt(1, targetLevel)
	return float64(target) / float64(base)
}

func inventoryProgressionBandName(name string, label string) string {
	baseName := strings.TrimSpace(name)
	for _, band := range inventoryProgressionBands {
		prefix := strings.ToLower(strings.TrimSpace(band.Label)) + " "
		if strings.HasPrefix(strings.ToLower(baseName), prefix) {
			baseName = strings.TrimSpace(baseName[len(prefix):])
			break
		}
	}
	if baseName == "" {
		baseName = "Item"
	}
	return fmt.Sprintf("%s %s", label, baseName)
}

func inventoryProgressionFlavorText(base string, band inventoryProgressionBand) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		return fmt.Sprintf("A %s piece tuned for adventurers around level %d.", strings.ToLower(band.Label), band.TargetLevel)
	}
	return fmt.Sprintf("%s Tuned for the %s band.", trimmed, strings.ToLower(band.Label))
}

func inventoryProgressionEffectText(base string, band inventoryProgressionBand) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		return fmt.Sprintf("Scaled for level %d performance.", band.TargetLevel)
	}
	return fmt.Sprintf("%s Scaled for level %d.", trimmed, band.TargetLevel)
}

func scaleSignedValue(value int, ratio float64) int {
	if value == 0 {
		return 0
	}
	scaled := int(math.Round(float64(value) * ratio))
	if scaled == 0 {
		if value > 0 {
			return 1
		}
		return -1
	}
	return scaled
}

func scalePositiveValue(value int, ratio float64, minValue int, maxValue int) int {
	if value <= 0 {
		return 0
	}
	scaled := int(math.Round(float64(value) * ratio))
	if scaled < minValue {
		scaled = minValue
	}
	if maxValue > 0 && scaled > maxValue {
		scaled = maxValue
	}
	return scaled
}

func scaleOptionalInt(value *int, ratio float64, minValue int, maxValue int) *int {
	if value == nil {
		return nil
	}
	scaled := scalePositiveValue(*value, ratio, minValue, maxValue)
	return intPtr(scaled)
}

func scaleInventorySuggestionStatuses(
	input models.ScenarioFailureStatusTemplates,
	ratio float64,
) models.ScenarioFailureStatusTemplates {
	if len(input) == 0 {
		return models.ScenarioFailureStatusTemplates{}
	}
	out := make(models.ScenarioFailureStatusTemplates, 0, len(input))
	for _, status := range input {
		scaled := status
		scaled.DamagePerTick = scalePositiveValue(status.DamagePerTick, ratio, 1, 1000000)
		scaled.HealthPerTick = scaleSignedValue(status.HealthPerTick, ratio)
		scaled.ManaPerTick = scaleSignedValue(status.ManaPerTick, ratio)
		scaled.DurationSeconds = scalePositiveValue(status.DurationSeconds, math.Sqrt(ratio), 1, 86400)
		scaled.StrengthMod = scaleSignedValue(status.StrengthMod, ratio)
		scaled.DexterityMod = scaleSignedValue(status.DexterityMod, ratio)
		scaled.ConstitutionMod = scaleSignedValue(status.ConstitutionMod, ratio)
		scaled.IntelligenceMod = scaleSignedValue(status.IntelligenceMod, ratio)
		scaled.WisdomMod = scaleSignedValue(status.WisdomMod, ratio)
		scaled.CharismaMod = scaleSignedValue(status.CharismaMod, ratio)
		scaled.PhysicalDamageBonusPercent = scaleSignedValue(status.PhysicalDamageBonusPercent, ratio)
		scaled.PiercingDamageBonusPercent = scaleSignedValue(status.PiercingDamageBonusPercent, ratio)
		scaled.SlashingDamageBonusPercent = scaleSignedValue(status.SlashingDamageBonusPercent, ratio)
		scaled.BludgeoningDamageBonusPercent = scaleSignedValue(status.BludgeoningDamageBonusPercent, ratio)
		scaled.FireDamageBonusPercent = scaleSignedValue(status.FireDamageBonusPercent, ratio)
		scaled.IceDamageBonusPercent = scaleSignedValue(status.IceDamageBonusPercent, ratio)
		scaled.LightningDamageBonusPercent = scaleSignedValue(status.LightningDamageBonusPercent, ratio)
		scaled.PoisonDamageBonusPercent = scaleSignedValue(status.PoisonDamageBonusPercent, ratio)
		scaled.ArcaneDamageBonusPercent = scaleSignedValue(status.ArcaneDamageBonusPercent, ratio)
		scaled.HolyDamageBonusPercent = scaleSignedValue(status.HolyDamageBonusPercent, ratio)
		scaled.ShadowDamageBonusPercent = scaleSignedValue(status.ShadowDamageBonusPercent, ratio)
		scaled.PhysicalResistancePercent = scaleSignedValue(status.PhysicalResistancePercent, ratio)
		scaled.PiercingResistancePercent = scaleSignedValue(status.PiercingResistancePercent, ratio)
		scaled.SlashingResistancePercent = scaleSignedValue(status.SlashingResistancePercent, ratio)
		scaled.BludgeoningResistancePercent = scaleSignedValue(status.BludgeoningResistancePercent, ratio)
		scaled.FireResistancePercent = scaleSignedValue(status.FireResistancePercent, ratio)
		scaled.IceResistancePercent = scaleSignedValue(status.IceResistancePercent, ratio)
		scaled.LightningResistancePercent = scaleSignedValue(status.LightningResistancePercent, ratio)
		scaled.PoisonResistancePercent = scaleSignedValue(status.PoisonResistancePercent, ratio)
		scaled.ArcaneResistancePercent = scaleSignedValue(status.ArcaneResistancePercent, ratio)
		scaled.HolyResistancePercent = scaleSignedValue(status.HolyResistancePercent, ratio)
		scaled.ShadowResistancePercent = scaleSignedValue(status.ShadowResistancePercent, ratio)
		out = append(out, scaled)
	}
	return out
}

func cloneAndScaleInventoryRecipes(
	input models.InventoryRecipes,
	ratio float64,
	recipeTier int,
) (models.InventoryRecipes, map[string]string) {
	if len(input) == 0 {
		return models.InventoryRecipes{}, map[string]string{}
	}
	out := make(models.InventoryRecipes, 0, len(input))
	idMap := map[string]string{}
	for _, recipe := range input {
		newID := uuid.NewString()
		idMap[recipe.ID] = newID
		ingredients := make([]models.InventoryRecipeIngredient, 0, len(recipe.Ingredients))
		for _, ingredient := range recipe.Ingredients {
			ingredients = append(ingredients, models.InventoryRecipeIngredient{
				ItemID:   ingredient.ItemID,
				Quantity: scalePositiveValue(ingredient.Quantity, math.Sqrt(ratio), 1, 999),
			})
		}
		out = append(out, models.InventoryRecipe{
			ID:          newID,
			Tier:        maxInt(1, recipeTier),
			IsPublic:    recipe.IsPublic,
			Ingredients: ingredients,
		})
	}
	return out, idMap
}

func remapTeachRecipeIDs(input models.StringArray, idMap map[string]string) models.StringArray {
	if len(input) == 0 {
		return models.StringArray{}
	}
	out := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, recipeID := range input {
		trimmed := strings.TrimSpace(recipeID)
		if trimmed == "" {
			continue
		}
		mapped := idMap[trimmed]
		if mapped == "" {
			mapped = trimmed
		}
		if _, exists := seen[mapped]; exists {
			continue
		}
		seen[mapped] = struct{}{}
		out = append(out, mapped)
	}
	return out
}

func inferInventorySuggestionCategory(item *models.InventoryItem) string {
	if item == nil {
		return "material"
	}
	if item.EquipSlot != nil && strings.TrimSpace(*item.EquipSlot) != "" {
		return "equippable"
	}
	if item.ConsumeCreateBase ||
		item.ConsumeHealthDelta != 0 ||
		item.ConsumeManaDelta != 0 ||
		item.ConsumeRevivePartyMemberHealth > 0 ||
		item.ConsumeReviveAllDownedPartyMembersHealth > 0 ||
		item.ConsumeDealDamage > 0 ||
		item.ConsumeDealDamageAllEnemies > 0 ||
		len(item.ConsumeStatusesToAdd) > 0 ||
		len(item.ConsumeStatusesToRemove) > 0 {
		return "consumable"
	}
	if item.IsCaptureType || item.UnlockLocksStrength != nil {
		return "utility"
	}
	return "material"
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func dedupeInventorySuggestionWarnings(input []string) models.StringArray {
	seen := map[string]struct{}{}
	out := make(models.StringArray, 0, len(input))
	for _, raw := range input {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func inventoryProgressionAbsInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (s *server) normalizeInventoryItemUpsertRequest(
	ctx *gin.Context,
	requestBody inventoryItemUpsertRequest,
	existingItem *models.InventoryItem,
) (*models.InventoryItem, error) {
	requestBody.Name = strings.TrimSpace(requestBody.Name)
	requestBody.FlavorText = strings.TrimSpace(requestBody.FlavorText)
	requestBody.EffectText = strings.TrimSpace(requestBody.EffectText)
	requestBody.RarityTier = strings.TrimSpace(requestBody.RarityTier)
	if requestBody.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if requestBody.RarityTier == "" {
		return nil, fmt.Errorf("rarityTier is required")
	}
	genreID := uuid.Nil
	var genre *models.ZoneGenre
	if existingItem != nil {
		genreID = existingItem.GenreID
		genre = existingItem.Genre
	}
	if strings.TrimSpace(requestBody.GenreID) != "" || genreID == uuid.Nil {
		resolvedGenre, err := s.resolveZoneGenre(ctx, requestBody.GenreID)
		if err != nil {
			return nil, err
		}
		genreID = resolvedGenre.ID
		genre = resolvedGenre
	}

	itemLevel := 1
	if existingItem != nil && existingItem.ItemLevel > 0 {
		itemLevel = existingItem.ItemLevel
	}
	if requestBody.ItemLevel != nil {
		itemLevel = *requestBody.ItemLevel
	}
	if itemLevel < 1 {
		return nil, fmt.Errorf("itemLevel must be 1 or greater")
	}
	if requestBody.BuyPrice != nil && *requestBody.BuyPrice < 0 {
		return nil, fmt.Errorf("buyPrice must be 0 or greater")
	}
	var resourceTypeID *uuid.UUID
	var resourceType *models.ResourceType
	if existingItem != nil {
		resourceTypeID = existingItem.ResourceTypeID
		resourceType = existingItem.ResourceType
	}
	if requestBody.ResourceTypeID != nil || requestBody.ResourceType != nil {
		resolvedResourceTypeID, resolvedResourceType, err := s.resolveResourceTypeReference(
			ctx,
			requestBody.ResourceTypeID,
			requestBody.ResourceType,
		)
		if err != nil {
			return nil, err
		}
		resourceTypeID = resolvedResourceTypeID
		resourceType = resolvedResourceType
	}
	if requestBody.UnlockLocksStrength != nil &&
		(*requestBody.UnlockLocksStrength < 1 || *requestBody.UnlockLocksStrength > 100) {
		return nil, fmt.Errorf("unlockLocksStrength must be between 1 and 100")
	}

	consumeDealDamageHits := 0
	if requestBody.ConsumeDealDamage > 0 {
		consumeDealDamageHits = 1
		if requestBody.ConsumeDealDamageHits != nil {
			consumeDealDamageHits = *requestBody.ConsumeDealDamageHits
		}
		if consumeDealDamageHits < 1 {
			return nil, fmt.Errorf("consumeDealDamageHits must be 1 or greater when consumeDealDamage is set")
		}
	}

	consumeDealDamageAllEnemiesHits := 0
	if requestBody.ConsumeDealDamageAllEnemies > 0 {
		consumeDealDamageAllEnemiesHits = 1
		if requestBody.ConsumeDealDamageAllEnemiesHits != nil {
			consumeDealDamageAllEnemiesHits = *requestBody.ConsumeDealDamageAllEnemiesHits
		}
		if consumeDealDamageAllEnemiesHits < 1 {
			return nil, fmt.Errorf("consumeDealDamageAllEnemiesHits must be 1 or greater when consumeDealDamageAllEnemies is set")
		}
	}

	var equipSlot *string
	if requestBody.EquipSlot != nil {
		trimmed := strings.TrimSpace(*requestBody.EquipSlot)
		if trimmed != "" {
			if !models.IsValidInventoryEquipSlot(trimmed) {
				return nil, fmt.Errorf("invalid equip slot")
			}
			equipSlot = &trimmed
		}
	}

	handAttrsInput := models.HandEquipmentAttributes{
		HandItemCategory:        requestBody.HandItemCategory,
		Handedness:              requestBody.Handedness,
		DamageMin:               requestBody.DamageMin,
		DamageMax:               requestBody.DamageMax,
		DamageAffinity:          requestBody.DamageAffinity,
		SwipesPerAttack:         requestBody.SwipesPerAttack,
		BlockPercentage:         requestBody.BlockPercentage,
		DamageBlocked:           requestBody.DamageBlocked,
		SpellDamageBonusPercent: requestBody.SpellDamageBonusPercent,
	}
	handAttrsInput = mergeGeneratedInventoryItemHandDefaults(equipSlot, requestBody.RarityTier, handAttrsInput)
	handAttrs, err := models.NormalizeAndValidateHandEquipment(equipSlot, handAttrsInput)
	if err != nil {
		return nil, err
	}

	consumeStatusesToAdd, err := parseScenarioFailureStatusTemplates(requestBody.ConsumeStatusesToAdd, "consumeStatusesToAdd")
	if err != nil {
		return nil, err
	}
	consumeStatusesToRemove := parseInventoryConsumeStatusNames(requestBody.ConsumeStatusesToRemove)
	consumeSpellIDs, err := parseInventoryConsumeSpellIDs(requestBody.ConsumeSpellIDs)
	if err != nil {
		return nil, err
	}
	var currentItemID *int
	if existingItem != nil {
		currentItemID = &existingItem.ID
	}
	alchemyRecipes, workshopRecipes, consumeTeachRecipeIDs, err := s.parseInventoryRecipeConfiguration(
		ctx,
		requestBody.AlchemyRecipes,
		requestBody.WorkshopRecipes,
		requestBody.ConsumeTeachRecipeIDs,
		currentItemID,
	)
	if err != nil {
		return nil, err
	}
	internalTags := parseInventoryInternalTags(requestBody.InternalTags)
	for idx, rawSpellID := range consumeSpellIDs {
		spellID, _ := uuid.Parse(rawSpellID)
		if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("consumeSpellIds[%d] not found", idx)
			}
			return nil, err
		}
	}

	archived := requestBody.Archived != nil && *requestBody.Archived
	if existingItem != nil {
		archived = existingItem.Archived
		if requestBody.Archived != nil {
			archived = *requestBody.Archived
		}
	}

	item := &models.InventoryItem{
		Archived:                                 archived,
		Name:                                     requestBody.Name,
		GenreID:                                  genreID,
		Genre:                                    genre,
		ImageURL:                                 requestBody.ImageURL,
		FlavorText:                               requestBody.FlavorText,
		EffectText:                               requestBody.EffectText,
		RarityTier:                               requestBody.RarityTier,
		ResourceTypeID:                           resourceTypeID,
		ResourceType:                             resourceType,
		IsCaptureType:                            requestBody.IsCaptureType,
		BuyPrice:                                 requestBody.BuyPrice,
		UnlockTier:                               requestBody.UnlockTier,
		UnlockLocksStrength:                      requestBody.UnlockLocksStrength,
		ItemLevel:                                itemLevel,
		EquipSlot:                                equipSlot,
		StrengthMod:                              requestBody.StrengthMod,
		DexterityMod:                             requestBody.DexterityMod,
		ConstitutionMod:                          requestBody.ConstitutionMod,
		IntelligenceMod:                          requestBody.IntelligenceMod,
		WisdomMod:                                requestBody.WisdomMod,
		CharismaMod:                              requestBody.CharismaMod,
		PhysicalDamageBonusPercent:               requestBody.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:               requestBody.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:               requestBody.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent:            requestBody.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:                   requestBody.FireDamageBonusPercent,
		IceDamageBonusPercent:                    requestBody.IceDamageBonusPercent,
		LightningDamageBonusPercent:              requestBody.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:                 requestBody.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:                 requestBody.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:                   requestBody.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:                 requestBody.ShadowDamageBonusPercent,
		PhysicalResistancePercent:                requestBody.PhysicalResistancePercent,
		PiercingResistancePercent:                requestBody.PiercingResistancePercent,
		SlashingResistancePercent:                requestBody.SlashingResistancePercent,
		BludgeoningResistancePercent:             requestBody.BludgeoningResistancePercent,
		FireResistancePercent:                    requestBody.FireResistancePercent,
		IceResistancePercent:                     requestBody.IceResistancePercent,
		LightningResistancePercent:               requestBody.LightningResistancePercent,
		PoisonResistancePercent:                  requestBody.PoisonResistancePercent,
		ArcaneResistancePercent:                  requestBody.ArcaneResistancePercent,
		HolyResistancePercent:                    requestBody.HolyResistancePercent,
		ShadowResistancePercent:                  requestBody.ShadowResistancePercent,
		HandItemCategory:                         handAttrs.HandItemCategory,
		Handedness:                               handAttrs.Handedness,
		DamageMin:                                handAttrs.DamageMin,
		DamageMax:                                handAttrs.DamageMax,
		DamageAffinity:                           handAttrs.DamageAffinity,
		SwipesPerAttack:                          handAttrs.SwipesPerAttack,
		BlockPercentage:                          handAttrs.BlockPercentage,
		DamageBlocked:                            handAttrs.DamageBlocked,
		SpellDamageBonusPercent:                  handAttrs.SpellDamageBonusPercent,
		ConsumeHealthDelta:                       requestBody.ConsumeHealthDelta,
		ConsumeManaDelta:                         requestBody.ConsumeManaDelta,
		ConsumeRevivePartyMemberHealth:           requestBody.ConsumeRevivePartyMemberHealth,
		ConsumeReviveAllDownedPartyMembersHealth: requestBody.ConsumeReviveAllDownedPartyMembersHealth,
		ConsumeDealDamage:                        requestBody.ConsumeDealDamage,
		ConsumeDealDamageHits:                    consumeDealDamageHits,
		ConsumeDealDamageAllEnemies:              requestBody.ConsumeDealDamageAllEnemies,
		ConsumeDealDamageAllEnemiesHits:          consumeDealDamageAllEnemiesHits,
		ConsumeCreateBase:                        requestBody.ConsumeCreateBase,
		ConsumeStatusesToAdd:                     consumeStatusesToAdd,
		ConsumeStatusesToRemove:                  consumeStatusesToRemove,
		ConsumeSpellIDs:                          consumeSpellIDs,
		ConsumeTeachRecipeIDs:                    consumeTeachRecipeIDs,
		AlchemyRecipes:                           alchemyRecipes,
		WorkshopRecipes:                          workshopRecipes,
		InternalTags:                             internalTags,
		ImageGenerationStatus: func() string {
			if requestBody.ImageURL != "" {
				return models.InventoryImageGenerationStatusComplete
			}
			return models.InventoryImageGenerationStatusNone
		}(),
	}
	return item, nil
}

func mergeGeneratedInventoryItemHandDefaults(
	equipSlot *string,
	rarityTier string,
	input models.HandEquipmentAttributes,
) models.HandEquipmentAttributes {
	if equipSlot == nil || !models.IsHandEquipSlot(strings.TrimSpace(*equipSlot)) {
		return input
	}
	if input.HandItemCategory == nil || input.Handedness == nil {
		return input
	}
	generated := generateInventoryItemHandAttributes(rarityTier, *input.HandItemCategory, *input.Handedness)
	if input.DamageMin == nil {
		input.DamageMin = generated.DamageMin
	}
	if input.DamageMax == nil {
		input.DamageMax = generated.DamageMax
	}
	if input.DamageAffinity == nil {
		input.DamageAffinity = generated.DamageAffinity
	}
	if input.SwipesPerAttack == nil {
		input.SwipesPerAttack = generated.SwipesPerAttack
	}
	if input.BlockPercentage == nil {
		input.BlockPercentage = generated.BlockPercentage
	}
	if input.DamageBlocked == nil {
		input.DamageBlocked = generated.DamageBlocked
	}
	if input.SpellDamageBonusPercent == nil {
		input.SpellDamageBonusPercent = generated.SpellDamageBonusPercent
	}
	return input
}

func inventoryRecipePayloadsFromModels(recipes models.InventoryRecipes) []inventoryRecipePayload {
	out := make([]inventoryRecipePayload, 0, len(recipes))
	for _, recipe := range recipes {
		ingredients := make([]inventoryRecipeIngredientPayload, 0, len(recipe.Ingredients))
		for _, ingredient := range recipe.Ingredients {
			ingredients = append(ingredients, inventoryRecipeIngredientPayload{
				ItemID:   ingredient.ItemID,
				Quantity: ingredient.Quantity,
			})
		}
		out = append(out, inventoryRecipePayload{
			ID:          recipe.ID,
			Tier:        recipe.Tier,
			IsPublic:    recipe.IsPublic,
			Ingredients: ingredients,
		})
	}
	return out
}

func inventoryItemStatusPayloadsFromTemplates(
	templates models.ScenarioFailureStatusTemplates,
) []scenarioFailureStatusPayload {
	out := make([]scenarioFailureStatusPayload, 0, len(templates))
	for _, template := range templates {
		positive := template.Positive
		out = append(out, scenarioFailureStatusPayload{
			Name:                          template.Name,
			Description:                   template.Description,
			Effect:                        template.Effect,
			EffectType:                    string(template.EffectType),
			Positive:                      &positive,
			DamagePerTick:                 template.DamagePerTick,
			HealthPerTick:                 template.HealthPerTick,
			ManaPerTick:                   template.ManaPerTick,
			DurationSeconds:               template.DurationSeconds,
			StrengthMod:                   template.StrengthMod,
			DexterityMod:                  template.DexterityMod,
			ConstitutionMod:               template.ConstitutionMod,
			IntelligenceMod:               template.IntelligenceMod,
			WisdomMod:                     template.WisdomMod,
			CharismaMod:                   template.CharismaMod,
			PhysicalDamageBonusPercent:    template.PhysicalDamageBonusPercent,
			PiercingDamageBonusPercent:    template.PiercingDamageBonusPercent,
			SlashingDamageBonusPercent:    template.SlashingDamageBonusPercent,
			BludgeoningDamageBonusPercent: template.BludgeoningDamageBonusPercent,
			FireDamageBonusPercent:        template.FireDamageBonusPercent,
			IceDamageBonusPercent:         template.IceDamageBonusPercent,
			LightningDamageBonusPercent:   template.LightningDamageBonusPercent,
			PoisonDamageBonusPercent:      template.PoisonDamageBonusPercent,
			ArcaneDamageBonusPercent:      template.ArcaneDamageBonusPercent,
			HolyDamageBonusPercent:        template.HolyDamageBonusPercent,
			ShadowDamageBonusPercent:      template.ShadowDamageBonusPercent,
			PhysicalResistancePercent:     template.PhysicalResistancePercent,
			PiercingResistancePercent:     template.PiercingResistancePercent,
			SlashingResistancePercent:     template.SlashingResistancePercent,
			BludgeoningResistancePercent:  template.BludgeoningResistancePercent,
			FireResistancePercent:         template.FireResistancePercent,
			IceResistancePercent:          template.IceResistancePercent,
			LightningResistancePercent:    template.LightningResistancePercent,
			PoisonResistancePercent:       template.PoisonResistancePercent,
			ArcaneResistancePercent:       template.ArcaneResistancePercent,
			HolyResistancePercent:         template.HolyResistancePercent,
			ShadowResistancePercent:       template.ShadowResistancePercent,
		})
	}
	return out
}

func zeroIntToNil(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}
