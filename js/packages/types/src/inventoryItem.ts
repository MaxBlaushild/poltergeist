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
  GoldCoin = 8,
  Dagger = 9,
	Damage = 10,
	Entseed = 11,
	Ale = 12,
  Witchflame = 13,
  WickedSpellbook = 14,
};

export const ItemsUsabledInMenu = [
	ItemType.WickedSpellbook,
	ItemType.CipherOfTheLaughingMonkey,
	ItemType.CortezsCutlass,
	ItemType.RustedMusket,
	ItemType.Dagger,
	ItemType.Ale,
];

export const PointOfInterestEffectingItems = [
	ItemType.CipherOfTheLaughingMonkey,
];

export const ItemsUsabledOnPointOfInterest = [
	ItemType.GoldenTelescope,
];

export const ItemsRequiringTeamId = [
	ItemType.WickedSpellbook,
	ItemType.CortezsCutlass,
	ItemType.RustedMusket,
	ItemType.Dagger,
];
