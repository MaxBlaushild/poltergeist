import { PointOfInterestChallenge } from "./pointOfInterestChallenge";
import { PointOfInterestChallengeSubmission } from "./pointOfInterestChallengeSubmission";
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
}
export declare const getControllingTeamForPoi: (pointOfInterest: PointOfInterest) => {
    submission: PointOfInterestChallengeSubmission | null;
    challenge: PointOfInterestChallenge | null;
};
