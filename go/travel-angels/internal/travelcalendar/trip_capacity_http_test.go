package travelcalendar

import (
	"bytes"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Use independent PostgreSQL connections so simultaneous approvals exercise
// row locking rather than being serialized by the fixture's connection pool.
func tripCapacityConcurrentPool(t *testing.T, f *guestTestFixture) {
	t.Helper()
	var schema string
	if err := f.db.Raw("SELECT current_schema()").Scan(&schema).Error; err != nil {
		t.Fatal(err)
	}
	dsn := os.Getenv("TC_TEST_DATABASE_URL")
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		dsn = parsed.String()
	} else {
		dsn += " search_path=" + schema
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(4)
	t.Cleanup(func() { pool.Close() })
	f.db = db
	f.service.db = db
}

func tripCapacityRevision(t *testing.T, f *guestTestFixture, id string) int {
	t.Helper()
	var stop Stop
	if err := f.db.First(&stop, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return stop.Revision
}

func tripCapacityCreate(t *testing.T, f *guestTestFixture) (string, string) {
	t.Helper()
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "link"}, f.owner, "", 200)
	created := calendarResponse(t, f, "POST", "/calendar/stops", calendarTestContent("Limited places"), f.owner, "", 201)["stop"].(map[string]interface{})
	if created["hostModel"] != true {
		t.Fatal("new trip did not use explicit hosts")
	}
	id := created["id"].(string)
	calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/publish", gin.H{"revision": created["revision"]}, f.owner, "", 200)
	tripCapacitySettings(t, f, id, 2, 200)
	return id, created["shareToken"].(string)
}

func tripCapacitySettings(t *testing.T, f *guestTestFixture, id string, maximum, expected int) {
	t.Helper()
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/settings", gin.H{"revision": tripCapacityRevision(t, f, id), "maxAttendees": maximum, "waitlistEnabled": true, "showAttendeeCount": true, "rosterVisibility": "attendees"}, f.owner, "", expected)
}

func tripCapacityPeople(t *testing.T, f *guestTestFixture) (uuid.UUID, uuid.UUID) {
	t.Helper()
	a, b := uuid.New(), uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?), (?)", a, b).Error; err != nil {
		t.Fatal(err)
	}
	return a, b
}

func tripCapacityRequest(t *testing.T, f *guestTestFixture, id, token string, person uuid.UUID, status, start, end string, expected int) map[string]interface{} {
	t.Helper()
	return calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/participation", gin.H{"status": status, "startDate": start, "endDate": end, "stopToken": token, "rosterConsentVersion": 1}, person, "", expected)
}

func TestTripConcurrentApprovalsRespectWholeTripMaximum(t *testing.T) {
	f := newGuestTestFixture(t)
	tripCapacityConcurrentPool(t, f)
	id, token := tripCapacityCreate(t, f)
	a, b := tripCapacityPeople(t, f)
	first := tripCapacityRequest(t, f, id, token, a, "requested", "2032-12-21", "2032-12-22", 200)
	second := tripCapacityRequest(t, f, id, token, b, "requested", "2032-12-29", "2032-12-30", 200)
	type approval struct {
		code     int
		body, id string
	}
	start := make(chan struct{})
	results := make(chan approval, 2)
	for _, p := range []map[string]interface{}{first, second} {
		participationID := p["id"].(string)
		go func() {
			<-start
			r := httptest.NewRequest("PUT", "/travel-angels/calendar/stops/"+id+"/participations/"+participationID, bytes.NewBufferString(`{"status":"joined"}`))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-Test-User", f.owner.String())
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, r)
			results <- approval{w.Code, w.Body.String(), participationID}
		}()
	}
	close(start)
	one, two := <-results, <-results
	if !((one.code == 200 && two.code == 409) || (one.code == 409 && two.code == 200)) {
		t.Fatalf("expected exactly one approval and one full response: %+v %+v", one, two)
	}
	confirmed, reserved, err := f.service.tripOccupancy(f.db, uuid.MustParse(id))
	if err != nil || confirmed != 2 || reserved != 0 {
		t.Fatalf("capacity overbooked: %d confirmed, %d reserved: %v", confirmed, reserved, err)
	}
	tripCapacitySettings(t, f, id, 1, 400)
	winner, loser := one.id, two.id
	if one.code != 200 {
		winner, loser = two.id, one.id
	}
	var waiting Participation
	if err = f.db.First(&waiting, "id = ?", loser).Error; err != nil {
		t.Fatal(err)
	}
	if waiting.Status != "requested" || waiting.Reserved {
		t.Fatal("first pending request acquired a reservation or was silently waitlisted")
	}
	tripCapacityRequest(t, f, id, token, waiting.UserID, "waitlisted", waiting.StartDate, waiting.EndDate, 200)
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/participations/"+winner, gin.H{"status": "removed"}, f.owner, "", 200)
	if err = f.db.First(&waiting, "id = ?", loser).Error; err != nil {
		t.Fatal(err)
	}
	if waiting.Status != "waitlisted" {
		t.Fatal("a vacancy silently promoted the waitlist")
	}
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/participations/"+loser, gin.H{"status": "joined"}, f.owner, "", 200)
}

func TestTripReservedPlaceSurvivesReconfirmationAndCanBeReleased(t *testing.T) {
	f := newGuestTestFixture(t)
	id, token := tripCapacityCreate(t, f)
	a, b := tripCapacityPeople(t, f)
	first := tripCapacityRequest(t, f, id, token, a, "requested", "2032-12-21", "2032-12-22", 200)
	second := tripCapacityRequest(t, f, id, token, b, "requested", "2032-12-29", "2032-12-30", 200)
	review := "/calendar/stops/" + id + "/participations/" + first["id"].(string)
	calendarResponse(t, f, "PUT", review, gin.H{"status": "joined"}, f.owner, "", 200)
	changed := calendarTestContent("Limited places")
	changed["startDate"], changed["endDate"] = "2033-02-01", "2033-02-15"
	changed["revision"] = tripCapacityRevision(t, f, id)
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id, changed, f.owner, "", 200)
	calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/publish", gin.H{"revision": tripCapacityRevision(t, f, id)}, f.owner, "", 200)
	var p, pending Participation
	f.db.First(&p, "id = ?", first["id"])
	f.db.First(&pending, "id = ?", second["id"])
	if p.Status != "needs_reconfirmation" || !p.Reserved || pending.Reserved {
		t.Fatal("material change did not distinguish confirmed and pending places")
	}
	tripCapacitySettings(t, f, id, 1, 400)
	resubmitted := tripCapacityRequest(t, f, id, token, a, "requested", "2033-02-03", "2033-02-06", 200)
	if resubmitted["reserved"] != true {
		t.Fatal("resubmission lost a reserved place")
	}
	tripCapacityRequest(t, f, id, token, b, "requested", "2033-02-08", "2033-02-10", 409)
	calendarResponse(t, f, "PUT", review, gin.H{"status": "joined"}, f.owner, "", 200)
	confirmed, reserved, err := f.service.tripOccupancy(f.db, uuid.MustParse(id))
	if err != nil || confirmed != 1 || reserved != 1 {
		t.Fatalf("approval double-counted a reservation: %d/%d %v", confirmed, reserved, err)
	}
	resubmitted = tripCapacityRequest(t, f, id, token, a, "requested", "2033-02-04", "2033-02-07", 200)
	if resubmitted["reserved"] != true {
		t.Fatal("changing approved dates lost the place")
	}
	calendarResponse(t, f, "PUT", review, gin.H{"status": "release_reservation"}, f.owner, "", 200)
	tripCapacityRequest(t, f, id, token, b, "requested", "2033-02-08", "2033-02-10", 200)
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/participations/"+second["id"].(string), gin.H{"status": "joined"}, f.owner, "", 200)
	calendarResponse(t, f, "PUT", review, gin.H{"status": "joined"}, f.owner, "", 409)
}
