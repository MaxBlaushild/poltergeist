package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// GenerateSpellsBulkProcessor creates spells/techniques in the background.
type GenerateSpellsBulkProcessor struct {
	dbClient    db.DbClient
	redisClient *redis.Client
}

func NewGenerateSpellsBulkProcessor(dbClient db.DbClient, redisClient *redis.Client) GenerateSpellsBulkProcessor {
	log.Println("Initializing GenerateSpellsBulkProcessor")
	return GenerateSpellsBulkProcessor{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

func containsAnyKeyword(haystack string, keywords []string) bool {
	normalizedHaystack := strings.ToLower(haystack)
	tokenSet := map[string]struct{}{}
	for _, token := range strings.FieldsFunc(normalizedHaystack, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if token == "" {
			continue
		}
		tokenSet[token] = struct{}{}
	}

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
		if normalizedKeyword == "" {
			continue
		}
		if strings.Contains(normalizedKeyword, " ") {
			if strings.Contains(normalizedHaystack, normalizedKeyword) {
				return true
			}
			continue
		}
		if _, exists := tokenSet[normalizedKeyword]; exists {
			return true
		}
	}
	return false
}

func inferGeneratedAbilityEffects(
	spec jobs.SpellCreationSpec,
	abilityType models.SpellAbilityType,
	manaCost int,
) models.SpellEffects {
	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		spec.Name,
		spec.Description,
		spec.EffectText,
		spec.SchoolOfMagic,
	}, " ")))

	if abilityType == models.SpellAbilityTypeTechnique {
		damage := 10
		if containsAnyKeyword(text, []string{"heavy", "crush", "breaker", "slam", "assault"}) {
			damage = 14
		}
		return models.SpellEffects{
			{
				Type:   models.SpellEffectTypeDealDamage,
				Amount: damage,
			},
		}
	}

	if containsAnyKeyword(text, []string{"heal", "renew", "restore", "revive", "recovery", "vital"}) {
		amount := 12 + (manaCost / 2)
		if containsAnyKeyword(text, []string{"all", "party", "group", "aura"}) {
			return models.SpellEffects{
				{
					Type:   models.SpellEffectTypeRestoreLifeAllParty,
					Amount: amount,
				},
			}
		}
		return models.SpellEffects{
			{
				Type:   models.SpellEffectTypeRestoreLifePartyMember,
				Amount: amount,
			},
		}
	}

	if containsAnyKeyword(text, []string{"ward", "barrier", "guard", "shield", "stance", "fortify"}) {
		return models.SpellEffects{
			{
				Type: models.SpellEffectTypeApplyBeneficialStatus,
				StatusesToApply: models.ScenarioFailureStatusTemplates{
					{
						Name:            "Fortified",
						Description:     "Hardened guard improves survivability.",
						Effect:          "Increased constitution and resilience.",
						Positive:        true,
						DurationSeconds: 45,
						ConstitutionMod: 2,
					},
				},
			},
		}
	}

	if containsAnyKeyword(text, []string{"cleanse", "purge", "dispel"}) {
		return models.SpellEffects{
			{
				Type: models.SpellEffectTypeRemoveDetrimental,
				StatusesToRemove: models.StringArray{
					"poisoned",
					"burning",
					"bleeding",
				},
			},
		}
	}

	damage := 14 + (manaCost / 3)
	if damage < 10 {
		damage = 10
	}
	return models.SpellEffects{
		{
			Type:   models.SpellEffectTypeDealDamage,
			Amount: damage,
		},
	}
}

func (p *GenerateSpellsBulkProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate spells bulk task: %v", task.Type())

	var payload jobs.GenerateSpellsBulkTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal generate spells bulk payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	abilityType := string(models.NormalizeSpellAbilityType(payload.AbilityType))
	statusKey := jobs.SpellBulkStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.SpellBulkStatus{
		JobID:        payload.JobID,
		Status:       jobs.SpellBulkStatusInProgress,
		Source:       strings.TrimSpace(payload.Source),
		AbilityType:  abilityType,
		TotalCount:   payload.TotalCount,
		CreatedCount: 0,
		StartedAt:    &now,
		UpdatedAt:    now,
	}
	if status.TotalCount <= 0 {
		status.TotalCount = len(payload.Spells)
	}
	if status.Source == "" {
		status.Source = "seed_generated"
	}
	p.setStatus(ctx, statusKey, status)

	if len(payload.Spells) == 0 {
		err := fmt.Errorf("no spells provided for bulk generation")
		p.markFailed(ctx, statusKey, status, err)
		return err
	}

	for index, spec := range payload.Spells {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			if abilityType == string(models.SpellAbilityTypeTechnique) {
				name = fmt.Sprintf("Technique %d", index+1)
			} else {
				name = fmt.Sprintf("Spell %d", index+1)
			}
		}
		description := strings.TrimSpace(spec.Description)
		effectText := strings.TrimSpace(spec.EffectText)
		if effectText == "" {
			effectText = description
		}
		schoolOfMagic := strings.TrimSpace(spec.SchoolOfMagic)
		if schoolOfMagic == "" {
			schoolOfMagic = "Arcane"
		}

		manaCost := spec.ManaCost
		if manaCost < 0 {
			manaCost = 0
		}
		if abilityType == string(models.SpellAbilityTypeTechnique) {
			manaCost = 0
		}
		emptyError := ""
		spell := &models.Spell{
			Name:                  name,
			Description:           description,
			AbilityType:           models.SpellAbilityType(abilityType),
			EffectText:            effectText,
			SchoolOfMagic:         schoolOfMagic,
			ManaCost:              manaCost,
			Effects:               inferGeneratedAbilityEffects(spec, models.SpellAbilityType(abilityType), manaCost),
			ImageGenerationStatus: models.SpellImageGenerationStatusNone,
			ImageGenerationError:  &emptyError,
		}

		if err := p.dbClient.Spell().Create(ctx, spell); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to create %s %d/%d: %w", abilityType, index+1, len(payload.Spells), err)
		}

		status.CreatedCount = index + 1
		status.UpdatedAt = time.Now().UTC()
		p.setStatus(ctx, statusKey, status)
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.SpellBulkStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
	return nil
}

func (p *GenerateSpellsBulkProcessor) markFailed(ctx context.Context, statusKey string, status jobs.SpellBulkStatus, cause error) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.SpellBulkStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *GenerateSpellsBulkProcessor) setStatus(ctx context.Context, statusKey string, status jobs.SpellBulkStatus) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal spell bulk status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.SpellBulkStatusTTL).Err(); err != nil {
		log.Printf("Failed to write spell bulk status: %v", err)
	}
}
