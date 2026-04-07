package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func normalizeMainStoryTextList(input []string) models.StringArray {
	values := make(models.StringArray, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		values = append(values, trimmed)
	}
	if values == nil {
		return models.StringArray{}
	}
	return values
}

func normalizeMainStoryDialogueLines(input []string) models.StringArray {
	lines := make(models.StringArray, 0, len(input))
	for _, raw := range input {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
	}
	if lines == nil {
		return models.StringArray{}
	}
	return lines
}

func normalizeMainStoryStepList(input []string) []string {
	values := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		values = append(values, trimmed)
	}
	if values == nil {
		return []string{}
	}
	return values
}

func (s *server) normalizeMainStoryBeatStep(
	ctx *gin.Context,
	index int,
	step models.QuestArchetypeSuggestionStep,
) (models.QuestArchetypeSuggestionStep, error) {
	step.Source = strings.TrimSpace(step.Source)
	step.Content = strings.TrimSpace(step.Content)
	step.LocationConcept = strings.TrimSpace(step.LocationConcept)
	step.LocationArchetypeName = strings.TrimSpace(step.LocationArchetypeName)
	step.TemplateConcept = strings.TrimSpace(step.TemplateConcept)
	step.ChallengeQuestion = strings.TrimSpace(step.ChallengeQuestion)
	step.ChallengeDescription = strings.TrimSpace(step.ChallengeDescription)
	step.ScenarioPrompt = strings.TrimSpace(step.ScenarioPrompt)
	if step.ChallengeProficiency != nil {
		proficiency := strings.TrimSpace(*step.ChallengeProficiency)
		if proficiency == "" {
			step.ChallengeProficiency = nil
		} else {
			step.ChallengeProficiency = &proficiency
		}
	}
	if step.DistanceMeters != nil && *step.DistanceMeters < 0 {
		return step, fmt.Errorf("beats steps[%d].distanceMeters must be zero or greater", index)
	}
	if step.LocationArchetypeID != nil && *step.LocationArchetypeID != uuid.Nil {
		locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, *step.LocationArchetypeID)
		if err != nil {
			return step, fmt.Errorf("beats steps[%d].locationArchetypeId could not be loaded", index)
		}
		if locationArchetype == nil {
			return step, fmt.Errorf("beats steps[%d].locationArchetypeId could not be loaded", index)
		}
		step.LocationArchetypeName = strings.TrimSpace(locationArchetype.Name)
	} else {
		step.LocationArchetypeID = nil
	}
	step.LocationMetadataTags = normalizeMainStoryStepList(step.LocationMetadataTags)
	step.PotentialContent = normalizeMainStoryStepList(step.PotentialContent)
	step.ChallengeStatTags = normalizeMainStoryStepList(step.ChallengeStatTags)
	step.ScenarioBeats = normalizeMainStoryStepList(step.ScenarioBeats)
	step.MonsterTemplateNames = normalizeMainStoryStepList(step.MonsterTemplateNames)
	step.MonsterTemplateIDs = normalizeMainStoryStepList(step.MonsterTemplateIDs)
	step.EncounterTone = normalizeMainStoryStepList(step.EncounterTone)
	return step, nil
}

func (s *server) normalizeMainStoryBeatDraft(
	ctx *gin.Context,
	index int,
	beat models.MainStoryBeatDraft,
) (models.MainStoryBeatDraft, error) {
	beat.OrderIndex = index + 1
	if beat.Act <= 0 {
		beat.Act = 1
	}
	beat.StoryRole = strings.TrimSpace(beat.StoryRole)
	beat.ChapterTitle = strings.TrimSpace(beat.ChapterTitle)
	if beat.ChapterTitle == "" {
		return beat, fmt.Errorf("beats[%d].chapterTitle is required", index)
	}
	beat.ChapterSummary = strings.TrimSpace(beat.ChapterSummary)
	beat.Purpose = strings.TrimSpace(beat.Purpose)
	beat.WhatChanges = strings.TrimSpace(beat.WhatChanges)
	beat.Name = strings.TrimSpace(beat.Name)
	if beat.Name == "" {
		beat.Name = beat.ChapterTitle
	}
	beat.Hook = strings.TrimSpace(beat.Hook)
	beat.Description = strings.TrimSpace(beat.Description)
	beat.QuestGiverCharacterKey = strings.TrimSpace(beat.QuestGiverCharacterKey)
	beat.QuestGiverCharacterName = strings.TrimSpace(beat.QuestGiverCharacterName)
	beat.QuestGiverAfterDescription = strings.TrimSpace(beat.QuestGiverAfterDescription)
	beat.QuestGiverAfterDialogue = normalizeMainStoryDialogueLines(beat.QuestGiverAfterDialogue)
	beat.AcceptanceDialogue = normalizeMainStoryDialogueLines(beat.AcceptanceDialogue)
	beat.IntroducedCharacterKeys = normalizeMainStoryTextList([]string(beat.IntroducedCharacterKeys))
	beat.RequiredCharacterKeys = normalizeMainStoryTextList([]string(beat.RequiredCharacterKeys))
	beat.IntroducedRevealKeys = normalizeMainStoryTextList([]string(beat.IntroducedRevealKeys))
	beat.RequiredRevealKeys = normalizeMainStoryTextList([]string(beat.RequiredRevealKeys))
	beat.RequiredZoneTags = normalizeMainStoryTextList([]string(beat.RequiredZoneTags))
	locationArchetypeIDs, err := s.normalizeQuestArchetypeSuggestionLocationArchetypeIDs(
		ctx,
		beat.RequiredLocationArchetypeIDs,
	)
	if err != nil {
		return beat, fmt.Errorf("beats[%d].requiredLocationArchetypeIds: %w", index, err)
	}
	beat.RequiredLocationArchetypeIDs = locationArchetypeIDs
	beat.PreferredContentMix = normalizeMainStoryTextList([]string(beat.PreferredContentMix))
	beat.RequiredStoryFlags = normalizeStoryFlagKeys(beat.RequiredStoryFlags)
	beat.SetStoryFlags = normalizeStoryFlagKeys(beat.SetStoryFlags)
	beat.ClearStoryFlags = normalizeStoryFlagKeys(beat.ClearStoryFlags)
	beat.CharacterTags = normalizeQuestTemplateCharacterTags(beat.CharacterTags)
	beat.InternalTags = normalizeQuestTemplateInternalTags(beat.InternalTags)
	beat.QuestGiverRelationshipEffects = normalizeCharacterRelationshipState(beat.QuestGiverRelationshipEffects)
	beat.WhyThisScales = strings.TrimSpace(beat.WhyThisScales)
	beat.ChallengeTemplateSeeds = normalizeMainStoryTextList([]string(beat.ChallengeTemplateSeeds))
	beat.ScenarioTemplateSeeds = normalizeMainStoryTextList([]string(beat.ScenarioTemplateSeeds))
	beat.MonsterTemplateSeeds = normalizeMainStoryTextList([]string(beat.MonsterTemplateSeeds))
	beat.Warnings = normalizeMainStoryTextList([]string(beat.Warnings))

	if beat.Difficulty <= 0 {
		beat.Difficulty = 1
	}
	difficultyMode := beat.DifficultyMode
	if difficultyMode == "" {
		difficultyMode = models.QuestDifficultyModeScale
	}
	normalizedDifficultyMode, normalizedDifficulty, err := normalizeQuestDifficultySettings(
		string(difficultyMode),
		&beat.Difficulty,
		models.QuestDifficultyModeScale,
		1,
	)
	if err != nil {
		return beat, fmt.Errorf("beats[%d] difficulty settings invalid: %w", index, err)
	}
	beat.DifficultyMode = normalizedDifficultyMode
	beat.Difficulty = normalizedDifficulty

	normalizedTargetLevel, err := normalizeMonsterEncounterTargetLevel(
		&beat.MonsterEncounterTargetLevel,
		beat.Difficulty,
	)
	if err != nil {
		return beat, fmt.Errorf("beats[%d].monsterEncounterTargetLevel invalid: %w", index, err)
	}
	beat.MonsterEncounterTargetLevel = normalizedTargetLevel

	if beat.QuestArchetypeID != nil && *beat.QuestArchetypeID != uuid.Nil {
		questArchetype, err := s.dbClient.QuestArchetype().FindByID(ctx, *beat.QuestArchetypeID)
		if err != nil {
			return beat, fmt.Errorf("beats[%d].questArchetypeId could not be loaded", index)
		}
		if questArchetype == nil {
			return beat, fmt.Errorf("beats[%d].questArchetypeId could not be loaded", index)
		}
		beat.QuestArchetypeName = strings.TrimSpace(questArchetype.Name)
	} else {
		beat.QuestArchetypeID = nil
		beat.QuestArchetypeName = strings.TrimSpace(beat.QuestArchetypeName)
	}

	normalizedSteps := make(models.QuestArchetypeSuggestionSteps, 0, len(beat.Steps))
	for stepIndex, step := range beat.Steps {
		normalizedStep, err := s.normalizeMainStoryBeatStep(ctx, stepIndex, step)
		if err != nil {
			return beat, fmt.Errorf("beats[%d]: %w", index, err)
		}
		if normalizedStep.Source == "" && normalizedStep.Content == "" && normalizedStep.TemplateConcept == "" && normalizedStep.LocationConcept == "" {
			continue
		}
		normalizedSteps = append(normalizedSteps, normalizedStep)
	}
	beat.Steps = normalizedSteps

	normalizedWorldChanges := make([]models.MainStoryWorldChange, 0, len(beat.WorldChanges))
	for _, change := range beat.WorldChanges {
		change.Type = strings.TrimSpace(change.Type)
		change.TargetKey = strings.TrimSpace(change.TargetKey)
		change.CharacterKey = strings.TrimSpace(change.CharacterKey)
		change.PointOfInterestHint = strings.TrimSpace(change.PointOfInterestHint)
		change.DestinationHint = strings.TrimSpace(change.DestinationHint)
		change.Description = strings.TrimSpace(change.Description)
		change.Clue = strings.TrimSpace(change.Clue)
		change.ZoneTags = normalizeMainStoryTextList([]string(change.ZoneTags))
		if change.Type == "" && change.TargetKey == "" && change.Description == "" {
			continue
		}
		normalizedWorldChanges = append(normalizedWorldChanges, change)
	}
	beat.WorldChanges = normalizedWorldChanges

	normalizedUnlockedScenarios := make([]models.MainStoryUnlockedScenario, 0, len(beat.UnlockedScenarios))
	for _, scenario := range beat.UnlockedScenarios {
		scenario.Name = strings.TrimSpace(scenario.Name)
		scenario.Prompt = strings.TrimSpace(scenario.Prompt)
		scenario.PointOfInterestHint = strings.TrimSpace(scenario.PointOfInterestHint)
		scenario.InternalTags = normalizeMainStoryTextList([]string(scenario.InternalTags))
		if scenario.Name == "" && scenario.Prompt == "" {
			continue
		}
		if scenario.Difficulty <= 0 {
			scenario.Difficulty = beat.Difficulty
		}
		normalizedUnlockedScenarios = append(normalizedUnlockedScenarios, scenario)
	}
	beat.UnlockedScenarios = normalizedUnlockedScenarios

	normalizedUnlockedChallenges := make([]models.MainStoryUnlockedChallenge, 0, len(beat.UnlockedChallenges))
	for _, challenge := range beat.UnlockedChallenges {
		challenge.Question = strings.TrimSpace(challenge.Question)
		challenge.Description = strings.TrimSpace(challenge.Description)
		challenge.PointOfInterestHint = strings.TrimSpace(challenge.PointOfInterestHint)
		if challenge.Proficiency != nil {
			proficiency := strings.TrimSpace(*challenge.Proficiency)
			if proficiency == "" {
				challenge.Proficiency = nil
			} else {
				challenge.Proficiency = &proficiency
			}
		}
		challenge.StatTags = normalizeMainStoryTextList([]string(challenge.StatTags))
		if challenge.Question == "" && challenge.Description == "" {
			continue
		}
		if challenge.Difficulty <= 0 {
			challenge.Difficulty = beat.Difficulty
		}
		normalizedUnlockedChallenges = append(normalizedUnlockedChallenges, challenge)
	}
	beat.UnlockedChallenges = normalizedUnlockedChallenges

	normalizedUnlockedEncounters := make([]models.MainStoryUnlockedEncounter, 0, len(beat.UnlockedMonsterEncounters))
	for _, encounter := range beat.UnlockedMonsterEncounters {
		encounter.Name = strings.TrimSpace(encounter.Name)
		encounter.Description = strings.TrimSpace(encounter.Description)
		encounter.PointOfInterestHint = strings.TrimSpace(encounter.PointOfInterestHint)
		encounter.EncounterTone = normalizeMainStoryTextList([]string(encounter.EncounterTone))
		encounter.MonsterTemplateHints = normalizeMainStoryTextList([]string(encounter.MonsterTemplateHints))
		if encounter.Name == "" && encounter.Description == "" {
			continue
		}
		if encounter.MonsterCount <= 0 {
			encounter.MonsterCount = 1
		}
		if encounter.TargetLevel <= 0 {
			encounter.TargetLevel = beat.MonsterEncounterTargetLevel
		}
		if encounter.EncounterType == "" {
			encounter.EncounterType = models.MonsterEncounterTypeMonster
		}
		normalizedUnlockedEncounters = append(normalizedUnlockedEncounters, encounter)
	}
	beat.UnlockedMonsterEncounters = normalizedUnlockedEncounters

	return beat, nil
}

func (s *server) normalizeMainStoryTemplate(
	ctx *gin.Context,
	input models.MainStoryTemplate,
	existingID *uuid.UUID,
) (*models.MainStoryTemplate, error) {
	template := &models.MainStoryTemplate{
		Name:              strings.TrimSpace(input.Name),
		Premise:           strings.TrimSpace(input.Premise),
		DistrictFit:       strings.TrimSpace(input.DistrictFit),
		Tone:              strings.TrimSpace(input.Tone),
		ThemeTags:         normalizeMainStoryTextList([]string(input.ThemeTags)),
		InternalTags:      normalizeMainStoryTextList([]string(input.InternalTags)),
		FactionKeys:       normalizeMainStoryTextList([]string(input.FactionKeys)),
		CharacterKeys:     normalizeMainStoryTextList([]string(input.CharacterKeys)),
		RevealKeys:        normalizeMainStoryTextList([]string(input.RevealKeys)),
		ClimaxSummary:     strings.TrimSpace(input.ClimaxSummary),
		ResolutionSummary: strings.TrimSpace(input.ResolutionSummary),
		WhyItWorks:        strings.TrimSpace(input.WhyItWorks),
	}
	if template.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if template.Premise == "" {
		return nil, fmt.Errorf("premise is required")
	}

	beats := make(models.MainStoryBeatDrafts, 0, len(input.Beats))
	for beatIndex, beat := range input.Beats {
		normalizedBeat, err := s.normalizeMainStoryBeatDraft(ctx, beatIndex, beat)
		if err != nil {
			return nil, err
		}
		beats = append(beats, normalizedBeat)
	}
	template.Beats = beats

	now := time.Now()
	if existingID != nil && *existingID != uuid.Nil {
		template.ID = *existingID
		template.UpdatedAt = now
	} else {
		template.ID = uuid.New()
		template.CreatedAt = now
		template.UpdatedAt = now
	}

	return template, nil
}

func (s *server) createMainStoryTemplate(ctx *gin.Context) {
	var requestBody models.MainStoryTemplate
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.normalizeMainStoryTemplate(ctx, requestBody, nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MainStoryTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) updateMainStoryTemplate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid main story template ID"})
		return
	}
	existing, err := s.dbClient.MainStoryTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "main story template not found"})
		return
	}

	var requestBody models.MainStoryTemplate
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.normalizeMainStoryTemplate(ctx, requestBody, &id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template.CreatedAt = existing.CreatedAt

	if err := s.dbClient.MainStoryTemplate().Update(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, template)
}
