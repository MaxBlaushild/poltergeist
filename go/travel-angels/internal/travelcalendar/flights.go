package travelcalendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FlightSearchQuery struct {
	FlightNumber, DepartureDate, Origin, Destination string
}

type FlightSearchProvider interface {
	Search(context.Context, FlightSearchQuery) ([]TripFlightLeg, error)
}

// Adapters may expose known, static recovery guidance without exposing raw
// provider response bodies or credentials.
type SafeFlightSearchError interface {
	error
	PublicMessage() string
}

func (s *Service) ConfigureFlightSearchProvider(provider FlightSearchProvider) {
	s.flightSearchProvider = provider
}

// Personal choices deliberately live outside host-editable draft/published JSON.
type FlightSelection struct {
	StopID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	FlightID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	Status               string
	LegIDs               []string          `gorm:"serializer:json;type:jsonb"`
	LegVersions          map[string]string `gorm:"serializer:json;type:jsonb"`
	ShareWithTravelers   bool
	RosterConsentVersion int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (FlightSelection) TableName() string { return "tc_flight_selections" }

func materialFlightLeg(leg TripFlightLeg) TripFlightLeg {
	// Provider lookup IDs and source annotations are not itinerary changes.
	leg.Provider, leg.ProviderID, leg.ID = "", "", ""
	return leg
}

func ensureItineraryLegIDs(flight *TripFlight) {
	occurrences := map[string]int{}
	for i := range flight.Legs {
		leg := &flight.Legs[i]
		if leg.ID != "" {
			continue
		}
		encoded, _ := json.Marshal(materialFlightLeg(*leg))
		identity := flight.ID + ":" + string(encoded)
		occurrences[identity]++
		leg.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity+":"+strconv.Itoa(occurrences[identity]))).String()
	}
}

func ensureFlightLegIDs(content *StopContent) {
	for i := range content.Flights {
		ensureItineraryLegIDs(&content.Flights[i])
	}
}

func legVersion(flight *TripFlight, leg TripFlightLeg) string {
	data, _ := json.Marshal(struct {
		Cancelled bool
		Leg       TripFlightLeg
	}{flight.Status == "cancelled", materialFlightLeg(leg)})
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func flightVersion(flight *TripFlight) string {
	versions := make([]string, 0, len(flight.Legs))
	for _, leg := range flight.Legs {
		versions = append(versions, leg.ID+":"+legVersion(flight, leg))
	}
	sort.Strings(versions)
	digest := sha256.Sum256([]byte(flight.ID + ":" + strings.Join(versions, ";")))
	return hex.EncodeToString(digest[:])
}

func selectionNeedsReconfirmation(selection *FlightSelection, flight *TripFlight) bool {
	if selection.Status == "removed" {
		return false
	}
	current := map[string]string{}
	for _, leg := range flight.Legs {
		current[leg.ID] = legVersion(flight, leg)
	}
	if len(selection.LegIDs) == 0 {
		return true
	}
	for _, id := range selection.LegIDs {
		if current[id] == "" || current[id] != selection.LegVersions[id] {
			return true
		}
	}
	return false
}

func (s *Service) registerFlightRoutes(g *gin.RouterGroup) {
	g.GET("/calendar/stops/:id/flight-search", s.searchFlights)
	g.GET("/calendar/stops/:id/flights/:flightId/travelers", s.flightTravelers)
	g.POST("/calendar/stops/:id/flights/:flightId/selection", s.selectFlight)
}

var flightSearchNumber = regexp.MustCompile(`^[A-Z0-9]{2,10}$`)
var flightSearchAirport = regexp.MustCompile(`^[A-Z]{3,4}$`)

func (s *Service) flightSearchAccess(c *gin.Context) error {
	viewer := s.viewer(c)
	return s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, err := s.lockedStop(tx, c.Param("id"))
		if err != nil {
			return err
		}
		if !s.canManageStop(tx, stop, viewer) {
			return errForbidden
		}
		return nil
	})
}

func (s *Service) searchFlights(c *gin.Context) {
	if err := s.flightSearchAccess(c); err != nil {
		respondError(c, err)
		return
	}
	query := FlightSearchQuery{
		FlightNumber:  strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(c.Query("flightNumber")), " ", "")),
		DepartureDate: c.Query("departureDate"),
		Origin:        strings.ToUpper(strings.TrimSpace(c.Query("origin"))),
		Destination:   strings.ToUpper(strings.TrimSpace(c.Query("destination"))),
	}
	validRoute := flightSearchAirport.MatchString(query.Origin) && flightSearchAirport.MatchString(query.Destination) && query.Origin != query.Destination
	if !validDate(query.DepartureDate) || (query.FlightNumber == "" && !validRoute) || (query.FlightNumber != "" && !flightSearchNumber.MatchString(query.FlightNumber)) || (query.Origin != "" && !flightSearchAirport.MatchString(query.Origin)) || (query.Destination != "" && !flightSearchAirport.MatchString(query.Destination)) {
		respondError(c, bad("Enter a departure date and a flight number or departure and arrival airport codes"))
		return
	}
	if s.flightSearchProvider == nil {
		c.JSON(503, gin.H{"configured": false, "flights": []TripFlightLeg{}, "error": "Flight lookup is not configured. You can enter flight details manually."})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	flights, err := s.flightSearchProvider.Search(ctx, query)
	cancel()
	if accessErr := s.flightSearchAccess(c); accessErr != nil {
		respondError(c, accessErr)
		return
	}
	if err != nil {
		message := "Flight lookup is temporarily unavailable. You can enter flight details manually."
		var public SafeFlightSearchError
		if errors.As(err, &public) && public.PublicMessage() != "" {
			message = public.PublicMessage()
		}
		c.JSON(503, gin.H{"configured": true, "flights": []TripFlightLeg{}, "error": message})
		return
	}
	if len(flights) > 20 {
		flights = flights[:20]
	}
	if flights == nil {
		flights = []TripFlightLeg{}
	}
	result := TripFlight{ID: "search", Legs: flights}
	ensureItineraryLegIDs(&result)
	c.JSON(200, gin.H{"configured": true, "flights": result.Legs})
}

func (s *Service) publishedFlight(tx *gorm.DB, c *gin.Context, viewer Viewer) (*Stop, *TripFlight, Viewer, error) {
	stop, err := s.lockedStop(tx, c.Param("id"))
	if err != nil {
		return nil, nil, viewer, err
	}
	viewer = s.currentTripMediaViewer(tx, stop, c, viewer)
	if !s.canViewStop(tx, stop, viewer) || stop.Published == nil {
		return nil, nil, viewer, errUnavailable
	}
	manager := s.canManageStop(tx, stop, viewer)
	for i := range stop.Published.Flights {
		flight := &stop.Published.Flights[i]
		if flight.ID == c.Param("flightId") && canReadLogistics(flight.Visibility, manager, s.tripAttendee(tx, stop, viewer)) {
			return stop, flight, viewer, nil
		}
	}
	return nil, nil, viewer, errUnavailable
}

func (s *Service) flightEligible(tx *gorm.DB, stop *Stop, viewer Viewer) (bool, bool, *Participation) {
	if viewer.UserID == uuid.Nil {
		return false, false, nil
	}
	manager := s.canManageStop(tx, stop, viewer)
	var participation Participation
	if tx.Where("stop_id = ? AND user_id = ?", stop.ID, viewer.UserID).First(&participation).Error != nil {
		return manager, manager, nil
	}
	joined := participation.Status == "joined" && s.tripAttendee(tx, stop, viewer)
	return manager || joined, manager, &participation
}

func (s *Service) flightTravelersDTO(tx *gorm.DB, stop *Stop, flight *TripFlight, viewer Viewer) (gin.H, error) {
	eligible, manager, participation := s.flightEligible(tx, stop, viewer)
	selectionOpen := flight.Status != "cancelled" && stop.Published.Status != "cancelled"
	canSelect := eligible && selectionOpen
	canView := manager || (eligible && stop.RosterVisibility == "attendees")
	tripConsentRequired := !manager && participation != nil && participation.RosterConsentVersion != stop.RosterVersion
	review := *flight
	review.HostNotes = ""
	review.AssignedUserIDs = []uuid.UUID{}
	result := gin.H{"flightVersion": flightVersion(flight), "rosterVersion": stop.RosterVersion,
		"flight":    review,
		"canSelect": canSelect, "requiresAccount": selectionOpen && viewer.UserID == uuid.Nil,
		"canRemoveSelection": false, "sharingAudience": stop.RosterVisibility,
		"canShareWithTravelers": canSelect && !tripConsentRequired,
		"requiresAttendance":    selectionOpen && viewer.UserID != uuid.Nil && !eligible,
		"canViewTravelers":      canView, "ownSelection": nil, "travelers": []gin.H{}}
	var selections []FlightSelection
	if err := tx.Where("stop_id = ? AND flight_id = ?", stop.ID, flight.ID).Order("created_at, user_id").Find(&selections).Error; err != nil {
		return nil, err
	}
	travelers := []gin.H{}
	for _, selection := range selections {
		stale := selectionNeedsReconfirmation(&selection, flight)
		if selection.UserID == viewer.UserID && selection.Status != "removed" {
			result["canRemoveSelection"] = true
			result["ownSelection"] = gin.H{"status": selection.Status, "legIds": selection.LegIDs, "needsReconfirmation": stale,
				"shareWithTravelers": selection.ShareWithTravelers, "rosterConsentRequired": tripConsentRequired || selection.RosterConsentVersion != stop.RosterVersion}
		}
		if !canView || stale || selection.Status == "removed" || !selection.ShareWithTravelers || selection.RosterConsentVersion != stop.RosterVersion || flight.Status == "cancelled" || stop.Published.Status == "cancelled" {
			continue
		}
		subject := s.participationViewer(tx, stop, selection.UserID, participationInput{})
		subjectEligible, subjectManager, participation := s.flightEligible(tx, stop, subject)
		if !subjectEligible || !s.canViewStop(tx, stop, subject) || !canReadLogistics(flight.Visibility, subjectManager, true) {
			continue
		}
		if !subjectManager && (participation == nil || participation.RosterConsentVersion != stop.RosterVersion) {
			continue
		}
		row := s.tripSafeIdentity(tx, selection.UserID)
		row["status"], row["legIds"], row["isHost"] = selection.Status, selection.LegIDs, subjectManager
		travelers = append(travelers, row)
	}
	result["travelers"] = travelers
	return result, nil
}

func (s *Service) flightTravelers(c *gin.Context) {
	viewer := s.viewer(c)
	var result gin.H
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, flight, current, err := s.publishedFlight(tx, c, viewer)
		if err != nil {
			return err
		}
		result, err = s.flightTravelersDTO(tx, stop, flight, current)
		return err
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, result)
}

func (s *Service) selectFlight(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Status             string   `json:"status"`
		LegIDs             []string `json:"legIds"`
		FlightVersion      string   `json:"flightVersion"`
		ShareWithTravelers bool     `json:"shareWithTravelers"`
		RosterVersion      int      `json:"rosterVersion"`
	}
	if c.ShouldBindJSON(&body) != nil || !oneOf(body.Status, "considering", "booked", "removed") || len(body.LegIDs) > 10 {
		respondError(c, bad("Choose a valid flight selection"))
		return
	}
	viewer := s.viewer(c)
	viewer.UserID = user.ID
	var result gin.H
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, flight, current, err := s.publishedFlight(tx, c, viewer)
		if err != nil {
			return err
		}
		eligible, manager, participation := s.flightEligible(tx, stop, current)
		if body.Status != "removed" && (!eligible || flight.Status == "cancelled" || stop.Published.Status == "cancelled") {
			return errForbidden
		}
		if body.Status != "removed" && body.FlightVersion != flightVersion(flight) {
			return tripConflict{"The flight changed. Review its details before selecting it", stop.Revision}
		}
		if body.Status != "removed" && body.ShareWithTravelers && body.RosterVersion != stop.RosterVersion {
			return tripConflict{"The traveler audience changed. Review it before sharing your selection", stop.Revision}
		}
		if body.Status != "removed" && body.ShareWithTravelers && !manager && (participation == nil || participation.RosterConsentVersion != stop.RosterVersion) {
			return tripConflict{"Review the trip's attendee-list privacy settings before sharing your flight plans", stop.Revision}
		}
		flightID, err := uuid.Parse(flight.ID)
		if err != nil {
			return errUnavailable
		}
		selection := FlightSelection{StopID: stop.ID, FlightID: flightID, UserID: user.ID}
		if err = tx.Where("stop_id = ? AND flight_id = ? AND user_id = ?", stop.ID, flight.ID, user.ID).First(&selection).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		selection.Status, selection.ShareWithTravelers, selection.RosterConsentVersion = body.Status, body.ShareWithTravelers, body.RosterVersion
		selection.LegIDs, selection.LegVersions = []string{}, map[string]string{}
		if body.Status != "removed" {
			available := map[string]TripFlightLeg{}
			for _, leg := range flight.Legs {
				available[leg.ID] = leg
			}
			if len(body.LegIDs) == 0 {
				return bad("Choose at least one flight leg")
			}
			for _, id := range body.LegIDs {
				leg, exists := available[id]
				if !exists || selection.LegVersions[id] != "" {
					return bad("Choose distinct legs from this itinerary")
				}
				selection.LegIDs = append(selection.LegIDs, id)
				selection.LegVersions[id] = legVersion(flight, leg)
			}
		} else {
			selection.ShareWithTravelers = false
		}
		if err = tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "stop_id"}, {Name: "flight_id"}, {Name: "user_id"}}, UpdateAll: true}).Create(&selection).Error; err != nil {
			return err
		}
		// The shared trip audit is visible to hosts. Personal choices must never
		// be copied there, including choices whose sharing is later withdrawn.
		result, err = s.flightTravelersDTO(tx, stop, flight, current)
		return err
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, result)
}
