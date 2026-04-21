package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func serializeZoneGenre(genre models.ZoneGenre) gin.H {
	return gin.H{
		"id":         genre.ID,
		"name":       genre.Name,
		"sortOrder":  genre.SortOrder,
		"active":     genre.Active,
		"promptSeed": genre.PromptSeed,
		"createdAt":  genre.CreatedAt,
		"updatedAt":  genre.UpdatedAt,
	}
}

func serializeZoneGenreScores(genres []models.ZoneGenre, zoneScores map[uuid.UUID]int) []gin.H {
	if len(genres) == 0 {
		return []gin.H{}
	}
	scores := make([]gin.H, 0, len(genres))
	for _, genre := range genres {
		scoreValue := 0
		if zoneScores != nil {
			scoreValue = zoneScores[genre.ID]
		}
		scores = append(scores, gin.H{
			"genreId": genre.ID,
			"genre":   serializeZoneGenre(genre),
			"score":   scoreValue,
		})
	}
	return scores
}

func serializeZone(
	zone *models.Zone,
	genres []models.ZoneGenre,
	zoneScores map[uuid.UUID]int,
	discovery *models.ZoneDiscovery,
) gin.H {
	if zone == nil {
		return gin.H{}
	}
	discovered := discovery != nil && discovery.ID != uuid.Nil
	return gin.H{
		"id":             zone.ID,
		"createdAt":      zone.CreatedAt,
		"updatedAt":      zone.UpdatedAt,
		"name":           zone.Name,
		"description":    zone.Description,
		"kind":           zone.Kind,
		"internalTags":   zone.InternalTags,
		"latitude":       zone.Latitude,
		"longitude":      zone.Longitude,
		"zoneImportId":   zone.ZoneImportID,
		"boundary":       zone.Boundary,
		"boundaryCoords": zone.BoundaryCoords,
		"points":         zone.Points,
		"genreScores":    serializeZoneGenreScores(genres, zoneScores),
		"discovered":     discovered,
		"discoveredAt": func() interface{} {
			if !discovered {
				return nil
			}
			return discovery.CreatedAt
		}(),
	}
}

func zoneGenreScoreIndex(scores []models.ZoneGenreScore) map[uuid.UUID]map[uuid.UUID]int {
	index := make(map[uuid.UUID]map[uuid.UUID]int, len(scores))
	for _, score := range scores {
		if score.ZoneID == uuid.Nil || score.GenreID == uuid.Nil {
			continue
		}
		if _, ok := index[score.ZoneID]; !ok {
			index[score.ZoneID] = map[uuid.UUID]int{}
		}
		index[score.ZoneID][score.GenreID] = score.Score
	}
	return index
}

func zoneDiscoveryIndex(
	discoveries []models.ZoneDiscovery,
) map[uuid.UUID]*models.ZoneDiscovery {
	index := make(map[uuid.UUID]*models.ZoneDiscovery, len(discoveries))
	for i := range discoveries {
		discovery := discoveries[i]
		if discovery.ZoneID == uuid.Nil {
			continue
		}
		discoveryCopy := discovery
		index[discovery.ZoneID] = &discoveryCopy
	}
	return index
}

func (s *server) serializeZonesWithGenresAndDiscoveries(
	ctx context.Context,
	zones []*models.Zone,
	discoveryByZone map[uuid.UUID]*models.ZoneDiscovery,
) ([]gin.H, error) {
	activeGenres, err := s.dbClient.ZoneGenre().FindActive(ctx)
	if err != nil {
		return nil, err
	}

	zoneIDs := make([]uuid.UUID, 0, len(zones))
	for _, zone := range zones {
		if zone == nil || zone.ID == uuid.Nil {
			continue
		}
		zoneIDs = append(zoneIDs, zone.ID)
	}

	scoreMapByZone := map[uuid.UUID]map[uuid.UUID]int{}
	if len(zoneIDs) > 0 {
		scores, err := s.dbClient.ZoneGenreScore().FindByZoneIDs(ctx, zoneIDs, false)
		if err != nil {
			return nil, err
		}
		scoreMapByZone = zoneGenreScoreIndex(scores)
	}

	serialized := make([]gin.H, 0, len(zones))
	for _, zone := range zones {
		if zone == nil {
			continue
		}
		serialized = append(
			serialized,
			serializeZone(
				zone,
				activeGenres,
				scoreMapByZone[zone.ID],
				discoveryByZone[zone.ID],
			),
		)
	}
	return serialized, nil
}

func (s *server) serializeZonesWithGenres(
	ctx context.Context,
	zones []*models.Zone,
) ([]gin.H, error) {
	return s.serializeZonesWithGenresAndDiscoveries(ctx, zones, nil)
}

func (s *server) serializeZonesWithGenresForUser(
	ctx context.Context,
	userID uuid.UUID,
	zones []*models.Zone,
) ([]gin.H, error) {
	discoveries, err := s.dbClient.ZoneDiscovery().GetDiscoveriesForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.serializeZonesWithGenresAndDiscoveries(
		ctx,
		zones,
		zoneDiscoveryIndex(discoveries),
	)
}

func (s *server) serializeSingleZoneWithGenres(ctx context.Context, zone *models.Zone) (gin.H, error) {
	if zone == nil {
		return gin.H{}, nil
	}
	serialized, err := s.serializeZonesWithGenres(ctx, []*models.Zone{zone})
	if err != nil {
		return nil, err
	}
	if len(serialized) == 0 {
		return gin.H{}, nil
	}
	return serialized[0], nil
}

func (s *server) serializeSingleZoneWithGenresForUser(
	ctx context.Context,
	userID uuid.UUID,
	zone *models.Zone,
) (gin.H, error) {
	if zone == nil {
		return gin.H{}, nil
	}
	serialized, err := s.serializeZonesWithGenresForUser(ctx, userID, []*models.Zone{zone})
	if err != nil {
		return nil, err
	}
	if len(serialized) == 0 {
		return gin.H{}, nil
	}
	return serialized[0], nil
}

func (s *server) getZoneGenres(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	includeInactive := false
	if raw := strings.TrimSpace(ctx.Query("includeInactive")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "includeInactive must be a boolean"})
			return
		}
		includeInactive = parsed
	}

	genres, err := s.dbClient.ZoneGenre().FindAll(ctx, includeInactive)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, 0, len(genres))
	for _, genre := range genres {
		response = append(response, serializeZoneGenre(genre))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createZoneGenre(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var body struct {
		Name       string `json:"name"`
		SortOrder  int    `json:"sortOrder"`
		Active     *bool  `json:"active"`
		PromptSeed string `json:"promptSeed"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	genre := &models.ZoneGenre{
		Name:       strings.TrimSpace(body.Name),
		SortOrder:  body.SortOrder,
		Active:     body.Active == nil || *body.Active,
		PromptSeed: strings.TrimSpace(body.PromptSeed),
	}
	if err := s.dbClient.ZoneGenre().Create(ctx, genre); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, serializeZoneGenre(*genre))
}

func (s *server) updateZoneGenre(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	genreID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || genreID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone genre ID"})
		return
	}

	existing, err := s.dbClient.ZoneGenre().FindByID(ctx, genreID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "zone genre not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var body struct {
		Name       string  `json:"name"`
		SortOrder  *int    `json:"sortOrder"`
		Active     *bool   `json:"active"`
		PromptSeed *string `json:"promptSeed"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Name = strings.TrimSpace(body.Name)
	if body.SortOrder != nil {
		existing.SortOrder = *body.SortOrder
	}
	if body.Active != nil {
		existing.Active = *body.Active
	}
	if body.PromptSeed != nil {
		existing.PromptSeed = strings.TrimSpace(*body.PromptSeed)
	}

	if err := s.dbClient.ZoneGenre().Update(ctx, existing); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.ZoneGenre().FindByID(ctx, genreID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, serializeZoneGenre(*updated))
}

func (s *server) deleteZoneGenre(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	genreID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || genreID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone genre ID"})
		return
	}

	if err := s.dbClient.ZoneGenre().Delete(ctx, genreID); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "zone genre not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "zone genre deleted successfully"})
}
