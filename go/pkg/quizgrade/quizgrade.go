// Package quizgrade holds the pure prompt-building and response-parsing logic for
// AI-graded quiz answers, shared by the enqueuer (vampire-ascendancy) and the
// worker (job-runner) so both agree on exactly how grading works.
package quizgrade

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BuildPrompt renders the grading instruction sent to the LLM oracle. The model
// is asked to score 0..maxBT on how well the answer matches the canonical rubric.
func BuildPrompt(rubric, prompt, answer string, maxBT int) string {
	return fmt.Sprintf(`You are grading a murder-mystery quiz answer for accuracy.

CANONICAL TRUTH (rubric):
%s

THE QUESTION ASKED:
%s

THE PLAYER'S ANSWER:
%s

Score the answer from 0 to %d on how accurately and completely it captures the canonical truth, where %d means it captures essentially everything and 0 means nothing relevant.

Apply any explicit scoring rules, minimum scores, or point allocations stated in the rubric above — they override your own judgment. When in doubt, grade generously.

Reply in EXACTLY this format and nothing else:
SCORE: <integer 0-%d>
WHY: <one short sentence, ~15 words max, on what the answer got right or missed>`,
		rubric, prompt, answer, maxBT, maxBT, maxBT)
}

var intRe = regexp.MustCompile(`-?\d+`)

// ParseScoreAndWhy pulls "SCORE: n" and "WHY: ..." out of the grader's reply,
// clamping the score to 0..maxBT.
func ParseScoreAndWhy(text string, maxBT int) (int, string) {
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
