export type InventoryItem = {
  id: ItemType;
  name: string;
  imageUrl: string;
  flavorText: string;
  effectText: string;
  rarityTier: Rarity;
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
  GoldCoin = 8
};

export const ItemsUsabledInMenu = [
	ItemType.CipherOfTheLaughingMonkey,
	ItemType.CortezsCutlass,
	ItemType.RustedMusket,
];

export const PointOfInterestEffectingItems = [
	ItemType.CipherOfTheLaughingMonkey,
];

export const ItemsUsabledOnPointOfInterest = [
	ItemType.GoldenTelescope,
];

export const ItemsRequiringTeamId = [
	ItemType.CortezsCutlass,
	ItemType.RustedMusket,
];
