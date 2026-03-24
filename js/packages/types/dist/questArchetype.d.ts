import { InventoryItem } from './inventoryItem';
import { LocationArchetype } from './locationArchetype';
import { QuestMaterialReward } from './quest';
import { Spell } from './spell';
export interface QuestArchetypeChallenge {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    reward: number;
    inventoryItemId?: number | null;
    proficiency?: string | null;
    difficulty?: number | null;
    unlockedNodeId?: string;
    unlockedNode?: QuestArchetypeNode;
}
export type QuestArchetypeNodeType = 'location' | 'monster_encounter' | 'scenario';
export interface QuestArchetypeNodeEncounterItemReward {
    inventoryItemId: number;
    quantity: number;
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
    encounterRewardMode?: 'explicit' | 'random';
    encounterRandomRewardSize?: 'small' | 'medium' | 'large';
    encounterRewardExperience?: number | null;
    encounterRewardGold?: number | null;
    encounterMaterialRewards?: QuestMaterialReward[];
    encounterItemRewards?: QuestArchetypeNodeEncounterItemReward[];
    encounterProximityMeters?: number | null;
    challenges: QuestArchetypeChallenge[];
    difficulty?: number | null;
}
export interface QuestArchetype {
    id: string;
    name: string;
    description: string;
    acceptanceDialogue?: string[];
    imageUrl?: string;
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
