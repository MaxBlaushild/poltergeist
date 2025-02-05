export type InventoryItem = {
    id: ItemType;
    name: string;
    imageUrl: string;
    flavorText: string;
    effectText: string;
    rarityTier: Rarity;
};
export declare enum Rarity {
    Common = "Common",
    Uncommon = "Uncommon",
    Epic = "Epic",
    Mythic = "Mythic"
}
export declare enum ItemType {
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
    Witchflame = 13
}
export declare const ItemsUsabledInMenu: ItemType[];
export declare const PointOfInterestEffectingItems: ItemType[];
export declare const ItemsUsabledOnPointOfInterest: ItemType[];
export declare const ItemsRequiringTeamId: ItemType[];
