package models

import "strings"

const ScenarioPlaceholderImageURL = "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png"

func IsScenarioPlaceholderImageURL(raw string) bool {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return false
	}
	return strings.Contains(normalized, "scenario-undiscovered.png")
}

func ResolveScenarioArtURLs(imageURL, thumbnailURL string) (string, string, bool) {
	normalizedImage := strings.TrimSpace(imageURL)
	normalizedThumbnail := strings.TrimSpace(thumbnailURL)

	imageIsActual := normalizedImage != "" && !IsScenarioPlaceholderImageURL(normalizedImage)
	thumbnailIsActual := normalizedThumbnail != "" && !IsScenarioPlaceholderImageURL(normalizedThumbnail)

	switch {
	case imageIsActual && thumbnailIsActual:
		return normalizedImage, normalizedThumbnail, false
	case imageIsActual:
		return normalizedImage, normalizedImage, false
	case thumbnailIsActual:
		return normalizedThumbnail, normalizedThumbnail, false
	default:
		return ScenarioPlaceholderImageURL, ScenarioPlaceholderImageURL, true
	}
}
