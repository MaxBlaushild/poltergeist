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

export const getControllingTeamForPoi = (pointOfInterest: PointOfInterest, pointOfInterestTeams: PointOfInterestTeam[]) => {
	const teams = pointOfInterestTeams.filter(team => team.pointOfInterestId === pointOfInterest.id);
	if (!teams.length) {
		return null;
	}
	const highestTierTeam = teams.reduce((prev, current) => {
		return (current.captureTier > prev.captureTier) ? current : prev;
	});
	return highestTierTeam;
}

