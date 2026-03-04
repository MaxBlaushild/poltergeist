package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const spellProgressionFromPromptSpellTemplate = `
You are designing ONE fantasy RPG spell concept from a creator brief.

Creator brief:
%s

Return JSON only:
{
  "spell": {
    "name": "2-4 words",
    "description": "20-60 words",
    "effectText": "one concise sentence",
    "schoolOfMagic": "short label",
    "manaCost": 0-60,
    "preferredEffectType": "deal_damage|restore_life_party_member|restore_life_all_party_members|apply_beneficial_statuses|remove_detrimental_statuses"
  }
}

Rules:
- Keep tone adventurous and original.
- Keep it safe and suitable for a public game.
- No copyrighted franchises or references.
`

const spellProgressionFromPromptTechniqueTemplate = `
You are designing ONE fantasy RPG combat technique concept from a creator brief.

Creator brief:
%s

Return JSON only:
{
  "spell": {
    "name": "2-4 words",
    "description": "20-60 words",
    "effectText": "one concise sentence",
    "schoolOfMagic": "short label, usually Martial",
    "manaCost": 0,
    "preferredEffectType": "deal_damage|restore_life_party_member|restore_life_all_party_members|apply_beneficial_statuses|remove_detrimental_statuses"
  }
}

Rules:
- This is a technique, not a spell.
- Keep it practical and grounded in physical execution.
- manaCost must be 0.
- Keep it safe and suitable for a public game.
- No copyrighted franchises or references.
`

var spellProgressionFromPromptLevelBands = []int{10, 25, 50, 70}

type generatedSpellProgressionPromptEnvelope struct {
	Spell               *generatedSpellProgressionPromptPayload `json:"spell"`
	Name                string                                  `json:"name"`
	Description         string                                  `json:"description"`
	EffectText          string                                  `json:"effectText"`
	SchoolOfMagic       string                                  `json:"schoolOfMagic"`
	ManaCost            *int                                    `json:"manaCost"`
	PreferredEffectType string                                  `json:"preferredEffectType"`
}

type generatedSpellProgressionPromptPayload struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	EffectText          string `json:"effectText"`
	SchoolOfMagic       string `json:"schoolOfMagic"`
	ManaCost            *int   `json:"manaCost"`
	PreferredEffectType string `json:"preferredEffectType"`
}

type GenerateSpellProgressionFromPromptProcessor struct {
	dbClient         db.DbClient
	redisClient      *redis.Client
	deepPriestClient deep_priest.DeepPriest
}

type spellProgressionFromPromptResult struct {
	ProgressionID uuid.UUID
	SeedSpellID   uuid.UUID
	CreatedSpell  []uuid.UUID
}

func NewGenerateSpellProgressionFromPromptProcessor(
	dbClient db.DbClient,
	redisClient *redis.Client,
	deepPriestClient deep_priest.DeepPriest,
) GenerateSpellProgressionFromPromptProcessor {
	log.Println("Initializing GenerateSpellProgressionFromPromptProcessor")
	return GenerateSpellProgressionFromPromptProcessor{
		dbClient:         dbClient,
		redisClient:      redisClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateSpellProgressionFromPromptProcessor) ProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	log.Printf("Processing spell progression prompt generation task: %v", task.Type())

	var payload jobs.GenerateSpellProgressionFromPromptTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}
	prompt := strings.TrimSpace(payload.Prompt)
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	abilityType := models.NormalizeSpellAbilityType(strings.TrimSpace(payload.AbilityType))
	if abilityType == "" {
		abilityType = models.SpellAbilityTypeSpell
	}

	statusKey := jobs.SpellProgressionPromptStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.SpellProgressionPromptStatus{
		JobID:        payload.JobID,
		Status:       jobs.SpellProgressionPromptStatusInProgress,
		Prompt:       prompt,
		AbilityType:  string(abilityType),
		CreatedCount: 0,
		StartedAt:    &now,
		UpdatedAt:    now,
	}
	p.setSpellProgressionPromptStatus(ctx, statusKey, status)

	result, err := p.generateSpellProgressionFromPrompt(ctx, prompt, abilityType)
	if err != nil {
		p.failSpellProgressionPromptStatus(ctx, statusKey, status, err)
		return err
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.SpellProgressionPromptStatusCompleted
	status.ProgressionID = &result.ProgressionID
	status.SeedSpellID = &result.SeedSpellID
	status.CreatedSpellIDs = result.CreatedSpell
	status.CreatedCount = len(result.CreatedSpell)
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setSpellProgressionPromptStatus(ctx, statusKey, status)
	return nil
}

func (p *GenerateSpellProgressionFromPromptProcessor) generateSpellProgressionFromPrompt(
	ctx context.Context,
	prompt string,
	abilityType models.SpellAbilityType,
) (*spellProgressionFromPromptResult, error) {
	existingSpells, err := p.dbClient.Spell().FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing spells: %w", err)
	}
	usedNames := map[string]struct{}{}
	for _, existing := range existingSpells {
		key := normalizeAbilityName(existing.Name)
		if key == "" {
			continue
		}
		usedNames[key] = struct{}{}
	}

	seedSpec, preferredEffectType, err := p.buildSeedSpec(ctx, prompt, abilityType)
	if err != nil {
		return nil, fmt.Errorf("failed to build seed spell spec: %w", err)
	}
	seedManaCost := seedSpec.ManaCost
	if seedManaCost < 0 {
		seedManaCost = 0
	}
	if seedManaCost > 300 {
		seedManaCost = 300
	}
	if abilityType == models.SpellAbilityTypeTechnique {
		seedManaCost = 0
	}
	effects := inferGeneratedAbilityEffectsWithPreference(
		seedSpec,
		abilityType,
		seedManaCost,
		preferredEffectType,
		nil,
	)
	if abilityType == models.SpellAbilityTypeTechnique {
		effects = reduceTechniqueProgressionEffectsPower(effects)
	}
	seedName := harmonizeGeneratedAbilityNameWithEffects(
		strings.TrimSpace(seedSpec.Name),
		abilityType,
		effects,
	)
	seedName = reserveGeneratedAbilityName(seedName, string(abilityType), 1, usedNames)
	seedDescription := harmonizeGeneratedAbilityDescriptionWithEffects(
		strings.TrimSpace(seedSpec.Description),
		abilityType,
		effects,
	)
	seedEffectText := strings.TrimSpace(seedSpec.EffectText)
	if seedEffectText == "" {
		seedEffectText = buildGeneratedAbilityEffectText(effects, abilityType)
	}
	school := strings.TrimSpace(seedSpec.SchoolOfMagic)
	if school == "" {
		if abilityType == models.SpellAbilityTypeTechnique {
			school = "Martial"
		} else {
			school = "Arcane"
		}
	}

	emptyError := ""
	now := time.Now()
	seedSpell := &models.Spell{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		Name:                  seedName,
		Description:           seedDescription,
		IconURL:               "",
		ImageGenerationStatus: models.SpellImageGenerationStatusNone,
		ImageGenerationError:  &emptyError,
		AbilityType:           abilityType,
		EffectText:            seedEffectText,
		SchoolOfMagic:         school,
		ManaCost:              seedManaCost,
		Effects:               effects,
	}
	if err := p.dbClient.Spell().Create(ctx, seedSpell); err != nil {
		return nil, fmt.Errorf("failed to create seed spell: %w", err)
	}

	progression := &models.SpellProgression{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:        fmt.Sprintf("%s Progression", strings.TrimSpace(seedSpell.Name)),
		AbilityType: abilityType,
	}
	if strings.TrimSpace(progression.Name) == "Progression" {
		progression.Name = fmt.Sprintf("%s Progression", promptSpellProgressionTheme(seedSpell))
	}
	if err := p.dbClient.Spell().CreateProgression(ctx, progression); err != nil {
		return nil, fmt.Errorf("failed to create spell progression: %w", err)
	}

	seedBand := promptInferSpellProgressionBand(seedSpell)
	if err := p.dbClient.Spell().UpsertProgressionMember(ctx, progression.ID, seedSpell.ID, seedBand); err != nil {
		return nil, fmt.Errorf("failed to assign seed spell to progression: %w", err)
	}

	createdSpellIDs := []uuid.UUID{seedSpell.ID}
	for _, targetBand := range spellProgressionFromPromptLevelBands {
		if targetBand == seedBand {
			continue
		}
		variant := buildPromptSpellProgressionVariant(seedSpell, seedBand, targetBand, usedNames, abilityType)
		if err := p.dbClient.Spell().Create(ctx, variant); err != nil {
			return nil, fmt.Errorf("failed to create spell variant for level %d: %w", targetBand, err)
		}
		if err := p.dbClient.Spell().UpsertProgressionMember(ctx, progression.ID, variant.ID, targetBand); err != nil {
			return nil, fmt.Errorf("failed to assign spell variant for level %d: %w", targetBand, err)
		}
		createdSpellIDs = append(createdSpellIDs, variant.ID)
	}

	return &spellProgressionFromPromptResult{
		ProgressionID: progression.ID,
		SeedSpellID:   seedSpell.ID,
		CreatedSpell:  createdSpellIDs,
	}, nil
}

func (p *GenerateSpellProgressionFromPromptProcessor) buildSeedSpec(
	ctx context.Context,
	prompt string,
	abilityType models.SpellAbilityType,
) (jobs.SpellCreationSpec, models.SpellEffectType, error) {
	fallback := fallbackSpellSpecFromPrompt(prompt, abilityType)
	if p.deepPriestClient == nil {
		return fallback, fallbackPreferredEffectType(prompt), nil
	}

	promptTemplate := spellProgressionFromPromptSpellTemplate
	if abilityType == models.SpellAbilityTypeTechnique {
		promptTemplate = spellProgressionFromPromptTechniqueTemplate
	}
	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{
		Question: fmt.Sprintf(promptTemplate, prompt),
	})
	if err != nil {
		return fallback, fallbackPreferredEffectType(prompt), nil
	}

	spec, preferred, parseErr := parseGeneratedSpellProgressionPromptSpec(answer.Answer)
	if parseErr != nil {
		return fallback, fallbackPreferredEffectType(prompt), nil
	}
	if strings.TrimSpace(spec.Name) == "" {
		spec.Name = fallback.Name
	}
	if strings.TrimSpace(spec.Description) == "" {
		spec.Description = fallback.Description
	}
	if strings.TrimSpace(spec.EffectText) == "" {
		spec.EffectText = fallback.EffectText
	}
	if strings.TrimSpace(spec.SchoolOfMagic) == "" {
		spec.SchoolOfMagic = fallback.SchoolOfMagic
	}
	if spec.ManaCost < 0 || spec.ManaCost > 300 {
		spec.ManaCost = fallback.ManaCost
	}
	if abilityType == models.SpellAbilityTypeTechnique {
		spec.ManaCost = 0
	}
	spec.AbilityType = string(abilityType)
	return spec, preferred, nil
}

func parseGeneratedSpellProgressionPromptSpec(
	raw string,
) (jobs.SpellCreationSpec, models.SpellEffectType, error) {
	parsed := generatedSpellProgressionPromptEnvelope{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(raw)), &parsed); err != nil {
		return jobs.SpellCreationSpec{}, "", err
	}

	payload := generatedSpellProgressionPromptPayload{
		Name:                parsed.Name,
		Description:         parsed.Description,
		EffectText:          parsed.EffectText,
		SchoolOfMagic:       parsed.SchoolOfMagic,
		ManaCost:            parsed.ManaCost,
		PreferredEffectType: parsed.PreferredEffectType,
	}
	if parsed.Spell != nil {
		payload = *parsed.Spell
	}

	spec := jobs.SpellCreationSpec{
		Name:          strings.TrimSpace(payload.Name),
		Description:   strings.TrimSpace(payload.Description),
		EffectText:    strings.TrimSpace(payload.EffectText),
		SchoolOfMagic: strings.TrimSpace(payload.SchoolOfMagic),
		AbilityType:   string(models.SpellAbilityTypeSpell),
		ManaCost:      18,
	}
	if payload.ManaCost != nil {
		spec.ManaCost = *payload.ManaCost
	}
	preferred := parsePreferredSpellEffectType(payload.PreferredEffectType)
	return spec, preferred, nil
}

func parsePreferredSpellEffectType(raw string) models.SpellEffectType {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case string(models.SpellEffectTypeDealDamage):
		return models.SpellEffectTypeDealDamage
	case string(models.SpellEffectTypeRestoreLifePartyMember):
		return models.SpellEffectTypeRestoreLifePartyMember
	case string(models.SpellEffectTypeRestoreLifeAllParty):
		return models.SpellEffectTypeRestoreLifeAllParty
	case string(models.SpellEffectTypeApplyBeneficialStatus):
		return models.SpellEffectTypeApplyBeneficialStatus
	case string(models.SpellEffectTypeRemoveDetrimental):
		return models.SpellEffectTypeRemoveDetrimental
	default:
		return ""
	}
}

func reduceTechniqueProgressionEffectsPower(effects models.SpellEffects) models.SpellEffects {
	if len(effects) == 0 {
		return effects
	}
	scaled := make(models.SpellEffects, 0, len(effects))
	for _, effect := range effects {
		next := effect
		if next.Amount > 0 {
			next.Amount = promptMaxInt(1, int(math.Round(float64(next.Amount)*0.82)))
		}
		if len(next.StatusesToApply) > 0 {
			statuses := make(models.ScenarioFailureStatusTemplates, 0, len(next.StatusesToApply))
			for _, status := range next.StatusesToApply {
				scaledStatus := status
				if scaledStatus.DamagePerTick > 0 {
					scaledStatus.DamagePerTick = promptMaxInt(1, int(math.Round(float64(scaledStatus.DamagePerTick)*0.82)))
				}
				statuses = append(statuses, scaledStatus)
			}
			next.StatusesToApply = statuses
		}
		scaled = append(scaled, next)
	}
	return scaled
}

func fallbackSpellSpecFromPrompt(
	prompt string,
	abilityType models.SpellAbilityType,
) jobs.SpellCreationSpec {
	name := fallbackSpellName(prompt, abilityType)
	description := strings.TrimSpace(prompt)
	if description == "" {
		if abilityType == models.SpellAbilityTypeTechnique {
			description = "A disciplined combat technique forged through repetition and precision."
		} else {
			description = "A focused magical technique forged from raw intent."
		}
	}
	effectText := description
	school := fallbackSpellSchool(prompt, abilityType)
	manaCost := 18
	if abilityType == models.SpellAbilityTypeTechnique {
		manaCost = 0
	}
	return jobs.SpellCreationSpec{
		Name:          name,
		Description:   description,
		EffectText:    effectText,
		SchoolOfMagic: school,
		AbilityType:   string(abilityType),
		ManaCost:      manaCost,
	}
}

func fallbackPreferredEffectType(prompt string) models.SpellEffectType {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch {
	case containsAnyKeyword(lower, []string{"heal", "restore", "mend", "recover", "support"}):
		if containsAnyKeyword(lower, []string{"all", "party", "group", "team", "everyone"}) {
			return models.SpellEffectTypeRestoreLifeAllParty
		}
		return models.SpellEffectTypeRestoreLifePartyMember
	case containsAnyKeyword(lower, []string{"cleanse", "purge", "remove curse", "dispel"}):
		return models.SpellEffectTypeRemoveDetrimental
	case containsAnyKeyword(lower, []string{"buff", "ward", "aegis", "shield", "fortify"}):
		return models.SpellEffectTypeApplyBeneficialStatus
	default:
		return models.SpellEffectTypeDealDamage
	}
}

func fallbackSpellName(prompt string, abilityType models.SpellAbilityType) string {
	if abilityType == models.SpellAbilityTypeTechnique {
		return "Iron Form"
	}
	affinity := inferGeneratedDamageAffinity(prompt, models.SpellAbilityTypeSpell)
	switch models.NormalizeDamageAffinity(affinity) {
	case models.DamageAffinityFire:
		return "Fire Wisp"
	case models.DamageAffinityIce:
		return "Frost Wisp"
	case models.DamageAffinityLightning:
		return "Storm Wisp"
	case models.DamageAffinityPoison:
		return "Venom Wisp"
	case models.DamageAffinityHoly:
		return "Radiant Wisp"
	case models.DamageAffinityShadow:
		return "Umbral Wisp"
	default:
		return "Arcane Wisp"
	}
}

func fallbackSpellSchool(prompt string, abilityType models.SpellAbilityType) string {
	if abilityType == models.SpellAbilityTypeTechnique {
		return "Martial"
	}
	affinity := inferGeneratedDamageAffinity(prompt, models.SpellAbilityTypeSpell)
	switch models.NormalizeDamageAffinity(affinity) {
	case models.DamageAffinityFire:
		return "Pyromancy"
	case models.DamageAffinityIce:
		return "Cryomancy"
	case models.DamageAffinityLightning:
		return "Tempest"
	case models.DamageAffinityPoison:
		return "Venomcraft"
	case models.DamageAffinityHoly:
		return "Radiance"
	case models.DamageAffinityShadow:
		return "Umbral"
	default:
		return "Arcane"
	}
}

func (p *GenerateSpellProgressionFromPromptProcessor) setSpellProgressionPromptStatus(
	ctx context.Context,
	statusKey string,
	status jobs.SpellProgressionPromptStatus,
) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal spell progression prompt status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.SpellProgressionPromptStatusTTL).Err(); err != nil {
		log.Printf("Failed to write spell progression prompt status: %v", err)
	}
}

func (p *GenerateSpellProgressionFromPromptProcessor) failSpellProgressionPromptStatus(
	ctx context.Context,
	statusKey string,
	status jobs.SpellProgressionPromptStatus,
	cause error,
) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.SpellProgressionPromptStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setSpellProgressionPromptStatus(ctx, statusKey, status)
}

func promptNormalizeSpellProgressionBand(levelBand int) int {
	if levelBand <= spellProgressionFromPromptLevelBands[0] {
		return spellProgressionFromPromptLevelBands[0]
	}
	if levelBand >= spellProgressionFromPromptLevelBands[len(spellProgressionFromPromptLevelBands)-1] {
		return spellProgressionFromPromptLevelBands[len(spellProgressionFromPromptLevelBands)-1]
	}

	best := spellProgressionFromPromptLevelBands[0]
	bestDistance := promptAbsInt(levelBand - best)
	for _, candidate := range spellProgressionFromPromptLevelBands[1:] {
		distance := promptAbsInt(levelBand - candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func promptInferSpellProgressionBand(spell *models.Spell) int {
	if spell == nil {
		return 25
	}
	powerScore := float64(promptMaxInt(spell.ManaCost, 0))
	for _, effect := range spell.Effects {
		powerScore += float64(promptMaxInt(effect.Amount, 0))
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
	return promptNormalizeSpellProgressionBand(bestBand)
}

func promptSpellProgressionPrimaryEffectType(spell *models.Spell) models.SpellEffectType {
	if spell == nil || len(spell.Effects) == 0 {
		return models.SpellEffectTypeDealDamage
	}
	return spell.Effects[0].Type
}

func promptSpellProgressionTheme(spell *models.Spell) string {
	if spell != nil && len(spell.Effects) > 0 {
		if affinity := promptFirstSpellDamageAffinity(spell); affinity != "" {
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
		if word := promptFirstWord(spell.SchoolOfMagic); word != "" {
			return promptTitleWord(word)
		}
		if word := promptFirstWord(spell.Name); word != "" {
			return promptTitleWord(word)
		}
	}
	return "Arcane"
}

func promptFirstSpellDamageAffinity(spell *models.Spell) string {
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

func promptSpellProgressionBandTerm(
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

func promptScaleSpellProgressionValue(base int, seedBand int, targetBand int, exponent float64) int {
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

func promptEstimateSpellProgressionMonsterHealth(levelBand int) int {
	level := promptNormalizeSpellProgressionBand(levelBand)
	curve := estimateMonsterCurveForTargetLevel(&level)
	if curve == nil || curve.EstimatedHealth <= 0 {
		baseConstitution := 12
		effectiveConstitution := promptMaxInt(1, baseConstitution+level-1)
		return effectiveConstitution * 10
	}
	return curve.EstimatedHealth
}

func promptSpellProgressionBandRatio(levelBand int, ratios map[int]float64) float64 {
	normalized := promptNormalizeSpellProgressionBand(levelBand)
	if ratio, ok := ratios[normalized]; ok {
		return ratio
	}
	return ratios[25]
}

func promptSpellProgressionTargetAmount(
	effectType models.SpellEffectType,
	levelBand int,
	abilityType models.SpellAbilityType,
) int {
	normalizedBand := promptNormalizeSpellProgressionBand(levelBand)
	if effectType == models.SpellEffectTypeDealDamage {
		damagePerLevel := 5
		if abilityType == models.SpellAbilityTypeTechnique {
			damagePerLevel = 4
		}
		return promptMaxInt(1, normalizedBand*damagePerLevel)
	}

	health := promptEstimateSpellProgressionMonsterHealth(levelBand)
	if health <= 0 {
		return 0
	}

	var ratio float64
	switch effectType {
	case models.SpellEffectTypeRestoreLifePartyMember:
		ratio = promptSpellProgressionBandRatio(levelBand, map[int]float64{
			10: 0.12,
			25: 0.18,
			50: 0.28,
			70: 0.40,
		})
	case models.SpellEffectTypeRestoreLifeAllParty:
		ratio = promptSpellProgressionBandRatio(levelBand, map[int]float64{
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
	return promptMaxInt(1, target)
}

func promptSpellProgressionTargetDamagePerTick(
	levelBand int,
	abilityType models.SpellAbilityType,
) int {
	directDamageTarget := promptSpellProgressionTargetAmount(
		models.SpellEffectTypeDealDamage,
		levelBand,
		abilityType,
	)
	return promptMaxInt(1, int(math.Round(float64(directDamageTarget)*0.2)))
}

func promptScaleSpellProgressionCombatAmount(
	base int,
	effectType models.SpellEffectType,
	seedBand int,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if base == 0 {
		return 0
	}
	legacy := promptScaleSpellProgressionValue(base, seedBand, targetBand, 1.15)
	target := promptSpellProgressionTargetAmount(effectType, targetBand, abilityType)
	if target <= 0 {
		return legacy
	}
	if targetBand > seedBand {
		return promptMaxInt(legacy, target)
	}
	if targetBand < seedBand {
		if legacy < target {
			return legacy
		}
		return target
	}
	return legacy
}

func promptScaleSpellProgressionDamagePerTick(
	base int,
	seedBand int,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if base == 0 {
		return 0
	}
	legacy := promptScaleSpellProgressionValue(base, seedBand, targetBand, 1.05)
	target := promptSpellProgressionTargetDamagePerTick(targetBand, abilityType)
	if target <= 0 {
		return legacy
	}
	if targetBand > seedBand {
		return promptMaxInt(legacy, target)
	}
	if targetBand < seedBand {
		if legacy < target {
			return legacy
		}
		return target
	}
	return legacy
}

func promptSpellProgressionBandFloor(levelBand int, floors map[int]int) int {
	normalized := promptNormalizeSpellProgressionBand(levelBand)
	if floor, ok := floors[normalized]; ok {
		return floor
	}
	return floors[25]
}

func promptEstimateSpellProgressionPlayerMaxMana(levelBand int) int {
	level := promptNormalizeSpellProgressionBand(levelBand)
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
	return promptMaxInt(20, estimatedMana)
}

func promptSpellProgressionTargetManaCost(
	effectType models.SpellEffectType,
	targetBand int,
	abilityType models.SpellAbilityType,
) int {
	if abilityType == models.SpellAbilityTypeTechnique {
		return 0
	}
	playerMana := promptEstimateSpellProgressionPlayerMaxMana(targetBand)

	switch effectType {
	case models.SpellEffectTypeDealDamage:
		bandFloor := promptSpellProgressionBandFloor(targetBand, map[int]int{
			10: 16,
			25: 36,
			50: 90,
			70: 180,
		})
		ratio := promptSpellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.09,
			25: 0.12,
			50: 0.17,
			70: 0.22,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return promptMaxInt(bandFloor, target)
	case models.SpellEffectTypeRestoreLifePartyMember:
		bandFloor := promptSpellProgressionBandFloor(targetBand, map[int]int{
			10: 14,
			25: 30,
			50: 76,
			70: 155,
		})
		ratio := promptSpellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.08,
			25: 0.11,
			50: 0.15,
			70: 0.19,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return promptMaxInt(bandFloor, target)
	case models.SpellEffectTypeRestoreLifeAllParty:
		bandFloor := promptSpellProgressionBandFloor(targetBand, map[int]int{
			10: 18,
			25: 42,
			50: 105,
			70: 210,
		})
		ratio := promptSpellProgressionBandRatio(targetBand, map[int]float64{
			10: 0.10,
			25: 0.14,
			50: 0.20,
			70: 0.27,
		})
		target := int(math.Round(float64(playerMana) * ratio))
		return promptMaxInt(bandFloor, target)
	case models.SpellEffectTypeApplyBeneficialStatus, models.SpellEffectTypeRemoveDetrimental:
		return promptSpellProgressionBandFloor(targetBand, map[int]int{
			10: 12,
			25: 26,
			50: 64,
			70: 130,
		})
	default:
		return promptSpellProgressionBandFloor(targetBand, map[int]int{
			10: 10,
			25: 22,
			50: 56,
			70: 112,
		})
	}
}

func promptScaleSpellProgressionManaCost(
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
	legacy := promptScaleSpellProgressionValue(baseMana, seedBand, targetBand, 1.25)
	target := promptSpellProgressionTargetManaCost(effectType, targetBand, abilityType)
	if target < 1 {
		target = 1
	}

	scaled := legacy
	if targetBand > seedBand {
		scaled = promptMaxInt(legacy, target)
	} else if targetBand < seedBand {
		scaled = promptMinInt(legacy, target)
	}
	if scaled < 1 {
		return 1
	}
	if scaled > 300 {
		return 300
	}
	return scaled
}

func promptCloneSpellEffectData(effectData map[string]interface{}) map[string]interface{} {
	if effectData == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(effectData))
	for key, value := range effectData {
		cloned[key] = value
	}
	return cloned
}

func buildPromptScaledSpellProgressionEffects(
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
			Amount:           promptScaleSpellProgressionCombatAmount(effect.Amount, effect.Type, seedBand, targetBand, abilityType),
			StatusesToRemove: append(models.StringArray(nil), effect.StatusesToRemove...),
			EffectData:       promptCloneSpellEffectData(effect.EffectData),
		}
		if effect.DamageAffinity != nil {
			affinity := strings.TrimSpace(*effect.DamageAffinity)
			next.DamageAffinity = &affinity
		}
		if len(effect.StatusesToApply) > 0 {
			statuses := make(models.ScenarioFailureStatusTemplates, 0, len(effect.StatusesToApply))
			for _, status := range effect.StatusesToApply {
				scaledStatus := status
				scaledStatus.DurationSeconds = promptMaxInt(1, promptScaleSpellProgressionValue(status.DurationSeconds, seedBand, targetBand, 0.35))
				scaledStatus.DamagePerTick = promptScaleSpellProgressionDamagePerTick(status.DamagePerTick, seedBand, targetBand, abilityType)
				scaledStatus.StrengthMod = promptScaleSpellProgressionValue(status.StrengthMod, seedBand, targetBand, 0.4)
				scaledStatus.DexterityMod = promptScaleSpellProgressionValue(status.DexterityMod, seedBand, targetBand, 0.4)
				scaledStatus.ConstitutionMod = promptScaleSpellProgressionValue(status.ConstitutionMod, seedBand, targetBand, 0.4)
				scaledStatus.IntelligenceMod = promptScaleSpellProgressionValue(status.IntelligenceMod, seedBand, targetBand, 0.4)
				scaledStatus.WisdomMod = promptScaleSpellProgressionValue(status.WisdomMod, seedBand, targetBand, 0.4)
				scaledStatus.CharismaMod = promptScaleSpellProgressionValue(status.CharismaMod, seedBand, targetBand, 0.4)
				statuses = append(statuses, scaledStatus)
			}
			next.StatusesToApply = statuses
		}
		scaled = append(scaled, next)
	}
	return scaled
}

func buildPromptSpellProgressionEffectText(effects models.SpellEffects) string {
	if len(effects) == 0 {
		return "A refined magical technique."
	}

	effect := effects[0]
	switch effect.Type {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return fmt.Sprintf("Restore %d health to one ally.", promptMaxInt(effect.Amount, 1))
	case models.SpellEffectTypeRestoreLifeAllParty:
		return fmt.Sprintf("Restore %d health to all allies.", promptMaxInt(effect.Amount, 1))
	case models.SpellEffectTypeApplyBeneficialStatus:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to allies.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies beneficial statuses to allies."
	case models.SpellEffectTypeRemoveDetrimental:
		return "Removes detrimental statuses from allies."
	default:
		affinity := "magical"
		if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
			affinity = strings.TrimSpace(*effect.DamageAffinity)
		}
		return fmt.Sprintf("Deals %d %s damage to a target.", promptMaxInt(effect.Amount, 1), affinity)
	}
}

func buildPromptSpellProgressionVariant(
	seed *models.Spell,
	seedBand int,
	targetBand int,
	usedNames map[string]struct{},
	abilityType models.SpellAbilityType,
) *models.Spell {
	primaryEffect := promptSpellProgressionPrimaryEffectType(seed)
	theme := promptSpellProgressionTheme(seed)
	bandTerm := promptSpellProgressionBandTerm(primaryEffect, targetBand, abilityType)
	name := reserveGeneratedAbilityName(
		fmt.Sprintf("%s %s", theme, bandTerm),
		string(abilityType),
		targetBand,
		usedNames,
	)
	effects := buildPromptScaledSpellProgressionEffects(seed.Effects, seedBand, targetBand, abilityType)
	manaCost := promptScaleSpellProgressionManaCost(promptMaxInt(seed.ManaCost, 1), primaryEffect, seedBand, targetBand, abilityType)
	if abilityType == models.SpellAbilityTypeTechnique {
		manaCost = 0
	}
	description := fmt.Sprintf("Level %d evolution of %s.", targetBand, strings.TrimSpace(seed.Name))
	if trimmed := strings.TrimSpace(seed.Description); trimmed != "" {
		description = fmt.Sprintf("%s %s", description, trimmed)
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
		EffectText:            buildPromptSpellProgressionEffectText(effects),
		SchoolOfMagic:         strings.TrimSpace(seed.SchoolOfMagic),
		ManaCost:              manaCost,
		Effects:               effects,
	}
}

func promptFirstWord(value string) string {
	for _, part := range strings.Fields(strings.TrimSpace(value)) {
		if part != "" {
			return part
		}
	}
	return ""
}

func promptTitleWord(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lowered := []rune(strings.ToLower(trimmed))
	lowered[0] = unicode.ToUpper(lowered[0])
	return string(lowered)
}

func promptAbsInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func promptMaxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func promptMinInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
