package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

const inventoryItemSuggestionPromptTemplate = `
You are designing draft inventory items for Unclaimed Streets, an urban fantasy MMORPG.

Generate exactly %d item drafts.

Requested direction:
- theme prompt: %s
- categories to bias toward: %s
- rarity tiers to bias toward: %s
- equip slots to bias toward: %s
- stat tags to support when relevant: %s
- benefit tags to support when relevant: %s
- statuses to apply when relevant: %s
- internal tags to bias toward: %s
- item level band: %d to %d

Existing inventory items to avoid echoing:
%s

Return JSON only:
{
  "drafts": [
    {
      "category": "equippable|consumable|material|utility",
      "whyItFits": "1-2 short sentences",
      "warnings": ["optional warning"],
      "item": {
        "name": "string",
        "flavorText": "1-3 sentences of evocative item description",
        "effectText": "one short sentence describing the gameplay effect",
        "rarityTier": "Common|Uncommon|Epic|Mythic|Not Droppable",
        "itemLevel": 1,
        "buyPrice": 10,
        "unlockTier": 1,
        "equipSlot": "hat|necklace|chest|legs|shoes|gloves|ring|dominant_hand|off_hand or empty",
        "strengthMod": 0,
        "dexterityMod": 0,
        "constitutionMod": 0,
        "intelligenceMod": 0,
        "wisdomMod": 0,
        "charismaMod": 0,
        "physicalDamageBonusPercent": 0,
        "piercingDamageBonusPercent": 0,
        "slashingDamageBonusPercent": 0,
        "bludgeoningDamageBonusPercent": 0,
        "fireDamageBonusPercent": 0,
        "iceDamageBonusPercent": 0,
        "lightningDamageBonusPercent": 0,
        "poisonDamageBonusPercent": 0,
        "arcaneDamageBonusPercent": 0,
        "holyDamageBonusPercent": 0,
        "shadowDamageBonusPercent": 0,
        "physicalResistancePercent": 0,
        "piercingResistancePercent": 0,
        "slashingResistancePercent": 0,
        "bludgeoningResistancePercent": 0,
        "fireResistancePercent": 0,
        "iceResistancePercent": 0,
        "lightningResistancePercent": 0,
        "poisonResistancePercent": 0,
        "arcaneResistancePercent": 0,
        "holyResistancePercent": 0,
        "shadowResistancePercent": 0,
        "handItemCategory": "weapon|shield|orb|staff or empty",
        "handedness": "one_handed|two_handed or empty",
        "damageMin": 0,
        "damageMax": 0,
        "damageAffinity": "physical|piercing|slashing|bludgeoning|fire|ice|lightning|poison|arcane|holy|shadow or empty",
        "swipesPerAttack": 0,
        "blockPercentage": 0,
        "damageBlocked": 0,
        "spellDamageBonusPercent": 0,
        "consumeHealthDelta": 0,
        "consumeManaDelta": 0,
        "consumeRevivePartyMemberHealth": 0,
        "consumeReviveAllDownedPartyMembersHealth": 0,
        "consumeDealDamage": 0,
        "consumeDealDamageHits": 0,
        "consumeDealDamageAllEnemies": 0,
        "consumeDealDamageAllEnemiesHits": 0,
        "consumeCreateBase": false,
        "consumeStatusesToAdd": [],
        "consumeStatusesToRemove": [],
        "consumeSpellIds": [],
        "consumeTeachRecipeIds": [],
        "alchemyRecipes": [],
        "workshopRecipes": [],
        "internalTags": ["snake_case_tag"]
      }
    }
  ]
}

Rules:
- Output exactly %d drafts.
- Output JSON only. No markdown.
- Keep the tone urban fantasy, tactile, and gameable.
- Drafts should feel materially distinct from one another.
- Bias toward items that support recognizable builds, professions, exploration, or social play.
- If stat tags are requested, strongly prefer items whose actual gameplay bonuses clearly support those stats.
- If benefit tags are requested, strongly prefer items whose gameplay effects clearly match those requested benefits.
- If status names are requested, strongly prefer consumables or utility items that actually apply those statuses.
- Do not invent recipe links, taught recipe IDs, or spell IDs in this generator version. Leave consumeSpellIds, consumeTeachRecipeIds, alchemyRecipes, and workshopRecipes empty.
- Do not set unlockLocksStrength in generated drafts. Leave lock-unlocking behavior empty unless it is added manually later.
- Equippable drafts should have coherent stats for their slot and fantasy.
- Consumables should create clear, concrete gameplay effects.
- Materials should still feel desirable and specific, not generic vendor trash.
- Utility items should unlock a play pattern, traversal trick, social angle, or base-related use.
- If equipSlot is dominant_hand or off_hand, use only valid hand item combinations:
  - dominant_hand: weapon or staff
  - off_hand: shield or orb
  - staff must be two_handed
  - off_hand items must be one_handed
- Keep effectText to one line.
- Use lowercase snake_case for internalTags.
`

type inventoryItemSuggestionResponse struct {
	Drafts []inventoryItemSuggestionDraftPayload `json:"drafts"`
}

type inventoryItemSuggestionDraftPayload struct {
	Category  string               `json:"category"`
	WhyItFits string               `json:"whyItFits"`
	Warnings  []string             `json:"warnings"`
	Item      models.InventoryItem `json:"item"`
}

type GenerateInventoryItemSuggestionsProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateInventoryItemSuggestionsProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateInventoryItemSuggestionsProcessor {
	log.Println("Initializing GenerateInventoryItemSuggestionsProcessor")
	return GenerateInventoryItemSuggestionsProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateInventoryItemSuggestionsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing inventory item suggestion task: %s", task.Type())

	var payload jobs.GenerateInventoryItemSuggestionsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.InventoryItemSuggestionJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Inventory item suggestion job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.InventoryItemSuggestionJobStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.InventoryItemSuggestionJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateDrafts(ctx, job); err != nil {
		return p.failJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateInventoryItemSuggestionsProcessor) generateDrafts(
	ctx context.Context,
	job *models.InventoryItemSuggestionJob,
) error {
	existingItems, err := p.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to load inventory items: %w", err)
	}

	prompt := fmt.Sprintf(
		inventoryItemSuggestionPromptTemplate,
		maxInt(1, job.Count),
		quotedOrNone(strings.TrimSpace(job.ThemePrompt)),
		renderTagList(job.Categories),
		renderTagList(job.RarityTiers),
		renderTagList(job.EquipSlots),
		renderTagList(job.StatTags),
		renderTagList(job.BenefitTags),
		renderTagList(job.StatusNames),
		renderTagList(job.InternalTags),
		maxInt(1, job.MinItemLevel),
		maxInt(maxInt(1, job.MinItemLevel), job.MaxItemLevel),
		buildInventoryItemSuggestionAvoidance(existingItems, 80),
		maxInt(1, job.Count),
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate inventory item suggestions: %w", err)
	}

	generated := &inventoryItemSuggestionResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse inventory item suggestion payload: %w", err)
	}
	if len(generated.Drafts) == 0 {
		return fmt.Errorf("inventory item suggestion payload did not include any drafts")
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
	for _, spec := range generated.Drafts {
		draft := sanitizeInventoryItemSuggestionDraft(spec, job, existingNames, localNames)
		draft.JobID = job.ID
		draft.Status = models.InventoryItemSuggestionDraftStatusSuggested
		if err := p.dbClient.InventoryItemSuggestionDraft().Create(ctx, draft); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create inventory item suggestion draft: %w", err)
		}
		createdCount++
	}

	job.CreatedCount = createdCount
	job.Status = models.InventoryItemSuggestionJobStatusCompleted
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.InventoryItemSuggestionJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update inventory item suggestion job: %w", err)
	}

	return nil
}

func (p *GenerateInventoryItemSuggestionsProcessor) failJob(
	ctx context.Context,
	job *models.InventoryItemSuggestionJob,
	err error,
) error {
	msg := err.Error()
	job.Status = models.InventoryItemSuggestionJobStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.InventoryItemSuggestionJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark inventory item suggestion job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func buildInventoryItemSuggestionAvoidance(items []models.InventoryItem, limit int) string {
	if len(items) == 0 {
		return "- none"
	}
	sorted := append([]models.InventoryItem(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
	})
	if limit > 0 && len(sorted) > limit {
		sorted = sorted[:limit]
	}
	lines := make([]string, 0, len(sorted))
	for _, item := range sorted {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		category := inferInventoryItemSuggestionCategory(item)
		summary := strings.TrimSpace(item.EffectText)
		if summary == "" {
			summary = strings.TrimSpace(item.FlavorText)
		}
		if len(summary) > 90 {
			summary = strings.TrimSpace(summary[:90]) + "..."
		}
		lines = append(lines, fmt.Sprintf("- %s [%s, level %d, %s]: %s", name, item.RarityTier, maxInt(1, item.ItemLevel), category, summary))
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func sanitizeInventoryItemSuggestionDraft(
	spec inventoryItemSuggestionDraftPayload,
	job *models.InventoryItemSuggestionJob,
	existingNames map[string]struct{},
	localNames map[string]struct{},
) *models.InventoryItemSuggestionDraft {
	item := spec.Item
	item.ID = 0
	item.CreatedAt = time.Time{}
	item.UpdatedAt = time.Time{}
	item.Archived = false
	item.ImageURL = ""
	item.ImageGenerationStatus = models.InventoryImageGenerationStatusNone
	item.ImageGenerationError = nil
	item.UnlockLocksStrength = nil

	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		item.Name = "Unnamed Item"
	}
	item.FlavorText = strings.TrimSpace(item.FlavorText)
	item.EffectText = strings.TrimSpace(item.EffectText)
	item.RarityTier = normalizeInventoryItemSuggestionRarity(item.RarityTier, job.RarityTiers)
	item.ItemLevel = clampInventoryItemSuggestionLevel(item.ItemLevel, job.MinItemLevel, job.MaxItemLevel)
	if item.ItemLevel < 1 {
		item.ItemLevel = 1
	}
	if item.BuyPrice == nil || *item.BuyPrice < 0 {
		item.BuyPrice = intPtr(defaultInventoryItemSuggestionBuyPrice(item))
	}
	if item.UnlockTier == nil || *item.UnlockTier < 1 {
		item.UnlockTier = intPtr(item.ItemLevel)
	}

	var warnings []string
	warnings = append(warnings, normalizeInventorySuggestionWarnings(spec.Warnings)...)

	category := normalizeInventorySuggestionCategory(spec.Category)
	if category == "" {
		category = inferInventoryItemSuggestionCategory(item)
	}

	if item.EquipSlot != nil {
		trimmed := strings.TrimSpace(*item.EquipSlot)
		if trimmed == "" || !models.IsValidInventoryEquipSlot(trimmed) {
			item.EquipSlot = nil
			warnings = append(warnings, "Equip slot was invalid and has been cleared.")
		} else {
			item.EquipSlot = &trimmed
		}
	}

	if !isHandInventorySuggestion(item) {
		item.HandItemCategory = nil
		item.Handedness = nil
		item.DamageMin = nil
		item.DamageMax = nil
		item.DamageAffinity = nil
		item.SwipesPerAttack = nil
		item.BlockPercentage = nil
		item.DamageBlocked = nil
		item.SpellDamageBonusPercent = nil
	}

	if len(item.AlchemyRecipes) > 0 || len(item.WorkshopRecipes) > 0 || len(item.ConsumeSpellIDs) > 0 || len(item.ConsumeTeachRecipeIDs) > 0 {
		warnings = append(warnings, "Linked recipes and taught spell/recipe references are cleared in this generator pass.")
	}
	item.AlchemyRecipes = models.InventoryRecipes{}
	item.WorkshopRecipes = models.InventoryRecipes{}
	item.ConsumeSpellIDs = models.StringArray{}
	item.ConsumeTeachRecipeIDs = models.StringArray{}
	item.InternalTags = normalizeInventorySuggestionTags(item.InternalTags, job.InternalTags)
	item.ConsumeStatusesToRemove = normalizeInventorySuggestionTags(item.ConsumeStatusesToRemove, nil)
	if item.ConsumeStatusesToAdd == nil {
		item.ConsumeStatusesToAdd = models.ScenarioFailureStatusTemplates{}
	}

	nameKey := normalizeInventoryItemSuggestionNameKey(item.Name)
	if _, exists := existingNames[nameKey]; exists {
		warnings = append(warnings, "This draft name collides with an existing inventory item.")
	}
	if _, exists := localNames[nameKey]; exists {
		warnings = append(warnings, "This draft name collides with another generated draft in the same batch.")
	}
	if nameKey != "" {
		localNames[nameKey] = struct{}{}
	}
	warnings = append(warnings, evaluateInventorySuggestionTargeting(item, job)...)

	whyItFits := strings.TrimSpace(spec.WhyItFits)
	if whyItFits == "" {
		whyItFits = "Generated to broaden the item pool for this band and theme."
	}

	return &models.InventoryItemSuggestionDraft{
		Name:         item.Name,
		Category:     category,
		RarityTier:   item.RarityTier,
		ItemLevel:    item.ItemLevel,
		EquipSlot:    item.EquipSlot,
		WhyItFits:    whyItFits,
		InternalTags: item.InternalTags,
		Warnings:     dedupeInventorySuggestionStrings(warnings),
		Payload: models.InventoryItemSuggestionPayloadValue{
			Category:  category,
			WhyItFits: whyItFits,
			Item:      item,
		},
	}
}

func normalizeInventorySuggestionCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "equippable":
		return "equippable"
	case "consumable":
		return "consumable"
	case "material":
		return "material"
	case "utility":
		return "utility"
	default:
		return ""
	}
}

func normalizeInventoryItemSuggestionRarity(raw string, allowed []string) string {
	candidate := normalizeGeneratedInventoryRarity(raw)
	if candidate == "" {
		candidate = "Common"
	}
	if len(allowed) == 0 {
		return candidate
	}
	allowedSet := map[string]struct{}{}
	for _, item := range allowed {
		normalized := normalizeGeneratedInventoryRarity(item)
		if normalized != "" {
			allowedSet[normalized] = struct{}{}
		}
	}
	if len(allowedSet) == 0 {
		return candidate
	}
	if _, exists := allowedSet[candidate]; exists {
		return candidate
	}
	for fallback := range allowedSet {
		return fallback
	}
	return candidate
}

func normalizeGeneratedInventoryRarity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "common":
		return "Common"
	case "uncommon":
		return "Uncommon"
	case "epic":
		return "Epic"
	case "mythic":
		return "Mythic"
	case "not droppable":
		return "Not Droppable"
	default:
		return ""
	}
}

func clampInventoryItemSuggestionLevel(level int, minLevel int, maxLevel int) int {
	if minLevel < 1 {
		minLevel = 1
	}
	if maxLevel < minLevel {
		maxLevel = minLevel
	}
	if level < minLevel {
		return minLevel
	}
	if level > maxLevel {
		return maxLevel
	}
	return level
}

func defaultInventoryItemSuggestionBuyPrice(item models.InventoryItem) int {
	base := maxInt(1, item.ItemLevel) * 8
	switch strings.ToLower(strings.TrimSpace(item.RarityTier)) {
	case "uncommon":
		base = maxInt(12, item.ItemLevel*14)
	case "epic":
		base = maxInt(24, item.ItemLevel*24)
	case "mythic":
		base = maxInt(40, item.ItemLevel*38)
	case "not droppable":
		return 0
	}
	if item.EquipSlot != nil && strings.TrimSpace(*item.EquipSlot) != "" {
		base += maxInt(2, item.ItemLevel/2)
	}
	if item.ConsumeDealDamage > 0 || item.ConsumeDealDamageAllEnemies > 0 || item.ConsumeHealthDelta > 0 || item.ConsumeManaDelta > 0 {
		base += maxInt(3, item.ItemLevel/3)
	}
	return base
}

func normalizeInventorySuggestionTags(primary []string, fallback []string) models.StringArray {
	seen := map[string]struct{}{}
	tags := make(models.StringArray, 0, len(primary)+len(fallback))
	for _, raw := range append(append([]string{}, primary...), fallback...) {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

func normalizeInventorySuggestionWarnings(input []string) []string {
	warnings := make([]string, 0, len(input))
	for _, raw := range input {
		trimmed := strings.TrimSpace(raw)
		if trimmed != "" {
			warnings = append(warnings, trimmed)
		}
	}
	return warnings
}

func dedupeInventorySuggestionStrings(input []string) models.StringArray {
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

func normalizeInventoryItemSuggestionNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func inferInventoryItemSuggestionCategory(item models.InventoryItem) string {
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

func isHandInventorySuggestion(item models.InventoryItem) bool {
	return item.EquipSlot != nil && models.IsHandEquipSlot(strings.TrimSpace(*item.EquipSlot))
}

func evaluateInventorySuggestionTargeting(
	item models.InventoryItem,
	job *models.InventoryItemSuggestionJob,
) []string {
	warnings := make([]string, 0)
	if job == nil {
		return warnings
	}
	if len(job.StatTags) > 0 && !inventorySuggestionMatchesRequestedStats(item, job.StatTags) {
		warnings = append(warnings, "This draft does not strongly express the requested stat focus.")
	}
	if len(job.BenefitTags) > 0 && !inventorySuggestionMatchesRequestedBenefits(item, job.BenefitTags) {
		warnings = append(warnings, "This draft does not clearly hit the requested benefit profile.")
	}
	if len(job.StatusNames) > 0 && !inventorySuggestionMatchesRequestedStatuses(item, job.StatusNames) {
		warnings = append(warnings, "This draft does not apply any of the requested statuses.")
	}
	return warnings
}

func inventorySuggestionMatchesRequestedStats(item models.InventoryItem, tags []string) bool {
	for _, tag := range tags {
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "strength":
			if item.StrengthMod != 0 {
				return true
			}
		case "dexterity":
			if item.DexterityMod != 0 {
				return true
			}
		case "constitution":
			if item.ConstitutionMod != 0 {
				return true
			}
		case "intelligence":
			if item.IntelligenceMod != 0 {
				return true
			}
		case "wisdom":
			if item.WisdomMod != 0 {
				return true
			}
		case "charisma":
			if item.CharismaMod != 0 {
				return true
			}
		}
	}
	return false
}

func inventorySuggestionMatchesRequestedBenefits(item models.InventoryItem, tags []string) bool {
	for _, raw := range tags {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "healing", "health_restore":
			if item.ConsumeHealthDelta > 0 || hasStatusEffectType(item.ConsumeStatusesToAdd, "health_over_time") {
				return true
			}
		case "mana_restore", "mana":
			if item.ConsumeManaDelta > 0 || hasStatusEffectType(item.ConsumeStatusesToAdd, "mana_over_time") {
				return true
			}
		case "revive":
			if item.ConsumeRevivePartyMemberHealth > 0 || item.ConsumeReviveAllDownedPartyMembersHealth > 0 {
				return true
			}
		case "damage", "offense":
			if hasDirectDamageBenefit(item) {
				return true
			}
		case "aoe_damage":
			if item.ConsumeDealDamageAllEnemies > 0 {
				return true
			}
		case "status_application", "status":
			if len(item.ConsumeStatusesToAdd) > 0 {
				return true
			}
		case "cleanse":
			if len(item.ConsumeStatusesToRemove) > 0 {
				return true
			}
		case "base_creation":
			if item.ConsumeCreateBase {
				return true
			}
		case "capture":
			if item.IsCaptureType {
				return true
			}
		case "unlocking", "locks":
			if item.UnlockLocksStrength != nil && *item.UnlockLocksStrength > 0 {
				return true
			}
		case "damage_bonus":
			if hasAnyDamageBonus(item) {
				return true
			}
		case "resistance", "defense":
			if hasAnyResistance(item) || item.BlockPercentage != nil || item.DamageBlocked != nil {
				return true
			}
		case "spellcasting", "spell_damage":
			if item.SpellDamageBonusPercent != nil && *item.SpellDamageBonusPercent > 0 {
				return true
			}
		}
	}
	return false
}

func inventorySuggestionMatchesRequestedStatuses(item models.InventoryItem, names []string) bool {
	if len(item.ConsumeStatusesToAdd) == 0 {
		return false
	}
	requested := map[string]struct{}{}
	for _, raw := range names {
		key := strings.ToLower(strings.TrimSpace(raw))
		if key != "" {
			requested[key] = struct{}{}
		}
	}
	for _, status := range item.ConsumeStatusesToAdd {
		if _, exists := requested[strings.ToLower(strings.TrimSpace(status.Name))]; exists {
			return true
		}
	}
	return false
}

func hasStatusEffectType(statuses models.ScenarioFailureStatusTemplates, effectType string) bool {
	for _, status := range statuses {
		if strings.EqualFold(strings.TrimSpace(status.EffectType), effectType) {
			return true
		}
	}
	return false
}

func hasDirectDamageBenefit(item models.InventoryItem) bool {
	if item.ConsumeDealDamage > 0 || item.ConsumeDealDamageAllEnemies > 0 {
		return true
	}
	return item.DamageMin != nil && item.DamageMax != nil && *item.DamageMax > 0
}

func hasAnyDamageBonus(item models.InventoryItem) bool {
	return item.PhysicalDamageBonusPercent != 0 ||
		item.PiercingDamageBonusPercent != 0 ||
		item.SlashingDamageBonusPercent != 0 ||
		item.BludgeoningDamageBonusPercent != 0 ||
		item.FireDamageBonusPercent != 0 ||
		item.IceDamageBonusPercent != 0 ||
		item.LightningDamageBonusPercent != 0 ||
		item.PoisonDamageBonusPercent != 0 ||
		item.ArcaneDamageBonusPercent != 0 ||
		item.HolyDamageBonusPercent != 0 ||
		item.ShadowDamageBonusPercent != 0
}

func hasAnyResistance(item models.InventoryItem) bool {
	return item.PhysicalResistancePercent != 0 ||
		item.PiercingResistancePercent != 0 ||
		item.SlashingResistancePercent != 0 ||
		item.BludgeoningResistancePercent != 0 ||
		item.FireResistancePercent != 0 ||
		item.IceResistancePercent != 0 ||
		item.LightningResistancePercent != 0 ||
		item.PoisonResistancePercent != 0 ||
		item.ArcaneResistancePercent != 0 ||
		item.HolyResistancePercent != 0 ||
		item.ShadowResistancePercent != 0
}
