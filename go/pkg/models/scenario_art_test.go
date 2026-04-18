package models

import "testing"

func TestResolveScenarioArtURLsUsesActualTemplateImage(t *testing.T) {
	imageURL := "https://example.com/scenario.png"

	resolvedImage, resolvedThumbnail, shouldGenerate := ResolveScenarioArtURLs(imageURL, "")
	if shouldGenerate {
		t.Fatal("expected existing image art to skip generation")
	}
	if resolvedImage != imageURL {
		t.Fatalf("expected image URL %q, got %q", imageURL, resolvedImage)
	}
	if resolvedThumbnail != imageURL {
		t.Fatalf("expected thumbnail to fall back to image URL %q, got %q", imageURL, resolvedThumbnail)
	}
}

func TestResolveScenarioArtURLsUsesActualTemplateThumbnail(t *testing.T) {
	thumbnailURL := "https://example.com/scenario-thumb.png"

	resolvedImage, resolvedThumbnail, shouldGenerate := ResolveScenarioArtURLs("", thumbnailURL)
	if shouldGenerate {
		t.Fatal("expected existing thumbnail art to skip generation")
	}
	if resolvedImage != thumbnailURL {
		t.Fatalf("expected image to fall back to thumbnail URL %q, got %q", thumbnailURL, resolvedImage)
	}
	if resolvedThumbnail != thumbnailURL {
		t.Fatalf("expected thumbnail URL %q, got %q", thumbnailURL, resolvedThumbnail)
	}
}

func TestResolveScenarioArtURLsQueuesGenerationWhenOnlyPlaceholderArtExists(t *testing.T) {
	resolvedImage, resolvedThumbnail, shouldGenerate := ResolveScenarioArtURLs(
		ScenarioPlaceholderImageURL,
		"thumbnails/placeholders/scenario-undiscovered.png",
	)
	if !shouldGenerate {
		t.Fatal("expected placeholder art to trigger scenario image generation")
	}
	if resolvedImage != ScenarioPlaceholderImageURL {
		t.Fatalf("expected placeholder image URL %q, got %q", ScenarioPlaceholderImageURL, resolvedImage)
	}
	if resolvedThumbnail != ScenarioPlaceholderImageURL {
		t.Fatalf("expected placeholder thumbnail URL %q, got %q", ScenarioPlaceholderImageURL, resolvedThumbnail)
	}
}

func TestResolveScenarioArtURLsQueuesGenerationWhenArtIsMissing(t *testing.T) {
	resolvedImage, resolvedThumbnail, shouldGenerate := ResolveScenarioArtURLs("", "")
	if !shouldGenerate {
		t.Fatal("expected missing art to trigger scenario image generation")
	}
	if resolvedImage != ScenarioPlaceholderImageURL {
		t.Fatalf("expected placeholder image URL %q, got %q", ScenarioPlaceholderImageURL, resolvedImage)
	}
	if resolvedThumbnail != ScenarioPlaceholderImageURL {
		t.Fatalf("expected placeholder thumbnail URL %q, got %q", ScenarioPlaceholderImageURL, resolvedThumbnail)
	}
}
