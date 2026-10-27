package travelcalendar

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errUnavailable = errors.New("unavailable")
var errForbidden = errors.New("forbidden")
var errPreview = errors.New("preview rollback")

type requestError struct{ message string }

func (e requestError) Error() string { return e.message }
func bad(message string) error       { return requestError{message} }
func respondError(c *gin.Context, e error) {
	var conflict tripConflict
	if errors.As(e, &conflict) {
		c.JSON(409, gin.H{"error": conflict.message, "revision": conflict.revision})
		return
	}
	if errors.Is(e, errUnavailable) || errors.Is(e, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "This plan is unavailable or requires access"})
		return
	}
	if errors.Is(e, errForbidden) {
		c.JSON(403, gin.H{"error": "You do not have permission to do this"})
		return
	}
	var r requestError
	if errors.As(e, &r) {
		c.JSON(400, gin.H{"error": r.message})
		return
	}
	c.JSON(500, gin.H{"error": "Unable to complete this calendar action"})
}
func (s *Service) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/travel-angels")
	g.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
		c.Next()
	})
	g.GET("/calendar", s.getCalendar)
	g.GET("/destinations", s.searchDestinations)
	g.PUT("/calendar", s.updateCalendar)
	g.GET("/calendar/permissions", s.getPermissions)
	g.PUT("/calendar/permissions", s.updatePermissions)
	g.POST("/calendar/permissions/preview", s.previewPermissions)
	g.POST("/calendar/share-link", s.updateShareLink)
	g.POST("/calendar/collaborators", s.addCollaborator)
	g.DELETE("/calendar/collaborators/:collaboratorId", s.removeCollaborator)
	g.POST("/calendar/collaborators/:collaboratorId/removal-preview", s.removeCollaborator)
	g.POST("/calendar/collaborators/:collaboratorId/accept", s.acceptCollaborator)
	g.POST("/calendar/stops", s.createStop)
	g.GET("/calendar/stops/:id", s.getStop)
	g.PUT("/calendar/stops/:id", s.updateStop)
	g.POST("/calendar/stops/:id/publish", s.publishStop)
	g.POST("/calendar/stops/:id/preview", s.previewStop)
	g.POST("/calendar/stops/:id/unpublish", s.unpublishStop)
	g.GET("/calendar/stops/:id/permissions", s.getPermissions)
	g.PUT("/calendar/stops/:id/permissions", s.updatePermissions)
	g.POST("/calendar/stops/:id/permissions/preview", s.previewPermissions)
	g.POST("/calendar/stops/:id/share-link", s.updateShareLink)
	g.POST("/calendar/stops/:id/collaborators", s.addCollaborator)
	g.DELETE("/calendar/stops/:id/collaborators/:collaboratorId", s.removeCollaborator)
	g.POST("/calendar/stops/:id/collaborators/:collaboratorId/removal-preview", s.removeCollaborator)
	g.POST("/calendar/stops/:id/participation", s.setParticipation)
	g.GET("/calendar/stops/:id/participations", s.listParticipations)
	g.PUT("/calendar/stops/:id/participations/:participationId", s.reviewParticipation)
	g.GET("/shared/calendars/:token", s.getSharedCalendar)
	g.GET("/shared/stops/:token", s.getSharedStop)
	s.registerTripRoutes(g)
	s.registerTripMediaRoutes(g)
	s.registerGuestRoutes(r)
}
func (s *Service) primaryCalendar(tx *gorm.DB, userID uuid.UUID) (*Calendar, error) {
	cal := Calendar{ID: uuid.New(), OwnerID: userID, Title: "My travel calendar", Visibility: "private", ShareToken: newToken(), SharingEnabled: true}
	if e := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "owner_id"}}, DoNothing: true}).Create(&cal).Error; e != nil {
		return nil, e
	}
	cal = Calendar{}
	if e := tx.Where("owner_id = ?", userID).First(&cal).Error; e != nil {
		return nil, e
	}
	return &cal, nil
}
func (s *Service) calendarDTO(cal *Calendar, v Viewer) gin.H {
	manager := v.UserID == cal.OwnerID
	if !manager && v.UserID != uuid.Nil {
		var count int64
		s.db.Model(&Collaborator{}).Where("calendar_id = ? AND user_id = ? AND accepted = ?", cal.ID, v.UserID, true).Count(&count)
		manager = count > 0
	}
	dto := gin.H{"id": cal.ID, "title": "Travel calendar", "canManage": manager, "canManagePermissions": v.UserID == cal.OwnerID, "canCreateTrips": v.UserID == cal.OwnerID}
	authorized := manager || (cal.Visibility == "specific" && hasGrant(s.db, "calendar_id", cal.ID, v)) || (cal.Visibility == "link" && cal.SharingEnabled && v.CalendarLinkID == cal.ID)
	if !authorized {
		var stops []Stop
		if s.db.Where("calendar_id = ?", cal.ID).Find(&stops).Error == nil {
			for i := range stops {
				if s.canViewStop(s.db, &stops[i], v) {
					authorized = true
					break
				}
			}
		}
	}
	if authorized {
		dto["title"] = cal.Title
		dto["ownerDisplayName"] = s.userDisplayName(cal.OwnerID)
	}
	if manager {
		dto["title"] = cal.Title
		dto["ownerId"] = cal.OwnerID
		dto["visibility"] = cal.Visibility
		dto["shareToken"] = cal.ShareToken
		dto["sharingEnabled"] = cal.SharingEnabled
		dto["shareUrl"] = s.baseURL + "/#/calendar/" + cal.ShareToken
	}
	return dto
}
func (s *Service) stopDTO(stop *Stop, v Viewer) gin.H {
	content := stop.Published
	manager := s.canManageStop(s.db, stop, v)
	if content == nil {
		content = &stop.Draft
	}
	dto := gin.H{"id": stop.ID, "calendarId": stop.CalendarID, "title": content.Title, "destination": content.Destination, "startDate": content.StartDate, "endDate": content.EndDate, "datePrecision": content.DatePrecision, "timeZone": effectiveTimeZone(content.TimeZone), "description": content.Description, "status": content.Status, "joinOpen": content.JoinOpen, "published": stop.Published != nil, "canManage": manager, "canManagePermissions": s.canManagePermissions(s.db, stop, v)}
	dto["destinations"] = contentDestinations(content)
	isPast := content.EndDate != "" && content.EndDate < todayForStop(content, time.Now())
	dto["isPast"] = isPast
	dto["acceptingParticipation"] = stop.Published != nil && content.JoinOpen && content.Status != "cancelled" && !isPast
	var owning Calendar
	if s.db.First(&owning, "id = ?", stop.CalendarID).Error == nil {
		dto["canManageCollaborators"] = v.UserID == owning.OwnerID
	}
	if manager {
		draft := stop.Draft
		draft.Destinations = contentDestinations(&draft)
		dto["draft"] = draft
		dto["visibility"] = stop.Visibility
		dto["shareToken"] = stop.ShareToken
		dto["sharingEnabled"] = stop.SharingEnabled
		dto["shareUrl"] = s.baseURL + "/#/stops/" + stop.ShareToken
	}
	if v.UserID != uuid.Nil {
		var p Participation
		if s.db.Where("stop_id = ? AND user_id = ?", stop.ID, v.UserID).First(&p).Error == nil {
			dto["ownParticipation"] = p
			if p.Status == "removed" {
				dto["acceptingParticipation"] = false
			}
		}
	}
	for key, value := range s.tripPageDTO(stop, v, manager) {
		dto[key] = value
	}
	for key, value := range s.tripLogisticsDTO(stop, content, v, manager) {
		dto[key] = value
	}
	return dto
}
func (s *Service) visibleStops(stops []Stop, v Viewer, c *gin.Context) []gin.H {
	result := make([]gin.H, 0)
	destination := strings.ToLower(strings.TrimSpace(c.Query("destination")))
	from, to := c.Query("from"), c.Query("to")
	for i := range stops {
		st := &stops[i]
		if !s.canViewStop(s.db, st, v) {
			continue
		}
		dto := s.stopDTO(st, v)
		content := st.Published
		if content == nil {
			content = &st.Draft
		}
		if destination != "" && !matchesDestinationQuery(content, destination) {
			continue
		}
		start, end := dto["startDate"].(string), dto["endDate"].(string)
		if start != "" && ((from != "" && end < from) || (to != "" && start > to)) {
			continue
		}
		result = append(result, dto)
	}
	sort.SliceStable(result, func(i, j int) bool {
		a, b := result[i]["startDate"].(string), result[j]["startDate"].(string)
		if a == "" {
			return false
		}
		if b == "" {
			return true
		}
		return a < b
	})
	return result
}
func (s *Service) getCalendar(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	cal, e := s.primaryCalendar(s.db, u.ID)
	if e != nil {
		respondError(c, e)
		return
	}
	var stops []Stop
	e = s.db.Where(`calendar_id = ? OR calendar_id IN (SELECT calendar_id FROM tc_collaborators WHERE user_id = ? AND accepted = true AND calendar_id IS NOT NULL) OR id IN (SELECT stop_id FROM tc_collaborators WHERE user_id = ? AND accepted = true AND stop_id IS NOT NULL) OR id IN (SELECT stop_id FROM tc_participations WHERE user_id = ? AND (status IN ('joined','needs_reconfirmation') OR reserved = true))`, cal.ID, u.ID, u.ID, u.ID).Find(&stops).Error
	if e != nil {
		respondError(c, e)
		return
	}
	v := s.viewer(c)
	v.UserID = u.ID
	// Owning a calendar is not a bearer capability for a referenced stop. Link
	// access is granted only when the referenced owning calendar still enables it.
	var invitations []Collaborator
	if e = s.db.Where("user_id = ? AND accepted = ? AND (stop_id IS NULL OR is_host = true OR stop_id IN (SELECT id FROM tc_stops WHERE host_model = false))", u.ID, false).Find(&invitations).Error; e != nil {
		respondError(c, e)
		return
	}
	var editable []Calendar
	if e = s.db.Where("owner_id = ? OR id IN (SELECT calendar_id FROM tc_collaborators WHERE user_id = ? AND accepted = true)", u.ID, u.ID).Find(&editable).Error; e != nil {
		respondError(c, e)
		return
	}
	editableDTO := make([]gin.H, 0, len(editable))
	for i := range editable {
		editableDTO = append(editableDTO, s.calendarDTO(&editable[i], v))
	}
	invitationDTO := []gin.H{}
	for _, invite := range invitations {
		row := gin.H{"id": invite.ID, "userId": invite.UserID, "stopId": invite.StopID, "calendarId": invite.CalendarID, "isHost": invite.IsHost, "canManagePermissions": invite.CanManagePermissions, "accepted": invite.Accepted}
		if invite.StopID != nil {
			var invited Stop
			if s.db.First(&invited, "id = ?", *invite.StopID).Error == nil {
				row["rosterVersion"] = invited.RosterVersion
				row["rosterVisibility"] = invited.RosterVisibility
				row["revision"] = invited.Revision
				row["hostModel"] = invited.HostModel
				row["tripTitle"] = "Trip host invitation"
				if invited.Published != nil {
					row["tripTitle"] = invited.Published.Title
				}
				if owner, e := s.tripOwner(s.db, &invited); e == nil {
					row["ownerDisplayName"] = s.userDisplayName(owner)
				}
			}
		}
		invitationDTO = append(invitationDTO, row)
	}
	c.JSON(200, gin.H{"calendar": s.calendarDTO(cal, v), "stops": s.visiblePersonalStops(stops, v, c), "cohostInvitations": invitationDTO, "editableCalendars": editableDTO})
}
func (s *Service) visiblePersonalStops(stops []Stop, v Viewer, c *gin.Context) []gin.H {
	// A joined reference remembers that the participant obtained a public link.
	// The originally observed sharing token must still be enabled and unchanged.
	result := make([]gin.H, 0)
	for _, stop := range stops {
		sv := v
		var participation Participation
		if s.db.Where("stop_id = ? AND user_id = ? AND (status IN ('joined','needs_reconfirmation') OR reserved = true)", stop.ID, v.UserID).First(&participation).Error == nil {
			stored := s.storedParticipationViewer(s.db, &stop, &participation)
			sv.CalendarLinkID = stored.CalendarLinkID
			sv.StopLinkID = stored.StopLinkID
		}
		result = append(result, s.visibleStops([]Stop{stop}, sv, c)...)
	}
	sort.SliceStable(result, func(i, j int) bool {
		a, b := result[i]["startDate"].(string), result[j]["startDate"].(string)
		if a == "" {
			return false
		}
		if b == "" {
			return true
		}
		return a < b
	})
	return result
}
func (s *Service) updateCalendar(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		Title string `json:"title"`
	}
	if c.ShouldBindJSON(&b) != nil || strings.TrimSpace(b.Title) == "" || len(b.Title) > 200 {
		respondError(c, bad("Provide a calendar title up to 200 characters"))
		return
	}
	cal, e := s.primaryCalendar(s.db, u.ID)
	if e == nil {
		e = s.db.Model(cal).Update("title", strings.TrimSpace(b.Title)).Error
	}
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, s.calendarDTO(cal, Viewer{UserID: u.ID}))
}
func (s *Service) getSharedCalendar(c *gin.Context) {
	var cal Calendar
	if e := s.db.Where("share_token = ? AND sharing_enabled = ?", c.Param("token"), true).First(&cal).Error; e != nil {
		respondError(c, errUnavailable)
		return
	}
	var stops []Stop
	if e := s.db.Where("calendar_id = ? OR id IN (SELECT stop_id FROM tc_collaborators WHERE user_id = ? AND accepted = true AND stop_id IS NOT NULL)", cal.ID, cal.OwnerID).Find(&stops).Error; e != nil {
		respondError(c, e)
		return
	}
	v := s.viewer(c)
	v.CalendarLinkID = cal.ID
	c.JSON(200, gin.H{"calendar": s.calendarDTO(&cal, v), "stops": s.visibleStops(stops, v, c)})
}
func (s *Service) getSharedStop(c *gin.Context) {
	var stop Stop
	if s.db.Where("share_token = ? AND sharing_enabled = ?", c.Param("token"), true).First(&stop).Error != nil {
		respondError(c, errUnavailable)
		return
	}
	v := s.viewer(c)
	v.StopLinkID = stop.ID
	if !s.canViewStop(s.db, &stop, v) {
		respondError(c, errUnavailable)
		return
	}
	c.JSON(200, gin.H{"stop": s.stopDTO(&stop, v)})
}
func (s *Service) getStop(c *gin.Context) {
	var stop Stop
	if id, e := uuid.Parse(c.Param("id")); e != nil || s.db.First(&stop, "id = ?", id).Error != nil {
		respondError(c, errUnavailable)
		return
	}
	v := s.viewer(c)
	if v.UserID != uuid.Nil {
		pv := s.participationViewer(s.db, &stop, v.UserID, participationInput{})
		if pv.CalendarLinkID != uuid.Nil {
			v.CalendarLinkID = pv.CalendarLinkID
		}
		if pv.StopLinkID != uuid.Nil {
			v.StopLinkID = pv.StopLinkID
		}
	}
	if !s.canViewStop(s.db, &stop, v) {
		respondError(c, errUnavailable)
		return
	}
	c.JSON(200, gin.H{"stop": s.stopDTO(&stop, v)})
}
func (s *Service) createStop(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		StopContent
		CalendarID *uuid.UUID `json:"calendarId"`
	}
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, bad("Invalid stop details"))
		return
	}
	var stop Stop
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var cal *Calendar
		if b.CalendarID == nil {
			var e error
			cal, e = s.primaryCalendar(tx, u.ID)
			if e != nil {
				return e
			}
		} else {
			cal = &Calendar{}
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(cal, "id = ?", *b.CalendarID).Error; e != nil {
				return errUnavailable
			}
			if cal.OwnerID != u.ID {
				return errForbidden
			}
		}
		stop = Stop{ID: uuid.New(), CalendarID: cal.ID, Draft: b.StopContent, Visibility: "inherit", ShareToken: newToken(), SharingEnabled: true, HostModel: true, Revision: 1, RosterVersion: 1, RosterVisibility: "attendees", ShowAttendeeCount: true}
		if e := s.prepareDestinations(tx, &stop.Draft, nil); e != nil {
			return e
		}
		if e := stop.Draft.validate(); e != nil {
			return bad(e.Error())
		}
		if e := s.validateTripContentReferences(tx, &stop, &stop.Draft); e != nil {
			return e
		}
		if e := tx.Create(&stop).Error; e != nil {
			return e
		}
		if e := s.recordDestinationUsage(tx, &stop, u.ID); e != nil {
			return e
		}
		p := Participation{ID: uuid.New(), StopID: stop.ID, UserID: u.ID, Status: "joined", RosterConsentVersion: 1}
		if stop.Draft.DatePrecision == "fixed" {
			p.StartDate = stop.Draft.StartDate
			p.EndDate = stop.Draft.EndDate
		}
		if e := tx.Create(&p).Error; e != nil {
			return e
		}
		return s.recordTripAudit(tx, stop.ID, u.ID, "trip_created", gin.H{"creatorAttending": true})
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"stop": s.stopDTO(&stop, Viewer{UserID: u.ID})})
}

// Lock the owning calendar before the stop for every mutation. Permission
// changes and requests therefore cannot race past one another's authorization.
func (s *Service) lockedStop(tx *gorm.DB, id string) (*Stop, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errUnavailable
	}
	var stop Stop
	if e = tx.First(&stop, "id = ?", uid).Error; e != nil {
		return nil, e
	}
	initialCalendar := stop.CalendarID
	var cal Calendar
	if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cal, "id = ?", stop.CalendarID).Error; e != nil {
		return nil, e
	}
	if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stop, "id = ?", uid).Error; e != nil {
		return nil, e
	}
	if stop.CalendarID != initialCalendar {
		return nil, tripConflict{"The trip owner changed. Reload the trip", stop.Revision}
	}
	return &stop, nil
}
func (s *Service) updateStop(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		StopContent
		Revision *int `json:"revision"`
	}
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, bad("Invalid stop details"))
		return
	}
	var stop *Stop
	var obsolete []TripMedia
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var e error
		stop, e = s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: u.ID}) {
			return errForbidden
		}
		if e = s.checkTripRevision(stop, b.Revision); e != nil {
			return e
		}
		if e = s.prepareDestinations(tx, &b.StopContent, &stop.Draft); e != nil {
			return e
		}
		if e = b.validate(); e != nil {
			return bad(e.Error())
		}
		if e = s.validateTripContentReferences(tx, stop, &b.StopContent); e != nil {
			return e
		}
		stop.Draft = b.StopContent
		if e = s.saveTripChange(tx, stop, u.ID, "draft_updated", nil); e != nil {
			return e
		}
		if e = s.recordDestinationUsage(tx, stop, u.ID); e != nil {
			return e
		}
		obsolete, e = pruneTripMedia(tx, stop)
		return e
	})
	if e != nil {
		respondError(c, e)
		return
	}
	s.deleteTripImageObjects(obsolete)
	c.JSON(200, gin.H{"stop": s.stopDTO(stop, Viewer{UserID: u.ID})})
}
func materialChange(before, after *StopContent) bool {
	return before != nil && (destinationIdentities(before) != destinationIdentities(after) || before.StartDate != after.StartDate || before.EndDate != after.EndDate || before.DatePrecision != after.DatePrecision || effectiveTimeZone(before.TimeZone) != effectiveTimeZone(after.TimeZone))
}
func (s *Service) publishStop(c *gin.Context)   { s.changePublication(c, true) }
func (s *Service) unpublishStop(c *gin.Context) { s.changePublication(c, false) }
func (s *Service) changePublication(c *gin.Context, publish bool) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision *int `json:"revision"`
	}
	if e := optionalTripBody(c, &body); e != nil {
		respondError(c, e)
		return
	}
	var stop *Stop
	var obsolete []TripMedia
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var e error
		stop, e = s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: u.ID}) {
			return errForbidden
		}
		if e = s.checkTripRevision(stop, body.Revision); e != nil {
			return e
		}
		prior := stop.Published
		if !publish && prior == nil {
			return nil
		}
		event := ""
		cancelled := false
		if publish {
			if e = s.prepareDestinations(tx, &stop.Draft, nil); e != nil {
				return e
			}
			if e = stop.Draft.validate(); e != nil {
				return bad(e.Error())
			}
			if e = s.validateTripContentReferences(tx, stop, &stop.Draft); e != nil {
				return e
			}
			next := stop.Draft
			stop.Published = &next
			cancelled = next.Status == "cancelled"
			if cancelled {
				event = "cancelled"
			} else if materialChange(prior, &next) {
				event = "needs_reconfirmation"
			}
		} else {
			stop.Published = nil
			event = "unpublished"
		}
		action := "trip_published"
		if !publish {
			action = "trip_unpublished"
		}
		if e = s.saveTripChange(tx, stop, u.ID, action, nil); e != nil {
			return e
		}
		if event != "" {
			var ps []Participation
			if e = tx.Where("stop_id = ? AND status IN ('interested','requested','joined','needs_reconfirmation','waitlisted')", stop.ID).Find(&ps).Error; e != nil {
				return e
			}
			for i := range ps {
				p := &ps[i]
				if cancelled {
					p.Status = "cancelled"
					p.Reserved = false
				} else if p.Status == "requested" || p.Status == "joined" {
					if p.Status == "joined" {
						p.Reserved = true
					}
					p.Status = "needs_reconfirmation"
				}
				if e = tx.Save(p).Error; e != nil {
					return e
				}
				if e = s.queueParticipationNotice(tx, stop, p, event); e != nil {
					return e
				}
			}
		}
		if publish && prior != nil && logisticsChanged(prior, stop.Published) {
			if e = s.queueTripLogisticsNotices(tx, stop, prior, stop.Published); e != nil {
				return e
			}
		}
		obsolete, e = pruneTripMedia(tx, stop)
		return e
	})
	if e != nil {
		respondError(c, e)
		return
	}
	s.deleteTripImageObjects(obsolete)
	c.JSON(200, gin.H{"stop": s.stopDTO(stop, Viewer{UserID: u.ID})})
}

// Resolve only a display label; never serialize an account's contacts or profile.
func (s *Service) userDisplayName(id uuid.UUID) string {
	var row struct {
		Name     string
		Username *string
	}
	if s.db.Table("users").Select("name, username").Where("id = ?", id).Take(&row).Error == nil {
		if strings.TrimSpace(row.Name) != "" {
			return row.Name
		}
		if row.Username != nil && strings.TrimSpace(*row.Username) != "" {
			return *row.Username
		}
	}
	return "Traveler"
}
