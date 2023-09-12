package models

import "gorm.io/gorm"

type Crystal struct {
	gorm.Model
	Name             string `json:"name"`
	Clue             string `json:"clue"`
	CaptureChallenge string `json:"captureChallenge"`
	AttuneChallenge  string `json:"attuneChallenge"`
	Captured         bool   `json:"captured"`
	Attuned          bool   `json:"attuned"`
	Lat              string `json:"lat"`
	Lng              string `json:"lng"`
	CaptureTeamID    uint   `json:"captureTeamId"`
}

func (u *Crystal) TableName() string {
	return "crisis_crystals"
}
