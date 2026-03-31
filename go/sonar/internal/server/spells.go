package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

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
	Hits             *int                           `json:"hits"`
	DamageAffinity   *string                        `json:"damageAffinity"`
	StatusesToApply  []scenarioFailureStatusPayload `json:"statusesToApply"`
	StatusesToRemove []string                       `json:"statusesToRemove"`
	EffectData       map[string]interface{}         `json:"effectData"`
}

type spellUpsertRequest struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	IconURL       string               `json:"iconUrl"`
	AbilityType   string               `json:"abilityType"`
	AbilityLevel  *int                 `json:"abilityLevel"`
	CooldownTurns int                  `json:"cooldownTurns"`
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
	Count        int                         `json:"count"`
	AbilityType  string                      `json:"abilityType"`
	TargetLevel  *int                        `json:"targetLevel"`
	EffectCounts *jobs.SpellBulkEffectCounts `json:"effectCounts"`
	// Deprecated: retained for backward compatibility with older clients.
	EffectMix *jobs.SpellBulkEffectCounts `json:"effectMix"`
}

type spellProgressionFromPromptRequest struct {
	Prompt      string `json:"prompt"`
	AbilityType string `json:"abilityType"`
}

type generatedAbilityPayload struct {
	Abilities  []jobs.SpellCreationSpec `json:"abilities"`
	Spells     []jobs.SpellCreationSpec `json:"spells"`
	Techniques []jobs.SpellCreationSpec `json:"techniques"`
}

type generatedSpellProgressionVariantFlavorEnvelope struct {
	Variants []generatedSpellProgressionVariantFlavor `json:"variants"`
}

type generatedSpellProgressionVariantFlavor struct {
	LevelBand    int    `json:"levelBand"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	AbilityLevel int    `json:"abilityLevel"`
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

const generateSpellProgressionVariantFlavorTemplate = `
You are designing missing level-band variants for a fantasy RPG %s progression.

Seed ability:
- Name: %s
- School: %s
- Description: %s
- Effect text: %s

Missing variants to create:
%s

Existing ability names to avoid:
%s

Return JSON only:
{
  "variants": [
    {
      "levelBand": 10,
      "name": "2-4 words",
      "description": "one vivid line, 8-18 words"
    }
  ]
}

Rules:
- Each name should feel related to the seed ability, but not be formulaic.
- Make the names feel like a true progression in intensity across the bands.
- Keep names distinct from each other and from the avoid list.
- Descriptions must be a single flavorful line with no level numbers, no tier labels, and no meta commentary.
- Spell descriptions should feel magical; technique descriptions should feel physical and martial.
- Do not mention progression bands, levels, rarity, or game systems.
- No copyrighted franchises or references.
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

var spellProgressionLevelBands = []int{10, 25, 50, 70}

const (
	spellProgressionPromptMinLength = 12
	spellProgressionPromptMaxLength = 2000
)

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

func isSpellDetrimentalStatusEffectType(effectType models.SpellEffectType) bool {
	switch effectType {
	case models.SpellEffectTypeApplyDetrimentalStatus, models.SpellEffectTypeApplyDetrimentalAll:
		return true
	default:
		return false
	}
}

func normalizeSpellStatusesForEffectType(
	effectType models.SpellEffectType,
	statuses models.ScenarioFailureStatusTemplates,
) models.ScenarioFailureStatusTemplates {
	if len(statuses) == 0 || !isSpellDetrimentalStatusEffectType(effectType) {
		return statuses
	}
	normalized := make(models.ScenarioFailureStatusTemplates, 0, len(statuses))
	for _, status := range statuses {
		next := status
		next.Positive = false
		normalized = append(normalized, next)
	}
	return normalized
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

func extractGeneratedJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```JSON")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func sanitizeGeneratedSpellProgressionVariantName(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.Trim(trimmed, "\"'`")
	trimmed = strings.Join(strings.Fields(trimmed), " ")
	return strings.TrimSpace(trimmed)
}

func sanitizeGeneratedSpellProgressionVariantDescription(value string) string {
	trimmed := stripSpellProgressionMetaSentences(value)
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func parseGeneratedSpellProgressionVariantFlavors(
	raw string,
) (map[int]generatedSpellProgressionVariantFlavor, error) {
	parsed := generatedSpellProgressionVariantFlavorEnvelope{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(raw)), &parsed); err != nil {
		return nil, err
	}

	byBand := make(map[int]generatedSpellProgressionVariantFlavor, len(parsed.Variants))
	for _, variant := range parsed.Variants {
		levelBand := variant.LevelBand
		if levelBand <= 0 {
			levelBand = variant.AbilityLevel
		}
		levelBand = normalizeSpellProgressionBand(levelBand)
		if levelBand <= 0 {
			continue
		}
		variant.LevelBand = levelBand
		variant.Name = sanitizeGeneratedSpellProgressionVariantName(variant.Name)
		variant.Description = sanitizeGeneratedSpellProgressionVariantDescription(variant.Description)
		byBand[levelBand] = variant
	}
	return byBand, nil
}

func formatSpellProgressionVariantPromptLines(
	seed *models.Spell,
	seedBand int,
	targetBands []int,
	abilityType models.SpellAbilityType,
) string {
	lines := make([]string, 0, len(targetBands))
	for _, targetBand := range targetBands {
		targetLevel := spellProgressionTargetLevelForBand(targetBand)
		effects := buildScaledSpellProgressionEffects(seed.Effects, seedBand, targetLevel, abilityType)
		lines = append(lines, fmt.Sprintf(
			"- Band %d: target level %d, effect text: %s",
			targetBand,
			targetLevel,
			buildSpellProgressionEffectText(effects),
		))
	}
	return strings.Join(lines, "\n")
}

func (s *server) generateSpellProgressionVariantFlavors(
	seed *models.Spell,
	seedBand int,
	targetBands []int,
	usedNames map[string]struct{},
	abilityType models.SpellAbilityType,
) map[int]generatedSpellProgressionVariantFlavor {
	if s.deepPriest == nil || seed == nil || len(targetBands) == 0 {
		return nil
	}

	abilityLabel := "spell"
	if abilityType == models.SpellAbilityTypeTechnique {
		abilityLabel = "technique"
	}

	existingNames := make([]string, 0, len(usedNames))
	for usedName := range usedNames {
		if trimmed := strings.TrimSpace(usedName); trimmed != "" {
			existingNames = append(existingNames, trimmed)
		}
	}

	prompt := fmt.Sprintf(
		generateSpellProgressionVariantFlavorTemplate,
		abilityLabel,
		strings.TrimSpace(seed.Name),
		strings.TrimSpace(seed.SchoolOfMagic),
		strings.TrimSpace(seed.Description),
		strings.TrimSpace(seed.EffectText),
		formatSpellProgressionVariantPromptLines(seed, seedBand, targetBands, abilityType),
		formatAbilityNamesForPrompt(existingNames),
	)
	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil
	}

	parsed, err := parseGeneratedSpellProgressionVariantFlavors(answer.Answer)
	if err != nil {
		return nil
	}
	return parsed
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

func clampAbilityLevel(level int) int {
	if level < 1 {
		return 1
	}
	if level > 100 {
		return 100
	}
	return level
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
	spec.AbilityLevel = clampAbilityLevel(spec.AbilityLevel)
	spec.ManaCost = clampBulkSpellManaCost(spec.ManaCost, abilityType)
	return spec
}

func sanitizeBulkAbilityEffectCounts(
	raw *jobs.SpellBulkEffectCounts,
	totalCount int,
) (*jobs.SpellBulkEffectCounts, error) {
	if raw == nil {
		return nil, nil
	}

	sanitized := &jobs.SpellBulkEffectCounts{
		DealDamage:               raw.DealDamage,
		DealDamageAllEnemies:     raw.DealDamageAllEnemies,
		RestoreLifePartyMember:   raw.RestoreLifePartyMember,
		RestoreLifeAllParty:      raw.RestoreLifeAllParty,
		ApplyBeneficialStatuses:  raw.ApplyBeneficialStatuses,
		RemoveDetrimentalEffects: raw.RemoveDetrimentalEffects,
	}

	configuredCounts := []struct {
		label string
		value int
	}{
		{label: "effectCounts.dealDamage", value: sanitized.DealDamage},
		{label: "effectCounts.dealDamageAllEnemies", value: sanitized.DealDamageAllEnemies},
		{label: "effectCounts.restoreLifePartyMember", value: sanitized.RestoreLifePartyMember},
		{label: "effectCounts.restoreLifeAllPartyMembers", value: sanitized.RestoreLifeAllParty},
		{label: "effectCounts.applyBeneficialStatuses", value: sanitized.ApplyBeneficialStatuses},
		{label: "effectCounts.removeDetrimentalStatuses", value: sanitized.RemoveDetrimentalEffects},
	}

	total := 0
	for _, configured := range configuredCounts {
		if configured.value < 0 {
			return nil, fmt.Errorf("%s must be greater than or equal to 0", configured.label)
		}
		if configured.value > totalCount {
			return nil, fmt.Errorf("%s must be less than or equal to count", configured.label)
		}
		total += configured.value
	}
	if total == 0 {
		return nil, fmt.Errorf("effectCounts must include at least one positive value")
	}
	if total != totalCount {
		return nil, fmt.Errorf("effectCounts must add up to count (%d)", totalCount)
	}
	return sanitized, nil
}

func sanitizeBulkAbilityTargetLevel(raw *int) (*int, error) {
	if raw == nil {
		return nil, nil
	}
	value := *raw
	if value < 1 || value > 100 {
		return nil, fmt.Errorf("targetLevel must be between 1 and 100")
	}
	return &value, nil
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

func (s *server) setSpellProgressionPromptStatus(ctx context.Context, status jobs.SpellProgressionPromptStatus) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client unavailable")
	}
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.redisClient.Set(
		ctx,
		jobs.SpellProgressionPromptStatusKey(status.JobID),
		payload,
		jobs.SpellProgressionPromptStatusTTL,
	).Err()
}

func (s *server) getSpellProgressionPromptStatus(
	ctx context.Context,
	jobID uuid.UUID,
) (*jobs.SpellProgressionPromptStatus, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client unavailable")
	}
	value, err := s.redisClient.Get(ctx, jobs.SpellProgressionPromptStatusKey(jobID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status jobs.SpellProgressionPromptStatus
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
	requestedEffectCounts := requestBody.EffectCounts
	if requestedEffectCounts == nil {
		requestedEffectCounts = requestBody.EffectMix
	}
	targetLevel, err := sanitizeBulkAbilityTargetLevel(requestBody.TargetLevel)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	effectCounts, err := sanitizeBulkAbilityEffectCounts(requestedEffectCounts, requestBody.Count)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	for i := range spellSpecs {
		if targetLevel != nil {
			spellSpecs[i].AbilityLevel = *targetLevel
			continue
		}
		spellSpecs[i].AbilityLevel = clampAbilityLevel(spellSpecs[i].AbilityLevel)
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
		TargetLevel:  targetLevel,
		EffectCounts: effectCounts,
		EffectMix:    effectCounts,
		QueuedAt:     &queuedAt,
		UpdatedAt:    queuedAt,
	}
	if err := s.setSpellBulkStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload := jobs.GenerateSpellsBulkTaskPayload{
		JobID:        jobID,
		Source:       source,
		AbilityType:  string(abilityType),
		TotalCount:   len(spellSpecs),
		TargetLevel:  targetLevel,
		EffectCounts: effectCounts,
		EffectMix:    effectCounts,
		Spells:       spellSpecs,
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
		"targetLevel":  status.TargetLevel,
		"effectCounts": status.EffectCounts,
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

func (s *server) queueSpellProgressionFromPromptWithType(
	ctx *gin.Context,
	forcedType *models.SpellAbilityType,
) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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

	var requestBody spellProgressionFromPromptRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prompt := strings.TrimSpace(requestBody.Prompt)
	if len(prompt) < spellProgressionPromptMinLength {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"prompt must be at least %d characters",
				spellProgressionPromptMinLength,
			),
		})
		return
	}
	if len(prompt) > spellProgressionPromptMaxLength {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"prompt must be at most %d characters",
				spellProgressionPromptMaxLength,
			),
		})
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

	jobID := uuid.New()
	queuedAt := time.Now().UTC()
	status := jobs.SpellProgressionPromptStatus{
		JobID:        jobID,
		Status:       jobs.SpellProgressionPromptStatusQueued,
		Prompt:       prompt,
		AbilityType:  string(abilityType),
		CreatedCount: 0,
		QueuedAt:     &queuedAt,
		UpdatedAt:    queuedAt,
	}
	if err := s.setSpellProgressionPromptStatus(ctx.Request.Context(), status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateSpellProgressionFromPromptTaskPayload{
		JobID:       jobID,
		Prompt:      prompt,
		AbilityType: string(abilityType),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateSpellProgressionFromPromptTaskType, payloadBytes)); err != nil {
		failedAt := time.Now().UTC()
		status.Status = jobs.SpellProgressionPromptStatusFailed
		status.Error = err.Error()
		status.CompletedAt = &failedAt
		status.UpdatedAt = failedAt
		_ = s.setSpellProgressionPromptStatus(ctx.Request.Context(), status)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, status)
}

func (s *server) queueSpellProgressionFromPrompt(ctx *gin.Context) {
	abilityType := models.SpellAbilityTypeSpell
	s.queueSpellProgressionFromPromptWithType(ctx, &abilityType)
}

func (s *server) queueTechniqueProgressionFromPrompt(ctx *gin.Context) {
	abilityType := models.SpellAbilityTypeTechnique
	s.queueSpellProgressionFromPromptWithType(ctx, &abilityType)
}

func (s *server) getSpellProgressionFromPromptStatus(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	jobID, err := uuid.Parse(ctx.Param("jobId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	status, err := s.getSpellProgressionPromptStatus(ctx.Request.Context(), jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if status == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "spell progression generation job not found"})
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
		hits := 0
		if effectPayload.Hits != nil {
			hits = *effectPayload.Hits
		}

		statusesToApply, err := parseScenarioFailureStatusTemplates(
			effectPayload.StatusesToApply,
			fmt.Sprintf("effects[%d].statusesToApply", index),
		)
		if err != nil {
			return nil, err
		}
		statusesToRemove := normalizeSpellStatusNames(effectPayload.StatusesToRemove)
		var damageAffinity *string

		switch effectType {
		case models.SpellEffectTypeDealDamage,
			models.SpellEffectTypeDealDamageAllEnemies,
			models.SpellEffectTypeRestoreLifePartyMember,
			models.SpellEffectTypeRestoreLifeAllParty,
			models.SpellEffectTypeRevivePartyMember,
			models.SpellEffectTypeReviveAllDownedParty:
			if amount <= 0 {
				return nil, fmt.Errorf("effects[%d].amount must be greater than 0", index)
			}
			if effectType == models.SpellEffectTypeDealDamage || effectType == models.SpellEffectTypeDealDamageAllEnemies {
				if hits <= 0 {
					hits = 1
				}
				rawAffinity := ""
				if effectPayload.DamageAffinity != nil {
					rawAffinity = strings.TrimSpace(*effectPayload.DamageAffinity)
				}
				normalized := models.NormalizeDamageAffinity(rawAffinity)
				normalizedValue := string(normalized)
				damageAffinity = &normalizedValue
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			if len(statusesToApply) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToApply is required", index)
			}
		case models.SpellEffectTypeApplyDetrimentalStatus,
			models.SpellEffectTypeApplyDetrimentalAll:
			if len(statusesToApply) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToApply is required", index)
			}
			statusesToApply = normalizeSpellStatusesForEffectType(effectType, statusesToApply)
		case models.SpellEffectTypeRemoveDetrimental:
			if len(statusesToRemove) == 0 {
				return nil, fmt.Errorf("effects[%d].statusesToRemove is required", index)
			}
		case models.SpellEffectTypeUnlockLocks:
			if amount < 1 || amount > 100 {
				return nil, fmt.Errorf("effects[%d].amount must be between 1 and 100", index)
			}
		default:
			// Allow new effect types without backend changes.
		}

		effects = append(effects, models.SpellEffect{
			Type:             effectType,
			Amount:           amount,
			Hits:             hits,
			DamageAffinity:   damageAffinity,
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

func (s *server) parseSpellUpsertRequest(body spellUpsertRequest, defaultAbilityLevel int) (*models.Spell, error) {
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
	abilityLevel := defaultAbilityLevel
	if abilityLevel < 1 {
		abilityLevel = 1
	}
	if body.AbilityLevel != nil {
		abilityLevel = *body.AbilityLevel
	}
	if abilityLevel < 1 {
		return nil, fmt.Errorf("abilityLevel must be 1 or greater")
	}
	if body.CooldownTurns < 0 {
		return nil, fmt.Errorf("cooldownTurns must be zero or greater")
	}
	manaCost := body.ManaCost
	cooldownTurns := 0
	if abilityType == models.SpellAbilityTypeTechnique {
		manaCost = 0
		cooldownTurns = body.CooldownTurns
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
		AbilityLevel:          abilityLevel,
		CooldownTurns:         cooldownTurns,
		EffectText:            strings.TrimSpace(body.EffectText),
		SchoolOfMagic:         schoolOfMagic,
		ManaCost:              manaCost,
		Effects:               effects,
	}, nil
}

func normalizeSpellProgressionBand(levelBand int) int {
	if levelBand <= spellProgressionLevelBands[0] {
		return spellProgressionLevelBands[0]
	}
	if levelBand >= spellProgressionLevelBands[len(spellProgressionLevelBands)-1] {
		return spellProgressionLevelBands[len(spellProgressionLevelBands)-1]
	}

	best := spellProgressionLevelBands[0]
	bestDistance := absInt(levelBand - best)
	for _, candidate := range spellProgressionLevelBands[1:] {
		distance := absInt(levelBand - candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func spellProgressionTargetLevelForBand(levelBand int) int {
	return normalizeSpellProgressionBand(levelBand)
}

func inferSpellProgressionBand(spell *models.Spell) int {
	if spell == nil {
		return 25
	}
	if spell.AbilityLevel > 0 {
		return normalizeSpellProgressionBand(spell.AbilityLevel)
	}
	powerScore := float64(spellMaxInt(spell.ManaCost, 0))
	for _, effect := range spell.Effects {
		effectMagnitude := spellMaxInt(effect.Amount, 0)
		if effect.Type == models.SpellEffectTypeDealDamage || effect.Type == models.SpellEffectTypeDealDamageAllEnemies {
			effectMagnitude *= spellMaxInt(effect.Hits, 1)
		}
		powerScore += float64(effectMagnitude)
		powerScore += float64(len(effect.StatusesToApply) * 8)
		if len(effect.StatusesToRemove) > 0 {
			powerScore += 6
		}
	}

	type bandProfile struct {
		band  int
		score float64
	}
	profiles := []bandProfile{
		{band: 10, score: 24},
		{band: 25, score: 52},
		{band: 50, score: 88},
		{band: 70, score: 120},
	}

	bestBand := profiles[0].band
	bestDistance := math.Abs(powerScore - profiles[0].score)
	for _, profile := range profiles[1:] {
		distance := math.Abs(powerScore - profile.score)
		if distance < bestDistance {
			bestBand = profile.band
			bestDistance = distance
		}
	}
	return bestBand
}

func selectSeedBandForProgression(inferredBand int, occupied map[int]uuid.UUID) int {
	normalized := normalizeSpellProgressionBand(inferredBand)
	if _, taken := occupied[normalized]; !taken {
		return normalized
	}

	available := make([]int, 0, len(spellProgressionLevelBands))
	for _, band := range spellProgressionLevelBands {
		if _, taken := occupied[band]; !taken {
			available = append(available, band)
		}
	}
	if len(available) == 0 {
		return normalized
	}

	sort.Slice(available, func(i, j int) bool {
		left := absInt(available[i] - normalized)
		right := absInt(available[j] - normalized)
		if left == right {
			return available[i] < available[j]
		}
		return left < right
	})
	return available[0]
}

func spellProgressionPrimaryEffectType(spell *models.Spell) models.SpellEffectType {
	if spell == nil || len(spell.Effects) == 0 {
		return models.SpellEffectTypeDealDamage
	}
	return spell.Effects[0].Type
}

func spellProgressionTheme(spell *models.Spell) string {
	if spell != nil && len(spell.Effects) > 0 {
		if affinity := firstSpellDamageAffinity(spell); affinity != "" {
			switch models.NormalizeDamageAffinity(affinity) {
			case models.DamageAffinityFire:
				return "Fire"
			case models.DamageAffinityIce:
				return "Frost"
			case models.DamageAffinityLightning:
				return "Storm"
			case models.DamageAffinityPoison:
				return "Venom"
			case models.DamageAffinityArcane:
				return "Arcane"
			case models.DamageAffinityHoly:
				return "Radiant"
			case models.DamageAffinityShadow:
				return "Umbral"
			default:
				return "Force"
			}
		}
	}
	if spell != nil {
		if word := firstWord(spell.SchoolOfMagic); word != "" {
			return titleWord(word)
		}
		if word := firstWord(spell.Name); word != "" {
			return titleWord(word)
		}
	}
	return "Arcane"
}

func firstSpellDamageAffinity(spell *models.Spell) string {
	if spell == nil {
		return ""
	}
	for _, effect := range spell.Effects {
		if effect.DamageAffinity == nil {
			continue
		}
		value := strings.TrimSpace(*effect.DamageAffinity)
		if value == "" {
			continue
		}
		return value
	}
	return ""
}

func spellProgressionBandTerm(
	effectType models.SpellEffectType,
	levelBand int,
	abilityType models.SpellAbilityType,
) string {
	if abilityType == models.SpellAbilityTypeTechnique {
		switch effectType {
		case models.SpellEffectTypeRestoreLifePartyMember:
			switch levelBand {
			case 10:
				return "Recovery Form"
			case 25:
				return "Renewal Form"
			case 50:
				return "Battle Recovery"
			default:
				return "Grand Recovery"
			}
		case models.SpellEffectTypeRestoreLifeAllParty:
			switch levelBand {
			case 10:
				return "Rally Stance"
			case 25:
				return "Rally Formation"
			case 50:
				return "War Anthem"
			default:
				return "Legend Rally"
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			switch levelBand {
			case 10:
				return "Guard Stance"
			case 25:
				return "Warden Form"
			case 50:
				return "Fortress Form"
			default:
				return "Unbroken Form"
			}
		case models.SpellEffectTypeRemoveDetrimental:
			switch levelBand {
			case 10:
				return "Shake Off"
			case 25:
				return "Clear Focus"
			case 50:
				return "Iron Focus"
			default:
				return "Perfect Focus"
			}
		case models.SpellEffectTypeDealDamageAllEnemies:
			switch levelBand {
			case 10:
				return "Sweeping Form"
			case 25:
				return "Cyclone Form"
			case 50:
				return "War Tempest"
			default:
				return "Master Tempest"
			}
		default:
			switch levelBand {
			case 10:
				return "Strike"
			case 25:
				return "Assault"
			case 50:
				return "Onslaught"
			default:
				return "Mastery"
			}
		}
	}
	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		switch levelBand {
		case 10:
			return "Mend"
		case 25:
			return "Renewal"
		case 50:
			return "Revitalize"
		default:
			return "Transcendence"
		}
	case models.SpellEffectTypeRestoreLifeAllParty:
		switch levelBand {
		case 10:
			return "Chorus"
		case 25:
			return "Hymn"
		case 50:
			return "Anthem"
		default:
			return "Apotheosis"
		}
	case models.SpellEffectTypeApplyBeneficialStatus:
		switch levelBand {
		case 10:
			return "Ward"
		case 25:
			return "Aegis"
		case 50:
			return "Bastion"
		default:
			return "Citadel"
		}
	case models.SpellEffectTypeApplyDetrimentalStatus:
		switch levelBand {
		case 10:
			return "Hex"
		case 25:
			return "Curse"
		case 50:
			return "Scourge"
		default:
			return "Malediction"
		}
	case models.SpellEffectTypeApplyDetrimentalAll:
		switch levelBand {
		case 10:
			return "Haze"
		case 25:
			return "Miasma"
		case 50:
			return "Blight"
		default:
			return "Plague"
		}
	case models.SpellEffectTypeRemoveDetrimental:
		switch levelBand {
		case 10:
			return "Cleanse"
		case 25:
			return "Purge"
		case 50:
			return "Sanctify"
		default:
			return "Absolution"
		}
	case models.SpellEffectTypeDealDamageAllEnemies:
		switch levelBand {
		case 10:
			return "Pulse"
		case 25:
			return "Wave"
		case 50:
			return "Tempest"
		default:
			return "Cataclysm"
		}
	default:
		switch levelBand {
		case 10:
			return "Wisp"
		case 25:
			return "Bolt"
		case 50:
			return "Sphere"
		default:
			return "Nova"
		}
	}
}

func scaleSpellProgressionValue(base int, seedBand int, targetBand int, exponent float64) int {
	if base == 0 {
		return 0
	}
	if seedBand <= 0 {
		seedBand = 25
	}
	ratio := float64(targetBand) / float64(seedBand)
	scaled := int(math.Round(float64(base) * math.Pow(ratio, exponent)))
	if base > 0 && scaled < 1 {
		return 1
	}
	if base < 0 && scaled > -1 {
		return -1
	}
	return scaled
}

func estimateSpellProgressionMonsterHealth(levelBand int) int {
	level := normalizeSpellProgressionBand(levelBand)
	baseConstitution := 12
	effectiveConstitution := spellMaxInt(1, baseConstitution+level-1)
	return effectiveConstitution * 10
}

func spellProgressionBandRatio(levelBand int, ratios map[int]float64) float64 {
	normalized := normalizeSpellProgressionBand(levelBand)
	if ratio, ok := ratios[normalized]; ok {
		return ratio
	}
	return ratios[25]
}

func spellProgressionTargetAmount(
	effectType models.SpellEffectType,
	levelBand int,
	abilityType models.SpellAbilityType,
) int {
	normalizedBand := normalizeSpellProgressionBand(levelBand)
	if effectType == models.SpellEffectTypeDealDamage {
		damagePerLevel := 10
		if abilityType == models.SpellAbilityTypeTechnique {
			damagePerLevel = 8
		}
		return spellMaxInt(1, normalizedBand*damagePerLevel)
	}
	if effectType == models.SpellEffectTypeDealDamageAllEnemies {
		damagePerLevel := 6
		if abilityType == models.SpellAbilityTypeTechnique {
			damagePerLevel = 5
		}
		return spellMaxInt(1, normalizedBand*damagePerLevel)
	}

	health := estimateSpellProgressionMonsterHealth(levelBand)
	if health <= 0 {
		return 0
	}

	var ratio float64
	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		ratio = spellProgressionBandRatio(levelBand, map[int]float64{
			10: 0.12,
			25: 0.18,
			50: 0.28,
			70: 0.40,
		})
	case models.SpellEffectTypeRestoreLifeAllParty:
		ratio = spellProgressionBandRatio(levelBand, map[int]float64{
			10: 0.07,
			25: 0.11,
			50: 0.17,
			70: 0.24,
		})
	default:
		return 0
	}
	if abilityType == models.SpellAbilityTypeTechnique {
		ratio *= 0.78
	}
	target := int(math.Round(float64(health) * ratio))
	return spellMaxInt(1, target)
}

func spellProgressionTargetDamagePerTick(
	levelBand int,
	abilityType models.SpellAbilityType,
) int {
	directDamageTarget := spellProgressionTargetAmount(
		models.SpellEffectTypeDealDamage,
		levelBand,
		abilityType,
	)
	return spellMaxInt(1, int(math.Round(float64(directDamageTarget)*0.2)))
}

func scaleSpellProgressionCombatAmount(
	base int,
	effectType models.SpellEffectType,
	seedBand int,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if base == 0 {
		return 0
	}
	legacy := scaleSpellProgressionValue(base, seedBand, targetBand, 1.15)
	target := spellProgressionTargetAmount(effectType, targetBand, abilityType)
	if target <= 0 {
		return legacy
	}
	if targetBand > seedBand {
		return spellMaxInt(legacy, target)
	}
	if targetBand < seedBand {
		if legacy < target {
			return legacy
		}
		return target
	}
	return legacy
}

func scaleSpellProgressionDamagePerTick(
	base int,
	seedBand int,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if base == 0 {
		return 0
	}
	legacy := scaleSpellProgressionValue(base, seedBand, targetBand, 1.05)
	target := spellProgressionTargetDamagePerTick(targetBand, abilityType)
	if target <= 0 {
		return legacy
	}
	if targetBand > seedBand {
		return spellMaxInt(legacy, target)
	}
	if targetBand < seedBand {
		if legacy < target {
			return legacy
		}
		return target
	}
	return legacy
}

func spellProgressionBandFloor(levelBand int, floors map[int]int) int {
	normalized := normalizeSpellProgressionBand(levelBand)
	if floor, ok := floors[normalized]; ok {
		return floor
	}
	return floors[25]
}

func estimateSpellProgressionPlayerMaxMana(levelBand int) int {
	level := normalizeSpellProgressionBand(levelBand)
	if level < 1 {
		level = 1
	}
	levelsGained := level - 1
	pointsGained := levelsGained * models.CharacterStatPointsPerLevel
	baseMental := models.CharacterStatBaseValue * 2

	// Assume caster builds allocate a growing share of stat points into INT/WIS.
	casterShare := 0.42
	if level > 10 {
		progress := float64(level-10) / 60.0
		if progress > 1 {
			progress = 1
		}
		casterShare += progress * 0.13
	}
	estimatedMental := float64(baseMental) + (float64(pointsGained) * casterShare)
	estimatedMana := int(math.Round(estimatedMental * 5.0))
	return spellMaxInt(20, estimatedMana)
}

func spellProgressionTargetManaCost(
	effectType models.SpellEffectType,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if abilityType == models.SpellAbilityTypeTechnique {
		return 0
	}
	playerMana := estimateSpellProgressionPlayerMaxMana(targetBand)

	switch effectType {
	case models.SpellEffectTypeDealDamage:
		bandFloor := spellProgressionBandFloor(targetBand, map[int]int{
			10: 32,
			25: 72,
			50: 180,
			70: 360,
		})
		ratio := spellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.18,
			25: 0.24,
			50: 0.34,
			70: 0.44,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return spellMaxInt(bandFloor, target)
	case models.SpellEffectTypeDealDamageAllEnemies:
		bandFloor := spellProgressionBandFloor(targetBand, map[int]int{
			10: 40,
			25: 96,
			50: 240,
			70: 480,
		})
		ratio := spellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.22,
			25: 0.32,
			50: 0.46,
			70: 0.60,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return spellMaxInt(bandFloor, target)
	case models.SpellEffectTypeRestoreLifePartyMember:
		bandFloor := spellProgressionBandFloor(targetBand, map[int]int{
			10: 28,
			25: 60,
			50: 152,
			70: 310,
		})
		ratio := spellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.16,
			25: 0.22,
			50: 0.30,
			70: 0.38,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return spellMaxInt(bandFloor, target)
	case models.SpellEffectTypeRestoreLifeAllParty:
		bandFloor := spellProgressionBandFloor(targetBand, map[int]int{
			10: 36,
			25: 84,
			50: 210,
			70: 420,
		})
		ratio := spellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.20,
			25: 0.28,
			50: 0.40,
			70: 0.54,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return spellMaxInt(bandFloor, target)
	case models.SpellEffectTypeApplyBeneficialStatus,
		models.SpellEffectTypeApplyDetrimentalStatus,
		models.SpellEffectTypeApplyDetrimentalAll,
		models.SpellEffectTypeRemoveDetrimental:
		return spellProgressionBandFloor(targetBand, map[int]int{
			10: 24,
			25: 52,
			50: 128,
			70: 260,
		})
	default:
		return spellProgressionBandFloor(targetBand, map[int]int{
			10: 20,
			25: 44,
			50: 112,
			70: 224,
		})
	}
}

func scaleSpellProgressionManaCost(
	baseMana int,
	effectType models.SpellEffectType,
	seedBand int,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if abilityType == models.SpellAbilityTypeTechnique {
		return 0
	}
	if baseMana <= 0 {
		return 0
	}
	legacy := scaleSpellProgressionValue(baseMana, seedBand, targetBand, 1.25)
	target := spellProgressionTargetManaCost(effectType, targetBand, abilityType)
	if target < 1 {
		target = 1
	}

	scaled := legacy
	if targetBand > seedBand {
		scaled = spellMaxInt(legacy, target)
	} else if targetBand < seedBand {
		scaled = spellMinInt(legacy, target)
	}
	if scaled < 1 {
		return 1
	}
	if scaled > 600 {
		return 600
	}
	return scaled
}

func cloneSpellEffectData(effectData map[string]interface{}) map[string]interface{} {
	if effectData == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(effectData))
	for key, value := range effectData {
		cloned[key] = value
	}
	return cloned
}

func buildScaledSpellProgressionEffects(
	seedEffects models.SpellEffects,
	seedBand int,
	targetBand int,
	abilityType models.SpellAbilityType,
) models.SpellEffects {
	if len(seedEffects) == 0 {
		return models.SpellEffects{}
	}

	scaled := make(models.SpellEffects, 0, len(seedEffects))
	for _, effect := range seedEffects {
		next := models.SpellEffect{
			Type:             effect.Type,
			Amount:           scaleSpellProgressionCombatAmount(effect.Amount, effect.Type, seedBand, targetBand, abilityType),
			Hits:             effect.Hits,
			StatusesToRemove: append(models.StringArray(nil), effect.StatusesToRemove...),
			EffectData:       cloneSpellEffectData(effect.EffectData),
		}
		if effect.DamageAffinity != nil {
			affinity := strings.TrimSpace(*effect.DamageAffinity)
			next.DamageAffinity = &affinity
		}
		if len(effect.StatusesToApply) > 0 {
			statuses := make(models.ScenarioFailureStatusTemplates, 0, len(effect.StatusesToApply))
			for _, status := range effect.StatusesToApply {
				scaledStatus := status
				scaledStatus.DurationSeconds = spellMaxInt(1, scaleSpellProgressionValue(status.DurationSeconds, seedBand, targetBand, 0.35))
				scaledStatus.DamagePerTick = scaleSpellProgressionDamagePerTick(status.DamagePerTick, seedBand, targetBand, abilityType)
				scaledStatus.HealthPerTick = scaleSpellProgressionDamagePerTick(status.HealthPerTick, seedBand, targetBand, abilityType)
				scaledStatus.ManaPerTick = scaleSpellProgressionDamagePerTick(status.ManaPerTick, seedBand, targetBand, abilityType)
				scaledStatus.StrengthMod = scaleSpellProgressionValue(status.StrengthMod, seedBand, targetBand, 0.4)
				scaledStatus.DexterityMod = scaleSpellProgressionValue(status.DexterityMod, seedBand, targetBand, 0.4)
				scaledStatus.ConstitutionMod = scaleSpellProgressionValue(status.ConstitutionMod, seedBand, targetBand, 0.4)
				scaledStatus.IntelligenceMod = scaleSpellProgressionValue(status.IntelligenceMod, seedBand, targetBand, 0.4)
				scaledStatus.WisdomMod = scaleSpellProgressionValue(status.WisdomMod, seedBand, targetBand, 0.4)
				scaledStatus.CharismaMod = scaleSpellProgressionValue(status.CharismaMod, seedBand, targetBand, 0.4)
				scaledStatus.PhysicalDamageBonusPercent = scaleSpellProgressionValue(status.PhysicalDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.PiercingDamageBonusPercent = scaleSpellProgressionValue(status.PiercingDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.SlashingDamageBonusPercent = scaleSpellProgressionValue(status.SlashingDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.BludgeoningDamageBonusPercent = scaleSpellProgressionValue(status.BludgeoningDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.FireDamageBonusPercent = scaleSpellProgressionValue(status.FireDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.IceDamageBonusPercent = scaleSpellProgressionValue(status.IceDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.LightningDamageBonusPercent = scaleSpellProgressionValue(status.LightningDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.PoisonDamageBonusPercent = scaleSpellProgressionValue(status.PoisonDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.ArcaneDamageBonusPercent = scaleSpellProgressionValue(status.ArcaneDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.HolyDamageBonusPercent = scaleSpellProgressionValue(status.HolyDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.ShadowDamageBonusPercent = scaleSpellProgressionValue(status.ShadowDamageBonusPercent, seedBand, targetBand, 0.35)
				scaledStatus.PhysicalResistancePercent = scaleSpellProgressionValue(status.PhysicalResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.PiercingResistancePercent = scaleSpellProgressionValue(status.PiercingResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.SlashingResistancePercent = scaleSpellProgressionValue(status.SlashingResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.BludgeoningResistancePercent = scaleSpellProgressionValue(status.BludgeoningResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.FireResistancePercent = scaleSpellProgressionValue(status.FireResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.IceResistancePercent = scaleSpellProgressionValue(status.IceResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.LightningResistancePercent = scaleSpellProgressionValue(status.LightningResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.PoisonResistancePercent = scaleSpellProgressionValue(status.PoisonResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.ArcaneResistancePercent = scaleSpellProgressionValue(status.ArcaneResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.HolyResistancePercent = scaleSpellProgressionValue(status.HolyResistancePercent, seedBand, targetBand, 0.35)
				scaledStatus.ShadowResistancePercent = scaleSpellProgressionValue(status.ShadowResistancePercent, seedBand, targetBand, 0.35)
				statuses = append(statuses, scaledStatus)
			}
			next.StatusesToApply = statuses
		}
		scaled = append(scaled, next)
	}
	return scaled
}

func buildSpellProgressionEffectText(effects models.SpellEffects) string {
	if len(effects) == 0 {
		return "A refined magical technique."
	}

	effect := effects[0]
	switch effect.Type {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return fmt.Sprintf("Restore %d health to one ally.", spellMaxInt(effect.Amount, 1))
	case models.SpellEffectTypeRestoreLifeAllParty:
		return fmt.Sprintf("Restore %d health to all allies.", spellMaxInt(effect.Amount, 1))
	case models.SpellEffectTypeApplyBeneficialStatus:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to allies.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies beneficial statuses to allies."
	case models.SpellEffectTypeApplyDetrimentalStatus:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to one enemy.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies detrimental statuses to one enemy."
	case models.SpellEffectTypeApplyDetrimentalAll:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to all enemies.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies detrimental statuses to all enemies."
	case models.SpellEffectTypeRemoveDetrimental:
		return "Removes detrimental statuses from allies."
	case models.SpellEffectTypeDealDamageAllEnemies:
		affinity := "magical"
		if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
			affinity = strings.TrimSpace(*effect.DamageAffinity)
		}
		if spellMaxInt(effect.Hits, 1) > 1 {
			return fmt.Sprintf("Deals %d %s damage to all enemies %d times.", spellMaxInt(effect.Amount, 1), affinity, spellMaxInt(effect.Hits, 1))
		}
		return fmt.Sprintf("Deals %d %s damage to all enemies.", spellMaxInt(effect.Amount, 1), affinity)
	default:
		affinity := "magical"
		if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
			affinity = strings.TrimSpace(*effect.DamageAffinity)
		}
		if spellMaxInt(effect.Hits, 1) > 1 {
			return fmt.Sprintf("Deals %d %s damage to a target %d times.", spellMaxInt(effect.Amount, 1), affinity, spellMaxInt(effect.Hits, 1))
		}
		return fmt.Sprintf("Deals %d %s damage to a target.", spellMaxInt(effect.Amount, 1), affinity)
	}
}

func stripSpellProgressionMetaSentences(description string) string {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return ""
	}

	sentences := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	kept := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		candidate := strings.TrimSpace(sentence)
		if candidate == "" {
			continue
		}
		lower := strings.ToLower(candidate)
		if strings.Contains(lower, "evolution") ||
			strings.Contains(lower, "progression") ||
			strings.Contains(lower, "level ") ||
			strings.Contains(lower, "level-") ||
			strings.Contains(lower, "level band") ||
			strings.Contains(lower, "tier ") {
			continue
		}
		kept = append(kept, candidate)
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, ". ") + "."
}

func buildSpellProgressionFlavorDescription(
	seed *models.Spell,
	primaryEffect models.SpellEffectType,
) string {
	if seed != nil {
		cleaned := stripSpellProgressionMetaSentences(seed.Description)
		if cleaned != "" {
			return cleaned
		}
	}

	school := "arcane"
	if seed != nil && strings.TrimSpace(seed.SchoolOfMagic) != "" {
		school = strings.ToLower(strings.TrimSpace(seed.SchoolOfMagic))
	}

	switch primaryEffect {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return "A soothing surge of mystic energy mends wounds and steadies one ally."
	case models.SpellEffectTypeRestoreLifeAllParty:
		return "A radiant wave of restorative power washes across the party, renewing battered allies."
	case models.SpellEffectTypeApplyBeneficialStatus:
		return "A focused invocation fortifies allies, sharpening their edge for the clash ahead."
	case models.SpellEffectTypeApplyDetrimentalStatus:
		return "A cruel working binds a single foe with a lingering detrimental condition."
	case models.SpellEffectTypeApplyDetrimentalAll:
		return "A spreading curse rolls across the battlefield, burdening every enemy it touches."
	case models.SpellEffectTypeRemoveDetrimental:
		return "A cleansing pulse strips away harmful effects and restores clarity in the heat of battle."
	case models.SpellEffectTypeDealDamageAllEnemies:
		return fmt.Sprintf("A sweeping burst of %s force erupts outward, overwhelming every foe in reach.", school)
	default:
		return fmt.Sprintf("A concentrated strike of %s power crashes into a single foe with ruthless force.", school)
	}
}

func spellProgressionIntensityWord(targetLevel int) string {
	switch {
	case targetLevel >= 70:
		return "cataclysmic"
	case targetLevel >= 50:
		return "surging"
	case targetLevel >= 25:
		return "focused"
	default:
		return "quick"
	}
}

func buildSpellProgressionVariantFlavorDescription(
	seed *models.Spell,
	primaryEffect models.SpellEffectType,
	targetLevel int,
	abilityType models.SpellAbilityType,
) string {
	intensity := spellProgressionIntensityWord(targetLevel)
	if abilityType == models.SpellAbilityTypeTechnique {
		switch primaryEffect {
		case models.SpellEffectTypeRestoreLifePartyMember:
			return fmt.Sprintf("A %s recovery form steadies one ally and restores their tempo.", intensity)
		case models.SpellEffectTypeRestoreLifeAllParty:
			return fmt.Sprintf("A %s rallying cadence renews the whole party in one motion.", intensity)
		case models.SpellEffectTypeApplyBeneficialStatus:
			return fmt.Sprintf("A %s combat stance hardens allied resolve and sharpens their edge.", intensity)
		case models.SpellEffectTypeApplyDetrimentalStatus:
			return fmt.Sprintf("A %s martial pressure point disrupts one foe's rhythm and balance.", intensity)
		case models.SpellEffectTypeApplyDetrimentalAll:
			return fmt.Sprintf("A %s sweeping technique throws every nearby foe off balance at once.", intensity)
		case models.SpellEffectTypeRemoveDetrimental:
			return fmt.Sprintf("A %s reset strips away pressure and restores the team's control.", intensity)
		case models.SpellEffectTypeDealDamageAllEnemies:
			return fmt.Sprintf("A %s sweeping technique tears through nearby enemies in one brutal pass.", intensity)
		default:
			return fmt.Sprintf("A %s combat technique drives concentrated force into a single enemy.", intensity)
		}
	}

	school := "arcane"
	if seed != nil && strings.TrimSpace(seed.SchoolOfMagic) != "" {
		school = strings.ToLower(strings.TrimSpace(seed.SchoolOfMagic))
	}

	switch primaryEffect {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return fmt.Sprintf("A %s restorative pulse swiftly mends one ally and steadies their footing.", intensity)
	case models.SpellEffectTypeRestoreLifeAllParty:
		return fmt.Sprintf("A %s healing surge washes over the party and renews battered allies.", intensity)
	case models.SpellEffectTypeApplyBeneficialStatus:
		return fmt.Sprintf("A %s invocation floods allies with sharpened focus and battle-ready resolve.", intensity)
	case models.SpellEffectTypeApplyDetrimentalStatus:
		return fmt.Sprintf("A %s hex settles onto one foe and steadily undermines their strength.", intensity)
	case models.SpellEffectTypeApplyDetrimentalAll:
		return fmt.Sprintf("A %s curse spills across the field and burdens every enemy it reaches.", intensity)
	case models.SpellEffectTypeRemoveDetrimental:
		return fmt.Sprintf("A %s cleansing pulse tears away harmful effects and restores control.", intensity)
	case models.SpellEffectTypeDealDamageAllEnemies:
		return fmt.Sprintf("A %s wave of %s force erupts outward and engulfs every nearby foe.", intensity, school)
	default:
		return fmt.Sprintf("A %s blast of %s power crashes into one foe with punishing force.", intensity, school)
	}
}

func buildSpellProgressionVariant(
	seed *models.Spell,
	seedBand int,
	targetBand int,
	usedNames map[string]struct{},
	flavorOverride *generatedSpellProgressionVariantFlavor,
	abilityType models.SpellAbilityType,
) *models.Spell {
	targetLevel := spellProgressionTargetLevelForBand(targetBand)
	primaryEffect := spellProgressionPrimaryEffectType(seed)
	fallbackName := fmt.Sprintf(
		"%s %s",
		spellProgressionTheme(seed),
		spellProgressionBandTerm(primaryEffect, targetLevel, abilityType),
	)
	nameBase := fallbackName
	if flavorOverride != nil && strings.TrimSpace(flavorOverride.Name) != "" {
		nameBase = flavorOverride.Name
	}
	name := nextUniqueAbilityName(nameBase, usedNames, abilityType)
	effects := buildScaledSpellProgressionEffects(seed.Effects, seedBand, targetLevel, abilityType)
	manaCost := scaleSpellProgressionManaCost(spellMaxInt(seed.ManaCost, 1), primaryEffect, seedBand, targetLevel, abilityType)
	cooldownTurns := 0
	if abilityType == models.SpellAbilityTypeTechnique {
		manaCost = 0
		if seed != nil && seed.CooldownTurns > 0 {
			cooldownTurns = seed.CooldownTurns
		}
	}
	description := buildSpellProgressionVariantFlavorDescription(seed, primaryEffect, targetLevel, abilityType)
	if flavorOverride != nil && strings.TrimSpace(flavorOverride.Description) != "" {
		description = flavorOverride.Description
	}
	emptyError := ""
	now := time.Now()
	return &models.Spell{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		Name:                  name,
		Description:           description,
		IconURL:               "",
		ImageGenerationStatus: models.SpellImageGenerationStatusNone,
		ImageGenerationError:  &emptyError,
		AbilityType:           abilityType,
		AbilityLevel:          targetLevel,
		CooldownTurns:         cooldownTurns,
		EffectText:            buildSpellProgressionEffectText(effects),
		SchoolOfMagic:         strings.TrimSpace(seed.SchoolOfMagic),
		ManaCost:              manaCost,
		Effects:               effects,
	}
}

func firstWord(value string) string {
	for _, part := range strings.Fields(strings.TrimSpace(value)) {
		if part != "" {
			return part
		}
	}
	return ""
}

func titleWord(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lowered := []rune(strings.ToLower(trimmed))
	lowered[0] = unicode.ToUpper(lowered[0])
	return string(lowered)
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func spellMaxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func spellMinInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
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

	spell, err := s.parseSpellUpsertRequest(requestBody, 1)
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
	defaultAbilityLevel := 1
	if existingSpell != nil && existingSpell.AbilityLevel > 0 {
		defaultAbilityLevel = existingSpell.AbilityLevel
	}
	spell, err := s.parseSpellUpsertRequest(requestBody, defaultAbilityLevel)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Spell().Update(ctx, spellID, map[string]interface{}{
		"name":           spell.Name,
		"description":    spell.Description,
		"icon_url":       spell.IconURL,
		"ability_type":   spell.AbilityType,
		"ability_level":  spell.AbilityLevel,
		"cooldown_turns": spell.CooldownTurns,
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

type abilityTomePricingTier struct {
	levelMax int
	rarity   string
	buyPrice int
}

var abilityTomePricingTiers = []abilityTomePricingTier{
	{levelMax: 15, rarity: "Common", buyPrice: 150},
	{levelMax: 30, rarity: "Uncommon", buyPrice: 350},
	{levelMax: 50, rarity: "Epic", buyPrice: 900},
	{levelMax: 100, rarity: "Mythic", buyPrice: 1800},
}

var abilityTomeStopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "as": {}, "at": {}, "by": {}, "for": {}, "from": {},
	"in": {}, "into": {}, "of": {}, "on": {}, "or": {}, "that": {}, "the": {}, "their": {},
	"through": {}, "to": {}, "up": {}, "with": {}, "your": {}, "one": {}, "all": {},
}

func abilityTomePricingTierForLevel(level int) abilityTomePricingTier {
	if level < 1 {
		level = 1
	}
	for _, tier := range abilityTomePricingTiers {
		if level <= tier.levelMax {
			return tier
		}
	}
	return abilityTomePricingTiers[len(abilityTomePricingTiers)-1]
}

func tokenizeAbilityTomeText(value string) []string {
	rawWords := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(value)), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	words := make([]string, 0, len(rawWords))
	for _, word := range rawWords {
		trimmed := strings.TrimSpace(word)
		if trimmed == "" {
			continue
		}
		words = append(words, trimmed)
	}
	return words
}

func abilityTomeKeywords(value string, maxWords int) []string {
	if maxWords <= 0 {
		return []string{}
	}
	seen := map[string]struct{}{}
	keywords := make([]string, 0, maxWords)
	for _, word := range tokenizeAbilityTomeText(value) {
		if _, skip := abilityTomeStopwords[word]; skip {
			continue
		}
		if _, exists := seen[word]; exists {
			continue
		}
		seen[word] = struct{}{}
		keywords = append(keywords, word)
		if len(keywords) >= maxWords {
			break
		}
	}
	if len(keywords) > 0 {
		return keywords
	}
	words := tokenizeAbilityTomeText(value)
	if len(words) > maxWords {
		return words[:maxWords]
	}
	return words
}

func trimAbilityTomeSentenceFragment(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.Trim(trimmed, " \t\r\n")
	trimmed = strings.TrimRight(trimmed, ".!?;,:")
	return strings.TrimSpace(trimmed)
}

func buildAbilityTomeName(ability *models.Spell) string {
	name := "Unknown Ability"
	if ability != nil {
		if trimmed := trimAbilityTomeSentenceFragment(ability.Name); trimmed != "" {
			name = trimmed
		}
	}
	return fmt.Sprintf("Tome of %s", name)
}

func buildLegacyAbilityTomeName(ability *models.Spell) string {
	return buildAbilityTomeName(ability) + "."
}

func abilityTomeCoverAdjective(ability *models.Spell) string {
	if ability == nil {
		return "weathered"
	}
	name := strings.ToLower(strings.TrimSpace(ability.Name))
	school := strings.ToLower(strings.TrimSpace(ability.SchoolOfMagic))
	switch {
	case strings.Contains(name, "ember") || strings.Contains(name, "fire") || strings.Contains(school, "pyro"):
		return "soot-dark"
	case strings.Contains(name, "frost") || strings.Contains(name, "ice") || strings.Contains(school, "cryo"):
		return "rime-lined"
	case strings.Contains(name, "storm") || strings.Contains(name, "lightning") || strings.Contains(school, "tempest"):
		return "storm-creased"
	case strings.Contains(name, "shadow") || strings.Contains(name, "night") || strings.Contains(school, "shadow") || strings.Contains(school, "umbral"):
		return "shadow-inked"
	case strings.Contains(name, "radiant") || strings.Contains(name, "solar") || strings.Contains(school, "radiance") || strings.Contains(school, "holy"):
		return "gold-filigreed"
	case normalizeSpellAbilityType(string(ability.AbilityType)) == models.SpellAbilityTypeTechnique:
		return "weathered"
	default:
		return "vellum-bound"
	}
}

func abilityTomeCoverFinish(ability *models.Spell) string {
	if ability == nil {
		return "inked"
	}
	name := strings.ToLower(strings.TrimSpace(ability.Name))
	school := strings.ToLower(strings.TrimSpace(ability.SchoolOfMagic))
	switch {
	case strings.Contains(name, "ember") || strings.Contains(name, "fire") || strings.Contains(school, "pyro"):
		return "copper foil"
	case strings.Contains(name, "frost") || strings.Contains(name, "ice") || strings.Contains(school, "cryo"):
		return "silver leaf"
	case strings.Contains(name, "storm") || strings.Contains(name, "lightning") || strings.Contains(school, "tempest"):
		return "blue-gold foil"
	case strings.Contains(name, "shadow") || strings.Contains(name, "night") || strings.Contains(school, "shadow") || strings.Contains(school, "umbral"):
		return "smoked silver"
	case strings.Contains(name, "radiant") || strings.Contains(name, "solar") || strings.Contains(school, "radiance") || strings.Contains(school, "holy"):
		return "sun-bright leaf"
	case normalizeSpellAbilityType(string(ability.AbilityType)) == models.SpellAbilityTypeTechnique:
		return "iron-black ink"
	default:
		return "gilded script"
	}
}

func abilityTomeInteriorDetail(ability *models.Spell) string {
	if ability == nil {
		return "practical lessons"
	}
	source := trimAbilityTomeSentenceFragment(ability.Description)
	if source == "" {
		source = trimAbilityTomeSentenceFragment(ability.EffectText)
	}
	if source == "" {
		if normalizeSpellAbilityType(string(ability.AbilityType)) == models.SpellAbilityTypeTechnique {
			return "well-worn combat drills"
		}
		return "annotated magical exercises"
	}
	if normalizeSpellAbilityType(string(ability.AbilityType)) == models.SpellAbilityTypeTechnique {
		return fmt.Sprintf("step-by-step diagrams for %s", strings.ToLower(source))
	}
	return fmt.Sprintf("margin notes on %s", strings.ToLower(source))
}

func buildAbilityTomeDescription(ability *models.Spell) string {
	abilityType := normalizeSpellAbilityType(string(ability.AbilityType))
	bookType := "grimoire"
	binding := "aged leather"
	if abilityType == models.SpellAbilityTypeTechnique {
		bookType = "manual"
		binding = "corded canvas"
	}
	tomeName := "Unknown Ability"
	if ability != nil {
		if trimmed := trimAbilityTomeSentenceFragment(ability.Name); trimmed != "" {
			tomeName = trimmed
		}
	}

	return fmt.Sprintf(
		"A %s %s bound in %s, titled %q in %s, with %s.",
		abilityTomeCoverAdjective(ability),
		bookType,
		binding,
		tomeName,
		abilityTomeCoverFinish(ability),
		abilityTomeInteriorDetail(ability),
	)
}

func buildAbilityTomeItem(ability *models.Spell) *models.InventoryItem {
	level := 1
	if ability != nil && ability.AbilityLevel > 0 {
		level = ability.AbilityLevel
	}
	tier := abilityTomePricingTierForLevel(level)
	abilityType := normalizeSpellAbilityType(string(ability.AbilityType))
	internalTags := models.StringArray{
		"tome",
		"ability_tome",
		fmt.Sprintf("%s_tome", abilityType),
	}
	return &models.InventoryItem{
		Name:                    buildAbilityTomeName(ability),
		FlavorText:              buildAbilityTomeDescription(ability),
		EffectText:              fmt.Sprintf("Consume to learn %s.", strings.TrimSpace(ability.Name)),
		RarityTier:              tier.rarity,
		BuyPrice:                intPtr(tier.buyPrice),
		ItemLevel:               level,
		ConsumeStatusesToAdd:    models.ScenarioFailureStatusTemplates{},
		ConsumeStatusesToRemove: models.StringArray{},
		ConsumeSpellIDs:         models.StringArray{ability.ID.String()},
		ConsumeTeachRecipeIDs:   models.StringArray{},
		AlchemyRecipes:          models.InventoryRecipes{},
		WorkshopRecipes:         models.InventoryRecipes{},
		InternalTags:            internalTags,
		ImageGenerationStatus:   models.InventoryImageGenerationStatusQueued,
	}
}

func abilityTomeUpdateMap(item *models.InventoryItem) map[string]interface{} {
	return map[string]interface{}{
		"archived":                           false,
		"name":                               item.Name,
		"flavor_text":                        item.FlavorText,
		"effect_text":                        item.EffectText,
		"rarity_tier":                        item.RarityTier,
		"buy_price":                          item.BuyPrice,
		"unlock_tier":                        nil,
		"unlock_locks_strength":              nil,
		"item_level":                         item.ItemLevel,
		"equip_slot":                         nil,
		"strength_mod":                       0,
		"dexterity_mod":                      0,
		"constitution_mod":                   0,
		"intelligence_mod":                   0,
		"wisdom_mod":                         0,
		"charisma_mod":                       0,
		"hand_item_category":                 nil,
		"handedness":                         nil,
		"damage_min":                         nil,
		"damage_max":                         nil,
		"damage_affinity":                    nil,
		"swipes_per_attack":                  nil,
		"block_percentage":                   nil,
		"damage_blocked":                     nil,
		"spell_damage_bonus_percent":         nil,
		"consume_health_delta":               0,
		"consume_mana_delta":                 0,
		"consume_revive_party_member_health": 0,
		"consume_revive_all_downed_party_members_health": 0,
		"consume_deal_damage":                            0,
		"consume_deal_damage_hits":                       0,
		"consume_deal_damage_all_enemies":                0,
		"consume_deal_damage_all_enemies_hits":           0,
		"consume_create_base":                            false,
		"consume_statuses_to_add":                        item.ConsumeStatusesToAdd,
		"consume_statuses_to_remove":                     item.ConsumeStatusesToRemove,
		"consume_spell_ids":                              item.ConsumeSpellIDs,
		"consume_teach_recipe_ids":                       item.ConsumeTeachRecipeIDs,
		"alchemy_recipes":                                item.AlchemyRecipes,
		"workshop_recipes":                               item.WorkshopRecipes,
		"internal_tags":                                  item.InternalTags,
		"image_generation_status":                        models.InventoryImageGenerationStatusQueued,
		"image_generation_error":                         nil,
	}
}

func parseAbilityIDList(input []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(input))
	seen := map[uuid.UUID]struct{}{}
	for idx, rawID := range input {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}
		parsed, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, fmt.Errorf("abilityIds[%d] must be a valid UUID", idx)
		}
		if _, exists := seen[parsed]; exists {
			continue
		}
		seen[parsed] = struct{}{}
		ids = append(ids, parsed)
	}
	return ids, nil
}

func (s *server) generateAbilityTomes(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		AbilityIDs []string `json:"abilityIds"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	abilityIDs, err := parseAbilityIDList(requestBody.AbilityIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(abilityIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one ability ID is required"})
		return
	}

	abilities, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	abilityByID := make(map[uuid.UUID]models.Spell, len(abilities))
	for _, ability := range abilities {
		abilityByID[ability.ID] = ability
	}

	missingAbilityIDs := make([]string, 0)
	for _, abilityID := range abilityIDs {
		if _, exists := abilityByID[abilityID]; !exists {
			missingAbilityIDs = append(missingAbilityIDs, abilityID.String())
		}
	}
	if len(missingAbilityIDs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":             "some abilities were not found",
			"missingAbilityIds": missingAbilityIDs,
		})
		return
	}

	existingItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	existingByName := make(map[string]models.InventoryItem, len(existingItems))
	for _, item := range existingItems {
		key := strings.ToLower(strings.TrimSpace(item.Name))
		if key == "" {
			continue
		}
		existingByName[key] = item
	}

	processedItems := make([]gin.H, 0, len(abilityIDs))
	warnings := make([]string, 0)
	createdCount := 0
	updatedCount := 0
	queuedImageCount := 0

	for _, abilityID := range abilityIDs {
		ability := abilityByID[abilityID]
		tome := buildAbilityTomeItem(&ability)
		key := strings.ToLower(strings.TrimSpace(tome.Name))
		legacyKey := strings.ToLower(strings.TrimSpace(buildLegacyAbilityTomeName(&ability)))
		action := "created"
		inventoryItemID := tome.ID

		existing, exists := existingByName[key]
		if !exists && legacyKey != key {
			existing, exists = existingByName[legacyKey]
		}
		if exists {
			if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, existing.ID, abilityTomeUpdateMap(tome)); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			inventoryItemID = existing.ID
			action = "updated"
			updatedCount++
		} else {
			if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, tome); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			inventoryItemID = tome.ID
			createdCount++
		}

		imageQueued := true
		var warning *string
		if err := s.enqueueInventoryItemImageGeneration(
			ctx,
			inventoryItemID,
			tome.Name,
			tome.FlavorText,
			tome.RarityTier,
		); err != nil {
			imageQueued = false
			errText := fmt.Sprintf("%s: %s", tome.Name, err.Error())
			warnings = append(warnings, errText)
			warning = &errText
		} else {
			queuedImageCount++
		}

		processedItems = append(processedItems, gin.H{
			"abilityId":       ability.ID,
			"abilityName":     ability.Name,
			"abilityType":     ability.AbilityType,
			"inventoryItemId": inventoryItemID,
			"tomeName":        tome.Name,
			"action":          action,
			"imageQueued":     imageQueued,
			"warning":         warning,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":            processedItems,
		"createdCount":     createdCount,
		"updatedCount":     updatedCount,
		"processedCount":   len(processedItems),
		"queuedImageCount": queuedImageCount,
		"warnings":         warnings,
	})
}

func (s *server) generateSpellProgression(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	spellID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid spell ID"})
		return
	}

	seedSpell, err := s.dbClient.Spell().FindByID(ctx, spellID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "spell not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	abilityType := normalizeSpellAbilityType(string(seedSpell.AbilityType))
	if abilityType != models.SpellAbilityTypeSpell && abilityType != models.SpellAbilityTypeTechnique {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "progressions are only supported for spells and techniques"})
		return
	}

	progression, err := s.dbClient.Spell().FindProgressionBySpellID(ctx, seedSpell.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if progression == nil {
		progression = &models.SpellProgression{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        fmt.Sprintf("%s Progression", strings.TrimSpace(seedSpell.Name)),
			AbilityType: abilityType,
		}
		if strings.TrimSpace(progression.Name) == "Progression" {
			progression.Name = fmt.Sprintf("%s Progression", spellProgressionTheme(seedSpell))
		}
		if err := s.dbClient.Spell().CreateProgression(ctx, progression); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	members, err := s.dbClient.Spell().FindProgressionMembers(ctx, progression.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	occupiedBands := map[int]uuid.UUID{}
	seedBand := 0
	seedBandNeedsNormalization := false
	for _, member := range members {
		occupiedBands[member.LevelBand] = member.SpellID
		targetLevel := spellProgressionTargetLevelForBand(member.LevelBand)
		if member.Spell.AbilityLevel != targetLevel {
			if err := s.dbClient.Spell().Update(ctx, member.SpellID, map[string]interface{}{
				"ability_level": targetLevel,
			}); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if member.SpellID == seedSpell.ID {
				seedSpell.AbilityLevel = targetLevel
			}
		}
		if member.SpellID == seedSpell.ID {
			normalizedSeedBand := targetLevel
			seedBand = normalizedSeedBand
			if normalizedSeedBand != member.LevelBand {
				seedBandNeedsNormalization = true
			}
		}
	}
	if seedBand == 0 {
		inferredSeedBand := inferSpellProgressionBand(seedSpell)
		seedBand = selectSeedBandForProgression(inferredSeedBand, occupiedBands)
		if err := s.dbClient.Spell().UpsertProgressionMember(ctx, progression.ID, seedSpell.ID, seedBand); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		occupiedBands[seedBand] = seedSpell.ID
	} else if seedBandNeedsNormalization {
		if err := s.dbClient.Spell().UpsertProgressionMember(ctx, progression.ID, seedSpell.ID, seedBand); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		occupiedBands[seedBand] = seedSpell.ID
	}
	if seedSpell.AbilityLevel != seedBand {
		if err := s.dbClient.Spell().Update(ctx, seedSpell.ID, map[string]interface{}{
			"ability_level": seedBand,
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		seedSpell.AbilityLevel = seedBand
	}

	existingSpells, err := s.dbClient.Spell().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	usedNames := map[string]struct{}{}
	for _, spell := range existingSpells {
		normalized := strings.ToLower(strings.TrimSpace(spell.Name))
		if normalized == "" {
			continue
		}
		usedNames[normalized] = struct{}{}
	}

	createdSpells := make([]models.Spell, 0)
	missingBands := make([]int, 0, len(spellProgressionLevelBands))
	for _, targetBand := range spellProgressionLevelBands {
		if _, occupied := occupiedBands[targetBand]; occupied {
			continue
		}
		missingBands = append(missingBands, targetBand)
	}
	llmFlavors := s.generateSpellProgressionVariantFlavors(seedSpell, seedBand, missingBands, usedNames, abilityType)
	for _, targetBand := range spellProgressionLevelBands {
		if _, occupied := occupiedBands[targetBand]; occupied {
			continue
		}
		var flavorOverride *generatedSpellProgressionVariantFlavor
		if llmFlavor, exists := llmFlavors[targetBand]; exists {
			llmFlavorCopy := llmFlavor
			flavorOverride = &llmFlavorCopy
		}
		variant := buildSpellProgressionVariant(seedSpell, seedBand, targetBand, usedNames, flavorOverride, abilityType)
		if err := s.dbClient.Spell().Create(ctx, variant); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := s.dbClient.Spell().UpsertProgressionMember(ctx, progression.ID, variant.ID, targetBand); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		occupiedBands[targetBand] = variant.ID
		createdSpells = append(createdSpells, *variant)
	}

	updatedProgression, err := s.dbClient.Spell().FindProgressionBySpellID(ctx, seedSpell.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"progression":   updatedProgression,
		"createdCount":  len(createdSpells),
		"createdSpells": createdSpells,
	})
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

func (s *server) applySpellReviveToUser(
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
	if currentHealth > 0 {
		return 0, currentHealth, maxHealth, nil
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

func spellHasCastableEffect(spell *models.Spell) bool {
	if spell == nil {
		return false
	}
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage,
			models.SpellEffectTypeDealDamageAllEnemies,
			models.SpellEffectTypeRestoreLifePartyMember,
			models.SpellEffectTypeRestoreLifeAllParty,
			models.SpellEffectTypeRevivePartyMember,
			models.SpellEffectTypeReviveAllDownedParty,
			models.SpellEffectTypeApplyBeneficialStatus,
			models.SpellEffectTypeApplyDetrimentalStatus,
			models.SpellEffectTypeApplyDetrimentalAll,
			models.SpellEffectTypeRemoveDetrimental:
			return true
		}
	}
	return false
}

func spellDealsMonsterDamage(spell *models.Spell) bool {
	if spell == nil {
		return false
	}
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage,
			models.SpellEffectTypeDealDamageAllEnemies:
			if effect.Amount > 0 {
				return true
			}
		}
	}
	return false
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
	var userSpellToCast *models.UserSpell
	for _, userSpell := range userSpells {
		if userSpell.SpellID == spellID {
			spell := userSpell.Spell
			if requiredType != nil && normalizeSpellAbilityType(string(spell.AbilityType)) != *requiredType {
				continue
			}
			spellToCast = &spell
			userSpellCopy := userSpell
			userSpellToCast = &userSpellCopy
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
	now := time.Now()
	cooldownRemaining := cooldownTurnsRemaining(*userSpellToCast, now)
	cooldownSeconds := cooldownSecondsRemaining(*userSpellToCast, now)
	if cooldownRemaining > 0 {
		label := "spell"
		if isTechnique {
			label = "technique"
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":                    fmt.Sprintf("%s is on cooldown", label),
			"cooldownTurnsRemaining":   cooldownRemaining,
			"cooldownSecondsRemaining": cooldownSeconds,
		})
		return
	}

	targetHealAmount := 0
	groupHealAmount := 0
	targetReviveAmount := 0
	groupReviveAmount := 0
	statusesToApply := models.ScenarioFailureStatusTemplates{}
	statusNamesToRemove := make([]string, 0)
	requiresMonsterTargetForStatuses := false
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
		case models.SpellEffectTypeRevivePartyMember:
			if effect.Amount > 0 {
				targetReviveAmount += effect.Amount
			}
		case models.SpellEffectTypeReviveAllDownedParty:
			if effect.Amount > 0 {
				groupReviveAmount += effect.Amount
			}
		case models.SpellEffectTypeApplyBeneficialStatus:
			statusesToApply = append(statusesToApply, effect.StatusesToApply...)
		case models.SpellEffectTypeApplyDetrimentalStatus,
			models.SpellEffectTypeApplyDetrimentalAll:
			requiresMonsterTargetForStatuses = true
			statusesToApply = append(
				statusesToApply,
				normalizeSpellStatusesForEffectType(effect.Type, effect.StatusesToApply)...,
			)
		case models.SpellEffectTypeRemoveDetrimental:
			statusNamesToRemove = append(statusNamesToRemove, effect.StatusesToRemove...)
		}
	}
	statusesToRemove := normalizeSpellStatusNames(statusNamesToRemove)
	hasStatusEffects := len(statusesToApply) > 0 || len(statusesToRemove) > 0

	if targetHealAmount <= 0 &&
		groupHealAmount <= 0 &&
		targetReviveAmount <= 0 &&
		groupReviveAmount <= 0 &&
		!hasStatusEffects &&
		!spellHasCastableEffect(spellToCast) {
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
	if (targetHealAmount > 0 || targetReviveAmount > 0) && !hasTargetUserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId is required for targeted heal or revive abilities"})
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
	if requiresMonsterTargetForStatuses && targetMonsterID == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetMonsterId is required for detrimental status abilities"})
		return
	}
	advanceCombatTurnOnCast := targetMonsterID != nil && !spellDealsMonsterDamage(spellToCast)
	var monsterBattle *models.MonsterBattle
	if targetMonsterID != nil {
		monsterBattle, err = s.getOrCreateActiveMonsterBattle(ctx, user.ID, *targetMonsterID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		monsterBattle, err = s.refreshMonsterBattleInviteState(ctx, monsterBattle.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if monsterBattle.State != string(models.MonsterBattleStateActive) {
			detail, detailErr := s.monsterBattleDetailResponse(ctx, monsterBattle)
			if detailErr != nil {
				ctx.JSON(http.StatusConflict, gin.H{"error": "battle is waiting for party invite responses"})
				return
			}
			ctx.JSON(http.StatusConflict, gin.H{
				"error":  "battle is waiting for party invite responses",
				"battle": detail,
			})
			return
		}
	}
	if targetMonsterID != nil {
		log.Printf(
			"[combat][cast] user=%s spell=%s abilityType=%s battle=%s monster=%s damageSpell=%t applyOnCast=%t statusesApply=%d statusesRemove=%d",
			user.ID,
			spellID,
			abilityType,
			monsterBattle.ID,
			*targetMonsterID,
			spellDealsMonsterDamage(spellToCast),
			advanceCombatTurnOnCast,
			len(statusesToApply),
			len(statusesToRemove),
		)
	}

	allowedTargets := map[uuid.UUID]bool{
		user.ID: true,
	}
	if targetHealAmount > 0 ||
		groupHealAmount > 0 ||
		targetReviveAmount > 0 ||
		groupReviveAmount > 0 ||
		hasTargetUserID {
		if monsterBattle != nil {
			participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, monsterBattle.ID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			for _, participant := range participants {
				allowedTargets[participant.UserID] = true
			}
		} else {
			partyMembers, err := s.dbClient.User().FindPartyMembers(ctx, user.ID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			for _, member := range partyMembers {
				allowedTargets[member.ID] = true
			}
		}
	}
	if hasTargetUserID {
		if !allowedTargets[targetUserID] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "targetUserId must be in your party"})
			return
		}
	}
	if advanceCombatTurnOnCast {
		if err := s.advanceUserCooldownsForCombatTurn(ctx, user.ID, &spellID, now); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if monsterBattle != nil {
			if err := s.advanceMonsterCooldownsForCombatTurn(ctx, monsterBattle, nil, now); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
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
	reviveByUser := map[uuid.UUID]int{}
	if targetReviveAmount > 0 {
		reviveByUser[targetUserID] += targetReviveAmount
	}
	if groupReviveAmount > 0 {
		for recipientID := range allowedTargets {
			reviveByUser[recipientID] += groupReviveAmount
		}
	}

	heals := []castSpellHealResult{}
	healResultIndexByUser := map[uuid.UUID]int{}
	upsertHealResult := func(result castSpellHealResult) {
		if existingIndex, exists := healResultIndexByUser[result.UserID]; exists {
			existing := heals[existingIndex]
			existing.Restored += result.Restored
			existing.Health = result.Health
			existing.MaxHealth = result.MaxHealth
			heals[existingIndex] = existing
			return
		}
		healResultIndexByUser[result.UserID] = len(heals)
		heals = append(heals, result)
	}
	for recipientID, totalRevive := range reviveByUser {
		restored, health, maxHealth, err := s.applySpellReviveToUser(ctx, recipientID, totalRevive)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if restored <= 0 {
			continue
		}
		upsertHealResult(castSpellHealResult{
			UserID:    recipientID,
			Restored:  restored,
			Health:    health,
			MaxHealth: maxHealth,
		})
	}
	for recipientID, totalHeal := range healByUser {
		restored, health, maxHealth, err := s.applySpellHealToUser(ctx, recipientID, totalHeal)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if restored <= 0 {
			continue
		}
		upsertHealResult(castSpellHealResult{
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
	battleTurnUserDotDamage := 0
	battleTurnMonsterDotDamage := 0

	if hasStatusEffects {
		if targetMonsterID != nil {
			if monsterBattle == nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "monster battle unavailable"})
				return
			}
			monsterBattleID = &monsterBattle.ID
			if err := s.dbClient.MonsterBattle().Touch(ctx, monsterBattle.ID, now); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			activeMonsterStatusNames := make([]string, 0, len(statusesToApply))
			for _, statusTemplate := range statusesToApply {
				name := strings.TrimSpace(statusTemplate.Name)
				if name == "" || statusTemplate.DurationSeconds <= 0 {
					continue
				}
				activeMonsterStatusNames = append(activeMonsterStatusNames, name)
			}
			if err := s.dbClient.MonsterStatus().DeleteActiveByBattleIDAndNames(ctx, monsterBattle.ID, activeMonsterStatusNames); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, statusTemplate := range statusesToApply {
				name := strings.TrimSpace(statusTemplate.Name)
				if name == "" || statusTemplate.DurationSeconds <= 0 {
					continue
				}
				if normalizeUserStatusEffectType(statusTemplate.EffectType) == models.UserStatusEffectTypeManaOverTime {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "mana_over_time statuses are not supported on monsters"})
					return
				}
				status := &models.MonsterStatus{
					UserID:                        user.ID,
					BattleID:                      monsterBattle.ID,
					MonsterID:                     *targetMonsterID,
					Name:                          name,
					Description:                   strings.TrimSpace(statusTemplate.Description),
					Effect:                        strings.TrimSpace(statusTemplate.Effect),
					Positive:                      statusTemplate.Positive,
					EffectType:                    normalizeMonsterStatusEffectType(statusTemplate.EffectType),
					DamagePerTick:                 statusTemplate.DamagePerTick,
					HealthPerTick:                 statusTemplate.HealthPerTick,
					StrengthMod:                   statusTemplate.StrengthMod,
					DexterityMod:                  statusTemplate.DexterityMod,
					ConstitutionMod:               statusTemplate.ConstitutionMod,
					IntelligenceMod:               statusTemplate.IntelligenceMod,
					WisdomMod:                     statusTemplate.WisdomMod,
					CharismaMod:                   statusTemplate.CharismaMod,
					PhysicalDamageBonusPercent:    statusTemplate.PhysicalDamageBonusPercent,
					PiercingDamageBonusPercent:    statusTemplate.PiercingDamageBonusPercent,
					SlashingDamageBonusPercent:    statusTemplate.SlashingDamageBonusPercent,
					BludgeoningDamageBonusPercent: statusTemplate.BludgeoningDamageBonusPercent,
					FireDamageBonusPercent:        statusTemplate.FireDamageBonusPercent,
					IceDamageBonusPercent:         statusTemplate.IceDamageBonusPercent,
					LightningDamageBonusPercent:   statusTemplate.LightningDamageBonusPercent,
					PoisonDamageBonusPercent:      statusTemplate.PoisonDamageBonusPercent,
					ArcaneDamageBonusPercent:      statusTemplate.ArcaneDamageBonusPercent,
					HolyDamageBonusPercent:        statusTemplate.HolyDamageBonusPercent,
					ShadowDamageBonusPercent:      statusTemplate.ShadowDamageBonusPercent,
					PhysicalResistancePercent:     statusTemplate.PhysicalResistancePercent,
					PiercingResistancePercent:     statusTemplate.PiercingResistancePercent,
					SlashingResistancePercent:     statusTemplate.SlashingResistancePercent,
					BludgeoningResistancePercent:  statusTemplate.BludgeoningResistancePercent,
					FireResistancePercent:         statusTemplate.FireResistancePercent,
					IceResistancePercent:          statusTemplate.IceResistancePercent,
					LightningResistancePercent:    statusTemplate.LightningResistancePercent,
					PoisonResistancePercent:       statusTemplate.PoisonResistancePercent,
					ArcaneResistancePercent:       statusTemplate.ArcaneResistancePercent,
					HolyResistancePercent:         statusTemplate.HolyResistancePercent,
					ShadowResistancePercent:       statusTemplate.ShadowResistancePercent,
					StartedAt:                     now,
					ExpiresAt:                     now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
				}
				if err := s.dbClient.MonsterStatus().Create(ctx, status); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				appliedMonsterStatuses = append(appliedMonsterStatuses, scenarioAppliedFailureStatus{
					Name:            status.Name,
					Description:     status.Description,
					Effect:          status.Effect,
					EffectType:      string(status.EffectType),
					Positive:        status.Positive,
					DamagePerTick:   status.DamagePerTick,
					HealthPerTick:   status.HealthPerTick,
					ManaPerTick:     0,
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
			activeUserStatusNames := make([]string, 0, len(statusesToApply))
			for _, statusTemplate := range statusesToApply {
				name := strings.TrimSpace(statusTemplate.Name)
				if name == "" || statusTemplate.DurationSeconds <= 0 {
					continue
				}
				activeUserStatusNames = append(activeUserStatusNames, name)
			}
			if err := s.dbClient.UserStatus().DeleteActiveByUserIDAndNames(ctx, statusTargetUserID, activeUserStatusNames); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			for _, statusTemplate := range statusesToApply {
				name := strings.TrimSpace(statusTemplate.Name)
				if name == "" || statusTemplate.DurationSeconds <= 0 {
					continue
				}
				status := &models.UserStatus{
					UserID:                        statusTargetUserID,
					Name:                          name,
					Description:                   strings.TrimSpace(statusTemplate.Description),
					Effect:                        strings.TrimSpace(statusTemplate.Effect),
					Positive:                      statusTemplate.Positive,
					EffectType:                    normalizeUserStatusEffectType(statusTemplate.EffectType),
					DamagePerTick:                 statusTemplate.DamagePerTick,
					HealthPerTick:                 statusTemplate.HealthPerTick,
					ManaPerTick:                   statusTemplate.ManaPerTick,
					StrengthMod:                   statusTemplate.StrengthMod,
					DexterityMod:                  statusTemplate.DexterityMod,
					ConstitutionMod:               statusTemplate.ConstitutionMod,
					IntelligenceMod:               statusTemplate.IntelligenceMod,
					WisdomMod:                     statusTemplate.WisdomMod,
					CharismaMod:                   statusTemplate.CharismaMod,
					PhysicalDamageBonusPercent:    statusTemplate.PhysicalDamageBonusPercent,
					PiercingDamageBonusPercent:    statusTemplate.PiercingDamageBonusPercent,
					SlashingDamageBonusPercent:    statusTemplate.SlashingDamageBonusPercent,
					BludgeoningDamageBonusPercent: statusTemplate.BludgeoningDamageBonusPercent,
					FireDamageBonusPercent:        statusTemplate.FireDamageBonusPercent,
					IceDamageBonusPercent:         statusTemplate.IceDamageBonusPercent,
					LightningDamageBonusPercent:   statusTemplate.LightningDamageBonusPercent,
					PoisonDamageBonusPercent:      statusTemplate.PoisonDamageBonusPercent,
					ArcaneDamageBonusPercent:      statusTemplate.ArcaneDamageBonusPercent,
					HolyDamageBonusPercent:        statusTemplate.HolyDamageBonusPercent,
					ShadowDamageBonusPercent:      statusTemplate.ShadowDamageBonusPercent,
					PhysicalResistancePercent:     statusTemplate.PhysicalResistancePercent,
					PiercingResistancePercent:     statusTemplate.PiercingResistancePercent,
					SlashingResistancePercent:     statusTemplate.SlashingResistancePercent,
					BludgeoningResistancePercent:  statusTemplate.BludgeoningResistancePercent,
					FireResistancePercent:         statusTemplate.FireResistancePercent,
					IceResistancePercent:          statusTemplate.IceResistancePercent,
					LightningResistancePercent:    statusTemplate.LightningResistancePercent,
					PoisonResistancePercent:       statusTemplate.PoisonResistancePercent,
					ArcaneResistancePercent:       statusTemplate.ArcaneResistancePercent,
					HolyResistancePercent:         statusTemplate.HolyResistancePercent,
					ShadowResistancePercent:       statusTemplate.ShadowResistancePercent,
					StartedAt:                     now,
					ExpiresAt:                     now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
				}
				if err := s.dbClient.UserStatus().Create(ctx, status); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				appliedUserStatuses = append(appliedUserStatuses, scenarioAppliedFailureStatus{
					Name:            status.Name,
					Description:     status.Description,
					Effect:          status.Effect,
					EffectType:      string(status.EffectType),
					Positive:        status.Positive,
					DamagePerTick:   status.DamagePerTick,
					HealthPerTick:   status.HealthPerTick,
					ManaPerTick:     status.ManaPerTick,
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
	if spellToCast.CooldownTurns > 0 {
		cooldownExpiresAt := cooldownExpiresAtFromTurns(spellToCast.CooldownTurns, now)
		if err := s.dbClient.UserSpell().UpdateCooldownExpiresAt(ctx, user.ID, spellToCast.ID, cooldownExpiresAt); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if monsterBattle != nil {
		monsterBattleID = &monsterBattle.ID
	}
	if monsterBattle != nil && advanceCombatTurnOnCast {
		if err := s.dbClient.MonsterBattle().Touch(ctx, monsterBattle.ID, time.Now()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userDotDamage, monsterDotDamage, dotErr := s.applyBattleTurnDamageOverTime(ctx, user.ID, monsterBattle.ID)
		if dotErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": dotErr.Error()})
			return
		}
		battleTurnUserDotDamage = userDotDamage
		battleTurnMonsterDotDamage = monsterDotDamage
		monsterBattle.MonsterHealthDeficit += monsterDotDamage
		if err := s.advanceBattleStatusDurations(ctx, user.ID, monsterBattle.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		monsterBattle, err = s.finalizeMonsterBattleIfDefeated(ctx, monsterBattle)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		monsterBattle, err = s.advanceMonsterBattleTurnState(ctx, monsterBattle, &user.ID, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if monsterBattle != nil && advanceCombatTurnOnCast {
		totalHeal := 0
		for _, heal := range heals {
			totalHeal += max(0, heal.Restored)
		}
		var targetUserUUID *uuid.UUID
		var targetMonsterUUID *uuid.UUID
		targetName := ""
		if hasTargetUserID {
			targetUserUUID = &targetUserID
			targetUser, targetErr := s.dbClient.User().FindByID(ctx, targetUserID)
			if targetErr != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": targetErr.Error()})
				return
			}
			targetName = monsterBattleUserDisplayName(targetUser)
		} else if len(heals) > 1 {
			targetName = "the party"
		}
		if targetMonsterID != nil && !hasTargetUserID {
			targetMonsterUUID = targetMonsterID
			targetMonster, targetErr := s.dbClient.Monster().FindByID(ctx, *targetMonsterID)
			if targetErr != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": targetErr.Error()})
				return
			}
			targetName = strings.TrimSpace(targetMonster.Name)
		}
		if err := s.recordMonsterBattleLastAction(ctx, monsterBattle, models.MonsterBattleLastAction{
			ActionType:      "ability",
			ActorType:       "user",
			ActorUserID:     &user.ID,
			ActorName:       monsterBattleUserDisplayName(user),
			AbilityID:       &spellToCast.ID,
			AbilityName:     strings.TrimSpace(spellToCast.Name),
			AbilityType:     string(abilityType),
			TargetUserID:    targetUserUUID,
			TargetMonsterID: targetMonsterUUID,
			TargetName:      targetName,
			Heal:            totalHeal,
			StatusesApplied: len(appliedUserStatuses) + len(appliedMonsterStatuses),
			StatusesRemoved: len(removedUserStatuses) + len(removedMonsterStatuses),
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	_, _, maxMana, _, manaAfter, err := s.getScenarioResourceState(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"spellId":                  spellToCast.ID,
		"spellName":                spellToCast.Name,
		"abilityType":              string(abilityType),
		"cooldownTurnsRemaining":   spellToCast.CooldownTurns,
		"cooldownSecondsRemaining": spellToCast.CooldownTurns * int(combatTurnDuration/time.Second),
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
	if monsterBattle != nil && monsterBattle.EndedAt != nil {
		response["battleEndedAt"] = monsterBattle.EndedAt
	}
	if monsterBattle != nil {
		battleDetail, detailErr := s.monsterBattleDetailResponse(ctx, monsterBattle)
		if detailErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": detailErr.Error()})
			return
		}
		response["battleDetail"] = battleDetail
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
	if battleTurnUserDotDamage > 0 {
		response["battleTurnUserDotDamage"] = battleTurnUserDotDamage
	}
	if battleTurnMonsterDotDamage > 0 {
		response["battleTurnMonsterDotDamage"] = battleTurnMonsterDotDamage
	}
	if monsterBattle != nil {
		log.Printf(
			"[combat][cast][result] user=%s spell=%s battle=%s turnIndex=%d monsterHealthDeficit=%d userDot=%d monsterDot=%d monsterStatusesApplied=%d monsterStatusesRemoved=%d",
			user.ID,
			spellID,
			monsterBattle.ID,
			monsterBattle.TurnIndex,
			monsterBattle.MonsterHealthDeficit,
			battleTurnUserDotDamage,
			battleTurnMonsterDotDamage,
			len(appliedMonsterStatuses),
			len(removedMonsterStatuses),
		)
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
