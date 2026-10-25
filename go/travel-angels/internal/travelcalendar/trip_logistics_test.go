package travelcalendar

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func tripLogisticsContent() StopContent {
	return StopContent{Title: "Portugal", Destination: "Lisbon", TimeZone: "Europe/Lisbon", StartDate: "2032-12-20", EndDate: "2033-01-10", DatePrecision: "fixed", Status: "confirmed", JoinOpen: true}
}
func TestTripLogisticsDefaultsAndValidation(t *testing.T) {
	content := tripLogisticsContent()
	content.Lodging = []TripLodging{{Name: "Group house", DatesTbd: true, Cost: &TripCost{Amount: "150.50", Currency: "eur", Basis: "night"}}}
	if err := validateTripLogistics(&content); err != nil {
		t.Fatal(err)
	}
	lodging := content.Lodging[0]
	if lodging.Visibility != "attendees" || lodging.Kind != "group" || lodging.Status != "considering" || lodging.TimeZone != "Europe/Lisbon" || lodging.Cost.Currency != "EUR" || lodging.ID == "" {
		t.Fatalf("unexpected defaults: %+v", lodging)
	}
	for name, edit := range map[string]func(*TripLodging){
		"unsafe link":          func(l *TripLodging) { l.URL = "javascript:alert(1)" },
		"embedded credentials": func(l *TripLodging) { l.URL = "https://private:secret@example.com" },
		"same-day checkout":    func(l *TripLodging) { l.DatesTbd = false; l.CheckInDate = "2032-12-21"; l.CheckOutDate = "2032-12-21" },
		"outside unreviewed":   func(l *TripLodging) { l.DatesTbd = false; l.CheckInDate = "2032-12-19"; l.CheckOutDate = "2032-12-21" },
		"private fields to guests": func(l *TripLodging) {
			l.PrivateDetails = &TripLodgingPrivateDetails{Visibility: "viewers", Address: "Secret address"}
		},
		"negative cost": func(l *TripLodging) { l.Cost = &TripCost{Amount: "-10", Currency: "EUR", Basis: "total"} },
	} {
		t.Run(name, func(t *testing.T) {
			copy := tripLogisticsContent()
			copy.Lodging = []TripLodging{lodging}
			edit(&copy.Lodging[0])
			if validateTripLogistics(&copy) == nil {
				t.Fatal("invalid entry accepted")
			}
		})
	}
	content.Lodging[0].DatesTbd = false
	content.Lodging[0].CheckInDate = "2032-12-19"
	content.Lodging[0].CheckOutDate = "2032-12-21"
	content.Lodging[0].OutsideTripDatesConfirmed = true
	if err := validateTripLogistics(&content); err != nil {
		t.Fatalf("intentional outside dates: %v", err)
	}
}
func TestTripLogisticsFlightChronologyUsesAirportZones(t *testing.T) {
	content := tripLogisticsContent()
	content.Flights = []TripFlight{{Label: "Westbound", Legs: []TripFlightLeg{{DepartureAirport: "LHR", ArrivalAirport: "JFK", DepartureDate: "2032-12-22", DepartureTime: "12:00", DepartureTimeZone: "Europe/London", ArrivalDate: "2032-12-22", ArrivalTime: "11:00", ArrivalTimeZone: "America/New_York"}}}}
	if err := validateTripLogistics(&content); err != nil {
		t.Fatalf("earlier local arrival is a later instant: %v", err)
	}
	content.Flights[0].Legs[0].ArrivalTime = "05:00"
	if validateTripLogistics(&content) == nil {
		t.Fatal("arrival before departure accepted")
	}
	content.Flights[0].Legs[0].ArrivalTime = "11:00"
	content.Flights[0].Legs = append(content.Flights[0].Legs, TripFlightLeg{DepartureAirport: "JFK", ArrivalAirport: "BOS", DepartureDate: "2032-12-22", DepartureTime: "10:00", DepartureTimeZone: "America/New_York", ArrivalDate: "2032-12-22", ArrivalTime: "12:00", ArrivalTimeZone: "America/New_York"})
	if validateTripLogistics(&content) == nil {
		t.Fatal("connection departed before arrival")
	}
	if _, err := flightInstant("2026-03-08", "02:30", "America/New_York"); err == nil {
		t.Fatal("nonexistent DST wall time accepted")
	}
	uncertain := tripLogisticsContent()
	uncertain.Flights = []TripFlight{{Label: "Suggested direct route", Legs: []TripFlightLeg{{DepartureAirport: "JFK", ArrivalAirport: "LIS"}}}}
	if err := validateTripLogistics(&uncertain); err != nil {
		t.Fatalf("suggestion without flight/time: %v", err)
	}
}
func TestTripLogisticsFilteringNeverExposesPrivateFields(t *testing.T) {
	self, other := uuid.New(), uuid.New()
	content := tripLogisticsContent()
	content.Lodging = []TripLodging{{ID: uuid.NewString(), Name: "Recommended hotel", Visibility: "viewers", AssignedUserIDs: []uuid.UUID{self, other}, HostNotes: "SECRET HOST", PrivateDetails: &TripLodgingPrivateDetails{Visibility: "attendees", Address: "SECRET ADDRESS"}}, {ID: uuid.NewString(), Name: "SECRET HOUSE", Visibility: "hosts"}, {ID: uuid.NewString(), Name: "Shared group stay", Visibility: "attendees"}}
	content.Flights = []TripFlight{{ID: uuid.NewString(), Label: "Public suggestion", Visibility: "viewers", AssignedUserIDs: []uuid.UUID{self, other}, HostNotes: "SECRET FLIGHT HOST"}, {ID: uuid.NewString(), Label: "SECRET FLIGHT", Visibility: "attendees"}}
	guest := filteredTripLogistics(&content, uuid.Nil, false, false)
	data, _ := json.Marshal(guest)
	if strings.Contains(string(data), "SECRET") || strings.Contains(string(data), self.String()) || strings.Contains(string(data), other.String()) {
		t.Fatalf("guest private data leaked: %s", data)
	}
	if len(guest["lodging"].([]TripLodging)) != 1 || len(guest["flights"].([]TripFlight)) != 1 {
		t.Fatal("guest saw hidden entry count")
	}
	attendee := filteredTripLogistics(&content, self, false, true)
	data, _ = json.Marshal(attendee)
	if strings.Contains(string(data), "SECRET HOST") || strings.Contains(string(data), "SECRET FLIGHT HOST") || strings.Contains(string(data), "SECRET HOUSE") || strings.Contains(string(data), other.String()) {
		t.Fatalf("attendee received private host/other-traveler data: %s", data)
	}
	if !strings.Contains(string(data), "SECRET ADDRESS") || !strings.Contains(string(data), self.String()) {
		t.Fatal("attendee lost permitted fields")
	}
	manager := filteredTripLogistics(&content, self, true, true)
	data, _ = json.Marshal(manager)
	if !strings.Contains(string(data), "SECRET HOST") || !strings.Contains(string(data), other.String()) {
		t.Fatal("host cannot manage assigned travelers")
	}
	if content.Lodging[0].HostNotes != "SECRET HOST" || len(content.Lodging[0].AssignedUserIDs) != 2 {
		t.Fatal("filter mutated persisted content")
	}
	hidden := filteredTripLogistics(&StopContent{Lodging: []TripLodging{{Visibility: "hosts"}}}, uuid.Nil, false, false)
	empty := filteredTripLogistics(&StopContent{}, uuid.Nil, false, false)
	h, _ := json.Marshal(hidden)
	e, _ := json.Marshal(empty)
	if string(h) != string(e) {
		t.Fatal("hidden entries are distinguishable from an empty section")
	}
}
func TestTripLogisticsPublishedPermissionsRequireImmediateControl(t *testing.T) {
	id := uuid.NewString()
	prior := StopContent{Lodging: []TripLodging{{ID: id, Visibility: "attendees", PrivateDetails: &TripLodgingPrivateDetails{Visibility: "hosts"}}}}
	next := StopContent{Lodging: []TripLodging{{ID: id, Visibility: "viewers", PrivateDetails: &TripLodgingPrivateDetails{Visibility: "hosts"}}}}
	if checkPublishedLogisticsPermissions(&prior, &next) == nil {
		t.Fatal("draft deferred a published permission change")
	}
	next.Lodging[0].Visibility = "attendees"
	next.Lodging[0].PrivateDetails.Visibility = "attendees"
	if checkPublishedLogisticsPermissions(&prior, &next) == nil {
		t.Fatal("draft broadened private fields")
	}
}
func TestTripLogisticsHTTPPrivacyAndImmediateRestriction(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	attendee, other := uuid.New(), uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?), (?)", attendee, other).Error; err != nil {
		t.Fatal(err)
	}
	participant := Participation{ID: uuid.New(), StopID: stop.ID, UserID: attendee, Status: "joined", StopLinkTokenHash: guestDigest(stop.ShareToken)}
	if err := f.db.Create(&participant).Error; err != nil {
		t.Fatal(err)
	}
	entryID := uuid.NewString()
	content := *stop.Published
	content.Lodging = []TripLodging{{ID: entryID, Name: "Public hotel", Visibility: "viewers", Kind: "recommendation", DatesTbd: true, AssignedUserIDs: []uuid.UUID{attendee, other}, HostNotes: "DO NOT EXPOSE", PrivateDetails: &TripLodgingPrivateDetails{Visibility: "attendees", Address: "PRIVATE ADDRESS"}}, {ID: uuid.NewString(), Name: "HIDDEN GROUP HOUSE", Visibility: "attendees", DatesTbd: true}}
	stop.Draft = content
	stop.Published = &content
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	path := "/calendar/stops/" + stop.ID.String() + "?stopToken=" + stop.ShareToken
	guest := calendarResponse(t, f, "GET", path, nil, uuid.Nil, "", 200)["stop"].(map[string]interface{})
	data, _ := json.Marshal(guest)
	if strings.Contains(string(data), "DO NOT EXPOSE") || strings.Contains(string(data), "PRIVATE ADDRESS") || strings.Contains(string(data), "HIDDEN GROUP HOUSE") {
		t.Fatalf("guest leaked private fields: %s", data)
	}
	joined := calendarResponse(t, f, "GET", path, nil, attendee, "", 200)["stop"].(map[string]interface{})
	data, _ = json.Marshal(joined)
	if !strings.Contains(string(data), "PRIVATE ADDRESS") || strings.Contains(string(data), other.String()) {
		t.Fatalf("attendee filtering failed: %s", data)
	}
	if err := f.db.Model(&participant).Updates(map[string]interface{}{"status": "requested", "reserved": true}).Error; err != nil {
		t.Fatal(err)
	}
	reserved := calendarResponse(t, f, "GET", path, nil, attendee, "", 200)["stop"].(map[string]interface{})
	data, _ = json.Marshal(reserved)
	if !strings.Contains(string(data), "PRIVATE ADDRESS") {
		t.Fatal("reserved reconfirming traveler lost attendee logistics")
	}
	calendarResponse(t, f, "PUT", "/calendar/stops/"+stop.ID.String()+"/logistics/permissions", gin.H{"revision": stop.Revision, "kind": "lodging", "id": entryID, "visibility": "hosts", "privateVisibility": "hosts"}, f.owner, "", 200)
	restricted := calendarResponse(t, f, "GET", path, nil, uuid.Nil, "", 200)["stop"].(map[string]interface{})
	if len(restricted["lodging"].([]interface{})) != 0 {
		t.Fatal("published restriction waited for draft publication")
	}
	if err := f.db.Model(&participant).Updates(map[string]interface{}{"status": "removed", "reserved": false}).Error; err != nil {
		t.Fatal(err)
	}
	removed := calendarResponse(t, f, "GET", path, nil, attendee, "", 200)["stop"].(map[string]interface{})
	if len(removed["lodging"].([]interface{})) != 0 {
		t.Fatal("removed attendee retained group-only entries")
	}
}

func TestTripLogisticsHistoricalAssignmentsDoNotBlockEdits(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	former := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", former).Error; err != nil {
		t.Fatal(err)
	}
	participant := Participation{ID: uuid.New(), StopID: stop.ID, UserID: former, Status: "removed"}
	if err := f.db.Create(&participant).Error; err != nil {
		t.Fatal(err)
	}
	stop.Draft.Lodging = []TripLodging{{ID: uuid.NewString(), Name: "Historical group stay", Visibility: "viewers", DatesTbd: true, AssignedUserIDs: []uuid.UUID{former}}}
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	next := stop.Draft
	next.Title = "Unrelated title correction"
	if err := f.service.validateTripContentReferences(f.db, &stop, &next); err != nil {
		t.Fatalf("historical assignment blocked unrelated edit: %v", err)
	}
	// Historical associations stay in the host management record only.
	filtered := filteredTripLogistics(&next, former, false, false)
	if len(filtered["lodging"].([]TripLodging)[0].AssignedUserIDs) != 0 {
		t.Fatal("former traveler received a stale assignment as active")
	}
	next.Lodging = append([]TripLodging{}, next.Lodging...)
	next.Lodging[0].ID = uuid.NewString()
	if err := f.service.validateTripContentReferences(f.db, &stop, &next); err == nil {
		t.Fatal("historical ineligible person was assigned to a new stay")
	}
}
