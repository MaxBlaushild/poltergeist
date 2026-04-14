package models

import (
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
)

type RandomRewardItemGrant struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type RandomRewardPlan struct {
	Experience int                     `json:"experience"`
	Gold       int                     `json:"gold"`
	ItemGrants []RandomRewardItemGrant `json:"itemGrants"`
}

type randomRewardProfile struct {
	xpPerLevel           int
	xpBase               int
	goldPerLevel         int
	goldBase             int
	consumableMinQty     int
	consumableMaxQty     int
	consumableChance     float64
	consumableGuaranteed bool
	equippableChance     float64
	equippableGuaranteed bool
}

func BuildRandomRewardPlan(
	level int,
	size RandomRewardSize,
	seed string,
	inventoryItems []InventoryItem,
) RandomRewardPlan {
	normalizedLevel := level
	if normalizedLevel < 1 {
		normalizedLevel = 1
	}
	normalizedSize := NormalizeRandomRewardSize(string(size))
	profile := randomRewardProfileForSize(normalizedSize)
	rng := rand.New(rand.NewSource(randomRewardSeed(seed, normalizedLevel, normalizedSize)))

	plan := RandomRewardPlan{
		Experience: maxRewardInt(25, normalizedLevel*profile.xpPerLevel+profile.xpBase),
		Gold:       maxRewardInt(10, normalizedLevel*profile.goldPerLevel+profile.goldBase),
		ItemGrants: []RandomRewardItemGrant{},
	}

	consumables := filterRewardItems(inventoryItems, normalizedLevel, false)
	equippables := filterRewardItems(inventoryItems, normalizedLevel, true)

	grantConsumable := profile.consumableGuaranteed
	if !grantConsumable && profile.consumableChance > 0 {
		grantConsumable = rng.Float64() < profile.consumableChance
	}
	if grantConsumable {
		if item := pickRewardItem(rng, consumables, normalizedLevel); item != nil {
			minQty := maxRewardInt(1, profile.consumableMinQty)
			maxQty := maxRewardInt(minQty, profile.consumableMaxQty)
			qty := minQty
			if maxQty > minQty {
				qty = minQty + rng.Intn((maxQty-minQty)+1)
			}
			plan.ItemGrants = append(plan.ItemGrants, RandomRewardItemGrant{
				InventoryItemID: item.ID,
				Quantity:        qty,
			})
		}
	}

	grantEquippable := profile.equippableGuaranteed
	if !grantEquippable && profile.equippableChance > 0 {
		grantEquippable = rng.Float64() < profile.equippableChance
	}
	if grantEquippable {
		if item := pickRewardItem(rng, equippables, normalizedLevel); item != nil {
			plan.ItemGrants = append(plan.ItemGrants, RandomRewardItemGrant{
				InventoryItemID: item.ID,
				Quantity:        1,
			})
		}
	}

	if len(plan.ItemGrants) > 1 {
		quantities := map[int]int{}
		for _, grant := range plan.ItemGrants {
			if grant.InventoryItemID <= 0 || grant.Quantity <= 0 {
				continue
			}
			quantities[grant.InventoryItemID] += grant.Quantity
		}
		ids := make([]int, 0, len(quantities))
		for itemID := range quantities {
			ids = append(ids, itemID)
		}
		sort.Ints(ids)
		merged := make([]RandomRewardItemGrant, 0, len(ids))
		for _, itemID := range ids {
			merged = append(merged, RandomRewardItemGrant{
				InventoryItemID: itemID,
				Quantity:        quantities[itemID],
			})
		}
		plan.ItemGrants = merged
	}

	return plan
}

func randomRewardProfileForSize(size RandomRewardSize) randomRewardProfile {
	switch size {
	case RandomRewardSizeMedium:
		return randomRewardProfile{
			xpPerLevel:           85,
			xpBase:               220,
			goldPerLevel:         16,
			goldBase:             85,
			consumableMinQty:     1,
			consumableMaxQty:     2,
			consumableGuaranteed: true,
			equippableChance:     0.35,
		}
	case RandomRewardSizeLarge:
		return randomRewardProfile{
			xpPerLevel:           120,
			xpBase:               400,
			goldPerLevel:         28,
			goldBase:             180,
			consumableMinQty:     2,
			consumableMaxQty:     3,
			consumableGuaranteed: true,
			equippableGuaranteed: true,
		}
	default:
		return randomRewardProfile{
			xpPerLevel:       55,
			xpBase:           120,
			goldPerLevel:     9,
			goldBase:         30,
			consumableMinQty: 1,
			consumableMaxQty: 1,
			consumableChance: 0.55,
			equippableChance: 0.1,
		}
	}
}

func randomRewardSeed(seed string, level int, size RandomRewardSize) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(seed)))
	_, _ = hasher.Write([]byte{0})
	_, _ = hasher.Write([]byte(size))
	_, _ = hasher.Write([]byte{0})
	_, _ = hasher.Write([]byte{byte(level & 0xff), byte((level >> 8) & 0xff)})
	value := int64(hasher.Sum64())
	if value == 0 {
		value = 1
	}
	return value
}

func filterRewardItems(items []InventoryItem, level int, equippable bool) []InventoryItem {
	filtered := make([]InventoryItem, 0, len(items))
	for _, item := range items {
		if item.ID <= 0 || item.IsCaptureType || item.Archived {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.RarityTier), "Not Droppable") {
			continue
		}
		itemIsEquippable := strings.TrimSpace(derefRewardString(item.EquipSlot)) != ""
		if equippable != itemIsEquippable {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})
	return filtered
}

func pickRewardItem(rng *rand.Rand, items []InventoryItem, level int) *InventoryItem {
	if len(items) == 0 {
		return nil
	}
	near := make([]InventoryItem, 0, len(items))
	withLevels := make([]InventoryItem, 0, len(items))
	for _, item := range items {
		if item.ItemLevel > 0 {
			withLevels = append(withLevels, item)
			if absRewardInt(item.ItemLevel-level) <= 6 {
				near = append(near, item)
			}
		}
	}
	candidates := near
	if len(candidates) == 0 && len(withLevels) > 0 {
		bestDelta := absRewardInt(withLevels[0].ItemLevel - level)
		for _, item := range withLevels[1:] {
			delta := absRewardInt(item.ItemLevel - level)
			if delta < bestDelta {
				bestDelta = delta
			}
		}
		candidates = make([]InventoryItem, 0, len(withLevels))
		for _, item := range withLevels {
			if absRewardInt(item.ItemLevel-level) == bestDelta {
				candidates = append(candidates, item)
			}
		}
	}
	if len(candidates) == 0 {
		candidates = items
	}
	if len(candidates) == 0 {
		return nil
	}
	picked := candidates[rng.Intn(len(candidates))]
	return &picked
}

func derefRewardString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func absRewardInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func maxRewardInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
