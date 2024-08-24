export const hasTeamDiscoveredPointOfInterest = (team, pointOfInterest) => {
    var _a;
    return (_a = team === null || team === void 0 ? void 0 : team.pointOfInterestTeams) === null || _a === void 0 ? void 0 : _a.some((pointOfInterestTeam) => pointOfInterestTeam.pointOfInterestId === pointOfInterest.id);
};
