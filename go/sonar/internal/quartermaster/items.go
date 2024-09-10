package quartermaster

type InventoryItem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ImageURL   string `json:"imageUrl"`
	FlavorText string `json:"flavorText"`
	EffectText string `json:"effectText"`
	RarityTier Rarity `json:"rarityTier"`
}

type Rarity string

const (
	RarityCommon   Rarity = "Common"
	RarityUncommon Rarity = "Uncommon"
	RarityEpic     Rarity = "Epic"
	RarityMythic   Rarity = "Mythic"
	NotDroppable   Rarity = "Not Droppable"
)

var PreDefinedItems = []InventoryItem{
	{
		ID:         1,
		Name:       "Cipher of the Laughing Monkey",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/cipher.png",
		FlavorText: "Unearthed in the heart of a dense jungle, this mysterious item lay among countless laughing skeletons.",
		EffectText: "Deploy to sow confusion among your rivals by warping their clue texts into bewildering riddles.",
		RarityTier: "Uncommon",
	},
	{
		ID:         2,
		Name:       "Golden Telescope",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/telescope-better.png",
		FlavorText: "Legend has it that a artificer parted with his sight to create this so that others might see the stars.",
		EffectText: "Instantly reveals a hidden point on the map. Tap this icon next to the \"I'm here!\" button on a hidden points of interest to use it.",
		RarityTier: "Uncommon",
	},
	{
		ID:         3,
		Name:       "Flawed Ruby",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/flawed-ruby.png",
		FlavorText: "This gem is chipped and disfigured, but will still fetch a decent price at market.",
		EffectText: "Instantly captures a tier one challenge. Tap this icon next to the \"Submit Answer\" button on any unlocked tier one challenge to use it.",
		RarityTier: "Uncommon",
	},
	{
		ID:         4,
		Name:       "Ruby",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/ruby.png",
		FlavorText: "A gem, sparkling more red than the blood you had to spill to procure it.",
		EffectText: "Instantly captures a tier two challenge. Tap this icon next to the \"Submit Answer\" button on any unlocked tier two challenge to use it.",
		RarityTier: "Epic",
	},
	{
		ID:         5,
		Name:       "Brilliant Ruby",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/brilliant-ruby.png",
		FlavorText: "You've hit the motherload! This gem will fetch a pirate's ransom.",
		EffectText: "Instantly captures a tier three challenge. Tap this icon next to the \"Submit Answer\" button on any unlocked tier three challenge to use it.",
		RarityTier: "Mythic",
	},
	{
		ID:         6,
		Name:       "Cortez’s Cutlass",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/cortez-cutlass.png",
		FlavorText: "A relic of the high seas, its blade still sharp enough to cut through the thickest of hides.",
		EffectText: "Steal all of another team's items.",
		RarityTier: "Not Droppable",
	},
	{
		ID:         7,
		Name:       "Rusted Musket",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/rusted-musket.png",
		FlavorText: "Found in a shipwreck, its barrel rusted and its stock worn.",
		EffectText: "Use on an opponent to lower their score by 2.",
		RarityTier: "Common",
	},
	{
		ID:         8,
		Name:       "Gold Coin",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/gold-coin.png",
		FlavorText: "A coin of pure gold. The currency of the high seas.",
		EffectText: "Hold in your inventory to increase your score by 1.",
		RarityTier: "Common",
	},
	{
		ID:         9,
		Name:       "Dagger",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/dagger.png",
		FlavorText: "A small, sharp blade. It's not much, but it's better than nothing.",
		EffectText: "Steal one item from an opponent at random.",
		RarityTier: "Epic",
	},
	{
		ID:         10,
		Name:       "Damage",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/bullet-hole.png",
		FlavorText: "You've been shot! Some ale will help.",
		EffectText: "Decreases score by 2 while held in inventory.",
		RarityTier: "Not Droppable",
	},
	{
		ID:         11,
		Name:       "Entseed",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/entseed.png",
		FlavorText: "This seed will grow into an Ent one day. For now, you can just bask in it's life energy.",
		EffectText: "Increase score by 3 and neutralize the effects of Damage while held in inventory.",
		RarityTier: "Not Droppable",
	},
	{
		ID:         12,
		Name:       "Ale",
		ImageURL:   "https://crew-points-of-interest.s3.amazonaws.com/ale.png",
		FlavorText: "A hearty brew, made from the finest ingredients.",
		EffectText: "Removes one damage when drank.",
		RarityTier: "Uncommon",
	},
}
