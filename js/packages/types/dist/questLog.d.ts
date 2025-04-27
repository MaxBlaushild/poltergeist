import { PointOfInterest } from './pointOfInterest';
import { PointOfInterestChallenge } from './pointOfInterestChallenge';
import { PointOfInterestChallengeSubmission } from './pointOfInterestChallengeSubmission';
export interface QuestObjective {
    challenge: PointOfInterestChallenge;
    isCompleted: boolean;
    submissions: PointOfInterestChallengeSubmission[];
}
export interface QuestNode {
    pointOfInterest: PointOfInterest;
    objectives: QuestObjective[];
    children: {
        [key: string]: QuestNode;
    };
}
export interface Quest {
    isCompleted: boolean;
    rootNode: QuestNode;
    imageUrl: string;
    name: string;
    description: string;
    id: string;
    getTags: () => string[];
}
export interface QuestLog {
    quests: Quest[];
    pendingTasks: Record<string, Task[]>;
    completedTasks: Record<string, Task[]>;
}
export interface Task {
    challenge: PointOfInterestChallenge;
    questId: string;
}
export declare function getQuestTags(quest: Quest): string[];
