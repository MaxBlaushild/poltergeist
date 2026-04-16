package server

import (
	"encoding/base64"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO8G6ioAAAAASUVORK5CYII="

func TestDecodeResourceTypeMapIconPayload(t *testing.T) {
	expected, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "plain base64",
			payload: tinyPNGBase64,
		},
		{
			name:    "data url",
			payload: "data:image/png;base64," + tinyPNGBase64,
		},
		{
			name:    "json payload",
			payload: "{\"data\":[{\"b64_json\":\"" + tinyPNGBase64 + "\"}]}",
		},
		{
			name:    "json array",
			payload: "[\"" + tinyPNGBase64 + "\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeResourceTypeMapIconPayload(tt.payload)
			if err != nil {
				t.Fatalf("decodeResourceTypeMapIconPayload() error = %v", err)
			}
			if string(got) != string(expected) {
				t.Fatalf("decoded bytes mismatch")
			}
		})
	}
}

func TestMatchedResourceTypesForInventoryItem(t *testing.T) {
	herbalismID := uuid.New()
	miningID := uuid.New()
	resourceTypeIndex := buildResourceTypeMatchIndex([]models.ResourceType{
		{
			ID:   herbalismID,
			Name: "Herbalism",
			Slug: "herbalism",
		},
		{
			ID:   miningID,
			Name: "Mining",
			Slug: "mining",
		},
	})

	tests := []struct {
		name        string
		item        models.InventoryItem
		expectedIDs []uuid.UUID
	}{
		{
			name: "matches lower-cased resource type name",
			item: models.InventoryItem{
				InternalTags: models.StringArray{"Herbalism", "crafted"},
			},
			expectedIDs: []uuid.UUID{herbalismID},
		},
		{
			name: "matches normalized slug tag",
			item: models.InventoryItem{
				InternalTags: models.StringArray{"MINING"},
			},
			expectedIDs: []uuid.UUID{miningID},
		},
		{
			name: "returns multiple matches in stable order",
			item: models.InventoryItem{
				InternalTags: models.StringArray{"mining", "herbalism"},
			},
			expectedIDs: []uuid.UUID{herbalismID, miningID},
		},
		{
			name: "returns no matches when tags do not align",
			item: models.InventoryItem{
				InternalTags: models.StringArray{"alchemy"},
			},
			expectedIDs: []uuid.UUID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matchedResourceTypesForInventoryItem(tt.item, resourceTypeIndex)
			ids := make([]uuid.UUID, 0, len(matches))
			for _, match := range matches {
				ids = append(ids, match.ID)
			}
			if diff := cmp.Diff(tt.expectedIDs, ids); diff != "" {
				t.Fatalf("matched resource type ids mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSelectGatherRewardInventoryItem(t *testing.T) {
	herbalismID := uuid.New()
	miningID := uuid.New()

	tests := []struct {
		name         string
		userLevel    int
		items        []models.InventoryItem
		expectedID   int
		expectErr    bool
		resourceType uuid.UUID
	}{
		{
			name:         "prefers items within the level band for the same resource type",
			userLevel:    22,
			resourceType: herbalismID,
			items: []models.InventoryItem{
				{ID: 101, ItemLevel: 15, ResourceTypeID: &herbalismID},
				{ID: 102, ItemLevel: 40, ResourceTypeID: &herbalismID},
				{ID: 201, ItemLevel: 18, ResourceTypeID: &miningID},
				{ID: 103, ItemLevel: 20, ResourceTypeID: &herbalismID, Archived: true},
				{ID: 104, ItemLevel: 19, ResourceTypeID: &herbalismID, RarityTier: "Not Droppable"},
			},
			expectedID: 101,
		},
		{
			name:         "falls back to the closest matching item level when nothing is in band",
			userLevel:    20,
			resourceType: herbalismID,
			items: []models.InventoryItem{
				{ID: 101, ItemLevel: 50, ResourceTypeID: &herbalismID},
				{ID: 102, ItemLevel: 41, ResourceTypeID: &herbalismID},
				{ID: 103, ItemLevel: 3, ResourceTypeID: &herbalismID},
			},
			expectedID: 103,
		},
		{
			name:         "errors when there are no active items for the resource type",
			userLevel:    20,
			resourceType: herbalismID,
			items: []models.InventoryItem{
				{ID: 201, ItemLevel: 18, ResourceTypeID: &miningID},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := selectGatherRewardInventoryItem(tt.resourceType, tt.userLevel, tt.items, nil)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("selectGatherRewardInventoryItem() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("selectGatherRewardInventoryItem() error = %v", err)
			}
			if item == nil {
				t.Fatalf("selectGatherRewardInventoryItem() returned nil item")
			}
			if item.ID != tt.expectedID {
				t.Fatalf("selectGatherRewardInventoryItem() id = %d, want %d", item.ID, tt.expectedID)
			}
		})
	}
}
