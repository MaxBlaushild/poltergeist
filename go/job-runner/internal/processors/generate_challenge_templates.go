package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const challengeTemplateGenerationPromptTemplate = `
You are designing %d reusable fantasy MMORPG challenge templates for a location archetype.

Location archetype:
- name: %s
- included place types: %s
- excluded place types: %s
- existing built-in archetype challenge examples: %s

Recent challenge templates for this archetype to avoid echoing:
%s

Return JSON only:
{
  "challenges": [
    {
      "question": "One short sentence (6-18 words) stating exactly what the player must do",
      "description": "40-140 words of scene flavor and context that supports the action",
      "submissionType": "photo or text",
      "difficulty": 0-40,
      "reward": 0-100,
      "statTags": ["0-3 of: strength,dexterity,constitution,intelligence,wisdom,charisma"],
      "proficiency": "optional short phrase or null"
    }
  ]
}

Hard rules:
- Output exactly %d challenges.
- These are reusable templates, not tied to any specific business name, coordinates, or one-off landmark.
- Each challenge should clearly fit the archetype and invite roleplay in a public-space fantasy MMO tone.
- Keep each challenge materially distinct from the others and from the recent templates list.
- Use only submissionType values: "photo" or "text".
`

type GenerateChallengeTemplatesProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateChallengeTemplatesProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateChallengeTemplatesProcessor {
	log.Println("Initializing GenerateChallengeTemplatesProcessor")
	return GenerateChallengeTemplatesProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateChallengeTemplatesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate challenge templates task: %v", task.Type())

	var payload jobs.GenerateChallengeTemplatesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ChallengeTemplateGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Challenge template generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ChallengeTemplateGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ChallengeTemplateGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateChallengeTemplates(ctx, job); err != nil {
		return p.failChallengeTemplateGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateChallengeTemplatesProcessor) generateChallengeTemplates(ctx context.Context, job *models.ChallengeTemplateGenerationJob) error {
	locationArchetype, err := p.dbClient.LocationArchetype().FindByID(ctx, job.LocationArchetypeID)
	if err != nil {
		return fmt.Errorf("failed to load location archetype: %w", err)
	}
	if locationArchetype == nil {
		return fmt.Errorf("location archetype not found")
	}

	prompt := fmt.Sprintf(
		challengeTemplateGenerationPromptTemplate,
		job.Count,
		strings.TrimSpace(locationArchetype.Name),
		joinPlaceTypes(placeTypesToStrings(locationArchetype.IncludedTypes)),
		joinPlaceTypes(placeTypesToStrings(locationArchetype.ExcludedTypes)),
		joinLocationArchetypeChallengeExamples(locationArchetype.Challenges),
		p.buildRecentChallengeTemplateAvoidance(ctx, job.LocationArchetypeID, 12),
		job.Count,
	)
	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate challenge templates: %w", err)
	}

	generated := &generatedChallengesResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse generated challenge template payload: %w", err)
	}

	challengeSpecs := sanitizeGeneratedChallenges(generated.Challenges, job.Count, locationArchetype.Name)
	createdCount := 0
	for _, spec := range challengeSpecs {
		template := &models.ChallengeTemplate{
			LocationArchetypeID: job.LocationArchetypeID,
			Question:            spec.Question,
			Description:         spec.Description,
			RewardMode:          models.RewardModeExplicit,
			RandomRewardSize:    models.RandomRewardSizeSmall,
			RewardExperience:    0,
			Reward:              spec.Reward,
			ItemChoiceRewards:   models.ChallengeTemplateItemChoiceRewards{},
			SubmissionType:      spec.SubmissionType,
			Difficulty:          spec.Difficulty,
			StatTags:            spec.StatTags,
			Proficiency:         spec.Proficiency,
		}
		if template.Reward == 0 {
			template.RewardMode = models.RewardModeRandom
		}
		if err := p.dbClient.ChallengeTemplate().Create(ctx, template); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create challenge template: %w", err)
		}
		createdCount++
	}

	job.Status = models.ChallengeTemplateGenerationStatusCompleted
	job.CreatedCount = createdCount
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	return p.dbClient.ChallengeTemplateGenerationJob().Update(ctx, job)
}

func (p *GenerateChallengeTemplatesProcessor) failChallengeTemplateGenerationJob(ctx context.Context, job *models.ChallengeTemplateGenerationJob, err error) error {
	msg := err.Error()
	job.Status = models.ChallengeTemplateGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ChallengeTemplateGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark challenge template generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func (p *GenerateChallengeTemplatesProcessor) buildRecentChallengeTemplateAvoidance(ctx context.Context, locationArchetypeID uuid.UUID, limit int) string {
	templates, err := p.dbClient.ChallengeTemplate().FindRecentByLocationArchetypeID(ctx, locationArchetypeID, limit)
	if err != nil || len(templates) == 0 {
		return "- none"
	}
	lines := make([]string, 0, len(templates))
	for _, template := range templates {
		question := strings.TrimSpace(template.Question)
		if question == "" {
			continue
		}
		lines = append(lines, "- "+question)
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func joinPlaceTypes(types []string) string {
	if len(types) == 0 {
		return "none"
	}
	return strings.Join(types, ", ")
}

func placeTypesToStrings[T ~string](types []T) []string {
	if len(types) == 0 {
		return nil
	}
	out := make([]string, 0, len(types))
	for _, item := range types {
		out = append(out, string(item))
	}
	return out
}

func joinLocationArchetypeChallengeExamples(challenges models.LocationArchetypeChallenges) string {
	if len(challenges) == 0 {
		return "none"
	}
	lines := make([]string, 0, len(challenges))
	for _, challenge := range challenges {
		question := strings.TrimSpace(challenge.Question)
		if question == "" {
			continue
		}
		lines = append(lines, question)
		if len(lines) >= 6 {
			break
		}
	}
	if len(lines) == 0 {
		return "none"
	}
	return strings.Join(lines, "; ")
}
