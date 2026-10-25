package travelcalendar

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func tripNotificationFixture(t *testing.T) (*guestTestFixture, Stop, Participation) {
	t.Helper()
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	friend := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", friend).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	recipient := GuestRecipient{ID: uuid.New(), Email: "trip-notices@example.com", UserID: &friend, VerifiedAt: &now, ParticipationNotices: true}
	if err := f.db.Create(&recipient).Error; err != nil {
		t.Fatal(err)
	}
	p := Participation{ID: uuid.New(), StopID: stop.ID, UserID: friend, Status: "joined", StopLinkTokenHash: guestDigest(stop.ShareToken)}
	if err := f.db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	content := *stop.Published
	content.Lodging = []TripLodging{{ID: uuid.NewString(), Name: "Group house", DatesTbd: true, Visibility: "attendees", PrivateDetails: &TripLodgingPrivateDetails{Visibility: "attendees", Address: "PRIVATE ADDRESS"}}}
	stop.Published = &content
	stop.Draft = content
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.service.queueParticipationNotice(f.db, &stop, &p, "logistics_updated"); err != nil {
		t.Fatal(err)
	}
	return f, stop, p
}
func tripNotificationRows(t *testing.T, f *guestTestFixture, user uuid.UUID) ([]interface{}, []interface{}) {
	t.Helper()
	response := calendarResponse(t, f, "GET", "/calendar-notifications", nil, user, "", 200)
	return response["notifications"].([]interface{}), response["deliveries"].([]interface{})
}
func TestTripNotificationsSuppressLogisticsAfterAudienceNarrows(t *testing.T) {
	for _, scope := range []string{"entry", "private details"} {
		t.Run(scope, func(t *testing.T) {
			f, stop, p := tripNotificationFixture(t)
			notices, deliveries := tripNotificationRows(t, f, p.UserID)
			if len(notices) != 1 || len(deliveries) != 1 {
				t.Fatal("eligible attendee did not receive queued update")
			}
			if scope == "entry" {
				stop.Published.Lodging[0].Visibility = "hosts"
			} else {
				stop.Published.Lodging[0].PrivateDetails.Visibility = "hosts"
			}
			// Bypass the immediate route cleanup to exercise dispatch/list defense in
			// depth: scope validation must independently reject this stale outbox item.
			if err := f.db.Save(&stop).Error; err != nil {
				t.Fatal(err)
			}
			if !f.service.tripAttendee(f.db, &stop, f.service.storedParticipationViewer(f.db, &stop, &p)) {
				t.Fatal("fixture must retain general attendee access")
			}
			notices, deliveries = tripNotificationRows(t, f, p.UserID)
			if len(notices) != 0 || len(deliveries) != 0 {
				t.Fatal("notification metadata revealed newly hidden logistics")
			}
			if err := f.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
				t.Fatal(err)
			}
			if f.sender.Attempts != 0 {
				t.Fatal("a queued email disclosed a now-hidden update")
			}
			var delivery calendarDelivery
			if err := f.db.First(&delivery).Error; err != nil {
				t.Fatal(err)
			}
			if delivery.Status != "skipped" {
				t.Fatalf("expected skipped delivery, got %s", delivery.Status)
			}
		})
	}
}
func TestTripNotificationsIgnoreHostOnlyChangesAndCoalesceNewPublications(t *testing.T) {
	f, stop, p := tripNotificationFixture(t)
	stop.Published.Lodging[0].HostNotes = "PRIVATE HOST NOTES"
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	notices, _ := tripNotificationRows(t, f, p.UserID)
	if len(notices) != 1 {
		t.Fatal("host-only edits invalidated a still-valid attendee update")
	}
	prior := *stop.Published
	next := prior
	next.Lodging = append([]TripLodging{}, prior.Lodging...)
	next.Lodging[0].Name = "Updated group house"
	stop.Published = &next
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.service.queueTripLogisticsNotices(f.db, &stop, &prior, &next); err != nil {
		t.Fatal(err)
	}
	notices, deliveries := tripNotificationRows(t, f, p.UserID)
	if len(notices) != 1 || len(deliveries) != 1 {
		t.Fatal("stale publication notices were not coalesced")
	}
	if err := f.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if f.sender.Attempts != 1 || len(f.sender.Messages) != 1 {
		t.Fatal("latest eligible update must send exactly once")
	}
}
func TestTripNotificationsHidePastHostActivityAndWithdrawnLogistics(t *testing.T) {
	f, stop, p := tripNotificationFixture(t)
	sid := stop.ID
	role := Collaborator{ID: uuid.New(), StopID: &sid, UserID: p.UserID, Accepted: true, IsHost: true, CanManagePermissions: true}
	if err := f.db.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	applicant := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", applicant).Error; err != nil {
		t.Fatal(err)
	}
	requested := Participation{ID: uuid.New(), StopID: stop.ID, UserID: applicant, Status: "requested"}
	if err := f.db.Create(&requested).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.service.queueParticipationNotice(f.db, &stop, &requested, "requested"); err != nil {
		t.Fatal(err)
	}
	notices, deliveries := tripNotificationRows(t, f, p.UserID)
	if len(notices) != 2 || len(deliveries) != 2 {
		t.Fatal("host should initially receive own logistics and request activity")
	}
	if err := f.db.Delete(&role).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&p).Update("status", "withdrawn").Error; err != nil {
		t.Fatal(err)
	}
	notices, deliveries = tripNotificationRows(t, f, p.UserID)
	if len(notices) != 0 || len(deliveries) != 0 {
		t.Fatal("removed role or attendee access retained private notification metadata")
	}
	if err := f.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if f.sender.Attempts != 0 {
		t.Fatal("email retained revoked host or attendee authority")
	}
}
func TestTripDeliveryDefersWhenOwnershipChangesBeforeCalendarLocks(t *testing.T) {
	f, stop, _ := tripNotificationFixture(t)
	targetOwner := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", targetOwner).Error; err != nil {
		t.Fatal(err)
	}
	target := Calendar{ID: uuid.New(), OwnerID: targetOwner, Title: "New owner", Visibility: "link", ShareToken: newToken(), SharingEnabled: true}
	if err := f.db.Create(&target).Error; err != nil {
		t.Fatal(err)
	}
	moved := false
	callbackName := "test:trip_owner_transfer_" + uuid.NewString()
	// Inject the committed ownership change immediately after the worker's
	// unlocked stop lookup. Its subsequent row lock sees the new calendar.
	if err := f.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if moved || tx.Statement.Table != "tc_stops" {
			return
		}
		if _, locked := tx.Statement.Clauses["FOR"]; locked {
			return
		}
		moved = true
		if err := tx.Session(&gorm.Session{NewDB: true}).Exec("UPDATE tc_stops SET calendar_id = ? WHERE id = ?", target.ID, stop.ID).Error; err != nil {
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.db.Callback().Query().Remove(callbackName) })
	if err := f.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if !moved {
		t.Fatal("ownership race was not injected")
	}
	if f.sender.Attempts != 0 {
		t.Fatal("worker sent while holding the wrong owning-calendar lock")
	}
	var delivery calendarDelivery
	if err := f.db.First(&delivery).Error; err != nil {
		t.Fatal(err)
	}
	if delivery.Status != "pending" {
		t.Fatalf("known unsent update should retry, got %s", delivery.Status)
	}
	f.db.Callback().Query().Remove(callbackName)
	if err := f.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if f.sender.Attempts != 1 {
		t.Fatal("retry with current calendar locks did not deliver")
	}
}
func TestTripNotificationSuppressionPreservesUncertainSends(t *testing.T) {
	f, stop, _ := tripNotificationFixture(t)
	var original calendarDelivery
	if err := f.db.First(&original).Error; err != nil {
		t.Fatal(err)
	}
	uncertain := original
	uncertain.ID = uuid.New()
	uncertain.DedupeKey += "-uncertain"
	uncertain.Status = "sending"
	failed := original
	failed.ID = uuid.New()
	failed.DedupeKey += "-failed"
	failed.Status = "failed"
	if err := f.db.Create(&uncertain).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Create(&failed).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Transaction(func(tx *gorm.DB) error {
		locked, err := f.service.lockedStop(tx, stop.ID.String())
		if err != nil {
			return err
		}
		return f.service.suppressTripLogisticsNotices(tx, locked.ID)
	}); err != nil {
		t.Fatal(err)
	}
	var rows []calendarDelivery
	if err := f.db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.ID == uncertain.ID {
			if row.Status != "sending" {
				t.Fatal("uncertain provider result was incorrectly reclassified")
			}
		} else if row.Status != "skipped" {
			t.Fatal("known unsent stale notice was not suppressed")
		}
	}
	var count int64
	if err := f.db.Model(&calendarNotice{}).Where("stop_id = ?", stop.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("stale in-app notices remained")
	}
}
