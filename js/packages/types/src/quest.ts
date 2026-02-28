import { InventoryItem } from './inventoryItem';
import { QuestNode } from './questNode';
import { Spell } from './spell';

export interface QuestItemReward {
  id?: string;
  questId?: string;
  inventoryItemId: number;
  inventoryItem?: InventoryItem;
  quantity: number;
}

export interface QuestSpellReward {
  id?: string;
  questId?: string;
  spellId: string;
  spell?: Spell;
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
  recurringQuestId?: string | null;
  recurrenceFrequency?: string | null;
  nextRecurrenceAt?: string | null;
  completionCount?: number;
  gold?: number;
  itemRewards?: QuestItemReward[];
  spellRewards?: QuestSpellReward[];
  nodes?: QuestNode[];
}
