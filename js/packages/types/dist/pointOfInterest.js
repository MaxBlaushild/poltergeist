export const getHighestFirstCompletedChallenge = (pointOfInterest, submissions) => {
    const sortedChallenges = pointOfInterest.pointOfInterestChallenges.sort((a, b) => b.tier - a.tier);
    let firstCorrectSubmission = null;
    let associatedChallenge = null;
    for (const challenge of sortedChallenges) {
        const correctSubmissions = submissions.filter(submission => submission.pointOfInterestChallengeId === challenge.id && submission.isCorrect);
        if ((correctSubmissions === null || correctSubmissions === void 0 ? void 0 : correctSubmissions.length) > 0) {
            firstCorrectSubmission = correctSubmissions[0];
            associatedChallenge = challenge;
            break;
        }
    }
    return { submission: firstCorrectSubmission, challenge: associatedChallenge };
};
