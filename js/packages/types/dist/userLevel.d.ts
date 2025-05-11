export interface UserLevel {
    id: string;
    createdAt: string;
    updatedAt: string;
    userId: string;
    level: number;
    experiencePointsOnLevel: number;
    totalExperiencePoints: number;
    levelsGained: number;
    experienceToNextLevel: number;
}
