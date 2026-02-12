import { InventoryItem } from './inventoryItem';
import { QuestNode } from './questNode';

export interface QuestItemReward {
  id?: string;
  questId?: string;
  inventoryItemId: number;
  inventoryItem?: InventoryItem;
  quantity: number;
}

export interface Quest {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  acceptanceDialogue?: string[];
  imageUrl?: string;
  zoneId?: string | null;
  questArchetypeId?: string | null;
  questGiverCharacterId?: string | null;
  gold?: number;
  itemRewards?: QuestItemReward[];
  nodes?: QuestNode[];
}
