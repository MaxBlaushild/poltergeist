export const getControllingTeamForPoi = (pointOfInterest) => {
    var _a;
    const sortedChallenges = pointOfInterest.pointOfInterestChallenges.sort((a, b) => b.tier - a.tier);
    let firstCorrectSubmission = null;
    let associatedChallenge = null;
    for (const challenge of sortedChallenges) {
        const correctSubmissions = (_a = challenge.pointOfInterestChallengeSubmissions) === null || _a === void 0 ? void 0 : _a.filter(submission => submission.isCorrect);
        if ((correctSubmissions === null || correctSubmissions === void 0 ? void 0 : correctSubmissions.length) > 0) {
            firstCorrectSubmission = correctSubmissions[0];
            associatedChallenge = challenge;
            break;
        }
    }
    return { submission: firstCorrectSubmission, challenge: associatedChallenge };
};
