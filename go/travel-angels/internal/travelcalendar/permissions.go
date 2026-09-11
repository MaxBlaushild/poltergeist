package travelcalendar

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// viewPolicy is deliberately independent of transport and database queries so
// every representation can use the same replacement (not additive) semantics.
type viewPolicy struct {
	Published          bool
	Manager            bool
	Visibility         string
	CalendarVisibility string
	StopGrant          bool
	CalendarGrant      bool
	HasEnabledLink     bool
}

func allowsView(p viewPolicy) bool {
	if p.Manager {
		return true
	}
	if !p.Published {
		return false
	}
	visibility := p.Visibility
	grant := p.StopGrant
	if visibility == "inherit" {
		visibility = p.CalendarVisibility
		grant = p.CalendarGrant
	}
	switch visibility {
	case "specific":
		return grant
	case "link":
		return p.HasEnabledLink
	default:
		return false
	}
}
func (s *Service) canManageStop(tx *gorm.DB, stop *Stop, v Viewer) bool {
	if v.UserID == uuid.Nil {
		return false
	}
	var cal Calendar
	if tx.First(&cal, "id = ?", stop.CalendarID).Error != nil {
		return false
	}
	if cal.OwnerID == v.UserID {
		return true
	}
	var n int64
	if tx.Model(&Collaborator{}).Where("user_id = ? AND accepted = ? AND (calendar_id = ? OR stop_id = ?)", v.UserID, true, stop.CalendarID, stop.ID).Count(&n).Error != nil {
		return false
	}
	return n > 0
}
func (s *Service) canManagePermissions(tx *gorm.DB, stop *Stop, v Viewer) bool {
	if v.UserID == uuid.Nil {
		return false
	}
	var cal Calendar
	if tx.First(&cal, "id = ?", stop.CalendarID).Error != nil {
		return false
	}
	if cal.OwnerID == v.UserID {
		return true
	}
	var n int64
	if tx.Model(&Collaborator{}).Where("user_id = ? AND accepted = ? AND can_manage_permissions = ? AND (calendar_id = ? OR stop_id = ?)", v.UserID, true, true, stop.CalendarID, stop.ID).Count(&n).Error != nil {
		return false
	}
	return n > 0
}
func hasGrant(tx *gorm.DB, scope string, id uuid.UUID, v Viewer) bool {
	if v.UserID == uuid.Nil && v.RecipientID == uuid.Nil {
		return false
	}
	var n int64
	q := tx.Model(&Grant{}).Where(scope+" = ?", id)
	if v.UserID != uuid.Nil && v.RecipientID != uuid.Nil {
		q = q.Where("(user_id = ? OR recipient_id = ? OR recipient_id IN (SELECT id FROM tc_guest_recipients WHERE user_id = ?))", v.UserID, v.RecipientID, v.UserID)
	} else if v.UserID != uuid.Nil {
		q = q.Where("(user_id = ? OR recipient_id IN (SELECT id FROM tc_guest_recipients WHERE user_id = ?))", v.UserID, v.UserID)
	} else {
		q = q.Where("recipient_id = ?", v.RecipientID)
	}
	return q.Count(&n).Error == nil && n > 0
}
func (s *Service) canViewStop(tx *gorm.DB, stop *Stop, v Viewer) bool {
	var cal Calendar
	if tx.First(&cal, "id = ?", stop.CalendarID).Error != nil {
		return false
	}
	return allowsView(viewPolicy{Published: stop.Published != nil, Manager: s.canManageStop(tx, stop, v), Visibility: stop.Visibility, CalendarVisibility: cal.Visibility, StopGrant: hasGrant(tx, "stop_id", stop.ID, v), CalendarGrant: hasGrant(tx, "calendar_id", cal.ID, v), HasEnabledLink: (v.StopLinkID == stop.ID && stop.SharingEnabled) || (v.CalendarLinkID == cal.ID && cal.SharingEnabled)})
}
func activeParticipation(status string) bool {
	return status == "interested" || status == "requested" || status == "joined" || status == "needs_reconfirmation"
}
