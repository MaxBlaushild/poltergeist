const getControllingTeamForPoi = (pointOfInterest, pointOfInterestTeams) => {
    const teams = pointOfInterestTeams.filter(team => team.pointOfInterestId === pointOfInterest.id);
    if (!teams.length) {
        return null;
    }
    const highestTierTeam = teams.reduce((prev, current) => {
        const prevMaxTier = Math.max(prev.tierOneCaptured ? 1 : 0, prev.tierTwoCaptured ? 2 : 0, prev.tierThreeCaptured ? 3 : 0);
        const currentMaxTier = Math.max(current.tierOneCaptured ? 1 : 0, current.tierTwoCaptured ? 2 : 0, current.tierThreeCaptured ? 3 : 0);
        return (currentMaxTier > prevMaxTier) ? current : prev;
    });
    return highestTierTeam;
};
export {};
