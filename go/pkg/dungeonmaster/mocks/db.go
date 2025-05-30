package mocks

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDbClient is a mock implementation of db.DbClient.
type MockDbClient struct {
	MockQuestArchetypeStore           *MockQuestArchetypeStore
	MockPointOfInterestGroupStore     *MockPointOfInterestGroupStore
	MockLocationArchetypeStore        *MockLocationArchetypeStore
	MockPointOfInterestChallengeStore *MockPointOfInterestChallengeStore
	MockQuestArchetypeNodeStore       *MockQuestArchetypeNodeStore
	MockPointOfInterestChildrenStore  *MockPointOfInterestChildrenStore

	BeginFunc func(ctx context.Context) (pgx.Tx, error)
	CloseFunc func()
	ExecFunc  func(ctx context.Context, q string) error
}

func NewMockDbClient() *MockDbClient {
	return &MockDbClient{
		MockQuestArchetypeStore:           &MockQuestArchetypeStore{},
		MockPointOfInterestGroupStore:     &MockPointOfInterestGroupStore{},
		MockLocationArchetypeStore:        &MockLocationArchetypeStore{},
		MockPointOfInterestChallengeStore: &MockPointOfInterestChallengeStore{},
		MockQuestArchetypeNodeStore:       &MockQuestArchetypeNodeStore{},
		MockPointOfInterestChildrenStore:  &MockPointOfInterestChildrenStore{},
	}
}

// Implement db.DbClient interface
func (m *MockDbClient) QuestArchetype() db.QuestArchetypeHandle {
	if m.MockQuestArchetypeStore != nil { return m.MockQuestArchetypeStore }
	return &MockQuestArchetypeStore{}
}
func (m *MockDbClient) PointOfInterestGroup() db.PointOfInterestGroupHandle {
	if m.MockPointOfInterestGroupStore != nil { return m.MockPointOfInterestGroupStore }
	return &MockPointOfInterestGroupStore{}
}
func (m *MockDbClient) LocationArchetype() db.LocationArchetypeHandle {
	if m.MockLocationArchetypeStore != nil { return m.MockLocationArchetypeStore }
	return &MockLocationArchetypeStore{}
}
func (m *MockDbClient) PointOfInterestChallenge() db.PointOfInterestChallengeHandle {
	if m.MockPointOfInterestChallengeStore != nil { return m.MockPointOfInterestChallengeStore }
	return &MockPointOfInterestChallengeStore{}
}
func (m *MockDbClient) QuestArchetypeNode() db.QuestArchetypeNodeHandle {
	if m.MockQuestArchetypeNodeStore != nil { return m.MockQuestArchetypeNodeStore }
	return &MockQuestArchetypeNodeStore{}
}
func (m *MockDbClient) PointOfInterestChildren() db.PointOfInterestChildrenHandle {
	if m.MockPointOfInterestChildrenStore != nil { return m.MockPointOfInterestChildrenStore }
	return &MockPointOfInterestChildrenStore{}
}
func (m *MockDbClient) Exec(ctx context.Context, q string) error {
	if m.ExecFunc != nil { return m.ExecFunc(ctx, q) }
	return nil
}

// Implementations for other DbClient methods returning nil or basic mocks
func (m *MockDbClient) Score() db.ScoreHandle                                     { return nil }
func (m *MockDbClient) User() db.UserHandle                                       { return nil }
func (m *MockDbClient) HowManyQuestion() db.HowManyQuestionHandle                 { return nil }
func (m *MockDbClient) HowManyAnswer() db.HowManyAnswerHandle                     { return nil }
func (m *MockDbClient) Team() db.TeamHandle                                       { return nil }
func (m *MockDbClient) UserTeam() db.UserTeamHandle                               { return nil }
func (m *MockDbClient) PointOfInterest() db.PointOfInterestHandle                 { return nil }
func (m *MockDbClient) NeighboringPointsOfInterest() db.NeighboringPointsOfInterestHandle { return nil }
func (m *MockDbClient) TextVerificationCode() db.TextVerificationCodeHandle         { return nil }
func (m *MockDbClient) SentText() db.SentTextHandle                               { return nil }
func (m *MockDbClient) HowManySubscription() db.HowManySubscriptionHandle           { return nil }
func (m *MockDbClient) SonarSurvey() db.SonarSurveyHandle                         { return nil }
func (m *MockDbClient) SonarSurveySubmission() db.SonarSurveySubmissionHandle       { return nil }
func (m *MockDbClient) SonarActivity() db.SonarActivityHandle                     { return nil }
func (m *MockDbClient) SonarCategory() db.SonarCategoryHandle                     { return nil }
func (m *MockDbClient) Match() db.MatchHandle                                     { return nil }
func (m *MockDbClient) VerificationCode() db.VerificationCodeHandle               { return nil }
func (m *MockDbClient) InventoryItem() db.InventoryItemHandle                     { return nil }
func (m *MockDbClient) AuditItem() db.AuditItemHandle                             { return nil }
func (m *MockDbClient) ImageGeneration() db.ImageGenerationHandle                 { return nil }
func (m *MockDbClient) PointOfInterestDiscovery() db.PointOfInterestDiscoveryHandle   { return nil }
func (m *MockDbClient) MatchUser() db.MatchUserHandle                             { return nil }
func (m *MockDbClient) Tag() db.TagHandle                                         { return nil }
func (m *MockDbClient) TagGroup() db.TagGroupHandle                               { return nil }
func (m *MockDbClient) Zone() db.ZoneHandle                                       { return nil }
func (m *MockDbClient) QuestArchetypeChallenge() db.QuestArchetypeChallengeHandle     { return nil }
func (m *MockDbClient) QuestArchetypeNodeChallenge() db.QuestArchetypeNodeChallengeHandle { return nil }
func (m *MockDbClient) ZoneQuestArchetype() db.ZoneQuestArchetypeHandle             { return nil }
func (m *MockDbClient) TrackedPointOfInterestGroup() db.TrackedPointOfInterestGroupHandle { return nil }
func (m *MockDbClient) Point() db.PointHandle                                     { return nil }
func (m *MockDbClient) UserLevel() db.UserLevelHandle                             { return nil }
func (m *MockDbClient) UserZoneReputation() db.UserZoneReputationHandle             { return nil }

// --- MockQuestArchetypeStore ---
type MockQuestArchetypeStore struct {
	CreateFunc   func(ctx context.Context, questArchetype *models.QuestArchetype) error
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error)
	FindAllFunc  func(ctx context.Context) ([]*models.QuestArchetype, error)
	UpdateFunc   func(ctx context.Context, questArchetype *models.QuestArchetype) error
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}
func (m *MockQuestArchetypeStore) Create(ctx context.Context, qa *models.QuestArchetype) error { if m.CreateFunc != nil { return m.CreateFunc(ctx, qa) }; return nil }
func (m *MockQuestArchetypeStore) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error) { if m.FindByIDFunc != nil { return m.FindByIDFunc(ctx, id) }; return &models.QuestArchetype{}, nil }
func (m *MockQuestArchetypeStore) FindAll(ctx context.Context) ([]*models.QuestArchetype, error) { if m.FindAllFunc != nil { return m.FindAllFunc(ctx) }; return []*models.QuestArchetype{}, nil }
func (m *MockQuestArchetypeStore) Update(ctx context.Context, qa *models.QuestArchetype) error { if m.UpdateFunc != nil { return m.UpdateFunc(ctx, qa) }; return nil }
func (m *MockQuestArchetypeStore) Delete(ctx context.Context, id uuid.UUID) error { if m.DeleteFunc != nil { return m.DeleteFunc(ctx, id) }; return nil }

// --- MockPointOfInterestGroupStore ---
type MockPointOfInterestGroupStore struct {
	CreateFunc      func(ctx context.Context, name string, description string, imageUrl string, typeValue models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error)
	FindByIDFunc    func(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error)
	FindAllFunc     func(ctx context.Context) ([]*models.PointOfInterestGroup, error)
	EditFunc        func(ctx context.Context, id uuid.UUID, name string, description string, typeValue models.PointOfInterestGroupType) error
	DeleteFunc      func(ctx context.Context, id uuid.UUID) error
	UpdateImageUrlFunc func(ctx context.Context, id uuid.UUID, imageUrl string) error
	FindByTypeFunc  func(ctx context.Context, typeValue models.PointOfInterestGroupType) ([]*models.PointOfInterestGroup, error)
	GetNearbyQuestsFunc func(ctx context.Context, userID uuid.UUID, lat float64, lng float64, radiusInMeters float64, tags []string) ([]models.PointOfInterestGroup, error)
	GetStartedQuestsFunc func(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error)
	AddMemberFunc   func(ctx context.Context, pointOfInterestID uuid.UUID, pointOfInterestGroupID uuid.UUID) (*models.PointOfInterestGroupMember, error)
	UpdateFunc      func(ctx context.Context, pointOfInterestGroupID uuid.UUID, updates *models.PointOfInterestGroup) error
	FindByIDsFunc   func(ctx context.Context, ids []uuid.UUID) ([]models.PointOfInterestGroup, error)
	GetQuestsInZoneFunc func(ctx context.Context, zoneID uuid.UUID, tags []string) ([]models.PointOfInterestGroup, error)
}
func (m *MockPointOfInterestGroupStore) Create(ctx context.Context, name string, description string, imageUrl string, typeValue models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) { if m.CreateFunc != nil { return m.CreateFunc(ctx, name, description, imageUrl, typeValue) }; return &models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error) { if m.FindByIDFunc != nil { return m.FindByIDFunc(ctx, id) }; return &models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error) { if m.FindAllFunc != nil { return m.FindAllFunc(ctx) }; return []*models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) Edit(ctx context.Context, id uuid.UUID, name string, description string, typeValue models.PointOfInterestGroupType) error { if m.EditFunc != nil { return m.EditFunc(ctx, id, name, description, typeValue) }; return nil }
func (m *MockPointOfInterestGroupStore) Delete(ctx context.Context, id uuid.UUID) error { if m.DeleteFunc != nil { return m.DeleteFunc(ctx, id) }; return nil }
func (m *MockPointOfInterestGroupStore) UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error { if m.UpdateImageUrlFunc != nil { return m.UpdateImageUrlFunc(ctx, id, imageUrl) }; return nil }
func (m *MockPointOfInterestGroupStore) FindByType(ctx context.Context, typeValue models.PointOfInterestGroupType) ([]*models.PointOfInterestGroup, error) { if m.FindByTypeFunc != nil { return m.FindByTypeFunc(ctx, typeValue) }; return []*models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) GetNearbyQuests(ctx context.Context, userID uuid.UUID, lat float64, lng float64, radiusInMeters float64, tags []string) ([]models.PointOfInterestGroup, error) { if m.GetNearbyQuestsFunc != nil { return m.GetNearbyQuestsFunc(ctx, userID, lat, lng, radiusInMeters, tags) }; return []models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) { if m.GetStartedQuestsFunc != nil { return m.GetStartedQuestsFunc(ctx, userID) }; return []models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) AddMember(ctx context.Context, pointOfInterestID uuid.UUID, pointOfInterestGroupID uuid.UUID) (*models.PointOfInterestGroupMember, error) { if m.AddMemberFunc != nil { return m.AddMemberFunc(ctx, pointOfInterestID, pointOfInterestGroupID) }; return &models.PointOfInterestGroupMember{}, nil }
func (m *MockPointOfInterestGroupStore) Update(ctx context.Context, pointOfInterestGroupID uuid.UUID, updates *models.PointOfInterestGroup) error { if m.UpdateFunc != nil { return m.UpdateFunc(ctx, pointOfInterestGroupID, updates) }; return nil }
func (m *MockPointOfInterestGroupStore) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.PointOfInterestGroup, error) { if m.FindByIDsFunc != nil { return m.FindByIDsFunc(ctx, ids) }; return []models.PointOfInterestGroup{}, nil }
func (m *MockPointOfInterestGroupStore) GetQuestsInZone(ctx context.Context, zoneID uuid.UUID, tags []string) ([]models.PointOfInterestGroup, error) { if m.GetQuestsInZoneFunc != nil { return m.GetQuestsInZoneFunc(ctx, zoneID, tags) }; return []models.PointOfInterestGroup{}, nil }

// --- MockLocationArchetypeStore ---
type MockLocationArchetypeStore struct {
	CreateFunc   func(ctx context.Context, locationArchetype *models.LocationArchetype) error
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error)
	FindAllFunc  func(ctx context.Context) ([]*models.LocationArchetype, error)
	UpdateFunc   func(ctx context.Context, locationArchetype *models.LocationArchetype) error
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}
func (m *MockLocationArchetypeStore) Create(ctx context.Context, la *models.LocationArchetype) error { if m.CreateFunc != nil { return m.CreateFunc(ctx, la) }; return nil }
func (m *MockLocationArchetypeStore) FindByID(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error) { if m.FindByIDFunc != nil { return m.FindByIDFunc(ctx, id) }; return &models.LocationArchetype{}, nil }
func (m *MockLocationArchetypeStore) FindAll(ctx context.Context) ([]*models.LocationArchetype, error) { if m.FindAllFunc != nil { return m.FindAllFunc(ctx) }; return []*models.LocationArchetype{}, nil }
func (m *MockLocationArchetypeStore) Update(ctx context.Context, la *models.LocationArchetype) error { if m.UpdateFunc != nil { return m.UpdateFunc(ctx, la) }; return nil }
func (m *MockLocationArchetypeStore) Delete(ctx context.Context, id uuid.UUID) error { if m.DeleteFunc != nil { return m.DeleteFunc(ctx, id) }; return nil }

// --- MockPointOfInterestChallengeStore ---
type MockPointOfInterestChallengeStore struct {
	CreateFunc                     func(ctx context.Context, pointOfInterestID uuid.UUID, tier int, question string, inventoryItemID int, pointOfInterestGroupID *uuid.UUID) (*models.PointOfInterestChallenge, error)
	SubmitAnswerForChallengeFunc   func(ctx context.Context, challengeID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error)
	FindByIDFunc                   func(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error)
	GetChallengeForPointOfInterestFunc func(ctx context.Context, pointOfInterestID uuid.UUID, tier int) (*models.PointOfInterestChallenge, error)
	DeleteFunc                     func(ctx context.Context, id uuid.UUID) error
	EditFunc                       func(ctx context.Context, id uuid.UUID, question string, inventoryItemID int, tier int) (*models.PointOfInterestChallenge, error)
	GetSubmissionsForMatchFunc     func(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error)
	GetSubmissionsForUserFunc      func(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error)
	DeleteAllForPointOfInterestFunc func(ctx context.Context, pointOfInterestID uuid.UUID) error
	GetChildrenForChallengeFunc    func(ctx context.Context, challengeID uuid.UUID) ([]models.PointOfInterestChildren, error)
}
func (m *MockPointOfInterestChallengeStore) Create(ctx context.Context, pointOfInterestID uuid.UUID, tier int, question string, inventoryItemID int, pointOfInterestGroupID *uuid.UUID) (*models.PointOfInterestChallenge, error) { if m.CreateFunc != nil { return m.CreateFunc(ctx, pointOfInterestID, tier, question, inventoryItemID, pointOfInterestGroupID) }; return &models.PointOfInterestChallenge{}, nil }
func (m *MockPointOfInterestChallengeStore) SubmitAnswerForChallenge(ctx context.Context, challengeID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error) { if m.SubmitAnswerForChallengeFunc != nil { return m.SubmitAnswerForChallengeFunc(ctx, challengeID, teamID, userID, text, imageURL, isCorrect) }; return &models.PointOfInterestChallengeSubmission{}, nil }
func (m *MockPointOfInterestChallengeStore) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error) { if m.FindByIDFunc != nil { return m.FindByIDFunc(ctx, id) }; return &models.PointOfInterestChallenge{}, nil }
func (m *MockPointOfInterestChallengeStore) GetChallengeForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID, tier int) (*models.PointOfInterestChallenge, error) { if m.GetChallengeForPointOfInterestFunc != nil { return m.GetChallengeForPointOfInterestFunc(ctx, pointOfInterestID, tier) }; return &models.PointOfInterestChallenge{}, nil }
func (m *MockPointOfInterestChallengeStore) Delete(ctx context.Context, id uuid.UUID) error { if m.DeleteFunc != nil { return m.DeleteFunc(ctx, id) }; return nil }
func (m *MockPointOfInterestChallengeStore) Edit(ctx context.Context, id uuid.UUID, question string, inventoryItemID int, tier int) (*models.PointOfInterestChallenge, error) { if m.EditFunc != nil { return m.EditFunc(ctx, id, question, inventoryItemID, tier) }; return &models.PointOfInterestChallenge{}, nil }
func (m *MockPointOfInterestChallengeStore) GetSubmissionsForMatch(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error) { if m.GetSubmissionsForMatchFunc != nil { return m.GetSubmissionsForMatchFunc(ctx, matchID) }; return []models.PointOfInterestChallengeSubmission{}, nil }
func (m *MockPointOfInterestChallengeStore) GetSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error) { if m.GetSubmissionsForUserFunc != nil { return m.GetSubmissionsForUserFunc(ctx, userID) }; return []models.PointOfInterestChallengeSubmission{}, nil }
func (m *MockPointOfInterestChallengeStore) DeleteAllForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID) error { if m.DeleteAllForPointOfInterestFunc != nil { return m.DeleteAllForPointOfInterestFunc(ctx, pointOfInterestID) }; return nil }
func (m *MockPointOfInterestChallengeStore) GetChildrenForChallenge(ctx context.Context, challengeID uuid.UUID) ([]models.PointOfInterestChildren, error) { if m.GetChildrenForChallengeFunc != nil { return m.GetChildrenForChallengeFunc(ctx, challengeID) }; return []models.PointOfInterestChildren{}, nil }

// --- MockQuestArchetypeNodeStore ---
type MockQuestArchetypeNodeStore struct {
	CreateFunc   func(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNode, error)
	FindAllFunc  func(ctx context.Context) ([]*models.QuestArchetypeNode, error)
	UpdateFunc   func(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}
func (m *MockQuestArchetypeNodeStore) Create(ctx context.Context, qan *models.QuestArchetypeNode) error { if m.CreateFunc != nil { return m.CreateFunc(ctx, qan) }; return nil }
func (m *MockQuestArchetypeNodeStore) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNode, error) { if m.FindByIDFunc != nil { return m.FindByIDFunc(ctx, id) }; return &models.QuestArchetypeNode{}, nil }
func (m *MockQuestArchetypeNodeStore) FindAll(ctx context.Context) ([]*models.QuestArchetypeNode, error) { if m.FindAllFunc != nil { return m.FindAllFunc(ctx) }; return []*models.QuestArchetypeNode{}, nil }
func (m *MockQuestArchetypeNodeStore) Update(ctx context.Context, qan *models.QuestArchetypeNode) error { if m.UpdateFunc != nil { return m.UpdateFunc(ctx, qan) }; return nil }
func (m *MockQuestArchetypeNodeStore) Delete(ctx context.Context, id uuid.UUID) error { if m.DeleteFunc != nil { return m.DeleteFunc(ctx, id) }; return nil }

// --- MockPointOfInterestChildrenStore ---
type MockPointOfInterestChildrenStore struct {
	CreateFunc func(ctx context.Context, pointOfInterestGroupMemberID uuid.UUID, nextPointOfInterestGroupMemberID uuid.UUID, pointOfInterestChallengeID uuid.UUID) error
	DeleteFunc func(ctx context.Context, id uuid.UUID) error
}
func (m *MockPointOfInterestChildrenStore) Create(ctx context.Context, pointOfInterestGroupMemberID uuid.UUID, nextPointOfInterestGroupMemberID uuid.UUID, pointOfInterestChallengeID uuid.UUID) error { if m.CreateFunc != nil { return m.CreateFunc(ctx, pointOfInterestGroupMemberID, nextPointOfInterestGroupMemberID, pointOfInterestChallengeID) }; return nil }
func (m *MockPointOfInterestChildrenStore) Delete(ctx context.Context, id uuid.UUID) error { if m.DeleteFunc != nil { return m.DeleteFunc(ctx, id) }; return nil }

// --- MockTx ---
type MockTx struct{
	CurrentMockLargeObjects pgx.LargeObjects // Field to hold a LargeObjects mock
}
func (mt *MockTx) Begin(ctx context.Context) (pgx.Tx, error) { return mt, nil }
func (mt *MockTx) Commit(ctx context.Context) error          { return nil }
func (mt *MockTx) Rollback(ctx context.Context) error        { return nil }
func (mt *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) { return 0, nil }
func (mt *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (mt *MockTx) LargeObjects() pgx.LargeObjects                             { return mt.CurrentMockLargeObjects } // Return the field
func (mt *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) { return nil, nil }
func (mt *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil }
func (mt *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) { return &MockRows{}, nil }
func (mt *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row { return &MockRow{} }
func (mt *MockTx) Conn() *pgx.Conn { return nil }

// --- MockRows ---
type MockRows struct{}
func (mr *MockRows) Close()                                {}
func (mr *MockRows) Err() error                              { return nil }
func (mr *MockRows) CommandTag() pgconn.CommandTag           { return pgconn.CommandTag{} }
func (mr *MockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (mr *MockRows) Next() bool                              { return false }
func (mr *MockRows) Scan(dest ...interface{}) error          { return nil }
func (mr *MockRows) Values() ([]interface{}, error)          { return nil, nil }
func (mr *MockRows) RawValues() [][]byte                     { return nil }
func (mr *MockRows) Conn() *pgx.Conn                         { return nil }

// --- MockRow ---
type MockRow struct{}
func (mr *MockRow) Scan(dest ...interface{}) error { return nil }

// MockLargeObjects struct is removed.
// Static assertion for MockLargeObjects is removed.
