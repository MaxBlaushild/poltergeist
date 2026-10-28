package travelcalendar

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func flightFixture(t *testing.T) (*guestTestFixture, Stop, string) {
	t.Helper()
	f := newGuestTestFixture(t)
	_, source, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations/000480_travel_flight_selections.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.db.Exec(string(data)).Error; err != nil {
		t.Fatal(err)
	}
	_, stop := f.plans(t)
	stop.HostModel, stop.RosterVersion, stop.RosterVisibility = true, 1, "attendees"
	id := uuid.NewString()
	content := tripLogisticsContent()
	content.Flights = []TripFlight{{ID: id, Label: "Japan flight", Kind: "group", Visibility: "viewers", Status: "planned", HostNotes: "PRIVATE HOST NOTE", AssignedUserIDs: []uuid.UUID{f.owner}, Legs: []TripFlightLeg{
		{Airline: "ANA", FlightNumber: "NH109", DepartureAirport: "JFK", ArrivalAirport: "HND", DepartureDate: "2032-12-22", DepartureTime: "13:00", DepartureTimeZone: "America/New_York", ArrivalDate: "2032-12-23", ArrivalTime: "17:00", ArrivalTimeZone: "Asia/Tokyo"},
		{Airline: "ANA", FlightNumber: "NH25", DepartureAirport: "HND", ArrivalAirport: "ITM", DepartureDate: "2032-12-24", DepartureTime: "12:00", DepartureTimeZone: "Asia/Tokyo", ArrivalDate: "2032-12-24", ArrivalTime: "13:00", ArrivalTimeZone: "Asia/Tokyo"},
	}}}
	// Deliberately persist legacy content without leg IDs.
	stop.Draft, stop.Published = content, &content
	if err = f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	if err = f.db.First(&stop, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	return f, stop, id
}

func flightPath(stop Stop, id string) string {
	return "/calendar/stops/" + stop.ID.String() + "/flights/" + id
}

func flightTraveler(t *testing.T, f *guestTestFixture, stop Stop) uuid.UUID {
	t.Helper()
	user := tripAuthorityAccount(t, f)
	p := Participation{ID: uuid.New(), StopID: stop.ID, UserID: user, Status: "joined", StartDate: stop.Draft.StartDate, EndDate: stop.Draft.EndDate, RosterConsentVersion: stop.RosterVersion, StopLinkTokenHash: guestDigest(stop.ShareToken)}
	if err := f.db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func flightGet(t *testing.T, f *guestTestFixture, stop Stop, id string, user uuid.UUID) map[string]interface{} {
	t.Helper()
	return calendarResponse(t, f, "GET", flightPath(stop, id)+"/travelers?stopToken="+stop.ShareToken, nil, user, "", 200)
}

func flightChoose(t *testing.T, f *guestTestFixture, stop Stop, id string, user uuid.UUID, state map[string]interface{}, legs []string, share bool, status string, expected int) map[string]interface{} {
	t.Helper()
	return calendarResponse(t, f, "POST", flightPath(stop, id)+"/selection?stopToken="+stop.ShareToken, gin.H{"status": status, "legIds": legs, "flightVersion": state["flightVersion"], "rosterVersion": state["rosterVersion"], "shareWithTravelers": share}, user, "", expected)
}

func TestFlightSelectionAccessGuestAndSelfOnly(t *testing.T) {
	f, stop, id := flightFixture(t)
	guest := flightGet(t, f, stop, id, uuid.Nil)
	if guest["requiresAccount"] != true || guest["canSelect"] != false || guest["canViewTravelers"] != false || len(guest["travelers"].([]interface{})) != 0 {
		t.Fatal("guest received participation controls or roster", guest)
	}
	leg := stop.Published.Flights[0].Legs[0].ID
	flightChoose(t, f, stop, id, uuid.Nil, guest, []string{leg}, true, "booked", 401)
	outsider := tripAuthorityAccount(t, f)
	view := flightGet(t, f, stop, id, outsider)
	if view["requiresAttendance"] != true {
		t.Fatal("unjoined viewer not prompted to join")
	}
	flightChoose(t, f, stop, id, outsider, view, []string{leg}, true, "booked", 403)
	attendee := flightTraveler(t, f, stop)
	state := flightGet(t, f, stop, id, attendee)
	if state["canSelect"] != true {
		t.Fatal("eligible attendee cannot select")
	}
	flightChoose(t, f, stop, id, attendee, state, []string{"unknown"}, true, "booked", 400)
	flightChoose(t, f, stop, id, attendee, state, []string{leg, leg}, true, "booked", 400)
	result := calendarResponse(t, f, "POST", flightPath(stop, id)+"/selection?stopToken="+stop.ShareToken, gin.H{"status": "booked", "legIds": []string{leg}, "flightVersion": state["flightVersion"], "rosterVersion": state["rosterVersion"], "shareWithTravelers": true, "userId": outsider}, attendee, "", 200)
	if result["ownSelection"].(map[string]interface{})["status"] != "booked" {
		t.Fatal("own selection missing")
	}
	var rows []FlightSelection
	f.db.Find(&rows)
	if len(rows) != 1 || rows[0].UserID != attendee {
		t.Fatal("request could select on behalf of another user")
	}
	serialized, _ := json.Marshal(result)
	if strings.Contains(string(serialized), "PRIVATE HOST NOTE") || strings.Contains(string(serialized), f.owner.String()) {
		t.Fatal("private assignments/host notes leaked", string(serialized))
	}
	// The host can coordinate their own flights without being an attendee.
	host := flightGet(t, f, stop, id, f.owner)
	flightChoose(t, f, stop, id, f.owner, host, []string{leg}, true, "considering", 200)
	calendarResponse(t, f, "GET", flightPath(stop, id)+"/travelers", nil, outsider, "", 404)
	stop.Published = nil
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "GET", flightPath(stop, id)+"/travelers", nil, f.owner, "", 404)
}

func TestFlightRosterConsentAndRevokedAttendance(t *testing.T) {
	f, stop, id := flightFixture(t)
	a, b := flightTraveler(t, f, stop), flightTraveler(t, f, stop)
	leg := stop.Published.Flights[0].Legs[0].ID
	state := flightGet(t, f, stop, id, a)
	flightChoose(t, f, stop, id, a, state, []string{leg}, false, "considering", 200)
	audit := calendarResponse(t, f, "GET", "/calendar/stops/"+stop.ID.String()+"/audit", nil, f.owner, "", 200)
	auditJSON, _ := json.Marshal(audit)
	if strings.Contains(string(auditJSON), a.String()) || strings.Contains(string(auditJSON), id) || strings.Contains(string(auditJSON), "considering") {
		t.Fatal("personal flight choice leaked through the host audit", string(auditJSON))
	}
	if got := flightGet(t, f, stop, id, f.owner); len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("unshared choice leaked even to host")
	}
	flightChoose(t, f, stop, id, a, state, []string{leg}, true, "booked", 200)
	if got := flightGet(t, f, stop, id, b); len(got["travelers"].([]interface{})) != 1 {
		t.Fatal("consenting selection missing")
	}
	if err := f.db.Model(&Participation{}).Where("stop_id = ? AND user_id = ?", stop.ID, a).Update("roster_consent_version", 0).Error; err != nil {
		t.Fatal(err)
	}
	if got := flightGet(t, f, stop, id, b); len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("unconsented attendee leaked")
	}
	f.db.Model(&Participation{}).Where("stop_id = ? AND user_id = ?", stop.ID, a).Update("roster_consent_version", 1)
	stop.RosterVisibility = "hosts"
	f.db.Save(&stop)
	if got := flightGet(t, f, stop, id, b); got["canViewTravelers"] != false || got["canShareWithTravelers"] != true || got["sharingAudience"] != "hosts" || len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("host-only roster leaked")
	}
	stop.RosterVisibility = "attendees"
	stop.RosterVersion = 2
	f.db.Save(&stop)
	if got := flightGet(t, f, stop, id, f.owner); len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("expanded audience reused stale flight sharing consent")
	}
	flightChoose(t, f, stop, id, a, state, []string{leg}, true, "booked", 409)
	stale := flightGet(t, f, stop, id, a)
	if stale["canShareWithTravelers"] != false || stale["ownSelection"].(map[string]interface{})["rosterConsentRequired"] != true {
		t.Fatal("missing trip consent not explained")
	}
	flightChoose(t, f, stop, id, a, stale, []string{leg}, true, "booked", 409)
	private := flightChoose(t, f, stop, id, a, stale, []string{leg}, false, "booked", 200)
	if private["ownSelection"].(map[string]interface{})["rosterConsentRequired"] != true {
		t.Fatal("saving private selection hid outstanding trip consent")
	}
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/roster-consent", gin.H{"rosterConsentVersion": 2, "shareDates": false}, a, "", 200)
	if current := flightGet(t, f, stop, id, a); current["canShareWithTravelers"] != true {
		t.Fatal("explicit roster consent did not enable sharing")
	}
	flightChoose(t, f, stop, id, a, flightGet(t, f, stop, id, a), []string{leg}, true, "booked", 200)
	f.db.Model(&Participation{}).Where("stop_id = ? AND user_id = ?", stop.ID, a).Update("status", "withdrawn")
	if got := flightGet(t, f, stop, id, b); len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("withdrawn attendee flight leaked")
	}
	flightChoose(t, f, stop, id, a, flightGet(t, f, stop, id, a), []string{leg}, true, "booked", 403)
	if current := flightGet(t, f, stop, id, a); current["canRemoveSelection"] != true {
		t.Fatal("withdrawn attendee cannot remove existing choice")
	}
	// Removing a past choice remains possible to an account that can still browse.
	removed := flightChoose(t, f, stop, id, a, state, nil, false, "removed", 200)
	if removed["ownSelection"] != nil {
		t.Fatal("removed selection remained active")
	}
}

func TestFlightSelectionSurvivesPublishingAndNeedsReviewForSelectedLegChanges(t *testing.T) {
	f, stop, id := flightFixture(t)
	a := flightTraveler(t, f, stop)
	selected := stop.Published.Flights[0].Legs[0].ID
	state := flightGet(t, f, stop, id, a)
	flightChoose(t, f, stop, id, a, state, []string{selected}, true, "booked", 200)
	// Host drafts do not change the published choice or its current review snapshot.
	stop.Draft.Flights[0].Legs[0].DepartureTime = "14:00"
	// Separate deep copies: fixture reads allocate draft and published independently.
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	before := flightGet(t, f, stop, id, a)
	if before["flightVersion"] != state["flightVersion"] {
		t.Fatal("draft changed published selection")
	}
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": stop.Revision}, f.owner, "", 200)
	after := flightGet(t, f, stop, id, a)
	if after["ownSelection"].(map[string]interface{})["needsReconfirmation"] != true || len(after["travelers"].([]interface{})) != 0 {
		t.Fatal("changed selected leg remained current", after)
	}
	flight := after["flight"].(map[string]interface{})
	if flight["legs"].([]interface{})[0].(map[string]interface{})["departureTime"] != "14:00" {
		t.Fatal("review snapshot did not match new version")
	}
	flightChoose(t, f, stop, id, a, state, []string{selected}, true, "booked", 409)
	confirmed := flightChoose(t, f, stop, id, a, after, []string{selected}, true, "booked", 200)
	if confirmed["ownSelection"].(map[string]interface{})["needsReconfirmation"] != false {
		t.Fatal("reconfirmation did not update material versions")
	}
	var latest Stop
	f.db.First(&latest, "id = ?", stop.ID)
	// Changes to an unselected connecting leg do not invalidate the selected leg.
	latest.Draft.Flights[0].Legs[1].DepartureTime = "12:30"
	f.db.Save(&latest)
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": latest.Revision}, f.owner, "", 200)
	if got := flightGet(t, f, stop, id, a); got["ownSelection"].(map[string]interface{})["needsReconfirmation"] != false {
		t.Fatal("unselected leg edit invalidated selection")
	}
	f.db.First(&latest, "id = ?", stop.ID)
	latest.Published.Flights = []TripFlight{}
	f.db.Save(&latest)
	calendarResponse(t, f, "GET", flightPath(stop, id)+"/travelers?stopToken="+stop.ShareToken, nil, a, "", 404)
}

func TestFlightLegacyLegIDsAndVersionsSurviveReadsAndReordering(t *testing.T) {
	f, stop, id := flightFixture(t)
	a := flightTraveler(t, f, stop)
	state := flightGet(t, f, stop, id, a)
	first := stop.Published.Flights[0].Legs[0].ID
	if parsed, err := uuid.Parse(first); err != nil || parsed == uuid.Nil {
		t.Fatal("legacy leg lacked stable id")
	}
	flightChoose(t, f, stop, id, a, state, []string{first}, true, "booked", 200)
	// Reordering old stored entries without IDs must not attach a selection to
	// whichever flight happens to occupy the same index.
	content := *stop.Published
	content.Flights[0].Legs = []TripFlightLeg{content.Flights[0].Legs[1], content.Flights[0].Legs[0]}
	for i := range content.Flights[0].Legs {
		content.Flights[0].Legs[i].ID = ""
	}
	f.db.Model(&Stop{}).Where("id = ?", stop.ID).Update("published", content)
	got := flightGet(t, f, stop, id, a)
	if got["flightVersion"] != state["flightVersion"] || got["ownSelection"].(map[string]interface{})["needsReconfirmation"] != false {
		t.Fatal("leg reorder changed identity", got)
	}
	legs := got["flight"].(map[string]interface{})["legs"].([]interface{})
	if legs[1].(map[string]interface{})["id"] != first {
		t.Fatal("selected stable ID followed index instead of flight")
	}
}

func TestFlightAudienceAndAccessAreRecheckedForTravelerRoster(t *testing.T) {
	f, stop, id := flightFixture(t)
	a := flightTraveler(t, f, stop)
	state := flightGet(t, f, stop, id, a)
	flightChoose(t, f, stop, id, a, state, []string{stop.Published.Flights[0].Legs[0].ID}, true, "booked", 200)
	stop.Published.Flights[0].Visibility = "hosts"
	f.db.Save(&stop)
	calendarResponse(t, f, "GET", flightPath(stop, id)+"/travelers?stopToken="+stop.ShareToken, nil, a, "", 404)
	if got := flightGet(t, f, stop, id, f.owner); len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("selection outlived access to flight")
	}
	stop.Published.Flights[0].Visibility = "viewers"
	stop.ShareToken = newToken()
	f.db.Save(&stop)
	if got := flightGet(t, f, stop, id, f.owner); len(got["travelers"].([]interface{})) != 0 {
		t.Fatal("revoked trip link left traveler in roster")
	}
}

func TestFlightCancelledSelectionCanOnlyBeRemoved(t *testing.T) {
	f, stop, id := flightFixture(t)
	a := flightTraveler(t, f, stop)
	leg := stop.Published.Flights[0].Legs[0].ID
	state := flightGet(t, f, stop, id, a)
	flightChoose(t, f, stop, id, a, state, []string{leg}, true, "booked", 200)
	stop.Published.Flights[0].Status = "cancelled"
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	current := flightGet(t, f, stop, id, a)
	if current["canSelect"] != false || current["canRemoveSelection"] != true || current["canShareWithTravelers"] != false || len(current["travelers"].([]interface{})) != 0 {
		t.Fatal("cancelled selection has incorrect available actions", current)
	}
	outsider := tripAuthorityAccount(t, f)
	for _, viewer := range []uuid.UUID{uuid.Nil, outsider} {
		state := flightGet(t, f, stop, id, viewer)
		if state["requiresAccount"] != false || state["requiresAttendance"] != false || state["canSelect"] != false {
			t.Fatal("cancelled flight invited a viewer to claim or join", state)
		}
	}
	flightChoose(t, f, stop, id, a, current, []string{leg}, true, "booked", 403)
	removed := flightChoose(t, f, stop, id, a, current, nil, false, "removed", 200)
	if removed["ownSelection"] != nil || removed["canRemoveSelection"] != false {
		t.Fatal("cancelled choice was not removed")
	}
	stop.Published.Flights[0].Status = "planned"
	stop.Published.Status = "cancelled"
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	for _, viewer := range []uuid.UUID{uuid.Nil, outsider} {
		state := flightGet(t, f, stop, id, viewer)
		if state["requiresAccount"] != false || state["requiresAttendance"] != false || state["canSelect"] != false {
			t.Fatal("cancelled trip invited a viewer to claim or join", state)
		}
	}
}

type testFlightSearch struct {
	queries      []FlightSearchQuery
	beforeReturn func()
	failed       bool
	bounded      bool
}

func (p *testFlightSearch) Search(ctx context.Context, q FlightSearchQuery) ([]TripFlightLeg, error) {
	p.queries = append(p.queries, q)
	_, p.bounded = ctx.Deadline()
	if p.beforeReturn != nil {
		p.beforeReturn()
	}
	if p.failed {
		return nil, errors.New("provider-secret-response")
	}
	return []TripFlightLeg{{Airline: "ANA", FlightNumber: "NH109", DepartureAirport: "JFK", ArrivalAirport: "HND", DepartureDate: q.DepartureDate}}, nil
}

func TestFlightSearchRequiresHostAndBoundsProviderWork(t *testing.T) {
	f, stop, _ := flightFixture(t)
	path := "/calendar/stops/" + stop.ID.String() + "/flight-search?flightNumber=nh%20109&departureDate=2032-12-22"
	calendarResponse(t, f, "GET", path, nil, uuid.Nil, "", 403)
	attendee := flightTraveler(t, f, stop)
	calendarResponse(t, f, "GET", path, nil, attendee, "", 403)
	unconfigured := calendarResponse(t, f, "GET", path, nil, f.owner, "", 503)
	if unconfigured["configured"] != false {
		t.Fatal("unconfigured search pretended active")
	}
	p := &testFlightSearch{}
	f.service.ConfigureFlightSearchProvider(p)
	result := calendarResponse(t, f, "GET", path, nil, f.owner, "", 200)
	if len(p.queries) != 1 || p.queries[0].FlightNumber != "NH109" || !p.bounded || result["flights"].([]interface{})[0].(map[string]interface{})["id"] == "" {
		t.Fatal("search lost normalization/bounds/leg identity")
	}
	calendarResponse(t, f, "GET", "/calendar/stops/"+stop.ID.String()+"/flight-search?origin=jfk&destination=hnd&departureDate=2032-12-22", nil, f.owner, "", 200)
	calendarResponse(t, f, "GET", "/calendar/stops/"+stop.ID.String()+"/flight-search?flightNumber=../secret&departureDate=2032-12-22", nil, f.owner, "", 400)
	if len(p.queries) != 2 {
		t.Fatal("invalid input reached provider")
	}
	p.failed = true
	failed := calendarResponse(t, f, "GET", path, nil, f.owner, "", 503)
	data, _ := json.Marshal(failed)
	if strings.Contains(string(data), "provider-secret") {
		t.Fatal("provider error leaked")
	}
	p.failed = false
	p.beforeReturn = func() {
		if err := f.db.Model(&Calendar{}).Where("id = ?", stop.CalendarID).Update("owner_id", attendee).Error; err != nil {
			t.Fatal(err)
		}
	}
	calendarResponse(t, f, "GET", path, nil, f.owner, "", 403)
}

func TestFlightSelectionMigrationRollbackAndCascade(t *testing.T) {
	f, stop, id := flightFixture(t)
	state := flightGet(t, f, stop, id, f.owner)
	flightChoose(t, f, stop, id, f.owner, state, []string{stop.Published.Flights[0].Legs[0].ID}, true, "booked", 200)
	if err := f.db.Delete(&Stop{}, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := f.db.Model(&FlightSelection{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatal("deleted trip retained personal flight data", err, count)
	}
	_, source, _, _ := runtime.Caller(0)
	for _, suffix := range []string{"down", "up"} {
		contents, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations/000480_travel_flight_selections."+suffix+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err = f.db.Exec(string(contents)).Error; err != nil {
			t.Fatal(err)
		}
	}
}
