package quizgrade

import (
	"strings"
	"testing"
)

func TestParseScoreAndWhy(t *testing.T) {
	cases := []struct {
		name      string
		text      string
		maxBT     int
		wantScore int
		wantWhy   string
	}{
		{"basic", "SCORE: 8\nWHY: nailed the motive", 10, 8, "nailed the motive"},
		{"clamp high", "SCORE: 20\nWHY: perfect", 10, 10, "perfect"},
		{"clamp negative", "SCORE: -3\nWHY: irrelevant", 10, 0, "irrelevant"},
		{"why single line only", "SCORE: 4\nWHY: missed the culprit\nSCORE again", 10, 4, "missed the culprit"},
		{"missing why", "SCORE: 5", 10, 5, ""},
		{"lowercase markers", "score: 6 why: partial credit", 10, 6, "partial credit"},
		{"no score defaults zero", "WHY: no number here", 10, 0, "no number here"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotScore, gotWhy := ParseScoreAndWhy(c.text, c.maxBT)
			if gotScore != c.wantScore {
				t.Errorf("score = %d, want %d", gotScore, c.wantScore)
			}
			if gotWhy != c.wantWhy {
				t.Errorf("why = %q, want %q", gotWhy, c.wantWhy)
			}
		})
	}
}

func TestParseFirstInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"the answer is 42 tokens", 42},
		{"-7 below zero", -7},
		{"no digits at all", 0},
		{"7 then 9", 7},
	}
	for _, c := range cases {
		if got := parseFirstInt(c.in); got != c.want {
			t.Errorf("parseFirstInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestClampInt(t *testing.T) {
	cases := []struct{ n, lo, hi, want int }{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{99, 0, 10, 10},
		{0, 0, 0, 0},
	}
	for _, c := range cases {
		if got := clampInt(c.n, c.lo, c.hi); got != c.want {
			t.Errorf("clampInt(%d,%d,%d) = %d, want %d", c.n, c.lo, c.hi, got, c.want)
		}
	}
}

func TestBuildPrompt_IncludesInputs(t *testing.T) {
	p := BuildPrompt("the butler did it", "who did it?", "the butler", 10)
	for _, want := range []string{"the butler did it", "who did it?", "the butler", "0 to 10", "SCORE:", "WHY:"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}
