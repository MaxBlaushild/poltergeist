import { PointOfInterestChallengeSubmission } from "./pointOfInterestChallengeSubmission";

export interface PointOfInterestChallenge {
    id: string;
    pointOfInterestId: string;
    pointOfInterestGroupId?: string;
    pointOfInterestChallengeSubmissions: PointOfInterestChallengeSubmission[];
    question: string;
    tier: number;
    createdAt: Date;
    updatedAt: Date;
    inventoryItemId: number;
}
