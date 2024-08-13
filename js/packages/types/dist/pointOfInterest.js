export const getControllingTeamForPoi = (pointOfInterest, pointOfInterestTeams) => {
    const teams = pointOfInterestTeams.filter(team => team.pointOfInterestId === pointOfInterest.id);
    if (!teams.length) {
        return null;
    }
    const highestTierTeam = teams.reduce((prev, current) => {
        return (current.captureTier > prev.captureTier) ? current : prev;
    });
    return highestTierTeam;
};
