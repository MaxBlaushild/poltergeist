package travelcalendar

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LibraryTrip provides only a safe personal-reference label. It reuses the
// calendar policy and any still-valid link previously observed during RSVP.
func (s *Service) LibraryTrip(tx *gorm.DB, userID, tripID uuid.UUID) (map[string]interface{}, error) {
	var stop Stop
	if err := tx.First(&stop, "id = ?", tripID).Error; err != nil {
		return nil, err
	}
	v := s.participationViewer(tx, &stop, userID, participationInput{})
	if !s.canViewStop(tx, &stop, v) {
		return nil, gorm.ErrRecordNotFound
	}
	content := &stop.Draft
	if stop.Published != nil {
		content = stop.Published
	}
	return map[string]interface{}{"id": stop.ID.String(), "title": content.Title, "destination": content.Destination, "destinations": contentDestinations(content), "startDate": content.StartDate, "endDate": content.EndDate}, nil
}

// The picker mirrors personal calendar references, not every discoverable plan.
func (s *Service) LibraryTrips(tx *gorm.DB, userID uuid.UUID) ([]map[string]interface{}, error) {
	stops := []Stop{}
	err := tx.Where(`calendar_id IN (SELECT id FROM tc_calendars WHERE owner_id = ?) OR
	 id IN (SELECT stop_id FROM tc_collaborators WHERE user_id = ? AND accepted = true) OR
	 calendar_id IN (SELECT calendar_id FROM tc_collaborators WHERE user_id = ? AND accepted = true) OR
	 id IN (SELECT stop_id FROM tc_participations WHERE user_id = ? AND (status IN ('joined','needs_reconfirmation') OR reserved = true))`, userID, userID, userID, userID).
		Order("created_at DESC, id").Find(&stops).Error
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for _, stop := range stops {
		row, err := s.LibraryTrip(tx, userID, stop.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}
