package server

import (
	"math"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

// ---- Part 1 grade parsing ----

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
			gotScore, gotWhy := parseScoreAndWhy(c.text, c.maxBT)
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

// ---- Part 2 size-normalized House Favor ----

func mcQuestion(correct string, hf float64) models.VampireQuizQuestion {
	return models.VampireQuizQuestion{ID: uuid.New(), QuestionType: "multiple_choice", CorrectAnswer: correct, HFValue: hf}
}

func ans(player, house, question uuid.UUID, answer string) db.Part2Answer {
	return db.Part2Answer{PlayerID: player, HouseID: house, QuestionID: question, Answer: answer}
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestScorePart2Favor_Normalizes(t *testing.T) {
	h1 := uuid.New()
	p1, p2 := uuid.New(), uuid.New()
	q1 := mcQuestion("A", 4)

	// 2 participants, 1 correct -> (1/2)*4 = 2.0
	got := scorePart2Favor(
		[]models.VampireQuizQuestion{q1},
		[]db.Part2Answer{ans(p1, h1, q1.ID, "A"), ans(p2, h1, q1.ID, "B")},
	)
	if !approx(got[h1], 2.0) {
		t.Fatalf("h1 favor = %v, want 2.0", got[h1])
	}
}

func TestScorePart2Favor_SumsQuestionsCaseInsensitive(t *testing.T) {
	h1 := uuid.New()
	p1, p2 := uuid.New(), uuid.New()
	q1 := mcQuestion("A", 4)
	q2 := mcQuestion("X", 2)

	got := scorePart2Favor(
		[]models.VampireQuizQuestion{q1, q2},
		[]db.Part2Answer{
			ans(p1, h1, q1.ID, "A"), ans(p2, h1, q1.ID, "B"), // q1: 1/2 * 4 = 2.0
			ans(p1, h1, q2.ID, "X"), ans(p2, h1, q2.ID, "x"), // q2: 2/2 * 2 = 2.0 (case-insensitive)
		},
	)
	if !approx(got[h1], 4.0) {
		t.Fatalf("h1 favor = %v, want 4.0", got[h1])
	}
}

func TestScorePart2Favor_SkipsNumericAndOmitsZero(t *testing.T) {
	h1, h2, h3 := uuid.New(), uuid.New(), uuid.New()
	p1, p2, p3 := uuid.New(), uuid.New(), uuid.New()
	q1 := mcQuestion("A", 4)
	qNum := models.VampireQuizQuestion{ID: uuid.New(), QuestionType: "number", HFValue: 5}

	got := scorePart2Favor(
		[]models.VampireQuizQuestion{q1, qNum},
		[]db.Part2Answer{
			ans(p1, h1, q1.ID, "A"),      // h1: 1/1 * 4 = 4.0
			ans(p2, h2, q1.ID, "wrong"),  // h2: all wrong -> omitted
			ans(p3, h3, qNum.ID, "99"),   // h3: only numeric -> not a participant -> omitted
		},
	)
	if !approx(got[h1], 4.0) {
		t.Fatalf("h1 favor = %v, want 4.0", got[h1])
	}
	if _, ok := got[h2]; ok {
		t.Fatalf("h2 should be omitted (zero correct), got %v", got[h2])
	}
	if _, ok := got[h3]; ok {
		t.Fatalf("h3 should be omitted (numeric-only, not a participant), got %v", got[h3])
	}
}

func TestScorePart2Favor_RoundsTo2dp(t *testing.T) {
	h1 := uuid.New()
	p1, p2, p3 := uuid.New(), uuid.New(), uuid.New()
	q1 := mcQuestion("A", 1)

	// 3 participants, 1 correct -> 1/3 * 1 = 0.3333 -> 0.33
	got := scorePart2Favor(
		[]models.VampireQuizQuestion{q1},
		[]db.Part2Answer{
			ans(p1, h1, q1.ID, "A"),
			ans(p2, h1, q1.ID, "B"),
			ans(p3, h1, q1.ID, "C"),
		},
	)
	if !approx(got[h1], 0.33) {
		t.Fatalf("h1 favor = %v, want 0.33", got[h1])
	}
}
