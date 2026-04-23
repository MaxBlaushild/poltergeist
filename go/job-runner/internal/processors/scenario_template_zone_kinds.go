package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

const scenarioTemplateZoneKindPromptTemplate = `
You are classifying the best-fit zone kind for a reusable fantasy MMORPG scenario template.
%s
%s

Scenario template:
- prompt: %s
- open ended: %t
- difficulty: %d
- reward model: %s
- option count: %d

Return JSON only:
{
  "zoneKind": "forest"
}

Rules:
- zoneKind must be one of the allowed slugs exactly as written.
- Choose the single best-fit zone kind for where this reusable template would most naturally belong.
- Base the decision on terrain, hazards, traversal, factions, props, creatures, weather, and environmental vibe implied by the template prompt.
- Do not pick a zone kind just because it is common; pick the strongest environmental fit.
`

type scenarioTemplateZoneKindPayload struct {
	ZoneKind string `json:"zoneKind"`
}

func classifyScenarioTemplateZoneKind(
	ctx context.Context,
	template *models.ScenarioTemplate,
	zoneKinds []models.ZoneKind,
	priest deep_priest.DeepPriest,
) string {
	if template == nil {
		return ""
	}
	if priest != nil && len(zoneKinds) > 0 {
		if generated, err := generateScenarioTemplateZoneKindWithLLM(
			ctx,
			template,
			zoneKinds,
			priest,
		); err == nil {
			return generated
		}
	}
	return deriveScenarioZoneKindHeuristically(
		zoneKinds,
		template.ZoneKind,
		strings.TrimSpace(template.Prompt),
		scenarioGenrePromptLabel(template.Genre),
	)
}

func generateScenarioTemplateZoneKindWithLLM(
	_ context.Context,
	template *models.ScenarioTemplate,
	zoneKinds []models.ZoneKind,
	priest deep_priest.DeepPriest,
) (string, error) {
	if template == nil {
		return "", fmt.Errorf("template missing")
	}
	if priest == nil {
		return "", fmt.Errorf("deep priest unavailable")
	}
	if len(zoneKinds) == 0 {
		return "", fmt.Errorf("zone kinds unavailable")
	}

	prompt := fmt.Sprintf(
		scenarioTemplateZoneKindPromptTemplate,
		scenarioGenreInstructionBlock(template.Genre),
		buildScenarioZoneKindInstructionBlock(
			zoneKinds,
			findZoneKindBySlug(zoneKinds, template.ZoneKind),
			"current zone kind",
		),
		quotedOrNone(strings.TrimSpace(template.Prompt)),
		template.OpenEnded,
		template.Difficulty,
		strings.TrimSpace(string(template.RewardMode)),
		len(template.Options),
	)

	answer, err := priest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return "", err
	}

	var payload scenarioTemplateZoneKindPayload
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &payload); err != nil {
		return "", err
	}

	return normalizeScenarioGeneratedZoneKind(
		payload.ZoneKind,
		zoneKinds,
		deriveScenarioZoneKindHeuristically(
			zoneKinds,
			template.ZoneKind,
			strings.TrimSpace(template.Prompt),
			scenarioGenrePromptLabel(template.Genre),
		),
	), nil
}
