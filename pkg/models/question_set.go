package models

import "gorm.io/gorm"

type QuestionSet struct {
	gorm.Model
	Questions []Question `json:"questions"`
}
