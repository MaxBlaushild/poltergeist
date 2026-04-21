import { DialogueMessage } from './characterAction';
import { PointOfInterest } from './pointOfInterest';
export type ExpositionRewardMode = 'explicit' | 'random';
export type ExpositionRandomRewardSize = 'small' | 'medium' | 'large';
export interface ExpositionMaterialReward {
    resourceKey: string;
    amount: number;
}
export interface ExpositionItemReward {
    id?: string;
    inventoryItemId: number;
    quantity: number;
}
export interface ExpositionSpellReward {
    id?: string;
    spellId: string;
    spell?: {
        id: string;
        name: string;
        iconUrl?: string;
    };
}
export interface Exposition {
    id: string;
    createdAt?: Date | string;
    updatedAt?: Date | string;
    zoneId: string;
    zoneKind?: string;
    pointOfInterestId?: string | null;
    pointOfInterest?: PointOfInterest | null;
    latitude: number;
    longitude: number;
    title: string;
    description: string;
    dialogue: DialogueMessage[];
    requiredStoryFlags?: string[];
    imageUrl: string;
    thumbnailUrl: string;
    rewardMode?: ExpositionRewardMode;
    randomRewardSize?: ExpositionRandomRewardSize;
    rewardExperience?: number;
    rewardGold?: number;
    materialRewards?: ExpositionMaterialReward[];
    itemRewards?: ExpositionItemReward[];
    spellRewards?: ExpositionSpellReward[];
}
