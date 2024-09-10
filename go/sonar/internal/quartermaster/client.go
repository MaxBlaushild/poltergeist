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
	UseItem(ctx context.Context, teamInventoryItemID uuid.UUID, metadata *UseItemMetadata) error
	GetItem(ctx context.Context, teamID uuid.UUID) (InventoryItem, error)
	FindItemForItemID(itemID int) (InventoryItem, error)
	GetInventoryItems() []InventoryItem
	ApplyInventoryItemEffects(ctx context.Context, userID uuid.UUID, match *models.Match) error
	GetItemSpecificItem(ctx context.Context, teamID uuid.UUID, itemID int) (InventoryItem, error)
}

type UseItemMetadata struct {
	PointOfInterestID uuid.UUID `json:"pointOfInterestId"`
	TargetTeamID      uuid.UUID `json:"targetTeamId"`
}

func NewClient(db db.DbClient) Quartermaster {
	return &client{db: db}
}

func (c *client) GetInventoryItems() []InventoryItem {
	return PreDefinedItems
}

func (c *client) FindItemForItemID(itemID int) (InventoryItem, error) {
	for _, item := range PreDefinedItems {
		if item.ID == itemID {
			return item, nil
		}
	}

	return InventoryItem{}, fmt.Errorf("item not found")
}

func (c *client) GetItem(ctx context.Context, teamID uuid.UUID) (InventoryItem, error) {
	item, err := c.getRandomItem()
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) GetItemSpecificItem(ctx context.Context, teamID uuid.UUID, itemID int) (InventoryItem, error) {
	item, err := c.FindItemForItemID(itemID)
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) UseItem(ctx context.Context, teamInventoryItemID uuid.UUID, metadata *UseItemMetadata) error {
	teamInventoryItem, err := c.db.InventoryItem().FindByID(ctx, teamInventoryItemID)
	if err != nil {
		return err
	}

	if err := c.db.InventoryItem().UseInventoryItem(ctx, teamInventoryItem.ID); err != nil {
		return err
	}

	teamMatch, err := c.db.Match().FindForTeamID(ctx, teamInventoryItem.TeamID)
	if err != nil {
		return err
	}

	if err := c.ApplyItemEffectByID(ctx, teamInventoryItem, teamMatch, metadata); err != nil {
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

	totalWeight := 0
	for _, item := range PreDefinedItems {
		totalWeight += rarityWeights[item.RarityTier]
	}

	for {
		randWeight := rand.Intn(totalWeight + 1)

		for _, item := range PreDefinedItems {
			randWeight -= rarityWeights[item.RarityTier]
			if randWeight < 0 {
				return item, nil
			}
		}
	}
}
