import { PointOfInterestChallenge } from "./pointOfInterestChallenge";
import { PointOfInterestChallengeSubmission } from "./pointOfInterestChallengeSubmission";
import { Tag } from "./tag";
export interface PointOfInterest {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    clue: string;
    lat: string;
    lng: string;
    imageURL: string;
    description: string;
    pointOfInterestChallenges: PointOfInterestChallenge[];
    tags: Tag[];
    googleMapsPlaceId: string;
    originalName: string;
    geometry: string;
    unlockTier?: number | null;
}
export declare const getHighestFirstCompletedChallenge: (pointOfInterest: PointOfInterest, submissions: PointOfInterestChallengeSubmission[]) => {
    submission: PointOfInterestChallengeSubmission | null;
    challenge: PointOfInterestChallenge | null;
};
