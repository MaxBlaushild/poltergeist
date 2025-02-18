export interface PointOfInterestChallengeSubmission {
    id: string;
    pointOfInterestChallengeId: string;
    teamId?: string | null;
    createdAt: Date;
    updatedAt: Date;
    text: string;
    imageUrl: string;
    isCorrect?: boolean;
    userId?: string | null;
}
