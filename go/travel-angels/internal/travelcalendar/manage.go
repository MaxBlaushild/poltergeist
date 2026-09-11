package travelcalendar

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type grantInput struct {
	UserID      *uuid.UUID `json:"userId"`
	RecipientID *uuid.UUID `json:"recipientId"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
}
type permissionInput struct {
	Visibility     string        `json:"visibility"`
	Grants         *[]grantInput `json:"grants"`
	ConfirmRemoval bool          `json:"confirmRemoval"`
}

func (s *Service) permissionScope(tx *gorm.DB, c *gin.Context, userID uuid.UUID) (*Calendar, *Stop, error) {
	if c.Param("id") != "" {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return nil, nil, e
		}
		if !s.canManagePermissions(tx, stop, Viewer{UserID: userID}) {
			return nil, nil, errForbidden
		}
		return nil, stop, nil
	}
	cal, e := s.primaryCalendar(tx, userID)
	if e != nil {
		return nil, nil, e
	}
	if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(cal, "id = ?", cal.ID).Error; e != nil {
		return nil, nil, e
	}
	return cal, nil, nil
}
func scopeWhere(cal *Calendar, stop *Stop) (string, uuid.UUID) {
	if stop != nil {
		return "stop_id", stop.ID
	}
	return "calendar_id", cal.ID
}
func (s *Service) getPermissions(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var grants []Grant
	var collaborators []Collaborator
	visibility := ""
	e := s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.permissionScope(tx, c, u.ID)
		if e != nil {
			return e
		}
		scope, id := scopeWhere(cal, stop)
		if stop != nil {
			visibility = stop.Visibility
		} else {
			visibility = cal.Visibility
		}
		if e = tx.Where(scope+" = ?", id).Find(&grants).Error; e != nil {
			return e
		}
		return tx.Where(scope+" = ?", id).Find(&collaborators).Error
	})
	if e != nil {
		respondError(c, e)
		return
	}
	rows := make([]gin.H, 0, len(grants))
	for _, g := range grants {
		row := gin.H{"id": g.ID, "userId": g.UserID, "recipientId": g.RecipientID}
		if g.RecipientID != nil {
			var recipient GuestRecipient
			if s.db.First(&recipient, "id = ?", *g.RecipientID).Error == nil {
				row["email"] = recipient.Email
				row["name"] = recipient.Name
			}
		}
		rows = append(rows, row)
	}
	collaboratorRows := make([]gin.H, 0, len(collaborators))
	for _, person := range collaborators {
		collaboratorRows = append(collaboratorRows, gin.H{"id": person.ID, "calendarId": person.CalendarID, "stopId": person.StopID, "userId": person.UserID, "canManagePermissions": person.CanManagePermissions, "accepted": person.Accepted, "displayName": s.userDisplayName(person.UserID)})
	}
	c.JSON(200, gin.H{"visibility": visibility, "grants": rows, "collaborators": collaboratorRows})
}
func (s *Service) previewPermissions(c *gin.Context) { s.savePermissions(c, true) }
func (s *Service) updatePermissions(c *gin.Context)  { s.savePermissions(c, false) }
func (s *Service) savePermissions(c *gin.Context, preview bool) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b permissionInput
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, bad("Invalid permissions"))
		return
	}
	removed := 0
	e := s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.permissionScope(tx, c, u.ID)
		if e != nil {
			return e
		}
		scope, id := scopeWhere(cal, stop)
		if b.Visibility != "" {
			if b.Visibility != "private" && b.Visibility != "specific" && b.Visibility != "link" && !(stop != nil && b.Visibility == "inherit") {
				return bad("Invalid visibility")
			}
			if stop != nil {
				stop.Visibility = b.Visibility
				if e = tx.Save(stop).Error; e != nil {
					return e
				}
			} else {
				cal.Visibility = b.Visibility
				if e = tx.Save(cal).Error; e != nil {
					return e
				}
			}
		}
		if b.Grants != nil {
			if len(*b.Grants) > 500 {
				return bad("Too many viewers")
			}
			if e = tx.Where(scope+" = ?", id).Delete(&Grant{}).Error; e != nil {
				return e
			}
			seen := map[string]bool{}
			for _, input := range *b.Grants {
				if input.Email != "" {
					if input.UserID != nil {
						return bad("Choose an email or account for each viewer")
					}
					recipient, e := s.ensureGuestRecipient(tx, input.Email, input.Name)
					if e != nil {
						return bad("Provide a valid viewer email")
					}
					input.RecipientID = &recipient.ID
				}
				if (input.UserID == nil) == (input.RecipientID == nil) {
					return bad("Each viewer needs one account or email")
				}
				key := ""
				if input.UserID != nil {
					if *input.UserID == uuid.Nil {
						return bad("Invalid viewer account")
					}
					var n int64
					if tx.Table("users").Where("id = ?", *input.UserID).Count(&n).Error != nil || n != 1 {
						return bad("Invalid viewer account")
					}
					key = "user:" + input.UserID.String()
				} else {
					if *input.RecipientID == uuid.Nil {
						return bad("Invalid recipient")
					}
					key = "recipient:" + input.RecipientID.String()
					var n int64
					if tx.Model(&GuestRecipient{}).Where("id = ?", *input.RecipientID).Count(&n).Error != nil || n != 1 {
						return bad("Invalid recipient")
					}
				}
				if seen[key] {
					continue
				}
				seen[key] = true
				grant := Grant{ID: uuid.New(), UserID: input.UserID, RecipientID: input.RecipientID}
				if stop != nil {
					grant.StopID = &id
				} else {
					grant.CalendarID = &id
				}
				if e = tx.Create(&grant).Error; e != nil {
					return e
				}
			}
		}
		stops := []Stop{}
		if stop != nil {
			stops = append(stops, *stop)
		} else {
			if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("calendar_id = ?", cal.ID).Find(&stops).Error; e != nil {
				return e
			}
		}
		for i := range stops {
			count, e := s.reconcileParticipationAccess(tx, &stops[i], !preview)
			if e != nil {
				return e
			}
			removed += count
		}
		if preview {
			return errPreview
		}
		if removed > 0 && !b.ConfirmRemoval {
			return errPreview
		}
		for i := range stops {
			if e = s.reconcileGuestAccess(tx, &stops[i]); e != nil {
				return e
			}
		}
		return nil
	})
	if errors.Is(e, errPreview) {
		if preview {
			c.JSON(200, gin.H{"removedParticipationCount": removed})
			return
		}
		c.JSON(409, gin.H{"error": "Confirm ending participation for people losing access", "removedParticipationCount": removed})
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"updated": true, "removedParticipationCount": removed})
}
func (s *Service) reconcileParticipationAccess(tx *gorm.DB, stop *Stop, apply bool) (int, error) {
	// An unpublished draft alone is not a grant revocation; managers may keep
	// drafting while participants await reconfirmation after republishing.
	var ps []Participation
	if e := tx.Where("stop_id = ? AND status IN ('interested','requested','joined','needs_reconfirmation')", stop.ID).Find(&ps).Error; e != nil {
		return 0, e
	}
	removed := 0
	policyStop := *stop
	if policyStop.Published == nil {
		content := stop.Draft
		policyStop.Published = &content
	}
	for i := range ps {
		p := &ps[i]
		v := s.storedParticipationViewer(tx, stop, p)
		if s.canViewStop(tx, &policyStop, v) {
			continue
		}
		removed++
		if apply {
			p.Status = "removed"
			if e := tx.Save(p).Error; e != nil {
				return 0, e
			}
			if e := s.queueParticipationNotice(tx, stop, p, "removed"); e != nil {
				return 0, e
			}
		}
	}
	return removed, nil
}
func (s *Service) updateShareLink(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		Enabled        bool `json:"enabled"`
		Rotate         bool `json:"rotate"`
		ConfirmRemoval bool `json:"confirmRemoval"`
	}
	if c.ShouldBindJSON(&b) != nil {
		respondError(c, bad("Invalid sharing options"))
		return
	}
	token := ""
	removed := 0
	e := s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.permissionScope(tx, c, u.ID)
		if e != nil {
			return e
		}
		stops := []Stop{}
		if stop != nil {
			stop.SharingEnabled = b.Enabled
			if b.Rotate {
				stop.ShareToken = newToken()
			}
			token = stop.ShareToken
			if e = tx.Save(stop).Error; e != nil {
				return e
			}
			stops = append(stops, *stop)
		} else {
			cal.SharingEnabled = b.Enabled
			if b.Rotate {
				cal.ShareToken = newToken()
			}
			token = cal.ShareToken
			if e = tx.Save(cal).Error; e != nil {
				return e
			}
			if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("calendar_id = ?", cal.ID).Find(&stops).Error; e != nil {
				return e
			}
		}
		for i := range stops {
			n, e := s.reconcileParticipationAccess(tx, &stops[i], true)
			if e != nil {
				return e
			}
			removed += n
			if e = s.reconcileGuestAccess(tx, &stops[i]); e != nil {
				return e
			}
		}
		if removed > 0 && !b.ConfirmRemoval {
			return errPreview
		}
		return nil
	})
	if errors.Is(e, errPreview) {
		c.JSON(409, gin.H{"error": "Confirm ending participation for people losing access", "removedParticipationCount": removed})
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"shareToken": token, "sharingEnabled": b.Enabled, "removedParticipationCount": removed})
}
func (s *Service) collaboratorScope(tx *gorm.DB, c *gin.Context, userID uuid.UUID) (*Calendar, *Stop, error) {
	cal, stop, e := s.permissionScope(tx, c, userID)
	if e != nil {
		return nil, nil, e
	}
	if stop != nil {
		var owner Calendar
		if e = tx.First(&owner, "id = ?", stop.CalendarID).Error; e != nil {
			return nil, nil, e
		}
		if owner.OwnerID != userID {
			return nil, nil, errForbidden
		}
	}
	return cal, stop, nil
}
func (s *Service) addCollaborator(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		UserID               uuid.UUID `json:"userId"`
		CanManagePermissions bool      `json:"canManagePermissions"`
	}
	if c.ShouldBindJSON(&b) != nil || b.UserID == uuid.Nil || b.UserID == u.ID {
		respondError(c, bad("Choose another claimed account"))
		return
	}
	var result Collaborator
	e := s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.collaboratorScope(tx, c, u.ID)
		if e != nil {
			return e
		}
		scope, id := scopeWhere(cal, stop)
		var n int64
		if tx.Table("users").Where("id = ?", b.UserID).Count(&n).Error != nil || n != 1 {
			return bad("Choose an existing account")
		}
		e = tx.Where(scope+" = ? AND user_id = ?", id, b.UserID).First(&result).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			result = Collaborator{ID: uuid.New(), UserID: b.UserID, CanManagePermissions: b.CanManagePermissions, Accepted: stop == nil}
			if stop != nil {
				result.StopID = &id
			} else {
				result.CalendarID = &id
			}
			return tx.Create(&result).Error
		}
		if e != nil {
			return e
		}
		result.CanManagePermissions = b.CanManagePermissions
		return tx.Save(&result).Error
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, result)
}
func (s *Service) removeCollaborator(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	collabID, e := uuid.Parse(c.Param("collaboratorId"))
	if e != nil {
		respondError(c, errUnavailable)
		return
	}
	removed := 0
	preview := c.Request.Method == "POST"
	e = s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.collaboratorScope(tx, c, u.ID)
		if e != nil {
			return e
		}
		scope, id := scopeWhere(cal, stop)
		if e = tx.Where("id = ? AND "+scope+" = ?", collabID, id).Delete(&Collaborator{}).Error; e != nil {
			return e
		}
		stops := []Stop{}
		if stop != nil {
			stops = append(stops, *stop)
		} else if e = tx.Where("calendar_id = ?", cal.ID).Find(&stops).Error; e != nil {
			return e
		}
		for i := range stops {
			n, e := s.reconcileParticipationAccess(tx, &stops[i], true)
			if e != nil {
				return e
			}
			removed += n
			if e = s.reconcileGuestAccess(tx, &stops[i]); e != nil {
				return e
			}
		}
		if preview || (removed > 0 && c.Query("confirmRemoval") != "true") {
			return errPreview
		}
		return nil
	})
	if errors.Is(e, errPreview) {
		if preview {
			c.JSON(200, gin.H{"removedParticipationCount": removed})
		} else {
			c.JSON(409, gin.H{"error": "Confirm ending participation for people losing management access", "removedParticipationCount": removed})
		}
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"removed": true})
}
func (s *Service) acceptCollaborator(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("collaboratorId"))
	if e != nil {
		respondError(c, errUnavailable)
		return
	}
	r := s.db.Model(&Collaborator{}).Where("id = ? AND user_id = ?", id, u.ID).Update("accepted", true)
	if r.Error != nil {
		respondError(c, r.Error)
		return
	}
	if r.RowsAffected == 0 {
		respondError(c, errUnavailable)
		return
	}
	c.JSON(200, gin.H{"accepted": true})
}

func normalizedStatus(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// previewStop evaluates the intended published policy without inheriting the
// requesting manager's privileges. Only managers can inspect draft content.
func (s *Service) previewStop(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	var b struct {
		Audience string     `json:"audience"`
		Email    string     `json:"email"`
		UserID   *uuid.UUID `json:"userId"`
	}
	if c.ShouldBindJSON(&b) != nil || (b.Audience != "anonymous" && b.Audience != "recipient") {
		respondError(c, bad("Choose anonymous or recipient preview"))
		return
	}
	var result gin.H
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: u.ID}) {
			return errForbidden
		}
		v := Viewer{CalendarLinkID: stop.CalendarID, StopLinkID: stop.ID}
		if b.Audience == "recipient" {
			if b.UserID != nil {
				v.UserID = *b.UserID
			} else {
				address, e := normalizedGuestEmail(b.Email)
				if e != nil {
					return bad("Provide a valid preview email")
				}
				var recipient GuestRecipient
				if tx.Where("email = ?", address).First(&recipient).Error == nil {
					v.RecipientID = recipient.ID
				}
			}
		}
		preview := *stop
		draft := stop.Draft
		preview.Published = &draft
		allowed := s.canViewStop(tx, &preview, v)
		result = gin.H{"canView": allowed, "draft": true}
		if allowed {
			result["stop"] = gin.H{"id": stop.ID, "calendarId": stop.CalendarID, "title": draft.Title, "destination": draft.Destination, "startDate": draft.StartDate, "endDate": draft.EndDate, "datePrecision": draft.DatePrecision, "timeZone": effectiveTimeZone(draft.TimeZone), "description": draft.Description, "status": draft.Status, "joinOpen": draft.JoinOpen, "published": true, "canManage": false, "canManagePermissions": false}
		}
		return nil
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, result)
}
