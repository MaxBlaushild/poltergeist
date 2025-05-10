package chat

import (
	"context"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/google/uuid"
)

type client struct {
	dbClient      db.DbClient
	quartermaster quartermaster.Quartermaster
}

const (
	CaptureMessage      = "%s captured %s at tier %s."
	CompleteTaskMessage = "%s completed a task at %s."
	CompleteQuestMessage = "%s completed a quest: %s."
)

type Client interface {
	AddUseItemMessage(ctx context.Context, ownedInventoryItem models.OwnedInventoryItem, metadata quartermaster.UseItemMetadata) error
	AddUnlockMessage(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, pointOfInterestID uuid.UUID) error
	AddCaptureMessage(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, challenge *models.PointOfInterestChallenge) error
	AddCompletedQuestMessage(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, challenge *models.PointOfInterestChallenge) error
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

func (c *client) AddCompletedQuestMessage(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, challenge *models.PointOfInterestChallenge) error {
	if challenge.PointOfInterestGroupID == uuid.Nil {
		return nil
	}

	pointOfInterestGroup, err := c.dbClient.PointOfInterestGroup().FindByID(ctx, *challenge.PointOfInterestGroupID)
	if err != nil {
		return err
	}

	var teamName string
	var matchID *uuid.UUID
	if userID != nil {
		teamName = "You"
		matchID = nil
	} else {
		team, err := c.dbClient.Team().GetByID(ctx, *teamID)
		if err != nil {
			return err
		}
		teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, team.ID)
		if err != nil {
			return err
		}
		teamName = c.makeTeamName(team.ID)
		matchID = &teamMatch.MatchID
	}

	message := fmt.Sprintf(
		CompleteQuestMessage,
		teamName,
		pointOfInterestGroup.Name,
	)

	return c.dbClient.AuditItem().Create(ctx, matchID, userID, message)
}

func (c *client) AddCaptureMessage(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, challenge *models.PointOfInterestChallenge) error {
	var teamName string
	var matchID *uuid.UUID
	if userID != nil {
		teamName = "You"
		matchID = nil
	} else {
		team, err := c.dbClient.Team().GetByID(ctx, *teamID)
		if err != nil {
			return err
		}
		teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, team.ID)
		if err != nil {
			return err
		}
		teamName = c.makeTeamName(team.ID)
		matchID = &teamMatch.MatchID
	}

	var message string
	if userID != nil {
		message = fmt.Sprintf(
			CompleteTaskMessage,
			teamName,
			c.makePointOfInterestName(challenge.PointOfInterestID),
		)
	} else {
		message = fmt.Sprintf(
			CaptureMessage,
			teamName,
			c.makePointOfInterestName(challenge.PointOfInterestID),
			c.makeChallengeTierName(challenge.Tier),
		)

	return c.dbClient.AuditItem().Create(ctx, matchID, userID, message)
}

func (c *client) AddUnlockMessage(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, pointOfInterestID uuid.UUID) error {
	var userName string
	var matchID *uuid.UUID
	if userID != nil {
		userName = "You"
		matchID = nil
	} else {
		team, err := c.dbClient.Team().GetByID(ctx, *teamID)
		if err != nil {
			return err
		}
		teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, *teamID)
		if err != nil {
			return err
		}
		userName = c.makeTeamName(team.ID)
		matchID = &teamMatch.MatchID
	}

	message := fmt.Sprintf("%s unlocked %s.", userName, c.makePointOfInterestName(pointOfInterestID))

	return c.dbClient.AuditItem().Create(ctx, matchID, userID, message)
}

func (c *client) AddUseItemMessage(ctx context.Context, ownedInventoryItem models.OwnedInventoryItem, metadata quartermaster.UseItemMetadata) error {
	if ownedInventoryItem.IsTeamItem() {
		return c.addUseItemMessageForTeam(ctx, ownedInventoryItem, metadata)
	}

	return c.addUseItemMessageForUser(ctx, ownedInventoryItem, metadata)
}

func (c *client) makeUseMessage(ctx context.Context, userName string, itemName string) string {
	return fmt.Sprintf("%s used a %s", userName, itemName)
}

func (c *client) addUseItemMessageForTeam(ctx context.Context, ownedInventoryItem models.OwnedInventoryItem, metadata quartermaster.UseItemMetadata) error {
	if ownedInventoryItem.TeamID == nil {
		return nil
	}

	team, err := c.dbClient.Team().GetByID(ctx, *ownedInventoryItem.TeamID)
	if err != nil {
		return err
	}

	teamMatch, err := c.dbClient.Match().FindForTeamID(ctx, team.ID)
	if err != nil {
		return err
	}

	item, err := c.quartermaster.FindItemForItemID(ownedInventoryItem.InventoryItemID)
	if err != nil {
		return err
	}

	message := c.makeUseMessage(ctx, c.makeTeamName(team.ID), c.makeInventoryItemName(item.ID))

	if metadata.TargetTeamID != uuid.Nil {
		message += fmt.Sprintf(" on %s", c.makeTeamName(metadata.TargetTeamID))
	} else if metadata.PointOfInterestID != uuid.Nil {
		message += fmt.Sprintf(" on %s", c.makePointOfInterestName(metadata.PointOfInterestID))
	}

	message += "."

	if err := c.dbClient.AuditItem().Create(ctx, &teamMatch.MatchID, nil, message); err != nil {
		return err
	}

	return nil
}

func (c *client) addUseItemMessageForUser(ctx context.Context, ownedInventoryItem models.OwnedInventoryItem, metadata quartermaster.UseItemMetadata) error {
	if ownedInventoryItem.UserID == nil {
		return nil
	}

	item, err := c.quartermaster.FindItemForItemID(ownedInventoryItem.InventoryItemID)
	if err != nil {
		return err
	}

	message := c.makeUseMessage(ctx, "You", c.makeInventoryItemName(item.ID))

	message += "."

	return c.dbClient.AuditItem().Create(ctx, nil, ownedInventoryItem.UserID, message)
}
