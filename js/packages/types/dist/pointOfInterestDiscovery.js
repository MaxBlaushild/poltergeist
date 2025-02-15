export const hasDiscoveredPointOfInterest = (pointOfInterestId, entityId, pointOfInterestDiscoveries) => {
    return pointOfInterestDiscoveries
        .filter((pointOfInterestDiscovery) => pointOfInterestDiscovery.pointOfInterestId === pointOfInterestId)
        .some((pointOfInterestDiscovery) => pointOfInterestDiscovery.userId === entityId || pointOfInterestDiscovery.teamId === entityId);
};
