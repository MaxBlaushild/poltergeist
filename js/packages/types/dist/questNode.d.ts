import { PointOfInterest } from './pointOfInterest';
import { Exposition } from './exposition';
import { Character } from './character';
import { InventoryItem } from './inventoryItem';
export type QuestNodeSubmissionType = 'text' | 'photo' | 'video' | (string & {});
export type QuestNodeObjectiveType = 'challenge' | 'fetch_quest' | 'story_flag' | 'scenario' | 'exposition' | 'monster_encounter' | 'monster' | (string & {});
export interface QuestNodeFetchRequirement {
    inventoryItemId: number;
    quantity: number;
    inventoryItem?: InventoryItem | null;
}
export interface QuestNodeObjective {
    id: string;
    type: QuestNodeObjectiveType;
    prompt: string;
    description?: string;
    imageUrl?: string;
    thumbnailUrl?: string;
    reward?: number;
    inventoryItemId?: number | null;
    storyFlagKey?: string | null;
    submissionType?: QuestNodeSubmissionType;
    difficulty?: number;
    statTags?: string[];
    proficiency?: string | null;
    characterId?: string | null;
    characterName?: string;
    fetchRequirements?: QuestNodeFetchRequirement[];
}
export interface QuestNodeChallengeDetails {
    id: string;
    pointOfInterestId?: string | null;
    pointOfInterest?: PointOfInterest | null;
    latitude: number;
    longitude: number;
    question: string;
    description?: string;
    submissionType?: QuestNodeSubmissionType;
    difficulty?: number;
    statTags?: string[];
    proficiency?: string | null;
    reward?: number;
    rewardExperience?: number;
}
export interface QuestNodeScenarioOptionDetails {
    id: string;
}
export interface QuestNodeScenarioDetails {
    id: string;
    zoneId: string;
    pointOfInterestId?: string | null;
    pointOfInterest?: PointOfInterest | null;
    latitude: number;
    longitude: number;
    prompt: string;
    difficulty?: number;
    openEnded?: boolean;
    options?: QuestNodeScenarioOptionDetails[];
}
export interface QuestNodeExpositionDetails extends Pick<Exposition, 'id' | 'zoneId' | 'pointOfInterestId' | 'pointOfInterest' | 'latitude' | 'longitude' | 'title' | 'description' | 'dialogue' | 'imageUrl' | 'thumbnailUrl' | 'rewardMode' | 'randomRewardSize' | 'rewardExperience' | 'rewardGold' | 'materialRewards' | 'itemRewards' | 'spellRewards'> {
}
export interface QuestNodeMonsterMemberDetails {
    slot: number;
    monster: {
        id: string;
        name: string;
    };
}
export interface QuestNodeMonsterEncounterDetails {
    id: string;
    zoneId: string;
    pointOfInterestId?: string | null;
    pointOfInterest?: PointOfInterest | null;
    latitude: number;
    longitude: number;
    name: string;
    description?: string;
    encounterType?: string;
    scaleWithUserLevel?: boolean;
    monsterCount?: number;
    members?: QuestNodeMonsterMemberDetails[];
}
export interface QuestNodeMonsterDetails {
    id: string;
    zoneId?: string;
    latitude: number;
    longitude: number;
    name: string;
    description?: string;
    level?: number;
}
export interface QuestNode {
    id: string;
    questId: string;
    orderIndex: number;
    submissionType?: QuestNodeSubmissionType;
    objectiveText?: string;
    objectiveDescription?: string;
    objective?: QuestNodeObjective | null;
    pointOfInterestId?: string | null;
    fetchCharacterId?: string | null;
    fetchCharacter?: Character | null;
    storyFlagKey?: string | null;
    scenarioId?: string | null;
    expositionId?: string | null;
    monsterId?: string | null;
    monsterEncounterId?: string | null;
    challengeId?: string | null;
    polygon?: string | null;
    polygonPoints?: [number, number][];
    scenario?: QuestNodeScenarioDetails | null;
    exposition?: QuestNodeExpositionDetails | null;
    monsterEncounter?: QuestNodeMonsterEncounterDetails | null;
    monster?: QuestNodeMonsterDetails | null;
    challenge?: QuestNodeChallengeDetails | null;
}
