package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	monsterBattlePartyInviteRadiusMeters  = 50.0
	monsterBattlePartyFreshLocationMaxAge = 5 * time.Minute
	monsterBattlePartyKnownFarMaxAge      = 30 * time.Minute
	monsterBattlePartyKnownFarFreshMeters = 100.0
	monsterBattlePartyKnownFarStaleMeters = 500.0
	monsterBattleInviteTTL                = 1 * time.Minute
	monsterBattleLowHealthHealThreshold   = 0.45
	monsterBattleBossEmergencyHealRatio   = 0.22
	monsterBattleBossHealLockTurns        = 2
	monsterBattleBossMaxHealPercent       = 18
	monsterAbilityDampenerPctLevel5       = 55
	monsterAbilityDampenerPctLevel10      = 70
	monsterAbilityDampenerPctLevel15      = 85
	monsterAbilitySingleCapPctLevel4      = 25
	monsterAbilityAoeCapPctLevel4         = 12
	monsterAbilityBossBurstCapPctLevel4   = 35
	monsterAbilitySingleCapPctLevel10     = 30
	monsterAbilityAoeCapPctLevel10        = 15
	monsterAbilityBossBurstCapPctLevel10  = 40
)

type monsterBattleInviteProximityDecision string

const (
	monsterBattleInviteProximityDecisionInvite   monsterBattleInviteProximityDecision = "invite"
	monsterBattleInviteProximityDecisionKnownFar monsterBattleInviteProximityDecision = "known_far"
	monsterBattleInviteProximityDecisionUnknown  monsterBattleInviteProximityDecision = "unknown"
)

type monsterBattleParticipantSummary struct {
	UserID      uuid.UUID `json:"userId"`
	IsInitiator bool      `json:"isInitiator"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type monsterBattleParticipantRewardSummary struct {
	UserID               uuid.UUID                  `json:"userId"`
	RewardExperience     int                        `json:"rewardExperience"`
	RewardGold           int                        `json:"rewardGold"`
	ItemsAwarded         []models.ItemAwarded       `json:"itemsAwarded"`
	BaseResourcesAwarded []models.BaseResourceDelta `json:"baseResourcesAwarded"`
}

type monsterBattleInviteSummary struct {
	ID            uuid.UUID  `json:"id"`
	BattleID      uuid.UUID  `json:"battleId"`
	InviterUserID uuid.UUID  `json:"inviterUserId"`
	InviteeUserID uuid.UUID  `json:"inviteeUserId"`
	MonsterID     uuid.UUID  `json:"monsterId"`
	Status        string     `json:"status"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	RespondedAt   *time.Time `json:"respondedAt,omitempty"`
}

type monsterBattleTurnOrderEntry struct {
	EntityType string     `json:"entityType"`
	UserID     *uuid.UUID `json:"userId,omitempty"`
	MonsterID  *uuid.UUID `json:"monsterId,omitempty"`
	Dexterity  int        `json:"dexterity"`
}

type monsterBattleDetail struct {
	Battle               *monsterBattleResponse                  `json:"battle"`
	Participants         []monsterBattleParticipantSummary       `json:"participants"`
	ParticipantRewards   []monsterBattleParticipantRewardSummary `json:"participantRewards"`
	ParticipantResources []monsterBattleUserResource             `json:"participantResources"`
	Invites              []monsterBattleInviteSummary            `json:"invites"`
	PendingResponses     int64                                   `json:"pendingResponses"`
	TurnOrder            []monsterBattleTurnOrderEntry           `json:"turnOrder"`
}

type monsterBattleUserResource struct {
	UserID    uuid.UUID                   `json:"userId"`
	Health    int                         `json:"health"`
	MaxHealth int                         `json:"maxHealth"`
	Mana      int                         `json:"mana"`
	MaxMana   int                         `json:"maxMana"`
	Bonuses   models.CharacterStatBonuses `json:"-"`
}

type monsterBattleActionSummary struct {
	ActionType             string                         `json:"actionType"`
	AbilityID              *uuid.UUID                     `json:"abilityId,omitempty"`
	AbilityName            string                         `json:"abilityName,omitempty"`
	AbilityType            string                         `json:"abilityType,omitempty"`
	ActorMonsterID         uuid.UUID                      `json:"actorMonsterId"`
	ActorMonsterName       string                         `json:"actorMonsterName"`
	TargetUserID           *uuid.UUID                     `json:"targetUserId,omitempty"`
	TargetUserIDs          []uuid.UUID                    `json:"targetUserIds,omitempty"`
	Damage                 int                            `json:"damage,omitempty"`
	Heal                   int                            `json:"heal,omitempty"`
	UserStatusesApplied    []scenarioAppliedFailureStatus `json:"userStatusesApplied,omitempty"`
	UserStatusesRemoved    []string                       `json:"userStatusesRemoved,omitempty"`
	MonsterStatusesApplied []scenarioAppliedFailureStatus `json:"monsterStatusesApplied,omitempty"`
	MonsterStatusesRemoved []string                       `json:"monsterStatusesRemoved,omitempty"`
}

func normalizeMonsterBattleActionType(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "attack":
		return "attack"
	case "item":
		return "item"
	case "ability", "spell", "technique":
		return "ability"
	default:
		return "attack"
	}
}

func monsterBattleUserDisplayName(user *models.User) string {
	if user == nil {
		return "Party Member"
	}
	if user.Username != nil {
		if username := strings.TrimSpace(*user.Username); username != "" {
			return "@" + username
		}
	}
	if name := strings.TrimSpace(user.Name); name != "" {
		return name
	}
	return "Party Member"
}

func monsterBattleInviteAnchor(
	monster *models.Monster,
	encounter *models.MonsterEncounter,
) (float64, float64, bool) {
	if encounter != nil &&
		isValidStandaloneCoordinate(encounter.Latitude, encounter.Longitude) {
		return encounter.Latitude, encounter.Longitude, true
	}
	if monster != nil &&
		isValidStandaloneCoordinate(monster.Latitude, monster.Longitude) {
		return monster.Latitude, monster.Longitude, true
	}
	return 0, 0, false
}

func normalizeLocationSnapshotAge(age time.Duration) time.Duration {
	if age < 0 {
		return 0
	}
	return age
}

func classifyMonsterBattleInviteProximity(
	snapshot *liveness.LocationSnapshot,
	anchorLat float64,
	anchorLng float64,
	now time.Time,
) (monsterBattleInviteProximityDecision, float64, time.Duration) {
	if snapshot == nil || snapshot.SeenAt.IsZero() {
		return monsterBattleInviteProximityDecisionUnknown, 0, 0
	}
	if !isValidStandaloneCoordinate(anchorLat, anchorLng) {
		return monsterBattleInviteProximityDecisionUnknown, 0, 0
	}

	age := normalizeLocationSnapshotAge(now.Sub(snapshot.SeenAt))
	if age > monsterBattlePartyKnownFarMaxAge {
		return monsterBattleInviteProximityDecisionUnknown, 0, age
	}

	locationStr := strings.TrimSpace(snapshot.Location)
	if locationStr == "" {
		return monsterBattleInviteProximityDecisionUnknown, 0, age
	}

	memberLat, memberLng, err := parseUserLocationString(locationStr)
	if err != nil {
		return monsterBattleInviteProximityDecisionUnknown, 0, age
	}

	distanceMeters := util.HaversineDistance(
		memberLat,
		memberLng,
		anchorLat,
		anchorLng,
	)
	if age <= monsterBattlePartyFreshLocationMaxAge &&
		distanceMeters <= monsterBattlePartyInviteRadiusMeters {
		return monsterBattleInviteProximityDecisionInvite, distanceMeters, age
	}
	if age <= monsterBattlePartyFreshLocationMaxAge &&
		distanceMeters >= monsterBattlePartyKnownFarFreshMeters {
		return monsterBattleInviteProximityDecisionKnownFar, distanceMeters, age
	}
	if distanceMeters >= monsterBattlePartyKnownFarStaleMeters {
		return monsterBattleInviteProximityDecisionKnownFar, distanceMeters, age
	}
	return monsterBattleInviteProximityDecisionUnknown, distanceMeters, age
}

func (s *server) recordMonsterBattleLastAction(
	ctx context.Context,
	battle *models.MonsterBattle,
	action models.MonsterBattleLastAction,
) error {
	if battle == nil || battle.EndedAt != nil {
		return nil
	}
	action.ActionType = normalizeMonsterBattleActionType(action.ActionType)
	if err := s.dbClient.MonsterBattle().RecordLastAction(ctx, battle.ID, action); err != nil {
		return err
	}
	battle.LastActionSequence += 1
	battle.LastAction = action
	return nil
}

func monsterBattleLastActionFromMonsterAction(
	action *monsterBattleActionSummary,
) models.MonsterBattleLastAction {
	if action == nil {
		return models.MonsterBattleLastAction{}
	}
	var actorMonsterID *uuid.UUID
	if action.ActorMonsterID != uuid.Nil {
		actorMonsterID = &action.ActorMonsterID
	}
	targetName := ""
	if len(action.TargetUserIDs) > 1 {
		targetName = "the party"
	}
	return models.MonsterBattleLastAction{
		ActionType:      normalizeMonsterBattleActionType(action.ActionType),
		ActorType:       "monster",
		ActorMonsterID:  actorMonsterID,
		ActorName:       strings.TrimSpace(action.ActorMonsterName),
		AbilityID:       action.AbilityID,
		AbilityName:     strings.TrimSpace(action.AbilityName),
		AbilityType:     strings.TrimSpace(action.AbilityType),
		TargetUserID:    action.TargetUserID,
		TargetName:      targetName,
		Damage:          max(0, action.Damage),
		Heal:            max(0, action.Heal),
		StatusesApplied: len(action.UserStatusesApplied) + len(action.MonsterStatusesApplied),
		StatusesRemoved: len(action.UserStatusesRemoved) + len(action.MonsterStatusesRemoved),
	}
}

func (s *server) findActiveMonsterBattleForUser(
	ctx context.Context,
	userID uuid.UUID,
	monsterID uuid.UUID,
) (*models.MonsterBattle, error) {
	ownerBattle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, userID, monsterID)
	if err != nil {
		return nil, err
	}
	participantBattle, err := s.dbClient.MonsterBattle().FindActiveByParticipantAndMonster(ctx, userID, monsterID)
	if err != nil {
		return nil, err
	}

	if ownerBattle == nil {
		return participantBattle, nil
	}
	if participantBattle == nil {
		return ownerBattle, nil
	}
	if participantBattle.StartedAt.After(ownerBattle.StartedAt) {
		return participantBattle, nil
	}
	if ownerBattle.StartedAt.After(participantBattle.StartedAt) {
		return ownerBattle, nil
	}
	// Prefer participant battle on ties so accepted party invites route to shared combat.
	return participantBattle, nil
}

func (s *server) userCanAccessMonsterBattle(
	ctx context.Context,
	userID uuid.UUID,
	battle *models.MonsterBattle,
) (bool, error) {
	if battle == nil {
		return false, nil
	}
	if battle.UserID == userID {
		return true, nil
	}
	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return false, err
	}
	for _, participant := range participants {
		if participant.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s *server) initializeMonsterBattlePartyState(ctx context.Context, battle *models.MonsterBattle) error {
	if battle == nil {
		log.Printf("[party-combat][invite] skipped: battle is nil")
		return nil
	}
	log.Printf(
		"[party-combat][invite] initializing battle=%s initiator=%s monster=%s",
		battle.ID,
		battle.UserID,
		battle.MonsterID,
	)
	now := time.Now()
	if err := s.dbClient.MonsterBattleParticipant().CreateOrUpdate(ctx, &models.MonsterBattleParticipant{
		BattleID:    battle.ID,
		UserID:      battle.UserID,
		IsInitiator: true,
		JoinedAt:    now,
	}); err != nil {
		return err
	}

	initiator, err := s.dbClient.User().FindByID(ctx, battle.UserID)
	if err != nil || initiator == nil || initiator.PartyID == nil {
		log.Printf(
			"[party-combat][invite] no-party-flow battle=%s initiator=%s err=%v partySet=%t",
			battle.ID,
			battle.UserID,
			err,
			initiator != nil && initiator.PartyID != nil,
		)
		battle.State = string(models.MonsterBattleStateActive)
		if err := s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State); err != nil {
			return err
		}
		return s.setMonsterBattleOpeningTurn(ctx, battle)
	}

	partyMembers, err := s.dbClient.User().FindPartyMembers(ctx, battle.UserID)
	if err != nil {
		return err
	}
	if len(partyMembers) == 0 {
		log.Printf(
			"[party-combat][invite] no eligible party members battle=%s initiator=%s",
			battle.ID,
			battle.UserID,
		)
		battle.State = string(models.MonsterBattleStateActive)
		if err := s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State); err != nil {
			return err
		}
		return s.setMonsterBattleOpeningTurn(ctx, battle)
	}
	log.Printf(
		"[party-combat][invite] evaluating party members battle=%s count=%d",
		battle.ID,
		len(partyMembers),
	)

	monster, err := s.dbClient.Monster().FindByID(ctx, battle.MonsterID)
	if err != nil {
		return err
	}
	var encounter *models.MonsterEncounter
	if battle.MonsterEncounterID != nil && *battle.MonsterEncounterID != uuid.Nil {
		encounter, err = s.dbClient.MonsterEncounter().FindByID(
			ctx,
			*battle.MonsterEncounterID,
		)
		if err != nil {
			return err
		}
	}
	inviteAnchorLat, inviteAnchorLng, hasInviteAnchor := monsterBattleInviteAnchor(
		monster,
		encounter,
	)
	if monster.OwnerUserID != nil {
		battle.State = string(models.MonsterBattleStateActive)
		if err := s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State); err != nil {
			return err
		}
		return s.setMonsterBattleOpeningTurn(ctx, battle)
	}

	inviteCount := 0
	for _, member := range partyMembers {
		locationSnapshot, err := s.getUserLocationSnapshot(ctx, member.ID)
		if err != nil {
			log.Printf(
				"[party-combat][invite] skipped member=%s battle=%s reason=location-snapshot-error err=%v",
				member.ID,
				battle.ID,
				err,
			)
			continue
		}
		if locationSnapshot == nil || locationSnapshot.SeenAt.IsZero() {
			log.Printf(
				"[party-combat][invite] skipped member=%s battle=%s reason=no-recent-location",
				member.ID,
				battle.ID,
			)
			continue
		}
		if hasInviteAnchor {
			decision, distanceMeters, age := classifyMonsterBattleInviteProximity(
				locationSnapshot,
				inviteAnchorLat,
				inviteAnchorLng,
				now,
			)
			switch decision {
			case monsterBattleInviteProximityDecisionInvite:
				// Fresh and near enough: proceed.
			case monsterBattleInviteProximityDecisionKnownFar:
				log.Printf(
					"[party-combat][invite] skipped member=%s battle=%s reason=known-far distance=%.2fm age=%s",
					member.ID,
					battle.ID,
					distanceMeters,
					age.Round(time.Second),
				)
				continue
			default:
				// If we have ambiguous or stale location data, fail open so combat
				// waits for a response instead of silently bypassing party flow.
				log.Printf(
					"[party-combat][invite] fallback-invite member=%s battle=%s reason=location-unknown distance=%.2fm age=%s",
					member.ID,
					battle.ID,
					distanceMeters,
					age.Round(time.Second),
				)
			}
		} else {
			log.Printf(
				"[party-combat][invite] fallback-invite member=%s battle=%s reason=anchor-unavailable",
				member.ID,
				battle.ID,
			)
		}

		invite := &models.MonsterBattleInvite{
			BattleID:      battle.ID,
			InviterUserID: battle.UserID,
			InviteeUserID: member.ID,
			MonsterID:     battle.MonsterID,
			Status:        string(models.MonsterBattleInviteStatusPending),
			ExpiresAt:     now.Add(monsterBattleInviteTTL),
		}
		if err := s.dbClient.MonsterBattleInvite().Create(ctx, invite); err != nil {
			log.Printf(
				"[party-combat][invite] failed create member=%s battle=%s err=%v",
				member.ID,
				battle.ID,
				err,
			)
			continue
		}
		inviteCount += 1
		log.Printf(
			"[party-combat][invite] created member=%s battle=%s invite=%s expiresAt=%s",
			member.ID,
			battle.ID,
			invite.ID,
			invite.ExpiresAt.UTC().Format(time.RFC3339),
		)
		s.createMonsterBattleInviteActivity(ctx, invite, monster, initiator)
		s.sendMonsterBattleInvitePush(ctx, invite, monster, initiator)
	}

	if inviteCount == 0 {
		log.Printf(
			"[party-combat][invite] no invites created battle=%s; state set active",
			battle.ID,
		)
		battle.State = string(models.MonsterBattleStateActive)
	} else {
		log.Printf(
			"[party-combat][invite] invites created battle=%s count=%d; waiting on responses",
			battle.ID,
			inviteCount,
		)
		battle.State = string(models.MonsterBattleStatePendingPartyResponses)
	}
	if err := s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State); err != nil {
		return err
	}
	if battle.State == string(models.MonsterBattleStateActive) {
		return s.setMonsterBattleOpeningTurn(ctx, battle)
	}
	return nil
}

func userDisplayName(user *models.User) string {
	if user == nil {
		return "A party member"
	}
	if user.Username != nil && strings.TrimSpace(*user.Username) != "" {
		return "@" + strings.TrimSpace(*user.Username)
	}
	if strings.TrimSpace(user.Name) != "" {
		return strings.TrimSpace(user.Name)
	}
	if strings.TrimSpace(user.PhoneNumber) != "" {
		return strings.TrimSpace(user.PhoneNumber)
	}
	return "A party member"
}

func (s *server) createMonsterBattleInviteActivity(
	ctx context.Context,
	invite *models.MonsterBattleInvite,
	monster *models.Monster,
	inviter *models.User,
) {
	if invite == nil {
		return
	}

	payload := map[string]interface{}{
		"inviteId":      invite.ID.String(),
		"battleId":      invite.BattleID.String(),
		"monsterId":     invite.MonsterID.String(),
		"monsterName":   strings.TrimSpace(monster.Name),
		"inviterUserId": invite.InviterUserID.String(),
		"inviterName":   userDisplayName(inviter),
		"expiresAt":     invite.ExpiresAt.UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = s.dbClient.Activity().CreateActivity(ctx, models.Activity{
		UserID:       invite.InviteeUserID,
		ActivityType: models.ActivityTypeMonsterBattleInvite,
		Data:         data,
		Seen:         false,
	})
}

func (s *server) sendMonsterBattleInvitePush(
	ctx context.Context,
	invite *models.MonsterBattleInvite,
	monster *models.Monster,
	inviter *models.User,
) {
	if invite == nil {
		log.Printf("[push][party-invite] skipped: invite is nil")
		return
	}
	if s.pushClient == nil {
		log.Printf(
			"[push][party-invite] skipped: push client not configured (battle=%s invite=%s invitee=%s)",
			invite.BattleID,
			invite.ID,
			invite.InviteeUserID,
		)
		return
	}
	tokens, err := s.dbClient.UserDeviceToken().FindByUserID(ctx, invite.InviteeUserID)
	if err != nil {
		log.Printf(
			"[push][party-invite] failed to fetch tokens (battle=%s invite=%s invitee=%s): %v",
			invite.BattleID,
			invite.ID,
			invite.InviteeUserID,
			err,
		)
		return
	}
	if len(tokens) == 0 {
		log.Printf(
			"[push][party-invite] skipped: no tokens (battle=%s invite=%s invitee=%s)",
			invite.BattleID,
			invite.ID,
			invite.InviteeUserID,
		)
		return
	}
	log.Printf(
		"[push][party-invite] sending push (battle=%s invite=%s invitee=%s tokens=%d)",
		invite.BattleID,
		invite.ID,
		invite.InviteeUserID,
		len(tokens),
	)
	title := "Party Combat Invite"
	monsterName := "a monster"
	if monster != nil && strings.TrimSpace(monster.Name) != "" {
		monsterName = strings.TrimSpace(monster.Name)
	}
	body := fmt.Sprintf("%s invited you to fight %s", userDisplayName(inviter), monsterName)
	data := map[string]string{
		"type":      "monster_battle_invite",
		"inviteId":  invite.ID.String(),
		"battleId":  invite.BattleID.String(),
		"monsterId": invite.MonsterID.String(),
		"expiresAt": invite.ExpiresAt.UTC().Format(time.RFC3339),
	}
	sentCount := 0
	failedCount := 0
	for _, token := range tokens {
		if err := s.pushClient.Send(ctx, token.Token, title, body, data); err != nil {
			failedCount++
			log.Printf(
				"[push][party-invite] send failed (battle=%s invite=%s invitee=%s platform=%s token=%s): %v",
				invite.BattleID,
				invite.ID,
				invite.InviteeUserID,
				token.Platform,
				tokenPreview(token.Token),
				err,
			)
			continue
		}
		sentCount++
	}
	log.Printf(
		"[push][party-invite] send complete (battle=%s invite=%s invitee=%s sent=%d failed=%d)",
		invite.BattleID,
		invite.ID,
		invite.InviteeUserID,
		sentCount,
		failedCount,
	)
}

func (s *server) refreshMonsterBattleInviteState(
	ctx context.Context,
	battleID uuid.UUID,
) (*models.MonsterBattle, error) {
	now := time.Now()
	if _, err := s.dbClient.MonsterBattleInvite().AutoDeclineExpiredByBattle(ctx, battleID, now); err != nil {
		return nil, err
	}
	invites, err := s.dbClient.MonsterBattleInvite().FindByBattleID(ctx, battleID)
	if err != nil {
		return nil, err
	}
	for _, invite := range invites {
		if invite.Status != string(models.MonsterBattleInviteStatusAccepted) {
			continue
		}
		if err := s.dbClient.MonsterBattleParticipant().CreateOrUpdate(ctx, &models.MonsterBattleParticipant{
			BattleID:    battleID,
			UserID:      invite.InviteeUserID,
			IsInitiator: false,
			JoinedAt:    invite.UpdatedAt,
		}); err != nil {
			return nil, err
		}
	}

	pendingCount, err := s.dbClient.MonsterBattleInvite().CountPendingByBattle(ctx, battleID, now)
	if err != nil {
		return nil, err
	}

	battle, err := s.dbClient.MonsterBattle().FindByID(ctx, battleID)
	if err != nil || battle == nil {
		return battle, err
	}
	targetState := string(models.MonsterBattleStateActive)
	if pendingCount > 0 {
		targetState = string(models.MonsterBattleStatePendingPartyResponses)
	}
	if battle.State != targetState {
		if err := s.dbClient.MonsterBattle().SetState(ctx, battle.ID, targetState); err != nil {
			return nil, err
		}
		if targetState == string(models.MonsterBattleStateActive) {
			battle.State = targetState
			if err := s.setMonsterBattleOpeningTurn(ctx, battle); err != nil {
				return nil, err
			}
		}
	}
	return s.dbClient.MonsterBattle().FindByID(ctx, battleID)
}

func (s *server) setMonsterBattleOpeningTurn(
	ctx context.Context,
	battle *models.MonsterBattle,
) error {
	if battle == nil || battle.EndedAt != nil {
		return nil
	}
	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return err
	}
	turnOrder, err := s.buildMonsterBattleTurnOrder(ctx, battle, participants)
	if err != nil {
		return err
	}
	for index, entry := range turnOrder {
		if strings.ToLower(strings.TrimSpace(entry.EntityType)) != "user" {
			continue
		}
		if err := s.dbClient.MonsterBattle().SetTurnIndex(ctx, battle.ID, index); err != nil {
			return err
		}
		battle.TurnIndex = index
		return nil
	}
	return nil
}

func (s *server) getUserBattleDexterity(ctx context.Context, userID uuid.UUID) (int, error) {
	stats, err := s.dbClient.UserCharacterStats().FindOrCreateForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		return 0, err
	}
	// Keep initiative stable during an active battle. Status effects still apply
	// to combat resolution, but they should not reshuffle turn order mid-fight.
	return stats.Dexterity + equipmentBonuses.Dexterity, nil
}

func (s *server) buildMonsterBattleTurnOrder(
	ctx context.Context,
	battle *models.MonsterBattle,
	participants []models.MonsterBattleParticipant,
) ([]monsterBattleTurnOrderEntry, error) {
	if battle == nil {
		return []monsterBattleTurnOrderEntry{}, nil
	}
	order := make([]monsterBattleTurnOrderEntry, 0, len(participants)+1)
	for _, participant := range participants {
		dexterity, err := s.getUserBattleDexterity(ctx, participant.UserID)
		if err != nil {
			return nil, err
		}
		userID := participant.UserID
		order = append(order, monsterBattleTurnOrderEntry{
			EntityType: "user",
			UserID:     &userID,
			Dexterity:  dexterity,
		})
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, battle.MonsterID)
	if err != nil {
		return nil, err
	}
	monsterStats := monster.EffectiveStats()
	monsterID := battle.MonsterID
	order = append(order, monsterBattleTurnOrderEntry{
		EntityType: "monster",
		MonsterID:  &monsterID,
		Dexterity:  monsterStats.Dexterity,
	})

	sort.SliceStable(order, func(i, j int) bool {
		if order[i].Dexterity == order[j].Dexterity {
			left := order[i].EntityType
			right := order[j].EntityType
			if left == right {
				leftID := ""
				rightID := ""
				if order[i].UserID != nil {
					leftID = order[i].UserID.String()
				}
				if order[i].MonsterID != nil {
					leftID = order[i].MonsterID.String()
				}
				if order[j].UserID != nil {
					rightID = order[j].UserID.String()
				}
				if order[j].MonsterID != nil {
					rightID = order[j].MonsterID.String()
				}
				return leftID < rightID
			}
			return left < right
		}
		return order[i].Dexterity > order[j].Dexterity
	})

	return order, nil
}

func nextMonsterBattleTurnIndex(
	currentIndex int,
	entryCount int,
	isAlive func(index int) bool,
) int {
	if entryCount <= 0 {
		return 0
	}
	if currentIndex < 0 || currentIndex >= entryCount {
		currentIndex = 0
	}
	for offset := 1; offset <= entryCount; offset++ {
		nextIndex := (currentIndex + offset) % entryCount
		if isAlive == nil || isAlive(nextIndex) {
			return nextIndex
		}
	}
	return currentIndex
}

func monsterBattleTurnEntryMatchesActor(
	entry monsterBattleTurnOrderEntry,
	actingUserID *uuid.UUID,
	actingMonsterID *uuid.UUID,
) bool {
	if actingUserID != nil && entry.UserID != nil && *actingUserID == *entry.UserID {
		return true
	}
	if actingMonsterID != nil && entry.MonsterID != nil && *actingMonsterID == *entry.MonsterID {
		return true
	}
	return false
}

func (s *server) monsterBattleTurnEntryIsAlive(
	ctx context.Context,
	battle *models.MonsterBattle,
	entry monsterBattleTurnOrderEntry,
) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(entry.EntityType)) {
	case "monster":
		monsterID := battle.MonsterID
		if entry.MonsterID != nil && *entry.MonsterID != uuid.Nil {
			monsterID = *entry.MonsterID
		}
		monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
		if err != nil {
			return false, err
		}
		if monster == nil {
			return false, nil
		}
		return monster.DerivedMaxHealth()-battle.MonsterHealthDeficit > 0, nil
	case "user":
		if entry.UserID == nil || *entry.UserID == uuid.Nil {
			return false, nil
		}
		_, _, _, health, _, err := s.getScenarioResourceState(ctx, *entry.UserID)
		if err != nil {
			return false, err
		}
		return health > 0, nil
	default:
		return false, nil
	}
}

func (s *server) advanceMonsterBattleTurnState(
	ctx context.Context,
	battle *models.MonsterBattle,
	actingUserID *uuid.UUID,
	actingMonsterID *uuid.UUID,
) (*models.MonsterBattle, error) {
	if battle == nil || battle.EndedAt != nil {
		return battle, nil
	}
	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	turnOrder, err := s.buildMonsterBattleTurnOrder(ctx, battle, participants)
	if err != nil {
		return nil, err
	}
	if len(turnOrder) == 0 {
		return battle, nil
	}
	aliveByIndex := make([]bool, len(turnOrder))
	for i := range turnOrder {
		alive, err := s.monsterBattleTurnEntryIsAlive(ctx, battle, turnOrder[i])
		if err != nil {
			return nil, err
		}
		aliveByIndex[i] = alive
	}
	currentIndex := battle.TurnIndex
	for i, entry := range turnOrder {
		if monsterBattleTurnEntryMatchesActor(entry, actingUserID, actingMonsterID) {
			currentIndex = i
			break
		}
	}
	nextIndex := nextMonsterBattleTurnIndex(
		currentIndex,
		len(turnOrder),
		func(index int) bool {
			return aliveByIndex[index]
		},
	)
	if err := s.dbClient.MonsterBattle().SetTurnIndex(ctx, battle.ID, nextIndex); err != nil {
		return nil, err
	}
	battle.TurnIndex = nextIndex
	return battle, nil
}

func monsterAbilityTargetsAllEnemies(spell *models.Spell) bool {
	if spell == nil {
		return false
	}
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamageAllEnemies,
			models.SpellEffectTypeApplyDetrimentalAll:
			return true
		}
	}
	return false
}

func monsterAbilityHasOffense(spell *models.Spell) bool {
	if spell == nil {
		return false
	}
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage,
			models.SpellEffectTypeDealDamageAllEnemies,
			models.SpellEffectTypeApplyDetrimentalStatus,
			models.SpellEffectTypeApplyDetrimentalAll:
			return true
		}
	}
	return false
}

func monsterAbilityHasSupport(spell *models.Spell) bool {
	if spell == nil {
		return false
	}
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeRestoreLifePartyMember,
			models.SpellEffectTypeRestoreLifeAllParty,
			models.SpellEffectTypeApplyBeneficialStatus,
			models.SpellEffectTypeRemoveDetrimental:
			return true
		}
	}
	return false
}

func monsterAbilityDamageDampenerPercent(abilityLevel int) int {
	switch {
	case abilityLevel <= 5:
		return monsterAbilityDampenerPctLevel5
	case abilityLevel <= 10:
		return monsterAbilityDampenerPctLevel10
	case abilityLevel <= 15:
		return monsterAbilityDampenerPctLevel15
	default:
		return 100
	}
}

func applyMonsterAbilityDamageDampener(damage int, abilityLevel int) int {
	if damage <= 0 {
		return 0
	}
	dampenerPercent := monsterAbilityDamageDampenerPercent(abilityLevel)
	if dampenerPercent >= 100 {
		return damage
	}
	return maxInt(1, ((damage*dampenerPercent)+50)/100)
}

func monsterAbilityDamageCapPercentForCombat(
	monster *models.Monster,
	spell *models.Spell,
	abilityLevel int,
) int {
	if spell == nil || abilityLevel > 10 {
		return 0
	}
	targetsAllEnemies := monsterAbilityTargetsAllEnemies(spell)
	isBossBurst := monsterUsesBossHealingRules(monster) && spell.CooldownTurns >= 2
	if abilityLevel <= 4 {
		switch {
		case targetsAllEnemies:
			return monsterAbilityAoeCapPctLevel4
		case isBossBurst:
			return monsterAbilityBossBurstCapPctLevel4
		default:
			return monsterAbilitySingleCapPctLevel4
		}
	}
	switch {
	case targetsAllEnemies:
		return monsterAbilityAoeCapPctLevel10
	case isBossBurst:
		return monsterAbilityBossBurstCapPctLevel10
	default:
		return monsterAbilitySingleCapPctLevel10
	}
}

func capMonsterAbilityDamageAgainstHealth(
	damage int,
	monster *models.Monster,
	spell *models.Spell,
	userLevel int,
	targetMaxHealth int,
) int {
	if damage <= 0 || targetMaxHealth <= 0 || monster == nil {
		return maxInt(0, damage)
	}
	abilityLevel := cappedMonsterAbilityLevelForUserLevel(
		monster.EffectiveLevel(),
		userLevel,
	)
	capPercent := monsterAbilityDamageCapPercentForCombat(
		monster,
		spell,
		abilityLevel,
	)
	if capPercent <= 0 {
		return damage
	}
	maxDamage := maxInt(1, (targetMaxHealth*capPercent)/100)
	return min(damage, maxDamage)
}

func monsterAbilityDamageForCombat(
	monster *models.Monster,
	spell *models.Spell,
	userLevel int,
) int {
	if monster == nil || spell == nil {
		return 0
	}
	abilityLevel := cappedMonsterAbilityLevelForUserLevel(
		monster.EffectiveLevel(),
		userLevel,
	)
	explicitDamage := 0
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage, models.SpellEffectTypeDealDamageAllEnemies:
			hits := effect.Hits
			if hits <= 0 && effect.Amount > 0 {
				hits = 1
			}
			explicitDamage += maxInt(0, effect.Amount) * maxInt(0, hits)
		}
	}
	if explicitDamage <= 0 {
		return 0
	}
	bonus := maxInt(0, monster.EffectiveLevel()/3)
	if models.NormalizeSpellAbilityType(string(spell.AbilityType)) == models.SpellAbilityTypeTechnique {
		bonus += maxInt(0, (monster.EffectiveStats().Strength-10)/2)
	}
	return applyMonsterAbilityDamageDampener(
		maxInt(1, explicitDamage+bonus),
		abilityLevel,
	)
}

func monsterAbilityDamageAffinity(spell *models.Spell) *string {
	if spell == nil {
		return nil
	}
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeDealDamage, models.SpellEffectTypeDealDamageAllEnemies:
			return models.NormalizeOptionalDamageAffinity(effect.DamageAffinity)
		}
	}
	return nil
}

func monsterBasicAttackAffinity(monster *models.Monster) *string {
	if monster == nil {
		return nil
	}
	if monster.DominantHandInventoryItem != nil {
		if affinity := models.NormalizeOptionalDamageAffinity(monster.DominantHandInventoryItem.DamageAffinity); affinity != nil {
			return affinity
		}
	}
	if monster.WeaponInventoryItem != nil {
		if affinity := models.NormalizeOptionalDamageAffinity(monster.WeaponInventoryItem.DamageAffinity); affinity != nil {
			return affinity
		}
	}
	affinity := string(models.DamageAffinityPhysical)
	return &affinity
}

func monsterAbilityHealingForCombat(spell *models.Spell) int {
	if spell == nil {
		return 0
	}
	total := 0
	for _, effect := range spell.Effects {
		switch effect.Type {
		case models.SpellEffectTypeRestoreLifePartyMember, models.SpellEffectTypeRestoreLifeAllParty:
			total += maxInt(0, effect.Amount)
		}
	}
	return total
}

func monsterAbilityHasHealing(spell *models.Spell) bool {
	return monsterAbilityHealingForCombat(spell) > 0
}

func monsterUsesBossHealingRules(monster *models.Monster) bool {
	if monster == nil || monster.Template == nil {
		return false
	}
	switch models.NormalizeMonsterTemplateType(string(monster.Template.MonsterType)) {
	case models.MonsterTemplateTypeBoss, models.MonsterTemplateTypeRaid:
		return true
	default:
		return false
	}
}

func bestMonsterHealingAbility(abilities []models.Spell) *models.Spell {
	if len(abilities) == 0 {
		return nil
	}
	best := abilities[0]
	bestHealing := monsterAbilityHealingForCombat(&best)
	for _, ability := range abilities[1:] {
		healing := monsterAbilityHealingForCombat(&ability)
		if healing > bestHealing {
			best = ability
			bestHealing = healing
		}
	}
	return &best
}

func adjustedMonsterAbilityHealingForCombat(
	monster *models.Monster,
	battle *models.MonsterBattle,
	spell *models.Spell,
	maxHealth int,
) int {
	healAmount := monsterAbilityHealingForCombat(spell)
	if healAmount <= 0 {
		return 0
	}
	if monsterUsesBossHealingRules(monster) && maxHealth > 0 {
		perCastCap := maxInt(1, (maxHealth*monsterBattleBossMaxHealPercent)/100)
		healAmount = min(healAmount, perCastCap)
	}
	if battle == nil {
		return healAmount
	}
	return min(healAmount, maxInt(0, battle.MonsterHealthDeficit))
}

func applyMonsterHealingLockout(
	monster *models.Monster,
	cooldowns models.MonsterBattleAbilityCooldowns,
	now time.Time,
) models.MonsterBattleAbilityCooldowns {
	if !monsterUsesBossHealingRules(monster) {
		return cooldowns
	}
	expiresAt := cooldownExpiresAtFromTurns(monsterBattleBossHealLockTurns, now)
	if expiresAt == nil {
		return cooldowns
	}
	if cooldowns == nil {
		cooldowns = models.MonsterBattleAbilityCooldowns{}
	}
	for _, ability := range monsterCombatAbilities(monster) {
		if !monsterAbilityHasHealing(&ability) {
			continue
		}
		abilityID := ability.ID.String()
		if current, exists := cooldowns[abilityID]; exists && current.After(*expiresAt) {
			continue
		}
		cooldowns[abilityID] = *expiresAt
	}
	return cooldowns
}

func monsterCombatAbilities(monster *models.Monster) []models.Spell {
	return monsterCombatAbilitiesForUserLevel(monster, 0)
}

func monsterCombatAbilitiesForUserLevel(
	monster *models.Monster,
	userLevel int,
) []models.Spell {
	if monster == nil {
		return []models.Spell{}
	}
	return monsterTemplateResolvedAbilitiesForLevel(
		monster.Template,
		cappedMonsterAbilityLevelForUserLevel(monster.EffectiveLevel(), userLevel),
	)
}

func chooseMonsterBattleAbility(
	monster *models.Monster,
	battle *models.MonsterBattle,
	currentHealth int,
	maxHealth int,
	currentMana int,
	userLevel int,
	now time.Time,
) *models.Spell {
	abilities := monsterCombatAbilitiesForUserLevel(monster, userLevel)
	if len(abilities) == 0 {
		return nil
	}

	healingSupport := make([]models.Spell, 0, len(abilities))
	utilitySupport := make([]models.Spell, 0, len(abilities))
	offense := make([]models.Spell, 0, len(abilities))
	cooldowns := models.MonsterBattleAbilityCooldowns{}
	if battle != nil {
		cooldowns = battle.MonsterAbilityCooldowns
	}
	for _, ability := range abilities {
		if normalizeSpellAbilityType(string(ability.AbilityType)) != models.SpellAbilityTypeTechnique &&
			ability.ManaCost > currentMana {
			continue
		}
		if monsterCooldownTurnsRemaining(cooldowns, ability.ID.String(), now) > 0 {
			continue
		}
		if monsterAbilityHasSupport(&ability) {
			if monsterAbilityHasHealing(&ability) {
				healingSupport = append(healingSupport, ability)
			} else {
				utilitySupport = append(utilitySupport, ability)
			}
		}
		if monsterAbilityHasOffense(&ability) {
			offense = append(offense, ability)
		}
	}

	healthRatio := 1.0
	if maxHealth > 0 {
		healthRatio = float64(currentHealth) / float64(maxHealth)
	}
	if monsterUsesBossHealingRules(monster) {
		if healthRatio <= monsterBattleBossEmergencyHealRatio {
			if best := bestMonsterHealingAbility(healingSupport); best != nil {
				return best
			}
		}
	} else if healthRatio <= monsterBattleLowHealthHealThreshold {
		if best := bestMonsterHealingAbility(healingSupport); best != nil {
			return best
		}
	}

	if len(offense) > 0 && rand.Intn(100) < 55 {
		best := offense[0]
		bestDamage := monsterAbilityDamageForCombat(monster, &best, userLevel)
		for _, ability := range offense[1:] {
			damage := monsterAbilityDamageForCombat(monster, &ability, userLevel)
			if damage > bestDamage {
				best = ability
				bestDamage = damage
			}
		}
		return &best
	}

	if len(utilitySupport) > 0 && len(offense) == 0 {
		best := utilitySupport[rand.Intn(len(utilitySupport))]
		return &best
	}
	if len(offense) > 0 {
		best := offense[rand.Intn(len(offense))]
		return &best
	}
	if len(utilitySupport) > 0 {
		best := utilitySupport[rand.Intn(len(utilitySupport))]
		return &best
	}
	if !monsterUsesBossHealingRules(monster) {
		if best := bestMonsterHealingAbility(healingSupport); best != nil {
			return best
		}
	}
	return nil
}

func (s *server) loadMonsterBattleUserResources(
	ctx context.Context,
	battle *models.MonsterBattle,
) ([]monsterBattleUserResource, error) {
	if battle == nil {
		return []monsterBattleUserResource{}, nil
	}
	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	resources := make([]monsterBattleUserResource, 0, len(participants))
	for _, participant := range participants {
		_, maxHealth, maxMana, health, mana, err := s.getScenarioResourceState(ctx, participant.UserID)
		if err != nil {
			return nil, err
		}
		bonuses, err := s.getCharacterTotalBonuses(ctx, participant.UserID)
		if err != nil {
			return nil, err
		}
		resources = append(resources, monsterBattleUserResource{
			UserID:    participant.UserID,
			Health:    health,
			MaxHealth: maxHealth,
			Mana:      mana,
			MaxMana:   maxMana,
			Bonuses:   bonuses,
		})
	}
	return resources, nil
}

func chooseMonsterBattleTarget(resources []monsterBattleUserResource) *monsterBattleUserResource {
	var target *monsterBattleUserResource
	for i := range resources {
		if resources[i].Health <= 0 {
			continue
		}
		if target == nil || resources[i].Health < target.Health {
			target = &resources[i]
		}
	}
	return target
}

func (s *server) applyMonsterBattleUserStatuses(
	ctx context.Context,
	targetUserIDs []uuid.UUID,
	statusTemplates models.ScenarioFailureStatusTemplates,
) ([]scenarioAppliedFailureStatus, error) {
	applied := make([]scenarioAppliedFailureStatus, 0, len(statusTemplates))
	now := time.Now()
	for _, targetUserID := range targetUserIDs {
		activeNames := make([]string, 0, len(statusTemplates))
		for _, statusTemplate := range statusTemplates {
			name := strings.TrimSpace(statusTemplate.Name)
			if name == "" || statusTemplate.DurationSeconds <= 0 {
				continue
			}
			activeNames = append(activeNames, name)
		}
		if err := s.dbClient.UserStatus().DeleteActiveByUserIDAndNames(ctx, targetUserID, activeNames); err != nil {
			return nil, err
		}
		for _, statusTemplate := range statusTemplates {
			name := strings.TrimSpace(statusTemplate.Name)
			if name == "" || statusTemplate.DurationSeconds <= 0 {
				continue
			}
			status := &models.UserStatus{
				UserID:                        targetUserID,
				Name:                          name,
				Description:                   strings.TrimSpace(statusTemplate.Description),
				Effect:                        strings.TrimSpace(statusTemplate.Effect),
				Positive:                      statusTemplate.Positive,
				EffectType:                    normalizeUserStatusEffectType(statusTemplate.EffectType),
				DamagePerTick:                 statusTemplate.DamagePerTick,
				HealthPerTick:                 statusTemplate.HealthPerTick,
				ManaPerTick:                   statusTemplate.ManaPerTick,
				StrengthMod:                   statusTemplate.StrengthMod,
				DexterityMod:                  statusTemplate.DexterityMod,
				ConstitutionMod:               statusTemplate.ConstitutionMod,
				IntelligenceMod:               statusTemplate.IntelligenceMod,
				WisdomMod:                     statusTemplate.WisdomMod,
				CharismaMod:                   statusTemplate.CharismaMod,
				PhysicalDamageBonusPercent:    statusTemplate.PhysicalDamageBonusPercent,
				PiercingDamageBonusPercent:    statusTemplate.PiercingDamageBonusPercent,
				SlashingDamageBonusPercent:    statusTemplate.SlashingDamageBonusPercent,
				BludgeoningDamageBonusPercent: statusTemplate.BludgeoningDamageBonusPercent,
				FireDamageBonusPercent:        statusTemplate.FireDamageBonusPercent,
				IceDamageBonusPercent:         statusTemplate.IceDamageBonusPercent,
				LightningDamageBonusPercent:   statusTemplate.LightningDamageBonusPercent,
				PoisonDamageBonusPercent:      statusTemplate.PoisonDamageBonusPercent,
				ArcaneDamageBonusPercent:      statusTemplate.ArcaneDamageBonusPercent,
				HolyDamageBonusPercent:        statusTemplate.HolyDamageBonusPercent,
				ShadowDamageBonusPercent:      statusTemplate.ShadowDamageBonusPercent,
				PhysicalResistancePercent:     statusTemplate.PhysicalResistancePercent,
				PiercingResistancePercent:     statusTemplate.PiercingResistancePercent,
				SlashingResistancePercent:     statusTemplate.SlashingResistancePercent,
				BludgeoningResistancePercent:  statusTemplate.BludgeoningResistancePercent,
				FireResistancePercent:         statusTemplate.FireResistancePercent,
				IceResistancePercent:          statusTemplate.IceResistancePercent,
				LightningResistancePercent:    statusTemplate.LightningResistancePercent,
				PoisonResistancePercent:       statusTemplate.PoisonResistancePercent,
				ArcaneResistancePercent:       statusTemplate.ArcaneResistancePercent,
				HolyResistancePercent:         statusTemplate.HolyResistancePercent,
				ShadowResistancePercent:       statusTemplate.ShadowResistancePercent,
				StartedAt:                     now,
				LastTickAt:                    &now,
				ExpiresAt:                     now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
			}
			if err := s.dbClient.UserStatus().Create(ctx, status); err != nil {
				return nil, err
			}
			applied = append(applied, scenarioAppliedFailureStatus{
				Name:            status.Name,
				Description:     status.Description,
				Effect:          status.Effect,
				EffectType:      string(status.EffectType),
				Positive:        status.Positive,
				DamagePerTick:   status.DamagePerTick,
				HealthPerTick:   status.HealthPerTick,
				ManaPerTick:     status.ManaPerTick,
				DurationSeconds: statusTemplate.DurationSeconds,
			})
		}
	}
	return applied, nil
}

func (s *server) applyMonsterBattleMonsterStatuses(
	ctx context.Context,
	battle *models.MonsterBattle,
	monster *models.Monster,
	statusTemplates models.ScenarioFailureStatusTemplates,
) ([]scenarioAppliedFailureStatus, error) {
	if battle == nil || monster == nil {
		return []scenarioAppliedFailureStatus{}, nil
	}
	applied := make([]scenarioAppliedFailureStatus, 0, len(statusTemplates))
	activeNames := make([]string, 0, len(statusTemplates))
	for _, statusTemplate := range statusTemplates {
		name := strings.TrimSpace(statusTemplate.Name)
		if name == "" || statusTemplate.DurationSeconds <= 0 {
			continue
		}
		activeNames = append(activeNames, name)
	}
	if err := s.dbClient.MonsterStatus().DeleteActiveByBattleIDAndNames(ctx, battle.ID, activeNames); err != nil {
		return nil, err
	}
	now := time.Now()
	for _, statusTemplate := range statusTemplates {
		name := strings.TrimSpace(statusTemplate.Name)
		if name == "" || statusTemplate.DurationSeconds <= 0 {
			continue
		}
		status := &models.MonsterStatus{
			UserID:                        battle.UserID,
			BattleID:                      battle.ID,
			MonsterID:                     monster.ID,
			Name:                          name,
			Description:                   strings.TrimSpace(statusTemplate.Description),
			Effect:                        strings.TrimSpace(statusTemplate.Effect),
			Positive:                      statusTemplate.Positive,
			EffectType:                    normalizeMonsterStatusEffectType(statusTemplate.EffectType),
			DamagePerTick:                 statusTemplate.DamagePerTick,
			HealthPerTick:                 statusTemplate.HealthPerTick,
			StrengthMod:                   statusTemplate.StrengthMod,
			DexterityMod:                  statusTemplate.DexterityMod,
			ConstitutionMod:               statusTemplate.ConstitutionMod,
			IntelligenceMod:               statusTemplate.IntelligenceMod,
			WisdomMod:                     statusTemplate.WisdomMod,
			CharismaMod:                   statusTemplate.CharismaMod,
			PhysicalDamageBonusPercent:    statusTemplate.PhysicalDamageBonusPercent,
			PiercingDamageBonusPercent:    statusTemplate.PiercingDamageBonusPercent,
			SlashingDamageBonusPercent:    statusTemplate.SlashingDamageBonusPercent,
			BludgeoningDamageBonusPercent: statusTemplate.BludgeoningDamageBonusPercent,
			FireDamageBonusPercent:        statusTemplate.FireDamageBonusPercent,
			IceDamageBonusPercent:         statusTemplate.IceDamageBonusPercent,
			LightningDamageBonusPercent:   statusTemplate.LightningDamageBonusPercent,
			PoisonDamageBonusPercent:      statusTemplate.PoisonDamageBonusPercent,
			ArcaneDamageBonusPercent:      statusTemplate.ArcaneDamageBonusPercent,
			HolyDamageBonusPercent:        statusTemplate.HolyDamageBonusPercent,
			ShadowDamageBonusPercent:      statusTemplate.ShadowDamageBonusPercent,
			PhysicalResistancePercent:     statusTemplate.PhysicalResistancePercent,
			PiercingResistancePercent:     statusTemplate.PiercingResistancePercent,
			SlashingResistancePercent:     statusTemplate.SlashingResistancePercent,
			BludgeoningResistancePercent:  statusTemplate.BludgeoningResistancePercent,
			FireResistancePercent:         statusTemplate.FireResistancePercent,
			IceResistancePercent:          statusTemplate.IceResistancePercent,
			LightningResistancePercent:    statusTemplate.LightningResistancePercent,
			PoisonResistancePercent:       statusTemplate.PoisonResistancePercent,
			ArcaneResistancePercent:       statusTemplate.ArcaneResistancePercent,
			HolyResistancePercent:         statusTemplate.HolyResistancePercent,
			ShadowResistancePercent:       statusTemplate.ShadowResistancePercent,
			StartedAt:                     now,
			ExpiresAt:                     now.Add(time.Duration(statusTemplate.DurationSeconds) * time.Second),
		}
		if err := s.dbClient.MonsterStatus().Create(ctx, status); err != nil {
			return nil, err
		}
		applied = append(applied, scenarioAppliedFailureStatus{
			Name:            status.Name,
			Description:     status.Description,
			Effect:          status.Effect,
			EffectType:      string(status.EffectType),
			Positive:        status.Positive,
			DamagePerTick:   status.DamagePerTick,
			HealthPerTick:   status.HealthPerTick,
			ManaPerTick:     0,
			DurationSeconds: statusTemplate.DurationSeconds,
		})
	}
	return applied, nil
}

func (s *server) executeMonsterBattleAction(
	ctx context.Context,
	battle *models.MonsterBattle,
	monster *models.Monster,
) (*monsterBattleActionSummary, []monsterBattleUserResource, error) {
	if battle == nil || monster == nil {
		return nil, nil, nil
	}
	resources, err := s.loadMonsterBattleUserResources(ctx, battle)
	if err != nil {
		return nil, nil, err
	}
	statusBonuses, err := s.dbClient.MonsterStatus().GetActiveStatBonuses(ctx, battle.ID)
	if err != nil {
		return nil, nil, err
	}
	totalCombatBonuses := statusBonuses
	if monster.Template != nil {
		totalCombatBonuses = monster.Template.AffinityBonuses().Add(totalCombatBonuses)
	}
	currentMonsterHealth := maxInt(0, monster.DerivedMaxHealthWithBonuses(statusBonuses)-battle.MonsterHealthDeficit)
	maxMonsterHealth := maxInt(1, monster.DerivedMaxHealthWithBonuses(statusBonuses))
	maxMonsterMana := maxInt(0, monster.DerivedMaxManaWithBonuses(statusBonuses))
	currentMonsterMana := maxInt(0, maxMonsterMana-battle.MonsterManaDeficit)
	battleScalingLevel, err := s.monsterBattleScalingLevel(ctx, battle)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	ability := chooseMonsterBattleAbility(
		monster,
		battle,
		currentMonsterHealth,
		maxMonsterHealth,
		currentMonsterMana,
		battleScalingLevel,
		now,
	)
	if ability == nil {
		log.Printf(
			"[party-combat][monster-ai] battle=%s monster=%s action=attack reason=no-usable-abilities currentMana=%d maxMana=%d cooldowns=%d",
			battle.ID,
			monster.ID,
			currentMonsterMana,
			maxMonsterMana,
			len(battle.MonsterAbilityCooldowns),
		)
		target := chooseMonsterBattleTarget(resources)
		if target == nil {
			return nil, resources, nil
		}
		damageMin, damageMax, swipesPerAttack := monster.DerivedAttackProfileWithBonuses(statusBonuses)
		totalDamage := 0
		for i := 0; i < swipesPerAttack; i++ {
			if damageMax <= damageMin {
				totalDamage += damageMin
				continue
			}
			totalDamage += damageMin + rand.Intn(damageMax-damageMin+1)
		}
		damageAffinity := monsterBasicAttackAffinity(monster)
		damageWithBonus, _, _ := applyAffinityDamageBonus(
			totalDamage,
			damageAffinity,
			totalCombatBonuses,
		)
		appliedDamage, _, _ := applyCharacterAffinityResistance(
			damageWithBonus,
			damageAffinity,
			target.Bonuses,
		)
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, target.UserID, appliedDamage, 0); err != nil {
			return nil, nil, err
		}
		updatedResources, err := s.loadMonsterBattleUserResources(ctx, battle)
		if err != nil {
			return nil, nil, err
		}
		return &monsterBattleActionSummary{
			ActionType:       "attack",
			ActorMonsterID:   monster.ID,
			ActorMonsterName: strings.TrimSpace(monster.Name),
			TargetUserID:     &target.UserID,
			TargetUserIDs:    []uuid.UUID{target.UserID},
			Damage:           appliedDamage,
		}, updatedResources, nil
	}

	summary := &monsterBattleActionSummary{
		ActionType:       "ability",
		ActorMonsterID:   monster.ID,
		ActorMonsterName: strings.TrimSpace(monster.Name),
		AbilityID:        &ability.ID,
		AbilityName:      strings.TrimSpace(ability.Name),
		AbilityType:      string(models.NormalizeSpellAbilityType(string(ability.AbilityType))),
	}
	log.Printf(
		"[party-combat][monster-ai] battle=%s monster=%s action=ability ability=%s manaCost=%d currentMana=%d cooldownTurns=%d",
		battle.ID,
		monster.ID,
		ability.ID,
		ability.ManaCost,
		currentMonsterMana,
		ability.CooldownTurns,
	)

	healAmount := adjustedMonsterAbilityHealingForCombat(
		monster,
		battle,
		ability,
		maxMonsterHealth,
	)
	if normalizeSpellAbilityType(string(ability.AbilityType)) != models.SpellAbilityTypeTechnique &&
		ability.ManaCost > 0 {
		battle.MonsterManaDeficit += ability.ManaCost
		if battle.MonsterManaDeficit > maxMonsterMana {
			battle.MonsterManaDeficit = maxMonsterMana
		}
	}
	cooldowns := normalizeMonsterAbilityCooldowns(battle.MonsterAbilityCooldowns, now)
	if ability.CooldownTurns > 0 {
		if cooldowns == nil {
			cooldowns = models.MonsterBattleAbilityCooldowns{}
		}
		if expiresAt := cooldownExpiresAtFromTurns(ability.CooldownTurns, now); expiresAt != nil {
			cooldowns[ability.ID.String()] = *expiresAt
		}
	}
	if healAmount > 0 {
		cooldowns = applyMonsterHealingLockout(monster, cooldowns, now)
	}
	battle.MonsterAbilityCooldowns = cooldowns
	if err := s.dbClient.MonsterBattle().UpdateMonsterCombatState(
		ctx,
		battle.ID,
		battle.MonsterManaDeficit,
		battle.MonsterAbilityCooldowns,
	); err != nil {
		return nil, nil, err
	}

	targetAll := monsterAbilityTargetsAllEnemies(ability)
	targetIDs := make([]uuid.UUID, 0, len(resources))
	if targetAll {
		for _, resource := range resources {
			if resource.Health > 0 {
				targetIDs = append(targetIDs, resource.UserID)
			}
		}
	} else if target := chooseMonsterBattleTarget(resources); target != nil {
		targetIDs = append(targetIDs, target.UserID)
		summary.TargetUserID = &target.UserID
	}
	if len(targetIDs) > 0 {
		summary.TargetUserIDs = append(summary.TargetUserIDs, targetIDs...)
	}
	resourceByUserID := make(map[uuid.UUID]monsterBattleUserResource, len(resources))
	for _, resource := range resources {
		resourceByUserID[resource.UserID] = resource
	}

	damage := monsterAbilityDamageForCombat(monster, ability, battleScalingLevel)
	if damage > 0 && len(targetIDs) > 0 {
		damageAffinity := monsterAbilityDamageAffinity(ability)
		damageWithBonus, _, _ := applyAffinityDamageBonus(
			damage,
			damageAffinity,
			totalCombatBonuses,
		)
		maxAppliedDamage := 0
		for _, userID := range targetIDs {
			resource := resourceByUserID[userID]
			damageAgainstTarget := capMonsterAbilityDamageAgainstHealth(
				damageWithBonus,
				monster,
				ability,
				battleScalingLevel,
				resource.MaxHealth,
			)
			appliedDamage, _, _ := applyCharacterAffinityResistance(
				damageAgainstTarget,
				damageAffinity,
				resource.Bonuses,
			)
			if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, userID, appliedDamage, 0); err != nil {
				return nil, nil, err
			}
			if appliedDamage > maxAppliedDamage {
				maxAppliedDamage = appliedDamage
			}
		}
		summary.Damage = maxAppliedDamage
	}

	if healAmount > 0 {
		if err := s.dbClient.MonsterBattle().AdjustMonsterHealthDeficit(ctx, battle.ID, -healAmount); err != nil {
			return nil, nil, err
		}
		summary.Heal = healAmount
	}

	for _, effect := range ability.Effects {
		switch effect.Type {
		case models.SpellEffectTypeApplyDetrimentalStatus, models.SpellEffectTypeApplyDetrimentalAll:
			statusTemplates := normalizeSpellStatusesForEffectType(effect.Type, effect.StatusesToApply)
			applied, err := s.applyMonsterBattleUserStatuses(ctx, targetIDs, statusTemplates)
			if err != nil {
				return nil, nil, err
			}
			summary.UserStatusesApplied = append(summary.UserStatusesApplied, applied...)
		case models.SpellEffectTypeApplyBeneficialStatus:
			applied, err := s.applyMonsterBattleMonsterStatuses(ctx, battle, monster, effect.StatusesToApply)
			if err != nil {
				return nil, nil, err
			}
			summary.MonsterStatusesApplied = append(summary.MonsterStatusesApplied, applied...)
		case models.SpellEffectTypeRemoveDetrimental:
			names := normalizeSpellStatusNames(effect.StatusesToRemove)
			if len(names) == 0 {
				continue
			}
			if err := s.dbClient.MonsterStatus().DeleteActiveByBattleIDAndNames(ctx, battle.ID, []string(names)); err != nil {
				return nil, nil, err
			}
			summary.MonsterStatusesRemoved = append(summary.MonsterStatusesRemoved, []string(names)...)
		}
	}

	updatedResources, err := s.loadMonsterBattleUserResources(ctx, battle)
	if err != nil {
		return nil, nil, err
	}
	return summary, updatedResources, nil
}

func (s *server) monsterBattleDetailResponse(
	ctx context.Context,
	battle *models.MonsterBattle,
) (*monsterBattleDetail, error) {
	if battle == nil {
		return &monsterBattleDetail{
			Battle:               nil,
			Participants:         []monsterBattleParticipantSummary{},
			ParticipantRewards:   []monsterBattleParticipantRewardSummary{},
			ParticipantResources: []monsterBattleUserResource{},
			Invites:              []monsterBattleInviteSummary{},
			TurnOrder:            []monsterBattleTurnOrderEntry{},
		}, nil
	}

	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	invites, err := s.dbClient.MonsterBattleInvite().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	pendingCount, err := s.dbClient.MonsterBattleInvite().CountPendingByBattle(ctx, battle.ID, now)
	if err != nil {
		return nil, err
	}

	participantSummaries := make([]monsterBattleParticipantSummary, 0, len(participants))
	participantRewards := make([]monsterBattleParticipantRewardSummary, 0, len(participants))
	for _, participant := range participants {
		participantSummaries = append(participantSummaries, monsterBattleParticipantSummary{
			UserID:      participant.UserID,
			IsInitiator: participant.IsInitiator,
			JoinedAt:    participant.JoinedAt,
		})
		participantRewards = append(participantRewards, monsterBattleParticipantRewardSummary{
			UserID:               participant.UserID,
			RewardExperience:     participant.RewardExperience,
			RewardGold:           participant.RewardGold,
			ItemsAwarded:         append([]models.ItemAwarded{}, participant.ItemsAwarded...),
			BaseResourcesAwarded: append([]models.BaseResourceDelta{}, participant.BaseResourcesAwarded...),
		})
	}

	inviteSummaries := make([]monsterBattleInviteSummary, 0, len(invites))
	for _, invite := range invites {
		inviteSummaries = append(inviteSummaries, monsterBattleInviteSummary{
			ID:            invite.ID,
			BattleID:      invite.BattleID,
			InviterUserID: invite.InviterUserID,
			InviteeUserID: invite.InviteeUserID,
			MonsterID:     invite.MonsterID,
			Status:        invite.Status,
			ExpiresAt:     invite.ExpiresAt,
			RespondedAt:   invite.RespondedAt,
		})
	}

	turnOrder, err := s.buildMonsterBattleTurnOrder(ctx, battle, participants)
	if err != nil {
		return nil, err
	}
	participantResources, err := s.loadMonsterBattleUserResources(ctx, battle)
	if err != nil {
		return nil, err
	}

	return &monsterBattleDetail{
		Battle:               monsterBattleResponseFrom(battle),
		Participants:         participantSummaries,
		ParticipantRewards:   participantRewards,
		ParticipantResources: participantResources,
		Invites:              inviteSummaries,
		PendingResponses:     pendingCount,
		TurnOrder:            turnOrder,
	}, nil
}

func splitRewardEvenly(total int, count int) []int {
	if total <= 0 || count <= 0 {
		return make([]int, max(0, count))
	}
	base := total / count
	remainder := total % count
	out := make([]int, count)
	for i := 0; i < count; i++ {
		out[i] = base
		if i < remainder {
			out[i] += 1
		}
	}
	return out
}

func monsterRewardItemsToScenarioRewards(rewards []models.MonsterItemReward) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID <= 0 {
			continue
		}
		quantity := reward.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        quantity,
		})
	}
	return out
}

func (s *server) finalizeMonsterBattleIfDefeated(
	ctx context.Context,
	battle *models.MonsterBattle,
) (*models.MonsterBattle, error) {
	if battle == nil || battle.EndedAt != nil {
		return battle, nil
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, battle.MonsterID)
	if err != nil {
		return nil, err
	}
	battleMonster, err := s.monsterScaledForBattle(ctx, battle, monster)
	if err != nil {
		return nil, err
	}
	statusBonuses, err := s.dbClient.MonsterStatus().GetActiveStatBonuses(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	monsterMaxHealth := battleMonster.DerivedMaxHealthWithBonuses(statusBonuses)
	if monsterMaxHealth <= 0 {
		monsterMaxHealth = 1
	}
	if battle.MonsterHealthDeficit < monsterMaxHealth {
		return battle, nil
	}
	log.Printf(
		"[monster-rewards][finalize][start] battle=%s monster=%s encounter=%v deficit=%d maxHealth=%d",
		battle.ID,
		battle.MonsterID,
		battle.MonsterEncounterID,
		battle.MonsterHealthDeficit,
		monsterMaxHealth,
	)

	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	if len(participants) == 0 {
		participant := models.MonsterBattleParticipant{
			BattleID:    battle.ID,
			UserID:      battle.UserID,
			IsInitiator: true,
			JoinedAt:    time.Now(),
		}
		if err := s.dbClient.MonsterBattleParticipant().CreateOrUpdate(ctx, &participant); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	var encounter *models.MonsterEncounter
	if battle.MonsterEncounterID != nil && *battle.MonsterEncounterID != uuid.Nil {
		encounter, err = s.dbClient.MonsterEncounter().FindByID(ctx, *battle.MonsterEncounterID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if encounter == nil {
		encounter, err = s.dbClient.MonsterEncounter().FindFirstByMonsterID(ctx, monster.ID)
		if err != nil {
			return nil, err
		}
	}
	resolvedEncounterID := uuid.Nil
	if encounter != nil {
		resolvedEncounterID = encounter.ID
	}
	log.Printf(
		"[monster-rewards][finalize][encounter] battle=%s resolvedEncounter=%s participants=%d",
		battle.ID,
		resolvedEncounterID,
		len(participants),
	)
	participantCount := len(participants)
	expRewards := splitRewardEvenly(max(0, monster.RewardExperience), participantCount)
	goldRewards := splitRewardEvenly(max(0, monster.RewardGold), participantCount)
	itemRewards := monsterRewardItemsToScenarioRewards(monster.ItemRewards)

	for index, participant := range participants {
		rewardExperience := expRewards[index]
		rewardGold := goldRewards[index]
		resolvedItemRewards := itemRewards
		if encounter != nil {
			_, _, rewardExperience, rewardGold, _, resolvedItemRewards, err = s.resolveMonsterEncounterRewardsForUser(
				ctx,
				participant.UserID,
				encounter,
			)
			if err != nil {
				return nil, err
			}
		}
		log.Printf(
			"[monster-rewards][finalize][resolved] battle=%s user=%s encounter=%s rewardExperience=%d rewardGold=%d itemRewardCount=%d",
			battle.ID,
			participant.UserID,
			resolvedEncounterID,
			rewardExperience,
			rewardGold,
			len(resolvedItemRewards),
		)
		itemsAwarded, _, err := s.awardScenarioRewards(
			ctx,
			participant.UserID,
			rewardExperience,
			rewardGold,
			resolvedItemRewards,
			[]scenarioRewardSpell{},
			[]string{},
		)
		if err != nil {
			return nil, err
		}
		rewardMode := monster.RewardMode
		sourceType := "monster"
		sourceID := monster.ID
		materialRewards := monster.MaterialRewards
		if encounter != nil {
			rewardMode = encounter.RewardMode
			sourceType = "monster_encounter"
			sourceID = encounter.ID
			materialRewards = encounter.MaterialRewards
		}
		baseResourcesAwarded, err := s.awardBaseResourcesToUser(
			ctx,
			participant.UserID,
			resolveBaseMaterialRewards(
				rewardMode,
				materialRewards,
				fmt.Sprintf("%s:%s:user:%s:materials", sourceType, sourceID, participant.UserID),
			),
			sourceType,
			&sourceID,
		)
		if err != nil {
			return nil, err
		}
		log.Printf(
			"[monster-rewards][finalize][awarded] battle=%s user=%s itemsAwarded=%d rewardExperience=%d rewardGold=%d",
			battle.ID,
			participant.UserID,
			len(itemsAwarded),
			rewardExperience,
			rewardGold,
		)
		if err := s.dbClient.MonsterBattleParticipant().UpdateRewards(
			ctx,
			battle.ID,
			participant.UserID,
			rewardExperience,
			rewardGold,
			itemsAwarded,
			baseResourcesAwarded,
		); err != nil {
			return nil, err
		}
		log.Printf(
			"[combat][defeat-recovery][finalize-battle] battle=%s user=%s outcome=victory-or-cleanup",
			battle.ID,
			participant.UserID,
		)
		if err := s.restoreUserToOneHealthIfDowned(ctx, participant.UserID); err != nil {
			return nil, err
		}
	}

	if err := s.dbClient.MonsterStatus().DeleteAllForBattleID(ctx, battle.ID); err != nil {
		return nil, err
	}
	endedAt := time.Now()
	if err := s.dbClient.MonsterBattle().End(ctx, battle.ID, endedAt); err != nil {
		return nil, err
	}
	battle.EndedAt = &endedAt
	battle.LastActivityAt = endedAt
	participantIDs := make([]uuid.UUID, 0, len(participants))
	for _, participant := range participants {
		participantIDs = append(participantIDs, participant.UserID)
	}
	if encounter != nil {
		for _, participant := range participants {
			if err := s.dbClient.UserMonsterEncounterVictory().Upsert(
				ctx,
				participant.UserID,
				encounter.ID,
			); err != nil {
				return nil, err
			}
		}
	}
	if err := s.completeQuestMonsterObjectives(
		ctx,
		participantIDs,
		encounter,
		monster.ID,
	); err != nil {
		return nil, err
	}
	if monster.OwnerUserID != nil {
		if encounter != nil {
			if err := s.dbClient.Tutorial().MarkMonsterCompleted(ctx, battle.UserID, encounter.ID); err != nil {
				return nil, err
			}
		}
	}
	log.Printf(
		"[monster-rewards][finalize][done] battle=%s encounter=%s participants=%d endedAt=%s",
		battle.ID,
		resolvedEncounterID,
		len(participants),
		endedAt.Format(time.RFC3339),
	)
	return battle, nil
}

func (s *server) completeQuestMonsterObjectives(
	ctx context.Context,
	userIDs []uuid.UUID,
	encounter *models.MonsterEncounter,
	defeatedMonsterID uuid.UUID,
) error {
	if len(userIDs) == 0 {
		return nil
	}

	now := time.Now()
	sharedQuestNodeIDs := map[uuid.UUID]struct{}{}
	for _, userID := range userIDs {
		user, err := s.dbClient.User().FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			continue
		}

		acceptances, err := s.dbClient.QuestAcceptanceV2().FindByUserID(ctx, userID)
		if err != nil {
			return err
		}
		for _, acceptance := range acceptances {
			if acceptance.IsClosed() {
				continue
			}
			quest, err := s.dbClient.Quest().FindByID(ctx, acceptance.QuestID)
			if err != nil {
				return err
			}
			if quest == nil {
				continue
			}
			currentNode, err := s.currentQuestNode(ctx, quest, &acceptance)
			if err != nil {
				return err
			}
			if currentNode == nil || !questNodeMatchesMonsterVictory(currentNode, encounter, defeatedMonsterID) {
				continue
			}

			completedNode, err := s.markQuestNodeCompleteForAcceptance(
				ctx,
				quest,
				&acceptance,
				currentNode.ID,
				now,
			)
			if err != nil {
				return err
			}
			if !completedNode {
				continue
			}

			if _, exists := sharedQuestNodeIDs[currentNode.ID]; exists {
				continue
			}
			sharedQuestNodeIDs[currentNode.ID] = struct{}{}
			s.shareQuestNodeCompletionWithEligiblePartyMembers(ctx, user, quest, currentNode)
		}
	}

	return nil
}

func questNodeMatchesMonsterVictory(
	node *models.QuestNode,
	encounter *models.MonsterEncounter,
	defeatedMonsterID uuid.UUID,
) bool {
	if node == nil {
		return false
	}
	if encounter != nil && node.MonsterEncounterID != nil && *node.MonsterEncounterID == encounter.ID {
		return true
	}
	if node.MonsterID != nil && *node.MonsterID == defeatedMonsterID {
		return true
	}
	return false
}

func (s *server) getMonsterBattleInvites(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	if _, err := s.dbClient.MonsterBattleInvite().AutoDeclineExpiredByInvitee(ctx, user.ID, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	invites, err := s.dbClient.MonsterBattleInvite().FindPendingByInvitee(ctx, user.ID, now)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	summaries := make([]monsterBattleInviteSummary, 0, len(invites))
	for _, invite := range invites {
		summaries = append(summaries, monsterBattleInviteSummary{
			ID:            invite.ID,
			BattleID:      invite.BattleID,
			InviterUserID: invite.InviterUserID,
			InviteeUserID: invite.InviteeUserID,
			MonsterID:     invite.MonsterID,
			Status:        invite.Status,
			ExpiresAt:     invite.ExpiresAt,
			RespondedAt:   invite.RespondedAt,
		})
	}
	ctx.JSON(http.StatusOK, summaries)
}

func (s *server) acceptMonsterBattleInvite(ctx *gin.Context) {
	s.respondToMonsterBattleInvite(ctx, string(models.MonsterBattleInviteStatusAccepted))
}

func (s *server) rejectMonsterBattleInvite(ctx *gin.Context) {
	s.respondToMonsterBattleInvite(ctx, string(models.MonsterBattleInviteStatusDeclined))
}

func (s *server) respondToMonsterBattleInvite(ctx *gin.Context, status string) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var requestBody struct {
		InviteID string `json:"inviteID"`
	}
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviteID, err := uuid.Parse(strings.TrimSpace(requestBody.InviteID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite ID"})
		return
	}

	invite, err := s.dbClient.MonsterBattleInvite().FindByID(ctx, inviteID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if invite == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	if invite.InviteeUserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not invitee"})
		return
	}

	now := time.Now()
	if _, err := s.dbClient.MonsterBattleInvite().AutoDeclineExpiredByBattle(ctx, invite.BattleID, now); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := s.dbClient.MonsterBattleInvite().UpdateStatus(ctx, inviteID, user.ID, status, &now)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rowsAffected == 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "invite is no longer pending"})
		return
	}

	if status == string(models.MonsterBattleInviteStatusAccepted) {
		ownerBattle, err := s.dbClient.MonsterBattle().FindActiveByUserAndMonster(ctx, user.ID, invite.MonsterID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ownerBattle != nil && ownerBattle.ID != invite.BattleID {
			if err := s.dbClient.MonsterStatus().DeleteAllForBattleID(ctx, ownerBattle.ID); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if err := s.dbClient.MonsterBattle().End(ctx, ownerBattle.ID, now); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			log.Printf(
				"[party-combat][invite] ended conflicting owner battle user=%s monster=%s oldBattle=%s invitedBattle=%s",
				user.ID,
				invite.MonsterID,
				ownerBattle.ID,
				invite.BattleID,
			)
		}

		if err := s.dbClient.MonsterBattleParticipant().CreateOrUpdate(ctx, &models.MonsterBattleParticipant{
			BattleID:    invite.BattleID,
			UserID:      user.ID,
			IsInitiator: false,
			JoinedAt:    now,
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	battle, err := s.refreshMonsterBattleInviteState(ctx, invite.BattleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	detail, err := s.monsterBattleDetailResponse(ctx, battle)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf(
		"[party-combat][invite] response user=%s invite=%s status=%s battle=%s pending=%d participants=%d",
		user.ID,
		invite.ID,
		status,
		invite.BattleID,
		detail.PendingResponses,
		len(detail.Participants),
	)
	ctx.JSON(http.StatusOK, detail)
}
