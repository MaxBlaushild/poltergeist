package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

// GenerateCharacterTagsProcessor reads a character's full content (bio,
// secrets, missions) via the LLM oracle and proposes personality/trait
// tags for the Invites tab's filter/search picker. One job per character,
// same "enqueue → poll status → done" shape as GradeQuizSubmissionProcessor.
type GenerateCharacterTagsProcessor struct {
	dbClient   db.DbClient
	deepPriest deep_priest.DeepPriest
}

func NewGenerateCharacterTagsProcessor(dbClient db.DbClient, deepPriest deep_priest.DeepPriest) GenerateCharacterTagsProcessor {
	return GenerateCharacterTagsProcessor{dbClient: dbClient, deepPriest: deepPriest}
}

const maxCharacterTags = 8

func (p *GenerateCharacterTagsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.GenerateCharacterTagsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal generate_character_tags payload: %w", err)
	}

	v := p.dbClient.Vampire()
	_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusGenerating, "")

	character, err := v.GetCharacterByID(ctx, payload.CharacterID)
	if err != nil {
		_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusFailed, err.Error())
		return fmt.Errorf("failed to load character %s: %w", payload.CharacterID, err)
	}
	if character == nil {
		_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusFailed, "character not found")
		return fmt.Errorf("character %s not found", payload.CharacterID)
	}

	prompt := buildCharacterTagsPrompt(character)
	ans, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusFailed, err.Error())
		return fmt.Errorf("deep priest tag generation failed for character %s: %w", payload.CharacterID, err)
	}
	tags := parseCharacterTags(ans.Answer)
	if len(tags) == 0 {
		_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusFailed, "the oracle returned no usable tags")
		return fmt.Errorf("no tags parsed from oracle response for character %s: %q", payload.CharacterID, ans.Answer)
	}

	if err := v.UpdateCharacter(ctx, payload.CharacterID, map[string]interface{}{
		"tags": models.StringArray(tags),
	}); err != nil {
		_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusFailed, err.Error())
		return fmt.Errorf("failed to save generated tags for character %s: %w", payload.CharacterID, err)
	}

	_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusGenerated, "")
	log.Printf("generated %d tags for character %s: %s", len(tags), payload.CharacterID, strings.Join(tags, ", "))
	return nil
}

// buildCharacterTagsPrompt packs everything the oracle needs to judge a
// character's personality from their in-world content — bio, secrets,
// missions — into one instruction.
func buildCharacterTagsPrompt(c *models.VampireCharacter) string {
	var b strings.Builder
	b.WriteString("You are reading a character dossier for a live murder-mystery party game. Propose short personality/trait tags for this character — the kind a host would use to quickly recall their vibe when picking who to invite whom to play as (examples: \"gambler\", \"musical\", \"aggressive\", \"risk taker\", \"loyal\", \"manipulative\", \"secretive\").\n\n")
	fmt.Fprintf(&b, "NAME: %s\n", c.Name)
	if c.Title != "" {
		fmt.Fprintf(&b, "TITLE: %s\n", c.Title)
	}
	if c.House != nil {
		fmt.Fprintf(&b, "HOUSE: %s\n", c.House.Name)
	}
	if c.PreEventInfo != "" {
		fmt.Fprintf(&b, "\nPRE-EVENT BIO:\n%s\n", c.PreEventInfo)
	}
	if c.PostAct1Context != "" {
		fmt.Fprintf(&b, "\nPOST-ACT BIO:\n%s\n", c.PostAct1Context)
	}
	if len(c.Secrets) > 0 {
		b.WriteString("\nSECRETS:\n")
		for _, s := range c.Secrets {
			fmt.Fprintf(&b, "- %s\n", s.Body)
		}
	}
	if len(c.Missions) > 0 {
		b.WriteString("\nMISSIONS:\n")
		for _, m := range c.Missions {
			fmt.Fprintf(&b, "- %s\n", m.Prompt)
		}
	}
	fmt.Fprintf(&b, "\nReply with EXACTLY a comma-separated list of %d-%d tags and nothing else — no numbering, no explanation. Each tag lowercase, one to three words, no trailing punctuation.\nExample: gambler, risk taker, secretive, loyal to house\n", 4, maxCharacterTags)
	return b.String()
}

// parseCharacterTags turns the oracle's comma-separated reply into a
// clean, deduped, capped tag list. Falls back to splitting on newlines if
// the model ignored the comma-separated instruction.
func parseCharacterTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	sep := ","
	if !strings.Contains(raw, ",") && strings.Contains(raw, "\n") {
		sep = "\n"
	}
	parts := strings.Split(raw, sep)

	seen := make(map[string]bool, len(parts))
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.ToLower(strings.Trim(strings.TrimSpace(p), ".- \t\"'•"))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		tags = append(tags, t)
		if len(tags) >= maxCharacterTags {
			break
		}
	}
	return tags
}
