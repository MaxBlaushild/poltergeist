package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type mainStorySuggestionJobRequest struct {
	Count                        int      `json:"count"`
	QuestCount                   int      `json:"questCount"`
	ThemePrompt                  string   `json:"themePrompt"`
	DistrictFit                  string   `json:"districtFit"`
	Tone                         string   `json:"tone"`
	FamilyTags                   []string `json:"familyTags"`
	CharacterTags                []string `json:"characterTags"`
	InternalTags                 []string `json:"internalTags"`
	RequiredLocationArchetypeIDs []string `json:"requiredLocationArchetypeIds"`
	RequiredLocationMetadataTags []string `json:"requiredLocationMetadataTags"`
}

func (s *server) createMainStorySuggestionJob(ctx *gin.Context) {
	var body mainStorySuggestionJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Count <= 0 {
		body.Count = 3
	}
	if body.Count > 25 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 25"})
		return
	}
	if body.QuestCount <= 0 {
		body.QuestCount = 15
	}
	if body.QuestCount > 30 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "questCount must be between 1 and 30"})
		return
	}

	requiredLocationArchetypeIDs, err := s.normalizeQuestArchetypeSuggestionLocationArchetypeIDs(
		ctx,
		body.RequiredLocationArchetypeIDs,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &models.MainStorySuggestionJob{
		ID:                           uuid.New(),
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
		Status:                       models.MainStorySuggestionJobStatusQueued,
		Count:                        body.Count,
		QuestCount:                   body.QuestCount,
		ThemePrompt:                  strings.TrimSpace(body.ThemePrompt),
		DistrictFit:                  strings.TrimSpace(body.DistrictFit),
		Tone:                         strings.TrimSpace(body.Tone),
		FamilyTags:                   normalizeQuestTemplateInternalTags(body.FamilyTags),
		CharacterTags:                normalizeQuestTemplateCharacterTags(body.CharacterTags),
		InternalTags:                 normalizeQuestTemplateInternalTags(body.InternalTags),
		RequiredLocationArchetypeIDs: requiredLocationArchetypeIDs,
		RequiredLocationMetadataTags: normalizeQuestTemplateInternalTags(body.RequiredLocationMetadataTags),
		CreatedCount:                 0,
	}
	if err := s.dbClient.MainStorySuggestionJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateMainStorySuggestionsTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateMainStorySuggestionsTaskType, payload)); err != nil {
		msg := err.Error()
		job.Status = models.MainStorySuggestionJobStatusFailed
		job.ErrorMessage = &msg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.MainStorySuggestionJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getMainStorySuggestionJobs(ctx *gin.Context) {
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	jobsList, err := s.dbClient.MainStorySuggestionJob().FindRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getMainStorySuggestionJob(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story suggestion job ID"})
		return
	}
	job, err := s.dbClient.MainStorySuggestionJob().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story suggestion job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) getMainStorySuggestionDrafts(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story suggestion job ID"})
		return
	}
	drafts, err := s.dbClient.MainStorySuggestionDraft().FindByJobID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, drafts)
}

func (s *server) deleteMainStorySuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.MainStorySuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story suggestion draft not found"})
		return
	}
	if draft.MainStoryTemplateID != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "converted drafts cannot be deleted"})
		return
	}
	if err := s.dbClient.MainStorySuggestionDraft().Delete(ctx, draftID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "main story suggestion draft deleted"})
}

func (s *server) convertMainStorySuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.MainStorySuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story suggestion draft not found"})
		return
	}
	if draft.MainStoryTemplateID != nil {
		existing, findErr := s.dbClient.MainStoryTemplate().FindByID(ctx, *draft.MainStoryTemplateID)
		if findErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": findErr.Error()})
			return
		}
		ctx.JSON(http.StatusOK, existing)
		return
	}

	template, err := s.materializeMainStorySuggestionDraft(ctx, draft)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) getMainStoryTemplates(ctx *gin.Context) {
	templates, err := s.dbClient.MainStoryTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (s *server) getMainStoryTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story template ID"})
		return
	}
	template, err := s.dbClient.MainStoryTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story template not found"})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) deleteMainStoryTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story template ID"})
		return
	}
	template, err := s.dbClient.MainStoryTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story template not found"})
		return
	}

	runs, err := s.dbClient.MainStoryDistrictRun().FindByMainStoryTemplateID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(runs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"clean up the %d district run(s) for this template before deleting it",
				len(runs),
			),
		})
		return
	}

	if err := s.removeMainStoryTemplateArtifacts(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "main story template deleted"})
}

func collectMainStoryTemplateCompletionFlags(
	beats models.MainStoryBeatDrafts,
) map[string]struct{} {
	flags := map[string]struct{}{}
	for _, beat := range beats {
		flag := strings.TrimSpace(mainStoryBeatCompletionFlag(beat))
		if flag == "" {
			continue
		}
		flags[flag] = struct{}{}
	}
	return flags
}

func hasAnyMainStoryTemplateCompletionFlag(
	requiredFlags []string,
	completionFlags map[string]struct{},
) bool {
	for _, raw := range requiredFlags {
		flag := strings.TrimSpace(strings.ToLower(raw))
		if flag == "" {
			continue
		}
		if _, exists := completionFlags[flag]; exists {
			return true
		}
	}
	return false
}

func (s *server) removeMainStoryTemplateArtifacts(
	ctx context.Context,
	template *models.MainStoryTemplate,
) error {
	if template == nil {
		return errRequired("template")
	}

	completionFlags := collectMainStoryTemplateCompletionFlags(template.Beats)
	questArchetypeIDs := map[uuid.UUID]struct{}{}
	questGiverCharacterIDs := map[uuid.UUID]struct{}{}
	for _, beat := range template.Beats {
		if beat.QuestArchetypeID != nil && *beat.QuestArchetypeID != uuid.Nil {
			questArchetypeIDs[*beat.QuestArchetypeID] = struct{}{}
		}
		if beat.QuestGiverCharacterID != nil && *beat.QuestGiverCharacterID != uuid.Nil {
			questGiverCharacterIDs[*beat.QuestGiverCharacterID] = struct{}{}
		}
	}

	drafts, err := s.dbClient.MainStorySuggestionDraft().FindByMainStoryTemplateID(ctx, template.ID)
	if err != nil {
		return fmt.Errorf("failed to load linked main story drafts: %w", err)
	}
	shouldDeleteGeneratedArtifacts := len(drafts) > 0
	now := time.Now()
	for _, draft := range drafts {
		draft.MainStoryTemplateID = nil
		draft.MainStoryTemplate = nil
		draft.Status = models.MainStorySuggestionDraftStatusSuggested
		draft.ConvertedAt = nil
		draft.UpdatedAt = now
		if err := s.dbClient.MainStorySuggestionDraft().Update(ctx, &draft); err != nil {
			return fmt.Errorf("failed to detach linked main story draft %s: %w", draft.ID.String(), err)
		}
	}

	if err := s.deleteMainStoryTemplateUnlockContent(ctx, completionFlags); err != nil {
		return err
	}

	if err := s.dbClient.StoryWorldChange().DeleteByMainStoryTemplateID(ctx, template.ID); err != nil {
		return fmt.Errorf("failed to delete story world changes for template %s: %w", template.ID.String(), err)
	}

	if !shouldDeleteGeneratedArtifacts {
		if err := s.dbClient.MainStoryTemplate().Delete(ctx, template.ID); err != nil {
			return fmt.Errorf("failed to delete main story template %s: %w", template.ID.String(), err)
		}
		return nil
	}

	for characterID := range questGiverCharacterIDs {
		if err := s.dbClient.QuestArchetype().ClearQuestGiverCharacterIDByCharacterID(ctx, characterID); err != nil {
			return fmt.Errorf("failed to clear quest giver references for generated story character %s: %w", characterID.String(), err)
		}
	}

	for questArchetypeID := range questArchetypeIDs {
		if err := s.dbClient.QuestArchetype().DeletePermanent(ctx, questArchetypeID); err != nil {
			return fmt.Errorf("failed to delete quest archetype %s: %w", questArchetypeID.String(), err)
		}
	}

	for characterID := range questGiverCharacterIDs {
		if err := s.dbClient.Character().Delete(ctx, characterID); err != nil {
			return fmt.Errorf("failed to delete generated story character %s: %w", characterID.String(), err)
		}
	}

	if err := s.dbClient.MainStoryTemplate().Delete(ctx, template.ID); err != nil {
		return fmt.Errorf("failed to delete main story template %s: %w", template.ID.String(), err)
	}

	return nil
}

func (s *server) deleteMainStoryTemplateUnlockContent(
	ctx context.Context,
	completionFlags map[string]struct{},
) error {
	if len(completionFlags) == 0 {
		return nil
	}

	scenarios, err := s.dbClient.Scenario().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load scenarios for template cleanup: %w", err)
	}
	for _, scenario := range scenarios {
		if !hasAnyMainStoryTemplateCompletionFlag([]string(scenario.RequiredStoryFlags), completionFlags) {
			continue
		}
		if !strings.Contains(strings.ToLower(strings.Join([]string(scenario.InternalTags), ",")), "main_story_unlock") {
			continue
		}
		if err := s.dbClient.Scenario().Delete(ctx, scenario.ID); err != nil {
			return fmt.Errorf("failed to delete unlocked scenario %s: %w", scenario.ID.String(), err)
		}
	}

	challenges, err := s.dbClient.Challenge().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load challenges for template cleanup: %w", err)
	}
	for _, challenge := range challenges {
		if !hasAnyMainStoryTemplateCompletionFlag([]string(challenge.RequiredStoryFlags), completionFlags) {
			continue
		}
		if err := s.dbClient.Challenge().Delete(ctx, challenge.ID); err != nil {
			return fmt.Errorf("failed to delete unlocked challenge %s: %w", challenge.ID.String(), err)
		}
	}

	encounters, err := s.dbClient.MonsterEncounter().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load monster encounters for template cleanup: %w", err)
	}
	for _, encounter := range encounters {
		if !hasAnyMainStoryTemplateCompletionFlag([]string(encounter.RequiredStoryFlags), completionFlags) {
			continue
		}
		loadedEncounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, encounter.ID)
		if err != nil {
			return fmt.Errorf("failed to load unlocked encounter %s: %w", encounter.ID.String(), err)
		}
		if loadedEncounter != nil {
			for _, member := range loadedEncounter.Members {
				if member.MonsterID == uuid.Nil {
					continue
				}
				if err := s.dbClient.Monster().Delete(ctx, member.MonsterID); err != nil {
					return fmt.Errorf("failed to delete encounter monster %s: %w", member.MonsterID.String(), err)
				}
			}
		}
		if err := s.dbClient.MonsterEncounter().Delete(ctx, encounter.ID); err != nil {
			return fmt.Errorf("failed to delete unlocked monster encounter %s: %w", encounter.ID.String(), err)
		}
	}

	return nil
}

func (s *server) materializeMainStorySuggestionDraft(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
) (*models.MainStoryTemplate, error) {
	if draft == nil {
		return nil, errRequired("draft")
	}
	if len(draft.Beats) == 0 {
		return nil, errInvalid("draft does not contain any beats")
	}

	generatedCharactersByKey, err := s.createMainStoryCharacters(ctx, draft)
	if err != nil {
		return nil, err
	}

	beatQuestGivers, err := s.resolveMainStoryBeatQuestGivers(ctx, draft, generatedCharactersByKey)
	if err != nil {
		return nil, err
	}

	templateID := uuid.New()
	templateBeats := make(models.MainStoryBeatDrafts, 0, len(draft.Beats))
	worldChanges := make([]models.StoryWorldChange, 0)
	for index, beat := range draft.Beats {
		resolution := beatQuestGivers[index]
		beat.QuestGiverCharacterKey = resolution.CharacterKey
		if resolution.Character != nil {
			beat.QuestGiverCharacterID = &resolution.Character.ID
			beat.QuestGiverCharacterName = strings.TrimSpace(resolution.Character.Name)
		}
		questArchetype, err := s.materializeMainStoryBeat(ctx, draft, beat, resolution)
		if err != nil {
			return nil, err
		}
		beat.QuestArchetypeID = &questArchetype.ID
		beat.QuestArchetypeName = questArchetype.Name
		resolvedWorldChanges, worldChangeWarnings, err := s.resolveMainStoryBeatWorldChanges(
			ctx,
			templateID,
			&questArchetype.ID,
			beat,
			resolution,
		)
		if err != nil {
			return nil, err
		}
		if len(worldChangeWarnings) > 0 {
			beat.Warnings = append(append(models.StringArray{}, beat.Warnings...), worldChangeWarnings...)
		}
		worldChanges = append(worldChanges, resolvedWorldChanges...)
		if unlockWarnings := s.materializeMainStoryBeatUnlocks(ctx, draft, beat, resolution); len(unlockWarnings) > 0 {
			beat.Warnings = append(append(models.StringArray{}, beat.Warnings...), unlockWarnings...)
		}
		beat.Steps = nil
		templateBeats = append(templateBeats, beat)
	}
	if err := s.applyMainStoryQuestGiverStoryVariants(ctx, templateBeats, beatQuestGivers); err != nil {
		return nil, err
	}

	template := &models.MainStoryTemplate{
		ID:                templateID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Name:              draft.Name,
		Premise:           draft.Premise,
		DistrictFit:       draft.DistrictFit,
		Tone:              draft.Tone,
		ThemeTags:         draft.ThemeTags,
		InternalTags:      draft.InternalTags,
		FactionKeys:       draft.FactionKeys,
		CharacterKeys:     draft.CharacterKeys,
		RevealKeys:        draft.RevealKeys,
		ClimaxSummary:     draft.ClimaxSummary,
		ResolutionSummary: draft.ResolutionSummary,
		WhyItWorks:        draft.WhyItWorks,
		Beats:             templateBeats,
	}
	if err := s.dbClient.MainStoryTemplate().Create(ctx, template); err != nil {
		return nil, err
	}
	if err := s.dbClient.StoryWorldChange().CreateBatch(ctx, worldChanges); err != nil {
		return nil, err
	}

	now := time.Now()
	draft.MainStoryTemplateID = &template.ID
	draft.MainStoryTemplate = template
	draft.Status = models.MainStorySuggestionDraftStatusConverted
	draft.ConvertedAt = &now
	draft.UpdatedAt = now
	draft.Beats = templateBeats
	if err := s.dbClient.MainStorySuggestionDraft().Update(ctx, draft); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *server) materializeMainStoryBeat(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
) (*models.QuestArchetype, error) {
	if resolution.Character == nil {
		return nil, fmt.Errorf("beat %q could not resolve a concrete main story quest giver", beat.Name)
	}
	questGiverCharacterID := &resolution.Character.ID
	tempDraft := &models.QuestArchetypeSuggestionDraft{
		ID:                          uuid.New(),
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
		Status:                      models.QuestArchetypeSuggestionDraftStatusSuggested,
		Name:                        beat.Name,
		Hook:                        beat.Hook,
		Description:                 beat.Description,
		AcceptanceDialogue:          models.StringArray(beat.AcceptanceDialogue),
		CharacterTags:               models.StringArray{},
		InternalTags:                normalizeQuestTemplateInternalTags(append(append([]string{}, beat.InternalTags...), append([]string(draft.InternalTags), "main_story", "story_role_"+strings.TrimSpace(beat.StoryRole))...)),
		DifficultyMode:              beat.DifficultyMode,
		Difficulty:                  beat.Difficulty,
		MonsterEncounterTargetLevel: beat.MonsterEncounterTargetLevel,
		WhyThisScales:               beat.WhyThisScales,
		Steps:                       beat.Steps,
		ChallengeTemplateSeeds:      beat.ChallengeTemplateSeeds,
		ScenarioTemplateSeeds:       beat.ScenarioTemplateSeeds,
		MonsterTemplateSeeds:        beat.MonsterTemplateSeeds,
		Warnings:                    beat.Warnings,
	}
	questArchetype, err := s.materializeQuestArchetypeSuggestionDraft(ctx, tempDraft)
	if err != nil {
		return nil, err
	}
	questArchetype.Category = models.QuestCategoryMainStory
	questArchetype.QuestGiverCharacterID = questGiverCharacterID
	questArchetype.CharacterTags = models.StringArray{}
	questArchetype.RequiredStoryFlags = normalizeStoryFlagKeys([]string(beat.RequiredStoryFlags))
	questArchetype.SetStoryFlags = normalizeStoryFlagKeys([]string(beat.SetStoryFlags))
	questArchetype.ClearStoryFlags = normalizeStoryFlagKeys([]string(beat.ClearStoryFlags))
	questArchetype.QuestGiverRelationshipEffects = normalizeCharacterRelationshipState(beat.QuestGiverRelationshipEffects)
	questArchetype.UpdatedAt = time.Now()
	if err := s.dbClient.QuestArchetype().Update(ctx, questArchetype); err != nil {
		return nil, err
	}
	return questArchetype, nil
}

func mainStoryBeatCompletionFlag(beat models.MainStoryBeatDraft) string {
	for _, flag := range beat.SetStoryFlags {
		normalized := strings.TrimSpace(strings.ToLower(flag))
		if normalized == "" {
			continue
		}
		if strings.Contains(normalized, "_beat_") && strings.HasSuffix(normalized, "_complete") {
			return normalized
		}
	}
	return ""
}

func buildMainStoryQuestGiverStoryVariant(
	beat models.MainStoryBeatDraft,
) *models.CharacterStoryVariant {
	completionFlag := mainStoryBeatCompletionFlag(beat)
	if completionFlag == "" {
		return nil
	}
	description := strings.TrimSpace(beat.QuestGiverAfterDescription)
	dialogueLines := make(models.DialogueSequence, 0, len(beat.QuestGiverAfterDialogue))
	for _, line := range beat.QuestGiverAfterDialogue {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		dialogueLines = append(dialogueLines, models.DialogueMessage{
			Speaker: "character",
			Text:    text,
			Order:   len(dialogueLines),
		})
	}
	if description == "" && len(dialogueLines) == 0 {
		return nil
	}
	return &models.CharacterStoryVariant{
		Priority:           1000 + max(0, beat.OrderIndex),
		RequiredStoryFlags: models.StringArray{completionFlag},
		Description:        description,
		Dialogue:           dialogueLines,
	}
}

func replaceCharacterStoryVariantForFlag(
	existing models.CharacterStoryVariants,
	flag string,
	replacement *models.CharacterStoryVariant,
) models.CharacterStoryVariants {
	filtered := make(models.CharacterStoryVariants, 0, len(existing)+1)
	for _, variant := range existing {
		match := false
		for _, requiredFlag := range variant.RequiredStoryFlags {
			if strings.TrimSpace(strings.ToLower(requiredFlag)) == flag {
				match = true
				break
			}
		}
		if match {
			continue
		}
		filtered = append(filtered, variant)
	}
	if replacement != nil {
		filtered = append(filtered, *replacement)
	}
	return filtered
}

func (s *server) applyMainStoryQuestGiverStoryVariants(
	ctx context.Context,
	beats models.MainStoryBeatDrafts,
	resolutions []mainStoryBeatQuestGiverResolution,
) error {
	if len(beats) == 0 || len(resolutions) == 0 {
		return nil
	}
	type characterVariantUpdate struct {
		character *models.Character
		variants  models.CharacterStoryVariants
	}
	updatesByCharacterID := map[uuid.UUID]*characterVariantUpdate{}
	for index, beat := range beats {
		if index >= len(resolutions) || resolutions[index].Character == nil {
			continue
		}
		variant := buildMainStoryQuestGiverStoryVariant(beat)
		if variant == nil {
			continue
		}
		character := resolutions[index].Character
		completionFlag := mainStoryBeatCompletionFlag(beat)
		entry, ok := updatesByCharacterID[character.ID]
		if !ok {
			entry = &characterVariantUpdate{
				character: character,
				variants:  append(models.CharacterStoryVariants{}, character.StoryVariants...),
			}
			updatesByCharacterID[character.ID] = entry
		}
		entry.variants = replaceCharacterStoryVariantForFlag(
			entry.variants,
			completionFlag,
			variant,
		)
	}
	for characterID, entry := range updatesByCharacterID {
		normalized := normalizeCharacterStoryVariants([]models.CharacterStoryVariant(entry.variants))
		if err := s.dbClient.Character().UpdateFields(ctx, characterID, map[string]interface{}{
			"story_variants": normalized,
		}); err != nil {
			return err
		}
	}
	return nil
}

type mainStoryBeatQuestGiverResolution struct {
	CharacterKey string
	Character    *models.Character
}

func (s *server) resolveMainStoryBeatQuestGivers(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
	generatedByKey map[string]*models.Character,
) ([]mainStoryBeatQuestGiverResolution, error) {
	if draft == nil {
		return nil, errRequired("draft")
	}
	generatedCharacters := make([]*models.Character, 0, len(generatedByKey))
	for _, character := range generatedByKey {
		if character == nil {
			continue
		}
		generatedCharacters = append(generatedCharacters, character)
	}
	var existingCharacters []*models.Character

	indexes := make([]int, 0, len(draft.Beats))
	for index := range draft.Beats {
		indexes = append(indexes, index)
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		left := draft.Beats[indexes[i]].OrderIndex
		right := draft.Beats[indexes[j]].OrderIndex
		if left != right {
			return left < right
		}
		return indexes[i] < indexes[j]
	})

	resolutions := make([]mainStoryBeatQuestGiverResolution, len(draft.Beats))
	assignedByKey := map[string]*models.Character{}
	usedCharacterIDs := map[uuid.UUID]string{}

	for _, beatIndex := range indexes {
		beat := draft.Beats[beatIndex]
		characterKey := resolveMainStoryBeatQuestGiverKey(beat, assignedByKey)
		if characterKey != "" {
			if generatedCharacter, ok := generatedByKey[characterKey]; ok && generatedCharacter != nil {
				assignedByKey[characterKey] = generatedCharacter
				usedCharacterIDs[generatedCharacter.ID] = characterKey
				resolutions[beatIndex] = mainStoryBeatQuestGiverResolution{
					CharacterKey: characterKey,
					Character:    generatedCharacter,
				}
				continue
			}
		}
		if characterKey != "" {
			if assignedCharacter, ok := assignedByKey[characterKey]; ok && assignedCharacter != nil {
				resolutions[beatIndex] = mainStoryBeatQuestGiverResolution{
					CharacterKey: characterKey,
					Character:    assignedCharacter,
				}
				continue
			}
		}

		candidates := rankedMainStoryQuestGiverCandidates(generatedCharacters, beat, characterKey)
		if len(candidates) == 0 {
			if existingCharacters == nil {
				var err error
				existingCharacters, err = s.dbClient.Character().FindAll(ctx)
				if err != nil {
					return nil, err
				}
			}
			candidates = rankedMainStoryQuestGiverCandidates(existingCharacters, beat, characterKey)
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf(
				"beat %q could not resolve a quest giver for story key %q",
				strings.TrimSpace(beat.Name),
				characterKey,
			)
		}

		selected, err := selectMainStoryQuestGiverCandidate(candidates, usedCharacterIDs, characterKey)
		if err != nil {
			return nil, fmt.Errorf("beat %q: %w", strings.TrimSpace(beat.Name), err)
		}

		if characterKey != "" {
			assignedByKey[characterKey] = selected
			usedCharacterIDs[selected.ID] = characterKey
		}
		resolutions[beatIndex] = mainStoryBeatQuestGiverResolution{
			CharacterKey: characterKey,
			Character:    selected,
		}
	}

	return resolutions, nil
}

func resolveMainStoryBeatQuestGiverKey(
	beat models.MainStoryBeatDraft,
	assignedByKey map[string]*models.Character,
) string {
	candidate := strings.TrimSpace(strings.ToLower(beat.QuestGiverCharacterKey))
	if candidate != "" {
		return candidate
	}
	for _, key := range beat.RequiredCharacterKeys {
		normalized := strings.TrimSpace(strings.ToLower(key))
		if normalized == "" {
			continue
		}
		if _, ok := assignedByKey[normalized]; ok {
			return normalized
		}
	}
	for _, key := range beat.RequiredCharacterKeys {
		normalized := strings.TrimSpace(strings.ToLower(key))
		if normalized != "" {
			return normalized
		}
	}
	for _, key := range beat.IntroducedCharacterKeys {
		normalized := strings.TrimSpace(strings.ToLower(key))
		if normalized == "" {
			continue
		}
		if _, ok := assignedByKey[normalized]; ok {
			return normalized
		}
	}
	for _, key := range beat.IntroducedCharacterKeys {
		normalized := strings.TrimSpace(strings.ToLower(key))
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

type mainStoryQuestGiverCandidate struct {
	Character   *models.Character
	Score       int
	TagMatches  int
	KeyMatched  bool
	CharacterID uuid.UUID
}

func rankedMainStoryQuestGiverCandidates(
	characters []*models.Character,
	beat models.MainStoryBeatDraft,
	characterKey string,
) []mainStoryQuestGiverCandidate {
	desiredTags := map[string]struct{}{}
	for _, tag := range beat.CharacterTags {
		normalized := strings.TrimSpace(strings.ToLower(tag))
		if normalized != "" {
			desiredTags[normalized] = struct{}{}
		}
	}
	if characterKey != "" {
		desiredTags[characterKey] = struct{}{}
	}

	candidates := make([]mainStoryQuestGiverCandidate, 0)
	for _, character := range characters {
		if character == nil || len(character.InternalTags) == 0 {
			continue
		}
		matchCount := 0
		keyMatched := false
		for _, tag := range character.InternalTags {
			normalized := strings.TrimSpace(strings.ToLower(tag))
			if normalized == "" {
				continue
			}
			if normalized == characterKey && characterKey != "" {
				keyMatched = true
			}
			if _, ok := desiredTags[normalized]; ok {
				matchCount++
			}
		}
		if matchCount == 0 {
			continue
		}
		score := matchCount
		if keyMatched {
			score += 100
		}
		candidates = append(candidates, mainStoryQuestGiverCandidate{
			Character:   character,
			Score:       score,
			TagMatches:  matchCount,
			KeyMatched:  keyMatched,
			CharacterID: character.ID,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].TagMatches != candidates[j].TagMatches {
			return candidates[i].TagMatches > candidates[j].TagMatches
		}
		leftName := strings.ToLower(strings.TrimSpace(candidates[i].Character.Name))
		rightName := strings.ToLower(strings.TrimSpace(candidates[j].Character.Name))
		if leftName != rightName {
			return leftName < rightName
		}
		return candidates[i].CharacterID.String() < candidates[j].CharacterID.String()
	})

	return candidates
}

func selectMainStoryQuestGiverCandidate(
	candidates []mainStoryQuestGiverCandidate,
	usedCharacterIDs map[uuid.UUID]string,
	characterKey string,
) (*models.Character, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidate quest givers available")
	}
	for _, candidate := range candidates {
		if existingKey, used := usedCharacterIDs[candidate.CharacterID]; used && existingKey != "" && existingKey != characterKey {
			continue
		}
		return candidate.Character, nil
	}
	if characterKey != "" {
		return nil, fmt.Errorf("all matching quest giver candidates are already assigned to other story characters for key %q", characterKey)
	}
	return candidates[0].Character, nil
}

func errRequired(field string) error {
	return &requestValidationError{message: field + " is required"}
}

func errInvalid(message string) error {
	return &requestValidationError{message: message}
}

type requestValidationError struct {
	message string
}

func (e *requestValidationError) Error() string {
	return e.message
}
