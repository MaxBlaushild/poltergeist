package quartermaster

type InventoryItem struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	ImageURL      string        `json:"imageUrl"`
	FlavorText    string        `json:"flavorText"`
	EffectText    string        `json:"effectText"`
	RarityTier    Rarity        `json:"rarityTier"`
	IsCaptureType bool          `json:"isCaptureType"`
	ItemType      ItemType      `json:"itemType"`
	EquipmentSlot EquipmentSlot `json:"equipmentSlot,omitempty"`
}

type Rarity string

const (
	RarityCommon   Rarity = "Common"
	RarityUncommon Rarity = "Uncommon"
	RarityEpic     Rarity = "Epic"
	RarityMythic   Rarity = "Mythic"
	NotDroppable   Rarity = "Not Droppable"
)

type ItemType string

const (
	ItemTypePassive     ItemType = "passive"
	ItemTypeConsumable  ItemType = "consumable"
	ItemTypeEquippable  ItemType = "equippable"
)

type EquipmentSlot string

const (
	EquipmentSlotHead      EquipmentSlot = "head"
	EquipmentSlotChest     EquipmentSlot = "chest"
	EquipmentSlotLegs      EquipmentSlot = "legs"
	EquipmentSlotFeet      EquipmentSlot = "feet"
	EquipmentSlotLeftHand  EquipmentSlot = "left_hand"
	EquipmentSlotRightHand EquipmentSlot = "right_hand"
	EquipmentSlotNeck      EquipmentSlot = "neck"
	EquipmentSlotRing      EquipmentSlot = "ring"
	EquipmentSlotBelt      EquipmentSlot = "belt"
	EquipmentSlotGloves    EquipmentSlot = "gloves"
)

var PreDefinedItems = []InventoryItem{
	{
		ID:            1,
		Name:          "Cipher of the Laughing Monkey",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/cipher.png",
		FlavorText:    "Unearthed in the heart of a dense jungle, this mysterious item lay among countless laughing skeletons.",
		EffectText:    "Deploy to sow confusion among your rivals by warping their clue texts into bewildering riddles.",
		RarityTier:    "Uncommon",
		IsCaptureType: false,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            2,
		Name:          "Golden Telescope",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/telescope-better.png",
		FlavorText:    "Legend has it that a artificer parted with his sight to create this so that others might see the stars.",
		EffectText:    "Instantly reveals a hidden point on the map. Tap this icon next to the \"I'm here!\" button on a hidden points of interest to use it.",
		RarityTier:    "Uncommon",
		IsCaptureType: false,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            3,
		Name:          "Flawed Ruby",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/flawed-ruby.png",
		FlavorText:    "This gem is chipped and disfigured, but will still fetch a decent price at market.",
		EffectText:    "Instantly captures a tier one challenge. Tap this icon next to the \"Submit Answer\" button on any unlocked tier one challenge to use it.",
		RarityTier:    "Uncommon",
		IsCaptureType: true,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            4,
		Name:          "Ruby",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/ruby.png",
		FlavorText:    "A gem, sparkling more red than the blood you had to spill to procure it.",
		EffectText:    "Instantly captures a tier two challenge. Tap this icon next to the \"Submit Answer\" button on any unlocked tier two challenge to use it.",
		RarityTier:    "Epic",
		IsCaptureType: true,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            5,
		Name:          "Brilliant Ruby",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/brilliant-ruby.png",
		FlavorText:    "You've hit the motherload! This gem will fetch a pirate's ransom.",
		EffectText:    "Instantly captures a tier three challenge. Tap this icon next to the \"Submit Answer\" button on any unlocked tier three challenge to use it.",
		RarityTier:    "Mythic",
		IsCaptureType: true,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            6,
		Name:          "Cortez's Cutlass",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/cortez-cutlass.png",
		FlavorText:    "A relic of the high seas, its blade still sharp enough to cut through the thickest of hides.",
		EffectText:    "Steal all of another team's items.",
		RarityTier:    "Not Droppable",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotRightHand,
	},
	{
		ID:            7,
		Name:          "Rusted Musket",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/rusted-musket.png",
		FlavorText:    "Found in a shipwreck, its barrel rusted and its stock worn.",
		EffectText:    "Use on an opponent to lower their score by 2.",
		RarityTier:    "Common",
		IsCaptureType: false,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            8,
		Name:          "Gold Coin",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/gold-coin.png",
		FlavorText:    "A coin of pure gold. The currency of the high seas.",
		EffectText:    "Hold in your inventory to increase your score by 1.",
		RarityTier:    "Common",
		IsCaptureType: false,
		ItemType:      ItemTypePassive,
	},
	{
		ID:            9,
		Name:          "Dagger",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/dagger.png",
		FlavorText:    "A small, sharp blade. It's not much, but it's better than nothing.",
		EffectText:    "Steal one item from an opponent at random.",
		RarityTier:    "Epic",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotLeftHand,
	},
	{
		ID:            10,
		Name:          "Damage",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/bullet-hole.png",
		FlavorText:    "You've been shot! Some ale will help.",
		EffectText:    "Decreases score by 2 while held in inventory.",
		RarityTier:    "Not Droppable",
		IsCaptureType: false,
		ItemType:      ItemTypePassive,
	},
	{
		ID:            11,
		Name:          "Entseed",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/entseed.png",
		FlavorText:    "This seed will grow into an Ent one day. For now, you can just bask in it's life energy.",
		EffectText:    "Increase score by 3 and neutralize the effects of Damage while held in inventory.",
		RarityTier:    "Not Droppable",
		IsCaptureType: false,
		ItemType:      ItemTypePassive,
	},
	{
		ID:            12,
		Name:          "Ale",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/ale.png",
		FlavorText:    "A hearty brew, made from the finest ingredients.",
		EffectText:    "Removes one damage when drank.",
		RarityTier:    "Uncommon",
		IsCaptureType: false,
		ItemType:      ItemTypeConsumable,
	},
	{
		ID:            13,
		Name:          "Witchflame",
		ImageURL:      "https://crew-points-of-interest.s3.us-east-1.amazonaws.com/witchflame.png",
		FlavorText:    "A flame that burns with a sinister glow.",
		EffectText:    "Removes all damage when held. Also increases score by 1 when held.",
		RarityTier:    "Not Droppable",
		IsCaptureType: false,
		ItemType:      ItemTypePassive,
	},
	{
		ID:            14,
		Name:          "Wicked Spellbook",
		ImageURL:      "https://crew-points-of-interest.s3.us-east-1.amazonaws.com/wicked-spellbook.png",
		FlavorText:    "The spellbook whispers to you. Ignore it.",
		EffectText:    "Steal all of another team's items.",
		RarityTier:    "Not Droppable",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotLeftHand,
	},
	{
		ID:            15,
		Name:          "The Compass of Peace",
		ImageURL:      "https://crew-points-of-interest.s3.us-east-1.amazonaws.com/compass-of-peace.png",
		FlavorText:    "Given to you by Shalimar the Merchant. The compass is said to point towards what the wearer needs most to heal.",
		EffectText:    "Negate up to 3 damage when held.",
		RarityTier:    "Not Droppable",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotNeck,
	},
	{
		ID:            16,
		Name:          "Pirate's Tricorn Hat",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/tricorn-hat.png",
		FlavorText:    "A weathered hat that has seen many adventures on the high seas. Its feathers still dance in the wind.",
		EffectText:    "Increases treasure finding by 10% when worn.",
		RarityTier:    "Uncommon",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotHead,
	},
	{
		ID:            17,
		Name:          "Captain's Coat",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/captains-coat.png",
		FlavorText:    "A noble coat worn by a captain of renown. Its brass buttons still gleam despite the salt and spray.",
		EffectText:    "Provides +5 defense against damage when worn.",
		RarityTier:    "Epic",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotChest,
	},
	{
		ID:            18,
		Name:          "Seafarer's Boots",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/seafarer-boots.png",
		FlavorText:    "Sturdy boots made for walking on both deck and shore. They've never failed their wearer.",
		EffectText:    "Increases movement speed by 15% when worn.",
		RarityTier:    "Common",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotFeet,
	},
	{
		ID:            19,
		Name:          "Enchanted Ring of Fortune",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/fortune-ring.png",
		FlavorText:    "A mysterious ring that pulses with magical energy. Those who wear it speak of incredible luck.",
		EffectText:    "Doubles reward chances for treasure hunting when worn.",
		RarityTier:    "Mythic",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotRing,
	},
	{
		ID:            20,
		Name:          "Leather Sailing Gloves",
		ImageURL:      "https://crew-points-of-interest.s3.amazonaws.com/sailing-gloves.png",
		FlavorText:    "Well-worn gloves that have handled countless ropes and rigging. They fit like a second skin.",
		EffectText:    "Improves grip and handling, reducing chance of dropping items by 50%.",
		RarityTier:    "Common",
		IsCaptureType: false,
		ItemType:      ItemTypeEquippable,
		EquipmentSlot: EquipmentSlotGloves,
	},
}
