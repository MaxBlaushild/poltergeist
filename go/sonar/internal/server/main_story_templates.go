package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type createMainStoryDistrictRunRequest struct {
	DistrictID string `json:"districtId"`
}

type createMainStoryZoneRunRequest struct {
	DistrictID string `json:"districtId"`
	ZoneID     string `json:"zoneId"`
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

func (s *server) getMainStoryDistrictRun(ctx *gin.Context) {
	runID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story district run ID"})
		return
	}
	run, err := s.dbClient.MainStoryDistrictRun().FindByID(ctx, runID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if run == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story district run not found"})
		return
	}
	ctx.JSON(http.StatusOK, run)
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

	run, err := s.createAndEnqueueMainStoryRun(ctx, template, district, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, run)
}

func (s *server) createMainStoryZoneRun(ctx *gin.Context) {
	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story template ID"})
		return
	}

	var body createMainStoryZoneRunRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	districtID, err := uuid.Parse(strings.TrimSpace(body.DistrictID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid district ID"})
		return
	}
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
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

	var targetZone *models.Zone
	for index := range district.Zones {
		if district.Zones[index].ID == zoneID {
			targetZone = &district.Zones[index]
			break
		}
	}
	if targetZone == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "zone does not belong to district"})
		return
	}

	run, err := s.createAndEnqueueMainStoryRun(ctx, template, district, targetZone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, run)
}

func (s *server) createAndEnqueueMainStoryRun(
	ctx context.Context,
	template *models.MainStoryTemplate,
	district *models.District,
	targetZone *models.Zone,
) (*models.MainStoryDistrictRun, error) {
	run := &models.MainStoryDistrictRun{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		MainStoryTemplateID:   template.ID,
		DistrictID:            district.ID,
		Status:                models.MainStoryDistrictRunStatusQueued,
		BeatRuns:              models.MainStoryDistrictBeatRuns{},
		GeneratedCharacterIDs: models.StringArray{},
	}
	if targetZone != nil {
		run.ZoneID = &targetZone.ID
	}
	if err := s.dbClient.MainStoryDistrictRun().Create(ctx, run); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(jobs.ProcessMainStoryDistrictRunTaskPayload{RunID: run.ID})
	if err != nil {
		errMsg := err.Error()
		run.Status = models.MainStoryDistrictRunStatusFailed
		run.ErrorMessage = &errMsg
		run.UpdatedAt = time.Now()
		_ = s.dbClient.MainStoryDistrictRun().Update(ctx, run)
		return run, err
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ProcessMainStoryDistrictRunTaskType, payload)); err != nil {
		errMsg := err.Error()
		run.Status = models.MainStoryDistrictRunStatusFailed
		run.ErrorMessage = &errMsg
		run.UpdatedAt = time.Now()
		_ = s.dbClient.MainStoryDistrictRun().Update(ctx, run)
		return run, err
	}
	return run, nil
}

func (s *server) deleteMainStoryDistrictRun(ctx *gin.Context) {
	runID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story district run ID"})
		return
	}

	run, err := s.dbClient.MainStoryDistrictRun().FindByID(ctx, runID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if run == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story district run not found"})
		return
	}

	if err := s.rollbackMainStoryDistrictRun(ctx, run); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "main story district run deleted"})
}

func (s *server) retryMainStoryDistrictRun(ctx *gin.Context) {
	runID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story district run ID"})
		return
	}

	run, err := s.dbClient.MainStoryDistrictRun().FindByID(ctx, runID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if run == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story district run not found"})
		return
	}
	if run.Status == models.MainStoryDistrictRunStatusInProgress || run.Status == models.MainStoryDistrictRunStatusQueued {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "main story district run is already pending"})
		return
	}

	template, err := s.dbClient.MainStoryTemplate().FindByID(ctx, run.MainStoryTemplateID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story template not found"})
		return
	}

	district, err := s.dbClient.District().FindByID(ctx, run.DistrictID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if district == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
		return
	}

	run.Status = models.MainStoryDistrictRunStatusQueued
	run.ErrorMessage = nil
	run.UpdatedAt = time.Now()
	if err := s.dbClient.MainStoryDistrictRun().Update(ctx, run); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.ProcessMainStoryDistrictRunTaskPayload{RunID: run.ID})
	if err != nil {
		errMsg := err.Error()
		run.Status = models.MainStoryDistrictRunStatusFailed
		run.ErrorMessage = &errMsg
		run.UpdatedAt = time.Now()
		_ = s.dbClient.MainStoryDistrictRun().Update(ctx, run)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.ProcessMainStoryDistrictRunTaskType, payload)); err != nil {
		errMsg := err.Error()
		run.Status = models.MainStoryDistrictRunStatusFailed
		run.ErrorMessage = &errMsg
		run.UpdatedAt = time.Now()
		_ = s.dbClient.MainStoryDistrictRun().Update(ctx, run)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, run)
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
			nil,
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

func (s *server) loadMainStoryDistrictRunCharacters(
	ctx context.Context,
	run *models.MainStoryDistrictRun,
) (map[string]*models.Character, error) {
	clonedByKey := make(map[string]*models.Character)
	if run == nil {
		return clonedByKey, errRequired("run")
	}

	for _, rawCharacterID := range run.GeneratedCharacterIDs {
		characterID, err := uuid.Parse(strings.TrimSpace(rawCharacterID))
		if err != nil || characterID == uuid.Nil {
			continue
		}
		character, err := s.dbClient.Character().FindByID(ctx, characterID)
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
			key := normalizeMainStoryCharacterKey(strings.TrimPrefix(tag, "story_character_"))
			if key == "" {
				continue
			}
			clonedByKey[key] = character
		}
	}

	return clonedByKey, nil
}

func findFirstIncompleteMainStoryDistrictBeatIndex(
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

func (s *server) loadMainStoryDistrictRunQuests(
	ctx context.Context,
	beatRuns models.MainStoryDistrictBeatRuns,
) ([]*models.Quest, error) {
	quests := make([]*models.Quest, 0, len(beatRuns))
	for _, beatRun := range beatRuns {
		if beatRun.QuestID == nil || *beatRun.QuestID == uuid.Nil {
			continue
		}
		quest, err := s.dbClient.Quest().FindByID(ctx, *beatRun.QuestID)
		if err != nil {
			return nil, err
		}
		if quest != nil {
			quests = append(quests, quest)
		}
	}
	return quests, nil
}

func (s *server) resumeMainStoryDistrictRun(
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

	retryStartIndex := findFirstIncompleteMainStoryDistrictBeatIndex(run, template)
	if retryStartIndex < 0 {
		return fmt.Errorf("main story district run has no failed or incomplete beats to retry")
	}

	clonedCharactersByKey, err := s.loadMainStoryDistrictRunCharacters(ctx, run)
	if err != nil {
		return err
	}

	var failedZoneID *uuid.UUID
	if retryStartIndex < len(run.BeatRuns) {
		failedZoneID = run.BeatRuns[retryStartIndex].ZoneID
	}

	for index := len(run.BeatRuns) - 1; index >= retryStartIndex; index-- {
		questID := run.BeatRuns[index].QuestID
		if questID == nil || *questID == uuid.Nil {
			continue
		}
		if err := s.dbClient.Quest().Delete(ctx, *questID); err != nil {
			return fmt.Errorf("failed to delete partially created quest %s: %w", questID.String(), err)
		}
	}

	completedBeatRuns := append(models.MainStoryDistrictBeatRuns{}, run.BeatRuns[:retryStartIndex]...)
	run.BeatRuns = completedBeatRuns
	run.Status = models.MainStoryDistrictRunStatusInProgress
	run.ErrorMessage = nil
	run.UpdatedAt = time.Now()

	completedQuests, err := s.loadMainStoryDistrictRunQuests(ctx, completedBeatRuns)
	if err != nil {
		return err
	}
	if err := s.linkMainStoryDistrictRunQuests(ctx, completedQuests); err != nil {
		return err
	}
	if err := s.dbClient.MainStoryDistrictRun().Update(ctx, run); err != nil {
		return err
	}

	createdQuests := append([]*models.Quest{}, completedQuests...)
	for beatIndex := retryStartIndex; beatIndex < len(template.Beats); beatIndex++ {
		deprioritizedZoneID := failedZoneID
		if beatIndex > retryStartIndex {
			deprioritizedZoneID = nil
		}

		beatRun, quest, err := s.materializeMainStoryDistrictBeatRun(
			ctx,
			template,
			district,
			template.Beats[beatIndex],
			clonedCharactersByKey,
			deprioritizedZoneID,
		)
		run.BeatRuns = append(run.BeatRuns, beatRun)
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

	return s.linkMainStoryDistrictRunQuests(ctx, createdQuests)
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

	rankedZones, err := rankMainStoryDistrictRunZones(district.Zones, beat, questArchetype, template, deprioritizedZoneID)
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
	deprioritizedZoneID *uuid.UUID,
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
		strings.Contains(message, "zone has no points of interest for quest giver placement") ||
		strings.Contains(message, "could not find enough places in zone after") ||
		strings.Contains(message, "no recently used places available as fallback")
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

func (s *server) rollbackMainStoryDistrictRun(
	ctx context.Context,
	run *models.MainStoryDistrictRun,
) error {
	if run == nil {
		return errRequired("run")
	}

	questIDsToDelete := map[uuid.UUID]struct{}{}
	for index := len(run.BeatRuns) - 1; index >= 0; index-- {
		questID := run.BeatRuns[index].QuestID
		if questID == nil || *questID == uuid.Nil {
			continue
		}
		questIDsToDelete[*questID] = struct{}{}
	}

	for _, rawCharacterID := range run.GeneratedCharacterIDs {
		characterID, err := uuid.Parse(strings.TrimSpace(rawCharacterID))
		if err != nil || characterID == uuid.Nil {
			continue
		}
		quests, err := s.dbClient.Quest().FindByQuestGiverCharacterID(ctx, characterID)
		if err != nil {
			return fmt.Errorf("failed to load quests for generated character %s: %w", characterID.String(), err)
		}
		for _, quest := range quests {
			if quest.ID == uuid.Nil {
				continue
			}
			questIDsToDelete[quest.ID] = struct{}{}
		}
	}

	for questID := range questIDsToDelete {
		quest, err := s.dbClient.Quest().FindByID(ctx, questID)
		if err != nil {
			return fmt.Errorf("failed to load generated quest %s for unlinking: %w", questID.String(), err)
		}
		if quest == nil {
			continue
		}
		quest.MainStoryPreviousQuestID = nil
		quest.MainStoryNextQuestID = nil
		quest.UpdatedAt = time.Now()
		if err := s.dbClient.Quest().Update(ctx, quest.ID, quest); err != nil {
			return fmt.Errorf("failed to unlink generated quest %s: %w", quest.ID.String(), err)
		}
	}

	for index := len(run.BeatRuns) - 1; index >= 0; index-- {
		questID := run.BeatRuns[index].QuestID
		if questID == nil || *questID == uuid.Nil {
			continue
		}
		if _, exists := questIDsToDelete[*questID]; !exists {
			continue
		}
		if err := s.dbClient.Quest().Delete(ctx, *questID); err != nil {
			return fmt.Errorf("failed to delete quest %s: %w", questID.String(), err)
		}
		delete(questIDsToDelete, *questID)
	}

	for questID := range questIDsToDelete {
		if err := s.dbClient.Quest().Delete(ctx, questID); err != nil {
			return fmt.Errorf("failed to delete generated quest %s: %w", questID.String(), err)
		}
	}

	for _, rawCharacterID := range run.GeneratedCharacterIDs {
		characterID, err := uuid.Parse(strings.TrimSpace(rawCharacterID))
		if err != nil || characterID == uuid.Nil {
			continue
		}
		if err := s.dbClient.Character().Delete(ctx, characterID); err != nil {
			return fmt.Errorf("failed to delete generated character %s: %w", characterID.String(), err)
		}
	}

	if err := s.dbClient.MainStoryDistrictRun().Delete(ctx, run.ID); err != nil {
		return fmt.Errorf("failed to delete main story district run: %w", err)
	}

	return nil
}
