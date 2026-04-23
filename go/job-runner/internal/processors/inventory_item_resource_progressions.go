package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type inventoryResourceProgressionTarget struct {
	Label      string
	Level      int
	RarityTier string
}

var inventoryResourceProgressionTargets = []inventoryResourceProgressionTarget{
	{Label: "Starter", Level: 1, RarityTier: "Common"},
	{Label: "Apprentice", Level: 12, RarityTier: "Common"},
	{Label: "Journeyman", Level: 23, RarityTier: "Common"},
	{Label: "Seasoned", Level: 34, RarityTier: "Common"},
	{Label: "Hardened", Level: 45, RarityTier: "Uncommon"},
	{Label: "Veteran", Level: 56, RarityTier: "Uncommon"},
	{Label: "Expert", Level: 67, RarityTier: "Uncommon"},
	{Label: "Master", Level: 78, RarityTier: "Epic"},
	{Label: "Mythic", Level: 89, RarityTier: "Epic"},
	{Label: "Apex", Level: 100, RarityTier: "Mythic"},
}

const inventoryResourceProgressionPromptTemplate = `
You are designing a 10-step progression of gatherable crafting resources for Unclaimed Streets, an urban fantasy MMORPG.
%s
%s
%s

Generate exactly %d draft materials. Each draft represents one rung in a single resource progression ladder for the same resource type and zone kind.

Target progression ladder:
%s

Requested creative direction:
- optional theme prompt: %s

Existing inventory items to avoid echoing:
%s

Return JSON only:
{
  "drafts": [
    {
      "category": "material",
      "whyItFits": "1-2 short sentences",
      "warnings": ["optional warning"],
      "item": {
        "name": "string",
        "flavorText": "1-3 sentences of evocative material description",
        "effectText": "one short sentence describing how crafters or gatherers use the material",
        "zoneKind": "exact requested zone kind slug",
        "rarityTier": "Common|Uncommon|Epic|Mythic",
        "itemLevel": 1,
        "buyPrice": 10,
        "unlockTier": 1,
        "internalTags": ["snake_case_tag"]
      }
    }
  ]
}

Rules:
- Output exactly %d drafts in the same order as the target progression ladder.
- Every draft must be a material item.
- Every draft must belong to the requested resource type and requested zone kind.
- The zone kind should visibly brand the materials' names, textures, geology, ecology, residue, provenance, and implied uses.
- The ten materials should feel like a coherent family with escalating scarcity, refinement, potency, and crafting value.
- Use the exact requested zoneKind slug for every draft.
- Match the listed target level and rarity for each rung.
- Keep itemLevel and unlockTier aligned to the target rung.
- Keep effectText to one gameplay-facing sentence about the material's crafting or gathering role.
- Do not add equip slots, combat stats, recipes, taught spells, consume effects, or base-creation behavior.
- Use lowercase snake_case for internalTags.
- Output JSON only. No markdown.
`

func (p *GenerateInventoryItemSuggestionsProcessor) generateResourceProgressionDrafts(
	ctx context.Context,
	job *models.InventoryItemSuggestionJob,
) error {
	if job == nil {
		return fmt.Errorf("inventory item suggestion job is required")
	}
	if job.ResourceTypeID == nil {
		return fmt.Errorf("resource progression jobs require resourceTypeId")
	}

	existingItems, err := p.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to load inventory items: %w", err)
	}
	zoneKinds, err := p.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load zone kinds: %w", err)
	}
	fixedZoneKind, err := loadOptionalZoneKind(ctx, p.dbClient, job.ZoneKind)
	if err != nil {
		return fmt.Errorf("failed to resolve resource progression zone kind: %w", err)
	}
	if fixedZoneKind == nil || models.NormalizeZoneKind(fixedZoneKind.Slug) == "" {
		return fmt.Errorf("resource progression jobs require a valid zone kind")
	}

	resourceType := job.ResourceType
	if resourceType == nil {
		resourceType, err = p.dbClient.ResourceType().FindByID(ctx, *job.ResourceTypeID)
		if err != nil {
			return fmt.Errorf("failed to load resource type: %w", err)
		}
	}
	if resourceType == nil {
		return fmt.Errorf("resource progression resource type could not be loaded")
	}

	prompt := fmt.Sprintf(
		inventoryResourceProgressionPromptTemplate,
		inventoryItemSuggestionGenreInstructionBlock(job.Genre),
		zoneKindInstructionBlock(fixedZoneKind),
		inventoryResourceTypeInstructionBlock(resourceType, fixedZoneKind),
		len(inventoryResourceProgressionTargets),
		renderInventoryResourceProgressionTargets(inventoryResourceProgressionTargets),
		quotedOrNone(strings.TrimSpace(job.ThemePrompt)),
		buildInventoryItemSuggestionAvoidance(existingItems, 80),
		len(inventoryResourceProgressionTargets),
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate inventory resource progression: %w", err)
	}

	generated := &inventoryItemSuggestionResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse inventory resource progression payload: %w", err)
	}
	if len(generated.Drafts) < len(inventoryResourceProgressionTargets) {
		return fmt.Errorf(
			"inventory resource progression payload returned %d drafts, expected %d",
			len(generated.Drafts),
			len(inventoryResourceProgressionTargets),
		)
	}

	existingNames := map[string]struct{}{}
	for _, item := range existingItems {
		nameKey := normalizeInventoryItemSuggestionNameKey(item.Name)
		if nameKey != "" {
			existingNames[nameKey] = struct{}{}
		}
	}
	localNames := map[string]struct{}{}
	createdCount := 0
	for index, target := range inventoryResourceProgressionTargets {
		draft := sanitizeInventoryItemSuggestionDraft(
			generated.Drafts[index],
			job,
			zoneKinds,
			existingNames,
			localNames,
		)
		enforceInventoryResourceProgressionDraft(draft, job, resourceType, target)
		draft.JobID = job.ID
		draft.Status = models.InventoryItemSuggestionDraftStatusSuggested
		if err := p.dbClient.InventoryItemSuggestionDraft().Create(ctx, draft); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create inventory resource progression draft: %w", err)
		}
		createdCount++
	}

	job.CreatedCount = createdCount
	job.Status = models.InventoryItemSuggestionJobStatusCompleted
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.InventoryItemSuggestionJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update inventory resource progression job: %w", err)
	}

	return nil
}

func inventoryResourceTypeInstructionBlock(
	resourceType *models.ResourceType,
	zoneKind *models.ZoneKind,
) string {
	resourceName := inventoryResourceTypeLabel(resourceType)
	resourceSlug := ""
	resourceDescription := ""
	if resourceType != nil {
		resourceSlug = strings.TrimSpace(resourceType.Slug)
		resourceDescription = strings.TrimSpace(resourceType.Description)
	}
	zoneLabel := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	if zoneLabel == "" {
		zoneLabel = strings.TrimSpace(models.ZoneKindPromptSlug(zoneKind))
	}

	lines := []string{
		"Resource progression direction:",
		fmt.Sprintf("- resource type: %s", resourceName),
	}
	if resourceSlug != "" {
		lines = append(lines, fmt.Sprintf("- resource type slug: %s", resourceSlug))
	}
	if resourceDescription != "" {
		lines = append(lines, fmt.Sprintf("- resource type description: %s", resourceDescription))
	}
	if zoneLabel != "" {
		lines = append(lines, fmt.Sprintf("- brand every material so it feels native to %s zones", strings.ToLower(zoneLabel)))
	}
	lines = append(
		lines,
		"- Think like a gatherer and crafter progression, not like combat gear.",
		"- Each rung should feel more valuable, more specialized, or harder to source than the previous rung.",
	)
	return strings.Join(lines, "\n")
}

func renderInventoryResourceProgressionTargets(
	targets []inventoryResourceProgressionTarget,
) string {
	lines := make([]string, 0, len(targets))
	for index, target := range targets {
		lines = append(
			lines,
			fmt.Sprintf(
				"%d. %s rung: level %d, rarity %s",
				index+1,
				target.Label,
				target.Level,
				target.RarityTier,
			),
		)
	}
	return strings.Join(lines, "\n")
}

func enforceInventoryResourceProgressionDraft(
	draft *models.InventoryItemSuggestionDraft,
	job *models.InventoryItemSuggestionJob,
	resourceType *models.ResourceType,
	target inventoryResourceProgressionTarget,
) {
	if draft == nil || job == nil {
		return
	}

	item := draft.Payload.Item
	warnings := append([]string{}, draft.Warnings...)
	normalizedZoneKind := models.NormalizeZoneKind(job.ZoneKind)

	item.ItemLevel = target.Level
	item.UnlockTier = intPtr(target.Level)
	item.RarityTier = target.RarityTier
	item.ZoneKind = normalizedZoneKind
	item.ResourceType = resourceType
	if job.ResourceTypeID != nil {
		resourceTypeID := *job.ResourceTypeID
		item.ResourceTypeID = &resourceTypeID
	}

	if item.BuyPrice == nil || *item.BuyPrice < 0 {
		item.BuyPrice = intPtr(defaultInventoryItemSuggestionBuyPrice(item))
	}

	if draft.Category != "material" || item.EquipSlot != nil || item.IsCaptureType || item.UnlockLocksStrength != nil {
		warnings = append(warnings, "Resource progression drafts are normalized to plain material items.")
	}

	draft.Category = "material"
	draft.ItemLevel = target.Level
	draft.RarityTier = target.RarityTier
	draft.EquipSlot = nil

	item.IsCaptureType = false
	item.EquipSlot = nil
	item.UnlockLocksStrength = nil
	item.StrengthMod = 0
	item.DexterityMod = 0
	item.ConstitutionMod = 0
	item.IntelligenceMod = 0
	item.WisdomMod = 0
	item.CharismaMod = 0
	item.PhysicalDamageBonusPercent = 0
	item.PiercingDamageBonusPercent = 0
	item.SlashingDamageBonusPercent = 0
	item.BludgeoningDamageBonusPercent = 0
	item.FireDamageBonusPercent = 0
	item.IceDamageBonusPercent = 0
	item.LightningDamageBonusPercent = 0
	item.PoisonDamageBonusPercent = 0
	item.ArcaneDamageBonusPercent = 0
	item.HolyDamageBonusPercent = 0
	item.ShadowDamageBonusPercent = 0
	item.PhysicalResistancePercent = 0
	item.PiercingResistancePercent = 0
	item.SlashingResistancePercent = 0
	item.BludgeoningResistancePercent = 0
	item.FireResistancePercent = 0
	item.IceResistancePercent = 0
	item.LightningResistancePercent = 0
	item.PoisonResistancePercent = 0
	item.ArcaneResistancePercent = 0
	item.HolyResistancePercent = 0
	item.ShadowResistancePercent = 0
	item.HandItemCategory = nil
	item.Handedness = nil
	item.DamageMin = nil
	item.DamageMax = nil
	item.DamageAffinity = nil
	item.SwipesPerAttack = nil
	item.BlockPercentage = nil
	item.DamageBlocked = nil
	item.SpellDamageBonusPercent = nil
	item.ConsumeHealthDelta = 0
	item.ConsumeManaDelta = 0
	item.ConsumeRevivePartyMemberHealth = 0
	item.ConsumeReviveAllDownedPartyMembersHealth = 0
	item.ConsumeDealDamage = 0
	item.ConsumeDealDamageHits = 0
	item.ConsumeDealDamageAllEnemies = 0
	item.ConsumeDealDamageAllEnemiesHits = 0
	item.ConsumeCreateBase = false
	item.ConsumeStatusesToAdd = models.ScenarioFailureStatusTemplates{}
	item.ConsumeStatusesToRemove = models.StringArray{}
	item.ConsumeSpellIDs = models.StringArray{}
	item.ConsumeTeachRecipeIDs = models.StringArray{}
	item.AlchemyRecipes = models.InventoryRecipes{}
	item.WorkshopRecipes = models.InventoryRecipes{}
	item.InternalTags = normalizeInventorySuggestionTags(
		append(
			append([]string{}, item.InternalTags...),
			strings.ToLower(strings.ReplaceAll(target.Label, " ", "_")),
		),
		[]string{
			"resource_progression",
			normalizeInventorySuggestionTagToken(inventoryResourceTypeLabel(resourceType)),
			normalizeInventorySuggestionTagToken(normalizedZoneKind),
		},
	)

	if strings.TrimSpace(item.EffectText) == "" {
		item.EffectText = fmt.Sprintf(
			"Used in %s crafting recipes tuned for level %d progression.",
			strings.ToLower(inventoryResourceTypeLabel(resourceType)),
			target.Level,
		)
	}
	if strings.TrimSpace(item.FlavorText) == "" {
		item.FlavorText = fmt.Sprintf(
			"A %s-tier %s material branded by %s zones.",
			strings.ToLower(target.Label),
			strings.ToLower(inventoryResourceTypeLabel(resourceType)),
			strings.ToLower(strings.ReplaceAll(normalizedZoneKind, "-", " ")),
		)
	}
	if strings.TrimSpace(draft.WhyItFits) == "" {
		draft.WhyItFits = fmt.Sprintf(
			"Fits the %s rung of this %s progression while staying visually loyal to %s zones.",
			strings.ToLower(target.Label),
			strings.ToLower(inventoryResourceTypeLabel(resourceType)),
			strings.ToLower(strings.ReplaceAll(normalizedZoneKind, "-", " ")),
		)
	}

	draft.InternalTags = item.InternalTags
	draft.Warnings = dedupeInventorySuggestionStrings(warnings)
	draft.Payload.Category = draft.Category
	draft.Payload.WhyItFits = draft.WhyItFits
	draft.Payload.Item = item
}

func inventoryResourceTypeLabel(resourceType *models.ResourceType) string {
	if resourceType == nil {
		return "Resource"
	}
	name := strings.TrimSpace(resourceType.Name)
	if name != "" {
		return name
	}
	slug := strings.TrimSpace(resourceType.Slug)
	if slug == "" {
		return "Resource"
	}
	return strings.ReplaceAll(slug, "_", " ")
}

func normalizeInventorySuggestionTagToken(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}
