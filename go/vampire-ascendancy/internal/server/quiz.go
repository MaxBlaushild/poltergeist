package server

import (
	"encoding/json"
	"hash/fnv"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// shuffledOptions decodes a question's stored option list and returns it in a
// per-player order. The order is deterministic in (playerID, questionID): every
// player sees a different arrangement, but a given player always sees the same
// one — so reloads and re-rendered locked answers don't reshuffle. Scoring is
// unaffected because submissions are matched by option text, not position.
func shuffledOptions(raw datatypes.JSON, playerID, questionID uuid.UUID) []string {
	options := []string{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &options); err != nil {
			return []string{}
		}
	}
	h := fnv.New64a()
	h.Write(playerID[:])
	h.Write(questionID[:])
	r := rand.New(rand.NewSource(int64(h.Sum64())))
	r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	return options
}

// GET /quiz — the two-part end quiz for players. Part 1 is a single open-end
// prompt (AI-graded → BT); Part 2 is multiple choice (normalized → HF). Neither
// part ever leaks the answer key or rubric.
func (s *server) getQuiz(ctx *gin.Context) {
	player := playerFromContext(ctx)
	v := s.dbClient.Vampire()

	state, err := v.GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	subs, err := v.ListQuizSubmissionsForPlayer(ctx, player.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	subByQ := map[string]string{}
	lockedByQ := map[string]bool{}
	for _, sub := range subs {
		subByQ[sub.QuestionID.String()] = sub.Answer
		lockedByQ[sub.QuestionID.String()] = sub.Locked
	}

	// ---- Part 1 ----
	part1 := gin.H{"open": state.QuizPart1Open, "openedAt": state.QuizPart1OpenedAt, "submitted": false, "prompt": "", "answer": ""}
	p1q, err := v.GetPart1Question(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p1q != nil {
		part1["prompt"] = p1q.Prompt
		part1["answer"] = subByQ[p1q.ID.String()]
		part1["submitted"] = lockedByQ[p1q.ID.String()]
	}

	// ---- Part 2 ----
	p2qs, err := v.ListQuizQuestionsByPart(ctx, 2, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p2Out := make([]gin.H, 0, len(p2qs))
	part2Submitted := false
	for _, q := range p2qs {
		if lockedByQ[q.ID.String()] {
			part2Submitted = true
		}
		p2Out = append(p2Out, gin.H{
			"id":      q.ID,
			"ordinal": q.Ordinal,
			"prompt":  q.Prompt,
			"tier":    q.Tier,
			"type":    q.QuestionType,
			"options": shuffledOptions(q.Options, player.ID, q.ID),
			"answer":  subByQ[q.ID.String()],
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"part1": part1,
		"part2": gin.H{
			"open":      state.QuizPart2Open,
			"submitted": part2Submitted,
			"questions": p2Out,
		},
	})
}

// POST /quiz/part1/submit — lock the player's open-end response. Grading is a
// separate GM-triggered step.
func (s *server) submitQuizPart1(ctx *gin.Context) {
	player := playerFromContext(ctx)
	v := s.dbClient.Vampire()

	state, err := v.GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !state.QuizPart1Open {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "part 1 is not open"})
		return
	}

	p1q, err := v.GetPart1Question(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p1q == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "no part 1 question configured"})
		return
	}

	if s.playerLockedQuestion(ctx, player.ID, p1q.ID) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "you have already answered"})
		return
	}

	var body struct {
		Answer string `json:"answer"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := v.UpsertQuizSubmission(ctx, player.ID, p1q.ID, body.Answer, nil, true); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /quiz/part2/submit — lock all of the player's multiple-choice answers.
// Auto-grades each (is_correct); House Favor is applied later at scoring time.
func (s *server) submitQuizPart2(ctx *gin.Context) {
	player := playerFromContext(ctx)
	v := s.dbClient.Vampire()

	state, err := v.GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !state.QuizPart2Open {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "part 2 is not open"})
		return
	}

	// One submission per player.
	existing, err := v.ListQuizSubmissionsForPlayer(ctx, player.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p2qs, err := v.ListQuizQuestionsByPart(ctx, 2, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p2ByID := map[string]string{} // questionID -> correctAnswer
	for _, q := range p2qs {
		p2ByID[q.ID.String()] = q.CorrectAnswer
	}
	for _, sub := range existing {
		if sub.Locked {
			if _, ok := p2ByID[sub.QuestionID.String()]; ok {
				ctx.JSON(http.StatusConflict, gin.H{"error": "you have already answered"})
				return
			}
		}
	}

	var body struct {
		Answers []struct {
			QuestionID string `json:"questionId"`
			Answer     string `json:"answer"`
		} `json:"answers"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, a := range body.Answers {
		correct, ok := p2ByID[a.QuestionID]
		if !ok {
			continue // not a part-2 question
		}
		qid, err := uuid.Parse(a.QuestionID)
		if err != nil {
			continue
		}
		isCorrect := correct != "" && strings.EqualFold(strings.TrimSpace(a.Answer), strings.TrimSpace(correct))
		if _, err := v.UpsertQuizSubmission(ctx, player.ID, qid, a.Answer, &isCorrect, true); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *server) playerLockedQuestion(ctx *gin.Context, playerID, questionID uuid.UUID) bool {
	subs, err := s.dbClient.Vampire().ListQuizSubmissionsForPlayer(ctx, playerID)
	if err != nil {
		return false
	}
	for _, sub := range subs {
		if sub.QuestionID == questionID && sub.Locked {
			return true
		}
	}
	return false
}
