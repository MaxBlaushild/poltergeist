export interface PointOfInterestTeam {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    teamId: string;
    pointOfInterestId: string;
    tierOneCaptured: boolean;
    tierTwoCaptured: boolean;
    tierThreeCaptured: boolean;
}
