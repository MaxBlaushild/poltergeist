package models

import "gorm.io/gorm"

type Match struct {
	gorm.Model
	QuestionSetID uint        `json:"questionSetId"`
	QuestionSet   QuestionSet `json:"questionSet"`
	HomeID        uint        `json:"homeId"`
	Home          User        `json:"home"`
	AwayID        uint        `json:"awayId"`
	Away          User        `json:"away"`
	WinnerID      *uint       `json:"winnerId"`
	Winner        User        `json:"winner"`
}
