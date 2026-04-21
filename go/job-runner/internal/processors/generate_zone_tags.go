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
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	zoneNeighborhoodFlavorTagGroupName = "zone_neighborhood_flavor"
	zoneNeighborhoodFlavorTagPoolSize  = 100
	zoneNeighborhoodSelectedTagCount   = 5
)

const zoneTagPromptTemplate = `
You are choosing internal neighborhood flavor tags for Unclaimed Streets, an urban fantasy MMORPG.

Zone:
- current name: %s
- current description: %s
- current internal tags: %s

Geometry context:
%s

Neighborhood evidence:
%s

Candidate shared tags (choose exactly %d and do not invent new ones):
%s

Return JSON only:
{
  "neighborhoodSummary": "2-4 sentences",
  "selectedTags": ["tag_one", "tag_two", "tag_three", "tag_four", "tag_five"]
}

Rules:
- selectedTags must contain exactly %d unique tags.
- Every selected tag must be copied exactly from the candidate tag list.
- Pick tags that describe the district's flavor, land use, social texture, activity, and atmosphere.
- Use the zone description, geometry, and points of interest as evidence.
- Prefer a balanced set instead of near-synonyms.
- Keep neighborhoodSummary grounded, vivid, and useful for internal worldbuilding.
- Do not output markdown or commentary outside the JSON object.
`

type zoneTagGenerationResponse struct {
	NeighborhoodSummary string   `json:"neighborhoodSummary"`
	SelectedTags        []string `json:"selectedTags"`
}

type GenerateZoneTagsProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateZoneTagsProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateZoneTagsProcessor {
	log.Println("Initializing GenerateZoneTagsProcessor")
	return GenerateZoneTagsProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateZoneTagsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate zone tags task: %s", task.Type())

	var payload jobs.GenerateZoneTagsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ZoneTagGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Zone tag generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ZoneTagGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneTagGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateTags(ctx, job); err != nil {
		return p.failZoneTagGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateZoneTagsProcessor) generateTags(ctx context.Context, job *models.ZoneTagGenerationJob) error {
	zone, err := p.dbClient.Zone().FindByID(ctx, job.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to load zone: %w", err)
	}
	if zone == nil {
		return fmt.Errorf("zone not found")
	}

	tagGroup, err := p.dbClient.TagGroup().FindByName(ctx, zoneNeighborhoodFlavorTagGroupName)
	if err != nil {
		return fmt.Errorf("failed to load zone neighborhood flavor tag group: %w", err)
	}
	if tagGroup == nil {
		return fmt.Errorf("zone neighborhood flavor tag group %q not found", zoneNeighborhoodFlavorTagGroupName)
	}

	groupTags, err := p.dbClient.Tag().FindByGroupID(ctx, tagGroup.ID)
	if err != nil {
		return fmt.Errorf("failed to load shared zone tags: %w", err)
	}

	candidateTags := buildZoneTagCandidatePool(groupTags, zoneNeighborhoodFlavorTagPoolSize)
	if len(candidateTags) < zoneNeighborhoodFlavorTagPoolSize {
		return fmt.Errorf(
			"zone neighborhood flavor tag pool expected %d tags, found %d",
			zoneNeighborhoodFlavorTagPoolSize,
			len(candidateTags),
		)
	}

	zoneName := strings.TrimSpace(zone.Name)
	if zoneName == "" {
		zoneName = "Unknown Zone"
	}
	currentDescription := strings.TrimSpace(zone.Description)
	if currentDescription == "" {
		currentDescription = "none"
	}
	currentTags := renderTagList(zone.InternalTags)
	geometrySummary := buildZoneFlavorGeometrySummary(*zone)
	neighborhoodEvidence, err := buildZoneNeighborhoodEvidence(ctx, p.dbClient, zone.ID)
	if err != nil {
		return fmt.Errorf("failed to load neighborhood evidence: %w", err)
	}
	contextSnapshot := fmt.Sprintf(
		"Zone:\n- current name: %s\n- current description: %s\n- current internal tags: %s\n\nGeometry context:\n%s\n\nNeighborhood evidence:\n%s",
		zoneName,
		currentDescription,
		currentTags,
		geometrySummary,
		neighborhoodEvidence,
	)
	job.ContextSnapshot = contextSnapshot
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneTagGenerationJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to persist zone tag generation context: %w", err)
	}

	prompt := fmt.Sprintf(
		zoneTagPromptTemplate,
		zoneName,
		currentDescription,
		currentTags,
		geometrySummary,
		neighborhoodEvidence,
		zoneNeighborhoodSelectedTagCount,
		renderTagList(candidateTags),
		zoneNeighborhoodSelectedTagCount,
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate zone tags: %w", err)
	}

	var generated zoneTagGenerationResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
		return fmt.Errorf("failed to parse generated zone tag payload: %w", err)
	}

	selectedTags, err := sanitizeSelectedZoneTags(generated.SelectedTags, candidateTags, zoneNeighborhoodSelectedTagCount)
	if err != nil {
		return err
	}
	summary := sanitizeZoneNeighborhoodSummary(generated.NeighborhoodSummary)
	if summary == "" {
		return fmt.Errorf("generated neighborhood summary was empty")
	}

	nextTags := applyGeneratedZoneTags(zone.InternalTags, selectedTags, candidateTags)
	if _, err := p.dbClient.Zone().UpdateMetadata(ctx, zone.ID, zone.Name, zone.Description, zone.Kind, nextTags); err != nil {
		return fmt.Errorf("failed to update zone tags: %w", err)
	}

	job.Status = models.ZoneTagGenerationStatusCompleted
	job.GeneratedSummary = &summary
	job.SelectedTags = models.StringArray(selectedTags)
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneTagGenerationJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update zone tag generation job: %w", err)
	}

	return nil
}

func buildZoneTagCandidatePool(tags []*models.Tag, limit int) []string {
	if limit <= 0 {
		limit = len(tags)
	}

	seen := map[string]struct{}{}
	candidateTags := make([]string, 0, minIntZoneTags(limit, len(tags)))
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		normalized := strings.ToLower(strings.TrimSpace(tag.Value))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		candidateTags = append(candidateTags, normalized)
		if len(candidateTags) >= limit {
			break
		}
	}

	return candidateTags
}

func buildZoneNeighborhoodEvidence(ctx context.Context, dbClient db.DbClient, zoneID uuid.UUID) (string, error) {
	pointsOfInterest, err := dbClient.PointOfInterest().FindAllForZone(ctx, zoneID)
	if err != nil {
		return "", err
	}
	if len(pointsOfInterest) == 0 {
		return "- no linked points of interest are available for this zone yet", nil
	}

	sort.Slice(pointsOfInterest, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(pointsOfInterest[i].Name))
		right := strings.ToLower(strings.TrimSpace(pointsOfInterest[j].Name))
		if left == right {
			return pointsOfInterest[i].CreatedAt.Before(pointsOfInterest[j].CreatedAt)
		}
		return left < right
	})

	tagCounts := map[string]int{}
	for _, poi := range pointsOfInterest {
		for _, tag := range poi.Tags {
			normalized := strings.ToLower(strings.TrimSpace(tag.Value))
			if normalized == "" {
				continue
			}
			tagCounts[normalized]++
		}
	}

	lines := []string{
		fmt.Sprintf("- point of interest count: %d", len(pointsOfInterest)),
	}
	if topTags := renderFrequentPOITags(tagCounts, 10); topTags != "" {
		lines = append(lines, fmt.Sprintf("- recurring point-of-interest tags: %s", topTags))
	}

	maxPoints := minIntZoneTags(len(pointsOfInterest), 12)
	for i := 0; i < maxPoints; i++ {
		poi := pointsOfInterest[i]
		name := firstNonEmpty(
			strings.TrimSpace(poi.Name),
			strings.TrimSpace(poi.OriginalName),
			strings.TrimSpace(stringPointerValue(poi.GoogleMapsPlaceName)),
			"unnamed point of interest",
		)
		tagValues := make([]string, 0, len(poi.Tags))
		for _, tag := range poi.Tags {
			normalized := strings.ToLower(strings.TrimSpace(tag.Value))
			if normalized == "" {
				continue
			}
			tagValues = append(tagValues, normalized)
		}
		sort.Strings(tagValues)
		tagSummary := "none"
		if len(tagValues) > 0 {
			tagSummary = strings.Join(tagValues, ", ")
		}
		description := firstNonEmpty(strings.TrimSpace(poi.Description), strings.TrimSpace(poi.Clue), "no extra description")
		lines = append(lines, fmt.Sprintf(
			"- %s | tags: %s | flavor: %s",
			name,
			tagSummary,
			trimForPrompt(description, 140),
		))
	}

	return strings.Join(lines, "\n"), nil
}

func renderFrequentPOITags(tagCounts map[string]int, limit int) string {
	type tagCount struct {
		tag   string
		count int
	}
	pairs := make([]tagCount, 0, len(tagCounts))
	for tag, count := range tagCounts {
		pairs = append(pairs, tagCount{tag: tag, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].tag < pairs[j].tag
		}
		return pairs[i].count > pairs[j].count
	})
	if limit > 0 && len(pairs) > limit {
		pairs = pairs[:limit]
	}
	if len(pairs) == 0 {
		return ""
	}
	rendered := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		rendered = append(rendered, fmt.Sprintf("%s (%d)", pair.tag, pair.count))
	}
	return strings.Join(rendered, ", ")
}

func sanitizeSelectedZoneTags(rawTags []string, candidateTags []string, requiredCount int) ([]string, error) {
	if requiredCount <= 0 {
		requiredCount = zoneNeighborhoodSelectedTagCount
	}

	allowed := map[string]struct{}{}
	allowedAliases := map[string]string{}
	for _, candidate := range candidateTags {
		allowed[candidate] = struct{}{}
		allowedAliases[simplifyTagKey(candidate)] = candidate
	}

	selected := make([]string, 0, len(rawTags))
	seen := map[string]struct{}{}
	for _, raw := range rawTags {
		normalized := normalizeGeneratedTag(raw)
		if normalized == "" {
			continue
		}
		if _, ok := allowed[normalized]; !ok {
			aliasMatch, hasAlias := allowedAliases[simplifyTagKey(normalized)]
			if !hasAlias {
				continue
			}
			normalized = aliasMatch
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		selected = append(selected, normalized)
		if len(selected) >= requiredCount {
			break
		}
	}

	if len(selected) != requiredCount {
		return nil, fmt.Errorf("generated zone tags did not include %d valid shared tags", requiredCount)
	}

	return selected, nil
}

func normalizeGeneratedTag(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.Join(strings.Fields(normalized), "_")
	return normalized
}

func simplifyTagKey(raw string) string {
	return strings.ReplaceAll(normalizeGeneratedTag(raw), "_", "")
}

func applyGeneratedZoneTags(existing models.StringArray, generated []string, candidatePool []string) models.StringArray {
	poolSet := map[string]struct{}{}
	for _, tag := range candidatePool {
		poolSet[tag] = struct{}{}
	}

	merged := make([]string, 0, len(existing)+len(generated))
	seen := map[string]struct{}{}
	appendTag := func(raw string) {
		normalized := strings.ToLower(strings.TrimSpace(raw))
		if normalized == "" {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}
		seen[normalized] = struct{}{}
		merged = append(merged, normalized)
	}

	for _, tag := range existing {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, isGeneratedPoolTag := poolSet[normalized]; isGeneratedPoolTag {
			continue
		}
		appendTag(normalized)
	}
	for _, tag := range generated {
		appendTag(tag)
	}

	return models.StringArray(merged)
}

func sanitizeZoneNeighborhoodSummary(raw string) string {
	summary := strings.TrimSpace(raw)
	summary = strings.Trim(summary, "\"")
	return strings.Join(strings.Fields(summary), " ")
}

func trimForPrompt(value string, maxLen int) string {
	trimmed := strings.TrimSpace(value)
	if maxLen <= 0 || len(trimmed) <= maxLen {
		return trimmed
	}
	if maxLen <= 3 {
		return trimmed[:maxLen]
	}
	return strings.TrimSpace(trimmed[:maxLen-3]) + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func minIntZoneTags(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *GenerateZoneTagsProcessor) failZoneTagGenerationJob(
	ctx context.Context,
	job *models.ZoneTagGenerationJob,
	err error,
) error {
	if job == nil {
		return err
	}
	errMsg := err.Error()
	job.Status = models.ZoneTagGenerationStatusFailed
	job.ErrorMessage = &errMsg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ZoneTagGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark zone tag generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}
