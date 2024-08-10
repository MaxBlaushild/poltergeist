import { PointOfInterestTeam } from "./pointOfInterestTeam";
import { User } from "./user";

export type Team = {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  name: string;
  users: User[];
  pointOfInterestTeams: PointOfInterestTeam[];
};
