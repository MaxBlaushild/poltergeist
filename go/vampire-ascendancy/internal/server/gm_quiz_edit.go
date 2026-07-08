package server

import (
	"encoding/json"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// GET /gm/quiz/questions — the editable quiz: the Part 1 open-end prompt/rubric
// and the Part 2 multiple-choice questions. The numeric "Blood Tokens on hand"
// question is managed automatically and isn't listed here.
func (s *server) gmGetQuizQuestions(ctx *gin.Context) {
	v := s.dbClient.Vampire()

	part1 := gin.H{"prompt": "", "rubric": "", "maxBt": 6}
	if p1, _ := v.GetPart1Question(ctx); p1 != nil {
		part1 = gin.H{"prompt": p1.Prompt, "rubric": p1.Rubric, "maxBt": p1.MaxBT}
	}

	p2qs, err := v.ListQuizQuestionsByPart(ctx, 2, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	part2 := make([]gin.H, 0, len(p2qs))
	for _, q := range p2qs {
		if q.QuestionType != "multiple_choice" {
			continue
		}
		var opts []string
		_ = json.Unmarshal(q.Options, &opts)
		part2 = append(part2, gin.H{
			"ordinal":       q.Ordinal,
			"prompt":        q.Prompt,
			"options":       opts,
			"correctAnswer": q.CorrectAnswer,
			"hfValue":       q.HFValue,
			"tier":          q.Tier,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"part1": part1, "part2": part2})
}

// PUT /gm/quiz/questions — replace the quiz from the editor. Rebuilds Part 1 and
// the Part 2 multiple-choice set, and preserves any numeric questions (the
// Blood-Tokens-on-hand question) so editing MC doesn't drop them. Replacing the
// question set clears existing quiz answers, so this is a pre-quiz operation.
func (s *server) gmUpdateQuizQuestions(ctx *gin.Context) {
	var body struct {
		Part1 struct {
			Prompt string `json:"prompt"`
			Rubric string `json:"rubric"`
			MaxBt  int    `json:"maxBt"`
		} `json:"part1"`
		Part2 []struct {
			Prompt        string   `json:"prompt"`
			Options       []string `json:"options"`
			CorrectAnswer string   `json:"correctAnswer"`
			HFValue       float64  `json:"hfValue"`
			Tier          string   `json:"tier"`
		} `json:"part2"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v := s.dbClient.Vampire()
	questions := make([]models.VampireQuizQuestion, 0, len(body.Part2)+2)

	maxBT := body.Part1.MaxBt
	if maxBT <= 0 {
		maxBT = 6
	}
	questions = append(questions, models.VampireQuizQuestion{
		Part:         1,
		Ordinal:      0,
		Prompt:       body.Part1.Prompt,
		QuestionType: "open",
		Rubric:       body.Part1.Rubric,
		MaxBT:        maxBT,
		Active:       true,
	})

	ord := 1
	for _, q := range body.Part2 {
		opts, _ := json.Marshal(q.Options)
		if len(q.Options) == 0 {
			opts = []byte("[]")
		}
		questions = append(questions, models.VampireQuizQuestion{
			Part:          2,
			Ordinal:       ord,
			Prompt:        q.Prompt,
			QuestionType:  "multiple_choice",
			Options:       opts,
			CorrectAnswer: q.CorrectAnswer,
			HFValue:       q.HFValue,
			Tier:          q.Tier,
			Active:        true,
		})
		ord++
	}

	// Preserve numeric questions (e.g. Blood Tokens on hand), appended after MC.
	if existing, err := v.ListQuizQuestionsByPart(ctx, 2, true); err == nil {
		for _, q := range existing {
			if q.QuestionType != "multiple_choice" {
				questions = append(questions, models.VampireQuizQuestion{
					Part:         2,
					Ordinal:      ord,
					Prompt:       q.Prompt,
					QuestionType: q.QuestionType,
					Options:      datatypes.JSON("[]"),
					Tier:         q.Tier,
					Active:       true,
				})
				ord++
			}
		}
	}

	if err := v.ReplaceQuizQuestions(ctx, questions); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "update_quiz_questions", map[string]interface{}{"part2Count": len(body.Part2)})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
