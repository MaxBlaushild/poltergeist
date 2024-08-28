import { PointOfInterestTeam } from "./pointOfInterestTeam";
import { PointOfInterest } from "./pointOfInterest";
import { User } from "./user";
import { TeamInventoryItem } from "./teamInventoryItem";
export type Team = {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    users: User[];
    pointOfInterestTeams: PointOfInterestTeam[];
    teamInventoryItems: TeamInventoryItem[];
};
export declare const hasTeamDiscoveredPointOfInterest: (team: Team | undefined, pointOfInterest: PointOfInterest) => boolean | undefined;
