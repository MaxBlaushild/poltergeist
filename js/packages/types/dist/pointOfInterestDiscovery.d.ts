export interface PointOfInterestDiscovery {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    teamId: string | undefined;
    userId: string | undefined;
    pointOfInterestId: string;
}
export declare const hasDiscoveredPointOfInterest: (pointOfInterestId: string, entityId: string, pointOfInterestDiscoveries: PointOfInterestDiscovery[]) => boolean;
