package server

import "github.com/MaxBlaushild/poltergeist/pkg/models"

func normalizeZoneKindRequest(value *string) string {
	if value == nil {
		return ""
	}
	return models.NormalizeZoneKind(*value)
}

func mergeZoneKindRequest(value *string, fallback string) string {
	if value == nil {
		return models.NormalizeZoneKind(fallback)
	}
	return models.NormalizeZoneKind(*value)
}
