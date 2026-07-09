package server

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// POST /gm/quiz/part1 — open or close Part 1 (the open-end round). Opening it
// stamps the start time (for the player countdown) and ensures Part 2 is closed.
func (s *server) gmSetPart1Open(ctx *gin.Context) {
	var body struct {
		Open bool `json:"open"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{"quiz_part1_open": body.Open}
	if body.Open {
		now := time.Now()
		updates["quiz_part1_opened_at"] = now
		updates["quiz_part2_open"] = false // the two parts are never open at once
	}
	state, err := s.dbClient.Vampire().UpdateGameState(ctx, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "set_quiz_part1_open", map[string]interface{}{"open": body.Open})
	ctx.JSON(http.StatusOK, gameStateResponse(state))
}

// POST /gm/quiz/part2 — open or close Part 2 (the MC round). Part 2 may not open
// while Part 1 is still open (its options would bias Part 1). Closing Part 2
// auto-scores it into House Favor.
func (s *server) gmSetPart2Open(ctx *gin.Context) {
	var body struct {
		Open bool `json:"open"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Part 2 can open independently of Part 1's state, so GMs can start the MC
	// round while they finish reviewing Part 1 scores.
	state, err := s.dbClient.Vampire().UpdateGameState(ctx, map[string]interface{}{"quiz_part2_open": body.Open})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Closing Part 2 finalizes scoring.
	if !body.Open {
		if err := s.scorePart2(ctx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	s.logGM(ctx, "set_quiz_part2_open", map[string]interface{}{"open": body.Open})
	ctx.JSON(http.StatusOK, gameStateResponse(state))
}

// POST /gm/quiz/part2/rescore — recompute Part 2 House Favor from the submitted
// answers (idempotent; replaces the prior Part 2 ledger entries).
func (s *server) gmRescorePart2(ctx *gin.Context) {
	if err := s.scorePart2(ctx); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "rescore_quiz_part2", map[string]interface{}{})
	standings, _ := s.dbClient.Vampire().Leaderboard(ctx)
	ctx.JSON(http.StatusOK, gin.H{"standings": standings})
}

// scorePart2 computes each house's Part 2 House Favor using the size-normalized
// formula and writes one ledger entry per house (replacing any prior ones).
//
//	house HF for a question = (correct answerers in house ÷ house participants) × question HF
//	house Part 2 HF         = sum of the above over all questions
//
// Participants = players in the house who submitted ≥1 answer. A house with zero
// participants earns 0 (and gets no entry).
func (s *server) scorePart2(ctx *gin.Context) error {
	v := s.dbClient.Vampire()

	questions, err := v.ListQuizQuestionsByPart(ctx, 2, true)
	if err != nil {
		return err
	}
	answers, err := v.ListPart2Answers(ctx)
	if err != nil {
		return err
	}

	favorByHouse := scorePart2Favor(questions, answers)

	if err := v.DeleteHouseFavorBySource(ctx, "quiz_part2"); err != nil {
		return err
	}

	for houseID, total := range favorByHouse {
		if total == 0 {
			continue
		}
		if err := v.AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
			HouseID: houseID,
			Delta:   total,
			Reason:  "End quiz (Part 2)",
			GMName:  "quiz",
			Source:  "quiz_part2",
		}); err != nil {
			return err
		}
	}
	return nil
}

// scorePart2Favor computes each house's size-normalized Part 2 House Favor (pure,
// so it can be unit-tested):
//
//	house HF for a question = (correct answerers in house ÷ house participants) × question HF
//	house Part 2 HF         = sum of the above over all questions, rounded to 2 dp
//
// Only multiple-choice questions score; the numeric "Blood Tokens on hand"
// question is skipped so it doesn't inflate the participant count. Houses with
// zero participants (or a zero total) are omitted from the result.
func scorePart2Favor(questions []models.VampireQuizQuestion, answers []db.Part2Answer) map[uuid.UUID]float64 {
	type q2 struct {
		correct string
		hf      float64
	}
	qmap := map[uuid.UUID]q2{}
	for _, q := range questions {
		if q.QuestionType != "multiple_choice" {
			continue
		}
		qmap[q.ID] = q2{correct: q.CorrectAnswer, hf: q.HFValue}
	}

	participants := map[uuid.UUID]map[uuid.UUID]bool{} // house -> set(player)
	correct := map[uuid.UUID]map[uuid.UUID]int{}       // house -> question -> #correct
	for _, a := range answers {
		qq, ok := qmap[a.QuestionID]
		if !ok {
			continue
		}
		if participants[a.HouseID] == nil {
			participants[a.HouseID] = map[uuid.UUID]bool{}
		}
		participants[a.HouseID][a.PlayerID] = true
		if qq.correct != "" && strings.EqualFold(strings.TrimSpace(a.Answer), strings.TrimSpace(qq.correct)) {
			if correct[a.HouseID] == nil {
				correct[a.HouseID] = map[uuid.UUID]int{}
			}
			correct[a.HouseID][a.QuestionID]++
		}
	}

	out := map[uuid.UUID]float64{}
	for houseID, players := range participants {
		n := len(players)
		if n == 0 {
			continue
		}
		total := 0.0
		for qid, qq := range qmap {
			c := 0
			if correct[houseID] != nil {
				c = correct[houseID][qid]
			}
			total += (float64(c) / float64(n)) * qq.hf
		}
		// Round to 2 decimals to keep the ledger tidy.
		total = float64(int(total*100+0.5)) / 100
		if total == 0 {
			continue
		}
		out[houseID] = total
	}
	return out
}

// POST /gm/quiz/part1/grade — kick off AI grading of every Part 1 response in
// the background (each response → Blood Tokens). The GM list polls for results.
func (s *server) gmGradePart1(ctx *gin.Context) {
	s.gradingMu.Lock()
	if s.grading {
		s.gradingMu.Unlock()
		ctx.JSON(http.StatusOK, gin.H{"status": "already grading"})
		return
	}
	s.grading = true
	s.gradingMu.Unlock()

	s.logGM(ctx, "grade_quiz_part1", map[string]interface{}{})

	// Run in the background so the request doesn't block on many LLM calls.
	go func() {
		defer func() {
			s.gradingMu.Lock()
			s.grading = false
			s.gradingMu.Unlock()
		}()
		s.gradePart1(context.Background())
	}()

	ctx.JSON(http.StatusOK, gin.H{"status": "grading started"})
}

func (s *server) gradePart1(ctx context.Context) {
	v := s.dbClient.Vampire()
	p1q, err := v.GetPart1Question(ctx)
	if err != nil || p1q == nil {
		return
	}
	maxBT := p1q.MaxBT
	if maxBT <= 0 {
		maxBT = 6
	}

	subs, err := v.ListQuizSubmissions(ctx)
	if err != nil {
		return
	}
	for _, sub := range subs {
		if sub.QuestionID != p1q.ID {
			continue
		}
		score, rationale := s.aiGradePart1(p1q, sub.Answer, maxBT)
		sf := float64(score)
		_ = v.UpdateQuizSubmissionGrade(ctx, sub.ID, &sf, score)
		_ = v.SetQuizSubmissionRationale(ctx, sub.ID, rationale)
		// Record the BT idempotently (one entry per player for Part 1).
		_ = v.DeleteBloodTokensBySourceForPlayer(ctx, sub.PlayerID, "quiz_part1")
		if score > 0 {
			_ = v.AddBloodTokens(ctx, &models.VampireBloodTokenLog{
				PlayerID: sub.PlayerID,
				Delta:    score,
				Reason:   "End quiz (Part 1)",
				Source:   "quiz_part1",
				GMName:   "quiz",
			})
		}
	}
}

func (s *server) aiGradePart1(q *models.VampireQuizQuestion, answer string, maxBT int) (int, string) {
	if strings.TrimSpace(answer) == "" || s.deepPriest == nil {
		return 0, ""
	}
	prompt := fmt.Sprintf(`You are grading a murder-mystery quiz answer for accuracy.

CANONICAL TRUTH (rubric):
%s

THE QUESTION ASKED:
%s

THE PLAYER'S ANSWER:
%s

Score the answer from 0 to %d on how accurately and completely it captures the canonical truth, where %d means it captures essentially everything and 0 means nothing relevant.

Reply in EXACTLY this format and nothing else:
SCORE: <integer 0-%d>
WHY: <one short sentence, ~15 words max, on what the answer got right or missed>`,
		q.Rubric, q.Prompt, answer, maxBT, maxBT, maxBT)

	ans, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil || ans == nil {
		return 0, "" // graceful: a GM can set the BT manually
	}
	return parseScoreAndWhy(ans.Answer, maxBT)
}

// parseScoreAndWhy pulls "SCORE: n" and "WHY: ..." out of the grader's reply.
func parseScoreAndWhy(text string, maxBT int) (int, string) {
	scoreText := text
	if i := strings.Index(strings.ToUpper(text), "SCORE:"); i >= 0 {
		scoreText = text[i+len("SCORE:"):]
	}
	score := clampInt(parseFirstInt(scoreText), 0, maxBT)

	rationale := ""
	if i := strings.Index(strings.ToUpper(text), "WHY:"); i >= 0 {
		rationale = strings.TrimSpace(text[i+len("WHY:"):])
		if nl := strings.IndexAny(rationale, "\r\n"); nl >= 0 {
			rationale = strings.TrimSpace(rationale[:nl])
		}
	}
	return score, rationale
}

// POST /gm/quiz/part1/override — a GM adjusts the BT awarded for one response.
func (s *server) gmOverridePart1BT(ctx *gin.Context) {
	var body struct {
		SubmissionID string `json:"submissionId"`
		AwardedBT    int    `json:"awardedBt"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	subID, err := uuid.Parse(body.SubmissionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	v := s.dbClient.Vampire()
	subs, err := v.ListQuizSubmissions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var playerID *uuid.UUID
	for _, sub := range subs {
		if sub.ID == subID {
			pid := sub.PlayerID
			playerID = &pid
			break
		}
	}
	if playerID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}

	if err := v.UpdateQuizSubmissionGrade(ctx, subID, nil, body.AwardedBT); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = v.DeleteBloodTokensBySourceForPlayer(ctx, *playerID, "quiz_part1")
	if body.AwardedBT > 0 {
		_ = v.AddBloodTokens(ctx, &models.VampireBloodTokenLog{
			PlayerID: *playerID,
			Delta:    body.AwardedBT,
			Reason:   "End quiz (Part 1, GM-adjusted)",
			Source:   "quiz_part1",
			GMName:   gmNameFromContext(ctx),
		})
	}
	s.logGM(ctx, "override_quiz_part1_bt", map[string]interface{}{"submissionId": body.SubmissionID, "awardedBt": body.AwardedBT})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /gm/quiz/submissions — all quiz answers (both parts) with context.
func (s *server) gmListQuizSubmissions(ctx *gin.Context) {
	details, err := s.dbClient.Vampire().ListQuizSubmissionsDetailed(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"submissions": details})
}

var intRe = regexp.MustCompile(`-?\d+`)

func parseFirstInt(s string) int {
	m := intRe.FindString(s)
	if m == "" {
		return 0
	}
	n, _ := strconv.Atoi(m)
	return n
}

func clampInt(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
