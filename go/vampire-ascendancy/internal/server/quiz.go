package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /quiz — the quiz for players. Returns questions without the answer key or
// scoring, plus whether this player has already submitted (answers lock once in).
func (s *server) getQuiz(ctx *gin.Context) {
	player := playerFromContext(ctx)

	state, err := s.dbClient.Vampire().GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	questions, err := s.dbClient.Vampire().ListQuizQuestions(ctx, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	subs, err := s.dbClient.Vampire().ListQuizSubmissionsForPlayer(ctx, player.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	answersByQ := map[string]string{}
	submitted := false
	for _, sub := range subs {
		answersByQ[sub.QuestionID.String()] = sub.Answer
		if sub.Locked {
			submitted = true
		}
	}

	out := make([]gin.H, 0, len(questions))
	for _, q := range questions {
		out = append(out, gin.H{
			"id":           q.ID,
			"ordinal":      q.Ordinal,
			"prompt":       q.Prompt,
			"questionType": q.QuestionType,
			"options":      q.Options,
			"answer":       answersByQ[q.ID.String()], // their prior answer, if any
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"quizOpen":  state.QuizOpen,
		"submitted": submitted,
		"questions": out,
	})
}

// POST /quiz/submit — submit and lock all answers at once. Multiple-choice
// answers are auto-graded and correct ones apply their House Favor effect;
// open-ended answers are stored for the GMs to read and adjudicate.
func (s *server) submitQuiz(ctx *gin.Context) {
	player := playerFromContext(ctx)

	state, err := s.dbClient.Vampire().GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !state.QuizOpen {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "the quiz is not open"})
		return
	}

	// One submission per player — reject if they already locked in.
	existing, err := s.dbClient.Vampire().ListQuizSubmissionsForPlayer(ctx, player.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, sub := range existing {
		if sub.Locked {
			ctx.JSON(http.StatusConflict, gin.H{"error": "you have already answered"})
			return
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

	// House name -> id, for applying HF effects.
	houses, err := s.dbClient.Vampire().ListHouses(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	houseByName := map[string]uuid.UUID{}
	for _, h := range houses {
		houseByName[h.Name] = h.ID
	}

	for _, a := range body.Answers {
		qid, err := uuid.Parse(a.QuestionID)
		if err != nil {
			continue
		}
		question, err := s.dbClient.Vampire().GetQuizQuestionByID(ctx, qid)
		if err != nil || question == nil {
			continue
		}

		var isCorrect *bool
		if question.QuestionType == "multiple_choice" && question.CorrectAnswer != "" {
			correct := strings.EqualFold(strings.TrimSpace(a.Answer), strings.TrimSpace(question.CorrectAnswer))
			isCorrect = &correct
		}

		if _, err := s.dbClient.Vampire().UpsertQuizSubmission(ctx, player.ID, qid, a.Answer, isCorrect, true); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Apply House Favor for a correct answer.
		if isCorrect != nil && *isCorrect && len(question.HFEffect) > 0 {
			effects := map[string]int{}
			if err := json.Unmarshal(question.HFEffect, &effects); err == nil {
				for houseName, delta := range effects {
					if delta == 0 {
						continue
					}
					houseID, ok := houseByName[houseName]
					if !ok {
						continue
					}
					_ = s.dbClient.Vampire().AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
						HouseID: houseID,
						Delta:   delta,
						Reason:  "quiz: " + question.Prompt,
						GMName:  "quiz",
						Source:  "quiz",
					})
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
