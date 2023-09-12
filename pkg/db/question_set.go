package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type questionSetHandle struct {
	db *gorm.DB
}

func (q *questionSetHandle) Insert(ctx context.Context, questions []models.Question) (*models.QuestionSet, error) {
	questionSet := models.QuestionSet{}

	if err := q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&questionSet).Error; err != nil {
			return err
		}

		for _, question := range questions {
			category := models.Category{Title: question.Category.Title}

			if err := tx.Where(&category).FirstOrCreate(&category).Error; err != nil {
				return err
			}

			question := models.Question{
				Category:    category,
				Prompt:      question.Prompt,
				QuestionSet: questionSet,
				Answer:      question.Answer,
			}

			if err := tx.Create(&question).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &questionSet, nil
}
