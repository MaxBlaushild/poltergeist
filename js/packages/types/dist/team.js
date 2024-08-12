export const hasTeamDiscoveredPointOfInterest = (team, pointOfInterest) => {
    return team === null || team === void 0 ? void 0 : team.pointOfInterestTeams.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.id);
};
export const hasTeamCapturedPointOfInterest = (team, pointOfInterest) => {
    return team === null || team === void 0 ? void 0 : team.pointOfInterestTeams.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.id && pointOfInterestTeam.captured);
};
export const hasTeamAttunedPointOfInterest = (team, pointOfInterest) => {
    return team === null || team === void 0 ? void 0 : team.pointOfInterestTeams.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.id && pointOfInterestTeam.attuned);
};
