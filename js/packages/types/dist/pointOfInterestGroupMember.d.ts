import { PointOfInterest } from "./pointOfInterest";
import { PointOfInterestGroup } from "./pointOfInterestGroup";
import { PointOfInterestChildren } from "./pointOfInterestChildren";
export interface PointOfInterestGroupMember {
    id: string;
    pointOfInterestGroupId: string;
    pointOfInterestGroup: PointOfInterestGroup;
    pointOfInterestId: string;
    pointOfInterest: PointOfInterest;
    createdAt: Date;
    updatedAt: Date;
    children: PointOfInterestChildren[];
}
