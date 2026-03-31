package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type questArchetypeSuggestionJobRequest struct {
	Count                        int      `json:"count"`
	ThemePrompt                  string   `json:"themePrompt"`
	FamilyTags                   []string `json:"familyTags"`
	CharacterTags                []string `json:"characterTags"`
	InternalTags                 []string `json:"internalTags"`
	RequiredLocationMetadataTags []string `json:"requiredLocationMetadataTags"`
}

func (s *server) createQuestArchetypeSuggestionJob(ctx *gin.Context) {
	var body questArchetypeSuggestionJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Count <= 0 {
		body.Count = 6
	}
	if body.Count > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "count must be between 1 and 100"})
		return
	}

	job := &models.QuestArchetypeSuggestionJob{
		ID:                           uuid.New(),
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
		Status:                       models.QuestArchetypeSuggestionJobStatusQueued,
		Count:                        body.Count,
		ThemePrompt:                  strings.TrimSpace(body.ThemePrompt),
		FamilyTags:                   normalizeQuestTemplateInternalTags(body.FamilyTags),
		CharacterTags:                normalizeQuestTemplateCharacterTags(body.CharacterTags),
		InternalTags:                 normalizeQuestTemplateInternalTags(body.InternalTags),
		RequiredLocationMetadataTags: normalizeQuestTemplateInternalTags(body.RequiredLocationMetadataTags),
		CreatedCount:                 0,
	}
	if err := s.dbClient.QuestArchetypeSuggestionJob().Create(ctx, job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(jobs.GenerateQuestArchetypeSuggestionsTaskPayload{JobID: job.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateQuestArchetypeSuggestionsTaskType, payload)); err != nil {
		msg := err.Error()
		job.Status = models.QuestArchetypeSuggestionJobStatusFailed
		job.ErrorMessage = &msg
		job.UpdatedAt = time.Now()
		_ = s.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, job)
}

func (s *server) getQuestArchetypeSuggestionJobs(ctx *gin.Context) {
	limit := 20
	if limitParam := strings.TrimSpace(ctx.Query("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	jobsList, err := s.dbClient.QuestArchetypeSuggestionJob().FindRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobsList)
}

func (s *server) getQuestArchetypeSuggestionJob(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype suggestion job ID"})
		return
	}
	job, err := s.dbClient.QuestArchetypeSuggestionJob().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype suggestion job not found"})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) getQuestArchetypeSuggestionDrafts(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype suggestion job ID"})
		return
	}
	drafts, err := s.dbClient.QuestArchetypeSuggestionDraft().FindByJobID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, drafts)
}

func (s *server) deleteQuestArchetypeSuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.QuestArchetypeSuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype suggestion draft not found"})
		return
	}
	if draft.QuestArchetypeID != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "converted drafts cannot be deleted"})
		return
	}
	if err := s.dbClient.QuestArchetypeSuggestionDraft().Delete(ctx, draftID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "quest archetype suggestion draft deleted"})
}

func (s *server) convertQuestArchetypeSuggestionDraft(ctx *gin.Context) {
	draftID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype suggestion draft ID"})
		return
	}
	draft, err := s.dbClient.QuestArchetypeSuggestionDraft().FindByID(ctx, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype suggestion draft not found"})
		return
	}
	if draft.QuestArchetypeID != nil {
		existing, findErr := s.dbClient.QuestArchetype().FindByID(ctx, *draft.QuestArchetypeID)
		if findErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": findErr.Error()})
			return
		}
		ctx.JSON(http.StatusOK, existing)
		return
	}

	questArchetype, err := s.materializeQuestArchetypeSuggestionDraft(ctx, draft)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, questArchetype)
}

func (s *server) materializeQuestArchetypeSuggestionDraft(
	ctx context.Context,
	draft *models.QuestArchetypeSuggestionDraft,
) (*models.QuestArchetype, error) {
	if draft == nil {
		return nil, fmt.Errorf("draft is required")
	}
	if len(draft.Steps) == 0 {
		return nil, fmt.Errorf("draft does not contain any steps")
	}
	if draft.Steps[0].Source == "proximity" {
		return nil, fmt.Errorf("the first step cannot use proximity")
	}

	monsterTemplates, err := s.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load monster templates: %w", err)
	}

	nodes := make([]*models.QuestArchetypeNode, 0, len(draft.Steps))
	for _, step := range draft.Steps {
		node, err := s.createQuestArchetypeSuggestionNode(ctx, step, draft, &monsterTemplates)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	for index, step := range draft.Steps {
		var nextNodeID *uuid.UUID
		if index+1 < len(nodes) {
			nextNodeID = &nodes[index+1].ID
		}
		if err := s.linkQuestArchetypeSuggestionStep(ctx, nodes[index], step, nextNodeID, draft); err != nil {
			return nil, err
		}
	}

	questArchetype := &models.QuestArchetype{
		ID:                          uuid.New(),
		Name:                        strings.TrimSpace(draft.Name),
		Description:                 strings.TrimSpace(draft.Description),
		AcceptanceDialogue:          normalizeQuestTemplateAcceptanceDialogue(draft.AcceptanceDialogue),
		ImageURL:                    "",
		DifficultyMode:              models.NormalizeQuestDifficultyMode(string(draft.DifficultyMode)),
		Difficulty:                  models.NormalizeQuestDifficulty(draft.Difficulty),
		MonsterEncounterTargetLevel: models.NormalizeMonsterEncounterTargetLevel(draft.MonsterEncounterTargetLevel),
		DefaultGold:                 0,
		RewardMode:                  models.RewardModeRandom,
		RandomRewardSize:            models.RandomRewardSizeSmall,
		RewardExperience:            0,
		MaterialRewards:             models.BaseMaterialRewards{},
		CharacterTags:               draft.CharacterTags,
		InternalTags:                draft.InternalTags,
		RootID:                      nodes[0].ID,
		ItemRewards:                 []models.QuestArchetypeItemReward{},
		SpellRewards:                []models.QuestArchetypeSpellReward{},
	}
	if _, _, dialogue, err := normalizeExplicitQuestTemplateContent(
		questArchetype.Name,
		questArchetype.Description,
		questArchetype.AcceptanceDialogue,
	); err != nil {
		return nil, err
	} else {
		questArchetype.AcceptanceDialogue = dialogue
	}

	if err := s.dbClient.QuestArchetype().Create(ctx, questArchetype); err != nil {
		return nil, fmt.Errorf("failed to create quest archetype: %w", err)
	}

	now := time.Now()
	draft.Status = models.QuestArchetypeSuggestionDraftStatusConverted
	draft.QuestArchetypeID = &questArchetype.ID
	draft.ConvertedAt = &now
	draft.UpdatedAt = now
	if err := s.dbClient.QuestArchetypeSuggestionDraft().Update(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to update suggestion draft: %w", err)
	}

	return s.dbClient.QuestArchetype().FindByID(ctx, questArchetype.ID)
}

func (s *server) createQuestArchetypeSuggestionNode(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
	monsterTemplates *[]models.MonsterTemplate,
) (*models.QuestArchetypeNode, error) {
	if step.Source == "location" && (step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil) {
		return nil, fmt.Errorf("%s step %q is missing a resolved location archetype", step.Content, step.LocationConcept)
	}

	payload := questArchetypeNodePayload{}
	switch step.Content {
	case "scenario":
		template, err := s.createQuestArchetypeSuggestionScenarioTemplate(ctx, step, draft)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeScenario)
		payload.ScenarioTemplateID = &template.ID
		payload.LocationArchetypeID = step.LocationArchetypeID
		if step.DistanceMeters != nil {
			payload.EncounterProximityMeters = step.DistanceMeters
		}
	case "monster":
		locationArchetype, err := s.resolveQuestArchetypeSuggestionStepLocationArchetype(ctx, step)
		if err != nil {
			return nil, err
		}
		templateIDs, err := s.ensureQuestMonsterTemplateIDs(
			ctx,
			monsterTemplates,
			questMonsterTemplateRequest{
				Count:             questArchetypeSuggestionMonsterTemplateCount(step),
				MonsterType:       models.MonsterTemplateTypeMonster,
				ThemePrompt:       strings.TrimSpace(draft.Name + " " + draft.Hook + " " + draft.Description),
				EncounterConcept:  questArchetypeSuggestionMonsterEncounterConcept(step, draft),
				LocationConcept:   strings.TrimSpace(step.LocationConcept),
				LocationArchetype: locationArchetype,
				EncounterTone:     step.EncounterTone,
				SeedHints:         questArchetypeSuggestionMonsterSeedHints(step, draft),
			},
			step.MonsterTemplateIDs,
		)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeMonsterEncounter)
		payload.LocationArchetypeID = step.LocationArchetypeID
		payload.MonsterTemplateIDs = templateIDs
		targetLevel := models.NormalizeMonsterEncounterTargetLevel(draft.MonsterEncounterTargetLevel)
		payload.TargetLevel = &targetLevel
		if step.DistanceMeters != nil {
			payload.EncounterProximityMeters = step.DistanceMeters
		}
	default:
		if step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil {
			return nil, fmt.Errorf("location challenge step %q is missing a resolved location archetype", step.LocationConcept)
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeLocation)
		payload.LocationArchetypeID = step.LocationArchetypeID
	}

	node := &models.QuestArchetypeNode{
		ID:         uuid.New(),
		NodeType:   models.QuestArchetypeNodeTypeLocation,
		Difficulty: 0,
	}
	if err := s.applyQuestArchetypeNodePayload(ctx, node, payload, true); err != nil {
		return nil, err
	}
	if err := s.dbClient.QuestArchetypeNode().Create(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to create quest archetype node: %w", err)
	}
	return node, nil
}

func (s *server) resolveQuestArchetypeSuggestionStepLocationArchetype(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
) (*models.LocationArchetype, error) {
	if step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil {
		return nil, nil
	}
	locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, *step.LocationArchetypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load location archetype for monster step: %w", err)
	}
	return locationArchetype, nil
}

func questArchetypeSuggestionMonsterTemplateCount(step models.QuestArchetypeSuggestionStep) int {
	count := len(step.MonsterTemplateIDs)
	if len(step.MonsterTemplateNames) > count {
		count = len(step.MonsterTemplateNames)
	}
	return maxInt(1, count)
}

func questArchetypeSuggestionMonsterEncounterConcept(
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	if concept := strings.TrimSpace(step.TemplateConcept); concept != "" {
		return concept
	}
	if len(step.PotentialContent) > 0 {
		return strings.TrimSpace(step.PotentialContent[0])
	}
	if draft == nil {
		return strings.TrimSpace(step.LocationConcept)
	}
	return strings.TrimSpace(draft.Name + " " + step.LocationConcept)
}

func questArchetypeSuggestionMonsterSeedHints(
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
) []string {
	hints := make([]string, 0, len(step.MonsterTemplateNames)+len(step.PotentialContent)+len(step.EncounterTone)+len(draft.MonsterTemplateSeeds)+2)
	hints = append(hints, step.MonsterTemplateNames...)
	hints = append(hints, step.PotentialContent...)
	hints = append(hints, step.EncounterTone...)
	if draft != nil {
		hints = append(hints, draft.MonsterTemplateSeeds...)
	}
	if concept := strings.TrimSpace(step.TemplateConcept); concept != "" {
		hints = append(hints, concept)
	}
	if location := strings.TrimSpace(step.LocationConcept); location != "" {
		hints = append(hints, location)
	}
	return hints
}

func (s *server) linkQuestArchetypeSuggestionStep(
	ctx context.Context,
	node *models.QuestArchetypeNode,
	step models.QuestArchetypeSuggestionStep,
	nextNodeID *uuid.UUID,
	draft *models.QuestArchetypeSuggestionDraft,
) error {
	if node == nil {
		return fmt.Errorf("node is required")
	}
	if step.Content != "challenge" && nextNodeID == nil {
		return nil
	}

	var challengeTemplateID *uuid.UUID
	if step.Content == "challenge" {
		template, err := s.createQuestArchetypeSuggestionChallengeTemplate(ctx, step, draft)
		if err != nil {
			return err
		}
		challengeTemplateID = &template.ID
	}

	challenge := &models.QuestArchetypeChallenge{
		ID:                  uuid.New(),
		ChallengeTemplateID: challengeTemplateID,
		Reward:              0,
		Difficulty:          0,
		UnlockedNodeID:      nextNodeID,
	}
	if err := s.dbClient.QuestArchetypeChallenge().Create(ctx, challenge); err != nil {
		return fmt.Errorf("failed to create quest archetype link: %w", err)
	}
	return s.dbClient.QuestArchetypeNodeChallenge().Create(ctx, &models.QuestArchetypeNodeChallenge{
		ID:                        uuid.New(),
		QuestArchetypeChallengeID: challenge.ID,
		QuestArchetypeNodeID:      node.ID,
	})
}

func (s *server) createQuestArchetypeSuggestionChallengeTemplate(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
) (*models.ChallengeTemplate, error) {
	if step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil {
		return nil, fmt.Errorf("challenge step %q is missing a resolved location archetype", step.LocationConcept)
	}
	submissionType := models.QuestNodeSubmissionType(strings.TrimSpace(strings.ToLower(string(step.ChallengeSubmissionType))))
	if submissionType == "" {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	if !submissionType.IsValid() {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	difficulty := models.NormalizeQuestDifficulty(draft.Difficulty)
	template := &models.ChallengeTemplate{
		LocationArchetypeID: *step.LocationArchetypeID,
		Question:            strings.TrimSpace(step.ChallengeQuestion),
		Description:         strings.TrimSpace(step.ChallengeDescription),
		ImageURL:            "",
		ThumbnailURL:        "",
		ScaleWithUserLevel:  false,
		RewardMode:          models.RewardModeRandom,
		RandomRewardSize:    models.RandomRewardSizeSmall,
		RewardExperience:    0,
		Reward:              0,
		InventoryItemID:     nil,
		ItemChoiceRewards:   models.ChallengeTemplateItemChoiceRewards{},
		SubmissionType:      submissionType,
		Difficulty:          difficulty,
		StatTags:            models.StringArray(step.ChallengeStatTags),
		Proficiency:         step.ChallengeProficiency,
	}
	if strings.TrimSpace(template.Question) == "" {
		return nil, fmt.Errorf("challenge step %q is missing question text", step.LocationConcept)
	}
	if strings.TrimSpace(template.Description) == "" {
		template.Description = "Generated challenge template."
	}
	if err := s.dbClient.ChallengeTemplate().Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create challenge template: %w", err)
	}
	return template, nil
}

func (s *server) createQuestArchetypeSuggestionScenarioTemplate(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
) (*models.ScenarioTemplate, error) {
	prompt := strings.TrimSpace(step.ScenarioPrompt)
	if prompt == "" {
		return nil, fmt.Errorf("scenario step %q is missing prompt text", step.LocationConcept)
	}
	template := &models.ScenarioTemplate{
		Prompt:                    prompt,
		ImageURL:                  "",
		ThumbnailURL:              "",
		ScaleWithUserLevel:        false,
		RewardMode:                models.RewardModeRandom,
		RandomRewardSize:          models.RandomRewardSizeSmall,
		Difficulty:                models.NormalizeQuestDifficulty(draft.Difficulty),
		RewardExperience:          0,
		RewardGold:                0,
		OpenEnded:                 step.ScenarioOpenEnded || len(step.ScenarioBeats) == 0,
		FailurePenaltyMode:        models.ScenarioFailurePenaltyModeShared,
		FailureHealthDrainType:    models.ScenarioFailureDrainTypeNone,
		FailureHealthDrainValue:   0,
		FailureManaDrainType:      models.ScenarioFailureDrainTypeNone,
		FailureManaDrainValue:     0,
		FailureStatuses:           models.ScenarioFailureStatusTemplates{},
		SuccessRewardMode:         models.ScenarioSuccessRewardModeShared,
		SuccessHealthRestoreType:  models.ScenarioFailureDrainTypeNone,
		SuccessHealthRestoreValue: 0,
		SuccessManaRestoreType:    models.ScenarioFailureDrainTypeNone,
		SuccessManaRestoreValue:   0,
		SuccessStatuses:           models.ScenarioFailureStatusTemplates{},
		Options:                   models.ScenarioTemplateOptions{},
		ItemRewards:               models.ScenarioTemplateRewards{},
		ItemChoiceRewards:         models.ScenarioTemplateRewards{},
		SpellRewards:              models.ScenarioTemplateSpellRewards{},
	}
	if err := s.dbClient.ScenarioTemplate().Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create scenario template: %w", err)
	}
	return template, nil
}
