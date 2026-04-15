import { QuestClosureMethod } from './quest';
export interface QuestAcceptanceV2 {
    id: string;
    userId: string;
    questId: string;
    currentQuestNodeId?: string | null;
    acceptedAt: string;
    objectivesCompletedAt?: string | null;
    closedAt?: string | null;
    closureMethod?: QuestClosureMethod | null;
    debriefPending?: boolean;
    debriefedAt?: string | null;
    turnedInAt?: string | null;
}
