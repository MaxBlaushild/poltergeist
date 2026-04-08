package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"sort"
)

type FetchQuestRequirement struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type FetchQuestRequirements []FetchQuestRequirement

func NormalizeFetchQuestRequirements(
	input []FetchQuestRequirement,
) FetchQuestRequirements {
	if len(input) == 0 {
		return FetchQuestRequirements{}
	}

	quantityByItemID := map[int]int{}
	orderByItemID := map[int]int{}
	for index, requirement := range input {
		if requirement.InventoryItemID <= 0 || requirement.Quantity <= 0 {
			continue
		}
		if _, seen := orderByItemID[requirement.InventoryItemID]; !seen {
			orderByItemID[requirement.InventoryItemID] = index
		}
		quantityByItemID[requirement.InventoryItemID] += requirement.Quantity
	}

	if len(quantityByItemID) == 0 {
		return FetchQuestRequirements{}
	}

	itemIDs := make([]int, 0, len(quantityByItemID))
	for inventoryItemID := range quantityByItemID {
		itemIDs = append(itemIDs, inventoryItemID)
	}
	sort.SliceStable(itemIDs, func(i, j int) bool {
		return orderByItemID[itemIDs[i]] < orderByItemID[itemIDs[j]]
	})

	normalized := make(FetchQuestRequirements, 0, len(itemIDs))
	for _, inventoryItemID := range itemIDs {
		normalized = append(normalized, FetchQuestRequirement{
			InventoryItemID: inventoryItemID,
			Quantity:        quantityByItemID[inventoryItemID],
		})
	}
	return normalized
}

func (r FetchQuestRequirements) Value() (driver.Value, error) {
	return json.Marshal(NormalizeFetchQuestRequirements(r))
}

func (r *FetchQuestRequirements) Scan(value interface{}) error {
	if value == nil {
		*r = FetchQuestRequirements{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan FetchQuestRequirements: value is not []byte")
	}

	var decoded []FetchQuestRequirement
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = NormalizeFetchQuestRequirements(decoded)
	return nil
}
