package models

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

type questNodeGraph struct {
	sortedNodes []QuestNode
	nodesByID   map[uuid.UUID]QuestNode
	inbound     map[uuid.UUID]int
}

func buildQuestNodeGraph(nodes []QuestNode) questNodeGraph {
	sortedNodes := append([]QuestNode(nil), nodes...)
	sort.SliceStable(sortedNodes, func(i, j int) bool {
		if sortedNodes[i].OrderIndex != sortedNodes[j].OrderIndex {
			return sortedNodes[i].OrderIndex < sortedNodes[j].OrderIndex
		}
		return sortedNodes[i].CreatedAt.Before(sortedNodes[j].CreatedAt)
	})

	nodesByID := make(map[uuid.UUID]QuestNode, len(sortedNodes))
	inbound := make(map[uuid.UUID]int, len(sortedNodes))
	for _, node := range sortedNodes {
		nodesByID[node.ID] = node
	}
	for _, node := range sortedNodes {
		for _, child := range node.Children {
			if _, ok := nodesByID[child.NextQuestNodeID]; !ok {
				continue
			}
			inbound[child.NextQuestNodeID]++
		}
	}

	return questNodeGraph{
		sortedNodes: sortedNodes,
		nodesByID:   nodesByID,
		inbound:     inbound,
	}
}

func (g questNodeGraph) entryNode() *QuestNode {
	for _, node := range g.sortedNodes {
		if g.inbound[node.ID] == 0 {
			nodeCopy := node
			return &nodeCopy
		}
	}
	if len(g.sortedNodes) == 0 {
		return nil
	}
	nodeCopy := g.sortedNodes[0]
	return &nodeCopy
}

func (g questNodeGraph) nodeByID(nodeID *uuid.UUID) *QuestNode {
	if nodeID == nil || *nodeID == uuid.Nil {
		return nil
	}
	node, ok := g.nodesByID[*nodeID]
	if !ok {
		return nil
	}
	nodeCopy := node
	return &nodeCopy
}

func (g questNodeGraph) transitionNode(
	nodeID uuid.UUID,
	outcome QuestNodeTransitionOutcome,
) *QuestNode {
	node, ok := g.nodesByID[nodeID]
	if !ok {
		return nil
	}

	for _, child := range node.Children {
		if child.TransitionOutcome() != outcome {
			continue
		}
		nextNode, exists := g.nodesByID[child.NextQuestNodeID]
		if !exists {
			continue
		}
		nextNodeCopy := nextNode
		return &nextNodeCopy
	}

	if outcome != QuestNodeTransitionOutcomeSuccess {
		return nil
	}

	for index, candidate := range g.sortedNodes {
		if candidate.ID != nodeID {
			continue
		}
		if index+1 >= len(g.sortedNodes) {
			return nil
		}
		nextNodeCopy := g.sortedNodes[index+1]
		return &nextNodeCopy
	}

	return nil
}

func (g questNodeGraph) bootstrapCurrentNode(
	progress map[uuid.UUID]QuestNodeProgressStatus,
) *QuestNode {
	current := g.entryNode()
	visited := map[uuid.UUID]struct{}{}
	for current != nil {
		if _, seen := visited[current.ID]; seen {
			return nil
		}
		visited[current.ID] = struct{}{}

		switch NormalizeQuestNodeProgressStatus(string(progress[current.ID])) {
		case QuestNodeProgressStatusCompleted:
			current = g.transitionNode(current.ID, QuestNodeTransitionOutcomeSuccess)
		case QuestNodeProgressStatusFailed:
			if current.FailurePolicyNormalized() == QuestNodeFailurePolicyTransition {
				nextNode := g.transitionNode(current.ID, QuestNodeTransitionOutcomeFailure)
				if nextNode != nil {
					current = nextNode
					continue
				}
			}
			return current
		default:
			return current
		}
	}
	return nil
}

func NormalizeStoryFlagKey(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ResolveCurrentQuestNode(
	nodes []QuestNode,
	currentNodeID *uuid.UUID,
	progress map[uuid.UUID]QuestNodeProgressStatus,
	activeStoryFlags map[string]bool,
) (*QuestNode, []uuid.UUID) {
	if len(nodes) == 0 {
		return nil, nil
	}

	localProgress := make(map[uuid.UUID]QuestNodeProgressStatus, len(progress))
	for nodeID, status := range progress {
		localProgress[nodeID] = NormalizeQuestNodeProgressStatus(string(status))
	}

	graph := buildQuestNodeGraph(nodes)
	currentNode := graph.nodeByID(currentNodeID)
	if currentNode == nil {
		currentNode = graph.bootstrapCurrentNode(localProgress)
	}

	autoCompleted := make([]uuid.UUID, 0)
	visited := map[uuid.UUID]struct{}{}
	for currentNode != nil {
		if _, seen := visited[currentNode.ID]; seen {
			return nil, autoCompleted
		}
		visited[currentNode.ID] = struct{}{}

		switch localProgress[currentNode.ID] {
		case QuestNodeProgressStatusCompleted:
			currentNode = graph.transitionNode(currentNode.ID, QuestNodeTransitionOutcomeSuccess)
			continue
		case QuestNodeProgressStatusFailed:
			if currentNode.FailurePolicyNormalized() == QuestNodeFailurePolicyTransition {
				nextNode := graph.transitionNode(currentNode.ID, QuestNodeTransitionOutcomeFailure)
				if nextNode != nil {
					currentNode = nextNode
					continue
				}
			}
		}

		storyFlagKey := currentNode.StoryFlagKeyNormalized()
		if storyFlagKey != "" && activeStoryFlags[storyFlagKey] {
			localProgress[currentNode.ID] = QuestNodeProgressStatusCompleted
			autoCompleted = append(autoCompleted, currentNode.ID)
			currentNode = graph.transitionNode(currentNode.ID, QuestNodeTransitionOutcomeSuccess)
			continue
		}

		nodeCopy := *currentNode
		return &nodeCopy, autoCompleted
	}

	return nil, autoCompleted
}

func ResolveNextQuestNode(
	nodes []QuestNode,
	currentNodeID uuid.UUID,
	outcome QuestNodeTransitionOutcome,
) *QuestNode {
	if len(nodes) == 0 || currentNodeID == uuid.Nil {
		return nil
	}
	graph := buildQuestNodeGraph(nodes)
	return graph.transitionNode(currentNodeID, outcome)
}
