package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userCharacterRelationshipHandle struct {
	db *gorm.DB
}

func (h *userCharacterRelationshipHandle) FindByUserAndCharacterIDs(
	ctx context.Context,
	userID uuid.UUID,
	characterIDs []uuid.UUID,
) ([]models.UserCharacterRelationship, error) {
	if len(characterIDs) == 0 {
		return []models.UserCharacterRelationship{}, nil
	}
	var relationships []models.UserCharacterRelationship
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND character_id IN ?", userID, characterIDs).
		Find(&relationships).Error; err != nil {
		return nil, err
	}
	return relationships, nil
}

func (h *userCharacterRelationshipHandle) FindByUserAndCharacterID(
	ctx context.Context,
	userID uuid.UUID,
	characterID uuid.UUID,
) (*models.UserCharacterRelationship, error) {
	var relationship models.UserCharacterRelationship
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND character_id = ?", userID, characterID).
		First(&relationship).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &relationship, nil
}

func (h *userCharacterRelationshipHandle) ApplyDelta(
	ctx context.Context,
	userID uuid.UUID,
	characterID uuid.UUID,
	delta models.CharacterRelationshipDelta,
) (*models.UserCharacterRelationship, error) {
	now := time.Now()
	relationship, err := h.FindByUserAndCharacterID(ctx, userID, characterID)
	if err != nil {
		return nil, err
	}
	if relationship == nil {
		relationship = &models.UserCharacterRelationship{
			ID:          uuid.New(),
			CreatedAt:   now,
			UpdatedAt:   now,
			UserID:      userID,
			CharacterID: characterID,
		}
	}
	relationship.Trust = clampCharacterRelationshipValue(relationship.Trust + delta.Trust)
	relationship.Respect = clampCharacterRelationshipValue(relationship.Respect + delta.Respect)
	relationship.Fear = clampCharacterRelationshipValue(relationship.Fear + delta.Fear)
	relationship.Debt = clampCharacterRelationshipValue(relationship.Debt + delta.Debt)
	relationship.UpdatedAt = now

	if relationship.CreatedAt.IsZero() {
		relationship.CreatedAt = now
	}
	if err := h.db.WithContext(ctx).Save(relationship).Error; err != nil {
		return nil, err
	}
	return relationship, nil
}

func clampCharacterRelationshipValue(value int) int {
	return normalizeCharacterRelationshipValue(value)
}
