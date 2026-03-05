import { InventoryItem } from "./inventoryItem";
import { Zone } from "./zone";
export interface TreasureChestItem {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    treasureChestId: string;
    inventoryItemId: number;
    inventoryItem: InventoryItem;
    quantity: number;
}
export interface TreasureChest {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    latitude: number;
    longitude: number;
    zoneId: string;
    zone: Zone;
    rewardMode: 'explicit' | 'random';
    randomRewardSize: 'small' | 'medium' | 'large';
    rewardExperience: number;
    gold: number | null;
    geometry: string;
    unlockTier: number | null;
    items: TreasureChestItem[];
    openedByUser?: boolean;
}
