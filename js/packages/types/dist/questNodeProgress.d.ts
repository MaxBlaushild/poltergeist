export interface QuestNodeProgress {
    id: string;
    questAcceptanceId: string;
    questNodeId: string;
    status?: 'active' | 'completed' | 'failed' | string | null;
    attemptCount?: number;
    lastFailedAt?: string | null;
    lastFailureReason?: string | null;
    completedAt?: string | null;
}
