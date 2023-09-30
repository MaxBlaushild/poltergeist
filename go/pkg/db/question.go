package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type questionHandle struct {
	db *gorm.DB
}

func (q *questionHandle) FindByQuestionSetID(ctx context.Context, questionSetID uint) ([]models.Question, error) {
	questions := []models.Question{}

	if err := q.db.WithContext(ctx).Where(&models.Question{QuestionSetID: questionSetID}).Find(&questions).Error; err != nil {
		return nil, err
	}

	return questions, nil
}

func (q *questionHandle) GetAllQuestions(ctx context.Context) ([]models.Question, error) {
	questions := []models.Question{}

	if err := q.db.WithContext(ctx).Find(&questions).Error; err != nil {
		return nil, err
	}

	return questions, nil
}
