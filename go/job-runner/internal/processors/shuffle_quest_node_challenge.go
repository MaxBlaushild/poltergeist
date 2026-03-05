package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

type ShuffleQuestNodeChallengeProcessor struct {
	dbClient   db.DbClient
	deepPriest deep_priest.DeepPriest
}

type shuffledQuestChallengeResponse struct {
	Question   string `json:"question"`
	Difficulty int    `json:"difficulty"`
}

const shuffleQuestNodeChallengePromptTemplate = `
You are a quest designer creating a replacement challenge for an existing real-world quest node.

Ignore any fantasy flavor. Base the challenge only on the real-world location and activity.

Zone: %s
Quest name: %s
Quest description: %s
Quest giver: %s
Quest giver description: %s

Point of Interest:
%s

Current challenge to replace:
%s

Create ONE new replacement challenge that is materially different from the current challenge.
Constraints:
- Safe, legal, and respectful. Do not require restricted areas or staff interaction.
- Single-input only: EITHER photo proof OR short text response (1-2 sentences), never both.
- Require meaningful participation in the POI's core activity (not just approaching it).
- Avoid hard-to-verify prompts. Prefer proof-of-participation in the activity itself.
- Do NOT use signage-only prompts (storefront sign, menu board, entrance, marquee, poster, or facade) as the main proof.
- If the POI is food/drink-focused, the challenge should involve getting a drink/food item and proving that selected item.
- Must be doable on-site by one player without external research.
- 1-2 short sentences.
- Difficulty must be 25-50 (inclusive).

Respond ONLY as JSON:
{
  "question": "string",
  "difficulty": 32
}
`

func NewShuffleQuestNodeChallengeProcessor(
	dbClient db.DbClient,
	deepPriest deep_priest.DeepPriest,
) ShuffleQuestNodeChallengeProcessor {
	log.Println("Initializing ShuffleQuestNodeChallengeProcessor")
	return ShuffleQuestNodeChallengeProcessor{
		dbClient:   dbClient,
		deepPriest: deepPriest,
	}
}

func (p *ShuffleQuestNodeChallengeProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing shuffle quest node challenge task: %v", task.Type())

	var payload jobs.ShuffleQuestNodeChallengeTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	challenge, err := p.dbClient.QuestNodeChallenge().FindByID(ctx, payload.QuestNodeChallengeID)
	if err != nil {
		return err
	}
	if challenge == nil {
		return nil
	}

	if challenge.ChallengeShuffleStatus == models.QuestNodeChallengeShuffleStatusCompleted {
		// Allow repeat shuffles; this is not terminal.
	}

	if err := p.updateChallengeShuffleStatus(ctx, challenge, models.QuestNodeChallengeShuffleStatusInProgress, nil); err != nil {
		return err
	}

	if err := p.shuffleChallenge(ctx, challenge); err != nil {
		errMsg := err.Error()
		_ = p.updateChallengeShuffleStatus(ctx, challenge, models.QuestNodeChallengeShuffleStatusFailed, &errMsg)
		return err
	}

	return nil
}

func (p *ShuffleQuestNodeChallengeProcessor) shuffleChallenge(ctx context.Context, challenge *models.QuestNodeChallenge) error {
	node, err := p.dbClient.QuestNode().FindByID(ctx, challenge.QuestNodeID)
	if err != nil {
		return fmt.Errorf("failed to find quest node: %w", err)
	}
	if node == nil {
		return fmt.Errorf("quest node not found")
	}

	quest, err := p.dbClient.Quest().FindByID(ctx, node.QuestID)
	if err != nil {
		return fmt.Errorf("failed to find quest: %w", err)
	}
	if quest == nil {
		return fmt.Errorf("quest not found")
	}

	var zone *models.Zone
	if quest.ZoneID != nil {
		zone, err = p.dbClient.Zone().FindByID(ctx, *quest.ZoneID)
		if err != nil {
			return fmt.Errorf("failed to find zone: %w", err)
		}
	}

	var questGiver *models.Character
	if quest.QuestGiverCharacterID != nil {
		questGiver, err = p.dbClient.Character().FindByID(ctx, *quest.QuestGiverCharacterID)
		if err != nil {
			return fmt.Errorf("failed to find quest giver: %w", err)
		}
	}

	locationDetails := "No location details available."
	fallbackLocationName := "this location"
	switch {
	case node.ChallengeID != nil:
		standaloneChallenge, err := p.dbClient.Challenge().FindByID(ctx, *node.ChallengeID)
		if err != nil {
			return fmt.Errorf("failed to find standalone challenge: %w", err)
		}
		if standaloneChallenge != nil {
			if strings.TrimSpace(standaloneChallenge.Description) != "" {
				locationDetails = fmt.Sprintf(
					"Name: Quest Challenge\nDescription: %s\nQuestion: %s",
					truncate(strings.TrimSpace(standaloneChallenge.Description), 220),
					truncate(strings.TrimSpace(standaloneChallenge.Question), 220),
				)
			} else {
				locationDetails = fmt.Sprintf("Name: Quest Challenge\nQuestion: %s", truncate(strings.TrimSpace(standaloneChallenge.Question), 220))
			}
		}
	case node.ScenarioID != nil:
		scenario, err := p.dbClient.Scenario().FindByID(ctx, *node.ScenarioID)
		if err != nil {
			return fmt.Errorf("failed to find scenario: %w", err)
		}
		if scenario != nil {
			fallbackLocationName = "scenario location"
			locationDetails = fmt.Sprintf(
				"Name: %s\nPrompt: %s",
				truncate(fallbackLocationName, 120),
				truncate(strings.TrimSpace(scenario.Prompt), 220),
			)
		}
	case node.MonsterID != nil:
		monster, err := p.dbClient.Monster().FindByID(ctx, *node.MonsterID)
		if err != nil {
			return fmt.Errorf("failed to find monster: %w", err)
		}
		if monster != nil {
			fallbackLocationName = strings.TrimSpace(monster.Name)
			if fallbackLocationName == "" {
				fallbackLocationName = "monster location"
			}
			locationDetails = fmt.Sprintf(
				"Name: %s\nDescription: %s",
				truncate(fallbackLocationName, 120),
				truncate(strings.TrimSpace(monster.Description), 220),
			)
		}
	}

	question, difficulty := p.generateReplacementChallenge(ctx, zone, quest, questGiver, locationDetails, fallbackLocationName, challenge.Question)
	question, submissionType := normalizeAppliedChallengeQuestion(question, nil, nil)
	if difficulty <= 0 {
		difficulty = randomQuestDifficulty()
	}
	difficulty = clampQuestDifficulty(difficulty)

	statTags := p.classifyShuffledChallengeStatTags(ctx, quest, question)
	if statTags == nil {
		statTags = models.StringArray{}
	}

	updates := copyQuestNodeChallengeForUpdate(challenge)
	updates.Question = question
	updates.Difficulty = difficulty
	updates.SubmissionType = submissionType
	updates.StatTags = statTags
	updates.ChallengeShuffleStatus = models.QuestNodeChallengeShuffleStatusCompleted
	updates.ChallengeShuffleError = nil
	updates.UpdatedAt = time.Now()

	updated, err := p.dbClient.QuestNodeChallenge().Update(ctx, challenge.ID, &updates)
	if err != nil {
		return fmt.Errorf("failed to update shuffled challenge: %w", err)
	}
	if updated != nil {
		*challenge = *updated
	}
	return nil
}

func (p *ShuffleQuestNodeChallengeProcessor) generateReplacementChallenge(
	ctx context.Context,
	zone *models.Zone,
	quest *models.Quest,
	questGiver *models.Character,
	locationDetails string,
	fallbackLocationName string,
	currentQuestion string,
) (string, int) {
	zoneName := ""
	if zone != nil {
		zoneName = zone.Name
	}
	questGiverName := ""
	questGiverDescription := ""
	if questGiver != nil {
		questGiverName = questGiver.Name
		questGiverDescription = questGiver.Description
	}
	prompt := fmt.Sprintf(
		shuffleQuestNodeChallengePromptTemplate,
		truncate(zoneName, 120),
		truncate(strings.TrimSpace(quest.Name), 120),
		truncate(strings.TrimSpace(quest.Description), 400),
		truncate(questGiverName, 80),
		truncate(questGiverDescription, 220),
		locationDetails,
		truncate(strings.TrimSpace(currentQuestion), 240),
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return buildHeuristicChallengeQuestion(fallbackLocationName, nil), randomQuestDifficulty()
	}

	var response shuffledQuestChallengeResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return buildHeuristicChallengeQuestion(fallbackLocationName, nil), randomQuestDifficulty()
	}

	question := strings.TrimSpace(response.Question)
	if question == "" {
		question = buildHeuristicChallengeQuestion(fallbackLocationName, nil)
	}
	return question, response.Difficulty
}

func (p *ShuffleQuestNodeChallengeProcessor) classifyShuffledChallengeStatTags(
	ctx context.Context,
	quest *models.Quest,
	question string,
) models.StringArray {
	if quest == nil {
		return models.StringArray{}
	}
	proxy := ApplyZoneSeedDraftProcessor{deepPriest: p.deepPriest}
	draft := models.ZoneSeedQuestDraft{
		Name:               strings.TrimSpace(quest.Name),
		Description:        strings.TrimSpace(quest.Description),
		ChallengeQuestion:  strings.TrimSpace(question),
		AcceptanceDialogue: []string(quest.AcceptanceDialogue),
	}
	return proxy.classifyQuestStatTags(ctx, draft)
}

func (p *ShuffleQuestNodeChallengeProcessor) updateChallengeShuffleStatus(
	ctx context.Context,
	challenge *models.QuestNodeChallenge,
	status string,
	errMsg *string,
) error {
	updates := copyQuestNodeChallengeForUpdate(challenge)
	updates.ChallengeShuffleStatus = status
	updates.ChallengeShuffleError = errMsg
	updates.UpdatedAt = time.Now()

	updated, err := p.dbClient.QuestNodeChallenge().Update(ctx, challenge.ID, &updates)
	if err != nil {
		return err
	}
	if updated != nil {
		*challenge = *updated
	}
	return nil
}

func copyQuestNodeChallengeForUpdate(challenge *models.QuestNodeChallenge) models.QuestNodeChallenge {
	if challenge == nil {
		return models.QuestNodeChallenge{
			StatTags: models.StringArray{},
		}
	}
	statTags := challenge.StatTags
	if statTags == nil {
		statTags = models.StringArray{}
	}
	return models.QuestNodeChallenge{
		Tier:                   challenge.Tier,
		Question:               challenge.Question,
		Reward:                 challenge.Reward,
		InventoryItemID:        challenge.InventoryItemID,
		SubmissionType:         challenge.SubmissionType,
		Difficulty:             challenge.Difficulty,
		StatTags:               statTags,
		Proficiency:            challenge.Proficiency,
		ChallengeShuffleStatus: challenge.ChallengeShuffleStatus,
		ChallengeShuffleError:  challenge.ChallengeShuffleError,
		UpdatedAt:              challenge.UpdatedAt,
	}
}
