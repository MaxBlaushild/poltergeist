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

func (q *client) ApplyItemEffectByID(ctx context.Context, teamInventoryItem *models.TeamInventoryItem, teamMatch *models.TeamMatch, metadata *UseItemMetadata) error {
	switch teamInventoryItem.InventoryItemID {
	case 1:
		// Make everything all squiggle for others when reading clues
		return q.AddEffectToMatch(ctx, teamMatch.MatchID, teamInventoryItem.InventoryItemID, teamInventoryItem.TeamID, itemDuration)
	case 2:
		// Instantly reveal a hidden point on the map.
		return q.db.PointOfInterest().Unlock(ctx, metadata.PointOfInterestID, teamMatch.TeamID)
	case 3:
		// 	// Instantly capture a tier one challenge.
		return q.captureChallenge(ctx, metadata.PointOfInterestID, teamMatch.TeamID, 1)
	case 4:
		// Instantly capture a tier two challenge.
		return q.captureChallenge(ctx, metadata.PointOfInterestID, teamMatch.TeamID, 2)
	case 5:
		// Instantly capture a tier three challenge.
		return q.captureChallenge(ctx, metadata.PointOfInterestID, teamMatch.TeamID, 3)
	case 6:
		// Steal all of another team's items. Must be within a 100 meter radius of the target team to use.
		return q.db.InventoryItem().StealItems(ctx, teamMatch.TeamID, metadata.TargetTeamID)
	case 7:
		// Inflict a wound on another team.
		return q.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, metadata.TargetTeamID, 10, 1)
	case 8:
		// Hold in your inventory to increase your score by 1.
		return nil
	case 9:
		// Steal an item from another team.
		items, err := q.db.InventoryItem().GetTeamsItems(ctx, metadata.TargetTeamID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		var randomItem models.TeamInventoryItem
		for {
			index := rand.Intn(len(items))
			if index != 10 {
				randomItem = items[index]
				break
			}
		}
		return q.db.InventoryItem().StealItem(ctx, teamMatch.TeamID, metadata.TargetTeamID, randomItem.InventoryItemID)
	case 10:
		// Steal an item from another team.
		return nil
	default:
		return errors.New("no effect found for this item")
	}
}

func (q *client) AddEffectToMatch(ctx context.Context, matchID uuid.UUID, inventoryItemID int, teamID uuid.UUID, duration time.Duration) error {
	return q.db.InventoryItem().ApplyInventoryItem(ctx, matchID, inventoryItemID, teamID, duration)
}

func (q *client) captureChallenge(ctx context.Context, pointOfInterestID uuid.UUID, teamID uuid.UUID, tier int) error {
	challenge, err := q.db.PointOfInterestChallenge().GetChallengeForPointOfInterest(ctx, pointOfInterestID, tier)
	if err != nil {
		return err
	}

	if _, err := q.db.PointOfInterestChallenge().SubmitAnswerForChallenge(ctx, challenge.ID, teamID, challengeAnswer, "", true); err != nil {
		return err
	}
	return nil
}
