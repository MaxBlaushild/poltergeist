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

export const hasTeamDiscoveredPointOfInterest = (team: Team | undefined, pointOfInterest: PointOfInterest) => {
  return team?.pointOfInterestTeams?.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.id);
};
