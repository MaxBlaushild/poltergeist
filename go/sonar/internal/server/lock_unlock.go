package server

import (
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const (
	treasureChestUnlockMethodItem  = "item"
	treasureChestUnlockMethodSpell = "spell"
)

type treasureChestUnlockSelection struct {
	Method               string
	OwnedInventoryItemID *uuid.UUID
	SpellID              *uuid.UUID
}

func inventoryItemLockUnlockStrength(item *models.InventoryItem) int {
	if item == nil {
		return 0
	}
	if item.UnlockLocksStrength != nil {
		return *item.UnlockLocksStrength
	}
	if item.UnlockTier != nil {
		return *item.UnlockTier
	}
	return 0
}

func spellLockUnlockStrength(spell *models.Spell) int {
	if spell == nil {
		return 0
	}
	best := 0
	for _, effect := range spell.Effects {
		if effect.Type != models.SpellEffectTypeUnlockLocks {
			continue
		}
		if effect.Amount > best {
			best = effect.Amount
		}
	}
	return best
}

func selectTreasureChestUnlockItemID(
	requiredStrength int,
	ownedItems []models.OwnedInventoryItem,
	inventoryByID map[int]*models.InventoryItem,
) *uuid.UUID {
	if requiredStrength <= 0 {
		return nil
	}
	var selected *models.OwnedInventoryItem
	selectedStrength := 0
	for i := range ownedItems {
		ownedItem := &ownedItems[i]
		if ownedItem.Quantity <= 0 {
			continue
		}
		strength := inventoryItemLockUnlockStrength(inventoryByID[ownedItem.InventoryItemID])
		if strength < requiredStrength {
			continue
		}
		if selected == nil || strength < selectedStrength {
			selected = ownedItem
			selectedStrength = strength
		}
	}
	if selected == nil {
		return nil
	}
	return &selected.ID
}

func selectTreasureChestUnlockSpellID(requiredStrength int, userSpells []models.UserSpell) *uuid.UUID {
	if requiredStrength <= 0 {
		return nil
	}
	var selected *models.UserSpell
	selectedStrength := 0
	for i := range userSpells {
		userSpell := &userSpells[i]
		strength := spellLockUnlockStrength(&userSpell.Spell)
		if strength < requiredStrength {
			continue
		}
		if selected == nil || strength < selectedStrength {
			selected = userSpell
			selectedStrength = strength
		}
	}
	if selected == nil {
		return nil
	}
	return &selected.SpellID
}

func normalizeTreasureChestUnlockMethod(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case treasureChestUnlockMethodItem:
		return treasureChestUnlockMethodItem
	case treasureChestUnlockMethodSpell:
		return treasureChestUnlockMethodSpell
	default:
		return ""
	}
}

func unlockMethodForResponse(unlockItemID *uuid.UUID, unlockSpellID *uuid.UUID) string {
	if unlockSpellID != nil {
		return treasureChestUnlockMethodSpell
	}
	if unlockItemID != nil {
		return treasureChestUnlockMethodItem
	}
	return ""
}

func buildTreasureChestUnlockSelection(
	method string,
	ownedInventoryItemID string,
	spellID string,
) (treasureChestUnlockSelection, error) {
	selection := treasureChestUnlockSelection{
		Method: normalizeTreasureChestUnlockMethod(method),
	}
	ownedInventoryItemID = strings.TrimSpace(ownedInventoryItemID)
	if ownedInventoryItemID != "" {
		parsedID, err := uuid.Parse(ownedInventoryItemID)
		if err != nil {
			return treasureChestUnlockSelection{}, fmt.Errorf("invalid ownedInventoryItemId")
		}
		selection.OwnedInventoryItemID = &parsedID
		if selection.Method == "" {
			selection.Method = treasureChestUnlockMethodItem
		}
	}
	spellID = strings.TrimSpace(spellID)
	if spellID != "" {
		parsedID, err := uuid.Parse(spellID)
		if err != nil {
			return treasureChestUnlockSelection{}, fmt.Errorf("invalid spellId")
		}
		selection.SpellID = &parsedID
		if selection.Method == "" {
			selection.Method = treasureChestUnlockMethodSpell
		}
	}
	if selection.Method == "" {
		return selection, nil
	}
	switch selection.Method {
	case treasureChestUnlockMethodItem:
		if selection.OwnedInventoryItemID == nil {
			return treasureChestUnlockSelection{}, fmt.Errorf("ownedInventoryItemId is required when unlockMethod is item")
		}
		if selection.SpellID != nil {
			return treasureChestUnlockSelection{}, fmt.Errorf("spellId cannot be provided when unlockMethod is item")
		}
	case treasureChestUnlockMethodSpell:
		if selection.SpellID == nil {
			return treasureChestUnlockSelection{}, fmt.Errorf("spellId is required when unlockMethod is spell")
		}
		if selection.OwnedInventoryItemID != nil {
			return treasureChestUnlockSelection{}, fmt.Errorf("ownedInventoryItemId cannot be provided when unlockMethod is spell")
		}
	default:
		return treasureChestUnlockSelection{}, fmt.Errorf("invalid unlockMethod")
	}
	return selection, nil
}

func resolveTreasureChestUnlock(
	requiredStrength int,
	ownedItems []models.OwnedInventoryItem,
	inventoryByID map[int]*models.InventoryItem,
	userSpells []models.UserSpell,
) (*uuid.UUID, *uuid.UUID, error) {
	if unlockSpellID := selectTreasureChestUnlockSpellID(requiredStrength, userSpells); unlockSpellID != nil {
		return nil, unlockSpellID, nil
	}
	if unlockItemID := selectTreasureChestUnlockItemID(requiredStrength, ownedItems, inventoryByID); unlockItemID != nil {
		return unlockItemID, nil, nil
	}
	return nil, nil, fmt.Errorf("you do not have an item or skill with sufficient lock strength to open this chest")
}

func resolveTreasureChestUnlockFromSelection(
	requiredStrength int,
	selection treasureChestUnlockSelection,
	ownedItems []models.OwnedInventoryItem,
	inventoryByID map[int]*models.InventoryItem,
	userSpells []models.UserSpell,
) (*uuid.UUID, *uuid.UUID, error) {
	if selection.Method == "" {
		return resolveTreasureChestUnlock(requiredStrength, ownedItems, inventoryByID, userSpells)
	}
	switch selection.Method {
	case treasureChestUnlockMethodItem:
		for i := range ownedItems {
			ownedItem := &ownedItems[i]
			if selection.OwnedInventoryItemID == nil || ownedItem.ID != *selection.OwnedInventoryItemID {
				continue
			}
			if ownedItem.Quantity <= 0 {
				return nil, nil, fmt.Errorf("selected unlock item is no longer available")
			}
			if inventoryItemLockUnlockStrength(inventoryByID[ownedItem.InventoryItemID]) < requiredStrength {
				return nil, nil, fmt.Errorf("selected unlock item is not strong enough to open this chest")
			}
			return &ownedItem.ID, nil, nil
		}
		return nil, nil, fmt.Errorf("selected unlock item is not available")
	case treasureChestUnlockMethodSpell:
		for i := range userSpells {
			userSpell := &userSpells[i]
			if selection.SpellID == nil || userSpell.SpellID != *selection.SpellID {
				continue
			}
			if spellLockUnlockStrength(&userSpell.Spell) < requiredStrength {
				return nil, nil, fmt.Errorf("selected skill is not strong enough to open this chest")
			}
			return nil, &userSpell.SpellID, nil
		}
		return nil, nil, fmt.Errorf("selected skill is not available")
	default:
		return nil, nil, fmt.Errorf("invalid unlockMethod")
	}
}

func hasTreasureChestUnlockSkill(requiredStrength int, userSpells []models.UserSpell) bool {
	if requiredStrength <= 0 {
		return true
	}
	for _, userSpell := range userSpells {
		if spellLockUnlockStrength(&userSpell.Spell) >= requiredStrength {
			return true
		}
	}
	return false
}
