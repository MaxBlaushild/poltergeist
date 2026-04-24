package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

const (
	poiMarkerCategoryIconKeyPrefix   = "thumbnails/placeholders/poi-marker-category"
	poiMarkerCategoryStatusKeyPrefix = "admin:thumbnails:poi-marker-category"
)

type pointOfInterestMarkerCategoryIconStatusResponse struct {
	Category              string  `json:"category"`
	Label                 string  `json:"label"`
	DefaultPrompt         string  `json:"defaultPrompt"`
	ThumbnailURL          string  `json:"thumbnailUrl"`
	EffectiveThumbnailURL string  `json:"effectiveThumbnailUrl"`
	DefaultThumbnailURL   string  `json:"defaultThumbnailUrl"`
	Status                string  `json:"status"`
	Exists                bool    `json:"exists"`
	DefaultExists         bool    `json:"defaultExists"`
	ActionPath            string  `json:"actionPath"`
	SupportsZoneKinds     bool    `json:"supportsZoneKinds"`
	ZoneKind              string  `json:"zoneKind,omitempty"`
	RequestedAt           *string `json:"requestedAt,omitempty"`
	LastModified          *string `json:"lastModified,omitempty"`
}

func allPointOfInterestMarkerCategories() []models.PointOfInterestMarkerCategory {
	return []models.PointOfInterestMarkerCategory{
		models.PointOfInterestMarkerCategoryGeneric,
		models.PointOfInterestMarkerCategoryCoffeehouse,
		models.PointOfInterestMarkerCategoryTavern,
		models.PointOfInterestMarkerCategoryEatery,
		models.PointOfInterestMarkerCategoryMarket,
		models.PointOfInterestMarkerCategoryArchive,
		models.PointOfInterestMarkerCategoryPark,
		models.PointOfInterestMarkerCategoryWaterfront,
		models.PointOfInterestMarkerCategoryMuseum,
		models.PointOfInterestMarkerCategoryTheater,
		models.PointOfInterestMarkerCategoryLandmark,
		models.PointOfInterestMarkerCategoryCivic,
		models.PointOfInterestMarkerCategoryArena,
	}
}

func parsePointOfInterestMarkerCategory(raw string) (models.PointOfInterestMarkerCategory, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(models.PointOfInterestMarkerCategoryGeneric):
		return models.PointOfInterestMarkerCategoryGeneric, true
	case string(models.PointOfInterestMarkerCategoryCoffeehouse):
		return models.PointOfInterestMarkerCategoryCoffeehouse, true
	case string(models.PointOfInterestMarkerCategoryTavern):
		return models.PointOfInterestMarkerCategoryTavern, true
	case string(models.PointOfInterestMarkerCategoryEatery):
		return models.PointOfInterestMarkerCategoryEatery, true
	case string(models.PointOfInterestMarkerCategoryMarket):
		return models.PointOfInterestMarkerCategoryMarket, true
	case string(models.PointOfInterestMarkerCategoryArchive):
		return models.PointOfInterestMarkerCategoryArchive, true
	case string(models.PointOfInterestMarkerCategoryPark):
		return models.PointOfInterestMarkerCategoryPark, true
	case string(models.PointOfInterestMarkerCategoryWaterfront):
		return models.PointOfInterestMarkerCategoryWaterfront, true
	case string(models.PointOfInterestMarkerCategoryMuseum):
		return models.PointOfInterestMarkerCategoryMuseum, true
	case string(models.PointOfInterestMarkerCategoryTheater):
		return models.PointOfInterestMarkerCategoryTheater, true
	case string(models.PointOfInterestMarkerCategoryLandmark):
		return models.PointOfInterestMarkerCategoryLandmark, true
	case string(models.PointOfInterestMarkerCategoryCivic):
		return models.PointOfInterestMarkerCategoryCivic, true
	case string(models.PointOfInterestMarkerCategoryArena):
		return models.PointOfInterestMarkerCategoryArena, true
	default:
		return "", false
	}
}

func pointOfInterestMarkerCategoryLabel(category models.PointOfInterestMarkerCategory) string {
	switch category {
	case models.PointOfInterestMarkerCategoryCoffeehouse:
		return "Coffeehouse"
	case models.PointOfInterestMarkerCategoryTavern:
		return "Tavern"
	case models.PointOfInterestMarkerCategoryEatery:
		return "Eatery"
	case models.PointOfInterestMarkerCategoryMarket:
		return "Market"
	case models.PointOfInterestMarkerCategoryArchive:
		return "Archive"
	case models.PointOfInterestMarkerCategoryPark:
		return "Park"
	case models.PointOfInterestMarkerCategoryWaterfront:
		return "Waterfront"
	case models.PointOfInterestMarkerCategoryMuseum:
		return "Museum"
	case models.PointOfInterestMarkerCategoryTheater:
		return "Theater"
	case models.PointOfInterestMarkerCategoryLandmark:
		return "Landmark"
	case models.PointOfInterestMarkerCategoryCivic:
		return "Civic"
	case models.PointOfInterestMarkerCategoryArena:
		return "Arena"
	default:
		return "Generic"
	}
}

func defaultPointOfInterestMarkerCategoryIconPrompt(category models.PointOfInterestMarkerCategory) string {
	switch category {
	case models.PointOfInterestMarkerCategoryCoffeehouse:
		return "A retro 16-bit RPG map marker icon for a discovered coffeehouse point of interest. Enchanted steaming cup sigil, cozy hearth motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryTavern:
		return "A retro 16-bit RPG map marker icon for a discovered tavern point of interest. Adventurer mug and lantern sigil, warm nightspot energy, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryEatery:
		return "A retro 16-bit RPG map marker icon for a discovered eatery point of interest. Hearthplate and cutlery sigil with welcoming feast motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryMarket:
		return "A retro 16-bit RPG map marker icon for a discovered market point of interest. Merchant stall and satchel sigil, bustling bazaar motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryArchive:
		return "A retro 16-bit RPG map marker icon for a discovered archive point of interest. Illuminated book and scroll sigil, scriptorium motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryPark:
		return "A retro 16-bit RPG map marker icon for a discovered park point of interest. Sacred tree and garden path sigil, verdant grove motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryWaterfront:
		return "A retro 16-bit RPG map marker icon for a discovered waterfront point of interest. Harbor wave and pier sigil, bright shoreline motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryMuseum:
		return "A retro 16-bit RPG map marker icon for a discovered museum point of interest. Relic plinth and curiosity sigil, grand gallery motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryTheater:
		return "A retro 16-bit RPG map marker icon for a discovered theater point of interest. Stage curtain and spotlight sigil, bardic performance motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryLandmark:
		return "A retro 16-bit RPG map marker icon for a discovered landmark point of interest. Monument spire and compass sigil, legendary waypoint motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryCivic:
		return "A retro 16-bit RPG map marker icon for a discovered civic point of interest. Sealed courier post and townhall sigil, public service motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	case models.PointOfInterestMarkerCategoryArena:
		return "A retro 16-bit RPG map marker icon for a discovered arena point of interest. Victory pennant and coliseum sigil, competitive grounds motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	default:
		return "A retro 16-bit RPG map marker icon for a discovered point of interest. Wayfinder sigil with cartographer rune motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette."
	}
}

func pointOfInterestMarkerCategoryThumbnailKey(category models.PointOfInterestMarkerCategory) string {
	return fmt.Sprintf("%s-%s.png", poiMarkerCategoryIconKeyPrefix, category)
}

func pointOfInterestMarkerCategoryThumbnailStatusKey(category models.PointOfInterestMarkerCategory) string {
	return fmt.Sprintf("%s:%s:requested-at", poiMarkerCategoryStatusKeyPrefix, category)
}

func (s *server) pointOfInterestMarkerCategoryStatusResponse(
	ctx *gin.Context,
	category models.PointOfInterestMarkerCategory,
	zoneKind *models.ZoneKind,
) (*pointOfInterestMarkerCategoryIconStatusResponse, error) {
	destinationKey := pointOfInterestMarkerCategoryVariantThumbnailKey(category, zoneKind)
	statusKey := pointOfInterestMarkerCategoryVariantStatusKey(category, zoneKind)
	status, exists, requestedAt, lastModified, err := s.readStaticThumbnailStatus(ctx, destinationKey, statusKey)
	if err != nil {
		return nil, err
	}
	defaultDestinationKey := pointOfInterestMarkerCategoryThumbnailKey(category)
	defaultExists := exists
	if contentMapMarkerZoneKindSlug(zoneKind) != "" {
		_, defaultExists, _, _, err = s.readStaticThumbnailStatus(
			ctx,
			defaultDestinationKey,
			pointOfInterestMarkerCategoryThumbnailStatusKey(category),
		)
		if err != nil {
			return nil, err
		}
	}

	response := &pointOfInterestMarkerCategoryIconStatusResponse{
		Category:              string(category),
		Label:                 pointOfInterestMarkerCategoryLabel(category),
		DefaultPrompt:         mergeZoneKindContentMapMarkerPrompt(defaultPointOfInterestMarkerCategoryIconPrompt(category), zoneKind),
		ThumbnailURL:          staticThumbnailURL(destinationKey),
		EffectiveThumbnailURL: staticThumbnailURL(defaultDestinationKey),
		DefaultThumbnailURL:   staticThumbnailURL(defaultDestinationKey),
		Status:                status,
		Exists:                exists,
		DefaultExists:         defaultExists,
		ActionPath:            fmt.Sprintf("/sonar/admin/thumbnails/poi-marker-categories/%s", category),
		SupportsZoneKinds:     true,
		ZoneKind:              contentMapMarkerZoneKindSlug(zoneKind),
	}
	if exists {
		response.EffectiveThumbnailURL = response.ThumbnailURL
	}
	if requestedAt != nil {
		value := requestedAt.UTC().Format(time.RFC3339Nano)
		response.RequestedAt = &value
	}
	if lastModified != nil {
		value := lastModified.UTC().Format(time.RFC3339Nano)
		response.LastModified = &value
	}
	return response, nil
}

func (s *server) listPointOfInterestMarkerCategoryIcons(ctx *gin.Context) {
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	responses := make([]*pointOfInterestMarkerCategoryIconStatusResponse, 0, len(allPointOfInterestMarkerCategories()))
	for _, category := range allPointOfInterestMarkerCategories() {
		response, err := s.pointOfInterestMarkerCategoryStatusResponse(ctx, category, zoneKind)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		responses = append(responses, response)
	}
	ctx.JSON(http.StatusOK, responses)
}

func (s *server) generatePointOfInterestMarkerCategoryIcon(ctx *gin.Context) {
	category, ok := parsePointOfInterestMarkerCategory(ctx.Param("category"))
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest marker category"})
		return
	}
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.queueGeneratedStaticThumbnail(
		ctx,
		mergeZoneKindContentMapMarkerPrompt(defaultPointOfInterestMarkerCategoryIconPrompt(category), zoneKind),
		pointOfInterestMarkerCategoryVariantThumbnailKey(category, zoneKind),
		pointOfInterestMarkerCategoryVariantStatusKey(category, zoneKind),
	)
}

func (s *server) getPointOfInterestMarkerCategoryIconStatus(ctx *gin.Context) {
	category, ok := parsePointOfInterestMarkerCategory(ctx.Param("category"))
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest marker category"})
		return
	}
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := s.pointOfInterestMarkerCategoryStatusResponse(ctx, category, zoneKind)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) deletePointOfInterestMarkerCategoryIcon(ctx *gin.Context) {
	category, ok := parsePointOfInterestMarkerCategory(ctx.Param("category"))
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid point of interest marker category"})
		return
	}
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.deleteStaticThumbnail(
		ctx,
		pointOfInterestMarkerCategoryVariantThumbnailKey(category, zoneKind),
		pointOfInterestMarkerCategoryVariantStatusKey(category, zoneKind),
	)
}
