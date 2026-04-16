package models

import (
	"fmt"
	"strings"
)

type ZoneContentFlushType string

const (
	ZoneContentFlushTypePointsOfInterest ZoneContentFlushType = "pointsOfInterest"
	ZoneContentFlushTypeQuests           ZoneContentFlushType = "quests"
	ZoneContentFlushTypeChallenges       ZoneContentFlushType = "challenges"
	ZoneContentFlushTypeScenarios        ZoneContentFlushType = "scenarios"
	ZoneContentFlushTypeExpositions      ZoneContentFlushType = "expositions"
	ZoneContentFlushTypeMonsters         ZoneContentFlushType = "monsters"
	ZoneContentFlushTypeTreasureChests   ZoneContentFlushType = "treasureChests"
	ZoneContentFlushTypeHealingFountains ZoneContentFlushType = "healingFountains"
	ZoneContentFlushTypeResources        ZoneContentFlushType = "resources"
	ZoneContentFlushTypeMovementPatterns ZoneContentFlushType = "movementPatterns"
	ZoneContentFlushTypeJobs             ZoneContentFlushType = "jobs"
)

var allZoneContentFlushTypes = []ZoneContentFlushType{
	ZoneContentFlushTypePointsOfInterest,
	ZoneContentFlushTypeQuests,
	ZoneContentFlushTypeChallenges,
	ZoneContentFlushTypeScenarios,
	ZoneContentFlushTypeExpositions,
	ZoneContentFlushTypeMonsters,
	ZoneContentFlushTypeTreasureChests,
	ZoneContentFlushTypeHealingFountains,
	ZoneContentFlushTypeResources,
	ZoneContentFlushTypeMovementPatterns,
	ZoneContentFlushTypeJobs,
}

type ZoneContentFlushOptions struct {
	ContentTypes []ZoneContentFlushType `json:"contentTypes"`
}

func AllZoneContentFlushTypes() []ZoneContentFlushType {
	return append([]ZoneContentFlushType(nil), allZoneContentFlushTypes...)
}

func DefaultZoneContentFlushOptions() ZoneContentFlushOptions {
	return NewZoneContentFlushOptions(AllZoneContentFlushTypes())
}

func NewZoneContentFlushOptions(contentTypes []ZoneContentFlushType) ZoneContentFlushOptions {
	return ZoneContentFlushOptions{
		ContentTypes: normalizeZoneContentFlushTypes(contentTypes),
	}
}

func ParseZoneContentFlushTypes(rawContentTypes []string) ([]ZoneContentFlushType, error) {
	if len(rawContentTypes) == 0 {
		return nil, nil
	}

	normalized := make([]ZoneContentFlushType, 0, len(rawContentTypes))
	seen := make(map[ZoneContentFlushType]struct{}, len(rawContentTypes))
	for _, rawContentType := range rawContentTypes {
		contentType, err := parseZoneContentFlushType(rawContentType)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[contentType]; exists {
			continue
		}
		seen[contentType] = struct{}{}
		normalized = append(normalized, contentType)
	}

	return normalized, nil
}

func (o ZoneContentFlushOptions) Includes(contentType ZoneContentFlushType) bool {
	for _, candidate := range o.ContentTypes {
		if candidate == contentType {
			return true
		}
	}
	return false
}

func normalizeZoneContentFlushTypes(contentTypes []ZoneContentFlushType) []ZoneContentFlushType {
	if len(contentTypes) == 0 {
		return nil
	}

	normalized := make([]ZoneContentFlushType, 0, len(contentTypes))
	seen := make(map[ZoneContentFlushType]struct{}, len(contentTypes))
	for _, contentType := range contentTypes {
		if contentType == "" {
			continue
		}
		if _, exists := seen[contentType]; exists {
			continue
		}
		seen[contentType] = struct{}{}
		normalized = append(normalized, contentType)
	}

	return normalized
}

func parseZoneContentFlushType(rawContentType string) (ZoneContentFlushType, error) {
	trimmed := strings.TrimSpace(rawContentType)
	if trimmed == "" {
		return "", fmt.Errorf("contentTypes entries cannot be empty")
	}

	for _, allowedType := range allZoneContentFlushTypes {
		if strings.EqualFold(trimmed, string(allowedType)) {
			return allowedType, nil
		}
	}

	return "", fmt.Errorf(
		"invalid contentType %q (expected one of: %s)",
		rawContentType,
		strings.Join(zoneContentFlushTypeValues(), ", "),
	)
}

func zoneContentFlushTypeValues() []string {
	values := make([]string, 0, len(allZoneContentFlushTypes))
	for _, contentType := range allZoneContentFlushTypes {
		values = append(values, string(contentType))
	}
	return values
}
