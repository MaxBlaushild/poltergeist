package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestGroupHandle struct {
	db *gorm.DB
}

func (c *pointOfInterestGroupHandle) GetQuestsInZone(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) ([]models.PointOfInterestGroup, error) {
	pointsOfInterest := []models.PointOfInterest{}
	query := c.db.WithContext(ctx).
		Table("points_of_interest poi").
		Joins("JOIN point_of_interest_zones pz ON pz.point_of_interest_id = poi.id").
		Where("pz.zone_id = ?", zoneID).
		Distinct("poi.*")

	if len(tags) > 0 {
		query = query.
			Joins("JOIN tag_entities te ON te.point_of_interest_id = poi.id").
			Joins("JOIN tags t ON t.id = te.tag_id").
			Where("t.value IN ?", tags)
	}

	if err := query.Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	pointOfInterestIDs := make([]uuid.UUID, len(pointsOfInterest))
	for i, poi := range pointsOfInterest {
		pointOfInterestIDs[i] = poi.ID
	}

	var pointOfInterestGroupMembers []models.PointOfInterestGroupMember
	if err := c.db.WithContext(ctx).Where("point_of_interest_id IN ?", pointOfInterestIDs).Find(&pointOfInterestGroupMembers).Error; err != nil {
		return nil, err
	}

	groupIDMap := make(map[uuid.UUID]bool)
	var groupIDs []uuid.UUID
	for _, member := range pointOfInterestGroupMembers {
		if !groupIDMap[member.PointOfInterestGroupID] {
			groupIDs = append(groupIDs, member.PointOfInterestGroupID)
			groupIDMap[member.PointOfInterestGroupID] = true
		}
	}

	// Get accepted quest IDs for this user
	var acceptedQuestIDs []uuid.UUID
	if err := c.db.WithContext(ctx).
		Table("quest_acceptances").
		Where("user_id = ?", userID).
		Pluck("point_of_interest_group_id", &acceptedQuestIDs).Error; err != nil {
		return nil, err
	}
	acceptedQuestIDMap := make(map[uuid.UUID]bool)
	for _, id := range acceptedQuestIDs {
		acceptedQuestIDMap[id] = true
	}

	var groups []models.PointOfInterestGroup
	if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
		Where("id IN ?", groupIDs).
		Where("hidden = false").
		Find(&groups).Error; err != nil {
		return nil, err
	}

	// Filter out quests with quest givers that haven't been accepted
	filteredGroups := []models.PointOfInterestGroup{}
	for _, group := range groups {
		// If quest has no quest giver, include it (backward compatibility)
		if group.QuestGiverCharacterID == nil {
			filteredGroups = append(filteredGroups, group)
			continue
		}
		// If quest has a quest giver, only include if user has accepted it
		if acceptedQuestIDMap[group.ID] {
			filteredGroups = append(filteredGroups, group)
		}
	}

	return filteredGroups, nil
}

func (c *pointOfInterestGroupHandle) GetNearbyQuests(ctx context.Context, userID uuid.UUID, lat float64, lng float64, radiusInMeters float64, tags []string) ([]models.PointOfInterestGroup, error) {
	pointsOfInterest := []models.PointOfInterest{}
	query := c.db.WithContext(ctx).
		Table("points_of_interest poi").
		Where("ST_DWithin(poi.geometry, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			lng, lat, radiusInMeters).
		Distinct("poi.*")

	if len(tags) > 0 {
		query = query.
			Joins("JOIN tag_entities te ON te.point_of_interest_id = poi.id").
			Joins("JOIN tags t ON t.id = te.tag_id").
			Where("t.value IN ?", tags)
	}

	if err := query.Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	pointOfInterestIDs := make([]uuid.UUID, len(pointsOfInterest))
	for i, poi := range pointsOfInterest {
		pointOfInterestIDs[i] = poi.ID
	}

	var pointOfInterestGroupMembers []models.PointOfInterestGroupMember
	if err := c.db.WithContext(ctx).Where("point_of_interest_id IN ?", pointOfInterestIDs).Find(&pointOfInterestGroupMembers).Error; err != nil {
		return nil, err
	}

	groupIDMap := make(map[uuid.UUID]bool)
	var groupIDs []uuid.UUID
	for _, member := range pointOfInterestGroupMembers {
		if !groupIDMap[member.PointOfInterestGroupID] {
			groupIDs = append(groupIDs, member.PointOfInterestGroupID)
			groupIDMap[member.PointOfInterestGroupID] = true
		}
	}

	var groups []models.PointOfInterestGroup
	if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
		Where("id IN ?", groupIDs).
		Where("hidden = false").
		Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// preloadPointOfInterestGroupRelations preloads all common relations for a PointOfInterestGroup query
func (c *pointOfInterestGroupHandle) preloadPointOfInterestGroupRelations(query *gorm.DB) *gorm.DB {
	return query.
		Preload("PointsOfInterest").
		Preload("PointsOfInterest.Characters").
		Preload("PointsOfInterest.Tags").
		Preload("PointsOfInterest.PointOfInterestChallenges").
		Preload("GroupMembers").
		Preload("GroupMembers.PointOfInterest").
		Preload("GroupMembers.PointOfInterest.Characters").
		Preload("GroupMembers.PointOfInterest.Tags").
		Preload("GroupMembers.PointOfInterest.PointOfInterestChallenges").
		Preload("GroupMembers.Children").
		Preload("QuestGiverCharacter")
}

func (c *pointOfInterestGroupHandle) FindByIDs(ctx context.Context, groupIDs []uuid.UUID) ([]models.PointOfInterestGroup, error) {
	var groups []models.PointOfInterestGroup
	if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
		Where("id IN ?", groupIDs).
		Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *pointOfInterestGroupHandle) GetStartedQuests(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestGroup, error) {
	var groups []models.PointOfInterestGroup
	query := c.db.WithContext(ctx).
		Distinct("pog.*").
		Table("point_of_interest_groups pog").
		Joins("JOIN point_of_interest_group_members pogm ON pogm.point_of_interest_group_id = pog.id").
		Joins("JOIN point_of_interest_challenges poc ON poc.point_of_interest_id = pogm.point_of_interest_id").
		Joins("JOIN point_of_interest_challenge_submissions pocs ON pocs.point_of_interest_challenge_id = poc.id").
		Where("pocs.user_id = ?", userID).
		Where("pog.type = ?", models.PointOfInterestGroupTypeQuest)

	if err := c.preloadPointOfInterestGroupRelations(query).Find(&groups).Error; err != nil {
		return nil, err
	}

	// Also include accepted quests with quest givers
	var acceptedQuests []models.PointOfInterestGroup
	acceptedQuestIDs := []uuid.UUID{}
	if err := c.db.WithContext(ctx).
		Table("quest_acceptances").
		Where("user_id = ?", userID).
		Pluck("point_of_interest_group_id", &acceptedQuestIDs).Error; err != nil {
		return nil, err
	}

	if len(acceptedQuestIDs) > 0 {
		if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
			Where("id IN ?", acceptedQuestIDs).
			Where("type = ?", models.PointOfInterestGroupTypeQuest).
			Find(&acceptedQuests).Error; err != nil {
			return nil, err
		}
	}

	// Merge and deduplicate
	groupMap := make(map[uuid.UUID]models.PointOfInterestGroup)
	for _, group := range groups {
		groupMap[group.ID] = group
	}
	for _, group := range acceptedQuests {
		groupMap[group.ID] = group
	}

	result := []models.PointOfInterestGroup{}
	for _, group := range groupMap {
		result = append(result, group)
	}

	return result, nil
}

func (c *pointOfInterestGroupHandle) Delete(ctx context.Context, id uuid.UUID) error {
	// Start a transaction since we'll be deleting multiple related records
	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}

	// First delete all child records associated with the group members
	if err := tx.Where("point_of_interest_group_member_id IN (SELECT id FROM point_of_interest_group_members WHERE point_of_interest_group_id = ?)", id).
		Delete(&models.PointOfInterestChildren{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("point_of_interest_group_id = ?", id).Delete(&models.PointOfInterestChallenge{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete related PointOfInterestGroupMember records
	if err := tx.Where("point_of_interest_group_id = ?", id).Delete(&models.PointOfInterestGroupMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the PointOfInterestGroup itself
	if err := tx.Delete(&models.PointOfInterestGroup{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (c *pointOfInterestGroupHandle) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	// Start a transaction since we'll be deleting multiple related records
	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return err
	}

	// First delete all child records associated with the group members
	if err := tx.Where("point_of_interest_group_member_id IN (SELECT id FROM point_of_interest_group_members WHERE point_of_interest_group_id IN ?)", ids).
		Delete(&models.PointOfInterestChildren{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("point_of_interest_group_id IN ?", ids).Delete(&models.PointOfInterestChallenge{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete related PointOfInterestGroupMember records
	if err := tx.Where("point_of_interest_group_id IN ?", ids).Delete(&models.PointOfInterestGroupMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the PointOfInterestGroups themselves
	if err := tx.Where("id IN ?", ids).Delete(&models.PointOfInterestGroup{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (c *pointOfInterestGroupHandle) UpdateImageUrl(ctx context.Context, id uuid.UUID, imageUrl string) error {
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", id).Updates(map[string]interface{}{
		"image_url":  imageUrl,
		"updated_at": time.Now(),
	}).Error
}

func (c *pointOfInterestGroupHandle) Edit(ctx context.Context, id uuid.UUID, name string, description string, typeValue models.PointOfInterestGroupType) error {
	group := models.PointOfInterestGroup{
		Name:        name,
		Description: description,
		Type:        typeValue,
		UpdatedAt:   time.Now(),
	}
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", id).Updates(group).Error
}

func (c *pointOfInterestGroupHandle) Create(ctx context.Context, name string, description string, imageUrl string, typeValue models.PointOfInterestGroupType) (*models.PointOfInterestGroup, error) {

	pointOfInterestGroup := models.PointOfInterestGroup{
		Name:        name,
		Description: description,
		ImageUrl:    imageUrl,
		Type:        typeValue,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := c.db.Create(&pointOfInterestGroup).Error; err != nil {
		return nil, err
	}

	return &pointOfInterestGroup, nil
}

func (c *pointOfInterestGroupHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error) {
	var pointOfInterestGroup models.PointOfInterestGroup
	if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
		First(&pointOfInterestGroup, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pointOfInterestGroup, nil
}

func (c *pointOfInterestGroupHandle) FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error) {
	var pointOfInterestGroups []*models.PointOfInterestGroup
	if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
		Find(&pointOfInterestGroups).Error; err != nil {
		return nil, err
	}
	return pointOfInterestGroups, nil
}

func (c *pointOfInterestGroupHandle) FindByType(ctx context.Context, typeValue models.PointOfInterestGroupType) ([]*models.PointOfInterestGroup, error) {
	var pointOfInterestGroups []*models.PointOfInterestGroup
	if err := c.preloadPointOfInterestGroupRelations(c.db.WithContext(ctx)).
		Where("type = ?", typeValue).
		Find(&pointOfInterestGroups).Error; err != nil {
		return nil, err
	}
	return pointOfInterestGroups, nil
}

func (c *pointOfInterestGroupHandle) AddMember(ctx context.Context, pointOfInterestID uuid.UUID, pointOfInterestGroupID uuid.UUID) (*models.PointOfInterestGroupMember, error) {
	member := models.PointOfInterestGroupMember{
		PointOfInterestID:      pointOfInterestID,
		PointOfInterestGroupID: pointOfInterestGroupID,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		ID:                     uuid.New(),
	}

	if err := c.db.Create(&member).Error; err != nil {
		return nil, err
	}

	return &member, nil
}

func (c *pointOfInterestGroupHandle) Update(ctx context.Context, pointOfInterestGroupID uuid.UUID, updates *models.PointOfInterestGroup) error {
	return c.db.Model(&models.PointOfInterestGroup{}).Where("id = ?", pointOfInterestGroupID).Updates(updates).Error
}
