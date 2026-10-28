package travelcalendar

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Logistics belong to the trip's draft/published content. Entry access is never
// a substitute for trip access; the narrower fields are filtered separately.
type TripCost struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Basis    string `json:"basis"`
}
type TripLodgingPrivateDetails struct {
	Visibility          string `json:"visibility"`
	Address             string `json:"address"`
	ArrivalInstructions string `json:"arrivalInstructions"`
}
type TripLodging struct {
	ID                        string                     `json:"id"`
	Name                      string                     `json:"name"`
	Type                      string                     `json:"type"`
	GeneralLocation           string                     `json:"generalLocation"`
	CheckInDate               string                     `json:"checkInDate"`
	CheckOutDate              string                     `json:"checkOutDate"`
	DatesTbd                  bool                       `json:"datesTbd"`
	CheckInTime               string                     `json:"checkInTime"`
	CheckOutTime              string                     `json:"checkOutTime"`
	TimeZone                  string                     `json:"timeZone"`
	Description               string                     `json:"description"`
	URL                       string                     `json:"url"`
	Status                    string                     `json:"status"`
	Kind                      string                     `json:"kind"`
	Visibility                string                     `json:"visibility"`
	Cost                      *TripCost                  `json:"cost,omitempty"`
	RoomsOrBeds               string                     `json:"roomsOrBeds"`
	AssignedUserIDs           []uuid.UUID                `json:"assignedUserIds"`
	PrivateDetails            *TripLodgingPrivateDetails `json:"privateDetails,omitempty"`
	HostNotes                 string                     `json:"hostNotes,omitempty"`
	OutsideTripDatesConfirmed bool                       `json:"outsideTripDatesConfirmed"`
	Property                  *TripLodgingProperty       `json:"property,omitempty"`
	Photos                    []TripLodgingPhoto         `json:"photos"`
}
type TripFlightLeg struct {
	ID                    string `json:"id"`
	Provider              string `json:"provider,omitempty"`
	ProviderID            string `json:"providerId,omitempty"`
	OperatingAirline      string `json:"operatingAirline,omitempty"`
	OperatingFlightNumber string `json:"operatingFlightNumber,omitempty"`
	Airline               string `json:"airline"`
	FlightNumber          string `json:"flightNumber"`
	DepartureAirport      string `json:"departureAirport"`
	ArrivalAirport        string `json:"arrivalAirport"`
	DepartureDate         string `json:"departureDate"`
	DepartureTime         string `json:"departureTime"`
	DepartureTimeZone     string `json:"departureTimeZone"`
	ArrivalDate           string `json:"arrivalDate"`
	ArrivalTime           string `json:"arrivalTime"`
	ArrivalTimeZone       string `json:"arrivalTimeZone"`
}
type TripFlight struct {
	ID                        string          `json:"id"`
	Label                     string          `json:"label"`
	Kind                      string          `json:"kind"`
	Visibility                string          `json:"visibility"`
	Status                    string          `json:"status"`
	Notes                     string          `json:"notes"`
	URL                       string          `json:"url"`
	AssignedUserIDs           []uuid.UUID     `json:"assignedUserIds"`
	HostNotes                 string          `json:"hostNotes,omitempty"`
	OutsideTripDatesConfirmed bool            `json:"outsideTripDatesConfirmed"`
	Legs                      []TripFlightLeg `json:"legs"`
}

func logisticsChanged(old, next *StopContent) bool {
	if old == nil || next == nil {
		return old != next
	}
	a, _ := json.Marshal(struct {
		Lodging []TripLodging
		Flights []TripFlight
	}{old.Lodging, old.Flights})
	b, _ := json.Marshal(struct {
		Lodging []TripLodging
		Flights []TripFlight
	}{next.Lodging, next.Flights})
	return string(a) != string(b)
}
func oneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
func validLogisticsURL(raw string) bool {
	if raw == "" {
		return true
	}
	if len(raw) > 2048 {
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil
}
func validLocalTime(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := time.Parse("15:04", value)
	return err == nil && parsed.Format("15:04") == value
}
func validLogisticsZone(zone string) bool {
	if zone == "" {
		return false
	}
	_, err := time.LoadLocation(zone)
	return err == nil && zone != "Local"
}
func outsideTripDate(c *StopContent, date string) bool {
	return date != "" && c.StartDate != "" && c.EndDate != "" && (date < c.StartDate || date > c.EndDate)
}
func validateEntryIdentity(id *string, seen map[string]bool) error {
	if *id == "" {
		*id = uuid.NewString()
	}
	parsed, err := uuid.Parse(*id)
	if err != nil || parsed == uuid.Nil || seen[parsed.String()] {
		return bad("Each lodging or flight entry needs a distinct valid ID")
	}
	*id = parsed.String()
	seen[*id] = true
	return nil
}
func validateLogisticsAudience(visibility *string, kind *string, assigned []uuid.UUID) error {
	if *visibility == "" {
		*visibility = "attendees"
	}
	if !oneOf(*visibility, "viewers", "attendees", "hosts") {
		return bad("Choose a valid logistics audience")
	}
	if *kind == "" {
		*kind = "group"
	}
	if !oneOf(*kind, "recommendation", "group") {
		return bad("Choose a recommended option or a group plan")
	}
	if len(assigned) > 200 {
		return bad("Too many traveler assignments")
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range assigned {
		if id == uuid.Nil || seen[id] {
			return bad("Traveler assignments must use distinct claimed accounts")
		}
		seen[id] = true
	}
	return nil
}
func validateTripLogistics(c *StopContent) error {
	if len(c.Lodging) > 60 || len(c.Flights) > 40 {
		return bad("A trip supports up to 60 lodging entries and 40 flight itineraries")
	}
	if c.Hero != nil {
		if err := c.Hero.validate(); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for i := range c.Lodging {
		l := &c.Lodging[i]
		if err := validateEntryIdentity(&l.ID, seen); err != nil {
			return err
		}
		if l.Property != nil {
			if l.Property.Provider != "google" || !validLodgingPlaceID(l.Property.PlaceID) {
				return bad("Choose a valid Google property or add lodging manually")
			}
		}
		if len(l.Photos) > 6 {
			return bad("Each lodging supports up to 6 uploaded photos")
		}
		photoIDs := map[uuid.UUID]bool{}
		for j := range l.Photos {
			photo := &l.Photos[j]
			photo.AltText = strings.TrimSpace(photo.AltText)
			if photo.MediaID == uuid.Nil || photoIDs[photo.MediaID] || photo.AltText == "" || utf8.RuneCountInString(photo.AltText) > 500 {
				return bad("Each lodging photo needs a distinct uploaded image and a description up to 500 characters")
			}
			photoIDs[photo.MediaID] = true
		}
		l.Name = strings.TrimSpace(l.Name)
		if l.Name == "" || utf8.RuneCountInString(l.Name) > 200 || len(l.Type) > 100 || len(l.GeneralLocation) > 500 || len(l.Description) > 10000 || len(l.HostNotes) > 10000 || len(l.RoomsOrBeds) > 2000 {
			return bad("Provide a lodging name and keep lodging fields within their limits")
		}
		if err := validateLogisticsAudience(&l.Visibility, &l.Kind, l.AssignedUserIDs); err != nil {
			return err
		}
		if l.Status == "" {
			l.Status = "considering"
		}
		if !oneOf(l.Status, "considering", "selected", "booked", "cancelled") {
			return bad("Choose a valid lodging planning status")
		}
		if !validLogisticsURL(l.URL) {
			return bad("Use an http or https lodging link without embedded credentials")
		}
		if l.TimeZone == "" {
			l.TimeZone = effectiveTimeZone(c.TimeZone)
		}
		if !validLogisticsZone(l.TimeZone) || !validLocalTime(l.CheckInTime) || !validLocalTime(l.CheckOutTime) {
			return bad("Choose valid lodging local times and an IANA time zone")
		}
		if l.DatesTbd {
			l.CheckInDate = ""
			l.CheckOutDate = ""
			l.CheckInTime = ""
			l.CheckOutTime = ""
		} else {
			if !validDate(l.CheckInDate) || !validDate(l.CheckOutDate) || l.CheckOutDate <= l.CheckInDate {
				return bad("Lodging checkout must be after check-in; checkout is exclusive")
			}
			if (outsideTripDate(c, l.CheckInDate) || outsideTripDate(c, l.CheckOutDate)) && !l.OutsideTripDatesConfirmed {
				return bad("Review lodging dates outside the trip, then confirm they are intentional")
			}
		}
		if l.Cost != nil {
			cost := l.Cost
			cost.Currency = strings.ToUpper(strings.TrimSpace(cost.Currency))
			cost.Amount = strings.TrimSpace(cost.Amount)
			validAmount := cost.Amount != "" && len(cost.Amount) <= 15
			dots := 0
			digits := 0
			for _, ch := range cost.Amount {
				if ch == '.' {
					dots++
					if dots > 1 {
						validAmount = false
					}
				} else if ch >= '0' && ch <= '9' {
					digits++
				} else {
					validAmount = false
				}
			}
			if digits == 0 || strings.HasPrefix(cost.Amount, ".") || strings.HasSuffix(cost.Amount, ".") {
				validAmount = false
			}
			currencyOK := len(cost.Currency) == 3
			for _, ch := range cost.Currency {
				if ch < 'A' || ch > 'Z' {
					currencyOK = false
				}
			}
			if !validAmount || !currencyOK || !oneOf(cost.Basis, "night", "person", "total") {
				return bad("Indicative costs need a nonnegative amount, three-letter currency, and night/person/total basis")
			}
		}
		if l.PrivateDetails != nil {
			d := l.PrivateDetails
			if d.Visibility == "" {
				d.Visibility = "attendees"
			}
			if !oneOf(d.Visibility, "attendees", "hosts") || len(d.Address) > 2000 || len(d.ArrivalInstructions) > 10000 {
				return bad("Private arrival details are limited to attendees or hosts")
			}
		}
	}
	for i := range c.Flights {
		f := &c.Flights[i]
		if err := validateEntryIdentity(&f.ID, seen); err != nil {
			return err
		}
		ensureItineraryLegIDs(f)
		legIDs := map[string]bool{}
		f.Label = strings.TrimSpace(f.Label)
		if f.Label == "" || utf8.RuneCountInString(f.Label) > 200 || len(f.Notes) > 10000 || len(f.HostNotes) > 10000 || len(f.Legs) == 0 || len(f.Legs) > 10 {
			return bad("Provide a flight label and between one and ten connecting legs")
		}
		if err := validateLogisticsAudience(&f.Visibility, &f.Kind, f.AssignedUserIDs); err != nil {
			return err
		}
		if f.Status == "" {
			f.Status = "suggested"
		}
		if !oneOf(f.Status, "suggested", "planned", "booked", "cancelled") {
			return bad("Choose a valid flight planning status")
		}
		if !validLogisticsURL(f.URL) {
			return bad("Use an http or https flight link without embedded credentials")
		}
		var previousArrival *time.Time
		outside := false
		for j := range f.Legs {
			leg := &f.Legs[j]
			if err := validateEntryIdentity(&leg.ID, legIDs); err != nil {
				return bad("Each flight leg needs a distinct valid ID")
			}
			if len(leg.Provider) > 50 || len(leg.ProviderID) > 300 || len(leg.OperatingAirline) > 150 || len(leg.OperatingFlightNumber) > 30 {
				return bad("Keep flight provider and operating carrier details within their limits")
			}
			if len(leg.Airline) > 150 || len(leg.FlightNumber) > 30 || len(leg.DepartureAirport) > 200 || len(leg.ArrivalAirport) > 200 || strings.TrimSpace(leg.DepartureAirport) == "" || strings.TrimSpace(leg.ArrivalAirport) == "" {
				return bad("Each flight leg needs departure and arrival airports")
			}
			departure, err := flightInstant(leg.DepartureDate, leg.DepartureTime, leg.DepartureTimeZone)
			if err != nil {
				return err
			}
			arrival, err := flightInstant(leg.ArrivalDate, leg.ArrivalTime, leg.ArrivalTimeZone)
			if err != nil {
				return err
			}
			if departure != nil && arrival != nil && arrival.Before(*departure) {
				return bad("Flight arrival must follow departure after accounting for airport time zones")
			}
			if previousArrival != nil && departure != nil && departure.Before(*previousArrival) {
				return bad("A connecting flight cannot depart before the previous leg arrives")
			}
			previousArrival = arrival
			outside = outside || outsideTripDate(c, leg.DepartureDate) || outsideTripDate(c, leg.ArrivalDate)
		}
		if outside && !f.OutsideTripDatesConfirmed {
			return bad("Review flight dates outside the trip, then confirm they are intentional")
		}
	}
	return nil
}

// Parse the local wall clock in its actual airport zone; UTC lexical comparison
// would reject valid westbound routes and miss impossible connections.
func flightInstant(date, clock, zone string) (*time.Time, error) {
	if date != "" && !validDate(date) {
		return nil, bad("Flight dates must use YYYY-MM-DD")
	}
	if !validLocalTime(clock) {
		return nil, bad("Flight local times must use HH:MM")
	}
	if zone != "" && !validLogisticsZone(zone) {
		return nil, bad("Choose valid IANA airport time zones")
	}
	if clock == "" {
		return nil, nil
	}
	if date == "" || zone == "" {
		return nil, bad("A flight time needs its local date and airport time zone")
	}
	location, _ := time.LoadLocation(zone)
	raw := date + " " + clock
	instant, err := time.ParseInLocation("2006-01-02 15:04", raw, location)
	if err != nil || instant.In(location).Format("2006-01-02 15:04") != raw {
		return nil, bad("This flight local time does not exist in its time zone")
	}
	return &instant, nil
}
func canReadLogistics(audience string, manager, attendee bool) bool {
	return manager || audience == "viewers" || (audience != "hosts" && attendee)
}
func logisticsAssignments(ids []uuid.UUID, viewer uuid.UUID, manager bool) []uuid.UUID {
	visible := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if manager || viewer != uuid.Nil && viewer == id {
			visible = append(visible, id)
		}
	}
	return visible
}
func filteredTripLogistics(content *StopContent, viewer uuid.UUID, manager, attendee bool) gin.H {
	lodging := make([]TripLodging, 0)
	flights := make([]TripFlight, 0)
	for _, entry := range content.Lodging {
		if !canReadLogistics(entry.Visibility, manager, attendee) {
			continue
		}
		assignmentViewer := viewer
		if !attendee && !manager {
			assignmentViewer = uuid.Nil
		}
		entry.AssignedUserIDs = logisticsAssignments(entry.AssignedUserIDs, assignmentViewer, manager)
		if !manager {
			entry.HostNotes = ""
			if entry.PrivateDetails != nil && !canReadLogistics(entry.PrivateDetails.Visibility, manager, attendee) {
				entry.PrivateDetails = nil
			}
		}
		lodging = append(lodging, entry)
	}
	for _, entry := range content.Flights {
		if !canReadLogistics(entry.Visibility, manager, attendee) {
			continue
		}
		assignmentViewer := viewer
		if !attendee && !manager {
			assignmentViewer = uuid.Nil
		}
		entry.AssignedUserIDs = logisticsAssignments(entry.AssignedUserIDs, assignmentViewer, manager)
		if !manager {
			entry.HostNotes = ""
		}
		flights = append(flights, entry)
	}
	return gin.H{"lodging": lodging, "flights": flights}
}
func (s *Service) tripLogisticsDTO(stop *Stop, content *StopContent, v Viewer, manager bool) gin.H {
	dto := filteredTripLogistics(content, v.UserID, manager, s.tripAttendee(s.db, stop, v))
	if content.Hero != nil {
		hero := *content.Hero
		hero.URL = tripHeroURL(stop.ID, hero.MediaID)
		dto["hero"] = hero
	}
	return dto
}

func (s *Service) validateTripContentReferences(tx *gorm.DB, stop *Stop, content *StopContent) error {
	// Published entry permissions have their own immediate mutation. A stale
	// draft must never silently undo a restriction when later published.
	if err := checkPublishedLogisticsPermissions(stop.Published, content); err != nil {
		return err
	}
	if content.Hero != nil {
		var media TripMedia
		if tx.Select("id, stop_id, width, height").Where("id = ? AND stop_id = ? AND lodging_id IS NULL", content.Hero.MediaID, stop.ID).First(&media).Error != nil {
			return bad("Upload a hero image for this trip before attaching it")
		}
		content.Hero.Width = media.Width
		content.Hero.Height = media.Height
		content.Hero.URL = tripHeroURL(stop.ID, content.Hero.MediaID)
	}
	for i := range content.Lodging {
		entry := &content.Lodging[i]
		for j := range entry.Photos {
			photo := &entry.Photos[j]
			var media TripMedia
			if tx.Select("id, width, height").Where("id = ? AND stop_id = ? AND lodging_id = ?", photo.MediaID, stop.ID, entry.ID).First(&media).Error != nil {
				return bad("Upload lodging photos to this trip and lodging before attaching them")
			}
			photo.Width, photo.Height = media.Width, media.Height
			photo.URL = lodgingPhotoURL(stop.ID, entry.ID, photo.MediaID)
		}
	}
	// Retain existing historical assignments for hosts after a traveler leaves.
	// They must not prevent unrelated content edits, but cannot be copied into
	// another entry or introduced as a new assignment while ineligible.
	history := map[string]map[uuid.UUID]bool{}
	remember := func(previous *StopContent) {
		if previous == nil {
			return
		}
		for _, entry := range previous.Lodging {
			key := "lodging:" + entry.ID
			if history[key] == nil {
				history[key] = map[uuid.UUID]bool{}
			}
			for _, id := range entry.AssignedUserIDs {
				history[key][id] = true
			}
		}
		for _, entry := range previous.Flights {
			key := "flight:" + entry.ID
			if history[key] == nil {
				history[key] = map[uuid.UUID]bool{}
			}
			for _, id := range entry.AssignedUserIDs {
				history[key][id] = true
			}
		}
	}
	if !stop.CreatedAt.IsZero() {
		remember(&stop.Draft)
		remember(stop.Published)
	}
	eligible := map[uuid.UUID]bool{}
	checked := map[uuid.UUID]bool{}
	checkAssignment := func(key string, id uuid.UUID) error {
		if checked[id] {
			if eligible[id] || history[key][id] {
				return nil
			}
			return bad("Assign only confirmed travelers with current trip access")
		}
		checked[id] = true
		var participation Participation
		if tx.Where("stop_id = ? AND user_id = ?", stop.ID, id).First(&participation).Error == nil {
			v := s.storedParticipationViewer(tx, stop, &participation)
			eligible[id] = s.tripAttendee(tx, stop, v)
		}
		if !eligible[id] && !history[key][id] {
			return bad("Assign only confirmed travelers with current trip access")
		}
		return nil
	}
	for _, entry := range content.Lodging {
		for _, id := range entry.AssignedUserIDs {
			if err := checkAssignment("lodging:"+entry.ID, id); err != nil {
				return err
			}
		}
	}
	for _, entry := range content.Flights {
		for _, id := range entry.AssignedUserIDs {
			if err := checkAssignment("flight:"+entry.ID, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkPublishedLogisticsPermissions(published, draft *StopContent) error {
	if published == nil || draft == nil {
		return nil
	}
	audience := func(value string) string {
		if value == "" {
			return "attendees"
		}
		return value
	}
	for _, next := range draft.Lodging {
		for _, prior := range published.Lodging {
			if next.ID != prior.ID {
				continue
			}
			if audience(next.Visibility) != audience(prior.Visibility) || (next.PrivateDetails != nil && prior.PrivateDetails != nil && audience(next.PrivateDetails.Visibility) != audience(prior.PrivateDetails.Visibility)) {
				return bad("Change published lodging permissions using the immediate audience control")
			}
		}
	}
	for _, next := range draft.Flights {
		for _, prior := range published.Flights {
			if next.ID == prior.ID && audience(next.Visibility) != audience(prior.Visibility) {
				return bad("Change published flight permissions using the immediate audience control")
			}
		}
	}
	return nil
}
func (s *Service) updateTripLogisticsPermissions(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, errUnavailable)
		return
	}
	var body struct {
		Revision          *int   `json:"revision"`
		Kind              string `json:"kind"`
		ID                string `json:"id"`
		Visibility        string `json:"visibility"`
		PrivateVisibility string `json:"privateVisibility"`
	}
	if c.ShouldBindJSON(&body) != nil || !oneOf(body.Kind, "lodging", "flight") || !oneOf(body.Visibility, "viewers", "attendees", "hosts") || body.PrivateVisibility != "" && !oneOf(body.PrivateVisibility, "attendees", "hosts") {
		respondError(c, bad("Choose a valid entry and audience"))
		return
	}
	var stop Stop
	err = s.db.Transaction(func(tx *gorm.DB) error {
		locked, err := s.lockedStop(tx, id.String())
		if err != nil {
			return err
		}
		stop = *locked
		if !s.canManagePermissions(tx, &stop, Viewer{UserID: user.ID}) {
			return errForbidden
		}
		if err := s.checkTripRevision(&stop, body.Revision); err != nil {
			return err
		}
		found := false
		update := func(content *StopContent) {
			if content == nil {
				return
			}
			if body.Kind == "lodging" {
				for i := range content.Lodging {
					entry := &content.Lodging[i]
					if entry.ID == body.ID {
						found = true
						entry.Visibility = body.Visibility
						if entry.PrivateDetails != nil && body.PrivateVisibility != "" {
							entry.PrivateDetails.Visibility = body.PrivateVisibility
						}
					}
				}
			} else {
				for i := range content.Flights {
					if content.Flights[i].ID == body.ID {
						found = true
						content.Flights[i].Visibility = body.Visibility
					}
				}
			}
		}
		update(&stop.Draft)
		update(stop.Published)
		if !found {
			return errUnavailable
		}
		return s.saveTripChange(tx, &stop, user.ID, "logistics_permissions_changed", gin.H{"kind": body.Kind, "entryId": body.ID, "visibility": body.Visibility, "privateVisibility": body.PrivateVisibility})
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"stop": s.stopDTO(&stop, Viewer{UserID: user.ID})})
}
