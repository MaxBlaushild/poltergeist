package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	Count                        int            `json:"count"`
	YeetIt                       bool           `json:"yeetIt"`
	ZoneKind                     string         `json:"zoneKind"`
	ThemePrompt                  string         `json:"themePrompt"`
	FamilyTags                   []string       `json:"familyTags"`
	FamilyMixTargets             map[string]int `json:"familyMixTargets"`
	CharacterTags                []string       `json:"characterTags"`
	InternalTags                 []string       `json:"internalTags"`
	RequiredLocationArchetypeIDs []string       `json:"requiredLocationArchetypeIds"`
	RequiredLocationMetadataTags []string       `json:"requiredLocationMetadataTags"`
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
	familyMixTargets := models.NormalizeQuestArchetypeSuggestionFamilyMixTargets(body.FamilyMixTargets)
	if sumQuestArchetypeSuggestionFamilyMixTargets(familyMixTargets) > body.Count {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "family mix targets cannot exceed the requested count"})
		return
	}
	zoneKind, err := s.resolveOptionalZoneKind(ctx, body.ZoneKind)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	job := &models.QuestArchetypeSuggestionJob{
		ID:                           uuid.New(),
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
		Status:                       models.QuestArchetypeSuggestionJobStatusQueued,
		Count:                        body.Count,
		YeetIt:                       body.YeetIt,
		ZoneKind:                     models.ZoneKindPromptSlug(zoneKind),
		ThemePrompt:                  strings.TrimSpace(body.ThemePrompt),
		FamilyTags:                   normalizeQuestTemplateInternalTags(body.FamilyTags),
		FamilyMixTargets:             familyMixTargets,
		CharacterTags:                normalizeQuestTemplateCharacterTags(body.CharacterTags),
		InternalTags:                 normalizeQuestTemplateInternalTags(body.InternalTags),
		RequiredLocationArchetypeIDs: requiredLocationArchetypeIDs,
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

func sumQuestArchetypeSuggestionFamilyMixTargets(
	targets models.QuestArchetypeSuggestionFamilyMixTargets,
) int {
	total := 0
	for _, count := range targets {
		if count > 0 {
			total += count
		}
	}
	return total
}

func (s *server) normalizeQuestArchetypeSuggestionLocationArchetypeIDs(
	ctx context.Context,
	rawIDs []string,
) (models.StringArray, error) {
	if len(rawIDs) == 0 {
		return models.StringArray{}, nil
	}
	locationArchetypes, err := s.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load location archetypes")
	}
	allowed := make(map[uuid.UUID]struct{}, len(locationArchetypes))
	for _, archetype := range locationArchetypes {
		if archetype == nil {
			continue
		}
		allowed[archetype.ID] = struct{}{}
	}
	normalized := make(models.StringArray, 0, len(rawIDs))
	seen := make(map[uuid.UUID]struct{}, len(rawIDs))
	for idx, rawID := range rawIDs {
		parsedID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil || parsedID == uuid.Nil {
			return nil, fmt.Errorf("requiredLocationArchetypeIds[%d] must be a valid UUID", idx)
		}
		if _, exists := allowed[parsedID]; !exists {
			return nil, fmt.Errorf("requiredLocationArchetypeIds[%d] could not be found", idx)
		}
		if _, exists := seen[parsedID]; exists {
			continue
		}
		seen[parsedID] = struct{}{}
		normalized = append(normalized, parsedID.String())
	}
	return normalized, nil
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
	for index := range jobsList {
		if !shouldAttemptYeetedQuestArchetypeSuggestionFinalization(&jobsList[index]) {
			continue
		}
		updatedJob, finalizeErr := s.maybeFinalizeYeetedQuestArchetypeSuggestionJob(ctx, &jobsList[index])
		if finalizeErr != nil {
			log.Printf("Failed to finalize yeeted quest archetype suggestion job %s: %v", jobsList[index].ID, finalizeErr)
		}
		if updatedJob != nil {
			jobsList[index] = *updatedJob
		}
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
	if shouldAttemptYeetedQuestArchetypeSuggestionFinalization(job) {
		updatedJob, finalizeErr := s.maybeFinalizeYeetedQuestArchetypeSuggestionJob(ctx, job)
		if finalizeErr != nil {
			log.Printf("Failed to finalize yeeted quest archetype suggestion job %s: %v", job.ID, finalizeErr)
		}
		if updatedJob != nil {
			job = updatedJob
		}
	}
	ctx.JSON(http.StatusOK, job)
}

func (s *server) getQuestArchetypeSuggestionDrafts(ctx *gin.Context) {
	jobID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quest archetype suggestion job ID"})
		return
	}
	job, err := s.dbClient.QuestArchetypeSuggestionJob().FindByID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "quest archetype suggestion job not found"})
		return
	}
	if shouldAttemptYeetedQuestArchetypeSuggestionFinalization(job) {
		if _, finalizeErr := s.maybeFinalizeYeetedQuestArchetypeSuggestionJob(ctx, job); finalizeErr != nil {
			log.Printf("Failed to finalize yeeted quest archetype suggestion job %s: %v", job.ID, finalizeErr)
		}
	}
	drafts, err := s.dbClient.QuestArchetypeSuggestionDraft().FindByJobID(ctx, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, drafts)
}

func (s *server) maybeFinalizeYeetedQuestArchetypeSuggestionJob(
	ctx context.Context,
	job *models.QuestArchetypeSuggestionJob,
) (*models.QuestArchetypeSuggestionJob, error) {
	if !shouldAttemptYeetedQuestArchetypeSuggestionFinalization(job) {
		return job, nil
	}

	drafts, err := s.dbClient.QuestArchetypeSuggestionDraft().FindByJobID(ctx, job.ID)
	if err != nil {
		return job, fmt.Errorf("failed to load quest archetype suggestion drafts: %w", err)
	}
	if len(drafts) == 0 {
		return job, nil
	}

	needsFinalization := false
	convertedCount := 0
	for _, draft := range drafts {
		if draft.QuestArchetypeID != nil {
			convertedCount++
			continue
		}
		needsFinalization = true
	}
	if !needsFinalization {
		if job.Status != models.QuestArchetypeSuggestionJobStatusCompleted ||
			job.CreatedCount != convertedCount ||
			job.ErrorMessage != nil {
			job.Status = models.QuestArchetypeSuggestionJobStatusCompleted
			job.CreatedCount = convertedCount
			job.ErrorMessage = nil
			job.UpdatedAt = time.Now()
			if err := s.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job); err != nil {
				return job, fmt.Errorf("failed to refresh yeeted quest archetype suggestion job counts: %w", err)
			}
		}
		return job, nil
	}

	for index := range drafts {
		draft := &drafts[index]
		if draft.QuestArchetypeID != nil {
			continue
		}
		if _, err := s.materializeQuestArchetypeSuggestionDraft(ctx, draft); err != nil {
			msg := fmt.Sprintf("failed to yeet draft %q into a live archetype: %v", draft.Name, err)
			job.Status = models.QuestArchetypeSuggestionJobStatusFailed
			job.ErrorMessage = &msg
			job.CreatedCount = convertedCount
			job.UpdatedAt = time.Now()
			if updateErr := s.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job); updateErr != nil {
				log.Printf("Failed to mark yeeted quest archetype suggestion job %s as failed: %v", job.ID, updateErr)
			}
			return job, err
		}
		convertedCount++
	}

	job.Status = models.QuestArchetypeSuggestionJobStatusCompleted
	job.ErrorMessage = nil
	job.CreatedCount = convertedCount
	job.UpdatedAt = time.Now()
	if err := s.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job); err != nil {
		return job, fmt.Errorf("failed to finalize yeeted quest archetype suggestion job: %w", err)
	}
	return job, nil
}

func shouldAttemptYeetedQuestArchetypeSuggestionFinalization(
	job *models.QuestArchetypeSuggestionJob,
) bool {
	if job == nil || !job.YeetIt {
		return false
	}
	switch job.Status {
	case models.QuestArchetypeSuggestionJobStatusCompleted,
		models.QuestArchetypeSuggestionJobStatusFailed:
		return true
	default:
		return false
	}
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
	preferredZoneKind, err := s.resolveOptionalZoneKind(ctx, draft.ZoneKind)
	if err != nil {
		return nil, fmt.Errorf("failed to load suggestion draft zone kind: %w", err)
	}
	locationArchetypes, err := s.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load location archetypes: %w", err)
	}
	requiredLocationArchetypeIDs := models.StringArray{}
	if draft.JobID != uuid.Nil {
		job, jobErr := s.dbClient.QuestArchetypeSuggestionJob().FindByID(ctx, draft.JobID)
		if jobErr != nil {
			return nil, fmt.Errorf("failed to load suggestion job for location repair: %w", jobErr)
		}
		if job != nil {
			requiredLocationArchetypeIDs = job.RequiredLocationArchetypeIDs
		}
	}
	if changed, ensureErr := s.ensureQuestArchetypeSuggestionDraftLocationArchetypes(
		ctx,
		draft,
		preferredZoneKind,
		locationArchetypes,
		requiredLocationArchetypeIDs,
	); ensureErr != nil {
		return nil, ensureErr
	} else if changed {
		draft.UpdatedAt = time.Now()
		if draft.JobID != uuid.Nil {
			if err := s.dbClient.QuestArchetypeSuggestionDraft().Update(ctx, draft); err != nil {
				return nil, fmt.Errorf("failed to persist repaired suggestion draft: %w", err)
			}
		}
	}
	draftNodes := questArchetypeSuggestionDraftNodes(draft)
	if len(draftNodes) == 0 {
		return nil, fmt.Errorf("draft does not contain any nodes")
	}
	if draftNodes[0].Source == "proximity" {
		return nil, fmt.Errorf("the first node cannot use proximity")
	}

	monsterTemplates, err := s.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load monster templates: %w", err)
	}

	nodes := make([]*models.QuestArchetypeNode, 0, len(draftNodes))
	nodeIDsByKey := make(map[string]uuid.UUID, len(draftNodes))
	for _, suggestionNode := range draftNodes {
		node, err := s.createQuestArchetypeSuggestionNode(
			ctx,
			suggestionNode,
			draft,
			preferredZoneKind,
			&monsterTemplates,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
		nodeIDsByKey[suggestionNode.NodeKey] = node.ID
	}

	for index, suggestionNode := range draftNodes {
		if err := s.linkQuestArchetypeSuggestionNode(
			ctx,
			nodes[index],
			suggestionNode,
			nodeIDsByKey,
		); err != nil {
			return nil, err
		}
	}

	questArchetype := &models.QuestArchetype{
		ID:                          uuid.New(),
		Name:                        strings.TrimSpace(draft.Name),
		Description:                 strings.TrimSpace(draft.Description),
		ZoneKind:                    models.ZoneKindPromptSlug(preferredZoneKind),
		AcceptanceDialogue:          dialogueSequenceFromLines(draft.AcceptanceDialogue),
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
	// Main story beat conversion reuses this helper with an in-memory draft that
	// is never persisted to quest_archetype_suggestion_drafts, so only write back
	// when there is a real suggestion-job parent.
	if draft.JobID != uuid.Nil {
		if err := s.dbClient.QuestArchetypeSuggestionDraft().Update(ctx, draft); err != nil {
			return nil, fmt.Errorf("failed to update suggestion draft: %w", err)
		}
	}

	return s.dbClient.QuestArchetype().FindByID(ctx, questArchetype.ID)
}

func questArchetypeSuggestionDraftNodes(
	draft *models.QuestArchetypeSuggestionDraft,
) models.QuestArchetypeSuggestionNodes {
	if draft == nil {
		return nil
	}
	if len(draft.Nodes) > 0 {
		nodes := make(models.QuestArchetypeSuggestionNodes, 0, len(draft.Nodes))
		for index, node := range draft.Nodes {
			if strings.TrimSpace(node.NodeKey) == "" {
				node.NodeKey = fmt.Sprintf("node_%d", index+1)
			}
			nodes = append(nodes, node)
		}
		return nodes
	}
	if len(draft.Steps) == 0 {
		return nil
	}
	nodes := make(models.QuestArchetypeSuggestionNodes, 0, len(draft.Steps))
	for index, step := range draft.Steps {
		node := questArchetypeSuggestionNodeFromStep(step, fmt.Sprintf("node_%d", index+1))
		if index+1 < len(draft.Steps) {
			node.Outcomes = models.QuestArchetypeSuggestionNodeOutcomes{
				{
					Outcome:     "success",
					NextNodeKey: fmt.Sprintf("node_%d", index+2),
				},
			}
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func questArchetypeSuggestionNodeFromStep(
	step models.QuestArchetypeSuggestionStep,
	nodeKey string,
) models.QuestArchetypeSuggestionNode {
	return models.QuestArchetypeSuggestionNode{
		NodeKey:                 strings.TrimSpace(nodeKey),
		Source:                  step.Source,
		Content:                 step.Content,
		LocationConcept:         step.LocationConcept,
		LocationArchetypeName:   step.LocationArchetypeName,
		LocationArchetypeID:     step.LocationArchetypeID,
		LocationMetadataTags:    append([]string(nil), step.LocationMetadataTags...),
		DistanceMeters:          step.DistanceMeters,
		TemplateConcept:         step.TemplateConcept,
		PotentialContent:        append([]string(nil), step.PotentialContent...),
		ChallengeQuestion:       step.ChallengeQuestion,
		ChallengeDescription:    step.ChallengeDescription,
		ChallengeSubmissionType: step.ChallengeSubmissionType,
		ChallengeProficiency:    step.ChallengeProficiency,
		ChallengeStatTags:       append([]string(nil), step.ChallengeStatTags...),
		ScenarioPrompt:          step.ScenarioPrompt,
		ScenarioOpenEnded:       step.ScenarioOpenEnded,
		ScenarioBeats:           append([]string(nil), step.ScenarioBeats...),
		ExpositionTitle:         step.ExpositionTitle,
		ExpositionDescription:   step.ExpositionDescription,
		ExpositionSpeakerName:   step.ExpositionSpeakerName,
		ExpositionPortraitURL:   step.ExpositionPortraitURL,
		ExpositionDialogue:      append([]string(nil), step.ExpositionDialogue...),
		MonsterTemplateNames:    append([]string(nil), step.MonsterTemplateNames...),
		MonsterTemplateIDs:      append([]string(nil), step.MonsterTemplateIDs...),
		EncounterTone:           append([]string(nil), step.EncounterTone...),
	}
}

func questArchetypeSuggestionStepFromNode(
	node models.QuestArchetypeSuggestionNode,
) models.QuestArchetypeSuggestionStep {
	return models.QuestArchetypeSuggestionStep{
		Source:                  node.Source,
		Content:                 node.Content,
		LocationConcept:         node.LocationConcept,
		LocationArchetypeName:   node.LocationArchetypeName,
		LocationArchetypeID:     node.LocationArchetypeID,
		LocationMetadataTags:    append([]string(nil), node.LocationMetadataTags...),
		DistanceMeters:          node.DistanceMeters,
		TemplateConcept:         node.TemplateConcept,
		PotentialContent:        append([]string(nil), node.PotentialContent...),
		ChallengeQuestion:       node.ChallengeQuestion,
		ChallengeDescription:    node.ChallengeDescription,
		ChallengeSubmissionType: node.ChallengeSubmissionType,
		ChallengeProficiency:    node.ChallengeProficiency,
		ChallengeStatTags:       append([]string(nil), node.ChallengeStatTags...),
		ScenarioPrompt:          node.ScenarioPrompt,
		ScenarioOpenEnded:       node.ScenarioOpenEnded,
		ScenarioBeats:           append([]string(nil), node.ScenarioBeats...),
		ExpositionTitle:         node.ExpositionTitle,
		ExpositionDescription:   node.ExpositionDescription,
		ExpositionSpeakerName:   node.ExpositionSpeakerName,
		ExpositionPortraitURL:   node.ExpositionPortraitURL,
		ExpositionDialogue:      append([]string(nil), node.ExpositionDialogue...),
		MonsterTemplateNames:    append([]string(nil), node.MonsterTemplateNames...),
		MonsterTemplateIDs:      append([]string(nil), node.MonsterTemplateIDs...),
		EncounterTone:           append([]string(nil), node.EncounterTone...),
	}
}

type questArchetypeSuggestionLocationRepairEntry struct {
	ID           uuid.UUID
	Name         string
	NameTokens   map[string]struct{}
	IntentTokens map[string]struct{}
}

func repairQuestArchetypeSuggestionDraftLocationArchetypes(
	draft *models.QuestArchetypeSuggestionDraft,
	locationArchetypes []*models.LocationArchetype,
	requiredLocationArchetypeIDs []string,
) bool {
	if draft == nil || len(locationArchetypes) == 0 {
		return false
	}
	entries := buildQuestArchetypeSuggestionLocationRepairEntries(locationArchetypes)
	requiredSet := buildQuestArchetypeSuggestionRequiredLocationIDSet(requiredLocationArchetypeIDs)
	changed := false

	if len(draft.Nodes) > 0 {
		for index := range draft.Nodes {
			if repairQuestArchetypeSuggestionNodeLocationArchetype(&draft.Nodes[index], entries, requiredSet) {
				changed = true
			}
		}
		if changed {
			draft.Steps = make(models.QuestArchetypeSuggestionSteps, 0, len(draft.Nodes))
			for _, node := range draft.Nodes {
				draft.Steps = append(draft.Steps, questArchetypeSuggestionStepFromNode(node))
			}
		}
		return changed
	}

	for index := range draft.Steps {
		if repairQuestArchetypeSuggestionStepLocationArchetype(&draft.Steps[index], entries, requiredSet) {
			changed = true
		}
	}
	return changed
}

func buildQuestArchetypeSuggestionLocationRepairEntries(
	locationArchetypes []*models.LocationArchetype,
) []questArchetypeSuggestionLocationRepairEntry {
	entries := make([]questArchetypeSuggestionLocationRepairEntry, 0, len(locationArchetypes))
	for _, archetype := range locationArchetypes {
		if archetype == nil || archetype.ID == uuid.Nil {
			continue
		}
		name := strings.TrimSpace(archetype.Name)
		if name == "" {
			continue
		}
		intentParts := make([]string, 0, len(archetype.IncludedTypes))
		for _, placeType := range archetype.IncludedTypes {
			trimmed := strings.TrimSpace(string(placeType))
			if trimmed != "" {
				intentParts = append(intentParts, trimmed)
			}
		}
		entries = append(entries, questArchetypeSuggestionLocationRepairEntry{
			ID:           archetype.ID,
			Name:         name,
			NameTokens:   questArchetypeSuggestionTokenSet(name),
			IntentTokens: questArchetypeSuggestionTokenSet(strings.Join(intentParts, " ")),
		})
	}
	return entries
}

func buildQuestArchetypeSuggestionRequiredLocationIDSet(
	requiredLocationArchetypeIDs []string,
) map[uuid.UUID]struct{} {
	if len(requiredLocationArchetypeIDs) == 0 {
		return nil
	}
	requiredSet := make(map[uuid.UUID]struct{}, len(requiredLocationArchetypeIDs))
	for _, rawID := range requiredLocationArchetypeIDs {
		parsedID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil || parsedID == uuid.Nil {
			continue
		}
		requiredSet[parsedID] = struct{}{}
	}
	return requiredSet
}

func repairQuestArchetypeSuggestionNodeLocationArchetype(
	node *models.QuestArchetypeSuggestionNode,
	entries []questArchetypeSuggestionLocationRepairEntry,
	requiredSet map[uuid.UUID]struct{},
) bool {
	if node == nil {
		return false
	}
	step := questArchetypeSuggestionStepFromNode(*node)
	if !repairQuestArchetypeSuggestionStepLocationArchetype(&step, entries, requiredSet) {
		return false
	}
	node.LocationArchetypeID = step.LocationArchetypeID
	node.LocationArchetypeName = step.LocationArchetypeName
	return true
}

func repairQuestArchetypeSuggestionStepLocationArchetype(
	step *models.QuestArchetypeSuggestionStep,
	entries []questArchetypeSuggestionLocationRepairEntry,
	requiredSet map[uuid.UUID]struct{},
) bool {
	if step == nil || step.Source != "location" {
		return false
	}
	if step.LocationArchetypeID != nil && *step.LocationArchetypeID != uuid.Nil {
		return false
	}
	resolvedID, resolvedName, ok := resolveQuestArchetypeSuggestionStepLocationArchetypeFallback(
		*step,
		entries,
		requiredSet,
	)
	if !ok {
		return false
	}
	step.LocationArchetypeID = &resolvedID
	step.LocationArchetypeName = resolvedName
	return true
}

func resolveQuestArchetypeSuggestionStepLocationArchetypeFallback(
	step models.QuestArchetypeSuggestionStep,
	entries []questArchetypeSuggestionLocationRepairEntry,
	requiredSet map[uuid.UUID]struct{},
) (uuid.UUID, string, bool) {
	if len(entries) == 0 {
		return uuid.Nil, "", false
	}

	filteredEntries := filterQuestArchetypeSuggestionLocationRepairEntries(entries, requiredSet)
	if len(filteredEntries) == 0 {
		filteredEntries = entries
	}

	queryName := strings.TrimSpace(step.LocationArchetypeName)
	for _, entry := range filteredEntries {
		if strings.EqualFold(strings.TrimSpace(entry.Name), queryName) && queryName != "" {
			return entry.ID, entry.Name, true
		}
	}

	nameTokens := questArchetypeSuggestionTokenSet(queryName)
	queryTokens := questArchetypeSuggestionTokenSet(strings.Join([]string{
		queryName,
		step.LocationConcept,
		step.TemplateConcept,
		strings.Join(step.LocationMetadataTags, " "),
	}, " "))

	if len(queryTokens) > 0 {
		if resolvedID, resolvedName, ok := scoreQuestArchetypeSuggestionLocationRepairEntries(
			filteredEntries,
			nameTokens,
			queryTokens,
			requiredSet,
		); ok {
			return resolvedID, resolvedName, true
		}
		if len(filteredEntries) != len(entries) {
			if resolvedID, resolvedName, ok := scoreQuestArchetypeSuggestionLocationRepairEntries(
				entries,
				nameTokens,
				queryTokens,
				requiredSet,
			); ok {
				return resolvedID, resolvedName, true
			}
		}
	}

	if len(filteredEntries) == 1 {
		return filteredEntries[0].ID, filteredEntries[0].Name, true
	}
	return uuid.Nil, "", false
}

func filterQuestArchetypeSuggestionLocationRepairEntries(
	entries []questArchetypeSuggestionLocationRepairEntry,
	requiredSet map[uuid.UUID]struct{},
) []questArchetypeSuggestionLocationRepairEntry {
	if len(requiredSet) == 0 {
		return entries
	}
	filtered := make([]questArchetypeSuggestionLocationRepairEntry, 0, len(requiredSet))
	for _, entry := range entries {
		if _, exists := requiredSet[entry.ID]; !exists {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func scoreQuestArchetypeSuggestionLocationRepairEntries(
	entries []questArchetypeSuggestionLocationRepairEntry,
	nameTokens map[string]struct{},
	queryTokens map[string]struct{},
	requiredSet map[uuid.UUID]struct{},
) (uuid.UUID, string, bool) {
	bestScore := 0
	bestBaseScore := 0
	bestEntry := questArchetypeSuggestionLocationRepairEntry{}
	for _, entry := range entries {
		baseScore := 0
		if len(nameTokens) > 0 {
			baseScore += questArchetypeSuggestionTokenOverlap(nameTokens, entry.NameTokens) * 5
		}
		baseScore += questArchetypeSuggestionTokenOverlap(queryTokens, entry.NameTokens) * 4
		baseScore += questArchetypeSuggestionTokenOverlap(queryTokens, entry.IntentTokens) * 2
		if baseScore <= 0 {
			continue
		}
		score := baseScore
		if _, exists := requiredSet[entry.ID]; exists {
			score += 2
		}
		if score > bestScore || (score == bestScore && baseScore > bestBaseScore) {
			bestScore = score
			bestBaseScore = baseScore
			bestEntry = entry
		}
	}
	if bestEntry.ID == uuid.Nil || bestBaseScore <= 0 {
		return uuid.Nil, "", false
	}
	return bestEntry.ID, bestEntry.Name, true
}

func questArchetypeSuggestionTokenSet(raw string) map[string]struct{} {
	out := map[string]struct{}{}
	normalized := strings.ToLower(strings.TrimSpace(raw))
	for _, part := range strings.FieldsFunc(normalized, func(char rune) bool {
		return !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9')
	}) {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) < 3 {
			continue
		}
		out[trimmed] = struct{}{}
	}
	return out
}

func questArchetypeSuggestionTokenOverlap(
	left map[string]struct{},
	right map[string]struct{},
) int {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	count := 0
	for token := range left {
		if _, exists := right[token]; exists {
			count++
		}
	}
	return count
}

func (s *server) createQuestArchetypeSuggestionNode(
	ctx context.Context,
	suggestionNode models.QuestArchetypeSuggestionNode,
	draft *models.QuestArchetypeSuggestionDraft,
	preferredZoneKind *models.ZoneKind,
	monsterTemplates *[]models.MonsterTemplate,
) (*models.QuestArchetypeNode, error) {
	step := questArchetypeSuggestionStepFromNode(suggestionNode)
	if step.Source == "location" && (step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil) {
		return nil, fmt.Errorf("%s step %q is missing a resolved location archetype", step.Content, step.LocationConcept)
	}

	payload := questArchetypeNodePayload{}
	if questArchetypeSuggestionNodeHasFailureOutcome(suggestionNode) {
		payload.FailurePolicy = string(models.QuestNodeFailurePolicyTransition)
	}
	switch step.Content {
	case "scenario":
		template, err := s.createQuestArchetypeSuggestionScenarioTemplate(ctx, step, draft, preferredZoneKind)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeScenario)
		payload.ScenarioTemplateID = &template.ID
		payload.LocationArchetypeID = step.LocationArchetypeID
		if step.DistanceMeters != nil {
			payload.EncounterProximityMeters = step.DistanceMeters
		}
	case "exposition":
		if step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil {
			return nil, fmt.Errorf("location exposition step %q is missing a resolved location archetype", step.LocationConcept)
		}
		template, err := s.createQuestArchetypeSuggestionExpositionTemplate(ctx, step, preferredZoneKind)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeExposition)
		payload.LocationArchetypeID = step.LocationArchetypeID
		payload.ExpositionTemplateID = &template.ID
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
				PreferredZoneKind: preferredZoneKind,
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
		template, err := s.createQuestArchetypeSuggestionChallengeTemplate(ctx, step, draft)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeChallenge)
		payload.LocationArchetypeID = step.LocationArchetypeID
		payload.ChallengeTemplateID = &template.ID
		if step.DistanceMeters != nil {
			payload.EncounterProximityMeters = step.DistanceMeters
		}
	}

	node := &models.QuestArchetypeNode{
		ID:         uuid.New(),
		NodeType:   models.QuestArchetypeNodeTypeChallenge,
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

func questArchetypeSuggestionNodeHasFailureOutcome(
	node models.QuestArchetypeSuggestionNode,
) bool {
	for _, outcome := range node.Outcomes {
		if strings.EqualFold(strings.TrimSpace(outcome.Outcome), "failure") &&
			strings.TrimSpace(outcome.NextNodeKey) != "" {
			return true
		}
	}
	return false
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
	seedCount := 0
	if draft != nil {
		seedCount = len(draft.MonsterTemplateSeeds)
	}
	hints := make([]string, 0, len(step.MonsterTemplateNames)+len(step.PotentialContent)+len(step.EncounterTone)+seedCount+2)
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

func (s *server) linkQuestArchetypeSuggestionNode(
	ctx context.Context,
	node *models.QuestArchetypeNode,
	suggestionNode models.QuestArchetypeSuggestionNode,
	nodeIDsByKey map[string]uuid.UUID,
) error {
	if node == nil {
		return fmt.Errorf("node is required")
	}
	successNodeID, err := questArchetypeSuggestionOutcomeNodeID(
		suggestionNode.Outcomes,
		"success",
		nodeIDsByKey,
	)
	if err != nil {
		return err
	}
	failureNodeID, err := questArchetypeSuggestionOutcomeNodeID(
		suggestionNode.Outcomes,
		"failure",
		nodeIDsByKey,
	)
	if err != nil {
		return err
	}
	if suggestionNode.Content != "challenge" && successNodeID == nil && failureNodeID == nil {
		return nil
	}

	challenge := &models.QuestArchetypeChallenge{
		ID:                    uuid.New(),
		Reward:                0,
		Difficulty:            0,
		UnlockedNodeID:        successNodeID,
		FailureUnlockedNodeID: failureNodeID,
	}
	if err := s.dbClient.QuestArchetypeChallenge().Create(ctx, challenge); err != nil {
		return fmt.Errorf("failed to create quest archetype link: %w", err)
	}
	log.Printf(
		"[main-story-convert][quest-archetype-suggestion][link] node=%s challenge=%s successNext=%v failureNext=%v creating explicit node-challenge join",
		node.ID.String(),
		challenge.ID.String(),
		successNodeID,
		failureNodeID,
	)
	return s.dbClient.QuestArchetypeNodeChallenge().Create(ctx, &models.QuestArchetypeNodeChallenge{
		ID:                        uuid.New(),
		QuestArchetypeChallengeID: challenge.ID,
		QuestArchetypeNodeID:      node.ID,
	})
}

func questArchetypeSuggestionOutcomeNodeID(
	outcomes models.QuestArchetypeSuggestionNodeOutcomes,
	outcomeKind string,
	nodeIDsByKey map[string]uuid.UUID,
) (*uuid.UUID, error) {
	for _, outcome := range outcomes {
		if !strings.EqualFold(strings.TrimSpace(outcome.Outcome), outcomeKind) {
			continue
		}
		nextNodeKey := strings.TrimSpace(outcome.NextNodeKey)
		if nextNodeKey == "" {
			return nil, fmt.Errorf("%s branch is missing a next node key", outcomeKind)
		}
		nodeID, ok := nodeIDsByKey[nextNodeKey]
		if !ok {
			return nil, fmt.Errorf("%s branch points to unknown node %q", outcomeKind, nextNodeKey)
		}
		return &nodeID, nil
	}
	return nil, nil
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

func buildQuestArchetypeSuggestionExpositionTemplate(
	step models.QuestArchetypeSuggestionStep,
	preferredZoneKind *models.ZoneKind,
) *models.ExpositionTemplate {
	title := strings.TrimSpace(step.ExpositionTitle)
	if title == "" {
		title = strings.TrimSpace(step.TemplateConcept)
	}
	if title == "" {
		title = "Lingering Echo"
	}

	description := strings.TrimSpace(step.ExpositionDescription)
	if description == "" {
		description = "Generated exposition template."
	}

	dialogue := models.DialogueSequenceFromSpeakerIdentityLines(
		step.ExpositionDialogue,
		questArchetypeSuggestionExpositionSpeakerName(step),
		questArchetypeSuggestionExpositionPortraitURL(step),
	)
	if len(dialogue) == 0 {
		dialogue = models.DialogueSequenceFromSpeakerIdentityLines([]string{
			"Something here is still trying to warn the next person through.",
			"Read the scene closely before you move on.",
		}, questArchetypeSuggestionExpositionSpeakerName(step), questArchetypeSuggestionExpositionPortraitURL(step))
	}

	return &models.ExpositionTemplate{
		ZoneKind:           models.ZoneKindPromptSlug(preferredZoneKind),
		Title:              title,
		Description:        description,
		Dialogue:           dialogue,
		RequiredStoryFlags: models.StringArray{},
		ImageURL:           "",
		ThumbnailURL:       "",
		RewardMode:         models.RewardModeRandom,
		RandomRewardSize:   models.RandomRewardSizeSmall,
		RewardExperience:   0,
		RewardGold:         0,
		MaterialRewards:    models.BaseMaterialRewards{},
		ItemRewards:        models.ExpositionTemplateItemRewards{},
		SpellRewards:       models.ExpositionTemplateSpellRewards{},
	}
}

func questArchetypeSuggestionExpositionSpeakerName(
	step models.QuestArchetypeSuggestionStep,
) string {
	if name := strings.TrimSpace(step.ExpositionSpeakerName); name != "" {
		return name
	}
	if name := strings.TrimSpace(step.ExpositionTitle); name != "" {
		return name
	}
	if name := strings.TrimSpace(step.TemplateConcept); name != "" {
		return name
	}
	if location := strings.TrimSpace(step.LocationConcept); location != "" {
		return location + " Echo"
	}
	return "Witness Echo"
}

func questArchetypeSuggestionExpositionPortraitURL(
	step models.QuestArchetypeSuggestionStep,
) string {
	if portraitURL := strings.TrimSpace(step.ExpositionPortraitURL); portraitURL != "" {
		return portraitURL
	}
	return "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png"
}

func (s *server) createQuestArchetypeSuggestionExpositionTemplate(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
	preferredZoneKind *models.ZoneKind,
) (*models.ExpositionTemplate, error) {
	template := buildQuestArchetypeSuggestionExpositionTemplate(step, preferredZoneKind)
	if err := s.dbClient.ExpositionTemplate().Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create exposition template: %w", err)
	}
	return template, nil
}

func (s *server) createQuestArchetypeSuggestionScenarioTemplate(
	ctx context.Context,
	step models.QuestArchetypeSuggestionStep,
	draft *models.QuestArchetypeSuggestionDraft,
	preferredZoneKind *models.ZoneKind,
) (*models.ScenarioTemplate, error) {
	prompt := strings.TrimSpace(step.ScenarioPrompt)
	if prompt == "" {
		return nil, fmt.Errorf("scenario step %q is missing prompt text", step.LocationConcept)
	}
	template := &models.ScenarioTemplate{
		ZoneKind:                  models.ZoneKindPromptSlug(preferredZoneKind),
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
