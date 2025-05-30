package dungeonmaster

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster/mocks"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/lib/pq" // Added for pq.StringArray
	"github.com/stretchr/testify/assert"
)

// generateQuestCopyInternalFunc and generateQuestImageInternalFunc are defined in client.go

func TestGenerateQuest(t *testing.T) {
	ctx := context.Background()
	testZone := &models.Zone{ID: uuid.New(), Name: "Test Zone"}
	testQuestArchetypeID := uuid.New()
	mockErr := errors.New("mock error")

	originalGenerateQuestCopyFunc := generateQuestCopyInternalFunc
	originalGenerateQuestImageFunc := generateQuestImageInternalFunc
	defer func() {
		generateQuestCopyInternalFunc = originalGenerateQuestCopyFunc
		generateQuestImageInternalFunc = originalGenerateQuestImageFunc
	}()

	defaultLocationArchetypeID := uuid.New()
	defaultQuestArchetypeNodeID := uuid.New()
	defaultQuestArchetypeChallengeID := uuid.New()

	defaultLocationArchetype := &models.LocationArchetype{
		ID:            defaultLocationArchetypeID,
		Name:          "Test Location Archetype",
		IncludedTypes: []googlemaps.PlaceType{googlemaps.PlaceType("park")},
		ExcludedTypes: []googlemaps.PlaceType{},
		Challenges:    pq.StringArray{"Defeat the dragon"}, // Correctly initialize Challenges for GetRandomChallenge
	}

	defaultQuestArchetype := &models.QuestArchetype{
		ID:   testQuestArchetypeID,
		Name: "Test Archetype",
		Root: models.QuestArchetypeNode{
			ID:                  defaultQuestArchetypeNodeID,
			LocationArchetypeID: defaultLocationArchetypeID,
			LocationArchetype:   *defaultLocationArchetype,   // Embed the configured LocationArchetype
			Challenges: []models.QuestArchetypeChallenge{
				{ID: defaultQuestArchetypeChallengeID, Reward: 100},
			},
		},
	}


	defaultPOI := &models.PointOfInterest{
		ID:          uuid.New(),
		Name:        "Test POI",
		Description: "A point of interest",
		Lat:         "1.0",
		Lng:         "1.0",
	}

	defaultQuestGroupID := uuid.New()
	defaultQuestGroup := &models.PointOfInterestGroup{
		ID:                           defaultQuestGroupID,
		Name:                         "Initial Quest Name",
		Description:                  "Initial Quest Description",
		Type:                         models.PointOfInterestGroupTypeQuest,
	}

	testCases := []struct {
		name                   string
		setup                  func(
			dbMock *mocks.MockDbClient,
			locationSeederMock *mocks.MockLocationSeederClient,
			t *testing.T,
		)
		questArchetypeID       uuid.UUID
		zone                   *models.Zone
		expectedErr            error
		expectQuest            bool
		mockGenerateQuestCopy  func(ctx context.Context, locations []string, descriptions []string, challenges []string) (*QuestCopy, error)
		mockGenerateQuestImage func(ctx context.Context, questCopy QuestCopy) (string, error)
	}{
		{
			name: "Successful quest generation",
			setup: func(dbMock *mocks.MockDbClient, locationSeederMock *mocks.MockLocationSeederClient, t *testing.T) {
				questArchetypeStoreMock := dbMock.QuestArchetype().(*mocks.MockQuestArchetypeStore)
				pointOfInterestGroupStoreMock := dbMock.PointOfInterestGroup().(*mocks.MockPointOfInterestGroupStore)
				locationArchetypeStoreMock := dbMock.LocationArchetype().(*mocks.MockLocationArchetypeStore)
				pointOfInterestChallengeStoreMock := dbMock.PointOfInterestChallenge().(*mocks.MockPointOfInterestChallengeStore)

				questArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) {
					// Return a copy to prevent modification across tests if any test modifies the default
					qaCopy := *defaultQuestArchetype
					laCopy := *defaultLocationArchetype
					qaCopy.Root.LocationArchetype = laCopy // Ensure the LA with challenges is used
					return &qaCopy, nil
				}
				pointOfInterestGroupStoreMock.CreateFunc = func(ctx context.Context, name, description, imageUrl string, groupType models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) {
					newGroup := *defaultQuestGroup
					newGroup.Name = name
					newGroup.Description = description
					newGroup.ImageUrl = imageUrl
					return &newGroup, nil
				}
				locationArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error) {
					assert.Equal(t, defaultLocationArchetypeID, id)
					laCopy := *defaultLocationArchetype // Return a copy
					return &laCopy, nil
				}
				locationSeederMock.SeedPointsOfInterestFunc = func(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error) {
					assert.Equal(t, testZone.ID, zone.ID)
					assert.Contains(t, includedTypes, googlemaps.PlaceType("park"))
					return []*models.PointOfInterest{defaultPOI}, nil
				}
				pointOfInterestGroupStoreMock.AddMemberFunc = func(ctx context.Context, poiID uuid.UUID, groupID uuid.UUID) (*models.PointOfInterestGroupMember, error) {
					assert.Equal(t, defaultPOI.ID, poiID)
					assert.Equal(t, defaultQuestGroupID, groupID)
					return &models.PointOfInterestGroupMember{ID: uuid.New(), PointOfInterestID: poiID, PointOfInterestGroupID: groupID}, nil
				}
				pointOfInterestChallengeStoreMock.CreateFunc = func(ctx context.Context, poiID uuid.UUID, tier int, question string, inventoryItemID int, pointOfInterestGroupID *uuid.UUID) (*models.PointOfInterestChallenge, error) {
					assert.Equal(t, defaultPOI.ID, poiID)
					assert.NotNil(t, pointOfInterestGroupID)
					assert.Equal(t, defaultQuestGroupID, *pointOfInterestGroupID)
					assert.Equal(t, "Defeat the dragon", question)
					assert.Equal(t, 100, inventoryItemID)
					return &models.PointOfInterestChallenge{ID: uuid.New(), PointOfInterestID: poiID, Tier: tier, Question: question, InventoryItemID: inventoryItemID, PointOfInterestGroupID: pointOfInterestGroupID}, nil
				}
				pointOfInterestGroupStoreMock.UpdateFunc = func(ctx context.Context, id uuid.UUID, updates *models.PointOfInterestGroup) error {
					assert.Equal(t, defaultQuestGroupID, id)
					assert.Equal(t, "Generated Quest Name", updates.Name)
					assert.Equal(t, "Generated Quest Description", updates.Description)
					assert.Equal(t, "http://image.url", updates.ImageUrl)
					return nil
				}
				pointOfInterestGroupStoreMock.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					t.Errorf("DeleteFunc was called in successful quest generation")
					return nil
				}
			},
			questArchetypeID: testQuestArchetypeID,
			zone:             testZone,
			expectedErr:      nil,
			expectQuest:      true,
			mockGenerateQuestCopy: func(ctx context.Context, locations []string, descriptions []string, challenges []string) (*QuestCopy, error) {
				return &QuestCopy{Name: "Generated Quest Name", Description: "Generated Quest Description"}, nil
			},
			mockGenerateQuestImage: func(ctx context.Context, questCopy QuestCopy) (string, error) {
				return "http://image.url", nil
			},
		},
		{
			name: "Error when dbClient.QuestArchetype().FindByID fails",
			setup: func(dbMock *mocks.MockDbClient, locationSeederMock *mocks.MockLocationSeederClient, t *testing.T) {
				questArchetypeStoreMock := dbMock.QuestArchetype().(*mocks.MockQuestArchetypeStore)
				questArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) {
					return nil, mockErr
				}
			},
			questArchetypeID: testQuestArchetypeID,
			zone:             testZone,
			expectedErr:      mockErr,
			expectQuest:      false,
		},
		{
			name: "Error when dbClient.PointOfInterestGroup().Create fails",
			setup: func(dbMock *mocks.MockDbClient, locationSeederMock *mocks.MockLocationSeederClient, t *testing.T) {
				questArchetypeStoreMock := dbMock.QuestArchetype().(*mocks.MockQuestArchetypeStore)
				questArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) {
					return defaultQuestArchetype, nil
				}
				pointOfInterestGroupStoreMock := dbMock.PointOfInterestGroup().(*mocks.MockPointOfInterestGroupStore)
				pointOfInterestGroupStoreMock.CreateFunc = func(ctx context.Context, name, description, imageUrl string, groupType models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) {
					return nil, mockErr
				}
			},
			questArchetypeID: testQuestArchetypeID,
			zone:             testZone,
			expectedErr:      mockErr,
			expectQuest:      false,
		},
		{
			name: "Error when locationSeeder.SeedPointsOfInterest fails for root node",
			setup: func(dbMock *mocks.MockDbClient, locationSeederMock *mocks.MockLocationSeederClient, t *testing.T) {
				questArchetypeStoreMock := dbMock.QuestArchetype().(*mocks.MockQuestArchetypeStore)
				pointOfInterestGroupStoreMock := dbMock.PointOfInterestGroup().(*mocks.MockPointOfInterestGroupStore)
				locationArchetypeStoreMock := dbMock.LocationArchetype().(*mocks.MockLocationArchetypeStore)

				questArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) {
					return defaultQuestArchetype, nil
				}
				pointOfInterestGroupStoreMock.CreateFunc = func(ctx context.Context, name, description, imageUrl string, groupType models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) {
					return defaultQuestGroup, nil
				}
				locationArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error) {
					return defaultLocationArchetype, nil
				}
				locationSeederMock.SeedPointsOfInterestFunc = func(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error) {
					return nil, mockErr
				}
				pointOfInterestGroupStoreMock.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, defaultQuestGroup.ID, id, "DeleteFunc should be called for the created quest group ID")
					return nil
				}
			},
			questArchetypeID: testQuestArchetypeID,
			zone:             testZone,
			expectedErr:      mockErr,
			expectQuest:      false,
		},
		{
			name: "Scenario where no points of interest are found by locationseeder for root node",
			setup: func(dbMock *mocks.MockDbClient, locationSeederMock *mocks.MockLocationSeederClient, t *testing.T) {
				questArchetypeStoreMock := dbMock.QuestArchetype().(*mocks.MockQuestArchetypeStore)
				pointOfInterestGroupStoreMock := dbMock.PointOfInterestGroup().(*mocks.MockPointOfInterestGroupStore)
				locationArchetypeStoreMock := dbMock.LocationArchetype().(*mocks.MockLocationArchetypeStore)

				questArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) {
					return defaultQuestArchetype, nil
				}
				pointOfInterestGroupStoreMock.CreateFunc = func(ctx context.Context, name, description, imageUrl string, groupType models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) {
					return defaultQuestGroup, nil
				}
				locationArchetypeStoreMock.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error) {
					return defaultLocationArchetype, nil
				}
				locationSeederMock.SeedPointsOfInterestFunc = func(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error) {
					return []*models.PointOfInterest{}, nil
				}
				pointOfInterestGroupStoreMock.DeleteFunc = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, defaultQuestGroup.ID, id)
					return nil
				}
			},
			questArchetypeID: testQuestArchetypeID,
			zone:             testZone,
			expectedErr:      errors.New("no points of interest found"),
			expectQuest:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDbClient := mocks.NewMockDbClient()
			mockLocationSeederClient := &mocks.MockLocationSeederClient{}

			if tc.mockGenerateQuestCopy != nil {
				generateQuestCopyInternalFunc = tc.mockGenerateQuestCopy
			} else {
				generateQuestCopyInternalFunc = func(ctx context.Context, locations []string, descriptions []string, challenges []string) (*QuestCopy, error) {
					return nil, fmt.Errorf("generateQuestCopyInternalFunc not set for test: %s", tc.name)
				}
			}
			if tc.mockGenerateQuestImage != nil {
				generateQuestImageInternalFunc = tc.mockGenerateQuestImage
			} else {
				generateQuestImageInternalFunc = func(ctx context.Context, questCopy QuestCopy) (string, error) {
					return "", fmt.Errorf("generateQuestImageInternalFunc not set for test: %s", tc.name)
				}
			}

			if tc.setup != nil {
				tc.setup(mockDbClient, mockLocationSeederClient, t)
			}

			client := NewClient(nil, mockDbClient, nil, mockLocationSeederClient, nil)

			quest, err := client.GenerateQuest(ctx, tc.zone, tc.questArchetypeID)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			if tc.expectQuest {
				assert.NotNil(t, quest)
				if quest != nil && defaultQuestGroup != nil {
					assert.Equal(t, defaultQuestGroup.ID, quest.ID)
				}
			} else {
				assert.Nil(t, quest)
			}
		})
	}
}
