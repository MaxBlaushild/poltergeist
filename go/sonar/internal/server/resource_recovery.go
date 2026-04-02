package server

import (
	"context"
	"log"

	"github.com/google/uuid"
)

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
