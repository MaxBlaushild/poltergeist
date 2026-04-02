import { InventoryItem } from './inventoryItem';
import { LocationArchetype } from './locationArchetype';
import { QuestDifficultyMode } from './questDifficulty';
import { QuestMaterialReward } from './quest';
import { Spell } from './spell';
import { Character } from './character';

export interface QuestArchetypeChallenge {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  challengeTemplateId?: string | null;
  challengeTemplate?: QuestArchetypeChallengeTemplate | null;
  proficiency?: string | null;
  difficulty?: number | null;
  unlockedNodeId?: string;
  unlockedNode?: QuestArchetypeNode;
}

export interface QuestArchetypeChallengeTemplate {
  id: string;
  locationArchetypeId: string;
  question: string;
  description?: string;
  submissionType?: string;
  difficulty?: number | null;
  proficiency?: string | null;
}

export type QuestArchetypeNodeType =
  | 'location'
  | 'monster_encounter'
  | 'scenario';

export interface QuestArchetypeItemReward {
  id?: string;
  questArchetypeId?: string;
  inventoryItemId: number;
  inventoryItem?: InventoryItem;
  quantity: number;
}

export interface QuestArchetypeSpellReward {
  id?: string;
  questArchetypeId?: string;
  spellId: string;
  spell?: Spell;
}

export interface QuestArchetypeNode {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  nodeType?: QuestArchetypeNodeType;
  locationArchetype?: LocationArchetype | null;
  locationArchetypeId?: string | null;
  scenarioTemplateId?: string | null;
  monsterTemplateIds?: string[];
  targetLevel?: number | null;
  encounterProximityMeters?: number | null;
  challenges: QuestArchetypeChallenge[];
  difficulty?: number | null;
}

export interface QuestArchetype {
  id: string;
  name: string;
  description: string;
  category?: 'side' | 'main_story';
  questGiverCharacterId?: string | null;
  questGiverCharacter?: Character | null;
  acceptanceDialogue?: string[];
  imageUrl?: string;
  difficultyMode?: QuestDifficultyMode;
  difficulty?: number;
  monsterEncounterTargetLevel?: number;
  defaultGold: number;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience?: number;
  recurrenceFrequency?: string | null;
  materialRewards?: QuestMaterialReward[];
  characterTags?: string[];
  internalTags?: string[];
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  root: QuestArchetypeNode;
  rootId: string;
  itemRewards?: QuestArchetypeItemReward[];
  spellRewards?: QuestArchetypeSpellReward[];
}
