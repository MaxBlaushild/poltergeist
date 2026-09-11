package travelcalendar

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func calendarResponse(t *testing.T, f *guestTestFixture, method, path string, body interface{}, user uuid.UUID, guest string, expected int) map[string]interface{} {
	t.Helper()
	status, result := f.request(t, method, path, body, user, guest)
	if status != expected {
		t.Fatalf("%s %s: expected %d, got %d: %#v", method, path, expected, status, result)
	}
	return result
}

func calendarTestContent(title string) gin.H {
	return gin.H{"title": title, "destination": "Lisbon, Portugal", "startDate": "2032-12-20", "endDate": "2033-01-10", "datePrecision": "fixed", "description": "Meet us for the holidays", "status": "confirmed", "joinOpen": true}
}

func calendarCreateStop(t *testing.T, f *guestTestFixture, title string) (string, string) {
	t.Helper()
	result := calendarResponse(t, f, "POST", "/calendar/stops", calendarTestContent(title), f.owner, "", 201)
	stop := result["stop"].(map[string]interface{})
	return stop["id"].(string), stop["shareToken"].(string)
}

func TestCalendarPublishingAndReplacingPermissions(t *testing.T) {
	f := newGuestTestFixture(t)
	first := calendarResponse(t, f, "GET", "/calendar", nil, f.owner, "", 200)
	second := calendarResponse(t, f, "GET", "/calendar", nil, f.owner, "", 200)
	cal := first["calendar"].(map[string]interface{})
	if cal["id"] != second["calendar"].(map[string]interface{})["id"] || cal["visibility"] != "private" {
		t.Fatal("a user must retain one initially private ongoing calendar")
	}
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "link"}, f.owner, "", 200)
	id, token := calendarCreateStop(t, f, "Original public title")
	sharedPath := "/shared/calendars/" + cal["shareToken"].(string)
	draftView := calendarResponse(t, f, "GET", sharedPath, nil, uuid.Nil, "", 200)
	if len(draftView["stops"].([]interface{})) != 0 {
		t.Fatal("draft leaked to the shared calendar")
	}
	calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/publish", nil, f.owner, "", 200)
	public := calendarResponse(t, f, "GET", "/shared/stops/"+token, nil, uuid.Nil, "", 200)["stop"].(map[string]interface{})
	for _, field := range []string{"draft", "shareToken", "grants", "recipientId"} {
		if _, exists := public[field]; exists {
			t.Fatalf("guest stop contains privileged field %s", field)
		}
	}
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id, calendarTestContent("Unpublished revision"), f.owner, "", 200)
	public = calendarResponse(t, f, "GET", "/shared/stops/"+token, nil, uuid.Nil, "", 200)["stop"].(map[string]interface{})
	if public["title"] != "Original public title" {
		t.Fatal("saving a draft replaced published information")
	}
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/permissions", gin.H{"visibility": "specific", "grants": []gin.H{}}, f.owner, "", 200)
	calendarResponse(t, f, "GET", "/shared/stops/"+token, nil, uuid.Nil, "", 404)
	restricted := calendarResponse(t, f, "GET", sharedPath, nil, uuid.Nil, "", 200)
	if len(restricted["stops"].([]interface{})) != 0 {
		t.Fatal("specific-stop override did not replace the broadly shared calendar audience")
	}
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "private"}, f.owner, "", 200)
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "link"}, f.owner, "", 200)
	calendarResponse(t, f, "GET", "/shared/stops/"+token, nil, uuid.Nil, "", 404)
}

func TestCalendarClaimParticipationChangesAndRevocation(t *testing.T) {
	f := newGuestTestFixture(t)
	id, link := calendarCreateStop(t, f, "Restricted Lisbon stop")
	calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/publish", nil, f.owner, "", 200)
	guest, recipient := f.verify(t, "calendar-friend@example.com")
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/permissions", gin.H{"visibility": "specific", "grants": []gin.H{{"recipientId": recipient}}}, f.owner, "", 200)
	friend, stranger := uuid.New(), uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?), (?)", friend, stranger).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "GET", "/shared/stops/"+link, nil, uuid.Nil, guest, 200)
	request := gin.H{"status": "requested", "startDate": "2032-12-22", "endDate": "2032-12-28"}
	participationPath := "/calendar/stops/" + id + "/participation"
	calendarResponse(t, f, "POST", participationPath, request, uuid.Nil, guest, 401)
	// Logging in cannot borrow the guest's unclaimed invitation by skipping claim.
	calendarResponse(t, f, "POST", participationPath, request, friend, guest, 404)
	calendarResponse(t, f, "POST", "/calendar-guests/claim", nil, friend, guest, 200)
	calendarResponse(t, f, "POST", "/calendar-guests/claim", nil, stranger, guest, 409)
	joinedRequest := calendarResponse(t, f, "POST", participationPath, request, friend, "", 200)
	if joinedRequest["status"] != "requested" {
		t.Fatal("claiming implicitly approved attendance")
	}
	second := calendarResponse(t, f, "POST", participationPath, request, friend, "", 200)
	if second["id"] != joinedRequest["id"] {
		t.Fatal("repeated join created a duplicate participation")
	}
	reviewPath := "/calendar/stops/" + id + "/participations/" + joinedRequest["id"].(string)
	calendarResponse(t, f, "PUT", reviewPath, gin.H{"status": "joined"}, stranger, "", 403)
	calendarResponse(t, f, "PUT", reviewPath, gin.H{"status": "joined"}, f.owner, "", 200)
	revised := calendarTestContent("Restricted Lisbon stop")
	revised["startDate"], revised["endDate"] = "2033-02-01", "2033-02-15"
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id, revised, f.owner, "", 200)
	calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/publish", nil, f.owner, "", 200)
	var participation Participation
	if err := f.db.First(&participation, "id = ?", joinedRequest["id"]).Error; err != nil {
		t.Fatal(err)
	}
	if participation.Status != "needs_reconfirmation" || participation.StartDate != "2032-12-22" {
		t.Fatal("published change silently moved the friend's commitment")
	}
	permissionsPath := "/calendar/stops/" + id + "/permissions"
	preview := calendarResponse(t, f, "POST", permissionsPath+"/preview", gin.H{"visibility": "private"}, f.owner, "", 200)
	if preview["removedParticipationCount"] != float64(1) {
		t.Fatalf("permission preview did not identify affected attendance: %#v", preview)
	}
	calendarResponse(t, f, "GET", "/shared/stops/"+link, nil, friend, "", 200)
	calendarResponse(t, f, "PUT", permissionsPath, gin.H{"visibility": "private"}, f.owner, "", 409)
	calendarResponse(t, f, "PUT", permissionsPath, gin.H{"visibility": "private", "confirmRemoval": true}, f.owner, "", 200)
	calendarResponse(t, f, "GET", "/shared/stops/"+link, nil, friend, guest, 404)
	f.db.First(&participation, "id = ?", participation.ID)
	if participation.Status != "removed" {
		t.Fatal("revoked participant remained active")
	}
}

func TestCalendarCohostReferencesDoNotExpandAudience(t *testing.T) {
	f := newGuestTestFixture(t)
	id, _ := calendarCreateStop(t, f, "Private shared stop")
	calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/publish", nil, f.owner, "", 200)
	partner, editor := uuid.New(), uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?), (?)", partner, editor).Error; err != nil {
		t.Fatal(err)
	}
	partnerResult := calendarResponse(t, f, "GET", "/calendar", nil, partner, "", 200)
	partnerCalendar := partnerResult["calendar"].(map[string]interface{})
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "link"}, partner, "", 200)
	invite := calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/collaborators", gin.H{"userId": partner}, f.owner, "", 200)
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+invite["id"].(string)+"/accept", nil, partner, "", 200)
	calendarResponse(t, f, "GET", "/calendar/stops/"+id, nil, partner, "", 200)
	calendarResponse(t, f, "PUT", "/calendar/stops/"+id+"/permissions", gin.H{"visibility": "link"}, partner, "", 403)
	shared := calendarResponse(t, f, "GET", "/shared/calendars/"+partnerCalendar["shareToken"].(string), nil, uuid.Nil, "", 200)
	if len(shared["stops"].([]interface{})) != 0 {
		t.Fatal("cohost calendar broadened the owning calendar's private audience")
	}
	editorInvite := calendarResponse(t, f, "POST", "/calendar/collaborators", gin.H{"userId": editor}, partner, "", 200)
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+editorInvite["id"].(string)+"/accept", nil, editor, "", 200)
	calendarResponse(t, f, "GET", "/calendar/stops/"+id, nil, editor, "", 404)
}

func TestCalendarMigrationsCanRollbackAndReapply(t *testing.T) {
	f := newGuestTestFixture(t)
	_, source, _, _ := runtime.Caller(0)
	for _, name := range []string{
		"000474_travel_calendar_guests.down.sql", "000473_travel_calendars.down.sql",
		"000473_travel_calendars.up.sql", "000474_travel_calendar_guests.up.sql",
	} {
		contents, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.db.Exec(string(contents)).Error; err != nil {
			t.Fatalf("migration %s: %v", name, err)
		}
	}
	calendarResponse(t, f, "GET", "/calendar", nil, f.owner, "", 200)
}
