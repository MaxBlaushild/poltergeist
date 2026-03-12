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

func TestSelectTreasureChestUnlockSpellIDChoosesLowestSufficientStrength(t *testing.T) {
	weakID := uuid.New()
	strongID := uuid.New()
	selected := selectTreasureChestUnlockSpellID(
		35,
		[]models.UserSpell{
			{
				SpellID: strongID,
				Spell: models.Spell{
					Effects: models.SpellEffects{
						{Type: models.SpellEffectTypeUnlockLocks, Amount: 80},
					},
				},
			},
			{
				SpellID: weakID,
				Spell: models.Spell{
					Effects: models.SpellEffects{
						{Type: models.SpellEffectTypeUnlockLocks, Amount: 40},
					},
				},
			},
		},
	)
	if selected == nil {
		t.Fatal("expected unlock spell to be selected")
	}
	if *selected != weakID {
		t.Fatalf("expected lowest sufficient spell %s, got %s", weakID, *selected)
	}
}

func TestResolveTreasureChestUnlockPrefersSkillOverItem(t *testing.T) {
	itemID := uuid.New()
	spellID := uuid.New()
	selectedItemID, selectedSpellID, err := resolveTreasureChestUnlock(
		35,
		[]models.OwnedInventoryItem{
			{ID: itemID, InventoryItemID: 1, Quantity: 1},
		},
		map[int]*models.InventoryItem{
			1: {ID: 1, UnlockLocksStrength: lockIntPtr(40)},
		},
		[]models.UserSpell{
			{
				SpellID: spellID,
				Spell: models.Spell{
					Effects: models.SpellEffects{
						{Type: models.SpellEffectTypeUnlockLocks, Amount: 40},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("expected chest unlock to be allowed, got %v", err)
	}
	if selectedItemID != nil {
		t.Fatalf("expected no item to be selected when a skill is available, got %s", *selectedItemID)
	}
	if selectedSpellID == nil {
		t.Fatal("expected skill to be preferred when available")
	}
	if *selectedSpellID != spellID {
		t.Fatalf("expected spell %s to be selected, got %s", spellID, *selectedSpellID)
	}
}

func TestResolveTreasureChestUnlockFallsBackToItem(t *testing.T) {
	itemID := uuid.New()
	selectedItemID, selectedSpellID, err := resolveTreasureChestUnlock(
		35,
		[]models.OwnedInventoryItem{
			{ID: itemID, InventoryItemID: 1, Quantity: 1},
		},
		map[int]*models.InventoryItem{
			1: {ID: 1, UnlockLocksStrength: lockIntPtr(40)},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("expected matching item to allow unlock, got %v", err)
	}
	if selectedSpellID != nil {
		t.Fatalf("expected no skill to be selected when only an item is available, got %s", *selectedSpellID)
	}
	if selectedItemID == nil {
		t.Fatal("expected matching item to be selected")
	}
	if *selectedItemID != itemID {
		t.Fatalf("expected item %s to be selected, got %s", itemID, *selectedItemID)
	}
}

func TestResolveTreasureChestUnlockFromSelectionUsesExplicitItem(t *testing.T) {
	itemID := uuid.New()
	selectedItemID, selectedSpellID, err := resolveTreasureChestUnlockFromSelection(
		35,
		treasureChestUnlockSelection{
			Method:               treasureChestUnlockMethodItem,
			OwnedInventoryItemID: &itemID,
		},
		[]models.OwnedInventoryItem{
			{ID: itemID, InventoryItemID: 1, Quantity: 1},
		},
		map[int]*models.InventoryItem{
			1: {ID: 1, UnlockLocksStrength: lockIntPtr(40)},
		},
		[]models.UserSpell{
			{
				SpellID: uuid.New(),
				Spell: models.Spell{
					Effects: models.SpellEffects{
						{Type: models.SpellEffectTypeUnlockLocks, Amount: 99},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("expected explicit item selection to work, got %v", err)
	}
	if selectedSpellID != nil {
		t.Fatalf("expected no spell to be selected, got %s", *selectedSpellID)
	}
	if selectedItemID == nil || *selectedItemID != itemID {
		t.Fatalf("expected item %s to be selected, got %v", itemID, selectedItemID)
	}
}

func TestResolveTreasureChestUnlockFromSelectionUsesExplicitSpell(t *testing.T) {
	spellID := uuid.New()
	selectedItemID, selectedSpellID, err := resolveTreasureChestUnlockFromSelection(
		35,
		treasureChestUnlockSelection{
			Method:  treasureChestUnlockMethodSpell,
			SpellID: &spellID,
		},
		[]models.OwnedInventoryItem{
			{ID: uuid.New(), InventoryItemID: 1, Quantity: 1},
		},
		map[int]*models.InventoryItem{
			1: {ID: 1, UnlockLocksStrength: lockIntPtr(40)},
		},
		[]models.UserSpell{
			{
				SpellID: spellID,
				Spell: models.Spell{
					Effects: models.SpellEffects{
						{Type: models.SpellEffectTypeUnlockLocks, Amount: 40},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("expected explicit spell selection to work, got %v", err)
	}
	if selectedItemID != nil {
		t.Fatalf("expected no item to be selected, got %s", *selectedItemID)
	}
	if selectedSpellID == nil || *selectedSpellID != spellID {
		t.Fatalf("expected spell %s to be selected, got %v", spellID, selectedSpellID)
	}
}
