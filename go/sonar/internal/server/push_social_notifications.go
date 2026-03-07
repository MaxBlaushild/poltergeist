package server

import (
	"context"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func (s *server) sendPartyInvitePushNotification(
	ctx context.Context,
	invite *models.PartyInvite,
	inviter *models.User,
) {
	if invite == nil {
		log.Printf("[push][party-invite-social] skipped: invite is nil")
		return
	}
	data := map[string]string{
		"type":      "party_invite",
		"inviteId":  invite.ID.String(),
		"inviterId": invite.InviterID.String(),
		"inviteeId": invite.InviteeID.String(),
		"sentAt":    time.Now().UTC().Format(time.RFC3339),
	}
	s.sendSocialPushToUser(
		ctx,
		"party-invite-social",
		invite.InviteeID,
		"Party Invite",
		userDisplayName(inviter)+" invited you to join their party.",
		data,
	)
}

func (s *server) sendFriendInvitePushNotification(
	ctx context.Context,
	invite *models.FriendInvite,
	inviter *models.User,
) {
	if invite == nil {
		log.Printf("[push][friend-invite] skipped: invite is nil")
		return
	}
	data := map[string]string{
		"type":      "friend_invite",
		"inviteId":  invite.ID.String(),
		"inviterId": invite.InviterID.String(),
		"inviteeId": invite.InviteeID.String(),
		"sentAt":    time.Now().UTC().Format(time.RFC3339),
	}
	s.sendSocialPushToUser(
		ctx,
		"friend-invite",
		invite.InviteeID,
		"Friend Invite",
		userDisplayName(inviter)+" sent you a friend invite.",
		data,
	)
}

func (s *server) sendFriendInviteAcceptedPushNotification(
	ctx context.Context,
	invite *models.FriendInvite,
	accepter *models.User,
) {
	if invite == nil {
		log.Printf("[push][friend-invite-accepted] skipped: invite is nil")
		return
	}
	data := map[string]string{
		"type":      "friend_invite_accepted",
		"inviteId":  invite.ID.String(),
		"inviterId": invite.InviterID.String(),
		"inviteeId": invite.InviteeID.String(),
		"sentAt":    time.Now().UTC().Format(time.RFC3339),
	}
	s.sendSocialPushToUser(
		ctx,
		"friend-invite-accepted",
		invite.InviterID,
		"Friend Invite Accepted",
		userDisplayName(accepter)+" accepted your friend invite.",
		data,
	)
}

func (s *server) sendSocialPushToUser(
	ctx context.Context,
	logScope string,
	userID uuid.UUID,
	title string,
	body string,
	data map[string]string,
) {
	if s.pushClient == nil {
		log.Printf(
			"[push][%s] skipped: push client not configured user=%s",
			logScope,
			userID,
		)
		return
	}
	tokens, err := s.dbClient.UserDeviceToken().FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("[push][%s] failed to fetch tokens user=%s: %v", logScope, userID, err)
		return
	}
	if len(tokens) == 0 {
		log.Printf("[push][%s] skipped: no tokens user=%s", logScope, userID)
		return
	}
	log.Printf("[push][%s] sending push user=%s tokens=%d", logScope, userID, len(tokens))

	sentCount := 0
	failedCount := 0
	for _, token := range tokens {
		if err := s.pushClient.Send(ctx, token.Token, title, body, data); err != nil {
			failedCount++
			log.Printf(
				"[push][%s] send failed user=%s platform=%s token=%s: %v",
				logScope,
				userID,
				token.Platform,
				tokenPreview(token.Token),
				err,
			)
			continue
		}
		sentCount++
	}
	log.Printf(
		"[push][%s] send complete user=%s sent=%d failed=%d",
		logScope,
		userID,
		sentCount,
		failedCount,
	)
}
