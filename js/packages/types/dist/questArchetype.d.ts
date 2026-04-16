import { InventoryItem } from './inventoryItem';
import { LocationArchetype } from './locationArchetype';
import { QuestDifficultyMode } from './questDifficulty';
import { CharacterRelationshipState, QuestClosurePolicy, QuestDebriefPolicy, QuestMaterialReward } from './quest';
import { Spell } from './spell';
import { Character } from './character';
import { CharacterTemplate } from './characterTemplate';
import type { DialogueMessage } from './characterAction';
import type { ExpositionMaterialReward, ExpositionRewardMode, ExpositionRandomRewardSize } from './exposition';
import type { ExpositionTemplate } from './expositionTemplate';
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
    failureUnlockedNodeId?: string | null;
    failureUnlockedNode?: QuestArchetypeNode | null;
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
export type QuestArchetypeNodeType = 'challenge' | 'monster_encounter' | 'scenario' | 'exposition' | 'fetch_quest' | 'story_flag';
export interface QuestArchetypeNodeFetchRequirement {
    inventoryItemId: number;
    quantity: number;
}
export type QuestArchetypeNodeLocationSelectionMode = 'random' | 'closest' | 'same_as_previous';
export type QuestMonsterEncounterType = 'monster' | 'boss' | 'raid' | (string & {});
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
    locationSelectionMode?: QuestArchetypeNodeLocationSelectionMode;
    challengeTemplateId?: string | null;
    challengeTemplate?: QuestArchetypeChallengeTemplate | null;
    scenarioTemplateId?: string | null;
    fetchCharacter?: Character | null;
    fetchCharacterId?: string | null;
    fetchCharacterTemplate?: CharacterTemplate | null;
    fetchCharacterTemplateId?: string | null;
    fetchRequirements?: QuestArchetypeNodeFetchRequirement[];
    objectiveDescription?: string | null;
    failurePolicy?: 'retry' | 'transition' | string | null;
    storyFlagKey?: string | null;
    monsterTemplateIds?: string[];
    encounterType?: QuestMonsterEncounterType | null;
    targetLevel?: number | null;
    encounterProximityMeters?: number | null;
    expositionTemplate?: ExpositionTemplate | null;
    expositionTemplateId?: string | null;
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
    closurePolicy?: QuestClosurePolicy;
    debriefPolicy?: QuestDebriefPolicy;
    returnBonusGold?: number;
    returnBonusExperience?: number;
    returnBonusRelationshipEffects?: CharacterRelationshipState;
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
