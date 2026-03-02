import { QuestNodeSubmissionType } from './questNode';
export interface QuestNodeChallenge {
    id: string;
    questNodeId: string;
    tier: number;
    question: string;
    reward: number;
    inventoryItemId?: number | null;
    submissionType?: QuestNodeSubmissionType;
    difficulty?: number;
    statTags?: string[];
    proficiency?: string | null;
    challengeShuffleStatus?: string;
    challengeShuffleError?: string | null;
}
