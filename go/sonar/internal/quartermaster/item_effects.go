package quartermaster

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

const (
	itemDuration    = 5 * time.Minute
	challengeAnswer = "This territory was claimed with cold, hard gemstones."
)

func (q *client) ApplyItemEffectByID(ctx context.Context, ownedInventoryItem models.OwnedInventoryItem, metadata *UseItemMetadata) error {
	switch ownedInventoryItem.InventoryItemID {
	case 1:
		teamMatch, err := q.db.Match().FindForTeamID(ctx, *ownedInventoryItem.TeamID)
		if err != nil {
			return err
		}
		// Make everything all squiggle for others when reading clues
		return q.AddEffectToMatch(ctx, teamMatch.MatchID, ownedInventoryItem, itemDuration)
	case 2:
		// Instantly reveal a hidden point on the map.
		return q.db.PointOfInterest().Unlock(ctx, metadata.PointOfInterestID, ownedInventoryItem.TeamID, ownedInventoryItem.UserID)
	case 3:
		// 	// Instantly capture a tier one challenge.
		return q.captureChallenge(ctx, metadata.PointOfInterestID, ownedInventoryItem, 1)
	case 4:
		// Instantly capture a tier two challenge.
		return q.captureChallenge(ctx, metadata.PointOfInterestID, ownedInventoryItem, 2)
	case 5:
		// Instantly capture a tier three challenge.
		return q.captureChallenge(ctx, metadata.PointOfInterestID, ownedInventoryItem, 3)
	case 6:
		if ownedInventoryItem.IsTeamItem() {
			// Steal all of another team's items. Must be within a 100 meter radius of the target team to use.
			return q.db.InventoryItem().StealItems(ctx, *ownedInventoryItem.TeamID, metadata.TargetTeamID)
		}

		return nil
	case 7:
		// Inflict a wound on another team.
		return q.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, &metadata.TargetTeamID, nil, 10, 1)
	case 8:
		// Hold in your inventory to increase your score by 1.
		return nil
	case 9:
		if ownedInventoryItem.IsTeamItem() {
			// Steal an item from another team.
			items, err := q.db.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{TeamID: &metadata.TargetTeamID})
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return nil
			}
			var randomItem models.OwnedInventoryItem
			for {
				index := rand.Intn(len(items))
				if index != 10 {
					randomItem = items[index]
					break
				}
			}
			return q.db.InventoryItem().StealItem(ctx, *ownedInventoryItem.TeamID, metadata.TargetTeamID, randomItem.InventoryItemID)
		}

		return nil
	case 10:
		// It's damage
		return nil
	case 11:
		// You have an acorn!
		return nil
	case 12:
		// Drink to remove one damage
		items, err := q.db.InventoryItem().GetItems(ctx, ownedInventoryItem)
		if err != nil {
			return err
		}

		for _, item := range items {
			if item.InventoryItemID == 10 {
				return q.db.InventoryItem().UseInventoryItem(ctx, item.ID)
			}
		}

		return nil
	case 13:
		// Remove all damage when held.
		return nil
	case 14:
		// Steal all of another team's items.
		return q.db.InventoryItem().StealItems(ctx, *ownedInventoryItem.TeamID, metadata.TargetTeamID)
	case 15:
		// Negate up to 3 damage when held.
		return nil
	default:
		return errors.New("no effect found for this item")
	}
}

func (q *client) AddEffectToMatch(ctx context.Context, matchID uuid.UUID, inventoryItem models.OwnedInventoryItem, duration time.Duration) error {
	if inventoryItem.TeamID == nil {
		return errors.New("cant use item without a team")
	}

	return q.db.InventoryItem().ApplyInventoryItem(ctx, matchID, inventoryItem.InventoryItemID, *inventoryItem.TeamID, duration)
}

func (q *client) captureChallenge(ctx context.Context, pointOfInterestID uuid.UUID, ownedInventoryItem models.OwnedInventoryItem, tier int) error {
	challenge, err := q.db.PointOfInterestChallenge().GetChallengeForPointOfInterest(ctx, pointOfInterestID, tier)
	if err != nil {
		return err
	}

	if _, err := q.db.PointOfInterestChallenge().SubmitAnswerForChallenge(ctx, challenge.ID, ownedInventoryItem.TeamID, ownedInventoryItem.UserID, challengeAnswer, "", true); err != nil {
		return err
	}

	if challenge.InventoryItemID == 0 {
		item, err := q.getRandomItem()
		if err != nil {
			return err
		}
		return q.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, ownedInventoryItem.TeamID, ownedInventoryItem.UserID, item.ID, 1)
	} else {
		return q.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, ownedInventoryItem.TeamID, ownedInventoryItem.UserID, challenge.InventoryItemID, 1)
	}
}
