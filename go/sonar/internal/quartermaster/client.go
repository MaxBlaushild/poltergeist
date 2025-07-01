package quartermaster

import (
	"context"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

type client struct {
	db db.DbClient
}

type Quartermaster interface {
	UseItem(ctx context.Context, ownedInventoryItemID uuid.UUID, metadata *UseItemMetadata) error
	GetItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID) (InventoryItem, error)
	FindItemForItemID(itemID int) (InventoryItem, error)
	GetInventoryItems() []InventoryItem
	ApplyInventoryItemEffects(ctx context.Context, userID uuid.UUID, match *models.Match) error
	GetItemSpecificItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, itemID int) (InventoryItem, error)
	EquipItem(ctx context.Context, userID uuid.UUID, ownedInventoryItemID uuid.UUID) error
	UnequipItem(ctx context.Context, userID uuid.UUID, equipmentSlot string) error
	GetUserEquipment(ctx context.Context, userID uuid.UUID) ([]models.UserEquipment, error)
}

type UseItemMetadata struct {
	PointOfInterestID uuid.UUID `json:"pointOfInterestId"`
	TargetTeamID      uuid.UUID `json:"targetTeamId"`
	ChallengeID       uuid.UUID `json:"challengeId"`
}

func NewClient(db db.DbClient) Quartermaster {
	return &client{db: db}
}

func (c *client) GetInventoryItems() []InventoryItem {
	// Get items from database
	dbItems, err := c.db.InventoryItem().FindAll(context.Background())
	if err != nil {
		// Fallback to hardcoded items if database fails
		return PreDefinedItems
	}
	
	// Convert database models to quartermaster models
	items := make([]InventoryItem, len(dbItems))
	for i, dbItem := range dbItems {
		equipmentSlot := EquipmentSlot("")
		if dbItem.EquipmentSlot != nil {
			equipmentSlot = EquipmentSlot(*dbItem.EquipmentSlot)
		}
		
		items[i] = InventoryItem{
			ID:            dbItem.ID,
			Name:          dbItem.Name,
			ImageURL:      dbItem.ImageURL,
			FlavorText:    dbItem.FlavorText,
			EffectText:    dbItem.EffectText,
			RarityTier:    Rarity(dbItem.RarityTier),
			IsCaptureType: dbItem.IsCaptureType,
			ItemType:      ItemType(dbItem.ItemType),
			EquipmentSlot: equipmentSlot,
		}
	}
	return items
}

func (c *client) FindItemForItemID(itemID int) (InventoryItem, error) {
	// Try to get item from database first
	dbItem, err := c.db.InventoryItem().FindByID(context.Background(), itemID)
	if err == nil {
		equipmentSlot := EquipmentSlot("")
		if dbItem.EquipmentSlot != nil {
			equipmentSlot = EquipmentSlot(*dbItem.EquipmentSlot)
		}
		
		return InventoryItem{
			ID:            dbItem.ID,
			Name:          dbItem.Name,
			ImageURL:      dbItem.ImageURL,
			FlavorText:    dbItem.FlavorText,
			EffectText:    dbItem.EffectText,
			RarityTier:    Rarity(dbItem.RarityTier),
			IsCaptureType: dbItem.IsCaptureType,
			ItemType:      ItemType(dbItem.ItemType),
			EquipmentSlot: equipmentSlot,
		}, nil
	}
	
	// Fallback to hardcoded items if database fails
	for _, item := range PreDefinedItems {
		if item.ID == itemID {
			return item, nil
		}
	}

	return InventoryItem{}, fmt.Errorf("item not found")
}

func (c *client) GetItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID) (InventoryItem, error) {
	item, err := c.getRandomItem()
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.OwnedInventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, userID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) GetItemSpecificItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, itemID int) (InventoryItem, error) {
	item, err := c.FindItemForItemID(itemID)
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.OwnedInventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, userID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) UseItem(ctx context.Context, ownedInventoryItemID uuid.UUID, metadata *UseItemMetadata) error {
	ownedInventoryItem, err := c.db.OwnedInventoryItem().FindByID(ctx, ownedInventoryItemID)
	if err != nil {
		return err
	}

	if err := c.db.OwnedInventoryItem().UseInventoryItem(ctx, ownedInventoryItem.ID); err != nil {
		return err
	}

	if err := c.ApplyItemEffectByID(ctx, *ownedInventoryItem, metadata); err != nil {
		return err
	}

	return nil
}

func (c *client) getRandomItem() (InventoryItem, error) {
	rand.Seed(uint64(time.Now().UnixNano()))

	const (
		weightCommon       = 50
		weightUncommon     = 30
		weightEpic         = 15
		weightMythic       = 5
		weightNotDroppable = 0
	)

	rarityWeights := map[Rarity]int{
		RarityCommon:   weightCommon,
		RarityUncommon: weightUncommon,
		RarityEpic:     weightEpic,
		RarityMythic:   weightMythic,
		NotDroppable:   weightNotDroppable,
	}

	// Get items from database or fallback to hardcoded
	items := c.GetInventoryItems()

	totalWeight := 0
	for _, item := range items {
		totalWeight += rarityWeights[item.RarityTier]
	}

	for {
		randWeight := rand.Intn(totalWeight + 1)

		for _, item := range items {
			randWeight -= rarityWeights[item.RarityTier]
			if randWeight < 0 {
				return item, nil
			}
		}
	}
}

func (c *client) EquipItem(ctx context.Context, userID uuid.UUID, ownedInventoryItemID uuid.UUID) error {
	// First, get the owned inventory item to determine what it is
	ownedItem, err := c.db.OwnedInventoryItem().FindByID(ctx, ownedInventoryItemID)
	if err != nil {
		return fmt.Errorf("failed to find owned inventory item: %w", err)
	}

	// Verify the item belongs to the user
	if ownedItem.UserID == nil || *ownedItem.UserID != userID {
		return fmt.Errorf("item does not belong to user")
	}

	// Get the item definition to check if it's equippable
	item, err := c.FindItemForItemID(ownedItem.InventoryItemID)
	if err != nil {
		return fmt.Errorf("failed to find item definition: %w", err)
	}

	// Check if the item is equippable
	if item.ItemType != ItemTypeEquippable {
		return fmt.Errorf("item is not equippable")
	}

	// Equip the item
	_, err = c.db.UserEquipment().EquipItem(ctx, userID, ownedInventoryItemID, string(item.EquipmentSlot))
	if err != nil {
		return fmt.Errorf("failed to equip item: %w", err)
	}

	return nil
}

func (c *client) UnequipItem(ctx context.Context, userID uuid.UUID, equipmentSlot string) error {
	err := c.db.UserEquipment().UnequipItem(ctx, userID, equipmentSlot)
	if err != nil {
		return fmt.Errorf("failed to unequip item: %w", err)
	}
	return nil
}

func (c *client) GetUserEquipment(ctx context.Context, userID uuid.UUID) ([]models.UserEquipment, error) {
	equipment, err := c.db.UserEquipment().GetUserEquipment(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user equipment: %w", err)
	}
	return equipment, nil
}
