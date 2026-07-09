package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

// ---- helpers ----

func bp(b bool) *bool { return &b }

func mkPlayer(name, house string) models.VampirePlayer {
	return models.VampirePlayer{
		ID: uuid.New(),
		Character: &models.VampireCharacter{
			Name:  name,
			House: &models.VampireHouse{Name: house},
		},
	}
}

func p1Sub(pid uuid.UUID, awardedBT int) db.QuizSubmissionDetail {
	return db.QuizSubmissionDetail{PlayerID: pid, Part: 1, QuestionType: "open", AwardedBT: awardedBT}
}

func onHandSub(pid uuid.UUID, answer string) db.QuizSubmissionDetail {
	return db.QuizSubmissionDetail{PlayerID: pid, Part: 2, QuestionType: "number", Answer: answer}
}

func mcSub(pid uuid.UUID, correct bool) db.QuizSubmissionDetail {
	return db.QuizSubmissionDetail{PlayerID: pid, Part: 2, QuestionType: "multiple_choice", IsCorrect: bp(correct)}
}

func held(owner uuid.UUID, target *uuid.UUID, it models.VampireItem) models.VampirePlayerItem {
	return models.VampirePlayerItem{PlayerID: owner, TargetPlayerID: target, Item: &it}
}

func rowFor(t *testing.T, rows []tallyRow, character string) tallyRow {
	t.Helper()
	for _, r := range rows {
		if r.Character == character {
			return r
		}
	}
	t.Fatalf("no tally row for %q (rows: %+v)", character, rows)
	return tallyRow{}
}

func absent(rows []tallyRow, character string) bool {
	for _, r := range rows {
		if r.Character == character {
			return false
		}
	}
	return true
}

// ---- base tally ----

func TestResolveTally_BaseNoItems(t *testing.T) {
	alice := mkPlayer("Alice", "Ashvale")
	subs := []db.QuizSubmissionDetail{
		p1Sub(alice.ID, 6),
		onHandSub(alice.ID, "10"),
		mcSub(alice.ID, true),
		mcSub(alice.ID, false),
	}
	rows := resolveTally([]models.VampirePlayer{alice}, subs, nil, nil)

	r := rowFor(t, rows, "Alice")
	if r.QuizBt != 6 || r.PhysicalBt != 10 {
		t.Fatalf("base: quizBt=%d physicalBt=%d, want 6/10", r.QuizBt, r.PhysicalBt)
	}
	if r.Correct != 1 || r.McTotal != 2 {
		t.Fatalf("mc: correct=%d mcTotal=%d, want 1/2", r.Correct, r.McTotal)
	}
	if r.ItemBt != 0 || r.FinalBt != 16 {
		t.Fatalf("final: itemBt=%d finalBt=%d, want 0/16", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_OnHandNonNumericStripped(t *testing.T) {
	p := mkPlayer("Nox", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(p.ID, "about 12 tokens")}
	r := rowFor(t, resolveTally([]models.VampirePlayer{p}, subs, nil, nil), "Nox")
	if r.PhysicalBt != 12 {
		t.Fatalf("physicalBt=%d, want 12 (digits extracted)", r.PhysicalBt)
	}
}

// ---- self-additive item effects ----

func TestResolveTally_FlatSelfBT(t *testing.T) {
	bob := mkPlayer("Bob", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(bob.ID, "5")}
	scepter := models.VampireItem{Name: "Gilded Moonstone Scepter", BTSelf: 15}
	rows := resolveTally([]models.VampirePlayer{bob}, subs, nil,
		[]models.VampirePlayerItem{held(bob.ID, nil, scepter)})

	r := rowFor(t, rows, "Bob")
	if r.ItemBt != 15 || r.FinalBt != 20 {
		t.Fatalf("itemBt=%d finalBt=%d, want 15/20", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_QuizPercentBonusRounds(t *testing.T) {
	// 50% of 7 quiz BT = 3.5 -> rounds to 4.
	carol := mkPlayer("Carol", "Ashvale")
	subs := []db.QuizSubmissionDetail{p1Sub(carol.ID, 7)}
	puzzle := models.VampireItem{Name: "Millenium Puzzle", QuizBTPct: 50}
	r := rowFor(t, resolveTally([]models.VampirePlayer{carol}, subs, nil,
		[]models.VampirePlayerItem{held(carol.ID, nil, puzzle)}), "Carol")
	if r.ItemBt != 4 {
		t.Fatalf("itemBt=%d, want 4 (round(3.5))", r.ItemBt)
	}
	if r.FinalBt != 11 { // 7 quiz + 0 hand + 4 bonus
		t.Fatalf("finalBt=%d, want 11", r.FinalBt)
	}
}

func TestResolveTally_DoubleGameBT(t *testing.T) {
	dave := mkPlayer("Dave", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(dave.ID, "0")}
	buster := models.VampireItem{Name: "Buster Sword", DoubleGameBT: true}
	gameBt := map[string]int{dave.ID.String(): 8}
	r := rowFor(t, resolveTally([]models.VampirePlayer{dave}, subs, gameBt,
		[]models.VampirePlayerItem{held(dave.ID, nil, buster)}), "Dave")
	if r.ItemBt != 8 || r.FinalBt != 8 {
		t.Fatalf("itemBt=%d finalBt=%d, want 8/8", r.ItemBt, r.FinalBt)
	}
}

// ---- targeted steals / deductions ----

func TestResolveTally_Steal(t *testing.T) {
	att := mkPlayer("Attacker", "Ashvale")
	vic := mkPlayer("Victim", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(att.ID, "0"), onHandSub(vic.ID, "20")}
	craven := models.VampireItem{Name: "Craven Edge", BTFromTarget: 8}
	rows := resolveTally([]models.VampirePlayer{att, vic}, subs, nil,
		[]models.VampirePlayerItem{held(att.ID, &vic.ID, craven)})

	if r := rowFor(t, rows, "Attacker"); r.ItemBt != 8 || r.FinalBt != 8 {
		t.Fatalf("attacker itemBt=%d finalBt=%d, want 8/8", r.ItemBt, r.FinalBt)
	}
	if r := rowFor(t, rows, "Victim"); r.ItemBt != -8 || r.FinalBt != 12 {
		t.Fatalf("victim itemBt=%d finalBt=%d, want -8/12", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_DeductNoGain(t *testing.T) {
	att := mkPlayer("Attacker", "Ashvale")
	vic := mkPlayer("Victim", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(att.ID, "0"), onHandSub(vic.ID, "20")}
	dagger := models.VampireItem{Name: "Dagger", BTDeductTarget: 5}
	rows := resolveTally([]models.VampirePlayer{att, vic}, subs, nil,
		[]models.VampirePlayerItem{held(att.ID, &vic.ID, dagger)})

	if r := rowFor(t, rows, "Attacker"); r.ItemBt != 0 {
		t.Fatalf("attacker itemBt=%d, want 0 (deduct gives no gain)", r.ItemBt)
	}
	if r := rowFor(t, rows, "Victim"); r.ItemBt != -5 || r.FinalBt != 15 {
		t.Fatalf("victim itemBt=%d finalBt=%d, want -5/15", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_ImmuneCancels(t *testing.T) {
	att := mkPlayer("Attacker", "Ashvale")
	vic := mkPlayer("Victim", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(att.ID, "0"), onHandSub(vic.ID, "20")}
	craven := models.VampireItem{Name: "Craven Edge", BTFromTarget: 8}
	tiara := models.VampireItem{Name: "Tiara", Immune: true}
	rows := resolveTally([]models.VampirePlayer{att, vic}, subs, nil, []models.VampirePlayerItem{
		held(att.ID, &vic.ID, craven),
		held(vic.ID, nil, tiara),
	})

	if r := rowFor(t, rows, "Attacker"); r.ItemBt != 0 {
		t.Fatalf("attacker itemBt=%d, want 0 (steal nullified)", r.ItemBt)
	}
	if r := rowFor(t, rows, "Victim"); r.ItemBt != 0 || r.FinalBt != 20 {
		t.Fatalf("victim itemBt=%d finalBt=%d, want 0/20 (immune)", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_ReflectBouncesToAttacker(t *testing.T) {
	att := mkPlayer("Attacker", "Ashvale")
	vic := mkPlayer("Victim", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(att.ID, "10"), onHandSub(vic.ID, "20")}
	craven := models.VampireItem{Name: "Craven Edge", BTFromTarget: 8}
	ouro := models.VampireItem{Name: "Ouroboros", Reflect: true}
	rows := resolveTally([]models.VampirePlayer{att, vic}, subs, nil, []models.VampirePlayerItem{
		held(att.ID, &vic.ID, craven),
		held(vic.ID, nil, ouro),
	})

	if r := rowFor(t, rows, "Attacker"); r.ItemBt != -8 || r.FinalBt != 2 {
		t.Fatalf("attacker itemBt=%d finalBt=%d, want -8/2 (reflected)", r.ItemBt, r.FinalBt)
	}
	if r := rowFor(t, rows, "Victim"); r.ItemBt != 0 || r.FinalBt != 20 {
		t.Fatalf("victim itemBt=%d finalBt=%d, want 0/20 (unharmed)", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_StripResistanceOverridesImmune(t *testing.T) {
	att := mkPlayer("Attacker", "Ashvale")
	vic := mkPlayer("Victim", "Ashvale")
	subs := []db.QuizSubmissionDetail{onHandSub(att.ID, "0"), onHandSub(vic.ID, "20")}
	tiara := models.VampireItem{Name: "Tiara", Immune: true}
	claws := models.VampireItem{Name: "Claws of Rending", BTDeductTarget: 1, StripResistance: true}
	rows := resolveTally([]models.VampirePlayer{att, vic}, subs, nil, []models.VampirePlayerItem{
		held(vic.ID, nil, tiara),
		held(att.ID, &vic.ID, claws),
	})

	if r := rowFor(t, rows, "Victim"); r.ItemBt != -1 || r.FinalBt != 19 {
		t.Fatalf("victim itemBt=%d finalBt=%d, want -1/19 (strip beats immune)", r.ItemBt, r.FinalBt)
	}
}

// ---- table membership ----

func TestResolveTally_OmitsInactiveIncludesTargeted(t *testing.T) {
	att := mkPlayer("Attacker", "Ashvale")
	silent := mkPlayer("Silent", "Ashvale") // no submission, but gets stolen from
	ghost := mkPlayer("Ghost", "Ashvale")    // no submission, no items -> omitted
	subs := []db.QuizSubmissionDetail{onHandSub(att.ID, "0")}
	craven := models.VampireItem{Name: "Craven Edge", BTFromTarget: 8}
	rows := resolveTally([]models.VampirePlayer{att, silent, ghost}, subs, nil,
		[]models.VampirePlayerItem{held(att.ID, &silent.ID, craven)})

	if !absent(rows, "Ghost") {
		t.Fatalf("Ghost should be omitted (no sub, no item effect)")
	}
	if r := rowFor(t, rows, "Silent"); r.ItemBt != -8 || r.FinalBt != -8 {
		t.Fatalf("Silent itemBt=%d finalBt=%d, want -8/-8", r.ItemBt, r.FinalBt)
	}
}

func TestResolveTally_StackedSelfEffects(t *testing.T) {
	p := mkPlayer("Stacker", "Ashvale")
	subs := []db.QuizSubmissionDetail{p1Sub(p.ID, 10), onHandSub(p.ID, "3")}
	scepter := models.VampireItem{Name: "Scepter", BTSelf: 15}
	puzzle := models.VampireItem{Name: "Puzzle", QuizBTPct: 100} // +10
	rows := resolveTally([]models.VampirePlayer{p}, subs, nil, []models.VampirePlayerItem{
		held(p.ID, nil, scepter),
		held(p.ID, nil, puzzle),
	})
	// base 10 quiz + 3 hand = 13; items +15 +10 = +25; final 38.
	r := rowFor(t, rows, "Stacker")
	if r.ItemBt != 25 || r.FinalBt != 38 {
		t.Fatalf("itemBt=%d finalBt=%d, want 25/38", r.ItemBt, r.FinalBt)
	}
}
