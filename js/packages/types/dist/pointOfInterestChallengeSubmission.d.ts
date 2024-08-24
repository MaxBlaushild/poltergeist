export interface PointOfInterestChallengeSubmission {
    id: string;
    pointOfInterestChallengeId: string;
    teamId: string;
    createdAt: Date;
    updatedAt: Date;
    text: string;
    imageUrl: string;
    isCorrect?: boolean;
}
