package travelcalendar

import (
	"errors"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type guestScope struct {
	CalendarID *uuid.UUID `json:"calendarId"`
	StopID     *uuid.UUID `json:"stopId"`
	ShareToken string     `json:"shareToken"`
}

func (scope guestScope) valid() bool {
	return (scope.CalendarID != nil && *scope.CalendarID != uuid.Nil && scope.StopID == nil) || (scope.StopID != nil && *scope.StopID != uuid.Nil && scope.CalendarID == nil)
}
func (scope guestScope) query(tx *gorm.DB) *gorm.DB {
	if scope.StopID != nil {
		return tx.Where("stop_id = ?", *scope.StopID)
	}
	return tx.Where("calendar_id = ?", *scope.CalendarID)
}

func (s *Service) lockGuestScope(tx *gorm.DB, scope guestScope) error {
	if scope.StopID != nil {
		_, err := s.lockedStop(tx, scope.StopID.String())
		return err
	}
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&Calendar{}, "id = ?", *scope.CalendarID).Error
}
func queryGuestScope(c *gin.Context) (guestScope, bool) {
	scope := guestScope{}
	if value := c.Query("calendarId"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return scope, false
		}
		scope.CalendarID = &id
	}
	if value := c.Query("stopId"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return scope, false
		}
		scope.StopID = &id
	}
	return scope, scope.valid()
}

func (s *Service) guestCanManageCalendar(tx *gorm.DB, cal *Calendar, userID uuid.UUID) bool {
	if userID == uuid.Nil {
		return false
	}
	if cal.OwnerID == userID {
		return true
	}
	var count int64
	return tx.Model(&Collaborator{}).Where("calendar_id = ? AND user_id = ? AND accepted = ?", cal.ID, userID, true).Count(&count).Error == nil && count > 0
}

func (s *Service) manageGuestScope(tx *gorm.DB, scope guestScope, viewer Viewer, permission bool) bool {
	if !scope.valid() {
		return false
	}
	if scope.StopID != nil {
		var stop Stop
		if tx.First(&stop, "id = ?", *scope.StopID).Error != nil {
			return false
		}
		if permission {
			return s.canManagePermissions(tx, &stop, viewer)
		}
		return s.canManageStop(tx, &stop, viewer)
	}
	var cal Calendar
	if tx.First(&cal, "id = ?", *scope.CalendarID).Error != nil {
		return false
	}
	if permission {
		return cal.OwnerID == viewer.UserID && viewer.UserID != uuid.Nil
	}
	return s.guestCanManageCalendar(tx, &cal, viewer.UserID)
}

func (s *Service) guestCanViewCalendar(tx *gorm.DB, cal *Calendar, v Viewer) bool {
	if s.guestCanManageCalendar(tx, cal, v.UserID) {
		return true
	}
	if cal.Visibility == "specific" && hasGrant(tx, "calendar_id", cal.ID, v) {
		return true
	}
	if cal.Visibility == "link" && cal.SharingEnabled && v.CalendarLinkID == cal.ID {
		return true
	}
	var stops []Stop
	if tx.Where("published IS NOT NULL AND (calendar_id = ? OR id IN (SELECT stop_id FROM tc_collaborators WHERE user_id = ? AND accepted = true AND stop_id IS NOT NULL))", cal.ID, cal.OwnerID).Find(&stops).Error != nil {
		return false
	}
	for i := range stops {
		if s.canViewStop(tx, &stops[i], v) {
			return true
		}
	}
	return false
}

func recipientViewer(recipient *GuestRecipient) Viewer {
	viewer := Viewer{RecipientID: recipient.ID}
	if recipient.UserID != nil {
		viewer.UserID = *recipient.UserID
	}
	return viewer
}

// trustedLink is only used for a sender's invitation/delivery. Browser mutations
// must prove possession of the supplied current share token before using links.
func (s *Service) guestCanViewScope(tx *gorm.DB, scope guestScope, recipient *GuestRecipient, trustedLink bool) bool {
	if !scope.valid() {
		return false
	}
	viewer := recipientViewer(recipient)
	if scope.StopID != nil {
		var stop Stop
		if tx.First(&stop, "id = ?", *scope.StopID).Error != nil {
			return false
		}
		if trustedLink || (scope.ShareToken != "" && scope.ShareToken == stop.ShareToken) {
			viewer.StopLinkID = stop.ID
		}
		// A stop discovered through an owning calendar can also be followed.
		var cal Calendar
		if tx.First(&cal, "id = ?", stop.CalendarID).Error == nil && (trustedLink || (scope.ShareToken != "" && scope.ShareToken == cal.ShareToken)) {
			viewer.CalendarLinkID = cal.ID
		}
		return s.canViewStop(tx, &stop, viewer)
	}
	var cal Calendar
	if tx.First(&cal, "id = ?", *scope.CalendarID).Error != nil {
		return false
	}
	if trustedLink || (scope.ShareToken != "" && scope.ShareToken == cal.ShareToken) {
		viewer.CalendarLinkID = cal.ID
	}
	return s.guestCanViewCalendar(tx, &cal, viewer)
}

func (s *Service) observedSubscriptionLinks(tx *gorm.DB, scope guestScope) (string, string) {
	if scope.ShareToken == "" {
		return "", ""
	}
	var calendar Calendar
	if scope.CalendarID != nil {
		if tx.First(&calendar, "id = ?", *scope.CalendarID).Error == nil && calendar.SharingEnabled && calendar.ShareToken == scope.ShareToken {
			return guestDigest(scope.ShareToken), ""
		}
		return "", ""
	}
	var stop Stop
	if tx.First(&stop, "id = ?", *scope.StopID).Error != nil {
		return "", ""
	}
	if stop.SharingEnabled && stop.ShareToken == scope.ShareToken {
		return "", guestDigest(scope.ShareToken)
	}
	if tx.First(&calendar, "id = ?", stop.CalendarID).Error == nil && calendar.SharingEnabled && calendar.ShareToken == scope.ShareToken {
		return guestDigest(scope.ShareToken), ""
	}
	return "", ""
}

func (s *Service) listGuestSubscriptions(c *gin.Context) {
	recipient, ok := s.requireGuestRecipient(c)
	if !ok {
		return
	}
	var subscriptions []calendarSubscription
	if s.db.Where("recipient_id = ?", recipient.ID).Order("created_at").Find(&subscriptions).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load update preferences"})
		return
	}
	if subscriptions == nil {
		subscriptions = []calendarSubscription{}
	}
	c.JSON(200, gin.H{"subscriptions": subscriptions})
}

func (s *Service) setGuestSubscription(c *gin.Context) {
	recipient, ok := s.requireGuestRecipient(c)
	if !ok {
		return
	}
	var body struct {
		guestScope
		Subscribed bool `json:"subscribed"`
	}
	if c.ShouldBindJSON(&body) != nil || !body.guestScope.valid() {
		c.JSON(400, gin.H{"error": "Choose one calendar or stop"})
		return
	}
	var subscription calendarSubscription
	denied := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.lockGuestScope(tx, body.guestScope); err != nil {
			denied = true
			return nil
		}
		// Serialize subscriptions for the same contact, including repeat clicks.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&GuestRecipient{}, "id = ?", recipient.ID).Error; err != nil {
			return err
		}
		if body.Subscribed && !s.guestCanViewScope(tx, body.guestScope, recipient, false) {
			denied = true
			return nil
		}
		err := body.guestScope.query(tx).Where("recipient_id = ?", recipient.ID).First(&subscription).Error
		status := "unsubscribed"
		if body.Subscribed {
			status = "active"
		}
		calendarHash, stopHash := s.observedSubscriptionLinks(tx, body.guestScope)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			subscription = calendarSubscription{ID: uuid.New(), RecipientID: recipient.ID, CalendarID: body.CalendarID, StopID: body.StopID, Status: status, CalendarLinkTokenHash: calendarHash, StopLinkTokenHash: stopHash}
			if err := tx.Create(&subscription).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			if err := tx.Model(&subscription).Updates(map[string]interface{}{"status": status, "calendar_link_token_hash": calendarHash, "stop_link_token_hash": stopHash}).Error; err != nil {
				return err
			}
			subscription.Status = status
		}
		if body.Subscribed && body.StopID != nil {
			return tx.Where("recipient_id = ? AND stop_id = ?", recipient.ID, *body.StopID).Delete(&subscriptionExclusion{}).Error
		}
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not save update preferences"})
		return
	}
	if denied {
		c.JSON(404, gin.H{"error": "These plans are unavailable or require access"})
		return
	}
	c.JSON(200, subscription)
}

func (s *Service) preferenceToken(tx *gorm.DB, recipientID uuid.UUID) (string, error) {
	raw := newToken()
	now := time.Now().UTC()
	token := guestToken{ID: uuid.New(), RecipientID: recipientID, TokenHash: guestDigest(raw), Purpose: "preferences", VerifiedAt: now, ExpiresAt: now.Add(365 * 24 * time.Hour)}
	return raw, tx.Create(&token).Error
}

func (s *Service) preferencesResponse(c *gin.Context, recipientID uuid.UUID) {
	var recipient GuestRecipient
	var subscriptions []calendarSubscription
	if s.db.First(&recipient, "id = ?", recipientID).Error != nil || s.db.Where("recipient_id = ?", recipientID).Order("created_at").Find(&subscriptions).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load preferences"})
		return
	}
	if subscriptions == nil {
		subscriptions = []calendarSubscription{}
	}
	// No title, destination or travel dates: a preference capability is never a
	// viewing capability, even if the recipient is authorized for a stop.
	c.JSON(200, gin.H{"email": recipient.Email, "subscriptions": subscriptions, "participationNotices": recipient.ParticipationNotices})
}

func (s *Service) getGuestPreferences(c *gin.Context) {
	token, err := s.findGuestToken(c.Param("token"), "preferences")
	if err != nil {
		c.JSON(401, gin.H{"error": "This preference link has expired. Verify your email to manage updates."})
		return
	}
	s.preferencesResponse(c, token.RecipientID)
}

func (s *Service) updateGuestPreferences(c *gin.Context) {
	token, err := s.findGuestToken(c.Param("token"), "preferences")
	if err != nil {
		c.JSON(401, gin.H{"error": "This preference link has expired. Verify your email to manage updates."})
		return
	}
	var body struct {
		SubscriptionIDs      *[]uuid.UUID `json:"subscriptionIds"`
		UnsubscribeAll       bool         `json:"unsubscribeAll"`
		ParticipationNotices *bool        `json:"participationNotices"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "Invalid preferences"})
		return
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&GuestRecipient{}, "id = ?", token.RecipientID).Error; err != nil {
			return err
		}
		if body.UnsubscribeAll || body.SubscriptionIDs != nil {
			q := tx.Model(&calendarSubscription{}).Where("recipient_id = ?", token.RecipientID)
			if !body.UnsubscribeAll && len(*body.SubscriptionIDs) > 0 {
				q = q.Where("id NOT IN ?", *body.SubscriptionIDs)
			}
			if err := q.Update("status", "unsubscribed").Error; err != nil {
				return err
			}
		}
		// Preference links may remove updates; they cannot revive paused scopes.
		if body.ParticipationNotices != nil {
			return tx.Model(&GuestRecipient{}).Where("id = ?", token.RecipientID).Update("participation_notices", *body.ParticipationNotices).Error
		}
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not save preferences"})
		return
	}
	s.preferencesResponse(c, token.RecipientID)
}

func (s *Service) inviteCalendarGuest(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		guestScope
		Email       string `json:"email"`
		Name        string `json:"name"`
		GrantAccess bool   `json:"grantAccess"`
	}
	if c.ShouldBindJSON(&body) != nil || !body.guestScope.valid() {
		c.JSON(400, gin.H{"error": "Choose a calendar or stop and email address"})
		return
	}
	address, err := normalizedGuestEmail(body.Email)
	if err != nil || len(body.Name) > 200 {
		c.JSON(400, gin.H{"error": "Enter a valid email and a name shorter than 200 characters"})
		return
	}
	viewer := Viewer{UserID: user.ID}
	if !s.manageGuestScope(s.db, body.guestScope, viewer, body.GrantAccess) {
		c.JSON(404, gin.H{"error": "These plans are unavailable or require management access"})
		return
	}
	if !s.deliveryConfigured(c) {
		return
	}
	if !s.allowGuestAction("invite-user:"+user.ID.String(), 30) || !s.allowGuestAction("invite-email:"+address, 5) {
		c.JSON(429, gin.H{"error": "Too many invitations. Try again in 15 minutes."})
		return
	}
	var recipient *GuestRecipient
	var invitation calendarInvitation
	validationError := ""
	validationFailure := errors.New("invitation validation failed")
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		if err := s.lockGuestScope(tx, body.guestScope); err != nil {
			validationError = "These plans are unavailable"
			return validationFailure
		}
		recipient, err = s.ensureGuestRecipient(tx, address, body.Name)
		if err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(recipient, "id = ?", recipient.ID).Error; err != nil {
			return err
		}
		if !s.manageGuestScope(tx, body.guestScope, viewer, body.GrantAccess) {
			validationError = "Management access has changed"
			return validationFailure
		}
		if body.GrantAccess {
			visibility := ""
			if body.StopID != nil {
				var stop Stop
				if err := tx.First(&stop, "id = ?", *body.StopID).Error; err != nil {
					return err
				}
				visibility = stop.Visibility
			} else {
				var cal Calendar
				if err := tx.First(&cal, "id = ?", *body.CalendarID).Error; err != nil {
					return err
				}
				visibility = cal.Visibility
			}
			if visibility != "specific" {
				validationError = "Set this scope to Specific people before adding a viewer. Invitations do not change visibility."
				return validationFailure
			}
			var count int64
			if err := body.guestScope.query(tx.Model(&Grant{})).Where("recipient_id = ?", recipient.ID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				grant := Grant{ID: uuid.New(), CalendarID: body.CalendarID, StopID: body.StopID, RecipientID: &recipient.ID}
				if err := tx.Create(&grant).Error; err != nil {
					return err
				}
			}
		}
		if !s.guestCanViewScope(tx, body.guestScope, recipient, true) {
			validationError = "This recipient cannot view the published plans. Grant appropriate viewing access and publish before inviting."
			return validationFailure
		}
		err = body.guestScope.query(tx).Where("recipient_id = ?", recipient.ID).First(&invitation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			invitation = calendarInvitation{ID: uuid.New(), RecipientID: recipient.ID, CalendarID: body.CalendarID, StopID: body.StopID, SenderID: user.ID, Status: "pending"}
			return tx.Create(&invitation).Error
		}
		return err
	})
	if validationError != "" {
		c.JSON(400, gin.H{"error": validationError})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not prepare invitation"})
		return
	}
	// Repeat clicks and invitations by another co-host reuse the scope identity.
	if invitation.Status == "sent" || invitation.Status == "sending" {
		s.invitationResponse(c, &invitation, recipient)
		return
	}
	claimed := s.db.Model(&calendarInvitation{}).Where("id = ? AND status IN ?", invitation.ID, []string{"pending", "failed"}).Updates(map[string]interface{}{"status": "sending", "last_error": ""})
	if claimed.Error != nil || claimed.RowsAffected == 0 {
		c.JSON(409, gin.H{"error": "This invitation is already being sent. Refresh its delivery status."})
		return
	}
	// Recheck just before provider dispatch; generic invitation text never
	// includes restricted details even if access changes during delivery.
	if !s.manageGuestScope(s.db, body.guestScope, viewer, false) || !s.guestCanViewScope(s.db, body.guestScope, recipient, true) {
		s.db.Model(&invitation).Updates(map[string]interface{}{"status": "failed", "last_error": "Viewing or management access changed before delivery"})
		c.JSON(409, gin.H{"error": "Access changed before the invitation was sent"})
		return
	}
	link, err := s.guestScopeLink(s.db, body.guestScope)
	if err == nil {
		err = s.guestDelivery.Sender.SendMail(email.Email{Email: recipient.Email, Name: recipient.Name, Subject: "You are invited to view travel plans on Travel Angels", PlainTextContent: "You have been invited to browse travel plans on Travel Angels.\n\nOpen the plans: " + link + "\n\nYou can browse without installing the app or creating an account. Restricted plans require a fresh code sent to your own email. Following updates is optional; this invitation has not subscribed you. To express interest or request to join, verify this email and claim or sign in to your account.\n\nForwarding this message does not give someone else access to plans restricted to your email."})
	}
	if err != nil {
		if deliveryErrorUncertain(err) {
			invitation.Status = "sending"
			invitation.LastError = "Delivery confirmation is uncertain. This invitation will not be automatically resent."
			s.db.Model(&invitation).Updates(map[string]interface{}{"status": invitation.Status, "last_error": invitation.LastError})
			s.invitationResponse(c, &invitation, recipient)
			return
		}
		s.db.Model(&invitation).Updates(map[string]interface{}{"status": "failed", "last_error": "Email provider did not confirm delivery. Retry from the invitation screen."})
		invitation.Status = "failed"
		invitation.LastError = "Email delivery failed"
	} else {
		now := time.Now().UTC()
		s.db.Model(&invitation).Updates(map[string]interface{}{"status": "sent", "sent_at": now, "last_error": ""})
		invitation.Status = "sent"
		invitation.SentAt = &now
	}
	s.invitationResponse(c, &invitation, recipient)
}

func (s *Service) invitationResponse(c *gin.Context, invitation *calendarInvitation, recipient *GuestRecipient) {
	c.JSON(200, gin.H{"id": invitation.ID, "recipientId": recipient.ID, "calendarId": invitation.CalendarID, "stopId": invitation.StopID, "email": recipient.Email, "name": recipient.Name, "claimed": recipient.UserID != nil, "status": invitation.Status, "lastError": invitation.LastError})
}

func (s *Service) listCalendarInvitations(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	scope, valid := queryGuestScope(c)
	if !valid || !s.manageGuestScope(s.db, scope, Viewer{UserID: user.ID}, false) {
		c.JSON(404, gin.H{"error": "These plans are unavailable or require management access"})
		return
	}
	var invitations []calendarInvitation
	if scope.query(s.db).Order("created_at DESC").Find(&invitations).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load invitations"})
		return
	}
	result := []gin.H{}
	for _, invitation := range invitations {
		var recipient GuestRecipient
		if s.db.First(&recipient, "id = ?", invitation.RecipientID).Error != nil {
			continue
		}
		var subscription calendarSubscription
		scope.query(s.db).Where("recipient_id = ?", recipient.ID).First(&subscription)
		status := subscription.Status
		if status == "" {
			status = "not_following"
		}
		result = append(result, gin.H{"id": invitation.ID, "recipientId": recipient.ID, "email": recipient.Email, "name": recipient.Name, "claimed": recipient.UserID != nil, "status": invitation.Status, "subscriptionStatus": status, "lastError": invitation.LastError})
	}
	c.JSON(200, gin.H{"invitations": result})
}

func (s *Service) guestScopeLink(tx *gorm.DB, scope guestScope) (string, error) {
	base := s.baseURL
	if s.guestDelivery != nil && s.guestDelivery.BaseURL != "" {
		base = s.guestDelivery.BaseURL
	}
	base = strings.TrimRight(base, "/")
	if scope.StopID != nil {
		var stop Stop
		if err := tx.First(&stop, "id = ?", *scope.StopID).Error; err != nil {
			return "", err
		}
		if stop.SharingEnabled {
			return base + "/#/stops/" + stop.ShareToken, nil
		}
		var cal Calendar
		if tx.First(&cal, "id = ?", stop.CalendarID).Error == nil && cal.SharingEnabled {
			return base + "/#/calendar/" + cal.ShareToken, nil
		}
		return "", errors.New("Sharing links are disabled")
	}
	var cal Calendar
	if err := tx.First(&cal, "id = ?", *scope.CalendarID).Error; err != nil {
		return "", err
	}
	if !cal.SharingEnabled {
		return "", errors.New("Sharing links are disabled")
	}
	return base + "/#/calendar/" + cal.ShareToken, nil
}
