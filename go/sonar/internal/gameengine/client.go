package gameengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/google/uuid"
	"github.com/paulmach/orb"
)

const (
	BaseReputationPointsAwardedForSuccessfulSubmission = 100
	BaseExperiencePointsAwardedForSuccessfulSubmission = 100
	BaseExperiencePointsAwardedForFinishedQuest        = 250
	BaseReputationPointsAwardedForFinishedQuest        = 250
)

type Submission struct {
	ChallengeID uuid.UUID
	TeamID      *uuid.UUID
	UserID      *uuid.UUID
	ImageURL    string
	Text        string
}

type SubmissionResult struct {
	Successful     bool   `json:"successful"`
	Reason         string `json:"reason"`
	QuestCompleted bool   `json:"questCompleted"`
}

type GameEngineClient interface {
	ProcessSuccessfulSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*SubmissionResult, error)
	ProcessSubmission(ctx context.Context, submission Submission) (*SubmissionResult, error)
	AwardQuestTurnInRewards(ctx context.Context, userID uuid.UUID, pointOfInterestGroupID uuid.UUID, teamID *uuid.UUID) (goldAwarded int, itemsAwarded []models.ItemAwarded, err error)
	AwardQuestNodeSubmissionRewards(ctx context.Context, userID uuid.UUID, teamID *uuid.UUID, quest *models.Quest, node *models.QuestNode, challenge *models.QuestNodeChallenge, questCompleted bool) error
}

type gameEngineClient struct {
	db             db.DbClient
	judge          judge.Client
	quartermaster  quartermaster.Quartermaster
	chatClient     chat.Client
	livenessClient liveness.LivenessClient
}

func NewGameEngineClient(
	db db.DbClient,
	judge judge.Client,
	quartermaster quartermaster.Quartermaster,
	chatClient chat.Client,
	livenessClient liveness.LivenessClient,
) GameEngineClient {
	return &gameEngineClient{db: db, judge: judge, quartermaster: quartermaster, chatClient: chatClient, livenessClient: livenessClient}
}

// isUserInZone checks if the given coordinates are within the zone's polygon boundary
func (c *gameEngineClient) isUserInZone(lat, lng float64, zone *models.Zone) bool {
	if zone == nil {
		log.Printf("[DEBUG] Zone is nil, returning false")
		return false
	}

	polygon := zone.GetPolygon()
	if len(polygon) == 0 {
		log.Printf("[DEBUG] Zone %s has no polygon points, returning false", zone.ID)
		return false
	}

	log.Printf("[DEBUG] Checking if point (%.6f, %.6f) is in zone %s with %d polygon rings", lat, lng, zone.ID, len(polygon))

	// Use ray-casting algorithm to check if point is inside polygon
	result := c.isPointInPolygon(orb.Point{lng, lat}, polygon)
	log.Printf("[DEBUG] Point (%.6f, %.6f) is in zone %s: %v", lat, lng, zone.ID, result)

	return result
}

// isPointInPolygon implements the ray-casting algorithm to check if a point is inside a polygon
func (c *gameEngineClient) isPointInPolygon(point orb.Point, polygon orb.Polygon) bool {
	if len(polygon) == 0 {
		log.Printf("[DEBUG] Polygon has no rings")
		return false
	}

	ring := polygon[0] // Get the outer ring
	n := len(ring)
	log.Printf("[DEBUG] Polygon has %d points in outer ring", n)

	inside := false
	intersections := 0

	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		pi := ring[i]
		pj := ring[j]

		if ((pi[1] > point[1]) != (pj[1] > point[1])) &&
			(point[0] < (pj[0]-pi[0])*(point[1]-pi[1])/(pj[1]-pi[1])+pi[0]) {
			inside = !inside
			intersections++
		}
	}

	log.Printf("[DEBUG] Ray-casting algorithm: %d intersections, point is inside: %v", intersections, inside)
	return inside
}

// getPartyMembers returns all party members if the user is in a party, otherwise returns just the user
// Filters by active status and zone location
func (c *gameEngineClient) getPartyMembers(ctx context.Context, userID *uuid.UUID, challengeZoneID uuid.UUID) ([]models.User, error) {
	if userID == nil {
		return []models.User{}, nil
	}

	user, err := c.db.User().FindByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	var allMembers []models.User
	// If user is in a party, get all party members
	if user.PartyID != nil {
		allMembers, err = c.db.User().FindPartyMembers(ctx, *userID)
		allMembers = append(allMembers, *user)
		if err != nil {
			return nil, err
		}
	} else {
		// If not in a party, just this user
		allMembers = []models.User{*user}
	}

	// Get the zone for filtering
	zone, err := c.db.Zone().FindByID(ctx, challengeZoneID)
	if err != nil {
		return nil, err
	}

	// Filter members by active status and zone location
	var filteredMembers []models.User
	log.Printf("[DEBUG] Filtering %d party members for zone %s", len(allMembers), challengeZoneID)

	for _, member := range allMembers {
		log.Printf("[DEBUG] Checking member %s (username: %s)", member.ID, *member.Username)

		// Check if user is active
		isActive, err := c.livenessClient.IsActive(ctx, member.ID)
		if err != nil {
			log.Printf("[DEBUG] Error checking if member %s is active: %v", member.ID, err)
			if member.ID == *userID {
				log.Printf("[DEBUG] Treating requestor %s as active due to liveness error", member.ID)
			} else {
				continue
			}
		}
		if !isActive && member.ID != *userID {
			log.Printf("[DEBUG] Member %s is not active, skipping", member.ID)
			continue
		}
		if isActive {
			log.Printf("[DEBUG] Member %s is active", member.ID)
		} else {
			log.Printf("[DEBUG] Member %s is not active but is requestor, continuing", member.ID)
		}

		// Get user location
		locationStr, err := c.livenessClient.GetUserLocation(ctx, member.ID)
		if err != nil {
			log.Printf("[DEBUG] Error getting location for member %s: %v", member.ID, err)
			if member.ID == *userID {
				log.Printf("[DEBUG] No location for requestor %s; skipping zone filter", member.ID)
				filteredMembers = append(filteredMembers, member)
				continue
			}
			continue
		}
		if locationStr == "" {
			log.Printf("[DEBUG] No location data for member %s, skipping", member.ID)
			if member.ID == *userID {
				log.Printf("[DEBUG] No location for requestor %s; skipping zone filter", member.ID)
				filteredMembers = append(filteredMembers, member)
			}
			continue
		}
		log.Printf("[DEBUG] Member %s location string: %s", member.ID, locationStr)

		// Parse location (format: "lat,lng,accuracy")
		parts := strings.Split(locationStr, ",")
		if len(parts) < 2 {
			log.Printf("[DEBUG] Invalid location format for member %s: %s (expected lat,lng,accuracy)", member.ID, locationStr)
			if member.ID == *userID {
				log.Printf("[DEBUG] Invalid location for requestor %s; skipping zone filter", member.ID)
				filteredMembers = append(filteredMembers, member)
			}
			continue
		}

		lat, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			log.Printf("[DEBUG] Error parsing latitude for member %s: %v (value: %s)", member.ID, err, parts[0])
			if member.ID == *userID {
				log.Printf("[DEBUG] Invalid latitude for requestor %s; skipping zone filter", member.ID)
				filteredMembers = append(filteredMembers, member)
			}
			continue
		}

		lng, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Printf("[DEBUG] Error parsing longitude for member %s: %v (value: %s)", member.ID, err, parts[1])
			if member.ID == *userID {
				log.Printf("[DEBUG] Invalid longitude for requestor %s; skipping zone filter", member.ID)
				filteredMembers = append(filteredMembers, member)
			}
			continue
		}

		log.Printf("[DEBUG] Member %s parsed coordinates: lat=%.6f, lng=%.6f", member.ID, lat, lng)

		// Check if user is in the zone
		isInZone := c.isUserInZone(lat, lng, zone)
		log.Printf("[DEBUG] Member %s is in zone %s: %v", member.ID, challengeZoneID, isInZone)

		if !isInZone {
			log.Printf("[DEBUG] Member %s is not in zone, skipping", member.ID)
			if member.ID == *userID {
				log.Printf("[DEBUG] Requestor %s not in zone; including anyway", member.ID)
				filteredMembers = append(filteredMembers, member)
			}
			continue
		}

		log.Printf("[DEBUG] Member %s passed all filters, adding to filtered list", member.ID)
		filteredMembers = append(filteredMembers, member)
	}

	log.Printf("[DEBUG] Filtered party members: %d out of %d members will receive rewards", len(filteredMembers), len(allMembers))

	return filteredMembers, nil
}

// validateUserProximity checks if the user is within 100 meters of the point of interest
func (c *gameEngineClient) validateUserProximity(ctx context.Context, userID *uuid.UUID, poiLat, poiLng string) error {
	if userID == nil {
		return fmt.Errorf("user ID is required for proximity validation")
	}

	// Get user location from Redis
	locationStr, err := c.livenessClient.GetUserLocation(ctx, *userID)
	if err != nil {
		return fmt.Errorf("unable to get user location: %w", err)
	}
	if locationStr == "" {
		return fmt.Errorf("user location not available")
	}

	// Parse user location (format: "lat,lng,accuracy")
	parts := strings.Split(locationStr, ",")
	if len(parts) < 2 {
		return fmt.Errorf("invalid location format")
	}

	userLat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude in user location: %w", err)
	}

	userLng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude in user location: %w", err)
	}

	// Parse POI coordinates
	poiLatFloat, err := strconv.ParseFloat(poiLat, 64)
	if err != nil {
		return fmt.Errorf("invalid POI latitude: %w", err)
	}

	poiLngFloat, err := strconv.ParseFloat(poiLng, 64)
	if err != nil {
		return fmt.Errorf("invalid POI longitude: %w", err)
	}

	// Calculate distance using Haversine formula
	distance := c.calculateDistance(userLat, userLng, poiLatFloat, poiLngFloat)

	// Check if user is within 100 meters
	if distance > 100 {
		return fmt.Errorf("you must be within 100 meters of the location to submit an answer. Currently %.0f meters away", distance)
	}

	return nil
}

// calculateDistance calculates the distance between two points using the Haversine formula
func (c *gameEngineClient) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371e3 // Earth's radius in meters

	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	distance := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * distance
}

func (c *gameEngineClient) ProcessSubmission(ctx context.Context, submission Submission) (*SubmissionResult, error) {
	challenge, err := c.db.PointOfInterestChallenge().FindByID(ctx, submission.ChallengeID)
	if err != nil {
		return nil, err
	}

	// Get the POI to access its coordinates
	poi, err := c.db.PointOfInterest().FindByID(ctx, challenge.PointOfInterestID)
	if err != nil {
		return nil, err
	}

	// Validate user proximity before processing the submission
	if err := c.validateUserProximity(ctx, submission.UserID, poi.Lat, poi.Lng); err != nil {
		return &SubmissionResult{
			Successful: false,
			Reason:     err.Error(),
		}, nil
	}

	judgementResult, err := c.judgeSubmission(ctx, submission, challenge)
	if err != nil {
		return nil, err
	}

	if !judgementResult.IsSuccessful() {
		return &SubmissionResult{
			Successful: false,
			Reason:     judgementResult.Judgement.Reason,
		}, nil
	}

	return c.ProcessSuccessfulSubmission(ctx, submission, challenge)
}

func (c *gameEngineClient) judgeSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*judge.JudgeSubmissionResponse, error) {
	judgementResult, err := c.judge.JudgeSubmission(ctx, judge.JudgeSubmissionRequest{
		Challenge:          challenge,
		TeamID:             submission.TeamID,
		UserID:             submission.UserID,
		ImageSubmissionUrl: submission.ImageURL,
		TextSubmission:     submission.Text,
	})
	if err != nil {
		return nil, err
	}

	return judgementResult, nil
}

func (c *gameEngineClient) ProcessSuccessfulSubmission(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) (*SubmissionResult, error) {
	questCompleted, err := c.HasCompletedQuest(ctx, challenge)
	if err != nil {
		return nil, err
	}

	submissionResult := SubmissionResult{
		QuestCompleted: questCompleted,
		Successful:     true,
		Reason:         "Challenge completed successfully!",
	}

	// Calculate experience points
	experiencePoints := BaseExperiencePointsAwardedForSuccessfulSubmission
	if questCompleted {
		experiencePoints += BaseExperiencePointsAwardedForFinishedQuest
	}

	// Calculate reputation points
	reputationPoints := BaseReputationPointsAwardedForSuccessfulSubmission
	if questCompleted {
		reputationPoints += BaseReputationPointsAwardedForFinishedQuest
	}

	// Get zone information
	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return nil, err
	}
	fullZone, err := c.db.Zone().FindByID(ctx, zone.ZoneID)
	if err != nil {
		return nil, err
	}

	// Get quest/POI group information
	var questID uuid.UUID
	var questName string
	if challenge.PointOfInterestGroupID != nil {
		questGroup, err := c.db.PointOfInterestGroup().FindByID(ctx, *challenge.PointOfInterestGroupID)
		if err != nil {
			return nil, err
		}
		questID = questGroup.ID
		questName = questGroup.Name
	} else {
		// Fallback if no group
		questID = challenge.ID
		questName = "Quest"
	}

	// Get current POI information
	currentPOI, err := c.db.PointOfInterest().FindByID(ctx, challenge.PointOfInterestID)
	if err != nil {
		return nil, err
	}

	// Get next POI information if quest not completed
	var nextPOI *models.POIInfo
	if !questCompleted {
		children, err := c.db.PointOfInterestChallenge().GetChildrenForChallenge(ctx, challenge.ID)
		if err == nil && len(children) > 0 {
			nextGroupMember, err := c.db.PointOfInterestGroupMember().FindByID(ctx, children[0].NextPointOfInterestGroupMemberID)
			if err == nil {
				nextPointOfInterest, err := c.db.PointOfInterest().FindByID(ctx, nextGroupMember.PointOfInterestID)
				if err == nil {
					nextPOI = &models.POIInfo{
						ID:       nextPointOfInterest.ID,
						Name:     nextPointOfInterest.Name,
						ImageURL: nextPointOfInterest.ImageUrl,
					}
				}
			}
		}
	}

	// Gold and items are awarded on quest turn-in, not on challenge completion
	itemsAwarded := []models.ItemAwarded{}
	goldAwarded := 0

	// Create activity for challenge completed with full context
	challengeActivityData, err := json.Marshal(models.ChallengeCompletedActivity{
		ChallengeID:       challenge.ID,
		Successful:        true,
		Reason:            "Challenge completed successfully!",
		SubmitterID:       submission.UserID,
		ExperienceAwarded: experiencePoints,
		ReputationAwarded: reputationPoints,
		ItemsAwarded:      itemsAwarded,
		GoldAwarded:       goldAwarded,
		QuestID:           questID,
		QuestName:         questName,
		QuestCompleted:    questCompleted,
		CurrentPOI: models.POIInfo{
			ID:       currentPOI.ID,
			Name:     currentPOI.Name,
			ImageURL: currentPOI.ImageUrl,
		},
		NextPOI:  nextPOI,
		ZoneID:   fullZone.ID,
		ZoneName: fullZone.Name,
	})
	if err != nil {
		return nil, err
	}
	// Get filtered party members for activity creation
	filteredMembers, err := c.getPartyMembers(ctx, submission.UserID, zone.ZoneID)
	if err != nil {
		return nil, err
	}

	// Create activities for filtered party members only
	for _, member := range filteredMembers {
		if err := c.db.Activity().CreateActivity(ctx, models.Activity{
			UserID:       member.ID,
			ActivityType: models.ActivityTypeChallengeCompleted,
			Data:         challengeActivityData,
			Seen:         false,
		}); err != nil {
			return nil, err
		}
	}

	// QuestCompleted activity is created on turn-in, not when last challenge completes

	if err = c.awardExperiencePoints(ctx, submission, questCompleted); err != nil {
		return nil, err
	}

	if err = c.awardReputationPoints(ctx, submission, challenge, questCompleted); err != nil {
		return nil, err
	}

	// Gold and items are awarded on quest turn-in, not here

	if err := c.addTaskCompleteMessage(ctx, submission, challenge, &submissionResult); err != nil {
		return nil, err
	}

	// Auto-track quest for party members if quest is not completed and has a group
	if !questCompleted && challenge.PointOfInterestGroupID != nil {
		if err := c.trackQuestForPartyMembers(ctx, challenge, submission.UserID); err != nil {
			// Log error but don't fail the submission
			// TODO: Add proper logging here
		}
	}

	return &submissionResult, nil
}

func (c *gameEngineClient) awardGold(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) error {
	if challenge.PointOfInterestGroupID == nil {
		return nil
	}

	// Determine gold from the quest group
	group, err := c.db.PointOfInterestGroup().FindByID(ctx, *challenge.PointOfInterestGroupID)
	if err != nil {
		return err
	}

	gold := group.Gold
	if gold <= 0 {
		return nil
	}

	// Get zone for filtering party members
	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	partyMembers, err := c.getPartyMembers(ctx, submission.UserID, zone.ZoneID)
	if err != nil {
		return err
	}

	for _, member := range partyMembers {
		if err := c.db.User().AddGold(ctx, member.ID, gold); err != nil {
			return err
		}
	}

	return nil
}

func (c *gameEngineClient) awardItems(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge) error {
	if challenge.PointOfInterestGroupID == nil {
		return nil
	}

	// Get the quest group to determine items to award
	group, err := c.db.PointOfInterestGroup().FindByID(ctx, *challenge.PointOfInterestGroupID)
	if err != nil {
		return err
	}

	// Get zone for the challenge
	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	// Get filtered party members (active and in zone)
	partyMembers, err := c.getPartyMembers(ctx, submission.UserID, zone.ZoneID)
	if err != nil {
		return err
	}

	// Award items to each party member
	for _, member := range partyMembers {
		memberID := member.ID

		var item quartermaster.InventoryItem
		// Use quest's InventoryItemID if available, otherwise get random item
		if group.InventoryItemID == nil || *group.InventoryItemID == 0 {
			item, err = c.quartermaster.GetItem(ctx, submission.TeamID, &memberID)
			if err != nil {
				return err
			}
		} else {
			item, err = c.quartermaster.GetItemSpecificItem(ctx, submission.TeamID, &memberID, *group.InventoryItemID)
			if err != nil {
				return err
			}
		}

		// Create activity for item received for this specific member
		activityData, err := json.Marshal(models.ItemReceivedActivity{
			ItemID:   item.ID,
			ItemName: item.Name,
		})
		if err != nil {
			return err
		}
		if err := c.db.Activity().CreateActivity(ctx, models.Activity{
			UserID:       memberID,
			ActivityType: models.ActivityTypeItemReceived,
			Data:         activityData,
			Seen:         false,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *gameEngineClient) HasCompletedQuest(ctx context.Context, challenge *models.PointOfInterestChallenge) (bool, error) {
	children, err := c.db.PointOfInterestChallenge().GetChildrenForChallenge(ctx, challenge.ID)
	if err != nil {
		return false, err
	}
	return len(children) == 0, nil
}

func (c *gameEngineClient) addTaskCompleteMessage(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, submissionResult *SubmissionResult) error {
	if err := c.chatClient.AddCaptureMessage(ctx, submission.TeamID, submission.UserID, challenge); err != nil {
		return err
	}

	if submissionResult.QuestCompleted {
		return c.chatClient.AddCompletedQuestMessage(ctx, submission.TeamID, submission.UserID, challenge)
	}

	return nil
}

func (c *gameEngineClient) AwardQuestTurnInRewards(ctx context.Context, userID uuid.UUID, questID uuid.UUID, teamID *uuid.UUID) (goldAwarded int, itemsAwarded []models.ItemAwarded, err error) {
	quest, err := c.db.Quest().FindByID(ctx, questID)
	if err != nil {
		return 0, nil, err
	}
	if quest == nil {
		return 0, nil, fmt.Errorf("quest not found")
	}

	var zoneID uuid.UUID
	if quest.ZoneID != nil {
		zoneID = *quest.ZoneID
	} else {
		for _, node := range quest.Nodes {
			if node.PointOfInterestID == nil {
				continue
			}
			zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, *node.PointOfInterestID)
			if err == nil {
				zoneID = zone.ZoneID
				break
			}
		}
	}
	if zoneID == uuid.Nil {
		return 0, nil, fmt.Errorf("quest has no zone for reward distribution")
	}

	partyMembers, err := c.getPartyMembers(ctx, &userID, zoneID)
	if err != nil {
		return 0, nil, err
	}
	if len(partyMembers) == 0 {
		user, err := c.db.User().FindByID(ctx, userID)
		if err != nil {
			return 0, nil, err
		}
		if user != nil {
			log.Printf("[DEBUG] No eligible party members for rewards, falling back to user %s", userID)
			partyMembers = []models.User{*user}
		}
	}

	goldAwarded = quest.Gold
	if goldAwarded > 0 {
		for _, member := range partyMembers {
			if err := c.db.User().AddGold(ctx, member.ID, goldAwarded); err != nil {
				return goldAwarded, itemsAwarded, err
			}
		}
	}

	itemsAwarded = []models.ItemAwarded{}
	for _, reward := range quest.ItemRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		item := reward.InventoryItem
		if item.ID == 0 {
			itemRecord, err := c.db.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID)
			if err != nil {
				return goldAwarded, itemsAwarded, err
			}
			item = *itemRecord
		}
		for _, member := range partyMembers {
			if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, nil, &member.ID, reward.InventoryItemID, reward.Quantity); err != nil {
				return goldAwarded, itemsAwarded, err
			}
		}
		itemsAwarded = append(itemsAwarded, models.ItemAwarded{
			ID:       item.ID,
			Name:     item.Name,
			ImageURL: item.ImageURL,
			Quantity: reward.Quantity,
		})
	}

	questActivityData, err := json.Marshal(models.QuestCompletedActivity{
		QuestID:      quest.ID,
		GoldAwarded:  goldAwarded,
		ItemsAwarded: itemsAwarded,
	})
	if err != nil {
		return goldAwarded, itemsAwarded, err
	}
	for _, member := range partyMembers {
		if err := c.db.Activity().CreateActivity(ctx, models.Activity{
			UserID:       member.ID,
			ActivityType: models.ActivityTypeQuestCompleted,
			Data:         questActivityData,
			Seen:         false,
		}); err != nil {
			return goldAwarded, itemsAwarded, err
		}
	}

	return goldAwarded, itemsAwarded, nil
}

func (c *gameEngineClient) trackQuestForPartyMembers(ctx context.Context, challenge *models.PointOfInterestChallenge, userID *uuid.UUID) error {
	// Get zone for the challenge
	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	// Get filtered party members (active and in zone)
	partyMembers, err := c.getPartyMembers(ctx, userID, zone.ZoneID)
	if err != nil {
		return err
	}

	// Track quest for each party member
	for _, member := range partyMembers {
		err := c.db.TrackedPointOfInterestGroup().Create(ctx, *challenge.PointOfInterestGroupID, member.ID)
		if err != nil {
			// Log error but continue with other members
			// TODO: Add proper logging here
			continue
		}
	}

	return nil
}

func (c *gameEngineClient) awardExperiencePoints(ctx context.Context, submission Submission, questCompleted bool) error {
	// Get challenge to find zone
	challenge, err := c.db.PointOfInterestChallenge().FindByID(ctx, submission.ChallengeID)
	if err != nil {
		return err
	}

	// Get zone for the challenge
	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	// Get filtered party members (active and in zone)
	partyMembers, err := c.getPartyMembers(ctx, submission.UserID, zone.ZoneID)
	if err != nil {
		return err
	}

	experiencePoints := BaseExperiencePointsAwardedForSuccessfulSubmission
	if questCompleted {
		experiencePoints += BaseExperiencePointsAwardedForFinishedQuest
	}

	// Award experience points to each party member
	for _, member := range partyMembers {
		userLevel, err := c.db.UserLevel().ProcessExperiencePointAdditions(ctx, member.ID, experiencePoints)
		if err != nil {
			return err
		}

		// Only create level-up activity for this member if they actually leveled up
		if userLevel.LevelsGained > 0 {
			if _, err := c.db.UserCharacterStats().EnsureLevelPoints(ctx, member.ID, userLevel.Level); err != nil {
				return err
			}
			activityData, err := json.Marshal(models.LevelUpActivity{
				NewLevel: userLevel.Level,
			})
			if err != nil {
				return err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       member.ID,
				ActivityType: models.ActivityTypeLevelUp,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *gameEngineClient) awardReputationPoints(ctx context.Context, submission Submission, challenge *models.PointOfInterestChallenge, questCompleted bool) error {
	// Get zone for the challenge
	zone, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, challenge.PointOfInterestID)
	if err != nil {
		return err
	}

	// Get filtered party members (active and in zone)
	partyMembers, err := c.getPartyMembers(ctx, submission.UserID, zone.ZoneID)
	if err != nil {
		return err
	}

	reputationPoints := BaseReputationPointsAwardedForSuccessfulSubmission
	if questCompleted {
		reputationPoints += BaseReputationPointsAwardedForFinishedQuest
	}

	// Award reputation points to each party member
	for _, member := range partyMembers {
		userZoneReputation, err := c.db.UserZoneReputation().ProcessReputationPointAdditions(ctx, member.ID, zone.ZoneID, reputationPoints)
		if err != nil {
			return err
		}

		// Only create reputation-up activity for this member if they actually gained reputation levels
		if userZoneReputation.LevelsGained > 0 {
			// Get full zone details
			fullZone, err := c.db.Zone().FindByID(ctx, zone.ZoneID)
			if err != nil {
				return err
			}

			activityData, err := json.Marshal(models.ReputationUpActivity{
				NewLevel: userZoneReputation.Level,
				ZoneName: fullZone.Name,
				ZoneID:   zone.ZoneID,
			})
			if err != nil {
				return err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       member.ID,
				ActivityType: models.ActivityTypeReputationUp,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *gameEngineClient) awardExperiencePointsForZone(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, questCompleted bool) (int, error) {
	experiencePoints := BaseExperiencePointsAwardedForSuccessfulSubmission
	if questCompleted {
		experiencePoints += BaseExperiencePointsAwardedForFinishedQuest
	}

	partyMembers, err := c.getPartyMembers(ctx, &userID, zoneID)
	if err != nil {
		return 0, err
	}

	for _, member := range partyMembers {
		userLevel, err := c.db.UserLevel().ProcessExperiencePointAdditions(ctx, member.ID, experiencePoints)
		if err != nil {
			return 0, err
		}

		if userLevel.LevelsGained > 0 {
			if _, err := c.db.UserCharacterStats().EnsureLevelPoints(ctx, member.ID, userLevel.Level); err != nil {
				return 0, err
			}
			activityData, err := json.Marshal(models.LevelUpActivity{
				NewLevel: userLevel.Level,
			})
			if err != nil {
				return 0, err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       member.ID,
				ActivityType: models.ActivityTypeLevelUp,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return 0, err
			}
		}
	}

	return experiencePoints, nil
}

func (c *gameEngineClient) awardReputationPointsForZone(ctx context.Context, userID uuid.UUID, zone *models.Zone, questCompleted bool) (int, error) {
	if zone == nil {
		return 0, fmt.Errorf("zone is required")
	}
	reputationPoints := BaseReputationPointsAwardedForSuccessfulSubmission
	if questCompleted {
		reputationPoints += BaseReputationPointsAwardedForFinishedQuest
	}

	partyMembers, err := c.getPartyMembers(ctx, &userID, zone.ID)
	if err != nil {
		return 0, err
	}

	for _, member := range partyMembers {
		userZoneReputation, err := c.db.UserZoneReputation().ProcessReputationPointAdditions(ctx, member.ID, zone.ID, reputationPoints)
		if err != nil {
			return 0, err
		}

		if userZoneReputation.LevelsGained > 0 {
			activityData, err := json.Marshal(models.ReputationUpActivity{
				NewLevel: userZoneReputation.Level,
				ZoneName: zone.Name,
				ZoneID:   zone.ID,
			})
			if err != nil {
				return 0, err
			}
			if err := c.db.Activity().CreateActivity(ctx, models.Activity{
				UserID:       member.ID,
				ActivityType: models.ActivityTypeReputationUp,
				Data:         activityData,
				Seen:         false,
			}); err != nil {
				return 0, err
			}
		}
	}

	return reputationPoints, nil
}

func (c *gameEngineClient) AwardQuestNodeSubmissionRewards(ctx context.Context, userID uuid.UUID, teamID *uuid.UUID, quest *models.Quest, node *models.QuestNode, challenge *models.QuestNodeChallenge, questCompleted bool) error {
	if quest == nil || node == nil || challenge == nil {
		return fmt.Errorf("quest, node, and challenge are required")
	}

	var zone *models.Zone
	var currentPOI *models.PointOfInterest
	if node.PointOfInterestID != nil {
		poi, err := c.db.PointOfInterest().FindByID(ctx, *node.PointOfInterestID)
		if err != nil {
			return err
		}
		currentPOI = poi
		zoneInfo, err := c.db.PointOfInterest().FindZoneForPointOfInterest(ctx, poi.ID)
		if err != nil {
			return err
		}
		zone, err = c.db.Zone().FindByID(ctx, zoneInfo.ZoneID)
		if err != nil {
			return err
		}
	} else if quest.ZoneID != nil {
		z, err := c.db.Zone().FindByID(ctx, *quest.ZoneID)
		if err != nil {
			return err
		}
		zone = z
	}
	if zone == nil {
		return fmt.Errorf("quest node has no zone context")
	}

	experienceAwarded, err := c.awardExperiencePointsForZone(ctx, userID, zone.ID, questCompleted)
	if err != nil {
		return err
	}
	reputationAwarded, err := c.awardReputationPointsForZone(ctx, userID, zone, questCompleted)
	if err != nil {
		return err
	}

	itemsAwarded := []models.ItemAwarded{}
	goldAwarded := 0

	var currentPOIInfo models.POIInfo
	if currentPOI != nil {
		currentPOIInfo = models.POIInfo{
			ID:       currentPOI.ID,
			Name:     currentPOI.Name,
			ImageURL: currentPOI.ImageUrl,
		}
	}

	var nextPOI *models.POIInfo
	nextNode := c.nextQuestNode(quest, node)
	if nextNode != nil && nextNode.PointOfInterestID != nil {
		poi, err := c.db.PointOfInterest().FindByID(ctx, *nextNode.PointOfInterestID)
		if err == nil && poi != nil {
			nextPOI = &models.POIInfo{
				ID:       poi.ID,
				Name:     poi.Name,
				ImageURL: poi.ImageUrl,
			}
		}
	}

	challengeActivityData, err := json.Marshal(models.ChallengeCompletedActivity{
		ChallengeID:       challenge.ID,
		Successful:        true,
		Reason:            "Challenge completed successfully!",
		SubmitterID:       &userID,
		ExperienceAwarded: experienceAwarded,
		ReputationAwarded: reputationAwarded,
		ItemsAwarded:      itemsAwarded,
		GoldAwarded:       goldAwarded,
		QuestID:           quest.ID,
		QuestName:         quest.Name,
		QuestCompleted:    questCompleted,
		CurrentPOI:        currentPOIInfo,
		NextPOI:           nextPOI,
		ZoneID:            zone.ID,
		ZoneName:          zone.Name,
	})
	if err != nil {
		return err
	}

	partyMembers, err := c.getPartyMembers(ctx, &userID, zone.ID)
	if err != nil {
		return err
	}
	for _, member := range partyMembers {
		if err := c.db.Activity().CreateActivity(ctx, models.Activity{
			UserID:       member.ID,
			ActivityType: models.ActivityTypeChallengeCompleted,
			Data:         challengeActivityData,
			Seen:         false,
		}); err != nil {
			return err
		}
	}

	if err := c.chatClient.AddQuestNodeCaptureMessage(ctx, teamID, &userID, quest, node); err != nil {
		return err
	}
	if questCompleted {
		if err := c.chatClient.AddQuestCompletedMessage(ctx, teamID, &userID, quest); err != nil {
			return err
		}
	}

	return nil
}

func (c *gameEngineClient) nextQuestNode(quest *models.Quest, current *models.QuestNode) *models.QuestNode {
	if quest == nil || current == nil {
		return nil
	}
	var next *models.QuestNode
	for i := range quest.Nodes {
		node := &quest.Nodes[i]
		if node.OrderIndex > current.OrderIndex {
			if next == nil || node.OrderIndex < next.OrderIndex {
				next = node
			}
		}
	}
	return next
}
