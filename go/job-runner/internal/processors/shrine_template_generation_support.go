package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

const shrineTemplateGenerationPromptTemplate = `
You are designing %d reusable fantasy MMORPG shrine templates.

Recent shrine templates to avoid echoing:
%s

Return JSON only:
{
  "templates": [
    {
      "name": "Shrine of Vigor",
      "blessingName": "Blessing of Vigor",
      "description": "2-4 vivid sentences describing the shrine itself",
      "effectDescription": "One sentence explaining the boon the player receives",
      "effectKind": "strength|dexterity|constitution|intelligence|wisdom|charisma|health_regen|mana_regen|physical_damage|arcane_damage|holy_damage|shadow_damage|fire_resistance|ice_resistance|lightning_resistance|poison_resistance|physical_resistance|warding",
      "baseMagnitude": 1-4
    }
  ]
}

Hard rules:
- Output exactly %d templates.
- Every shrine is beneficial, mystical, and reusable across many locations.
- The template name must clearly imply the effect.
- The blessingName must sound like a status the player receives after invoking the shrine.
- description should describe the shrine itself, not a one-off event.
- effectDescription must describe the benefit in player-facing language.
- baseMagnitude should stay conservative because the blessing scales with player level.
- Keep the templates materially distinct from one another and from the recent shrine list.
`

type generatedShrineTemplatePayload struct {
	Name              string `json:"name"`
	BlessingName      string `json:"blessingName"`
	Description       string `json:"description"`
	EffectDescription string `json:"effectDescription"`
	EffectKind        string `json:"effectKind"`
	BaseMagnitude     int    `json:"baseMagnitude"`
}

type generatedShrineTemplatesResponse struct {
	Templates []generatedShrineTemplatePayload `json:"templates"`
}

func generateShrineTemplatesForZoneKind(
	ctx context.Context,
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	rawZoneKind string,
	count int,
) ([]models.ShrineTemplate, error) {
	if count <= 0 {
		return []models.ShrineTemplate{}, nil
	}

	zoneKind, err := loadOptionalZoneKind(ctx, dbClient, rawZoneKind)
	if err != nil {
		return nil, fmt.Errorf("failed to load shrine template zone kind: %w", err)
	}

	prompt := fmt.Sprintf(
		shrineTemplateGenerationPromptTemplate,
		count,
		buildRecentShrineTemplateAvoidance(ctx, dbClient, rawZoneKind, 12),
		count,
	)
	if zoneKindBlock := zoneKindInstructionBlock(zoneKind); zoneKindBlock != "" {
		prompt = strings.TrimSpace(zoneKindBlock + "\n\n" + prompt)
	}

	answer, err := deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, fmt.Errorf("failed to generate shrine templates: %w", err)
	}

	generated := &generatedShrineTemplatesResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return nil, fmt.Errorf("failed to parse generated shrine template payload: %w", err)
	}

	sanitized := sanitizeGeneratedShrineTemplates(generated.Templates, count, zoneKind)
	created := make([]models.ShrineTemplate, 0, len(sanitized))
	for _, template := range sanitized {
		next := template
		if err := dbClient.ShrineTemplate().Create(ctx, &next); err != nil {
			return created, fmt.Errorf("failed to create shrine template: %w", err)
		}
		created = append(created, next)
	}
	return created, nil
}

func buildRecentShrineTemplateAvoidance(
	ctx context.Context,
	dbClient db.DbClient,
	rawZoneKind string,
	limit int,
) string {
	if dbClient == nil || limit <= 0 {
		return "- none"
	}
	templates, err := dbClient.ShrineTemplate().FindByZoneKind(ctx, rawZoneKind)
	if err != nil || len(templates) == 0 {
		return "- none"
	}
	lines := make([]string, 0, min(len(templates), limit))
	for _, template := range templates {
		name := strings.TrimSpace(template.Name)
		if name == "" {
			continue
		}
		effect := strings.TrimSpace(template.EffectDescription)
		if effect == "" {
			effect = string(template.EffectKind)
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", name, effect))
		if len(lines) >= limit {
			break
		}
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func sanitizeGeneratedShrineTemplates(
	input []generatedShrineTemplatePayload,
	count int,
	zoneKind *models.ZoneKind,
) []models.ShrineTemplate {
	out := make([]models.ShrineTemplate, 0, count)
	seenNames := map[string]struct{}{}
	effectKinds := models.AllShrineEffectKinds()
	for _, spec := range input {
		template := sanitizeGeneratedShrineTemplate(spec, zoneKind)
		if template.Name == "" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(template.Name))
		if _, exists := seenNames[key]; exists {
			continue
		}
		seenNames[key] = struct{}{}
		out = append(out, template)
		if len(out) >= count {
			return out
		}
	}

	for _, kind := range effectKinds {
		if len(out) >= count {
			break
		}
		template := fallbackShrineTemplate(kind, zoneKind)
		key := strings.ToLower(strings.TrimSpace(template.Name))
		if _, exists := seenNames[key]; exists {
			continue
		}
		seenNames[key] = struct{}{}
		out = append(out, template)
	}
	return out
}

func sanitizeGeneratedShrineTemplate(
	spec generatedShrineTemplatePayload,
	zoneKind *models.ZoneKind,
) models.ShrineTemplate {
	effectKind := models.NormalizeShrineEffectKind(spec.EffectKind)
	template := models.ShrineTemplate{
		ZoneKind:          models.ZoneKindPromptSlug(zoneKind),
		Name:              sanitizeShrineTemplateName(spec.Name, effectKind),
		Description:       sanitizeShrineText(spec.Description, 320),
		BlessingName:      sanitizeShrineBlessingName(spec.BlessingName, effectKind),
		EffectDescription: sanitizeShrineEffectDescription(spec.EffectDescription, effectKind),
		EffectKind:        effectKind,
		BaseMagnitude:     clampInt(spec.BaseMagnitude, 1, 4),
	}
	if template.Description == "" {
		template.Description = fallbackShrineTemplate(effectKind, zoneKind).Description
	}
	return template
}

func sanitizeShrineTemplateName(raw string, effectKind models.ShrineEffectKind) string {
	name := sanitizeShrineText(raw, 80)
	if countWords(name) > 6 {
		name = strings.Join(strings.Fields(name)[:6], " ")
	}
	if name != "" {
		return name
	}
	return fallbackShrineTemplate(effectKind, nil).Name
}

func sanitizeShrineBlessingName(raw string, effectKind models.ShrineEffectKind) string {
	name := sanitizeShrineText(raw, 80)
	if name != "" {
		return name
	}
	return fallbackShrineTemplate(effectKind, nil).BlessingName
}

func sanitizeShrineEffectDescription(raw string, effectKind models.ShrineEffectKind) string {
	text := sanitizeShrineText(raw, 140)
	if text != "" {
		return text
	}
	return fallbackShrineTemplate(effectKind, nil).EffectDescription
}

func sanitizeShrineText(raw string, maxLen int) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
	if trimmed == "" {
		return ""
	}
	if maxLen > 0 && len(trimmed) > maxLen {
		trimmed = strings.TrimSpace(trimmed[:maxLen])
	}
	return trimmed
}

type shrineFallbackSpec struct {
	Name              string
	BlessingName      string
	EffectDescription string
	Description       string
}

func fallbackShrineTemplate(
	effectKind models.ShrineEffectKind,
	zoneKind *models.ZoneKind,
) models.ShrineTemplate {
	spec := shrineFallbacksByEffect()[effectKind]
	if spec.Name == "" {
		spec = shrineFallbacksByEffect()[models.ShrineEffectKindStrength]
	}
	label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	description := spec.Description
	if label != "" {
		description = fmt.Sprintf("%s Its magic feels deeply attuned to %s places.", description, strings.ToLower(label))
	}
	return models.ShrineTemplate{
		ZoneKind:          models.ZoneKindPromptSlug(zoneKind),
		Name:              spec.Name,
		Description:       description,
		BlessingName:      spec.BlessingName,
		EffectDescription: spec.EffectDescription,
		EffectKind:        effectKind,
		BaseMagnitude:     2,
	}
}

func shrineFallbacksByEffect() map[models.ShrineEffectKind]shrineFallbackSpec {
	return map[models.ShrineEffectKind]shrineFallbackSpec{
		models.ShrineEffectKindStrength: {
			Name:              "Shrine of Vigor",
			BlessingName:      "Blessing of Vigor",
			EffectDescription: "Bolsters your strength for a day.",
			Description:       "A battle-worn shrine hums with a grounded, iron-hearted power that answers bold hands and steady breath.",
		},
		models.ShrineEffectKindDexterity: {
			Name:              "Shrine of Grace",
			BlessingName:      "Blessing of Grace",
			EffectDescription: "Sharpens your agility for a day.",
			Description:       "Slender ribbons and precise carvings circle this shrine, and every movement near it feels just a little lighter.",
		},
		models.ShrineEffectKindConstitution: {
			Name:              "Shrine of Endurance",
			BlessingName:      "Blessing of Endurance",
			EffectDescription: "Fortifies your constitution for a day.",
			Description:       "This heavy, patient shrine radiates a settling warmth like stone that has weathered a thousand hard seasons.",
		},
		models.ShrineEffectKindIntelligence: {
			Name:              "Shrine of Insight",
			BlessingName:      "Blessing of Insight",
			EffectDescription: "Heightens your intelligence for a day.",
			Description:       "Runes drift beneath the shrine's surface, arranging themselves into patterns that reward a sharpened mind.",
		},
		models.ShrineEffectKindWisdom: {
			Name:              "Shrine of Clarity",
			BlessingName:      "Blessing of Clarity",
			EffectDescription: "Deepens your wisdom for a day.",
			Description:       "Soft lantern-light gathers around this shrine, and its calm presence seems to steady intuition and judgment alike.",
		},
		models.ShrineEffectKindCharisma: {
			Name:              "Shrine of Splendor",
			BlessingName:      "Blessing of Splendor",
			EffectDescription: "Magnifies your charisma for a day.",
			Description:       "Polished surfaces and inviting light give this shrine an almost ceremonial magnetism that lingers after contact.",
		},
		models.ShrineEffectKindHealthRegen: {
			Name:              "Shrine of Renewal",
			BlessingName:      "Blessing of Renewal",
			EffectDescription: "Fills you with restorative vitality for a day.",
			Description:       "Living light wells up through the shrine's seams, promising steady recovery instead of sudden force.",
		},
		models.ShrineEffectKindManaRegen: {
			Name:              "Shrine of Clarity's Well",
			BlessingName:      "Blessing of Flowing Thought",
			EffectDescription: "Quickens your mana recovery for a day.",
			Description:       "A lucid arcane pulse moves through this shrine in measured waves, as though it were breathing hidden power.",
		},
		models.ShrineEffectKindPhysicalDamage: {
			Name:              "Shrine of Fury",
			BlessingName:      "Blessing of Fury",
			EffectDescription: "Empowers your physical strikes for a day.",
			Description:       "Chiseled weapon motifs line the shrine, and its aura sparks like anticipation before a decisive blow.",
		},
		models.ShrineEffectKindArcaneDamage: {
			Name:              "Shrine of Sorcery",
			BlessingName:      "Blessing of Sorcery",
			EffectDescription: "Amplifies your arcane power for a day.",
			Description:       "Thin threads of violet-free eldritch light arc across this shrine's face in disciplined geometries.",
		},
		models.ShrineEffectKindHolyDamage: {
			Name:              "Shrine of Radiance",
			BlessingName:      "Blessing of Radiance",
			EffectDescription: "Brightens your holy power for a day.",
			Description:       "This shrine glows with a clean dawn-like brilliance that turns resolve into something almost sacred.",
		},
		models.ShrineEffectKindShadowDamage: {
			Name:              "Shrine of Dusk",
			BlessingName:      "Blessing of Dusk",
			EffectDescription: "Deepens your shadow power for a day.",
			Description:       "Dark stone and silver tracery lend this shrine a restrained menace, as if it prefers silence over spectacle.",
		},
		models.ShrineEffectKindFireResistance: {
			Name:              "Shrine of Cindershield",
			BlessingName:      "Blessing of Cindershield",
			EffectDescription: "Wards you against fire for a day.",
			Description:       "Scorched edges frame the shrine, but its heart remains cool and protective against consuming flame.",
		},
		models.ShrineEffectKindIceResistance: {
			Name:              "Shrine of the Thaw",
			BlessingName:      "Blessing of the Thaw",
			EffectDescription: "Wards you against frost for a day.",
			Description:       "A pale sheen coats the shrine without freezing it, and its warmth resists the bite of winter magic.",
		},
		models.ShrineEffectKindLightningResistance: {
			Name:              "Shrine of Still Skies",
			BlessingName:      "Blessing of Still Skies",
			EffectDescription: "Wards you against lightning for a day.",
			Description:       "Forked engravings fade into calm circles here, as though storms lose their will upon reaching the stone.",
		},
		models.ShrineEffectKindPoisonResistance: {
			Name:              "Shrine of Antivena",
			BlessingName:      "Blessing of Antivena",
			EffectDescription: "Wards you against poison for a day.",
			Description:       "Fresh herbs and bitter resins scent the air around this shrine, whose magic purifies rather than dazzles.",
		},
		models.ShrineEffectKindPhysicalResistance: {
			Name:              "Shrine of the Bulwark",
			BlessingName:      "Blessing of the Bulwark",
			EffectDescription: "Hardens you against physical harm for a day.",
			Description:       "Broad-shouldered architecture and shield motifs make this shrine feel less like an altar and more like a promise.",
		},
		models.ShrineEffectKindAllDamageResistance: {
			Name:              "Shrine of Warding",
			BlessingName:      "Blessing of Warding",
			EffectDescription: "Surrounds you with broad magical protection for a day.",
			Description:       "Layered rings of sigils drift around this shrine in patient cycles, knitting a general-purpose ward against many harms.",
		},
	}
}

func ensureShrineTemplatePool(
	ctx context.Context,
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	rawZoneKind string,
	count int,
) ([]models.ShrineTemplate, error) {
	templates, err := dbClient.ShrineTemplate().FindByZoneKind(ctx, rawZoneKind)
	if err != nil {
		return nil, err
	}
	if len(templates) >= count {
		return templates, nil
	}
	missing := count - len(templates)
	generated, err := generateShrineTemplatesForZoneKind(ctx, dbClient, deepPriestClient, rawZoneKind, missing)
	if err != nil {
		return nil, err
	}
	templates = append(templates, generated...)
	sort.SliceStable(templates, func(i, j int) bool {
		return templates[i].CreatedAt.After(templates[j].CreatedAt)
	})
	return templates, nil
}
