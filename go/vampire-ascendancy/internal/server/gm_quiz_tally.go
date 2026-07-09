package server

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// tallyRow is one character's final Blood Token standing, with the item effects
// that moved it. Base tally = self-reported "on hand" + Part 1 quiz BT (the
// existing rule); item effects are adjustments layered on top.
type tallyRow struct {
	PlayerID   string   `json:"playerId"`
	Character  string   `json:"character"`
	House      string   `json:"house"`
	Correct    int      `json:"correct"`
	McTotal    int      `json:"mcTotal"`
	QuizBt     int      `json:"quizBt"`
	PhysicalBt int      `json:"physicalBt"`
	ItemBt     int      `json:"itemBt"`
	FinalBt    int      `json:"finalBt"`
	Notes      []string `json:"notes"`
}

var digits = regexp.MustCompile(`[^0-9]`)

// computeBtTally loads the data the resolution engine needs and delegates the
// (pure) computation to resolveTally.
func (s *server) computeBtTally(ctx context.Context) ([]tallyRow, error) {
	v := s.dbClient.Vampire()

	players, err := v.ListPlayers(ctx)
	if err != nil {
		return nil, err
	}
	subs, err := v.ListQuizSubmissionsDetailed(ctx)
	if err != nil {
		return nil, err
	}
	gameBt := map[string]int{}
	if totals, err := v.BloodTokenTotalsBySource(ctx, "game"); err == nil {
		for _, t := range totals {
			gameBt[t.PlayerID.String()] = t.Total
		}
	}
	pis, err := v.ListAllPlayerItems(ctx)
	if err != nil {
		return nil, err
	}
	return resolveTally(players, subs, gameBt, pis), nil
}

// resolveTally is the Blood Token resolution engine, kept pure (no DB) so it can
// be unit-tested. Precedence for the targeted steals/deductions: strip-resistance
// overrides everything, else the target's immunity cancels the loss, else the
// target's reflect bounces it back to the attacker; self-additive effects (flat,
// quiz %, double-games) are order independent and applied separately.
func resolveTally(
	players []models.VampirePlayer,
	subs []db.QuizSubmissionDetail,
	gameBt map[string]int,
	pis []models.VampirePlayerItem,
) []tallyRow {
	rows := map[string]*tallyRow{}
	hasSub := map[string]bool{}
	nameByID := map[string]string{}
	for _, p := range players {
		if p.Character == nil || p.Character.Name == "" {
			continue
		}
		house := ""
		if p.Character.House != nil {
			house = p.Character.House.Name
		}
		rows[p.ID.String()] = &tallyRow{
			PlayerID:  p.ID.String(),
			Character: p.Character.Name,
			House:     house,
			Notes:     []string{},
		}
		nameByID[p.ID.String()] = p.Character.Name
	}
	name := func(id uuid.UUID) string {
		if n, ok := nameByID[id.String()]; ok {
			return n
		}
		return "someone"
	}

	// Base tally from quiz submissions.
	for _, sub := range subs {
		r := rows[sub.PlayerID.String()]
		if r == nil {
			continue
		}
		hasSub[sub.PlayerID.String()] = true
		if sub.Part == 1 {
			r.QuizBt = sub.AwardedBT
		} else if sub.Part == 2 {
			if sub.QuestionType == "number" {
				n, _ := strconv.Atoi(digits.ReplaceAllString(sub.Answer, ""))
				r.PhysicalBt = n
			} else {
				r.McTotal++
				if sub.IsCorrect != nil && *sub.IsCorrect {
					r.Correct++
				}
			}
		}
	}

	// Pass 1: resistance map — who is immune, who reflects.
	immune := map[string]bool{}
	reflect := map[string]bool{}
	for _, pi := range pis {
		if pi.Item == nil {
			continue
		}
		if pi.Item.Immune {
			immune[pi.PlayerID.String()] = true
		}
		if pi.Item.Reflect {
			reflect[pi.PlayerID.String()] = true
		}
	}

	adj := func(pid string, delta int, note string) {
		if r := rows[pid]; r != nil {
			r.ItemBt += delta
			if note != "" {
				r.Notes = append(r.Notes, note)
			}
		}
	}

	// Pass 2: self-additive effects (order independent).
	for _, pi := range pis {
		it := pi.Item
		if it == nil {
			continue
		}
		owner := pi.PlayerID.String()
		if rows[owner] == nil {
			continue
		}
		if it.BTSelf != 0 {
			adj(owner, it.BTSelf, fmt.Sprintf("%s %+d BT", it.Name, it.BTSelf))
		}
		if it.QuizBTPct != 0 {
			bonus := int(math.Round(float64(it.QuizBTPct) / 100 * float64(rows[owner].QuizBt)))
			adj(owner, bonus, fmt.Sprintf("%s %+d BT (%d%% of %d quiz BT)", it.Name, bonus, it.QuizBTPct, rows[owner].QuizBt))
		}
		if it.DoubleGameBT {
			g := gameBt[owner]
			adj(owner, g, fmt.Sprintf("%s %+d BT (double game winnings)", it.Name, g))
		}
	}

	// Pass 3: targeted steals / deductions, with precedence.
	for _, pi := range pis {
		it := pi.Item
		if it == nil || pi.TargetPlayerID == nil {
			continue
		}
		if it.BTFromTarget == 0 && it.BTDeductTarget == 0 {
			continue
		}
		owner := pi.PlayerID.String()
		target := pi.TargetPlayerID.String()
		if rows[owner] == nil || rows[target] == nil || owner == target {
			continue
		}
		isSteal := it.BTFromTarget > 0
		amount := it.BTDeductTarget
		if isSteal {
			amount = it.BTFromTarget
		}
		tName := name(*pi.TargetPlayerID)
		oName := name(pi.PlayerID)

		switch {
		case !it.StripResistance && immune[target]:
			adj(owner, 0, fmt.Sprintf("%s vs %s — nullified by their immunity", it.Name, tName))
		case !it.StripResistance && reflect[target]:
			// Bounced back onto the attacker; any steal gain is cancelled.
			adj(owner, -amount, fmt.Sprintf("%s reflected by %s — you lose %d BT", it.Name, tName, amount))
		default:
			adj(target, -amount, fmt.Sprintf("%s: %s took %d BT from you", it.Name, oName, amount))
			if isSteal {
				adj(owner, amount, fmt.Sprintf("%s: stole %d BT from %s", it.Name, amount, tName))
			} else {
				adj(owner, 0, fmt.Sprintf("%s: removed %d BT from %s", it.Name, amount, tName))
			}
		}
	}

	out := make([]tallyRow, 0, len(rows))
	for pid, r := range rows {
		if !hasSub[pid] && r.ItemBt == 0 {
			continue // no quiz activity and untouched by items — omit for a clean table
		}
		r.FinalBt = r.QuizBt + r.PhysicalBt + r.ItemBt
		out = append(out, *r)
	}
	return out
}

// GET /gm/quiz/tally — per-character final Blood Token tally with item effects.
func (s *server) gmQuizTally(ctx *gin.Context) {
	rows, err := s.computeBtTally(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"players": rows})
}
