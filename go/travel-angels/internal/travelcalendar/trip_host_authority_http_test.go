package travelcalendar

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func tripAuthorityRevision(t *testing.T, f *guestTestFixture, id string) int {
	t.Helper()
	var stop Stop
	if err := f.db.First(&stop, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return stop.Revision
}
func tripAuthorityCreate(t *testing.T, f *guestTestFixture) map[string]interface{} {
	t.Helper()
	created := calendarResponse(t, f, "POST", "/calendar/stops", calendarTestContent("Host authority"), f.owner, "", 201)["stop"].(map[string]interface{})
	if created["hostModel"] != true {
		t.Fatal("new trip must use explicit hosts")
	}
	return created
}
func tripAuthorityAccount(t *testing.T, f *guestTestFixture) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", id).Error; err != nil {
		t.Fatal(err)
	}
	return id
}
func tripAuthorityInvite(t *testing.T, f *guestTestFixture, id string, person uuid.UUID) string {
	t.Helper()
	invitation := calendarResponse(t, f, "POST", "/calendar/stops/"+id+"/collaborators", gin.H{"revision": tripAuthorityRevision(t, f, id), "userId": person, "canManagePermissions": false}, f.owner, "", 200)
	if invitation["accepted"] != false || invitation["isHost"] != true || invitation["canManagePermissions"] != true {
		t.Fatal("new trip must issue a pending full-host invitation")
	}
	return invitation["id"].(string)
}
func tripAuthorityAcceptOrganizer(t *testing.T, f *guestTestFixture, invitation string, person uuid.UUID) {
	t.Helper()
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+invitation+"/accept", gin.H{"attending": false, "rosterConsentVersion": 1}, person, "", 200)
}

func TestTripHostAuthorityRequiresNamedAcceptanceAndSeparatesCalendarEditors(t *testing.T) {
	f := newGuestTestFixture(t)
	created := tripAuthorityCreate(t, f)
	id := created["id"].(string)
	path := "/calendar/stops/" + id
	editor, host, attendee := tripAuthorityAccount(t, f), tripAuthorityAccount(t, f), tripAuthorityAccount(t, f)
	editorInvitation := calendarResponse(t, f, "POST", "/calendar/collaborators", gin.H{"userId": editor, "canManagePermissions": true}, f.owner, "", 200)
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+editorInvitation["id"].(string)+"/accept", nil, editor, "", 200)
	otherCalendarTrip := calendarTestContent("Cannot create for another owner")
	otherCalendarTrip["calendarId"] = created["calendarId"]
	calendarResponse(t, f, "POST", "/calendar/stops", otherCalendarTrip, editor, "", 403)
	editorDraft := calendarTestContent("Editor must not overwrite this trip")
	editorDraft["revision"] = tripAuthorityRevision(t, f, id)
	calendarResponse(t, f, "PUT", path, editorDraft, editor, "", 403)
	calendarResponse(t, f, "GET", path, nil, editor, "", 404)
	invitation := tripAuthorityInvite(t, f, id, host)
	pendingDraft := calendarTestContent("Pending invitation is not authority")
	pendingDraft["revision"] = tripAuthorityRevision(t, f, id)
	calendarResponse(t, f, "PUT", path, pendingDraft, host, "", 403)
	calendarResponse(t, f, "PUT", path+"/permissions", gin.H{"revision": tripAuthorityRevision(t, f, id), "visibility": "link"}, host, "", 403)
	calendarResponse(t, f, "GET", path, nil, host, "", 404)
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+invitation+"/accept", gin.H{"attending": false}, editor, "", 404)
	calendarResponse(t, f, "POST", "/calendar/collaborators/"+invitation+"/accept", gin.H{"attending": false}, uuid.Nil, "", 401)
	tripAuthorityAcceptOrganizer(t, f, invitation, host)
	organizer := calendarResponse(t, f, "GET", path, nil, host, "", 200)["stop"].(map[string]interface{})
	if organizer["canManage"] != true || organizer["canManagePermissions"] != true || organizer["canManageCollaborators"] != true {
		t.Fatal("accepted organizer did not receive full host authority")
	}
	capacity := organizer["capacity"].(map[string]interface{})
	if capacity["confirmedCount"] != float64(1) || capacity["reservedCount"] != float64(0) {
		t.Fatal("non-attending host consumed a place")
	}
	if organizer["ownParticipation"].(map[string]interface{})["status"] != "withdrawn" {
		t.Fatal("organizing silently joined travel")
	}
	draft := calendarTestContent("Edited by accepted organizer")
	draft["revision"] = tripAuthorityRevision(t, f, id)
	calendarResponse(t, f, "PUT", path, draft, host, "", 200)
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": tripAuthorityRevision(t, f, id)}, host, "", 200)
	calendarResponse(t, f, "PUT", path+"/permissions", gin.H{"revision": tripAuthorityRevision(t, f, id), "visibility": "link"}, host, "", 200)
	requested := calendarResponse(t, f, "POST", path+"/participation", gin.H{"status": "requested", "startDate": "2032-12-22", "endDate": "2032-12-28", "stopToken": created["shareToken"], "rosterConsentVersion": 1}, attendee, "", 200)
	calendarResponse(t, f, "PUT", path+"/participations/"+requested["id"].(string), gin.H{"revision": tripAuthorityRevision(t, f, id), "status": "joined"}, host, "", 200)
	joined := calendarResponse(t, f, "GET", path, nil, attendee, "", 200)["stop"].(map[string]interface{})
	if joined["canManage"] != false || joined["canManagePermissions"] != false {
		t.Fatal("attendance granted host authority")
	}
	attendeeDraft := calendarTestContent("Attendee cannot edit")
	attendeeDraft["revision"] = tripAuthorityRevision(t, f, id)
	calendarResponse(t, f, "PUT", path, attendeeDraft, attendee, "", 403)
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": tripAuthorityRevision(t, f, id)}, attendee, "", 403)
	calendarResponse(t, f, "PUT", path+"/permissions", gin.H{"revision": tripAuthorityRevision(t, f, id), "visibility": "private"}, attendee, "", 403)
	calendarResponse(t, f, "POST", path+"/collaborators", gin.H{"revision": tripAuthorityRevision(t, f, id), "userId": editor}, attendee, "", 403)
	calendarResponse(t, f, "PUT", path+"/settings", gin.H{"revision": tripAuthorityRevision(t, f, id), "maxAttendees": 10, "rosterVisibility": "attendees", "showAttendeeCount": true}, attendee, "", 403)
	calendarResponse(t, f, "DELETE", path, gin.H{"revision": tripAuthorityRevision(t, f, id), "confirm": true}, attendee, "", 403)
}

func TestTripOtherHostCannotRemoveOwnerOrUseOwnerContinuityPowers(t *testing.T) {
	f := newGuestTestFixture(t)
	created := tripAuthorityCreate(t, f)
	id := created["id"].(string)
	path := "/calendar/stops/" + id
	host := tripAuthorityAccount(t, f)
	invitation := tripAuthorityInvite(t, f, id, host)
	tripAuthorityAcceptOrganizer(t, f, invitation, host)
	// An owner may also have a collaborator row after previously serving as a
	// co-host. Its existence must not make ownership removable through this API.
	stopID := uuid.MustParse(id)
	ownerRole := Collaborator{ID: uuid.New(), StopID: &stopID, UserID: f.owner, IsHost: true, Accepted: true, CanManagePermissions: true}
	if err := f.db.Create(&ownerRole).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "DELETE", path+"/collaborators/"+ownerRole.ID.String(), gin.H{"revision": tripAuthorityRevision(t, f, id), "confirmRemoval": true}, host, "", 400)
	calendarResponse(t, f, "POST", path+"/collaborators/"+ownerRole.ID.String()+"/removal-preview", gin.H{"revision": tripAuthorityRevision(t, f, id)}, host, "", 400)
	calendarResponse(t, f, "DELETE", path, gin.H{"revision": tripAuthorityRevision(t, f, id), "confirm": true}, host, "", 403)
	calendarResponse(t, f, "POST", path+"/transfer-ownership", gin.H{"revision": tripAuthorityRevision(t, f, id), "userId": f.owner, "confirm": true}, host, "", 403)
	calendarResponse(t, f, "POST", path+"/leave-host", gin.H{"revision": tripAuthorityRevision(t, f, id), "confirmRemoval": true}, f.owner, "", 400)
	owner := calendarResponse(t, f, "GET", path, nil, f.owner, "", 200)["stop"].(map[string]interface{})
	if owner["ownerId"] != f.owner.String() || owner["canManage"] != true {
		t.Fatal("failed removal changed owner authority")
	}
	var count int64
	if err := f.db.Model(&Collaborator{}).Where("id = ?", ownerRole.ID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("owner collaborator record was removed")
	}
}
