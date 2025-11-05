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
	ctx := context.Background()
	dbItems, err := c.db.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		return []InventoryItem{}
	}

	items := make([]InventoryItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = InventoryItem{
			ID:            dbItem.ID,
			Name:          dbItem.Name,
			ImageURL:      dbItem.ImageURL,
			FlavorText:    dbItem.FlavorText,
			EffectText:    dbItem.EffectText,
			RarityTier:    Rarity(dbItem.RarityTier),
			IsCaptureType: dbItem.IsCaptureType,
		}
	}
	return items
}

func (c *client) FindItemForItemID(itemID int) (InventoryItem, error) {
	ctx := context.Background()
	dbItem, err := c.db.InventoryItem().FindInventoryItemByID(ctx, itemID)
	if err != nil {
		return InventoryItem{}, fmt.Errorf("item not found: %w", err)
	}

	return InventoryItem{
		ID:            dbItem.ID,
		Name:          dbItem.Name,
		ImageURL:      dbItem.ImageURL,
		FlavorText:    dbItem.FlavorText,
		EffectText:    dbItem.EffectText,
		RarityTier:    Rarity(dbItem.RarityTier),
		IsCaptureType: dbItem.IsCaptureType,
	}, nil
}

func (c *client) GetItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID) (InventoryItem, error) {
	item, err := c.getRandomItem()
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, userID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) GetItemSpecificItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, itemID int) (InventoryItem, error) {
	item, err := c.FindItemForItemID(itemID)
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, userID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) UseItem(ctx context.Context, ownedInventoryItemID uuid.UUID, metadata *UseItemMetadata) error {
	ownedInventoryItem, err := c.db.InventoryItem().FindByID(ctx, ownedInventoryItemID)
	if err != nil {
		return err
	}

	if err := c.db.InventoryItem().UseInventoryItem(ctx, ownedInventoryItem.ID); err != nil {
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
