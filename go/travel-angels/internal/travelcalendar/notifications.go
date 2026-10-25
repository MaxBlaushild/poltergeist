package travelcalendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type calendarNotice struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid" json:"-"`
	StopID          uuid.UUID  `gorm:"type:uuid" json:"-"`
	ParticipationID uuid.UUID  `gorm:"type:uuid" json:"-"`
	Event           string     `json:"event"`
	DedupeKey       string     `json:"-"`
	ReadAt          *time.Time `json:"readAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

func (calendarNotice) TableName() string { return "tc_calendar_notices" }

func (s *Service) queueParticipationNotice(tx *gorm.DB, stop *Stop, participation *Participation, event string) error {
	targets := map[uuid.UUID]string{participation.UserID: event}
	if event == "requested" || event == "withdrawn" || event == "interested" || event == "waitlisted" {
		var cal Calendar
		if err := tx.First(&cal, "id = ?", stop.CalendarID).Error; err != nil {
			return err
		}
		targets[cal.OwnerID] = "host_" + event
		var managers []Collaborator
		if err := tx.Where("accepted = ? AND (calendar_id = ? OR stop_id = ?)", true, stop.CalendarID, stop.ID).Find(&managers).Error; err != nil {
			return err
		}
		for _, manager := range managers {
			if s.canManageStop(tx, stop, Viewer{UserID: manager.UserID}) {
				targets[manager.UserID] = "host_" + event
			}
		}
	}
	for userID, targetEvent := range targets {
		dedupe := fmt.Sprintf("participation:%s:%s:%s:%d", participation.ID, userID, targetEvent, participation.UpdatedAt.UnixNano())
		if targetEvent == "logistics_updated" {
			// The private deduplication key also binds this generic notification to
			// exactly the logistics this recipient was allowed to see at publication.
			// A later change or audience restriction makes it stale, without storing
			// private content in the outbox or exposing which hidden entry changed.
			snapshot, err := tripLogisticsNoticeSnapshot(stop.Published, userID)
			if err != nil {
				return err
			}
			dedupe += ":logistics:" + snapshot
		}
		notice := calendarNotice{ID: uuid.New(), UserID: userID, StopID: stop.ID, ParticipationID: participation.ID, Event: targetEvent, DedupeKey: dedupe}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "dedupe_key"}}, DoNothing: true}).Create(&notice).Error; err != nil {
			return err
		}
		var recipients []GuestRecipient
		if err := tx.Where("user_id = ? AND verified_at IS NOT NULL AND participation_notices = ?", userID, true).Find(&recipients).Error; err != nil {
			return err
		}
		for _, recipient := range recipients {
			delivery := calendarDelivery{ID: uuid.New(), RecipientID: recipient.ID, StopID: &stop.ID, ParticipationID: &participation.ID, Event: targetEvent, DedupeKey: dedupe + ":" + recipient.ID.String(), Status: "pending"}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "dedupe_key"}}, DoNothing: true}).Create(&delivery).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) reconcileGuestAccess(tx *gorm.DB, stop *Stop) error {
	var subscriptions []calendarSubscription
	calendarIDs, err := s.followableCalendarsForStop(tx, stop)
	if err != nil {
		return err
	}
	if err := tx.Where("status = ? AND (calendar_id IN ? OR stop_id = ?)", "active", calendarIDs, stop.ID).Find(&subscriptions).Error; err != nil {
		return err
	}
	seen := map[uuid.UUID]bool{}
	for _, subscription := range subscriptions {
		if seen[subscription.RecipientID] {
			continue
		}
		seen[subscription.RecipientID] = true
		var recipient GuestRecipient
		if err := tx.First(&recipient, "id = ?", subscription.RecipientID).Error; err != nil {
			return err
		}
		if s.hasEligibleStopSubscription(tx, &recipient, stop) {
			continue
		}
		if err := tx.Model(&calendarSubscription{}).Where("recipient_id = ? AND stop_id = ? AND status = ?", recipient.ID, stop.ID, "active").Update("status", "paused").Error; err != nil {
			return err
		}
		exclusion := subscriptionExclusion{RecipientID: recipient.ID, StopID: stop.ID}
		insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&exclusion)
		if insert.Error != nil {
			return insert.Error
		}
		if insert.RowsAffected == 0 {
			continue
		}
		// A removed participant already receives the participation notice. Other
		// followers receive a generic access notice with no restricted details.
		var removed int64
		if recipient.UserID != nil {
			if err := tx.Model(&Participation{}).Where("stop_id = ? AND user_id = ? AND status = ?", stop.ID, *recipient.UserID, "removed").Count(&removed).Error; err != nil {
				return err
			}
		}
		if removed == 0 && recipient.VerifiedAt != nil {
			delivery := calendarDelivery{ID: uuid.New(), RecipientID: recipient.ID, StopID: &stop.ID, Event: "access_removed", DedupeKey: fmt.Sprintf("access:%s:%s:%d", stop.ID, recipient.ID, time.Now().UTC().UnixNano()), Status: "pending"}
			if err := tx.Create(&delivery).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// RunDeliveryWorker drains the durable outbox. Multiple server replicas safely
// compete using a conditional pending-to-sending transition. An interrupted
// sending row stays uncertain instead of being automatically duplicated.
func (s *Service) RunDeliveryWorker(ctx context.Context) {
	if s.guestDelivery == nil || s.guestDelivery.Sender == nil {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		_ = s.DispatchPendingDeliveries(ctx, 25)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) DispatchPendingDeliveries(ctx context.Context, limit int) error {
	if s.guestDelivery == nil || s.guestDelivery.Sender == nil {
		return errors.New("email delivery is not configured")
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	var deliveries []calendarDelivery
	if err := s.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at").Limit(limit).Find(&deliveries).Error; err != nil {
		return err
	}
	var firstError error
	for i := range deliveries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := s.dispatchCalendarDelivery(ctx, deliveries[i].ID); err != nil && firstError == nil {
			firstError = err
		}
	}
	return firstError
}

func (s *Service) dispatchCalendarDelivery(ctx context.Context, id uuid.UUID) error {
	claim := s.db.WithContext(ctx).Model(&calendarDelivery{}).Where("id = ? AND status = ?", id, "pending").Updates(map[string]interface{}{"status": "sending", "attempts": gorm.Expr("attempts + 1")})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected == 0 {
		return nil
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var delivery calendarDelivery
		if err := tx.First(&delivery, "id = ?", id).Error; err != nil {
			return err
		}
		var broadcast calendarBroadcast
		var stops []Stop
		if delivery.BroadcastID != nil {
			if err := tx.First(&broadcast, "id = ?", *delivery.BroadcastID).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", broadcast.StopIDs).Find(&stops).Error; err != nil {
				return err
			}
			if len(stops) != len(broadcast.StopIDs) {
				return s.skipCalendarDelivery(tx, &delivery, "One or more plans are unavailable")
			}
		} else if delivery.StopID != nil {
			var stop Stop
			if err := tx.First(&stop, "id = ?", *delivery.StopID).Error; err != nil {
				return err
			}
			stops = append(stops, stop)
		} else {
			return s.skipCalendarDelivery(tx, &delivery, "No eligible plans")
		}
		// Use the core permission mutation lock order: calendars before stops.
		// Holding those locks through dispatch closes the check/send revocation
		// race; the recipient lock likewise serializes an unsubscribe.
		calendars := map[uuid.UUID]bool{}
		for _, stop := range stops {
			calendars[stop.CalendarID] = true
		}
		calendarIDs := []uuid.UUID{}
		for calendarID := range calendars {
			calendarIDs = append(calendarIDs, calendarID)
		}
		sort.Slice(calendarIDs, func(i, j int) bool { return calendarIDs[i].String() < calendarIDs[j].String() })
		for _, calendarID := range calendarIDs {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&Calendar{}, "id = ?", calendarID).Error; err != nil {
				return err
			}
		}
		sort.Slice(stops, func(i, j int) bool { return stops[i].ID.String() < stops[j].ID.String() })
		for i := range stops {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stops[i], "id = ?", stops[i].ID).Error; err != nil {
				return err
			}
			if !calendars[stops[i].CalendarID] {
				// An ownership transfer completed after the initial lookup. No email
				// has been attempted, so retry with the new owning calendar's lock.
				return tx.Model(&delivery).Updates(map[string]interface{}{"status": "pending", "last_error": ""}).Error
			}
		}
		// Permission changes may suppress a queued notification while this worker
		// waits for the trip locks. Never send a stale pre-lock copy of that row.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&delivery, "id = ?", id).Error; err != nil {
			return err
		}
		if delivery.Status != "sending" {
			return nil
		}
		var recipient GuestRecipient
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&recipient, "id = ?", delivery.RecipientID).Error; err != nil {
			return err
		}
		var subject, content string
		if delivery.BroadcastID != nil {
			if !s.canManageBroadcast(tx, &broadcast, broadcast.SenderID) || !s.broadcastRecipientEligible(tx, &recipient, stops) {
				return s.skipCalendarDelivery(tx, &delivery, "Viewing, management access, or following preference changed")
			}
			subject = "Travel Angels: an update to plans you follow"
			content = broadcast.Message + "\n\n"
			for _, stop := range stops {
				link, err := s.guestScopeLink(tx, guestScope{StopID: &stop.ID})
				if err != nil {
					return s.skipCalendarDelivery(tx, &delivery, "Sharing links are disabled")
				}
				content += stop.Published.Title + " — " + stop.Published.Destination + "\n" + stopDateLabel(stop.Published) + "\n" + link + "\n\n"
			}
		} else {
			var allowed bool
			subject, content, allowed = s.participationEmail(tx, &delivery, &recipient, &stops[0])
			if !allowed {
				return s.skipCalendarDelivery(tx, &delivery, "Participation or notification preferences changed")
			}
		}
		preferences, err := s.preferenceToken(tx, recipient.ID)
		if err != nil {
			return err
		}
		content += "\nManage email preferences or unsubscribe: " + strings.TrimRight(s.guestDelivery.BaseURL, "/") + "/#/preferences/" + preferences + "\n\nFollowing updates does not join a plan. Participation actions require your claimed account."
		err = s.guestDelivery.Sender.SendMail(email.Email{Email: recipient.Email, Name: recipient.Name, Subject: subject, PlainTextContent: content})
		if err != nil {
			if deliveryErrorUncertain(err) {
				return tx.Model(&delivery).Updates(map[string]interface{}{"status": "sending", "last_error": "Delivery confirmation is uncertain. This message will not be automatically resent."}).Error
			}
			return tx.Model(&delivery).Updates(map[string]interface{}{"status": "failed", "last_error": "The email provider did not confirm delivery. Retry is available."}).Error
		}
		now := time.Now().UTC()
		return tx.Model(&delivery).Updates(map[string]interface{}{"status": "sent", "sent_at": now, "last_error": ""}).Error
	})
	// A database/provider interruption can happen after provider acceptance.
	// Keep 'sending' as uncertain: retrying it could duplicate a successful send.
	if err != nil {
		s.db.Model(&calendarDelivery{}).Where("id = ? AND status = ?", id, "sending").Update("last_error", "Delivery confirmation is uncertain. This message will not be automatically resent.")
	}
	return err
}

func deliveryErrorUncertain(err error) bool {
	var networkError net.Error
	return errors.As(err, &networkError) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func (s *Service) skipCalendarDelivery(tx *gorm.DB, delivery *calendarDelivery, reason string) error {
	return tx.Model(delivery).Updates(map[string]interface{}{"status": "skipped", "last_error": reason}).Error
}

func stopDateLabel(content *StopContent) string {
	if content.DatePrecision == "tbd" {
		return "Dates to be decided"
	}
	label := content.StartDate + " to " + content.EndDate
	if content.DatePrecision == "approximate" {
		label += " (approximate)"
	}
	return label
}

func noticeMessage(event, title string) string {
	if event == "removed" || event == "access_removed" {
		return "Your access to a travel plan has ended. Any active participation in it has been removed. Updates for that plan are paused."
	}
	if event == "unpublished" {
		return "A travel plan is no longer published. Review and reconfirm your participation when the host publishes it again."
	}
	if title == "" {
		return "A travel plan's participation has changed. Open Travel Angels to review the plans currently available to you."
	}
	switch event {
	case "host_waitlisted":
		return "A traveler joined the waitlist for " + title + ". Review the waitlist in Travel Angels."
	case "waitlisted":
		return "You are on the waitlist for " + title + ". A host must approve your place before you are confirmed."
	case "reservation_released":
		return "Your reserved place for " + title + " was released. Your dates remain saved; any renewed approval depends on availability."
	case "logistics_updated":
		return "Lodging or flights for " + title + " were updated. Open the trip to review the details currently shared with you."
	case "host_requested":
		return "A traveler requested to join " + title + ". Review their request in Travel Angels."
	case "host_interested":
		return "A traveler is interested in " + title + "."
	case "host_withdrawn":
		return "A traveler withdrew from " + title + "."
	case "requested":
		return "Your request to join " + title + " is awaiting host approval."
	case "interested":
		return "Your interest in " + title + " has been recorded."
	case "joined":
		return "Your request to join " + title + " was approved."
	case "declined":
		return "Your request to join " + title + " was declined."
	case "withdrawn":
		return "Your participation in " + title + " was withdrawn. Participation updates have ended."
	case "cancelled":
		return title + " was cancelled. Your participation has been cancelled."
	case "needs_reconfirmation":
		return title + " changed. Your previous dates are saved; review the new plan and resubmit for host approval. You are not counted as confirmed until approval."
	default:
		return "Your participation in " + title + " changed. Review it in Travel Angels."
	}
}

func (s *Service) participationEmail(tx *gorm.DB, delivery *calendarDelivery, recipient *GuestRecipient, stop *Stop) (string, string, bool) {
	if recipient.VerifiedAt == nil {
		return "", "", false
	}
	if delivery.Event != "access_removed" && !recipient.ParticipationNotices {
		return "", "", false
	}
	viewer := recipientViewer(recipient)
	if strings.HasPrefix(delivery.Event, "host_") && !s.canManageStop(tx, stop, viewer) {
		return "", "", false
	}
	if delivery.ParticipationID != nil {
		var participation Participation
		if tx.First(&participation, "id = ?", *delivery.ParticipationID).Error != nil {
			return "", "", false
		}
		if participation.Status == "withdrawn" && delivery.Event != "withdrawn" && delivery.Event != "host_withdrawn" {
			return "", "", false
		}
		if participation.Status == "removed" && delivery.Event != "removed" {
			return "", "", false
		}
		if recipient.UserID == nil {
			return "", "", false
		}
		if *recipient.UserID == participation.UserID {
			viewer = s.storedParticipationViewer(tx, stop, &participation)
			viewer.RecipientID = recipient.ID
		}
	}
	if delivery.Event == "logistics_updated" && (!s.tripAttendee(tx, stop, viewer) || !tripLogisticsNoticeCurrent(stop, viewer.UserID, delivery.DedupeKey)) {
		return "", "", false
	}
	canView := stop.Published != nil && s.canViewStop(tx, stop, viewer)
	title := ""
	if canView {
		title = stop.Published.Title
	}
	content := noticeMessage(delivery.Event, title) + "\n\n"
	if canView && delivery.Event != "removed" && delivery.Event != "access_removed" && delivery.Event != "unpublished" {
		if link, err := s.guestScopeLink(tx, guestScope{StopID: &stop.ID}); err == nil {
			content += "Review the current plan: " + link + "\n"
		}
	}
	return "Travel Angels: participation update", content, true
}

func (s *Service) listCalendarNotifications(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var notices []calendarNotice
	if s.db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(100).Find(&notices).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load notifications"})
		return
	}
	result := []gin.H{}
	for _, notice := range notices {
		title := ""
		var stop Stop
		available := false
		if s.db.First(&stop, "id = ?", notice.StopID).Error == nil {
			viewer := Viewer{UserID: user.ID}
			var participation Participation
			if s.db.First(&participation, "id = ? AND user_id = ?", notice.ParticipationID, user.ID).Error == nil {
				viewer = s.storedParticipationViewer(s.db, &stop, &participation)
			}
			if !s.tripNoticeAuthority(s.db, &stop, viewer, notice.Event, notice.DedupeKey) {
				continue
			}
			available = stop.Published != nil && s.canViewStop(s.db, &stop, viewer)
			if available {
				title = stop.Published.Title
			}
		} else if notice.Event == "logistics_updated" || strings.HasPrefix(notice.Event, "host_") {
			continue
		}
		item := gin.H{"id": notice.ID, "event": notice.Event, "message": noticeMessage(notice.Event, title), "createdAt": notice.CreatedAt, "readAt": notice.ReadAt}
		if available {
			item["stopId"] = stop.ID
		}
		result = append(result, item)
	}
	var recipients []GuestRecipient
	if s.db.Where("user_id = ?", user.ID).Find(&recipients).Error != nil {
		c.JSON(500, gin.H{"error": "Could not load notifications"})
		return
	}
	ids := []uuid.UUID{}
	for _, recipient := range recipients {
		ids = append(ids, recipient.ID)
	}
	deliveries := []gin.H{}
	if len(ids) > 0 {
		var rows []calendarDelivery
		if s.db.Where("recipient_id IN ? AND broadcast_id IS NULL", ids).Order("created_at DESC").Limit(100).Find(&rows).Error != nil {
			c.JSON(500, gin.H{"error": "Could not load deliveries"})
			return
		}
		for _, row := range rows {
			if row.Event == "logistics_updated" || strings.HasPrefix(row.Event, "host_") {
				if row.StopID == nil {
					continue
				}
				var stop Stop
				if s.db.First(&stop, "id = ?", *row.StopID).Error != nil {
					continue
				}
				viewer := s.participationViewer(s.db, &stop, user.ID, participationInput{})
				if !s.tripNoticeAuthority(s.db, &stop, viewer, row.Event, row.DedupeKey) {
					continue
				}
			}
			deliveries = append(deliveries, gin.H{"id": row.ID, "status": row.Status, "event": row.Event, "lastError": row.LastError, "createdAt": row.CreatedAt})
		}
	}
	c.JSON(200, gin.H{"notifications": result, "deliveries": deliveries})
}

func tripLogisticsNoticeSnapshot(content *StopContent, userID uuid.UUID) (string, error) {
	if content == nil {
		return "", nil
	}
	encoded, err := json.Marshal(filteredTripLogistics(content, userID, false, true))
	if err != nil {
		return "", err
	}
	return guestDigest(string(encoded)), nil
}

func tripLogisticsNoticeCurrent(stop *Stop, userID uuid.UUID, dedupe string) bool {
	if stop.Published == nil || userID == uuid.Nil {
		return false
	}
	_, suffix, found := strings.Cut(dedupe, ":logistics:")
	if !found {
		return false
	}
	expected, _, _ := strings.Cut(suffix, ":")
	current, err := tripLogisticsNoticeSnapshot(stop.Published, userID)
	return err == nil && current != "" && expected == current
}

func (s *Service) tripNoticeAuthority(tx *gorm.DB, stop *Stop, viewer Viewer, event, dedupe string) bool {
	if strings.HasPrefix(event, "host_") {
		return s.canManageStop(tx, stop, viewer)
	}
	if event == "logistics_updated" {
		return s.tripAttendee(tx, stop, viewer) && tripLogisticsNoticeCurrent(stop, viewer.UserID, dedupe)
	}
	return true
}

// Call under the trip/calendar permission locks when narrowing logistics access.
// Uncertain sends remain uncertain; they are never automatically retried.
func (s *Service) suppressTripLogisticsNotices(tx *gorm.DB, stopID uuid.UUID) error {
	if err := tx.Model(&calendarDelivery{}).Where("stop_id = ? AND event = ? AND status IN ?", stopID, "logistics_updated", []string{"pending", "failed"}).Updates(map[string]interface{}{"status": "skipped", "last_error": "Logistics access changed"}).Error; err != nil {
		return err
	}
	return tx.Where("stop_id = ? AND event = ?", stopID, "logistics_updated").Delete(&calendarNotice{}).Error
}

func (s *Service) retryCalendarNotifications(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	if !s.deliveryConfigured(c) {
		return
	}
	result := s.db.Model(&calendarDelivery{}).Where("broadcast_id IS NULL AND status = ? AND recipient_id IN (SELECT id FROM tc_guest_recipients WHERE user_id = ?)", "failed", user.ID).Updates(map[string]interface{}{"status": "pending", "last_error": ""})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Could not retry notifications"})
		return
	}
	c.JSON(200, gin.H{"pending": result.RowsAffected})
}

// Logistics notices compare only the recipient-visible material. A change to
// host notes or another traveler's assignments must not reveal its existence.
func (s *Service) queueTripLogisticsNotices(tx *gorm.DB, stop *Stop, before, after *StopContent) error {
	var ps []Participation
	if e := tx.Where("stop_id = ? AND (status = 'joined' OR (reserved = true AND status IN ('requested','needs_reconfirmation')))", stop.ID).Find(&ps).Error; e != nil {
		return e
	}
	for i := range ps {
		p := &ps[i]
		v := s.storedParticipationViewer(tx, stop, p)
		if !s.tripAttendee(tx, stop, v) {
			continue
		}
		oldView := filteredTripLogistics(before, p.UserID, false, true)
		newView := filteredTripLogistics(after, p.UserID, false, true)
		oldJSON, e := json.Marshal(oldView)
		if e != nil {
			return e
		}
		newJSON, e := json.Marshal(newView)
		if e != nil {
			return e
		}
		if string(oldJSON) == string(newJSON) {
			continue
		}
		// Include the publication revision in the event key without changing the
		// actual attendance timestamp or allocating a second attendance record.
		noticeP := *p
		noticeP.UpdatedAt = stop.UpdatedAt
		if e = s.queueParticipationNotice(tx, stop, &noticeP, "logistics_updated"); e != nil {
			return e
		}
	}
	return nil
}
