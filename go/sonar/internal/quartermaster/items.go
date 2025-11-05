package quartermaster

type InventoryItem struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ImageURL      string `json:"imageUrl"`
	FlavorText    string `json:"flavorText"`
	EffectText    string `json:"effectText"`
	RarityTier    Rarity `json:"rarityTier"`
	IsCaptureType bool   `json:"isCaptureType"`
}

type Rarity string

const (
	RarityCommon   Rarity = "Common"
	RarityUncommon Rarity = "Uncommon"
	RarityEpic     Rarity = "Epic"
	RarityMythic   Rarity = "Mythic"
	NotDroppable   Rarity = "Not Droppable"
)

// PreDefinedItems removed - items are now loaded from database
