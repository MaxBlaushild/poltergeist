package chat

import (
	"context"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/google/uuid"
)

type client struct {
	dbClient      db.DbClient
	quartermaster quartermaster.Quartermaster
}

type Client interface {
	AddUseItemMessage(ctx context.Context, teamInventoryItemID uuid.UUID, metadata quartermaster.UseItemMetadata) error
	AddUnlockMessage(ctx context.Context, teamID uuid.UUID, pointOfInterestID uuid.UUID) error
	AddCaptureMessage(ctx context.Context, teamID uuid.UUID, challengeID uuid.UUID) error
}

func NewClient(dbClient db.DbClient, quartermaster quartermaster.Quartermaster) Client {
	return &client{
		dbClient:      dbClient,
		quartermaster: quartermaster,
	}
}

func (c *client) makeTeamName(id uuid.UUID) string {
	return fmt.Sprintf("{Team|%s}", id)
}

func (c *client) makePointOfInterestName(id uuid.UUID) string {
	return fmt.Sprintf("{PointOfInterest|%s}", id)
}

func (c *client) makeInventoryItemName(id int) string {
	return fmt.Sprintf("{InventoryItem|%d}", id)
}

func (c *client) makeChallengeTierName(tier int) string {
	switch tier {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	case 4:
		return "IV"
	case 5:
		return "V"
	case 6:
		return "VI"
	case 7:
		return "VII"
	case 8:
		return "VIII"
	case 9:
		return "IX"
	case 10:
		return "X"
	default:
		return fmt.Sprintf("Tier %d", tier)
	}
}

func (c *client) AddCaptureMessage(ctx context.Context, teamID uuid.UUID, challengeID uuid.UUID) error {
	teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, teamID)
	if err != nil {
		return err
	}

	challenge, err := c.dbClient.PointOfInterestChallenge().FindByID(ctx, challengeID)
	if err != nil {
		return err
	}

	message := fmt.Sprintf(
		"%s captured %s at tier %s.",
		c.makeTeamName(teamID),
		c.makePointOfInterestName(challenge.PointOfInterestID),
		c.makeChallengeTierName(challenge.Tier),
	)

	return c.dbClient.AuditItem().Create(ctx, teamMatch.MatchID, message)
}

func (c *client) AddUnlockMessage(ctx context.Context, teamID uuid.UUID, pointOfInterestID uuid.UUID) error {
	teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, teamID)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("%s unlocked %s.", c.makeTeamName(teamID), c.makePointOfInterestName(pointOfInterestID))

	return c.dbClient.AuditItem().Create(ctx, teamMatch.MatchID, message)
}

func (c *client) AddUseItemMessage(ctx context.Context, teamInventoryItemID uuid.UUID, metadata quartermaster.UseItemMetadata) error {
	teamInventoryItem, err := c.dbClient.InventoryItem().FindByID(ctx, teamInventoryItemID)
	if err != nil {
		return err
	}

	team, err := c.dbClient.Team().GetByID(ctx, teamInventoryItem.TeamID)
	if err != nil {
		return err
	}

	teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, team.ID)
	if err != nil {
		return err
	}

	item, err := c.quartermaster.FindItemForItemID(teamInventoryItem.InventoryItemID)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("%s used a %s", c.makeTeamName(team.ID), c.makeInventoryItemName(item.ID))

	if metadata.TargetTeamID != uuid.Nil {
		message += fmt.Sprintf(" on %s", c.makeTeamName(metadata.TargetTeamID))
	} else if metadata.PointOfInterestID != uuid.Nil {
		message += fmt.Sprintf(" on %s", c.makePointOfInterestName(metadata.PointOfInterestID))
	}

	message += "."

	if err := c.dbClient.AuditItem().Create(ctx, teamMatch.MatchID, message); err != nil {
		return err
	}

	return nil
}
