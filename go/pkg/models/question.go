package models

import (
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	CategoryID    uint        `json:"categoryId"`
	Category      Category    `json:"category"`
	QuestionSetID uint        `json:"questionSetId"`
	QuestionSet   QuestionSet `json:"questionSet"`
	Prompt        string      `json:"prompt"`
	Answer        string      `json:"answer"`
}
