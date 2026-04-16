package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type scenarioQuestHandoff struct {
	QuestID                  uuid.UUID  `json:"questId"`
	QuestName                string     `json:"questName"`
	ResolvedNodeID           uuid.UUID  `json:"resolvedNodeId"`
	NextNodeID               *uuid.UUID `json:"nextNodeId,omitempty"`
	NextObjectiveText        string     `json:"nextObjectiveText,omitempty"`
	NextObjectiveDescription string     `json:"nextObjectiveDescription,omitempty"`
	NextPointOfInterestName  string     `json:"nextPointOfInterestName,omitempty"`
	NextCharacterName        string     `json:"nextCharacterName,omitempty"`
	HandoffText              string     `json:"handoffText,omitempty"`
}

type scenarioQuestNodeSummary struct {
	objectiveText        string
	objectiveDescription string
	pointOfInterestName  string
	characterName        string
}

func sanitizeScenarioHandoffText(input string, fallback string) string {
	text := strings.TrimSpace(input)
	if text == "" {
		text = strings.TrimSpace(fallback)
	}
	if len(text) > 320 {
		text = strings.TrimSpace(text[:320])
	}
	return text
}

func (s *server) buildScenarioFreeformQuestContext(
	ctx context.Context,
	target questNodeCompletionTarget,
) string {
	currentSummary := s.buildScenarioQuestNodeSummary(ctx, target.Node)
	successNode := previewScenarioQuestNodeForOutcome(target.Quest, target.Node, true)
	failureNode := previewScenarioQuestNodeForOutcome(target.Quest, target.Node, false)
	successSummary := s.buildScenarioQuestNodeSummary(ctx, successNode)
	failureSummary := s.buildScenarioQuestNodeSummary(ctx, failureNode)

	currentObjective := currentSummary.objectiveText
	if currentObjective == "" {
		currentObjective = "Continue the quest"
	}

	questName := strings.TrimSpace(target.Quest.Name)
	if questName == "" {
		questName = "Unnamed quest"
	}

	lines := []string{
		fmt.Sprintf("Quest: %s", questName),
		fmt.Sprintf("Current objective: %s", currentObjective),
		fmt.Sprintf("If the response succeeds: %s", describeScenarioQuestPreview(target.Quest, target.Node, successNode, successSummary, true)),
		fmt.Sprintf("If the response fails: %s", describeScenarioQuestPreview(target.Quest, target.Node, failureNode, failureSummary, false)),
		"Use the quest context when writing handoff text so the result clearly points toward the next story beat.",
	}
	return strings.Join(lines, "\n")
}

func (s *server) buildScenarioQuestHandoffs(
	ctx context.Context,
	targets []questNodeCompletionTarget,
	success bool,
	preferredText string,
) []scenarioQuestHandoff {
	if len(targets) == 0 {
		return nil
	}

	trimmedPreferredText := strings.TrimSpace(preferredText)
	if len(targets) > 1 {
		trimmedPreferredText = ""
	}

	handoffs := make([]scenarioQuestHandoff, 0, len(targets))
	for _, target := range targets {
		if target.Quest == nil || target.Node == nil {
			continue
		}

		nextNode := previewScenarioQuestNodeForOutcome(target.Quest, target.Node, success)
		if target.Acceptance != nil {
			if resolvedNode, err := s.currentQuestNode(ctx, target.Quest, target.Acceptance); err != nil {
				log.Printf(
					"buildScenarioQuestHandoffs: failed to refresh current quest node quest=%s acceptance=%s err=%v",
					target.Quest.ID,
					target.Acceptance.ID,
					err,
				)
			} else {
				nextNode = resolvedNode
			}
		}

		nextSummary := s.buildScenarioQuestNodeSummary(ctx, nextNode)
		fallback := buildScenarioQuestHandoffFallback(
			target.Quest,
			target.Acceptance,
			target.Node,
			nextNode,
			nextSummary,
			success,
		)

		questName := strings.TrimSpace(target.Quest.Name)
		if questName == "" {
			questName = "Quest"
		}

		nextNodeID := questNodeIDPointer(nextNode)
		handoffs = append(handoffs, scenarioQuestHandoff{
			QuestID:                  target.Quest.ID,
			QuestName:                questName,
			ResolvedNodeID:           target.Node.ID,
			NextNodeID:               nextNodeID,
			NextObjectiveText:        nextSummary.objectiveText,
			NextObjectiveDescription: nextSummary.objectiveDescription,
			NextPointOfInterestName:  nextSummary.pointOfInterestName,
			NextCharacterName:        nextSummary.characterName,
			HandoffText:              sanitizeScenarioHandoffText(trimmedPreferredText, fallback),
		})
	}

	return handoffs
}

func previewScenarioQuestNodeForOutcome(
	quest *models.Quest,
	node *models.QuestNode,
	success bool,
) *models.QuestNode {
	if quest == nil || node == nil {
		return nil
	}
	if success {
		return models.ResolveNextQuestNode(
			quest.Nodes,
			node.ID,
			models.QuestNodeTransitionOutcomeSuccess,
		)
	}
	if node.FailurePolicyNormalized() == models.QuestNodeFailurePolicyTransition {
		nextNode := models.ResolveNextQuestNode(
			quest.Nodes,
			node.ID,
			models.QuestNodeTransitionOutcomeFailure,
		)
		if nextNode != nil {
			return nextNode
		}
	}
	return node
}

func describeScenarioQuestPreview(
	quest *models.Quest,
	currentNode *models.QuestNode,
	nextNode *models.QuestNode,
	nextSummary scenarioQuestNodeSummary,
	success bool,
) string {
	switch {
	case nextNode == nil:
		questName := "the quest"
		if quest != nil && strings.TrimSpace(quest.Name) != "" {
			questName = strings.TrimSpace(quest.Name)
		}
		return fmt.Sprintf("The current thread resolves and %s is ready to close.", questName)
	case !success && currentNode != nil && nextNode.ID == currentNode.ID:
		if nextSummary.objectiveText != "" {
			return "The same objective remains active: " + nextSummary.objectiveText
		}
		return "The same quest objective remains active."
	case nextSummary.objectiveText != "":
		return nextSummary.objectiveText
	case nextSummary.characterName != "":
		return "Follow up with " + nextSummary.characterName
	case nextSummary.pointOfInterestName != "":
		return "Travel to " + nextSummary.pointOfInterestName
	case !success:
		return "The story continues through the failure consequences."
	default:
		return "The story continues to the next quest beat."
	}
}

func buildScenarioQuestHandoffFallback(
	quest *models.Quest,
	acceptance *models.QuestAcceptanceV2,
	currentNode *models.QuestNode,
	nextNode *models.QuestNode,
	nextSummary scenarioQuestNodeSummary,
	success bool,
) string {
	questName := "the quest"
	if quest != nil && strings.TrimSpace(quest.Name) != "" {
		questName = strings.TrimSpace(quest.Name)
	}

	switch {
	case nextNode == nil:
		return fmt.Sprintf(
			"That resolves the last open thread in %s. The quest is ready to close.",
			questName,
		)
	case !success && currentNode != nil && nextNode.ID == currentNode.ID:
		if nextSummary.objectiveText != "" {
			return "The problem is not settled yet. You still need to " + trimScenarioPreviewText(nextSummary.objectiveText) + "."
		}
		return "The problem is not settled yet. The same objective is still active."
	case !success && nextSummary.objectiveText != "":
		return "Even in failure, the story keeps moving: " + trimScenarioPreviewText(nextSummary.objectiveText) + "."
	case success && nextSummary.objectiveText != "":
		return "That opens the next lead: " + trimScenarioPreviewText(nextSummary.objectiveText) + "."
	case nextSummary.characterName != "":
		return "That points you toward " + nextSummary.characterName + "."
	case nextSummary.pointOfInterestName != "":
		return "That points you toward " + nextSummary.pointOfInterestName + "."
	case acceptance != nil && (acceptance.IsClosed() || acceptance.ObjectivesCompletedAt != nil):
		return fmt.Sprintf(
			"That resolves the last open thread in %s. The quest is ready to close.",
			questName,
		)
	case !success:
		return "The attempt falls short, but the quest thread continues from here."
	default:
		return "The quest thread continues from here."
	}
}

func trimScenarioPreviewText(input string) string {
	return strings.TrimRight(strings.TrimSpace(input), ".!?")
}

func (s *server) buildScenarioQuestNodeSummary(
	ctx context.Context,
	node *models.QuestNode,
) scenarioQuestNodeSummary {
	if node == nil {
		return scenarioQuestNodeSummary{}
	}

	summary := scenarioQuestNodeSummary{
		objectiveText:        strings.TrimSpace(node.ObjectiveDescription),
		objectiveDescription: strings.TrimSpace(node.ObjectiveDescription),
	}

	if node.IsFetchQuestNode() {
		character := node.FetchCharacter
		if character == nil && node.FetchCharacterID != nil && *node.FetchCharacterID != uuid.Nil {
			loadedCharacter, err := s.dbClient.Character().FindByID(ctx, *node.FetchCharacterID)
			if err != nil {
				log.Printf("buildScenarioQuestNodeSummary: fetch character lookup failed %s: %v", node.FetchCharacterID.String(), err)
			} else {
				character = loadedCharacter
			}
		}
		if character != nil {
			summary.characterName = strings.TrimSpace(character.Name)
			if character.PointOfInterest != nil {
				summary.pointOfInterestName = strings.TrimSpace(character.PointOfInterest.Name)
			}
		}
		if summary.objectiveText == "" {
			if summary.characterName != "" {
				summary.objectiveText = "Bring the required items to " + summary.characterName
			} else {
				summary.objectiveText = "Deliver the required items"
			}
		}
		if summary.objectiveDescription == "" {
			summary.objectiveDescription = "Hand over the requested items to continue the quest."
		}
		return summary
	}

	if node.ScenarioID != nil && *node.ScenarioID != uuid.Nil {
		scenario, err := s.dbClient.Scenario().FindByID(ctx, *node.ScenarioID)
		if err != nil {
			log.Printf("buildScenarioQuestNodeSummary: scenario lookup failed %s: %v", node.ScenarioID.String(), err)
		} else if scenario != nil {
			if summary.objectiveText == "" {
				summary.objectiveText = strings.TrimSpace(scenario.Prompt)
			}
			if summary.objectiveDescription == "" {
				summary.objectiveDescription = strings.TrimSpace(scenario.Prompt)
			}
			if scenario.PointOfInterest != nil {
				summary.pointOfInterestName = strings.TrimSpace(scenario.PointOfInterest.Name)
			}
		}
	}

	if node.ChallengeID != nil && *node.ChallengeID != uuid.Nil {
		challenge, err := s.dbClient.Challenge().FindByID(ctx, *node.ChallengeID)
		if err != nil {
			log.Printf("buildScenarioQuestNodeSummary: challenge lookup failed %s: %v", node.ChallengeID.String(), err)
		} else if challenge != nil {
			if summary.objectiveText == "" {
				summary.objectiveText = strings.TrimSpace(challenge.Question)
			}
			if summary.objectiveDescription == "" {
				summary.objectiveDescription = strings.TrimSpace(challenge.Description)
			}
			if challenge.PointOfInterest != nil {
				summary.pointOfInterestName = strings.TrimSpace(challenge.PointOfInterest.Name)
			}
		}
	}

	if node.ExpositionID != nil && *node.ExpositionID != uuid.Nil {
		exposition, err := s.dbClient.Exposition().FindByID(ctx, *node.ExpositionID)
		if err != nil {
			log.Printf("buildScenarioQuestNodeSummary: exposition lookup failed %s: %v", node.ExpositionID.String(), err)
		} else if exposition != nil {
			title := strings.TrimSpace(exposition.Title)
			if summary.objectiveText == "" {
				if title != "" {
					summary.objectiveText = "Complete the dialogue: " + title
				} else {
					summary.objectiveText = "Complete the dialogue"
				}
			}
			if summary.objectiveDescription == "" {
				summary.objectiveDescription = strings.TrimSpace(exposition.Description)
			}
			if exposition.PointOfInterest != nil {
				summary.pointOfInterestName = strings.TrimSpace(exposition.PointOfInterest.Name)
			}
		}
	}

	if node.MonsterEncounterID != nil && *node.MonsterEncounterID != uuid.Nil {
		encounter, err := s.dbClient.MonsterEncounter().FindByID(ctx, *node.MonsterEncounterID)
		if err != nil {
			log.Printf("buildScenarioQuestNodeSummary: monster encounter lookup failed %s: %v", node.MonsterEncounterID.String(), err)
		} else if encounter != nil {
			name := strings.TrimSpace(encounter.Name)
			if summary.objectiveText == "" {
				switch models.NormalizeMonsterEncounterType(string(encounter.EncounterType)) {
				case models.MonsterEncounterTypeBoss:
					if name != "" {
						summary.objectiveText = "Defeat the boss: " + name
					} else {
						summary.objectiveText = "Defeat the boss encounter"
					}
				case models.MonsterEncounterTypeRaid:
					if name != "" {
						summary.objectiveText = "Complete the raid encounter: " + name
					} else {
						summary.objectiveText = "Complete the raid encounter"
					}
				default:
					if name != "" {
						summary.objectiveText = "Defeat " + name
					} else {
						summary.objectiveText = "Defeat the monster encounter"
					}
				}
			}
			if summary.objectiveDescription == "" {
				summary.objectiveDescription = strings.TrimSpace(encounter.Description)
			}
			if encounter.PointOfInterest != nil {
				summary.pointOfInterestName = strings.TrimSpace(encounter.PointOfInterest.Name)
			}
		}
	}

	if node.MonsterID != nil && *node.MonsterID != uuid.Nil {
		monster, err := s.dbClient.Monster().FindByID(ctx, *node.MonsterID)
		if err != nil {
			log.Printf("buildScenarioQuestNodeSummary: monster lookup failed %s: %v", node.MonsterID.String(), err)
		} else if monster != nil {
			name := strings.TrimSpace(monster.Name)
			if summary.objectiveText == "" {
				if name != "" {
					summary.objectiveText = "Defeat " + name
				} else {
					summary.objectiveText = "Defeat the monster"
				}
			}
			if summary.objectiveDescription == "" {
				summary.objectiveDescription = strings.TrimSpace(monster.Description)
			}
		}
	}

	if summary.objectiveText == "" {
		summary.objectiveText = "Continue the quest"
	}
	if summary.objectiveDescription == "" {
		summary.objectiveDescription = summary.objectiveText
	}

	return summary
}
