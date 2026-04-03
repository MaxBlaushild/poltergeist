package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createMainStoryDistrictRunRequest struct {
	DistrictID string `json:"districtId"`
}

type mainStoryDistrictZoneCandidate struct {
	zone       *models.Zone
	matchCount int
}

func (s *server) getMainStoryDistrictRuns(ctx *gin.Context) {
	templateIDRaw := strings.TrimSpace(ctx.Query("templateId"))
	if templateIDRaw != "" {
		templateID, err := uuid.Parse(templateIDRaw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story template ID"})
			return
		}
		runs, err := s.dbClient.MainStoryDistrictRun().FindByMainStoryTemplateID(ctx, templateID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, runs)
		return
	}

	runs, err := s.dbClient.MainStoryDistrictRun().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, runs)
}

func (s *server) createMainStoryDistrictRun(ctx *gin.Context) {
	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story template ID"})
		return
	}

	var body createMainStoryDistrictRunRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	districtID, err := uuid.Parse(strings.TrimSpace(body.DistrictID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid district ID"})
		return
	}

	template, err := s.dbClient.MainStoryTemplate().FindByID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story template not found"})
		return
	}

	district, err := s.dbClient.District().FindByID(ctx, districtID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if district == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
		return
	}
	if len(district.Zones) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "district has no child zones"})
		return
	}

	run := &models.MainStoryDistrictRun{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		MainStoryTemplateID:   template.ID,
		DistrictID:            district.ID,
		Status:                models.MainStoryDistrictRunStatusInProgress,
		BeatRuns:              models.MainStoryDistrictBeatRuns{},
		GeneratedCharacterIDs: models.StringArray{},
	}
	if err := s.dbClient.MainStoryDistrictRun().Create(ctx, run); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.materializeMainStoryDistrictRun(ctx, template, district, run); err != nil {
		msg := err.Error()
		run.Status = models.MainStoryDistrictRunStatusFailed
		run.ErrorMessage = &msg
		run.UpdatedAt = time.Now()
		_ = s.dbClient.MainStoryDistrictRun().Update(ctx, run)
		ctx.JSON(http.StatusOK, run)
		return
	}

	run.Status = models.MainStoryDistrictRunStatusCompleted
	run.ErrorMessage = nil
	run.UpdatedAt = time.Now()
	if err := s.dbClient.MainStoryDistrictRun().Update(ctx, run); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, run)
}

func (s *server) materializeMainStoryDistrictRun(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	run *models.MainStoryDistrictRun,
) error {
	if template == nil {
		return errRequired("template")
	}
	if district == nil {
		return errRequired("district")
	}
	if run == nil {
		return errRequired("run")
	}
	if len(template.Beats) == 0 {
		return errInvalid("main story template has no beats")
	}

	clonedCharactersByKey, generatedCharacterIDs, err := s.cloneMainStoryDistrictRunCharacters(
		ctx,
		template,
		district,
	)
	if err != nil {
		return err
	}
	run.GeneratedCharacterIDs = generatedCharacterIDs

	beatRuns := make(models.MainStoryDistrictBeatRuns, 0, len(template.Beats))
	createdQuests := make([]*models.Quest, 0, len(template.Beats))
	for _, beat := range template.Beats {
		beatRun, quest, err := s.materializeMainStoryDistrictBeatRun(
			ctx,
			template,
			district,
			beat,
			clonedCharactersByKey,
		)
		beatRuns = append(beatRuns, beatRun)
		run.BeatRuns = beatRuns
		run.UpdatedAt = time.Now()
		if updateErr := s.dbClient.MainStoryDistrictRun().Update(ctx, run); updateErr != nil {
			return updateErr
		}
		if err != nil {
			if linkErr := s.linkMainStoryDistrictRunQuests(ctx, createdQuests); linkErr != nil {
				return linkErr
			}
			return err
		}
		if quest != nil {
			createdQuests = append(createdQuests, quest)
		}
	}

	if err := s.linkMainStoryDistrictRunQuests(ctx, createdQuests); err != nil {
		return err
	}
	run.BeatRuns = beatRuns
	return nil
}

func (s *server) cloneMainStoryDistrictRunCharacters(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
) (map[string]*models.Character, models.StringArray, error) {
	clonedByKey := make(map[string]*models.Character)
	generatedIDs := make(models.StringArray, 0)
	sourceCharactersByID := make(map[uuid.UUID]*models.Character)

	for _, beat := range template.Beats {
		key := normalizeMainStoryCharacterKey(beat.QuestGiverCharacterKey)
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
				sourceCharacter, _ = s.dbClient.Character().FindByID(ctx, *beat.QuestGiverCharacterID)
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
			InternalTags:          buildMainStoryDistrictRunCharacterTags(template, district, key, sourceCharacter),
			ImageGenerationStatus: models.CharacterImageGenerationStatusNone,
		}
		if sourceCharacter != nil {
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
		if err := s.dbClient.Character().Create(ctx, character); err != nil {
			return nil, nil, err
		}
		clonedByKey[key] = character
		generatedIDs = append(generatedIDs, character.ID.String())
	}

	return clonedByKey, generatedIDs, nil
}

func buildMainStoryDistrictRunCharacterTags(
	template *models.MainStoryTemplate,
	district *models.District,
	characterKey string,
	sourceCharacter *models.Character,
) models.StringArray {
	tags := []string{
		"main_story",
		"story_character",
		"district_main_story_run",
		"story_character_" + characterKey,
	}
	if template != nil {
		tags = append(tags, "main_story_template_"+template.ID.String())
		tags = append(tags, []string(template.InternalTags)...)
	}
	if district != nil {
		tags = append(tags, "district_"+district.ID.String())
	}
	if sourceCharacter != nil {
		tags = append(tags, []string(sourceCharacter.InternalTags)...)
	}
	return normalizeQuestTemplateInternalTags(tags)
}

func normalizeMainStoryCharacterKey(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func (s *server) materializeMainStoryDistrictBeatRun(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	beat models.MainStoryBeatDraft,
	clonedCharactersByKey map[string]*models.Character,
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
	questArchetype, err := s.dbClient.QuestArchetype().FindByID(ctx, *beat.QuestArchetypeID)
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

	character := clonedCharactersByKey[normalizeMainStoryCharacterKey(beat.QuestGiverCharacterKey)]
	if character == nil {
		beatRun.Status = models.MainStoryDistrictRunStatusFailed
		beatRun.ErrorMessage = "beat quest giver could not be resolved"
		return beatRun, nil, fmt.Errorf("%s", beatRun.ErrorMessage)
	}
	beatRun.QuestGiverCharacterID = &character.ID
	if strings.TrimSpace(character.Name) != "" {
		beatRun.QuestGiverCharacterName = strings.TrimSpace(character.Name)
	}

	rankedZones, err := rankMainStoryDistrictRunZones(district.Zones, beat, questArchetype, template)
	if err != nil {
		beatRun.Status = models.MainStoryDistrictRunStatusFailed
		beatRun.ErrorMessage = err.Error()
		return beatRun, nil, err
	}

	var lastErr error
	for zoneIndex, candidate := range rankedZones {
		if candidate.zone == nil {
			continue
		}

		beatRun.ZoneID = &candidate.zone.ID
		beatRun.ZoneName = strings.TrimSpace(candidate.zone.Name)

		pointOfInterest, err := s.ensureDistrictRunCharacterPointOfInterest(
			ctx,
			character,
			candidate.zone,
			beat,
		)
		if err != nil {
			lastErr = err
			if shouldRetryMainStoryDistrictBeatInAnotherZone(err) && zoneIndex+1 < len(rankedZones) {
				continue
			}
			break
		}
		if pointOfInterest != nil {
			beatRun.PointOfInterestID = &pointOfInterest.ID
			beatRun.PointOfInterestName = strings.TrimSpace(pointOfInterest.Name)
		}

		quest, err := s.dungeonmaster.GenerateQuest(ctx, candidate.zone, questArchetype.ID, &character.ID)
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
		if shouldRetryMainStoryDistrictBeatInAnotherZone(err) && zoneIndex+1 < len(rankedZones) {
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

func (s *server) ensureDistrictRunCharacterPointOfInterest(
	ctx context.Context,
	character *models.Character,
	zone *models.Zone,
	beat models.MainStoryBeatDraft,
) (*models.PointOfInterest, error) {
	if character == nil {
		return nil, errRequired("character")
	}
	if zone == nil {
		return nil, errRequired("zone")
	}

	if character.PointOfInterestID != nil && *character.PointOfInterestID != uuid.Nil {
		pointOfInterest, err := s.dbClient.PointOfInterest().FindByID(ctx, *character.PointOfInterestID)
		if err == nil && pointOfInterest != nil {
			if pointOfInterestZone, zoneErr := s.dbClient.Zone().FindByPointOfInterestID(ctx, pointOfInterest.ID); zoneErr == nil &&
				pointOfInterestZone != nil &&
				pointOfInterestZone.ID == zone.ID {
				return pointOfInterest, nil
			}
		}
	}

	pointsOfInterest, err := s.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
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
	best := pickBestPointOfInterestForHint(pointsOfInterest, hint, nil)
	if best == nil {
		return nil, fmt.Errorf("no point of interest could be resolved for quest giver placement")
	}

	if err := s.dbClient.Character().UpdateFields(ctx, character.ID, map[string]interface{}{
		"point_of_interest_id": best.ID,
		"updated_at":           time.Now(),
	}); err != nil {
		return nil, err
	}
	character.PointOfInterestID = &best.ID
	character.PointOfInterest = best
	return best, nil
}

func rankMainStoryDistrictRunZones(
	zones []models.Zone,
	beat models.MainStoryBeatDraft,
	questArchetype *models.QuestArchetype,
	template *models.MainStoryTemplate,
) ([]mainStoryDistrictZoneCandidate, error) {
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
	normalized := normalizeMainStoryDistrictSeedTags(desiredTags)
	ranked := make([]mainStoryDistrictZoneCandidate, 0, len(zones))
	for index := range zones {
		ranked = append(ranked, mainStoryDistrictZoneCandidate{
			zone:       &zones[index],
			matchCount: mainStoryDistrictSeedMatchCount(zones[index].InternalTags, normalized),
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].matchCount > ranked[j].matchCount
	})
	return ranked, nil
}

func mainStoryDistrictSeedMatchCount(tags models.StringArray, desired map[string]struct{}) int {
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

func normalizeMainStoryDistrictSeedTags(tags models.StringArray) map[string]struct{} {
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

func shouldRetryMainStoryDistrictBeatInAnotherZone(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no points of interest found for location archetype") ||
		strings.Contains(message, "no unused points of interest found for location archetype") ||
		strings.Contains(message, "zone has no points of interest for quest giver placement")
}

func (s *server) linkMainStoryDistrictRunQuests(
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
		existing, err := s.dbClient.Quest().FindByID(ctx, quest.ID)
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
		if err := s.dbClient.Quest().Update(ctx, existing.ID, existing); err != nil {
			return err
		}
	}
	return nil
}
