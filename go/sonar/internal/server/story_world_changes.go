package server

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func filterActiveStoryWorldChanges(
	changes []models.StoryWorldChange,
	activeFlags map[string]bool,
) []models.StoryWorldChange {
	active := make([]models.StoryWorldChange, 0, len(changes))
	for _, change := range changes {
		if !hasRequiredStoryFlags(change.RequiredStoryFlags, activeFlags) {
			continue
		}
		active = append(active, change)
	}
	sort.SliceStable(active, func(i, j int) bool {
		if active[i].Priority != active[j].Priority {
			return active[i].Priority > active[j].Priority
		}
		if active[i].BeatOrder != active[j].BeatOrder {
			return active[i].BeatOrder > active[j].BeatOrder
		}
		return active[i].UpdatedAt.After(active[j].UpdatedAt)
	})
	return active
}

func pointOfInterestTextChangesByID(
	changes []models.StoryWorldChange,
) map[uuid.UUID]models.StoryWorldChange {
	byID := map[uuid.UUID]models.StoryWorldChange{}
	for _, change := range changes {
		if change.EffectType != models.StoryWorldChangeTypeShowPOIText || change.PointOfInterestID == nil {
			continue
		}
		if _, exists := byID[*change.PointOfInterestID]; exists {
			continue
		}
		byID[*change.PointOfInterestID] = change
	}
	return byID
}

func characterMoveChangesByID(
	changes []models.StoryWorldChange,
) map[uuid.UUID]models.StoryWorldChange {
	byID := map[uuid.UUID]models.StoryWorldChange{}
	for _, change := range changes {
		if change.EffectType != models.StoryWorldChangeTypeMoveCharacter || change.CharacterID == nil {
			continue
		}
		if _, exists := byID[*change.CharacterID]; exists {
			continue
		}
		byID[*change.CharacterID] = change
	}
	return byID
}

func applyStoryWorldTextChangesToPointOfInterests(
	pointsOfInterest []models.PointOfInterest,
	changes []models.StoryWorldChange,
) {
	textChanges := pointOfInterestTextChangesByID(changes)
	for index := range pointsOfInterest {
		change, ok := textChanges[pointsOfInterest[index].ID]
		if !ok {
			continue
		}
		if strings.TrimSpace(change.Description) != "" {
			pointsOfInterest[index].Description = strings.TrimSpace(change.Description)
		}
		if strings.TrimSpace(change.Clue) != "" {
			pointsOfInterest[index].Clue = strings.TrimSpace(change.Clue)
		}
	}
}

func removeCharacterFromPointOfInterests(
	pointsOfInterest []models.PointOfInterest,
	characterID uuid.UUID,
) (*models.Character, bool) {
	for pointIndex := range pointsOfInterest {
		filtered := make([]models.Character, 0, len(pointsOfInterest[pointIndex].Characters))
		var found *models.Character
		for _, character := range pointsOfInterest[pointIndex].Characters {
			if character.ID == characterID {
				characterCopy := character
				found = &characterCopy
				continue
			}
			filtered = append(filtered, character)
		}
		pointsOfInterest[pointIndex].Characters = filtered
		if found != nil {
			return found, true
		}
	}
	return nil, false
}

func appendCharacterToPointOfInterest(
	pointsOfInterest []models.PointOfInterest,
	pointOfInterestID uuid.UUID,
	character models.Character,
) {
	for pointIndex := range pointsOfInterest {
		if pointsOfInterest[pointIndex].ID != pointOfInterestID {
			continue
		}
		for _, existing := range pointsOfInterest[pointIndex].Characters {
			if existing.ID == character.ID {
				return
			}
		}
		character.PointOfInterestID = &pointOfInterestID
		pointsOfInterest[pointIndex].Characters = append(pointsOfInterest[pointIndex].Characters, character)
		return
	}
}

func (s *server) applyStoryWorldChangesToPointOfInterests(
	ctx context.Context,
	pointsOfInterest []models.PointOfInterest,
	activeFlags map[string]bool,
) error {
	if len(pointsOfInterest) == 0 || len(activeFlags) == 0 {
		return nil
	}
	allChanges, err := s.dbClient.StoryWorldChange().FindAll(ctx)
	if err != nil {
		return err
	}
	activeChanges := filterActiveStoryWorldChanges(allChanges, activeFlags)
	if len(activeChanges) == 0 {
		return nil
	}
	applyStoryWorldTextChangesToPointOfInterests(pointsOfInterest, activeChanges)

	moveChanges := characterMoveChangesByID(activeChanges)
	if len(moveChanges) == 0 {
		return nil
	}
	for characterID, change := range moveChanges {
		if change.DestinationPointOfInterestID == nil {
			continue
		}
		character, found := removeCharacterFromPointOfInterests(pointsOfInterest, characterID)
		if !found {
			loaded, err := s.dbClient.Character().FindByID(ctx, characterID)
			if err != nil || loaded == nil {
				continue
			}
			character = loaded
		}
		appendCharacterToPointOfInterest(pointsOfInterest, *change.DestinationPointOfInterestID, *character)
	}
	return nil
}

func (s *server) applyStoryWorldChangesToCharacter(
	ctx context.Context,
	character *models.Character,
	activeFlags map[string]bool,
) error {
	if character == nil || len(activeFlags) == 0 {
		return nil
	}
	allChanges, err := s.dbClient.StoryWorldChange().FindAll(ctx)
	if err != nil {
		return err
	}
	moveChanges := characterMoveChangesByID(filterActiveStoryWorldChanges(allChanges, activeFlags))
	change, ok := moveChanges[character.ID]
	if !ok || change.DestinationPointOfInterestID == nil {
		return nil
	}
	character.PointOfInterestID = change.DestinationPointOfInterestID
	pointOfInterest, err := s.dbClient.PointOfInterest().FindByID(ctx, *change.DestinationPointOfInterestID)
	if err != nil {
		return err
	}
	if pointOfInterest != nil {
		character.PointOfInterest = pointOfInterest
	}
	return nil
}

func scorePointOfInterestForHint(
	pointOfInterest models.PointOfInterest,
	hint string,
) int {
	if strings.TrimSpace(hint) == "" {
		return 0
	}
	haystacks := []string{
		pointOfInterest.Name,
		pointOfInterest.OriginalName,
		pointOfInterest.Description,
		pointOfInterest.Clue,
	}
	for _, tag := range pointOfInterest.Tags {
		haystacks = append(haystacks, tag.Value)
	}
	score := 0
	for _, token := range strings.Fields(strings.ToLower(strings.TrimSpace(hint))) {
		if len(token) < 3 {
			continue
		}
		for _, haystack := range haystacks {
			if strings.Contains(strings.ToLower(haystack), token) {
				score++
				break
			}
		}
	}
	return score
}

func pickBestPointOfInterestForHint(
	pointsOfInterest []models.PointOfInterest,
	hint string,
	excludeID *uuid.UUID,
) *models.PointOfInterest {
	var best *models.PointOfInterest
	bestScore := -1
	for index := range pointsOfInterest {
		pointOfInterest := pointsOfInterest[index]
		if excludeID != nil && pointOfInterest.ID == *excludeID {
			continue
		}
		score := scorePointOfInterestForHint(pointOfInterest, hint)
		if score > bestScore {
			best = &pointsOfInterest[index]
			bestScore = score
		}
	}
	if bestScore <= 0 && len(pointsOfInterest) > 0 {
		for index := range pointsOfInterest {
			if excludeID != nil && pointsOfInterest[index].ID == *excludeID {
				continue
			}
			return &pointsOfInterest[index]
		}
	}
	return best
}

func buildResolvedStoryWorldChanges(
	templateID uuid.UUID,
	questArchetypeID *uuid.UUID,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
	changes []models.MainStoryWorldChange,
	resolvedPointOfInterestID *uuid.UUID,
	destinationPointOfInterestID *uuid.UUID,
) []models.StoryWorldChange {
	requiredFlags := normalizeStoryFlagKeys([]string(beat.RequiredStoryFlags))
	if completionFlag := mainStoryBeatCompletionFlag(beat); completionFlag != "" {
		requiredFlags = normalizeStoryFlagKeys([]string{completionFlag})
	}
	resolved := make([]models.StoryWorldChange, 0, len(changes))
	now := time.Now()
	for _, change := range changes {
		entry := models.StoryWorldChange{
			ID:                  uuid.New(),
			CreatedAt:           now,
			UpdatedAt:           now,
			MainStoryTemplateID: templateID,
			QuestArchetypeID:    questArchetypeID,
			BeatOrder:           beat.OrderIndex,
			Priority:            1000 + max(0, beat.OrderIndex),
			EffectType:          models.NormalizeStoryWorldChangeType(change.Type),
			TargetKey:           strings.TrimSpace(change.TargetKey),
			RequiredStoryFlags:  requiredFlags,
			Description:         strings.TrimSpace(change.Description),
			Clue:                strings.TrimSpace(change.Clue),
		}
		if resolution.Character != nil {
			entry.CharacterID = &resolution.Character.ID
		}
		switch entry.EffectType {
		case models.StoryWorldChangeTypeShowPOIText:
			entry.PointOfInterestID = resolvedPointOfInterestID
		case models.StoryWorldChangeTypeMoveCharacter:
			entry.PointOfInterestID = resolvedPointOfInterestID
			entry.DestinationPointOfInterestID = destinationPointOfInterestID
		}
		resolved = append(resolved, entry)
	}
	return resolved
}

func (s *server) resolveMainStoryBeatWorldChanges(
	ctx context.Context,
	templateID uuid.UUID,
	questArchetypeID *uuid.UUID,
	beat models.MainStoryBeatDraft,
	resolution mainStoryBeatQuestGiverResolution,
) ([]models.StoryWorldChange, models.StringArray, error) {
	if len(beat.WorldChanges) == 0 || resolution.Character == nil {
		return nil, nil, nil
	}

	var sameZonePoints []models.PointOfInterest
	if resolution.Character.PointOfInterestID != nil {
		poiZone, err := s.dbClient.PointOfInterest().FindZoneForPointOfInterest(ctx, *resolution.Character.PointOfInterestID)
		if err == nil && poiZone != nil {
			sameZonePoints, _ = s.dbClient.PointOfInterest().FindAllForZone(ctx, poiZone.ZoneID)
		}
	}

	warnings := models.StringArray{}
	resolved := make([]models.StoryWorldChange, 0, len(beat.WorldChanges))
	for _, change := range beat.WorldChanges {
		changeType := models.NormalizeStoryWorldChangeType(change.Type)
		if changeType == "" {
			continue
		}

		var resolvedPointOfInterestID *uuid.UUID
		var destinationPointOfInterestID *uuid.UUID

		switch changeType {
		case models.StoryWorldChangeTypeShowPOIText:
			resolvedPointOfInterestID = resolution.Character.PointOfInterestID
			if resolvedPointOfInterestID == nil && strings.TrimSpace(change.PointOfInterestHint) != "" {
				candidates := sameZonePoints
				if len(candidates) == 0 {
					allPoints, err := s.dbClient.PointOfInterest().FindAll(ctx)
					if err != nil {
						return nil, nil, err
					}
					candidates = allPoints
				}
				if best := pickBestPointOfInterestForHint(candidates, change.PointOfInterestHint, nil); best != nil {
					resolvedPointOfInterestID = &best.ID
				}
			}
			if resolvedPointOfInterestID == nil {
				warnings = append(warnings, "show_poi_text world change could not resolve a point of interest")
				continue
			}
		case models.StoryWorldChangeTypeMoveCharacter:
			resolvedPointOfInterestID = resolution.Character.PointOfInterestID
			candidates := sameZonePoints
			if len(candidates) == 0 {
				allPoints, err := s.dbClient.PointOfInterest().FindAll(ctx)
				if err != nil {
					return nil, nil, err
				}
				candidates = allPoints
			}
			best := pickBestPointOfInterestForHint(candidates, change.DestinationHint, resolvedPointOfInterestID)
			if best == nil {
				warnings = append(warnings, "move_character world change could not resolve a destination point of interest")
				continue
			}
			destinationPointOfInterestID = &best.ID
		}

		resolved = append(
			resolved,
			buildResolvedStoryWorldChanges(
				templateID,
				questArchetypeID,
				beat,
				resolution,
				[]models.MainStoryWorldChange{change},
				resolvedPointOfInterestID,
				destinationPointOfInterestID,
			)...,
		)
	}

	return resolved, warnings, nil
}
