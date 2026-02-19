package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
)

type client struct {
	db                                *gorm.DB
	scoreHandle                       *scoreHandler
	userHandle                        *userHandle
	howManyQuestionHandle             *howManyQuestionHandle
	howManyAnswerHandle               *howManyAnswerHandle
	teamHandle                        *teamHandle
	userTeamHandle                    *userTeamHandle
	pointOfInterestHandle             *pointOfInterestHandle
	neighboringPointsOfInterestHandle *neighboringPointsOfInterestHandle
	textVerificationCodeHandle        *textVerificationCodeHandle
	sentTextHandle                    *sentTextHandle
	howManySubscriptionHandle         *howManySubscriptionHandle
	sonarSurveyHandle                 *sonarSurveyHandle
	sonarSurveySubmissionHandle       *sonarSurveySubmissionHandle
	sonarActivityHandle               *sonarActivityHandle
	sonarCategoryHandle               *sonarCategoryHandle
	matchHandle                       *matchHandle
	verificationCodeHandle            *verificationCodeHandler
	pointOfInterestGroupHandle        *pointOfInterestGroupHandle
	pointOfInterestChallengeHandle    *pointOfInterestChallengeHandle
	pointOfInterestImportHandle       *pointOfInterestImportHandle
	inventoryItemHandle               *inventoryItemHandler
	newUserStarterConfigHandle        *newUserStarterConfigHandle
	auditItemHandle                   *auditItemHandler
	imageGenerationHandle             *imageGenerationHandle
	outfitProfileGenerationHandle     *outfitProfileGenerationHandle
	pointOfInterestChildrenHandle     *pointOfInterestChildrenHandle
	pointOfInterestDiscoveryHandle    *pointOfInterestDiscoveryHandle
	matchUserHandle                   *matchUserHandle
	tagHandle                         *tagHandle
	tagGroupHandle                    *tagGroupHandle
	zoneHandle                        *zoneHandler
	locationArchetypeHandle           *locationArchetypeHandle
	questArchetypeHandle              *questArchetypeHandle
	questArchetypeNodeHandle          *questArchetypeNodeHandle
	questArchetypeChallengeHandle     *questArchetypeChallengeHandle
	questArchetypeNodeChallengeHandle *questArchetypeNodeChallengeHandle
	zoneQuestArchetypeHandle          *zoneQuestArchetypeHandle
	trackedPointOfInterestGroupHandle *trackedPointOfInterestGroupHandle
	trackedQuestHandle                *trackedQuestHandle
	pointHandle                       *pointHandler
	userLevelHandle                   *userLevelHandler
	userZoneReputationHandle          *userZoneReputationHandler
	partyHandle                       *partyHandle
	friendHandle                      *friendHandle
	friendInviteHandle                *friendInviteHandle
	partyInviteHandle                 *partyInviteHandle
	postHandle                        *postHandle
	postTagHandle                     *postTagHandle
	postFlagHandle                    *postFlagHandle
	postReactionHandle                *postReactionHandle
	postCommentHandle                 *postCommentHandle
	activityHandle                    *activityHandle
	pointOfInterestGroupMemberHandle  *pointOfInterestGroupMemberHandle
	characterHandle                   *characterHandler
	characterLocationHandle           *characterLocationHandle
	characterActionHandle             *characterActionHandler
	questAcceptanceHandle             *questAcceptanceHandle
	questAcceptanceV2Handle           *questAcceptanceV2Handle
	questHandle                       *questHandle
	questItemRewardHandle             *questItemRewardHandle
	questNodeHandle                   *questNodeHandle
	questNodeChallengeHandle          *questNodeChallengeHandle
	questNodeChildHandle              *questNodeChildHandle
	questNodeProgressHandle           *questNodeProgressHandle
	movementPatternHandle             *movementPatternHandler
	treasureChestHandle               *treasureChestHandle
	documentHandle                    *documentHandler
	documentTagHandle                 *documentTagHandler
	documentLocationHandle            *documentLocationHandler
	googleDriveTokenHandle            *googleDriveTokenHandler
	dropboxTokenHandle                *dropboxTokenHandler
	hueTokenHandle                    *hueTokenHandler
	trendingDestinationHandle         *trendingDestinationHandler
	quickDecisionRequestHandle        *quickDecisionRequestHandler
	communityPollHandle               *communityPollHandler
	utilityClosetPuzzleHandle         *utilityClosetPuzzleHandler
	feteRoomHandle                    *feteRoomHandler
	feteTeamHandle                    *feteTeamHandler
	feteRoomLinkedListTeamHandle      *feteRoomLinkedListTeamHandler
	feteRoomTeamHandle                *feteRoomTeamHandler
	blockchainTransactionHandle       *blockchainTransactionHandle
	userCertificateHandle             *userCertificateHandle
	albumHandle                       *albumHandle
	albumMemberHandle                 *albumMemberHandle
	albumInviteHandle                 *albumInviteHandle
	albumPostHandle                   *albumPostHandle
	albumShareHandle                  *albumShareHandle
	notificationHandle                *notificationHandle
	userDeviceTokenHandle             *userDeviceTokenHandle
	userRecentPostTagHandle           *userRecentPostTagHandle
	socialAccountHandle               *socialAccountHandler
	insiderTradeHandle                *insiderTradeHandle
}

type ClientConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func NewClient(cfg ClientConfig) (DbClient, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return &client{
		db:                                db,
		scoreHandle:                       &scoreHandler{db: db},
		userHandle:                        &userHandle{db: db},
		howManyQuestionHandle:             &howManyQuestionHandle{db: db},
		howManyAnswerHandle:               &howManyAnswerHandle{db: db},
		teamHandle:                        &teamHandle{db: db},
		userTeamHandle:                    &userTeamHandle{db: db},
		pointOfInterestHandle:             &pointOfInterestHandle{db: db},
		neighboringPointsOfInterestHandle: &neighboringPointsOfInterestHandle{db: db},
		textVerificationCodeHandle:        &textVerificationCodeHandle{db: db},
		sentTextHandle:                    &sentTextHandle{db: db},
		howManySubscriptionHandle:         &howManySubscriptionHandle{db: db},
		sonarSurveyHandle:                 &sonarSurveyHandle{db: db},
		sonarSurveySubmissionHandle:       &sonarSurveySubmissionHandle{db: db},
		sonarActivityHandle:               &sonarActivityHandle{db: db},
		sonarCategoryHandle:               &sonarCategoryHandle{db: db},
		matchHandle:                       &matchHandle{db: db},
		verificationCodeHandle:            &verificationCodeHandler{db: db},
		pointOfInterestGroupHandle:        &pointOfInterestGroupHandle{db: db},
		pointOfInterestChallengeHandle:    &pointOfInterestChallengeHandle{db: db},
		pointOfInterestImportHandle:       &pointOfInterestImportHandle{db: db},
		inventoryItemHandle:               &inventoryItemHandler{db: db},
		newUserStarterConfigHandle:        &newUserStarterConfigHandle{db: db},
		auditItemHandle:                   &auditItemHandler{db: db},
		imageGenerationHandle:             &imageGenerationHandle{db: db},
		outfitProfileGenerationHandle:     &outfitProfileGenerationHandle{db: db},
		pointOfInterestChildrenHandle:     &pointOfInterestChildrenHandle{db: db},
		pointOfInterestDiscoveryHandle:    &pointOfInterestDiscoveryHandle{db: db},
		matchUserHandle:                   &matchUserHandle{db: db},
		tagHandle:                         &tagHandle{db: db},
		tagGroupHandle:                    &tagGroupHandle{db: db},
		zoneHandle:                        &zoneHandler{db: db},
		locationArchetypeHandle:           &locationArchetypeHandle{db: db},
		questArchetypeHandle:              &questArchetypeHandle{db: db},
		questArchetypeNodeHandle:          &questArchetypeNodeHandle{db: db},
		questArchetypeChallengeHandle:     &questArchetypeChallengeHandle{db: db},
		questArchetypeNodeChallengeHandle: &questArchetypeNodeChallengeHandle{db: db},
		zoneQuestArchetypeHandle:          &zoneQuestArchetypeHandle{db: db},
		trackedPointOfInterestGroupHandle: &trackedPointOfInterestGroupHandle{db: db},
		trackedQuestHandle:                &trackedQuestHandle{db: db},
		pointHandle:                       &pointHandler{db: db},
		userLevelHandle:                   &userLevelHandler{db: db},
		userZoneReputationHandle:          &userZoneReputationHandler{db: db},
		partyHandle:                       &partyHandle{db: db},
		friendHandle:                      &friendHandle{db: db},
		friendInviteHandle:                &friendInviteHandle{db: db},
		partyInviteHandle:                 &partyInviteHandle{db: db},
		postHandle:                        &postHandle{db: db},
		postTagHandle:                     &postTagHandle{db: db},
		postFlagHandle:                    &postFlagHandle{db: db},
		albumHandle:                       &albumHandle{db: db},
		albumMemberHandle:                 &albumMemberHandle{db: db},
		albumInviteHandle:                 &albumInviteHandle{db: db},
		albumPostHandle:                   &albumPostHandle{db: db},
		albumShareHandle:                  &albumShareHandle{db: db},
		notificationHandle:                &notificationHandle{db: db},
		userDeviceTokenHandle:             &userDeviceTokenHandle{db: db},
		userRecentPostTagHandle:           &userRecentPostTagHandle{db: db},
		postReactionHandle:                &postReactionHandle{db: db},
		postCommentHandle:                 &postCommentHandle{db: db},
		activityHandle:                    &activityHandle{db: db},
		pointOfInterestGroupMemberHandle:  &pointOfInterestGroupMemberHandle{db: db},
		characterHandle:                   &characterHandler{db: db},
		characterLocationHandle:           &characterLocationHandle{db: db},
		characterActionHandle:             &characterActionHandler{db: db},
		questAcceptanceHandle:             &questAcceptanceHandle{db: db},
		questAcceptanceV2Handle:           &questAcceptanceV2Handle{db: db},
		questHandle:                       &questHandle{db: db},
		questItemRewardHandle:             &questItemRewardHandle{db: db},
		questNodeHandle:                   &questNodeHandle{db: db},
		questNodeChallengeHandle:          &questNodeChallengeHandle{db: db},
		questNodeChildHandle:              &questNodeChildHandle{db: db},
		questNodeProgressHandle:           &questNodeProgressHandle{db: db},
		movementPatternHandle:             &movementPatternHandler{db: db},
		treasureChestHandle:               &treasureChestHandle{db: db},
		documentHandle:                    &documentHandler{db: db},
		documentTagHandle:                 &documentTagHandler{db: db},
		documentLocationHandle:            &documentLocationHandler{db: db},
		googleDriveTokenHandle:            &googleDriveTokenHandler{db: db},
		dropboxTokenHandle:                &dropboxTokenHandler{db: db},
		hueTokenHandle:                    &hueTokenHandler{db: db},
		trendingDestinationHandle:         &trendingDestinationHandler{db: db},
		quickDecisionRequestHandle:        &quickDecisionRequestHandler{db: db},
		communityPollHandle:               &communityPollHandler{db: db},
		utilityClosetPuzzleHandle:         &utilityClosetPuzzleHandler{db: db},
		feteRoomHandle:                    &feteRoomHandler{db: db},
		feteTeamHandle:                    &feteTeamHandler{db: db},
		feteRoomLinkedListTeamHandle:      &feteRoomLinkedListTeamHandler{db: db},
		feteRoomTeamHandle:                &feteRoomTeamHandler{db: db},
		blockchainTransactionHandle:       &blockchainTransactionHandle{db: db},
		userCertificateHandle:             &userCertificateHandle{db: db},
		socialAccountHandle:               &socialAccountHandler{db: db},
		insiderTradeHandle:                &insiderTradeHandle{db: db},
	}, nil
}

func (c *client) Activity() ActivityHandle {
	return c.activityHandle
}

func (c *client) PointOfInterestGroupMember() PointOfInterestGroupMemberHandle {
	return c.pointOfInterestGroupMemberHandle
}

func (c *client) PartyInvite() PartyInviteHandle {
	return c.partyInviteHandle
}

func (c *client) FriendInvite() FriendInviteHandle {
	return c.friendInviteHandle
}

func (c *client) Friend() FriendHandle {
	return c.friendHandle
}

func (c *client) Post() PostHandle {
	return c.postHandle
}

func (c *client) PostTag() PostTagHandle {
	return c.postTagHandle
}

func (c *client) PostFlag() PostFlagHandle {
	return c.postFlagHandle
}

func (c *client) Album() AlbumHandle {
	return c.albumHandle
}

func (c *client) AlbumMember() AlbumMemberHandle {
	return c.albumMemberHandle
}

func (c *client) AlbumInvite() AlbumInviteHandle {
	return c.albumInviteHandle
}

func (c *client) AlbumPost() AlbumPostHandle {
	return c.albumPostHandle
}

func (c *client) AlbumShare() AlbumShareHandle {
	return c.albumShareHandle
}

func (c *client) Notification() NotificationHandle {
	return c.notificationHandle
}

func (c *client) UserDeviceToken() UserDeviceTokenHandle {
	return c.userDeviceTokenHandle
}

func (c *client) UserRecentPostTag() UserRecentPostTagHandle {
	return c.userRecentPostTagHandle
}

func (c *client) PostReaction() PostReactionHandle {
	return c.postReactionHandle
}

func (c *client) PostComment() PostCommentHandle {
	return c.postCommentHandle
}

func (c *client) Party() PartyHandle {
	return c.partyHandle
}

func (c *client) UserZoneReputation() UserZoneReputationHandle {
	return c.userZoneReputationHandle
}

func (c *client) UserLevel() UserLevelHandle {
	return c.userLevelHandle
}

func (c *client) Point() PointHandle {
	return c.pointHandle
}

func (c *client) TrackedPointOfInterestGroup() TrackedPointOfInterestGroupHandle {
	return c.trackedPointOfInterestGroupHandle
}

func (c *client) TrackedQuest() TrackedQuestHandle {
	return c.trackedQuestHandle
}

func (c *client) ZoneQuestArchetype() ZoneQuestArchetypeHandle {
	return c.zoneQuestArchetypeHandle
}

func (c *client) LocationArchetype() LocationArchetypeHandle {
	return c.locationArchetypeHandle
}

func (c *client) QuestArchetype() QuestArchetypeHandle {
	return c.questArchetypeHandle
}

func (c *client) QuestArchetypeNode() QuestArchetypeNodeHandle {
	return c.questArchetypeNodeHandle
}

func (c *client) QuestArchetypeChallenge() QuestArchetypeChallengeHandle {
	return c.questArchetypeChallengeHandle
}

func (c *client) QuestArchetypeNodeChallenge() QuestArchetypeNodeChallengeHandle {
	return c.questArchetypeNodeChallengeHandle
}

func (c *client) Zone() ZoneHandle {
	return c.zoneHandle
}

func (c *client) Tag() TagHandle {
	return c.tagHandle
}

func (c *client) TagGroup() TagGroupHandle {
	return c.tagGroupHandle
}

func (c *client) MatchUser() MatchUserHandle {
	return c.matchUserHandle
}

func (c *client) PointOfInterestDiscovery() PointOfInterestDiscoveryHandle {
	return c.pointOfInterestDiscoveryHandle
}

func (c *client) PointOfInterestChildren() PointOfInterestChildrenHandle {
	return c.pointOfInterestChildrenHandle
}

func (c *client) AuditItem() AuditItemHandle {
	return c.auditItemHandle
}

func (c *client) InventoryItem() InventoryItemHandle {
	return c.inventoryItemHandle
}

func (c *client) NewUserStarterConfig() NewUserStarterConfigHandle {
	return c.newUserStarterConfigHandle
}

func (c *client) PointOfInterestChallenge() PointOfInterestChallengeHandle {
	return c.pointOfInterestChallengeHandle
}

func (c *client) PointOfInterestImport() PointOfInterestImportHandle {
	return c.pointOfInterestImportHandle
}

func (c *client) PointOfInterestGroup() PointOfInterestGroupHandle {
	return c.pointOfInterestGroupHandle
}

func (c *client) VerificationCode() VerificationCodeHandle {
	return c.verificationCodeHandle
}

func (c *client) Match() MatchHandle {
	return c.matchHandle
}

func (c *client) Score() ScoreHandle {
	return c.scoreHandle
}

func (c *client) Exec(ctx context.Context, q string) error {
	return c.db.WithContext(ctx).Exec(q).Error
}

func (c *client) HowManyAnswer() HowManyAnswerHandle {
	return c.howManyAnswerHandle
}

func (c *client) HowManyQuestion() HowManyQuestionHandle {
	return c.howManyQuestionHandle
}

func (c *client) User() UserHandle {
	return c.userHandle
}

func (c *client) PointOfInterest() PointOfInterestHandle {
	return c.pointOfInterestHandle
}

func (c *client) NeighboringPointsOfInterest() NeighboringPointsOfInterestHandle {
	return c.neighboringPointsOfInterestHandle
}

func (c *client) Team() TeamHandle {
	return c.teamHandle
}

func (c *client) UserTeam() UserTeamHandle {
	return c.userTeamHandle
}

func (c *client) TextVerificationCode() TextVerificationCodeHandle {
	return c.textVerificationCodeHandle
}

func (c *client) SentText() SentTextHandle {
	return c.sentTextHandle
}

func (c *client) HowManySubscription() HowManySubscriptionHandle {
	return c.howManySubscriptionHandle
}

func (c *client) SonarSurvey() SonarSurveyHandle {
	return c.sonarSurveyHandle
}

func (c *client) SonarSurveySubmission() SonarSurveySubmissionHandle {
	return c.sonarSurveySubmissionHandle
}

func (c *client) SonarActivity() SonarActivityHandle {
	return c.sonarActivityHandle
}

func (c *client) SonarCategory() SonarCategoryHandle {
	return c.sonarCategoryHandle
}

func (c *client) ImageGeneration() ImageGenerationHandle {
	return c.imageGenerationHandle
}

func (c *client) OutfitProfileGeneration() OutfitProfileGenerationHandle {
	return c.outfitProfileGenerationHandle
}

func (c *client) Character() CharacterHandle {
	return c.characterHandle
}

func (c *client) CharacterLocation() CharacterLocationHandle {
	return c.characterLocationHandle
}

func (c *client) CharacterAction() CharacterActionHandle {
	return c.characterActionHandle
}

func (c *client) QuestAcceptance() QuestAcceptanceHandle {
	return c.questAcceptanceHandle
}

func (c *client) QuestAcceptanceV2() QuestAcceptanceV2Handle {
	return c.questAcceptanceV2Handle
}

func (c *client) MovementPattern() MovementPatternHandle {
	return c.movementPatternHandle
}

func (c *client) Quest() QuestHandle {
	return c.questHandle
}

func (c *client) QuestItemReward() QuestItemRewardHandle {
	return c.questItemRewardHandle
}

func (c *client) QuestNode() QuestNodeHandle {
	return c.questNodeHandle
}

func (c *client) QuestNodeChallenge() QuestNodeChallengeHandle {
	return c.questNodeChallengeHandle
}

func (c *client) QuestNodeChild() QuestNodeChildHandle {
	return c.questNodeChildHandle
}

func (c *client) QuestNodeProgress() QuestNodeProgressHandle {
	return c.questNodeProgressHandle
}

func (c *client) TreasureChest() TreasureChestHandle {
	return c.treasureChestHandle
}

func (c *client) Document() DocumentHandle {
	return c.documentHandle
}

func (c *client) DocumentTag() DocumentTagHandle {
	return c.documentTagHandle
}

func (c *client) DocumentLocation() DocumentLocationHandle {
	return c.documentLocationHandle
}

func (c *client) GoogleDriveToken() GoogleDriveTokenHandle {
	return c.googleDriveTokenHandle
}

func (c *client) DropboxToken() DropboxTokenHandle {
	return c.dropboxTokenHandle
}

func (c *client) HueToken() HueTokenHandle {
	return c.hueTokenHandle
}

func (c *client) TrendingDestination() TrendingDestinationHandle {
	return c.trendingDestinationHandle
}

func (c *client) QuickDecisionRequest() QuickDecisionRequestHandle {
	return c.quickDecisionRequestHandle
}

func (c *client) CommunityPoll() CommunityPollHandle {
	return c.communityPollHandle
}

func (c *client) UtilityClosetPuzzle() UtilityClosetPuzzleHandle {
	return c.utilityClosetPuzzleHandle
}

func (c *client) FeteRoom() FeteRoomHandle {
	return c.feteRoomHandle
}

func (c *client) FeteTeam() FeteTeamHandle {
	return c.feteTeamHandle
}

func (c *client) FeteRoomLinkedListTeam() FeteRoomLinkedListTeamHandle {
	return c.feteRoomLinkedListTeamHandle
}

func (c *client) FeteRoomTeam() FeteRoomTeamHandle {
	return c.feteRoomTeamHandle
}

func (c *client) BlockchainTransaction() BlockchainTransactionHandle {
	return c.blockchainTransactionHandle
}

func (c *client) UserCertificate() UserCertificateHandle {
	return c.userCertificateHandle
}

func (c *client) SocialAccount() SocialAccountHandle {
	return c.socialAccountHandle
}

func (c *client) InsiderTrade() InsiderTradeHandle {
	return c.insiderTradeHandle
}
