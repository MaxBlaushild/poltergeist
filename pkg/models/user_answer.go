package models

import "gorm.io/gorm"

type UserAnswer struct {
	gorm.Model
	User             User           `json:"user"`
	UserID           uint           `json:"userId"`
	Question         Question       `json:"question"`
	QuestionID       uint           `json:"questionId"`
	Answer           string         `json:"answer"`
	Correct          bool           `json:"correct"`
	UserSubmission   UserSubmission `json:"userSubmission"`
	UserSubmissionID uint           `json:"userSubmissionId"`
	Points           uint           `json:"points"`
}
