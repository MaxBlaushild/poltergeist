import { PointOfInterest } from "./pointOfInterest";
export type PointOfInterestGroup = {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    pointsOfInterest: PointOfInterest[];
    description: string;
};
