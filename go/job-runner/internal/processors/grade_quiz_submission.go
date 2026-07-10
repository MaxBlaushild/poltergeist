package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/quizgrade"
	"github.com/hibiken/asynq"
)

// GradeQuizSubmissionProcessor grades one Part 1 quiz answer via the LLM oracle
// and applies the resulting Blood Tokens. One job per submission, so the runner's
// worker pool grades the whole quiz in parallel (fan-out).
type GradeQuizSubmissionProcessor struct {
	dbClient   db.DbClient
	deepPriest deep_priest.DeepPriest
}

func NewGradeQuizSubmissionProcessor(dbClient db.DbClient, deepPriest deep_priest.DeepPriest) GradeQuizSubmissionProcessor {
	return GradeQuizSubmissionProcessor{dbClient: dbClient, deepPriest: deepPriest}
}

func (p *GradeQuizSubmissionProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.GradeQuizSubmissionTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal grade_quiz_submission payload: %w", err)
	}

	maxBT := payload.MaxBT
	if maxBT <= 0 {
		maxBT = 6
	}

	v := p.dbClient.Vampire()
	// State: → grading (stamps start time, bumps attempt count).
	_ = v.MarkQuizGradeStarted(ctx, payload.SubmissionID)

	// Empty answers score 0 without troubling the oracle.
	score := 0
	rationale := ""
	if answer := strings.TrimSpace(payload.Answer); answer != "" {
		prompt := quizgrade.BuildPrompt(payload.Rubric, payload.Prompt, payload.Answer, maxBT)
		ans, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			// State: → failed. Return the error so asynq retries (transient blips).
			_ = v.SetQuizGradeStatus(ctx, payload.SubmissionID, models.QuizGradeStatusFailed, err.Error())
			return fmt.Errorf("deep priest grading failed for submission %s: %w", payload.SubmissionID, err)
		}
		if ans != nil {
			score, rationale = quizgrade.ParseScoreAndWhy(ans.Answer, maxBT)
		}
	}

	sf := float64(score)
	if err := p.applyGrade(ctx, payload, score, sf, rationale); err != nil {
		_ = v.SetQuizGradeStatus(ctx, payload.SubmissionID, models.QuizGradeStatusFailed, err.Error())
		return err
	}

	// State: → graded.
	_ = v.SetQuizGradeStatus(ctx, payload.SubmissionID, models.QuizGradeStatusGraded, "")
	log.Printf("graded quiz submission %s → %d BT", payload.SubmissionID, score)
	return nil
}

// applyGrade persists the score, rationale, and Blood Token award for one graded
// submission. BT is idempotent (one Part 1 entry per player).
func (p *GradeQuizSubmissionProcessor) applyGrade(ctx context.Context, payload jobs.GradeQuizSubmissionTaskPayload, score int, sf float64, rationale string) error {
	v := p.dbClient.Vampire()
	if err := v.UpdateQuizSubmissionGrade(ctx, payload.SubmissionID, &sf, score); err != nil {
		return fmt.Errorf("failed to store grade: %w", err)
	}
	if err := v.SetQuizSubmissionRationale(ctx, payload.SubmissionID, rationale); err != nil {
		return fmt.Errorf("failed to store rationale: %w", err)
	}
	if err := v.DeleteBloodTokensBySourceForPlayer(ctx, payload.PlayerID, "quiz_part1"); err != nil {
		return fmt.Errorf("failed to clear prior BT: %w", err)
	}
	if score > 0 {
		if err := v.AddBloodTokens(ctx, &models.VampireBloodTokenLog{
			PlayerID: payload.PlayerID,
			Delta:    score,
			Reason:   "End quiz (Part 1)",
			Source:   "quiz_part1",
			GMName:   "quiz",
		}); err != nil {
			return fmt.Errorf("failed to award BT: %w", err)
		}
	}
	return nil
}
