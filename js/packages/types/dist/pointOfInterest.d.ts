import { PointOfInterestChallenge } from './pointOfInterestChallenge';
import { PointOfInterestChallengeSubmission } from './pointOfInterestChallengeSubmission';
import { Tag } from './tag';
import { ZoneGenre } from './zone';
export type PointOfInterestRewardMode = 'explicit' | 'random';
export type PointOfInterestRandomRewardSize = 'small' | 'medium' | 'large';
export interface PointOfInterestMaterialReward {
    resourceKey: string;
    amount: number;
}
export interface PointOfInterestItemReward {
    id?: string;
    inventoryItemId: number;
    quantity: number;
}
export interface PointOfInterestSpellReward {
    id?: string;
    spellId: string;
    spell?: {
        id: string;
        name: string;
        iconUrl?: string;
    };
}
export interface PointOfInterestStoryVariant {
    id?: string;
    createdAt?: Date;
    updatedAt?: Date;
    priority?: number;
    requiredStoryFlags?: string[];
    description?: string;
    clue?: string;
}
export interface PointOfInterest {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    clue: string;
    lat: string;
    lng: string;
    imageURL: string;
    thumbnailUrl?: string;
    imageGenerationStatus?: string;
    imageGenerationError?: string | null;
    description: string;
    pointOfInterestChallenges: PointOfInterestChallenge[];
    genreId: string;
    genre?: ZoneGenre | null;
    zoneKind?: string;
    tags: Tag[];
    googleMapsPlaceId: string;
    googleMapsPlaceName?: string | null;
    originalName: string;
    geometry: string;
    unlockTier?: number | null;
    storyVariants?: PointOfInterestStoryVariant[];
    rewardMode?: PointOfInterestRewardMode;
    randomRewardSize?: PointOfInterestRandomRewardSize;
    rewardExperience?: number;
    rewardGold?: number;
    materialRewards?: PointOfInterestMaterialReward[];
    itemRewards?: PointOfInterestItemReward[];
    spellRewards?: PointOfInterestSpellReward[];
}
export declare const getHighestFirstCompletedChallenge: (pointOfInterest: PointOfInterest, submissions: PointOfInterestChallengeSubmission[]) => {
    submission: PointOfInterestChallengeSubmission | null;
    challenge: PointOfInterestChallenge | null;
};
