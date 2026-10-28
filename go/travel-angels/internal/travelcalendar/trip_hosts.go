package travelcalendar

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The confirmation binds the source revision to both calendars' actual audience,
// because editing the destination calendar does not change the trip revision.
func (s *Service) transferAudience(tx *gorm.DB, stop *Stop, old, target *Calendar) (string, gin.H, error) {
	type subject struct {
		Key    string    `json:"key"`
		UserID uuid.UUID `json:"userId"`
	}
	read := func(cal *Calendar) ([]subject, error) {
		var grants []Grant
		if err := tx.Where("calendar_id = ?", cal.ID).Find(&grants).Error; err != nil {
			return nil, err
		}
		result := []subject{}
		for _, grant := range grants {
			if grant.UserID != nil {
				result = append(result, subject{"user:" + grant.UserID.String(), *grant.UserID})
				continue
			}
			if grant.RecipientID != nil {
				var recipient GuestRecipient
				if err := tx.First(&recipient, "id = ?", *grant.RecipientID).Error; err != nil {
					return nil, err
				}
				id := uuid.Nil
				if recipient.UserID != nil {
					id = *recipient.UserID
				}
				result = append(result, subject{"recipient:" + grant.RecipientID.String(), id})
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
		return result, nil
	}
	before, err := read(old)
	if err != nil {
		return "", nil, err
	}
	after, err := read(target)
	if err != nil {
		return "", nil, err
	}
	var hosts []Collaborator
	if err = tx.Where("stop_id = ? AND accepted = ? AND is_host = ?", stop.ID, true, true).Order("user_id").Find(&hosts).Error; err != nil {
		return "", nil, err
	}
	stable := map[uuid.UUID]bool{old.OwnerID: true, target.OwnerID: true}
	for _, host := range hosts {
		stable[host.UserID] = true
	}
	set := func(cal *Calendar, subjects []subject) map[string]subject {
		out := map[string]subject{}
		if stop.Visibility != "inherit" || cal.Visibility != "specific" {
			return out
		}
		for _, item := range subjects {
			if stable[item.UserID] {
				continue
			}
			key := item.Key
			if item.UserID != uuid.Nil {
				key = "user:" + item.UserID.String()
			}
			out[key] = item
		}
		return out
	}
	oldSet, newSet := set(old, before), set(target, after)
	difference := func(a, b map[string]subject) []gin.H {
		keys := []string{}
		for key := range a {
			if _, found := b[key]; !found {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		rows := []gin.H{}
		for _, key := range keys {
			item := a[key]
			// Moving a trip must not reveal the new owner's unrelated contact emails.
			row := gin.H{"subjectType": "invited_contact", "displayName": "Invited contact"}
			if item.UserID != uuid.Nil {
				row = s.tripSafeIdentity(tx, item.UserID)
				row["subjectType"] = "account"
			}
			rows = append(rows, row)
		}
		return rows
	}
	added, removed := difference(newSet, oldSet), difference(oldSet, newSet)
	fromVisibility, toVisibility := stop.Visibility, stop.Visibility
	if stop.Visibility == "inherit" {
		fromVisibility = old.Visibility
		toVisibility = target.Visibility
	}
	linkChanged := (fromVisibility == "link" && old.SharingEnabled) || (toVisibility == "link" && target.SharingEnabled)
	var participants []Participation
	if err = tx.Where("stop_id = ?", stop.ID).Order("id").Find(&participants).Error; err != nil {
		return "", nil, err
	}
	snapshot := gin.H{"stopId": stop.ID, "revision": stop.Revision, "visibility": stop.Visibility, "oldCalendar": old, "newCalendar": target, "oldGrants": before, "newGrants": after, "hosts": hosts, "participations": participants}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "", nil, err
	}
	return guestDigest(string(encoded)), gin.H{"fromVisibility": fromVisibility, "toVisibility": toVisibility, "inheritanceChanged": stop.Visibility == "inherit", "oldCalendarId": old.ID, "newCalendarId": target.ID, "addedViewers": added, "removedViewers": removed, "addedViewerCount": len(added), "removedViewerCount": len(removed), "linkAccessChanged": linkChanged}, nil
}

type tripConversionInput struct {
	Revision    *int        `json:"revision"`
	HostUserIDs []uuid.UUID `json:"hostUserIds"`
	Confirm     bool        `json:"confirm"`
}

func (s *Service) conversionCandidates(tx *gorm.DB, stop *Stop) ([]gin.H, error) {
	var roles []Collaborator
	if e := tx.Where("calendar_id = ? OR stop_id = ?", stop.CalendarID, stop.ID).Order("created_at").Find(&roles).Error; e != nil {
		return nil, e
	}
	owner, e := s.tripOwner(tx, stop)
	if e != nil {
		return nil, e
	}
	rows := []gin.H{}
	seen := map[uuid.UUID]int{}
	for _, role := range roles {
		if role.UserID == owner {
			continue
		}
		source := "stop collaborator"
		if role.CalendarID != nil {
			source = "calendar editor"
		}
		if index, ok := seen[role.UserID]; ok {
			rows[index]["source"] = rows[index]["source"].(string) + ", " + source
			if role.Accepted {
				rows[index]["accepted"] = true
			}
			continue
		}
		row := s.tripSafeIdentity(tx, role.UserID)
		row["source"] = source
		row["accepted"] = role.Accepted
		row["canManagePermissions"] = role.CanManagePermissions
		seen[role.UserID] = len(rows)
		rows = append(rows, row)
	}
	return rows, nil
}
func (s *Service) getHostConversion(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var response gin.H
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		owner, e := s.tripOwner(tx, stop)
		if e != nil {
			return e
		}
		if owner != user.ID {
			return errForbidden
		}
		candidates, e := s.conversionCandidates(tx, stop)
		if e != nil {
			return e
		}
		response = gin.H{"hostModel": stop.HostModel, "revision": stop.Revision, "candidates": candidates}
		return nil
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, response)
}
func (s *Service) convertTripHosts(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body tripConversionInput
	if c.ShouldBindJSON(&body) != nil || len(body.HostUserIDs) > 100 {
		respondError(c, bad("Choose up to 100 named hosts"))
		return
	}
	preview := strings.HasSuffix(c.FullPath(), "/preview")
	response := gin.H{}
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
		if user.ID != owner {
			return errForbidden
		}
		if stop.HostModel {
			return bad("This trip already uses full host permissions")
		}
		if body.Revision == nil || *body.Revision != stop.Revision {
			return tripConflict{"Reload the conversion preview before continuing", stop.Revision}
		}
		candidates, e := s.conversionCandidates(tx, stop)
		if e != nil {
			return e
		}
		selected := map[uuid.UUID]bool{}
		for _, id := range body.HostUserIDs {
			if id == uuid.Nil || id == owner {
				return bad("Choose other existing accounts as additional hosts")
			}
			var n int64
			if e = tx.Table("users").Where("id = ?", id).Count(&n).Error; e != nil {
				return e
			}
			if n != 1 {
				return bad("Choose an existing account")
			}
			selected[id] = true
		}
		losing := []gin.H{}
		for _, candidate := range candidates {
			if accepted, _ := candidate["accepted"].(bool); accepted {
				candidate["pendingHostInvitation"] = selected[candidate["userId"].(uuid.UUID)]
				losing = append(losing, candidate)
			}
		}
		// The explicit activation replaces old stop roles. Calendar-wide grants stay
		// untouched for every other trip; selected people receive fresh invitations.
		if e = tx.Where("stop_id = ?", stop.ID).Delete(&Collaborator{}).Error; e != nil {
			return e
		}
		for id := range selected {
			sid := stop.ID
			if e = tx.Create(&Collaborator{ID: uuid.New(), StopID: &sid, UserID: id, IsHost: true, CanManagePermissions: true, Accepted: false}).Error; e != nil {
				return e
			}
		}
		stop.HostModel = true
		if e = tx.Save(stop).Error; e != nil {
			return e
		}
		removed, e := s.reconcileParticipationAccess(tx, stop, !preview)
		if e != nil {
			return e
		}
		response = gin.H{"revision": stop.Revision, "losingManagement": losing, "removedParticipationCount": removed, "message": "Activating this model ends the listed legacy editing access. Selected people receive full host invitations and gain authority only after accepting."}
		if preview || !body.Confirm {
			return errPreview
		}
		if e = s.reconcileGuestAccess(tx, stop); e != nil {
			return e
		}
		return s.saveTripChange(tx, stop, user.ID, "host_model_converted", gin.H{"invitedHostIds": body.HostUserIDs, "removedParticipationCount": removed})
	})
	if errors.Is(e, errPreview) {
		if preview {
			c.JSON(200, response)
		} else {
			response["error"] = "Review and explicitly confirm the host conversion"
			c.JSON(409, response)
		}
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	response["revision"] = stop.Revision
	response["stop"] = s.stopDTO(stop, Viewer{UserID: user.ID})
	c.JSON(200, response)
}
func (s *Service) transferTripOwnership(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision            *int      `json:"revision"`
		UserID              uuid.UUID `json:"userId"`
		Confirm             bool      `json:"confirm"`
		AudienceFingerprint string    `json:"audienceFingerprint"`
	}
	if c.ShouldBindJSON(&body) != nil || body.UserID == uuid.Nil || body.UserID == user.ID {
		respondError(c, bad("Choose an accepted host as the next owner"))
		return
	}
	preview := strings.HasSuffix(c.FullPath(), "/preview")
	response := gin.H{}
	var stop Stop
	// Initialize the target calendar only after checking the caller's authority.
	// Persisting it before the preview transaction makes its ID stable across the
	// rolled-back preview and the explicitly confirmed transfer.
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, errUnavailable)
		return
	}
	if err = s.db.First(&stop, "id = ?", id).Error; err != nil {
		respondError(c, err)
		return
	}
	owner, err := s.tripOwner(s.db, &stop)
	if err != nil {
		respondError(c, err)
		return
	}
	if owner != user.ID {
		respondError(c, errForbidden)
		return
	}
	var targetHost Collaborator
	if !stop.HostModel || s.db.Where("stop_id = ? AND user_id = ? AND accepted = ? AND is_host = ?", stop.ID, body.UserID, true, true).First(&targetHost).Error != nil {
		respondError(c, bad("The new owner must be an accepted trip host"))
		return
	}
	target, err := s.primaryCalendar(s.db, body.UserID)
	if err != nil {
		respondError(c, err)
		return
	}
	e := s.db.Transaction(func(tx *gorm.DB) error {
		id, e := uuid.Parse(c.Param("id"))
		if e != nil {
			return errUnavailable
		}
		if e = tx.First(&stop, "id = ?", id).Error; e != nil {
			return e
		}
		initialCalendar := stop.CalendarID
		// Lock both calendars in one stable order before the trip, avoiding opposite
		// transfers deadlocking on the two calendars.
		var calendars []Calendar
		if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", []uuid.UUID{initialCalendar, target.ID}).Order("id").Find(&calendars).Error; e != nil {
			return e
		}
		for i := range calendars {
			if calendars[i].ID == target.ID {
				target = &calendars[i]
			}
		}
		if e = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stop, "id = ?", id).Error; e != nil {
			return e
		}
		if stop.CalendarID != initialCalendar {
			return tripConflict{"The owner changed. Reload the trip", stop.Revision}
		}
		owner, e := s.tripOwner(tx, &stop)
		if e != nil {
			return e
		}
		if owner != user.ID {
			return errForbidden
		}
		if !stop.HostModel {
			return bad("Convert this trip to full hosts before transferring ownership")
		}
		if e = s.checkTripRevision(&stop, body.Revision); e != nil {
			return e
		}
		var host Collaborator
		if e = tx.Where("stop_id = ? AND user_id = ? AND accepted = ? AND is_host = ?", stop.ID, body.UserID, true, true).First(&host).Error; e != nil {
			return bad("The new owner must be an accepted trip host")
		}
		var old Calendar
		if e = tx.First(&old, "id = ?", stop.CalendarID).Error; e != nil {
			return e
		}
		fingerprint, audience, e := s.transferAudience(tx, &stop, &old, target)
		if e != nil {
			return e
		}
		if !preview && body.AudienceFingerprint != fingerprint {
			return tripConflict{"The calendar audience changed. Review a fresh ownership-transfer preview", stop.Revision}
		}
		sid := stop.ID
		oldHost := Collaborator{ID: uuid.New(), StopID: &sid, UserID: user.ID, Accepted: true, IsHost: true, CanManagePermissions: true}
		if e = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&oldHost).Error; e != nil {
			return e
		}
		if e = tx.Model(&Collaborator{}).Where("stop_id = ? AND user_id = ?", stop.ID, user.ID).Updates(map[string]any{"accepted": true, "is_host": true, "can_manage_permissions": true}).Error; e != nil {
			return e
		}
		stop.CalendarID = target.ID
		if e = tx.Save(&stop).Error; e != nil {
			return e
		}
		removed, e := s.reconcileParticipationAccess(tx, &stop, !preview)
		if e != nil {
			return e
		}
		response = gin.H{"revision": stop.Revision, "audienceFingerprint": fingerprint, "removedParticipationCount": removed, "audienceChange": audience, "message": "Named audience changes are listed; anyone holding a calendar link cannot be enumerated. Explicit trip audiences stay the same. Calendar inheritance and calendar-link access now use the new owner's calendar. The previous owner remains a host until they leave."}
		if preview || !body.Confirm {
			return errPreview
		}
		if e = s.reconcileGuestAccess(tx, &stop); e != nil {
			return e
		}
		return s.saveTripChange(tx, &stop, user.ID, "ownership_transferred", gin.H{"previousOwnerId": user.ID, "ownerId": body.UserID, "removedParticipationCount": removed})
	})
	if errors.Is(e, errPreview) {
		if preview {
			c.JSON(200, response)
		} else {
			response["error"] = "Review and confirm the ownership and audience changes"
			c.JSON(409, response)
		}
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	response["revision"] = stop.Revision
	response["stop"] = s.stopDTO(&stop, Viewer{UserID: user.ID})
	c.JSON(200, response)
}
func (s *Service) leaveTripHost(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision       *int `json:"revision"`
		ConfirmRemoval bool `json:"confirmRemoval"`
	}
	if e := optionalTripBody(c, &body); e != nil {
		respondError(c, e)
		return
	}
	removed := 0
	revision := 0
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		owner, e := s.tripOwner(tx, stop)
		if e != nil {
			return e
		}
		if owner == user.ID {
			return bad("Transfer ownership to an accepted host before leaving")
		}
		if !s.canManageStop(tx, stop, Viewer{UserID: user.ID}) {
			return errForbidden
		}
		if e = s.checkTripRevision(stop, body.Revision); e != nil {
			return e
		}
		result := tx.Where("stop_id = ? AND user_id = ?", stop.ID, user.ID).Delete(&Collaborator{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return bad("Calendar-wide editing access must be removed by the calendar owner")
		}
		removed, e = s.reconcileParticipationAccess(tx, stop, true)
		if e != nil {
			return e
		}
		if removed > 0 && !body.ConfirmRemoval {
			return errPreview
		}
		if e = s.reconcileGuestAccess(tx, stop); e != nil {
			return e
		}
		if e = s.saveTripChange(tx, stop, user.ID, "host_left", nil); e != nil {
			return e
		}
		revision = stop.Revision
		return nil
	})
	if errors.Is(e, errPreview) {
		c.JSON(409, gin.H{"error": "Leaving hosting also ends attendance that depended on that access", "removedParticipationCount": removed})
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"left": true, "revision": revision, "removedParticipationCount": removed})
}
func (s *Service) deleteTrip(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision *int `json:"revision"`
		Confirm  bool `json:"confirm"`
	}
	if e := optionalTripBody(c, &body); e != nil {
		respondError(c, e)
		return
	}
	if !body.Confirm {
		respondError(c, bad("Explicitly confirm permanent deletion"))
		return
	}
	var obsolete []TripMedia
	var obsoleteDocuments []TripDocument
	e := s.db.Transaction(func(tx *gorm.DB) error {
		stop, e := s.lockedStop(tx, c.Param("id"))
		if e != nil {
			return e
		}
		owner, e := s.tripOwner(tx, stop)
		if e != nil {
			return e
		}
		if owner != user.ID {
			return errForbidden
		}
		if e = s.checkTripRevision(stop, mutationRevision(c, body.Revision)); e != nil {
			return e
		}
		if e = s.recordTripAudit(tx, stop.ID, user.ID, "trip_deleted", nil); e != nil {
			return e
		}
		if e = tx.Select("id, storage_bucket, storage_key").Where("stop_id = ?", stop.ID).Find(&obsolete).Error; e != nil {
			return e
		}
		if e = tx.Select("id, storage_bucket, storage_key").Where("stop_id = ? AND kind = 'file'", stop.ID).Find(&obsoleteDocuments).Error; e != nil {
			return e
		}
		return tx.Delete(stop).Error
	})
	if e != nil {
		respondError(c, e)
		return
	}
	s.deleteTripImageObjects(obsolete)
	for i := range obsoleteDocuments {
		s.deleteTripDocumentObject(&obsoleteDocuments[i])
	}
	c.JSON(200, gin.H{"deleted": true})
}

func (s *Service) addCollaborator(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision             *int      `json:"revision"`
		UserID               uuid.UUID `json:"userId"`
		CanManagePermissions bool      `json:"canManagePermissions"`
	}
	if c.ShouldBindJSON(&body) != nil || body.UserID == uuid.Nil || body.UserID == user.ID {
		respondError(c, bad("Choose another claimed account"))
		return
	}
	var result Collaborator
	revision := 0
	e := s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.collaboratorScope(tx, c, user.ID)
		if e != nil {
			return e
		}
		if stop != nil {
			if e = s.checkTripRevision(stop, body.Revision); e != nil {
				return e
			}
			owner, e := s.tripOwner(tx, stop)
			if e != nil {
				return e
			}
			if body.UserID == owner {
				return bad("The trip owner is already a host")
			}
		}
		var n int64
		if e = tx.Table("users").Where("id = ?", body.UserID).Count(&n).Error; e != nil {
			return e
		}
		if n != 1 {
			return bad("Choose an existing account")
		}
		scope, id := scopeWhere(cal, stop)
		e = tx.Where(scope+" = ? AND user_id = ?", id, body.UserID).First(&result).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			result = Collaborator{ID: uuid.New(), UserID: body.UserID, CanManagePermissions: body.CanManagePermissions, Accepted: stop == nil}
			if stop != nil {
				result.StopID = &id
				result.IsHost = stop.HostModel
				if stop.HostModel {
					result.CanManagePermissions = true
				}
			} else {
				result.CalendarID = &id
			}
			if e = tx.Create(&result).Error; e != nil {
				return e
			}
		} else if e != nil {
			return e
		} else {
			if stop != nil && stop.HostModel {
				if !result.IsHost {
					result.Accepted = false
				}
				result.IsHost = true
				result.CanManagePermissions = true
			} else {
				result.CanManagePermissions = body.CanManagePermissions
			}
			if e = tx.Save(&result).Error; e != nil {
				return e
			}
		}
		if stop != nil {
			if e = s.saveTripChange(tx, stop, user.ID, "host_invited", gin.H{"userId": body.UserID, "isHost": result.IsHost}); e != nil {
				return e
			}
			revision = stop.Revision
		}
		return nil
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"id": result.ID, "userId": result.UserID, "stopId": result.StopID, "calendarId": result.CalendarID, "accepted": result.Accepted, "isHost": result.IsHost, "canManagePermissions": result.CanManagePermissions, "revision": revision})
}
func (s *Service) removeCollaborator(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body struct {
		Revision       *int `json:"revision"`
		ConfirmRemoval bool `json:"confirmRemoval"`
	}
	if e := optionalTripBody(c, &body); e != nil {
		respondError(c, e)
		return
	}
	id, e := uuid.Parse(c.Param("collaboratorId"))
	if e != nil {
		respondError(c, errUnavailable)
		return
	}
	preview := c.Request.Method == "POST"
	removed := 0
	revision := 0
	e = s.db.Transaction(func(tx *gorm.DB) error {
		cal, stop, e := s.collaboratorScope(tx, c, user.ID)
		if e != nil {
			return e
		}
		if stop != nil {
			if e = s.checkTripRevision(stop, mutationRevision(c, body.Revision)); e != nil {
				return e
			}
			revision = stop.Revision
		}
		scope, scopeID := scopeWhere(cal, stop)
		var collaborator Collaborator
		if e = tx.Where("id = ? AND "+scope+" = ?", id, scopeID).First(&collaborator).Error; e != nil {
			return e
		}
		if stop != nil {
			owner, e := s.tripOwner(tx, stop)
			if e != nil {
				return e
			}
			if collaborator.UserID == owner {
				return bad("Transfer ownership before removing the owner")
			}
		}
		if e = tx.Delete(&collaborator).Error; e != nil {
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
		if preview || (removed > 0 && !body.ConfirmRemoval && c.Query("confirmRemoval") != "true") {
			return errPreview
		}
		if stop != nil {
			if e = s.saveTripChange(tx, stop, user.ID, "host_removed", gin.H{"userId": collaborator.UserID, "removedParticipationCount": removed}); e != nil {
				return e
			}
			revision = stop.Revision
		}
		return nil
	})
	if errors.Is(e, errPreview) {
		result := gin.H{"removedParticipationCount": removed, "revision": revision}
		if preview {
			c.JSON(200, result)
		} else {
			result["error"] = "Confirm ending attendance for people losing management access"
			c.JSON(409, result)
		}
		return
	}
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"removed": true, "revision": revision, "removedParticipationCount": removed})
}
func (s *Service) acceptCollaborator(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("collaboratorId"))
	if e != nil {
		respondError(c, errUnavailable)
		return
	}
	var body hostAttendanceInput
	if e = optionalTripBody(c, &body); e != nil {
		respondError(c, e)
		return
	}
	revision := 0
	e = s.db.Transaction(func(tx *gorm.DB) error {
		var role Collaborator
		if e := tx.Where("id = ? AND user_id = ?", id, user.ID).First(&role).Error; e != nil {
			return e
		}
		if role.StopID == nil {
			return tx.Model(&role).Update("accepted", true).Error
		}
		stop, e := s.lockedStop(tx, role.StopID.String())
		if e != nil {
			return e
		}
		if e = tx.Where("id = ? AND user_id = ?", id, user.ID).First(&role).Error; e != nil {
			return e
		}
		if stop.HostModel && !role.IsHost {
			return errUnavailable
		}
		if stop.HostModel && body.Attending == nil {
			return bad("Choose whether you are attending when accepting full host access")
		}
		if role.Accepted {
			return nil
		}
		role.Accepted = true
		if stop.HostModel {
			role.CanManagePermissions = true
		}
		if e = tx.Save(&role).Error; e != nil {
			return e
		}
		if stop.HostModel {
			if _, e = s.setHostAttendance(tx, stop, user.ID, body); e != nil {
				return e
			}
		}
		if e = s.saveTripChange(tx, stop, user.ID, "host_accepted", gin.H{"fullHost": role.IsHost}); e != nil {
			return e
		}
		revision = stop.Revision
		return nil
	})
	if e != nil {
		respondError(c, e)
		return
	}
	c.JSON(200, gin.H{"accepted": true, "revision": revision})
}
