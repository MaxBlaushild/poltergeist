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
  gold: number | null;
  geometry: string;
  items: TreasureChestItem[];
}

