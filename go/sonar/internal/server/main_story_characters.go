package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type mainStoryCharacterProfile struct {
	CharacterKey string   `json:"characterKey"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	InternalTags []string `json:"internalTags"`
}

type mainStoryCharacterProfileResponse struct {
	Characters []mainStoryCharacterProfile `json:"characters"`
}

func collectMainStoryCharacterKeys(
	draft *models.MainStorySuggestionDraft,
) []string {
	if draft == nil {
		return nil
	}
	seen := map[string]struct{}{}
	keys := make([]string, 0)
	addKey := func(raw string) {
		key := strings.TrimSpace(strings.ToLower(raw))
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	for _, key := range draft.CharacterKeys {
		addKey(key)
	}
	for _, beat := range draft.Beats {
		addKey(beat.QuestGiverCharacterKey)
		for _, key := range beat.IntroducedCharacterKeys {
			addKey(key)
		}
		for _, key := range beat.RequiredCharacterKeys {
			addKey(key)
		}
	}
	return keys
}

func buildMainStoryCharacterGenerationPrompt(
	draft *models.MainStorySuggestionDraft,
	keys []string,
) string {
	beatLines := make([]string, 0, len(draft.Beats))
	for _, beat := range draft.Beats {
		beatLines = append(beatLines, fmt.Sprintf(
			"- beat %d %q: role=%s, quest giver key=%s, character tags=%s, summary=%s",
			beat.OrderIndex,
			strings.TrimSpace(beat.Name),
			strings.TrimSpace(beat.StoryRole),
			strings.TrimSpace(beat.QuestGiverCharacterKey),
			strings.Join([]string(beat.CharacterTags), ", "),
			strings.TrimSpace(beat.ChapterSummary),
		))
	}
	return fmt.Sprintf(`
You are designing recurring NPCs for a district-scale urban fantasy main story in Unclaimed Streets.

Create one concrete, reusable NPC for each required story character key.
These characters should feel like they belong to the same story, district, and tone as the beats below.
They should be vivid, human, and playable as recurring questgivers, not abstract placeholders.

Story:
- name: %s
- premise: %s
- district fit: %s
- tone: %s
- theme tags: %s
- faction keys: %s
- character keys: %s

Beat context:
%s

Return JSON only:
{
  "characters": [
    {
      "characterKey": "one of the requested keys",
      "name": "full character name",
      "description": "2-4 sentence portrait of who they are, how they present, and what role they play in the story",
      "internalTags": ["lowercase_snake_case_tag", "lowercase_snake_case_tag"]
    }
  ]
}

Rules:
- Output exactly one character for each required key.
- Every characterKey must match one of the requested keys exactly.
- Names should feel distinct, memorable, and human.
- Descriptions should be grounded urban fantasy one-paragraph character portraits, not backstory dumps.
- internalTags should be sparse and useful for future linking: include role, vibe, faction, and social position tags where appropriate.
- Do not include markdown.
`,
		strings.TrimSpace(draft.Name),
		strings.TrimSpace(draft.Premise),
		strings.TrimSpace(draft.DistrictFit),
		strings.TrimSpace(draft.Tone),
		strings.Join([]string(draft.ThemeTags), ", "),
		strings.Join([]string(draft.FactionKeys), ", "),
		strings.Join(keys, ", "),
		strings.Join(beatLines, "\n"),
	)
}

func humanizeStoryCharacterKey(key string) string {
	parts := strings.Fields(strings.ReplaceAll(strings.TrimSpace(key), "_", " "))
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func fallbackMainStoryCharacterProfile(
	draft *models.MainStorySuggestionDraft,
	key string,
) mainStoryCharacterProfile {
	title := humanizeStoryCharacterKey(key)
	premise := strings.TrimSpace(draft.Premise)
	description := fmt.Sprintf(
		"%s is a recurring figure in this main story, woven into the district's tensions and asked to carry the thread associated with %s.",
		title,
		title,
	)
	if premise != "" {
		description = fmt.Sprintf(
			"%s is a recurring figure in %s, carrying the story thread associated with %s while feeling grounded in the district's daily life.",
			title,
			premise,
			title,
		)
	}
	return mainStoryCharacterProfile{
		CharacterKey: key,
		Name:         title,
		Description:  description,
		InternalTags: []string{key},
	}
}

func normalizeMainStoryCharacterProfiles(
	draft *models.MainStorySuggestionDraft,
	keys []string,
	raw []mainStoryCharacterProfile,
) []mainStoryCharacterProfile {
	normalized := make([]mainStoryCharacterProfile, 0, len(keys))
	byKey := map[string]mainStoryCharacterProfile{}
	for _, profile := range raw {
		key := strings.TrimSpace(strings.ToLower(profile.CharacterKey))
		if key == "" {
			continue
		}
		byKey[key] = mainStoryCharacterProfile{
			CharacterKey: key,
			Name:         strings.TrimSpace(profile.Name),
			Description:  strings.TrimSpace(profile.Description),
			InternalTags: parseCharacterInternalTags(profile.InternalTags),
		}
	}
	for _, key := range keys {
		profile, ok := byKey[key]
		if !ok {
			profile = fallbackMainStoryCharacterProfile(draft, key)
		}
		if strings.TrimSpace(profile.Name) == "" {
			profile.Name = humanizeStoryCharacterKey(key)
		}
		if strings.TrimSpace(profile.Description) == "" {
			profile.Description = fallbackMainStoryCharacterProfile(draft, key).Description
		}
		profile.CharacterKey = key
		profile.InternalTags = parseCharacterInternalTags(append([]string{key}, profile.InternalTags...))
		normalized = append(normalized, profile)
	}
	return normalized
}

func (s *server) generateMainStoryCharacterProfiles(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
) ([]mainStoryCharacterProfile, error) {
	keys := collectMainStoryCharacterKeys(draft)
	if len(keys) == 0 {
		return nil, nil
	}
	if s.deepPriest == nil {
		return normalizeMainStoryCharacterProfiles(draft, keys, nil), nil
	}

	answer, err := s.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: buildMainStoryCharacterGenerationPrompt(draft, keys),
	})
	if err != nil {
		return normalizeMainStoryCharacterProfiles(draft, keys, nil), nil
	}

	var response mainStoryCharacterProfileResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(answer.Answer)), &response); err != nil {
		return normalizeMainStoryCharacterProfiles(draft, keys, nil), nil
	}

	return normalizeMainStoryCharacterProfiles(draft, keys, response.Characters), nil
}

func buildMainStoryCharacterInternalTags(
	draft *models.MainStorySuggestionDraft,
	profile mainStoryCharacterProfile,
) models.StringArray {
	tags := make([]string, 0, len(profile.InternalTags)+len(draft.InternalTags)+4)
	tags = append(tags, profile.InternalTags...)
	tags = append(tags, draft.InternalTags...)
	tags = append(tags, "main_story", "story_character", "story_character_"+profile.CharacterKey)
	return parseCharacterInternalTags(tags)
}

func (s *server) enqueueGeneratedMainStoryCharacterImage(
	ctx context.Context,
	character *models.Character,
) {
	if character == nil || s.asyncClient == nil {
		return
	}
	payloadBytes, err := json.Marshal(jobs.GenerateCharacterImageTaskPayload{
		CharacterID: character.ID,
		Name:        character.Name,
		Description: character.Description,
	})
	if err != nil {
		errMsg := err.Error()
		_ = s.dbClient.Character().UpdateFields(ctx, character.ID, map[string]interface{}{
			"image_generation_status": models.CharacterImageGenerationStatusFailed,
			"image_generation_error":  &errMsg,
		})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateCharacterImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = s.dbClient.Character().UpdateFields(ctx, character.ID, map[string]interface{}{
			"image_generation_status": models.CharacterImageGenerationStatusFailed,
			"image_generation_error":  &errMsg,
		})
	}
}

func (s *server) createMainStoryCharacters(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
) (map[string]*models.Character, error) {
	profiles, err := s.generateMainStoryCharacterProfiles(ctx, draft)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return map[string]*models.Character{}, nil
	}

	createdByKey := make(map[string]*models.Character, len(profiles))
	for _, profile := range profiles {
		key := strings.TrimSpace(strings.ToLower(profile.CharacterKey))
		if key == "" {
			continue
		}
		character := &models.Character{
			ID:                    uuid.New(),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			Name:                  strings.TrimSpace(profile.Name),
			Description:           strings.TrimSpace(profile.Description),
			InternalTags:          buildMainStoryCharacterInternalTags(draft, profile),
			ImageGenerationStatus: models.CharacterImageGenerationStatusQueued,
		}
		if s.asyncClient == nil {
			character.ImageGenerationStatus = models.CharacterImageGenerationStatusNone
		}
		if err := s.dbClient.Character().Create(ctx, character); err != nil {
			return nil, err
		}
		s.enqueueGeneratedMainStoryCharacterImage(ctx, character)
		createdByKey[key] = character
	}
	return createdByKey, nil
}
