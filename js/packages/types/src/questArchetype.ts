import { InventoryItem } from './inventoryItem';
import { LocationArchetype } from './locationArchetype';
import { QuestDifficultyMode } from './questDifficulty';
import { CharacterRelationshipState, QuestMaterialReward } from './quest';
import { Spell } from './spell';
import { Character } from './character';
import type { DialogueMessage } from './characterAction';
import type {
  ExpositionMaterialReward,
  ExpositionRewardMode,
  ExpositionRandomRewardSize,
} from './exposition';

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
  | 'scenario'
  | 'exposition';

export interface QuestArchetypeNodeExpositionItemReward {
  inventoryItemId: number;
  quantity: number;
}

export interface QuestArchetypeNodeExpositionSpellReward {
  spellId: string;
}

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
  expositionTitle?: string | null;
  expositionDescription?: string | null;
  expositionDialogue?: DialogueMessage[];
  expositionRewardMode?: ExpositionRewardMode;
  expositionRandomRewardSize?: ExpositionRandomRewardSize;
  expositionRewardExperience?: number | null;
  expositionRewardGold?: number | null;
  expositionMaterialRewards?: ExpositionMaterialReward[];
  expositionItemRewards?: QuestArchetypeNodeExpositionItemReward[];
  expositionSpellRewards?: QuestArchetypeNodeExpositionSpellReward[];
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
  acceptanceDialogue?: DialogueMessage[];
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
  requiredStoryFlags?: string[];
  setStoryFlags?: string[];
  clearStoryFlags?: string[];
  questGiverRelationshipEffects?: CharacterRelationshipState;
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
