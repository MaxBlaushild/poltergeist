import { PointOfInterest } from '@poltergeist/types';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';

export const mapCaptureTiers = (
  pointOfInterest: PointOfInterest,
  submissions: PointOfInterestChallengeSubmission[]
): { [key: number]: boolean } => {
  const completedForTier: { [key: number]: boolean } = {};
  pointOfInterest.pointOfInterestChallenges
    .sort((a, b) => a.tier - b.tier)
    .forEach((challenge) => {
      const completed = submissions.some(
        (submission) => submission.isCorrect
      );
      if (completed) {
        completedForTier[challenge.tier] = completed;
        for (let j = 0; j < challenge.tier; j++) {
          completedForTier[j] = true;
        }
      }
    });
  return completedForTier;
};
