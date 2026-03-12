package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func lockIntPtr(v int) *int {
	return &v
}

func TestInventoryItemLockUnlockStrengthPrefersNewField(t *testing.T) {
	item := &models.InventoryItem{
		UnlockTier:          lockIntPtr(25),
		UnlockLocksStrength: lockIntPtr(60),
	}
	if got := inventoryItemLockUnlockStrength(item); got != 60 {
		t.Fatalf("expected unlock locks strength 60, got %d", got)
	}
}

func TestSelectTreasureChestUnlockItemIDChoosesLowestSufficientStrength(t *testing.T) {
	weakID := uuid.New()
	strongID := uuid.New()
	selected := selectTreasureChestUnlockItemID(
		40,
		[]models.OwnedInventoryItem{
			{ID: strongID, InventoryItemID: 1, Quantity: 1},
			{ID: weakID, InventoryItemID: 2, Quantity: 1},
		},
		map[int]*models.InventoryItem{
			1: {ID: 1, UnlockLocksStrength: lockIntPtr(80)},
			2: {ID: 2, UnlockLocksStrength: lockIntPtr(45)},
		},
	)
	if selected == nil {
		t.Fatal("expected unlock item to be selected")
	}
	if *selected != weakID {
		t.Fatalf("expected lowest sufficient item %s, got %s", weakID, *selected)
	}
}

func TestHasTreasureChestUnlockSkillDetectsMatchingSpellEffect(t *testing.T) {
	if !hasTreasureChestUnlockSkill(35, []models.UserSpell{
		{
			Spell: models.Spell{
				Effects: models.SpellEffects{
					{Type: models.SpellEffectTypeUnlockLocks, Amount: 40},
				},
			},
		},
	}) {
		t.Fatal("expected unlock skill to satisfy chest requirement")
	}
}
