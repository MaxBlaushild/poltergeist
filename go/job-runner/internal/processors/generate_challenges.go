package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const challengeGenerationPromptTemplate = `
You are designing %d location-based fantasy RPG map challenges for one zone.

Zone:
- name: %s
- description: %s

Recent scenarios in this zone (match tone/theme, but do not copy):
%s

Recent challenges in this zone to avoid echoing:
%s

Return JSON only:
{
  "challenges": [
    {
      "question": "2-4 sentences asking player to roleplay a concrete action in the real world",
      "description": "40-140 words of visual flavor and atmosphere for later image generation",
      "submissionType": "photo or text",
      "difficulty": 0-40,
      "reward": 0-100,
      "statTags": ["0-3 of: strength,dexterity,constitution,intelligence,wisdom,charisma"],
      "proficiency": "optional short phrase or null"
    }
  ]
}

Hard rules:
- Output exactly %d challenges.
- Challenges must feel similar in fantasy tone to scenarios in this zone.
- Each challenge must require ROLEPLAYING an action in the environment.
- Proof must be either:
  - photo: player submits a photo of what they did
  - text: player submits a short written report of what they did
- Use only submissionType values: "photo" or "text".
- Keep content safe for public spaces and legal behavior.
- Avoid minors, explicit content, illegal activity, harassment, trespassing, or dangerous stunts.
- Make each challenge materially distinct from the others and from recent challenges.
`

var challengeGenerationValidStatTags = map[string]struct{}{
	"strength":     {},
	"dexterity":    {},
	"constitution": {},
	"intelligence": {},
	"wisdom":       {},
	"charisma":     {},
}

var challengeGenerationDefaultStatTags = []string{
	"strength",
	"dexterity",
	"constitution",
	"intelligence",
	"wisdom",
	"charisma",
}

var challengeGenerationFallbackActions = []string{
	"quietly inspect the scene and act as a vigilant ranger",
	"deliver a ceremonial message as if you were a town envoy",
	"reconstruct what happened as a field investigator",
	"perform a brief warding ritual to steady local spirits",
	"act as a guild scout and verify safe passage",
	"play the role of a chronicler documenting a strange event",
}

type generatedChallengesResponse struct {
	Challenges []generatedChallengePayload `json:"challenges"`
}

type generatedChallengePayload struct {
	Question       string   `json:"question"`
	Description    string   `json:"description"`
	SubmissionType string   `json:"submissionType"`
	Difficulty     *int     `json:"difficulty"`
	Reward         *int     `json:"reward"`
	StatTags       []string `json:"statTags"`
	Proficiency    *string  `json:"proficiency"`
}

type GenerateChallengesProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	asyncClient      *asynq.Client
}

func NewGenerateChallengesProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	asyncClient *asynq.Client,
) GenerateChallengesProcessor {
	log.Println("Initializing GenerateChallengesProcessor")
	return GenerateChallengesProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		asyncClient:      asyncClient,
	}
}

func (p *GenerateChallengesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate challenges task: %v", task.Type())

	var payload jobs.GenerateChallengesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ChallengeGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Challenge generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ChallengeGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ChallengeGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateChallenges(ctx, job); err != nil {
		return p.failChallengeGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateChallengesProcessor) generateChallenges(ctx context.Context, job *models.ChallengeGenerationJob) error {
	zone, err := p.dbClient.Zone().FindByID(ctx, job.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to load zone: %w", err)
	}
	if zone == nil {
		return fmt.Errorf("zone not found")
	}

	zoneName := strings.TrimSpace(zone.Name)
	if zoneName == "" {
		zoneName = "Unknown Zone"
	}
	zoneDescription := strings.TrimSpace(zone.Description)
	if zoneDescription == "" {
		zoneDescription = "No description available."
	}

	scenarios, err := p.dbClient.Scenario().FindByZoneID(ctx, job.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to load zone scenarios: %w", err)
	}
	scenarioHints := buildRecentScenarioThemeHints(scenarios, 8)
	challengeAvoidance := p.buildRecentChallengeAvoidance(ctx, job.ZoneID, 12)

	prompt := fmt.Sprintf(
		challengeGenerationPromptTemplate,
		job.Count,
		zoneName,
		zoneDescription,
		scenarioHints,
		challengeAvoidance,
		job.Count,
	)
	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate challenges: %w", err)
	}

	generated := &generatedChallengesResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse generated challenges payload: %w", err)
	}

	challengeSpecs := sanitizeGeneratedChallenges(generated.Challenges, job.Count, zoneName)
	createdCount := 0

	for i := 0; i < len(challengeSpecs); i++ {
		spec := challengeSpecs[i]
		lat, lng := challengeGenerationLocation(*zone, i)
		now := time.Now()

		challenge := &models.Challenge{
			ID:             uuid.New(),
			CreatedAt:      now,
			UpdatedAt:      now,
			ZoneID:         job.ZoneID,
			Latitude:       lat,
			Longitude:      lng,
			Question:       spec.Question,
			Description:    spec.Description,
			Reward:         spec.Reward,
			SubmissionType: spec.SubmissionType,
			Difficulty:     spec.Difficulty,
			StatTags:       spec.StatTags,
			Proficiency:    spec.Proficiency,
		}
		if err := p.dbClient.Challenge().Create(ctx, challenge); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create challenge: %w", err)
		}

		createdCount++
		job.CreatedCount = createdCount
		job.UpdatedAt = time.Now()
		if err := p.dbClient.ChallengeGenerationJob().Update(ctx, job); err != nil {
			return fmt.Errorf("failed to update challenge generation progress: %w", err)
		}

		if p.asyncClient != nil {
			imagePayload, err := json.Marshal(jobs.GenerateChallengeImageTaskPayload{
				ChallengeID: challenge.ID,
			})
			if err == nil {
				if _, enqueueErr := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateChallengeImageTaskType, imagePayload)); enqueueErr != nil {
					log.Printf("Failed to enqueue challenge image generation for challenge %s: %v", challenge.ID, enqueueErr)
				}
			}
		}
	}

	job.Status = models.ChallengeGenerationStatusCompleted
	job.CreatedCount = createdCount
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ChallengeGenerationJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update challenge generation job: %w", err)
	}

	return nil
}

func (p *GenerateChallengesProcessor) failChallengeGenerationJob(ctx context.Context, job *models.ChallengeGenerationJob, err error) error {
	msg := err.Error()
	job.Status = models.ChallengeGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ChallengeGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark challenge generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

type sanitizedGeneratedChallenge struct {
	Question       string
	Description    string
	SubmissionType models.QuestNodeSubmissionType
	Difficulty     int
	Reward         int
	StatTags       models.StringArray
	Proficiency    *string
}

func sanitizeGeneratedChallenges(raw []generatedChallengePayload, targetCount int, zoneName string) []sanitizedGeneratedChallenge {
	if targetCount <= 0 {
		return []sanitizedGeneratedChallenge{}
	}

	result := make([]sanitizedGeneratedChallenge, 0, targetCount)
	seenQuestions := map[string]struct{}{}

	for _, item := range raw {
		if len(result) >= targetCount {
			break
		}
		sanitized := sanitizeGeneratedChallenge(item, len(result), zoneName)
		key := strings.ToLower(strings.TrimSpace(sanitized.Question))
		if key == "" {
			continue
		}
		if _, exists := seenQuestions[key]; exists {
			continue
		}
		seenQuestions[key] = struct{}{}
		result = append(result, sanitized)
	}

	for len(result) < targetCount {
		fallback := fallbackGeneratedChallenge(zoneName, len(result))
		key := strings.ToLower(strings.TrimSpace(fallback.Question))
		if _, exists := seenQuestions[key]; exists {
			fallback.Question = fmt.Sprintf("%s (Variation %d)", fallback.Question, len(result)+1)
			key = strings.ToLower(strings.TrimSpace(fallback.Question))
		}
		seenQuestions[key] = struct{}{}
		result = append(result, fallback)
	}

	return result
}

func sanitizeGeneratedChallenge(raw generatedChallengePayload, index int, zoneName string) sanitizedGeneratedChallenge {
	submissionType := sanitizeChallengeSubmissionType(raw.SubmissionType, index)

	question := strings.TrimSpace(raw.Question)
	if question == "" {
		return fallbackGeneratedChallenge(zoneName, index)
	}
	if len(question) > 700 {
		question = strings.TrimSpace(question[:700])
	}
	question = ensureChallengeProofInstruction(question, submissionType)

	description := strings.TrimSpace(raw.Description)
	if description == "" {
		description = fmt.Sprintf(
			"A roleplay challenge unfolding in %s. Use scene details that clearly support the player action and produce strong visual cues for image generation.",
			zoneName,
		)
	}
	if len(description) > 1400 {
		description = strings.TrimSpace(description[:1400])
	}

	difficulty := 20
	if raw.Difficulty != nil {
		difficulty = clampInt(*raw.Difficulty, 0, 40)
	}
	reward := 30
	if raw.Reward != nil {
		reward = clampInt(*raw.Reward, 0, 100)
	}

	statTags := sanitizeChallengeStatTags(raw.StatTags, index)
	proficiency := sanitizeChallengeProficiency(raw.Proficiency)

	return sanitizedGeneratedChallenge{
		Question:       question,
		Description:    description,
		SubmissionType: submissionType,
		Difficulty:     difficulty,
		Reward:         reward,
		StatTags:       statTags,
		Proficiency:    proficiency,
	}
}

func sanitizeChallengeSubmissionType(raw string, index int) models.QuestNodeSubmissionType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(models.QuestNodeSubmissionTypePhoto):
		return models.QuestNodeSubmissionTypePhoto
	case string(models.QuestNodeSubmissionTypeText):
		return models.QuestNodeSubmissionTypeText
	default:
		if index%2 == 0 {
			return models.QuestNodeSubmissionTypePhoto
		}
		return models.QuestNodeSubmissionTypeText
	}
}

func sanitizeChallengeStatTags(raw []string, index int) models.StringArray {
	if len(raw) == 0 {
		tag := challengeGenerationDefaultStatTags[index%len(challengeGenerationDefaultStatTags)]
		return models.StringArray{tag}
	}

	seen := map[string]struct{}{}
	result := make(models.StringArray, 0, 3)
	for _, tag := range raw {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := challengeGenerationValidStatTags[normalized]; !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
		if len(result) >= 3 {
			break
		}
	}
	if len(result) == 0 {
		tag := challengeGenerationDefaultStatTags[index%len(challengeGenerationDefaultStatTags)]
		return models.StringArray{tag}
	}
	return result
}

func sanitizeChallengeProficiency(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	if len(trimmed) > 80 {
		trimmed = strings.TrimSpace(trimmed[:80])
	}
	return &trimmed
}

func ensureChallengeProofInstruction(question string, submissionType models.QuestNodeSubmissionType) string {
	normalized := strings.ToLower(question)
	switch submissionType {
	case models.QuestNodeSubmissionTypePhoto:
		if strings.Contains(normalized, "photo") ||
			strings.Contains(normalized, "picture") ||
			strings.Contains(normalized, "image") ||
			strings.Contains(normalized, "selfie") ||
			strings.Contains(normalized, "snapshot") {
			return question
		}
		return strings.TrimSpace(question + " Take a photo as proof.")
	case models.QuestNodeSubmissionTypeText:
		if strings.Contains(normalized, "text") ||
			strings.Contains(normalized, "write") ||
			strings.Contains(normalized, "written") ||
			strings.Contains(normalized, "describe") ||
			strings.Contains(normalized, "report") ||
			strings.Contains(normalized, "journal") {
			return question
		}
		return strings.TrimSpace(question + " Submit a short text write-up as proof.")
	default:
		return question
	}
}

func fallbackGeneratedChallenge(zoneName string, index int) sanitizedGeneratedChallenge {
	action := challengeGenerationFallbackActions[index%len(challengeGenerationFallbackActions)]
	submissionType := sanitizeChallengeSubmissionType("", index)

	question := fmt.Sprintf(
		"Roleplay a local adventurer in %s and %s. Keep it immersive and specific to your surroundings.",
		zoneName,
		action,
	)
	question = ensureChallengeProofInstruction(question, submissionType)

	description := fmt.Sprintf(
		"Set this challenge in %s with distinct environmental flavor: weathered architecture, ambient movement, and small grounded details that imply history. The player action should feel intentional, theatrical, and believable in a public setting. Emphasize props, posture, and nearby textures so the generated image can show a clear roleplay moment.",
		zoneName,
	)

	tag := challengeGenerationDefaultStatTags[index%len(challengeGenerationDefaultStatTags)]
	proficiency := tag + " tactics"
	if len(tag) > 0 {
		proficiency = strings.ToUpper(tag[:1]) + tag[1:] + " tactics"
	}

	return sanitizedGeneratedChallenge{
		Question:       question,
		Description:    description,
		SubmissionType: submissionType,
		Difficulty:     clampInt(14+(index%5)*4, 0, 40),
		Reward:         clampInt(20+(index%6)*8, 0, 100),
		StatTags:       models.StringArray{challengeGenerationDefaultStatTags[index%len(challengeGenerationDefaultStatTags)]},
		Proficiency:    &proficiency,
	}
}

func challengeGenerationLocation(zone models.Zone, offset int) (float64, float64) {
	point := zone.GetRandomPoint()
	if point.X() != 0 || point.Y() != 0 {
		return point.Y(), point.X()
	}
	if zone.Latitude != 0 || zone.Longitude != 0 {
		return zone.Latitude, zone.Longitude
	}

	baseLat := 40.7589
	baseLng := -73.98513
	jitter := float64((offset%7)-3) * 0.00012
	return baseLat + jitter, baseLng - jitter
}

func buildRecentScenarioThemeHints(scenarios []models.Scenario, limit int) string {
	if len(scenarios) == 0 || limit <= 0 {
		return "- none"
	}

	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].CreatedAt.After(scenarios[j].CreatedAt)
	})

	lines := make([]string, 0, limit)
	for _, scenario := range scenarios {
		prompt := strings.TrimSpace(scenario.Prompt)
		if prompt == "" {
			continue
		}
		prompt = strings.ReplaceAll(prompt, "\n", " ")
		if len(prompt) > 220 {
			prompt = strings.TrimSpace(prompt[:220]) + "..."
		}
		lines = append(lines, "- "+prompt)
		if len(lines) >= limit {
			break
		}
	}

	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func (p *GenerateChallengesProcessor) buildRecentChallengeAvoidance(ctx context.Context, zoneID uuid.UUID, limit int) string {
	if limit <= 0 {
		return "- none"
	}

	challenges, err := p.dbClient.Challenge().FindByZoneID(ctx, zoneID)
	if err != nil || len(challenges) == 0 {
		return "- none"
	}

	sort.Slice(challenges, func(i, j int) bool {
		return challenges[i].CreatedAt.After(challenges[j].CreatedAt)
	})

	lines := make([]string, 0, limit)
	for _, challenge := range challenges {
		question := strings.TrimSpace(challenge.Question)
		if question == "" {
			continue
		}
		question = strings.ReplaceAll(question, "\n", " ")
		if len(question) > 220 {
			question = strings.TrimSpace(question[:220]) + "..."
		}
		lines = append(lines, "- "+question)
		if len(lines) >= limit {
			break
		}
	}

	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}
