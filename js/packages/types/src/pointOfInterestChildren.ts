import { PointOfInterest } from "./pointOfInterest";
import { PointOfInterestChallenge } from "./pointOfInterestChallenge";

export interface PointOfInterestChildren {
    id: string;
    pointOfInterestGroupMemberId: string;
    pointOfInterestId: string;
    pointOfInterest: PointOfInterest;
    createdAt: Date;
    updatedAt: Date;
    pointOfInterestChallengeId: string;
    pointOfInterestChallenge: PointOfInterestChallenge;
}
