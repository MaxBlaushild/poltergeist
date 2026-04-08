package models

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

func NormalizeStoryFlagKey(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ResolveCurrentQuestNode(
	nodes []QuestNode,
	completed map[uuid.UUID]bool,
	activeStoryFlags map[string]bool,
) (*QuestNode, []uuid.UUID) {
	if len(nodes) == 0 {
		return nil, nil
	}

	sortedNodes := append([]QuestNode(nil), nodes...)
	sort.SliceStable(sortedNodes, func(i, j int) bool {
		if sortedNodes[i].OrderIndex != sortedNodes[j].OrderIndex {
			return sortedNodes[i].OrderIndex < sortedNodes[j].OrderIndex
		}
		return sortedNodes[i].CreatedAt.Before(sortedNodes[j].CreatedAt)
	})

	localCompleted := make(map[uuid.UUID]bool, len(completed))
	for nodeID, isCompleted := range completed {
		localCompleted[nodeID] = isCompleted
	}

	autoCompleted := make([]uuid.UUID, 0)
	for _, node := range sortedNodes {
		if localCompleted[node.ID] {
			continue
		}
		storyFlagKey := node.StoryFlagKeyNormalized()
		if storyFlagKey != "" && activeStoryFlags[storyFlagKey] {
			localCompleted[node.ID] = true
			autoCompleted = append(autoCompleted, node.ID)
			continue
		}

		nodeCopy := node
		return &nodeCopy, autoCompleted
	}

	return nil, autoCompleted
}
