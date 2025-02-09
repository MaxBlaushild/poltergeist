import { PointOfInterestChallengeSubmission } from "./pointOfInterestChallengeSubmission";
export interface PointOfInterestChallenge {
    id: string;
    pointOfInterestId: string;
    pointOfInterestChallengeSubmissions: PointOfInterestChallengeSubmission[];
    question: string;
    tier: number;
    createdAt: Date;
    updatedAt: Date;
    inventoryItemId: number;
}
