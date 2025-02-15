export interface PointOfInterestDiscovery {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    teamId: string;
    userId: string;
    pointOfInterestId: string;
}
export declare const hasDiscoveredPointOfInterest: (pointOfInterestId: string, entityId: string, pointOfInterestDiscoveries: PointOfInterestDiscovery[]) => boolean;
