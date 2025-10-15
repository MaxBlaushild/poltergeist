// Type guards
export function isLevelUpActivity(activity) {
    return activity.activityType === 'level_up';
}
export function isChallengeCompletedActivity(activity) {
    return activity.activityType === 'challenge_completed';
}
export function isQuestCompletedActivity(activity) {
    return activity.activityType === 'quest_completed';
}
export function isItemReceivedActivity(activity) {
    return activity.activityType === 'item_received';
}
export function isReputationUpActivity(activity) {
    return activity.activityType === 'reputation_up';
}
