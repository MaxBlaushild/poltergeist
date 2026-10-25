package travelcalendar

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tripConflict struct {
	message  string
	revision int
}

func (e tripConflict) Error() string { return e.message }
func (s *Service) checkTripRevision(stop *Stop, expected *int) error {
	if expected == nil {
		if stop.HostModel {
			return tripConflict{"Reload this trip before saving your changes", stop.Revision}
		}
		return nil
	}
	if *expected != stop.Revision {
		return tripConflict{"This trip changed. Reload it and review your changes before saving", stop.Revision}
	}
	return nil
}
func mutationRevision(c *gin.Context, value *int) *int {
	if value != nil {
		return value
	}
	if n, e := strconv.Atoi(c.Query("revision")); e == nil {
		return &n
	}
	return nil
}
func optionalTripBody(c *gin.Context, body any) error {
	if e := c.ShouldBindJSON(body); e != nil && !errors.Is(e, io.EOF) {
		return bad("Invalid trip action")
	}
	return nil
}

type tripAuditEvent struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	StopID    uuid.UUID      `gorm:"type:uuid" json:"stopId"`
	ActorID   uuid.UUID      `gorm:"type:uuid" json:"actorId"`
	Action    string         `json:"action"`
	Details   map[string]any `gorm:"serializer:json;type:jsonb" json:"details"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (tripAuditEvent) TableName() string { return "tc_trip_audit_events" }
func (s *Service) recordTripAudit(tx *gorm.DB, stopID, actorID uuid.UUID, action string, details any) error {
	fields := map[string]any{}
	if details != nil {
		encoded, e := json.Marshal(details)
		if e != nil {
			return e
		}
		if e = json.Unmarshal(encoded, &fields); e != nil {
			return e
		}
	}
	return tx.Create(&tripAuditEvent{ID: uuid.New(), StopID: stopID, ActorID: actorID, Action: action, Details: fields}).Error
}
func (s *Service) saveTripChange(tx *gorm.DB, stop *Stop, actorID uuid.UUID, action string, details any) error {
	stop.Revision++
	if e := tx.Save(stop).Error; e != nil {
		return e
	}
	return s.recordTripAudit(tx, stop.ID, actorID, action, details)
}
func (s *Service) tripOwner(tx *gorm.DB, stop *Stop) (uuid.UUID, error) {
	var cal Calendar
	if e := tx.First(&cal, "id = ?", stop.CalendarID).Error; e != nil {
		return uuid.Nil, e
	}
	return cal.OwnerID, nil
}
func (s *Service) tripAttendee(tx *gorm.DB, stop *Stop, v Viewer) bool {
	if v.UserID == uuid.Nil {
		return false
	}
	var p Participation
	if tx.Where("stop_id = ? AND user_id = ?", stop.ID, v.UserID).First(&p).Error != nil {
		return false
	}
	if p.Status != "joined" && !(p.Reserved && (p.Status == "requested" || p.Status == "needs_reconfirmation")) {
		return false
	}
	stored := s.storedParticipationViewer(tx, stop, &p)
	if v.CalendarLinkID == uuid.Nil {
		v.CalendarLinkID = stored.CalendarLinkID
	}
	if v.StopLinkID == uuid.Nil {
		v.StopLinkID = stored.StopLinkID
	}
	return s.canViewStop(tx, stop, v)
}
func participationOccupies(p *Participation) bool {
	return p != nil && (p.Status == "joined" || p.Reserved)
}
func (s *Service) tripOccupancy(tx *gorm.DB, stopID uuid.UUID) (int, int, error) {
	var confirmed, reserved int64
	if e := tx.Model(&Participation{}).Where("stop_id = ? AND status = ?", stopID, "joined").Count(&confirmed).Error; e != nil {
		return 0, 0, e
	}
	if e := tx.Model(&Participation{}).Where("stop_id = ? AND reserved = ? AND status IN ('requested','needs_reconfirmation')", stopID, true).Count(&reserved).Error; e != nil {
		return 0, 0, e
	}
	return int(confirmed), int(reserved), nil
}
func (s *Service) ensureTripCapacity(tx *gorm.DB, stop *Stop, p *Participation) error {
	if stop.MaxAttendees == nil || participationOccupies(p) {
		return nil
	}
	confirmed, reserved, e := s.tripOccupancy(tx, stop.ID)
	if e != nil {
		return e
	}
	if confirmed+reserved >= *stop.MaxAttendees {
		return tripConflict{"This trip is full. A host must free a place or raise the maximum before confirming another attendee", stop.Revision}
	}
	return nil
}
func (s *Service) tripSafeIdentity(tx *gorm.DB, id uuid.UUID) gin.H {
	var row struct {
		Name              string
		Username          *string
		ProfilePictureURL string
	}
	name := "Traveler"
	avatar := ""
	if tx.Table("users").Select("name, username, profile_picture_url").Where("id = ?", id).Take(&row).Error == nil {
		if row.Name != "" {
			name = row.Name
		} else if row.Username != nil && *row.Username != "" {
			name = *row.Username
		}
		avatar = row.ProfilePictureURL
	}
	return gin.H{"userId": id, "displayName": name, "avatarUrl": avatar}
}
func (s *Service) tripPageDTO(stop *Stop, v Viewer, manager bool) gin.H {
	owner, e := s.tripOwner(s.db, stop)
	if e != nil {
		return gin.H{}
	}
	hosts := []gin.H{}
	hostIDs := map[uuid.UUID]bool{owner: true}
	var roles []Collaborator
	query := s.db.Where("stop_id = ? AND accepted = ?", stop.ID, true)
	if stop.HostModel {
		query = query.Where("is_host = ?", true)
	} else {
		query = query.Where("is_host = ?", true)
	}
	if query.Find(&roles).Error == nil {
		for _, role := range roles {
			hostIDs[role.UserID] = true
		}
	}
	var participants []Participation
	if s.db.Where("stop_id = ?", stop.ID).Order("created_at").Find(&participants).Error != nil {
		participants = []Participation{}
	}
	attendance := map[uuid.UUID]Participation{}
	for _, p := range participants {
		attendance[p.UserID] = p
	}
	orderedHosts := []uuid.UUID{owner}
	for _, role := range roles {
		if role.UserID != owner {
			orderedHosts = append(orderedHosts, role.UserID)
		}
	}
	for _, id := range orderedHosts {
		row := s.tripSafeIdentity(s.db, id)
		p, ok := attendance[id]
		row["isOwner"] = id == owner
		row["isAttending"] = ok && p.Status == "joined"
		hosts = append(hosts, row)
	}
	attendee := s.tripAttendee(s.db, stop, v)
	canSee := manager || (attendee && stop.RosterVisibility == "attendees")
	roster := []gin.H{}
	for _, p := range participants {
		if p.Status != "joined" || !canSee {
			continue
		}
		if !manager && (p.RosterConsentVersion != stop.RosterVersion || stop.RosterVisibility != "attendees") {
			continue
		}
		row := s.tripSafeIdentity(s.db, p.UserID)
		row["isHost"] = hostIDs[p.UserID]
		if manager || p.UserID == v.UserID || p.ShareDates {
			row["startDate"] = p.StartDate
			row["endDate"] = p.EndDate
		}
		roster = append(roster, row)
	}
	confirmed, reserved, e := s.tripOccupancy(s.db, stop.ID)
	if e != nil {
		confirmed = 0
		reserved = 0
	}
	full := stop.MaxAttendees != nil && confirmed+reserved >= *stop.MaxAttendees
	capacity := gin.H{"full": full, "waitlistEnabled": stop.WaitlistEnabled, "showCount": stop.ShowAttendeeCount || manager}
	joinOpen := stop.Draft.JoinOpen
	if stop.Published != nil {
		joinOpen = stop.Published.JoinOpen
	}
	settings := gin.H{"waitlistEnabled": stop.WaitlistEnabled, "rosterVisibility": stop.RosterVisibility, "showAttendeeCount": stop.ShowAttendeeCount, "joinOpen": joinOpen}
	if manager || stop.ShowAttendeeCount {
		capacity["maximum"] = stop.MaxAttendees
		capacity["confirmedCount"] = confirmed
		capacity["reservedCount"] = reserved
		if stop.MaxAttendees != nil {
			remaining := *stop.MaxAttendees - confirmed - reserved
			if remaining < 0 {
				remaining = 0
			}
			capacity["remainingPlaces"] = remaining
		} else {
			capacity["remainingPlaces"] = nil
		}
		settings["maxAttendees"] = stop.MaxAttendees
	}
	own, hasOwn := attendance[v.UserID]
	return gin.H{"hostModel": stop.HostModel, "revision": stop.Revision, "ownerId": owner, "hosts": hosts, "attendees": roster, "canViewAttendees": canSee, "canManageCollaborators": manager && (stop.HostModel || owner == v.UserID), "canManageTripSettings": manager && (stop.HostModel || owner == v.UserID), "rosterVersion": stop.RosterVersion, "ownRosterConsentRequired": hasOwn && own.RosterConsentVersion != stop.RosterVersion, "tripSettings": settings, "capacity": capacity}
}
func (s *Service) registerTripRoutes(g *gin.RouterGroup) {
	g.PUT("/calendar/stops/:id/settings", s.updateTripSettings)
	g.POST("/calendar/stops/:id/roster-consent", s.updateRosterConsent)
	g.POST("/calendar/stops/:id/host-attendance", s.updateHostAttendance)
	g.GET("/calendar/stops/:id/host-conversion", s.getHostConversion)
	g.POST("/calendar/stops/:id/host-conversion/preview", s.convertTripHosts)
	g.POST("/calendar/stops/:id/host-conversion", s.convertTripHosts)
	g.POST("/calendar/stops/:id/transfer-ownership/preview", s.transferTripOwnership)
	g.POST("/calendar/stops/:id/transfer-ownership", s.transferTripOwnership)
	g.POST("/calendar/stops/:id/leave-host", s.leaveTripHost)
	g.DELETE("/calendar/stops/:id", s.deleteTrip)
	g.GET("/calendar/stops/:id/audit", s.getTripAudit)
}
func (s *Service) updateTripSettings(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision          *int   `json:"revision"`
		MaxAttendees      *int   `json:"maxAttendees"`
		WaitlistEnabled   bool   `json:"waitlistEnabled"`
		RosterVisibility  string `json:"rosterVisibility"`
		ShowAttendeeCount bool   `json:"showAttendeeCount"`
		JoinOpen          *bool  `json:"joinOpen"`
	}
	if c.ShouldBindJSON(&body) != nil || (body.MaxAttendees != nil && *body.MaxAttendees < 1) || (body.RosterVisibility != "hosts" && body.RosterVisibility != "attendees") {
		respondError(c, bad("Choose a positive maximum or unlimited, and a valid roster audience"))
		return
	}
	var stop *Stop
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var e error
		stop, e = s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		owner, e := s.tripOwner(tx, stop)
		if e != nil {
			return e
		}
		if !s.canManagePermissions(tx, stop, Viewer{UserID: user.ID}) || (!stop.HostModel && owner != user.ID) {
			return errForbidden
		}
		if e = s.checkTripRevision(stop, body.Revision); e != nil {
			return e
		}
		confirmed, reserved, e := s.tripOccupancy(tx, stop.ID)
		if e != nil {
			return e
		}
		if body.MaxAttendees != nil && *body.MaxAttendees < confirmed+reserved {
			return bad("Resolve confirmed attendance and reserved places before lowering the maximum")
		}
		if stop.RosterVisibility == "hosts" && body.RosterVisibility == "attendees" {
			stop.RosterVersion++
		}
		stop.MaxAttendees = body.MaxAttendees
		stop.WaitlistEnabled = body.WaitlistEnabled
		stop.RosterVisibility = body.RosterVisibility
		stop.ShowAttendeeCount = body.ShowAttendeeCount
		if body.JoinOpen != nil {
			stop.Draft.JoinOpen = *body.JoinOpen
			if stop.Published != nil {
				live := *stop.Published
				live.JoinOpen = *body.JoinOpen
				stop.Published = &live
			}
		}
		return s.saveTripChange(tx, stop, user.ID, "settings_updated", gin.H{"maxAttendees": body.MaxAttendees, "waitlistEnabled": body.WaitlistEnabled, "rosterVisibility": body.RosterVisibility, "showAttendeeCount": body.ShowAttendeeCount, "joinOpen": body.JoinOpen})
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"stop": s.stopDTO(stop, Viewer{UserID: user.ID}), "revision": stop.Revision})
}
func (s *Service) updateRosterConsent(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		RosterConsentVersion int  `json:"rosterConsentVersion"`
		ShareDates           bool `json:"shareDates"`
	}
	if c.ShouldBindJSON(&body) != nil {
		respondError(c, bad("Invalid roster preferences"))
		return
	}
	var result Participation
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if e = tx.Where("stop_id = ? AND user_id = ?", stop.ID, user.ID).First(&result).Error; e != nil {
			return e
		}
		if !s.canViewStop(tx, stop, s.participationViewer(tx, stop, user.ID, participationInput{})) {
			return errUnavailable
		}
		if body.RosterConsentVersion != stop.RosterVersion {
			return tripConflict{"The roster audience changed. Review it before consenting", stop.Revision}
		}
		result.RosterConsentVersion = stop.RosterVersion
		result.ShareDates = body.ShareDates
		if e = tx.Save(&result).Error; e != nil {
			return e
		}
		return s.recordTripAudit(tx, stop.ID, user.ID, "roster_consent_updated", gin.H{"consentVersion": stop.RosterVersion, "shareDates": body.ShareDates})
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, result)
}
func (s *Service) updateHostAttendance(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body hostAttendanceInput
	if c.ShouldBindJSON(&body) != nil {
		respondError(c, bad("Choose whether you are attending"))
		return
	}
	var result *Participation
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: user.ID}) {
			return errForbidden
		}
		if !stop.HostModel {
			owner, e := s.tripOwner(tx, stop)
			if e != nil {
				return e
			}
			if owner != user.ID {
				return errForbidden
			}
		}
		result, e = s.setHostAttendance(tx, stop, user.ID, body)
		return e
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"participation": result})
}

type hostAttendanceInput struct {
	Attending            *bool  `json:"attending"`
	RosterConsentVersion int    `json:"rosterConsentVersion"`
	ShareDates           bool   `json:"shareDates"`
	StartDate            string `json:"startDate"`
	EndDate              string `json:"endDate"`
}

func (s *Service) setHostAttendance(tx *gorm.DB, stop *Stop, userID uuid.UUID, body hostAttendanceInput) (*Participation, error) {
	if body.Attending == nil {
		return nil, bad("Choose whether you are attending")
	}
	var p Participation
	e := tx.Where("stop_id = ? AND user_id = ?", stop.ID, userID).First(&p).Error
	if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, e
	}
	if errors.Is(e, gorm.ErrRecordNotFound) {
		p = Participation{ID: uuid.New(), StopID: stop.ID, UserID: userID}
	}
	if !*body.Attending {
		p.Status = "withdrawn"
		p.Reserved = false
	} else {
		content := stop.Published
		if content == nil {
			content = &stop.Draft
		}
		if content.Status == "cancelled" || (content.EndDate != "" && content.EndDate < todayForStop(content, time.Now())) {
			return nil, bad("You cannot attend a cancelled or past trip")
		}
		if body.RosterConsentVersion != stop.RosterVersion {
			return nil, tripConflict{"Review who can see the roster before confirming attendance", stop.Revision}
		}
		if e = s.ensureTripCapacity(tx, stop, &p); e != nil {
			return nil, e
		}
		start, end := body.StartDate, body.EndDate
		if content.DatePrecision == "fixed" {
			if start == "" && end == "" {
				start = content.StartDate
				end = content.EndDate
			}
			if !validDate(start) || !validDate(end) || start > end || start < content.StartDate || end > content.EndDate {
				return nil, bad("Choose attendance dates within the trip")
			}
		} else {
			start = ""
			end = ""
		}
		p.Status = "joined"
		p.Reserved = false
		p.StartDate = start
		p.EndDate = end
		p.RosterConsentVersion = stop.RosterVersion
		p.ShareDates = body.ShareDates
	}
	if e = tx.Save(&p).Error; e != nil {
		return nil, e
	}
	if e = s.recordTripAudit(tx, stop.ID, userID, "host_attendance_updated", gin.H{"attending": *body.Attending}); e != nil {
		return nil, e
	}
	return &p, s.queueParticipationNotice(tx, stop, &p, p.Status)
}
func (s *Service) getTripAudit(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("id"))
	if e != nil {
		respondError(c, errUnavailable)
		return
	}
	var stop Stop
	if s.db.First(&stop, "id = ?", id).Error != nil || !s.canManageStop(s.db, &stop, Viewer{UserID: user.ID}) {
		respondError(c, errUnavailable)
		return
	}
	var events []tripAuditEvent
	if e = s.db.Where("stop_id = ?", id).Order("created_at DESC").Limit(100).Find(&events).Error; e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"events": events})
}
