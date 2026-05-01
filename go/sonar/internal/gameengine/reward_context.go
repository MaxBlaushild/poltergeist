package gameengine

import (
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func buildRandomRewardContextForQuestTurnIn(
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
		}
		if node.Scenario != nil {
			context.InternalTags = append(context.InternalTags, []string(node.Scenario.InternalTags)...)
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

	context.StatTags = normalizeRewardContextStrings(context.StatTags)
	context.Proficiencies = normalizeRewardContextStrings(context.Proficiencies)
	context.InternalTags = normalizeRewardContextStrings(context.InternalTags)
	context.ElementalTags = normalizeRewardContextStrings(context.ElementalTags)
	return context
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
	return normalizeRewardContextStrings(tags)
}

func rewardElementalTagsForEncounter(encounter *models.MonsterEncounter) []string {
	if encounter == nil {
		return []string{}
	}

	tags := []string{}
	for _, member := range encounter.Members {
		tags = append(tags, rewardElementalTagsForMonster(&member.Monster)...)
	}
	return normalizeRewardContextStrings(tags)
}

func normalizeRewardContextStrings(values []string) []string {
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
