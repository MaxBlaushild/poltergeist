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
	Reward      int                                         `json:"reward"`
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
	GetQuestLog(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) (*QuestLog, error)
}

type questlogClient struct {
	dbClient db.DbClient
}

const (
	NearbyDistanceInMeters = 30000 // 30km
)

func NewClient(dbClient db.DbClient) QuestlogClient {
	log.Printf("Creating new QuestlogClient")
	return &questlogClient{dbClient: dbClient}
}

func (c *questlogClient) GetQuestLog(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) (*QuestLog, error) {
	log.Printf("Getting quest log for user %s in zone %s with tags %v", userID, zoneID, tags)

	groups, err := c.GetQuestsInZone(ctx, userID, zoneID, tags)
	if err != nil {
		log.Printf("Error getting nearby quests: %v", err)
		return nil, err
	}

	startedQuests, err := c.GetStartedQuests(ctx, userID)
	if err != nil {
		log.Printf("Error getting started quests: %v", err)
		return nil, err
	}

	trackedQuests, err := c.GetTrackedQuests(ctx, userID)
	if err != nil {
		log.Printf("Error getting tracked quests: %v", err)
		return nil, err
	}

	submissions, err := c.dbClient.PointOfInterestChallenge().GetSubmissionsForUser(ctx, userID)
	if err != nil {
		log.Printf("Error getting challenge submissions: %v", err)
		return nil, err
	}

	log.Printf("Found %d nearby quests, %d started quests, %d tracked quests, and %d submissions", len(groups), len(startedQuests), len(trackedQuests), len(submissions))

	uniqueGroups := c.mergeQuestGroups(groups, startedQuests, trackedQuests)
	trackedQuestIDs := c.getTrackedQuestIDs(trackedQuests)
	allQuests := c.convertToQuestSlice(uniqueGroups)

	pendingTasks := make(map[uuid.UUID][]QuestChallenge)
	completedTasks := make(map[uuid.UUID][]QuestChallenge)
	quests := c.buildQuestsWithCompletedNodes(allQuests, submissions, pendingTasks, completedTasks)

	log.Printf("Built quest log with %d quests, %d pending tasks, %d completed tasks", len(quests), len(pendingTasks), len(completedTasks))

	return &QuestLog{
		Quests:          quests,
		PendingTasks:    pendingTasks,
		CompletedTasks:  completedTasks,
		TrackedQuestIDs: trackedQuestIDs,
	}, nil
}

func (c *questlogClient) mergeQuestGroups(groups ...[]models.PointOfInterestGroup) map[uuid.UUID]models.PointOfInterestGroup {
	log.Printf("Merging %d quest group sets", len(groups))
	uniqueGroups := make(map[uuid.UUID]models.PointOfInterestGroup)
	for i, groupSet := range groups {
		log.Printf("Processing group set %d with %d groups", i+1, len(groupSet))
		for _, group := range groupSet {
			uniqueGroups[group.ID] = group
		}
	}
	log.Printf("Merged into %d unique groups", len(uniqueGroups))
	return uniqueGroups
}

func (c *questlogClient) getTrackedQuestIDs(trackedQuests []models.PointOfInterestGroup) []uuid.UUID {
	log.Printf("Getting tracked quest IDs from %d quests", len(trackedQuests))
	ids := make([]uuid.UUID, len(trackedQuests))
	for i, quest := range trackedQuests {
		ids[i] = quest.ID
	}
	return ids
}

func (c *questlogClient) convertToQuestSlice(groupMap map[uuid.UUID]models.PointOfInterestGroup) []models.PointOfInterestGroup {
	log.Printf("Converting map of %d groups to slice", len(groupMap))
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
	log.Printf("Building quests with completed nodes from %d quests and %d submissions", len(quests), len(submissions))
	result := make([]Quest, 0, len(quests))
	for i, group := range quests {
		log.Printf("Processing quest %d/%d: %s", i+1, len(quests), group.Name)
		quest := c.ConvertToQuestWithCompletedNodes(group, submissions, pendingTasks, completedTasks)
		result = append(result, *quest)
	}
	return result
}

func (c *questlogClient) GetQuestsInZone(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) ([]models.PointOfInterestGroup, error) {
	log.Printf("Getting quests in zone %s for user %s with tags %v", zoneID, userID, tags)
	groups, err := c.dbClient.PointOfInterestGroup().GetQuestsInZone(ctx, zoneID, tags)
	if err != nil {
		log.Printf("Error getting quests in zone: %v", err)
		return nil, err
	}
	log.Printf("Found %d quests in zone", len(groups))
	return groups, nil
}

func (c *questlogClient) GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) {
	log.Printf("Getting started quests for user %s", userID)
	groups, err := c.dbClient.PointOfInterestGroup().GetStartedQuests(ctx, userID)
	if err != nil {
		log.Printf("Error getting started quests: %v", err)
		return nil, err
	}
	log.Printf("Found %d started quests", len(groups))
	return groups, nil
}

func (c *questlogClient) GetTrackedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) {
	log.Printf("Getting tracked quests for user %s", userID)
	shallowGroups, err := c.dbClient.TrackedPointOfInterestGroup().GetByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error getting tracked quest IDs: %v", err)
		return nil, err
	}

	groupIDs := make([]uuid.UUID, 0, len(shallowGroups))
	for _, group := range shallowGroups {
		groupIDs = append(groupIDs, group.PointOfInterestGroupID)
	}

	log.Printf("Found %d tracked quest IDs, fetching full quest details", len(groupIDs))
	groups, err := c.dbClient.PointOfInterestGroup().FindByIDs(ctx, groupIDs)
	if err != nil {
		log.Printf("Error getting tracked quests: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d tracked quests", len(groups))
	return groups, nil
}

func (c *questlogClient) GetSubmissionLookup(submissions []models.PointOfInterestChallengeSubmission) map[uuid.UUID][]models.PointOfInterestChallengeSubmission {
	log.Printf("Building submission lookup from %d submissions", len(submissions))
	submissionsByChallenge := make(map[uuid.UUID][]models.PointOfInterestChallengeSubmission)
	for _, submission := range submissions {
		submissionsByChallenge[submission.PointOfInterestChallengeID] = append(submissionsByChallenge[submission.PointOfInterestChallengeID], submission)
	}
	log.Printf("Built lookup with %d challenge entries", len(submissionsByChallenge))
	return submissionsByChallenge
}

func (c *questlogClient) IsChallengeCompleted(challenge models.PointOfInterestChallenge, submissionsLookup map[uuid.UUID][]models.PointOfInterestChallengeSubmission) (bool, bool) {
	log.Printf("Checking completion status for challenge %s", challenge.ID)
	submissions, exists := submissionsLookup[challenge.ID]
	if exists {
		for _, submission := range submissions {
			if submission.IsCorrect != nil && *submission.IsCorrect {
				log.Printf("Challenge %s is completed", challenge.ID)
				return true, exists
			}
		}
	}
	log.Printf("Challenge %s is not completed", challenge.ID)
	return false, exists
}

func (c *questlogClient) buildObjectives(
	poi models.PointOfInterest,
	group models.PointOfInterestGroup,
	submissionsByChallenge map[uuid.UUID][]models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) []QuestObjective {
	log.Printf("Building objectives for POI %s in group %s", poi.ID, group.ID)
	objectives := make([]QuestObjective, 0)
	for _, challenge := range poi.PointOfInterestChallenges {
		if !c.isChallengeInGroup(challenge, group.ID) {
			log.Printf("Challenge %s is not in group %s, skipping", challenge.ID, group.ID)
			continue
		}

		objective := c.createObjective(challenge, poi.ID, group.ID, submissionsByChallenge, pendingTasks, completedTasks)
		objectives = append(objectives, objective)
	}
	log.Printf("Built %d objectives for POI %s", len(objectives), poi.ID)
	return objectives
}

func (c *questlogClient) isChallengeInGroup(challenge models.PointOfInterestChallenge, groupID uuid.UUID) bool {
	log.Printf("Checking if challenge %s belongs to group %s", challenge.ID, groupID)
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
	log.Printf("Creating objective for challenge %s in POI %s", challenge.ID, poiID)
	isCompleted, exists := c.IsChallengeCompleted(challenge, submissionsByChallenge)

	objective := QuestObjective{
		Challenge:   challenge,
		IsCompleted: isCompleted,
		Submissions: []models.PointOfInterestChallengeSubmission{},
		Reward:      challenge.InventoryItemID,
	}

	if exists {
		log.Printf("Found %d submissions for challenge %s", len(submissionsByChallenge[challenge.ID]), challenge.ID)
		objective.Submissions = submissionsByChallenge[challenge.ID]
	} else {
		log.Printf("No submissions found for challenge %s", challenge.ID)
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
	log.Printf("Updating task maps for challenge %s in POI %s", challenge.ID, poiID)
	questChallenge := QuestChallenge{
		Challenge: challenge,
		QuestID:   groupID,
	}

	if isCompleted {
		log.Printf("Adding completed task for challenge %s", challenge.ID)
		c.addTaskIfNotExists(poiID, questChallenge, completedTasks)
	} else {
		log.Printf("Adding pending task for challenge %s", challenge.ID)
		c.addTaskIfNotExists(poiID, questChallenge, pendingTasks)
	}
}

func (c *questlogClient) addTaskIfNotExists(poiID uuid.UUID, task QuestChallenge, taskMap map[uuid.UUID][]QuestChallenge) {
	log.Printf("Checking if task for challenge %s already exists for POI %s", task.Challenge.ID, poiID)
	for _, existingTask := range taskMap[poiID] {
		if existingTask.Challenge.ID == task.Challenge.ID {
			log.Printf("Task already exists, skipping")
			return
		}
	}
	log.Printf("Adding new task for challenge %s", task.Challenge.ID)
	taskMap[poiID] = append(taskMap[poiID], task)
}

func (c *questlogClient) ConvertToQuestWithCompletedNodes(
	group models.PointOfInterestGroup,
	submissions []models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
) *Quest {
	log.Printf("Converting group %s to quest with completed nodes", group.ID)
	submissionsByChallenge := c.GetSubmissionLookup(submissions)
	visited := make(map[uuid.UUID]bool)
	rootGroupMember := group.GetRootMember()
	rootNode := c.buildQuestNodeTree(rootGroupMember, group, submissionsByChallenge, pendingTasks, completedTasks, visited)

	quest := &Quest{
		IsCompleted: c.isQuestCompleted(rootNode),
		RootNode:    rootNode,
		ImageUrl:    group.ImageUrl,
		Name:        group.Name,
		Description: group.Description,
		ID:          group.ID,
	}
	log.Printf("Converted group %s to quest (completed: %v)", group.ID, quest.IsCompleted)
	return quest
}

func (c *questlogClient) buildQuestNodeTree(
	member *models.PointOfInterestGroupMember,
	group models.PointOfInterestGroup,
	submissionsByChallenge map[uuid.UUID][]models.PointOfInterestChallengeSubmission,
	pendingTasks map[uuid.UUID][]QuestChallenge,
	completedTasks map[uuid.UUID][]QuestChallenge,
	visited map[uuid.UUID]bool,
) *QuestNode {
	log.Printf("Building quest node tree for POI %s in group %s", member.PointOfInterest.ID, group.ID)

	if visited[member.PointOfInterest.ID] {
		log.Printf("Already visited POI %s in group %s, skipping to prevent infinite recursion", member.PointOfInterest.ID, group.ID)
		return nil
	}
	visited[member.PointOfInterest.ID] = true

	objectives := c.buildObjectives(member.PointOfInterest, group, submissionsByChallenge, pendingTasks, completedTasks)
	log.Printf("Built %d objectives for POI %s", len(objectives), member.PointOfInterest.ID)

	node := &QuestNode{
		PointOfInterest: member.PointOfInterest,
		Objectives:      objectives,
		Children:        make(map[uuid.UUID]*QuestNode),
	}

	log.Printf("Building child nodes for POI %s, found %d children", member.PointOfInterest.ID, len(member.Children))
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
	for i, child := range children {
		log.Printf("Processing child %d/%d for POI %s, challenge ID: %s", i+1, len(children), parentNode.PointOfInterest.ID, child.PointOfInterestChallengeID)

		childMember := c.findGroupMember(child.NextPointOfInterestGroupMemberID, group.GroupMembers)
		if childMember == nil {
			log.Printf("Could not find group member with ID %s for challenge %s", child.NextPointOfInterestGroupMemberID, child.PointOfInterestChallengeID)
			continue
		}

		log.Printf("Found child member %s for challenge %s", childMember.ID, child.PointOfInterestChallengeID)
		childNode := c.buildQuestNodeTree(childMember, group, submissionsByChallenge, pendingTasks, completedTasks, visited)

		if childNode != nil {
			log.Printf("Successfully built child node for POI %s, linking to parent POI %s", childNode.PointOfInterest.ID, parentNode.PointOfInterest.ID)
			parentNode.Children[child.PointOfInterestChallengeID] = childNode
			c.linkObjectiveToChildNode(parentNode, child.PointOfInterestChallengeID, childNode)
		} else {
			log.Printf("Child node was nil for challenge %s", child.PointOfInterestChallengeID)
		}
	}
}

func (c *questlogClient) findGroupMember(memberID uuid.UUID, members []models.PointOfInterestGroupMember) *models.PointOfInterestGroupMember {
	log.Printf("Searching for group member with ID %s among %d members", memberID, len(members))
	for _, member := range members {
		if member.ID == memberID {
			log.Printf("Found matching group member %s", memberID)
			return &member
		}
	}
	log.Printf("No matching group member found for ID %s", memberID)
	return nil
}

func (c *questlogClient) linkObjectiveToChildNode(node *QuestNode, challengeID uuid.UUID, childNode *QuestNode) {
	log.Printf("Linking objectives for POI %s to child POI %s", node.PointOfInterest.ID, childNode.PointOfInterest.ID)
	for i, objective := range node.Objectives {
		if objective.Challenge.ID == challengeID {
			log.Printf("Found matching objective for challenge %s, linking to child node", challengeID)
			node.Objectives[i].NextNode = childNode
		}
	}
}

func (c *questlogClient) isQuestCompleted(node *QuestNode) bool {
	if node == nil {
		log.Printf("Node is nil, considering quest completed")
		return true
	}

	log.Printf("Checking completion status for POI %s", node.PointOfInterest.ID)

	for _, objective := range node.Objectives {
		if !objective.IsCompleted {
			log.Printf("Objective %s for POI %s is not completed", objective.Challenge.ID, node.PointOfInterest.ID)
			return false
		}
	}

	log.Printf("All objectives completed for POI %s, checking children", node.PointOfInterest.ID)
	for _, child := range node.Children {
		if !c.isQuestCompleted(child) {
			log.Printf("Child POI %s is not completed", child.PointOfInterest.ID)
			return false
		}
	}

	log.Printf("POI %s and all its children are completed", node.PointOfInterest.ID)
	return true
}
