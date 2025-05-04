import { PointOfInterestChallenge } from "./pointOfInterestChallenge";
export interface PointOfInterestChildren {
    id: string;
    pointOfInterestGroupMemberId: string;
    createdAt: Date;
    updatedAt: Date;
    pointOfInterestChallengeId: string;
    pointOfInterestChallenge: PointOfInterestChallenge;
    nextPointOfInterestGroupMemberId: string;
}
