import { PointOfInterestTeam } from "./pointOfInterestTeam";
export interface PointOfInterest {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    clue: string;
    tierOneChallenge: string;
    tierTwoChallenge: string;
    tierThreeChallenge: string;
    lat: string;
    lng: string;
    imageURL: string;
    description: string;
}
export declare const getControllingTeamForPoi: (pointOfInterest: PointOfInterest, pointOfInterestTeams: PointOfInterestTeam[]) => PointOfInterestTeam | null;
