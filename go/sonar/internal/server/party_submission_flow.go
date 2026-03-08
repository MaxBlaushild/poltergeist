package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	partySubmissionContentTypeScenario  = "scenario"
	partySubmissionContentTypeChallenge = "challenge"

	partySubmissionStatusProcessing = "processing"
	partySubmissionStatusCompleted  = "completed"

	partySubmissionProcessingTTL  = 5 * time.Minute
	partySubmissionCompletedTTL   = 24 * time.Hour
	partySubmissionResultQueueTTL = 48 * time.Hour
)

type partySubmissionLockState struct {
	PartyID           uuid.UUID  `json:"partyId"`
	ContentType       string     `json:"contentType"`
	ContentID         uuid.UUID  `json:"contentId"`
	Status            string     `json:"status"`
	SubmittedByUserID uuid.UUID  `json:"submittedByUserId"`
	SubmittedAt       time.Time  `json:"submittedAt"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
}

type partySubmissionResultEnvelope struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"createdAt"`
}

func partySubmissionLockRedisKey(
	partyID uuid.UUID,
	contentType string,
	contentID uuid.UUID,
) string {
	return fmt.Sprintf(
		"party_submission_lock:%s:%s:%s",
		partyID.String(),
		contentType,
		contentID.String(),
	)
}

func partySubmissionResultQueueRedisKey(userID uuid.UUID) string {
	return fmt.Sprintf("party_submission_results:%s", userID.String())
}

func normalizePartySubmissionContentType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case partySubmissionContentTypeScenario:
		return partySubmissionContentTypeScenario
	case partySubmissionContentTypeChallenge:
		return partySubmissionContentTypeChallenge
	default:
		return ""
	}
}

func partySubmissionStatusPayloadFromLock(
	lock *partySubmissionLockState,
	locked bool,
) map[string]interface{} {
	payload := map[string]interface{}{
		"locked": locked,
	}
	if lock == nil {
		return payload
	}
	payload["contentType"] = lock.ContentType
	payload["contentId"] = lock.ContentID.String()
	payload["status"] = lock.Status
	payload["submittedByUserId"] = lock.SubmittedByUserID.String()
	payload["submittedAt"] = lock.SubmittedAt.UTC().Format(time.RFC3339)
	if lock.CompletedAt != nil {
		payload["completedAt"] = lock.CompletedAt.UTC().Format(time.RFC3339)
	}
	return payload
}

func partySubmissionParticipantIDs(users []models.User) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(users))
	seen := map[uuid.UUID]struct{}{}
	for _, user := range users {
		if user.ID == uuid.Nil {
			continue
		}
		if _, exists := seen[user.ID]; exists {
			continue
		}
		seen[user.ID] = struct{}{}
		out = append(out, user.ID)
	}
	return out
}

func (s *server) resolvePartySubmissionParticipants(
	ctx context.Context,
	userID uuid.UUID,
) (*uuid.UUID, []models.User, error) {
	user, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	participants := []models.User{*user}
	if user.PartyID == nil {
		return nil, participants, nil
	}

	members, err := s.dbClient.User().FindPartyMembers(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	seen := map[uuid.UUID]struct{}{
		user.ID: {},
	}
	for _, member := range members {
		if member.ID == uuid.Nil {
			continue
		}
		if _, exists := seen[member.ID]; exists {
			continue
		}
		seen[member.ID] = struct{}{}
		participants = append(participants, member)
	}

	return user.PartyID, participants, nil
}

func (s *server) readPartySubmissionLock(
	ctx context.Context,
	partyID uuid.UUID,
	contentType string,
	contentID uuid.UUID,
) (*partySubmissionLockState, error) {
	if s.redisClient == nil {
		return nil, nil
	}
	key := partySubmissionLockRedisKey(partyID, contentType, contentID)
	raw, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		// redis.Nil is returned as an error here too.
		return nil, nil
	}
	lock := &partySubmissionLockState{}
	if err := json.Unmarshal([]byte(raw), lock); err != nil {
		return nil, nil
	}
	return lock, nil
}

func (s *server) acquirePartySubmissionLock(
	ctx context.Context,
	partyID uuid.UUID,
	contentType string,
	contentID uuid.UUID,
	submittedByUserID uuid.UUID,
) (*partySubmissionLockState, bool, error) {
	if s.redisClient == nil {
		return nil, true, nil
	}

	now := time.Now().UTC()
	lock := &partySubmissionLockState{
		PartyID:           partyID,
		ContentType:       contentType,
		ContentID:         contentID,
		Status:            partySubmissionStatusProcessing,
		SubmittedByUserID: submittedByUserID,
		SubmittedAt:       now,
	}

	payload, err := json.Marshal(lock)
	if err != nil {
		return nil, false, err
	}

	key := partySubmissionLockRedisKey(partyID, contentType, contentID)
	acquired, err := s.redisClient.SetNX(ctx, key, string(payload), partySubmissionProcessingTTL).Result()
	if err != nil {
		return nil, false, err
	}
	if acquired {
		return lock, true, nil
	}

	existing, err := s.readPartySubmissionLock(ctx, partyID, contentType, contentID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}
	return nil, false, nil
}

func (s *server) markPartySubmissionLockCompleted(
	ctx context.Context,
	partyID uuid.UUID,
	contentType string,
	contentID uuid.UUID,
	submittedByUserID uuid.UUID,
	submittedAt time.Time,
) error {
	if s.redisClient == nil {
		return nil
	}

	completedAt := time.Now().UTC()
	lock := &partySubmissionLockState{
		PartyID:           partyID,
		ContentType:       contentType,
		ContentID:         contentID,
		Status:            partySubmissionStatusCompleted,
		SubmittedByUserID: submittedByUserID,
		SubmittedAt:       submittedAt.UTC(),
		CompletedAt:       &completedAt,
	}

	payload, err := json.Marshal(lock)
	if err != nil {
		return err
	}
	key := partySubmissionLockRedisKey(partyID, contentType, contentID)
	return s.redisClient.Set(ctx, key, string(payload), partySubmissionCompletedTTL).Err()
}

func (s *server) releasePartySubmissionLock(
	ctx context.Context,
	partyID uuid.UUID,
	contentType string,
	contentID uuid.UUID,
) error {
	if s.redisClient == nil {
		return nil
	}
	key := partySubmissionLockRedisKey(partyID, contentType, contentID)
	return s.redisClient.Del(ctx, key).Err()
}

func (s *server) queuePartySubmissionResultForUser(
	ctx context.Context,
	userID uuid.UUID,
	contentType string,
	contentID uuid.UUID,
	envelope partySubmissionResultEnvelope,
) error {
	title := "Party Submission Result"
	body := "A party member resolved an encounter. Tap to view your result."
	switch envelope.Type {
	case "scenarioOutcome":
		title = "Scenario Result Ready"
		body = "A party member resolved a scenario. View the roll and rewards."
	case "challengeOutcome":
		title = "Challenge Result Ready"
		body = "A party member resolved a challenge. View the roll and rewards."
	}
	data := map[string]string{
		"type":        "party_submission_result",
		"resultId":    envelope.ID.String(),
		"resultType":  envelope.Type,
		"contentType": contentType,
		"contentId":   contentID.String(),
	}

	if s.redisClient == nil {
		s.sendSocialPushToUser(
			ctx,
			"party-submission-result",
			userID,
			title,
			body,
			data,
		)
		return nil
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	queueKey := partySubmissionResultQueueRedisKey(userID)
	pipe := s.redisClient.TxPipeline()
	pipe.RPush(ctx, queueKey, string(raw))
	pipe.Expire(ctx, queueKey, partySubmissionResultQueueTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	s.sendSocialPushToUser(
		ctx,
		"party-submission-result",
		userID,
		title,
		body,
		data,
	)
	return nil
}

func (s *server) queuePartySubmissionResultsForParty(
	ctx context.Context,
	recipientUserIDs []uuid.UUID,
	excludeUserID *uuid.UUID,
	contentType string,
	contentID uuid.UUID,
	modalType string,
	modalData map[string]interface{},
) {
	now := time.Now().UTC()
	for _, userID := range recipientUserIDs {
		if excludeUserID != nil && *excludeUserID == userID {
			continue
		}
		err := s.queuePartySubmissionResultForUser(
			ctx,
			userID,
			contentType,
			contentID,
			partySubmissionResultEnvelope{
				ID:        uuid.New(),
				Type:      modalType,
				Data:      modalData,
				CreatedAt: now,
			},
		)
		if err != nil {
			log.Printf(
				"[party-submission][result] failed to queue result user=%s contentType=%s contentId=%s err=%v",
				userID,
				contentType,
				contentID,
				err,
			)
		}
	}
}

func (s *server) getPartySubmissionStatus(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	contentType := normalizePartySubmissionContentType(ctx.Query("contentType"))
	if contentType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid contentType"})
		return
	}
	contentID, err := uuid.Parse(strings.TrimSpace(ctx.Query("contentId")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid contentId"})
		return
	}

	partyID, _, err := s.resolvePartySubmissionParticipants(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if partyID == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"contentType": contentType,
			"contentId":   contentID.String(),
			"locked":      false,
		})
		return
	}

	lock, err := s.readPartySubmissionLock(ctx.Request.Context(), *partyID, contentType, contentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	payload := partySubmissionStatusPayloadFromLock(lock, lock != nil)
	payload["contentType"] = contentType
	payload["contentId"] = contentID.String()
	ctx.JSON(http.StatusOK, payload)
}

func (s *server) getPendingPartySubmissionResults(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if s.redisClient == nil {
		ctx.JSON(http.StatusOK, []partySubmissionResultEnvelope{})
		return
	}

	key := partySubmissionResultQueueRedisKey(user.ID)
	rawItems, err := s.redisClient.LRange(ctx.Request.Context(), key, 0, -1).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(rawItems) > 0 {
		_ = s.redisClient.Del(ctx.Request.Context(), key).Err()
	}

	out := make([]partySubmissionResultEnvelope, 0, len(rawItems))
	for _, raw := range rawItems {
		var result partySubmissionResultEnvelope
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			continue
		}
		out = append(out, result)
	}
	ctx.JSON(http.StatusOK, out)
}

func (s *server) awardScenarioRewardsToParticipants(
	ctx context.Context,
	participantIDs []uuid.UUID,
	submitterID uuid.UUID,
	rewardExperience int,
	rewardGold int,
	rewardItems []scenarioRewardItem,
	rewardSpells []scenarioRewardSpell,
	proficiencies []string,
) ([]models.ItemAwarded, []models.SpellAwarded, error) {
	itemsAwarded := []models.ItemAwarded{}
	spellsAwarded := []models.SpellAwarded{}
	for _, participantID := range participantIDs {
		items, spells, err := s.awardScenarioRewards(
			ctx,
			participantID,
			rewardExperience,
			rewardGold,
			rewardItems,
			rewardSpells,
			proficiencies,
		)
		if err != nil {
			return nil, nil, err
		}
		if participantID == submitterID {
			itemsAwarded = items
			spellsAwarded = spells
		}
	}
	return itemsAwarded, spellsAwarded, nil
}
