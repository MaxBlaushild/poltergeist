package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
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

// enqueueGrade queues one grading job for a submission and flips it to "queued".
// asynq auto-retries are bounded (MaxRetry) so a persistent failure lands in the
// "failed" state for the GM to retry manually, rather than retrying forever.
func (s *server) enqueueGrade(ctx context.Context, p1q *models.VampireQuizQuestion, sub models.VampireQuizSubmission, maxBT int) error {
	payload, err := json.Marshal(jobs.GradeQuizSubmissionTaskPayload{
		SubmissionID: sub.ID,
		PlayerID:     sub.PlayerID,
		Prompt:       p1q.Prompt,
		Rubric:       p1q.Rubric,
		Answer:       sub.Answer,
		MaxBT:        maxBT,
	})
	if err != nil {
		return err
	}
	if _, err := s.asyncClient.Enqueue(
		asynq.NewTask(jobs.GradeQuizSubmissionTaskType, payload),
		asynq.Queue("grading"),
		asynq.MaxRetry(3),
	); err != nil {
		return err
	}
	_ = s.dbClient.Vampire().SetQuizGradeStatus(ctx, sub.ID, models.QuizGradeStatusQueued, "")
	return nil
}

// POST /gm/quiz/part1/grade — enqueue one AI-grading job per Part 1 response.
// The job-runner grades them in parallel and applies the Blood Tokens directly;
// the GM list polls for results. Grades apply automatically (no confirmation).
func (s *server) gmGradePart1(ctx *gin.Context) {
	s.gradePart1(ctx, false, nil)
}

// POST /gm/quiz/part1/regrade — re-enqueue grading. With a submissionId, retries
// just that one; otherwise retries every submission not already "graded" (i.e.
// queued/grading/failed/never — the ones that could be stuck).
func (s *server) gmRegradePart1(ctx *gin.Context) {
	var body struct {
		SubmissionID string `json:"submissionId"`
	}
	_ = ctx.ShouldBindJSON(&body)
	var only *uuid.UUID
	if body.SubmissionID != "" {
		id, err := uuid.Parse(body.SubmissionID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
			return
		}
		only = &id
	}
	s.gradePart1(ctx, true, only)
}

// gradePart1 enqueues grading jobs. incompleteOnly skips already-graded answers
// (for retries); only, when set, restricts to a single submission.
func (s *server) gradePart1(ctx *gin.Context, incompleteOnly bool, only *uuid.UUID) {
	if s.asyncClient == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "grading queue is not configured"})
		return
	}
	v := s.dbClient.Vampire()
	p1q, err := v.GetPart1Question(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p1q == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no Part 1 question configured"})
		return
	}
	maxBT := p1q.MaxBT
	if maxBT <= 0 {
		maxBT = 6
	}
	subs, err := v.ListQuizSubmissions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queued := 0
	for _, sub := range subs {
		if sub.QuestionID != p1q.ID {
			continue
		}
		if only != nil && sub.ID != *only {
			continue
		}
		if incompleteOnly && sub.GradeStatus == models.QuizGradeStatusGraded {
			continue
		}
		if err := s.enqueueGrade(ctx, p1q, sub, maxBT); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue grading: " + err.Error()})
			return
		}
		queued++
	}

	s.logGM(ctx, "grade_quiz_part1", map[string]interface{}{"queued": queued, "retry": incompleteOnly})
	ctx.JSON(http.StatusOK, gin.H{"status": "grading queued", "queued": queued})
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

