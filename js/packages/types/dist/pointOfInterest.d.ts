import { PointOfInterestChallenge } from './pointOfInterestChallenge';
import { PointOfInterestChallengeSubmission } from './pointOfInterestChallengeSubmission';
import { Tag } from './tag';
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
    tags: Tag[];
    googleMapsPlaceId: string;
    googleMapsPlaceName?: string | null;
    originalName: string;
    geometry: string;
    unlockTier?: number | null;
    storyVariants?: PointOfInterestStoryVariant[];
}
export declare const getHighestFirstCompletedChallenge: (pointOfInterest: PointOfInterest, submissions: PointOfInterestChallengeSubmission[]) => {
    submission: PointOfInterestChallengeSubmission | null;
    challenge: PointOfInterestChallenge | null;
};
