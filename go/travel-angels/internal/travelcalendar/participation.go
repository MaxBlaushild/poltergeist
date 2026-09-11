package travelcalendar

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type participationInput struct {
	Status        string `json:"status"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	CalendarToken string `json:"calendarToken"`
	StopToken     string `json:"stopToken"`
}

func validateParticipation(content *StopContent, input *participationInput, today string) error {
	if content == nil || content.Status == "cancelled" || !content.JoinOpen {
		return bad("This stop is not accepting participation")
	}
	if content.EndDate != "" && content.EndDate < today {
		return bad("This stop has already ended")
	}
	if input.Status != "interested" && input.Status != "requested" {
		return bad("Choose interested or requested")
	}
	if input.Status == "requested" && content.DatePrecision != "fixed" {
		return bad("Requesting to join requires fixed dates")
	}
	if content.DatePrecision != "fixed" {
		if input.StartDate != "" || input.EndDate != "" {
			return bad("Choose interest without dates until the plan has fixed dates")
		}
		return nil
	}
	if input.StartDate == "" && input.EndDate == "" {
		input.StartDate = content.StartDate
		input.EndDate = content.EndDate
	}
	if !validDate(input.StartDate) || !validDate(input.EndDate) || input.EndDate < input.StartDate || input.StartDate < content.StartDate || input.EndDate > content.EndDate {
		return bad("Participation dates must be within the stop's dates")
	}
	if input.EndDate < today {
		return bad("Choose dates that have not ended")
	}
	return nil
}
func (s *Service) participationViewer(tx *gorm.DB, stop *Stop, userID uuid.UUID, b participationInput) Viewer {
	// Guest tokens confer browsing only. Recipient grants used to RSVP must be
	// linked to this claimed account, which hasGrant verifies independently.
	v := Viewer{UserID: userID}
	var cal Calendar
	if b.CalendarToken != "" && tx.Where("id = ? AND share_token = ? AND sharing_enabled = ?", stop.CalendarID, b.CalendarToken, true).First(&cal).Error == nil {
		v.CalendarLinkID = cal.ID
	}
	if b.StopToken != "" && b.StopToken == stop.ShareToken && stop.SharingEnabled {
		v.StopLinkID = stop.ID
	}
	var p Participation
	if tx.Where("stop_id = ? AND user_id = ?", stop.ID, userID).First(&p).Error == nil && activeParticipation(p.Status) {
		stored := s.storedParticipationViewer(tx, stop, &p)
		if stored.CalendarLinkID != uuid.Nil {
			v.CalendarLinkID = stored.CalendarLinkID
		}
		if stored.StopLinkID != uuid.Nil {
			v.StopLinkID = stored.StopLinkID
		}
	}
	return v
}
func (s *Service) setParticipation(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b participationInput
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, bad("Invalid participation"))
		return
	}
	b.Status = normalizedStatus(b.Status)
	if b.CalendarToken == "" {
		b.CalendarToken = c.Query("calendarToken")
	}
	if b.StopToken == "" {
		b.StopToken = c.Query("stopToken")
	}
	var result Participation
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		e = tx.Where("stop_id = ? AND user_id = ?", stop.ID, u.ID).First(&result).Error
		exists := e == nil
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if b.Status == "withdrawn" {
			if !exists {
				return errUnavailable
			}
			if result.Status == "removed" {
				return bad("Ask a host to restore your eligibility")
			}
			if result.Status == "withdrawn" {
				return nil
			}
			result.Status = "withdrawn"
			if e = tx.Save(&result).Error; e != nil {
				return e
			}
			return s.queueParticipationNotice(tx, stop, &result, "withdrawn")
		}
		if exists && result.Status == "removed" {
			return bad("Ask a host to restore your eligibility")
		}
		if !s.canViewStop(tx, stop, s.participationViewer(tx, stop, u.ID, b)) {
			return errUnavailable
		}
		if e = validateParticipation(stop.Published, &b, todayForStop(stop.Published, time.Now())); e != nil {
			return e
		}
		unchanged := exists && result.Status == b.Status && result.StartDate == b.StartDate && result.EndDate == b.EndDate
		priorCalendarHash, priorStopHash := result.CalendarLinkTokenHash, result.StopLinkTokenHash
		if !exists {
			result = Participation{ID: uuid.New(), StopID: stop.ID, UserID: u.ID}
		}
		if b.StopToken != "" && b.StopToken == stop.ShareToken && stop.SharingEnabled {
			result.StopLinkTokenHash = guestDigest(b.StopToken)
		}
		if b.CalendarToken != "" {
			var cal Calendar
			if tx.Where("id = ? AND share_token = ? AND sharing_enabled = ?", stop.CalendarID, b.CalendarToken, true).First(&cal).Error == nil {
				result.CalendarLinkTokenHash = guestDigest(b.CalendarToken)
			}
		}
		if unchanged {
			if result.CalendarLinkTokenHash != priorCalendarHash || result.StopLinkTokenHash != priorStopHash {
				return tx.Save(&result).Error
			}
			return nil
		}
		result.Status = b.Status
		result.StartDate = b.StartDate
		result.EndDate = b.EndDate
		if e = tx.Save(&result).Error; e != nil {
			return e
		}
		return s.queueParticipationNotice(tx, stop, &result, b.Status)
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, result)
}
func (s *Service) listParticipations(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var rows []Participation
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: u.ID}) {
			return errForbidden
		}
		return tx.Where("stop_id = ?", stop.ID).Order("created_at ASC").Find(&rows).Error
	})
	if e != nil {
		respondError(c, e)
		return
	}
	result := make([]gin.H, 0, len(rows))
	for _, p := range rows {
		result = append(result, gin.H{"id": p.ID, "stopId": p.StopID, "userId": p.UserID, "status": p.Status, "startDate": p.StartDate, "endDate": p.EndDate, "createdAt": p.CreatedAt, "updatedAt": p.UpdatedAt, "displayName": s.userDisplayName(p.UserID)})
	}
	c.JSON(200, gin.H{"participations": result})
}
func validateReview(current, target string) error {
	switch target {
	case "joined":
		if current != "requested" {
			return bad("Only a current join request can be approved")
		}
	case "declined":
		if current != "requested" && current != "interested" && current != "needs_reconfirmation" {
			return bad("There is no pending participation to decline")
		}
	case "removed":
		if !activeParticipation(current) {
			return bad("There is no active participation to remove")
		}
	case "withdrawn":
		if current != "removed" {
			return bad("Only removed participation can have eligibility restored")
		}
	default:
		return bad("Invalid review status")
	}
	return nil
}
func (s *Service) reviewParticipation(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		Status string `json:"status"`
	}
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, bad("Invalid review"))
		return
	}
	b.Status = normalizedStatus(b.Status)
	id, e := uuid.Parse(c.Param("participationId"))
	if e != nil {
		respondError(c, errUnavailable)
		return
	}
	var result Participation
	e = s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: u.ID}) {
			return errForbidden
		}
		if e = tx.Where("id = ? AND stop_id = ?", id, stop.ID).First(&result).Error; e != nil {
			return e
		}
		if result.Status == b.Status {
			return nil
		}
		if e = validateReview(result.Status, b.Status); e != nil {
			return e
		}
		if b.Status == "joined" {
			v := s.storedParticipationViewer(tx, stop, &result)
			if !s.canViewStop(tx, stop, v) {
				return bad("This person no longer has viewing access")
			}
			input := participationInput{Status: "requested", StartDate: result.StartDate, EndDate: result.EndDate}
			if e = validateParticipation(stop.Published, &input, todayForStop(stop.Published, time.Now())); e != nil {
				return e
			}
		}
		result.Status = b.Status
		if e = tx.Save(&result).Error; e != nil {
			return e
		}
		return s.queueParticipationNotice(tx, stop, &result, b.Status)
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, result)
}

// Participation remembers only bearer links the user actually supplied. Hashes
// prevent raw sharing secrets from appearing in participation APIs. Rotation or
// disabling invalidates these capabilities while independent grants still work.
func (s *Service) storedParticipationViewer(tx *gorm.DB, stop *Stop, p *Participation) Viewer {
	v := Viewer{UserID: p.UserID}
	if stop.SharingEnabled && p.StopLinkTokenHash != "" && p.StopLinkTokenHash == guestDigest(stop.ShareToken) {
		v.StopLinkID = stop.ID
	}
	if p.CalendarLinkTokenHash != "" {
		var cal Calendar
		if tx.First(&cal, "id = ?", stop.CalendarID).Error == nil && cal.SharingEnabled && p.CalendarLinkTokenHash == guestDigest(cal.ShareToken) {
			v.CalendarLinkID = cal.ID
		}
	}
	return v
}

// Dates stay local YYYY-MM-DD values. The destination zone is used only to
// determine today's date, so an RSVP does not close at another city's midnight.
func todayForStop(content *StopContent, now time.Time) string {
	location := time.UTC
	if content != nil && content.TimeZone != "" {
		if destination, err := time.LoadLocation(content.TimeZone); err == nil {
			location = destination
		}
	}
	return now.In(location).Format("2006-01-02")
}
