package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type DbClient interface {
	Score() ScoreHandle
	User() UserHandle
	HowManyQuestion() HowManyQuestionHandle
	HowManyAnswer() HowManyAnswerHandle
	Team() TeamHandle
	UserTeam() UserTeamHandle
	PointOfInterest() PointOfInterestHandle
	NeighboringPointsOfInterest() NeighboringPointsOfInterestHandle
	TextVerificationCode() TextVerificationCodeHandle
	SentText() SentTextHandle
	HowManySubscription() HowManySubscriptionHandle
	SonarSurvey() SonarSurveyHandle
	SonarSurveySubmission() SonarSurveySubmissionHandle
	SonarActivity() SonarActivityHandle
	SonarCategory() SonarCategoryHandle
	Match() MatchHandle
	VerificationCode() VerificationCodeHandle
	PointOfInterestGroup() PointOfInterestGroupHandle
	PointOfInterestChallenge() PointOfInterestChallengeHandle
	InventoryItem() InventoryItemHandle
	AuditItem() AuditItemHandle
	ImageGeneration() ImageGenerationHandle
	PointOfInterestChildren() PointOfInterestChildrenHandle
	PointOfInterestDiscovery() PointOfInterestDiscoveryHandle
	MatchUser() MatchUserHandle
	Tag() TagHandle
	TagGroup() TagGroupHandle
	Zone() ZoneHandle
	LocationArchetype() LocationArchetypeHandle
	QuestArchetype() QuestArchetypeHandle
	QuestArchetypeNode() QuestArchetypeNodeHandle
	QuestArchetypeChallenge() QuestArchetypeChallengeHandle
	QuestArchetypeNodeChallenge() QuestArchetypeNodeChallengeHandle
	ZoneQuestArchetype() ZoneQuestArchetypeHandle
	TrackedPointOfInterestGroup() TrackedPointOfInterestGroupHandle
	Point() PointHandle
	UserLevel() UserLevelHandle
	UserZoneReputation() UserZoneReputationHandle
	Friend() FriendHandle
	Party() PartyHandle
	FriendInvite() FriendInviteHandle
	PartyInvite() PartyInviteHandle
	Activity() ActivityHandle
	PointOfInterestGroupMember() PointOfInterestGroupMemberHandle
	Character() CharacterHandle
	CharacterAction() CharacterActionHandle
	QuestAcceptance() QuestAcceptanceHandle
	MovementPattern() MovementPatternHandle
	TreasureChest() TreasureChestHandle
	Document() DocumentHandle
	DocumentTag() DocumentTagHandle
	GoogleDriveToken() GoogleDriveTokenHandle
	DropboxToken() DropboxTokenHandle
	HueToken() HueTokenHandle
	FeteRoom() FeteRoomHandle
	FeteTeam() FeteTeamHandle
	FeteRoomLinkedListTeam() FeteRoomLinkedListTeamHandle
	Exec(ctx context.Context, q string) error
}

type ScoreHandle interface {
	Upsert(ctx context.Context, username string) (*models.Score, error)
	FindAll(ctx context.Context) ([]models.Score, error)
}

type HowManyAnswerHandle interface {
	FindByQuestionIDAndUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.HowManyAnswer, error)
	Insert(ctx context.Context, a *models.HowManyAnswer) (*models.HowManyAnswer, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type HowManyQuestionHandle interface {
	Insert(ctx context.Context, text string, explanation string, howMany int, promptSeedIndex int, prompt string) (*models.HowManyQuestion, error)
	FindAll(ctx context.Context) ([]*models.HowManyQuestion, error)
	MarkValid(ctx context.Context, howManyQuestionID uuid.UUID) error
	MarkDone(ctx context.Context, howManyQuestionID uuid.UUID) error
	FindTodaysQuestion(ctx context.Context) (*models.HowManyQuestion, error)
	FindById(ctx context.Context, id uuid.UUID) (*models.HowManyQuestion, error)
	ValidQuestionsRemaining(ctx context.Context) (int64, error)
}

type UserHandle interface {
	Insert(ctx context.Context, name string, phoneNumber string, id *uuid.UUID, username *string, dateOfBirth *time.Time, gender *string, latitude *float64, longitude *float64, locationAddress *string, bio *string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error)
	FindUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	DeleteAll(ctx context.Context) error
	UpdateProfilePictureUrl(ctx context.Context, userID uuid.UUID, url string) error
	UpdateHasSeenTutorial(ctx context.Context, userID uuid.UUID, hasSeenTutorial bool) error
	JoinParty(ctx context.Context, inviterID uuid.UUID, inviteeID uuid.UUID) error
	FindPartyMembers(ctx context.Context, userID uuid.UUID) ([]models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindLikeByUsername(ctx context.Context, username string) ([]*models.User, error)
	SetUsername(ctx context.Context, userID uuid.UUID, username string) error
	Update(ctx context.Context, userID uuid.UUID, updates models.User) error
	LeaveParty(ctx context.Context, userID uuid.UUID) error
	AddGold(ctx context.Context, userID uuid.UUID, amount int) error
	SetGold(ctx context.Context, userID uuid.UUID, amount int) error
	SubtractGold(ctx context.Context, userID uuid.UUID, amount int) error
	AddCredits(ctx context.Context, userID uuid.UUID, amount int) error
	SetCredits(ctx context.Context, userID uuid.UUID, amount int) error
	SubtractCredits(ctx context.Context, userID uuid.UUID, amount int) error
}

type TeamHandle interface {
	GetAll(ctx context.Context) ([]models.Team, error)
	Create(ctx context.Context, userIDs []uuid.UUID, teamName string, matchID uuid.UUID) (*models.Team, error)
	AddUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
	RemoveUserFromMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) error
	UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) (*models.Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error)
	GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.Team, error)
}

type UserTeamHandle interface {
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PointOfInterestHandle interface {
	FindAll(ctx context.Context) ([]models.PointOfInterest, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterest, error)
	Create(ctx context.Context, crystal models.PointOfInterest) error
	Unlock(ctx context.Context, pointOfInterestID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID) error
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.PointOfInterest, error)
	Edit(ctx context.Context, id uuid.UUID, name string, description string, lat string, lng string, unlockTier *int) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error
	CreateForGroup(ctx context.Context, pointOfInterest *models.PointOfInterest, pointOfInterestGroupID uuid.UUID) error
	FindAllForZone(ctx context.Context, zoneID uuid.UUID) ([]models.PointOfInterest, error)
	FindByGoogleMapsPlaceID(ctx context.Context, googleMapsPlaceID string) (*models.PointOfInterest, error)
	Update(ctx context.Context, pointOfInterestID uuid.UUID, updates *models.PointOfInterest) error
	FindZoneForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID) (*models.PointOfInterestZone, error)
	UpdateLastUsedInQuest(ctx context.Context, pointOfInterestID uuid.UUID) error
	FindRecentlyUsedInZone(ctx context.Context, zoneID uuid.UUID, since time.Time) (map[string]bool, error)
}

type NeighboringPointsOfInterestHandle interface {
	Create(ctx context.Context, crystalOneID uuid.UUID, crystalTwoID uuid.UUID) error
	FindAll(ctx context.Context) ([]models.NeighboringPointsOfInterest, error)
}

type TextVerificationCodeHandle interface {
	Insert(ctx context.Context, phoneNumber string) (*models.TextVerificationCode, error)
	Find(ctx context.Context, phoneNumber string, code string) (*models.TextVerificationCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

type SentTextHandle interface {
	GetCount(ctx context.Context, phoneNumber string, textType string) (int64, error)
	Insert(ctx context.Context, textType string, phoneNumber string, text string) (*models.SentText, error)
}

type HowManySubscriptionHandle interface {
	Insert(ctx context.Context, userID uuid.UUID) (*models.HowManySubscription, error)
	FindAll(ctx context.Context) ([]models.HowManySubscription, error)
	IncrementNumFreeQuestions(ctx context.Context, userID uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.HowManySubscription, error)
	SetSubscribed(ctx context.Context, userID uuid.UUID, stripeID string) error
	DeleteByStripeID(ctx context.Context, stripeID string) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type SonarSurveyHandle interface {
	GetSurveys(ctx context.Context, userID uuid.UUID) ([]models.SonarSurvey, error)
	CreateSurvey(ctx context.Context, userID uuid.UUID, title string, activityIDs []uuid.UUID) (*models.SonarSurvey, error)
	GetSurveyByID(ctx context.Context, surveyID uuid.UUID) (*models.SonarSurvey, error)
}

type SonarSurveySubmissionHandle interface {
	CreateSubmission(ctx context.Context, surveryID uuid.UUID, userID uuid.UUID, activityIDS []uuid.UUID, downs []bool) (*models.SonarSurveySubmission, error)
	GetUserSubmissionForSurvey(ctx context.Context, userID uuid.UUID, surveyID uuid.UUID) (*models.SonarSurveySubmission, error)
	GetAllSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.SonarSurveySubmission, error)
	GetSubmissionByID(ctx context.Context, submissionID uuid.UUID) (*models.SonarSurveySubmission, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type SonarActivityHandle interface {
	GetAllActivities(ctx context.Context) ([]models.SonarActivity, error)
	CreateActivity(ctx context.Context, activity models.SonarActivity) (models.SonarActivity, error)
	UpdateActivity(ctx context.Context, activity models.SonarActivity) (models.SonarActivity, error)
	DeleteActivity(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetActivityByID(ctx context.Context, id uuid.UUID) (models.SonarActivity, error)
}

type SonarCategoryHandle interface {
	GetCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]models.SonarCategory, error)
	GetAllCategoriesWithActivities(ctx context.Context) ([]models.SonarCategory, error)
	CreateCategory(ctx context.Context, category models.SonarCategory) (models.SonarCategory, error)
	UpdateCategory(ctx context.Context, category models.SonarCategory) (models.SonarCategory, error)
	DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetCategoryByID(ctx context.Context, id uuid.UUID) (models.SonarCategory, error)
}

type MatchHandle interface {
	Create(ctx context.Context, creatorID uuid.UUID, pointsOfInterestIDs []uuid.UUID) (*models.Match, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Match, error)
	StartMatch(ctx context.Context, matchID uuid.UUID) error
	EndMatch(ctx context.Context, matchID uuid.UUID) error
	FindCurrentMatchForUser(ctx context.Context, userId uuid.UUID) (*models.Match, error)
	FindForTeamID(ctx context.Context, teamID uuid.UUID) (*models.TeamMatch, error)
	FindCurrentMatchIDForUser(ctx context.Context, userId uuid.UUID) (*uuid.UUID, error)
}

type VerificationCodeHandle interface {
	Create(ctx context.Context) (*models.VerificationCode, error)
}

type PointOfInterestGroupHandle interface {
	Create(ctx context.Context, name string, description string, imageUrl string, typeValue models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error)
	FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error)
	Edit(ctx context.Context, id uuid.UUID, name string, description string, typeValue models.PointOfInterestGroupType) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByIDs(ctx context.Context, ids []uuid.UUID) error
	UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error
	FindByType(ctx context.Context, typeValue models.PointOfInterestGroupType) ([]*models.PointOfInterestGroup, error)
	GetNearbyQuests(ctx context.Context, userID uuid.UUID, lat float64, lng float64, radiusInMeters float64, tags []string) ([]models.PointOfInterestGroup, error)
	GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error)
	AddMember(ctx context.Context, pointOfInterestID uuid.UUID, pointOfInterestGroupID uuid.UUID) (*models.PointOfInterestGroupMember, error)
	Update(ctx context.Context, pointOfInterestGroupID uuid.UUID, updates *models.PointOfInterestGroup) error
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.PointOfInterestGroup, error)
	GetQuestsInZone(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) ([]models.PointOfInterestGroup, error)
}

type PointOfInterestChallengeHandle interface {
	Create(ctx context.Context, pointOfInterestID uuid.UUID, tier int, question string, inventoryItemID int, pointOfInterestGroupID *uuid.UUID) (*models.PointOfInterestChallenge, error)
	SubmitAnswerForChallenge(ctx context.Context, challengeID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error)
	GetChallengeForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID, tier int) (*models.PointOfInterestChallenge, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Edit(ctx context.Context, id uuid.UUID, question string, inventoryItemID int, tier int) (*models.PointOfInterestChallenge, error)
	GetSubmissionsForMatch(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error)
	GetSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error)
	DeleteAllForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID) error
	GetChildrenForChallenge(ctx context.Context, challengeID uuid.UUID) ([]models.PointOfInterestChildren, error)
	DeleteSubmission(ctx context.Context, submissionID uuid.UUID) error
	DeleteAllSubmissionsForUser(ctx context.Context, userID uuid.UUID) error
}

type InventoryItemHandle interface {
	CreateOrIncrementInventoryItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, inventoryItemID int, quantity int) error
	UseInventoryItem(ctx context.Context, ownedInventoryItemID uuid.UUID) error
	ApplyInventoryItem(ctx context.Context, matchID uuid.UUID, inventoryItemID int, teamID uuid.UUID, duration time.Duration) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.OwnedInventoryItem, error)
	StealItems(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID) error
	GetItems(ctx context.Context, userOrTeam models.OwnedInventoryItem) ([]models.OwnedInventoryItem, error)
	StealItem(ctx context.Context, thiefTeamID uuid.UUID, victimTeamID uuid.UUID, inventoryItemID int) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
	DecrementUserInventoryItem(ctx context.Context, userID uuid.UUID, inventoryItemID int, quantity int) error
	// CRUD methods for inventory items
	CreateInventoryItem(ctx context.Context, item *models.InventoryItem) error
	FindInventoryItemByID(ctx context.Context, id int) (*models.InventoryItem, error)
	FindAllInventoryItems(ctx context.Context) ([]models.InventoryItem, error)
	UpdateInventoryItem(ctx context.Context, id int, item *models.InventoryItem) error
	DeleteInventoryItem(ctx context.Context, id int) error
}

type AuditItemHandle interface {
	Create(ctx context.Context, matchID *uuid.UUID, userID *uuid.UUID, message string) error
	GetAuditItemsForMatch(ctx context.Context, matchID uuid.UUID) ([]*models.AuditItem, error)
	GetAuditItemsForUser(ctx context.Context, userID uuid.UUID) ([]*models.AuditItem, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type ImageGenerationHandle interface {
	Create(ctx context.Context, imageGeneration *models.ImageGeneration) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.ImageGeneration, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.ImageGeneration, error)
	UpdateState(ctx context.Context, imageGenerationID uuid.UUID, state models.GenerationStatus) error
	FindByState(ctx context.Context, state models.GenerationStatus) ([]models.ImageGeneration, error)
	SetOptions(ctx context.Context, imageGenerationID uuid.UUID, options []string) error
	Updates(ctx context.Context, imageGenerationID uuid.UUID, updates *models.ImageGeneration) error
	GetCompleteGenerationsForUser(ctx context.Context, userID uuid.UUID) ([]models.ImageGeneration, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PointOfInterestChildrenHandle interface {
	Create(ctx context.Context, pointOfInterestGroupMemberID uuid.UUID, nextPointOfInterestGroupMemberID uuid.UUID, pointOfInterestChallengeID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PointOfInterestDiscoveryHandle interface {
	GetDiscoveriesForTeam(teamID uuid.UUID) ([]models.PointOfInterestDiscovery, error)
	GetDiscoveriesForUser(userID uuid.UUID) ([]models.PointOfInterestDiscovery, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, discovery *models.PointOfInterestDiscovery) error
}

type MatchUserHandle interface {
	Create(ctx context.Context, matchUser *models.MatchUser) error
	FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.MatchUser, error)
	FindUsersForMatch(ctx context.Context, matchID uuid.UUID) ([]models.User, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type TagHandle interface {
	FindAll(ctx context.Context) ([]*models.Tag, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Tag, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*models.Tag, error)
	Create(ctx context.Context, tag *models.Tag) error
	Update(ctx context.Context, tag *models.Tag) error
	AddTagToPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error
	RemoveTagFromPointOfInterest(ctx context.Context, tagID uuid.UUID, pointOfInterestID uuid.UUID) error
	Upsert(ctx context.Context, tag *models.Tag) error
	FindByValue(ctx context.Context, value string) (*models.Tag, error)
	MoveTagToTagGroup(ctx context.Context, tagID uuid.UUID, tagGroupID uuid.UUID) error
	CreateTagGroup(ctx context.Context, tagGroup *models.TagGroup) error
}

type TagGroupHandle interface {
	FindAll(ctx context.Context) ([]*models.TagGroup, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.TagGroup, error)
	Create(ctx context.Context, tagGroup *models.TagGroup) error
	Update(ctx context.Context, tagGroup *models.TagGroup) error
	FindByName(ctx context.Context, name string) (*models.TagGroup, error)
}

type ZoneHandle interface {
	Create(ctx context.Context, zone *models.Zone) error
	FindAll(ctx context.Context) ([]*models.Zone, error)
	Update(ctx context.Context, zone *models.Zone) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Zone, error)
	Delete(ctx context.Context, zoneID uuid.UUID) error
	AddPointOfInterestToZone(ctx context.Context, zoneID uuid.UUID, pointOfInterestID uuid.UUID) error
	RemovePointOfInterestFromZone(ctx context.Context, zoneID uuid.UUID, pointOfInterestID uuid.UUID) error
	UpdateBoundary(ctx context.Context, zoneID uuid.UUID, boundary [][]float64) error
	UpdateNameAndDescription(ctx context.Context, zoneID uuid.UUID, name string, description string) error
	FindByPointOfInterestID(ctx context.Context, pointOfInterestID uuid.UUID) (*models.Zone, error)
}

type LocationArchetypeHandle interface {
	Create(ctx context.Context, locationArchetype *models.LocationArchetype) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.LocationArchetype, error)
	FindAll(ctx context.Context) ([]*models.LocationArchetype, error)
	Update(ctx context.Context, locationArchetype *models.LocationArchetype) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuestArchetypeHandle interface {
	Create(ctx context.Context, questArchetype *models.QuestArchetype) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetype, error)
	FindAll(ctx context.Context) ([]*models.QuestArchetype, error)
	Update(ctx context.Context, questArchetype *models.QuestArchetype) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuestArchetypeNodeHandle interface {
	Create(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNode, error)
	FindAll(ctx context.Context) ([]*models.QuestArchetypeNode, error)
	Update(ctx context.Context, questArchetypeNode *models.QuestArchetypeNode) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuestArchetypeChallengeHandle interface {
	Create(ctx context.Context, questArchetypeChallenge *models.QuestArchetypeChallenge) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeChallenge, error)
	FindAll(ctx context.Context) ([]*models.QuestArchetypeChallenge, error)
	Update(ctx context.Context, questArchetypeChallenge *models.QuestArchetypeChallenge) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAllByNodeID(ctx context.Context, nodeID uuid.UUID) ([]*models.QuestArchetypeChallenge, error)
}

type QuestArchetypeNodeChallengeHandle interface {
	Create(ctx context.Context, questArchetypeNodeChallenge *models.QuestArchetypeNodeChallenge) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeNodeChallenge, error)
	FindAll(ctx context.Context) ([]*models.QuestArchetypeNodeChallenge, error)
	Update(ctx context.Context, questArchetypeNodeChallenge *models.QuestArchetypeNodeChallenge) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ZoneQuestArchetypeHandle interface {
	Create(ctx context.Context, zoneQuestArchetype *models.ZoneQuestArchetype) error
	FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]*models.ZoneQuestArchetype, error)
	FindByZoneIDAndQuestArchetypeID(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID) (*models.ZoneQuestArchetype, error)
	Delete(ctx context.Context, zoneQuestArchetypeID uuid.UUID) error
	DeleteByZoneIDAndQuestArchetypeID(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID) error
	DeleteByZoneID(ctx context.Context, zoneID uuid.UUID) error
	DeleteByQuestArchetypeID(ctx context.Context, questArchetypeID uuid.UUID) error
	DeleteAll(ctx context.Context) error
	FindAll(ctx context.Context) ([]*models.ZoneQuestArchetype, error)
}

type TrackedPointOfInterestGroupHandle interface {
	Create(ctx context.Context, pointOfInterestGroupID uuid.UUID, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TrackedPointOfInterestGroup, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PointHandle interface {
	CreatePoint(ctx context.Context, latitude float64, longitude float64) (*models.Point, error)
}

type UserLevelHandle interface {
	ProcessExperiencePointAdditions(ctx context.Context, userID uuid.UUID, experiencePoints int) (*models.UserLevel, error)
	FindOrCreateForUser(ctx context.Context, userID uuid.UUID) (*models.UserLevel, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserLevel, error)
	Create(ctx context.Context, userLevel *models.UserLevel) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type UserZoneReputationHandle interface {
	ProcessReputationPointAdditions(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, reputationPoints int) (*models.UserZoneReputation, error)
	FindOrCreateForUserAndZone(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID) (*models.UserZoneReputation, error)
	FindAllForUser(ctx context.Context, userID uuid.UUID) ([]*models.UserZoneReputation, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PartyHandle interface {
	Create(ctx context.Context) (*models.Party, error)
	SetLeader(ctx context.Context, partyID uuid.UUID, leaderID uuid.UUID, userID uuid.UUID) error
	LeaveParty(ctx context.Context, user *models.User) error
	FindUsersParty(ctx context.Context, partyID uuid.UUID) (*models.Party, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type FriendHandle interface {
	Create(ctx context.Context, firstUserID uuid.UUID, secondUserID uuid.UUID) (*models.Friend, error)
	FindAllFriends(ctx context.Context, userID uuid.UUID) ([]models.User, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type FriendInviteHandle interface {
	Create(ctx context.Context, inviterID uuid.UUID, inviteeID uuid.UUID) (*models.FriendInvite, error)
	FindAllInvites(ctx context.Context, userID uuid.UUID) ([]models.FriendInvite, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.FriendInvite, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PartyInviteHandle interface {
	Create(ctx context.Context, inviter *models.User, inviteeID uuid.UUID) (*models.PartyInvite, error)
	FindAllInvites(ctx context.Context, userID uuid.UUID) ([]models.PartyInvite, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PartyInvite, error)
	Accept(ctx context.Context, id uuid.UUID, user *models.User) (*models.PartyInvite, error)
	Reject(ctx context.Context, id uuid.UUID, user *models.User) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type ActivityHandle interface {
	GetFeed(ctx context.Context, userID uuid.UUID) ([]models.Activity, error)
	MarkAsSeen(ctx context.Context, activityIDs []uuid.UUID) error
	CreateActivity(ctx context.Context, activity models.Activity) error
	CreateActivitiesForPartyMembers(ctx context.Context, partyID *uuid.UUID, userID *uuid.UUID, activityType models.ActivityType, data []byte) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PointOfInterestGroupMemberHandle interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroupMember, error)
	FindByPointOfInterestAndGroup(ctx context.Context, pointOfInterestID uuid.UUID, groupID uuid.UUID) (*models.PointOfInterestGroupMember, error)
}

type CharacterHandle interface {
	Create(ctx context.Context, character *models.Character) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Character, error)
	FindAll(ctx context.Context) ([]*models.Character, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.Character) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByMovementPatternType(ctx context.Context, patternType models.MovementPatternType) ([]*models.Character, error)
}

type CharacterActionHandle interface {
	Create(ctx context.Context, characterAction *models.CharacterAction) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.CharacterAction, error)
	FindAll(ctx context.Context) ([]*models.CharacterAction, error)
	FindByCharacterID(ctx context.Context, characterID uuid.UUID) ([]*models.CharacterAction, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.CharacterAction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuestAcceptanceHandle interface {
	Create(ctx context.Context, questAcceptance *models.QuestAcceptance) error
	FindByUserAndQuest(ctx context.Context, userID uuid.UUID, pointOfInterestGroupID uuid.UUID) (*models.QuestAcceptance, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.QuestAcceptance, error)
}

type MovementPatternHandle interface {
	Create(ctx context.Context, movementPattern *models.MovementPattern) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.MovementPattern, error)
	FindAll(ctx context.Context) ([]*models.MovementPattern, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.MovementPattern) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByType(ctx context.Context, patternType models.MovementPatternType) ([]*models.MovementPattern, error)
	FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]*models.MovementPattern, error)
}

type TreasureChestHandle interface {
	Create(ctx context.Context, treasureChest *models.TreasureChest) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.TreasureChest, error)
	FindAll(ctx context.Context) ([]models.TreasureChest, error)
	FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.TreasureChest, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.TreasureChest) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddItem(ctx context.Context, treasureChestID uuid.UUID, inventoryItemID int, quantity int) error
	RemoveItem(ctx context.Context, treasureChestID uuid.UUID, inventoryItemID int) error
	UpdateItemQuantity(ctx context.Context, treasureChestID uuid.UUID, inventoryItemID int, quantity int) error
	InvalidateByZoneID(ctx context.Context, zoneID uuid.UUID) error
	HasUserOpenedChest(ctx context.Context, userID uuid.UUID, chestID uuid.UUID) (bool, error)
	CreateUserTreasureChestOpening(ctx context.Context, opening *models.UserTreasureChestOpening) error
	FindByIDWithUserStatus(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.TreasureChest, bool, error)
	FindAllWithUserStatus(ctx context.Context, userID *uuid.UUID) ([]models.TreasureChest, map[uuid.UUID]bool, error)
	FindByZoneIDWithUserStatus(ctx context.Context, zoneID uuid.UUID, userID *uuid.UUID) ([]models.TreasureChest, map[uuid.UUID]bool, error)
}

type DocumentHandle interface {
	Create(ctx context.Context, document *models.Document, existingTagIDs []uuid.UUID, newTagTexts []string) (*models.Document, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Document, error)
	FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.Document, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Document, error)
	Update(ctx context.Context, document *models.Document) error
	UpdateTags(ctx context.Context, documentID uuid.UUID, existingTagIDs []uuid.UUID, newTagTexts []string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DocumentTagHandle interface {
	FindOrCreateByText(ctx context.Context, text string) (*models.DocumentTag, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.DocumentTag, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.DocumentTag, error)
}

type GoogleDriveTokenHandle interface {
	Create(ctx context.Context, token *models.GoogleDriveToken) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.GoogleDriveToken, error)
	Update(ctx context.Context, token *models.GoogleDriveToken) error
	Delete(ctx context.Context, userID uuid.UUID) error
	RefreshToken(ctx context.Context, userID uuid.UUID, newAccessToken string, expiresAt time.Time) error
}

type DropboxTokenHandle interface {
	Create(ctx context.Context, token *models.DropboxToken) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.DropboxToken, error)
	Update(ctx context.Context, token *models.DropboxToken) error
	Delete(ctx context.Context, userID uuid.UUID) error
	RefreshToken(ctx context.Context, userID uuid.UUID, newAccessToken string, expiresAt time.Time) error
}

type HueTokenHandle interface {
	Create(ctx context.Context, token *models.HueToken) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.HueToken, error)
	FindLatest(ctx context.Context) (*models.HueToken, error)
	Update(ctx context.Context, token *models.HueToken) error
	Delete(ctx context.Context, id uuid.UUID) error
	RefreshToken(ctx context.Context, id uuid.UUID, newAccessToken string, expiresAt time.Time) error
}

type FeteRoomHandle interface {
	Create(ctx context.Context, room *models.FeteRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.FeteRoom, error)
	FindAll(ctx context.Context) ([]models.FeteRoom, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.FeteRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FeteTeamHandle interface {
	Create(ctx context.Context, team *models.FeteTeam) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.FeteTeam, error)
	FindAll(ctx context.Context) ([]models.FeteTeam, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.FeteTeam) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FeteRoomLinkedListTeamHandle interface {
	Create(ctx context.Context, item *models.FeteRoomLinkedListTeam) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.FeteRoomLinkedListTeam, error)
	FindAll(ctx context.Context) ([]models.FeteRoomLinkedListTeam, error)
	Update(ctx context.Context, id uuid.UUID, updates *models.FeteRoomLinkedListTeam) error
	Delete(ctx context.Context, id uuid.UUID) error
}
