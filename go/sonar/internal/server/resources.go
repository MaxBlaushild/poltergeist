package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sort"
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
	Name               string                                   `json:"name"`
	Slug               *string                                  `json:"slug"`
	Description        string                                   `json:"description"`
	MapIconURL         string                                   `json:"mapIconUrl"`
	MapIconPrompt      string                                   `json:"mapIconPrompt"`
	GatherRequirements []resourceGatherRequirementUpsertRequest `json:"gatherRequirements"`
}

type resourceUpsertRequest struct {
	ZoneID             string                                   `json:"zoneId"`
	ZoneKind           *string                                  `json:"zoneKind"`
	ResourceTypeID     string                                   `json:"resourceTypeId"`
	GatherRequirements []resourceGatherRequirementUpsertRequest `json:"gatherRequirements"`
	Quantity           *int                                     `json:"quantity"`
	Latitude           *float64                                 `json:"latitude"`
	Longitude          *float64                                 `json:"longitude"`
}

type resourceGatherRequirementUpsertRequest struct {
	MinLevel                *int `json:"minLevel"`
	MaxLevel                *int `json:"maxLevel"`
	RequiredInventoryItemID int  `json:"requiredInventoryItemId"`
}

type resourceTypeInventoryItemSyncConflict struct {
	InventoryItemID       int      `json:"inventoryItemId"`
	InventoryItemName     string   `json:"inventoryItemName"`
	MatchingResourceTypes []string `json:"matchingResourceTypes"`
}

type resourceTypeInventoryItemSyncSummary struct {
	TotalItemCount      int                                     `json:"totalItemCount"`
	UpdatedCount        int                                     `json:"updatedCount"`
	AlreadyMatchedCount int                                     `json:"alreadyMatchedCount"`
	UnmatchedCount      int                                     `json:"unmatchedCount"`
	AmbiguousCount      int                                     `json:"ambiguousCount"`
	AmbiguousItems      []resourceTypeInventoryItemSyncConflict `json:"ambiguousItems"`
}

type resourceRequirementGenerationBand struct {
	Label      string
	MinLevel   int
	MaxLevel   int
	RarityTier string
	NamePrefix string
}

type resourceRequirementToolProfile struct {
	Noun            string
	DescriptionStem string
}

type resourceRequirementGenerationResponse struct {
	Resource       *models.Resource       `json:"resource"`
	CreatedItems   []models.InventoryItem `json:"createdItems"`
	ReusedItems    []models.InventoryItem `json:"reusedItems"`
	GeneratedCount int                    `json:"generatedCount"`
	ReusedCount    int                    `json:"reusedCount"`
	Message        string                 `json:"message"`
}

type resourceTypeRequirementGenerationResponse struct {
	ResourceType         *models.ResourceType   `json:"resourceType"`
	UpdatedResources     []models.Resource      `json:"updatedResources"`
	CreatedItems         []models.InventoryItem `json:"createdItems"`
	ReusedItems          []models.InventoryItem `json:"reusedItems"`
	UpdatedResourceCount int                    `json:"updatedResourceCount"`
	GeneratedCount       int                    `json:"generatedCount"`
	ReusedCount          int                    `json:"reusedCount"`
	Message              string                 `json:"message"`
}

var defaultResourceRequirementGenerationBands = []resourceRequirementGenerationBand{
	{Label: "starter", MinLevel: 1, MaxLevel: 20, RarityTier: "Common", NamePrefix: "Apprentice"},
	{Label: "journeyman", MinLevel: 21, MaxLevel: 40, RarityTier: "Uncommon", NamePrefix: "Journeyman"},
	{Label: "expert", MinLevel: 41, MaxLevel: 60, RarityTier: "Uncommon", NamePrefix: "Expert"},
	{Label: "master", MinLevel: 61, MaxLevel: 80, RarityTier: "Epic", NamePrefix: "Masterwork"},
	{Label: "legend", MinLevel: 81, MaxLevel: 100, RarityTier: "Mythic", NamePrefix: "Grandmaster"},
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

func humanizeNormalizedSlug(value string) string {
	normalized := normalizeResourceTypeSlug(value)
	if normalized == "" {
		return ""
	}
	parts := strings.Split(normalized, "-")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func resourceTypeMatchKeys(resourceType models.ResourceType) []string {
	keys := make([]string, 0, 3)
	seen := map[string]struct{}{}
	for _, candidate := range []string{
		strings.ToLower(strings.TrimSpace(resourceType.Name)),
		strings.ToLower(strings.TrimSpace(resourceType.Slug)),
		normalizeResourceTypeSlug(resourceType.Name),
	} {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		keys = append(keys, candidate)
	}
	return keys
}

func resourceRequirementToolProfileForType(resourceType *models.ResourceType) resourceRequirementToolProfile {
	slug := ""
	name := ""
	if resourceType != nil {
		slug = normalizeResourceTypeSlug(resourceType.Slug)
		if slug == "" {
			slug = normalizeResourceTypeSlug(resourceType.Name)
		}
		name = strings.TrimSpace(resourceType.Name)
	}
	switch slug {
	case "mining":
		return resourceRequirementToolProfile{
			Noun:            "Pickaxe",
			DescriptionStem: "built to crack ore seams and mineral veins",
		}
	case "herbalism":
		return resourceRequirementToolProfile{
			Noun:            "Herbalist Kit",
			DescriptionStem: "packed for careful harvesting of roots, herbs, and blooms",
		}
	case "logging":
		return resourceRequirementToolProfile{
			Noun:            "Hatchet",
			DescriptionStem: "balanced for felling timber and splitting tough bark",
		}
	case "skinning":
		return resourceRequirementToolProfile{
			Noun:            "Skinning Knife",
			DescriptionStem: "sharpened for clean field dressing and hide work",
		}
	case "fishing":
		return resourceRequirementToolProfile{
			Noun:            "Fishing Rod",
			DescriptionStem: "rigged for landing sturdy river and coastal catches",
		}
	default:
		displayName := strings.TrimSpace(name)
		if displayName == "" {
			displayName = humanizeNormalizedSlug(slug)
		}
		if displayName == "" {
			displayName = "Gathering"
		}
		return resourceRequirementToolProfile{
			Noun:            fmt.Sprintf("%s Tool", displayName),
			DescriptionStem: fmt.Sprintf("made for gathering %s nodes", strings.ToLower(displayName)),
		}
	}
}

func generatedResourceRequirementTags(resourceTypeSlug string, band resourceRequirementGenerationBand) []string {
	return parseInventoryInternalTags([]string{
		"gathering-tool",
		"resource-requirement",
		fmt.Sprintf("tool-for-%s", resourceTypeSlug),
		fmt.Sprintf("resource-band-%d-%d", band.MinLevel, band.MaxLevel),
	})
}

func inventoryItemHasNormalizedTags(item models.InventoryItem, tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	itemTags := parseInventoryInternalTags([]string(item.InternalTags))
	if len(itemTags) == 0 {
		return false
	}
	tagSet := make(map[string]struct{}, len(itemTags))
	for _, tag := range itemTags {
		tagSet[string(tag)] = struct{}{}
	}
	for _, tag := range tags {
		if _, exists := tagSet[tag]; !exists {
			return false
		}
	}
	return true
}

func findExistingGeneratedResourceRequirementItem(
	items []models.InventoryItem,
	requiredTags []string,
	itemLevel int,
) *models.InventoryItem {
	for index := range items {
		item := &items[index]
		if item.Archived || item.ID <= 0 {
			continue
		}
		if item.ItemLevel != itemLevel {
			continue
		}
		if !inventoryItemHasNormalizedTags(*item, requiredTags) {
			continue
		}
		return item
	}
	return nil
}

func resourceNameForDisplay(resourceType *models.ResourceType) string {
	if resourceType == nil {
		return "resource"
	}
	resourceName := strings.TrimSpace(resourceType.Name)
	if resourceName == "" {
		resourceName = humanizeNormalizedSlug(resourceType.Slug)
	}
	if resourceName == "" {
		resourceName = "resource"
	}
	return resourceName
}

func resourceTypeWithGatherRequirementsUpdate(
	resourceType *models.ResourceType,
	gatherRequirements []models.ResourceGatherRequirement,
) *models.ResourceType {
	if resourceType == nil {
		return &models.ResourceType{
			GatherRequirements: gatherRequirements,
		}
	}
	return &models.ResourceType{
		Name:               resourceType.Name,
		Slug:               resourceType.Slug,
		Description:        resourceType.Description,
		MapIconURL:         resourceType.MapIconURL,
		MapIconPrompt:      resourceType.MapIconPrompt,
		GatherRequirements: gatherRequirements,
	}
}

func buildGeneratedResourceRequirementItemRequest(
	resourceType *models.ResourceType,
	band resourceRequirementGenerationBand,
) inventoryItemUpsertRequest {
	resourceTypeSlug := ""
	resourceTypeName := ""
	if resourceType != nil {
		resourceTypeSlug = normalizeResourceTypeSlug(resourceType.Slug)
		if resourceTypeSlug == "" {
			resourceTypeSlug = normalizeResourceTypeSlug(resourceType.Name)
		}
		resourceTypeName = strings.TrimSpace(resourceType.Name)
	}
	if resourceTypeName == "" {
		resourceTypeName = humanizeNormalizedSlug(resourceTypeSlug)
	}
	if resourceTypeName == "" {
		resourceTypeName = "resource"
	}

	toolProfile := resourceRequirementToolProfileForType(resourceType)
	itemLevel := band.MinLevel
	name := fmt.Sprintf("%s %s", band.NamePrefix, toolProfile.Noun)
	flavorText := fmt.Sprintf(
		"A %s %s for adventurers entering level %d gathering.",
		strings.ToLower(band.NamePrefix),
		toolProfile.DescriptionStem,
		band.MinLevel,
	)
	effectText := fmt.Sprintf(
		"Required to gather %s resources for characters level %d-%d.",
		strings.ToLower(resourceTypeName),
		band.MinLevel,
		band.MaxLevel,
	)

	return inventoryItemUpsertRequest{
		Name:          name,
		FlavorText:    flavorText,
		EffectText:    effectText,
		RarityTier:    band.RarityTier,
		IsCaptureType: false,
		ItemLevel:     &itemLevel,
		InternalTags:  generatedResourceRequirementTags(resourceTypeSlug, band),
	}
}

func (s *server) buildGeneratedResourceGatherRequirements(
	ctx *gin.Context,
	resourceType *models.ResourceType,
) ([]models.InventoryItem, []models.InventoryItem, []models.ResourceGatherRequirement, error) {
	resourceTypeSlug := ""
	if resourceType != nil {
		resourceTypeSlug = normalizeResourceTypeSlug(resourceType.Slug)
		if resourceTypeSlug == "" {
			resourceTypeSlug = normalizeResourceTypeSlug(resourceType.Name)
		}
	}
	if resourceTypeSlug == "" {
		return nil, nil, nil, fmt.Errorf("resource type is missing a usable slug")
	}

	activeItems, err := s.dbClient.InventoryItem().FindAllActiveInventoryItems(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	createdItems := make([]models.InventoryItem, 0, len(defaultResourceRequirementGenerationBands))
	reusedItems := make([]models.InventoryItem, 0, len(defaultResourceRequirementGenerationBands))
	gatherRequirements := make([]models.ResourceGatherRequirement, 0, len(defaultResourceRequirementGenerationBands))

	for _, band := range defaultResourceRequirementGenerationBands {
		requiredTags := generatedResourceRequirementTags(resourceTypeSlug, band)
		item := findExistingGeneratedResourceRequirementItem(activeItems, requiredTags, band.MinLevel)
		if item == nil {
			request := buildGeneratedResourceRequirementItemRequest(resourceType, band)
			normalizedItem, err := s.normalizeInventoryItemUpsertRequest(ctx, request, nil)
			if err != nil {
				return nil, nil, nil, err
			}
			if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, normalizedItem); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create inventory item: %w", err)
			}
			if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, normalizedItem.ID, map[string]interface{}{
				"image_generation_status": models.InventoryImageGenerationStatusQueued,
				"image_generation_error":  "",
			}); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to queue inventory item image generation: %w", err)
			}
			if err := s.enqueueInventoryItemImageGeneration(
				ctx,
				normalizedItem.ID,
				normalizedItem.Name,
				normalizedItem.FlavorText,
				normalizedItem.RarityTier,
			); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to queue inventory item image generation: %w", err)
			}
			item, err = s.dbClient.InventoryItem().FindInventoryItemByID(ctx, normalizedItem.ID)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to fetch generated inventory item: %w", err)
			}
			if item == nil {
				return nil, nil, nil, fmt.Errorf("generated inventory item could not be loaded")
			}
			createdItems = append(createdItems, *item)
			activeItems = append(activeItems, *item)
		} else {
			reusedItems = append(reusedItems, *item)
		}

		gatherRequirements = append(gatherRequirements, models.ResourceGatherRequirement{
			MinLevel:                band.MinLevel,
			MaxLevel:                band.MaxLevel,
			RequiredInventoryItemID: item.ID,
			RequiredInventoryItem:   item,
		})
	}

	return createdItems, reusedItems, gatherRequirements, nil
}

func (s *server) updateResourceTypeGatherRequirements(
	ctx *gin.Context,
	resourceType *models.ResourceType,
	gatherRequirements []models.ResourceGatherRequirement,
) (*models.ResourceType, error) {
	if resourceType == nil {
		return nil, fmt.Errorf("resource type is required")
	}
	updates := resourceTypeWithGatherRequirementsUpdate(resourceType, gatherRequirements)
	if err := s.dbClient.ResourceType().Update(ctx, resourceType.ID, updates); err != nil {
		return nil, err
	}
	return s.dbClient.ResourceType().FindByID(ctx, resourceType.ID)
}

func (s *server) findResourcesForType(
	ctx *gin.Context,
	resourceTypeID uuid.UUID,
) ([]models.Resource, error) {
	allResources, err := s.dbClient.Resource().FindAll(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]models.Resource, 0)
	for _, resource := range allResources {
		if resource.Invalidated || resource.ResourceTypeID != resourceTypeID {
			continue
		}
		filtered = append(filtered, resource)
	}
	return filtered, nil
}

func buildResourceTypeMatchIndex(resourceTypes []models.ResourceType) map[string][]models.ResourceType {
	index := make(map[string][]models.ResourceType, len(resourceTypes))
	for _, resourceType := range resourceTypes {
		for _, key := range resourceTypeMatchKeys(resourceType) {
			index[key] = append(index[key], resourceType)
		}
	}
	return index
}

func matchedResourceTypesForInventoryItem(
	item models.InventoryItem,
	resourceTypeIndex map[string][]models.ResourceType,
) []models.ResourceType {
	if len(resourceTypeIndex) == 0 {
		return nil
	}

	normalizedTags := parseInventoryInternalTags([]string(item.InternalTags))
	if len(normalizedTags) == 0 {
		return nil
	}

	matchesByID := make(map[uuid.UUID]models.ResourceType)
	for _, tag := range normalizedTags {
		for _, resourceType := range resourceTypeIndex[string(tag)] {
			matchesByID[resourceType.ID] = resourceType
		}
	}
	if len(matchesByID) == 0 {
		return nil
	}

	matches := make([]models.ResourceType, 0, len(matchesByID))
	for _, resourceType := range matchesByID {
		matches = append(matches, resourceType)
	}
	sort.Slice(matches, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(matches[i].Name))
		rightName := strings.ToLower(strings.TrimSpace(matches[j].Name))
		if leftName == rightName {
			return strings.ToLower(strings.TrimSpace(matches[i].Slug)) <
				strings.ToLower(strings.TrimSpace(matches[j].Slug))
		}
		return leftName < rightName
	})
	return matches
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

func decodeResourceTypeMapIconPayload(encoded string) ([]byte, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return nil, fmt.Errorf("image payload was empty")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return downloadResourceTypeMapIconSource(trimmed)
	}

	if strings.HasPrefix(trimmed, "[") {
		var payload []string
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			for _, entry := range payload {
				if strings.TrimSpace(entry) == "" {
					continue
				}
				return decodeResourceTypeMapIconPayload(entry)
			}
			return nil, fmt.Errorf("image payload array contained no data")
		}
	}

	if strings.HasPrefix(trimmed, "{") {
		var payload struct {
			Data []struct {
				B64JSON string `json:"b64_json"`
				URL     string `json:"url"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			for _, entry := range payload.Data {
				if strings.TrimSpace(entry.B64JSON) != "" {
					return decodeResourceTypeMapIconPayload(entry.B64JSON)
				}
				if strings.TrimSpace(entry.URL) != "" {
					return decodeResourceTypeMapIconPayload(entry.URL)
				}
			}
			return nil, fmt.Errorf("image payload object contained no data")
		}
	}

	if strings.HasPrefix(trimmed, "data:") {
		if comma := strings.Index(trimmed, ","); comma != -1 {
			trimmed = trimmed[comma+1:]
		}
	}

	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(trimmed)
		if err != nil {
			continue
		}
		if len(decoded) == 0 {
			return nil, fmt.Errorf("decoded image was empty")
		}
		return decoded, nil
	}

	return nil, fmt.Errorf("failed to decode image payload as base64")
}

func downloadResourceTypeMapIconSource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("downloaded image was empty")
	}
	return body, nil
}

func (s *server) uploadResourceTypeMapIcon(resourceTypeID uuid.UUID, imageBytes []byte) (string, error) {
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("no image data provided")
	}

	imageFormat, err := util.DetectImageFormat(imageBytes)
	if err != nil {
		return "", err
	}

	imageExtension, err := util.GetImageExtension(imageFormat)
	if err != nil {
		return "", err
	}

	imageName := fmt.Sprintf(
		"resource-types/%s-map-icon-%d.%s",
		resourceTypeID.String(),
		time.Now().UnixNano(),
		imageExtension,
	)
	return s.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
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

func activeResourceGatherRequirementForLevel(
	requirements []models.ResourceGatherRequirement,
	level int,
) *models.ResourceGatherRequirement {
	normalizedLevel := level
	if normalizedLevel < 1 {
		normalizedLevel = 1
	}
	for index := range requirements {
		requirement := &requirements[index]
		if requirement.MinLevel <= normalizedLevel && normalizedLevel <= requirement.MaxLevel {
			return requirement
		}
	}
	return nil
}

func resourceGatherRequirementItemName(requirement *models.ResourceGatherRequirement) string {
	if requirement == nil {
		return "required equipment"
	}
	if requirement.RequiredInventoryItem != nil {
		name := strings.TrimSpace(requirement.RequiredInventoryItem.Name)
		if name != "" {
			return name
		}
	}
	if requirement.RequiredInventoryItemID > 0 {
		return fmt.Sprintf("item #%d", requirement.RequiredInventoryItemID)
	}
	return "required equipment"
}

func userOwnsInventoryItem(
	ownedItems []models.OwnedInventoryItem,
	inventoryItemID int,
) bool {
	if inventoryItemID <= 0 {
		return false
	}
	for _, ownedItem := range ownedItems {
		if ownedItem.InventoryItemID == inventoryItemID && ownedItem.Quantity > 0 {
			return true
		}
	}
	return false
}

func (s *server) validateResourceGatherRequirements(
	ctx context.Context,
	requests []resourceGatherRequirementUpsertRequest,
) ([]models.ResourceGatherRequirement, error) {
	if requests == nil {
		return nil, nil
	}
	if len(requests) == 0 {
		return []models.ResourceGatherRequirement{}, nil
	}

	requirements := make([]models.ResourceGatherRequirement, 0, len(requests))
	for index, request := range requests {
		if request.MinLevel == nil || request.MaxLevel == nil {
			return nil, fmt.Errorf("gatherRequirements[%d] must include minLevel and maxLevel", index)
		}
		if *request.MinLevel < 1 || *request.MaxLevel < 1 {
			return nil, fmt.Errorf("gatherRequirements[%d] levels must be 1 or greater", index)
		}
		if *request.MaxLevel < *request.MinLevel {
			return nil, fmt.Errorf("gatherRequirements[%d] maxLevel must be greater than or equal to minLevel", index)
		}
		if request.RequiredInventoryItemID <= 0 {
			return nil, fmt.Errorf("gatherRequirements[%d] requiredInventoryItemId must be positive", index)
		}

		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, request.RequiredInventoryItemID)
		if err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("gatherRequirements[%d] requiredInventoryItemId not found", index)
			}
			return nil, err
		}
		if item == nil || item.Archived {
			return nil, fmt.Errorf("gatherRequirements[%d] requiredInventoryItemId must reference an active inventory item", index)
		}

		requirements = append(requirements, models.ResourceGatherRequirement{
			MinLevel:                *request.MinLevel,
			MaxLevel:                *request.MaxLevel,
			RequiredInventoryItemID: item.ID,
			RequiredInventoryItem:   item,
		})
	}

	sort.Slice(requirements, func(i, j int) bool {
		if requirements[i].MinLevel == requirements[j].MinLevel {
			if requirements[i].MaxLevel == requirements[j].MaxLevel {
				return requirements[i].RequiredInventoryItemID < requirements[j].RequiredInventoryItemID
			}
			return requirements[i].MaxLevel < requirements[j].MaxLevel
		}
		return requirements[i].MinLevel < requirements[j].MinLevel
	})

	for index := 1; index < len(requirements); index++ {
		previous := requirements[index-1]
		current := requirements[index]
		if current.MinLevel <= previous.MaxLevel {
			return nil, fmt.Errorf(
				"gather requirement level bands cannot overlap (%d-%d overlaps %d-%d)",
				previous.MinLevel,
				previous.MaxLevel,
				current.MinLevel,
				current.MaxLevel,
			)
		}
	}

	return requirements, nil
}

func normalizedGatherRewardItemLevel(item models.InventoryItem) int {
	if item.ItemLevel > 0 {
		return item.ItemLevel
	}
	return 1
}

func gatherRewardCandidatesForResourceType(
	resourceTypeID uuid.UUID,
	items []models.InventoryItem,
) []models.InventoryItem {
	filtered := make([]models.InventoryItem, 0, len(items))
	for _, item := range items {
		if item.ID <= 0 || item.Archived || item.IsCaptureType {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.RarityTier), "Not Droppable") {
			continue
		}
		if !resourceTypeIDsMatch(&item, resourceTypeID) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool {
		leftLevel := normalizedGatherRewardItemLevel(filtered[i])
		rightLevel := normalizedGatherRewardItemLevel(filtered[j])
		if leftLevel == rightLevel {
			return filtered[i].ID < filtered[j].ID
		}
		return leftLevel < rightLevel
	})
	return filtered
}

func selectGatherRewardInventoryItem(
	resourceTypeID uuid.UUID,
	userLevel int,
	items []models.InventoryItem,
	rng *rand.Rand,
) (*models.InventoryItem, error) {
	candidates := gatherRewardCandidatesForResourceType(resourceTypeID, items)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no active inventory items are configured for this resource type")
	}

	targetLevel := userLevel
	if targetLevel < 1 {
		targetLevel = 1
	}
	minLevel := targetLevel - 10
	if minLevel < 1 {
		minLevel = 1
	}
	maxLevel := targetLevel + 10

	inBand := make([]models.InventoryItem, 0, len(candidates))
	closest := make([]models.InventoryItem, 0, len(candidates))
	bestDelta := -1
	for _, item := range candidates {
		itemLevel := normalizedGatherRewardItemLevel(item)
		if itemLevel >= minLevel && itemLevel <= maxLevel {
			inBand = append(inBand, item)
			continue
		}
		delta := itemLevel - targetLevel
		if delta < 0 {
			delta = -delta
		}
		if bestDelta == -1 || delta < bestDelta {
			bestDelta = delta
			closest = []models.InventoryItem{item}
			continue
		}
		if delta == bestDelta {
			closest = append(closest, item)
		}
	}

	pool := inBand
	if len(pool) == 0 {
		pool = closest
	}
	if len(pool) == 0 {
		pool = candidates
	}
	if len(pool) == 0 {
		return nil, fmt.Errorf("no gather reward candidates were available")
	}

	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	selected := pool[rng.Intn(len(pool))]
	return &selected, nil
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
	gatherRequirements, err := s.validateResourceGatherRequirements(ctx, requestBody.GatherRequirements)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resourceType := &models.ResourceType{
		Name:               name,
		Slug:               slug,
		Description:        strings.TrimSpace(requestBody.Description),
		MapIconURL:         strings.TrimSpace(requestBody.MapIconURL),
		MapIconPrompt:      strings.TrimSpace(requestBody.MapIconPrompt),
		GatherRequirements: gatherRequirements,
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
	gatherRequirements := existing.GatherRequirements
	if requestBody.GatherRequirements != nil {
		gatherRequirements, err = s.validateResourceGatherRequirements(ctx, requestBody.GatherRequirements)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	updates := &models.ResourceType{
		Name:               name,
		Slug:               slug,
		Description:        strings.TrimSpace(requestBody.Description),
		MapIconURL:         strings.TrimSpace(requestBody.MapIconURL),
		MapIconPrompt:      strings.TrimSpace(requestBody.MapIconPrompt),
		GatherRequirements: gatherRequirements,
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
	imagePayload, err := s.deepPriest.GenerateImage(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	imageBytes, err := decodeResourceTypeMapIconPayload(imagePayload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode generated icon: " + err.Error()})
		return
	}
	mapIconURL, err := s.uploadResourceTypeMapIcon(resourceType.ID, imageBytes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload generated icon: " + err.Error()})
		return
	}

	updates := &models.ResourceType{
		Name:               resourceType.Name,
		Slug:               resourceType.Slug,
		Description:        resourceType.Description,
		MapIconURL:         mapIconURL,
		MapIconPrompt:      prompt,
		GatherRequirements: resourceType.GatherRequirements,
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

func (s *server) generateResourceTypeRequirementItems(ctx *gin.Context) {
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

	createdItems, reusedItems, gatherRequirements, err := s.buildGeneratedResourceGatherRequirements(ctx, resourceType)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "usable slug") {
			statusCode = http.StatusBadRequest
		}
		ctx.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	updatedResourceType, err := s.updateResourceTypeGatherRequirements(ctx, resourceType, gatherRequirements)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource type requirements: " + err.Error()})
		return
	}
	updatedResources, err := s.findResourcesForType(ctx, resourceType.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated resources: " + err.Error()})
		return
	}

	resourceName := resourceNameForDisplay(resourceType)
	ctx.JSON(http.StatusOK, resourceTypeRequirementGenerationResponse{
		ResourceType:         updatedResourceType,
		UpdatedResources:     updatedResources,
		CreatedItems:         createdItems,
		ReusedItems:          reusedItems,
		UpdatedResourceCount: len(updatedResources),
		GeneratedCount:       len(createdItems),
		ReusedCount:          len(reusedItems),
		Message: fmt.Sprintf(
			"Generated %d required item(s), reused %d, and applied the default bands to %d %s node(s).",
			len(createdItems),
			len(reusedItems),
			len(updatedResources),
			strings.ToLower(resourceName),
		),
	})
}

func (s *server) syncResourceTypesToInventoryItems(ctx *gin.Context) {
	resourceTypes, err := s.dbClient.ResourceType().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load resource types: " + err.Error()})
		return
	}

	inventoryItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load inventory items: " + err.Error()})
		return
	}

	resourceTypeIndex := buildResourceTypeMatchIndex(resourceTypes)
	summary := resourceTypeInventoryItemSyncSummary{
		TotalItemCount: len(inventoryItems),
		AmbiguousItems: []resourceTypeInventoryItemSyncConflict{},
	}

	for _, item := range inventoryItems {
		matches := matchedResourceTypesForInventoryItem(item, resourceTypeIndex)
		if len(matches) == 0 {
			summary.UnmatchedCount++
			continue
		}

		if item.ResourceTypeID != nil {
			for _, match := range matches {
				if match.ID == *item.ResourceTypeID {
					summary.AlreadyMatchedCount++
					goto nextItem
				}
			}
		}

		if len(matches) > 1 {
			summary.AmbiguousCount++
			if len(summary.AmbiguousItems) < 10 {
				matchNames := make([]string, 0, len(matches))
				for _, match := range matches {
					name := strings.TrimSpace(match.Name)
					if name == "" {
						name = strings.TrimSpace(match.Slug)
					}
					matchNames = append(matchNames, name)
				}
				itemName := strings.TrimSpace(item.Name)
				if itemName == "" {
					itemName = fmt.Sprintf("Item #%d", item.ID)
				}
				summary.AmbiguousItems = append(summary.AmbiguousItems, resourceTypeInventoryItemSyncConflict{
					InventoryItemID:       item.ID,
					InventoryItemName:     itemName,
					MatchingResourceTypes: matchNames,
				})
			}
			continue
		}

		if err := s.dbClient.InventoryItem().UpdateInventoryItem(ctx, item.ID, map[string]interface{}{
			"resource_type_id": matches[0].ID,
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update inventory item resource type: " + err.Error()})
			return
		}
		summary.UpdatedCount++

	nextItem:
	}

	ctx.JSON(http.StatusOK, summary)
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
	if requestBody.GatherRequirements != nil {
		gatherRequirements, err := s.validateResourceGatherRequirements(ctx, requestBody.GatherRequirements)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resourceType, err = s.updateResourceTypeGatherRequirements(ctx, resourceType, gatherRequirements)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource type requirements: " + err.Error()})
			return
		}
	}

	resource := &models.Resource{
		ZoneID:         zoneID,
		ZoneKind:       normalizeZoneKindRequest(requestBody.ZoneKind),
		ResourceTypeID: resourceType.ID,
		Quantity:       *requestBody.Quantity,
		Latitude:       *requestBody.Latitude,
		Longitude:      *requestBody.Longitude,
		Invalidated:    false,
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
	if requestBody.GatherRequirements != nil {
		gatherRequirements, err := s.validateResourceGatherRequirements(ctx, requestBody.GatherRequirements)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resourceType, err = s.updateResourceTypeGatherRequirements(ctx, resourceType, gatherRequirements)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource type requirements: " + err.Error()})
			return
		}
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
		ZoneID:         zoneID,
		ZoneKind:       mergeZoneKindRequest(requestBody.ZoneKind, existing.ZoneKind),
		ResourceTypeID: resourceType.ID,
		Quantity:       quantity,
		Latitude:       latitude,
		Longitude:      longitude,
		Invalidated:    existing.Invalidated,
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

func (s *server) generateResourceRequirementItems(ctx *gin.Context) {
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

	resourceType := &resource.ResourceType
	createdItems, reusedItems, gatherRequirements, err := s.buildGeneratedResourceGatherRequirements(ctx, resourceType)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "usable slug") {
			statusCode = http.StatusBadRequest
		}
		ctx.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.updateResourceTypeGatherRequirements(ctx, resourceType, gatherRequirements); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update resource type requirements: " + err.Error()})
		return
	}

	updatedResource, err := s.dbClient.Resource().FindByID(ctx, resource.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated resource: " + err.Error()})
		return
	}

	resourceName := resourceNameForDisplay(resourceType)

	ctx.JSON(http.StatusOK, resourceRequirementGenerationResponse{
		Resource:       updatedResource,
		CreatedItems:   createdItems,
		ReusedItems:    reusedItems,
		GeneratedCount: len(createdItems),
		ReusedCount:    len(reusedItems),
		Message: fmt.Sprintf(
			"Generated %d required item(s) and reused %d for %s. All nodes of this type now inherit those bands.",
			len(createdItems),
			len(reusedItems),
			resourceName,
		),
	})
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
	if resource.Quantity < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "resource is not configured for gathering"})
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

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	activeRequirement := activeResourceGatherRequirementForLevel(
		resource.GatherRequirements,
		userLevel.Level,
	)
	if activeRequirement != nil {
		ownedItems, err := s.dbClient.InventoryItem().GetUsersItems(ctx, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !userOwnsInventoryItem(ownedItems, activeRequirement.RequiredInventoryItemID) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf(
					"requires %s to gather at your current level",
					resourceGatherRequirementItemName(activeRequirement),
				),
			})
			return
		}
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

	inventoryItems, err := s.dbClient.InventoryItem().FindAllActiveInventoryItems(ctx)
	if err != nil {
		_ = s.dbClient.Exec(ctx, fmt.Sprintf("DELETE FROM user_resource_gatherings WHERE id = '%s'", gathering.ID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	selectedItem, err := selectGatherRewardInventoryItem(
		resource.ResourceTypeID,
		userLevel.Level,
		inventoryItems,
		nil,
	)
	if err != nil {
		_ = s.dbClient.Exec(ctx, fmt.Sprintf("DELETE FROM user_resource_gatherings WHERE id = '%s'", gathering.ID))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rewardExperience := modestResourceExperienceReward(userLevel)
	itemsAwarded, spellsAwarded, err := s.awardScenarioRewards(
		ctx,
		user.ID,
		rewardExperience,
		0,
		[]scenarioRewardItem{{
			InventoryItemID: selectedItem.ID,
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
