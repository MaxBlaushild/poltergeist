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

const openEndedScenarioTemplatePromptTemplate = `
You are designing %d reusable fantasy MMORPG scenario templates.

Recent scenario templates to avoid echoing:
%s

Return JSON only:
{
  "templates": [
    {
      "prompt": "2-4 vivid sentences",
      "difficulty": 0-40,
      "rewardExperience": 0-120,
      "rewardGold": 0-120,
      "itemRewards": [
        { "inventoryItemId": <id from allowed list>, "quantity": 1-3 }
      ]
    }
  ]
}

Rules:
- Output exactly %d templates.
- These are generic templates, not tied to any specific zone, city, landmark, or coordinates.
- Each prompt should describe a clear fantasy conflict or opportunity that could fit many places.
- Keep the scenarios materially distinct from one another and from the recent templates list.
- itemRewards can be empty.
- Use only inventoryItemId values from this allowed list:
%s
`

const choiceScenarioTemplatePromptTemplate = `
You are designing %d reusable fantasy MMORPG scenario templates.

Recent scenario templates to avoid echoing:
%s

Return JSON only:
{
  "templates": [
    {
      "prompt": "2-4 vivid sentences",
      "difficulty": 0-40,
      "options": [
        {
          "optionText": "player action text",
          "successText": "one or two sentences",
          "failureText": "one or two sentences",
          "statTag": "strength|dexterity|constitution|intelligence|wisdom|charisma",
          "proficiencies": ["0-3 short proficiencies"],
          "difficulty": null or 0-40,
          "rewardExperience": 0-80,
          "rewardGold": 0-80,
          "itemRewards": [
            { "inventoryItemId": <id from allowed list>, "quantity": 1-2 }
          ]
        }
      ]
    }
  ]
}

Rules:
- Output exactly %d templates.
- These are generic templates, not tied to any specific zone, city, landmark, or coordinates.
- Each template prompt should describe a reusable fantasy situation with 3 distinct player options.
- Keep templates materially distinct from one another and from the recent templates list.
- Each template must have exactly 3 options.
- itemRewards can be empty.
- Use only inventoryItemId values from this allowed list:
%s
`

type openEndedScenarioTemplatePayload struct {
	Prompt           string                            `json:"prompt"`
	Difficulty       *int                              `json:"difficulty"`
	RewardExperience int                               `json:"rewardExperience"`
	RewardGold       int                               `json:"rewardGold"`
	ItemRewards      []scenarioGenerationRewardPayload `json:"itemRewards"`
}

type openEndedScenarioTemplatesResponse struct {
	Templates []openEndedScenarioTemplatePayload `json:"templates"`
}

type choiceScenarioTemplatePayload struct {
	Prompt     string                                  `json:"prompt"`
	Difficulty *int                                    `json:"difficulty"`
	Options    []choiceScenarioGenerationOptionPayload `json:"options"`
}

type choiceScenarioTemplatesResponse struct {
	Templates []choiceScenarioTemplatePayload `json:"templates"`
}

type GenerateScenarioTemplatesProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateScenarioTemplatesProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateScenarioTemplatesProcessor {
	log.Println("Initializing GenerateScenarioTemplatesProcessor")
	return GenerateScenarioTemplatesProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateScenarioTemplatesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate scenario templates task: %v", task.Type())

	var payload jobs.GenerateScenarioTemplatesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ScenarioTemplateGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Scenario template generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ScenarioTemplateGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ScenarioTemplateGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateScenarioTemplates(ctx, job); err != nil {
		return p.failScenarioTemplateGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateScenarioTemplatesProcessor) generateScenarioTemplates(ctx context.Context, job *models.ScenarioTemplateGenerationJob) error {
	genre, err := loadScenarioGenre(ctx, p.dbClient, job.GenreID, job.Genre)
	if err != nil {
		return fmt.Errorf("failed to load scenario template genre: %w", err)
	}
	inventoryItems, err := p.dbClient.InventoryItem().FindAllActiveInventoryItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to load inventory items: %w", err)
	}
	allowedItemIDs := make(map[int]struct{}, len(inventoryItems))
	for _, item := range inventoryItems {
		allowedItemIDs[item.ID] = struct{}{}
	}
	allowedItemsPrompt := buildAllowedItemsPrompt(inventoryItems)
	recentAvoidance := p.buildRecentScenarioTemplateAvoidance(ctx, genre, 12)

	createdCount := 0
	if job.OpenEnded {
		prompt := buildOpenEndedScenarioTemplatePrompt(
			job.Count,
			recentAvoidance,
			allowedItemsPrompt,
			genre,
		)
		answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			return fmt.Errorf("failed to generate open-ended scenario templates: %w", err)
		}
		generated := &openEndedScenarioTemplatesResponse{}
		if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
			return fmt.Errorf("failed to parse open-ended scenario template payload: %w", err)
		}
		for _, spec := range generated.Templates {
			template := &models.ScenarioTemplate{
				GenreID:                  genre.ID,
				Genre:                    genre,
				Prompt:                   sanitizeScenarioPrompt(spec.Prompt),
				ScaleWithUserLevel:       false,
				RewardMode:               models.RewardModeExplicit,
				RandomRewardSize:         models.RandomRewardSizeSmall,
				Difficulty:               sanitizeScenarioDifficulty(spec.Difficulty, 24),
				RewardExperience:         clampInt(spec.RewardExperience, 0, 120),
				RewardGold:               clampInt(spec.RewardGold, 0, 120),
				OpenEnded:                true,
				FailurePenaltyMode:       models.ScenarioFailurePenaltyModeShared,
				FailureHealthDrainType:   models.ScenarioFailureDrainTypeNone,
				FailureManaDrainType:     models.ScenarioFailureDrainTypeNone,
				FailureStatuses:          models.ScenarioFailureStatusTemplates{},
				SuccessRewardMode:        models.ScenarioSuccessRewardModeShared,
				SuccessHealthRestoreType: models.ScenarioFailureDrainTypeNone,
				SuccessManaRestoreType:   models.ScenarioFailureDrainTypeNone,
				SuccessStatuses:          models.ScenarioFailureStatusTemplates{},
				Options:                  models.ScenarioTemplateOptions{},
				ItemRewards:              scenarioItemRewardsToTemplateRewards(sanitizeScenarioRewards(spec.ItemRewards, allowedItemIDs, 3)),
				ItemChoiceRewards:        models.ScenarioTemplateRewards{},
				SpellRewards:             models.ScenarioTemplateSpellRewards{},
			}
			if template.RewardExperience == 0 && template.RewardGold == 0 && len(template.ItemRewards) == 0 {
				template.RewardMode = models.RewardModeRandom
			}
			if err := p.dbClient.ScenarioTemplate().Create(ctx, template); err != nil {
				job.CreatedCount = createdCount
				return fmt.Errorf("failed to create scenario template: %w", err)
			}
			createdCount++
		}
	} else {
		prompt := buildChoiceScenarioTemplatePrompt(
			job.Count,
			recentAvoidance,
			allowedItemsPrompt,
			genre,
		)
		answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			return fmt.Errorf("failed to generate choice scenario templates: %w", err)
		}
		generated := &choiceScenarioTemplatesResponse{}
		if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
			return fmt.Errorf("failed to parse choice scenario template payload: %w", err)
		}
		for _, spec := range generated.Templates {
			options := sanitizeScenarioOptions(spec.Options, allowedItemIDs)
			if len(options) == 0 {
				options = append(options, fallbackScenarioOption())
			}
			template := &models.ScenarioTemplate{
				GenreID:                  genre.ID,
				Genre:                    genre,
				Prompt:                   sanitizeScenarioPrompt(spec.Prompt),
				ScaleWithUserLevel:       false,
				RewardMode:               models.RewardModeRandom,
				RandomRewardSize:         models.RandomRewardSizeSmall,
				Difficulty:               sanitizeScenarioDifficulty(spec.Difficulty, 24),
				RewardExperience:         0,
				RewardGold:               0,
				OpenEnded:                false,
				FailurePenaltyMode:       models.ScenarioFailurePenaltyModeShared,
				FailureHealthDrainType:   models.ScenarioFailureDrainTypeNone,
				FailureManaDrainType:     models.ScenarioFailureDrainTypeNone,
				FailureStatuses:          models.ScenarioFailureStatusTemplates{},
				SuccessRewardMode:        models.ScenarioSuccessRewardModeShared,
				SuccessHealthRestoreType: models.ScenarioFailureDrainTypeNone,
				SuccessManaRestoreType:   models.ScenarioFailureDrainTypeNone,
				SuccessStatuses:          models.ScenarioFailureStatusTemplates{},
				Options:                  scenarioOptionsToTemplateOptions(options),
				ItemRewards:              models.ScenarioTemplateRewards{},
				ItemChoiceRewards:        models.ScenarioTemplateRewards{},
				SpellRewards:             models.ScenarioTemplateSpellRewards{},
			}
			if err := p.dbClient.ScenarioTemplate().Create(ctx, template); err != nil {
				job.CreatedCount = createdCount
				return fmt.Errorf("failed to create scenario template: %w", err)
			}
			createdCount++
		}
	}

	job.Status = models.ScenarioTemplateGenerationStatusCompleted
	job.CreatedCount = createdCount
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	return p.dbClient.ScenarioTemplateGenerationJob().Update(ctx, job)
}

func (p *GenerateScenarioTemplatesProcessor) failScenarioTemplateGenerationJob(ctx context.Context, job *models.ScenarioTemplateGenerationJob, err error) error {
	msg := err.Error()
	job.Status = models.ScenarioTemplateGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ScenarioTemplateGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark scenario template generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func (p *GenerateScenarioTemplatesProcessor) buildRecentScenarioTemplateAvoidance(
	ctx context.Context,
	genre *models.ZoneGenre,
	limit int,
) string {
	var (
		templates []models.ScenarioTemplate
		err       error
	)
	if genre != nil && genre.ID != uuid.Nil {
		templates, err = p.dbClient.ScenarioTemplate().FindRecentByGenre(
			ctx,
			genre.ID,
			limit,
		)
	} else {
		templates, err = p.dbClient.ScenarioTemplate().FindRecent(ctx, limit)
	}
	if err != nil || len(templates) == 0 {
		return "- none"
	}
	lines := make([]string, 0, len(templates))
	for _, template := range templates {
		prompt := strings.TrimSpace(template.Prompt)
		if prompt == "" {
			continue
		}
		if len(prompt) > 180 {
			prompt = strings.TrimSpace(prompt[:180]) + "..."
		}
		lines = append(lines, "- "+prompt)
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func buildOpenEndedScenarioTemplatePrompt(
	count int,
	recentAvoidance string,
	allowedItemsPrompt string,
	genre *models.ZoneGenre,
) string {
	base := fmt.Sprintf(
		openEndedScenarioTemplatePromptTemplate,
		count,
		recentAvoidance,
		count,
		allowedItemsPrompt,
	)
	if isBaselineFantasyScenarioGenre(genre) {
		return base
	}
	return strings.TrimSpace(scenarioGenreInstructionBlock(genre) + "\n" + base)
}

func buildChoiceScenarioTemplatePrompt(
	count int,
	recentAvoidance string,
	allowedItemsPrompt string,
	genre *models.ZoneGenre,
) string {
	base := fmt.Sprintf(
		choiceScenarioTemplatePromptTemplate,
		count,
		recentAvoidance,
		count,
		allowedItemsPrompt,
	)
	if isBaselineFantasyScenarioGenre(genre) {
		return base
	}
	return strings.TrimSpace(scenarioGenreInstructionBlock(genre) + "\n" + base)
}

func scenarioItemRewardsToTemplateRewards(rewards []models.ScenarioItemReward) models.ScenarioTemplateRewards {
	out := make(models.ScenarioTemplateRewards, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, models.ScenarioTemplateReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func scenarioOptionsToTemplateOptions(options []models.ScenarioOption) models.ScenarioTemplateOptions {
	out := make(models.ScenarioTemplateOptions, 0, len(options))
	for _, option := range options {
		out = append(out, models.ScenarioTemplateOption{
			OptionText:                option.OptionText,
			SuccessText:               option.SuccessText,
			FailureText:               option.FailureText,
			SuccessHandoffText:        option.SuccessHandoffText,
			FailureHandoffText:        option.FailureHandoffText,
			StatTag:                   option.StatTag,
			Proficiencies:             option.Proficiencies,
			Difficulty:                option.Difficulty,
			RewardExperience:          option.RewardExperience,
			RewardGold:                option.RewardGold,
			FailureHealthDrainType:    option.FailureHealthDrainType,
			FailureHealthDrainValue:   option.FailureHealthDrainValue,
			FailureManaDrainType:      option.FailureManaDrainType,
			FailureManaDrainValue:     option.FailureManaDrainValue,
			FailureStatuses:           option.FailureStatuses,
			SuccessHealthRestoreType:  option.SuccessHealthRestoreType,
			SuccessHealthRestoreValue: option.SuccessHealthRestoreValue,
			SuccessManaRestoreType:    option.SuccessManaRestoreType,
			SuccessManaRestoreValue:   option.SuccessManaRestoreValue,
			SuccessStatuses:           option.SuccessStatuses,
			ItemRewards:               scenarioOptionItemRewardsToTemplateRewards(option.ItemRewards),
			ItemChoiceRewards:         scenarioOptionItemChoiceRewardsToTemplateRewards(option.ItemChoiceRewards),
			SpellRewards:              scenarioOptionSpellRewardsToTemplateRewards(option.SpellRewards),
		})
	}
	return out
}

func scenarioOptionItemRewardsToTemplateRewards(rewards []models.ScenarioOptionItemReward) models.ScenarioTemplateRewards {
	out := make(models.ScenarioTemplateRewards, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, models.ScenarioTemplateReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func scenarioOptionItemChoiceRewardsToTemplateRewards(rewards []models.ScenarioOptionItemChoiceReward) models.ScenarioTemplateRewards {
	out := make(models.ScenarioTemplateRewards, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, models.ScenarioTemplateReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func scenarioOptionSpellRewardsToTemplateRewards(rewards []models.ScenarioOptionSpellReward) models.ScenarioTemplateSpellRewards {
	out := make(models.ScenarioTemplateSpellRewards, 0, len(rewards))
	for _, reward := range rewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		out = append(out, models.ScenarioTemplateSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return out
}
