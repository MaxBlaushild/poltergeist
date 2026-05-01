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
	"github.com/hibiken/asynq"
)

const expositionTemplateGenerationPromptTemplate = `
You are designing %d reusable fantasy MMORPG exposition templates for map encounters.

Recent exposition templates to avoid echoing:
%s

Return JSON only:
{
  "templates": [
    {
      "title": "2-5 word evocative title",
      "description": "2-4 vivid sentences of setup and atmosphere",
      "dialogue": [
        "3-6 short lines of in-world dialogue"
      ],
      "rewardMode": "random or explicit",
      "randomRewardSize": "small or medium or large",
      "rewardExperience": 0-70,
      "rewardGold": 0-70
    }
  ]
}

Hard rules:
- Output exactly %d templates.
- These are reusable templates, not tied to a specific business, landmark, district, or coordinates.
- Each exposition should feel like a brief discoverable encounter, omen, witness account, magical residue, local warning, or atmospheric vignette.
- Let the requested zone kind strongly influence imagery, props, hazards, folklore, traversal, and mood.
- Dialogue should work without a specific named NPC. It can sound like a traveler, spirit, ranger, sentry, survivor, scavenger, or ambient supernatural voice.
- Keep dialogue lines short, punchy, and easy to present one after another.
- Keep each template materially distinct from the others and from the recent templates list.
- If rewardMode is "random", set rewardExperience and rewardGold to 0.
- If rewardMode is "explicit", keep rewardExperience and rewardGold modest but meaningful.
`

type generatedExpositionTemplateSpec struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Dialogue         []string `json:"dialogue"`
	RewardMode       string   `json:"rewardMode"`
	RandomRewardSize string   `json:"randomRewardSize"`
	RewardExperience int      `json:"rewardExperience"`
	RewardGold       int      `json:"rewardGold"`
}

type generatedExpositionTemplatesResponse struct {
	Templates []generatedExpositionTemplateSpec `json:"templates"`
}

type GenerateExpositionTemplatesProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateExpositionTemplatesProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateExpositionTemplatesProcessor {
	log.Println("Initializing GenerateExpositionTemplatesProcessor")
	return GenerateExpositionTemplatesProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateExpositionTemplatesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate exposition templates task: %v", task.Type())

	var payload jobs.GenerateExpositionTemplatesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ExpositionTemplateGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Exposition template generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ExpositionTemplateGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ExpositionTemplateGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateExpositionTemplates(ctx, job); err != nil {
		return p.failExpositionTemplateGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateExpositionTemplatesProcessor) generateExpositionTemplates(
	ctx context.Context,
	job *models.ExpositionTemplateGenerationJob,
) error {
	zoneKind, err := loadOptionalZoneKind(ctx, p.dbClient, job.ZoneKind)
	if err != nil {
		return fmt.Errorf("failed to load exposition template zone kind: %w", err)
	}
	if zoneKind == nil || strings.TrimSpace(zoneKind.Slug) == "" {
		return fmt.Errorf("exposition template generation requires a zone kind")
	}

	prompt := fmt.Sprintf(
		expositionTemplateGenerationPromptTemplate,
		job.Count,
		p.buildRecentExpositionTemplateAvoidance(ctx, zoneKind.Slug, 12),
		job.Count,
	)
	if zoneKindBlock := zoneKindInstructionBlock(zoneKind); zoneKindBlock != "" {
		prompt = strings.TrimSpace(zoneKindBlock + "\n\n" + prompt)
	}

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate exposition templates: %w", err)
	}

	generated := &generatedExpositionTemplatesResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse exposition template payload: %w", err)
	}

	createdCount := 0
	for index, spec := range generated.Templates {
		template := buildGeneratedExpositionTemplate(zoneKind.Slug, spec, index)
		if err := p.dbClient.ExpositionTemplate().Create(ctx, template); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create exposition template: %w", err)
		}
		createdCount++
	}

	job.Status = models.ExpositionTemplateGenerationStatusCompleted
	job.CreatedCount = createdCount
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	return p.dbClient.ExpositionTemplateGenerationJob().Update(ctx, job)
}

func (p *GenerateExpositionTemplatesProcessor) failExpositionTemplateGenerationJob(
	ctx context.Context,
	job *models.ExpositionTemplateGenerationJob,
	err error,
) error {
	msg := err.Error()
	job.Status = models.ExpositionTemplateGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ExpositionTemplateGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark exposition template generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func (p *GenerateExpositionTemplatesProcessor) buildRecentExpositionTemplateAvoidance(
	ctx context.Context,
	zoneKind string,
	limit int,
) string {
	templates, err := p.dbClient.ExpositionTemplate().FindAll(ctx)
	if err != nil || len(templates) == 0 {
		return "- none"
	}
	normalizedZoneKind := models.NormalizeZoneKind(zoneKind)
	lines := make([]string, 0, limit)
	for _, template := range templates {
		if normalizedZoneKind != "" && models.NormalizeZoneKind(template.ZoneKind) != normalizedZoneKind {
			continue
		}
		title := strings.TrimSpace(template.Title)
		description := strings.TrimSpace(template.Description)
		if title == "" && description == "" {
			continue
		}
		if description == "" {
			lines = append(lines, "- "+title)
		} else {
			lines = append(lines, fmt.Sprintf("- %s: %s", title, description))
		}
		if len(lines) >= limit {
			break
		}
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func buildGeneratedExpositionTemplate(
	zoneKind string,
	spec generatedExpositionTemplateSpec,
	index int,
) *models.ExpositionTemplate {
	title := strings.TrimSpace(spec.Title)
	if title == "" {
		title = fmt.Sprintf("%s Echo %d", expositionTemplateZoneKindLabel(zoneKind), index+1)
	}
	description := strings.TrimSpace(spec.Description)
	if description == "" {
		description = "A lingering fragment of local memory clings to this place."
	}

	dialogue := models.DialogueSequenceFromStringLines(spec.Dialogue)
	if len(dialogue) == 0 {
		dialogue = models.DialogueSequenceFromStringLines([]string{
			"The place remembers more than it should.",
			"Something in the air is trying to warn passersby away.",
		})
	}

	rewardMode := models.NormalizeRewardMode(spec.RewardMode)
	if rewardMode == "" {
		rewardMode = models.RewardModeRandom
	}
	randomRewardSize := models.NormalizeRandomRewardSize(spec.RandomRewardSize)
	if randomRewardSize == "" {
		randomRewardSize = models.RandomRewardSizeSmall
	}
	rewardExperience := clampInt(spec.RewardExperience, 0, 70)
	rewardGold := clampInt(spec.RewardGold, 0, 70)
	if rewardMode == models.RewardModeRandom {
		rewardExperience = 0
		rewardGold = 0
	}
	if rewardMode == models.RewardModeExplicit && rewardExperience == 0 && rewardGold == 0 {
		rewardMode = models.RewardModeRandom
	}

	return &models.ExpositionTemplate{
		ZoneKind:           zoneKind,
		Title:              title,
		Description:        description,
		Dialogue:           dialogue,
		RequiredStoryFlags: models.StringArray{},
		ImageURL:           "",
		ThumbnailURL:       "",
		RewardMode:         rewardMode,
		RandomRewardSize:   randomRewardSize,
		RewardExperience:   rewardExperience,
		RewardGold:         rewardGold,
		MaterialRewards:    models.BaseMaterialRewards{},
		ItemRewards:        models.ExpositionTemplateItemRewards{},
		SpellRewards:       models.ExpositionTemplateSpellRewards{},
	}
}

func expositionTemplateZoneKindLabel(zoneKind string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(zoneKind, "-", " "), "_", " "))
	fields := strings.Fields(normalized)
	if len(fields) == 0 {
		return "Wilderness"
	}

	for index, field := range fields {
		lower := strings.ToLower(field)
		if lower == "" {
			continue
		}
		fields[index] = strings.ToUpper(lower[:1]) + lower[1:]
	}
	return strings.Join(fields, " ")
}
