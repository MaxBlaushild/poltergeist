package models

import "gorm.io/gorm"

type HowManyQuestion struct {
	gorm.Model
	Text        string `json:"text"`
	HowMany     int    `json:"howMany"`
	Explanation string `json:"explanation"`
	Valid       bool   `json:"valid"`
	Done        bool   `json:"done"`
}

func (h *HowManyQuestion) TableName() string {
	return "how_many_qs"
}
