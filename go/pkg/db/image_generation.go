package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type imageGenerationHandle struct {
	db *gorm.DB
}

func (h *imageGenerationHandle) Create(ctx context.Context, imageGeneration *models.ImageGeneration) error {
	return h.db.WithContext(ctx).Create(imageGeneration).Error
}

func (h *imageGenerationHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ImageGeneration, error) {
	var imageGeneration models.ImageGeneration
	if err := h.db.WithContext(ctx).First(&imageGeneration, id).Error; err != nil {
		return nil, err
	}
	return &imageGeneration, nil
}

func (h *imageGenerationHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.ImageGeneration, error) {
	var imageGenerations []models.ImageGeneration
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Find(&imageGenerations).Error; err != nil {
		return nil, err
	}
	return imageGenerations, nil
}

func (h *imageGenerationHandle) UpdateState(ctx context.Context, imageGenerationID uuid.UUID, state models.GenerationStatus) error {
	imgGen, err := h.FindByID(ctx, imageGenerationID)
	if err != nil {
		return err
	}

	if imgGen.Status >= state {
		return errors.New("image generation is already in this state or a more progressed state")
	}

	imgGen.Status = state
	return h.db.WithContext(ctx).Save(imgGen).Error
}

func (h *imageGenerationHandle) SetOptions(ctx context.Context, imageGenerationID uuid.UUID, options []string) error {
	imgGen, err := h.FindByID(ctx, imageGenerationID)
	if err != nil {
		return err
	}

	imgGen.OptionOne = &options[0]
	imgGen.OptionTwo = &options[1]
	imgGen.OptionThree = &options[2]
	imgGen.OptionFour = &options[3]

	return h.db.WithContext(ctx).Save(imgGen).Error
}

func (h *imageGenerationHandle) FindByState(ctx context.Context, state models.GenerationStatus) ([]models.ImageGeneration, error) {
	var imageGenerations []models.ImageGeneration
	if err := h.db.WithContext(ctx).Where("status = ?", state).Find(&imageGenerations).Error; err != nil {
		return nil, err
	}
	return imageGenerations, nil
}
