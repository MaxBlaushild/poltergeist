package models

import (
	"testing"
	"time"
)

func TestQuestAcceptanceV2IsTurnedIn(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		acceptance *QuestAcceptanceV2
		want       bool
	}{
		{
			name:       "nil acceptance",
			acceptance: nil,
			want:       false,
		},
		{
			name: "closed without debrief",
			acceptance: &QuestAcceptanceV2{
				ClosedAt: &now,
			},
			want: false,
		},
		{
			name: "turned in timestamp set",
			acceptance: &QuestAcceptanceV2{
				TurnedInAt: &now,
			},
			want: true,
		},
		{
			name: "legacy debriefed timestamp set",
			acceptance: &QuestAcceptanceV2{
				DebriefedAt: &now,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.acceptance.IsTurnedIn(); got != tt.want {
				t.Fatalf("IsTurnedIn() = %v, want %v", got, tt.want)
			}
		})
	}
}
