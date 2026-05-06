package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

const questArchetypeSuggestionPresetPromptTemplate = `
You are helping a game designer set up one reusable quest archetype generation job for StreetSekai, an urban fantasy MMORPG.

Design one compelling preset that feels specific, playable, and varied enough to produce a fun batch of quests.

Available zone kinds:
%s

Available location archetypes:
%s

Known family slugs:
%s

Current partial form values to riff from when useful:
- current count: %s
- current preferred zone kind: %s
- current preferred zone kind flavor seed: %s
- current theme prompt: %s
- current family tags: %s
- current family mix targets: %s
- current character tags: %s
- current internal tags: %s
- current required location archetypes: %s
- current required location metadata tags: %s

Return JSON only:
{
  "count": 8,
  "zoneKind": "zone-kind-slug-or-empty-string",
  "themePrompt": "string",
  "familyTags": ["tag_a", "tag_b"],
  "familyMixTargets": {
    "investigation": 2,
    "negotiation": 1
  },
  "characterTags": ["tag_a", "tag_b"],
  "internalTags": ["tag_a", "tag_b"],
  "requiredLocationArchetypeNames": ["Exact Archetype Name"],
  "requiredLocationMetadataTags": ["tag_a", "tag_b"]
}

Rules:
- Output JSON only. No markdown.
- Keep count between 1 and 12.
- When a current count is provided, preserve it unless the requested family mix would exceed it.
- Pick a coherent preset, not a grab bag.
- Make the theme prompt vivid and reusable, not a one-off plot synopsis.
- Favor street-level urban fantasy with faction pressure, traversal texture, and tangible stakes.
- If a preferred zone kind is provided, keep the preset anchored to that same zone kind unless it is invalid.
- When a zone kind is present, make the theme prompt unmistakably about that environment in its opening sentence.
- Name or strongly evoke zone-specific routes, landmarks, institutions, factions, terrain, or social pressure.
- Do not return a theme prompt that could fit a different zone kind unchanged.
- Let the zone kind also shape the tags, route texture, and location choices.
- familyTags, characterTags, internalTags, and requiredLocationMetadataTags should use lowercase snake_case when possible.
- familyMixTargets keys must be chosen only from the known family slugs list.
- Keep familyMixTargets ambitious but coherent. The total should not exceed count.
- Required location archetypes are optional. Returning none is valid when the theme works better with flexible routing.
- When the current required location archetypes are none, do not force specific ones unless they materially improve the preset.
- Prefer 2-4 character tags, 3-6 internal tags, 2-5 family tags, 2-4 required metadata tags, and 0-2 required location archetypes.
- requiredLocationArchetypeNames must match the available location archetype names exactly.
- Avoid generic filler like fantasy, adventure, generic, miscellaneous, or urban_fantasy.
- Avoid named proper nouns that make the preset too one-off.
- Prefer fun mixes like diplomacy under pressure, scramble logistics, occult malfunction, labor standoff, monster ecology, civic panic, rooftop pursuit, or fail-forward rescue.
`

type questArchetypeSuggestionPresetPayload struct {
	Count                        int            `json:"count"`
	ZoneKind                     string         `json:"zoneKind"`
	ThemePrompt                  string         `json:"themePrompt"`
	FamilyTags                   []string       `json:"familyTags"`
	FamilyMixTargets             map[string]int `json:"familyMixTargets"`
	CharacterTags                []string       `json:"characterTags"`
	InternalTags                 []string       `json:"internalTags"`
	RequiredLocationArchetypeIDs []string       `json:"requiredLocationArchetypeIds"`
	RequiredLocationMetadataTags []string       `json:"requiredLocationMetadataTags"`
}

type questArchetypeSuggestionPresetLLMResponse struct {
	Count                          int            `json:"count"`
	ZoneKind                       string         `json:"zoneKind"`
	ThemePrompt                    string         `json:"themePrompt"`
	FamilyTags                     []string       `json:"familyTags"`
	FamilyMixTargets               map[string]int `json:"familyMixTargets"`
	CharacterTags                  []string       `json:"characterTags"`
	InternalTags                   []string       `json:"internalTags"`
	RequiredLocationArchetypeNames []string       `json:"requiredLocationArchetypeNames"`
	RequiredLocationMetadataTags   []string       `json:"requiredLocationMetadataTags"`
}

type questArchetypeSuggestionPresetResponse struct {
	Count                        int                                             `json:"count"`
	ZoneKind                     string                                          `json:"zoneKind"`
	ThemePrompt                  string                                          `json:"themePrompt"`
	FamilyTags                   models.StringArray                              `json:"familyTags"`
	FamilyMixTargets             models.QuestArchetypeSuggestionFamilyMixTargets `json:"familyMixTargets"`
	CharacterTags                models.StringArray                              `json:"characterTags"`
	InternalTags                 models.StringArray                              `json:"internalTags"`
	RequiredLocationArchetypeIDs models.StringArray                              `json:"requiredLocationArchetypeIds"`
	RequiredLocationMetadataTags models.StringArray                              `json:"requiredLocationMetadataTags"`
}

func (s *server) generateQuestArchetypeSuggestionPreset(ctx *gin.Context) {
	var body questArchetypeSuggestionPresetPayload
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preset, err := s.buildQuestArchetypeSuggestionPreset(ctx.Request.Context(), body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, preset)
}

func (s *server) buildQuestArchetypeSuggestionPreset(
	ctx context.Context,
	hints questArchetypeSuggestionPresetPayload,
) (*questArchetypeSuggestionPresetResponse, error) {
	if s == nil || s.deepPriest == nil {
		return nil, fmt.Errorf("quest archetype preset generation is unavailable")
	}

	locationArchetypes, err := s.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load location archetypes: %w", err)
	}
	zoneKinds, err := s.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load zone kinds: %w", err)
	}
	normalizedHintLocationArchetypeIDs, err := s.normalizeQuestArchetypeSuggestionLocationArchetypeIDs(
		ctx,
		hints.RequiredLocationArchetypeIDs,
	)
	if err != nil {
		return nil, err
	}

	preferredZoneKind, err := s.resolveOptionalZoneKind(ctx, hints.ZoneKind)
	if err != nil {
		preferredZoneKind = nil
	}
	prompt := fmt.Sprintf(
		questArchetypeSuggestionPresetPromptTemplate,
		models.ZoneKindsPromptOptions(zoneKinds),
		buildQuestArchetypeSuggestionPresetLocationArchetypePrompt(locationArchetypes),
		strings.Join(models.QuestArchetypeSuggestionKnownFamilySlugs(), ", "),
		renderQuestArchetypeSuggestionPresetCountHint(hints.Count),
		renderQuestArchetypeSuggestionPresetZoneKindHint(preferredZoneKind, hints.ZoneKind),
		renderQuestArchetypeSuggestionPresetZoneKindSeedHint(preferredZoneKind),
		renderQuestArchetypeSuggestionPresetQuotedHint(hints.ThemePrompt),
		renderQuestArchetypeSuggestionPresetTagHint(hints.FamilyTags),
		renderQuestArchetypeSuggestionPresetFamilyMixTargets(hints.FamilyMixTargets),
		renderQuestArchetypeSuggestionPresetTagHint(hints.CharacterTags),
		renderQuestArchetypeSuggestionPresetTagHint(hints.InternalTags),
		buildQuestArchetypeSuggestionPresetHintLocationArchetypePrompt(
			normalizedHintLocationArchetypeIDs,
			locationArchetypes,
		),
		renderQuestArchetypeSuggestionPresetTagHint(hints.RequiredLocationMetadataTags),
	)

	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, fmt.Errorf("failed to generate quest archetype preset: %w", err)
	}

	var generated questArchetypeSuggestionPresetLLMResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
		return nil, fmt.Errorf("failed to parse quest archetype preset: %w", err)
	}

	hints.RequiredLocationArchetypeIDs = normalizedHintLocationArchetypeIDs
	return s.sanitizeQuestArchetypeSuggestionPresetResponse(
		ctx,
		generated,
		hints,
		locationArchetypes,
	)
}

func (s *server) sanitizeQuestArchetypeSuggestionPresetResponse(
	ctx context.Context,
	generated questArchetypeSuggestionPresetLLMResponse,
	hints questArchetypeSuggestionPresetPayload,
	locationArchetypes []*models.LocationArchetype,
) (*questArchetypeSuggestionPresetResponse, error) {
	count := generated.Count
	if hints.Count > 0 {
		count = hints.Count
	}
	if count <= 0 {
		count = 2
	}
	if count < 1 {
		count = 1
	}
	if count > 12 {
		count = 12
	}

	familyMixTargets := models.NormalizeQuestArchetypeSuggestionFamilyMixTargets(generated.FamilyMixTargets)
	if len(familyMixTargets) == 0 {
		familyMixTargets = models.NormalizeQuestArchetypeSuggestionFamilyMixTargets(hints.FamilyMixTargets)
	}
	familyMixCount := sumQuestArchetypeSuggestionFamilyMixTargets(familyMixTargets)
	if familyMixCount > count {
		familyMixTargets = trimQuestArchetypeSuggestionPresetFamilyMixTargets(familyMixTargets, count)
	}

	zoneKindRaw := strings.TrimSpace(generated.ZoneKind)
	if zoneKindRaw == "" {
		zoneKindRaw = strings.TrimSpace(hints.ZoneKind)
	}
	resolvedZoneKind, err := s.resolveOptionalZoneKind(ctx, zoneKindRaw)
	if err != nil {
		resolvedZoneKind = nil
	}
	zoneKind := models.ZoneKindPromptSlug(resolvedZoneKind)
	if zoneKind == "" {
		zoneKind = models.NormalizeZoneKind(zoneKindRaw)
	}

	familyTags := normalizeQuestTemplateInternalTags(generated.FamilyTags)
	if len(familyTags) == 0 {
		familyTags = normalizeQuestTemplateInternalTags(hints.FamilyTags)
	}
	if len(familyTags) == 0 {
		familyTags = deriveQuestArchetypeSuggestionPresetFamilyTags(familyMixTargets)
	}

	characterTags := normalizeQuestTemplateCharacterTags(generated.CharacterTags)
	if len(characterTags) == 0 {
		characterTags = normalizeQuestTemplateCharacterTags(hints.CharacterTags)
	}

	internalTags := normalizeQuestTemplateInternalTags(generated.InternalTags)
	if len(internalTags) == 0 {
		internalTags = normalizeQuestTemplateInternalTags(hints.InternalTags)
	}
	if len(internalTags) == 0 {
		internalTags = deriveQuestArchetypeSuggestionPresetInternalTags(familyMixTargets, familyTags)
	}

	requiredLocationMetadataTags := normalizeQuestTemplateInternalTags(generated.RequiredLocationMetadataTags)
	if len(requiredLocationMetadataTags) == 0 {
		requiredLocationMetadataTags = normalizeQuestTemplateInternalTags(hints.RequiredLocationMetadataTags)
	}

	requiredLocationArchetypeIDs := buildQuestArchetypeSuggestionPresetLocationArchetypeIDs(
		generated.RequiredLocationArchetypeNames,
		locationArchetypes,
	)
	if len(requiredLocationArchetypeIDs) == 0 {
		requiredLocationArchetypeIDs = normalizeQuestTemplateInternalTags(hints.RequiredLocationArchetypeIDs)
	}

	themePrompt := strings.TrimSpace(generated.ThemePrompt)
	if themePrompt == "" {
		themePrompt = strings.TrimSpace(hints.ThemePrompt)
	}
	if themePrompt == "" {
		themePrompt = buildQuestArchetypeSuggestionPresetFallbackThemePrompt(
			resolvedZoneKind,
			zoneKind,
			familyMixTargets,
			familyTags,
			internalTags,
		)
	}
	themePrompt = ensureQuestArchetypeSuggestionPresetThemeMatchesZoneKind(
		themePrompt,
		resolvedZoneKind,
		zoneKind,
	)

	return &questArchetypeSuggestionPresetResponse{
		Count:                        count,
		ZoneKind:                     zoneKind,
		ThemePrompt:                  themePrompt,
		FamilyTags:                   familyTags,
		FamilyMixTargets:             familyMixTargets,
		CharacterTags:                characterTags,
		InternalTags:                 internalTags,
		RequiredLocationArchetypeIDs: requiredLocationArchetypeIDs,
		RequiredLocationMetadataTags: requiredLocationMetadataTags,
	}, nil
}

func buildQuestArchetypeSuggestionPresetLocationArchetypePrompt(
	locationArchetypes []*models.LocationArchetype,
) string {
	if len(locationArchetypes) == 0 {
		return "none"
	}
	sorted := append([]*models.LocationArchetype(nil), locationArchetypes...)
	sort.Slice(sorted, func(left, right int) bool {
		leftName := ""
		if sorted[left] != nil {
			leftName = sorted[left].Name
		}
		rightName := ""
		if sorted[right] != nil {
			rightName = sorted[right].Name
		}
		return strings.ToLower(strings.TrimSpace(leftName)) <
			strings.ToLower(strings.TrimSpace(rightName))
	})
	lines := make([]string, 0, len(sorted))
	for _, archetype := range sorted {
		if archetype == nil {
			continue
		}
		name := strings.TrimSpace(archetype.Name)
		if name == "" {
			continue
		}
		included := make([]string, 0, len(archetype.IncludedTypes))
		for _, placeType := range archetype.IncludedTypes {
			trimmed := strings.TrimSpace(string(placeType))
			if trimmed != "" {
				included = append(included, trimmed)
			}
		}
		if len(included) == 0 {
			lines = append(lines, fmt.Sprintf("- %s", name))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s (types: %s)", name, strings.Join(included, ", ")))
	}
	if len(lines) == 0 {
		return "none"
	}
	return strings.Join(lines, "\n")
}

func buildQuestArchetypeSuggestionPresetHintLocationArchetypePrompt(
	rawIDs []string,
	locationArchetypes []*models.LocationArchetype,
) string {
	if len(rawIDs) == 0 {
		return "none"
	}
	byID := make(map[string]string, len(locationArchetypes))
	for _, archetype := range locationArchetypes {
		if archetype == nil {
			continue
		}
		byID[archetype.ID.String()] = strings.TrimSpace(archetype.Name)
	}
	names := make([]string, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		if name := strings.TrimSpace(byID[strings.TrimSpace(rawID)]); name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}

func buildQuestArchetypeSuggestionPresetLocationArchetypeIDs(
	names []string,
	locationArchetypes []*models.LocationArchetype,
) models.StringArray {
	if len(names) == 0 || len(locationArchetypes) == 0 {
		return models.StringArray{}
	}
	byName := make(map[string]string, len(locationArchetypes))
	for _, archetype := range locationArchetypes {
		if archetype == nil {
			continue
		}
		name := normalizeQuestArchetypeSuggestionPresetLocationArchetypeName(archetype.Name)
		if name == "" {
			continue
		}
		byName[name] = archetype.ID.String()
	}
	out := models.StringArray{}
	seen := map[string]struct{}{}
	for _, rawName := range names {
		normalized := normalizeQuestArchetypeSuggestionPresetLocationArchetypeName(rawName)
		if normalized == "" {
			continue
		}
		id, exists := byName[normalized]
		if !exists {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func normalizeQuestArchetypeSuggestionPresetLocationArchetypeName(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func renderQuestArchetypeSuggestionPresetCountHint(count int) string {
	if count <= 0 {
		return "none"
	}
	return fmt.Sprintf("%d", count)
}

func renderQuestArchetypeSuggestionPresetZoneKindHint(
	zoneKind *models.ZoneKind,
	raw string,
) string {
	if zoneKind != nil {
		label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
		slug := strings.TrimSpace(models.ZoneKindPromptSlug(zoneKind))
		if label != "" && slug != "" {
			return fmt.Sprintf("%s (%s)", label, slug)
		}
		if slug != "" {
			return slug
		}
	}
	normalized := models.NormalizeZoneKind(raw)
	if normalized == "" {
		return "none"
	}
	return normalized
}

func renderQuestArchetypeSuggestionPresetZoneKindSeedHint(zoneKind *models.ZoneKind) string {
	seed := strings.TrimSpace(models.ZoneKindPromptSeed(zoneKind))
	if seed == "" {
		return "none"
	}
	return seed
}

func renderQuestArchetypeSuggestionPresetQuotedHint(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}

func renderQuestArchetypeSuggestionPresetTagHint(tags []string) string {
	normalized := normalizeQuestTemplateInternalTags(tags)
	if len(normalized) == 0 {
		return "none"
	}
	return strings.Join(normalized, ", ")
}

func renderQuestArchetypeSuggestionPresetFamilyMixTargets(
	targets map[string]int,
) string {
	normalized := models.NormalizeQuestArchetypeSuggestionFamilyMixTargets(targets)
	if len(normalized) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(normalized))
	for _, slug := range models.QuestArchetypeSuggestionKnownFamilySlugs() {
		count := normalized[slug]
		if count <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s x%d", slug, count))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func trimQuestArchetypeSuggestionPresetFamilyMixTargets(
	targets models.QuestArchetypeSuggestionFamilyMixTargets,
	limit int,
) models.QuestArchetypeSuggestionFamilyMixTargets {
	if limit <= 0 || len(targets) == 0 {
		return models.QuestArchetypeSuggestionFamilyMixTargets{}
	}
	remaining := limit
	out := models.QuestArchetypeSuggestionFamilyMixTargets{}
	for _, slug := range models.QuestArchetypeSuggestionKnownFamilySlugs() {
		if remaining <= 0 {
			break
		}
		count := targets[slug]
		if count <= 0 {
			continue
		}
		if count > remaining {
			count = remaining
		}
		out[slug] = count
		remaining -= count
	}
	return out
}

func deriveQuestArchetypeSuggestionPresetFamilyTags(
	targets models.QuestArchetypeSuggestionFamilyMixTargets,
) models.StringArray {
	out := models.StringArray{}
	for _, slug := range models.QuestArchetypeSuggestionKnownFamilySlugs() {
		if targets[slug] <= 0 {
			continue
		}
		out = append(out, slug)
		if len(out) >= 4 {
			break
		}
	}
	return out
}

func deriveQuestArchetypeSuggestionPresetInternalTags(
	targets models.QuestArchetypeSuggestionFamilyMixTargets,
	familyTags []string,
) models.StringArray {
	out := models.StringArray{}
	seen := map[string]struct{}{}
	for _, slug := range models.QuestArchetypeSuggestionKnownFamilySlugs() {
		if targets[slug] <= 0 {
			continue
		}
		out = append(out, slug)
		seen[slug] = struct{}{}
		if len(out) >= 3 {
			return out
		}
	}
	for _, tag := range normalizeQuestTemplateInternalTags(familyTags) {
		if _, exists := seen[tag]; exists {
			continue
		}
		out = append(out, tag)
		seen[tag] = struct{}{}
		if len(out) >= 4 {
			break
		}
	}
	return out
}

func buildQuestArchetypeSuggestionPresetFallbackThemePrompt(
	zoneKind *models.ZoneKind,
	zoneKindSlug string,
	familyMixTargets models.QuestArchetypeSuggestionFamilyMixTargets,
	familyTags []string,
	internalTags []string,
) string {
	families := deriveQuestArchetypeSuggestionPresetFamilyTags(familyMixTargets)
	if len(families) == 0 {
		families = normalizeQuestTemplateInternalTags(familyTags)
	}
	if len(families) == 0 {
		families = normalizeQuestTemplateInternalTags(internalTags)
	}
	parts := make([]string, 0, len(families))
	for _, family := range families {
		trimmed := strings.ReplaceAll(strings.TrimSpace(family), "_", " ")
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
		if len(parts) >= 3 {
			break
		}
	}
	familyText := "varied street-level conflicts"
	if len(parts) > 0 {
		familyText = strings.Join(parts, ", ")
	}
	if strings.TrimSpace(zoneKindSlug) != "" {
		base := fmt.Sprintf(
			"Reusable %s quests shaped by the flavor and pressures of %s zones.",
			familyText,
			strings.ReplaceAll(strings.TrimSpace(zoneKindSlug), "-", " "),
		)
		return appendQuestArchetypeSuggestionPresetSentence(
			base,
			buildQuestArchetypeSuggestionPresetZoneKindThemeSentence(zoneKind, zoneKindSlug),
		)
	}
	return fmt.Sprintf(
		"Reusable urban fantasy quests built around %s with strong route texture and social pressure.",
		familyText,
	)
}

func ensureQuestArchetypeSuggestionPresetThemeMatchesZoneKind(
	themePrompt string,
	zoneKind *models.ZoneKind,
	zoneKindSlug string,
) string {
	trimmed := strings.TrimSpace(themePrompt)
	if trimmed == "" {
		return ""
	}
	if questArchetypeSuggestionPresetThemeReflectsZoneKind(trimmed, zoneKind, zoneKindSlug) {
		return trimmed
	}
	return appendQuestArchetypeSuggestionPresetSentence(
		trimmed,
		buildQuestArchetypeSuggestionPresetZoneKindThemeSentence(zoneKind, zoneKindSlug),
	)
}

func questArchetypeSuggestionPresetThemeReflectsZoneKind(
	themePrompt string,
	zoneKind *models.ZoneKind,
	zoneKindSlug string,
) bool {
	normalizedTheme := strings.ToLower(strings.TrimSpace(themePrompt))
	if normalizedTheme == "" {
		return false
	}

	label := strings.ToLower(strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind)))
	if label != "" && strings.Contains(normalizedTheme, label) {
		return true
	}

	slugPhrase := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(zoneKindSlug)), "-", " ")
	if slugPhrase != "" && strings.Contains(normalizedTheme, slugPhrase) {
		return true
	}

	themeTokens := tokenizeQuestArchetypeSuggestionPresetThemeText(themePrompt)
	if len(themeTokens) == 0 {
		return false
	}
	themeTokenSet := make(map[string]struct{}, len(themeTokens))
	for _, token := range themeTokens {
		themeTokenSet[token] = struct{}{}
	}

	matched := 0
	for _, cue := range buildQuestArchetypeSuggestionPresetZoneKindCueTokens(zoneKind, zoneKindSlug) {
		if _, exists := themeTokenSet[cue]; !exists {
			continue
		}
		matched++
		if matched >= 2 {
			return true
		}
	}

	return false
}

func buildQuestArchetypeSuggestionPresetZoneKindCueTokens(
	zoneKind *models.ZoneKind,
	zoneKindSlug string,
) []string {
	candidates := []string{
		models.ZoneKindPromptLabel(zoneKind),
		zoneKindSlug,
	}
	if zoneKind != nil {
		candidates = append(candidates, zoneKind.Description)
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 8)
	for _, candidate := range candidates {
		for _, token := range tokenizeQuestArchetypeSuggestionPresetThemeText(candidate) {
			if _, exists := seen[token]; exists {
				continue
			}
			seen[token] = struct{}{}
			out = append(out, token)
		}
	}
	return out
}

func tokenizeQuestArchetypeSuggestionPresetThemeText(raw string) []string {
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(raw)), func(char rune) bool {
		return !unicode.IsLetter(char) && !unicode.IsNumber(char)
	})
	if len(parts) == 0 {
		return nil
	}

	stopwords := map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "built": {},
		"by": {}, "for": {}, "from": {}, "in": {}, "into": {}, "it": {}, "its": {}, "of": {},
		"on": {}, "or": {}, "that": {}, "the": {}, "their": {}, "them": {}, "these": {},
		"this": {}, "to": {}, "under": {}, "with": {}, "zones": {}, "zone": {},
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) < 4 {
			continue
		}
		if _, stopword := stopwords[part]; stopword {
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

func buildQuestArchetypeSuggestionPresetZoneKindThemeSentence(
	zoneKind *models.ZoneKind,
	zoneKindSlug string,
) string {
	label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	if label == "" {
		label = strings.ReplaceAll(strings.TrimSpace(zoneKindSlug), "-", " ")
	}
	if label == "" {
		return ""
	}

	description := ""
	if zoneKind != nil {
		description = strings.TrimSpace(zoneKind.Description)
	}
	if description != "" {
		return fmt.Sprintf(
			"Make it unmistakably suited to %s zones: %s",
			strings.ToLower(label),
			description,
		)
	}

	return fmt.Sprintf(
		"Make it unmistakably suited to %s zones, with routes, landmarks, and local tensions that only that environment could support.",
		strings.ToLower(label),
	)
}

func appendQuestArchetypeSuggestionPresetSentence(base string, extra string) string {
	base = strings.TrimSpace(base)
	extra = strings.TrimSpace(extra)
	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}
	if strings.EqualFold(base, extra) || strings.Contains(strings.ToLower(base), strings.ToLower(extra)) {
		return base
	}
	last := base[len(base)-1]
	if last != '.' && last != '!' && last != '?' {
		base += "."
	}
	return base + " " + extra
}
