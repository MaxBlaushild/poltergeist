package processors

import (
	"strings"
	"testing"
)

func TestBuildOutfitPromptFrontView(t *testing.T) {
	prompt := buildOutfitPrompt("a silver knight outfit", "short black hair", "", false)

	if !strings.Contains(prompt, "front-facing") {
		t.Fatalf("expected front-facing prompt, got: %s", prompt)
	}
	if !strings.Contains(prompt, "key facial features") {
		t.Fatalf("expected facial feature guidance, got: %s", prompt)
	}
	if strings.Contains(prompt, "face away from the viewer") {
		t.Fatalf("did not expect back-view guidance in front prompt: %s", prompt)
	}
}

func TestBuildOutfitPromptBackView(t *testing.T) {
	prompt := buildOutfitPrompt("a silver knight outfit", "short black hair", "", true)

	if !strings.Contains(prompt, "back-facing") {
		t.Fatalf("expected back-facing prompt, got: %s", prompt)
	}
	if !strings.Contains(prompt, "face away from the viewer") {
		t.Fatalf("expected explicit back-view direction, got: %s", prompt)
	}
	if !strings.Contains(prompt, "Do not show the face") {
		t.Fatalf("expected no-face constraint, got: %s", prompt)
	}
}

func TestBuildOutfitPromptIncludesGenderPresentation(t *testing.T) {
	prompt := buildOutfitPrompt("a silver knight outfit", "short black hair", "woman", false)

	if !strings.Contains(prompt, "Preserve the person's gender presentation explicitly: woman.") {
		t.Fatalf("expected explicit gender guidance, got: %s", prompt)
	}
}
