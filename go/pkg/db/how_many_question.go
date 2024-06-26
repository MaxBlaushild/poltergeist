package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type howManyQuestionHandle struct {
	db *gorm.DB
}

func (h *howManyQuestionHandle) Insert(ctx context.Context, text string, explanation string, howMany int, promptSeedIndex int, prompt string) (*models.HowManyQuestion, error) {
	howManyQuestion := models.HowManyQuestion{
		Text:            text,
		Explanation:     explanation,
		HowMany:         howMany,
		PromptSeedIndex: promptSeedIndex,
		Prompt:          prompt,
		Valid:           true,
	}

	if err := h.db.WithContext(ctx).Create(&howManyQuestion).Error; err != nil {
		return nil, err
	}

	return &howManyQuestion, nil
}

func (h *howManyQuestionHandle) FindAll(ctx context.Context) ([]*models.HowManyQuestion, error) {
	howManyQuestions := []*models.HowManyQuestion{}

	if err := h.db.WithContext(ctx).Find(&howManyQuestions).Error; err != nil {
		return nil, err
	}

	return howManyQuestions, nil
}

func (h *howManyQuestionHandle) FindById(ctx context.Context, id uuid.UUID) (*models.HowManyQuestion, error) {
	howManyQuestion := models.HowManyQuestion{}

	if err := h.db.WithContext(ctx).First(&howManyQuestion, id).Error; err != nil {
		return nil, err
	}

	return &howManyQuestion, nil
}

func (h *howManyQuestionHandle) ValidQuestionsRemaining(ctx context.Context) (int64, error) {
	var count int64

	if err := h.db.WithContext(ctx).Model(&models.HowManyQuestion{}).Where("valid = true AND done = false").Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

func (h *howManyQuestionHandle) MarkValid(ctx context.Context, howManyQuestionID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.HowManyQuestion{}).Where("id = ?", howManyQuestionID).Update("valid", true).Error
}

func (h *howManyQuestionHandle) MarkDone(ctx context.Context, howManyQuestionID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.HowManyQuestion{}).Where("id = ?", howManyQuestionID).Update("done", true).Error
}

func (h *howManyQuestionHandle) FindTodaysQuestion(ctx context.Context) (*models.HowManyQuestion, error) {
	question := models.HowManyQuestion{}

	if err := h.db.WithContext(ctx).Order("created_at desc").Where("valid = true AND done = false").First(&question).Error; err != nil {
		return nil, err
	}

	return &question, nil
}
