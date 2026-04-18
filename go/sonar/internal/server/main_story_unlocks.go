package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func mainStoryBeatUnlockFlags(beat models.MainStoryBeatDraft) models.StringArray {
	if completionFlag := mainStoryBeatCompletionFlag(beat); completionFlag != "" {
		return models.StringArray{completionFlag}
	}
	return normalizeStoryFlagKeys([]string(beat.RequiredStoryFlags))
}

func (s *server) resolveMainStoryUnlockPointOfInterest(
	ctx context.Context,
	resolution mainStoryBeatQuestGiverResolution,
	hint string,
) (*models.PointOfInterest, error) {
	if resolution.Character != nil && resolution.Character.PointOfInterestID != nil {
		if strings.TrimSpace(hint) == "" {
			return s.dbClient.PointOfInterest().FindByID(ctx, *resolution.Character.PointOfInterestID)
		}
	}

	var candidates []models.PointOfInterest
	if resolution.Character != nil && resolution.Character.PointOfInterestID != nil {
		zone, err := s.dbClient.PointOfInterest().FindZoneForPointOfInterest(ctx, *resolution.Character.PointOfInterestID)
		if err == nil && zone != nil {
			candidates, _ = s.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ZoneID)
		}
	}
	if len(candidates) == 0 {
		allPoints, err := s.dbClient.PointOfInterest().FindAll(ctx)
		if err != nil {
			return nil, err
		}
		candidates = allPoints
	}
	best := pickBestPointOfInterestForHint(candidates, hint, nil)
	if best == nil && resolution.Character != nil && resolution.Character.PointOfInterestID != nil {
		return s.dbClient.PointOfInterest().FindByID(ctx, *resolution.Character.PointOfInterestID)
	}
	if best == nil {
		return nil, fmt.Errorf("no point of interest could be resolved")
	}
	return best, nil
}

func (s *server) materializeMainStoryUnlockedScenario(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
	spec models.MainStoryUnlockedScenario,
	requiredFlags models.StringArray,
) error {
	pointOfInterest, err := s.resolveMainStoryUnlockPointOfInterest(ctx, resolution, spec.PointOfInterestHint)
	if err != nil {
		return err
	}
	zone, err := s.dbClient.Zone().FindByPointOfInterestID(ctx, pointOfInterest.ID)
	if err != nil {
		return err
	}
	latitude, longitude, err := parsePointOfInterestCoordinates(pointOfInterest)
	if err != nil {
		return err
	}
	imageURL, thumbnailURL, shouldGenerateImage := models.ResolveScenarioArtURLs(
		"",
		"",
	)

	scenario := &models.Scenario{
		ZoneID:             zone.ID,
		PointOfInterestID:  &pointOfInterest.ID,
		Latitude:           latitude,
		Longitude:          longitude,
		Prompt:             strings.TrimSpace(spec.Prompt),
		RequiredStoryFlags: requiredFlags,
		InternalTags: normalizeQuestTemplateInternalTags(append(
			append(
				append([]string{}, []string(draft.InternalTags)...),
				[]string(beat.InternalTags)...,
			),
			append([]string{"main_story_unlock", "main_story_scenario"}, []string(spec.InternalTags)...)...,
		)),
		ImageURL:           imageURL,
		ThumbnailURL:       thumbnailURL,
		ScaleWithUserLevel: true,
		RewardMode:         models.RewardModeRandom,
		RandomRewardSize:   models.RandomRewardSizeSmall,
		Difficulty:         max(1, spec.Difficulty),
		OpenEnded:          true,
	}
	if err := s.dbClient.Scenario().Create(ctx, scenario); err != nil {
		return err
	}
	if shouldGenerateImage {
		if err := s.enqueueScenarioImageGenerationTask(scenario.ID); err != nil {
			log.Printf(
				"materializeMainStoryUnlockedScenario: failed to queue scenario image generation for %s: %v",
				scenario.ID,
				err,
			)
		}
	}
	return nil
}

func (s *server) materializeMainStoryUnlockedChallenge(
	ctx context.Context,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
	spec models.MainStoryUnlockedChallenge,
	requiredFlags models.StringArray,
) error {
	pointOfInterest, err := s.resolveMainStoryUnlockPointOfInterest(ctx, resolution, spec.PointOfInterestHint)
	if err != nil {
		return err
	}
	zone, err := s.dbClient.Zone().FindByPointOfInterestID(ctx, pointOfInterest.ID)
	if err != nil {
		return err
	}
	latitude, longitude, err := parsePointOfInterestCoordinates(pointOfInterest)
	if err != nil {
		return err
	}

	submissionType := models.QuestNodeSubmissionType(strings.TrimSpace(strings.ToLower(string(spec.SubmissionType))))
	if !submissionType.IsValid() {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	challenge := &models.Challenge{
		ID:                 uuid.New(),
		ZoneID:             zone.ID,
		PointOfInterestID:  &pointOfInterest.ID,
		Latitude:           latitude,
		Longitude:          longitude,
		Question:           strings.TrimSpace(spec.Question),
		Description:        strings.TrimSpace(spec.Description),
		RequiredStoryFlags: requiredFlags,
		ScaleWithUserLevel: true,
		RewardMode:         models.RewardModeRandom,
		RandomRewardSize:   models.RandomRewardSizeSmall,
		SubmissionType:     submissionType,
		Difficulty:         max(1, spec.Difficulty),
		StatTags:           parseChallengeStatTags([]string(spec.StatTags)),
		Proficiency:        spec.Proficiency,
	}
	return s.dbClient.Challenge().Create(ctx, challenge)
}

func monsterTemplateTypeForEncounterType(
	encounterType models.MonsterEncounterType,
) models.MonsterTemplateType {
	switch encounterType {
	case models.MonsterEncounterTypeBoss:
		return models.MonsterTemplateTypeBoss
	case models.MonsterEncounterTypeRaid:
		return models.MonsterTemplateTypeRaid
	default:
		return models.MonsterTemplateTypeMonster
	}
}

func (s *server) materializeMainStoryUnlockedEncounter(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
	spec models.MainStoryUnlockedEncounter,
	requiredFlags models.StringArray,
) error {
	pointOfInterest, err := s.resolveMainStoryUnlockPointOfInterest(ctx, resolution, spec.PointOfInterestHint)
	if err != nil {
		return err
	}
	zone, err := s.dbClient.Zone().FindByPointOfInterestID(ctx, pointOfInterest.ID)
	if err != nil {
		return err
	}
	latitude, longitude, err := parsePointOfInterestCoordinates(pointOfInterest)
	if err != nil {
		return err
	}

	templates, err := s.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return err
	}
	templateIDs, err := s.ensureQuestMonsterTemplateIDs(
		ctx,
		&templates,
		questMonsterTemplateRequest{
			Count:            max(1, spec.MonsterCount),
			MonsterType:      monsterTemplateTypeForEncounterType(spec.EncounterType),
			ThemePrompt:      strings.TrimSpace(draft.Premise),
			EncounterConcept: strings.TrimSpace(spec.Name + " " + spec.Description),
			LocationConcept:  strings.TrimSpace(pointOfInterest.Name + " " + pointOfInterest.Description),
			EncounterTone:    append([]string{}, spec.EncounterTone...),
			SeedHints: append(
				append([]string{}, spec.MonsterTemplateHints...),
				[]string(beat.MonsterTemplateSeeds)...,
			),
		},
		nil,
	)
	if err != nil {
		return err
	}
	templateByID := make(map[string]models.MonsterTemplate, len(templates))
	for _, template := range templates {
		templateByID[template.ID.String()] = template
	}

	encounter := &models.MonsterEncounter{
		Name:               strings.TrimSpace(spec.Name),
		Description:        strings.TrimSpace(spec.Description),
		EncounterType:      models.NormalizeMonsterEncounterType(string(spec.EncounterType)),
		RequiredStoryFlags: requiredFlags,
		ScaleWithUserLevel: true,
		RewardMode:         models.RewardModeRandom,
		RandomRewardSize:   models.RandomRewardSizeSmall,
		ZoneID:             zone.ID,
		PointOfInterestID:  &pointOfInterest.ID,
		Latitude:           latitude,
		Longitude:          longitude,
	}
	if encounter.EncounterType == "" {
		encounter.EncounterType = models.MonsterEncounterTypeMonster
	}
	if err := s.dbClient.MonsterEncounter().Create(ctx, encounter); err != nil {
		return err
	}

	level := max(1, spec.TargetLevel)
	if level <= 1 {
		level = max(1, beat.MonsterEncounterTargetLevel)
	}
	members := make([]models.MonsterEncounterMember, 0, len(templateIDs))
	for index, rawTemplateID := range templateIDs {
		template, ok := templateByID[strings.TrimSpace(rawTemplateID)]
		if !ok {
			continue
		}
		templateID := template.ID
		monster := &models.Monster{
			Name:                  strings.TrimSpace(template.Name),
			Description:           strings.TrimSpace(template.Description),
			ImageURL:              strings.TrimSpace(template.ImageURL),
			ThumbnailURL:          strings.TrimSpace(template.ThumbnailURL),
			ZoneID:                zone.ID,
			GenreID:               template.GenreID,
			Genre:                 template.Genre,
			Latitude:              latitude,
			Longitude:             longitude,
			TemplateID:            &templateID,
			Level:                 level,
			RewardMode:            models.RewardModeExplicit,
			RandomRewardSize:      models.RandomRewardSizeSmall,
			RewardExperience:      0,
			RewardGold:            0,
			ImageGenerationStatus: models.MonsterImageGenerationStatusNone,
		}
		if err := s.dbClient.Monster().Create(ctx, monster); err != nil {
			return err
		}
		members = append(members, models.MonsterEncounterMember{
			MonsterID: monster.ID,
			Slot:      index,
		})
	}
	if len(members) == 0 {
		return fmt.Errorf("no monster members could be created")
	}
	return s.dbClient.MonsterEncounter().ReplaceMembers(ctx, encounter.ID, members)
}

func (s *server) materializeMainStoryBeatUnlocks(
	ctx context.Context,
	draft *models.MainStorySuggestionDraft,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
) models.StringArray {
	requiredFlags := mainStoryBeatUnlockFlags(beat)
	warnings := models.StringArray{}

	for _, scenario := range beat.UnlockedScenarios {
		if err := s.materializeMainStoryUnlockedScenario(
			ctx,
			draft,
			beat,
			resolution,
			scenario,
			requiredFlags,
		); err != nil {
			warnings = append(warnings, fmt.Sprintf("scenario unlock %q failed: %v", strings.TrimSpace(scenario.Name), err))
		}
	}
	for _, challenge := range beat.UnlockedChallenges {
		if err := s.materializeMainStoryUnlockedChallenge(
			ctx,
			beat,
			resolution,
			challenge,
			requiredFlags,
		); err != nil {
			warnings = append(warnings, fmt.Sprintf("challenge unlock %q failed: %v", strings.TrimSpace(challenge.Question), err))
		}
	}
	for _, encounter := range beat.UnlockedMonsterEncounters {
		if err := s.materializeMainStoryUnlockedEncounter(
			ctx,
			draft,
			beat,
			resolution,
			encounter,
			requiredFlags,
		); err != nil {
			warnings = append(warnings, fmt.Sprintf("monster unlock %q failed: %v", strings.TrimSpace(encounter.Name), err))
		}
	}

	return warnings
}
