export type InventoryItem = {
  id: number;
  name: string;
  imageUrl: string;
  flavorText: string;
  effectText: string;
  rarityTier: Rarity;
  isCaptureType?: boolean;
  itemType?: string;
  equipmentSlot?: string;
};

export enum Rarity {
  Common = "Common",
  Uncommon = "Uncommon",
  Epic = "Epic",
  Mythic = "Mythic"
};

export enum ItemType {
  CipherOfTheLaughingMonkey = 1,
  GoldenTelescope = 2,
  FlawedRuby = 3,
  Ruby = 4,
  BrilliantRuby = 5,
  CortezsCutlass = 6,
  RustedMusket = 7,
  GoldCoin = 8,
  Dagger = 9,
	Damage = 10,
	Entseed = 11,
	Ale = 12,
  Witchflame = 13,
  WickedSpellbook = 14,
  CompassOfPeace = 15,
};

export const ItemsUsabledInMenu = [
	14, // WickedSpellbook
	1,  // CipherOfTheLaughingMonkey
	6,  // CortezsCutlass
	7,  // RustedMusket
	9,  // Dagger
	12, // Ale
];

export const PointOfInterestEffectingItems = [
	1, // CipherOfTheLaughingMonkey
];

export const ItemsUsabledOnPointOfInterest = [
	2, // GoldenTelescope
];

export const ItemsRequiringTeamId = [
	14, // WickedSpellbook
	6,  // CortezsCutlass
	7,  // RustedMusket
	9,  // Dagger
];
