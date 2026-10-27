package travelcalendar

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestTripTitleLimitCountsCharacters(t *testing.T) {
	content := StopContent{Title: strings.Repeat("旅", 200), Destination: "京都", DatePrecision: "tbd"}
	if err := content.validate(); err != nil {
		t.Fatalf("200 Unicode characters rejected: %v", err)
	}
	content.Title += "旅"
	if content.validate() == nil {
		t.Fatal("201 characters accepted")
	}
}

func TestTripLegacyConversionRequiresActivationAndFreshHostAcceptance(t *testing.T) {
	f := newGuestTestFixture(t)
	id, _ := calendarCreateLegacyStop(t, f, "Legacy calendar stop")
	path := "/calendar/stops/" + id
	editor := tripAuthorityAccount(t, f)
	invite := calendarResponse(t, f, "POST", "/calendar/collaborators", gin.H{"userId": editor, "canManagePermissions": true}, f.owner, "", 200)
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+invite["id"].(string)+"/accept", nil, editor, "", 200)
	calendarResponse(t, f, "GET", path, nil, editor, "", 200)
	input := gin.H{"revision": tripAuthorityRevision(t, f, id), "hostUserIds": []uuid.UUID{editor}}
	preview := calendarResponse(t, f, "POST", path+"/host-conversion/preview", input, f.owner, "", 200)
	if len(preview["losingManagement"].([]interface{})) != 1 {
		t.Fatalf("missing affected editor: %#v", preview)
	}
	calendarResponse(t, f, "GET", path, nil, editor, "", 200)
	var prior Stop
	f.db.First(&prior, "id = ?", id)
	if prior.HostModel {
		t.Fatal("preview activated new authority")
	}
	input["confirm"] = true
	calendarResponse(t, f, "POST", path+"/host-conversion", input, editor, "", 403)
	calendarResponse(t, f, "POST", path+"/host-conversion", input, f.owner, "", 200)
	calendarResponse(t, f, "GET", path, nil, editor, "", 404)
	var pending Collaborator
	if err := f.db.Where("stop_id = ? AND user_id = ?", id, editor).First(&pending).Error; err != nil {
		t.Fatal(err)
	}
	if pending.Accepted || !pending.IsHost {
		t.Fatal("conversion silently promoted existing editor")
	}
	var attendance int64
	f.db.Model(&Participation{}).Where("stop_id = ?", id).Count(&attendance)
	if attendance != 0 {
		t.Fatal("legacy conversion fabricated owner or editor attendance")
	}
	tripAuthorityAcceptOrganizer(t, f, pending.ID.String(), editor)
	current := calendarResponse(t, f, "GET", path, nil, editor, "", 200)["stop"].(map[string]interface{})
	if current["canManagePermissions"] != true {
		t.Fatal("accepted full host lacks permission authority")
	}
}

func TestTripOwnershipTransferBindsAudiencePreviewAndPreservesContinuity(t *testing.T) {
	f := newGuestTestFixture(t)
	created := tripAuthorityCreate(t, f)
	id := created["id"].(string)
	path := "/calendar/stops/" + id
	target := tripAuthorityAccount(t, f)
	invitation := tripAuthorityInvite(t, f, id, target)
	tripAuthorityAcceptOrganizer(t, f, invitation, target)
	oldViewer, newViewer := tripCapacityPeople(t, f)
	otherViewer := tripAuthorityAccount(t, f)
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "specific", "grants": []gin.H{{"userId": oldViewer}}}, f.owner, "", 200)
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "specific", "grants": []gin.H{{"userId": newViewer}}}, target, "", 200)
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": tripAuthorityRevision(t, f, id)}, f.owner, "", 200)
	request := calendarResponse(t, f, "POST", path+"/participation", gin.H{"status": "requested", "startDate": "2032-12-22", "endDate": "2032-12-23", "rosterConsentVersion": 1}, oldViewer, "", 200)
	calendarResponse(t, f, "PUT", path+"/participations/"+request["id"].(string), gin.H{"status": "joined"}, f.owner, "", 200)
	input := gin.H{"revision": tripAuthorityRevision(t, f, id), "userId": target}
	preview := calendarResponse(t, f, "POST", path+"/transfer-ownership/preview", input, f.owner, "", 200)
	change := preview["audienceChange"].(map[string]interface{})
	if change["addedViewerCount"] != float64(1) || change["removedViewerCount"] != float64(1) || preview["removedParticipationCount"] != float64(1) {
		t.Fatalf("incomplete preview: %#v", preview)
	}
	calendarResponse(t, f, "GET", path, nil, oldViewer, "", 200)
	calendarResponse(t, f, "GET", path, nil, newViewer, "", 404)
	input["confirm"] = true
	calendarResponse(t, f, "POST", path+"/transfer-ownership", input, f.owner, "", 409)
	input["audienceFingerprint"] = preview["audienceFingerprint"]
	// The source revision stays unchanged when the target calendar audience changes.
	calendarResponse(t, f, "PUT", "/calendar/permissions", gin.H{"visibility": "specific", "grants": []gin.H{{"userId": newViewer}, {"userId": otherViewer}}}, target, "", 200)
	calendarResponse(t, f, "POST", path+"/transfer-ownership", input, f.owner, "", 409)
	preview = calendarResponse(t, f, "POST", path+"/transfer-ownership/preview", input, f.owner, "", 200)
	input["audienceFingerprint"] = preview["audienceFingerprint"]
	moved := calendarResponse(t, f, "POST", path+"/transfer-ownership", input, f.owner, "", 200)["stop"].(map[string]interface{})
	if moved["ownerId"] != target.String() || moved["canManage"] != true {
		t.Fatal("ownership/previous-host continuity lost")
	}
	calendarResponse(t, f, "GET", path, nil, oldViewer, "", 404)
	calendarResponse(t, f, "GET", path, nil, newViewer, "", 200)
	var participation Participation
	f.db.First(&participation, "id = ?", request["id"])
	if participation.Status != "removed" || participation.Reserved {
		t.Fatal("transfer retained attendee without effective access")
	}
	calendarResponse(t, f, "DELETE", path, gin.H{"revision": tripAuthorityRevision(t, f, id), "confirm": true}, f.owner, "", 403)
	calendarResponse(t, f, "DELETE", path, gin.H{"revision": tripAuthorityRevision(t, f, id), "confirm": true}, target, "", 200)
	var stops, audits int64
	f.db.Model(&Stop{}).Where("id = ?", id).Count(&stops)
	f.db.Model(&tripAuditEvent{}).Where("stop_id = ? AND action = ?", id, "trip_deleted").Count(&audits)
	if stops != 0 || audits != 1 {
		t.Fatalf("delete/audit continuity failed: stops=%d audits=%d", stops, audits)
	}
}

func TestTripCloseAttendanceIsImmediateWithoutPublishingOtherDrafts(t *testing.T) {
	f := newGuestTestFixture(t)
	id, token := tripCapacityCreate(t, f)
	path := "/calendar/stops/" + id
	revised := calendarTestContent("Unpublished next title")
	revised["revision"] = tripAuthorityRevision(t, f, id)
	calendarResponse(t, f, "PUT", path, revised, f.owner, "", 200)
	calendarResponse(t, f, "PUT", path+"/settings", gin.H{"revision": tripAuthorityRevision(t, f, id), "maxAttendees": 2, "rosterVisibility": "attendees", "showAttendeeCount": true, "joinOpen": false}, f.owner, "", 200)
	public := calendarResponse(t, f, "GET", "/shared/stops/"+token, nil, uuid.Nil, "", 200)["stop"].(map[string]interface{})
	if public["joinOpen"] != false || public["title"] != "Limited places" {
		t.Fatal("closure was deferred or published unrelated draft content")
	}
	person := tripAuthorityAccount(t, f)
	tripCapacityRequest(t, f, id, token, person, "requested", "2032-12-22", "2032-12-23", 400)
	var stop Stop
	f.db.First(&stop, "id = ?", id)
	if stop.Draft.JoinOpen || stop.Published.JoinOpen || stop.Draft.Title != "Unpublished next title" {
		t.Fatal("immediate control did not preserve draft state")
	}
}

func TestTripMigrationPreservesLegacyAuthorityAndDoesNotInventAttendance(t *testing.T) {
	f := newGuestTestFixture(t)
	_, source, _, _ := runtime.Caller(0)
	migrate := func(name string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if err = f.db.Exec(string(data)).Error; err != nil {
			t.Fatal(err)
		}
	}
	migrate("000479_travel_lodging_media.down.sql")
	migrate("000478_travel_trip_media_storage.down.sql")
	migrate("000475_travel_trip_pages.down.sql")
	cal, err := f.service.primaryCalendar(f.db, f.owner)
	if err != nil {
		t.Fatal(err)
	}
	stopID := uuid.New()
	editor := tripAuthorityAccount(t, f)
	if err = f.db.Exec("INSERT INTO tc_stops(id,calendar_id,draft,visibility,share_token,sharing_enabled) VALUES (?,?,'{}','inherit',?,true)", stopID, cal.ID, newToken()).Error; err != nil {
		t.Fatal(err)
	}
	if err = f.db.Exec("INSERT INTO tc_collaborators(id,stop_id,user_id,accepted,can_manage_permissions) VALUES (?,?,?,true,true)", uuid.New(), stopID, editor).Error; err != nil {
		t.Fatal(err)
	}
	migrate("000475_travel_trip_pages.up.sql")
	migrate("000478_travel_trip_media_storage.up.sql")
	migrate("000479_travel_lodging_media.up.sql")
	var stop Stop
	if err = f.db.First(&stop, "id = ?", stopID).Error; err != nil {
		t.Fatal(err)
	}
	if stop.HostModel || !f.service.canManagePermissions(f.db, &stop, Viewer{UserID: editor}) {
		t.Fatal("migration changed legacy editor authority")
	}
	var count int64
	f.db.Model(&Participation{}).Where("stop_id = ?", stopID).Count(&count)
	if count != 0 {
		t.Fatal("migration invented attendance")
	}
}
