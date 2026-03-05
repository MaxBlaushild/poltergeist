export type InventoryConsumeStatus = {
  name: string;
  description: string;
  effect: string;
  positive: boolean;
  durationSeconds: number;
  strengthMod: number;
  dexterityMod: number;
  constitutionMod: number;
  intelligenceMod: number;
  wisdomMod: number;
  charismaMod: number;
};

type DamageAffinity =
  | 'physical'
  | 'fire'
  | 'ice'
  | 'lightning'
  | 'poison'
  | 'arcane'
  | 'holy'
  | 'shadow'
  | string;

export type InventoryItem = {
  id: number;
  name: string;
  imageUrl: string;
  flavorText: string;
  effectText: string;
  rarityTier: Rarity | string;
  isCaptureType: boolean;
  sellValue?: number;
  unlockTier?: number;
  itemLevel?: number;
  equipSlot?: string | null;
  strengthMod?: number;
  dexterityMod?: number;
  constitutionMod?: number;
  intelligenceMod?: number;
  wisdomMod?: number;
  charismaMod?: number;
  handItemCategory?: string | null;
  handedness?: string | null;
  damageMin?: number | null;
  damageMax?: number | null;
  damageAffinity?: DamageAffinity | null;
  swipesPerAttack?: number | null;
  blockPercentage?: number | null;
  damageBlocked?: number | null;
  spellDamageBonusPercent?: number | null;
  consumeHealthDelta?: number;
  consumeManaDelta?: number;
  consumeStatusesToAdd?: InventoryConsumeStatus[];
  consumeStatusesToRemove?: string[];
  consumeSpellIds?: string[];
  internalTags?: string[];
  imageGenerationStatus?: string;
  imageGenerationError?: string;
  createdAt?: string;
  updatedAt?: string;
};

export enum Rarity {
  Common = "Common",
  Uncommon = "Uncommon",
  Epic = "Epic",
  Mythic = "Mythic",
  NotDroppable = "Not Droppable"
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
