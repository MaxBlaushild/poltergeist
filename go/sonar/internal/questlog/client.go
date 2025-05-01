package questlog

import (
	"context"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type QuestObjective struct {
	Challenge   models.PointOfInterestChallenge             `json:"challenge"`
	IsCompleted bool                                        `json:"isCompleted"`
	Submissions []models.PointOfInterestChallengeSubmission `json:"submissions"`
	NextNode    *QuestNode                                  `json:"nextNode"`
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
	ID          uuid.UUID  `json:"id"`
}

type QuestChallenge struct {
	Challenge models.PointOfInterestChallenge `json:"challenge"`
	QuestID   uuid.UUID                       `json:"questId"`
}

type QuestLog struct {
	Quests          []Quest                        `json:"quests"`
	PendingTasks    map[uuid.UUID][]QuestChallenge `json:"pendingTasks"`
	CompletedTasks  map[uuid.UUID][]QuestChallenge `json:"completedTasks"`
	TrackedQuestIDs []uuid.UUID                    `json:"trackedQuestIds"`
}

type QuestlogClient interface {
	GetQuestLog(ctx context.Context, userID uuid.UUID, lat float64, lng float64, tags []string) (*QuestLog, error)
}

type questlogClient struct {
	dbClient db.DbClient
}

const (
	NearbyDistanceInMeters = 30000 // 30km
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

	trackedQuests, err := c.GetTrackedQuests(ctx, userID)
	if err != nil {
		return nil, err
	}

	submissions, err := c.dbClient.PointOfInterestChallenge().GetSubmissionsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	uniqueGroups := c.mergeQuestGroups(groups, startedQuests, trackedQuests)
	trackedQuestIDs := c.getTrackedQuestIDs(trackedQuests)
	allQuests := c.convertToQuestSlice(uniqueGroups)

	pendingTasks := make(map[uuid.UUID][]QuestChallenge)
	completedTasks := make(map[uuid.UUID][]QuestChallenge)
	quests := c.buildQuestsWithCompletedNodes(allQuests, submissions, pendingTasks, completedTasks)

	return &QuestLog{
		Quests:          quests,
		PendingTasks:    pendingTasks,
		CompletedTasks:  completedTasks,
		TrackedQuestIDs: trackedQuestIDs,
	}, nil
}

func (c *questlogClient) mergeQuestGroups(groups ...[]models.PointOfInterestGroup) map[uuid.UUID]models.PointOfInterestGroup {
	uniqueGroups := make(map[uuid.UUID]models.PointOfInterestGroup)
	for _, groupSet := range groups {
		for _, group := range groupSet {
			uniqueGroups[group.ID] = group
		}
	}
	return uniqueGroups
}

func (c *questlogClient) getTrackedQuestIDs(trackedQuests []models.PointOfInterestGroup) []uuid.UUID {
	ids := make([]uuid.UUID, len(trackedQuests))
	for i, quest := range trackedQuests {
		ids[i] = quest.ID
	}
	return ids
}

func (c *questlogClient) convertToQuestSlice(groupMap map[uuid.UUID]models.PointOfInterestGroup) []models.PointOfInterestGroup {
	quests := make([]models.PointOfInterestGroup, 0, len(groupMap))
	for _, group := range groupMap {
		quests = append(quests, group)
	}
	return quests
}

func (c *questlogClient) buildQuestsWithCompletedNodes(
	quests []models.PointOfInterestGroup,
	submissions []models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) []Quest {
	result := make([]Quest, 0, len(quests))
	for _, group := range quests {
		quest := c.ConvertToQuestWithCompletedNodes(group, submissions, pendingTasks, completedTasks)
		result = append(result, *quest)
	}
	return result
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

func (c *questlogClient) GetTrackedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) {
	shallowGroups, err := c.dbClient.TrackedPointOfInterestGroup().GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]uuid.UUID, 0, len(shallowGroups))
	for _, group := range shallowGroups {
		groupIDs = append(groupIDs, group.PointOfInterestGroupID)
	}

	groups, err := c.dbClient.PointOfInterestGroup().FindByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *questlogClient) GetSubmissionLookup(submissions []models.PointOfInterestChallengeSubmission) map[uuid.UUID][]models.PointOfInterestChallengeSubmission {
	submissionsByChallenge := make(map[uuid.UUID][]models.PointOfInterestChallengeSubmission)
	for _, submission := range submissions {
		submissionsByChallenge[submission.PointOfInterestChallengeID] = append(submissionsByChallenge[submission.PointOfInterestChallengeID], submission)
	}
	return submissionsByChallenge
}

func (c *questlogClient) IsChallengeCompleted(challenge models.PointOfInterestChallenge, submissionsLookup map[uuid.UUID][]models.PointOfInterestChallengeSubmission) (bool, bool) {
	submissions, exists := submissionsLookup[challenge.ID]
	if exists {
		for _, submission := range submissions {
			if submission.IsCorrect != nil && *submission.IsCorrect {
				return true, exists
			}
		}
	}
	return false, exists
}

func (c *questlogClient) buildObjectives(
	poi models.PointOfInterest,
	group models.PointOfInterestGroup,
	submissionsByChallenge map[uuid.UUID][]models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) []QuestObjective {
	objectives := make([]QuestObjective, 0)
	for _, challenge := range poi.PointOfInterestChallenges {
		// if !c.isChallengeInGroup(challenge, group.ID) {
		// 	continue
		// }

		objective := c.createObjective(challenge, poi.ID, group.ID, submissionsByChallenge, pendingTasks, completedTasks)
		objectives = append(objectives, objective)
	}
	return objectives
}

func (c *questlogClient) isChallengeInGroup(challenge models.PointOfInterestChallenge, groupID uuid.UUID) bool {
	return challenge.PointOfInterestGroupID != nil && *challenge.PointOfInterestGroupID == groupID
}

func (c *questlogClient) createObjective(
	challenge models.PointOfInterestChallenge,
	poiID uuid.UUID,
	groupID uuid.UUID,
	submissionsByChallenge map[uuid.UUID][]models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) QuestObjective {
	isCompleted, exists := c.IsChallengeCompleted(challenge, submissionsByChallenge)

	objective := QuestObjective{
		Challenge:   challenge,
		IsCompleted: isCompleted,
		Submissions: []models.PointOfInterestChallengeSubmission{},
	}

	if exists {
		objective.Submissions = submissionsByChallenge[challenge.ID]
	} else {
		objective.Submissions = []models.PointOfInterestChallengeSubmission{}
	}

	c.updateTaskMaps(challenge, poiID, groupID, isCompleted, pendingTasks, completedTasks)

	return objective
}

func (c *questlogClient) updateTaskMaps(
	challenge models.PointOfInterestChallenge,
	poiID uuid.UUID,
	groupID uuid.UUID,
	isCompleted bool,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) {
	questChallenge := QuestChallenge{
		Challenge: challenge,
		QuestID:   groupID,
	}

	if isCompleted {
		c.addTaskIfNotExists(poiID, questChallenge, completedTasks)
	} else {
		c.addTaskIfNotExists(poiID, questChallenge, pendingTasks)
	}
}

func (c *questlogClient) addTaskIfNotExists(poiID uuid.UUID, task QuestChallenge, taskMap map[uuid.UUID][]QuestChallenge) {
	for _, existingTask := range taskMap[poiID] {
		if existingTask.Challenge.ID == task.Challenge.ID {
			return
		}
	}
	taskMap[poiID] = append(taskMap[poiID], task)
}

func (c *questlogClient) ConvertToQuestWithCompletedNodes(
	group models.PointOfInterestGroup,
	submissions []models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) *Quest {
	submissionsByChallenge := c.GetSubmissionLookup(submissions)
	visited := make(map[uuid.UUID]bool)
	rootGroupMember := group.GetRootMember()
	rootNode := c.buildQuestNodeTree(rootGroupMember, group, submissionsByChallenge, pendingTasks, completedTasks, visited)

	return &Quest{
		IsCompleted: c.isQuestCompleted(rootNode),
		RootNode:    rootNode,
		ImageUrl:    group.ImageUrl,
		Name:        group.Name,
		Description: group.Description,
		ID:          group.ID,
	}
}

func (c *questlogClient) buildQuestNodeTree(
	member *models.PointOfInterestGroupMember,
	group models.PointOfInterestGroup,
	submissionsByChallenge map[uuid.UUID][]models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
	visited map[uuid.UUID]bool,
) *QuestNode {
	if visited[member.PointOfInterest.ID] {
		log.Printf("Already visited POI %s in group %s, skipping to prevent infinite recursion", member.PointOfInterest.ID, group.ID)
		return nil
	}
	visited[member.PointOfInterest.ID] = true

	node := &QuestNode{
		PointOfInterest: member.PointOfInterest,
		Objectives:      c.buildObjectives(member.PointOfInterest, group, submissionsByChallenge, pendingTasks, completedTasks),
		Children:        make(map[uuid.UUID]*QuestNode),
	}

	c.buildChildNodes(node, member.Children, group, submissionsByChallenge, pendingTasks, completedTasks, visited)

	return node
}

func (c *questlogClient) buildChildNodes(
	parentNode *QuestNode,
	children []models.PointOfInterestChildren,
	group models.PointOfInterestGroup,
	submissionsByChallenge map[uuid.UUID][]models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
	visited map[uuid.UUID]bool,
) {
	for _, child := range children {
		childMember := c.findGroupMember(child.NextPointOfInterestGroupMemberID, group.GroupMembers)
		if childMember == nil {
			continue
		}

		childNode := c.buildQuestNodeTree(childMember, group, submissionsByChallenge, pendingTasks, completedTasks, visited)
		if childNode != nil {
			parentNode.Children[child.PointOfInterestChallengeID] = childNode
			c.linkObjectiveToChildNode(parentNode, child.PointOfInterestChallengeID, childNode)
		}
	}
}

func (c *questlogClient) findGroupMember(memberID uuid.UUID, members []models.PointOfInterestGroupMember) *models.PointOfInterestGroupMember {
	for _, member := range members {
		if member.ID == memberID {
			return &member
		}
	}
	return nil
}

func (c *questlogClient) linkObjectiveToChildNode(node *QuestNode, challengeID uuid.UUID, childNode *QuestNode) {
	for i, objective := range node.Objectives {
		if objective.Challenge.ID == challengeID {
			node.Objectives[i].NextNode = childNode
		}
	}
}

func (c *questlogClient) isQuestCompleted(node *QuestNode) bool {
	if node == nil {
		return true
	}

	for _, objective := range node.Objectives {
		if !objective.IsCompleted {
			return false
		}
	}

	for _, child := range node.Children {
		if !c.isQuestCompleted(child) {
			return false
		}
	}

	return true
}
