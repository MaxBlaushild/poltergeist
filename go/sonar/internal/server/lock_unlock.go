package server

import (
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

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
