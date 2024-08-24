import { PointOfInterestTeam } from "./pointOfInterestTeam";
import { PointOfInterest } from "./pointOfInterest";
import { User } from "./user";
export type Team = {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    users: User[];
    pointOfInterestTeams: PointOfInterestTeam[];
};
export declare const hasTeamDiscoveredPointOfInterest: (team: Team | undefined, pointOfInterest: PointOfInterest) => boolean | undefined;
