export type User = {
    phoneNumber: string;
    name: string;
    id: string;
    profilePictureUrl: string;
    partyId: string | null;
    username: string;
    isActive: boolean | null;
    gold: number;
};
export type UserZoneReputationSummary = {
    zoneId: string;
    level: number;
    totalReputation: number;
    reputationOnLevel: number;
    updatedAt: string;
};
