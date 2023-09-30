package models

import "gorm.io/gorm"

type UserSubmission struct {
	gorm.Model
	User          User         `json:"user"`
	UserID        uint         `json:"userId"`
	QuestionSet   QuestionSet  `json:"questionSet"`
	QuestionSetID uint         `json:"questionSetId"`
	UserAnswers   []UserAnswer `json:"userAnswers"`
}
