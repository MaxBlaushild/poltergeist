package questlog

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type QuestObjective struct {
	Challenge   models.PointOfInterestChallenge             `json:"challenge"`
	IsCompleted bool                                        `json:"isCompleted"`
	Submissions []models.PointOfInterestChallengeSubmission `json:"submissions"`
}

type QuestNode struct {
	PointOfInterest models.PointOfInterest   `json:"pointOfInterest"`
	Objectives      []QuestObjective         `json:"objectives"`
	Children        map[uuid.UUID]*QuestNode `json:"children"`
}

type Quest struct {
	IsCompleted bool       `json:"isCompleted"`
	RootNode    *QuestNode `json:"rootNode"`
	ImageUrl    string     `json:"imageUrl"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}

type QuestLog struct {
	Quests         []Quest                                         `json:"quests"`
	PendingTasks   map[uuid.UUID][]models.PointOfInterestChallenge `json:"pendingTasks"`
	CompletedTasks map[uuid.UUID][]models.PointOfInterestChallenge `json:"completedTasks"`
}

type QuestlogClient interface {
	GetQuestLog(ctx context.Context, userID uuid.UUID, lat float64, lng float64, tags []string) (*QuestLog, error)
}

type questlogClient struct {
	dbClient db.DbClient
}

const (
	NearbyDistanceInMeters = 51900 // Approximately the width of Delhi in meters (51.9 km)
)

func NewClient(dbClient db.DbClient) QuestlogClient {
	return &questlogClient{dbClient: dbClient}
}

func (c *questlogClient) GetQuestLog(ctx context.Context, userID uuid.UUID, lat float64, lng float64, tags []string) (*QuestLog, error) {
	groups, err := c.GetNearbyQuests(ctx, userID, lat, lng, tags)
	if err != nil {
		return nil, err
	}
	startedQuests, err := c.GetStartedQuests(ctx, userID)
	if err != nil {
		return nil, err
	}
	submissions, err := c.dbClient.PointOfInterestChallenge().GetSubmissionsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	uniqueGroups := make(map[uuid.UUID]models.PointOfInterestGroup)
	for _, group := range groups {
		uniqueGroups[group.ID] = group
	}
	for _, group := range startedQuests {
		uniqueGroups[group.ID] = group
	}

	allQuests := make([]models.PointOfInterestGroup, 0, len(uniqueGroups))
	for _, group := range uniqueGroups {
		allQuests = append(allQuests, group)
	}
	pendingTasks := make(map[uuid.UUID][]models.PointOfInterestChallenge)
	completedTasks := make(map[uuid.UUID][]models.PointOfInterestChallenge)
	quests := make([]Quest, 0)
	for _, group := range allQuests {
		quest := c.ConvertToQuestWithCompletedNodes(group, submissions, pendingTasks, completedTasks)
		quests = append(quests, *quest)
	}

	return &QuestLog{
		Quests:         quests,
		PendingTasks:   pendingTasks,
		CompletedTasks: completedTasks,
	}, nil
}

func (c *questlogClient) GetNearbyQuests(ctx context.Context, userID uuid.UUID, lat float64, lng float64, tags []string) ([]models.PointOfInterestGroup, error) {
	groups, err := c.dbClient.PointOfInterestGroup().GetNearbyQuests(ctx, userID, lat, lng, NearbyDistanceInMeters, tags)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *questlogClient) GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) {
	groups, err := c.dbClient.PointOfInterestGroup().GetStartedQuests(ctx, userID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *questlogClient) ConvertToQuestWithCompletedNodes(
	group models.PointOfInterestGroup,
	submissions []models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]models.PointOfInterestChallenge,
	completedTasks map[uuid.UUID][]models.PointOfInterestChallenge,
) *Quest {
	submissionsByChallenge := make(map[uuid.UUID][]models.PointOfInterestChallengeSubmission)
	for _, submission := range submissions {
		submissionsByChallenge[submission.PointOfInterestChallengeID] = append(submissionsByChallenge[submission.PointOfInterestChallengeID], submission)
	}

	// Helper function to build objectives for a point of interest
	buildObjectives := func(poi models.PointOfInterest) []QuestObjective {
		objectives := make([]QuestObjective, 0)
		for _, challenge := range poi.PointOfInterestChallenges {
			if challenge.PointOfInterestGroupID != nil && *challenge.PointOfInterestGroupID != group.ID {
				continue
			}

			submissions, exists := submissionsByChallenge[challenge.ID]
			isCompleted := false
			if exists {
				for _, submission := range submissions {
					if submission.IsCorrect != nil && *submission.IsCorrect {
						isCompleted = true
						break
					}
				}
			}

			objective := QuestObjective{
				Challenge:   challenge,
				IsCompleted: isCompleted,
				Submissions: []models.PointOfInterestChallengeSubmission{},
			}
			if exists {
				objective.Submissions = submissions
			}
			objectives = append(objectives, objective)

			// Add to pending or completed tasks
			if isCompleted {
				completedTasks[poi.ID] = append(completedTasks[poi.ID], challenge)
			} else {
				pendingTasks[poi.ID] = append(pendingTasks[poi.ID], challenge)
			}
		}
		return objectives
	}

	// Helper function to recursively build quest nodes
	var buildQuestNode func(poi models.PointOfInterest, children []models.PointOfInterestChildren) *QuestNode
	buildQuestNode = func(poi models.PointOfInterest, children []models.PointOfInterestChildren) *QuestNode {
		node := &QuestNode{
			PointOfInterest: poi,
			Objectives:      buildObjectives(poi),
			Children:        make(map[uuid.UUID]*QuestNode),
		}

		for _, child := range children {
			var childGroupMember *models.PointOfInterestGroupMember
			for _, member := range group.GroupMembers {
				if member.PointOfInterestID == child.PointOfInterestID {
					childGroupMember = &member
					break
				}
			}
			if childGroupMember == nil {
				continue
			}
			childNode := buildQuestNode(childGroupMember.PointOfInterest, childGroupMember.Children)
			node.Children[child.PointOfInterestChallengeID] = childNode
		}

		return node
	}

	// Build the root node and all its children
	rootNode := buildQuestNode(group.GroupMembers[0].PointOfInterest, group.GroupMembers[0].Children)

	// Check if all objectives are completed
	allCompleted := true
	var checkNode func(node *QuestNode) bool
	checkNode = func(node *QuestNode) bool {
		// Check objectives for current node
		for _, objective := range node.Objectives {
			if !objective.IsCompleted {
				return false
			}
		}

		// Recursively check all child nodes
		for _, child := range node.Children {
			if !checkNode(child) {
				return false
			}
		}

		return true
	}

	allCompleted = checkNode(rootNode)

	return &Quest{
		IsCompleted: allCompleted,
		RootNode:    rootNode,
		ImageUrl:    group.ImageUrl,
		Name:        group.Name,
		Description: group.Description,
	}
}
