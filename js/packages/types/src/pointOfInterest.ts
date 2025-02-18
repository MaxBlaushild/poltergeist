import { PointOfInterestChallenge } from "./pointOfInterestChallenge";
import { PointOfInterestChallengeSubmission } from "./pointOfInterestChallengeSubmission";
import { PointOfInterestDiscovery } from "./pointOfInterestDiscovery";

export interface PointOfInterest {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    clue: string;
    lat: string;
    lng: string;
    imageURL: string;
    description: string;
	pointOfInterestChallenges: PointOfInterestChallenge[];
}

export const getHighestFirstCompletedChallenge = (pointOfInterest: PointOfInterest, submissions: PointOfInterestChallengeSubmission[]): { submission: PointOfInterestChallengeSubmission | null, challenge: PointOfInterestChallenge | null } => {
	const sortedChallenges = pointOfInterest.pointOfInterestChallenges.sort((a, b) => b.tier - a.tier);
	let firstCorrectSubmission = null;
	let associatedChallenge = null;

	for (const challenge of sortedChallenges) {
		const correctSubmissions = submissions.filter(submission => submission.pointOfInterestChallengeId === challenge.id && submission.isCorrect);

		if (correctSubmissions?.length > 0) {
			firstCorrectSubmission = correctSubmissions[0];
			associatedChallenge = challenge;
			break;
		}
	}

	return { submission: firstCorrectSubmission, challenge: associatedChallenge };
};
