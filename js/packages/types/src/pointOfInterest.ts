import { PointOfInterestTeam } from "./pointOfInterestTeam";

export interface PointOfInterest {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    clue: string;
    tierOneChallenge: string;
    tierTwoChallenge: string;
    tierThreeChallenge: string;
    lat: string;
    lng: string;
    imageURL: string;
    description: string;
}

const getControllingTeamForPoi = (pointOfInterest: PointOfInterest, pointOfInterestTeams: PointOfInterestTeam[]) => {
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
}

