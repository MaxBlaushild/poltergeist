package server

import "testing"

func TestNextMonsterBattleTurnIndex(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		entryCount   int
		alive        []bool
		want         int
	}{
		{
			name:         "advances to next alive entry",
			currentIndex: 0,
			entryCount:   3,
			alive:        []bool{true, false, true},
			want:         2,
		},
		{
			name:         "wraps around to first alive entry",
			currentIndex: 2,
			entryCount:   3,
			alive:        []bool{true, false, true},
			want:         0,
		},
		{
			name:         "keeps current when nobody is alive",
			currentIndex: 1,
			entryCount:   3,
			alive:        []bool{false, false, false},
			want:         1,
		},
		{
			name:         "normalizes invalid current index",
			currentIndex: 99,
			entryCount:   3,
			alive:        []bool{false, true, true},
			want:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextMonsterBattleTurnIndex(
				tt.currentIndex,
				tt.entryCount,
				func(index int) bool {
					if index < 0 || index >= len(tt.alive) {
						return false
					}
					return tt.alive[index]
				},
			)
			if got != tt.want {
				t.Fatalf("nextMonsterBattleTurnIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}
