package server

import (
	"context"
	stdErrors "errors"
	"math"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const (
	shopModeExplicit         = "explicit"
	shopModeTags             = "tags"
	shopTaggedLevelBandDelta = 15
)

var errShopHasNoInventory = stdErrors.New("shop has no inventory")

type shopInventoryItem struct {
	ItemID int `json:"itemId"`
	Price  int `json:"price"`
}

func normalizeShopMode(raw interface{}) string {
	mode := strings.ToLower(strings.TrimSpace(interfaceToString(raw)))
	switch mode {
	case shopModeTags:
		return shopModeTags
	case shopModeExplicit:
		return shopModeExplicit
	default:
		return ""
	}
}

func parseShopInventoryItems(raw interface{}) []shopInventoryItem {
	switch v := raw.(type) {
	case []shopInventoryItem:
		copied := make([]shopInventoryItem, len(v))
		copy(copied, v)
		return copied
	}

	rawEntries, ok := raw.([]interface{})
	if !ok {
		return []shopInventoryItem{}
	}

	items := make([]shopInventoryItem, 0, len(rawEntries))
	seen := map[int]struct{}{}
	for _, entry := range rawEntries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		itemID, ok := interfaceToInt(entryMap["itemId"])
		if !ok || itemID <= 0 {
			continue
		}
		price, ok := interfaceToInt(entryMap["price"])
		if !ok || price < 0 {
			continue
		}
		if _, exists := seen[itemID]; exists {
			continue
		}
		seen[itemID] = struct{}{}
		items = append(items, shopInventoryItem{
			ItemID: itemID,
			Price:  price,
		})
	}
	return items
}

func parseShopItemTags(raw interface{}) []string {
	switch v := raw.(type) {
	case []string:
		rawTags := make([]interface{}, 0, len(v))
		for _, tag := range v {
			rawTags = append(rawTags, tag)
		}
		return parseShopItemTags(rawTags)
	}

	rawTags, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}
	tags := make([]string, 0, len(rawTags))
	seen := map[string]struct{}{}
	for _, entry := range rawTags {
		tag := strings.ToLower(strings.TrimSpace(interfaceToString(entry)))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func normalizeShopMetadata(raw map[string]interface{}) models.MetadataJSONB {
	normalized := models.MetadataJSONB{}
	if raw == nil {
		normalized["shopMode"] = shopModeExplicit
		normalized["inventory"] = []shopInventoryItem{}
		normalized["shopItemTags"] = []string{}
		return normalized
	}

	mode := normalizeShopMode(raw["shopMode"])
	inventory := parseShopInventoryItems(raw["inventory"])
	tags := parseShopItemTags(raw["shopItemTags"])

	if mode == "" {
		if len(tags) > 0 {
			mode = shopModeTags
		} else {
			mode = shopModeExplicit
		}
	}

	normalized["shopMode"] = mode
	normalized["inventory"] = inventory
	normalized["shopItemTags"] = tags
	return normalized
}

func interfaceToString(raw interface{}) string {
	switch v := raw.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func interfaceToInt(raw interface{}) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func inventoryItemEffectiveLevel(item models.InventoryItem) int {
	if item.ItemLevel > 0 {
		return item.ItemLevel
	}
	return 1
}

func defaultShopPriceForItem(item models.InventoryItem) int {
	if item.SellValue != nil && *item.SellValue > 0 {
		return maxInt(*item.SellValue*2, 1)
	}

	level := inventoryItemEffectiveLevel(item)
	base := float64(level * 10)
	rarity := strings.ToLower(strings.TrimSpace(item.RarityTier))
	rarityMultiplier := 1.0
	switch rarity {
	case "uncommon":
		rarityMultiplier = 1.25
	case "rare":
		rarityMultiplier = 1.5
	case "epic":
		rarityMultiplier = 1.8
	case "mythic":
		rarityMultiplier = 2.25
	}
	return maxInt(int(math.Round(base*rarityMultiplier)), 1)
}

func itemHasAnyInternalTag(item models.InventoryItem, tagSet map[string]struct{}) bool {
	for _, rawTag := range []string(item.InternalTags) {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		if _, ok := tagSet[tag]; ok {
			return true
		}
	}
	return false
}

func resolveTaggedShopInventory(items []models.InventoryItem, userLevel int, tags []string) []shopInventoryItem {
	if len(tags) == 0 {
		return []shopInventoryItem{}
	}

	minLevel := maxInt(1, userLevel-shopTaggedLevelBandDelta)
	maxLevel := maxInt(minLevel, userLevel+shopTaggedLevelBandDelta)
	tagSet := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		tagSet[normalized] = struct{}{}
	}
	if len(tagSet) == 0 {
		return []shopInventoryItem{}
	}

	type candidate struct {
		itemID int
		level  int
		price  int
	}
	candidates := make([]candidate, 0)
	seen := map[int]struct{}{}
	for _, item := range items {
		if item.ID <= 0 {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		if !itemHasAnyInternalTag(item, tagSet) {
			continue
		}
		level := inventoryItemEffectiveLevel(item)
		if level < minLevel || level > maxLevel {
			continue
		}
		seen[item.ID] = struct{}{}
		candidates = append(candidates, candidate{
			itemID: item.ID,
			level:  level,
			price:  defaultShopPriceForItem(item),
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].level != candidates[j].level {
			return candidates[i].level < candidates[j].level
		}
		return candidates[i].itemID < candidates[j].itemID
	})

	resolved := make([]shopInventoryItem, 0, len(candidates))
	for _, entry := range candidates {
		resolved = append(resolved, shopInventoryItem{
			ItemID: entry.itemID,
			Price:  entry.price,
		})
	}
	return resolved
}

func (s *server) resolveShopInventoryForUser(
	ctx context.Context,
	userID uuid.UUID,
	action *models.CharacterAction,
) ([]shopInventoryItem, error) {
	if action == nil || action.ActionType != models.ActionTypeShop {
		return nil, stdErrors.New("action is not a shop action")
	}

	metadata := action.Metadata
	if metadata == nil {
		metadata = models.MetadataJSONB{}
	}

	mode := normalizeShopMode(metadata["shopMode"])
	inventory := parseShopInventoryItems(metadata["inventory"])
	tags := parseShopItemTags(metadata["shopItemTags"])

	if mode == "" {
		if len(tags) > 0 {
			mode = shopModeTags
		} else {
			mode = shopModeExplicit
		}
	}

	switch mode {
	case shopModeTags:
		if len(tags) == 0 {
			return nil, errShopHasNoInventory
		}
		userLevel, err := s.currentUserLevel(ctx, userID)
		if err != nil {
			return nil, err
		}
		items, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
		if err != nil {
			return nil, err
		}
		return resolveTaggedShopInventory(items, userLevel, tags), nil
	default:
		if len(inventory) == 0 {
			return nil, errShopHasNoInventory
		}
		return inventory, nil
	}
}

func (s *server) hydrateShopActionMetadataForUser(
	ctx context.Context,
	userID uuid.UUID,
	action *models.CharacterAction,
) error {
	if action == nil || action.ActionType != models.ActionTypeShop {
		return nil
	}
	if action.Metadata == nil {
		action.Metadata = models.MetadataJSONB{}
	}

	mode := normalizeShopMode(action.Metadata["shopMode"])
	tags := parseShopItemTags(action.Metadata["shopItemTags"])
	if mode == "" {
		if len(tags) > 0 {
			mode = shopModeTags
		} else {
			mode = shopModeExplicit
		}
	}
	action.Metadata["shopMode"] = mode
	action.Metadata["shopItemTags"] = tags

	inventory, err := s.resolveShopInventoryForUser(ctx, userID, action)
	if err != nil {
		if stdErrors.Is(err, errShopHasNoInventory) {
			action.Metadata["inventory"] = []shopInventoryItem{}
			return nil
		}
		return err
	}
	action.Metadata["inventory"] = inventory
	return nil
}
