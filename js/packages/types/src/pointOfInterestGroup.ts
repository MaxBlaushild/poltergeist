import { PointOfInterest } from "./pointOfInterest";
import { PointOfInterestGroupMember } from "./pointOfInterestGroupMember";

export type PointOfInterestGroup = {
  id: string; 
  createdAt: Date;
  updatedAt: Date;
  name: string;
  pointsOfInterest: PointOfInterest[];
  description: string;
  imageUrl: string;
  groupMembers: PointOfInterestGroupMember[];
};
