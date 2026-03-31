package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

const monsterTemplateAffinityPromptTemplate = `
You are tuning combat affinity stats for a fantasy RPG monster template.

Monster template:
- Type: %s
- Name: %s
- Description: %s
- Strength: %d
- Dexterity: %d
- Constitution: %d
- Intelligence: %d
- Wisdom: %d
- Charisma: %d

Return JSON only:
{
  "affinityDamageBonuses": {
    "fire": 0
  },
  "affinityResistances": {
    "fire": 0
  }
}

Hard rules:
- Use only these affinities: physical, piercing, slashing, bludgeoning, fire, ice, lightning, poison, arcane, holy, shadow.
- Values are integer percentages.
- Most affinities should stay at 0. Prefer 1-3 meaningful bonuses and 1-3 meaningful resistances.
- Damage bonus values should usually be between -25 and 60.
- Resistance values should usually be between -50 and 60.
- Only assign a negative value when the monster fantasy strongly suggests a vulnerability or clumsiness.
- Favor stats that reinforce the monster fantasy:
  - physical/piercing/slashing/bludgeoning for martial or beastly monsters
  - fire/ice/lightning/poison/arcane/holy/shadow for obviously elemental or magical monsters
- Resistances should reflect what the monster shrugs off or is vulnerable to.
- Damage bonuses should reflect what kind of damage the monster is best at dealing.
`

type generatedMonsterTemplateAffinityPayload struct {
	AffinityDamageBonuses map[string]int `json:"affinityDamageBonuses"`
	AffinityResistances   map[string]int `json:"affinityResistances"`
}

func scoreMonsterTemplateAffinities(
	ctx context.Context,
	template *models.MonsterTemplate,
	priest deep_priest.DeepPriest,
) models.CharacterStatBonuses {
	if template == nil {
		return models.CharacterStatBonuses{}
	}

	if priest != nil {
		if generated, err := generateMonsterTemplateAffinitiesWithLLM(ctx, template, priest); err == nil {
			return generated
		}
	}

	return deriveMonsterTemplateAffinitiesHeuristically(template)
}

func generateMonsterTemplateAffinitiesWithLLM(
	_ context.Context,
	template *models.MonsterTemplate,
	priest deep_priest.DeepPriest,
) (models.CharacterStatBonuses, error) {
	if template == nil {
		return models.CharacterStatBonuses{}, fmt.Errorf("template missing")
	}
	if priest == nil {
		return models.CharacterStatBonuses{}, fmt.Errorf("deep priest unavailable")
	}

	prompt := fmt.Sprintf(
		monsterTemplateAffinityPromptTemplate,
		strings.TrimSpace(string(models.NormalizeMonsterTemplateType(string(template.MonsterType)))),
		strings.TrimSpace(template.Name),
		strings.TrimSpace(template.Description),
		template.BaseStrength,
		template.BaseDexterity,
		template.BaseConstitution,
		template.BaseIntelligence,
		template.BaseWisdom,
		template.BaseCharisma,
	)

	answer, err := priest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return models.CharacterStatBonuses{}, err
	}

	var payload generatedMonsterTemplateAffinityPayload
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &payload); err != nil {
		return models.CharacterStatBonuses{}, err
	}
	return sanitizeMonsterTemplateAffinityPayload(payload), nil
}

func sanitizeMonsterTemplateAffinityPayload(payload generatedMonsterTemplateAffinityPayload) models.CharacterStatBonuses {
	bonuses := models.CharacterStatBonuses{}
	for key, value := range payload.AffinityDamageBonuses {
		setAffinityDamageBonusPercent(&bonuses, key, normalizeAffinityPercent(value, -25, 60))
	}
	for key, value := range payload.AffinityResistances {
		setAffinityResistancePercent(&bonuses, key, normalizeAffinityPercent(value, -50, 60))
	}
	return bonuses
}

func normalizeAffinityPercent(value int, minValue int, maxValue int) int {
	if value < minValue {
		value = minValue
	}
	if value > maxValue {
		value = maxValue
	}
	if value == 0 {
		return 0
	}
	rounded := (value / 5) * 5
	if value > 0 && value%5 != 0 {
		rounded += 5
	}
	if value < 0 && value%5 != 0 {
		rounded -= 5
	}
	if rounded < minValue {
		return minValue
	}
	if rounded > maxValue {
		return maxValue
	}
	return rounded
}

func deriveMonsterTemplateAffinitiesHeuristically(template *models.MonsterTemplate) models.CharacterStatBonuses {
	if template == nil {
		return models.CharacterStatBonuses{}
	}

	var bonuses models.CharacterStatBonuses
	text := strings.ToLower(strings.TrimSpace(template.Name + " " + template.Description))

	addDamageBonus := func(affinity models.DamageAffinity, delta int) {
		current := bonuses.DamageBonusPercentForAffinity(string(affinity))
		setAffinityDamageBonusPercent(&bonuses, string(affinity), clampAffinityPercent(current+delta, -25, 60))
	}
	addResistance := func(affinity models.DamageAffinity, delta int) {
		current := bonuses.ResistancePercentForAffinity(string(affinity))
		setAffinityResistancePercent(&bonuses, string(affinity), clampAffinityPercent(current+delta, -50, 60))
	}

	type keywordRule struct {
		keywords      []string
		damage        []models.DamageAffinity
		resist        []models.DamageAffinity
		vulnerableTo  []models.DamageAffinity
		damageDelta   int
		resistDelta   int
		vulnerablePct int
	}

	rules := []keywordRule{
		{keywords: []string{"fire", "ember", "cinder", "flame", "inferno", "lava", "ash"}, damage: []models.DamageAffinity{models.DamageAffinityFire}, resist: []models.DamageAffinity{models.DamageAffinityFire}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityIce}, damageDelta: 40, resistDelta: 50, vulnerablePct: -25},
		{keywords: []string{"frost", "ice", "glacier", "winter", "rime"}, damage: []models.DamageAffinity{models.DamageAffinityIce}, resist: []models.DamageAffinity{models.DamageAffinityIce}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityFire}, damageDelta: 40, resistDelta: 50, vulnerablePct: -25},
		{keywords: []string{"storm", "lightning", "thunder", "spark", "volt"}, damage: []models.DamageAffinity{models.DamageAffinityLightning}, resist: []models.DamageAffinity{models.DamageAffinityLightning}, damageDelta: 40, resistDelta: 45},
		{keywords: []string{"venom", "poison", "toxic", "corrosive", "ooze", "acid"}, damage: []models.DamageAffinity{models.DamageAffinityPoison}, resist: []models.DamageAffinity{models.DamageAffinityPoison}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityFire}, damageDelta: 35, resistDelta: 50, vulnerablePct: -15},
		{keywords: []string{"arcane", "mage", "sorcer", "wizard", "psionic", "mind", "eldritch"}, damage: []models.DamageAffinity{models.DamageAffinityArcane}, resist: []models.DamageAffinity{models.DamageAffinityArcane}, damageDelta: 35, resistDelta: 30},
		{keywords: []string{"holy", "radiant", "sun", "angel", "saint"}, damage: []models.DamageAffinity{models.DamageAffinityHoly}, resist: []models.DamageAffinity{models.DamageAffinityHoly}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityShadow}, damageDelta: 35, resistDelta: 35, vulnerablePct: -20},
		{keywords: []string{"shadow", "umbral", "void", "necrot", "undead", "wraith", "ghost", "specter", "spectre", "shade"}, damage: []models.DamageAffinity{models.DamageAffinityShadow}, resist: []models.DamageAffinity{models.DamageAffinityShadow}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityHoly}, damageDelta: 35, resistDelta: 45, vulnerablePct: -30},
		{keywords: []string{"skeleton", "zombie", "bone", "grave"}, resist: []models.DamageAffinity{models.DamageAffinityPiercing}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityBludgeoning}, resistDelta: 20, vulnerablePct: -25},
		{keywords: []string{"stone", "iron", "armored", "armoured", "golem"}, resist: []models.DamageAffinity{models.DamageAffinitySlashing, models.DamageAffinityPiercing}, vulnerableTo: []models.DamageAffinity{models.DamageAffinityBludgeoning}, resistDelta: 20, vulnerablePct: -20},
	}

	for _, rule := range rules {
		if !containsAnyMonsterKeyword(text, rule.keywords) {
			continue
		}
		for _, affinity := range rule.damage {
			addDamageBonus(affinity, rule.damageDelta)
		}
		for _, affinity := range rule.resist {
			addResistance(affinity, rule.resistDelta)
		}
		for _, affinity := range rule.vulnerableTo {
			addResistance(affinity, rule.vulnerablePct)
		}
	}

	physicalScore := template.BaseStrength + template.BaseDexterity + template.BaseConstitution
	mentalScore := template.BaseIntelligence + template.BaseWisdom + template.BaseCharisma
	if physicalScore >= mentalScore+6 {
		if template.BaseDexterity >= template.BaseStrength+3 {
			addDamageBonus(models.DamageAffinityPiercing, 25)
			addResistance(models.DamageAffinityPiercing, 10)
		} else if template.BaseStrength >= template.BaseDexterity+3 {
			addDamageBonus(models.DamageAffinityBludgeoning, 25)
			addResistance(models.DamageAffinityPhysical, 10)
		} else {
			addDamageBonus(models.DamageAffinitySlashing, 20)
			addResistance(models.DamageAffinityPhysical, 10)
		}
	}
	if mentalScore >= physicalScore+6 && !hasMeaningfulMagicalDamageBonus(bonuses) {
		addDamageBonus(models.DamageAffinityArcane, 25)
		addResistance(models.DamageAffinityArcane, 15)
	}
	if template.BaseConstitution >= 15 {
		addResistance(models.DamageAffinityPhysical, 10)
	}

	return bonuses
}

func containsAnyMonsterKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func hasMeaningfulMagicalDamageBonus(bonuses models.CharacterStatBonuses) bool {
	return bonuses.FireDamageBonusPercent > 0 ||
		bonuses.IceDamageBonusPercent > 0 ||
		bonuses.LightningDamageBonusPercent > 0 ||
		bonuses.PoisonDamageBonusPercent > 0 ||
		bonuses.ArcaneDamageBonusPercent > 0 ||
		bonuses.HolyDamageBonusPercent > 0 ||
		bonuses.ShadowDamageBonusPercent > 0
}

func clampAffinityPercent(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func applyAffinityBonusesToMonsterTemplate(template *models.MonsterTemplate, bonuses models.CharacterStatBonuses) {
	if template == nil {
		return
	}
	template.PhysicalDamageBonusPercent = bonuses.PhysicalDamageBonusPercent
	template.PiercingDamageBonusPercent = bonuses.PiercingDamageBonusPercent
	template.SlashingDamageBonusPercent = bonuses.SlashingDamageBonusPercent
	template.BludgeoningDamageBonusPercent = bonuses.BludgeoningDamageBonusPercent
	template.FireDamageBonusPercent = bonuses.FireDamageBonusPercent
	template.IceDamageBonusPercent = bonuses.IceDamageBonusPercent
	template.LightningDamageBonusPercent = bonuses.LightningDamageBonusPercent
	template.PoisonDamageBonusPercent = bonuses.PoisonDamageBonusPercent
	template.ArcaneDamageBonusPercent = bonuses.ArcaneDamageBonusPercent
	template.HolyDamageBonusPercent = bonuses.HolyDamageBonusPercent
	template.ShadowDamageBonusPercent = bonuses.ShadowDamageBonusPercent
	template.PhysicalResistancePercent = bonuses.PhysicalResistancePercent
	template.PiercingResistancePercent = bonuses.PiercingResistancePercent
	template.SlashingResistancePercent = bonuses.SlashingResistancePercent
	template.BludgeoningResistancePercent = bonuses.BludgeoningResistancePercent
	template.FireResistancePercent = bonuses.FireResistancePercent
	template.IceResistancePercent = bonuses.IceResistancePercent
	template.LightningResistancePercent = bonuses.LightningResistancePercent
	template.PoisonResistancePercent = bonuses.PoisonResistancePercent
	template.ArcaneResistancePercent = bonuses.ArcaneResistancePercent
	template.HolyResistancePercent = bonuses.HolyResistancePercent
	template.ShadowResistancePercent = bonuses.ShadowResistancePercent
}

func setAffinityDamageBonusPercent(bonuses *models.CharacterStatBonuses, affinity string, value int) {
	if bonuses == nil {
		return
	}
	switch models.NormalizeDamageAffinity(affinity) {
	case models.DamageAffinityPiercing:
		bonuses.PiercingDamageBonusPercent = value
	case models.DamageAffinitySlashing:
		bonuses.SlashingDamageBonusPercent = value
	case models.DamageAffinityBludgeoning:
		bonuses.BludgeoningDamageBonusPercent = value
	case models.DamageAffinityFire:
		bonuses.FireDamageBonusPercent = value
	case models.DamageAffinityIce:
		bonuses.IceDamageBonusPercent = value
	case models.DamageAffinityLightning:
		bonuses.LightningDamageBonusPercent = value
	case models.DamageAffinityPoison:
		bonuses.PoisonDamageBonusPercent = value
	case models.DamageAffinityArcane:
		bonuses.ArcaneDamageBonusPercent = value
	case models.DamageAffinityHoly:
		bonuses.HolyDamageBonusPercent = value
	case models.DamageAffinityShadow:
		bonuses.ShadowDamageBonusPercent = value
	default:
		bonuses.PhysicalDamageBonusPercent = value
	}
}

func setAffinityResistancePercent(bonuses *models.CharacterStatBonuses, affinity string, value int) {
	if bonuses == nil {
		return
	}
	switch models.NormalizeDamageAffinity(affinity) {
	case models.DamageAffinityPiercing:
		bonuses.PiercingResistancePercent = value
	case models.DamageAffinitySlashing:
		bonuses.SlashingResistancePercent = value
	case models.DamageAffinityBludgeoning:
		bonuses.BludgeoningResistancePercent = value
	case models.DamageAffinityFire:
		bonuses.FireResistancePercent = value
	case models.DamageAffinityIce:
		bonuses.IceResistancePercent = value
	case models.DamageAffinityLightning:
		bonuses.LightningResistancePercent = value
	case models.DamageAffinityPoison:
		bonuses.PoisonResistancePercent = value
	case models.DamageAffinityArcane:
		bonuses.ArcaneResistancePercent = value
	case models.DamageAffinityHoly:
		bonuses.HolyResistancePercent = value
	case models.DamageAffinityShadow:
		bonuses.ShadowResistancePercent = value
	default:
		bonuses.PhysicalResistancePercent = value
	}
}
