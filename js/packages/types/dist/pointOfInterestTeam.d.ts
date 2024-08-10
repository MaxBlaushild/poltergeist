export interface PointOfInterestTeam {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    teamId: string;
    pointOfInterestId: string;
    unlocked: boolean;
    captured: boolean;
    attuned: boolean;
}
