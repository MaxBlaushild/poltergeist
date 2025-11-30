export var Rarity;
(function (Rarity) {
    Rarity["Common"] = "Common";
    Rarity["Uncommon"] = "Uncommon";
    Rarity["Epic"] = "Epic";
    Rarity["Mythic"] = "Mythic";
    Rarity["NotDroppable"] = "Not Droppable";
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
    ItemType[ItemType["Dagger"] = 9] = "Dagger";
    ItemType[ItemType["Damage"] = 10] = "Damage";
    ItemType[ItemType["Entseed"] = 11] = "Entseed";
    ItemType[ItemType["Ale"] = 12] = "Ale";
    ItemType[ItemType["Witchflame"] = 13] = "Witchflame";
    ItemType[ItemType["WickedSpellbook"] = 14] = "WickedSpellbook";
    ItemType[ItemType["CompassOfPeace"] = 15] = "CompassOfPeace";
})(ItemType || (ItemType = {}));
;
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
