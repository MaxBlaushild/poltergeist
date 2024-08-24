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

export const hasTeamDiscoveredPointOfInterest = (team: Team | undefined, pointOfInterest: PointOfInterest) => {
  return team?.pointOfInterestTeams?.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.id);
};
