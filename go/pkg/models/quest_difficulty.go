package models

import "strings"

type QuestDifficultyMode string

const (
	QuestDifficultyModeScale QuestDifficultyMode = "scale"
	QuestDifficultyModeFixed QuestDifficultyMode = "fixed"
)

func NormalizeQuestDifficultyMode(raw string) QuestDifficultyMode {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(QuestDifficultyModeFixed):
		return QuestDifficultyModeFixed
	default:
		return QuestDifficultyModeScale
	}
}

func NormalizeQuestDifficulty(value int) int {
	if value < 1 {
		return 1
	}
	return value
}
