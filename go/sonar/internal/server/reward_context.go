package server

import (
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func buildRandomRewardContextForPointOfInterest(
	pointOfInterest *models.PointOfInterest,
) *models.RandomRewardContext {
	if pointOfInterest == nil {
		return nil
	}

	return &models.RandomRewardContext{
		ContentKind:             models.RandomRewardContentPointOfInterest,
		ZoneKind:                pointOfInterest.ZoneKind,
		GenreName:               rewardContextGenreName(pointOfInterest.Genre),
		PointOfInterestCategory: models.NormalizePointOfInterestMarkerCategory(string(pointOfInterest.MarkerCategory)),
		InternalTags:            rewardContextTagsFromPointOfInterest(pointOfInterest),
	}
}

func buildRandomRewardContextForZone(
	zone *models.Zone,
) *models.RandomRewardContext {
	if zone == nil {
		return nil
	}
	return &models.RandomRewardContext{
		ContentKind:  models.RandomRewardContentPointOfInterest,
		ZoneKind:     zone.Kind,
		InternalTags: rewardContextNormalizeUniqueStrings([]string(zone.InternalTags)),
	}
}

func buildRandomRewardContextForExposition(
	exposition *models.Exposition,
) *models.RandomRewardContext {
	if exposition == nil {
		return nil
	}

	context := &models.RandomRewardContext{
		ContentKind:  models.RandomRewardContentExposition,
		ZoneKind:     exposition.ZoneKind,
		InternalTags: []string{"lore", "guide"},
	}
	if exposition.PointOfInterest != nil {
		context.GenreName = rewardContextGenreName(exposition.PointOfInterest.Genre)
		context.PointOfInterestCategory = models.NormalizePointOfInterestMarkerCategory(
			string(exposition.PointOfInterest.MarkerCategory),
		)
		context.InternalTags = append(context.InternalTags, rewardContextTagsFromPointOfInterest(exposition.PointOfInterest)...)
	}
	context.InternalTags = rewardContextNormalizeUniqueStrings(context.InternalTags)
	return context
}

func buildRandomRewardContextForChallenge(
	challenge *models.Challenge,
) *models.RandomRewardContext {
	if challenge == nil {
		return nil
	}

	context := &models.RandomRewardContext{
		ContentKind:    models.RandomRewardContentChallenge,
		ZoneKind:       challenge.ZoneKind,
		SubmissionType: challenge.SubmissionType,
		StatTags:       rewardContextNormalizeUniqueStrings([]string(challenge.StatTags)),
	}
	if challenge.PointOfInterest != nil {
		context.PointOfInterestCategory = models.NormalizePointOfInterestMarkerCategory(
			string(challenge.PointOfInterest.MarkerCategory),
		)
		context.InternalTags = rewardContextTagsFromPointOfInterest(challenge.PointOfInterest)
		context.GenreName = rewardContextGenreName(challenge.PointOfInterest.Genre)
	}
	if challenge.Proficiency != nil {
		context.Proficiencies = append(context.Proficiencies, strings.TrimSpace(*challenge.Proficiency))
	}
	context.Proficiencies = rewardContextNormalizeUniqueStrings(context.Proficiencies)
	return context
}

func buildRandomRewardContextForScenario(
	scenario *models.Scenario,
	selectedOption *models.ScenarioOption,
	statTag string,
	proficiencies []string,
) *models.RandomRewardContext {
	if scenario == nil {
		return nil
	}

	context := &models.RandomRewardContext{
		ContentKind:    models.RandomRewardContentScenario,
		ZoneKind:       scenario.ZoneKind,
		GenreName:      rewardContextGenreName(scenario.Genre),
		PrimaryStatTag: strings.TrimSpace(statTag),
		Proficiencies:  rewardContextNormalizeUniqueStrings(proficiencies),
		InternalTags:   rewardContextNormalizeUniqueStrings([]string(scenario.InternalTags)),
	}
	if scenario.PointOfInterest != nil {
		context.PointOfInterestCategory = models.NormalizePointOfInterestMarkerCategory(
			string(scenario.PointOfInterest.MarkerCategory),
		)
		context.InternalTags = rewardContextNormalizeUniqueStrings(
			append(context.InternalTags, rewardContextTagsFromPointOfInterest(scenario.PointOfInterest)...),
		)
	}
	if selectedOption != nil {
		context.StatTags = rewardContextNormalizeUniqueStrings(
			append(context.StatTags, strings.TrimSpace(selectedOption.StatTag)),
		)
		context.Proficiencies = rewardContextNormalizeUniqueStrings(
			append(context.Proficiencies, []string(selectedOption.Proficiencies)...),
		)
	}
	return context
}

func buildRandomRewardContextForMonster(
	monster *models.Monster,
) *models.RandomRewardContext {
	if monster == nil {
		return nil
	}

	return &models.RandomRewardContext{
		ContentKind:   models.RandomRewardContentMonster,
		ZoneKind:      monster.ZoneKind,
		GenreName:     rewardContextGenreName(monster.Genre),
		ElementalTags: rewardElementalTagsForMonster(monster),
	}
}

func buildRandomRewardContextForMonsterEncounter(
	encounter *models.MonsterEncounter,
) *models.RandomRewardContext {
	if encounter == nil {
		return nil
	}

	context := &models.RandomRewardContext{
		ContentKind:   models.RandomRewardContentMonsterEncounter,
		ZoneKind:      encounter.ZoneKind,
		ElementalTags: rewardElementalTagsForEncounter(encounter),
	}
	if encounter.PointOfInterest != nil {
		context.PointOfInterestCategory = models.NormalizePointOfInterestMarkerCategory(
			string(encounter.PointOfInterest.MarkerCategory),
		)
		context.GenreName = rewardContextGenreName(encounter.PointOfInterest.Genre)
		context.InternalTags = rewardContextTagsFromPointOfInterest(encounter.PointOfInterest)
	}
	return context
}

func buildRandomRewardContextForTreasureChest(
	treasureChest *models.TreasureChest,
) *models.RandomRewardContext {
	if treasureChest == nil {
		return nil
	}
	return &models.RandomRewardContext{
		ContentKind: models.RandomRewardContentTreasureChest,
		ZoneKind:    treasureChest.ZoneKind,
	}
}

func buildRandomRewardContextForQuest(
	quest *models.Quest,
) *models.RandomRewardContext {
	if quest == nil {
		return nil
	}

	context := &models.RandomRewardContext{
		ContentKind: models.RandomRewardContentQuestTurnIn,
		ZoneKind:    quest.ZoneKind,
	}
	if models.IsMainStoryQuestCategory(quest.Category) {
		context.InternalTags = append(context.InternalTags, "main_story")
	}
	for _, node := range quest.Nodes {
		if node.Challenge != nil {
			context.SubmissionType = node.Challenge.SubmissionType
			context.StatTags = append(context.StatTags, []string(node.Challenge.StatTags)...)
			if node.Challenge.Proficiency != nil {
				context.Proficiencies = append(context.Proficiencies, strings.TrimSpace(*node.Challenge.Proficiency))
			}
		}
		if node.Exposition != nil {
			context.InternalTags = append(context.InternalTags, "scholar", "arcane", "relic")
			context.InternalTags = append(
				context.InternalTags,
				rewardContextTagsFromPointOfInterest(node.Exposition.PointOfInterest)...,
			)
		}
		if node.Scenario != nil {
			context.InternalTags = append(context.InternalTags, []string(node.Scenario.InternalTags)...)
			if node.Scenario.PointOfInterest != nil {
				context.InternalTags = append(
					context.InternalTags,
					rewardContextTagsFromPointOfInterest(node.Scenario.PointOfInterest)...,
				)
			}
		}
		if node.Monster != nil {
			context.InternalTags = append(context.InternalTags, "martial", "hunter")
			context.ElementalTags = append(context.ElementalTags, rewardElementalTagsForMonster(node.Monster)...)
		}
		if node.MonsterEncounter != nil {
			context.InternalTags = append(context.InternalTags, "martial", "hunter")
			context.ElementalTags = append(context.ElementalTags, rewardElementalTagsForEncounter(node.MonsterEncounter)...)
		}
		if node.IsFetchQuestNode() {
			context.InternalTags = append(context.InternalTags, "wild", "guide", "gathering")
		}
	}

	context.StatTags = rewardContextNormalizeUniqueStrings(context.StatTags)
	context.Proficiencies = rewardContextNormalizeUniqueStrings(context.Proficiencies)
	context.InternalTags = rewardContextNormalizeUniqueStrings(context.InternalTags)
	context.ElementalTags = rewardContextNormalizeUniqueStrings(context.ElementalTags)
	return context
}

func rewardContextTagsFromPointOfInterest(
	pointOfInterest *models.PointOfInterest,
) []string {
	if pointOfInterest == nil {
		return []string{}
	}

	tags := make([]string, 0, len(pointOfInterest.Tags)+4)
	switch models.NormalizePointOfInterestMarkerCategory(string(pointOfInterest.MarkerCategory)) {
	case models.PointOfInterestMarkerCategoryArchive, models.PointOfInterestMarkerCategoryMuseum:
		tags = append(tags, "scholar", "arcane", "guide")
	case models.PointOfInterestMarkerCategoryLandmark, models.PointOfInterestMarkerCategoryCivic:
		tags = append(tags, "relic", "guide")
	case models.PointOfInterestMarkerCategoryMarket:
		tags = append(tags, "social", "broker", "street")
	case models.PointOfInterestMarkerCategoryPark, models.PointOfInterestMarkerCategoryWaterfront:
		tags = append(tags, "nature", "wild", "scout")
	case models.PointOfInterestMarkerCategoryArena:
		tags = append(tags, "martial", "hunter")
	case models.PointOfInterestMarkerCategoryCoffeehouse, models.PointOfInterestMarkerCategoryTavern, models.PointOfInterestMarkerCategoryEatery, models.PointOfInterestMarkerCategoryTheater:
		tags = append(tags, "social", "guide", "court")
	}
	for _, tag := range pointOfInterest.Tags {
		tags = append(tags, strings.TrimSpace(tag.Value))
	}
	return rewardContextNormalizeUniqueStrings(tags)
}

func rewardContextGenreName(genre *models.ZoneGenre) string {
	if genre == nil {
		return ""
	}
	return strings.TrimSpace(genre.Name)
}

func rewardContextNormalizeUniqueStrings(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func rewardElementalTagsForMonster(monster *models.Monster) []string {
	if monster == nil || monster.Template == nil {
		return []string{}
	}

	bonuses := monster.Template.AffinityBonuses()
	tags := []string{}
	if bonuses.FireDamageBonusPercent > 0 || bonuses.FireResistancePercent > 0 {
		tags = append(tags, "fire")
	}
	if bonuses.IceDamageBonusPercent > 0 || bonuses.IceResistancePercent > 0 {
		tags = append(tags, "ice")
	}
	if bonuses.LightningDamageBonusPercent > 0 || bonuses.LightningResistancePercent > 0 {
		tags = append(tags, "lightning", "storm")
	}
	if bonuses.PoisonDamageBonusPercent > 0 || bonuses.PoisonResistancePercent > 0 {
		tags = append(tags, "poison")
	}
	if bonuses.ArcaneDamageBonusPercent > 0 || bonuses.ArcaneResistancePercent > 0 {
		tags = append(tags, "arcane")
	}
	if bonuses.HolyDamageBonusPercent > 0 || bonuses.HolyResistancePercent > 0 {
		tags = append(tags, "holy")
	}
	if bonuses.ShadowDamageBonusPercent > 0 || bonuses.ShadowResistancePercent > 0 {
		tags = append(tags, "shadow")
	}
	return rewardContextNormalizeUniqueStrings(tags)
}

func rewardElementalTagsForEncounter(
	encounter *models.MonsterEncounter,
) []string {
	if encounter == nil {
		return []string{}
	}

	tags := []string{}
	for _, member := range encounter.Members {
		tags = append(tags, rewardElementalTagsForMonster(&member.Monster)...)
	}
	return rewardContextNormalizeUniqueStrings(tags)
}
