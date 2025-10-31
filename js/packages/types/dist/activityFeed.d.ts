export type ActivityType = 'level_up' | 'challenge_completed' | 'quest_completed' | 'item_received' | 'reputation_up';
export interface LevelUpActivityData {
    newLevel: number;
}
export interface ChallengeCompletedActivityData {
    challengeId: string;
    successful: boolean;
    reason: string;
    submitterId?: string;
    experienceAwarded: number;
    reputationAwarded: number;
    itemsAwarded: Array<{
        id: number;
        name: string;
        imageUrl: string;
    }>;
    goldAwarded: number;
    questId: string;
    questName: string;
    questCompleted: boolean;
    currentPOI: {
        id: string;
        name: string;
        imageURL: string;
    };
    nextPOI?: {
        id: string;
        name: string;
        imageURL: string;
    };
    zoneId: string;
    zoneName: string;
}
export interface QuestCompletedActivityData {
    questId: string;
    goldAwarded: number;
    itemAwarded?: {
        id: number;
        name: string;
        imageUrl: string;
    };
}
export interface ItemReceivedActivityData {
    itemId: number;
    itemName: string;
}
export interface ReputationUpActivityData {
    newLevel: number;
    zoneName: string;
    zoneId: string;
}
export type ActivityData = LevelUpActivityData | ChallengeCompletedActivityData | QuestCompletedActivityData | ItemReceivedActivityData | ReputationUpActivityData;
export interface ActivityFeed {
    id: string;
    userId: string;
    activityType: ActivityType;
    data: ActivityData;
    seen: boolean;
    createdAt: Date;
    updatedAt: Date;
}
export declare function isLevelUpActivity(activity: ActivityFeed): activity is ActivityFeed & {
    data: LevelUpActivityData;
};
export declare function isChallengeCompletedActivity(activity: ActivityFeed): activity is ActivityFeed & {
    data: ChallengeCompletedActivityData;
};
export declare function isQuestCompletedActivity(activity: ActivityFeed): activity is ActivityFeed & {
    data: QuestCompletedActivityData;
};
export declare function isItemReceivedActivity(activity: ActivityFeed): activity is ActivityFeed & {
    data: ItemReceivedActivityData;
};
export declare function isReputationUpActivity(activity: ActivityFeed): activity is ActivityFeed & {
    data: ReputationUpActivityData;
};
