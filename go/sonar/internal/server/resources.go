package server

import (
	"context"
	stdErrors "errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const resourceInteractRadiusMeters = 50.0

type resourceWithUserStatus struct {
	models.Resource
	GatheredByUser bool `json:"gatheredByUser"`
}

type resourceTypeUpsertRequest struct {
	Name          string  `json:"name"`
	Slug          *string `json:"slug"`
	Description   string  `json:"description"`
	MapIconURL    string  `json:"mapIconUrl"`
	MapIconPrompt string  `json:"mapIconPrompt"`
}

type resourceUpsertRequest struct {
	ZoneID          string   `json:"zoneId"`
	ResourceTypeID  string   `json:"resourceTypeId"`
	InventoryItemID int      `json:"inventoryItemId"`
	Quantity        *int     `json:"quantity"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
}

func normalizeResourceTypeSlug(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return ""
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '_' || r == '-':
			if builder.Len() == 0 || lastDash {
				continue
			}
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func defaultResourceTypeMapIconPrompt(resourceType *models.ResourceType) string {
	name := "resource"
	description := ""
	if resourceType != nil {
		if trimmed := strings.TrimSpace(resourceType.Name); trimmed != "" {
			name = trimmed
		}
		description = strings.TrimSpace(resourceType.Description)
	}
	if description == "" {
		description = "Gatherable world resource."
	}
	return fmt.Sprintf(
		"A retro 16-bit RPG map marker icon for %s. %s Top-down map-ready icon art, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.",
		name,
		description,
	)
}

func (s *server) resolveResourceTypeReference(
	ctx context.Context,
	resourceTypeIDRaw *string,
	legacyResourceTypeRaw *string,
) (*uuid.UUID, *models.ResourceType, error) {
	if resourceTypeIDRaw != nil {
		trimmed := strings.TrimSpace(*resourceTypeIDRaw)
		if trimmed != "" {
			resourceTypeID, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid resourceTypeId")
			}
			resourceType, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID)
			if err != nil {
				if stdErrors.Is(err, gorm.ErrRecordNotFound) {
					return nil, nil, fmt.Errorf("resourceTypeId not found")
				}
				return nil, nil, err
			}
			return &resourceType.ID, resourceType, nil
		}
	}
	if legacyResourceTypeRaw != nil {
		slug := normalizeResourceTypeSlug(*legacyResourceTypeRaw)
		if slug != "" {
			resourceType, err := s.dbClient.ResourceType().FindBySlug(ctx, slug)
			if err != nil {
				return nil, nil, err
			}
			if resourceType == nil {
				return nil, nil, fmt.Errorf("resourceType not found")
			}
			return &resourceType.ID, resourceType, nil
		}
	}
	return nil, nil, nil
}

func resourceTypeIDsMatch(item *models.InventoryItem, resourceTypeID uuid.UUID) bool {
	if item == nil || item.ResourceTypeID == nil {
		return false
	}
	return *item.ResourceTypeID == resourceTypeID
}

func modestResourceExperienceReward(userLevel *models.UserLevel) int {
	if userLevel == nil {
		return 12
	}
	xpToNext := userLevel.XPToNextLevel()
	if xpToNext < 1 {
		xpToNext = 1
	}
	reward := int(math.Round(float64(xpToNext) * 0.08))
	if reward < 12 {
		reward = 12
	}
	if reward > 220 {
		reward = 220
	}
	return reward
}

func (s *server) getResourceTypes(ctx *gin.Context) {
	resourceTypes, err := s.dbClient.ResourceType().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resourceTypes)
}

func (s *server) getResourceType(ctx *gin.Context) {
	resourceTypeID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
		return
	}
	resourceType, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource type not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resourceType)
}

func (s *server) createResourceType(ctx *gin.Context) {
	var requestBody resourceTypeUpsertRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(requestBody.Name)
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	slugSource := name
	if requestBody.Slug != nil && strings.TrimSpace(*requestBody.Slug) != "" {
		slugSource = *requestBody.Slug
	}
	slug := normalizeResourceTypeSlug(slugSource)
	if slug == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	existing, err := s.dbClient.ResourceType().FindBySlug(ctx, slug)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource type slug already exists"})
		return
	}

	resourceType := &models.ResourceType{
		Name:          name,
		Slug:          slug,
		Description:   strings.TrimSpace(requestBody.Description),
		MapIconURL:    strings.TrimSpace(requestBody.MapIconURL),
		MapIconPrompt: strings.TrimSpace(requestBody.MapIconPrompt),
	}
	if err := s.dbClient.ResourceType().Create(ctx, resourceType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create resource type: " + err.Error()})
		return
	}
	created, err := s.dbClient.ResourceType().FindByID(ctx, resourceType.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch created resource type: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateResourceType(ctx *gin.Context) {
	resourceTypeID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
		return
	}
	existing, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource type not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody resourceTypeUpsertRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := existing.Name
	if trimmed := strings.TrimSpace(requestBody.Name); trimmed != "" {
		name = trimmed
	}
	slugSource := name
	if requestBody.Slug != nil && strings.TrimSpace(*requestBody.Slug) != "" {
		slugSource = *requestBody.Slug
	} else if strings.TrimSpace(existing.Slug) != "" {
		slugSource = existing.Slug
	}
	slug := normalizeResourceTypeSlug(slugSource)
	if slug == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	if other, err := s.dbClient.ResourceType().FindBySlug(ctx, slug); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if other != nil && other.ID != existing.ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource type slug already exists"})
		return
	}

	updates := &models.ResourceType{
		Name:          name,
		Slug:          slug,
		Description:   strings.TrimSpace(requestBody.Description),
		MapIconURL:    strings.TrimSpace(requestBody.MapIconURL),
		MapIconPrompt: strings.TrimSpace(requestBody.MapIconPrompt),
	}
	if updates.Description == "" {
		updates.Description = existing.Description
	}
	if strings.TrimSpace(requestBody.MapIconURL) == "" {
		updates.MapIconURL = existing.MapIconURL
	}
	if strings.TrimSpace(requestBody.MapIconPrompt) == "" {
		updates.MapIconPrompt = existing.MapIconPrompt
	}
	if err := s.dbClient.ResourceType().Update(ctx, existing.ID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource type: " + err.Error()})
		return
	}
	updated, err := s.dbClient.ResourceType().FindByID(ctx, existing.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated resource type: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteResourceType(ctx *gin.Context) {
	resourceTypeID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
		return
	}
	if _, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource type not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ResourceType().Delete(ctx, resourceTypeID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete resource type: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource type deleted successfully"})
}

func (s *server) generateResourceTypeMapIcon(ctx *gin.Context) {
	resourceTypeID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
		return
	}
	resourceType, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource type not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Prompt *string `json:"prompt"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && !stdErrors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := strings.TrimSpace(resourceType.MapIconPrompt)
	if requestBody.Prompt != nil {
		prompt = strings.TrimSpace(*requestBody.Prompt)
	}
	if prompt == "" {
		prompt = defaultResourceTypeMapIconPrompt(resourceType)
	}

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)
	sourceImageURL, err := s.deepPriest.GenerateImage(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updates := &models.ResourceType{
		Name:          resourceType.Name,
		Slug:          resourceType.Slug,
		Description:   resourceType.Description,
		MapIconURL:    sourceImageURL,
		MapIconPrompt: prompt,
	}
	if err := s.dbClient.ResourceType().Update(ctx, resourceType.ID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource type icon: " + err.Error()})
		return
	}
	updated, err := s.dbClient.ResourceType().FindByID(ctx, resourceType.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated resource type: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) getResources(ctx *gin.Context) {
	user, userErr := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if userErr == nil {
		userID = &user.ID
	}
	resources, gatheredMap, err := s.dbClient.Resource().FindAllWithUserStatus(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]resourceWithUserStatus, 0, len(resources))
	for _, resource := range resources {
		if resource.Invalidated {
			continue
		}
		response = append(response, resourceWithUserStatus{
			Resource:       resource,
			GatheredByUser: gatheredMap[resource.ID],
		})
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getResource(ctx *gin.Context) {
	resourceID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
		return
	}
	user, userErr := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if userErr == nil {
		userID = &user.ID
	}
	resource, gatheredByUser, err := s.dbClient.Resource().FindByIDWithUserStatus(ctx, resourceID, userID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resource == nil || resource.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}
	ctx.JSON(http.StatusOK, resourceWithUserStatus{
		Resource:       *resource,
		GatheredByUser: gatheredByUser,
	})
}

func (s *server) getResourcesForZone(ctx *gin.Context) {
	zoneID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	user, userErr := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if userErr == nil {
		userID = &user.ID
	}
	resources, gatheredMap, err := s.dbClient.Resource().FindByZoneIDWithUserStatus(ctx, zoneID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]resourceWithUserStatus, 0, len(resources))
	for _, resource := range resources {
		if resource.Invalidated {
			continue
		}
		response = append(response, resourceWithUserStatus{
			Resource:       resource,
			GatheredByUser: gatheredMap[resource.ID],
		})
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createResource(ctx *gin.Context) {
	var requestBody resourceUpsertRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Latitude == nil || requestBody.Longitude == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return
	}
	if requestBody.Quantity == nil || *requestBody.Quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be 1 or greater"})
		return
	}
	zoneID, err := uuid.Parse(strings.TrimSpace(requestBody.ZoneID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	resourceTypeID, err := uuid.Parse(strings.TrimSpace(requestBody.ResourceTypeID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
		return
	}
	resourceType, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource type not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, requestBody.InventoryItemID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inventoryItem == nil || !resourceTypeIDsMatch(inventoryItem, resourceTypeID) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventory item resource type must match the resource type"})
		return
	}

	resource := &models.Resource{
		ZoneID:          zoneID,
		ResourceTypeID:  resourceType.ID,
		InventoryItemID: inventoryItem.ID,
		Quantity:        *requestBody.Quantity,
		Latitude:        *requestBody.Latitude,
		Longitude:       *requestBody.Longitude,
		Invalidated:     false,
	}
	if err := s.dbClient.Resource().Create(ctx, resource); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create resource: " + err.Error()})
		return
	}
	created, err := s.dbClient.Resource().FindByID(ctx, resource.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch created resource: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateResource(ctx *gin.Context) {
	resourceID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
		return
	}
	existing, err := s.dbClient.Resource().FindByID(ctx, resourceID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody resourceUpsertRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID := existing.ZoneID
	if trimmed := strings.TrimSpace(requestBody.ZoneID); trimmed != "" {
		zoneID, err = uuid.Parse(trimmed)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
			return
		}
	}
	resourceTypeID := existing.ResourceTypeID
	if trimmed := strings.TrimSpace(requestBody.ResourceTypeID); trimmed != "" {
		resourceTypeID, err = uuid.Parse(trimmed)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type ID"})
			return
		}
	}
	resourceType, err := s.dbClient.ResourceType().FindByID(ctx, resourceTypeID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource type not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inventoryItemID := existing.InventoryItemID
	if requestBody.InventoryItemID > 0 {
		inventoryItemID = requestBody.InventoryItemID
	}
	inventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, inventoryItemID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventory item not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inventoryItem == nil || !resourceTypeIDsMatch(inventoryItem, resourceType.ID) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "inventory item resource type must match the resource type"})
		return
	}
	quantity := existing.Quantity
	if requestBody.Quantity != nil {
		quantity = *requestBody.Quantity
	}
	if quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be 1 or greater"})
		return
	}
	latitude := existing.Latitude
	if requestBody.Latitude != nil {
		latitude = *requestBody.Latitude
	}
	longitude := existing.Longitude
	if requestBody.Longitude != nil {
		longitude = *requestBody.Longitude
	}

	updates := &models.Resource{
		ZoneID:          zoneID,
		ResourceTypeID:  resourceType.ID,
		InventoryItemID: inventoryItem.ID,
		Quantity:        quantity,
		Latitude:        latitude,
		Longitude:       longitude,
		Invalidated:     existing.Invalidated,
	}
	if err := s.dbClient.Resource().Update(ctx, existing.ID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource: " + err.Error()})
		return
	}
	updated, err := s.dbClient.Resource().FindByID(ctx, existing.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated resource: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteResource(ctx *gin.Context) {
	resourceID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
		return
	}
	if _, err := s.dbClient.Resource().FindByID(ctx, resourceID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Resource().Delete(ctx, resourceID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete resource: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

func (s *server) gatherResource(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	resourceID, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
		return
	}
	resource, err := s.dbClient.Resource().FindByID(ctx, resourceID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resource == nil || resource.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}
	if resource.Quantity < 1 || resource.InventoryItemID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource is not configured for gathering"})
		return
	}
	if !resourceTypeIDsMatch(&resource.InventoryItem, resource.ResourceTypeID) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource inventory item must match the resource type"})
		return
	}
	hasGathered, err := s.dbClient.Resource().HasUserGathered(ctx, user.ID, resource.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasGathered {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource already gathered"})
		return
	}
	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	distance := util.HaversineDistance(userLat, userLng, resource.Latitude, resource.Longitude)
	if distance > resourceInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"you must be within %.0f meters of the resource. Currently %.0f meters away",
				resourceInteractRadiusMeters,
				distance,
			),
		})
		return
	}

	gathering := &models.UserResourceGathering{
		UserID:     user.ID,
		ResourceID: resource.ID,
		GatheredAt: time.Now(),
	}
	if err := s.dbClient.Resource().CreateUserGathering(ctx, gathering); err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "duplicate key") || strings.Contains(lowerErr, "user_resource_gatherings_user_resource_unique") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource already gathered"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		_ = s.dbClient.Exec(ctx, fmt.Sprintf("DELETE FROM user_resource_gatherings WHERE id = '%s'", gathering.ID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rewardExperience := modestResourceExperienceReward(userLevel)
	itemsAwarded, spellsAwarded, err := s.awardScenarioRewards(
		ctx,
		user.ID,
		rewardExperience,
		0,
		[]scenarioRewardItem{{
			InventoryItemID: resource.InventoryItemID,
			Quantity:        resource.Quantity,
		}},
		[]scenarioRewardSpell{},
		[]string{},
	)
	if err != nil {
		_ = s.dbClient.Exec(ctx, fmt.Sprintf("DELETE FROM user_resource_gatherings WHERE id = '%s'", gathering.ID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated user: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":          "resource gathered successfully",
		"resourceId":       resource.ID,
		"rewardExperience": rewardExperience,
		"itemsAwarded":     itemsAwarded,
		"spellsAwarded":    spellsAwarded,
		"user":             updatedUser,
	})
}
