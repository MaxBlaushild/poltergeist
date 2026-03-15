package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	monsterBattlePartyInviteRadiusMeters = 50.0
	monsterBattleInviteTTL               = 1 * time.Minute
)

type monsterBattleParticipantSummary struct {
	UserID      uuid.UUID `json:"userId"`
	IsInitiator bool      `json:"isInitiator"`
	JoinedAt    time.Time `json:"joinedAt"`
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
	Battle           *monsterBattleResponse            `json:"battle"`
	Participants     []monsterBattleParticipantSummary `json:"participants"`
	Invites          []monsterBattleInviteSummary      `json:"invites"`
	PendingResponses int64                             `json:"pendingResponses"`
	TurnOrder        []monsterBattleTurnOrderEntry     `json:"turnOrder"`
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
		return s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State)
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
		return s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State)
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
	if monster.OwnerUserID != nil {
		battle.State = string(models.MonsterBattleStateActive)
		return s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State)
	}

	inviteCount := 0
	for _, member := range partyMembers {
		isActive, err := s.livenessClient.HasRecentLocation(ctx, member.ID)
		if err != nil {
			log.Printf(
				"[party-combat][invite] skipped member=%s battle=%s reason=active-check-error err=%v",
				member.ID,
				battle.ID,
				err,
			)
			continue
		}
		if !isActive {
			log.Printf(
				"[party-combat][invite] skipped member=%s battle=%s reason=not-active",
				member.ID,
				battle.ID,
			)
			continue
		}

		memberLat, memberLng, err := s.getUserLatLng(ctx, member.ID)
		if err == nil {
			distanceMeters := util.HaversineDistance(memberLat, memberLng, monster.Latitude, monster.Longitude)
			if distanceMeters > monsterBattlePartyInviteRadiusMeters {
				log.Printf(
					"[party-combat][invite] skipped member=%s battle=%s reason=too-far distance=%.2fm max=%.2fm",
					member.ID,
					battle.ID,
					distanceMeters,
					monsterBattlePartyInviteRadiusMeters,
				)
				continue
			}
		} else {
			// If we know the member is active but cannot read precise location,
			// fail open so combat waits for a response instead of silently bypassing party flow.
			log.Printf(
				"[party-combat][invite] fallback-invite member=%s battle=%s reason=location-unavailable err=%v",
				member.ID,
				battle.ID,
				err,
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
	return s.dbClient.MonsterBattle().SetState(ctx, battle.ID, battle.State)
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
	}
	return s.dbClient.MonsterBattle().FindByID(ctx, battleID)
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

func (s *server) monsterBattleDetailResponse(
	ctx context.Context,
	battle *models.MonsterBattle,
) (*monsterBattleDetail, error) {
	if battle == nil {
		return &monsterBattleDetail{
			Battle:       nil,
			Participants: []monsterBattleParticipantSummary{},
			Invites:      []monsterBattleInviteSummary{},
			TurnOrder:    []monsterBattleTurnOrderEntry{},
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
	for _, participant := range participants {
		participantSummaries = append(participantSummaries, monsterBattleParticipantSummary{
			UserID:      participant.UserID,
			IsInitiator: participant.IsInitiator,
			JoinedAt:    participant.JoinedAt,
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

	return &monsterBattleDetail{
		Battle:           monsterBattleResponseFrom(battle),
		Participants:     participantSummaries,
		Invites:          inviteSummaries,
		PendingResponses: pendingCount,
		TurnOrder:        turnOrder,
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
	statusBonuses, err := s.dbClient.MonsterStatus().GetActiveStatBonuses(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	monsterMaxHealth := monster.DerivedMaxHealthWithBonuses(statusBonuses)
	if monsterMaxHealth <= 0 {
		monsterMaxHealth = 1
	}
	if battle.MonsterHealthDeficit < monsterMaxHealth {
		return battle, nil
	}

	participants, err := s.dbClient.MonsterBattleParticipant().FindByBattleID(ctx, battle.ID)
	if err != nil {
		return nil, err
	}
	if len(participants) == 0 {
		participants = append(participants, models.MonsterBattleParticipant{
			BattleID:    battle.ID,
			UserID:      battle.UserID,
			IsInitiator: true,
			JoinedAt:    time.Now(),
		})
	}

	participantCount := len(participants)
	expRewards := splitRewardEvenly(max(0, monster.RewardExperience), participantCount)
	goldRewards := splitRewardEvenly(max(0, monster.RewardGold), participantCount)
	itemRewards := monsterRewardItemsToScenarioRewards(monster.ItemRewards)

	for index, participant := range participants {
		_, _, err := s.awardScenarioRewards(
			ctx,
			participant.UserID,
			expRewards[index],
			goldRewards[index],
			itemRewards,
			[]scenarioRewardSpell{},
			[]string{},
		)
		if err != nil {
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
	encounter, err := s.dbClient.MonsterEncounter().FindFirstByMonsterID(ctx, monster.ID)
	if err != nil {
		return nil, err
	}
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
			if acceptance.TurnedInAt != nil {
				continue
			}
			quest, err := s.dbClient.Quest().FindByID(ctx, acceptance.QuestID)
			if err != nil {
				return err
			}
			if quest == nil {
				continue
			}
			currentNode, err := s.currentQuestNode(ctx, quest, acceptance.ID)
			if err != nil {
				return err
			}
			if currentNode == nil || !questNodeMatchesMonsterVictory(currentNode, encounter, defeatedMonsterID) {
				continue
			}

			completedNode, err := s.markQuestNodeCompleteForAcceptance(
				ctx,
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
