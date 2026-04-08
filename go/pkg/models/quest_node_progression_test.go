package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestResolveCurrentQuestNodeAutoCompletesSatisfiedStoryFlagNodes(t *testing.T) {
	now := time.Now()
	waitNodeID := uuid.New()
	challengeNodeID := uuid.New()

	currentNode, autoCompleted := ResolveCurrentQuestNode(
		[]QuestNode{
			{
				ID:           waitNodeID,
				OrderIndex:   0,
				CreatedAt:    now,
				StoryFlagKey: "intro_complete",
			},
			{
				ID:         challengeNodeID,
				OrderIndex: 1,
				CreatedAt:  now.Add(time.Second),
			},
		},
		map[uuid.UUID]bool{},
		map[string]bool{"intro_complete": true},
	)

	if len(autoCompleted) != 1 || autoCompleted[0] != waitNodeID {
		t.Fatalf("expected story-flag node %s to auto-complete, got %+v", waitNodeID, autoCompleted)
	}
	if currentNode == nil {
		t.Fatalf("expected a current node after auto-completing the story-flag node")
	}
	if currentNode.ID != challengeNodeID {
		t.Fatalf("expected challenge node %s, got %s", challengeNodeID, currentNode.ID)
	}
}

func TestResolveCurrentQuestNodeStopsOnUnsatisfiedStoryFlagNode(t *testing.T) {
	waitNodeID := uuid.New()

	currentNode, autoCompleted := ResolveCurrentQuestNode(
		[]QuestNode{
			{
				ID:           waitNodeID,
				OrderIndex:   0,
				StoryFlagKey: "boss_seen",
			},
		},
		map[uuid.UUID]bool{},
		map[string]bool{"other_flag": true},
	)

	if len(autoCompleted) != 0 {
		t.Fatalf("expected no auto-completed nodes, got %+v", autoCompleted)
	}
	if currentNode == nil {
		t.Fatalf("expected the waiting story-flag node to remain current")
	}
	if currentNode.ID != waitNodeID {
		t.Fatalf("expected story-flag node %s, got %s", waitNodeID, currentNode.ID)
	}
}
