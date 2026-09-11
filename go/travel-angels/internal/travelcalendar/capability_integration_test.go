package travelcalendar

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestParticipationRetainsOnlyObservedSharingCapabilities(t *testing.T) {
	f := newGuestTestFixture(t)
	ownerCalendar, err := f.service.primaryCalendar(f.db, f.owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.primaryCalendar(f.db, f.owner); err != nil {
		t.Fatalf("primary calendar must be idempotent: %v", err)
	}
	content := StopContent{Title: "Meet in Tokyo", Destination: "Tokyo", DatePrecision: "fixed", StartDate: "2030-05-01", EndDate: "2030-05-10", Status: "confirmed", JoinOpen: true}
	stop := Stop{ID: uuid.New(), CalendarID: ownerCalendar.ID, Draft: content, Published: &content, Visibility: "link", ShareToken: newToken(), SharingEnabled: true}
	if err = f.db.Create(&stop).Error; err != nil {
		t.Fatal(err)
	}
	users := []uuid.UUID{uuid.New(), uuid.New()}
	for _, id := range users {
		if err = f.db.Exec("INSERT INTO users(id) VALUES (?)", id).Error; err != nil {
			t.Fatal(err)
		}
	}
	path := "/calendar/stops/" + stop.ID.String()
	for i, user := range users {
		body := map[string]interface{}{"status": "requested", "startDate": "2030-05-03", "endDate": "2030-05-06"}
		if i == 0 {
			body["stopToken"] = stop.ShareToken
		} else {
			body["calendarToken"] = ownerCalendar.ShareToken
		}
		code, response := f.request(t, http.MethodPost, path+"/participation", body, user, "")
		if code != 200 {
			t.Fatalf("request %d = %d: %v", i, code, response)
		}
		code, response = f.request(t, http.MethodGet, path, nil, user, "")
		if code != 200 {
			t.Fatalf("stored capability cannot reopen its stop: %d %v", code, response)
		}
	}
	code, response := f.request(t, http.MethodPost, path+"/share-link", map[string]interface{}{"enabled": true, "rotate": true}, f.owner, "")
	if code != 409 || response["removedParticipationCount"] != float64(1) {
		t.Fatalf("rotation must preview only the stop-link participant: %d %v", code, response)
	}
	var unchanged Stop
	if err = f.db.First(&unchanged, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.ShareToken != stop.ShareToken {
		t.Fatal("unconfirmed rotation must roll back")
	}
	code, response = f.request(t, http.MethodPost, path+"/share-link", map[string]interface{}{"enabled": true, "rotate": true, "confirmRemoval": true}, f.owner, "")
	if code != 200 {
		t.Fatalf("rotation failed: %d %v", code, response)
	}
	for i, user := range users {
		var p Participation
		if err = f.db.Where("stop_id = ? AND user_id = ?", stop.ID, user).First(&p).Error; err != nil {
			t.Fatal(err)
		}
		want := "removed"
		if i == 1 {
			want = "requested"
		}
		if p.Status != want {
			t.Fatalf("participant %d is %s, want %s", i, p.Status, want)
		}
	}
	code, _ = f.request(t, http.MethodGet, "/shared/stops/"+stop.ShareToken, nil, uuid.Nil, "")
	if code != 404 {
		t.Fatal("rotated token still resolves")
	}
	code, _ = f.request(t, http.MethodGet, path, nil, users[0], "")
	if code != 404 {
		t.Fatal("removed participation manufactured a new alternate capability")
	}
	code, _ = f.request(t, http.MethodGet, path, nil, users[1], "")
	if code != 200 {
		t.Fatal("independently held calendar link should retain access")
	}
}
