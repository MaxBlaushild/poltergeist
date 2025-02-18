export interface PointOfInterestDiscovery {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    teamId: string | undefined;
    userId: string | undefined;
    pointOfInterestId: string;
}

export const hasDiscoveredPointOfInterest = (pointOfInterestId: string, entityId: string, pointOfInterestDiscoveries: PointOfInterestDiscovery[]) => {
    return pointOfInterestDiscoveries
        .filter((pointOfInterestDiscovery) => pointOfInterestDiscovery.pointOfInterestId === pointOfInterestId)
        .some((pointOfInterestDiscovery) => pointOfInterestDiscovery.userId === entityId || pointOfInterestDiscovery.teamId === entityId);
  };