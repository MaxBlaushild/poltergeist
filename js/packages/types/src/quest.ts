import { InventoryItem } from './inventoryItem';
import { QuestDifficultyMode } from './questDifficulty';
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

export interface QuestMaterialReward {
  resourceKey: string;
  amount: number;
}

export interface Quest {
  id: string;
  createdAt: string;
  updatedAt: string;
  nodeCount?: number;
  detailLoaded?: boolean;
  name: string;
  description: string;
  category?: 'side' | 'main_story';
  acceptanceDialogue?: string[];
  imageUrl?: string;
  zoneId?: string | null;
  questArchetypeId?: string | null;
  questGiverCharacterId?: string | null;
  mainStoryPreviousQuestId?: string | null;
  mainStoryNextQuestId?: string | null;
  recurringQuestId?: string | null;
  recurrenceFrequency?: string | null;
  nextRecurrenceAt?: string | null;
  completionCount?: number;
  difficultyMode?: QuestDifficultyMode;
  difficulty?: number;
  monsterEncounterTargetLevel?: number;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience?: number;
  gold?: number;
  materialRewards?: QuestMaterialReward[];
  itemRewards?: QuestItemReward[];
  spellRewards?: QuestSpellReward[];
  nodes?: QuestNode[];
}
