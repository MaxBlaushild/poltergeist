package travelcalendar

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type broadcastRequest struct {
	CalendarID *uuid.UUID  `json:"calendarId"`
	StopIDs    []uuid.UUID `json:"stopIds"`
	Message    string      `json:"message"`
}

func (s *Service) prepareBroadcast(tx *gorm.DB, request broadcastRequest, senderID uuid.UUID) ([]Stop, []GuestRecipient, error) {
	if strings.TrimSpace(request.Message) == "" || len(request.Message) > 5000 {
		return nil, nil, errors.New("Write a message of up to 5000 characters")
	}
	if len(request.StopIDs) > 100 {
		return nil, nil, errors.New("Choose at most 100 stops per message")
	}
	viewer := Viewer{UserID: senderID}
	if request.CalendarID != nil {
		if !s.manageGuestScope(tx, guestScope{CalendarID: request.CalendarID}, viewer, false) {
			return nil, nil, errors.New("Calendar management access is required")
		}
	}
	var stops []Stop
	if len(request.StopIDs) > 0 {
		if err := tx.Where("id IN ?", request.StopIDs).Find(&stops).Error; err != nil {
			return nil, nil, err
		}
		ids := map[uuid.UUID]bool{}
		for _, id := range request.StopIDs {
			ids[id] = true
		}
		if len(stops) != len(ids) {
			return nil, nil, errors.New("One or more selected stops are unavailable")
		}
	} else if request.CalendarID != nil {
		var calendar Calendar
		if err := tx.First(&calendar, "id = ?", *request.CalendarID).Error; err != nil {
			return nil, nil, err
		}
		if err := tx.Where("published IS NOT NULL AND (calendar_id = ? OR id IN (SELECT stop_id FROM tc_collaborators WHERE user_id = ? AND accepted = true AND stop_id IS NOT NULL))", calendar.ID, calendar.OwnerID).Find(&stops).Error; err != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, errors.New("Choose a calendar or published stops")
	}
	if len(stops) == 0 {
		return nil, nil, errors.New("Publish at least one stop before broadcasting")
	}
	if len(stops) > 100 {
		return nil, nil, errors.New("Select up to 100 stops for this broadcast")
	}
	for i := range stops {
		if stops[i].Published == nil || !s.canManageStop(tx, &stops[i], viewer) {
			return nil, nil, errors.New("You must manage every published stop included in this message")
		}
	}
	var recipients []GuestRecipient
	calendarIDs := []uuid.UUID{}
	stopIDs := []uuid.UUID{}
	seenCalendars := map[uuid.UUID]bool{}
	for _, stop := range stops {
		stopIDs = append(stopIDs, stop.ID)
		related, err := s.followableCalendarsForStop(tx, &stop)
		if err != nil {
			return nil, nil, err
		}
		for _, calendarID := range related {
			if !seenCalendars[calendarID] {
				seenCalendars[calendarID] = true
				calendarIDs = append(calendarIDs, calendarID)
			}
		}
	}
	// Recipients are considered only after explicitly following. Merely being
	// invited, granted access, verified, or claimed cannot opt someone in.
	if err := tx.Where("verified_at IS NOT NULL AND id IN (SELECT recipient_id FROM tc_calendar_subscriptions WHERE status = ? AND (calendar_id IN ? OR stop_id IN ?))", "active", calendarIDs, stopIDs).Find(&recipients).Error; err != nil {
		return nil, nil, err
	}
	eligible := []GuestRecipient{}
	for i := range recipients {
		if s.broadcastRecipientEligible(tx, &recipients[i], stops) {
			eligible = append(eligible, recipients[i])
		}
	}
	sort.Slice(stops, func(i, j int) bool { return stops[i].ID.String() < stops[j].ID.String() })
	return stops, eligible, nil
}

func (s *Service) followableCalendarsForStop(tx *gorm.DB, stop *Stop) ([]uuid.UUID, error) {
	var calendars []Calendar
	if err := tx.Where("id = ? OR owner_id IN (SELECT user_id FROM tc_collaborators WHERE stop_id = ? AND accepted = true)", stop.CalendarID, stop.ID).Find(&calendars).Error; err != nil {
		return nil, err
	}
	ids := []uuid.UUID{}
	for _, calendar := range calendars {
		ids = append(ids, calendar.ID)
	}
	return ids, nil
}

func (s *Service) hasEligibleStopSubscription(tx *gorm.DB, recipient *GuestRecipient, stop *Stop) bool {
	calendarIDs, err := s.followableCalendarsForStop(tx, stop)
	if err != nil {
		return false
	}
	var subscriptions []calendarSubscription
	if tx.Where("recipient_id = ? AND status = ? AND (calendar_id IN ? OR stop_id = ?)", recipient.ID, "active", calendarIDs, stop.ID).Find(&subscriptions).Error != nil {
		return false
	}
	for _, subscription := range subscriptions {
		viewer := recipientViewer(recipient)
		if subscription.StopID != nil && subscription.StopLinkTokenHash != "" && subscription.StopLinkTokenHash == guestDigest(stop.ShareToken) && stop.SharingEnabled {
			viewer.StopLinkID = stop.ID
		}
		calendarID := stop.CalendarID
		if subscription.CalendarID != nil {
			calendarID = *subscription.CalendarID
		}
		var calendar Calendar
		if subscription.CalendarLinkTokenHash != "" && tx.First(&calendar, "id = ?", calendarID).Error == nil && calendar.SharingEnabled && subscription.CalendarLinkTokenHash == guestDigest(calendar.ShareToken) {
			viewer.CalendarLinkID = calendar.ID
		}
		if s.canViewStop(tx, stop, viewer) {
			return true
		}
	}
	return false
}

func (s *Service) broadcastRecipientEligible(tx *gorm.DB, recipient *GuestRecipient, stops []Stop) bool {
	if len(stops) == 0 || recipient.VerifiedAt == nil {
		return false
	}
	for i := range stops {
		stop := &stops[i]
		if stop.Published == nil {
			return false
		}
		var exclusion int64
		if tx.Model(&subscriptionExclusion{}).Where("recipient_id = ? AND stop_id = ?", recipient.ID, stop.ID).Count(&exclusion).Error != nil || exclusion > 0 {
			return false
		}
		if !s.hasEligibleStopSubscription(tx, recipient, stop) {
			return false
		}
	}
	// Shared free-form text can mention any included stop; therefore every
	// included stop must be both authorized and followed by the recipient.
	return true
}

func (s *Service) previewCalendarBroadcast(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body broadcastRequest
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "Invalid broadcast"})
		return
	}
	stops, recipients, err := s.prepareBroadcast(s.db, body, user.ID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ids := []uuid.UUID{}
	for _, stop := range stops {
		ids = append(ids, stop.ID)
	}
	audience := []gin.H{}
	for _, recipient := range recipients {
		audience = append(audience, gin.H{"email": recipient.Email, "name": recipient.Name})
	}
	c.JSON(200, gin.H{"recipientCount": len(recipients), "recipients": audience, "stopIds": ids, "message": strings.TrimSpace(body.Message), "audienceExplanation": "Recipients must follow and currently be allowed to view every included stop. Send separate messages for different audiences."})
}

func (s *Service) createCalendarBroadcast(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body broadcastRequest
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "Invalid broadcast"})
		return
	}
	if !s.deliveryConfigured(c) {
		return
	}
	if !s.allowGuestAction("broadcast-user:"+user.ID.String(), 10) {
		c.JSON(429, gin.H{"error": "Too many broadcasts. Try again in 15 minutes."})
		return
	}
	var broadcast calendarBroadcast
	err := s.db.Transaction(func(tx *gorm.DB) error {
		stops, recipients, err := s.prepareBroadcast(tx, body, user.ID)
		if err != nil {
			return err
		}
		if len(recipients) == 0 {
			return errors.New("No followers can currently receive every stop in this message. Check permissions and update preferences or send separate messages.")
		}
		broadcast = calendarBroadcast{ID: uuid.New(), SenderID: user.ID, CalendarID: body.CalendarID, StopIDs: []uuid.UUID{}, Message: strings.TrimSpace(body.Message)}
		for _, stop := range stops {
			broadcast.StopIDs = append(broadcast.StopIDs, stop.ID)
		}
		if err := tx.Create(&broadcast).Error; err != nil {
			return err
		}
		for _, recipient := range recipients {
			delivery := calendarDelivery{ID: uuid.New(), BroadcastID: &broadcast.ID, RecipientID: recipient.ID, Event: "broadcast", DedupeKey: fmt.Sprintf("broadcast:%s:%s", broadcast.ID, recipient.ID), Status: "pending"}
			if err := tx.Create(&delivery).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	s.broadcastResponse(c, &broadcast)
}

func broadcastStatus(deliveries []calendarDelivery) string {
	pending, failed, sent := false, false, false
	for _, delivery := range deliveries {
		switch delivery.Status {
		case "pending", "sending":
			pending = true
		case "failed":
			failed = true
		case "sent":
			sent = true
		}
	}
	if pending {
		return "pending"
	}
	if failed {
		return "failed"
	}
	if sent {
		return "sent"
	}
	return "skipped"
}

func (s *Service) broadcastResponse(c *gin.Context, broadcast *calendarBroadcast) {
	deliveries := []calendarDelivery{}
	if s.db.Where("broadcast_id = ?", broadcast.ID).Order("created_at").Find(&deliveries).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load delivery status"})
		return
	}
	c.JSON(200, gin.H{"id": broadcast.ID, "calendarId": broadcast.CalendarID, "stopIds": broadcast.StopIDs, "message": broadcast.Message, "createdAt": broadcast.CreatedAt, "status": broadcastStatus(deliveries), "deliveries": deliveries})
}

func (s *Service) canManageBroadcast(tx *gorm.DB, broadcast *calendarBroadcast, userID uuid.UUID) bool {
	if len(broadcast.StopIDs) == 0 {
		return false
	}
	var stops []Stop
	if tx.Where("id IN ?", broadcast.StopIDs).Find(&stops).Error != nil || len(stops) != len(broadcast.StopIDs) {
		return false
	}
	for i := range stops {
		if !s.canManageStop(tx, &stops[i], Viewer{UserID: userID}) {
			return false
		}
	}
	return true
}

func (s *Service) listCalendarBroadcasts(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	scope, valid := queryGuestScope(c)
	if !valid || !s.manageGuestScope(s.db, scope, Viewer{UserID: user.ID}, false) {
		c.JSON(404, gin.H{"error": "These plans are unavailable or require management access"})
		return
	}
	// Scope managers can see broadcasts covering their scope only if they
	// still manage all of the message's included stops and audience.
	var broadcasts []calendarBroadcast
	q := s.db.Order("created_at DESC").Limit(100)
	if scope.CalendarID != nil {
		q = q.Where("calendar_id = ?", *scope.CalendarID)
	} else {
		q = q.Where("stop_ids @> ?::jsonb", fmt.Sprintf("[\"%s\"]", *scope.StopID))
	}
	if q.Find(&broadcasts).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load broadcasts"})
		return
	}
	result := []gin.H{}
	for _, broadcast := range broadcasts {
		if !s.canManageBroadcast(s.db, &broadcast, user.ID) {
			continue
		}
		deliveries := []calendarDelivery{}
		if s.db.Where("broadcast_id = ?", broadcast.ID).Find(&deliveries).Error != nil {
			continue
		}
		result = append(result, gin.H{"id": broadcast.ID, "calendarId": broadcast.CalendarID, "stopIds": broadcast.StopIDs, "message": broadcast.Message, "createdAt": broadcast.CreatedAt, "status": broadcastStatus(deliveries), "deliveries": deliveries})
	}
	c.JSON(200, gin.H{"broadcasts": result})
}

func (s *Service) retryCalendarBroadcast(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": "Broadcast unavailable"})
		return
	}
	var broadcast calendarBroadcast
	if s.db.First(&broadcast, "id = ?", id).Error != nil || !s.canManageBroadcast(s.db, &broadcast, user.ID) {
		c.JSON(404, gin.H{"error": "Broadcast unavailable"})
		return
	}
	if !s.deliveryConfigured(c) {
		return
	}
	// A provider-successful send, including an uncertain in-flight send after
	// process death, is never blindly sent again by the retry action.
	if s.db.Model(&calendarDelivery{}).Where("broadcast_id = ? AND status = ?", id, "failed").Updates(map[string]interface{}{"status": "pending", "last_error": ""}).Error != nil {
		c.JSON(500, gin.H{"error": "Could not retry deliveries"})
		return
	}
	s.broadcastResponse(c, &broadcast)
}
