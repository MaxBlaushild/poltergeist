import { PointOfInterest } from './pointOfInterest';
import { PointOfInterestChallenge } from './pointOfInterestChallenge';
import { PointOfInterestChallengeSubmission } from './pointOfInterestChallengeSubmission';
export interface QuestObjective {
    challenge: PointOfInterestChallenge;
    isCompleted: boolean;
    submissions: PointOfInterestChallengeSubmission[];
    nextNode?: QuestNode | null;
}
export interface QuestNode {
    pointOfInterest: PointOfInterest;
    objectives: QuestObjective[];
    children: Record<string, QuestNode>;
}
export interface Quest {
    isCompleted: boolean;
    rootNode: QuestNode;
    imageUrl: string;
    name: string;
    description: string;
    id: string;
}
export interface QuestLog {
    quests: Quest[];
    pendingTasks: Record<string, Task[]>;
    completedTasks: Record<string, Task[]>;
    trackedQuestIds: string[];
}
export interface Task {
    challenge: PointOfInterestChallenge;
    questId: string;
}
export declare function getQuestTags(quest: Quest): string[];
