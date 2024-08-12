export const hasTeamDiscoveredPointOfInterest = (team, pointOfInterest) => {
    return team === null || team === void 0 ? void 0 : team.pointOfInterestTeams.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.ID);
};
export const hasTeamCapturedPointOfInterest = (team, pointOfInterest) => {
    return team === null || team === void 0 ? void 0 : team.pointOfInterestTeams.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.ID && pointOfInterestTeam.captured);
};
export const hasTeamAttunedPointOfInterest = (team, pointOfInterest) => {
    return team === null || team === void 0 ? void 0 : team.pointOfInterestTeams.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.ID && pointOfInterestTeam.attuned);
};
