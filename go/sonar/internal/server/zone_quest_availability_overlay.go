package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type zoneQuestAvailabilityOverlayCharacterResponse struct {
	ID                         uuid.UUID `json:"id"`
	HasAvailableQuest          bool      `json:"hasAvailableQuest"`
	HasAvailableMainStoryQuest bool      `json:"hasAvailableMainStoryQuest"`
}

type zoneQuestAvailabilityOverlayPointOfInterestResponse struct {
	ID                         uuid.UUID                                       `json:"id"`
	HasAvailableQuest          bool                                            `json:"hasAvailableQuest"`
	HasAvailableMainStoryQuest bool                                            `json:"hasAvailableMainStoryQuest"`
	Characters                 []zoneQuestAvailabilityOverlayCharacterResponse `json:"characters"`
}

type zoneQuestAvailabilityOverlayResponse struct {
	PointsOfInterest []zoneQuestAvailabilityOverlayPointOfInterestResponse `json:"pointsOfInterest"`
	Characters       []zoneQuestAvailabilityOverlayCharacterResponse       `json:"characters"`
}

func (s *server) getZoneQuestAvailabilityOverlay(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := s.zoneQuestAvailabilityOverlayForUser(
		ctx,
		user,
		zone,
		activeStoryFlags,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) zoneQuestAvailabilityOverlayForUser(
	ctx *gin.Context,
	user *models.User,
	zone *models.Zone,
	activeStoryFlags map[string]bool,
) (zoneQuestAvailabilityOverlayResponse, error) {
	pointsOfInterest, err := s.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return zoneQuestAvailabilityOverlayResponse{}, err
	}

	if err := s.applyStoryWorldChangesToPointOfInterests(
		ctx,
		pointsOfInterest,
		activeStoryFlags,
	); err != nil {
		return zoneQuestAvailabilityOverlayResponse{}, err
	}

	zonePoiIDsList := make([]uuid.UUID, 0, len(pointsOfInterest))
	zonePoiIDs := make(map[uuid.UUID]struct{}, len(pointsOfInterest))
	for i := range pointsOfInterest {
		poi := pointsOfInterest[i]
		if poi.ID != uuid.Nil {
			zonePoiIDsList = append(zonePoiIDsList, poi.ID)
			zonePoiIDs[poi.ID] = struct{}{}
		}
	}

	characters, err := s.dbClient.Character().FindPotentiallyInZone(
		ctx,
		zone,
		zonePoiIDsList,
	)
	if err != nil {
		return zoneQuestAvailabilityOverlayResponse{}, err
	}

	relevantCharacterIDs := collectCharacterIDsFromPointsOfInterest(pointsOfInterest)
	for i := range characters {
		ch := characters[i]
		if ch == nil || ch.ID == uuid.Nil {
			continue
		}
		relevantCharacterIDs = append(relevantCharacterIDs, ch.ID)
	}

	availability, err := s.questAvailabilityByCharacterIDs(
		ctx,
		user.ID,
		relevantCharacterIDs,
	)
	if err != nil {
		return zoneQuestAvailabilityOverlayResponse{}, err
	}

	activeFetchCharacterIDs, err := s.activeFetchQuestCharacterIDsForUser(
		ctx,
		user.ID,
	)
	if err != nil {
		return zoneQuestAvailabilityOverlayResponse{}, err
	}

	poiResponse := make(
		[]zoneQuestAvailabilityOverlayPointOfInterestResponse,
		0,
		len(pointsOfInterest),
	)
	for i := range pointsOfInterest {
		poi := pointsOfInterest[i]
		characterResponse := make(
			[]zoneQuestAvailabilityOverlayCharacterResponse,
			0,
			len(poi.Characters),
		)
		hasAvailable := false
		hasAvailableMainStory := false
		for _, character := range poi.Characters {
			entry := availability[character.ID]
			if entry.HasAvailableQuest {
				hasAvailable = true
			}
			if entry.HasAvailableMainStoryQuest {
				hasAvailableMainStory = true
			}
			characterResponse = append(
				characterResponse,
				zoneQuestAvailabilityOverlayCharacterResponse{
					ID:                         character.ID,
					HasAvailableQuest:          entry.HasAvailableQuest,
					HasAvailableMainStoryQuest: entry.HasAvailableMainStoryQuest,
				},
			)
		}
		poiResponse = append(
			poiResponse,
			zoneQuestAvailabilityOverlayPointOfInterestResponse{
				ID:                         poi.ID,
				HasAvailableQuest:          hasAvailable,
				HasAvailableMainStoryQuest: hasAvailableMainStory,
				Characters:                 characterResponse,
			},
		)
	}

	characterResponse := make(
		[]zoneQuestAvailabilityOverlayCharacterResponse,
		0,
		len(characters),
	)
	for i := range characters {
		ch := characters[i]
		if ch == nil ||
			!characterVisibleToUser(user.ID, ch) ||
			!fetchQuestCharacterVisibleToUser(ch, activeFetchCharacterIDs) {
			continue
		}
		if err := s.applyStoryWorldChangesToCharacter(
			ctx,
			ch,
			activeStoryFlags,
		); err != nil {
			return zoneQuestAvailabilityOverlayResponse{}, err
		}
		if !characterInZone(zone, zonePoiIDs, ch) {
			continue
		}
		entry := availability[ch.ID]
		characterResponse = append(
			characterResponse,
			zoneQuestAvailabilityOverlayCharacterResponse{
				ID:                         ch.ID,
				HasAvailableQuest:          entry.HasAvailableQuest,
				HasAvailableMainStoryQuest: entry.HasAvailableMainStoryQuest,
			},
		)
	}

	return zoneQuestAvailabilityOverlayResponse{
		PointsOfInterest: poiResponse,
		Characters:       characterResponse,
	}, nil
}
