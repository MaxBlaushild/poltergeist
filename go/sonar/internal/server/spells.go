package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
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
	TargetUserID    *string `json:"targetUserId"`
	TargetMonsterID *string `json:"targetMonsterId"`
}

type castSpellHealResult struct {
	UserID    uuid.UUID `json:"userId"`
	Restored  int       `json:"restored"`
	Health    int       `json:"health"`
	MaxHealth int       `json:"maxHealth"`
}

type bulkGenerateSpellsRequest struct {
	Count       int    `json:"count"`
	AbilityType string `json:"abilityType"`
}

type generatedAbilityPayload struct {
	Abilities  []jobs.SpellCreationSpec `json:"abilities"`
	Spells     []jobs.SpellCreationSpec `json:"spells"`
	Techniques []jobs.SpellCreationSpec `json:"techniques"`
}

const generateAbilitiesPromptTemplate = `
You are designing %d original %s for a fantasy action RPG.

Avoid these existing %s names:
%s

Hard constraints:
- Output exactly %d %s.
- Use unique names (2-4 words) not present in the existing list.
- Keep descriptions concise and practical (8-18 words), focused on combat utility.
- Do not reference DnD, tabletop, or copyrighted franchises.
- School of magic must be a concise label.
- %s
- Return JSON only.

Respond as:
{
  "abilities": [
    {
      "name": "string",
      "description": "string",
      "effectText": "string",
      "schoolOfMagic": "string",
      "manaCost": 12
    }
  ]
}
`

var spellBulkSeeds = []jobs.SpellCreationSpec{
	{Name: "Ember Lance", Description: "Launch a focused flame spike that pierces a single target.", EffectText: "A concentrated fire bolt scorches one enemy.", SchoolOfMagic: "Pyromancy", ManaCost: 12, AbilityType: "spell"},
	{Name: "Frostbind", Description: "Coat an enemy in brittle ice that slows aggressive movement.", EffectText: "Frigid chains hinder movement and reaction speed.", SchoolOfMagic: "Cryomancy", ManaCost: 10, AbilityType: "spell"},
	{Name: "Storm Javelin", Description: "Hurl a charged spear of lightning into clustered foes.", EffectText: "Electric impact arcs through nearby enemies.", SchoolOfMagic: "Tempest", ManaCost: 16, AbilityType: "spell"},
	{Name: "Verdant Renewal", Description: "Infuse allies with restorative nature energy over a short duration.", EffectText: "Life energy restores a portion of party vitality.", SchoolOfMagic: "Druidic", ManaCost: 14, AbilityType: "spell"},
	{Name: "Rune Barrier", Description: "Raise a shimmering ward that absorbs incoming magical pressure.", EffectText: "A runic shield dampens incoming spell damage.", SchoolOfMagic: "Abjuration", ManaCost: 18, AbilityType: "spell"},
	{Name: "Nightveil Hex", Description: "Place a lingering curse that erodes confidence and focus.", EffectText: "A shadow curse weakens enemy output over time.", SchoolOfMagic: "Hexcraft", ManaCost: 11, AbilityType: "spell"},
	{Name: "Solar Flare", Description: "Burst radiant light that staggers foes and disrupts channeling.", EffectText: "Blinding radiance interrupts hostile casting.", SchoolOfMagic: "Radiance", ManaCost: 17, AbilityType: "spell"},
	{Name: "Echo Pulse", Description: "Release a resonant wave that destabilizes enemy formations.", EffectText: "A sonic pulse disorients foes in a line.", SchoolOfMagic: "Resonance", ManaCost: 9, AbilityType: "spell"},
}

var techniqueBulkSeeds = []jobs.SpellCreationSpec{
	{Name: "Iron Counter", Description: "Time a precise counterstrike immediately after blocking an attack.", EffectText: "A disciplined counter punishes overextended enemies.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Shadow Step", Description: "Shift your stance to reposition quickly around an opponent.", EffectText: "Rapid footwork grants a superior flanking angle.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Crushing Palm", Description: "Deliver a close-range strike that breaks defensive rhythm.", EffectText: "Impact technique reduces enemy guard stability.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Viper Feint", Description: "Use deceptive movement to draw out and punish a reaction.", EffectText: "A feint opens a brief window for high precision hits.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Guard Breaker", Description: "Commit to a heavy swing built to shatter active defense.", EffectText: "Forceful technique cracks defensive posture.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Whirl Cut", Description: "Spin through nearby enemies with fast consecutive slashes.", EffectText: "Circular blade motion pressures multiple targets.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Falcon Rush", Description: "Explode forward with a burst of momentum into melee range.", EffectText: "A swift gap-closer catches ranged enemies off balance.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
	{Name: "Stone Stance", Description: "Anchor your footing to resist displacement and stagger effects.", EffectText: "Grounded posture improves stability under pressure.", SchoolOfMagic: "Martial", ManaCost: 0, AbilityType: "technique"},
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

func nextUniqueAbilityName(base string, used map[string]struct{}, abilityType models.SpellAbilityType) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		if abilityType == models.SpellAbilityTypeTechnique {
			trimmed = "Technique"
		} else {
			trimmed = "Spell"
		}
	}
	candidate := trimmed
	suffix := 2
	for {
		key := strings.ToLower(strings.TrimSpace(candidate))
		if _, exists := used[key]; !exists {
			used[key] = struct{}{}
			return candidate
		}
		candidate = fmt.Sprintf("%s %d", trimmed, suffix)
		suffix++
	}
}

func clampBulkSpellManaCost(manaCost int, abilityType models.SpellAbilityType) int {
	if abilityType == models.SpellAbilityTypeTechnique {
		return 0
	}
	if manaCost < 0 {
		return 0
	}
	if manaCost > 100 {
		return 100
	}
	return manaCost
}

func sanitizeGeneratedAbilitySpec(spec jobs.SpellCreationSpec, abilityType models.SpellAbilityType) jobs.SpellCreationSpec {
	spec.Name = strings.TrimSpace(spec.Name)
	spec.Description = strings.TrimSpace(spec.Description)
	if spec.Description == "" {
		if abilityType == models.SpellAbilityTypeTechnique {
			spec.Description = "A practical combat maneuver with reliable battlefield utility."
		} else {
			spec.Description = "A focused magical ability with practical battlefield utility."
		}
	}
	spec.EffectText = strings.TrimSpace(spec.EffectText)
	if spec.EffectText == "" {
		spec.EffectText = spec.Description
	}
	spec.SchoolOfMagic = strings.TrimSpace(spec.SchoolOfMagic)
	if spec.SchoolOfMagic == "" {
		if abilityType == models.SpellAbilityTypeTechnique {
			spec.SchoolOfMagic = "Martial"
		} else {
			spec.SchoolOfMagic = "Arcane"
		}
	}
	spec.AbilityType = string(abilityType)
	spec.ManaCost = clampBulkSpellManaCost(spec.ManaCost, abilityType)
	return spec
}

func formatAbilityNamesForPrompt(names []string) string {
	if len(names) == 0 {
		return "(none)"
	}

	sorted := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		normalized := strings.ToLower(trimmed)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		sorted = append(sorted, trimmed)
	}
	sort.Strings(sorted)
	if len(sorted) == 0 {
		return "(none)"
	}

	const maxNames = 200
	limited := sorted
	remaining := 0
	if len(sorted) > maxNames {
		limited = sorted[:maxNames]
		remaining = len(sorted) - maxNames
	}

	var builder strings.Builder
	for _, name := range limited {
		builder.WriteString("- ")
		builder.WriteString(name)
		builder.WriteByte('\n')
	}
	if remaining > 0 {
		builder.WriteString(fmt.Sprintf("- ... and %d more\n", remaining))
	}
	return strings.TrimSpace(builder.String())
}

func parseGeneratedAbilitySpecs(raw string, abilityType models.SpellAbilityType) ([]jobs.SpellCreationSpec, error) {
	payload := extractJSONPayload(raw)
	if payload == "" {
		return nil, fmt.Errorf("empty generation payload")
	}

	var wrapped generatedAbilityPayload
	if err := json.Unmarshal([]byte(payload), &wrapped); err == nil {
		candidates := make([]jobs.SpellCreationSpec, 0, len(wrapped.Abilities)+len(wrapped.Spells)+len(wrapped.Techniques))
		candidates = append(candidates, wrapped.Abilities...)
		if abilityType == models.SpellAbilityTypeTechnique {
			candidates = append(candidates, wrapped.Techniques...)
		} else {
			candidates = append(candidates, wrapped.Spells...)
		}
		if len(candidates) > 0 {
			return candidates, nil
		}
	}

	var list []jobs.SpellCreationSpec
	if err := json.Unmarshal([]byte(payload), &list); err == nil && len(list) > 0 {
		return list, nil
	}
	return nil, fmt.Errorf("invalid ability generation payload")
}

func buildBulkSpellSpecsFromSeeds(
	count int,
	abilityType models.SpellAbilityType,
	usedNames map[string]struct{},
) []jobs.SpellCreationSpec {
	specs := make([]jobs.SpellCreationSpec, 0, count)
	if count <= 0 {
		return specs
	}

	seeds := spellBulkSeeds
	if abilityType == models.SpellAbilityTypeTechnique {
		seeds = techniqueBulkSeeds
	}
	if len(seeds) == 0 {
		return specs
	}

	for i := 0; i < count; i++ {
		spec := sanitizeGeneratedAbilitySpec(seeds[i%len(seeds)], abilityType)
		spec.Name = nextUniqueAbilityName(spec.Name, usedNames, abilityType)
		specs = append(specs, spec)
	}
	return specs
}

func (s *server) generateAbilitySpecsWithLLM(
	count int,
	abilityType models.SpellAbilityType,
	usedNames map[string]struct{},
	existingNames []string,
) ([]jobs.SpellCreationSpec, error) {
	specs := make([]jobs.SpellCreationSpec, 0, count)
	if count <= 0 {
		return specs, nil
	}

	abilityLabel := "spells"
	abilityConstraint := "manaCost must be an integer from 0 to 60."
	if abilityType == models.SpellAbilityTypeTechnique {
		abilityLabel = "techniques"
		abilityConstraint = "manaCost must always be 0 for techniques."
	}

	denyList := make([]string, 0, len(existingNames)+len(usedNames))
	denyList = append(denyList, existingNames...)
	for used := range usedNames {
		denyList = append(denyList, used)
	}

	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts && len(specs) < count; attempt++ {
		remaining := count - len(specs)
		prompt := fmt.Sprintf(
			generateAbilitiesPromptTemplate,
			remaining,
			abilityLabel,
			abilityLabel,
			formatAbilityNamesForPrompt(denyList),
			remaining,
			abilityLabel,
			abilityConstraint,
		)
		answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			continue
		}

		candidates, err := parseGeneratedAbilitySpecs(answer.Answer, abilityType)
		if err != nil {
			continue
		}
		for _, candidate := range candidates {
			if len(specs) >= count {
				break
			}
			candidate = sanitizeGeneratedAbilitySpec(candidate, abilityType)
			if candidate.Name == "" {
				continue
			}
			normalized := strings.ToLower(candidate.Name)
			if _, exists := usedNames[normalized]; exists {
				continue
			}
			usedNames[normalized] = struct{}{}
			denyList = append(denyList, candidate.Name)
			specs = append(specs, candidate)
		}
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("failed to generate %s with llm", abilityLabel)
	}
	return specs, nil
}

func (s *server) buildBulkAbilitySpecs(
	count int,
	abilityType models.SpellAbilityType,
	usedNames map[string]struct{},
	existingNames []string,
) ([]jobs.SpellCreationSpec, string, error) {
	if count <= 0 {
		return []jobs.SpellCreationSpec{}, "none", nil
	}

	specs := make([]jobs.SpellCreationSpec, 0, count)
	source := "seed_generated"
	if s.deepPriest != nil {
		aiSpecs, err := s.generateAbilitySpecsWithLLM(count, abilityType, usedNames, existingNames)
		if err == nil && len(aiSpecs) > 0 {
			specs = append(specs, aiSpecs...)
			source = "ai_generated"
		}
	}

	if remaining := count - len(specs); remaining > 0 {
		fallback := buildBulkSpellSpecsFromSeeds(remaining, abilityType, usedNames)
		specs = append(specs, fallback...)
		if source == "ai_generated" {
			source = "ai_generated_with_seed_fallback"
		}
	}
	if len(specs) == 0 {
		return nil, "none", fmt.Errorf("no abilities prepared for generation")
	}
	if len(specs) > count {
		specs = specs[:count]
	}
	return specs, source, nil
}

func (s *server) setSpellBulkStatus(ctx context.Context, status jobs.SpellBulkStatus) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client unavailable")
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.redisClient.Set(
		ctx,
		jobs.SpellBulkStatusKey(status.JobID),
		payload,
		jobs.SpellBulkStatusTTL,
	).Err()
}

func (s *server) getSpellBulkStatus(ctx context.Context, jobID uuid.UUID) (*jobs.SpellBulkStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client unavailable")
	}
	value, err := s.redisClient.Get(ctx, jobs.SpellBulkStatusKey(jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status jobs.SpellBulkStatus
	if err := json.Unmarshal([]byte(value), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *server) bulkGenerateAbilities(ctx *gin.Context, forcedType *models.SpellAbilityType) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody bulkGenerateSpellsRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Count < 1 || requestBody.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}
	if s.asyncClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "async client unavailable"})
		return
	}
	if s.redisClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "redis client unavailable"})
		return
	}

	abilityType := models.SpellAbilityTypeSpell
	if forcedType != nil {
		abilityType = *forcedType
	} else if strings.TrimSpace(requestBody.AbilityType) != "" {
		if !models.IsValidSpellAbilityType(requestBody.AbilityType) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "abilityType must be spell or technique"})
			return
		}
		abilityType = models.NormalizeSpellAbilityType(requestBody.AbilityType)
	}

	existingSpells, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	existingNames := make([]string, 0, len(existingSpells))
	usedNames := make(map[string]struct{}, len(existingSpells)+requestBody.Count)
	for _, spell := range existingSpells {
		if normalizeSpellAbilityType(string(spell.AbilityType)) != abilityType {
			continue
		}
		name := strings.TrimSpace(spell.Name)
		if name == "" {
			continue
		}
		existingNames = append(existingNames, name)
		usedNames[strings.ToLower(name)] = struct{}{}
	}

	spellSpecs, source, err := s.buildBulkAbilitySpecs(requestBody.Count, abilityType, usedNames, existingNames)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New()
	queuedAt := time.Now().UTC()
	status := jobs.SpellBulkStatus{
		JobID:        jobID,
		Status:       jobs.SpellBulkStatusQueued,
		Source:       source,
		AbilityType:  string(abilityType),
		TotalCount:   len(spellSpecs),
		CreatedCount: 0,
		QueuedAt:     &queuedAt,
		UpdatedAt:    queuedAt,
	}
	if err := s.setSpellBulkStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload := jobs.GenerateSpellsBulkTaskPayload{
		JobID:       jobID,
		Source:      source,
		AbilityType: string(abilityType),
		TotalCount:  len(spellSpecs),
		Spells:      spellSpecs,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateSpellsBulkTaskType, payloadBytes)); err != nil {
		failedAt := time.Now().UTC()
		status.Status = jobs.SpellBulkStatusFailed
		status.Error = err.Error()
		status.CompletedAt = &failedAt
		status.UpdatedAt = failedAt
		_ = s.setSpellBulkStatus(ctx.Request.Context(), status)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"jobId":        status.JobID,
		"status":       status.Status,
		"source":       status.Source,
		"abilityType":  status.AbilityType,
		"totalCount":   status.TotalCount,
		"createdCount": status.CreatedCount,
		"queuedAt":     status.QueuedAt,
		"updatedAt":    status.UpdatedAt,
	})
}

func (s *server) bulkGenerateSpells(ctx *gin.Context) {
	s.bulkGenerateAbilities(ctx, nil)
}

func (s *server) bulkGenerateTechniques(ctx *gin.Context) {
	abilityType := models.SpellAbilityTypeTechnique
	s.bulkGenerateAbilities(ctx, &abilityType)
}

func (s *server) getBulkGenerateSpellsStatus(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	jobID, err := uuid.Parse(ctx.Param("jobId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	status, err := s.getSpellBulkStatus(ctx.Request.Context(), jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "bulk generation job not found"})
		return
	}
	ctx.JSON(http.StatusOK, status)
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
	statusesToApply := models.ScenarioFailureStatusTemplates{}
	statusNamesToRemove := make([]string, 0)
	for _, effect := range spellToCast.Effects {
		switch effect.Type {
		case models.SpellEffectTypeRestoreLifePartyMember:
			if effect.Amount > 0 {
				targetHealAmount += effect.Amount
			}
		case models.SpellEffectTypeRestoreLifeAllParty:
			if effect.Amount > 0 {
				groupHealAmount += effect.Amount
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			statusesToApply = append(statusesToApply, effect.StatusesToApply...)
		case models.SpellEffectTypeRemoveDetrimental:
			statusNamesToRemove = append(statusNamesToRemove, effect.StatusesToRemove...)
		}
	}
	statusesToRemove := normalizeSpellStatusNames(statusNamesToRemove)
	hasStatusEffects := len(statusesToApply) > 0 || len(statusesToRemove) > 0

	if targetHealAmount <= 0 && groupHealAmount <= 0 && !hasStatusEffects {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "this ability has no castable effect"})
		return
	}

	var request castSpellRequest
	if err := ctx.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var targetUserID uuid.UUID
	hasTargetUserID := false
	if request.TargetUserID != nil && strings.TrimSpace(*request.TargetUserID) != "" {
		targetUserID, err = uuid.Parse(strings.TrimSpace(*request.TargetUserID))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId must be a valid UUID"})
			return
		}
		hasTargetUserID = true
	}
	if targetHealAmount > 0 && !hasTargetUserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId is required for targeted heal abilities"})
		return
	}

	var targetMonsterID *uuid.UUID
	if request.TargetMonsterID != nil && strings.TrimSpace(*request.TargetMonsterID) != "" {
		parsedMonsterID, parseErr := uuid.Parse(strings.TrimSpace(*request.TargetMonsterID))
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetMonsterId must be a valid UUID"})
			return
		}
		if _, err := s.dbClient.Monster().FindByID(ctx, parsedMonsterID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "target monster not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		targetMonsterID = &parsedMonsterID
	}

	allowedTargets := map[uuid.UUID]bool{
		user.ID: true,
	}
	if targetHealAmount > 0 || groupHealAmount > 0 || hasTargetUserID {
		partyMembers, err := s.dbClient.User().FindPartyMembers(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, member := range partyMembers {
			allowedTargets[member.ID] = true
		}
	}
	if hasTargetUserID {
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

	appliedUserStatuses := []scenarioAppliedFailureStatus{}
	removedUserStatuses := []string{}
	appliedMonsterStatuses := []scenarioAppliedFailureStatus{}
	removedMonsterStatuses := []string{}
	var monsterBattleID *uuid.UUID

	if hasStatusEffects {
		now := time.Now()
		if targetMonsterID != nil {
			monsterBattle, err := s.getOrCreateActiveMonsterBattle(ctx, user.ID, *targetMonsterID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			monsterBattleID = &monsterBattle.ID
			if err := s.dbClient.MonsterBattle().Touch(ctx, monsterBattle.ID, now); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, statusTemplate := range statusesToApply {
				name := strings.TrimSpace(statusTemplate.Name)
				if name == "" || statusTemplate.DurationSeconds <= 0 {
					continue
				}
				status := &models.MonsterStatus{
					UserID:          user.ID,
					BattleID:        monsterBattle.ID,
					MonsterID:       *targetMonsterID,
					Name:            name,
					Description:     strings.TrimSpace(statusTemplate.Description),
					Effect:          strings.TrimSpace(statusTemplate.Effect),
					Positive:        statusTemplate.Positive,
					EffectType:      models.MonsterStatusEffectTypeStatModifier,
					StrengthMod:     statusTemplate.StrengthMod,
					DexterityMod:    statusTemplate.DexterityMod,
					ConstitutionMod: statusTemplate.ConstitutionMod,
					IntelligenceMod: statusTemplate.IntelligenceMod,
					WisdomMod:       statusTemplate.WisdomMod,
					CharismaMod:     statusTemplate.CharismaMod,
					StartedAt:       now,
					ExpiresAt:       now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
				}
				if err := s.dbClient.MonsterStatus().Create(ctx, status); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				appliedMonsterStatuses = append(appliedMonsterStatuses, scenarioAppliedFailureStatus{
					Name:            status.Name,
					Description:     status.Description,
					Effect:          status.Effect,
					Positive:        status.Positive,
					DurationSeconds: statusTemplate.DurationSeconds,
				})
			}

			if len(statusesToRemove) > 0 {
				if err := s.dbClient.MonsterStatus().DeleteActiveByBattleIDAndNames(ctx, monsterBattle.ID, []string(statusesToRemove)); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				removedMonsterStatuses = append(removedMonsterStatuses, []string(statusesToRemove)...)
			}
		} else {
			statusTargetUserID := user.ID
			if hasTargetUserID {
				statusTargetUserID = targetUserID
			}
			for _, statusTemplate := range statusesToApply {
				name := strings.TrimSpace(statusTemplate.Name)
				if name == "" || statusTemplate.DurationSeconds <= 0 {
					continue
				}
				status := &models.UserStatus{
					UserID:          statusTargetUserID,
					Name:            name,
					Description:     strings.TrimSpace(statusTemplate.Description),
					Effect:          strings.TrimSpace(statusTemplate.Effect),
					Positive:        statusTemplate.Positive,
					EffectType:      models.UserStatusEffectTypeStatModifier,
					StrengthMod:     statusTemplate.StrengthMod,
					DexterityMod:    statusTemplate.DexterityMod,
					ConstitutionMod: statusTemplate.ConstitutionMod,
					IntelligenceMod: statusTemplate.IntelligenceMod,
					WisdomMod:       statusTemplate.WisdomMod,
					CharismaMod:     statusTemplate.CharismaMod,
					StartedAt:       now,
					ExpiresAt:       now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
				}
				if err := s.dbClient.UserStatus().Create(ctx, status); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				appliedUserStatuses = append(appliedUserStatuses, scenarioAppliedFailureStatus{
					Name:            status.Name,
					Description:     status.Description,
					Effect:          status.Effect,
					Positive:        status.Positive,
					DurationSeconds: statusTemplate.DurationSeconds,
				})
			}

			if len(statusesToRemove) > 0 {
				if err := s.dbClient.UserStatus().DeleteActiveByUserIDAndNames(ctx, statusTargetUserID, []string(statusesToRemove)); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				removedUserStatuses = append(removedUserStatuses, []string(statusesToRemove)...)
			}
		}
	}

	_, _, maxMana, _, manaAfter, err := s.getScenarioResourceState(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
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
	}
	if hasTargetUserID {
		response["targetUserId"] = targetUserID
	}
	if targetMonsterID != nil {
		response["targetMonsterId"] = *targetMonsterID
	}
	if monsterBattleID != nil {
		response["battleId"] = *monsterBattleID
	}
	if len(appliedUserStatuses) > 0 {
		response["userStatusesApplied"] = appliedUserStatuses
	}
	if len(removedUserStatuses) > 0 {
		response["userStatusesRemoved"] = removedUserStatuses
	}
	if len(appliedMonsterStatuses) > 0 {
		response["monsterStatusesApplied"] = appliedMonsterStatuses
	}
	if len(removedMonsterStatuses) > 0 {
		response["monsterStatusesRemoved"] = removedMonsterStatuses
	}

	ctx.JSON(http.StatusOK, response)
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
