package quartermaster

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/gofrs/uuid"
	"golang.org/x/exp/rand"
)

type client struct {
	db *db.DbClient
}

type Quartermaster interface {
	UseItem(ctx context.Context, itemID int) error
	GetItem(ctx context.Context, teamID uuid.UUID) (InventoryItem, error)
}

func NewClient(db *db.DbClient) Quartermaster {
	return &client{db: db}
}

func (c *client) GetItem(ctx context.Context, teamID uuid.UUID) (InventoryItem, error) {
	item, err := c.getRandomItem()
	if err != nil {
		return InventoryItem{}, err
	}
	return item, nil
}

func (c *client) UseItem(ctx context.Context, itemID int) error {
	return nil
}

func (c *client) getRandomItem() (InventoryItem, error) {
	rand.Seed(uint64(time.Now().UnixNano()))

	const (
		weightCommon   = 50
		weightUncommon = 30
		weightEpic     = 15
		weightMythic   = 5
	)

	rarityWeights := map[Rarity]int{
		RarityCommon:   weightCommon,
		RarityUncommon: weightUncommon,
		RarityEpic:     weightEpic,
		RarityMythic:   weightMythic,
	}

	totalWeight := 0
	for _, item := range preDefinedItems {
		totalWeight += rarityWeights[item.RarityTier]
	}

	randWeight := rand.Intn(totalWeight)

	for _, item := range preDefinedItems {
		randWeight -= rarityWeights[item.RarityTier]
		if randWeight <= 0 {
			return item, nil
		}
	}

	return InventoryItem{}, errors.New("failed to select a random item")
}
