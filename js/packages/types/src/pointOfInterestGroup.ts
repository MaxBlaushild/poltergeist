import { PointOfInterest } from "./pointOfInterest";
import { PointOfInterestGroupMember } from "./pointOfInterestGroupMember";

export enum PointOfInterestGroupType {
  Unassigned = 0,
  Arena = 1,
  Quest = 2,
}

export type PointOfInterestGroup = {
  id: string; 
  createdAt: Date;
  updatedAt: Date;
  name: string;
  pointsOfInterest: PointOfInterest[];
  description: string;
  imageUrl: string;
  groupMembers: PointOfInterestGroupMember[];
  type: PointOfInterestGroupType;
};