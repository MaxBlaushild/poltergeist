package db

import (
	"context"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type questArchetypeNodeHandle struct {
	db *gorm.DB
}

func (h *questArchetypeNodeHandle) Create(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error {
	if questArchetypeNode != nil && len(questArchetypeNode.Challenges) > 0 {
		log.Printf(
			"[main-story-convert][quest-archetype-node][create] node=%s has %d in-memory challenges; omitting associations on save",
			questArchetypeNode.ID.String(),
			len(questArchetypeNode.Challenges),
		)
	}
	return h.db.WithContext(ctx).
		Omit(clause.Associations).
		Create(questArchetypeNode).Error
}

func (h *questArchetypeNodeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNode, error) {
	var questArchetypeNode models.QuestArchetypeNode
	if err := h.db.WithContext(ctx).
		Preload("ChallengeTemplate").
		Preload("Challenges").
		Preload("Challenges.ChallengeTemplate").
		Preload("FetchCharacter").
		Preload("FetchCharacterTemplate").
		Preload("LocationArchetype").
		Preload("ScenarioTemplate").
		Preload("ExpositionTemplate").
		First(&questArchetypeNode, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &questArchetypeNode, nil
}

func (h *questArchetypeNodeHandle) FindAll(ctx context.Context) ([]*models.QuestArchetypeNode, error) {
	var questArchetypeNodes []*models.QuestArchetypeNode
	if err := h.db.WithContext(ctx).
		Preload("ChallengeTemplate").
		Preload("Challenges").
		Preload("Challenges.ChallengeTemplate").
		Preload("FetchCharacter").
		Preload("FetchCharacterTemplate").
		Preload("LocationArchetype").
		Preload("ScenarioTemplate").
		Preload("ExpositionTemplate").
		Find(&questArchetypeNodes).Error; err != nil {
		return nil, err
	}
	return questArchetypeNodes, nil
}

func (h *questArchetypeNodeHandle) Update(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error {
	if questArchetypeNode != nil && len(questArchetypeNode.Challenges) > 0 {
		log.Printf(
			"[main-story-convert][quest-archetype-node][update] node=%s has %d in-memory challenges; omitting associations on save",
			questArchetypeNode.ID.String(),
			len(questArchetypeNode.Challenges),
		)
	}
	return h.db.WithContext(ctx).
		Omit(clause.Associations).
		Save(questArchetypeNode).Error
}

func (h *questArchetypeNodeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetypeNode{}, "id = ?", id).Error
}
