package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

var expositionTemplateAbstractSpeakerKeywords = []string{
	"echo",
	"warning",
	"whisper",
	"imprint",
	"trace",
	"residue",
	"overheard",
}

type expositionTemplateSpeakerPortraitCandidate struct {
	SpeakerName string
	Description string
}

type GenerateExpositionTemplateSpeakerPortraitsProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateExpositionTemplateSpeakerPortraitsProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateExpositionTemplateSpeakerPortraitsProcessor {
	return GenerateExpositionTemplateSpeakerPortraitsProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateExpositionTemplateSpeakerPortraitsProcessor) ProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload jobs.GenerateExpositionTemplateSpeakerPortraitsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	template, err := p.dbClient.ExpositionTemplate().FindByID(ctx, payload.ExpositionTemplateID)
	if err != nil {
		return fmt.Errorf("failed to find exposition template: %w", err)
	}
	if template == nil {
		return fmt.Errorf("exposition template not found")
	}

	candidates := collectExpositionTemplateSpeakerPortraitCandidates(template)
	if len(candidates) == 0 {
		return nil
	}

	portraitURLs := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		request := deep_priest.GenerateImageRequest{
			Prompt: characterImagePrompt(candidate.SpeakerName, candidate.Description, nil),
		}
		deep_priest.ApplyGenerateImageDefaults(&request)

		imagePayload, err := p.deepPriestClient.GenerateImage(request)
		if err != nil {
			return fmt.Errorf("failed to generate exposition speaker portrait for %q: %w", candidate.SpeakerName, err)
		}
		imageBytes, err := decodeCharacterImagePayload(imagePayload)
		if err != nil {
			return fmt.Errorf("failed to decode exposition speaker portrait for %q: %w", candidate.SpeakerName, err)
		}
		imageURL, err := p.uploadExpositionSpeakerPortrait(ctx, template.ID, candidate.SpeakerName, imageBytes)
		if err != nil {
			return fmt.Errorf("failed to upload exposition speaker portrait for %q: %w", candidate.SpeakerName, err)
		}
		portraitURLs[normalizeExpositionSpeakerPortraitKey(candidate.SpeakerName)] = imageURL
	}

	dialogue, changed := applyExpositionSpeakerPortraitURLs(template.Dialogue, portraitURLs)
	if !changed {
		return nil
	}

	template.Dialogue = dialogue
	template.UpdatedAt = time.Now()
	if err := p.dbClient.ExpositionTemplate().Update(ctx, template.ID, template); err != nil {
		return fmt.Errorf("failed to update exposition template dialogue portraits: %w", err)
	}

	if err := p.syncLinkedExpositions(ctx, template.ID, portraitURLs); err != nil {
		return err
	}
	return nil
}

func (p *GenerateExpositionTemplateSpeakerPortraitsProcessor) uploadExpositionSpeakerPortrait(
	ctx context.Context,
	templateID uuid.UUID,
	speakerName string,
	imageBytes []byte,
) (string, error) {
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("no image data provided")
	}

	imageFormat, err := util.DetectImageFormat(imageBytes)
	if err != nil {
		return "", err
	}
	imageExtension, err := util.GetImageExtension(imageFormat)
	if err != nil {
		return "", err
	}

	imageName := fmt.Sprintf(
		"exposition-template-speakers/%s/%s-%d.%s",
		templateID.String(),
		normalizeExpositionSpeakerPortraitFilename(speakerName),
		time.Now().UnixNano(),
		imageExtension,
	)
	return p.awsClient.UploadImageToS3("crew-profile-icons", imageName, imageBytes)
}

func (p *GenerateExpositionTemplateSpeakerPortraitsProcessor) syncLinkedExpositions(
	ctx context.Context,
	templateID uuid.UUID,
	portraitURLs map[string]string,
) error {
	linkedExpositions, err := p.dbClient.Exposition().FindByTemplateID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to load linked expositions for portrait sync: %w", err)
	}
	for index := range linkedExpositions {
		dialogue, changed := applyExpositionSpeakerPortraitURLs(linkedExpositions[index].Dialogue, portraitURLs)
		if !changed {
			continue
		}
		linkedExpositions[index].Dialogue = dialogue
		linkedExpositions[index].UpdatedAt = time.Now()
		if err := p.dbClient.Exposition().Update(ctx, linkedExpositions[index].ID, &linkedExpositions[index]); err != nil {
			return fmt.Errorf("failed to sync linked exposition %s dialogue portraits: %w", linkedExpositions[index].ID.String(), err)
		}
	}
	return nil
}

func collectExpositionTemplateSpeakerPortraitCandidates(
	template *models.ExpositionTemplate,
) []expositionTemplateSpeakerPortraitCandidate {
	if template == nil || len(template.Dialogue) == 0 {
		return nil
	}

	actualPortraitSpeakers := map[string]struct{}{}
	for _, message := range template.Dialogue {
		speakerName := strings.TrimSpace(message.SpeakerName)
		if speakerName == "" {
			continue
		}
		if !expositionSpeakerPortraitNeedsGeneration(message.PortraitURL) {
			actualPortraitSpeakers[normalizeExpositionSpeakerPortraitKey(speakerName)] = struct{}{}
		}
	}

	seen := map[string]struct{}{}
	candidates := make([]expositionTemplateSpeakerPortraitCandidate, 0, len(template.Dialogue))
	for _, message := range template.Dialogue {
		speakerName := strings.TrimSpace(message.SpeakerName)
		if speakerName == "" || !shouldGeneratePortraitForExpositionSpeaker(speakerName) {
			continue
		}
		if !expositionSpeakerPortraitNeedsGeneration(message.PortraitURL) {
			continue
		}
		key := normalizeExpositionSpeakerPortraitKey(speakerName)
		if _, exists := actualPortraitSpeakers[key]; exists {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, expositionTemplateSpeakerPortraitCandidate{
			SpeakerName: speakerName,
			Description: buildExpositionTemplateSpeakerPortraitDescription(template, speakerName),
		})
	}
	return candidates
}

func applyExpositionSpeakerPortraitURLs(
	dialogue models.DialogueSequence,
	portraitURLs map[string]string,
) (models.DialogueSequence, bool) {
	if len(dialogue) == 0 || len(portraitURLs) == 0 {
		return dialogue, false
	}
	updated := append(models.DialogueSequence{}, dialogue...)
	changed := false
	for index, message := range updated {
		if !expositionSpeakerPortraitNeedsGeneration(message.PortraitURL) {
			continue
		}
		url := strings.TrimSpace(portraitURLs[normalizeExpositionSpeakerPortraitKey(message.SpeakerName)])
		if url == "" {
			continue
		}
		updated[index].PortraitURL = url
		changed = true
	}
	return updated, changed
}

func expositionSpeakerPortraitNeedsGeneration(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return true
	}
	normalized := strings.ToLower(trimmed)
	return strings.Contains(normalized, "character-undiscovered") ||
		strings.Contains(normalized, "loading-image.gif")
}

func shouldGeneratePortraitForExpositionSpeaker(speakerName string) bool {
	normalized := strings.ToLower(strings.TrimSpace(speakerName))
	if normalized == "" {
		return false
	}
	for _, keyword := range expositionTemplateAbstractSpeakerKeywords {
		if strings.Contains(normalized, keyword) {
			return false
		}
	}
	return true
}

func buildExpositionTemplateSpeakerPortraitDescription(
	template *models.ExpositionTemplate,
	speakerName string,
) string {
	parts := []string{
		fmt.Sprintf("%s is a recurring street-level urban fantasy character implied by this exposition scene.", strings.TrimSpace(speakerName)),
	}
	if template != nil {
		if title := strings.TrimSpace(template.Title); title != "" {
			parts = append(parts, fmt.Sprintf("Scene title: %s.", title))
		}
		if description := strings.TrimSpace(template.Description); description != "" {
			parts = append(parts, description)
		}
	}
	speakerLines := collectExpositionTemplateSpeakerLines(template, speakerName, 2)
	if len(speakerLines) > 0 {
		parts = append(parts, fmt.Sprintf("Dialogue cues: %s", strings.Join(speakerLines, " ")))
	}
	return truncate(strings.Join(parts, " "), 700)
}

func collectExpositionTemplateSpeakerLines(
	template *models.ExpositionTemplate,
	speakerName string,
	limit int,
) []string {
	if template == nil || len(template.Dialogue) == 0 || limit <= 0 {
		return nil
	}
	lines := make([]string, 0, limit)
	for _, message := range template.Dialogue {
		if !strings.EqualFold(strings.TrimSpace(message.SpeakerName), strings.TrimSpace(speakerName)) {
			continue
		}
		text := strings.TrimSpace(message.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
		if len(lines) >= limit {
			break
		}
	}
	if len(lines) > 0 {
		return lines
	}
	for _, message := range template.Dialogue {
		text := strings.TrimSpace(message.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
		if len(lines) >= limit {
			break
		}
	}
	return lines
}

func normalizeExpositionSpeakerPortraitKey(speakerName string) string {
	return strings.ToLower(strings.TrimSpace(speakerName))
}

func normalizeExpositionSpeakerPortraitFilename(speakerName string) string {
	normalized := strings.ToLower(strings.TrimSpace(speakerName))
	if normalized == "" {
		return "speaker"
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, char := range normalized {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'):
			builder.WriteRune(char)
			lastUnderscore = false
		case !lastUnderscore && builder.Len() > 0:
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	value := strings.Trim(builder.String(), "_")
	if value == "" {
		return "speaker"
	}
	return value
}
