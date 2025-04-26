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
}
export interface QuestLog {
    quests: Quest[];
}
