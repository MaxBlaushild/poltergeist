package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ProcessMainStoryDistrictRunProcessor struct {
	dbClient      db.DbClient
	dungeonmaster dungeonmaster.Client
}

func NewProcessMainStoryDistrictRunProcessor(
	dbClient db.DbClient,
	dungeonmaster dungeonmaster.Client,
) ProcessMainStoryDistrictRunProcessor {
	return ProcessMainStoryDistrictRunProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmaster,
	}
}

func (p *ProcessMainStoryDistrictRunProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.ProcessMainStoryDistrictRunTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	run, err := p.dbClient.MainStoryDistrictRun().FindByID(ctx, payload.RunID)
	if err != nil {
		return err
	}
	if run == nil {
		return nil
	}
	if run.Status == models.MainStoryDistrictRunStatusCompleted {
		return nil
	}

	template, err := p.dbClient.MainStoryTemplate().FindByID(ctx, run.MainStoryTemplateID)
	if err != nil {
		return p.failRun(ctx, run, err)
	}
	if template == nil {
		return p.failRun(ctx, run, fmt.Errorf("main story template not found"))
	}

	district, err := p.dbClient.District().FindByID(ctx, run.DistrictID)
	if err != nil {
		return p.failRun(ctx, run, err)
	}
	if district == nil {
		return p.failRun(ctx, run, fmt.Errorf("district not found"))
	}
	if len(district.Zones) == 0 {
		return p.failRun(ctx, run, fmt.Errorf("district has no child zones"))
	}

	targetZone, err := resolveProcessMainStoryTargetZone(district, run.ZoneID)
	if err != nil {
		return p.failRun(ctx, run, err)
	}

	run.Status = models.MainStoryDistrictRunStatusInProgress
	run.ErrorMessage = nil
	run.UpdatedAt = time.Now()
	if err := p.dbClient.MainStoryDistrictRun().Update(ctx, run); err != nil {
		return err
	}

	if len(run.BeatRuns) == 0 && len(run.GeneratedCharacterIDs) == 0 {
		if err := p.materializeRun(ctx, template, district, targetZone, run); err != nil {
			return p.failRun(ctx, run, err)
		}
	} else {
		if err := p.resumeRun(ctx, template, district, targetZone, run); err != nil {
			return p.failRun(ctx, run, err)
		}
	}

	run.Status = models.MainStoryDistrictRunStatusCompleted
	run.ErrorMessage = nil
	run.UpdatedAt = time.Now()
	return p.dbClient.MainStoryDistrictRun().Update(ctx, run)
}

func (p *ProcessMainStoryDistrictRunProcessor) failRun(
	ctx context.Context,
	run *models.MainStoryDistrictRun,
	err error,
) error {
	if run != nil {
		msg := err.Error()
		run.Status = models.MainStoryDistrictRunStatusFailed
		run.ErrorMessage = &msg
		run.UpdatedAt = time.Now()
		if updateErr := p.dbClient.MainStoryDistrictRun().Update(ctx, run); updateErr != nil {
			return fmt.Errorf("main story district run failed: %v (also failed to persist state: %w)", err, updateErr)
		}
	}
	return err
}

func (p *ProcessMainStoryDistrictRunProcessor) materializeRun(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	targetZone *models.Zone,
	run *models.MainStoryDistrictRun,
) error {
	if template == nil {
		return fmt.Errorf("template is required")
	}
	if district == nil {
		return fmt.Errorf("district is required")
	}
	if run == nil {
		return fmt.Errorf("run is required")
	}
	if len(template.Beats) == 0 {
		return fmt.Errorf("main story template has no beats")
	}

	clonedCharactersByKey, generatedCharacterIDs, err := p.cloneRunCharacters(ctx, template, district, targetZone)
	if err != nil {
		return err
	}
	run.GeneratedCharacterIDs = generatedCharacterIDs

	beatRuns := make(models.MainStoryDistrictBeatRuns, 0, len(template.Beats))
	createdQuests := make([]*models.Quest, 0, len(template.Beats))
	for _, beat := range template.Beats {
		beatRun, quest, err := p.materializeBeatRun(ctx, template, district, targetZone, beat, clonedCharactersByKey, nil)
		beatRuns = append(beatRuns, beatRun)
		run.BeatRuns = beatRuns
		run.UpdatedAt = time.Now()
		if updateErr := p.dbClient.MainStoryDistrictRun().Update(ctx, run); updateErr != nil {
			return updateErr
		}
		if err != nil {
			if linkErr := p.linkRunQuests(ctx, createdQuests); linkErr != nil {
				return linkErr
			}
			return err
		}
		if quest != nil {
			createdQuests = append(createdQuests, quest)
		}
	}

	if err := p.linkRunQuests(ctx, createdQuests); err != nil {
		return err
	}
	run.BeatRuns = beatRuns
	return nil
}

func (p *ProcessMainStoryDistrictRunProcessor) cloneRunCharacters(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	targetZone *models.Zone,
) (map[string]*models.Character, models.StringArray, error) {
	clonedByKey := make(map[string]*models.Character)
	generatedIDs := make(models.StringArray, 0)
	sourceCharactersByID := make(map[uuid.UUID]*models.Character)

	for _, beat := range template.Beats {
		key := normalizeProcessMainStoryCharacterKey(beat.QuestGiverCharacterKey)
		if key == "" {
			continue
		}
		if _, exists := clonedByKey[key]; exists {
			continue
		}

		var sourceCharacter *models.Character
		if beat.QuestGiverCharacterID != nil && *beat.QuestGiverCharacterID != uuid.Nil {
			if cached, exists := sourceCharactersByID[*beat.QuestGiverCharacterID]; exists {
				sourceCharacter = cached
			} else {
				sourceCharacter, _ = p.dbClient.Character().FindByID(ctx, *beat.QuestGiverCharacterID)
				if sourceCharacter != nil {
					sourceCharactersByID[sourceCharacter.ID] = sourceCharacter
				}
			}
		}

		name := strings.TrimSpace(beat.QuestGiverCharacterName)
		description := ""
		if sourceCharacter != nil {
			if name == "" {
				name = strings.TrimSpace(sourceCharacter.Name)
			}
			description = strings.TrimSpace(sourceCharacter.Description)
		}
		if name == "" {
			name = strings.TrimSpace(beat.QuestGiverCharacterKey)
		}
		if description == "" {
			description = strings.TrimSpace(beat.Description)
		}

		character := &models.Character{
			ID:                    uuid.New(),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			Name:                  name,
			Description:           description,
			InternalTags:          buildProcessMainStoryCharacterTags(template, district, targetZone, key, sourceCharacter),
			ImageGenerationStatus: models.CharacterImageGenerationStatusNone,
		}
		if sourceCharacter != nil {
			character.GenreID = sourceCharacter.GenreID
			character.Genre = sourceCharacter.Genre
			character.MapIconURL = sourceCharacter.MapIconURL
			character.DialogueImageURL = sourceCharacter.DialogueImageURL
			character.ThumbnailURL = sourceCharacter.ThumbnailURL
			character.StoryVariants = append(models.CharacterStoryVariants{}, sourceCharacter.StoryVariants...)
			if strings.TrimSpace(sourceCharacter.DialogueImageURL) != "" ||
				strings.TrimSpace(sourceCharacter.ThumbnailURL) != "" ||
				strings.TrimSpace(sourceCharacter.MapIconURL) != "" {
				character.ImageGenerationStatus = models.CharacterImageGenerationStatusComplete
			}
		}
		if err := p.dbClient.Character().Create(ctx, character); err != nil {
			return nil, nil, err
		}
		clonedByKey[key] = character
		generatedIDs = append(generatedIDs, character.ID.String())
	}

	return clonedByKey, generatedIDs, nil
}

func (p *ProcessMainStoryDistrictRunProcessor) loadRunCharacters(
	ctx context.Context,
	run *models.MainStoryDistrictRun,
) (map[string]*models.Character, error) {
	clonedByKey := make(map[string]*models.Character)
	if run == nil {
		return clonedByKey, fmt.Errorf("run is required")
	}

	for _, rawCharacterID := range run.GeneratedCharacterIDs {
		characterID, err := uuid.Parse(strings.TrimSpace(rawCharacterID))
		if err != nil || characterID == uuid.Nil {
			continue
		}
		character, err := p.dbClient.Character().FindByID(ctx, characterID)
		if err != nil {
			return nil, err
		}
		if character == nil {
			continue
		}
		for _, rawTag := range []string(character.InternalTags) {
			tag := strings.TrimSpace(strings.ToLower(rawTag))
			if !strings.HasPrefix(tag, "story_character_") {
				continue
			}
			key := normalizeProcessMainStoryCharacterKey(strings.TrimPrefix(tag, "story_character_"))
			if key == "" {
				continue
			}
			clonedByKey[key] = character
		}
	}

	return clonedByKey, nil
}

func findFirstIncompleteProcessMainStoryBeatIndex(
	run *models.MainStoryDistrictRun,
	template *models.MainStoryTemplate,
) int {
	if run == nil || template == nil {
		return -1
	}
	for index, beat := range template.Beats {
		if index >= len(run.BeatRuns) {
			return index
		}
		if run.BeatRuns[index].Status != models.MainStoryDistrictRunStatusCompleted {
			return index
		}
		if run.BeatRuns[index].QuestID == nil || *run.BeatRuns[index].QuestID == uuid.Nil {
			return index
		}
		if beat.QuestArchetypeID != nil &&
			run.BeatRuns[index].QuestArchetypeID != nil &&
			*beat.QuestArchetypeID != *run.BeatRuns[index].QuestArchetypeID {
			return index
		}
	}
	return -1
}

func (p *ProcessMainStoryDistrictRunProcessor) loadRunQuests(
	ctx context.Context,
	beatRuns models.MainStoryDistrictBeatRuns,
) ([]*models.Quest, error) {
	quests := make([]*models.Quest, 0, len(beatRuns))
	for _, beatRun := range beatRuns {
		if beatRun.QuestID == nil || *beatRun.QuestID == uuid.Nil {
			continue
		}
		quest, err := p.dbClient.Quest().FindByID(ctx, *beatRun.QuestID)
		if err != nil {
			return nil, err
		}
		if quest != nil {
			quests = append(quests, quest)
		}
	}
	return quests, nil
}

func (p *ProcessMainStoryDistrictRunProcessor) resumeRun(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	targetZone *models.Zone,
	run *models.MainStoryDistrictRun,
) error {
	if template == nil {
		return fmt.Errorf("template is required")
	}
	if district == nil {
		return fmt.Errorf("district is required")
	}
	if run == nil {
		return fmt.Errorf("run is required")
	}

	retryStartIndex := findFirstIncompleteProcessMainStoryBeatIndex(run, template)
	if retryStartIndex < 0 {
		return fmt.Errorf("main story district run has no failed or incomplete beats to retry")
	}

	clonedCharactersByKey, err := p.loadRunCharacters(ctx, run)
	if err != nil {
		return err
	}

	var failedZoneID *uuid.UUID
	if targetZone == nil && retryStartIndex < len(run.BeatRuns) {
		failedZoneID = run.BeatRuns[retryStartIndex].ZoneID
	}

	for index := len(run.BeatRuns) - 1; index >= retryStartIndex; index-- {
		questID := run.BeatRuns[index].QuestID
		if questID == nil || *questID == uuid.Nil {
			continue
		}
		if err := p.dbClient.Quest().Delete(ctx, *questID); err != nil {
			return fmt.Errorf("failed to delete partially created quest %s: %w", questID.String(), err)
		}
	}

	completedBeatRuns := append(models.MainStoryDistrictBeatRuns{}, run.BeatRuns[:retryStartIndex]...)
	run.BeatRuns = completedBeatRuns
	run.ErrorMessage = nil
	run.UpdatedAt = time.Now()

	completedQuests, err := p.loadRunQuests(ctx, completedBeatRuns)
	if err != nil {
		return err
	}
	if err := p.linkRunQuests(ctx, completedQuests); err != nil {
		return err
	}
	if err := p.dbClient.MainStoryDistrictRun().Update(ctx, run); err != nil {
		return err
	}

	createdQuests := append([]*models.Quest{}, completedQuests...)
	for beatIndex := retryStartIndex; beatIndex < len(template.Beats); beatIndex++ {
		deprioritizedZoneID := failedZoneID
		if beatIndex > retryStartIndex {
			deprioritizedZoneID = nil
		}
		beatRun, quest, err := p.materializeBeatRun(
			ctx,
			template,
			district,
			targetZone,
			template.Beats[beatIndex],
			clonedCharactersByKey,
			deprioritizedZoneID,
		)
		run.BeatRuns = append(run.BeatRuns, beatRun)
		run.UpdatedAt = time.Now()
		if updateErr := p.dbClient.MainStoryDistrictRun().Update(ctx, run); updateErr != nil {
			return updateErr
		}
		if err != nil {
			if linkErr := p.linkRunQuests(ctx, createdQuests); linkErr != nil {
				return linkErr
			}
			return err
		}
		if quest != nil {
			createdQuests = append(createdQuests, quest)
		}
	}

	return p.linkRunQuests(ctx, createdQuests)
}

func buildProcessMainStoryCharacterTags(
	template *models.MainStoryTemplate,
	district *models.District,
	targetZone *models.Zone,
	characterKey string,
	sourceCharacter *models.Character,
) models.StringArray {
	tags := []string{
		"main_story",
		"main_story_run",
		"story_character",
		"story_character_" + characterKey,
	}
	if targetZone != nil {
		tags = append(tags, "zone_main_story_run", "zone_"+targetZone.ID.String())
	} else {
		tags = append(tags, "district_main_story_run")
	}
	if template != nil {
		tags = append(tags, "main_story_template_"+template.ID.String())
		tags = append(tags, []string(template.InternalTags)...)
	}
	if district != nil {
		tags = append(tags, "district_"+district.ID.String())
	}
	if targetZone != nil {
		tags = append(tags, []string(targetZone.InternalTags)...)
	}
	if sourceCharacter != nil {
		tags = append(tags, []string(sourceCharacter.InternalTags)...)
	}
	return normalizeProcessMainStoryTags(tags)
}

func normalizeProcessMainStoryCharacterKey(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func (p *ProcessMainStoryDistrictRunProcessor) materializeBeatRun(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	targetZone *models.Zone,
	beat models.MainStoryBeatDraft,
	clonedCharactersByKey map[string]*models.Character,
	deprioritizedZoneID *uuid.UUID,
) (models.MainStoryDistrictBeatRun, *models.Quest, error) {
	beatRun := models.MainStoryDistrictBeatRun{
		OrderIndex:              beat.OrderIndex,
		Act:                     beat.Act,
		ChapterTitle:            strings.TrimSpace(beat.ChapterTitle),
		StoryRole:               strings.TrimSpace(beat.StoryRole),
		Status:                  models.MainStoryDistrictRunStatusInProgress,
		QuestArchetypeID:        beat.QuestArchetypeID,
		QuestArchetypeName:      strings.TrimSpace(beat.QuestArchetypeName),
		QuestGiverCharacterName: strings.TrimSpace(beat.QuestGiverCharacterName),
	}

	if beat.QuestArchetypeID == nil || *beat.QuestArchetypeID == uuid.Nil {
		beatRun.Status = models.MainStoryDistrictRunStatusFailed
		beatRun.ErrorMessage = "beat is missing a quest archetype"
		return beatRun, nil, fmt.Errorf("%s", beatRun.ErrorMessage)
	}
	questArchetype, err := p.dbClient.QuestArchetype().FindByID(ctx, *beat.QuestArchetypeID)
	if err != nil {
		beatRun.Status = models.MainStoryDistrictRunStatusFailed
		beatRun.ErrorMessage = fmt.Sprintf("failed to load quest archetype: %v", err)
		return beatRun, nil, err
	}
	if questArchetype == nil {
		beatRun.Status = models.MainStoryDistrictRunStatusFailed
		beatRun.ErrorMessage = "quest archetype not found"
		return beatRun, nil, fmt.Errorf("%s", beatRun.ErrorMessage)
	}

	character := clonedCharactersByKey[normalizeProcessMainStoryCharacterKey(beat.QuestGiverCharacterKey)]
	if character == nil {
		beatRun.Status = models.MainStoryDistrictRunStatusFailed
		beatRun.ErrorMessage = "beat quest giver could not be resolved"
		return beatRun, nil, fmt.Errorf("%s", beatRun.ErrorMessage)
	}
	beatRun.QuestGiverCharacterID = &character.ID
	if strings.TrimSpace(character.Name) != "" {
		beatRun.QuestGiverCharacterName = strings.TrimSpace(character.Name)
	}

	rankedZones := []processMainStoryZoneCandidate{}
	if targetZone != nil {
		rankedZones = append(rankedZones, processMainStoryZoneCandidate{zone: targetZone})
	} else {
		rankedZones, err = rankProcessMainStoryZones(district.Zones, beat, questArchetype, template, deprioritizedZoneID)
		if err != nil {
			beatRun.Status = models.MainStoryDistrictRunStatusFailed
			beatRun.ErrorMessage = err.Error()
			return beatRun, nil, err
		}
	}

	var lastErr error
	for zoneIndex, candidate := range rankedZones {
		if candidate.zone == nil {
			continue
		}

		beatRun.ZoneID = &candidate.zone.ID
		beatRun.ZoneName = strings.TrimSpace(candidate.zone.Name)

		pointOfInterest, err := p.ensureCharacterPointOfInterest(ctx, character, candidate.zone, beat)
		if err != nil {
			lastErr = err
			if shouldRetryProcessMainStoryBeatInAnotherZone(err) && zoneIndex+1 < len(rankedZones) {
				continue
			}
			break
		}
		if pointOfInterest != nil {
			beatRun.PointOfInterestID = &pointOfInterest.ID
			beatRun.PointOfInterestName = strings.TrimSpace(pointOfInterest.Name)
		}

		quest, err := p.dungeonmaster.GenerateQuest(ctx, candidate.zone, questArchetype.ID, &character.ID)
		if err == nil {
			if quest == nil {
				lastErr = fmt.Errorf("quest generation returned no quest")
				break
			}
			beatRun.QuestID = &quest.ID
			beatRun.QuestName = strings.TrimSpace(quest.Name)
			beatRun.Status = models.MainStoryDistrictRunStatusCompleted
			beatRun.ErrorMessage = ""
			return beatRun, quest, nil
		}

		lastErr = err
		if shouldRetryProcessMainStoryBeatInAnotherZone(err) && zoneIndex+1 < len(rankedZones) {
			continue
		}
		break
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("quest generation failed")
	}
	beatRun.Status = models.MainStoryDistrictRunStatusFailed
	beatRun.ErrorMessage = lastErr.Error()
	return beatRun, nil, lastErr
}

func (p *ProcessMainStoryDistrictRunProcessor) ensureCharacterPointOfInterest(
	ctx context.Context,
	character *models.Character,
	zone *models.Zone,
	beat models.MainStoryBeatDraft,
) (*models.PointOfInterest, error) {
	if character == nil {
		return nil, fmt.Errorf("character is required")
	}
	if zone == nil {
		return nil, fmt.Errorf("zone is required")
	}

	if character.PointOfInterestID != nil && *character.PointOfInterestID != uuid.Nil {
		pointOfInterest, err := p.dbClient.PointOfInterest().FindByID(ctx, *character.PointOfInterestID)
		if err == nil && pointOfInterest != nil {
			if pointOfInterestZone, zoneErr := p.dbClient.Zone().FindByPointOfInterestID(ctx, pointOfInterest.ID); zoneErr == nil &&
				pointOfInterestZone != nil &&
				pointOfInterestZone.ID == zone.ID {
				return pointOfInterest, nil
			}
		}
	}

	pointsOfInterest, err := p.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return nil, err
	}
	if len(pointsOfInterest) == 0 {
		return nil, fmt.Errorf("zone %s has no points of interest for quest giver placement", zone.ID)
	}
	hint := strings.Join([]string{
		strings.TrimSpace(beat.ChapterTitle),
		strings.TrimSpace(beat.Name),
		strings.TrimSpace(beat.Hook),
		strings.TrimSpace(beat.Description),
	}, " ")
	best := pickBestPointOfInterestForHintForProcess(pointsOfInterest, hint, nil)
	if best == nil {
		return nil, fmt.Errorf("no point of interest could be resolved for quest giver placement")
	}

	if err := p.dbClient.Character().UpdateFields(ctx, character.ID, map[string]interface{}{
		"point_of_interest_id": best.ID,
		"updated_at":           time.Now(),
	}); err != nil {
		return nil, err
	}
	character.PointOfInterestID = &best.ID
	character.PointOfInterest = best
	return best, nil
}

type processMainStoryZoneCandidate struct {
	zone       *models.Zone
	matchCount int
}

func resolveProcessMainStoryTargetZone(
	district *models.District,
	targetZoneID *uuid.UUID,
) (*models.Zone, error) {
	if targetZoneID == nil || *targetZoneID == uuid.Nil {
		return nil, nil
	}
	if district == nil {
		return nil, fmt.Errorf("district is required")
	}
	for index := range district.Zones {
		if district.Zones[index].ID == *targetZoneID {
			return &district.Zones[index], nil
		}
	}
	return nil, fmt.Errorf("zone %s is not part of district %s", targetZoneID.String(), district.ID.String())
}

func rankProcessMainStoryZones(
	zones []models.Zone,
	beat models.MainStoryBeatDraft,
	questArchetype *models.QuestArchetype,
	template *models.MainStoryTemplate,
	deprioritizedZoneID *uuid.UUID,
) ([]processMainStoryZoneCandidate, error) {
	if len(zones) == 0 {
		return nil, fmt.Errorf("district has no zones")
	}
	desiredTags := make(models.StringArray, 0)
	desiredTags = append(desiredTags, beat.RequiredZoneTags...)
	if len(desiredTags) == 0 && questArchetype != nil {
		desiredTags = append(desiredTags, questArchetype.InternalTags...)
	}
	if len(desiredTags) == 0 && template != nil {
		desiredTags = append(desiredTags, template.ThemeTags...)
	}
	normalized := normalizeProcessMainStoryTagSet(desiredTags)
	ranked := make([]processMainStoryZoneCandidate, 0, len(zones))
	for index := range zones {
		ranked = append(ranked, processMainStoryZoneCandidate{
			zone:       &zones[index],
			matchCount: processMainStoryTagMatchCount(zones[index].InternalTags, normalized),
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].matchCount != ranked[j].matchCount {
			return ranked[i].matchCount > ranked[j].matchCount
		}
		if deprioritizedZoneID != nil && *deprioritizedZoneID != uuid.Nil {
			iIsDeprioritized := ranked[i].zone != nil && ranked[i].zone.ID == *deprioritizedZoneID
			jIsDeprioritized := ranked[j].zone != nil && ranked[j].zone.ID == *deprioritizedZoneID
			if iIsDeprioritized != jIsDeprioritized {
				return !iIsDeprioritized
			}
		}
		return ranked[i].matchCount > ranked[j].matchCount
	})
	return ranked, nil
}

func processMainStoryTagMatchCount(tags models.StringArray, desired map[string]struct{}) int {
	if len(desired) == 0 {
		return 0
	}
	count := 0
	seen := map[string]struct{}{}
	for _, rawTag := range []string(tags) {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		if _, exists := desired[tag]; !exists {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		count++
	}
	return count
}

func normalizeProcessMainStoryTagSet(tags models.StringArray) map[string]struct{} {
	normalized := make(map[string]struct{}, len(tags))
	for _, rawTag := range []string(tags) {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		normalized[tag] = struct{}{}
	}
	return normalized
}

func shouldRetryProcessMainStoryBeatInAnotherZone(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no points of interest found for location archetype") ||
		strings.Contains(message, "no unused points of interest found for location archetype") ||
		strings.Contains(message, "zone has no points of interest for quest giver placement") ||
		strings.Contains(message, "could not find enough places in zone after") ||
		strings.Contains(message, "no recently used places available as fallback")
}

func (p *ProcessMainStoryDistrictRunProcessor) linkRunQuests(
	ctx context.Context,
	quests []*models.Quest,
) error {
	if len(quests) == 0 {
		return nil
	}
	for index, quest := range quests {
		if quest == nil {
			continue
		}
		existing, err := p.dbClient.Quest().FindByID(ctx, quest.ID)
		if err != nil {
			return err
		}
		if existing == nil {
			continue
		}
		if index > 0 && quests[index-1] != nil {
			existing.MainStoryPreviousQuestID = &quests[index-1].ID
		} else {
			existing.MainStoryPreviousQuestID = nil
		}
		if index+1 < len(quests) && quests[index+1] != nil {
			existing.MainStoryNextQuestID = &quests[index+1].ID
		} else {
			existing.MainStoryNextQuestID = nil
		}
		existing.UpdatedAt = time.Now()
		if err := p.dbClient.Quest().Update(ctx, existing.ID, existing); err != nil {
			return err
		}
	}
	return nil
}

func normalizeProcessMainStoryTags(input []string) models.StringArray {
	tags := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, rawTag := range input {
		normalized := strings.ToLower(strings.TrimSpace(rawTag))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		tags = append(tags, normalized)
	}
	return tags
}

func scorePointOfInterestForHintForProcess(
	pointOfInterest models.PointOfInterest,
	hint string,
) int {
	if strings.TrimSpace(hint) == "" {
		return 0
	}
	haystacks := []string{
		pointOfInterest.Name,
		pointOfInterest.OriginalName,
		pointOfInterest.Description,
		pointOfInterest.Clue,
	}
	for _, tag := range pointOfInterest.Tags {
		haystacks = append(haystacks, tag.Value)
	}
	score := 0
	for _, token := range strings.Fields(strings.ToLower(strings.TrimSpace(hint))) {
		if len(token) < 3 {
			continue
		}
		for _, haystack := range haystacks {
			if strings.Contains(strings.ToLower(haystack), token) {
				score++
				break
			}
		}
	}
	return score
}

func pickBestPointOfInterestForHintForProcess(
	pointsOfInterest []models.PointOfInterest,
	hint string,
	excludeID *uuid.UUID,
) *models.PointOfInterest {
	var best *models.PointOfInterest
	bestScore := -1
	for index := range pointsOfInterest {
		pointOfInterest := pointsOfInterest[index]
		if excludeID != nil && pointOfInterest.ID == *excludeID {
			continue
		}
		score := scorePointOfInterestForHintForProcess(pointOfInterest, hint)
		if score > bestScore {
			best = &pointsOfInterest[index]
			bestScore = score
		}
	}
	if bestScore <= 0 && len(pointsOfInterest) > 0 {
		for index := range pointsOfInterest {
			if excludeID != nil && pointsOfInterest[index].ID == *excludeID {
				continue
			}
			return &pointsOfInterest[index]
		}
	}
	return best
}
