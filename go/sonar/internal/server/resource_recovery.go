package server

import (
	"context"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const (
	monsterBattleDefeatWoundedStatusName  = "Wounded"
	monsterBattleDefeatHealthFloorPercent = 30
	monsterBattleDefeatManaFloorPercent   = 25
	monsterBattleDefeatWoundedDuration    = 15 * time.Minute
)

func monsterBattleDefeatResourceFloor(maxResource int, floorPercent int) int {
	if maxResource <= 0 {
		return 0
	}
	target := (maxResource*floorPercent + 99) / 100
	if target < 1 {
		target = 1
	}
	if target > maxResource {
		target = maxResource
	}
	return target
}

func monsterBattleDefeatWoundedStatusTemplate() models.ScenarioFailureStatusTemplate {
	return models.ScenarioFailureStatusTemplate{
		Name:            monsterBattleDefeatWoundedStatusName,
		Description:     "The monster's attack leaves you shaken and off-balance.",
		Effect:          "Combat stats are reduced while you recover from defeat.",
		EffectType:      string(models.UserStatusEffectTypeStatModifier),
		Positive:        false,
		DurationSeconds: int(monsterBattleDefeatWoundedDuration.Seconds()),
		StrengthMod:     -2,
		DexterityMod:    -2,
		IntelligenceMod: -2,
		WisdomMod:       -2,
	}
}

func (s *server) applyMonsterBattleDefeatPenalty(
	ctx context.Context,
	userID uuid.UUID,
) error {
	stats, maxHealth, maxMana, currentHealth, currentMana, err := s.getScenarioResourceState(ctx, userID)
	if err != nil {
		log.Printf("[combat][defeat-penalty][error] user=%s stage=load-before error=%v", userID, err)
		return err
	}

	targetHealth := monsterBattleDefeatResourceFloor(
		maxHealth,
		monsterBattleDefeatHealthFloorPercent,
	)
	targetMana := monsterBattleDefeatResourceFloor(
		maxMana,
		monsterBattleDefeatManaFloorPercent,
	)
	targetHealthDeficit := maxHealth - targetHealth
	targetManaDeficit := maxMana - targetMana
	healthDeficitDelta := targetHealthDeficit - stats.HealthDeficit
	manaDeficitDelta := targetManaDeficit - stats.ManaDeficit
	log.Printf(
		"[combat][defeat-penalty][attempt] user=%s currentHealth=%d currentMana=%d maxHealth=%d maxMana=%d currentHealthDeficit=%d currentManaDeficit=%d targetHealth=%d targetMana=%d healthDelta=%d manaDelta=%d",
		userID,
		currentHealth,
		currentMana,
		maxHealth,
		maxMana,
		stats.HealthDeficit,
		stats.ManaDeficit,
		targetHealth,
		targetMana,
		healthDeficitDelta,
		manaDeficitDelta,
	)
	if healthDeficitDelta != 0 || manaDeficitDelta != 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(
			ctx,
			userID,
			healthDeficitDelta,
			manaDeficitDelta,
		); err != nil {
			log.Printf("[combat][defeat-penalty][error] user=%s stage=adjust error=%v", userID, err)
			return err
		}
	}

	if _, err := s.applyMonsterBattleUserStatuses(
		ctx,
		[]uuid.UUID{userID},
		models.ScenarioFailureStatusTemplates{
			monsterBattleDefeatWoundedStatusTemplate(),
		},
	); err != nil {
		log.Printf("[combat][defeat-penalty][error] user=%s stage=status error=%v", userID, err)
		return err
	}

	_, postMaxHealth, postMaxMana, postCurrentHealth, postCurrentMana, err := s.getScenarioResourceState(ctx, userID)
	if err != nil {
		log.Printf("[combat][defeat-penalty][after-adjust-load-error] user=%s error=%v", userID, err)
		return err
	}

	log.Printf(
		"[combat][defeat-penalty][result] user=%s currentHealth=%d currentMana=%d maxHealth=%d maxMana=%d",
		userID,
		postCurrentHealth,
		postCurrentMana,
		postMaxHealth,
		postMaxMana,
	)
	return nil
}

func (s *server) restoreUserToOneHealthIfDowned(
	ctx context.Context,
	userID uuid.UUID,
) error {
	stats, maxHealth, _, currentHealth, _, err := s.getScenarioResourceState(ctx, userID)
	if err != nil {
		log.Printf("[combat][defeat-recovery][error] user=%s stage=load-before error=%v", userID, err)
		return err
	}
	if currentHealth > 0 {
		log.Printf(
			"[combat][defeat-recovery][skip] user=%s currentHealth=%d maxHealth=%d healthDeficit=%d reason=already-above-zero",
			userID,
			currentHealth,
			maxHealth,
			stats.HealthDeficit,
		)
		return nil
	}

	targetHealth := 1
	if maxHealth < targetHealth {
		targetHealth = maxHealth
	}
	if targetHealth < 0 {
		targetHealth = 0
	}

	targetHealthDeficit := maxHealth - targetHealth
	healthDeficitDelta := targetHealthDeficit - stats.HealthDeficit
	log.Printf(
		"[combat][defeat-recovery][attempt] user=%s currentHealth=%d maxHealth=%d currentDeficit=%d targetHealth=%d targetDeficit=%d deficitDelta=%d",
		userID,
		currentHealth,
		maxHealth,
		stats.HealthDeficit,
		targetHealth,
		targetHealthDeficit,
		healthDeficitDelta,
	)
	if healthDeficitDelta == 0 {
		log.Printf("[combat][defeat-recovery][noop] user=%s reason=delta-zero", userID)
		return nil
	}

	updatedStats, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(
		ctx,
		userID,
		healthDeficitDelta,
		0,
	)
	if err != nil {
		log.Printf("[combat][defeat-recovery][error] user=%s stage=adjust error=%v", userID, err)
		return err
	}

	_, postMaxHealth, _, postCurrentHealth, _, err := s.getScenarioResourceState(ctx, userID)
	if err != nil {
		log.Printf(
			"[combat][defeat-recovery][after-adjust-load-error] user=%s storedDeficit=%d error=%v",
			userID,
			updatedStats.HealthDeficit,
			err,
		)
		return err
	}

	log.Printf(
		"[combat][defeat-recovery][result] user=%s storedDeficit=%d currentHealth=%d maxHealth=%d",
		userID,
		updatedStats.HealthDeficit,
		postCurrentHealth,
		postMaxHealth,
	)
	return nil
}
