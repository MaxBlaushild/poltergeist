package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type RebalanceSpellDamageProcessor struct {
	dbClient    db.DbClient
	redisClient *redis.Client
}

func NewRebalanceSpellDamageProcessor(
	dbClient db.DbClient,
	redisClient *redis.Client,
) RebalanceSpellDamageProcessor {
	log.Println("Initializing RebalanceSpellDamageProcessor")
	return RebalanceSpellDamageProcessor{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

func (p *RebalanceSpellDamageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing rebalance spell damage task: %v", task.Type())

	var payload jobs.RebalanceSpellDamageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	statusKey := jobs.SpellDamageRebalanceStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.SpellDamageRebalanceStatus{
		JobID:      payload.JobID,
		Status:     jobs.SpellDamageRebalanceStatusInProgress,
		SpellIDs:   payload.SpellIDs,
		StartedAt:  &now,
		UpdatedAt:  now,
		TotalCount: 0,
	}
	p.setStatus(ctx, statusKey, status)

	spells, err := p.dbClient.Spell().FindAll(ctx)
	if err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to load spells: %w", err)
	}

	selected := make(map[uuid.UUID]struct{}, len(payload.SpellIDs))
	for _, id := range payload.SpellIDs {
		selected[id] = struct{}{}
	}

	targets := make([]models.Spell, 0, len(spells))
	for _, spell := range spells {
		if models.NormalizeSpellAbilityType(string(spell.AbilityType)) != models.SpellAbilityTypeSpell {
			continue
		}
		if len(selected) > 0 {
			if _, ok := selected[spell.ID]; !ok {
				continue
			}
		} else if len(spell.ProgressionLinks) == 0 {
			continue
		}
		targets = append(targets, spell)
	}

	status.TotalCount = len(targets)
	status.UpdatedAt = time.Now().UTC()
	p.setStatus(ctx, statusKey, status)

	for _, spell := range targets {
		nextEffects, changed := rebalanceSpellDamageEffectsToLegacyBaseline(spell)
		if !changed {
			status.UpdatedAt = time.Now().UTC()
			p.setStatus(ctx, statusKey, status)
			continue
		}

		if err := p.dbClient.Spell().Update(ctx, spell.ID, map[string]interface{}{
			"effects":     nextEffects,
			"effect_text": buildRebalancedSpellEffectText(nextEffects),
		}); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to update spell %s: %w", spell.ID, err)
		}
		status.UpdatedCount++
		status.UpdatedAt = time.Now().UTC()
		p.setStatus(ctx, statusKey, status)
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.SpellDamageRebalanceStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
	return nil
}

func (p *RebalanceSpellDamageProcessor) markFailed(
	ctx context.Context,
	statusKey string,
	status jobs.SpellDamageRebalanceStatus,
	cause error,
) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.SpellDamageRebalanceStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *RebalanceSpellDamageProcessor) setStatus(
	ctx context.Context,
	statusKey string,
	status jobs.SpellDamageRebalanceStatus,
) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal spell damage rebalance status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.SpellDamageRebalanceStatusTTL).Err(); err != nil {
		log.Printf("Failed to write spell damage rebalance status: %v", err)
	}
}

func rebalanceSpellDamageEffectsToLegacyBaseline(spell models.Spell) (models.SpellEffects, bool) {
	nextEffects := make(models.SpellEffects, 0, len(spell.Effects))
	changed := false
	levelBand := normalizeRebalanceSpellBand(spell.AbilityLevel)

	for _, effect := range spell.Effects {
		next := effect
		switch effect.Type {
		case models.SpellEffectTypeDealDamage:
			target := legacySpellProgressionTargetAmount(effect.Type, levelBand, models.SpellAbilityTypeSpell)
			if next.Amount != target {
				next.Amount = target
				changed = true
			}
		case models.SpellEffectTypeDealDamageAllEnemies:
			target := legacySpellProgressionTargetAmount(effect.Type, levelBand, models.SpellAbilityTypeSpell)
			if next.Amount != target {
				next.Amount = target
				changed = true
			}
		}

		if len(effect.StatusesToApply) > 0 {
			nextStatuses := make(models.ScenarioFailureStatusTemplates, 0, len(effect.StatusesToApply))
			for _, status := range effect.StatusesToApply {
				nextStatus := status
				targetTick := legacySpellProgressionDamagePerTick(levelBand, models.SpellAbilityTypeSpell)
				if status.DamagePerTick != 0 && nextStatus.DamagePerTick != targetTick {
					nextStatus.DamagePerTick = targetTick
					changed = true
				}
				nextStatuses = append(nextStatuses, nextStatus)
			}
			if !reflect.DeepEqual(effect.StatusesToApply, nextStatuses) {
				next.StatusesToApply = nextStatuses
			}
		}
		nextEffects = append(nextEffects, next)
	}

	if !changed && reflect.DeepEqual(spell.Effects, nextEffects) {
		return spell.Effects, false
	}
	return nextEffects, changed
}

func normalizeRebalanceSpellBand(level int) int {
	bands := []int{10, 25, 50, 70}
	if level <= bands[0] {
		return bands[0]
	}
	if level >= bands[len(bands)-1] {
		return bands[len(bands)-1]
	}
	best := bands[0]
	bestDistance := absInt(level - best)
	for _, candidate := range bands[1:] {
		distance := absInt(level - candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func legacySpellProgressionTargetAmount(
	effectType models.SpellEffectType,
	levelBand int,
	abilityType models.SpellAbilityType,
) int {
	normalizedBand := normalizeRebalanceSpellBand(levelBand)
	if effectType == models.SpellEffectTypeDealDamage {
		damagePerLevel := 5
		if abilityType == models.SpellAbilityTypeTechnique {
			damagePerLevel = 4
		}
		return maxInt(1, normalizedBand*damagePerLevel)
	}
	if effectType == models.SpellEffectTypeDealDamageAllEnemies {
		damagePerLevel := 4
		if abilityType == models.SpellAbilityTypeTechnique {
			damagePerLevel = 3
		}
		return maxInt(1, normalizedBand*damagePerLevel)
	}
	return 0
}

func legacySpellProgressionDamagePerTick(
	levelBand int,
	abilityType models.SpellAbilityType,
) int {
	directDamageTarget := legacySpellProgressionTargetAmount(
		models.SpellEffectTypeDealDamage,
		levelBand,
		abilityType,
	)
	return maxInt(1, int(float64(directDamageTarget)*0.2+0.5))
}

func buildRebalancedSpellEffectText(effects models.SpellEffects) string {
	if len(effects) == 0 {
		return "A refined magical technique."
	}

	effect := effects[0]
	switch effect.Type {
	case models.SpellEffectTypeRestoreLifePartyMember:
		return fmt.Sprintf("Restore %d health to one ally.", maxInt(effect.Amount, 1))
	case models.SpellEffectTypeRestoreLifeAllParty:
		return fmt.Sprintf("Restore %d health to all allies.", maxInt(effect.Amount, 1))
	case models.SpellEffectTypeApplyBeneficialStatus:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to allies.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies beneficial statuses to allies."
	case models.SpellEffectTypeApplyDetrimentalStatus:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to one enemy.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies detrimental statuses to one enemy."
	case models.SpellEffectTypeApplyDetrimentalAll:
		if len(effect.StatusesToApply) > 0 && strings.TrimSpace(effect.StatusesToApply[0].Name) != "" {
			return fmt.Sprintf("Applies %s to all enemies.", strings.TrimSpace(effect.StatusesToApply[0].Name))
		}
		return "Applies detrimental statuses to all enemies."
	case models.SpellEffectTypeRemoveDetrimental:
		return "Removes detrimental statuses from allies."
	case models.SpellEffectTypeDealDamageAllEnemies:
		affinity := "magical"
		if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
			affinity = strings.TrimSpace(*effect.DamageAffinity)
		}
		if maxInt(effect.Hits, 1) > 1 {
			return fmt.Sprintf("Deals %d %s damage to all enemies %d times.", maxInt(effect.Amount, 1), affinity, maxInt(effect.Hits, 1))
		}
		return fmt.Sprintf("Deals %d %s damage to all enemies.", maxInt(effect.Amount, 1), affinity)
	default:
		affinity := "magical"
		if effect.DamageAffinity != nil && strings.TrimSpace(*effect.DamageAffinity) != "" {
			affinity = strings.TrimSpace(*effect.DamageAffinity)
		}
		if maxInt(effect.Hits, 1) > 1 {
			return fmt.Sprintf("Deals %d %s damage to a target %d times.", maxInt(effect.Amount, 1), affinity, maxInt(effect.Hits, 1))
		}
		return fmt.Sprintf("Deals %d %s damage to a target.", maxInt(effect.Amount, 1), affinity)
	}
}
