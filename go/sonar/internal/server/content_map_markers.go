package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sharedContentMapMarkerDefinition struct {
	ID                string
	Label             string
	DefaultPrompt     string
	DefaultKey        string
	DefaultStatusKey  string
	ActionPath        string
	SupportsZoneKinds bool
}

type contentMapMarkerStatusResponse struct {
	ID                    string  `json:"id"`
	Label                 string  `json:"label"`
	ThumbnailURL          string  `json:"thumbnailUrl"`
	EffectiveThumbnailURL string  `json:"effectiveThumbnailUrl"`
	DefaultThumbnailURL   string  `json:"defaultThumbnailUrl"`
	Status                string  `json:"status"`
	Exists                bool    `json:"exists"`
	DefaultExists         bool    `json:"defaultExists"`
	RequestedAt           *string `json:"requestedAt,omitempty"`
	LastModified          *string `json:"lastModified,omitempty"`
	DefaultPrompt         string  `json:"defaultPrompt"`
	ActionPath            string  `json:"actionPath"`
	SupportsZoneKinds     bool    `json:"supportsZoneKinds"`
	ZoneKind              string  `json:"zoneKind,omitempty"`
}

type resourceTypeMapMarkerStatusResponse struct {
	ResourceTypeID        string  `json:"resourceTypeId"`
	Name                  string  `json:"name"`
	Slug                  string  `json:"slug"`
	Description           string  `json:"description"`
	ThumbnailURL          string  `json:"thumbnailUrl"`
	EffectiveThumbnailURL string  `json:"effectiveThumbnailUrl"`
	DefaultThumbnailURL   string  `json:"defaultThumbnailUrl"`
	Status                string  `json:"status"`
	Exists                bool    `json:"exists"`
	DefaultExists         bool    `json:"defaultExists"`
	RequestedAt           *string `json:"requestedAt,omitempty"`
	LastModified          *string `json:"lastModified,omitempty"`
	DefaultPrompt         string  `json:"defaultPrompt"`
	ZoneKind              string  `json:"zoneKind,omitempty"`
	SupportsZoneKinds     bool    `json:"supportsZoneKinds"`
	CanDelete             bool    `json:"canDelete"`
}

type contentMapMarkerZoneKindSummary struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type contentMapMarkersPageResponse struct {
	ZoneKind            *contentMapMarkerZoneKindSummary                  `json:"zoneKind,omitempty"`
	SharedMarkers       []contentMapMarkerStatusResponse                  `json:"sharedMarkers"`
	PoiCategoryMarkers  []pointOfInterestMarkerCategoryIconStatusResponse `json:"poiCategoryMarkers"`
	ResourceTypeMarkers []resourceTypeMapMarkerStatusResponse             `json:"resourceTypeMarkers"`
}

type contentMapMarkerExistenceCache map[string]bool

var sharedContentMapMarkerDefinitions = []sharedContentMapMarkerDefinition{
	{
		ID:                "poi-undiscovered",
		Label:             "Undiscovered POI",
		DefaultPrompt:     poiUndiscoveredIconText,
		DefaultKey:        poiUndiscoveredIconKey,
		DefaultStatusKey:  poiUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/poi-undiscovered",
		SupportsZoneKinds: true,
	},
	{
		ID:                "scenario-undiscovered",
		Label:             "Scenario Marker",
		DefaultPrompt:     scenarioUndiscoveredIconText,
		DefaultKey:        scenarioUndiscoveredIconKey,
		DefaultStatusKey:  scenarioUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/scenario-undiscovered",
		SupportsZoneKinds: true,
	},
	{
		ID:                "exposition-undiscovered",
		Label:             "Exposition Marker",
		DefaultPrompt:     expositionUndiscoveredIconText,
		DefaultKey:        expositionUndiscoveredIconKey,
		DefaultStatusKey:  expositionUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/exposition-undiscovered",
		SupportsZoneKinds: true,
	},
	{
		ID:                "treasure-chest-undiscovered",
		Label:             "Treasure Chest Marker",
		DefaultPrompt:     treasureChestUndiscoveredIconText,
		DefaultKey:        treasureChestUndiscoveredIconKey,
		DefaultStatusKey:  treasureChestUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/treasure-chest-undiscovered",
		SupportsZoneKinds: true,
	},
	{
		ID:                "monster-undiscovered",
		Label:             "Monster Marker",
		DefaultPrompt:     monsterUndiscoveredIconText,
		DefaultKey:        monsterUndiscoveredIconKey,
		DefaultStatusKey:  monsterUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/monster-undiscovered/monster",
		SupportsZoneKinds: true,
	},
	{
		ID:                "boss-undiscovered",
		Label:             "Boss Marker",
		DefaultPrompt:     bossUndiscoveredIconText,
		DefaultKey:        bossUndiscoveredIconKey,
		DefaultStatusKey:  bossUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/monster-undiscovered/boss",
		SupportsZoneKinds: true,
	},
	{
		ID:                "raid-undiscovered",
		Label:             "Raid Marker",
		DefaultPrompt:     raidUndiscoveredIconText,
		DefaultKey:        raidUndiscoveredIconKey,
		DefaultStatusKey:  raidUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/monster-undiscovered/raid",
		SupportsZoneKinds: true,
	},
	{
		ID:                "character-undiscovered",
		Label:             "Character Marker",
		DefaultPrompt:     characterUndiscoveredIconText,
		DefaultKey:        characterUndiscoveredIconKey,
		DefaultStatusKey:  characterUndiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/character-undiscovered",
		SupportsZoneKinds: true,
	},
	{
		ID:                "healing-fountain-discovered",
		Label:             "Healing Fountain Marker",
		DefaultPrompt:     healingFountainDiscoveredIconText,
		DefaultKey:        healingFountainDiscoveredIconKey,
		DefaultStatusKey:  healingFountainDiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/healing-fountain-discovered",
		SupportsZoneKinds: true,
	},
	{
		ID:                "base-discovered",
		Label:             "Base Marker",
		DefaultPrompt:     baseDiscoveredIconText,
		DefaultKey:        baseDiscoveredIconKey,
		DefaultStatusKey:  baseDiscoveredStatusKey,
		ActionPath:        "/sonar/admin/thumbnails/base",
		SupportsZoneKinds: false,
	},
}

func monsterEncounterContentMapMarkerDefinition(
	encounterType models.MonsterEncounterType,
) sharedContentMapMarkerDefinition {
	switch encounterType {
	case models.MonsterEncounterTypeBoss:
		return sharedContentMapMarkerDefinitions[5]
	case models.MonsterEncounterTypeRaid:
		return sharedContentMapMarkerDefinitions[6]
	default:
		return sharedContentMapMarkerDefinitions[4]
	}
}

func contentMapMarkerZoneKindSlug(zoneKind *models.ZoneKind) string {
	if zoneKind == nil {
		return ""
	}
	return models.NormalizeZoneKind(zoneKind.Slug)
}

func buildContentMapMarkerZoneKindSummary(zoneKind *models.ZoneKind) *contentMapMarkerZoneKindSummary {
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if slug == "" {
		return nil
	}
	return &contentMapMarkerZoneKindSummary{
		Slug:        slug,
		Name:        strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind)),
		Description: strings.TrimSpace(zoneKind.Description),
	}
}

func mergeZoneKindContentMapMarkerPrompt(basePrompt string, zoneKind *models.ZoneKind) string {
	basePrompt = strings.TrimSpace(basePrompt)
	if basePrompt == "" {
		return ""
	}
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if slug == "" {
		return basePrompt
	}
	label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	if label == "" {
		label = slug
	}
	flavor := strings.TrimSpace(models.ZoneKindPromptSeed(zoneKind))
	if flavor == "" {
		flavor = fmt.Sprintf(
			"Keep the icon immediately legible as %s-flavored content.",
			strings.ToLower(label),
		)
	}
	return strings.TrimSpace(fmt.Sprintf(
		"%s Adapt the icon so it unmistakably belongs in a %s zone. %s Keep the gameplay role identical, but let the silhouette, props, materials, and motif shift toward the zone's most iconic version of that role when useful, such as turning a generic herbalism node into a crop stalk in farmland. Preserve clean top-down map readability.",
		basePrompt,
		strings.ToLower(label),
		flavor,
	))
}

func sharedContentMapMarkerDestinationKey(
	definition sharedContentMapMarkerDefinition,
	zoneKind *models.ZoneKind,
) string {
	if !definition.SupportsZoneKinds {
		return definition.DefaultKey
	}
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if slug == "" {
		return definition.DefaultKey
	}
	return fmt.Sprintf("thumbnails/placeholders/content-map-markers/%s/%s", definition.ID, slug)
}

func sharedContentMapMarkerStatusKey(
	definition sharedContentMapMarkerDefinition,
	zoneKind *models.ZoneKind,
) string {
	if !definition.SupportsZoneKinds {
		return definition.DefaultStatusKey
	}
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if slug == "" {
		return definition.DefaultStatusKey
	}
	return fmt.Sprintf("admin:content-map-markers:%s:%s:requested-at", definition.ID, slug)
}

func pointOfInterestMarkerCategoryVariantThumbnailKey(
	category models.PointOfInterestMarkerCategory,
	zoneKind *models.ZoneKind,
) string {
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if slug == "" {
		return pointOfInterestMarkerCategoryThumbnailKey(category)
	}
	return fmt.Sprintf(
		"thumbnails/placeholders/content-map-markers/poi-category/%s/%s",
		category,
		slug,
	)
}

func pointOfInterestMarkerCategoryVariantStatusKey(
	category models.PointOfInterestMarkerCategory,
	zoneKind *models.ZoneKind,
) string {
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if slug == "" {
		return pointOfInterestMarkerCategoryThumbnailStatusKey(category)
	}
	return fmt.Sprintf(
		"admin:content-map-markers:poi-category:%s:%s:requested-at",
		category,
		slug,
	)
}

func resourceTypeContentMapMarkerDestinationKey(
	resourceTypeID uuid.UUID,
	zoneKind *models.ZoneKind,
) string {
	slug := contentMapMarkerZoneKindSlug(zoneKind)
	if resourceTypeID == uuid.Nil || slug == "" {
		return ""
	}
	return fmt.Sprintf(
		"thumbnails/placeholders/content-map-markers/resources/%s/%s",
		resourceTypeID.String(),
		slug,
	)
}

func resourceTypeContentMapMarkerPrompt(
	resourceType *models.ResourceType,
	zoneKind *models.ZoneKind,
) string {
	basePrompt := defaultResourceTypeMapIconPrompt(resourceType)
	if resourceType != nil {
		if custom := strings.TrimSpace(resourceType.MapIconPrompt); custom != "" {
			basePrompt = custom
		}
	}
	return mergeZoneKindContentMapMarkerPrompt(basePrompt, zoneKind)
}

func effectiveContentMapMarkerZoneKind(recordZoneKind string, zone *models.Zone) string {
	if zone != nil {
		if normalized := models.NormalizeZoneKind(zone.Kind); normalized != "" {
			return normalized
		}
	}
	return models.NormalizeZoneKind(recordZoneKind)
}

func (s *server) resolveContentMapMarkerZoneKind(
	ctx *gin.Context,
) (*models.ZoneKind, error) {
	rawZoneKind := strings.TrimSpace(ctx.Query("zoneKind"))
	if rawZoneKind == "" {
		return nil, nil
	}
	return s.resolveOptionalZoneKind(ctx.Request.Context(), rawZoneKind)
}

func (s *server) staticThumbnailExistsCached(
	ctx context.Context,
	cache contentMapMarkerExistenceCache,
	destinationKey string,
) bool {
	destinationKey = strings.TrimSpace(destinationKey)
	if destinationKey == "" {
		return false
	}
	if exists, ok := cache[destinationKey]; ok {
		return exists
	}
	lastModified, err := s.awsClient.GetObjectLastModified(jobs.ThumbnailBucket, destinationKey)
	exists := err == nil && lastModified != nil
	cache[destinationKey] = exists
	return exists
}

func (s *server) sharedContentMapMarkerStatusResponse(
	ctx *gin.Context,
	definition sharedContentMapMarkerDefinition,
	zoneKind *models.ZoneKind,
) (*contentMapMarkerStatusResponse, error) {
	targetKey := sharedContentMapMarkerDestinationKey(definition, zoneKind)
	targetStatusKey := sharedContentMapMarkerStatusKey(definition, zoneKind)
	status, exists, requestedAt, lastModified, err := s.readStaticThumbnailStatus(
		ctx,
		targetKey,
		targetStatusKey,
	)
	if err != nil {
		return nil, err
	}

	defaultExists := exists
	if contentMapMarkerZoneKindSlug(zoneKind) != "" && definition.SupportsZoneKinds {
		_, defaultExists, _, _, err = s.readStaticThumbnailStatus(
			ctx,
			definition.DefaultKey,
			definition.DefaultStatusKey,
		)
		if err != nil {
			return nil, err
		}
	}

	response := &contentMapMarkerStatusResponse{
		ID:                    definition.ID,
		Label:                 definition.Label,
		ThumbnailURL:          staticThumbnailURL(targetKey),
		EffectiveThumbnailURL: staticThumbnailURL(definition.DefaultKey),
		DefaultThumbnailURL:   staticThumbnailURL(definition.DefaultKey),
		Status:                status,
		Exists:                exists,
		DefaultExists:         defaultExists,
		DefaultPrompt:         mergeZoneKindContentMapMarkerPrompt(definition.DefaultPrompt, zoneKind),
		ActionPath:            definition.ActionPath,
		SupportsZoneKinds:     definition.SupportsZoneKinds,
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

func (s *server) resourceTypeContentMapMarkerStatusResponse(
	ctx *gin.Context,
	resourceType models.ResourceType,
	zoneKind *models.ZoneKind,
) (resourceTypeMapMarkerStatusResponse, error) {
	defaultThumbnailURL := strings.TrimSpace(resourceType.MapIconURL)
	defaultExists := defaultThumbnailURL != ""
	response := resourceTypeMapMarkerStatusResponse{
		ResourceTypeID:        resourceType.ID.String(),
		Name:                  resourceType.Name,
		Slug:                  resourceType.Slug,
		Description:           resourceType.Description,
		ThumbnailURL:          defaultThumbnailURL,
		EffectiveThumbnailURL: defaultThumbnailURL,
		DefaultThumbnailURL:   defaultThumbnailURL,
		Status:                "missing",
		Exists:                defaultExists,
		DefaultExists:         defaultExists,
		DefaultPrompt:         resourceTypeContentMapMarkerPrompt(&resourceType, zoneKind),
		ZoneKind:              contentMapMarkerZoneKindSlug(zoneKind),
		SupportsZoneKinds:     true,
		CanDelete:             contentMapMarkerZoneKindSlug(zoneKind) != "",
	}

	if contentMapMarkerZoneKindSlug(zoneKind) == "" {
		if defaultExists {
			response.Status = "completed"
		}
		return response, nil
	}

	destinationKey := resourceTypeContentMapMarkerDestinationKey(resourceType.ID, zoneKind)
	lastModified, err := s.awsClient.GetObjectLastModified(jobs.ThumbnailBucket, destinationKey)
	if err != nil {
		return resourceTypeMapMarkerStatusResponse{}, err
	}
	response.ThumbnailURL = staticThumbnailURL(destinationKey)
	response.Exists = lastModified != nil
	response.Status = "missing"
	if lastModified != nil {
		response.Status = "completed"
		value := lastModified.UTC().Format(time.RFC3339Nano)
		response.LastModified = &value
		response.EffectiveThumbnailURL = response.ThumbnailURL
	} else {
		response.EffectiveThumbnailURL = defaultThumbnailURL
	}
	return response, nil
}

func (s *server) uploadStaticContentMapMarker(
	destinationKey string,
	imageBytes []byte,
) (string, error) {
	destinationKey = strings.TrimSpace(destinationKey)
	if destinationKey == "" {
		return "", fmt.Errorf("missing content marker destination key")
	}
	return s.awsClient.UploadImageToS3(jobs.ThumbnailBucket, destinationKey, imageBytes)
}

func (s *server) getContentMapMarkersPageData(ctx *gin.Context) {
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sharedMarkers := make([]contentMapMarkerStatusResponse, 0, len(sharedContentMapMarkerDefinitions))
	for _, definition := range sharedContentMapMarkerDefinitions {
		response, err := s.sharedContentMapMarkerStatusResponse(ctx, definition, zoneKind)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sharedMarkers = append(sharedMarkers, *response)
	}

	poiCategoryMarkers := make([]pointOfInterestMarkerCategoryIconStatusResponse, 0, len(allPointOfInterestMarkerCategories()))
	for _, category := range allPointOfInterestMarkerCategories() {
		response, err := s.pointOfInterestMarkerCategoryStatusResponse(ctx, category, zoneKind)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		poiCategoryMarkers = append(poiCategoryMarkers, *response)
	}
	sort.Slice(poiCategoryMarkers, func(i, j int) bool {
		return poiCategoryMarkers[i].Label < poiCategoryMarkers[j].Label
	})

	resourceTypes, err := s.dbClient.ResourceType().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resourceTypeMarkers := make([]resourceTypeMapMarkerStatusResponse, 0, len(resourceTypes))
	for i := range resourceTypes {
		entry, err := s.resourceTypeContentMapMarkerStatusResponse(ctx, resourceTypes[i], zoneKind)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resourceTypeMarkers = append(resourceTypeMarkers, entry)
	}
	sort.Slice(resourceTypeMarkers, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(resourceTypeMarkers[i].Name))
		right := strings.ToLower(strings.TrimSpace(resourceTypeMarkers[j].Name))
		if left == right {
			return resourceTypeMarkers[i].Slug < resourceTypeMarkers[j].Slug
		}
		return left < right
	})

	ctx.JSON(http.StatusOK, contentMapMarkersPageResponse{
		ZoneKind:            buildContentMapMarkerZoneKindSummary(zoneKind),
		SharedMarkers:       sharedMarkers,
		PoiCategoryMarkers:  poiCategoryMarkers,
		ResourceTypeMarkers: resourceTypeMarkers,
	})
}

func (s *server) generateSharedContentMapMarker(
	ctx *gin.Context,
	definition sharedContentMapMarkerDefinition,
) {
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.queueGeneratedStaticThumbnail(
		ctx,
		mergeZoneKindContentMapMarkerPrompt(definition.DefaultPrompt, zoneKind),
		sharedContentMapMarkerDestinationKey(definition, zoneKind),
		sharedContentMapMarkerStatusKey(definition, zoneKind),
	)
}

func (s *server) getSharedContentMapMarkerStatus(
	ctx *gin.Context,
	definition sharedContentMapMarkerDefinition,
) {
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := s.sharedContentMapMarkerStatusResponse(ctx, definition, zoneKind)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) deleteSharedContentMapMarker(
	ctx *gin.Context,
	definition sharedContentMapMarkerDefinition,
) {
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.deleteStaticThumbnail(
		ctx,
		sharedContentMapMarkerDestinationKey(definition, zoneKind),
		sharedContentMapMarkerStatusKey(definition, zoneKind),
	)
}

func (s *server) deleteResourceTypeContentMapMarker(ctx *gin.Context) {
	zoneKind, err := s.resolveContentMapMarkerZoneKind(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if contentMapMarkerZoneKindSlug(zoneKind) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "zoneKind is required to delete a resource type override marker"})
		return
	}
	resourceTypeID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
		return
	}
	destinationKey := resourceTypeContentMapMarkerDestinationKey(resourceTypeID, zoneKind)
	if err := s.awsClient.DeleteObjectFromS3(jobs.ThumbnailBucket, destinationKey); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":       "deleted",
		"thumbnailUrl": staticThumbnailURL(destinationKey),
		"zoneKind":     contentMapMarkerZoneKindSlug(zoneKind),
	})
}

func (s *server) resolveSharedContentMapMarkerURL(
	ctx context.Context,
	definition sharedContentMapMarkerDefinition,
	zoneKind string,
	fallbackURL string,
	cache contentMapMarkerExistenceCache,
) string {
	normalizedZoneKind := models.NormalizeZoneKind(zoneKind)
	if definition.SupportsZoneKinds && normalizedZoneKind != "" {
		variantKey := fmt.Sprintf("thumbnails/placeholders/content-map-markers/%s/%s", definition.ID, normalizedZoneKind)
		if s.staticThumbnailExistsCached(ctx, cache, variantKey) {
			return staticThumbnailURL(variantKey)
		}
	}
	if strings.TrimSpace(fallbackURL) != "" {
		return strings.TrimSpace(fallbackURL)
	}
	return staticThumbnailURL(definition.DefaultKey)
}

func (s *server) resolveResourceTypeMapMarkerURL(
	ctx context.Context,
	resourceType *models.ResourceType,
	zoneKind string,
	cache contentMapMarkerExistenceCache,
) string {
	if resourceType == nil {
		return ""
	}
	normalizedZoneKind := models.NormalizeZoneKind(zoneKind)
	if normalizedZoneKind != "" {
		variantKey := fmt.Sprintf(
			"thumbnails/placeholders/content-map-markers/resources/%s/%s",
			resourceType.ID.String(),
			normalizedZoneKind,
		)
		if s.staticThumbnailExistsCached(ctx, cache, variantKey) {
			return staticThumbnailURL(variantKey)
		}
	}
	return strings.TrimSpace(resourceType.MapIconURL)
}

func (s *server) resolvePointOfInterestMapMarkerURL(
	ctx context.Context,
	category models.PointOfInterestMarkerCategory,
	zoneKind string,
	cache contentMapMarkerExistenceCache,
) string {
	normalizedZoneKind := models.NormalizeZoneKind(zoneKind)
	if normalizedZoneKind != "" {
		variantKey := fmt.Sprintf(
			"thumbnails/placeholders/content-map-markers/poi-category/%s/%s",
			category,
			normalizedZoneKind,
		)
		if s.staticThumbnailExistsCached(ctx, cache, variantKey) {
			return staticThumbnailURL(variantKey)
		}
	}
	defaultKey := pointOfInterestMarkerCategoryThumbnailKey(category)
	if s.staticThumbnailExistsCached(ctx, cache, defaultKey) {
		return staticThumbnailURL(defaultKey)
	}
	return ""
}
