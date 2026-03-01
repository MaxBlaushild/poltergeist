package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type spellEffectPayload struct {
	Type             string                         `json:"type"`
	Amount           *int                           `json:"amount"`
	StatusesToApply  []scenarioFailureStatusPayload `json:"statusesToApply"`
	StatusesToRemove []string                       `json:"statusesToRemove"`
	EffectData       map[string]interface{}         `json:"effectData"`
}

type spellUpsertRequest struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	IconURL       string               `json:"iconUrl"`
	AbilityType   string               `json:"abilityType"`
	EffectText    string               `json:"effectText"`
	SchoolOfMagic string               `json:"schoolOfMagic"`
	ManaCost      int                  `json:"manaCost"`
	Effects       []spellEffectPayload `json:"effects"`
}

type castSpellRequest struct {
	TargetUserID *string `json:"targetUserId"`
}

type castSpellHealResult struct {
	UserID    uuid.UUID `json:"userId"`
	Restored  int       `json:"restored"`
	Health    int       `json:"health"`
	MaxHealth int       `json:"maxHealth"`
}

func normalizeSpellAbilityType(value string) models.SpellAbilityType {
	return models.NormalizeSpellAbilityType(strings.TrimSpace(strings.ToLower(value)))
}

func isSpellOfType(spell *models.Spell, abilityType models.SpellAbilityType) bool {
	if spell == nil {
		return false
	}
	return normalizeSpellAbilityType(string(spell.AbilityType)) == abilityType
}

func normalizeSpellStatusNames(values []string) models.StringArray {
	out := models.StringArray{}
	seen := map[string]bool{}
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, normalized)
	}
	return out
}

func (s *server) parseSpellEffects(input []spellEffectPayload) (models.SpellEffects, error) {
	effects := models.SpellEffects{}
	for index, effectPayload := range input {
		effectType := models.SpellEffectType(strings.TrimSpace(strings.ToLower(effectPayload.Type)))
		if effectType == "" {
			return nil, fmt.Errorf("effects[%d].type is required", index)
		}

		amount := 0
		if effectPayload.Amount != nil {
			amount = *effectPayload.Amount
		}

		statusesToApply, err := parseScenarioFailureStatusTemplates(
			effectPayload.StatusesToApply,
			fmt.Sprintf("effects[%d].statusesToApply", index),
		)
		if err != nil {
			return nil, err
		}
		statusesToRemove := normalizeSpellStatusNames(effectPayload.StatusesToRemove)

		switch effectType {
		case models.SpellEffectTypeDealDamage,
			models.SpellEffectTypeRestoreLifePartyMember,
			models.SpellEffectTypeRestoreLifeAllParty:
			if amount <= 0 {
				return nil, fmt.Errorf("effects[%d].amount must be greater than 0", index)
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			if len(statusesToApply) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToApply is required", index)
			}
		case models.SpellEffectTypeRemoveDetrimental:
			if len(statusesToRemove) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToRemove is required", index)
			}
		default:
			// Allow new effect types without backend changes.
		}

		effects = append(effects, models.SpellEffect{
			Type:             effectType,
			Amount:           amount,
			StatusesToApply:  statusesToApply,
			StatusesToRemove: statusesToRemove,
			EffectData:       effectPayload.EffectData,
		})
	}
	if effects == nil {
		return models.SpellEffects{}, nil
	}
	return effects, nil
}

func (s *server) parseSpellUpsertRequest(body spellUpsertRequest) (*models.Spell, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	rawAbilityType := strings.TrimSpace(strings.ToLower(body.AbilityType))
	abilityType := models.SpellAbilityTypeSpell
	if rawAbilityType != "" {
		if !models.IsValidSpellAbilityType(rawAbilityType) {
			return nil, fmt.Errorf("abilityType must be spell or technique")
		}
		abilityType = models.SpellAbilityType(rawAbilityType)
	}
	schoolOfMagic := strings.TrimSpace(body.SchoolOfMagic)
	if schoolOfMagic == "" {
		return nil, fmt.Errorf("schoolOfMagic is required")
	}
	if body.ManaCost < 0 {
		return nil, fmt.Errorf("manaCost must be zero or greater")
	}
	manaCost := body.ManaCost
	if abilityType == models.SpellAbilityTypeTechnique {
		manaCost = 0
	}

	effects, err := s.parseSpellEffects(body.Effects)
	if err != nil {
		return nil, err
	}

	return &models.Spell{
		Name:                  name,
		Description:           strings.TrimSpace(body.Description),
		IconURL:               strings.TrimSpace(body.IconURL),
		ImageGenerationStatus: models.SpellImageGenerationStatusNone,
		AbilityType:           abilityType,
		EffectText:            strings.TrimSpace(body.EffectText),
		SchoolOfMagic:         schoolOfMagic,
		ManaCost:              manaCost,
		Effects:               effects,
	}, nil
}

func (s *server) getSpells(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spells, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spells)
}

func filterSpellsByType(spells []models.Spell, abilityType models.SpellAbilityType) []models.Spell {
	filtered := make([]models.Spell, 0, len(spells))
	for _, spell := range spells {
		if normalizeSpellAbilityType(string(spell.AbilityType)) != abilityType {
			continue
		}
		filtered = append(filtered, spell)
	}
	return filtered
}

func (s *server) getTechniques(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spells, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, filterSpellsByType(spells, models.SpellAbilityTypeTechnique))
}

func (s *server) getSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	spell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spell)
}

func (s *server) getTechnique(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid technique ID"})
		return
	}

	spell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isSpellOfType(spell, models.SpellAbilityTypeTechnique) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found"})
		return
	}
	ctx.JSON(http.StatusOK, spell)
}

func (s *server) createSpell(ctx *gin.Context) {
	var requestBody spellUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.createSpellWithBoundRequest(ctx, requestBody)
}

func (s *server) createTechnique(ctx *gin.Context) {
	var requestBody spellUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestBody.AbilityType = string(models.SpellAbilityTypeTechnique)
	requestBody.ManaCost = 0
	s.createSpellWithBoundRequest(ctx, requestBody)
}

func (s *server) createSpellWithBoundRequest(ctx *gin.Context, requestBody spellUpsertRequest) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spell, err := s.parseSpellUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if spell.IconURL != "" {
		spell.ImageGenerationStatus = models.SpellImageGenerationStatusComplete
		clearErr := ""
		spell.ImageGenerationError = &clearErr
	} else {
		spell.ImageGenerationStatus = models.SpellImageGenerationStatusNone
	}

	if err := s.dbClient.Spell().Create(ctx, spell); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.Spell().FindByID(ctx, spell.ID)
	if err != nil {
		ctx.JSON(http.StatusCreated, spell)
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateSpellWithBoundRequest(
	ctx *gin.Context,
	spellID uuid.UUID,
	existingSpell *models.Spell,
	requestBody spellUpsertRequest,
) {
	spell, err := s.parseSpellUpsertRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Spell().Update(ctx, spellID, map[string]interface{}{
		"name":         spell.Name,
		"description":  spell.Description,
		"icon_url":     spell.IconURL,
		"ability_type": spell.AbilityType,
		"image_generation_status": func() string {
			if spell.IconURL != "" {
				return models.SpellImageGenerationStatusComplete
			}
			if existingSpell.ImageGenerationStatus == models.SpellImageGenerationStatusQueued ||
				existingSpell.ImageGenerationStatus == models.SpellImageGenerationStatusInProgress {
				return existingSpell.ImageGenerationStatus
			}
			return models.SpellImageGenerationStatusNone
		}(),
		"image_generation_error": func() interface{} {
			if spell.IconURL != "" {
				return ""
			}
			if existingSpell.ImageGenerationStatus == models.SpellImageGenerationStatusQueued ||
				existingSpell.ImageGenerationStatus == models.SpellImageGenerationStatusInProgress {
				return existingSpell.ImageGenerationError
			}
			return ""
		}(),
		"effect_text":     spell.EffectText,
		"school_of_magic": spell.SchoolOfMagic,
		"mana_cost":       spell.ManaCost,
		"effects":         spell.Effects,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": spellID})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	existingSpell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody spellUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.updateSpellWithBoundRequest(ctx, spellID, existingSpell, requestBody)
}

func (s *server) updateTechnique(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	techniqueID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid technique ID"})
		return
	}

	existing, err := s.dbClient.Spell().FindByID(ctx, techniqueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isSpellOfType(existing, models.SpellAbilityTypeTechnique) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found"})
		return
	}

	var requestBody spellUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestBody.AbilityType = string(models.SpellAbilityTypeTechnique)
	requestBody.ManaCost = 0
	s.updateSpellWithBoundRequest(ctx, techniqueID, existing, requestBody)
}

func (s *server) deleteSpell(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Spell().Delete(ctx, spellID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "spell deleted successfully"})
}

func (s *server) deleteTechnique(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	techniqueID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid technique ID"})
		return
	}

	spell, err := s.dbClient.Spell().FindByID(ctx, techniqueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isSpellOfType(spell, models.SpellAbilityTypeTechnique) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found"})
		return
	}

	if err := s.dbClient.Spell().Delete(ctx, techniqueID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "technique deleted successfully"})
}

func (s *server) generateSpellIconForType(
	ctx *gin.Context,
	notFoundLabel string,
	requiredType *models.SpellAbilityType,
) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s ID", notFoundLabel)})
		return
	}

	spell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", notFoundLabel)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requiredType != nil && !isSpellOfType(spell, *requiredType) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", notFoundLabel)})
		return
	}

	if err := s.dbClient.Spell().Update(ctx, spellID, map[string]interface{}{
		"image_generation_status": models.SpellImageGenerationStatusQueued,
		"image_generation_error":  "",
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue spell icon generation: " + err.Error()})
		return
	}

	payload := jobs.GenerateSpellIconTaskPayload{
		SpellID:       spell.ID,
		Name:          spell.Name,
		Description:   spell.Description,
		SchoolOfMagic: spell.SchoolOfMagic,
		EffectText:    spell.EffectText,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateSpellIconTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = s.dbClient.Spell().Update(ctx, spellID, map[string]interface{}{
			"image_generation_status": models.SpellImageGenerationStatusFailed,
			"image_generation_error":  errMsg,
		})
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedSpell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedSpell)
}

func (s *server) generateSpellIcon(ctx *gin.Context) {
	s.generateSpellIconForType(ctx, "spell", nil)
}

func (s *server) generateTechniqueIcon(ctx *gin.Context) {
	abilityType := models.SpellAbilityTypeTechnique
	s.generateSpellIconForType(ctx, "technique", &abilityType)
}

func (s *server) applySpellHealToUser(
	ctx context.Context,
	userID uuid.UUID,
	amount int,
) (restored int, health int, maxHealth int, err error) {
	if amount <= 0 {
		return 0, 0, 0, nil
	}

	stats, maxHealth, _, currentHealth, _, err := s.getScenarioResourceState(ctx, userID)
	if err != nil {
		return 0, 0, 0, err
	}
	if stats.HealthDeficit <= 0 {
		return 0, currentHealth, maxHealth, nil
	}

	restoreAmount := amount
	if restoreAmount > stats.HealthDeficit {
		restoreAmount = stats.HealthDeficit
	}
	if restoreAmount <= 0 {
		return 0, currentHealth, maxHealth, nil
	}

	if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, -restoreAmount, 0); err != nil {
		return 0, 0, 0, err
	}

	currentHealth += restoreAmount
	if currentHealth > maxHealth {
		currentHealth = maxHealth
	}
	return restoreAmount, currentHealth, maxHealth, nil
}

func filterUserSpellsByType(
	userSpells []models.UserSpell,
	abilityType models.SpellAbilityType,
) []models.UserSpell {
	filtered := make([]models.UserSpell, 0, len(userSpells))
	for _, userSpell := range userSpells {
		if normalizeSpellAbilityType(string(userSpell.Spell.AbilityType)) != abilityType {
			continue
		}
		filtered = append(filtered, userSpell)
	}
	return filtered
}

func (s *server) castSpellWithType(ctx *gin.Context, requiredType *models.SpellAbilityType) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	userSpells, err := s.dbClient.UserSpell().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var spellToCast *models.Spell
	for _, userSpell := range userSpells {
		if userSpell.SpellID == spellID {
			spell := userSpell.Spell
			if requiredType != nil && normalizeSpellAbilityType(string(spell.AbilityType)) != *requiredType {
				continue
			}
			spellToCast = &spell
			break
		}
	}
	if spellToCast == nil {
		if requiredType != nil && *requiredType == models.SpellAbilityTypeTechnique {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "technique not found for user"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found for user"})
		return
	}
	abilityType := normalizeSpellAbilityType(string(spellToCast.AbilityType))
	isTechnique := abilityType == models.SpellAbilityTypeTechnique

	targetHealAmount := 0
	groupHealAmount := 0
	for _, effect := range spellToCast.Effects {
		if effect.Amount <= 0 {
			continue
		}
		switch effect.Type {
		case models.SpellEffectTypeRestoreLifePartyMember:
			targetHealAmount += effect.Amount
		case models.SpellEffectTypeRestoreLifeAllParty:
			groupHealAmount += effect.Amount
		}
	}
	if targetHealAmount <= 0 && groupHealAmount <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "this ability has no castable healing effect"})
		return
	}

	var request castSpellRequest
	if err := ctx.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	partyMembers, err := s.dbClient.User().FindPartyMembers(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	allowedTargets := map[uuid.UUID]bool{
		user.ID: true,
	}
	for _, member := range partyMembers {
		allowedTargets[member.ID] = true
	}

	var targetUserID uuid.UUID
	if targetHealAmount > 0 {
		if request.TargetUserID == nil || strings.TrimSpace(*request.TargetUserID) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId is required for targeted heal abilities"})
			return
		}
		targetUserID, err = uuid.Parse(strings.TrimSpace(*request.TargetUserID))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId must be a valid UUID"})
			return
		}
		if !allowedTargets[targetUserID] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId must be in your party"})
			return
		}
	}

	if !isTechnique {
		_, _, _, _, currentMana, err := s.getScenarioResourceState(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if currentMana < spellToCast.ManaCost {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":       "not enough mana",
				"currentMana": currentMana,
				"manaCost":    spellToCast.ManaCost,
			})
			return
		}

		if spellToCast.ManaCost > 0 {
			if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, user.ID, 0, spellToCast.ManaCost); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	healByUser := map[uuid.UUID]int{}
	if targetHealAmount > 0 {
		healByUser[targetUserID] += targetHealAmount
	}
	if groupHealAmount > 0 {
		for recipientID := range allowedTargets {
			healByUser[recipientID] += groupHealAmount
		}
	}

	heals := []castSpellHealResult{}
	for recipientID, totalHeal := range healByUser {
		restored, health, maxHealth, err := s.applySpellHealToUser(ctx, recipientID, totalHeal)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if restored <= 0 {
			continue
		}
		heals = append(heals, castSpellHealResult{
			UserID:    recipientID,
			Restored:  restored,
			Health:    health,
			MaxHealth: maxHealth,
		})
	}

	_, _, maxMana, _, manaAfter, err := s.getScenarioResourceState(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"spellId":     spellToCast.ID,
		"spellName":   spellToCast.Name,
		"abilityType": string(abilityType),
		"manaSpent": func() int {
			if isTechnique {
				return 0
			}
			return spellToCast.ManaCost
		}(),
		"currentMana": manaAfter,
		"maxMana":     maxMana,
		"heals":       heals,
	})
}

func (s *server) castSpell(ctx *gin.Context) {
	s.castSpellWithType(ctx, nil)
}

func (s *server) castTechnique(ctx *gin.Context) {
	abilityType := models.SpellAbilityTypeTechnique
	s.castSpellWithType(ctx, &abilityType)
}

func (s *server) getCurrentUserSpells(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userSpells, err := s.dbClient.UserSpell().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, filterUserSpellsByType(userSpells, models.SpellAbilityTypeSpell))
}

func (s *server) getCurrentUserTechniques(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userSpells, err := s.dbClient.UserSpell().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, filterUserSpellsByType(userSpells, models.SpellAbilityTypeTechnique))
}
