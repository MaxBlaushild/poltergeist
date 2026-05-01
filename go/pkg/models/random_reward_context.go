package models

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

type RandomRewardContentKind string

const (
	RandomRewardContentGeneric          RandomRewardContentKind = "generic"
	RandomRewardContentChallenge        RandomRewardContentKind = "challenge"
	RandomRewardContentPointOfInterest  RandomRewardContentKind = "point_of_interest"
	RandomRewardContentExposition       RandomRewardContentKind = "exposition"
	RandomRewardContentScenario         RandomRewardContentKind = "scenario"
	RandomRewardContentMonster          RandomRewardContentKind = "monster"
	RandomRewardContentMonsterEncounter RandomRewardContentKind = "monster_encounter"
	RandomRewardContentTreasureChest    RandomRewardContentKind = "treasure_chest"
	RandomRewardContentQuestTurnIn      RandomRewardContentKind = "quest_turn_in"
)

type RandomRewardContext struct {
	ContentKind             RandomRewardContentKind
	ZoneKind                string
	GenreName               string
	PointOfInterestCategory PointOfInterestMarkerCategory
	SubmissionType          QuestNodeSubmissionType
	PrimaryStatTag          string
	StatTags                []string
	Proficiencies           []string
	InternalTags            []string
	ResourceTypeIDs         []uuid.UUID
	ElementalTags           []string
}

func (c *RandomRewardContext) PreferredRewardTags() []string {
	if c == nil {
		return []string{}
	}

	tags := make([]string, 0, 24)
	seen := map[string]struct{}{}
	addTag := func(raw string) {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			return
		}
		if _, exists := seen[tag]; exists {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	addMany := func(values ...string) {
		for _, value := range values {
			addTag(value)
		}
	}

	for _, rawTag := range c.InternalTags {
		addTag(rawTag)
	}
	for _, elementalTag := range c.ElementalTags {
		addTag(elementalTag)
	}

	switch c.ContentKind {
	case RandomRewardContentChallenge:
		switch c.SubmissionType {
		case QuestNodeSubmissionTypeText:
			addMany("social", "scholar", "guide")
		case QuestNodeSubmissionTypeVideo:
			addMany("scout", "guide", "street", "wild")
		default:
			addMany("scout", "guide", "street")
		}
	case RandomRewardContentPointOfInterest:
		switch c.PointOfInterestCategory {
		case PointOfInterestMarkerCategoryArchive, PointOfInterestMarkerCategoryMuseum, PointOfInterestMarkerCategoryLandmark, PointOfInterestMarkerCategoryCivic:
			addMany("guide", "scholar", "relic", "arcane")
		case PointOfInterestMarkerCategoryMarket:
			addMany("broker", "social", "guide", "street")
		case PointOfInterestMarkerCategoryPark, PointOfInterestMarkerCategoryWaterfront:
			addMany("nature", "wild", "scout")
		case PointOfInterestMarkerCategoryCoffeehouse, PointOfInterestMarkerCategoryTavern, PointOfInterestMarkerCategoryEatery, PointOfInterestMarkerCategoryTheater:
			addMany("social", "guide", "court")
		case PointOfInterestMarkerCategoryArena:
			addMany("martial", "hunter", "frontline")
		default:
			addMany("guide", "street", "scout")
		}
	case RandomRewardContentExposition:
		addMany("scholar", "guide", "arcane", "ritual", "relic")
	case RandomRewardContentScenario:
		for _, proficiencyTag := range randomRewardTagsFromProficiencies(c.Proficiencies) {
			addTag(proficiencyTag)
		}
		if strings.Contains(strings.ToLower(strings.Join(c.InternalTags, ",")), "main_story") {
			addMany("guide", "relic")
		}
	case RandomRewardContentMonster, RandomRewardContentMonsterEncounter:
		addMany("martial", "hunter", "frontline", "defender")
	case RandomRewardContentQuestTurnIn:
		if strings.Contains(strings.ToLower(strings.Join(c.InternalTags, ",")), "main_story") {
			addMany("guide", "relic")
		}
		for _, proficiencyTag := range randomRewardTagsFromProficiencies(c.Proficiencies) {
			addTag(proficiencyTag)
		}
	}

	for _, statTag := range c.normalizedStatTags() {
		switch statTag {
		case "strength", "constitution":
			addMany("martial", "frontline", "defender")
		case "dexterity":
			addMany("scout", "rogue", "skirmisher", "street")
		case "intelligence", "wisdom":
			addMany("scholar", "arcane", "seer", "ritual")
		case "charisma":
			addMany("social", "leader", "envoy", "broker", "court")
		}
	}

	for _, word := range randomRewardNormalizedWords(c.GenreName) {
		addTag(word)
	}

	return tags
}

func (c *RandomRewardContext) HasSignals() bool {
	if c == nil {
		return false
	}
	if NormalizeZoneKind(c.ZoneKind) != "" {
		return true
	}
	if len(c.ResourceTypeIDs) > 0 {
		return true
	}
	return len(c.PreferredRewardTags()) > 0
}

func (c *RandomRewardContext) normalizedStatTags() []string {
	if c == nil {
		return []string{}
	}

	tags := make([]string, 0, len(c.StatTags)+1)
	seen := map[string]struct{}{}
	add := func(raw string) {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			return
		}
		if _, exists := seen[tag]; exists {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	add(c.PrimaryStatTag)
	for _, statTag := range c.StatTags {
		add(statTag)
	}

	return tags
}

func filterRewardItemsForContext(
	items []InventoryItem,
	level int,
	equippable bool,
	rewardContext *RandomRewardContext,
) []InventoryItem {
	filtered := filterRewardItems(items, level, equippable)
	if rewardContext == nil || len(filtered) == 0 || !rewardContext.HasSignals() {
		return filtered
	}

	preferredTags := rewardContext.PreferredRewardTags()
	bestScore := 0
	scored := make([]scoredRewardItem, 0, len(filtered))
	for _, item := range filtered {
		score := scoreRewardItemForContext(item, equippable, rewardContext, preferredTags)
		scored = append(scored, scoredRewardItem{
			Item:  item,
			Score: score,
		})
		if score > bestScore {
			bestScore = score
		}
	}
	if bestScore <= 0 {
		return filtered
	}

	threshold := bestScore - 14
	if threshold < 1 {
		threshold = 1
	}

	selected := make([]scoredRewardItem, 0, len(scored))
	for _, candidate := range scored {
		if candidate.Score < threshold {
			continue
		}
		selected = append(selected, candidate)
	}
	if len(selected) == 0 {
		return filtered
	}

	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Score == selected[j].Score {
			return selected[i].Item.ID < selected[j].Item.ID
		}
		return selected[i].Score > selected[j].Score
	})

	out := make([]InventoryItem, 0, len(selected))
	for _, candidate := range selected {
		out = append(out, candidate.Item)
	}
	return out
}

type scoredRewardItem struct {
	Item  InventoryItem
	Score int
}

func scoreRewardItemForContext(
	item InventoryItem,
	equippable bool,
	rewardContext *RandomRewardContext,
	preferredTags []string,
) int {
	if rewardContext == nil {
		return 0
	}

	score := 0
	if zoneKind := NormalizeZoneKind(rewardContext.ZoneKind); zoneKind != "" && NormalizeZoneKind(item.ZoneKind) == zoneKind {
		score += 30
	}

	if len(rewardContext.ResourceTypeIDs) > 0 && item.ResourceTypeID != nil {
		for _, resourceTypeID := range rewardContext.ResourceTypeIDs {
			if resourceTypeID != uuid.Nil && *item.ResourceTypeID == resourceTypeID {
				score += 45
				break
			}
		}
	}

	matchingTagCount := rewardItemMatchingTagCount(item, preferredTags)
	if matchingTagCount > 0 {
		score += min(28, matchingTagCount*7)
	}

	if rewardItemProvidesKnowledgeProgression(item) && rewardContextPrefersKnowledgeRewards(rewardContext) {
		score += 10
	}
	if rewardItemProvidesUtility(item) && rewardContextPrefersUtilityRewards(rewardContext) {
		score += 8
	}
	if equippable && rewardContextPrefersCombatEquipment(rewardContext) {
		score += 8
	}
	if !equippable && rewardContextPrefersNonEquipment(rewardContext) {
		score += 8
	}

	if item.DamageAffinity != nil {
		damageAffinity := strings.ToLower(strings.TrimSpace(*item.DamageAffinity))
		for _, preferredTag := range preferredTags {
			if preferredTag == damageAffinity {
				score += 10
				break
			}
		}
	}

	return score
}

func rewardItemMatchingTagCount(item InventoryItem, preferredTags []string) int {
	if len(preferredTags) == 0 || len(item.InternalTags) == 0 {
		return 0
	}

	itemTags := map[string]struct{}{}
	for _, rawTag := range item.InternalTags {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		itemTags[tag] = struct{}{}
	}

	matches := 0
	for _, preferredTag := range preferredTags {
		if _, exists := itemTags[preferredTag]; exists {
			matches++
		}
	}
	return matches
}

func rewardItemProvidesUtility(item InventoryItem) bool {
	return item.ConsumeHealthDelta != 0 ||
		item.ConsumeManaDelta != 0 ||
		item.ConsumeRevivePartyMemberHealth != 0 ||
		item.ConsumeReviveAllDownedPartyMembersHealth != 0 ||
		item.ConsumeCreateBase ||
		len(item.ConsumeStatusesToAdd) > 0 ||
		len(item.ConsumeStatusesToRemove) > 0 ||
		len(item.ConsumeSpellIDs) > 0 ||
		len(item.ConsumeTeachRecipeIDs) > 0
}

func rewardItemProvidesKnowledgeProgression(item InventoryItem) bool {
	return len(item.ConsumeSpellIDs) > 0 ||
		len(item.ConsumeTeachRecipeIDs) > 0 ||
		len(item.AlchemyRecipes) > 0 ||
		len(item.WorkshopRecipes) > 0
}

func rewardContextPrefersKnowledgeRewards(rewardContext *RandomRewardContext) bool {
	if rewardContext == nil {
		return false
	}

	for _, tag := range rewardContext.PreferredRewardTags() {
		switch tag {
		case "scholar", "arcane", "seer", "ritual", "relic", "guide":
			return true
		}
	}
	return false
}

func rewardContextPrefersUtilityRewards(rewardContext *RandomRewardContext) bool {
	if rewardContext == nil {
		return false
	}

	switch rewardContext.ContentKind {
	case RandomRewardContentPointOfInterest, RandomRewardContentExposition:
		return true
	}

	for _, tag := range rewardContext.PreferredRewardTags() {
		switch tag {
		case "social", "scholar", "guide", "scout", "street", "nature", "wild", "ritual":
			return true
		}
	}
	return false
}

func rewardContextPrefersCombatEquipment(rewardContext *RandomRewardContext) bool {
	if rewardContext == nil {
		return false
	}

	switch rewardContext.ContentKind {
	case RandomRewardContentMonster, RandomRewardContentMonsterEncounter:
		return true
	}

	for _, tag := range rewardContext.PreferredRewardTags() {
		switch tag {
		case "martial", "frontline", "defender", "hunter", "rogue", "skirmisher":
			return true
		}
	}
	return false
}

func rewardContextPrefersNonEquipment(rewardContext *RandomRewardContext) bool {
	if rewardContext == nil {
		return false
	}

	switch rewardContext.ContentKind {
	case RandomRewardContentPointOfInterest, RandomRewardContentExposition:
		return true
	}

	for _, tag := range rewardContext.PreferredRewardTags() {
		switch tag {
		case "social", "scholar", "guide", "ritual", "relic", "seer":
			return true
		}
	}
	return false
}

func randomRewardTagsFromProficiencies(proficiencies []string) []string {
	tags := make([]string, 0, len(proficiencies)*2)
	seen := map[string]struct{}{}
	add := func(raw string) {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			return
		}
		if _, exists := seen[tag]; exists {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	for _, proficiency := range proficiencies {
		key := strings.ToLower(strings.TrimSpace(proficiency))
		switch {
		case strings.Contains(key, "arc"), strings.Contains(key, "magic"), strings.Contains(key, "spell"), strings.Contains(key, "ritual"), strings.Contains(key, "lore"), strings.Contains(key, "scholar"):
			add("arcane")
			add("scholar")
		case strings.Contains(key, "stealth"), strings.Contains(key, "lock"), strings.Contains(key, "rogue"), strings.Contains(key, "scout"), strings.Contains(key, "sneak"):
			add("rogue")
			add("scout")
		case strings.Contains(key, "social"), strings.Contains(key, "persua"), strings.Contains(key, "decept"), strings.Contains(key, "perform"), strings.Contains(key, "diplom"):
			add("social")
			add("court")
		case strings.Contains(key, "surviv"), strings.Contains(key, "nature"), strings.Contains(key, "hunt"), strings.Contains(key, "track"):
			add("wild")
			add("nature")
		case strings.Contains(key, "fight"), strings.Contains(key, "guard"), strings.Contains(key, "blade"), strings.Contains(key, "melee"), strings.Contains(key, "brawl"):
			add("martial")
			add("frontline")
		}
	}

	return tags
}

func randomRewardNormalizedWords(value string) []string {
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(value)), func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '/' || r == ','
	})
	if len(parts) == 0 {
		return []string{}
	}

	ignored := map[string]struct{}{
		"":        {},
		"and":     {},
		"the":     {},
		"of":      {},
		"fantasy": {},
		"urban":   {},
	}
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		if _, skip := ignored[part]; skip {
			continue
		}
		if _, exists := seen[part]; exists {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}
