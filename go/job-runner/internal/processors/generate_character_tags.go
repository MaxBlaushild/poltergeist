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
	tags, err := parseCharacterTagsAnswer(ans.Answer)
	if err != nil {
		_ = v.SetCharacterTagsStatus(ctx, payload.CharacterID, models.CharacterTagsStatusFailed, err.Error())
		return fmt.Errorf("failed to parse oracle response for character %s: %w (raw: %q)", payload.CharacterID, err, ans.Answer)
	}
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
	if c.Bio != "" {
		fmt.Fprintf(&b, "\nBIO:\n%s\n", c.Bio)
	}
	if len(c.PostAct1Contexts) > 0 {
		// Spans every mystery this character's been cast in, same as
		// SECRETS below — not scoped to any one mystery, since tag
		// generation reads the character's full picture.
		b.WriteString("\nPOST-ACT BIO:\n")
		for _, pc := range c.PostAct1Contexts {
			if pc.PostAct1Context != "" {
				fmt.Fprintf(&b, "%s\n", pc.PostAct1Context)
			}
		}
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
	// The Fount forces JSON-object output on every completion regardless of
	// what's asked for (see fount-of-erebos), so — unlike an earlier
	// version of this prompt that asked for a bare comma-separated list and
	// got it back JSON-wrapped anyway (garbling a naive comma-split) — ask
	// for the JSON shape it's already going to produce and parse that.
	fmt.Fprintf(&b, "\nReturn JSON only, shaped exactly like this: {\"tags\": [\"tag one\", \"tag two\", ...]} — %d to %d tags, nothing else. Each tag lowercase, one to three words, no trailing punctuation. Do not output markdown or commentary outside the JSON object.\n", 4, maxCharacterTags)
	return b.String()
}

// characterTagsEnvelope is the {"tags": [...]} shape the prompt asks for.
// Tags is left as json.RawMessage so parseCharacterTagsAnswer can also
// recover from the oracle sending a comma-separated string instead of a
// real array, rather than failing outright.
type characterTagsEnvelope struct {
	Tags json.RawMessage `json:"tags"`
}

// parseCharacterTagsAnswer parses the oracle's JSON reply (the Fount always
// returns JSON — see buildCharacterTagsPrompt) into a clean, deduped,
// capped tag list. extractGeneratedJSONObject strips markdown fences if
// the model added any, matching the pattern every other JSON-mode
// processor in this package uses.
func parseCharacterTagsAnswer(raw string) ([]string, error) {
	var env characterTagsEnvelope
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(raw)), &env); err != nil {
		return nil, fmt.Errorf("could not parse oracle response as JSON: %w", err)
	}

	var asArray []string
	if err := json.Unmarshal(env.Tags, &asArray); err == nil {
		return sanitizeCharacterTags(asArray), nil
	}

	// Defense-in-depth: if the oracle ignored the array instruction and
	// sent "tags" as one comma-separated string instead, salvage it rather
	// than saving that raw string as a single garbage tag.
	var asString string
	if err := json.Unmarshal(env.Tags, &asString); err == nil {
		return sanitizeCharacterTags(strings.Split(asString, ",")), nil
	}

	return nil, fmt.Errorf(`the oracle's "tags" field was neither a list of strings nor a comma-separated string`)
}

// sanitizeCharacterTags cleans, dedupes, and caps the oracle's proposed tags.
func sanitizeCharacterTags(raw []string) []string {
	seen := make(map[string]bool, len(raw))
	tags := make([]string, 0, len(raw))
	for _, p := range raw {
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
