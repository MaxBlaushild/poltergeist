export var Rarity;
(function (Rarity) {
    Rarity["Common"] = "Common";
    Rarity["Uncommon"] = "Uncommon";
    Rarity["Epic"] = "Epic";
    Rarity["Mythic"] = "Mythic";
})(Rarity || (Rarity = {}));
;
export var ItemType;
(function (ItemType) {
    ItemType[ItemType["CipherOfTheLaughingMonkey"] = 1] = "CipherOfTheLaughingMonkey";
    ItemType[ItemType["GoldenTelescope"] = 2] = "GoldenTelescope";
    ItemType[ItemType["FlawedRuby"] = 3] = "FlawedRuby";
    ItemType[ItemType["Ruby"] = 4] = "Ruby";
    ItemType[ItemType["BrilliantRuby"] = 5] = "BrilliantRuby";
    ItemType[ItemType["CortezsCutlass"] = 6] = "CortezsCutlass";
    ItemType[ItemType["RustedMusket"] = 7] = "RustedMusket";
    ItemType[ItemType["GoldCoin"] = 8] = "GoldCoin";
})(ItemType || (ItemType = {}));
;
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
